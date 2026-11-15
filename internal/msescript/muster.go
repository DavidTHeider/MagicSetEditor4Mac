package msescript

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/dlclark/regexp2"
)

// Muster ist ein Muster von MSE — an **einer** Stelle übersetzt.
//
// # Welcher Dialekt
//
// MSE benutzt `boost::regex` in der Perl-Fassung (`src/util/regex.hpp`, Zeile
// 27 von twanvl/MagicSetEditor2, Stand 2026-08-16). Gos `regexp` ist RE2 und
// kennt weder Vorausblick noch Rückblick noch Rückverweise — dafür garantiert
// es lineare Laufzeit.
//
// Ausgezählt über alle `match:`-Angaben von `M15-Magic-Pack`:
//
// | | |
// | --- | ---: |
// | verschiedene Muster | 1645 |
// | von RE2 getragen | 1630 |
// | brauchen Vorausblick, Rückblick, Rückverweis oder `(?x)` | 15 |
//
// # Zwei Maschinen, und warum nicht eine
//
// **RE2 zuerst, `regexp2` als Rückfalllinie.** Die 1.630 gehen den schnellen
// Weg; die fünfzehn den anderen.
//
// Eine einzige Maschine wäre sauberer, und beide Wahlen kosten etwas:
//
//   - **Nur RE2**: fünfzehn Muster scheitern. Eines davon steht in
//     `magic.mse-game` an zwölf Zeilen, die **jeder** der 190 Stile ausführt —
//     gemessen 2.280 Befunde aus einem Prozent der Muster.
//   - **Nur `regexp2`**: Rückverfolgung ohne Laufzeitgrenze. `match` wird beim
//     Deuten einer einzigen Karte 616-mal gerufen, und ein bösartiges Muster in
//     einem **bezogenen** Paket hinge die Anwendung auf. Fremde Daten sind hier
//     der Normalfall (§8), und ein Paket von 2007 ist nicht auf uns geprüft.
//
// Die geteilte Lösung gibt beides: die Regel bleibt für 99 Prozent linear, und
// die fünfzehn laufen überhaupt.
//
// **Was das nicht ist:** eine Zusage, dass ein übersetztes Muster genau dasselbe
// trifft wie in MSE. RE2, .NET und Perl decken sich in den ausgezählten Formen;
// ein Unterschied bei einem exotischen Muster wäre hier nicht zu sehen.
type Muster struct {
	// quelle ist das Muster, wie es in der Vorlage stand.
	//
	// **Nur für die Meldung.**
	quelle string

	// schnell ist RE2 — gesetzt, wenn das Muster damit übersetzbar war.
	schnell *regexp.Regexp

	// breit ist die Rückfalllinie mit Rückverfolgung.
	breit *regexp2.Regexp

	// frist ist die Zeitgrenze, mit der [breit] übersetzt wurde — damit
	// [musterspeicher] ein Muster nicht unter einer anderen Frist hergibt.
	frist time.Duration

	// kontext ist `in_context:` — ein Treffer zählt nur, wo er passt
	// ([Muster.MitKontext]). Leer heißt: jeder Treffer zählt.
	kontext *Muster
}

// ErrMuster meldet ein Muster, das keine der beiden Maschinen liest.
var ErrMuster = errors.New("kein brauchbares Muster")

// musterFrist begrenzt die Rückverfolgung.
//
// **Eine Sekunde ist großzügig und trotzdem eine Grenze.** Kein Muster der
// bezogenen Pakete braucht mehr als Millisekunden; die Frist ist nicht für sie
// da, sondern für das Paket, das jemand morgen bezieht.
//
// **Eine Veränderliche und keine Konstante**, damit `muster_intern_test.go` sie
// kurz herabsetzen kann. Der Abbruch ist ein Fehlerweg wie jeder andere und
// gehört betreten (§3) — mit der echten Sekunde dauerte jeder Lauf der
// Pipeline eine Sekunde länger, und zwar für jeden der fünf Aufrufer.
//
// Sie wird sonst nirgends geschrieben; wer das täte, änderte eine Zusage über
// fremde Daten an einer Stelle, an der niemand sie sucht.
var musterFrist = time.Second

