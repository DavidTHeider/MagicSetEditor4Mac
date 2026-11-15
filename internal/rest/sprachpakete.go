package rest

// **Die eigenen Übersetzungen über REST** — LOC-02
// (#1688).
//
// Einordnung nach §16: **erweiternd** — sieben Wege kommen hinzu. Eine
// Datei ist ein Text im Rumpf, wie bei der Symbolbibliothek: der Dienst steht
// an keiner Platte, die der Rufer sieht.

import (
	"context"
	"net/http"
)

// Sprachwerkstatt ist, was REST für die eigenen Übersetzungen braucht.
type Sprachwerkstatt interface {
	Sprachpakete(ctx context.Context) (any, error)
	Sprachpaket(ctx context.Context, kennung string) (any, error)
	SprachpaketAnlegen(ctx context.Context, kennung, name, basis string) (any, error)
	SprachpaketAendern(ctx context.Context, kennung string, angaben, texte map[string]string) (any, error)
	SprachpaketLoeschen(ctx context.Context, kennung string) (any, error)
	SprachpaketAusgabe(ctx context.Context, kennung string) (any, error)
	SprachpaketEinlesen(ctx context.Context, inhalt string, ersetzen bool) (any, error)
}

// sprachwege hängt die sieben Wege ein.
func (d *Dienst) sprachwege(weg *http.ServeMux) {
	weg.HandleFunc("GET /sprachpakete", d.sprachpakete)
	weg.HandleFunc("POST /sprachpakete", d.sprachpaketAnlegen)
	weg.HandleFunc("POST /sprachpakete/einlesen", d.sprachpaketEinlesen)
	weg.HandleFunc("GET /sprachpakete/{kennung}", d.sprachpaket)
	weg.HandleFunc("PATCH /sprachpakete/{kennung}", d.sprachpaketAendern)
	weg.HandleFunc("DELETE /sprachpakete/{kennung}", d.sprachpaketLoeschen)
	weg.HandleFunc("GET /sprachpakete/{kennung}/datei", d.sprachpaketDatei)
}

func (d *Dienst) sprachpakete(schreiben http.ResponseWriter, anfrage *http.Request) {
	antworteMit(schreiben, http.StatusOK)(d.naehte.Sprachpakete.Sprachpakete(anfrage.Context()))
}

func (d *Dienst) sprachpaket(schreiben http.ResponseWriter, anfrage *http.Request) {
	antworteMit(schreiben, http.StatusOK)(d.naehte.Sprachpakete.Sprachpaket(
		anfrage.Context(), anfrage.PathValue("kennung")))
}

func (d *Dienst) sprachpaketDatei(schreiben http.ResponseWriter, anfrage *http.Request) {
	antworteMit(schreiben, http.StatusOK)(d.naehte.Sprachpakete.SprachpaketAusgabe(
		anfrage.Context(), anfrage.PathValue("kennung")))
}

func (d *Dienst) sprachpaketLoeschen(schreiben http.ResponseWriter, anfrage *http.Request) {
	antworteMit(schreiben, http.StatusOK)(d.naehte.Sprachpakete.SprachpaketLoeschen(
		anfrage.Context(), anfrage.PathValue("kennung")))
}

func (d *Dienst) sprachpaketAnlegen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Kennung string `json:"kennung"`
		Name    string `json:"name"`
		Basis   string `json:"basis"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	antworteMit(schreiben, http.StatusCreated)(d.naehte.Sprachpakete.SprachpaketAnlegen(
		anfrage.Context(), rumpf.Kennung, rumpf.Name, rumpf.Basis))
}

// sprachpaketAendern nimmt nur, was im Rumpf steht — **was fehlt, bleibt**.
func (d *Dienst) sprachpaketAendern(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Name     *string           `json:"name"`
		Stand    *string           `json:"stand"`
		Lizenz   *string           `json:"lizenz"`
		Herkunft *string           `json:"herkunft"`
		Texte    map[string]string `json:"texte"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	angaben := map[string]string{}
	for schluessel, wert := range map[string]*string{
		"name": rumpf.Name, "stand": rumpf.Stand, "lizenz": rumpf.Lizenz, "herkunft": rumpf.Herkunft,
	} {
		if wert != nil {
			angaben[schluessel] = *wert
		}
	}

	antworteMit(schreiben, http.StatusOK)(d.naehte.Sprachpakete.SprachpaketAendern(
		anfrage.Context(), anfrage.PathValue("kennung"), angaben, rumpf.Texte))
}

func (d *Dienst) sprachpaketEinlesen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Inhalt   string `json:"inhalt"`
		Ersetzen bool   `json:"ersetzen"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	inhalt, gut := pflichttext(schreiben, "inhalt", rumpf.Inhalt)
	if !gut {
		return
	}

	antworteMit(schreiben, http.StatusCreated)(d.naehte.Sprachpakete.SprachpaketEinlesen(
		anfrage.Context(), inhalt, rumpf.Ersetzen))
}
