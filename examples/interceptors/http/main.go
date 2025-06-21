package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type CallToolRequest struct {
	Params CallToolParams `json:"params"`
}

type CallToolParams struct {
	Name      string `json:"name"`
	Arguments any    `json:"arguments,omitempty"`
}

type CallToolResult struct {
	Content []TextContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type TextContent struct {
	Text string `json:"text"`
}

func main() {
	http.HandleFunc("/before", func(w http.ResponseWriter, r *http.Request) {
		var toolCall CallToolRequest
		if err := json.NewDecoder(r.Body).Decode(&toolCall); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		fmt.Fprintf(os.Stderr, "Calling tool [%s] with arguments: %v", toolCall.Params.Name, toolCall.Params.Arguments)

		// Here, instead of returning an empty 200 response, we could bypass the tool call
		// totally and return our own response.
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/after", func(w http.ResponseWriter, r *http.Request) {
		var result CallToolResult
		if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		fmt.Fprintf(os.Stderr, "Tool gave a response of: %d characters", len(result.Content[0].Text))
		w.WriteHeader(http.StatusOK)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalln("Server failed:", err)
	}
}
