package main

import (
	"badstuff/requests"
	"badstuff/spider"
)

const output = "output/entries.json"

func main() {
	spider := &spider.Spider{
		Generate:   generate,
		Parse:      parse,
		Process:    process,
		NParallels: 128,
		SessionOptions: &requests.SessionOptions{
			NoRedirect: true,
			Header: map[string][]string{
				"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; rv:109.0) Gecko/20100101 Firefox/115.0"},
			},
		},
	}
	spider.Run()
}
