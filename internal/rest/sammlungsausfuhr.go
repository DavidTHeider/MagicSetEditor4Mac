// **Die Sammlungsausfuhr über den Dienst** — Werkstück 4 von Phase 4.5
// (#1292).
//
// Einordnung nach §16: **erweiternd** — zwei Wege kommen hinzu, keiner ändert
// seine Bedeutung.

package rest

import (
	"context"
	"net/http"
)

// Sammlungsausfuhrwerkstatt ist, was der Dienst zum Ausgeben braucht.
type Sammlungsausfuhrwerkstatt interface {
	SammlungAusgeben(ctx context.Context, form string) (string, error)
	SammlungCSVEinlesen(ctx context.Context, text, quelle, zusammenfuehrung string) (any, error)

	// SammlungCSVVorschauen sagt, was das Einlesen täte — DOM-03.AK-01 (#1698).
	SammlungCSVVorschauen(ctx context.Context, text string) (any, error)
}

// sammlungAusgeben schreibt die Sammlung als Text.
//
// **Die Form steht in der Abfrage** (`?form=csv`), weil sie die Antwort
// auswählt und nicht ändert — ein `GET` mit Rumpf wäre hier fehl am Platz.
func (d *Dienst) sammlungAusgeben(schreiben http.ResponseWriter, anfrage *http.Request) {
	form := anfrage.URL.Query().Get("form")
	if form == "" {
		form = "csv"
	}

	text, fehler := d.naehte.Sammlungsausfuhr.SammlungAusgeben(anfrage.Context(), form)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Als JSON und nicht als `text/csv`.** Jeder andere Weg dieses Dienstes
	// antwortet mit JSON, und ein einzelner mit einem anderen Typ zwänge jeden
	// Klienten zu einer Ausnahme. Wer eine Datei will, schreibt sie selbst.
	ja(schreiben, http.StatusOK, map[string]any{"form": form, "text": text})
}

// sammlungCSVEinlesen liest ein CSV zurück.
func (d *Dienst) sammlungCSVEinlesen(schreiben http.ResponseWriter, anfrage *http.Request) {
	// **`quelle` ist freiwillig — erweiternd im Sinne von 0033.** Ein älterer
	// Aufrufer schickt sie nicht und bekommt dasselbe wie bisher.
	var rumpf struct {
		Text             string `json:"text"`
		Quelle           string `json:"quelle"`
		Zusammenfuehrung string `json:"zusammenfuehrung"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	bericht, fehler := d.naehte.Sammlungsausfuhr.SammlungCSVEinlesen(
		anfrage.Context(), rumpf.Text, rumpf.Quelle, rumpf.Zusammenfuehrung)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, bericht)
}

// sammlungCSVVorschauen sagt, was das Einlesen eines CSV täte — **ohne zu
// schreiben**, auch keinen Druck (DOM-03.AK-01, #1698).
func (d *Dienst) sammlungCSVVorschauen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Text string `json:"text"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	vorschau, fehler := d.naehte.Sammlungsausfuhr.SammlungCSVVorschauen(anfrage.Context(), rumpf.Text)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, vorschau)
}
