//go:build !solution

package structtags

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var fieldCache sync.Map // reflect.Type -> map[string]int (var's name -> position in the struct)

func buildFieldIndexes(t reflect.Type) map[string]int {
	fields := make(map[string]int)

	for i := 0; i < t.NumField(); i++ {
		fieldInfo := t.Field(i)
		tag := fieldInfo.Tag

		name := tag.Get("http")
		if name == "" {
			name = strings.ToLower(fieldInfo.Name)
		}

		fields[name] = i
	}

	return fields
}

func fieldsFor(t reflect.Type) map[string]int {
	if cached, ok := fieldCache.Load(t); ok {
		return cached.(map[string]int)
	}

	fields := buildFieldIndexes(t)
	actual, _ := fieldCache.LoadOrStore(t, fields)

	return actual.(map[string]int)
}

func Unpack(req *http.Request, ptr interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	v := reflect.ValueOf(ptr).Elem()
	fields := fieldsFor(v.Type())

	for name, values := range req.Form {
		index, ok := fields[name]
		if !ok || len(values) == 0 {
			continue
		}

		f := v.Field(index)

		if f.Kind() == reflect.Slice {
			start := f.Len()
			f.Grow(len(values))
			tail := f.Slice(start, start+len(values))

			for i, value := range values {
				if err := populate(tail.Index(i), value); err != nil {
					return fmt.Errorf("%s: %v", name, err)
				}

				f.SetLen(start + i + 1)
			}

			continue
		}

		for _, value := range values {
			if err := populate(f, value); err != nil {
				return fmt.Errorf("%s: %v", name, err)
			}
		}
	}

	return nil
}

func populate(v reflect.Value, value string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(value)

	case reflect.Int:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}

		v.SetInt(i)

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}

		v.SetBool(b)

	default:
		return fmt.Errorf("unsupported kind %s", v.Type())
	}

	return nil
}
