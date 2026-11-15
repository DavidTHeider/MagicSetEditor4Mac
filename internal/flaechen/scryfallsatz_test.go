// **Dieselbe Fähigkeit, überall dieselbe Antwort — für Scryfall**
// (#1028).
//
// # Was diese Datei prüft
//
// §16 endet mit:
//
//	Ein Wächter zählt es. Was er nicht hält, steht neben ihm: er sieht, dass
//	eine Fähigkeit vorhanden ist, nicht dass sie dasselbe tut. Dafür gibt es
//	je Fähigkeit einen Lauf, der denselben Arbeitsablauf über jede Fläche
//	fährt und dieselben fachlichen Fehler erwartet.
//
// Für die Scryfall-Fähigkeiten gab es diesen Lauf nicht — `grep -rn scryfall
// internal/flaechen/` fand am 2026-08-24 keine Fundstelle. Der Fähigkeitszähler
// hält, dass in jeder Spalte ein Name steht; **keiner hielt, dass hinter den
// Namen dieselbe Antwort steht.**
//
// Dass das kein Papiertiger ist, hat Phase V selbst gezeigt: MCP wies einen
// leeren Manakosten nicht ab, während CLI und REST es taten (#1017). Gefunden
// wurde das durch einen Abdeckungslauf, nicht durch einen Wächter.
//
// # Was diese Läufe **nicht** halten
//
//   - **Nicht, dass die Antworten gleich aussehen.** Sie sehen verschieden
//     aus, und das ist richtig: die Kommandozeile schreibt für Menschen, MCP
//     für ein Sprachmodell, REST JSON mit Statuscode. Geprüft ist die
//     **Wirkung** im Bestand und die **Klasse** des Fehlers.
//   - **Das Fenster nur, wo es dasteht.** Hier stand, es sei keine vierte
//     Fläche — seit C04 ruft `fensterAus` seine Bindungen über
//     `NeueBindungen`, mit Toren und Rufkontext; welche Flächen ein Lauf
//     anspricht, misst `internal/werkzeuge/flaechenparitaetflaechen_test.go`
//     und steht in `docs/FLAECHENPARITAET.md`.
//   - **Nicht den echten Dienst.** Die Gegenstelle steht in
//     `scryfallgegenstelle_test.go`; ob Scryfall wirklich so antwortet, sagen
//     die Läufe hinter dem Tag `netzlauf`.
package flaechen_test

import (
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Einen ganzen Satz einlesen — über jede Fläche.**
func TestEinenSatzEinlesenUeberJedeFlaeche(t *testing.T) {
	var gefragt atomic.Int64

	gegenstelle := gefragteGegenstelle(t, &gefragt)

	// **Und am Ende steht, dass die Gegenstelle wirklich gefragt wurde.**
	// Ohne diese Zusicherung ließe sich „kein Aufruf ging ans echte Netz" nur
	// behaupten: eine Umlenkung, die nicht ankommt, schickte jede Fläche zu
	// Scryfall, und der Lauf sagte nicht, welches von beidem er geprüft hat.
	t.Cleanup(func() {
		if gefragt.Load() == 0 {
			t.Error("die Gegenstelle wurde nie gefragt — die Umlenkung über " +
				"MSE_SCRYFALL_ADRESSE hat nicht gegriffen, und die Aufrufe " +
				"gingen ans echte Netz")
		}
	})

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)

		aus, fehler, code := kommandozeile(t, umgebung, "scryfall-set", "prf", "magic", "--json")
		if code != 0 {
			t.Fatalf("scryfall-set: %d %s%s", code, aus, fehler)
		}

		satzImBestandPruefen(t, umgebung)
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)

		text, istFehler := mcpRuf(t, umgebung, "schreiben.scryfall.set",
			map[string]any{"code": "prf", "spiel": "magic"})
		if istFehler {
			t.Fatalf("schreiben.scryfall.set: %s", text)
		}

		satzImBestandPruefen(t, umgebung)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/sets/scryfall",
			`{"code":"prf","spiel":"magic"}`)
		if status != http.StatusCreated {
			t.Fatalf("POST /sets/scryfall: %d (%s)", status, rumpf)
		}

		satzImBestandPruefen(t, umgebung)
	})
}