// UebersetzeMuster liest ein Muster in der Fassung, die MSE-Vorlagen schreiben.
//
// # Mehrzeilig, weil MSE es ist
//
// `^` trifft an **jedem** Zeilenanfang, `$` vor **jedem** Umbruch. MSE
// benutzt `boost::regex` (`src/util/regex.cpp`, Zeile 20, Fassung f9d9356 vom
// 2021-07-06) und übergibt keine Match-Flags — weder bei `regex_search`
// (`regex.hpp`, Zeilen 74 und 80) noch bei `regex_replace` (`regex.cpp`,
// Zeile 28; `boost::format_sed` ist eine **Format**-Flagge, keine
// Match-Flagge). Es gilt also `match_default`, und darin ist
// `match_single_line` **nicht** gesetzt — genau die Flagge, mit der man das
// Verhalten abschaltete.
//
// **Was daran hing:** `magic.mse-game/script`, Zeile 1943, klammert jede
// Zeile, die mit „• " beginnt, in `<li><bullet>…</bullet>…</li>`. Ohne
// Mehrzeiligkeit traf die Regel nur, wenn der **ganze** Regeltext so anfing —
// und das tut er nie, denn davor steht „Choose one —". Über `Full-Magic-Pack`
// tragen **86 von 1.903** verschiedenen `match:`-Angaben ein `^` oder `$`;
// gezählt an der Angabe in der Datei, nicht an Auswertungen zur Laufzeit (§10)
// (#435).
//
// `\A` und `\z` bleiben davon unberührt — sie meinen den Text, nicht die
// Zeile.
func UebersetzeMuster(quelle string) (*Muster, error) {
	frist := musterFrist

	if gemerkt, da := musterspeicher.hole(quelle, frist); da {
		return gemerkt, nil
	}

	muster, err := uebersetzeMuster(quelle, frist)
	if err != nil {
		return nil, err
	}

	musterspeicher.merke(quelle, muster)

	return muster, nil
}

// uebersetzeMuster übersetzt, ohne zu merken.
func uebersetzeMuster(quelle string, frist time.Duration) (*Muster, error) {
	gesetzt := ohneKollationselemente(quelle)

	if schnell, err := regexp.Compile("(?m)" + gesetzt); err == nil {
		return &Muster{quelle: quelle, schnell: schnell}, nil
	}

	ausgeschrieben, err := posixKlassenAusgeschrieben(gesetzt)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrMuster, quelle, err)
	}

	breit, err := regexp2.Compile(ausgeschrieben, regexp2.Multiline)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrMuster, quelle, err)
	}

	breit.MatchTimeout = frist

	return &Muster{quelle: quelle, breit: breit, frist: frist}, nil
}

// ErrLauf meldet, dass die Rückfalllinie ihren Lauf abgebrochen hat.
//
// **Fast immer die Zeitgrenze.** Ein Muster mit Rückverfolgung kann an einer
// ungünstigen Eingabe beliebig lange rechnen; [musterFrist] bricht dann ab.
//
// # Warum das einen eigenen Kanal braucht
//
// Bis zum Zweitblick über die Portierung gaben diese Verfahren bei einem
// Abbruch schlicht „kein Treffer" oder einen **halben** Stand zurück — der Text
// war dann teils ersetzt und sah aus wie ein Ergebnis. Das ist der stille Weg,
// den §8 verbietet: ein Set, von dem etwas nicht lesbar ist, wird mit einem
// benannten Befund geöffnet, nicht mit einer falschen Antwort.
var ErrLauf = errors.New("der Mustervergleich wurde abgebrochen")

// Passt sagt, ob das Muster irgendwo trifft.
func (m *Muster) Passt(text string) (bool, error) {
	if m.schnell != nil {
		return m.schnell.MatchString(text), nil
	}

	trifft, err := m.breit.MatchString(text)
	if err != nil {
		return false, fmt.Errorf("%w: %s: %w", ErrLauf, m.quelle, err)
	}

	return trifft, nil
}

// AlleTreffer gibt jeden Treffer als Text — für `filter_text` und
// `break_text`, **je Treffer neu ab seinem Ende** ([Muster.jederTrefferMitNeustart]).
func (m *Muster) AlleTreffer(text string) ([]string, error) {
	gefunden := make([]string, 0, 4)

	err := m.jederTrefferMitNeustart(text, func(_, _ int, gruppen []string) {
		gefunden = append(gefunden, gruppen[0])
	})

	return gefunden, err
}

// Teile zerlegt einen Text an jedem Treffer — für `split_text`, **je Treffer
// neu ab seinem Ende** wie dort (`regex.cpp`, Zeilen 212–232).
func (m *Muster) Teile(text string) ([]string, error) {
	zeichen := []rune(text)
	teile := make([]string, 0, 4)
	gelesen := 0

	err := m.jederTrefferMitNeustart(text, func(index, laenge int, _ []string) {
		teile = append(teile, string(zeichen[gelesen:index]))
		gelesen = index + laenge
	})
	if err != nil {
		return nil, err
	}

	return append(teile, string(zeichen[gelesen:])), nil
}

// jederTreffer läuft über alle Treffer der Rückfalllinie.
func (m *Muster) jederTreffer(text string, tue func(*regexp2.Match)) error {
	treffer, err := m.breit.FindStringMatch(text)

	for err == nil && treffer != nil {
		tue(treffer)

		treffer, err = m.breit.FindNextMatch(treffer)
	}

	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrLauf, m.quelle, err)
	}

	return nil
}
