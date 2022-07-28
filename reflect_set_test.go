package goload

import (
	"fmt"
	"reflect"
	"testing"
)

type Cal struct {
	Num1 float64
	Num2 float64
}

func (cal Cal) GetSub(name string) {
	res := cal.Num1 - cal.Num2
	fmt.Printf("%s完成了减法运算，%f-%f=%f\n", name, cal.Num1, cal.Num2, res)
}

func TestReflectCal(t *testing.T) {
	in := &Cal{
		Num1: 18,
		Num2: 2,
	}

	inType := reflect.TypeOf(in)
	inVal := reflect.ValueOf(in)

	for i := 0; i < inType.Elem().NumField(); i++ {
		fmt.Printf("第%d个字段名为%v,类型为%v\n", i, inType.Elem().Field(i).Name, inType.Elem().Field(i).Type)
	}

	if inVal.Elem().Field(0).CanSet() {
		inVal.Elem().Field(0).SetFloat(8.0)
		fmt.Println("num1 is ok!")
	} else {
		fmt.Println("num1 is not ok!")
	}
	if inVal.Elem().Field(1).CanSet() {
		inVal.Elem().Field(1).SetFloat(3.0)
		fmt.Println("num2 is ok!")
	} else {
		fmt.Println("num2 is not ok!")
	}

	var inArg []reflect.Value

	inArg = append(inArg, reflect.ValueOf("Tom"))
	// append(inArg, reflect.ValueOf("Lisa"))

	inVal.Elem().Method(0).Call(inArg)

}
