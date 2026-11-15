// Wohin eine Antwort geht — und was, wenn das Schreiben scheitert.
//
// # Warum das eine eigene Frage ist
//
// `mse sets --json | head -1` schließt die Röhre, sobald `head` genug hat.
// Jedes weitere `Fprintln` scheitert dann. Ein Programm, das den Fehler
// wegwirft und trotzdem 0 zurückgibt, sagt einem Skript, es habe alles
// ausgegeben — und das stimmt nicht.
//
// **Deshalb wird er gemerkt und einmal geprüft**, statt an jedem `Fprintln`.
// Sechzehn `if`-Zweige um eine Ausgabe herum prüfte kein Test einzeln, und
// jeder wäre eine Behauptung (§3).

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// schreiber ist, wohin eine Antwort geht — und in welcher Form.
type schreiber struct {
	ziel    io.Writer
	fehler  io.Writer
	alsJSON bool

	// eingabe ist die Standardeingabe — **ohne Fenster die Zwischenablage**
	// (#1168).
	//
	// **`nil` ist erlaubt** und heißt „nichts": jeder Befehl außer
	// `ablage-einfuegen` liest gar nicht, und ein Test, der sie nicht setzt,
	// soll nicht am Terminal des Prüfenden hängen.
	eingabe io.Reader

	// erwartet ist `--erwartet=<json>` — die Erwartung für `felder-setzen`
	// (U6, #1417); leer heißt: ohne Vergleich.
	erwartet string

	// vorgang ist die Kennung dieses Laufs — für den Fehler im JSON (F03).
	vorgang domain.Vorgangskennung

	// erster ist der erste Schreibfehler, den es gab. **Der erste und nicht
	// der letzte**: die folgenden sind seine Folge.
	erster error
}

// sag schreibt eine Zeile auf das Ziel.
func (s *schreiber) sag(format string, werte ...any) {
	if _, err := fmt.Fprintf(s.ziel, format, werte...); err != nil && s.erster == nil {
		s.erster = err
	}
}

// klage schreibt eine Zeile auf den Fehlerstrom.
//
// **Ein Fehler beim Klagen wird nicht gemerkt.** Wer sich darüber beschweren
// wollte, bräuchte einen dritten Strom.
//
// **Und eine Zeile endet mit ihrem Zeilenende** (ENG-05.AK-03, #1730): 35
// Stellen gaben eine Meldung aus Go ohne eines aus, und die Vorgangskennung
// danach klebte an ihr — `…liegt schon daVorgang: dad7dc7a`. Hier, an der
// einen Stelle, statt an jeder der 35.
func (s *schreiber) klage(format string, werte ...any) {
	zeile := fmt.Sprintf(format, werte...)
	if !strings.HasSuffix(zeile, "\n") {
		zeile += "\n"
	}

	_, _ = io.WriteString(s.fehler, zeile) //nolint:errcheck // siehe darüber
}

// abschluss gibt den Rückgabewert einer geglückten Arbeit — und meldet einen
// Schreibfehler, falls es einen gab.
//
// **Ohne Parameter**, weil er nur auf dem geglückten Weg gerufen wird: ein
// fachlicher Fehler ist längst zurückgegeben, wenn es hierher kommt. Der erste
// Anlauf nahm den bisherigen Wert entgegen und bekam immer denselben — ein
// Parameter, der nie zwei Werte hat, ist eine Verzweigung, die keine ist.
func (s *schreiber) abschluss() int {
	if s.erster == nil {
		return CodeGut
	}

	s.klageUeber(fmt.Errorf("die Ausgabe ließ sich nicht vollständig schreiben: %w", s.erster))

	return CodeFachlich
}

