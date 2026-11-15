// **Wie dieses Programm hinausgeht, am MCP-Server** — Werkstück N5 von
// Phase 5.4
// (#1571).
//
// # Warum es das Werkzeug gibt, obwohl der Proxy etwas über dieses Netz sagt
//
// **Weil `lesen.einstellungen` ihn ohnehin nennt.** Der erste Anlauf lehnte
// ein Werkzeug hier ab — mit derselben Begründung wie bei
// `PeerkontaktAdresse`: eine Adresse im eigenen Netz geht ein Modell nichts
// an. Nachgesehen statt erinnert: die Einstellungen kommen dort **vollständig**
// heraus, Proxy eingeschlossen. Eine Ablehnung, die einen Wert schützen soll,
// der eine Zeile weiter offen steht, ist eine Ausrede und keine Begründung
// (§10).
//
// **Was dieses Werkzeug darüber hinaus sagt**, ist das, was die rohe
// Einstellung nicht sagt: was **wirksam** ist, und **woher** es kommt. Ein
// Modell, dessen Anfrage an einem Proxy scheitert, sieht hier den Grund statt
// eines „connection refused".

package mcp

import (
	"context"
	"encoding/json"
)

// Netzweg ist, was das eine Werkzeug dieser Datei braucht.
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

// netzwegwerkzeug baut das eine.
func netzwegwerkzeug(quelle Netzweg) []Werkzeug {
	if quelle == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.netzweg", Art: Lesend, Idempotent: true,
		Erklaerung: "Sagt, wie dieser Prozess ins Netz geht: Proxy, Fristen und " +
			"woher jeder Wert stammt — aus den Einstellungen oder aus der Umgebung. " +
			"ÄNDERT NICHTS. Geändert wird über »schreiben.einstellung« mit den " +
			"Schlüsseln, die unter »netz.« stehen.",
		Schema: objekt(nil),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			angaben, err := quelle.Netzstand(ctx)
			if err != nil {
				return "", err
			}

			return alsJSON(angaben)
		},
	}}
}
