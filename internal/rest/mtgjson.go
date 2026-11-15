// **MTG JSON über den Dienst** —
// #1291.
//
// **Eine eigene Datei**, weil `endpunkte.go` an der Zeilengrenze aus §2 steht.

package rest

import (
	"net/http"
	"strconv"
)

// mtgjsonLesen fährt eine MTG-JSON-Quelle in den Bestand ein.
//
// **`POST /mtgjson` und nicht `POST /sets`** — jener fährt ein **Setpaket**
// ein. Zwei Formate hinter einer Adresse wären eine Zusage, die niemand lesen
// kann (#891).
//
// Einordnung nach §16: **erweiternd** — ein Weg kommt hinzu, keiner ändert
// seine Bedeutung.
func (d *Dienst) mtgjsonLesen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Pfad  string `json:"pfad"`
		Spiel string `json:"spiel"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	angelegt, befunde, fehler := d.naehte.Ausfuhr.MTGJSONLesen(
		anfrage.Context(), rumpf.Pfad, rumpf.Spiel)

	// **Die Kennungen als Zeichenketten**, wie überall im Dienst: eine
	// JavaScript-Zahl trägt keine 64 Bit.
	kennungen := make([]string, 0, len(angelegt))
	for _, kennung := range angelegt {
		kennungen = append(kennungen, strconv.FormatInt(kennung, 10))
	}

	teilstand := map[string]any{
		"angelegt": kennungen, "befunde": ohneLeeretexte(befunde),
	}

	// **Der Teilerfolg geht mit dem Fehler hinaus** — Befund 1 des Zweitblicks
	// über Phase 4.4. Eine AllSets-Datei führt Hunderte Sets, jedes in einer
	// eigenen Transaktion; bricht sie mitten im Lauf ab, **stehen die
	// vorherigen im Bestand**. Wer nur `fehlerhaft` schriebe, ließe sie
	// zurück, ohne dass jemand von ihnen weiß (§7 Regel 3, #1041).
	if fehler != nil {
		fehlerhaftMit(schreiben, fehler, teilstand)

		return
	}

	ja(schreiben, http.StatusCreated, teilstand)
}

// mtgjsonSchreiben schreibt ein Set als MTG JSON.
func (d *Dienst) mtgjsonSchreiben(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		Spiel string `json:"spiel"`
		Ziel  string `json:"ziel"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	karten, befunde, fehler := d.naehte.Ausfuhr.MTGJSONSchreiben(
		anfrage.Context(), setID, rumpf.Spiel, rumpf.Ziel)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{
		"ziel": rumpf.Ziel, "karten": karten, "befunde": befunde,
	})
}

// mtgjsonSpiele nennt die Spiele, für die es eine Zuordnung gibt.
func (d *Dienst) mtgjsonSpiele(schreiben http.ResponseWriter, _ *http.Request) {
	ja(schreiben, http.StatusOK, map[string]any{"spiele": d.naehte.Ausfuhr.MTGJSONSpiele()})
}
