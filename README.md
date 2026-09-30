## bucket-finder
Reads in a list of domains from stdin and crawls them to find S3 buckets in the HTML source and from within referenced JavaScript files. My primary motivation for this tool is to list S3 buckets related to a target with a view to pipe the output into other tooling to look for S3 misconfigurations. Your use-case may vary.

## Recommended Usage

`$ cat domains | bucket-finder`

or, you can use as part of your bug-bounty recon workflow, e.g.

`$ assetfinder -subs-only example.com | bucket-finder`


or a single target as an argument:

`$ bucket-finder example.com`

## Output

One bucket name per line on stdout, de-duplicated across the whole run, so it pipes straight into other tooling. With `-v`, each hit is also explained on stderr with the URL it was found at and the chain of pages that led there.

Bucket references are recognised in every shape AWS uses: virtual host (`bucket.s3.amazonaws.com`, `bucket.s3.eu-west-1.amazonaws.com`, `bucket.s3-website-us-east-1.amazonaws.com`, dualstack), path style (`s3.amazonaws.com/bucket`, `s3-eu-west-1.amazonaws.com/bucket`) and `s3://bucket`, including inside JavaScript strings with escaped slashes.

## Options

```
-c int
    set the concurrency level (default 50)

-d int
    set the crawling depth (default 5). 1 is the page itself, 2 adds the pages it links to, and so on.

-external
    follow page links to other domains as well. By default links stay on the target's own registrable domain. Script files are always fetched wherever they are hosted, since CDN bundles are where bucket references tend to hide.

-v  get more info on attempts (printed to stderr)
```

## Install

You need to have [Go installed](https://golang.org/doc/install) and configured (i.e. with $GOPATH/bin in your $PATH):

`go install github.com/cybercdh/bucket-finder@latest`

## Thanks

`bucket-finder` uses the [colly](https://github.com/gocolly/colly) framework for crawling, which makes this type of code super-simple to implement.