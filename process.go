package main

import (
	"encoding/json"
	"os"
	"time"
)

var lastProcessTime int

func process(i interface{}) {
	if i == nil {
		return
	}
	results = append(results, i.(*ParseResult))

	currentTime := int(time.Now().Unix())
	if currentTime == lastProcessTime {
		return
	}
	lastProcessTime = currentTime
	j, _ := json.MarshalIndent(results, "", "  ")
	os.WriteFile(output, j, 0644)
}
