package app

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Skripte einer Paketdatei prüfen, während jemand sie schreibt** —
// Werkstück B von Phase 4.9
// (#1324).
//
// # Warum neben der Vorschau
//
// Die Vorschau zeichnet eine Karte; ihr Fehler kann aus **jeder** beteiligten
// Datei kommen — dem Spielskript, einem eingebundenen Paket, der bearbeiteten
// Datei. Eine Zeile daraus im offenen Editor anzustreichen wäre eine
// Behauptung, die niemand geprüft hat, und ein Zeiger auf die falsche Zeile
// ist schlechter als keiner.
//
// Diese Prüfung liest **nur den Text, der im Feld steht**.
//
// # Ohne Fehlerumschlag
//
// Wie [App.Skriptfarben]: es gibt nichts, was hier schiefgehen könnte. Eine
// Datei, die sich nicht zerteilen lässt, hat keine Skriptwerte — und dann
// gibt es keinen Befund, nicht einen Fehler über den Befund.

// Skriptbefundantwort trägt die Stelle eines Syntaxfehlers in einer
// Paketdatei.
type Skriptbefundantwort struct {
	// Stelle ist der Ort — **nil**, wenn kein Skriptwert einen Fehler hat.
	Stelle *Skriptstelle `json:"stelle"`

	// Meldung ist der deutsche Satz (§9); das Fenster setzt über
	// [Skriptbefundantwort.Schluessel] seinen eigenen ein.
	Meldung string `json:"meldung"`

	// Schluessel benennt die Meldung — derselbe Katalogschlüssel wie bei der
	// Konsole.
	Schluessel string `json:"schluessel"`

	// Werte füllt die Platzhalter des Satzes.
	//
	// **Ohne `omitempty`**, und aus dem Grund aus §3: eine leere Liste
	// verschwände aus der Antwort, und im Fenster stünde `undefined` statt
	// einer leeren Liste.
	Werte []string `json:"werte"`
}

// PaketdateiPruefen gibt den ersten Syntaxfehler in den Skripten einer
// Paketdatei.
func (a *App) PaketdateiPruefen(inhalt string) Skriptbefundantwort {
	befund := service.PaketSkriptfehler(inhalt)
	if befund == nil {
		return Skriptbefundantwort{Werte: []string{}}
	}

	werte := befund.Werte
	if werte == nil {
		werte = []string{}
	}

	return Skriptbefundantwort{
		Stelle:     &Skriptstelle{Zeile: befund.Zeile, Spalte: befund.Spalte},
		Meldung:    befund.Meldung,
		Schluessel: string(befund.Schluessel),
		Werte:      werte,
	}
}
