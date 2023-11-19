package main

import (
	"regexp"
)

var userAgentList = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.2 Safari/605.1.15",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.2 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (iPad; CPU OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.2 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edge/91.0.864.59",
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0",
	"Mozilla/5.0 (Android 11; Mobile; rv:89.0) Gecko/89.0 Firefox/89.0",
	"Mozilla/5.0 (Linux; Android 11; SM-G986U1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Mobile Safari/537.36",
}

var excludedDomains = []string{
	"facebook.com",
	"whatsapp.com",
	"linkedin.com",
	"google.com",
	"youtube.com",
	"apple.com",
	"twimg.com",
	"mailto:",
	"tel:",
	"javascript:void(0)",
	"twitter.com",
	"googleapis.com",
	"jquery.com",
	"instagram.com",
	"github.com",
}

// Matches bucketname.s3...amazonaws.com patterns
var reBucketFront = regexp.MustCompile(`^([\w\-]+)\.s3(?:-[\w-]+)?(?:\.dualstack)?(?:\.[\w-]+)?\.amazonaws\.com(?:\.cn)?$`)

// Matches s3...amazonaws.com/bucketname patterns
var reS3Front = regexp.MustCompile(`^s3(?:-[\w-]+)?(?:\.dualstack)?(?:\.[\w-]+)?\.amazonaws\.com(?:\.cn)?/([\w\-]+)$`)

// Matches s3://bucketname/file pattern
var reS3Scheme = regexp.MustCompile(`^s3://([\w\-]+)/`)

var reS3Original = regexp.MustCompile(`[\w\-\.]*\.s3\.?(?:[\w\-\.]+)?\.amazonaws\.com`)

var patternMap = map[string]*regexp.Regexp{
	"re1": reBucketFront,
	"re2": reS3Front,
	"re3": reS3Scheme,
	"re4": reS3Original,
}
