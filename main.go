package main

import (
	"badstuff/requests"
	"badstuff/spider"
)

func main() {
	session := requests.NewSession(&requests.SessionOptions{
		Header: map[string][]string{
			"User-Agent": {"Mozilla/5.0"},
		},
	})
	spider := &spider.Spider[*HResult]{
		Workflow: &HWorkflow{
			OutputPath: "output/entries.json",
		},
		Session:    session,
		NParallels: 128,
	}
	spider.Run()
}
