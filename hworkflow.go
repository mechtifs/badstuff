package main

import (
	"badstuff/requests"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var re = map[string]*regexp.Regexp{
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

type HResult struct {
	Index      int       `json:"id"`
	Title      string    `json:"title"`
	Time       time.Time `json:"time"`
	Categories []string  `json:"categories"`
	Tags       []string  `json:"tags"`
	Magnets    []string  `json:"magnets"`
}

type HWorkflow struct {
	BaseUri         string
	Results         []*HResult
	OutputPath      string
	lastProcessTime int
}

func (w *HWorkflow) getEndId() int {
	r, err := requests.Get(fmt.Sprintf(w.BaseUri, 0)+`/`, nil)
	if err != nil {
		panic("Cannot fetch end ID")
	}
	articles := re["article"].FindAllSubmatch(r.Body, -1)
	var endId int
	for _, article := range articles {
		if !strings.Contains(string(article[1]), "sticky") {
			endId, err = strconv.Atoi(strings.Split(string(article[1]), " ")[0])
			if err != nil {
				panic("Cannot parse end ID")
			}
			break
		}
	}
	return endId
}

func (w *HWorkflow) dumpResults() {
	j, _ := json.MarshalIndent(w.Results, "", "  ")
	os.WriteFile(w.OutputPath, j, 0644)
}

func (w *HWorkflow) Init() {
	f, err := os.ReadFile(w.OutputPath)
	if err == nil {
		json.Unmarshal(f, &w.Results)
	}
}

func (w *HWorkflow) Generate(c chan string) {
	startId := 1
	if len(w.Results) > 0 {
		startId = w.Results[len(w.Results)-1].Index + 1
	}
	endId := w.getEndId()
	for i := startId; i <= endId; i++ {
		c <- fmt.Sprintf(w.BaseUri, i)
	}
	close(c)
}

func (w *HWorkflow) Parse(r *requests.Response) *HResult {
	contents := re["content"].FindSubmatch(r.Body)
	if contents == nil {
		return nil
	}

	id, _ := strconv.Atoi(re["index"].FindStringSubmatch(r.Url)[1])

	magnetStrs := []string{}
	magnetsHex := re["magnetHex"].FindAllSubmatch(contents[1], -1)
	for i := range len(magnetsHex) {
		builder := &strings.Builder{}
		builder.WriteString("magnet:?xt=urn:btih:")
		builder.WriteString(strings.ToLower(string(magnetsHex[i][1])))
		magnetStrs = append(magnetStrs, builder.String())
	}
	magnetsB32 := re["magnetB32"].FindAllSubmatch(contents[1], -1)
	for i := range len(magnetsB32) {
		decoded, err := base32.StdEncoding.DecodeString(string(magnetsB32[i][1]))
		if err != nil {
			continue
		}
		builder := &strings.Builder{}
		builder.WriteString("magnet:?xt=urn:btih:")
		builder.WriteString(hex.EncodeToString(decoded))
		magnetStrs = append(magnetStrs, builder.String())
	}

	var titleStr string
	titles := re["title"].FindSubmatch(r.Body)
	if titles != nil {
		titleStr = html.UnescapeString(string(titles[1]))
		if titleStr == "未找到页面" {
			return nil
		}
	}

	timeIso8601, _ := time.Parse(time.RFC3339, string(re["time"].FindSubmatch(r.Body)[1]))

	categoryStrs := []string{}
	categories := re["category"].FindAllSubmatch(r.Body, -1)
	for i := range len(categories) {
		categoryStrs = append(categoryStrs, html.UnescapeString(string(categories[i][1])))
	}

	tagStrs := []string{}
	tags := re["tag"].FindAllSubmatch(r.Body, -1)
	for i := range len(tags) {
		tagStrs = append(tagStrs, html.UnescapeString(string(tags[i][1])))
	}

	return &HResult{
		Index:      id,
		Title:      titleStr,
		Time:       timeIso8601,
		Categories: categoryStrs,
		Tags:       tagStrs,
		Magnets:    magnetStrs,
	}
}

func (w *HWorkflow) Process(i *HResult) {
	if i == nil {
		return
	}
	w.Results = append(w.Results, i)
	currentTime := int(time.Now().Unix())
	if currentTime == w.lastProcessTime {
		return
	}
	w.lastProcessTime = currentTime
	w.dumpResults()
}

func (w *HWorkflow) Finalize() {
	w.dumpResults()
}
