package gocsv

import (
	"fmt"
	"time"
)

type MarshallerStruct struct {
	Foo string
	Bar int
}

func (m MarshallerStruct) MarshalCSV() (string, error) {
	return fmt.Sprintf("%s %d", m.Foo, m.Bar), nil
}

func (m MarshallerStruct) UnmarshalCSV(s string) error {
	_, err := fmt.Sscanf("%s %d", s, &m.Foo, &m.Bar)
	return err
}

type Sample struct {
	Foo        string           `csv:"foo"`
	Bar        int              `csv:"BAR"`
	Baz        string           `csv:"Baz"`
	Frop       float64          `csv:"Quux"`
	Blah       *int             `csv:"Blah"`
	SPtr       *string          `csv:"SPtr"`
	Marshaller MarshallerStruct `csv:"Marshaller"`
	Omit       *string          `csv:"Omit,omitempty"`
}

type EmbedSample struct {
	Qux string `csv:"first"`
	Sample
	Ignore string  `csv:"-"`
	Grault float64 `csv:"garply"`
	Quux   string  `csv:"last"`
}

type SkipFieldSample struct {
	EmbedSample
	MoreIgnore string `csv:"-"`
	Corge      string `csv:"abc"`
}

// Testtype for unmarshal/marshal functions on renamed basic types
type RenamedFloat64Unmarshaler float64
type RenamedFloat64Default float64

type RenamedSample struct {
	RenamedFloatUnmarshaler RenamedFloat64Unmarshaler `csv:"foo"`
	RenamedFloatDefault     RenamedFloat64Default     `csv:"bar"`
}

type MultiTagSample struct {
	Foo string `csv:"Baz,foo"`
	Bar int    `csv:"BAR"`
}

type TagSeparatorSample struct {
	Foo string `csv:"Baz|foo"`
	Bar int    `csv:"BAR"`
}

type DateTime struct {
	Foo time.Time `csv:"Foo"`
}
