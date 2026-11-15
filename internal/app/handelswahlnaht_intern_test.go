package app

// **Die Wahl eines Handels an der Naht** — UX-08.AK-04
// (#1680).
//
// Über einen echten Bestand, mit derselben Vorrichtung wie
// `handelsleitungnaht_intern_test.go`: ein Handel ohne Kennung bekommt eine,
// und die Liste mit einem Gegenüber nennt ihn — und mit einem anderen nicht.

import (
	"strings"
	"testing"
)

// **Eröffnen ohne Kennung, dann aus der Liste wählen.**
func TestDieNahtVergibtDieKennungUndNenntDieHandel(t *testing.T) {
	anwendung, _ := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	wunsch := einLeitungswunsch()
	wunsch.Vorgang = ""

	angelegt := anwendung.HandelEroeffnen(ctx, wunsch)
	if angelegt.Fehler != nil {
		t.Fatalf("Eröffnen: %+v", angelegt.Fehler)
	}

	kennung := angelegt.Vorgang.Vorgang
	if !strings.HasPrefix(kennung, "h-") {
		t.Errorf("vergeben wurde %q", kennung)
	}

	mit := anwendung.HandelVorgaenge(ctx, dasLeitungsgegenueber)
	if mit.Fehler != nil {
		t.Fatalf("Vorgänge: %+v", mit.Fehler)
	}

	if len(mit.Vorgaenge) != 1 || mit.Vorgaenge[0].Vorgang != kennung {
		t.Errorf("mit dem Gegenüber: %+v", mit.Vorgaenge)
	}

	// **Mit einem anderen keiner** — und eine leere Liste, kein `null`.
	ohne := anwendung.HandelVorgaenge(ctx, "bjemandanders")
	if ohne.Fehler != nil || ohne.Vorgaenge == nil || len(ohne.Vorgaenge) != 0 {
		t.Errorf("mit einem anderen: %+v", ohne)
	}
}

// **Eine Absage ist eine Antwort mit leerer Liste** und keinem Nullwert —
// das Fenster liest sie, ohne nachzusehen, ob es sie gibt.
func TestDieNahtSagtAbOhneGegenueber(t *testing.T) {
	anwendung, _ := eineLeitungsflaeche(t)

	antwort := anwendung.HandelVorgaenge(riegelkontext(t), "")
	if antwort.Fehler == nil {
		t.Fatal("ohne Gegenüber kam eine Liste")
	}

	if antwort.Vorgaenge == nil {
		t.Error("die Absage trug keine Liste")
	}
}
