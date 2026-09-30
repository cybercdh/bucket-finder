package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gookit/color"
	"golang.org/x/net/publicsuffix"
)

var (
	concurrency int
	depth       int
	verbose     bool
	external    bool

	// reported dedupes bucket names across every crawl so stdout is a clean,
	// unique list that can be piped into other tooling.
	reportedMu sync.Mutex
	reported   = make(map[string]bool)
)

func main() {
	flag.IntVar(&concurrency, "c", 50, "set the concurrency level")
	flag.IntVar(&depth, "d", 5, "set the crawling depth")
	flag.BoolVar(&verbose, "v", false, "See more info on attempts (printed to stderr)")
	flag.BoolVar(&external, "external", false, "follow page links to other domains as well (script files are always fetched wherever they live)")
	flag.Parse()

	if concurrency < 1 {
		fmt.Fprintf(os.Stderr, "-c must be at least 1 (got %d)\n", concurrency)
		os.Exit(2)
	}

	jobs := make(chan string, 100)

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				crawl(job)
			}
		}()
	}

	var input io.Reader = os.Stdin
	if arg := flag.Arg(0); arg != "" {
		input = strings.NewReader(arg)
	} else if info, err := os.Stdin.Stat(); err == nil && info.Mode()&os.ModeCharDevice != 0 {
		fmt.Fprintln(os.Stderr, "Expected usage: echo <domain> | bucket-finder   or   bucket-finder <domain>")
		os.Exit(2)
	}

	sc := bufio.NewScanner(input)
	for sc.Scan() {
		if u, ok := normalizeURL(sc.Text()); ok {
			jobs <- u
		}
	}
	close(jobs)
	if err := sc.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to read input: %s\n", err)
	}

	wg.Wait()
}

// crawl walks one seed URL to the configured depth, scanning every response
// body (HTML and JavaScript alike) for S3 bucket references.
func crawl(seed string) {
	scope := registrable(seed)

	c := colly.NewCollector(colly.MaxDepth(depth))
	c.SetRequestTimeout(15 * time.Second)

	c.OnHTML("a[href], script[src]", func(e *colly.HTMLElement) {
		var link string
		isScript := false
		if href := e.Attr("href"); href != "" {
			link = e.Request.AbsoluteURL(href)
		} else if src := e.Attr("src"); src != "" {
			link = e.Request.AbsoluteURL(src)
			isScript = true
		}
		if link == "" || shouldExclude(link) {
			return
		}
		// Page links stay on the seed's registrable domain unless -external is
		// set. Scripts are fetched from anywhere: a CDN-hosted bundle is exactly
		// where bucket references hide, and a script is a leaf (no HTML to
		// follow), so it cannot widen the crawl.
		if !isScript && !external && registrable(link) != scope {
			return
		}

		chain := append(append([]string(nil), e.Request.Ctx.GetAny("urlChain").([]string)...), link)
		ctx := colly.NewContext()
		ctx.Put("urlChain", chain)

		// Collector.Request always starts at depth 1, which made -d a no-op:
		// every link looked like a fresh crawl. Building the request from the
		// current one lets us carry the real depth so MaxDepth applies.
		req, err := e.Request.New("GET", link, nil)
		if err != nil {
			return
		}
		req.Ctx = ctx
		req.Depth = e.Request.Depth + 1
		_ = req.Do() // ErrMaxDepth and already-visited are expected here
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", RandomString(userAgentList))
		if verbose {
			fmt.Fprintln(os.Stderr, "Visiting", r.URL)
		}
	})

	c.OnResponse(func(r *colly.Response) {
		for _, name := range extractBuckets(r.Body) {
			report(name, r)
		}
	})

	ctx := colly.NewContext()
	ctx.Put("urlChain", []string{seed})
	if err := c.Request("GET", seed, nil, ctx, nil); err != nil && verbose {
		fmt.Fprintf(os.Stderr, "%s: %s\n", seed, err)
	}
}

// report prints a bucket name once. Verbose mode adds where it was seen and
// the chain of pages that led there, on stderr so stdout stays pipeable.
func report(name string, r *colly.Response) {
	reportedMu.Lock()
	first := !reported[name]
	reported[name] = true
	reportedMu.Unlock()

	if first {
		fmt.Println(name)
	}
	if !verbose {
		return
	}
	fmt.Fprint(os.Stderr, color.Green.Sprintf("S3 Bucket Found: %s\n", name))
	fmt.Fprint(os.Stderr, color.Green.Sprintf("At URL: %s\n", r.Request.URL))
	if chain, ok := r.Ctx.GetAny("urlChain").([]string); ok {
		fmt.Fprintln(os.Stderr, "URL Chain:")
		for _, u := range chain {
			fmt.Fprint(os.Stderr, color.Yellow.Sprintf("%s\n", u))
		}
	}
	fmt.Fprintln(os.Stderr, "------")
}

// normalizeURL trims an input line and gives it an https:// scheme when none
// is present. Detection is by "://": the old "http" prefix test treated a
// hostname such as httpbin.org as a URL.
func normalizeURL(line string) (string, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", false
	}
	if !strings.Contains(line, "://") {
		line = "https://" + line
	}
	u, err := url.Parse(line)
	if err != nil || u.Host == "" {
		return "", false
	}
	return u.String(), true
}

// registrable returns the eTLD+1 of a URL's host, or the host itself when the
// public suffix list has no answer (IPs, localhost, single labels).
func registrable(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	if root, err := publicsuffix.EffectiveTLDPlusOne(host); err == nil {
		return root
	}
	return host
}

func RandomString(userAgentList []string) string {
	return userAgentList[rand.Intn(len(userAgentList))]
}

func shouldExclude(link string) bool {
	for _, domain := range excludedDomains {
		if strings.Contains(link, domain) {
			return true
		}
	}
	return false
}
