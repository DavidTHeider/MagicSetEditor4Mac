package domain

import (
	"strings"
	"testing"
)

// **Eine Funktion lässt sich mit Vorgabe „aus“ ausliefern** — OPS-10.AK-04:
// ohne eigene Wahl ist sie aus, mit „an“ an, und die Einstellung trägt die
// Vorgabe „aus“.
func TestEineVorschauKommtAbgeschaltet(t *testing.T) {
	vorher := schalterverzeichnis
	t.Cleanup(func() { schalterverzeichnis = vorher })

	schalterverzeichnis = []Schalter{{Kennung: "neu", Art: SchalterFunktion, Reife: ReifeVorschau}}

	if !SchalterstandAus(nil).Aus("neu") {
		t.Error("ohne Wahl an")
	}

	if SchalterstandAus(map[string]string{SchalterEinstellung("neu"): SchalterAn}).Aus("neu") {
		t.Error("mit „an“ aus")
	}

	if zeilen := schaltereinstellungen(); len(zeilen) != 1 || zeilen[0].Vorgabe != SchalterAus {
		t.Errorf("%+v", zeilen)
	}
}

// **Eine unlesbare Tafel hält an**, statt jeden Schalter wirkungslos zu
// machen.
func TestEineUnlesbareWegetafelHaeltAn(t *testing.T) {
	defer func() {
		if grund, ist := recover().(string); !ist || !strings.Contains(grund, "unlesbar") {
			t.Errorf("Grund: %v", grund)
		}
	}()

	wegetafelAus([]byte("{"))
	t.Error("eine unlesbare Tafel wurde gelesen")
}
