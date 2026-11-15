package mcp_test

// **Die vier Verwaltungswerkzeuge einer Partie** — PLAY-03.AK-05
// (#1711).

import (
	"errors"
	"strings"
	"testing"
)

// **Jedes der vier reicht sein Argument durch und gibt die Antwort zurück.**
func TestDieVierVerwaltungswerkzeugeGehenDurch(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		argumente string
		erwartet  string
	}{
		"lesen.partie.ausfuhr": {
			`{"partieID":"p1"}`, `"fassung\":1}`,
		},
		"schreiben.partie.loeschen": {`{"partieID":"p1"}`, `"befunde"`},
		"schreiben.partie.kopieren": {`{"partieID":"p1"}`, "p1-kopie"},
		"schreiben.partie.einfuhr": {
			`{"inhalt":"{\"fassung\":1}"}`, "ein fremdes Feld",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			probe := &partienprobe{}

			aus, fehler := rufePartiewerkzeug(t, probe, name, fall.argumente)
			if fehler != nil {
				t.Fatalf("%s: %v", name, fehler)
			}

			if !strings.Contains(aus, fall.erwartet) {
				t.Errorf("die Antwort lautet %q", aus)
			}
		})
	}

	// **Und das Argument kommt unten an** — nicht bloß der Ruf (§10).
	probe := &partienprobe{}

	if _, fehler := rufePartiewerkzeug(
		t, probe, "schreiben.partie.kopieren", `{"partieID":"p7"}`,
	); fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	if probe.gesehen.kennung != "p7" {
		t.Errorf("unten kam %q an", probe.gesehen.kennung)
	}
}

// **Ein Ruf ohne Angabe wird abgewiesen** — und zwar mit einem Satz, der das
// fehlende Feld nennt (§8).
func TestDieVerwaltungswerkzeugeWeisenLeereArgumenteAb(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		argumente string
		teil      string
	}{
		"lesen.partie.ausfuhr":      {`{"partieID":""}`, "Kennung"},
		"schreiben.partie.loeschen": {`{"partieID":""}`, "Kennung"},
		"schreiben.partie.kopieren": {`{}`, "Kennung"},
		"schreiben.partie.einfuhr":  {`{"inhalt":""}`, "Text"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := rufePartiewerkzeug(
				t, &partienprobe{}, name, fall.argumente)
			if fehler == nil {
				t.Fatal("der leere Ruf ging durch")
			}

			if !strings.Contains(fehler.Error(), fall.teil) {
				t.Errorf("die Meldung %q nennt %q nicht", fehler, fall.teil)
			}
		})
	}
}

// **Und unlesbare Argumente ebenso** — der zweite Weg, und er ist ein
// anderer als der leere (§10).
func TestDieVerwaltungswerkzeugeWeisenUnlesbareArgumenteAb(t *testing.T) {
	t.Parallel()

	// **Jedes Werkzeug bekommt seine eigene Zahl.** Ein `partieID: 42` an
	// der Einfuhr wäre ein Feld, das sie gar nicht liest — der Ruf ginge
	// durch, und der Zweig dahinter bliebe unbetreten (§10).
	faelle := map[string]string{
		"lesen.partie.ausfuhr":      `{"partieID":42}`,
		"schreiben.partie.loeschen": `{"partieID":42}`,
		"schreiben.partie.kopieren": `{"partieID":42}`,
		"schreiben.partie.einfuhr":  `{"inhalt":42}`,
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := rufePartiewerkzeug(
				t, &partienprobe{}, name, argumente,
			); fehler == nil {
				t.Error("die Zahl ging als Angabe durch")
			}
		})
	}
}

// **Ein klagender Bestand klagt nach oben.**
func TestDieVerwaltungswerkzeugeReichenDenFehlerDurch(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"lesen.partie.ausfuhr":      `{"partieID":"p1"}`,
		"schreiben.partie.loeschen": `{"partieID":"p1"}`,
		"schreiben.partie.kopieren": `{"partieID":"p1"}`,
		"schreiben.partie.einfuhr":  `{"inhalt":"{}"}`,
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			probe := &partienprobe{klagt: errors.New("der Bestand schweigt")}

			if _, fehler := rufePartiewerkzeug(
				t, probe, name, argumente); fehler == nil {
				t.Error("die Klage ging verloren")
			}
		})
	}
}
