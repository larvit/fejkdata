package datatype

import "testing"

func TestEachDatatypeIsSpelledAsDataWritesIt(t *testing.T) {
	for d, want := range map[DataType][2]string{
		String:  {"string", "text"},
		Integer: {"integer", "an integer"},
		Number:  {"number", "a number"},
		Boolean: {"boolean", "a boolean"},
	} {
		if d.String() != want[0] || Noun(d) != want[1] {
			t.Errorf("%d: String = %q, Noun = %q, want %q", int(d), d.String(), Noun(d), want)
		}
	}
	if Count != 4 {
		t.Errorf("Count = %d, want the four datatypes", Count)
	}
}

func TestStringNamesADatatypeNoDataWrites(t *testing.T) {
	if got := DataType(9).String(); got != "DataType(9)" {
		t.Errorf("String = %q, want DataType(9)", got)
	}
}
