package main

import (
	"badstuff/requests"
	"badstuff/spider"
)

const output = "output/entries.json"

func main() {
	session := requests.NewSession(&requests.SessionOptions{
		Header: map[string][]string{
			"User-Agent": {"Mozilla/5.0"},
		},
	})
	spider := &spider.Spider[*ParseResult]{
		Generate:   generate,
		Parse:      parse,
		Process:    process,
		Finalize:   finalize,
		Session:    session,
		NParallels: 128,
	}
	spider.Run()
}
