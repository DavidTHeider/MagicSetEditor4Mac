// **Die Werkzeuge der Preise** — Werkstück 3 von Phase 3.4
// (#1165).
//
// Aus `sammlung_test.go` herausgelöst, als es über 400 Zeilen ging (§2) — die
// Preisreihe hat mit A3 (#1442) sieben Teile bekommen, und jeder Fall nennt
// sie alle. Der Schnitt folgt der Frage: dort stehen die Werkzeuge der
// **Sammlung**, hier die der **Preise**.
package mcp_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Die Preise tragen ihre Angaben.**
func TestDieWerkzeugeDerPreiseGebenIhreAngaben(t *testing.T) {
	t.Parallel()

	for name, argumente := range map[string]string{
		"lesen.preise":          `{"printing":"probe-druck"}`,
		"schreiben.preiseholen": `{"printing":"probe-druck"}`,
		"lesen.preisverlauf": `{"printing":"probe-druck","quelle":"scryfall","markt":"paper","art":"retail","waehrung":"EUR",` +
			`"finish":"nonfoil"}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort, err := rufeSammlungswerkzeug(t, neueSammlungsprobe(),
				neuePreisprobe(), name, argumente)
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}

			// **Quelle und Alter stehen dabei.** Eine Zahl ohne sie wäre eine
			// Behauptung, für die niemand einsteht.
			for _, teil := range []string{"scryfall", "alterStunden", "4.17"} {
				if !strings.Contains(antwort, teil) {
					t.Errorf("die Antwort nennt %q nicht: %s", teil, antwort)
				}
			}
		})
	}
}

// **Die Übersicht sagt, wie viele ohne Preis blieben.**
func TestDasWerkzeugPortfolioNenntDieOhnePreis(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, neueSammlungsprobe(), neuePreisprobe(),
		"lesen.portfolio", `{}`)
	if err != nil {
		t.Fatalf("lesen.portfolio: %v", err)
	}

	if !strings.Contains(antwort, "ohnePreis") {
		t.Errorf("die Antwort ist %s", antwort)
	}
}

// **Und ein Fehler der Naht kommt durch — auf jedem der vier Preiswege.**
func TestEinFehlerDerPreisnahtKommtDurchsWerkzeug(t *testing.T) {
	t.Parallel()

	boese := &stolpernderPreis{}

	for name, argumente := range map[string]string{
		"lesen.preise":          `{"printing":"probe-druck"}`,
		"schreiben.preiseholen": `{"printing":"probe-druck"}`,
		"lesen.portfolio":       `{}`,
		"lesen.preisverlauf": `{"printing":"probe-druck","quelle":"scryfall","markt":"paper","art":"retail","waehrung":"EUR",` +
			`"finish":"nonfoil"}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := rufeSammlungswerkzeug(t, neueSammlungsprobe(), boese,
				name, argumente); err == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}
}

