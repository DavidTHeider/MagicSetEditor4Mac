package mcp

// **Die Werkzeuge zu den bekannten Gegenüber** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Was hier gemessen wird
//
// **Dass ein Modell lesen und sperren darf — und sonst nichts.** Bestätigen,
// Entsperren und Freigeben sind Vertrauensentscheidungen; ein Werkzeug dafür
// hieße, dass ein Modell entscheidet, wem dieser Rechner Karten überlässt.
//
// **Und dass „keine Naht" kein Werkzeug ergibt** — sonst stünde es im Katalog
// und antwortete mit einem Sturz.

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// kontaktprobe gibt eine feste Liste — oder einen Fehler.
type kontaktprobe struct {
	liste    Peerkontakte
	fehler   error
	gesperrt string
}

func (k *kontaktprobe) Peerkontakte(context.Context) (Peerkontakte, error) {
	return k.liste, k.fehler
}

func (k *kontaktprobe) PeerkontaktSperren(
	_ context.Context, kennung string,
) (Peerkontakte, error) {
	k.gesperrt = kennung

	return k.liste, k.fehler
}

// eineProbe gibt eine Liste mit einem Gegenüber.
func eineProbe() Peerkontakte {
	return Peerkontakte{
		Kontakte: []Peerkontakt{{
			Kennung:       "bwer",
			Fingerabdruck: "1111-2222",
			Stand:         "unbestaetigt",
			Rechte:        []string{},
		}},
		Zwecke: []string{"paarung", "teilen", "handel", "partie"},
	}
}

// **Zwei Werkzeuge, und nur eines schreibt.**
func TestDieKontaktwerkzeugeSindZweiUndEinesSchreibt(t *testing.T) {
	t.Parallel()

	lesend, schreibend := kontaktwerkzeuge(&kontaktprobe{liste: eineProbe()})

	if len(lesend) != 1 || len(schreibend) != 1 {
		t.Fatalf("%d lesende und %d schreibende", len(lesend), len(schreibend))
	}

	if lesend[0].Art != Lesend || !lesend[0].Idempotent {
		t.Errorf("das lesende ist %+v", lesend[0])
	}

	if schreibend[0].Art != Schreibend {
		t.Errorf("das schreibende ist %+v", schreibend[0])
	}
}

// **Das lesende nennt die Kontakte und ihre Zwecke.**
func TestDasLesendeKontaktwerkzeugNenntDieListe(t *testing.T) {
	t.Parallel()

	lesend, _ := kontaktwerkzeuge(&kontaktprobe{liste: eineProbe()})

	antwort, fehler := lesend[0].Ruf(t.Context(), nil)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	for _, teil := range []string{"bwer", "unbestaetigt", "handel"} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("%q steht nicht in %q", teil, antwort)
		}
	}
}

// **Das schreibende sperrt genau den genannten.**
func TestDasSchreibendeKontaktwerkzeugSperrtDenGenannten(t *testing.T) {
	t.Parallel()

	probe := &kontaktprobe{liste: eineProbe()}
	_, schreibend := kontaktwerkzeuge(probe)

	if _, fehler := schreibend[0].Ruf(
		t.Context(), []byte(`{"kennung":"bwer"}`)); fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if probe.gesperrt != "bwer" {
		t.Fatalf("gesperrt wurde %q", probe.gesperrt)
	}
}

// **Krumme Argumente enden vor der Naht.**
func TestKrummeArgumenteEndenVorDerKontaktnaht(t *testing.T) {
	t.Parallel()

	probe := &kontaktprobe{liste: eineProbe()}
	_, schreibend := kontaktwerkzeuge(probe)

	if _, fehler := schreibend[0].Ruf(
		t.Context(), []byte(`{,}`)); fehler == nil {
		t.Fatal("der krumme Rumpf kam durch")
	}

	if probe.gesperrt != "" {
		t.Fatalf("gesperrt wurde trotzdem %q", probe.gesperrt)
	}
}

// **Ein Fehler der Naht kommt als Fehler zurück** — an beiden Werkzeugen.
func TestEinFehlerDerKontaktnahtKommtZurueck(t *testing.T) {
	t.Parallel()

	lesend, schreibend := kontaktwerkzeuge(
		&kontaktprobe{fehler: errors.New("kein Bestand")})

	if _, fehler := lesend[0].Ruf(t.Context(), nil); fehler == nil {
		t.Error("das lesende schwieg")
	}

	if _, fehler := schreibend[0].Ruf(
		t.Context(), []byte(`{"kennung":"bwer"}`)); fehler == nil {
		t.Error("das schreibende schwieg")
	}
}

// **Ohne Naht gibt es keine Werkzeuge.**
func TestOhneKontaktnahtGibtEsKeineWerkzeuge(t *testing.T) {
	t.Parallel()

	if lesend, schreibend := kontaktwerkzeuge(nil); lesend != nil ||
		schreibend != nil {
		t.Fatalf("es kamen %v und %v", lesend, schreibend)
	}
}
