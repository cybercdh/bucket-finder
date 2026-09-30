package main

import (
	"regexp"
	"strings"
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

// Bucket references come in three shapes:
//
//	virtual-host  bucket.s3.amazonaws.com, bucket.s3.eu-west-1.amazonaws.com,
//	              bucket.s3-website-us-east-1.amazonaws.com, bucket.s3.dualstack.<region>.amazonaws.com
//	path-style    s3.amazonaws.com/bucket, s3-eu-west-1.amazonaws.com/bucket, s3.<region>.amazonaws.com/bucket
//	scheme        s3://bucket/key
//
// The previous patterns for the first two were anchored with ^ and $ and run
// against whole page bodies, so they could never match; only a loose fourth
// pattern did any work, and it missed path-style and s3:// entirely. Rather
// than one regex per shape, a loose candidate match is followed by a small
// parser that finds the "s3*" label and takes what is in front of it (virtual
// host) or the first path segment after it (path style).
var (
	reAmazonHost = regexp.MustCompile(`(?i)(?:[a-z0-9][a-z0-9.\-]*\.)?s3[a-z0-9\-]*(?:\.[a-z0-9\-]+)*\.amazonaws\.com(?:\.cn)?(?:/[a-z0-9][a-z0-9.\-_]*)?`)
	reS3Scheme   = regexp.MustCompile(`(?i)\bs3://([a-z0-9][a-z0-9.\-]{1,61}[a-z0-9])`)
)

// notBucketLabels are s3* service labels whose leading labels are not buckets.
var notBucketLabels = map[string]bool{
	"s3-control": true, "s3-outposts": true, "s3-accesspoint": true, "s3-object-lambda": true,
}

// extractBuckets returns the unique, lower-cased bucket names referenced in
// body, in first-seen order. JavaScript-escaped slashes ("\/") are unescaped
// first so references inside JSON and JS string literals are found too.
func extractBuckets(body []byte) []string {
	text := strings.ReplaceAll(string(body), `\/`, `/`)
	var out []string
	seen := make(map[string]bool)
	add := func(name string) {
		name = strings.ToLower(strings.Trim(name, "."))
		if len(name) < 3 || len(name) > 63 || seen[name] {
			return
		}
		seen[name] = true
		out = append(out, name)
	}

	for _, m := range reAmazonHost.FindAllString(text, -1) {
		if name, ok := bucketFromAmazonRef(m); ok {
			add(name)
		}
	}
	for _, m := range reS3Scheme.FindAllStringSubmatch(text, -1) {
		add(m[1])
	}
	return out
}

// bucketFromAmazonRef parses one host[/segment] match into a bucket name.
func bucketFromAmazonRef(ref string) (string, bool) {
	ref = strings.ToLower(ref)
	host, path, _ := strings.Cut(ref, "/")
	labels := strings.Split(host, ".")
	for i, l := range labels {
		if !strings.HasPrefix(l, "s3") {
			continue
		}
		if notBucketLabels[l] {
			return "", false
		}
		if i > 0 {
			// virtual-host style: everything before the s3 label is the bucket
			return strings.Join(labels[:i], "."), true
		}
		// path style: the bucket is the first path segment
		if path != "" {
			return path, true
		}
		return "", false
	}
	return "", false
}
