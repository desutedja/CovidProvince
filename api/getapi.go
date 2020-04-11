package getapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
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
	Longitude float64 `json:"x"`
	Latitude  float64 `json:"y"`
}

type CovidResponse struct {
	CovidProvinsi []DataResult `json:"covidprovinsi"`
}

type DataResult struct {
	FID            int     `json:"fid"`
	KodeProvinsi   int     `json:"kodeprofinsi"`
	Provinsi       string  `json:"profinsi"`
	KasusPositif   int     `json:"kasuspositif"`
	KasusSembuh    int     `json:"kasussembuh"`
	KasusMeninggal int     `json:"kasusmeninggal"`
	Longitude      float64 `json:"longitude"`
	Latitude       float64 `json:"latitude"`
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "docker"
	dbname   = "covidprovinsi"
)

var db *sql.DB

func conDB() {
	// Create connection string
	connString := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Create connection pool
	var err error
	db, err = sql.Open("postgres", connString)
	if err != nil {
		log.Fatal("Error creating connection pool: " + err.Error())
	}
	// defer db.Close()

	// err = db.Ping()
	// if err != nil {
	// 	panic(err)
	// }

	log.Printf("Connected!\n")
}

//CreateTable is a function to create first initial database table
func CreateTable(w http.ResponseWriter, r *http.Request) {
	conDB()

	query := `CREATE TABLE IF NOT EXISTS Covid (
		FID integer NOT NULL,
		KodeProvinsi integer NOT NULL,
		Provinsi varchar(100) NOT NULL,
		KasusPositif integer NOT NULL,
		KasusSembuh integer NOT NULL,
		KasusMeninggal integer NOT NULL,
		Latitude double precision,
		Longitude double precision,
		DateTimeCreated timestamp
	  )`

	_, err := db.Exec(query)
	if err != nil {
		db.Close()
		log.Fatal("Create table failed:", err.Error())
	}

	log.Printf("Create table success!\n")
	db.Close()
}

//function for check different max hour - current time, from table covid
func checkhour() int {
	conDB()

	//if null set hour to 10
	query := `select COALESCE(DATE_PART('day',now() at time zone 'asia/jakarta'-max(DateTimeCreated)) * 24 
			+ DATE_PART('hour', now() at time zone 'asia/jakarta'-max(DateTimeCreated)),10) as hour from covid`

	var res int
	err := db.QueryRow(query).Scan(&res)
	if err != nil {
		db.Close()
		log.Fatal("query Check Hour failed:", err.Error())
	}
	db.Close()

	return res
}

func (box *CovidResponse) AddItem(item DataResult) {
	box.CovidProvinsi = append(box.CovidProvinsi, item)
}

//GetAPI is a function to call api from source(BNPB)
func GetAPI(w http.ResponseWriter, r *http.Request) {
	var hour int
	var query string
	hour = checkhour()

	conDB()
	var result CovidResponse

	//if last insert data more than two hour then insert new
	if hour > 2 {
		log.Printf("Inserting Data!\n")
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

		trans, err := db.Begin()
		if err != nil {
			log.Fatal("Begin transaction error : " + err.Error())
		}

		stmt, err := trans.Prepare(`INSERT INTO Covid (FID,KodeProvinsi,Provinsi,KasusPositif,KasusSembuh,KasusMeninggal,Latitude,Longitude,DateTimeCreated)
								VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9);`)
		if err != nil {
			trans.Rollback()
			log.Fatal("Preparation transaction error : " + err.Error())
		}

		var dr DataResult
		for i := 0; i < len(responseObject.Features); i++ {
			attr := responseObject.Features[i].Attributes
			geo := responseObject.Features[i].Geometries

			_, err := stmt.Exec(attr.FID, attr.KodeProvinsi, attr.Provinsi, attr.KasusPositif, attr.KasusSembuh, attr.KasusMeninggal, geo.Latitude, geo.Longitude, time.Now())
			if err != nil {
				trans.Rollback()
				log.Fatal("Execute transaction error : " + err.Error()) // return an error too, we may want to wrap them
			}

			dr.FID = attr.FID
			dr.KodeProvinsi = attr.KodeProvinsi
			dr.Provinsi = attr.Provinsi
			dr.KasusPositif = attr.KasusPositif
			dr.KasusSembuh = attr.KasusSembuh
			dr.KasusMeninggal = attr.KasusMeninggal
			dr.Latitude = geo.Latitude
			dr.Longitude = geo.Longitude

			result.AddItem(dr)
		}

		defer stmt.Close()
		trans.Commit()
		db.Close()

		json.NewEncoder(w).Encode(result)
		return
	}

	log.Printf("Read Data!\n")
	query = `select FID,KodeProvinsi,Provinsi,KasusPositif,KasusSembuh,KasusMeninggal,Latitude,Longitude 
				from covid where TO_CHAR(datetimecreated, 'YYYYMMDDHH24') = TO_CHAR((select MAX(datetimecreated) from covid),'YYYYMMDDHH24')`

	selectQuery, err := db.Query(query)
	for selectQuery.Next() {
		var r DataResult
		err = selectQuery.Scan(&r.FID, &r.KodeProvinsi, &r.Provinsi, &r.KasusPositif, &r.KasusSembuh, &r.KasusMeninggal, &r.Latitude, &r.Longitude)
		if err != nil {
			panic(err.Error())
		}

		result.AddItem(r)
	}
	db.Close()

	json.NewEncoder(w).Encode(result)
}
