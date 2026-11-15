// **Ein Bild als Symbol nachzeichnen, über den Dienst** — Werkstück F von
// Phase 4.11
// (#1340).
//
// Einordnung nach §16: **erweiternd** — ein Weg kommt hinzu.

package rest

import (
	"context"
	"encoding/json"
	"net/http"
)

// Symbolwerkstatt ist, was der Dienst am Symbol eines Sets braucht.
//
// **Eine Schnittstelle und nicht zwei.** Das Nachzeichnen kam zuerst; der
// Editor ist am 2026-09-02 dazugekommen
// (#1391),
// und beide arbeiten am selben Gegenstand. Zwei Nähte nebeneinander hießen,
// dass ein Aufrufer beide einsetzen muss, um ein Symbol zu bedienen.
type Symbolwerkstatt interface {
	SymbolNachzeichnen(ctx context.Context, setID int64, bildpfad string) (any, error)

	// Lesen gibt das abgelegte Symbol eines Sets — **im Profil der Anfrage**
	// (#1815): der Kontext trägt es, und nur darin gibt es das Set.
	SymbolLesen(ctx context.Context, setID int64, name string) (teile any, fehler error)

	// Ablegen schreibt eine Zeichnung neben das Set.
	SymbolAblegen(ctx context.Context, setID int64, teile json.RawMessage, unter string) (name string, fehler error)

	// Rechnen führt eine der Rechnungen des Editors aus — **ohne Bestand**.
	SymbolRechnen(was string, argumente json.RawMessage) (ergebnis any, fehler error)

	// Palette gibt, was ein Kartenfeld zum Einfügen anbietet.
	Palette(ctx context.Context, karteID int64, feld string) (eintraege any, fehler error)
}

// symbolNachzeichnen macht aus einer Bilddatei das Symbol eines Sets.
//
// **`POST` und nicht `PUT`.** Zweimal dasselbe Bild gibt zwar dieselbe
// Zeichnung, aber `SymbolSetzen` legt beim zweiten Mal **dieselbe** Datei neu
// an oder eine zweite — das entscheidet der Name, nicht dieser Weg. Ein `PUT`
// versprächen Idempotenz, die hier eine Stelle tiefer entschieden wird.
func (d *Dienst) symbolNachzeichnen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var koerper struct {
		Bildpfad string `json:"bildpfad"`
	}

	if !gelesen(schreiben, anfrage, &koerper) {
		return
	}

	antwort, fehler := d.naehte.Symbol.SymbolNachzeichnen(
		anfrage.Context(), setID, koerper.Bildpfad)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, antwort)
}
