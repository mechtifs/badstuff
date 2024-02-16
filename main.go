package main

import (
	"badstuff/requests"
	"badstuff/spider"
	"encoding/json"
	"os"
)

const output = "output/entries.json"

var results = make([]*ParseResult, 0)

func main() {
	f, err := os.ReadFile(output)
	if err == nil {
		json.Unmarshal(f, &results)
	}

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
