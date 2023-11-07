package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"badstuff/requests"
)

const baseUrl = `https://www.hacg.lv/wp/%d.html`

func getEndId() int {
	r, err := requests.Get(fmt.Sprintf(baseUrl, 0)+`/`, nil)
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

func generate(c chan string) {
	startId := 1
	f, err := os.ReadFile(output)
	if err == nil {
		_ = json.Unmarshal(f, &results)
		startId = results[len(results)-1].Index + 1
	}
	endId := getEndId()
	for i := startId; i < endId; i++ {
		c <- fmt.Sprintf(baseUrl, i)
	}
	close(c)
}
