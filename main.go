package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/gookit/color"
)

var jobs = make(chan string, 100)
var concurrency int
var depth int
var verbose bool

var urlChain = make(map[string]string)

func main() {
	flag.IntVar(&concurrency, "c", 50, "set the concurrency level")
	flag.IntVar(&depth, "d", 5, "set the crawling depth")
	flag.BoolVar(&verbose, "v", false, "See more info on attempts")
	flag.Parse()

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// iterate the user input
			for job := range jobs {

				c := colly.NewCollector(
					colly.MaxDepth(depth),
				)

				c.OnHTML("a[href], script[src]", func(e *colly.HTMLElement) {
					var link string

					// parse hrefs and js src URIs
					if href := e.Attr("href"); href != "" {
						link = e.Request.AbsoluteURL(href)
					} else if src := e.Attr("src"); src != "" {
						link = e.Request.AbsoluteURL(src)
					}

					if link != "" && !shouldExclude(link) {
						// Update context with the current URL chain
						chain := append(e.Request.Ctx.GetAny("urlChain").([]string), link)
						ctx := colly.NewContext()
						ctx.Put("urlChain", chain)

						// Manually create a new request with the updated context
						c.Request("GET", link, nil, ctx, nil)
					}
				})

				c.OnRequest(func(r *colly.Request) {

					r.Headers.Set("User-Agent", RandomString(userAgentList))

					if verbose {
						fmt.Println("Visiting", r.URL)
					}
				})

				c.OnResponse(func(r *colly.Response) {
					body := string(r.Body)

					// iterate the regexp pattern map
					for _, re := range patternMap {
						matches := re.FindAllString(body, -1)

						for _, match := range matches {
							// Print the match (S3 bucket name) and the URL chain
							if verbose {
								color.Green.Println("S3 Bucket Found:", match)
								color.Green.Println("At URL:", r.Request.URL)
								fmt.Println("URL Chain:")
								if chain, ok := r.Ctx.GetAny("urlChain").([]string); ok {
									for _, u := range chain {
										color.Yellow.Println(u)
									}
								}
								fmt.Println("------")
							} else {
								fmt.Println("S3 Bucket Found:", match)
								fmt.Println("At URL:", r.Request.URL)
								fmt.Println("URL Chain:")
								if chain, ok := r.Ctx.GetAny("urlChain").([]string); ok {
									for _, u := range chain {
										fmt.Println(u)
									}
								}
								fmt.Println("------")
							}

						}
					}
				})

				// Initialize the context with the starting URL
				ctx := colly.NewContext()
				ctx.Put("urlChain", []string{job})
				c.Request("GET", job, nil, ctx, nil)
			}
		}()
	}

	// check for input piped to stdin
	info, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if info.Mode()&os.ModeCharDevice != 0 || (info.Mode()&os.ModeNamedPipe == 0 && info.Size() <= 0) {
		print_usage()
	}

	// get user input
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		domain := scanner.Text()

		if !strings.HasPrefix(domain, "http") {
			domain = "https://" + domain
		}
		jobs <- domain
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	// wait for workers
	close(jobs)
	wg.Wait()
}

func RandomString(userAgentList []string) string {
	randomIndex := rand.Intn(len(userAgentList))
	return userAgentList[randomIndex]
}

func shouldExclude(link string) bool {
	for _, domain := range excludedDomains {
		if strings.Contains(link, domain) {
			return true
		}
	}
	return false
}

func print_usage() {
	log.Fatalln("Expected usage: echo <domain> | bucket-finder")
}
