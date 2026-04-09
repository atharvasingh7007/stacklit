package internal

import "net/http"

func Handle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func processRequest(r *http.Request) error {
	return nil
}
