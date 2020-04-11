package main

import (
	"net/http"

	getapi "github.com/desutedja/CovidProvince/api"
	"goji.io"
	"goji.io/pat"
)

func main() {

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/"), getapi.GetAPI)
	mux.HandleFunc(pat.Get("/CreateTable"), getapi.CreateTable)

	http.ListenAndServe("localhost:8181", mux)
}