// **Und ebenso auf den fünf Wegen der Sammlung.**
func TestEinFehlerDerSammlungsnahtKommtDurchsWerkzeug(t *testing.T) {
	t.Parallel()

	boese := &stolperndeSammlung{}

	for name, argumente := range map[string]string{
		"lesen.sammlung":             `{}`,
		"lesen.sammlungfuer":         `{"printing":"x"}`,
		"schreiben.exemplar":         `{"printing":"x","menge":1}`,
		"schreiben.exemplaraendern":  `{"id":1,"printing":"x","menge":1}`,
		"schreiben.exemplarloeschen": `{"id":1}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := rufeSammlungswerkzeug(t, boese, neuePreisprobe(),
				name, argumente); err == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}
}

// nilprobe gibt `nil` statt leerer Listen.
//
// # Warum es diese Vorrichtung gibt
//
// Eine Naht darf `nil` zurückgeben — Go macht daraus beim Lesen aus einer
// leeren Abfrage genau das. Über JSON wird daraus `null`, und ein Modell, das
// eine Liste erwartet, bekommt etwas, das es nicht durchlaufen kann (§3). Die
// Werkzeuge fangen es ab; ohne diesen Lauf wäre das eine Behauptung.
type nilprobe struct{}

func (nilprobe) Sammlung(context.Context) ([]mcp.Exemplarsicht, error) {
	return nil, nil
}

func (nilprobe) SammlungFuer(context.Context, string) ([]mcp.Exemplarsicht, error) {
	return nil, nil
}

func (nilprobe) ExemplarAnlegen(context.Context, mcp.Exemplarsicht) (int64, error) {
	return 1, nil
}

func (nilprobe) ExemplarAendern(context.Context, mcp.Exemplarsicht) error { return nil }
func (nilprobe) ExemplarLoeschen(context.Context, int64, int64) error     { return nil }

func (nilprobe) PreiseHolen(context.Context, string) (
	[]mcp.Preissicht, []string, error,
) {
	return nil, nil, nil
}

func (nilprobe) Preise(context.Context, string) ([]mcp.Preissicht, error) {
	return nil, nil
}

func (nilprobe) Preisverlauf(context.Context, mcp.Preisreihe) (
	[]mcp.Preissicht, error,
) {
	return nil, nil
}

func (nilprobe) Preisbulk(context.Context, string, string) (
	mcp.Bulkbilanzsicht, error,
) {
	return mcp.Bulkbilanzsicht{}, nil
}

func (nilprobe) Marktangebote(context.Context, mcp.Marktwunsch) (
	mcp.Marktsicht, error,
) {
	return mcp.Marktsicht{}, nil
}

func (nilprobe) Anbieterstand(context.Context, string) (mcp.Anbietersicht, error) {
	return mcp.Anbietersicht{}, nil
}

func (nilprobe) Fremdpreisverlauf(context.Context, mcp.Verlaufsbezug) (
	mcp.Verlaufssicht, error,
) {
	return mcp.Verlaufssicht{}, nil
}

func (nilprobe) KIModelle(context.Context) (mcp.KIKatalogsicht, error) {
	return mcp.KIKatalogsicht{}, nil
}

func (nilprobe) KIVorschlag(context.Context, mcp.KIVorschlagsbezug) (
	mcp.KIVorschlagssicht, error,
) {
	return mcp.KIVorschlagssicht{}, nil
}

// KIBatch gibt **nichts** zurück — der Weg, auf dem die leeren Listen
// entstehen (A13.1, #1452).
func (nilprobe) KIBatch(context.Context, mcp.KIBatchbezug) (
	mcp.KIBatchsicht, error,
) {
	return mcp.KIBatchsicht{}, nil
}

func (nilprobe) Portfolio(context.Context) (mcp.Portfoliosicht, error) {
	return mcp.Portfoliosicht{}, nil
}

// **Aus `nil` wird eine leere Liste und kein `null`.**
func TestAusNilWirdEineLeereListe(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		argumente string
		feld      string
	}{
		"lesen.sammlung":        {`{}`, `"Exemplare": []`},
		"lesen.sammlungfuer":    {`{"printing":"x"}`, `"Exemplare": []`},
		"lesen.preise":          {`{"printing":"x"}`, `"Staende": []`},
		"schreiben.preiseholen": {`{"printing":"x"}`, `"Befunde": []`},
		"lesen.portfolio":       {`{}`, `"zeilen": []`},
		"lesen.preisverlauf": {
			`{"printing":"x","quelle":"scryfall","markt":"paper",` +
				`"art":"retail","waehrung":"EUR","finish":"nonfoil"}`,
			`"Staende": []`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, nilprobe{},
				name, fall.argumente)
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}

			if !strings.Contains(antwort, fall.feld) {
				t.Errorf("die Antwort ist %s", antwort)
			}
		})
	}
}

// **Ein Löschen gegen einen überholten Stand wird abgewiesen.**
//
// Der Fall, den Befund 10 des Nachlaufs über Phase 3.4 benannt hat: jemand
// liest, ein anderer ändert, der erste löscht seinen alten Stand — und nähme
// eine Änderung mit, die er nie gesehen hat.
func TestEinLoeschenGegenEinenAltenStandWirdUeberMCPAbgewiesen(t *testing.T) {
	t.Parallel()

	sammlung := neueSammlungsprobe()

	if _, err := rufeSammlungswerkzeug(t, sammlung, neuePreisprobe(),
		"schreiben.exemplaraendern",
		`{"id":1,"fassung":1,"printing":"probe-druck","menge":9}`); err != nil {
		t.Fatalf("ändern: %v", err)
	}

	if _, err := rufeSammlungswerkzeug(t, sammlung, neuePreisprobe(),
		"schreiben.exemplarloeschen",
		`{"id":1,"fassung":1}`); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("der Fehler ist %v — will ErrConflict", err)
	}

	if len(sammlung.zeilen) != 1 {
		t.Errorf("die Zeile ist trotz Konflikt weg")
	}
}

// **Eine halbe Preisreihe wird auch über MCP abgewiesen** — A3 (#1442).
//
// **Dieselbe Zusage an jeder Fläche** (§16): welche Reihe gemeint ist, sagt
// der Aufrufer. Eine Naht, die „Scryfall, Papier, Verkauf" ergänzte,
// beantwortete eine Frage, die niemand gestellt hat.
func TestEineHalbePreisreiheWirdUeberMCPAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		argumente string
		nennt     string
	}{
		"gar nichts": {
			`{"printing":"probe-druck"}`, "art, finish, markt, quelle, waehrung",
		},
		"ohne Quelle": {
			`{"printing":"probe-druck","markt":"paper","art":"retail",` +
				`"waehrung":"EUR","finish":"nonfoil"}`, "quelle",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
				"lesen.preisverlauf", fall.argumente)
			if err == nil {
				t.Fatal("es kam durch")
			}

			if !strings.Contains(err.Error(), fall.nennt) {
				t.Errorf("die Meldung nennt %q nicht: %v", fall.nennt, err)
			}
		})
	}
}

// **Das Werkzeug für den Bulkbezug gibt seine Bilanz.**
func TestDasWerkzeugPreisbulkGibtSeineBilanz(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.preise.bulk",
		`{"quelle":"mtgjson","umfang":"heute"}`)
	if err != nil {
		t.Fatalf("schreiben.preise.bulk: %v", err)
	}

	for _, teil := range []string{
		`"quelle": "mtgjson"`, `"umfang": "heute"`,
		`"drucke": 1`, `"staende": 1`, `"befunde": []`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}
}

// **Und ein Fehler der Naht kommt durch.**
func TestEinFehlerDesBulkbezugsKommtDurchsWerkzeug(t *testing.T) {
	t.Parallel()

	if _, err := rufeSammlungswerkzeug(t, nilprobe{}, stolpernderPreis{},
		"schreiben.preise.bulk",
		`{"quelle":"mtgjson","umfang":"heute"}`); !errors.Is(err, errGestolpertMCP) {
		t.Errorf("der Fehler ist %v", err)
	}
}

// **Ein Rumpf, der kein Objekt ist, wird abgewiesen.**
func TestEinKrummerRumpfDesBulkbezugsWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.preise.bulk", `[1, 2, 3]`); err == nil {
		t.Error("es kam durch")
	}
}

// **Und die zwei Wege der Messung** — A9, #1448.
func (nilprobe) KIModellPruefen(context.Context, mcp.KIProbenbezug) (
	mcp.KIProbensicht, error,
) {
	return mcp.KIProbensicht{}, nil
}

func (nilprobe) KIModellVergessen(context.Context, mcp.KIProbenbezug) error {
	return nil
}

func (nilprobe) KIGeraetemodellFuerSitzung(_ context.Context, an bool) (bool, error) {
	return an, nil
}

// KIErkennung gibt den **Nullwert** — A13.3 (#1452).
//
// **Das ist der Zweck dieses Doppels**: jede Liste darin ist `nil`, und der
// Lauf darüber prüft, dass daraus `[]` wird und nicht `null`. Für ein
// Sprachmodell sieht `null` aus wie „konnte nicht beantwortet werden".
func (nilprobe) KIErkennung(context.Context, mcp.KIErkennungsbezug) (
	mcp.KIErkennungssicht, error,
) {
	return mcp.KIErkennungssicht{}, nil
}
