package mcp_test

// **Das Werkzeug für den Verlauf** — Werkstück G5 von Phase 5.5
// (#1462).
//
// **Eine eigene Datei**, weil `partiewerkzeug_test.go` sonst über die
// Zeilengrenze aus §2 ginge — und weil der Schnitt der Frage folgt: dort
// steht, was eine Partie **tut**, hier, was sie **war**.
//
// # Wogegen hier geprüft wird
//
// **Dass die Grenze auch nachträglich hält.** Ein Modell vergisst nichts, was
// es einmal gelesen hat; ein Protokoll, das die fremde Ergänzung mitnähme,
// gäbe hinterher preis, was am Tisch niemand sehen durfte.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// verlaufsprobe gibt einen festen Strom — so, wie `service` ihn gefiltert
// hätte.
type verlaufsprobe struct {
	klagt error
	leer  bool

	gesehen struct{ kennung, fuer string }
}

func (p *verlaufsprobe) PartieVerlauf(_ context.Context, kennung, fuer string) (
	[]mcp.Partieereignis, error,
) {
	p.gesehen.kennung, p.gesehen.fuer = kennung, fuer

	if p.klagt != nil {
		return nil, p.klagt
	}

	if p.leer {
		// **`nil` und nicht eine leere Liste** — das ist, was ein Speicher
		// ohne Zeilen liefert, und genau der Fall soll hier ankommen.
		return nil, nil
	}

	return []mcp.Partieereignis{{
		Folge: "1", Art: "zone.blick",
		Oeffentlich: map[string]any{"sitz": fuer},
		Privat:      map[string]any{"karten": []any{"Wald"}},
		Kette:       "abcd",
	}}, nil
}

// rufeVerlaufswerkzeug ruft `lesen.partieverlauf`.
func rufeVerlaufswerkzeug(
	t *testing.T, quelle mcp.Verlaufswerkstatt, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: probenbestand(), Schreiber: probenbestand(),
		Partienverlauf: quelle,
	}) {
		if werkzeug.Name == "lesen.partieverlauf" {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatal("das Werkzeug lesen.partieverlauf gibt es nicht")

	return "", nil
}

// **Ohne Naht gibt es das Werkzeug gar nicht.**
func TestOhneVerlaufsnahtGibtEsKeinVerlaufswerkzeug(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: probenbestand(), Schreiber: probenbestand(),
	}) {
		if werkzeug.Name == "lesen.partieverlauf" {
			t.Error("das Werkzeug steht ohne Naht da")
		}
	}
}

// **Der Verlauf kommt — mit Partie und Sitz, wie sie gefragt wurden.**
func TestDasWerkzeugVerlaufGibtDenStromEinesSitzes(t *testing.T) {
	t.Parallel()

	probe := &verlaufsprobe{}

	antwort, err := rufeVerlaufswerkzeug(t, probe,
		`{"partieID":"p1","fuer":"sitz-1"}`)
	if err != nil {
		t.Fatalf("lesen.partieverlauf: %v", err)
	}

	if probe.gesehen.kennung != "p1" || probe.gesehen.fuer != "sitz-1" {
		t.Errorf("gerufen mit %+v", probe.gesehen)
	}

	for _, teil := range []string{
		`"folge": "1"`, `"art": "zone.blick"`, `"oeffentlich"`, `"privat"`,
		`"kette": "abcd"`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}
}

// **Ein leerer Verlauf ist eine leere Liste und kein `null`.**
//
// **Weil ein Modell aus `null` „kein Verlauf" läse** und aus `[]` „noch nichts
// geschehen" — zwei verschiedene Auskünfte.
func TestEinLeererVerlaufKommtAlsListe(t *testing.T) {
	t.Parallel()

	antwort, err := rufeVerlaufswerkzeug(t, &verlaufsprobe{leer: true},
		`{"partieID":"p1","fuer":"sitz-1"}`)
	if err != nil {
		t.Fatalf("lesen.partieverlauf: %v", err)
	}

	if strings.TrimSpace(antwort) != "[]" {
		t.Errorf("die Antwort ist %q, erwartet []", antwort)
	}
}

// **Was das Werkzeug abweist, bevor es die Naht erreicht.**
//
// **Dieselbe Abweisung wie an den anderen Flächen** (§16) — und der
// Erfolgsweg steht oben, sonst sagte „wird abgewiesen" nur, dass irgendetwas
// fehlt (§10).
func TestWasDasVerlaufswerkzeugAbweist(t *testing.T) {
	t.Parallel()

	for name, argumente := range map[string]string{
		"ohne Partie": `{"fuer":"sitz-1"}`,
		"ohne Sitz":   `{"partieID":"p1"}`,
		"mit Unsinn":  `{"partieID":1}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			probe := &verlaufsprobe{}

			if _, err := rufeVerlaufswerkzeug(
				t, probe, argumente,
			); err == nil {
				t.Errorf("%s kam durch", argumente)
			}

			// **Und die Naht wurde gar nicht erst gefragt.**
			if probe.gesehen.kennung != "" || probe.gesehen.fuer != "" {
				t.Errorf("die Naht wurde gerufen: %+v", probe.gesehen)
			}
		})
	}
}

// **Und was die Naht klagt, klagt das Werkzeug weiter.**
func TestWasDieVerlaufsnahtKlagtKommtDurch(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("der Bestand ist zu")

	_, err := rufeVerlaufswerkzeug(t, &verlaufsprobe{klagt: kaputt},
		`{"partieID":"p1","fuer":"sitz-1"}`)

	if !errors.Is(err, kaputt) {
		t.Errorf("Fehler %v, erwartet die Klage der Naht", err)
	}
}
