package main

import (
	"encoding/base32"
	"encoding/hex"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"

	"badstuff/requests"
)

type ParseResult struct {
	Index      int       `json:"id"`
	Title      string    `json:"title"`
	Time       time.Time `json:"time"`
	Categories []string  `json:"categories"`
	Tags       []string  `json:"tags"`
	Magnets    []string  `json:"magnets"`
}

var exp = map[string]*regexp.Regexp{
	"index":     regexp.MustCompile(`\/wp\/(\d+).html`),
	"article":   regexp.MustCompile(`class="post-(.+?)"`),
	"title":     regexp.MustCompile(`<title>(.+) \| .+?<\/title>`),
	"content":   regexp.MustCompile(`entry-content([\S\s]+?).entry-content`),
	"magnetHex": regexp.MustCompile(`[^\/=+0-9A-Fa-f]([0-9A-Fa-f]{32}|[0-9A-Fa-f]{40})[^\/=+0-9A-Fa-f]`),
	"magnetB32": regexp.MustCompile(`[^\/=+0-9A-Fa-f]([2-7A-Z]{32})[^\/=+0-9A-Fa-f]`),
	"category":  regexp.MustCompile(`rel="category tag">(.+?)<\/a>`),
	"tag":       regexp.MustCompile(`rel="tag">(.+?)<\/a>`),
	"time":      regexp.MustCompile(`datetime="(.+?)"`),
}

func parse(r *requests.Response) *ParseResult {
	contents := exp["content"].FindSubmatch(r.Body)
	if contents == nil {
		return nil
	}

	id, _ := strconv.Atoi(exp["index"].FindStringSubmatch(r.Url)[1])

	magnetStrs := []string{}
	magnetsHex := exp["magnetHex"].FindAllSubmatch(contents[1], -1)
	for i := range len(magnetsHex) {
		magnetStrs = append(magnetStrs, "magnet:?xt=urn:btih:"+strings.ToLower(string(magnetsHex[i][1])))
	}
	magnetsB32 := exp["magnetB32"].FindAllSubmatch(contents[1], -1)
	for i := range len(magnetsB32) {
		decoded, err := base32.StdEncoding.DecodeString(string(magnetsB32[i][1]))
		if err != nil {
			continue
		}
		magnetStrs = append(magnetStrs, "magnet:?xt=urn:btih:"+hex.EncodeToString(decoded))
	}

	var titleStr string
	titles := exp["title"].FindSubmatch(r.Body)
	if titles != nil {
		titleStr = html.UnescapeString(string(titles[1]))
		if titleStr == "未找到页面" {
			return nil
		}
	}

	timeIso8601, _ := time.Parse(time.RFC3339, string(exp["time"].FindSubmatch(r.Body)[1]))

	categoryStrs := []string{}
	categories := exp["category"].FindAllSubmatch(r.Body, -1)
	for i := range len(categories) {
		categoryStrs = append(categoryStrs, html.UnescapeString(string(categories[i][1])))
	}

	tagStrs := []string{}
	tags := exp["tag"].FindAllSubmatch(r.Body, -1)
	for i := range len(tags) {
		tagStrs = append(tagStrs, html.UnescapeString(string(tags[i][1])))
	}

	return &ParseResult{
		Index:      id,
		Title:      titleStr,
		Time:       timeIso8601,
		Categories: categoryStrs,
		Tags:       tagStrs,
		Magnets:    magnetStrs,
	}
}
