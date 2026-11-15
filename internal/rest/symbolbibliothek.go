// **Die Symbolbibliothek über den Dienst** — DOM-02
// (#1697).
//
// Einordnung nach §16: **erweiternd** — zwölf Wege kommen hinzu. **Eine
// Datei ist ein Text im Rumpf**: ein Dienst steht an keiner Platte des
// Rufers (Vermerk 0140).

package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

// Symbolbibliothekswerkstatt ist, was der Dienst für die Bibliothek braucht
// — **gleichlautend mit `mcp.Symbolbibliothek`**, und das ist Absicht:
// dieselbe Werkstatt bedient beide.
type Symbolbibliothekswerkstatt interface {
	Symbolbibliothek(ctx context.Context, suche string) (any, error)
	BibliothekssymbolText(ctx context.Context, id int64) (any, error)
	BibliothekssymbolAufnehmen(ctx context.Context, setID int64, name string) (any, error)
	BibliothekssymbolKopieren(ctx context.Context, id int64, name string) (any, error)
	BibliothekssymbolUmbenennen(ctx context.Context, id int64, name string) (any, error)
	BibliothekssymbolLoeschen(ctx context.Context, id int64, bestaetigt bool) (any, error)
	BibliothekssymbolEinsetzen(ctx context.Context, id, setID int64) (any, error)
	BibliothekssymbolAusText(ctx context.Context, inhalt, name string) (any, error)

	// Die Bearbeitung — DOM-02.AK-02, in `symbolbibliothekpflege.go`.
	BibliothekssymbolTeile(ctx context.Context, id int64) (any, error)
	BibliothekssymbolSichern(ctx context.Context, id int64, teile json.RawMessage) (any, error)
	BibliothekssymbolZuruecknehmen(ctx context.Context, id int64) (any, error)
	BibliothekssymbolNachziehen(ctx context.Context, id int64) (any, error)
}

// symbolbibliothekswege trägt die Wege ein.
//
// **Das Einlesen und das Aufnehmen sind ein `POST` auf die Sammlung**, weil
// ein neues Symbol entsteht; **das Einsetzen steht am Set**, weil sich dessen
// Symbol ändert.
func (d *Dienst) symbolbibliothekswege(weg *http.ServeMux) {
	weg.HandleFunc("GET /symbole", d.symbolbibliothek)
	weg.HandleFunc("POST /symbole", d.bibliothekssymbolAufnehmen)
	weg.HandleFunc("POST /symbole/einlesen", d.bibliothekssymbolEinlesen)
	weg.HandleFunc("GET /symbole/{symbolID}", d.bibliothekssymbolText)
	weg.HandleFunc("POST /symbole/{symbolID}/kopie", d.bibliothekssymbolKopieren)
	weg.HandleFunc("PUT /symbole/{symbolID}/name", d.bibliothekssymbolUmbenennen)
	weg.HandleFunc("DELETE /symbole/{symbolID}", d.bibliothekssymbolLoeschen)
	weg.HandleFunc("PUT /sets/{setID}/bibliothekssymbol", d.bibliothekssymbolEinsetzen)

	// **Die Bearbeitung** — die Teile sind eine Ressource des Symbols; Zurücknehmen
	// und Nachziehen sind Vorgänge, deshalb `POST`.
	weg.HandleFunc("GET /symbole/{symbolID}/teile", d.bibliothekssymbolTeile)
	weg.HandleFunc("PUT /symbole/{symbolID}/teile", d.bibliothekssymbolSichern)
	weg.HandleFunc("POST /symbole/{symbolID}/zuruecknahme", d.bibliothekssymbolZuruecknehmen)
	weg.HandleFunc("POST /symbole/{symbolID}/nachzug", d.bibliothekssymbolNachziehen)
}

// symbolbibliothek nennt die Symbole — `?suche=` schränkt ein.
func (d *Dienst) symbolbibliothek(schreiben http.ResponseWriter, anfrage *http.Request) {
	antworteMit(schreiben, http.StatusOK)(d.naehte.Symbolbibliothek.Symbolbibliothek(
		anfrage.Context(), anfrage.URL.Query().Get("suche")))
}

