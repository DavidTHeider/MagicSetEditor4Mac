package rest

import (
	"context"
	"net/http"
)

// Fremdabgleich ist, was die beiden Scryfall-Endpunkte brauchen.
//
// Die Schnittstelle sagt, was `rest` **braucht** (§2). Erfüllt wird sie im
// Composition Root.
type Fremdabgleich interface {
	Vorschlag(ctx context.Context, setID, karteID int64, name string) (Vorschlag, error)
	Uebernehmen(ctx context.Context, setID, karteID int64, name string, felder []string) (
		vorher []Aenderung, fehler error)
}

// Vorschlag ist, was eine Übernahme ändern würde.
type Vorschlag struct {
	Aenderungen  []Aenderung
	Befunde      []string
	Herkunft     string
	AlterStunden int
}

// Aenderung ist ein Feld, das sich unterscheidet.
type Aenderung struct {
	Feld string
	Alt  string
	Neu  string
}

// aenderungzeile ist eine Änderung, wie dieser Dienst sie herausgibt.
type aenderungzeile struct {
	Feld string `json:"feld"`
	Alt  string `json:"alt"`
	Neu  string `json:"neu"`
}

// scryfallVorschlag zeigt, was sich ändern würde — und ändert nichts.
func (d *Dienst) scryfallVorschlag(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gutesSet := kennung(schreiben, anfrage, "setID")
	karteID, guteKarte := kennung(schreiben, anfrage, "karteID")

	if !gutesSet || !guteKarte {
		return
	}

	name := anfrage.URL.Query().Get("name")
	if name == "" {
		nein(schreiben, http.StatusBadRequest, "validierung",
			"der Kartenname fehlt — er gehört als ?name= an die Adresse")

		return
	}

	vorschlag, fehler := d.naehte.Abgleich.Vorschlag(anfrage.Context(), setID, karteID, name)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"aenderungen":  alsAenderungzeilen(vorschlag.Aenderungen),
		"befunde":      alsListe(vorschlag.Befunde),
		"herkunft":     vorschlag.Herkunft,
		"alterStunden": vorschlag.AlterStunden,
	})
}

// scryfallUebernehmen schreibt die genannten Felder.
func (d *Dienst) scryfallUebernehmen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gutesSet := kennung(schreiben, anfrage, "setID")
	karteID, guteKarte := kennung(schreiben, anfrage, "karteID")

	if !gutesSet || !guteKarte {
		return
	}

	var rumpf struct {
		Name   string   `json:"name"`
		Felder []string `json:"felder"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	vorher, fehler := d.naehte.Abgleich.Uebernehmen(anfrage.Context(),
		setID, karteID, rumpf.Name, rumpf.Felder)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Die vorigen Werte reisen mit** — sonst weiß ein Aufrufer nicht, wie er
	// zurück soll. Dieselbe Zusage wie bei MCP.
	ja(schreiben, http.StatusOK, map[string]any{"vorher": alsAenderungzeilen(vorher)})
}

// alsAenderungzeilen übersetzt für die Antwort.
func alsAenderungzeilen(aenderungen []Aenderung) []aenderungzeile {
	zeilen := make([]aenderungzeile, 0, len(aenderungen))
	for _, eine := range aenderungen {
		zeilen = append(zeilen, aenderungzeile(eine))
	}

	return zeilen
}

// alsListe macht aus einem `nil`-Slice eine leere Liste.
//
// **Sonst käme `null` an**, und ein Aufrufer, der darüber schleift, bekäme
// einen Fehler statt keiner Runde — die Falle aus AGENTS.md §3.
func alsListe(texte []string) []string {
	if texte == nil {
		return []string{}
	}

	return texte
}
