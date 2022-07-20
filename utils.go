package main

import (
	"fmt"
	"os"
	"reflect"
)

// IsFileExist 判断文件是否存在
func IsFileExist(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil || os.IsNotExist(err) || info.IsDir() {
		return false
	}
	return true
}

// ToMapSetStrictE converts a slice or array to map set with error strictly.
// The result of map key type is equal to the element type of input.
func ToMapSetStrictE(i interface{}) (interface{}, error) {
	// check param
	if i == nil {
		return nil, fmt.Errorf("unable to converts %#v of type %T to map[interface{}]struct{}", i, i)
	}
	t := reflect.TypeOf(i)
	kind := t.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return nil, fmt.Errorf("the input %#v of type %T isn't a slice or array", i, i)
	}
	// execute the convert
	v := reflect.ValueOf(i)
	mT := reflect.MapOf(t.Elem(), reflect.TypeOf(struct{}{}))
	mV := reflect.MakeMapWithSize(mT, v.Len())
	for j := 0; j < v.Len(); j++ {
		mV.SetMapIndex(v.Index(j), reflect.ValueOf(struct{}{}))
	}
	return mV.Interface(), nil
}
