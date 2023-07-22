package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

const baseUrl = `https://www.hacg.sbs/wp/%d.html`

const coroCnt = 128

var lock = &sync.Mutex{}
var idChan = make(chan int)
var cChan = make(chan chan *ParseResult, coroCnt*2)

func main() {
	var output string
	var startId int

	if len(os.Args) > 0 {
		output = os.Args[1]
	} else {
		output = "output.json"
	}

	results := make([]*ParseResult, 0)
	f, err := os.ReadFile(output)
	if err == nil {
		_ = json.Unmarshal(f, &results)
		startId = results[len(results)-1].Index + 1
	} else {
		startId = 1
	}
	fmt.Println("Start ID:", startId)

	endId := GetEndId(baseUrl)
	fmt.Println("End ID:", endId)

	go func() {
		for i := startId; i <= endId; i++ {
			idChan <- i
		}
		close(idChan)
	}()

	for i := startId; i < startId+coroCnt && i <= endId; i++ {
		c := make(chan *ParseResult, 1)
		lock.Lock()
		cChan <- c
		go Process(c)
	}

	for i := startId; i <= endId; i++ {
		r := <-<-cChan
		if r != nil {
			results = append(results, r)
			j, _ := json.MarshalIndent(results, "", "  ")
			os.WriteFile(output, j, 0644)
		}
	}
}
