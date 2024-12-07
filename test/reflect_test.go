package test

import (
	"fmt"
	"reflect"
	"testing"
)

func TestReflect(t *testing.T) {
	//v := reflect.New(reflect.TypeOf(0)).Elem()
	//v.SetInt(50)
	//fmt.Println(v.Int())

	var x = 3.4

	v := reflect.ValueOf(&x).Elem()
	fmt.Println("Setting a value:")
	v.SetFloat(7.1) // 运行时会报错，因为v不是可设置的（settable）

}
