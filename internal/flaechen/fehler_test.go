// **Dieselben fachlichen Fehler an beiden Flächen** — die zweite Hälfte der
// Abnahme von Phase T
// (#886).
//
// # Was „dieselben" heißt und was nicht
//
// **Nicht derselbe Wortlaut.** Eine Kommandozeile schreibt auf den
// Fehlerstrom, MCP schickt ein `isError` mit Text; sie einander anzugleichen
// hieße, eine Fläche der anderen anzupassen, statt beide dem Fachkern.
//
// **Sondern derselbe Ausgang.** Wo die eine „nein" sagt, sagt die andere es
// auch — und aus demselben Grund. Eine Fläche, die ein unbekanntes Set still
// als leere Liste beantwortet, während die andere abweist, hat einen eigenen
// Fachkern.
package flaechen_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Dieselbe Sache, an beiden Flächen, mit vergleichbarer Meldung.**
//
// # Warum das anders aussieht als vorher
//
// Hier standen drei Tests, die je Fläche prüften, **dass** etwas scheitert —
// und beim unbekannten Set auf zwei **verschiedenen** Vorgängen: die
// Kommandozeile las Karten, MCP sicherte ein Set. Damit war die Abnahmezusage
// „mit denselben fachlichen Fehlern" nicht geprüft, sondern nur „beide sagen
// irgendwie nein" (Zweitblick über Phase T).
//
// Geprüft wird jetzt je Fall **derselbe** Vorgang, und beide Meldungen müssen
// das nennen, woran es lag. Das ist das Stärkste, was sich vergleichen lässt:
// die Kommandozeile gibt einen Rückgabewert und einen deutschen Satz, MCP ein
// `isError` und einen Text — eine gemeinsame Fehlerstruktur haben sie nicht.
//
// **Was er nicht prüft**: dass beide **denselben Satz** sagen. Sie sollen es
// nicht: eine Kommandozeile spricht mit einem Menschen, MCP mit einem Modell.
func TestBeideFlaechenWeisenDasselbeMitDemselbenGrundAb(t *testing.T) {
	faelle := []struct {
		name          string
		cli           func(ziel string) []string
		werkzeug      string
		eingabe       func(ziel string) map[string]any
		restMethode   string
		restWeg       func(ziel string) string
		restRumpf     func(ziel string) string
		mussVorkommen func(ziel string) string

		// klasse ist die Fehlerklasse, die **alle vier** melden müssen.
		klasse string

		// fenster ist derselbe Vorgang über die Bindung des Fensters (C04);
		// der Ort kommt dort aus dem Dialog.
		fenster func(b *app.Bindungen, ziel string) *app.Fehler
	}{
		{
			name:     "ein Set, das es nicht gibt",
			cli:      func(ziel string) []string { return []string{"sichern", "4711", filepath.Join(ziel, "x.mse-set")} },
			werkzeug: "schreiben.set.sichern",
			eingabe: func(ziel string) map[string]any {
				return map[string]any{"setID": 4711, "ziel": filepath.Join(ziel, "y.mse-set")}
			},
			restMethode:   "POST",
			restWeg:       func(string) string { return "/sets/4711/sicherung" },
			restRumpf:     func(ziel string) string { return `{"ziel":"` + ziel + `/z.mse-set"}` },
			mussVorkommen: func(string) string { return "4711" },
			klasse:        klasseNichtGefunden,
			fenster: func(b *app.Bindungen, _ string) *app.Fehler {
				return b.SetSpeichernUnter(context.Background(), "4711").Fehler
			},
		},
		{
			name:     "eine Karte, die es nicht gibt",
			cli:      func(ziel string) []string { return []string{"zeichnen", "4711", filepath.Join(ziel, "eins.png")} },
			werkzeug: "schreiben.karte.zeichnen",
			eingabe: func(ziel string) map[string]any {
				return map[string]any{"karteID": 4711, "ziel": filepath.Join(ziel, "zwei.png")}
			},
			restMethode:   "POST",
			restWeg:       func(string) string { return "/karten/4711/bild" },
			restRumpf:     func(ziel string) string { return `{"ziel":"` + ziel + `/drei.png"}` },
			mussVorkommen: func(string) string { return "4711" },
			klasse:        klasseNichtGefunden,
			fenster: func(b *app.Bindungen, ziel string) *app.Fehler {
				return b.BildSchreiben(context.Background(), "4711", ziel, "vier", "png", 0, "ersetzen", "").Fehler
			},
		},
		{
			name:     "ein Pfad, den es nicht gibt",
			cli:      func(ziel string) []string { return []string{"oeffnen", filepath.Join(ziel, "gibtsnicht.mse-set")} },
			werkzeug: "schreiben.set.einfahren",
			eingabe: func(ziel string) map[string]any {
				return map[string]any{"pfad": filepath.Join(ziel, "gibtsnicht.mse-set")}
			},
			restMethode: "POST",
			restWeg:     func(string) string { return "/sets" },
			restRumpf: func(ziel string) string {
				return `{"pfad":"` + filepath.Join(ziel, "gibtsnicht.mse-set") + `"}`
			},
			mussVorkommen: func(string) string { return "gibtsnicht" },
			klasse:        klasseNichtGefunden,
			fenster: func(b *app.Bindungen, ziel string) *app.Fehler {
				return b.SetAusDatei(context.Background(), filepath.Join(ziel, "gibtsnicht.mse-set")).Fehler
			},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			umgebung := eigenerBestand(t)
			ziel := t.TempDir()
			an := starteDienst(t, umgebung)

			// **Über `abweisungGleich` und nicht von Hand** — Befund 3 des
			// Nachlaufs über Phase 3.8
			// (#1253).
			//
			// Hier stand dreimal „nicht null" plus ein Textfragment. Damit
			// konnten Kommandozeile, MCP und Dienst **verschiedene Klassen**
			// melden und der Lauf blieb grün — genau die Zusage, die im Namen
			// steht, war nicht geprüft.
			abweisungGleich(t, fall.name, abweisung{
				klasse: fall.klasse,
				kern:   fall.mussVorkommen(ziel),
				cli: func() (string, int) {
					aus, fehlerstrom, code := kommandozeile(t, umgebung, fall.cli(ziel)...)

					return aus + fehlerstrom, code
				},
				mcp: func() (string, bool) {
					return mcpRuf(t, umgebung, fall.werkzeug, fall.eingabe(ziel))
				},
				rest: func() (int, string) {
					return an.ruf(t, fall.restMethode, fall.restWeg(ziel), fall.restRumpf(ziel))
				},
				fenster: func() *app.Fehler {
					return fall.fenster(fensterAus(t, umgebung, filepath.Join(ziel, "w.mse-set")), ziel)
				},
			})
		})
	}
}

