package main

import (
	"fmt"
	"net/http"

	getapi "github.com/desutedja/CovidProvince/api"
	"goji.io"
	"goji.io/pat"
)

func main() {

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/"), getapi.GetAPI)
	mux.HandleFunc(pat.Get("/CreateTable"), getapi.CreateTable)

	fmt.Println("listening on port 8181")
	http.ListenAndServe("localhost:8181", mux)
}
