// **Ein Symbol der Bibliothek bearbeiten, über den Dienst** — DOM-02.AK-02
// (#1697).
//
// Einordnung nach §16: **erweiternd**. Die Teile haben dieselbe Gestalt wie
// bei `POST /symbol/{rechnung}`.

package rest

import (
	"context"
	"encoding/json"
	"net/http"
)

// bibliothekssymbolTeile gibt die Zeichnung eines Symbols als Teile.
func (d *Dienst) bibliothekssymbolTeile(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.amSymbol(schreiben, anfrage, d.naehte.Symbolbibliothek.BibliothekssymbolTeile)
}

// bibliothekssymbolSichern legt eine neue Fassung ab.
func (d *Dienst) bibliothekssymbolSichern(schreiben http.ResponseWriter, anfrage *http.Request) {
	id, gut := kennung(schreiben, anfrage, "symbolID")
	if !gut {
		return
	}

	var rumpf struct {
		Teile json.RawMessage `json:"teile"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	antworteMit(schreiben, http.StatusOK)(d.naehte.Symbolbibliothek.BibliothekssymbolSichern(
		anfrage.Context(), id, rumpf.Teile))
}

// bibliothekssymbolZuruecknehmen holt die vorige Fassung zurück.
func (d *Dienst) bibliothekssymbolZuruecknehmen(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.amSymbol(schreiben, anfrage, d.naehte.Symbolbibliothek.BibliothekssymbolZuruecknehmen)
}

// bibliothekssymbolNachziehen legt die jetzige Fassung in die Sets, die die
// vorige tragen.
func (d *Dienst) bibliothekssymbolNachziehen(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.amSymbol(schreiben, anfrage, d.naehte.Symbolbibliothek.BibliothekssymbolNachziehen)
}

// amSymbol bedient die Wege, die nur ein Symbol nennen.
func (d *Dienst) amSymbol(
	schreiben http.ResponseWriter, anfrage *http.Request,
	arbeit func(context.Context, int64) (any, error),
) {
	id, gut := kennung(schreiben, anfrage, "symbolID")
	if !gut {
		return
	}

	antworteMit(schreiben, http.StatusOK)(arbeit(anfrage.Context(), id))
}
