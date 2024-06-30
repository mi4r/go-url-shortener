package main

import (
	"fmt"
	"net/http"
)

func GetHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", "https://practicum.yandex.ru/")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func PostHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}
	body := req.URL.Path + "id"
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(body))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/id`, GetHandler)
	mux.HandleFunc(`/`, PostHandler)
	if err := http.ListenAndServe(`localhost:8080`, mux); err != nil {
		fmt.Println(err.Error())
	}
}
