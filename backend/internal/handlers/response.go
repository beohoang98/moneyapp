package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const maxBodySize = 1 << 20 // 1 MB

type ListResponse[T any] struct {
	Data    []T `json:"data"`
	Total   int `json:"total"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type SingleResponse[T any] struct {
	Data T `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	if dec.More() {
		return fmt.Errorf("body must contain a single JSON object")
	}
	defer func() { io.Copy(io.Discard, r.Body) }()
	return nil
}

func respondList[T any](w http.ResponseWriter, status int, items []T, total, page, perPage int) {
	respondJSON(w, status, ListResponse[T]{
		Data:    items,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	})
}

func respondSingle[T any](w http.ResponseWriter, status int, item T) {
	respondJSON(w, status, SingleResponse[T]{Data: item})
}