// **Und beide sehen denselben Bestand.**
//
// Das ist die Zusage hinter allem: was die eine schreibt, findet die andere.
// Zwei Flächen auf zwei Datenbanken wären zwei Anwendungen.
func TestBeideFlaechenSehenDenselbenBestand(t *testing.T) {
	umgebung := eigenerBestand(t)

	// Über die Kommandozeile einfahren …
	aus, fehlerstrom, code := kommandozeile(t, umgebung, "oeffnen", probeset(t), "--json")
	if code != 0 {
		t.Fatalf("einfahren: %d (%s)", code, fehlerstrom)
	}

	setID := setkennungAus(t, aus)

	// … und über MCP wiederfinden.
	text, istFehler := mcpRuf(t, umgebung, "lesen.sets", map[string]any{})
	if istFehler {
		t.Fatalf("lesen.sets: %s", text)
	}

	// **Über den Namen und nicht über die Kennung.** `lesen.sets` gibt heute
	// die Feldnamen der Domänenentität heraus (`"ID"`, `"Name"`), die
	// Kommandozeile ihre eigenen (`"id"`, `"name"`) — ein Befund, der beim
	// Schreiben dieser Abnahme aufgefallen ist und in
	// #985
	// steht. Diese Abnahme prüft die **Fachlichkeit**, nicht die Gestalt: dass
	// beide Flächen denselben Bestand sehen.
	if !strings.Contains(text, "Mein Probeset") {
		t.Errorf("MCP kennt das eingefahrene Set nicht (%s):\n%s", setID, text)
	}
}

// **Ein zweites Einfahren legt an beiden Flächen kein zweites Set an.**
//
// Die Regel steht im Fachkern (#886); hier wird sie **an beiden Flächen**
// gefahren — denn genau dort ist sie einmal auseinandergelaufen.
func TestEinZweitesEinfahrenLegtAnBeidenFlaechenNichtsAn(t *testing.T) {
	umgebung := eigenerBestand(t)
	pfad := probeset(t)

	aus, _, code := kommandozeile(t, umgebung, "oeffnen", pfad, "--json")
	if code != 0 {
		t.Fatalf("erstes Einfahren: %d", code)
	}

	erst := setkennungAus(t, aus)

	text, istFehler := mcpRuf(t, umgebung, "schreiben.set.einfahren",
		map[string]any{"pfad": pfad})
	if istFehler {
		t.Fatalf("zweites Einfahren über MCP: %s", text)
	}

	if zweit := setkennungAus(t, text); zweit != erst {
		t.Errorf("MCP legte Set %s an, die Kommandozeile hatte %s", zweit, erst)
	}

	// Und im Bestand steht genau eines.
	aus, _, _ = kommandozeile(t, umgebung, "sets")
	if zeilen := strings.Count(strings.TrimSpace(aus), "\n") + 1; zeilen != 1 {
		t.Errorf("%d Sets im Bestand:\n%s", zeilen, aus)
	}
}
