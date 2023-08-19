## bucket-finder
Reads in a list of domains from stdin and crawls them to find S3 buckets in the HTML source and from within referenced JavaScript files. My primary motivation for this tool is to list S3 buckets related to a target with a view to pipe the output into other tooling to look for S3 misconfigurations. Your use-case may vary.

## Recommended Usage

`$ cat domains | bucket-finder`

or, you can use as part of your bug-bounty recon workflow, e.g.

`$ assetfinder -subs-only example.com | bucket-finder`


## Options

```
-c int
    set the concurrency level (default 50)

-v  get more info on attempts
```

## Install

You need to have [Go installed](https://golang.org/doc/install) and configured (i.e. with $GOPATH/bin in your $PATH):

`go install github.com/cybercdh/bucket-finder@latest`

## Thanks

`bucket-finder` uses the [colly](https://github.com/gocolly/colly) framework for crawling, which makes this type of code super-simple to implement.