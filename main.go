package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type ResponseData struct {
	Features []Feature `json:"features"`
}

type Feature struct {
	Attributes Attribute `json:"attributes"`
	Geometries Geometry  `json:"geometry"`
}

type Attribute struct {
	FID            int    `json:"FID"`
	KodeProvinsi   int    `json:"Kode_Provi"`
	Provinsi       string `json:"Provinsi"`
	KasusPositif   int    `json:"Kasus_Posi"`
	KasusSembuh    int    `json:"Kasus_Semb"`
	KasusMeninggal int    `json:"Kasus_Meni"`
}

type Geometry struct {
	Longitude string `json:"x"`
	Latitude  string `json:"y"`
}

func main() {
	response, err := http.Get("https://services5.arcgis.com/VS6HdKS0VfIhv8Ct/arcgis/rest/services/COVID19_Indonesia_per_Provinsi/FeatureServer/0/query?where=1%3D1&outFields=*&outSR=4326&f=json")

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var responseObject ResponseData
	json.Unmarshal(data, &responseObject)

	for i := 0; i < len(responseObject.Features); i++ {
		fmt.Println(responseObject.Features[i].Attributes)
	}
}
