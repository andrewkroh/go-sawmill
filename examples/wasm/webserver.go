//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Visit http://localhost:8081")
	if err := http.ListenAndServe(`localhost:8081`, http.FileServer(http.Dir(`.`))); err != nil {
		log.Fatal("ERROR:", err)
	}
}
