package gocsv

import (
	"fmt"
	"io"
	"reflect"
)

type encoder struct {
	out io.Writer
}

func newEncoder(out io.Writer) *encoder {
	return &encoder{out}
}

func writeFromChan(writer *SafeCSVWriter, c <-chan interface{}) error {
	// Get the first value. It wil determine the header structure.
	firstValue, ok := <-c
	if !ok {
		return fmt.Errorf("channel is closed")
	}
	inValue, inType := getConcreteReflectValueAndType(firstValue) // Get the concrete type
	if err := ensureStructOrPtr(inType); err != nil {
		return err
	}
	inInnerWasPointer := inType.Kind() == reflect.Ptr
	inInnerStructInfo := getStructInfo(inType) // Get the inner struct info to get CSV annotations
	csvHeadersLabels := make([]string, len(inInnerStructInfo.Fields))
	for i, fieldInfo := range inInnerStructInfo.Fields { // Used to write the header (first line) in CSV
		csvHeadersLabels[i] = fieldInfo.getFirstKey()
	}
	if err := writer.Write(csvHeadersLabels); err != nil {
		return err
	}
	write := func(val reflect.Value) error {
		if !val.CanAddr() {
			// Make an addressable copy
			//
			// Admittedly this is a kludge but downstream operations (like reflect.Value.Interface())
			// require addressability
			ptr := reflect.New(val.Type())
			ptr.Elem().Set(val)
			val = ptr.Elem()
		}
		for j, fieldInfo := range inInnerStructInfo.Fields {
			csvHeadersLabels[j] = ""
			inInnerFieldValue, err := getInnerField(val, inInnerWasPointer, fieldInfo.IndexChain) // Get the correct field header <-> position
			if err != nil {
				return err
			}
			csvHeadersLabels[j] = inInnerFieldValue
		}
		if err := writer.Write(csvHeadersLabels); err != nil {
			return err
		}
		return nil
	}
	if err := write(inValue); err != nil {
		return err
	}
	for v := range c {
		val, _ := getConcreteReflectValueAndType(v) // Get the concrete type (not pointer) (Slice<?> or Array<?>)
		if err := ensureStructOrPtr(inType); err != nil {
			return err
		}
		if err := write(val); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func writeTo(writer *SafeCSVWriter, in interface{}, omitHeaders bool) error {
	inValue, inType := getConcreteReflectValueAndType(in) // Get the concrete type (not pointer) (Slice<?> or Array<?>)
	if err := ensureInType(inType); err != nil {
		return err
	}
	inInnerWasPointer, inInnerType := getConcreteContainerInnerType(inType) // Get the concrete inner type (not pointer) (Container<"?">)
	if err := ensureInInnerType(inInnerType); err != nil {
		return err
	}
	inInnerStructInfo := getStructInfo(inInnerType) // Get the inner struct info to get CSV annotations
	toWrite := [][]string{}
	if !omitHeaders {
		csvHeadersLabels := make([]string, len(inInnerStructInfo.Fields))
		for i, fieldInfo := range inInnerStructInfo.Fields { // Used to write the header (first line) in CSV
			csvHeadersLabels[i] = fieldInfo.getFirstKey()
		}
		toWrite = append(toWrite, csvHeadersLabels)
	}
	inLen := inValue.Len()
	for i := 0; i < inLen; i++ { // Iterate over container rows
		row := make([]string, len(inInnerStructInfo.Fields))
		for j, fieldInfo := range inInnerStructInfo.Fields {
			inInnerFieldValue, err := getInnerField(inValue.Index(i), inInnerWasPointer, fieldInfo.IndexChain) // Get the correct field header <-> position
			if err != nil {
				return err
			}
			row[j] = inInnerFieldValue
		}
		toWrite = append(toWrite, row)
	}

	for _, row := range toWrite {
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func ensureStructOrPtr(t reflect.Type) error {
	switch t.Kind() {
	case reflect.Struct:
		fallthrough
	case reflect.Ptr:
		return nil
	}
	return fmt.Errorf("cannot use " + t.String() + ", only slice or array supported")
}

// Check if the inType is an array or a slice
func ensureInType(outType reflect.Type) error {
	switch outType.Kind() {
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		return nil
	}
	return fmt.Errorf("cannot use " + outType.String() + ", only slice or array supported")
}

// Check if the inInnerType is of type struct
func ensureInInnerType(outInnerType reflect.Type) error {
	switch outInnerType.Kind() {
	case reflect.Struct:
		return nil
	}
	return fmt.Errorf("cannot use " + outInnerType.String() + ", only struct supported")
}

func getInnerField(outInner reflect.Value, outInnerWasPointer bool, index []int) (string, error) {
	oi := outInner
	if outInnerWasPointer {
		oi = outInner.Elem()
	}
	return getFieldAsString(oi.FieldByIndex(index))
}
