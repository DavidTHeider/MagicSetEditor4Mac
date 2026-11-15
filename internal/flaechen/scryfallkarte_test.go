// Die übrigen schreibenden Scryfall-Fähigkeiten über jede Fläche
// (#1028).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`** — sie gilt für diese genauso, und zweimal
// dieselbe Begründung liefe auseinander.
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Eine Karte aus einem Scryfall-Namen anlegen — über jede Fläche.**
func TestEineKarteAnlegenUeberJedeFlaeche(t *testing.T) {
	gegenstelle := scryfallgegenstelle(t)

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		setID := magicset(t, umgebung)

		mussKommandozeile(t, umgebung, "scryfall-karte", setID, "Dritte Probe", "--json")
		karteMitNamenPruefen(t, umgebung, setID, "Dritte Probe")
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		setID := magicset(t, umgebung)

		text, istFehler := mcpRuf(t, umgebung, "schreiben.scryfall.karte",
			map[string]any{"setID": zahl(t, setID), "name": "Dritte Probe"})
		if istFehler {
			t.Fatalf("schreiben.scryfall.karte: %s", text)
		}

		karteMitNamenPruefen(t, umgebung, setID, "Dritte Probe")
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		setID := magicset(t, umgebung)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/karten/scryfall",
			`{"name":"Dritte Probe"}`)
		if status != http.StatusCreated {
			t.Fatalf("POST karten/scryfall: %d (%s)", status, rumpf)
		}

		karteMitNamenPruefen(t, umgebung, setID, "Dritte Probe")
	})
}

// **Und ein Name, den es nicht gibt, scheitert überall gleich.**
//
// **Über `abweisungGleich`** seit dem 2026-08-28 — Befund 3 des Nachlaufs über
// Phase 3.8 (#1253).
// Hier stand dreimal „irgendein Fehler", und damit war die Zusage im Namen
// nicht geprüft.
//
// **Eine Umgebung für alle drei**, nicht drei: der Lauf sagt zu, dass
// **dieselbe** Lage überall gleich ausgeht — drei Bestände wären drei Lagen.
func TestEinUnbekannterKartennameScheitertUeberallGleich(t *testing.T) {
	gegenstelle := scryfallgegenstelle(t)
	umgebung := scryfallumgebung(t, gegenstelle)
	setID := magicset(t, umgebung)
	an := starteDienst(t, umgebung)

	abweisungGleich(t, "ein Kartenname, den Scryfall nicht kennt", abweisung{
		klasse: klasseNichtGefunden,
		kern:   "Scryfall",
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).ScryfallKarteAnlegen(t.Context(), setID, "Gibtsnicht", "", "").Fehler
		},
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"scryfall-karte", setID, "Gibtsnicht")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.scryfall.karte",
				map[string]any{"setID": zahl(t, setID), "name": "Gibtsnicht"})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/sets/"+setID+"/karten/scryfall",
				`{"name":"Gibtsnicht"}`)
		},
	})
}

// karteMitNamenPruefen sieht nach, ob eine Karte dieses Namens im Set steht.
//
// **Über die Kommandozeile für alle drei** — die Begründung steht bei
// [satzImBestandPruefen].
func karteMitNamenPruefen(t *testing.T, umgebung []string, setID, name string) {
	t.Helper()

	aus := mussKommandozeile(t, umgebung, "karten", setID, "--json")
	if !strings.Contains(aus, name) {
		t.Errorf("die Karte %q steht nicht im Set — %s", name, aus)
	}
}
