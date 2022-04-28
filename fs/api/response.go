package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Response[T any] struct {
	Success *T     `json:",omitempty"`
	Error   *Error `json:",omitempty"`
}

type Error struct {
	Message string
	Code    ErrorCode
}

type ErrorCode string

const FileNotFound ErrorCode = "FileNotFound"
const ChunkNotFound ErrorCode = "ChunkNotFound"
const Unknown ErrorCode = "Unknown"

func HttpOk[T any](w http.ResponseWriter, body T) {
	HttpSuccess(w, body, http.StatusOK)
}

func HttpSuccess[T any](w http.ResponseWriter, body T, statusCode int) {
	resp := Response[T]{Success: &body}

	bytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		http.Error(w, "Failed to return response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(bytes)

	// TODO: debug
	// fmt.Printf("%s\n", string(bytes))
}

func HttpError(w http.ResponseWriter, message string, code ErrorCode, statusCode int) {
	resp := Response[any]{Error: &Error{Message: message, Code: code}}

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

func ParseResponse[T any](in io.ReadCloser) (T, error) {
	var success T
	bytes, err := io.ReadAll(in)
	if err != nil {
		return success, errors.New("failed to read body")
	}
	var body Response[T]
	err = json.Unmarshal(bytes, &body)
	if err != nil {
		return success, errors.New("failed to unmarshal body")
	}
	if body.Error != nil {
		return success, errors.New(body.Error.Message)
	}
	success = *body.Success
	return success, nil
}
