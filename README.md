# go-json-e2e-test
A simple library for server-level e2e testing in json

## Install

```
go get github.com/yama-koo/go-json-e2e-test
```

## Usage

### E2E(t *testing.T, handler http.Handler, path string, ignoreFields []string)
Recursively searches for `req*.json` in the folder specified by `path`. Make a request to api using the found json and compare the response to the `res*.json` of the same name.

## Example

directory example
```
.
├── go.mod
├── go.sum
├── main_test.go
└── testdata
    └── get
        ├── req.json
        └── res.json
```

testdata/get/req.json
```json
{
  "method": "GET",
  "path": "/get/1",
  "data": null
}
```

testdata/get/res.json
```json
{
  "message": "",
  "statusCode": 200,
  "data": {
    "message": "hello world"
  }
}
```

main_test.go
```go
package main

import (
  "encoding/json"
  "net/http"
  "testing"

  "github.com/go-chi/chi"
  "github.com/go-chi/chi/middleware"
  "github.com/yama-koo/go-json-e2e-test/e2e"
)

func TestMain(t *testing.T) {
  r := chi.NewRouter()
  r.Use(middleware.Logger)

  // simple api
  r.Get("/get/{id}", func(w http.ResponseWriter, r *http.Request) {
    res := map[string]interface{}{
      "id":      1,
      "message": "hello world",
    }

    w.Header().Add("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    byte, _ := json.Marshal(res)
    _, _ = w.Write(byte)
  })

  e2e.E2E(t, r, "testdata", []string{"id"})
}
```

```shell
go test ./...
```

output in case of failure
```log
2020/10/25 21:42:08 "GET http://127.0.0.1:49437/get HTTP/1.1" from 127.0.0.1:49438 - 200 32B in 103.935µs
--- FAIL: TestMain (0.00s)
    /Users/xxx/Documents/example/e2e.go:117: error:  testdata/get/req.json
    /Users/xxx/Documents/example/e2e.go:118:   map[string]interface{}{
            ... // 1 ignored entry
        - 	"message": string("hello world!!!"),
        + 	"message": string("hello world"),
          }
```

## Json format

req.json
```json
{
  "method": "HTTP method",
  "path": "Your api endpoint",
  "data": "Request body"
}
```

res.json
```json
{
  "message": "Expected message",
  "statusCode": "Expected status code",
  "data": "Your api response"
}
```

more [example](./e2e/testdata)

## Supported methods

||
|-|
|GET|
|POST|
|PUT|
|PATCH|
|DELETE|

## Tips
If you want to control the order of execution, you can use a numeric prefix.

```
.
├── go.mod
├── go.sum
├── main_test.go
├── testdata1
│   └── get
│       ├── req.json
│       └── res.json
└── testdata2
    ├── 10_post
    │   ├── req.json
    │   └── res.json
    └── 20_post
        ├── req.json
        └── res.json
```
