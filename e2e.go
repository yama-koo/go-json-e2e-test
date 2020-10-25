package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type Request struct {
	Method string      `json:"method"`
	Path   string      `json:"path"`
	Data   interface{} `json:"data"`
}

type Response struct {
	Message    string      `json:"message"`
	StatusCode int         `json:"statusCode"`
	Data       interface{} `json:"data"`
}

var igf []string

func walkMatch(root string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		pattern := `.+req.*\.json`
		result := regexp.MustCompile(pattern).Match([]byte(path))
		if result {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func E2E(t *testing.T, handler http.Handler, path string, ignoreFields []string) {
	igf = ignoreFields

	ts := httptest.NewServer(handler)
	defer ts.Close()
	client := ts.Client()

	dirs, err := walkMatch(path)
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, file := range dirs {
		err := exec(t, file, client, ts.URL)
		if err != nil {
			t.Error(err.Error())
		}
	}
}

func exec(t *testing.T, path string, client *http.Client, url string) error {
	reqFile, err := os.Open(path)
	if err != nil {
		return err
	}

	byteReqFile, err := ioutil.ReadAll(reqFile)
	if err != nil {
		return err
	}

	var reqMap Request
	err = json.Unmarshal(byteReqFile, &reqMap)
	if err != nil {
		return err
	}

	actual, err := request(client, reqMap, url)
	if err != nil {
		return err
	}

	expect, err := getExpect(path)
	if err != nil {
		return err
	}

	if diff := cmp.Diff(
		expect.StatusCode,
		actual.StatusCode,
		cmp.FilterPath(isIgnoreField, cmp.Ignore()),
	); diff != "" {
		t.Log("error: ", path)
		t.Error(diff)
	}
	if expect.Message != "" {
		if diff := cmp.Diff(expect.Message, actual.Message, cmp.FilterPath(isIgnoreField, cmp.Ignore())); diff != "" {
			t.Log("error: ", path)
			t.Error(diff)
		}
	}
	if diff := cmp.Diff(expect.Data, actual.Data, cmp.FilterPath(isIgnoreField, cmp.Ignore())); diff != "" {
		t.Log("error: ", path)
		t.Error(diff)
	}

	return nil
}

func isIgnoreField(p cmp.Path) bool {
	for _, field := range igf {
		f := `["` + field + `"]`
		if p.Last().String() == f {
			return true
		}
	}
	return false
}

func request(client *http.Client, reqMap Request, url string) (*Response, error) {
	method := reqMap.Method
	path := reqMap.Path
	switch method {
	case "GET":
		r, _ := client.Get(url + path)
		return convertToResponse(r)
	case "POST":
		body, err := json.Marshal(reqMap.Data)
		if err != nil {
			return nil, err
		}
		reader := bytes.NewReader(body)

		r, err := client.Post(url+path, "Content-Type application/json", reader)
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()
		return convertToResponse(r)
	case "PUT":
		body, err := json.Marshal(reqMap.Data)
		if err != nil {
			return nil, err
		}
		reader := bytes.NewReader(body)

		req, err := http.NewRequest("PUT", url+path, reader)
		if err != nil {
			return nil, err
		}

		r, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()
		return convertToResponse(r)
	case "PATCH":
		body, err := json.Marshal(reqMap.Data)
		if err != nil {
			return nil, err
		}
		reader := bytes.NewReader(body)

		req, err := http.NewRequest("PATCH", url+path, reader)
		if err != nil {
			return nil, err
		}

		r, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()
		return convertToResponse(r)
	case "DELETE":
		body, err := json.Marshal(reqMap.Data)
		if err != nil {
			return nil, err
		}
		reader := bytes.NewReader(body)
		req, err := http.NewRequest("DELETE", url+path, reader)
		if err != nil {
			return nil, err
		}

		r, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()
		return convertToResponse(r)
	default:
		return nil, errors.New("method not found")
	}
}

func convertToResponse(res *http.Response) (*Response, error) {
	actual := &Response{
		Message:    res.Status,
		StatusCode: res.StatusCode,
	}

	resByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if len(resByte) == 0 {
		actual.Data = nil
	} else if resByte[0] == '{' {
		var body interface{}
		err = json.Unmarshal(resByte, &body)
		if err != nil {
			return nil, err
		}
		actual.Data = body
	} else {
		actual.Data = string(resByte)
	}

	return actual, nil
}

func getExpect(path string) (*Response, error) {
	reg := regexp.MustCompile(`req`)
	resFile, err := os.Open(reg.ReplaceAllString(path, "res"))
	if err != nil {
		return nil, err
	}

	byteRes, err := ioutil.ReadAll(resFile)
	if err != nil {
		return nil, err
	}

	var expect Response
	err = json.Unmarshal(byteRes, &expect)
	if err != nil {
		return nil, err
	}
	return &expect, nil
}
