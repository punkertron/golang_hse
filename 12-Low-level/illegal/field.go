//go:build !solution

package illegal

import (
	"reflect"
	"unsafe"
)

func SetPrivateField(obj interface{}, name string, value interface{}) {
	rv := reflect.ValueOf(obj).Elem()
	t := rv.Type()

	n := t.NumField()
	for i := range n {
		if t.Field(i).Name == name {
			ptr := unsafe.Add(rv.Addr().UnsafePointer(), t.Field(i).Offset)
			v := reflect.NewAt(t.Field(i).Type, ptr).Elem()
			v.Set(reflect.ValueOf(value))

			return
		}
	}
}
