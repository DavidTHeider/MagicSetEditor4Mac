// **Den Bestand sichern über jede Fläche ohne Fenster** —
// #1169,
// Werkstück 2.
//
// # Was hier neu ist
//
// Sichern und Zurückspielen standen bis zum 2026-08-28 **nur** am Fenster,
// mit ✋ als Grund: „mindestens eine Bindung öffnet einen nativen
// Dateidialog". Der Dialog war nie der Kern — er beantwortet *welche Datei*,
// und eine Kommandozeile beantwortet sie im Argument. Was fehlte, war die
// Trennung von Weg und Lebenszyklus, und die hat #1175 gebracht.
//
// # Die Flächen sind nicht gleich, und das steht dabei
//
//   - **Sichern**: alle drei.
//   - **Hineinsehen**: alle drei.
//   - **Zurückspielen**: nur die Kommandozeile. MCP nicht, weil der ganze
//     Bestand nichts ist, was ein Sprachmodell still ersetzen darf (§16,
//     Regel 6); der Dienst nicht, weil er **eine** Verbindung über seine
//     Laufzeit hält und alle seine Nähte sie sich abgeholt haben.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Jede Fläche sichert, und die Sicherung ist wieder lesbar.**
//
// # Die Zusage
//
// Nicht „dieselbe Datei" — drei Sicherungen desselben Bestands unterscheiden
// sich im Zeitstempel ihres Manifests. Die Zusage ist: **jede Fläche kann es,
// und was dabei herauskommt, sagt an jeder Fläche dasselbe über sich** — ein
// Set, drei Karten.
func TestDenBestandSichernUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	eingefahren(t, umgebung)

	ziel := t.TempDir()

	an := starteDienst(t, umgebung)

	for name, sichere := range map[string]func(string){
		"Kommandozeile": func(pfad string) {
			mussKommandozeile(t, umgebung, "bestand-sichern", pfad)
		},
		"MCP": func(pfad string) {
			if text, istFehler := mcpRuf(t, umgebung, "schreiben.bestand.sichern",
				map[string]any{"ziel": pfad}); istFehler {
				t.Fatalf("schreiben.bestand.sichern: %s", text)
			}
		},
		"REST": func(pfad string) {
			status, rumpf := an.ruf(t, http.MethodPost, "/bestand/sicherung",
				`{"ziel":"`+pfad+`"}`)
			if status != http.StatusCreated {
				t.Fatalf("POST bestand/sicherung: %d (%s)", status, rumpf)
			}
		},
	} {
		pfad := filepath.Join(ziel, strings.ToLower(name)+".msesicherung")

		sichere(pfad)

		// **Angesehen wird über alle drei** — dieselbe Datei, und jede muss
		// dasselbe darüber sagen.
		// **Angesehen wird über alle drei — und der Verbund verglichen**,
		// nicht nach einem Wort durchsucht. Der erste Anlauf prüfte nur
		// `"sets": 1`; eine Fläche mit `"karten": 0` wäre durchgekommen
		// (Befund 4 des Zweitblicks über Phase 3.8).
		ausCLI := mussKommandozeile(t, umgebung, "--json", "bestand-schauen", pfad)

		ausMCP, istFehler := mcpRuf(t, umgebung, "lesen.sicherung",
			map[string]any{"pfad": pfad})
		if istFehler {
			t.Fatalf("lesen.sicherung: %s", ausMCP)
		}

		status, ausREST := an.ruf(t, http.MethodGet,
			"/bestand/sicherung?pfad="+url.QueryEscape(pfad), "")
		if status != http.StatusOK {
			t.Fatalf("GET bestand/sicherung: %d (%s)", status, ausREST)
		}

		gleicheAuskunft(t, "die Sicherung von "+name+" — Zeile gegen Modell", ausCLI, ausMCP)
		gleicheAuskunft(t, "die Sicherung von "+name+" — Modell gegen Dienst", ausMCP, ausREST)

		// **Und sie sagt das Richtige.** Zwei gleiche Antworten können beide
		// falsch sein; ein Set und drei Karten sind das, was eingefahren
		// wurde.
		if !strings.Contains(ausCLI, `"sets": 1`) || !strings.Contains(ausCLI, `"karten": 3`) {
			t.Errorf("die Auskunft nennt nicht ein Set und drei Karten:\n%s", ausCLI)
		}
	}
}

