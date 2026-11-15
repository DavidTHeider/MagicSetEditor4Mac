// **Wie dieses Programm hinausgeht, über den Dienst** — Werkstück N5 von
// Phase 5.4
// (#1571).
//
// **Rein lesend.** Geändert wird über die Einstellungen — `PUT
// /einstellungen/{schluessel}` —, und ein zweiter Schreibweg hier wäre eine
// zweite Fassung derselben Frage.
//
// **Und er ist der Beleg für Werkstück 2 des Issues**: dieser Dienst ist ein
// Klient wie das Fenster, und was er hier zeigt, ist der Weg, den er selbst
// nimmt.

package rest

import (
	"context"
	"net/http"
)

// Netzweg ist, was der eine Endpunkt dieser Datei braucht.
type Netzweg interface {
	Netzstand(ctx context.Context) (Netzstandangaben, error)
}

// Netzstandangaben sind der wirksame Netzweg samt Herkunft je Feld.
type Netzstandangaben struct {
	ProxyHTTP  Netzfeld `json:"proxyHTTP"`
	ProxyHTTPS Netzfeld `json:"proxyHTTPS"`
	ProxyAus   Netzfeld `json:"proxyAus"`

	FristVerbindung int `json:"fristVerbindung"`
	FristKopf       int `json:"fristKopf"`
	FristStillstand int `json:"fristStillstand"`

	ProxyBenutzt bool `json:"proxyBenutzt"`
}

// Netzfeld ist ein wirksamer Wert samt seiner Herkunft.
type Netzfeld struct {
	Wert string `json:"wert"`

	// Herkunft ist „einstellung", „umgebung" oder „ohne".
	Herkunft string `json:"herkunft"`
}

// netzwegSchauen sagt, wie dieser Prozess nach außen geht.
func (d *Dienst) netzwegSchauen(schreiben http.ResponseWriter, anfrage *http.Request) {
	angaben, fehler := d.naehte.Netzweg.Netzstand(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, angaben)
}
