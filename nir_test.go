package fejkdata

import (
	"testing"

	"github.com/larvit/fejkdata/internal/fejktest"
)

func TestValidNirAcceptsInseesWorkedExample(t *testing.T) {
	if !fejktest.ValidNir("121127713123250") {
		t.Error("121127713123250 is the example in INSEE's Courrier des statistiques, want it valid")
	}
}

func TestValidNirCountsCorsicasDepartmentsAs19And18(t *testing.T) {
	for _, good := range []string{"185052A04101227", "185052B04101254", "185051904101227", "185051804101254"} {
		if !fejktest.ValidNir(good) {
			t.Errorf("%s has the key of its digits with 2A as 19 and 2B as 18, want it valid", good)
		}
	}
	if fejktest.ValidNir("185052A04101254") {
		t.Error("185052A04101254 has the key of 2B, want it invalid")
	}
}

func TestValidNirRefusesAWrongKeyLengthOrCharacter(t *testing.T) {
	for _, bad := range []string{"121127713123251", "12112771312325", "1211277131232500", "12112771312325x", "185052C04101227", "185052a04101227", "1 21 12 77 131 232 50", ""} {
		t.Run(bad, func(t *testing.T) {
			if fejktest.ValidNir(bad) {
				t.Errorf("%q is no NIR with a valid key", bad)
			}
		})
	}
}

func TestBuiltinNirDrawsTheKeyOfTheThirteenCharactersBefore(t *testing.T) {
	f := emptyGenerator(1)
	for pre, key := range map[string]string{
		"1211277131232": "50",
		"185052A041012": "27",
		"185052B041012": "54",
		"1000000000046": "01",
		"1000000000038": "09",
		"1000000000047": "97",
	} {
		t.Run(pre, func(t *testing.T) {
			if got := mustRender(t, f, `{"format":"{pre}{nir()}","pre":"`+pre+`"}`); got != pre+key {
				t.Errorf("nir() after %s = %q, want the key %s", pre, got, key)
			}
		})
	}
}

func TestBuiltinNirReadsOnlyTheLastThirteenLettersAndDigitsEmitted(t *testing.T) {
	f := emptyGenerator(1)
	for _, pre := range []string{"1 21 12 77 131 232", "99-1211277131232", "x1211277131232", "NIR: 1.21.12.77.131.232"} {
		t.Run(pre, func(t *testing.T) {
			if got := mustRender(t, f, `{"format":"{pre}{nir()}","pre":"`+pre+`"}`); got != pre+"50" {
				t.Errorf("nir() after %q = %q, want the key 50 of 1211277131232", pre, got)
			}
		})
	}
}

func TestBuiltinNirRendersNothingWhereTheThirteenAreNoNumber(t *testing.T) {
	f := emptyGenerator(1)
	for _, pre := range []string{"", "121127713123", "18505cx041012", "185052C041012", "185052a041012", "2A12345678901"} {
		t.Run(pre, func(t *testing.T) {
			if got := mustRender(t, f, `{"format":"{pre}{nir()}","pre":"`+pre+`"}`); got != pre {
				t.Errorf("nir() after %q = %q, want nothing after it", pre, got)
			}
		})
	}
}
