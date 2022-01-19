//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Visit http://localhost:8080")
	http.ListenAndServe(`localhost:8080`, http.FileServer(http.Dir(`.`)))
}
