package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Person struct {
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
}

const (
	host     = "127.0.0.1"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
)

func OpenConnection() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}

func GETHandler(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()

	rows, err := db.Query("SELECT * FROM person")
	if err != nil {
		log.Fatal(err)
	}

	var people []Person

	for rows.Next() {
		var person Person
		rows.Scan(&person.Name, &person.Nickname)
		people = append(people, person)
	}

	peopleBytes, _ := json.MarshalIndent(people, "", "\t")

	w.Header().Set("Content-Type", "application/json")
	w.Write(peopleBytes)

	defer rows.Close()
	defer db.Close()
}

func POSTHandler(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()

	var p Person
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sqlStatement := `INSERT INTO person (name, nickname) VALUES ($1, $2)`
	_, err = db.Exec(sqlStatement, p.Name, p.Nickname)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/text")
	w.Write([]byte("Success!"))

	defer db.Close()
}

func DELETEHandler(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	err := sql.ErrNoRows
	vars := mux.Vars(r)
	nickname := vars["nickname"]
	sqlStatement := `DELETE FROM person WHERE nickname=$1`
	res, err := db.Exec(sqlStatement, nickname)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}
	if count == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/text")
		w.Write([]byte("No data found"))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/text")
		w.Write([]byte("Success!"))
	}

	defer db.Close()
}

func PUTHandler(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	vars := mux.Vars(r)
	nickname := vars["nickname"]
	var p Person
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sqlStatement := `UPDATE  person  SET name=$1,nickname=$2 where nickname=$3`
	res, err := db.Exec(sqlStatement, p.Name, p.Nickname,nickname)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}
	if count == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/text")
		w.Write([]byte("No data found"))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/text")
		w.Write([]byte("Success!"))
	}

	defer db.Close()
}

func GETByNicknameHandler(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	vars := mux.Vars(r)
	nickname := vars["nickname"]
	sqlStatement := `SELECT * FROM  person WHERE nickname=$1`
	var person Person
	err := db.QueryRow(sqlStatement,nickname).Scan(&person.Name, &person.Nickname)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		personByte, _ := json.MarshalIndent(person, "", "\t")
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/text")
		w.Write([]byte(personByte))
		defer db.Close()

	}
	
	}




func main() {
	router := mux.NewRouter().StrictSlash(true)
	port := ":8080"
	router.HandleFunc("/", POSTHandler).Methods("POST")
	router.HandleFunc("/", GETHandler)
	router.HandleFunc("/{nickname}", DELETEHandler).Methods("DELETE")
	router.HandleFunc("/{nickname}", PUTHandler).Methods("PUT")
    router.HandleFunc("/{nickname}",GETByNicknameHandler).Methods("GET")
	log.Println("listening on port", port)
	log.Fatal(http.ListenAndServe(port, router))
}
