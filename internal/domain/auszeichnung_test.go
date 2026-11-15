// **Was das Ersetzen mit den Auszeichnungen tut** — die Läufe über
// `internal/domain/auszeichnung.go`.
//
// # Warum eine eigene Datei
//
// `ersetzung_test.go` ging über die Zeilengrenze (§2), als der Nachlauf über
// Phase 3.7 die überspannten Marken verlangte. Der Schnitt folgt der Frage:
// dort steht, **was** getroffen und ersetzt wird; hier, **was dabei mit den
// Marken geschieht**.
//
// **Diese Datei ist die teuerste Lehre der Phase.** Dieselbe Klasse ist
// dreimal aufgetreten und zweimal nur halb geschlossen worden — erst am
// Zeichen, das beim Kleinschreiben kürzer wird, dann an der Marke hinter dem
// Treffer, zuletzt an der Marke **im** Treffer.
package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Zeichen, das beim Kleinschreiben kürzer wird** — Befund 9 des
// Zweitblicks über Phase 3.7
// (#1168).
//
// # Die Klasse, nicht der Fall
//
// Gesucht wird über **Bytes**, und die Zuordnung auf den rohen Wert entsteht
// beim Entmarken. Der erste Anlauf schrieb den Text **danach** klein
// (`strings.ToLower`) und benutzte die dort gefundenen Bytestellen weiter für
// die Grenzen des Originals. Das trägt nur, solange die Kleinschreibung die
// Bytelänge erhält.
//
// **In Unicode tut sie das nicht immer.** Go bildet `İ` (U+0130, zwei Bytes)
// auf `i` (ein Byte) ab und `K` (U+212A, drei Bytes) auf `k` (ein Byte). Jede
// Stelle **hinter** einem solchen Zeichen lief damit aus dem Takt.
//
// Die vorhandenen Läufe konnten es nicht sehen: sie prüfen `ß` vor einem
// späteren ASCII-Treffer — `ß` bleibt beim Kleinschreiben `ß`.
func TestErsetzungTrifftAuchWennDieKleinschreibungKuerzt(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name, roh, suche, ersatz, erwartet string
	}{
		{"das Zeichen selbst", "İ", "i", "X", "X"},
		{"und das Kelvinzeichen", "K", "k", "X", "X"},
		{"davor bleibt stehen", "aİb", "b", "X", "aİX"},
		{"und eine Auszeichnung dahinter", "İ<i>x</i>", "i", "X", "X<i>x</i>"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			frage := domain.Ersetzungsfrage{Text: fall.suche, Ersatz: fall.ersatz}

			if got := frage.Ersetze(fall.roh); got != fall.erwartet {
				t.Errorf("Ersetze(%q) = %q, erwartet %q", fall.roh, got, fall.erwartet)
			}
		})
	}
}

// **Eine Auszeichnung hinter dem Treffer bleibt stehen** — die Klasse hinter
// Befund 9 des Zweitblicks über Phase 3.7.
//
// # Warum das kein Unicode-Fall ist
//
// Der Befund nannte ein Zeichen, das beim Kleinschreiben kürzer wird. Die
// Klasse ist größer: das **Ende** einer Fundstelle war der Anfang des
// **nächsten** übernommenen Zeichens, und dazwischen können Auszeichnungen
// stehen. `Ersetze("a<i>b</i>")` mit `a → X` gab `"Xb</i>"` — das öffnende
// `<i>` war weg, bei reinem ASCII.
//
// **Der Fall am Ende des Wertes war gesehen** (`hinterDemLetzten`); jeder
// Fall in der Mitte war es nicht.
func TestEineAuszeichnungHinterDemTrefferBleibtStehen(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name, roh, suche, erwartet string
	}{
		{"öffnende Marke dahinter", "a<i>b</i>", "a", "X<i>b</i>"},
		{"schließende Marke dahinter", "<i>a</i>b", "a", "<i>X</i>b"},
		{"zwei Marken dazwischen", "a<i></b>c", "a", "X<i></b>c"},
		{"und mitten im Wort", "ab<i>c</i>", "b", "aX<i>c</i>"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			frage := domain.Ersetzungsfrage{Text: fall.suche, Ersatz: "X"}

			if got := frage.Ersetze(fall.roh); got != fall.erwartet {
				t.Errorf("Ersetze(%q) = %q, erwartet %q", fall.roh, got, fall.erwartet)
			}
		})
	}
}

