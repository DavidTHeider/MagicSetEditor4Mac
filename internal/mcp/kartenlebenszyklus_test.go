package mcp_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Lebenszyklus einer Karte über MCP** —
// #591.
//
// # Was „vollständig" heißt
//
// `AGENTS.md` §2 sagt, wofür es dieses Paket gibt: „Werkzeuge, mit denen ein
// Sprachmodell Karten **liest und ändert**". Vollständig ist es damit, wenn der
// Lebenszyklus einer Karte darüber erreichbar ist — anlegen, kopieren,
// löschen, umordnen —, und nicht erst, wenn jede der 29 schreibenden Bindungen
// des Fensters ein Gegenstück hat.
//
// **Bis zum 2026-08-22 gab es davon nichts.** Fünf Werkzeuge, davon **eines**
// schreibendes: `schreiben.feld`. Ein Sprachmodell konnte ein Feld setzen und
// sonst nichts ändern — keine Karte anlegen, keine löschen. Das ist der Befund,
// der [0039](../../docs/entscheidungen/0039-kein-dienst-auf-der-rueckschleife.md)
// gegen einen zweiten Zugang entschieden hat: die Lücke lag im **Umfang**.
//
// # Und wie man sie zurücknimmt
//
// [0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md) verlangt,
// dass eine Änderung sagt, wie sie sich zurücknehmen lässt. Bei einem Feld ist
// das der Vorwert; bei einer Karte geht es **über den Verlauf** —
// `KarteLoeschen` schreibt einen Kartenverlaufsschritt, und der Schreiber des
// Servers trägt `domain.UrheberServer`. Die Antwort sagt es, statt es
// vorauszusetzen.
func TestDerLebenszyklusEinerKarteStehtUeberMCP(t *testing.T) {
	faelle := []struct {
		werkzeug  string
		argumente string
		erwartet  string
	}{
		{"schreiben.karte.anlegen", `{"setID": 7}`, "karteID"},
		{"schreiben.karte.kopieren", `{"setID": 7, "karteID": 1}`, "karteID"},
		{"schreiben.karte.loeschen", `{"setID": 7, "karteID": 1}`, "zuruecknahme"},
		{"schreiben.karten.tauschen", `{"setID": 7, "eineID": 1, "andereID": 2}`, "zuruecknahme"},
	}

	for _, fall := range faelle {
		t.Run(fall.werkzeug, func(t *testing.T) {
			bestand := &bestandsprobe{karten: []*domain.CardRow{
				{ID: 1, Number: 1}, {ID: 2, Number: 2},
			}}

			antwort := rufeWerkzeug(t, werkzeugMitNamen(t, bestand, fall.werkzeug),
				fall.argumente)

			if antwort[fall.erwartet] == nil {
				t.Errorf("die Antwort nennt %q nicht: %v", fall.erwartet, antwort)
			}

			// **Jede Änderung sagt, wie sie zurückzunehmen ist** (0010).
			hinweis, istText := antwort["zuruecknahme"].(string)
			if !istText || strings.TrimSpace(hinweis) == "" {
				t.Errorf("keine Zurücknahme genannt: %v", antwort)
			}
		})
	}
}

// **Und jedes weist eine unbrauchbare Eingabe ab, mit dem Feld** (§8).
//
// Der Kasten in #591 sagt es so: „Jede Eingabe wird validiert, bevor sie einen
// Dienst erreicht; ein Befund nennt das **Feld**." Eine Kennung von null ist
// keine Kennung, und die Meldung muss sagen, welche gemeint war — sonst sucht
// der Aufrufer sie in vier Feldern.
func TestDerLebenszyklusWeistUnbrauchbareKennungenAb(t *testing.T) {
	faelle := map[string]struct{ werkzeug, argumente, feld string }{
		"anlegen ohne Set":       {"schreiben.karte.anlegen", `{"setID": 0}`, "setID"},
		"kopieren ohne Karte":    {"schreiben.karte.kopieren", `{"setID": 7, "karteID": 0}`, "karteID"},
		"löschen ohne Karte":     {"schreiben.karte.loeschen", `{"setID": 7, "karteID": 0}`, "karteID"},
		"tauschen ohne die eine": {"schreiben.karten.tauschen", `{"setID": 7, "eineID": 0, "andereID": 2}`, "eineID"},

		// **Und ohne das Set** — bei jedem der drei, seit die Setgrenze in
		// `service` geprüft wird
		// (#801).
		// Ohne diese drei Fälle wäre die neue Prüfung an drei von vier Stellen
		// unbetreten.
		"kopieren ohne Set": {"schreiben.karte.kopieren", `{"karteID": 1}`, "setID"},
		"löschen ohne Set":  {"schreiben.karte.loeschen", `{"karteID": 1}`, "setID"},
		"tauschen ohne Set": {"schreiben.karten.tauschen", `{"eineID": 1, "andereID": 2}`, "setID"},

		// **Und ohne die andere.** Zwei Kennungen heißen zwei Prüfungen; ein
		// Test für die erste ließe die zweite unbetreten — und ihre Meldung
		// nennt ein anderes Feld, sonst suchte der Aufrufer im falschen.
		"tauschen ohne die andere": {"schreiben.karten.tauschen", `{"setID": 7, "eineID": 1, "andereID": 0}`, "andereID"},
		"tauschen mit sich":        {"schreiben.karten.tauschen", `{"setID": 7, "eineID": 2, "andereID": 2}`, "andereID"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			bestand := &bestandsprobe{}

			_, err := werkzeugMitNamen(t, bestand, fall.werkzeug).
				Ruf(t.Context(), []byte(fall.argumente))
			if err == nil {
				t.Fatal("die Eingabe wurde übernommen")
			}

			if !strings.Contains(err.Error(), fall.feld) {
				t.Errorf("die Meldung nennt %q nicht: %v", fall.feld, err)
			}
		})
	}
}

// **Und jedes reicht die beiden Fehlerwege durch** (§3).
//
// **Zwei, nicht einer**: kaputte Argumente scheitern **vor** dem Bestand, ein
// klemmender Bestand **danach**. Ein Test für einen von beiden ließe den
// anderen unbetreten — und der zweite ist der, den ein Sprachmodell im Betrieb
// wirklich trifft.
func TestDerLebenszyklusReichtSeineFehlerDurch(t *testing.T) {
	werkzeuge := map[string]string{
		"schreiben.karte.anlegen":   `{"setID": 7}`,
		"schreiben.karte.kopieren":  `{"setID": 7, "karteID": 1}`,
		"schreiben.karte.loeschen":  `{"setID": 7, "karteID": 1}`,
		"schreiben.karten.tauschen": `{"setID": 7, "eineID": 1, "andereID": 2}`,
	}

	for name, argumente := range werkzeuge {
		t.Run(name, func(t *testing.T) {
			t.Run("kaputte Argumente", func(t *testing.T) {
				_, err := werkzeugMitNamen(t, &bestandsprobe{}, name).
					Ruf(t.Context(), []byte(`{"setID":`))
				if err == nil || !strings.Contains(err.Error(), "kein JSON") {
					t.Errorf("Fehler %v", err)
				}
			})

			t.Run("der Bestand klemmt", func(t *testing.T) {
				bestand := &bestandsprobe{schreibfehler: errKlemmt}

				if _, err := werkzeugMitNamen(t, bestand, name).
					Ruf(t.Context(), []byte(argumente)); err == nil {
					t.Error("der Fehler des Bestands ging verloren")
				}
			})
		})
	}
}

// errKlemmt ist der Fehler, den die Probe beim Schreiben gibt.
var errKlemmt = errors.New("die Ablage antwortet nicht")