// **Ein vorhandenes Ziel wird überall gleich abgewiesen.**
//
// Der Fehlerfall neben dem Erfolg (§16). Eine Sicherung, die eine vorhandene
// Datei überschriebe, nähme jemandem genau das, wofür er sie angelegt hat.
func TestEineSicherungUeberschreibtNirgendsEtwas(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	eingefahren(t, umgebung)

	ziel := filepath.Join(t.TempDir(), "belegt.msesicherung")

	mussKommandozeile(t, umgebung, "bestand-sichern", ziel)

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "ein belegtes Ziel", abweisung{
		kern:   "liegt schon da — eine Sicherung überschreibt nichts",
		klasse: klasseKonflikt,
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung, "bestand-sichern", ziel)

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.bestand.sichern",
				map[string]any{"ziel": ziel})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/bestand/sicherung",
				`{"ziel":"`+ziel+`"}`)
		},
		// **Der Ort kommt im Fenster aus dem Dialog** — dieselbe Datei.
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung, ziel).BestandSichern(t.Context(), "", false).Fehler
		},
	})
}

// **Zurückgespielt wird über die Kommandozeile — und der Zeitpunkt bestätigt.**
//
// Der Aufruf heißt `mse bestand-zurueckspielen <sicherung> <zeitpunkt>`.
//
// # Warum nur dort
//
// Steht im Dateikopf. Was hier zusätzlich geprüft wird: **die Bestätigung
// greift**. Ohne sie wäre `bestand-zurueckspielen` ein Befehl, den man aus der
// Verlaufsliste wiederholt und dabei einen Tagesstand wegwirft.
func TestDasZurueckspielenBrauchtDenZeitpunkt(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	eingefahren(t, umgebung)

	ziel := filepath.Join(t.TempDir(), "stand.msesicherung")

	mussKommandozeile(t, umgebung, "bestand-sichern", ziel)

	// **Danach ändert sich der Bestand** — ein zweites Set, das die Sicherung
	// nicht kennt. Ohne diese Änderung sagte der Lauf darunter nichts: ein
	// Zurückspielen, das nichts verändert, sieht aus wie eines, das wirkt.
	neuesSet(t, umgebung, "Nach der Sicherung")

	var angaben struct {
		AngelegtAm string `json:"angelegtAm"`
	}

	roh := mussKommandozeile(t, umgebung, "--json", "bestand-schauen", ziel)
	if err := json.Unmarshal([]byte(roh), &angaben); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	// **Der falsche Zeitpunkt hält es an.**
	if _, _, code := kommandozeile(t, umgebung,
		"bestand-zurueckspielen", ziel, "1999-01-01"); code == 0 {
		t.Error("das Zurückspielen kam ohne den richtigen Zeitpunkt durch")
	}

	if aus := mussKommandozeile(t, umgebung, "sets"); !strings.Contains(aus, "Nach der Sicherung") {
		t.Errorf("der Bestand wurde trotz falscher Bestätigung ersetzt:\n%s", aus)
	}

	// **Und der richtige lässt es durch.**
	mussKommandozeile(t, umgebung, "bestand-zurueckspielen", ziel, angaben.AngelegtAm)

	if aus := mussKommandozeile(t, umgebung, "sets"); strings.Contains(aus, "Nach der Sicherung") {
		t.Errorf("das Set von nach der Sicherung steht noch da:\n%s", aus)
	}
}
