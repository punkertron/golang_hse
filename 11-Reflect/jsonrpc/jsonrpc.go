//go:build !solution

package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"
)

func MakeHandler(service interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := strings.TrimPrefix(r.URL.Path, "/")

		rcpMethod := reflect.ValueOf(service).MethodByName(method)
		if !rcpMethod.IsValid() {
			http.NotFound(w, r)
			return
		}

		reqType := rcpMethod.Type().In(1)
		reqValue := reflect.New(reqType.Elem())

		if err := json.NewDecoder(r.Body).Decode(reqValue.Interface()); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		results := rcpMethod.Call([]reflect.Value{
			reflect.ValueOf(r.Context()),
			reqValue,
		})

		if !results[1].IsNil() {
			err := results[1].Interface().(error)
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(results[0].Interface())
	})
}

var client = &http.Client{
	Timeout: 5 * time.Second,
}

func Call(ctx context.Context, endpoint string, method string, req, rsp interface{}) error {
	buf, err := json.Marshal(req)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint+"/"+method,
		bytes.NewReader(buf),
	)
	if err != nil {
		return err
	}

	responce, err := client.Do(request)
	if err != nil {
		return err
	}

	defer responce.Body.Close()
	// Put this after defer resp.Body.Close().
	if responce.StatusCode < 200 || responce.StatusCode >= 300 {
		body, err := io.ReadAll(responce.Body)
		if err != nil {
			return fmt.Errorf("HTTP %d: read error body: %w", responce.StatusCode, err)
		}

		return fmt.Errorf("HTTP %d: %s", responce.StatusCode, strings.TrimSpace(string(body)))
	}

	return json.NewDecoder(responce.Body).Decode(rsp)
}
