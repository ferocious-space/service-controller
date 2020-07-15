package ServiceController

import (
	"reflect"
)

func GetType(i interface{}) reflect.Type {
	tp := reflect.TypeOf(i)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	return tp
}