// klageUeber meldet einen Fehler aus dem Unterbau.
//
// **Und nicht jeden gleich.** `%v` gibt die ganze Kette — bei einem belegten
// Bestand hieß das „Migration beginnen: der Bestand ist gerade belegt:
// database is locked (5) (SQLITE_BUSY)": vorn ein irreführender Zusammenhang,
// hinten eine rohe SQLite-Meldung, die §8 an keiner Oberfläche zulässt
// (#886).
//
// # Warum nicht app.Uebersetze — für den Satz
//
// Das gibt einen **Schlüssel**, und den Satz dazu halten die Kataloge im
// Frontend (§9). Eine Kommandozeile hat keinen Katalog; sie spricht Deutsch
// wie jede Meldung dieses Programms. **Für `--json` gerade doch** (F03,
// #2342): dort zählt der Code, nach dem ein Skript verzweigt, und der kommt
// von derselben Stelle wie im Fenster und im Dienst ([schreiber.fehlerAlsJSON]).
//
// # Was sie nicht tut
//
// **Sie deckt nicht jeden Fehler zu.** Nur der belegte Bestand hat einen
// eigenen Satz, weil nur er einen **Rat** hat, der nicht im Fehler steht:
// gleich nochmal. Alles andere bleibt, wie es kommt — eine Kommandozeile, die
// jede Ursache glättet, ist beim Nachsehen wertlos.
//
// **Ein fremder Dienst bekommt seinen Rat dazu** (D02, #2233), auf einer
// zweiten Zeile: der Fehler bleibt, wie er kommt — er nennt Fall, Wirt und
// Frist und keine Abfrage —, und darunter steht, was zu tun ist.
func (s *schreiber) klageUeber(fehler error) {
	if errors.Is(fehler, domain.ErrBelegt) {
		s.klageMitSatz(fehler, "%s — vermutlich hat das Fenster ihn offen "+
			"oder ein zweiter Befehl läuft. Bitte gleich noch einmal versuchen.\n",
			grossAmAnfang(domain.ErrBelegt.Error()))

		return
	}

	s.klage("%v\n", fehler)

	if rat := netzratFuer(fehler); rat != "" {
		s.klage("%s\n", rat)
	}

	if s.alsJSON {
		s.fehlerAlsJSON(fehler)
	}
}

// klageMitSatz meldet einen Fehler mit **eigenem Satz** — und mit `--json`
// trotzdem seine Form.
//
// **Für die Stellen, deren Satz mehr sagt als der Fehler**: der belegte
// Bestand mit seinem Rat, die bestätigte Sicherung, die nicht passt. Bis zum
// Zweitblick über Phase 1 von 0.9.0 (Befund 4) schrieben sie ihren Satz über
// `klage` und endeten mit Rückgabewert 1 — und mit `--json` stand kein Code
// da (`fehlerjsonweg_test.go` hält jetzt jeden Weg).
func (s *schreiber) klageMitSatz(fehler error, satz string, werte ...any) {
	s.klage(satz, werte...)

	if s.alsJSON {
		s.fehlerAlsJSON(fehler)
	}
}

// fehlerAlsJSON gibt einen Fehler **zusätzlich** als JSON aus — mit Code,
// Schlüssel, Werten und Vorgangskennung, in derselben Gestalt, die das
// Fenster und der Dienst bekommen (F03, #2342). Der Satz auf der
// Fehlerausgabe und der Rückgabewert bleiben, wie sie sind.
func (s *schreiber) fehlerAlsJSON(fehler error) {
	s.formAlsJSON(app.Uebersetze(app.MitVorgang(context.Background(), s.vorgang), fehler))
}

// klageUeberFehler meldet einen Fehler, den `internal/app` schon übersetzt
// hat: den Satz auf der Fehlerausgabe und mit `--json` **zusätzlich** seine
// Form (F03, #2342). Bis dahin gaben 53 Stellen nur `Fehler.Meldung` aus,
// und mit `--json` stand bei ihnen kein Code — ein Wächter hält den Weg
// (`fehlerjson_test.go`).
func (s *schreiber) klageUeberFehler(fehler *app.Fehler) {
	s.klage("%s\n", fehler.Meldung)

	if s.alsJSON {
		s.formAlsJSON(fehler)
	}
}

// formAlsJSON gibt die Form eines Fehlers als `{"fehler": …}` aus.
func (s *schreiber) formAlsJSON(fehler *app.Fehler) {
	roh, _ := json.Marshal(map[string]*app.Fehler{"fehler": fehler}) //nolint:errcheck // ein Fehler aus Zeichenketten lässt sich immer schreiben

	s.sag("%s\n", roh)
}

// grossAmAnfang schreibt den ersten Buchstaben groß.
//
// Die Fehlerwerte der Domäne beginnen klein, weil sie umschlossen werden
// (`fmt.Errorf("… : %w")`); am Satzanfang steht dann ein kleines Wort.
func grossAmAnfang(satz string) string {
	if satz == "" {
		return satz
	}

	zeichen := []rune(satz)

	return string(unicode.ToUpper(zeichen[0])) + string(zeichen[1:])
}
