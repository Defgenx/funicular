package utils

import (
	"errors"
	"reflect"
)

func CopyMap(originalMap interface{}) (interface{}, error) {
	var err error = nil
	var newMap = make(map[interface{}]interface{})

	switch reflect.TypeOf(originalMap).Kind() {
	case reflect.Map:
		s := reflect.ValueOf(originalMap)
		for _, key := range s.MapKeys() {
			newMap[key] = s.MapIndex(key)
		}
	default:
		err = errors.New("given parameter must be a map")
	}
	return newMap, err
}