package utils

import (
	"errors"
	"reflect"
)

func CopyMap(originalMap interface{}) (newMap interface{}, err error) {
	err = nil
	switch reflect.TypeOf(originalMap).Kind() {
	case reflect.Map:
		s := reflect.ValueOf(originalMap)
		var newMap = make(map[interface{}]interface{})
		for _, key := range s.MapKeys() {
			newMap[key] = s.MapIndex(key)
		}
	default:
		err = errors.New("given parameter must be a map")
	}
	return
}