// **Ein Treffer, der über eine Auszeichnung hinweggeht** — Befund 1 des
// Nachlaufs über Phase 3.7
// (#1168).
//
// # Warum die Endpunkte allein nicht reichen
//
// Die zwei Grenzlisten haben die Marke **hinter** einem Treffer gerettet. Geht
// der Treffer aber **über** eine Marke hinweg, schneidet der Ersatz sie
// mitsamt heraus — und was übrig bleibt, ist unausgeglichen:
//
//	Ersetze("a<i>b</i>", "ab" → "")  gab  "</i>"
//	Ersetze("<i>a</i>b", "ab" → "")  gab  "<i>"
//
// **MSE2 gleicht aus** (`tagged_substr_replace`,
// `src/util/tagged_string.cpp`, Zeilen 622–632, Fassung `f9d9356` vom
// 2021-07-06): was im geschnittenen Bereich offen blieb, wird davor
// geschlossen und dahinter wieder geöffnet — und danach vereinfacht, sodass
// ein leeres Paar verschwindet.
func TestEinTrefferUeberEineAuszeichnungHinweg(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name, roh, suche, ersatz, erwartet string
	}{
		{"öffnende Marke im Treffer", "a<i>b</i>", "ab", "", ""},
		{"schließende Marke im Treffer", "<i>a</i>b", "ab", "", ""},
		{"und mit einem Ersatz", "a<i>b</i>", "ab", "X", "X"},
		{
			"was drumherum steht, bleibt",
			"vor a<i>b</i> nach", "ab", "X", "vor X nach",
		},
		{
			"eine geschachtelte Marke",
			"a<i><b>c</b></i>", "ac", "X", "X",
		},
		{
			"eine Marke, die weiterläuft",
			"a<i>bc</i>", "ab", "X", "X<i>c</i>",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			frage := domain.Ersetzungsfrage{Text: fall.suche, Ersatz: fall.ersatz}

			if got := frage.Ersetze(fall.roh); got != fall.erwartet {
				t.Errorf("Ersetze(%q) = %q, erwartet %q", fall.roh, got, fall.erwartet)
			}
		})
	}
}

// **Die Randfälle des Markenausgleichs** — Befund 1 des Nachlaufs über
// Phase 3.7.
//
// # Warum sie einzeln stehen
//
// [domain.Markenausgleich] ist die eine Stelle, an der ein ausgezeichneter
// Wert beim Ersetzen ausgeglichen wird. Was sie **nicht** trifft, wäre ein
// Wert, der nach dem Schreiben nicht mehr zu lesen ist — und `msepaket`
// schreibt ihn dann so in die Datei.
func TestDerMarkenausgleichKenntSeineRaender(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name, roh                     string
		anfang, ende                  int
		erwartetDavor, erwartetDanach string
	}{
		{"ohne Marke bleibt beides leer", "abc", 0, 3, "", ""},
		{"ein Paar im Bereich hebt sich auf", "a<i>b</i>c", 0, 10, "", ""},
		{"eine offene bleibt offen", "a<i>b", 0, 5, "", "<i>"},
		{"eine schließende bleibt schließend", "a</i>b", 0, 6, "</i>", ""},
		{
			"zwei geschachtelte offene",
			"<i><b>", 0, 6, "", "<i><b>",
		},
		{
			"zwei schließende, von innen nach außen",
			"</b></i>", 0, 8, "</i></b>", "",
		},
		{
			"eine Klammer ohne Ende zählt nicht",
			"a<i", 0, 3, "", "",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			davor, danach := domain.Markenausgleich(fall.roh, fall.anfang, fall.ende)

			if davor != fall.erwartetDavor || danach != fall.erwartetDanach {
				t.Errorf("Markenausgleich(%q) = (%q, %q), erwartet (%q, %q)",
					fall.roh, davor, danach, fall.erwartetDavor, fall.erwartetDanach)
			}
		})
	}
}

// **Ein geschütztes Kleinerzeichen ist keine Marke** — MSEs `ESCAPED_LANGLE`.
//
// Ein `\x01` im Text sähe für einen naiven Markensucher wie ein Anfang aus;
// er ist keiner, und der Ausgleich darf ihn nicht als offene Marke zählen.
func TestEinGeschuetztesKleinerzeichenIstKeineMarke(t *testing.T) {
	t.Parallel()

	frage := domain.Ersetzungsfrage{Text: "b", Ersatz: "X"}

	if got := frage.Ersetze("\x01b"); got != "\x01X" {
		t.Errorf("Ersetze(%q) = %q, erwartet %q", "\x01b", got, "\x01X")
	}
}
