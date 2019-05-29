package map_helper

import "reflect"

func Copy(originalMap map[Type]Type1) map[string]interface{} {
	var newMap = make(map[string]interface{})
	for k,v := range originalMap {
		newMap[k] = v
	}
	return newMap
}

func Equal(firstMap map[string]interface{}, otherMap map[string]interface{}) bool {
	return reflect.DeepEqual(firstMap, otherMap)
}