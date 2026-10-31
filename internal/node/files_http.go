package node

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
)

func (s *FileShare) registerHTTP(mux *http.ServeMux, auth func(*http.Request) bool) {
	if s == nil {
		return
	}
	check := func(w http.ResponseWriter, r *http.Request) bool {
		if auth != nil && auth(r) {
			return true
		}
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	mux.HandleFunc("GET /v1/node/projects", func(w http.ResponseWriter, r *http.Request) {
		if !check(w, r) {
			return
		}
		writeJSON(w, map[string]any{"projects": s.Projects()})
	})
	mux.HandleFunc("GET /v1/node/files", func(w http.ResponseWriter, r *http.Request) {
		if !check(w, r) {
			return
		}
		depth := 1
		if raw := r.URL.Query().Get("depth"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil {
				http.Error(w, "invalid depth", http.StatusBadRequest)
				return
			}
			depth = parsed
		}
		list, err := s.List(r.URL.Query().Get("project"), r.URL.Query().Get("path"), depth)
		if err != nil {
			writeFileError(w, err)
			return
		}
		writeJSON(w, list)
	})
	mux.HandleFunc("GET /v1/node/files/read", func(w http.ResponseWriter, r *http.Request) {
		if !check(w, r) {
			return
		}
		body, err := s.ReadFile(r.URL.Query().Get("project"), r.URL.Query().Get("path"))
		if err != nil {
			writeFileError(w, err)
			return
		}
		writeJSON(w, body)
	})
	mux.HandleFunc("PUT /v1/node/files", func(w http.ResponseWriter, r *http.Request) {
		if !check(w, r) {
			return
		}
		var req struct {
			Project      string  `json:"project"`
			Path         string  `json:"path"`
			Content      *string `json:"content"`
			ExpectedETag *string `json:"expected_etag"`
		}
		if err := decodeFileBody(w, r, &req); err != nil {
			return
		}
		if req.Content == nil {
			http.Error(w, "missing content", http.StatusBadRequest)
			return
		}
		entry, err := s.WriteFile(req.Project, req.Path, []byte(*req.Content), req.ExpectedETag)
		if err != nil {
			writeFileError(w, err)
			return
		}
		writeJSON(w, entry)
	})
	mux.HandleFunc("POST /v1/node/files", func(w http.ResponseWriter, r *http.Request) {
		if !check(w, r) {
			return
		}
		var req struct {
			Project string `json:"project"`
			Path    string `json:"path"`
			Kind    string `json:"kind"`
		}
		if err := decodeFileBody(w, r, &req); err != nil {
			return
		}
		if req.Kind != "" && req.Kind != "dir" {
			http.Error(w, "POST creates a directory; use PUT to write a file", http.StatusBadRequest)
			return
		}
		entry, err := s.Mkdir(req.Project, req.Path)
		if err != nil {
			writeFileError(w, err)
			return
		}
		writeJSON(w, entry)
	})
	mux.HandleFunc("POST /v1/node/files/rename", func(w http.ResponseWriter, r *http.Request) {
		if !check(w, r) {
			return
		}
		var req struct {
			Project      string  `json:"project"`
			From         string  `json:"from"`
			To           string  `json:"to"`
			ExpectedETag *string `json:"expected_etag"`
		}
		if err := decodeFileBody(w, r, &req); err != nil {
			return
		}
		entry, err := s.RenameFile(req.Project, req.From, req.To, req.ExpectedETag)
		if err != nil {
			writeFileError(w, err)
			return
		}
		writeJSON(w, entry)
	})
	mux.HandleFunc("DELETE /v1/node/files", func(w http.ResponseWriter, r *http.Request) {
		if !check(w, r) {
			return
		}
		var req struct {
			Project      string  `json:"project"`
			Path         string  `json:"path"`
			ExpectedETag *string `json:"expected_etag"`
		}
		if err := decodeFileBody(w, r, &req); err != nil {
			return
		}
		if err := s.DeleteFile(req.Project, req.Path, req.ExpectedETag); err != nil {
			writeFileError(w, err)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	})
}

func decodeFileBody(w http.ResponseWriter, r *http.Request, dest any) error {
	err := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxFileBytes+1<<20)).Decode(dest)
	if err != nil {
		if errors.Is(err, io.EOF) {
			http.Error(w, "missing request body", http.StatusBadRequest)
			return err
		}
		http.Error(w, "invalid request", http.StatusBadRequest)
		return err
	}
	return nil
}

func writeFileError(w http.ResponseWriter, err error) {
	var conflict *ConflictError
	switch {
	case errors.As(err, &conflict):
		writeJSONStatus(w, http.StatusConflict, map[string]string{
			"error":   "conflict",
			"etag":    conflict.ETag,
			"message": conflict.Error(),
		})
	case errors.Is(err, os.ErrNotExist), errors.Is(err, errNoProject):
		writeJSONStatus(w, http.StatusNotFound, map[string]string{
			"error":   "not_found",
			"message": err.Error(),
		})
	case errors.Is(err, errNotEmpty):
		writeJSONStatus(w, http.StatusConflict, map[string]string{
			"error":   "not_empty",
			"message": err.Error(),
		})
	case errors.Is(err, errPathEscape), errors.Is(err, errSymlink), errors.Is(err, errETagRequired), errors.Is(err, errIsDir):
		writeJSONStatus(w, http.StatusBadRequest, map[string]string{
			"error":   "rejected",
			"message": err.Error(),
		})
	default:
		writeJSONStatus(w, http.StatusBadRequest, map[string]string{
			"error":   "rejected",
			"message": err.Error(),
		})
	}
}
