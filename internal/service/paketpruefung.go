package service

import (
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Die Skripte einer Paketdatei prüfen** — Werkstück B von Phase 4.9
// (#1324).
//
// # Warum neben der Vorschau und nicht in ihr
//
// Die Vorschau zeichnet eine Karte, und ihr Fehler kann aus **jeder**
// beteiligten Datei stammen: dem Spielskript, einem eingebundenen Paket, der
// bearbeiteten Datei. Eine Zeile daraus im offenen Editor anzustreichen wäre
// eine Behauptung, die niemand geprüft hat — und ein Zeiger, der auf die
// falsche Zeile zeigt, ist schlechter als keiner.
//
// Diese Prüfung liest **nur den Text, der im Feld steht**. Ihre Stelle
// bezeichnet damit sicher eine Zeile dieser Datei.
//
// # Warum hier und nicht in einem der beiden Blätter
//
// Dieselbe Naht wie bei [Paketfarben], und aus demselben Grund:
// `msepaket.Einfaerben` **benennt** einen Skriptwert, ohne ihn zu verstehen;
// `msescript` versteht ihn, ohne vom Paketformat zu wissen. Beide zusammen
// kennt nur diese Schicht (§2).

// PaketSkriptfehler gibt den **ersten** Syntaxfehler in den Skriptwerten
// einer Paketdatei — mit seiner Stelle in **dieser** Datei.
//
// **Der erste und nicht alle.** Ein Fehler in einem Skriptwert lässt die
// übrigen unberührt, aber ein Editor streicht eine Stelle an, nicht zwölf;
// und wer den ersten behebt, sieht den nächsten. MSE meldet mehrere, hat
// dafür aber eine Liste unter dem Feld (`concat(errors)`,
// `src/util/error.cpp`, Zeilen 146–153, Fassung f9d9356 vom 2021-07-06) —
// die gibt es hier nicht.
//
// **Und die erste Feldart, die MSE2 nicht kennt** (FMT-02, S10.NAT-02,
// #1691) — gemeldet wird, was weiter oben steht: ein Editor streicht eine
// Stelle an, und die erste ist die, an der jemand weiterliest.
//
// Gibt `nil`, wenn die Datei keinen Skriptwert hat oder alle zerteilen und
// jede Feldart bekannt ist.
func PaketSkriptfehler(quelltext string) *msescript.Fehler {
	var erster *msescript.Fehler

	for _, merkmal := range msepaket.Einfaerben(quelltext) {
		if befund := merkmalfehler(merkmal); befund != nil {
			erster = befund

			break
		}
	}

	feldarten := msepaket.Feldartbefunde(quelltext)
	if len(feldarten) > 0 && (erster == nil || feldarten[0].Stelle.Zeile < erster.Zeile) {
		erster = feldartfehler(feldarten[0])
	}

	return erster
}

// **Die Schlüssel eines Feldartbefunds** — keine Meldungen des Zerteilers,
// deshalb nicht in `internal/msescript` und nicht im Katalogblock `skript`:
// das Fenster findet sie unter `seiten.vorlagen`. Der Typ ist trotzdem der
// des Zerteilers, weil die Prüfung **eine** Antwort gibt, gleich woher der
// Befund stammt.
const (
	SchluesselFeldartFehlt     msescript.Meldungsschluessel = "feldartFehlt"
	SchluesselFeldartUnbekannt msescript.Meldungsschluessel = "feldartUnbekannt"
)

// feldartfehler gibt einen Feldartbefund in der Gestalt eines Skriptfehlers —
// mit Schlüssel und Werten, damit das Fenster ihn in seiner Sprache zeigt.
func feldartfehler(befund msepaket.Feldartbefund) *msescript.Fehler {
	fehler := &msescript.Fehler{
		Zeile: befund.Stelle.Zeile, Spalte: befund.Stelle.Spalte, Meldung: befund.Meldung,
		Schluessel: SchluesselFeldartFehlt, Werte: []string{},
	}

	if befund.Art != "" {
		fehler.Schluessel = SchluesselFeldartUnbekannt
		fehler.Werte = []string{befund.Art, strings.Join(msepaket.BekannteFeldarten(), ", ")}
	}

	return fehler
}

// zerteilfehler zerteilt einen Skriptwert und versetzt seine Fehlerstelle.
//
// **Dieselbe Versetzung wie in `skriptmerkmale`**, und zwar buchstäblich
// dieselbe Funktion: der Lexer zählt ab Zeile eins, Spalte eins, und der
// Rumpf steht mitten in einer Datei. Zwei Rechnungen für eine Frage liefen
// auseinander, sobald eine von beiden eine Form dazulernt — und genau das ist
// am 2026-09-02 geschehen, als der eingerückte Rumpf dazukam.
//
// **Ein neuer Verbund und keine Änderung am alten.** `msescript.Fehler` kommt
// aus einem Blatt; ihn hier zu verändern hieße, einen fremden Wert zu
// beschreiben, den ein Aufrufer schon in der Hand hält.
// merkmalfehler prüft einen Wert nach seiner Art: ein Skriptwert wird zerteilt,
// ein Textskriptwert streng als Zeichenkette mit Ausdruck gelesen (#1934),
// jeder andere hat keinen Fehler.
func merkmalfehler(merkmal msepaket.Farbmerkmal) *msescript.Fehler {
	switch merkmal.Art {
	case msepaket.FarbeSkriptwert:
		_, err := msescript.Zerteile(mitEinbindungen(merkmal.Text))

		return zerteilfehler(merkmal, err)

	case msepaket.FarbeTextskriptwert:
		return zerteilfehler(merkmal, msescript.TextskriptFehler(merkmal.Text))

	default:
		return nil
	}
}

func zerteilfehler(merkmal msepaket.Farbmerkmal, err error) *msescript.Fehler {
	if err == nil {
		return nil
	}

	// **Über [msescript.Befund] und nicht über ein eigenes `errors.As`.** Ein
	// Zweig für den Fall, dass der Fehler keiner des Zerteilers ist, bliebe
	// unbetreten — der Zerteiler und der Lexer geben nichts anderes, und ein
	// Zweig ohne Fall ist eine Behauptung darüber, was im Ernstfall geschieht
	// (§3).
	//
	// `Befund` steht in `msescript`, wo jeder Fehler hereinkommen darf und
	// beide Wege zu begehen sind.
	befund := msescript.Befund(err)

	zeile, spalte := versetzt(merkmal, befund.Zeile, befund.Spalte)

	return &msescript.Fehler{
		Zeile:      zeile,
		Spalte:     spalte,
		Meldung:    befund.Meldung,
		Schluessel: befund.Schluessel,
		Werte:      befund.Werte,
	}
}

// mitEinbindungen schreibt `include file: X` so um, wie MSE2s Lexer es liest:
// als Aufruf `include_file("X")`, der Rest der Zeile ist der Dateiname
// (`src/script/parser.cpp` Zeilen 185–195, Fassung f9d9356 vom 2021-07-06).
//
// **Nur für die Prüfung.** Gezeichnet wird über `msepaket`, das die
// Einbindung schon im Baum auflöst; ein Skript mit `include file:` erreicht
// den Zerteiler dort nie. Die Prüfung liest aber den rohen Text, und für sie
// war `init script: include file: einstieg` ein Ausdruck, der nicht aufhört —
// in drei Dateien der Saat (FMT-02, #1691).
func mitEinbindungen(text string) string {
	const marke = "include file:"

	zeilen := strings.Split(text, "\n")
	for nummer, zeile := range zeilen {
		vor, datei, gefunden := strings.Cut(zeile, marke)
		if gefunden {
			zeilen[nummer] = vor + "include_file(" + strconv.Quote(strings.TrimSpace(datei)) + ")"
		}
	}

	return strings.Join(zeilen, "\n")
}
