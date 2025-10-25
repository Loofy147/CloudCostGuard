package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/repos/test-org/test-repo/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
			return
		}
		fmt.Println("Mock GitHub server received a comment.")
		w.WriteHeader(http.StatusCreated)
	})

	fmt.Println("Mock GitHub server listening on :8081...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Failed to start mock server: %v", err)
	}
}
