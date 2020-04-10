package main

import (
	"github.com/desutedja/covid/api"
)

func main() {
	http.HandleFunc("/", GetApi())
}
