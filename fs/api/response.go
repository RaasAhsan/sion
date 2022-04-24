package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Success interface{} `json:",omitempty"`
	Error   *Error      `json:",omitempty"`
}

type Error struct {
	Message string
	Code    ErrorCode
}

type ErrorCode string

const FileNotFound ErrorCode = "FileNotFound"
const ChunkNotFound ErrorCode = "ChunkNotFound"
const Unknown ErrorCode = "Unknown"

func HttpOk(w http.ResponseWriter, body interface{}) {
	HttpSuccess(w, body, http.StatusOK)
}

func HttpSuccess(w http.ResponseWriter, body interface{}, statusCode int) {
	resp := Response{Success: body}

	bytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		http.Error(w, "Failed to return response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(bytes)

	// TODO: debug
	fmt.Printf("%s\n", string(bytes))
}

func HttpError(w http.ResponseWriter, message string, code ErrorCode, statusCode int) {
	resp := Response{Error: &Error{Message: message, Code: code}}

	bytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		http.Error(w, "Failed to return response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(bytes)

	// TODO: debug
	fmt.Printf("%s\n", string(bytes))
}
