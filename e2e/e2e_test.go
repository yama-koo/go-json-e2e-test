package e2e

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var r *chi.Mux

func init() {
	r = chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/get", func(w http.ResponseWriter, r *http.Request) {
		res := map[string]interface{}{
			"id":      1,
			"message": "hello world",
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		byte, _ := json.Marshal(res)
		_, _ = w.Write(byte)
	})
	r.Post("/post/{id}", func(w http.ResponseWriter, r *http.Request) {
		s := chi.URLParam(r, "id")
		id, err := strconv.Atoi(s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if id == 100 {
			body := map[string]interface{}{
				"error": "invalid argument",
			}
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			byte, _ := json.Marshal(body)
			_, _ = w.Write(byte)
			return
		}

		var req interface{}
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(b, &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		body := map[string]interface{}{
			"id":      id,
			"message": "hello world",
			"data":    req,
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		byte, _ := json.Marshal(body)
		_, _ = w.Write(byte)
	})
	r.Put("/put/{id}", func(w http.ResponseWriter, r *http.Request) {
		s := chi.URLParam(r, "id")
		id, err := strconv.Atoi(s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var req map[string]interface{}
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(b, &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		body := map[string]interface{}{
			"id":      id,
			"message": req["key"],
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		byte, _ := json.Marshal(body)
		_, _ = w.Write(byte)
	})
	r.Patch("/patch/{id}", func(w http.ResponseWriter, r *http.Request) {
		s := chi.URLParam(r, "id")
		id, err := strconv.Atoi(s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var req map[string]interface{}
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(b, &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		body := map[string]interface{}{
			"id":      id,
			"message": req["key"],
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		byte, _ := json.Marshal(body)
		_, _ = w.Write(byte)
	})
	r.Delete("/delete/{id}", func(w http.ResponseWriter, r *http.Request) {
		s := chi.URLParam(r, "id")
		_, err := strconv.Atoi(s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

func TestE2E(t *testing.T) {
	type args struct {
		t            *testing.T
		handler      http.Handler
		path         string
		ignoreFields []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				t:            t,
				handler:      r,
				path:         "testdata",
				ignoreFields: []string{"id"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			E2E(tt.args.t, tt.args.handler, tt.args.path, tt.args.ignoreFields)
		})
	}
}
