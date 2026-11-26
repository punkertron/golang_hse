//go:build !solution

package jsonlist

import (
	"encoding/json"
	"errors"
	"io"
	"reflect"
)

func Marshal(w io.Writer, slice interface{}) error {
	if slice == nil {
		return errors.New("jsonlist: expected a slice, got nil")
	}

	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return &json.UnsupportedTypeError{Type: v.Type()}
	}

	encoder := json.NewEncoder(w)
	for i := 0; i < v.Len(); i++ {
		if err := encoder.Encode(v.Index(i).Interface()); err != nil {
			return err
		}
	}

	return nil
}

func Unmarshal(r io.Reader, slice interface{}) error {
	if slice == nil {
		return errors.New("jsonlist: expected a pointer to a slice, got nil")
	}

	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Slice {
		return &json.UnsupportedTypeError{Type: v.Type()}
	}

	sl := v.Elem()
	elemType := sl.Type().Elem()
	decoder := json.NewDecoder(r)

	for {
		elem := reflect.New(elemType)

		err := decoder.Decode(elem.Interface())
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		sl.Set(reflect.Append(sl, elem.Elem()))
	}
}