// **Und derselbe fachliche Fehler an jeder Fläche.**
//
// # Warum das der eigentliche Lauf ist
//
// Dass eine Fähigkeit **da** ist, sagt der Fähigkeitszähler. Was er nicht
// sagt, ist, ob sie dasselbe **abweist**. Ein Satzcode ohne Treffer ist der
// Fall, den jede Fläche gleich beantworten muss — sonst hängt es am Weg, ob
// ein Fehler ein Fehler ist.
//
// **Über `abweisungGleich`** seit dem 2026-08-28 — Befund 3 des Nachlaufs über
// Phase 3.8 (#1253).
// Hier stand dreimal „irgendein Fehler und kein angelegtes Set".
func TestEinUnbekannterSatzScheitertUeberallGleich(t *testing.T) {
	gegenstelle := scryfallgegenstelle(t)
	umgebung := scryfallumgebung(t, gegenstelle)
	an := starteDienst(t, umgebung)

	abweisungGleich(t, "ein Satzcode, den Scryfall nicht kennt", abweisung{
		klasse: klasseNichtGefunden,
		kern:   "Scryfall kennt die Karten dieses Satzes nicht",
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).ScryfallSatzEinlesen(t.Context(), "gibtsnicht", "magic").Fehler
		},
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"scryfall-set", "gibtsnicht", "magic")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.scryfall.set",
				map[string]any{"code": "gibtsnicht", "spiel": "magic"})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/sets/scryfall",
				`{"code":"gibtsnicht","spiel":"magic"}`)
		},
	})

	// **Und keine Fläche hat dabei ein Set angelegt.** Das ist die zweite
	// Hälfte der Zusage: ein gescheiterter Abruf hinterlässt keinen halben
	// Stand (§8).
	keinSetImBestand(t, umgebung)
}

// **Ein unbekanntes Spiel ebenso** — und das ist der Fall, an dem #1017 hing.
//
// Dort wies MCP einen leeren Manakosten nicht ab, während CLI und REST es
// taten. Hier ist die Frage dieselbe in klein: eine Angabe, die keine
// Zuordnung hat, muss an jeder Fläche eine Abweisung sein.
func TestEinUnbekanntesSpielScheitertUeberallGleich(t *testing.T) {
	gegenstelle := scryfallgegenstelle(t)
	umgebung := scryfallumgebung(t, gegenstelle)
	an := starteDienst(t, umgebung)

	abweisungGleich(t, "ein Spiel, für das es keine Zuordnung gibt", abweisung{
		klasse: klasseValidierung,
		kern:   "gibtsnicht",
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).ScryfallSatzEinlesen(t.Context(), "prf", "gibtsnicht").Fehler
		},
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"scryfall-set", "prf", "gibtsnicht")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.scryfall.set",
				map[string]any{"code": "prf", "spiel": "gibtsnicht"})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/sets/scryfall",
				`{"code":"prf","spiel":"gibtsnicht"}`)
		},
	})
}

// satzImBestandPruefen sieht nach, was wirklich dasteht.
//
// # Warum über die Kommandozeile und nicht über die Antwort
//
// **Die Antworten sehen an jeder Fläche anders aus, und das ist richtig**: die
// Kommandozeile schreibt für Menschen, MCP führt seine Felder groß
// ([0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)),
// REST schreibt klein. Wer die Antworten verglich, prüfte die Schreibweise
// statt der Fähigkeit.
//
// **Geprüft ist deshalb die Wirkung** — ein Beobachter für alle drei. Dass
// gerade die Kommandozeile ihn stellt, ist eine Wahl und keine Aussage: sie
// ist die Fläche, die ohne laufenden Dienst und ohne Anmeldung auskommt.
func satzImBestandPruefen(t *testing.T, umgebung []string) {
	t.Helper()

	sets := setzeilenAus(t, mussKommandozeile(t, umgebung, "sets", "--json"))
	if len(sets) != 1 {
		t.Fatalf("%d Sets im Bestand, erwartet eines", len(sets))
	}

	if sets[0].Name != "Prüfsatz" {
		t.Errorf("das Set heißt %q, erwartet »Prüfsatz« aus den Satzangaben", sets[0].Name)
	}

	karten := kartenzeilenAus(t,
		mussKommandozeile(t, umgebung, "karten", sets[0].ID, "--json"))
	if len(karten) != 2 {
		t.Fatalf("%d Karten im Set, erwartet 2", len(karten))
	}
}

// keinSetImBestand hält fest, dass ein gescheiterter Aufruf nichts hinterließ.
func keinSetImBestand(t *testing.T, umgebung []string) {
	t.Helper()

	if sets := setzeilenAus(t, mussKommandozeile(t, umgebung, "sets", "--json")); len(sets) != 0 {
		t.Errorf("%d Sets im Bestand, erwartet keines", len(sets))
	}
}