// bibliothekssymbolText gibt ein Symbol als Text einer `.mse-symbol`.
func (d *Dienst) bibliothekssymbolText(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.amSymbol(schreiben, anfrage, d.naehte.Symbolbibliothek.BibliothekssymbolText)
}

// bibliothekssymbolAufnehmen legt das Symbol eines Sets in die Bibliothek.
func (d *Dienst) bibliothekssymbolAufnehmen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		SetID int64  `json:"setID"`
		Name  string `json:"name"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	antworteMit(schreiben, http.StatusCreated)(d.naehte.Symbolbibliothek.BibliothekssymbolAufnehmen(
		anfrage.Context(), rumpf.SetID, rumpf.Name))
}

// bibliothekssymbolEinlesen legt ein Symbol aus dem Text einer `.mse-symbol`
// an.
func (d *Dienst) bibliothekssymbolEinlesen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Inhalt string `json:"inhalt"`
		Name   string `json:"name"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	inhalt, gut := pflichttext(schreiben, "inhalt", rumpf.Inhalt)
	if !gut {
		return
	}

	antworteMit(schreiben, http.StatusCreated)(d.naehte.Symbolbibliothek.BibliothekssymbolAusText(
		anfrage.Context(), inhalt, rumpf.Name))
}

// bibliothekssymbolKopieren legt ein Symbol unter anderem Namen noch einmal an.
func (d *Dienst) bibliothekssymbolKopieren(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.mitSymbolnamen(schreiben, anfrage, http.StatusCreated,
		d.naehte.Symbolbibliothek.BibliothekssymbolKopieren)
}

// bibliothekssymbolUmbenennen gibt einem Symbol einen anderen Namen.
func (d *Dienst) bibliothekssymbolUmbenennen(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.mitSymbolnamen(schreiben, anfrage, http.StatusOK,
		d.naehte.Symbolbibliothek.BibliothekssymbolUmbenennen)
}

// mitSymbolnamen bedient die Wege, die ein Symbol und einen Namen nehmen.
func (d *Dienst) mitSymbolnamen(
	schreiben http.ResponseWriter, anfrage *http.Request, status int,
	arbeit func(context.Context, int64, string) (any, error),
) {
	id, gut := kennung(schreiben, anfrage, "symbolID")
	if !gut {
		return
	}

	var rumpf struct {
		Name string `json:"name"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	antworteMit(schreiben, status)(arbeit(anfrage.Context(), id, rumpf.Name))
}

// bibliothekssymbolLoeschen löscht — **ein getragenes nur mit
// `?bestaetigt=true`**, und die Absage nennt die Sets.
func (d *Dienst) bibliothekssymbolLoeschen(schreiben http.ResponseWriter, anfrage *http.Request) {
	id, gut := kennung(schreiben, anfrage, "symbolID")
	if !gut {
		return
	}

	bestaetigt, _ := strconv.ParseBool(anfrage.URL.Query().Get("bestaetigt")) //nolint:errcheck // unlesbar heißt nein

	antworteMit(schreiben, http.StatusOK)(d.naehte.Symbolbibliothek.BibliothekssymbolLoeschen(
		anfrage.Context(), id, bestaetigt))
}

// bibliothekssymbolEinsetzen macht ein Symbol der Bibliothek zu dem des Sets.
func (d *Dienst) bibliothekssymbolEinsetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		ID int64 `json:"id"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	antworteMit(schreiben, http.StatusOK)(d.naehte.Symbolbibliothek.BibliothekssymbolEinsetzen(
		anfrage.Context(), rumpf.ID, setID))
}

// antworteMit schreibt ein Ergebnis mit `status` — oder seinen Fehler.
func antworteMit(schreiben http.ResponseWriter, status int) func(any, error) {
	return func(ergebnis any, fehler error) {
		if fehler != nil {
			fehlerhaft(schreiben, fehler)

			return
		}

		ja(schreiben, status, ergebnis)
	}
}
