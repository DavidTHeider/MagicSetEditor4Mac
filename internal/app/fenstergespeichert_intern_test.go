package app

// **Das Fenster erfährt, was gespeichert ist** — #2535 (Nachlauf über das
// Gesamtreview über 0.9.0, Befund 1). Seit #2396 speichert der Feldweg den
// gerechneten Wert; der Entwurf im Fenster bestätigt genau diesen und nimmt
// einen fremden, der vor dem Nachladen kommt, nicht als seinen. Gefahren über
// dieselbe Naht wie das Fenster: Kartenaufsicht, Feldnaht, Feldweg.

import (
	"strconv"
	"testing"
)

func TestDasFensterNenntDenGespeichertenWert(t *testing.T) {
	anwendung, setID, karteID, _ := nachzugsfenster(t)

	gesetzt := anwendung.FeldSetzenBedingt(probekontext(t), setID,
		strconv.FormatInt(karteID, 10), "zeile", "Glut", nil)
	if gesetzt.Fehler != nil {
		t.Fatalf("FeldSetzenBedingt: %+v", gesetzt.Fehler)
	}

	if gesetzt.Gespeichert == nil || *gesetzt.Gespeichert != "[Glut]" {
		t.Errorf("gespeichert %v, erwartet [Glut]", gesetzt.Gespeichert)
	}
}
