package gocsv

import (
	"fmt"
	"io"
	"os"
	"reflect"
)

type decoder struct {
	in io.Reader
}

func newDecoder(in io.Reader) *decoder {
	return &decoder{in}
}

func (decode *decoder) getCSVRows() ([][]string, error) {
	return getCSVReader(decode.in).ReadAll()
}

func maybeMissingStructFields(structInfo []fieldInfo, headers []string) error {
	if len(structInfo) == 0 {
		return nil
	}

	headerMap := make(map[string]struct{}, len(headers))
	for idx := range headers {
		headerMap[headers[idx]] = struct{}{}
	}

	for _, info := range structInfo {
		if _, ok := headerMap[info.Key]; !ok {
			return fmt.Errorf("found unmatched struct tag %v", info.Key)
		}
	}
	return nil
}

func (decode *decoder) readTo(out interface{}) error {
	outValue, outType := getConcreteReflectValueAndType(out) // Get the concrete type (not pointer) (Slice<?> or Array<?>)
	if err := decode.ensureOutType(outType); err != nil {
		return err
	}
	outInnerWasPointer, outInnerType := getConcreteContainerInnerType(outType) // Get the concrete inner type (not pointer) (Container<"?">)
	if err := decode.ensureOutInnerType(outInnerType); err != nil {
		return err
	}
	csvRows, err := decode.getCSVRows() // Get the CSV csvRows
	if err != nil {
		return err
	}
	if err := decode.ensureOutCapacity(&outValue, len(csvRows)); err != nil { // Ensure the container is big enough to hold the CSV content
		return err
	}
	outInnerStructInfo := getStructInfo(outInnerType) // Get the inner struct info to get CSV annotations
	if len(outInnerStructInfo.Fields) == 0 {
		return fmt.Errorf("no csv struct tags found")
	}
	headers := csvRows[0]
	body := csvRows[1:]

	csvHeadersLabels := make(map[int]*fieldInfo, len(outInnerStructInfo.Fields)) // Used to store the correspondance header <-> position in CSV

	for i, csvColumnHeader := range headers {
		if fieldInfo := decode.getCSVFieldPosition(csvColumnHeader, outInnerStructInfo); fieldInfo != nil {
			csvHeadersLabels[i] = fieldInfo
		}
	}
	if err := maybeMissingStructFields(outInnerStructInfo.Fields, headers); err != nil {
		if FailIfUnmatchedStructTags {
			return err
		} else {
			fmt.Fprintf(os.Stdout, fmt.Sprintf("not all declared csv structtags matched given input header"))
		}
	}

	for i, csvRow := range body {
		outInner := decode.createNewOutInner(outInnerWasPointer, outInnerType)
		for j, csvColumnContent := range csvRow {
			if fieldInfo, ok := csvHeadersLabels[j]; ok { // Position found accordingly to header name
				if err := decode.setInnerField(&outInner, outInnerWasPointer, fieldInfo.Num, csvColumnContent); err != nil { // Set field of struct
					return err
				}
			}
		}
		outValue.Index(i).Set(outInner)
	}
	return nil
}

// Check if the outType is an array or a slice
func (decode *decoder) ensureOutType(outType reflect.Type) error {
	switch outType.Kind() {
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		return nil
	}
	return fmt.Errorf("cannot use " + outType.String() + ", only slice or array supported")
}

// Check if the outInnerType is of type struct
func (decode *decoder) ensureOutInnerType(outInnerType reflect.Type) error {
	switch outInnerType.Kind() {
	case reflect.Struct:
		return nil
	}
	return fmt.Errorf("cannot use " + outInnerType.String() + ", only struct supported")
}

func (decode *decoder) ensureOutCapacity(out *reflect.Value, csvLen int) error {
	switch out.Kind() {
	case reflect.Array:
		if out.Len() < csvLen-1 { // Array is not big enough to hold the CSV content (arrays are not addressable)
			return fmt.Errorf("array capacity problem: cannot store %d %s in %s", csvLen-1, out.Type().Elem().String(), out.Type().String())
		}
	case reflect.Slice:
		if !out.CanAddr() && out.Len() < csvLen-1 { // Slice is not big enough tho hold the CSV content and is not addressable
			return fmt.Errorf("slice capacity problem and is not addressable (did you forget &?)")
		} else if out.CanAddr() && out.Len() < csvLen-1 {
			out.Set(reflect.MakeSlice(out.Type(), csvLen-1, csvLen-1)) // Slice is not big enough, so grows it
		}
	}
	return nil
}

func (decode *decoder) getCSVFieldPosition(key string, structInfo *structInfo) *fieldInfo {
	for _, field := range structInfo.Fields {
		if field.Key == key {
			return &field
		}
	}
	return nil
}

func (decode *decoder) createNewOutInner(outInnerWasPointer bool, outInnerType reflect.Type) reflect.Value {
	if outInnerWasPointer {
		return reflect.New(outInnerType)
	}
	return reflect.New(outInnerType).Elem()
}

func (decode *decoder) setInnerField(outInner *reflect.Value, outInnerWasPointer bool, fieldPosition int, value string) error {
	if outInnerWasPointer {
		return setField(outInner.Elem().Field(fieldPosition), value)
	}
	return setField(outInner.Field(fieldPosition), value)
}
