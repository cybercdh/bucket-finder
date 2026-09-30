package main

import (
	"reflect"
	"testing"
)

func TestExtractBuckets(t *testing.T) {
	body := `
<link href="https://Static-Assets.s3.amazonaws.com/site.css">
<script src="//cdn-bundle.s3.eu-west-1.amazonaws.com/app.js"></script>
<img src="http://photos.example.s3-website-us-east-1.amazonaws.com/a.png">
<a href="https://s3.amazonaws.com/legacy-uploads/file.pdf">
<a href="https://s3-us-west-2.amazonaws.com/regional-path/x">
<a href="https://s3.dualstack.ap-southeast-2.amazonaws.com/dual-path/x">
var cfg = {"url":"https:\/\/js-escaped.s3.amazonaws.com\/data.json"};
const uri = "s3://scheme-bucket/key/object.txt";
<a href="https://s3-control.us-east-1.amazonaws.com/v20180820/">
<a href="https://s3.amazonaws.com/">
duplicate: https://static-assets.s3.amazonaws.com/other.css
`
	want := []string{
		"static-assets",
		"cdn-bundle",
		"photos.example",
		"legacy-uploads",
		"regional-path",
		"dual-path",
		"js-escaped",
		"scheme-bucket",
	}
	got := extractBuckets([]byte(body))
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %v\nwant %v", got, want)
	}
}

func TestNormalizeURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"example.com", "https://example.com", true},
		{"  http://example.com/x  ", "http://example.com/x", true},
		{"httpbin.org", "https://httpbin.org", true},
		{"", "", false},
		{"# c", "", false},
		{"http://", "", false},
	}
	for _, c := range cases {
		got, ok := normalizeURL(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("normalizeURL(%q) = (%q, %v), want (%q, %v)", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestRegistrable(t *testing.T) {
	cases := map[string]string{
		"https://www.Example.co.uk/x": "example.co.uk",
		"https://cdn.example.com/":    "example.com",
		"http://127.0.0.1:8080/":      "127.0.0.1",
		"http://localhost/":           "localhost",
	}
	for in, want := range cases {
		if got := registrable(in); got != want {
			t.Errorf("registrable(%q) = %q, want %q", in, got, want)
		}
	}
}
