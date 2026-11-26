//go:build !solution

package reversemap

import "reflect"

func ReverseMap(forward interface{}) interface{} {
	src := reflect.ValueOf(forward)
	t := src.Type()

	keyType := t.Key()
	valueType := t.Elem()

	mapType := reflect.MapOf(valueType, keyType)
	newM := reflect.MakeMapWithSize(mapType, src.Len())

	var iter reflect.MapIter
	iter.Reset(src)
	for iter.Next() {
		newM.SetMapIndex(iter.Value(), iter.Key())
	}

	return newM.Interface()
}
