package main

import (
	"bad_stuff/requests"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
)

var exp = map[string]*regexp.Regexp{
	"article":  regexp.MustCompile(`class="post-(.+?)"`),
	"title":    regexp.MustCompile(`<title>(.+) \| .+?<\/title>`),
	"content":  regexp.MustCompile(`entry-content([\S\s]+?).entry-content`),
	"magnet":   regexp.MustCompile(`[^/=+0-9a-fA-F]([0-9a-fA-F]{32}|[0-9a-fA-F]{40})[^/=+0-9a-fA-F]`),
	"category": regexp.MustCompile(`rel="category tag">(.+?)<\/a>`),
	"tag":      regexp.MustCompile(`rel="tag">(.+?)<\/a>`),
	"time":     regexp.MustCompile(`datetime="(.+?)"`),
}

type ParseResult struct {
	Index      int      `json:"id"`
	Title      string   `json:"title"`
	Time       string   `json:"time"`
	Categories []string `json:"categories"`
	Tags       []string `json:"tags"`
	Magnets    []string `json:"magnets"`
}

func GetEndId(baseUrl string) int {
	r, err := requests.Get(fmt.Sprintf(baseUrl, 0) + `/`)
	if err != nil {
		panic("Cannot fetch end ID")
	}
	articles := exp["article"].FindAllSubmatch(r.Content, -1)
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

func newTask() {
	cN := make(chan *ParseResult, 1)
	lock.Lock()
	cChan <- cN
	go Process(cN)
}

func Process(c chan *ParseResult) {
	id, ok := <-idChan
	lock.Unlock()
	if !ok {
		c <- nil
		return
	}

	var r *requests.Response
	var err error

	fmt.Println("Coro", id, "Started")

	for i := 0; ; i++ {
		if i >= 10 {
			fmt.Println("WARNING:", id, "has been retried for", i, "times")
		}
		r, err = requests.Get(fmt.Sprintf(baseUrl, id))
		if err != nil {
			fmt.Println(err)
			continue
		}
		if r.StatusCode < 500 {
			break
		}
	}

	if r.StatusCode != 200 {
		// fmt.Println(id, r.StatusCode)
		c <- nil
		fmt.Println("Coro", id, "finished", len(cChan))
		newTask()
		return
	}

	contents := exp["content"].FindSubmatch(r.Content)
	if contents == nil {
		// fmt.Println(id, "invalid content")
		c <- nil
		fmt.Println("Coro", id, "finished", len(cChan))
		newTask()
		return
	}

	magnets := exp["magnet"].FindAllSubmatch(contents[1], -1)
	var magnetStrs []string
	if magnets == nil {
		magnetStrs = []string{}
	} else {
		for i := 0; i < len(magnets); i++ {
			magnetStrs = append(magnetStrs, html.UnescapeString(strings.ToLower(string(magnets[i][1]))))
		}
	}

	var titleStr string
	titles := exp["title"].FindSubmatch(r.Content)
	if titles != nil {
		titleStr = html.UnescapeString(string(titles[1]))
		if titleStr == "未找到页面" {
			// fmt.Println(id, "not found")
			c <- nil
			fmt.Println("Coro", id, "finished", len(cChan))
			newTask()
			return
		}
	}

	timeStr := string(exp["time"].FindSubmatch(r.Content)[1])

	categories := exp["category"].FindAllSubmatch(r.Content, -1)
	var categoryStrs []string
	if categories == nil {
		categoryStrs = []string{}
	} else {
		for i := 0; i < len(categories); i++ {
			categoryStrs = append(categoryStrs, html.UnescapeString(strings.ToLower(string(categories[i][1]))))
		}
	}

	tags := exp["tag"].FindAllSubmatch(r.Content, -1)
	var tagStrs []string
	if tags == nil {
		tagStrs = []string{}
	} else if len(tagStrs) == 1 {
		tagStrs = strings.Split(tagStrs[0], "，")
	} else {
		for i := 0; i < len(tags); i++ {
			tagStrs = append(tagStrs, strings.TrimRight(html.UnescapeString(string(tags[i][1])), "| "))
		}
	}

	result := &ParseResult{
		Index:      id,
		Title:      titleStr,
		Time:       timeStr,
		Categories: categoryStrs,
		Tags:       tagStrs,
		Magnets:    magnetStrs,
	}
	c <- result
	fmt.Println("Coro", id, "finished", len(cChan))
	newTask()
}
