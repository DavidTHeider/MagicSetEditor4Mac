package rest

// **Der lokale Suchindex am Dienst** — Werkstück A15 von Phase 5.3
// (#1514).
//
// **Vier Wege, hinter der Marke wie jeder andere.** Der Aufbau ist der
// einzige schreibende; er läuft, solange die Anfrage läuft, und `DELETE` auf
// denselben Weg bricht ihn ab — kein zweiter Bezeichner für dieselbe Sache.

import (
	"context"
	"net/http"
)

// Indexauskunft ist die Naht zur Anwendung.
type Indexauskunft interface {
	Indexstand(ctx context.Context) (any, error)
	Indexsuche(ctx context.Context, frage string, grenze int,
		filter map[string][]string) (any, error)
	IndexAufbauen(ctx context.Context, modell, quelle, revision, datum string,
		fortschreiben bool) (any, error)
	IndexAbbrechen(ctx context.Context) (any, error)
}

// indexstand beantwortet `GET /index`.
func (d *Dienst) indexstand(schreiben http.ResponseWriter, anfrage *http.Request) {
	stand, fehler := d.naehte.Index.Indexstand(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, stand)
}

// indexsuche beantwortet `GET /index/suche`.
//
// **Die Filter stehen in der Abfrage, mehrfach erlaubt.**
// `?quelle=regelwerk&farbe=R&farbe=U` ist die Gestalt, in der HTTP eine
// Mehrfachauswahl kennt — ein JSON-Rumpf an einem `GET` wäre eine eigene
// Erfindung.
func (d *Dienst) indexsuche(schreiben http.ResponseWriter, anfrage *http.Request) {
	abfrage := anfrage.URL.Query()

	ergebnis, fehler := d.naehte.Index.Indexsuche(anfrage.Context(),
		abfrage.Get("frage"), grenzeAusAbfrage(abfrage.Get("grenze")),
		map[string][]string{
			"quellen":        abfrage["quelle"],
			"sprachen":       abfrage["sprache"],
			"sets":           abfrage["set"],
			"farben":         abfrage["farbe"],
			"farbidentitaet": abfrage["farbidentitaet"],
			"typen":          abfrage["typ"],
			"format":         {abfrage.Get("format")},
			"formatlage":     {abfrage.Get("formatlage")},
			"formatstand":    {abfrage.Get("formatstand")},
			"manawertVon":    {abfrage.Get("manawertVon")},
			"manawertBis":    {abfrage.Get("manawertBis")},
		})
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, ergebnis)
}

// indexaufbau beantwortet `POST /index/aufbau`.
func (d *Dienst) indexaufbau(schreiben http.ResponseWriter, anfrage *http.Request) {
	var eingabe struct {
		Modell     string `json:"modell"`
		Quelle     string `json:"quelle"`
		Revision   string `json:"revision"`
		Quelldatum string `json:"quelldatum"`

		// Fortschreiben baut auf der aktiven Generation auf — die Quelle
		// liefert dann nur, was sich geändert hat.
		Fortschreiben bool `json:"fortschreiben"`
	}

	if !gelesen(schreiben, anfrage, &eingabe) {
		return
	}

	bericht, fehler := d.naehte.Index.IndexAufbauen(anfrage.Context(),
		eingabe.Modell, eingabe.Quelle, eingabe.Revision, eingabe.Quelldatum,
		eingabe.Fortschreiben)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, bericht)
}

// indexabbruch beantwortet `DELETE /index/aufbau`.
//
// **Kein Fehler, wenn nichts lief.** „Es läuft nichts" ist eine Auskunft; ein
// 404 hieße, den Weg gäbe es nicht.
func (d *Dienst) indexabbruch(schreiben http.ResponseWriter, anfrage *http.Request) {
	antwort, fehler := d.naehte.Index.IndexAbbrechen(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, antwort)
}

// grenzeAusAbfrage liest die Trefferzahl — zehn, wo keine steht.
func grenzeAusAbfrage(roh string) int {
	if gelesen := zahlAus(roh); gelesen > 0 {
		return gelesen
	}

	return 10
}
