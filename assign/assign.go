package assign

import (
	"fmt"
	"reflect"
)

func Structs(target interface{}, source interface{}) error {
	tv := reflect.Indirect(reflect.ValueOf(target))
	sv := reflect.Indirect(reflect.ValueOf(source))

	if tv.Kind() != reflect.Struct || sv.Kind() != reflect.Struct {
		return fmt.Errorf("target and source should be structs")
	}

	for i := 0; i < tv.NumField(); i++ {
		name := tv.Type().Field(i).Name
		tField := tv.Field(i)
		sField := reflect.Indirect(sv.FieldByName(name))

		if tField.CanSet() && sField.IsValid() {
			if sField.Kind() == reflect.Interface && sField.IsNil() {
				continue
			}
			if sField.Kind() == reflect.Slice && sField.IsNil() {
				continue
			}
			if sField.Kind() == reflect.Map && sField.IsNil() {
				continue
			}
			if tField.Kind() == reflect.Ptr {
				tField.Elem().Set(sField)
			} else {
				tField.Set(sField)
			}
		}
	}

	return nil
}
