package main

import (
	"encoding/json"
	"os"
)

func finalize() {
	j, _ := json.MarshalIndent(results, "", "  ")
	os.WriteFile(output, j, 0644)
}
