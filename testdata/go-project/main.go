package main

import (
	"fmt"
	"net/http"
	"github.com/example/pkg/handler"
)

func main() {
	fmt.Println("hello")
	http.ListenAndServe(":8080", nil)
}
