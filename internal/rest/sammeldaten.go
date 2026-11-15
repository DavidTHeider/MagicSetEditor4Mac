package rest

import (
	"context"
	"net/http"
)

// Fremdsammeldaten ist, was die Sammeldatenendpunkte brauchen.
type Fremdsammeldaten interface {
	Stand(ctx context.Context) (Sammelstand, error)
	Holen(ctx context.Context) (Sammelstand, error)
}

// Sammelstand sagt, was auf der Platte liegt.
type Sammelstand struct {
	Art       string
	Pfad      string
	Vorhanden bool
	Bytes     int64
	Angeboten int64
	Stand     string
}

// sammeldatenZeigen sagt, was liegt.
//
// **GET, weil es nichts ändert** — und damit wiederholbar.
func (d *Dienst) sammeldatenZeigen(schreiben http.ResponseWriter, anfrage *http.Request) {
	stand, fehler := d.naehte.Sammeldaten.Stand(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, alsSammelzeile(stand))
}

// sammeldatenHolen lädt die Sammeldatei.
//
// **POST, weil es die Platte ändert** — rund 25 MB, und eine vorhandene Datei
// wird ersetzt. Ein GET dafür wäre die falsche Zusage: ein Zwischenspeicher
// dürfte ihn wiederholen.
func (d *Dienst) sammeldatenHolen(schreiben http.ResponseWriter, anfrage *http.Request) {
	stand, fehler := d.naehte.Sammeldaten.Holen(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, alsSammelzeile(stand))
}

// alsSammelzeile übersetzt für die Antwort.
func alsSammelzeile(stand Sammelstand) map[string]any {
	return map[string]any{
		"art":       stand.Art,
		"pfad":      stand.Pfad,
		"vorhanden": stand.Vorhanden,
		"bytes":     stand.Bytes,
		"angeboten": stand.Angeboten,
		"stand":     stand.Stand,
	}
}
