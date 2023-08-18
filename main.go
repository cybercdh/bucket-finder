package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	// "regexp"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/gookit/color"
)

var jobs = make(chan string, 100)
var concurrency int
var depth int
var verbose bool

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

				// crawl all js files
				c.OnHTML("script[src]", func(e *colly.HTMLElement) {
					e.Request.Visit(e.Attr("src"))
				})

				// crawl all hrefs
				c.OnHTML("a[href]", func(e *colly.HTMLElement) {
					link := e.Attr("href")
					if !shouldExclude(link) {
						e.Request.Visit(link)
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
							if verbose {
								color.Green.Println(match, r.Request.URL)
							} else {
								fmt.Println(match, r.Request.URL)
							}
						}
					}
				})

				c.Visit(job)
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
