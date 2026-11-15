// Die Skriptkonsole über jede Fläche
// (#1070).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ein Ausdruck läuft über jede Fläche — mit demselben Ergebnis.**
//
// Hier **darf** verglichen werden, anders als bei den Scryfall-Läufen: der
// Quelltext eines Werts ist die Antwort selbst und nicht ihre Verpackung.
// `"ab"` heißt an jeder Fläche dasselbe.
func TestEinenAusdruckAuswertenUeberJedeFlaeche(t *testing.T) {
	umgebung := eigenerBestand(t)
	setID := setImProbebestand(t, umgebung)
	karteID := ersteKarte(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))

	t.Run("Kommandozeile", func(t *testing.T) {
		aus := mussKommandozeile(t, umgebung, "skript", karteID, `"a" + "b"`)

		if strings.TrimSpace(aus) != `"ab"` {
			t.Errorf("Ausgabe %q", aus)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "lesen.skript",
			map[string]any{"karteID": zahl(t, karteID), "ausdruck": `"a" + "b"`})
		if istFehler {
			t.Fatalf("lesen.skript: %s", text)
		}

		if !strings.Contains(text, `\"ab\"`) {
			t.Errorf("Antwort %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/karten/"+karteID+"/skript",
			`{"ausdruck":"\"a\" + \"b\""}`)
		if status != http.StatusOK {
			t.Fatalf("POST skript: %d (%s)", status, rumpf)
		}

		if !strings.Contains(rumpf, `\"ab\"`) {
			t.Errorf("Antwort %s", rumpf)
		}
	})
}

// **Und ein unvollständiger Ausdruck scheitert überall gleich.**
//
// **Über `abweisungGleich`** seit dem 2026-08-28 — Befund 3 des Nachlaufs über
// Phase 3.8 (#1253).
// Hier stand dreimal „irgendein Fehler mit dem Wort »Zeile«", und damit war
// die Zusage im Namen nicht geprüft: drei verschiedene Klassen mit drei
// verschiedenen Sätzen hätten sie erfüllt.
func TestEinUnvollstaendigerAusdruckScheitertUeberallGleich(t *testing.T) {
	umgebung := eigenerBestand(t)
	setID := setImProbebestand(t, umgebung)
	karteID := ersteKarte(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))
	an := starteDienst(t, umgebung)

	abweisungGleich(t, "ein unvollständiger Ausdruck", abweisung{
		klasse: klasseValidierung,
		// **Die Stelle gehört in die Meldung** — an jeder Fläche.
		kern: "Zeile",
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung, "skript", karteID, "1 +")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "lesen.skript",
				map[string]any{"karteID": zahl(t, karteID), "ausdruck": "1 +"})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/karten/"+karteID+"/skript",
				`{"ausdruck":"1 +"}`)
		},
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).Skript(t.Context(), karteID, "1 +").Fehler
		},
	})
}

// **Die Namen der Umgebung stehen an jeder Fläche** —
// #1164.
//
// # Warum das nicht dem Fenster vorbehalten bleibt
//
// Hier stand eine Begründung, die keine Deckung hatte (§10): wer über MCP
// oder Dienst schreibe, habe „einen eigenen Editor mit eigener
// Vervollständigung". Bei MSE-Script hat er das gerade nicht — welche
// Funktionen zur Verfügung stehen, hängt am **Paket auf diesem Rechner** und
// steht in keinem Trainingsstand.
//
// **Verglichen wird die Liste selbst und nicht ihre Verpackung**: `length`
// heißt an jeder Fläche dasselbe.
//
// **Was dieser Lauf nicht zeigt**: dass eine Vorlage `card` und `set`
// dazulegt. Das Probeset hat kein Stilpaket, also bleibt die Bibliothek (§8);
// dass die Liste der **Umgebung** folgt und nicht einer festen Aufzählung,
// steht in `internal/service/skriptnamen_test.go` mit einem Stil daneben.
func TestDieNamenDerUmgebungStehenAnJederFlaeche(t *testing.T) {
	umgebung := eigenerBestand(t)
	setID := setImProbebestand(t, umgebung)
	karteID := ersteKarte(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))

	t.Run("Kommandozeile", func(t *testing.T) {
		aus := mussKommandozeile(t, umgebung, "skriptnamen", karteID)

		for _, name := range []string{"length", "to_upper"} {
			if !strings.Contains(aus, name) {
				t.Errorf("%q fehlt in der Ausgabe", name)
			}
		}
	})

	t.Run("MCP", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "lesen.skriptnamen",
			map[string]any{"karteID": zahl(t, karteID)})
		if istFehler {
			t.Fatalf("lesen.skriptnamen: %s", text)
		}

		for _, name := range []string{"length", "to_upper"} {
			if !strings.Contains(text, name) {
				t.Errorf("%q fehlt in %s", name, text)
			}
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodGet,
			"/karten/"+karteID+"/skriptnamen", "")
		if status != http.StatusOK {
			t.Fatalf("GET skriptnamen: %d (%s)", status, rumpf)
		}

		for _, name := range []string{"length", "to_upper"} {
			if !strings.Contains(rumpf, name) {
				t.Errorf("%q fehlt in %s", name, rumpf)
			}
		}
	})
}

// **Und eine Karte, die es nicht gibt, scheitert überall.**
//
// Die Gegenprobe zum Lauf darüber: ohne sie sagte er auch dann ja, wenn jede
// Fläche eine feste Liste zurückgäbe, ohne je eine Karte zu lesen.
func TestOhneKarteGibtEsNirgendsNamen(t *testing.T) {
	umgebung := eigenerBestand(t)

	t.Run("Kommandozeile", func(t *testing.T) {
		if aus, _, code := kommandozeile(t, umgebung, "skriptnamen", "999999"); code == 0 {
			t.Fatalf("die fehlende Karte kam durch: %s", aus)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "lesen.skriptnamen",
			map[string]any{"karteID": 999999})
		if !istFehler {
			t.Fatalf("die fehlende Karte kam durch: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		if status, rumpf := an.ruf(t, http.MethodGet,
			"/karten/999999/skriptnamen", ""); status == http.StatusOK {
			t.Fatalf("die fehlende Karte kam durch: %s", rumpf)
		}
	})
}

// setImProbebestand fährt das Probeset ein und gibt seine Kennung.
func setImProbebestand(t *testing.T, umgebung []string) string {
	t.Helper()

	return setkennungAus(t, mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))
}
