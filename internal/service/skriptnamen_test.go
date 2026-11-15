// **Was die Umgebung kennt, kann sie auch aufzählen** — Werkstück 1 von
// Phase 3.3 (#1164).
//
// # Warum die Liste aus der Umgebung kommt und nicht daneben steht
//
// Eine gepflegte Aufzählung liefe an dem Tag auseinander, an dem die
// Bibliothek eine Funktion dazubekommt — und die Vervollständigung böte etwas
// an, das nicht läuft. Dieser Lauf prüft deshalb nicht **welche** Namen es
// sind, sondern **dass sie dieselbe Quelle haben wie das Auswerten**.
package service_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Namen sind die der Bibliothek — sortiert und ohne Doppel.**
func TestDieNamenKommenAusDerUmgebung(t *testing.T) {
	t.Parallel()

	konsole := service.NeueSkriptkonsole(
		konsolenkarten{karte: &domain.CardRow{ID: 7}}, nil)

	namen, fehler := konsole.Namen(t.Context(), 7)
	if fehler != nil {
		t.Fatalf("Namen: %v", fehler)
	}

	if !slices.IsSorted(namen) {
		t.Errorf("die Namen sind unsortiert: %v", namen)
	}

	// **Die Probe ist ein Aufruf, keine Zeile in einer Liste.** Ein Name,
	// den die Konsole nennt, muss auch laufen — sonst bietet die
	// Vervollständigung etwas an, das es nicht gibt.
	for _, name := range namen {
		if _, fehler := konsole.Auswerten(t.Context(), 7, name); fehler != nil {
			t.Errorf("%q steht in der Liste, läuft aber nicht: %v", name, fehler)
		}
	}

	if len(namen) < 2 {
		t.Fatalf("es kamen nur %d Namen: %v", len(namen), namen)
	}
}

// **Die Wörter des Zerteilers stehen nicht darin.**
//
// Die Gegenprobe zum Lauf darüber: `if` ist ein **Schlüsselwort** und kein
// Name der Umgebung. Wer es in die Vervollständigung nähme, behauptete, sie
// kenne etwas, das sie nicht kennt — und der Aufruf oben fiele darüber.
func TestDieSchluesselwoerterStehenNichtInDerListe(t *testing.T) {
	t.Parallel()

	konsole := service.NeueSkriptkonsole(
		konsolenkarten{karte: &domain.CardRow{ID: 7}}, nil)

	namen, fehler := konsole.Namen(t.Context(), 7)
	if fehler != nil {
		t.Fatalf("Namen: %v", fehler)
	}

	for _, wort := range []string{"if", "then", "else", "and", "or", "not"} {
		if slices.Contains(namen, wort) {
			t.Errorf("%q steht in der Liste der Namen", wort)
		}
	}
}

// **Eine Karte, die es nicht gibt, hat keine Namen.**
//
// Nicht eine leere Liste: das wäre die Aussage „die Umgebung kennt nichts",
// und die stimmt nicht — es gibt gar keine Umgebung.
func TestOhneKarteGibtEsKeineNamen(t *testing.T) {
	t.Parallel()

	konsole := service.NeueSkriptkonsole(
		konsolenkarten{fehler: domain.ErrNotFound}, nil)

	namen, fehler := konsole.Namen(t.Context(), 7)
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("der Fehler ist %v — will ErrNotFound", fehler)
	}

	if namen != nil {
		t.Errorf("es kam trotzdem eine Liste: %v", namen)
	}
}

// **Mit Stilpaket kommen `card` und `set` dazu.**
//
// **Das ist die unterscheidende Zusicherung dieser Datei.** Ohne sie sagten
// die Läufe darüber auch dann ja, wenn `Namen` eine feste Aufzählung der
// Bibliothek zurückgäbe — die Liste folgt der Umgebung, und hier ist die
// Umgebung eine andere.
func TestMitStilpaketStehenKarteUndSetInDerListe(t *testing.T) {
	t.Parallel()

	bauer := service.NeuerPlanbauer().
		MitStilen(festerStil{vorlage: service.Vorlage{Stil: saatstil(t)}, da: true})

	konsole := service.NeueSkriptkonsole(
		konsolenkarten{karte: &domain.CardRow{ID: 7}}, bauer)

	namen, fehler := konsole.Namen(t.Context(), 7)
	if fehler != nil {
		t.Fatalf("Namen: %v", fehler)
	}

	for _, name := range []string{"card", "set"} {
		if !slices.Contains(namen, name) {
			t.Errorf("%q fehlt in %v", name, namen)
		}
	}
}
