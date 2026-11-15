// **Der Symboleditor über den Dienst**
// (#1391).
//
// # Warum `POST` für Rechnungen, die nichts ändern
//
// Weil die Teile im **Rumpf** stehen. Ein Symbol hat verschachtelte Teile mit
// Bézierpunkten; in einer Adresse hätte es nichts zu suchen. Geändert wird
// dabei nichts — das steht in `docs/FAEHIGKEITEN.md` und in der Erklärung
// jedes Werkzeugs, denn die Methode allein sagt es hier nicht. Derselbe Griff
// wie bei der Paketprüfung.

package rest

import (
	"encoding/json"
	"io"
	"net/http"
)

// symbolLesen gibt das abgelegte Symbol eines Sets.
func (d *Dienst) symbolLesen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	teile, fehler := d.naehte.Symbol.SymbolLesen(anfrage.Context(), setID, anfrage.URL.Query().Get("name"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"teile": teile})
}

// symbolSetzen legt eine Zeichnung neben das Set.
func (d *Dienst) symbolSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		Teile json.RawMessage `json:"teile"`
		Unter string          `json:"unter"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	name, fehler := d.naehte.Symbol.SymbolAblegen(anfrage.Context(), setID, rumpf.Teile, rumpf.Unter)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"name": name})
}

// symbolRechnen führt eine der Rechnungen des Editors aus.
//
// **Der Name steht im Weg, die Argumente im Rumpf.** So ist an der Adresse zu
// sehen, was geschieht — und ein Zugriffsprotokoll sagt mehr als
// „POST /symbol/rechnung".
func (d *Dienst) symbolRechnen(schreiben http.ResponseWriter, anfrage *http.Request) {
	rumpf, gelesenerRumpf := rohenRumpf(schreiben, anfrage)
	if !gelesenerRumpf {
		return
	}

	ergebnis, fehler := d.naehte.Symbol.SymbolRechnen(
		"lesen.symbol."+anfrage.PathValue("rechnung"), rumpf)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"ergebnis": ergebnis})
}

// rohenRumpf liest den Anfragerumpf als rohes JSON.
//
// **Nicht über [gelesen]**, weil der Rumpf hier keine feste Gestalt hat: die
// acht Rechnungen des Symboleditors haben acht verschiedene Argumentmengen,
// und `DisallowUnknownFields` gegen einen gemeinsamen Verbund wiese die
// Hälfte davon ab. Geprüft werden die Felder dort, wo sie gelesen werden.
func rohenRumpf(schreiben http.ResponseWriter, anfrage *http.Request) (json.RawMessage, bool) {
	const hoechstens = 1 << 20 // ein Mebibyte; ein Symbol ist viel kleiner

	roh, fehler := io.ReadAll(http.MaxBytesReader(schreiben, anfrage.Body, hoechstens))
	if fehler != nil {
		nein(schreiben, http.StatusRequestEntityTooLarge, "validierung",
			"der Rumpf ist zu groß")

		return nil, false
	}

	return roh, true
}

// symbolpalette gibt, was ein Kartenfeld zum Einfügen anbietet.
//
// **Am Feld und nicht am Symbol**: die Palette hängt an einer Karte und ihrer
// Vorlage, nicht an einer Zeichnung. Deshalb steht sie unter `/karten/…` und
// nicht unter `/symbol/…`.
func (d *Dienst) symbolpalette(schreiben http.ResponseWriter, anfrage *http.Request) {
	karteID, gut := kennung(schreiben, anfrage, "karteID")
	if !gut {
		return
	}

	eintraege, fehler := d.naehte.Symbol.Palette(anfrage.Context(), karteID, anfrage.PathValue("feld"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"eintraege": eintraege})
}
