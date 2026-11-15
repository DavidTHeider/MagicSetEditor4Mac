// **Ein Set löschen über die Rückschleife** —
// #1218.
//
// # Warum der Name in der Abfrage steht und nicht im Rumpf
//
// Weil `DELETE` keinen Rumpf haben soll — RFC 9110 §9.3.5 sagt, dass er keine
// definierte Bedeutung hat und dass ein Absenden ihn scheitern lassen kann.
// Derselbe Schnitt wie bei `DELETE /sammlung/{id}?fassung=…`: was die
// Bedingung des Löschens ist, steht in der Abfrage.

package rest

import (
	"context"
	"net/http"
)

// Setloescher ist, was der Dienst zum Löschen braucht.
type Setloescher interface {
	SetLoeschen(ctx context.Context, setID int64, name string) ([]string, error)
}

// setLoeschen nimmt ein Set aus dem Bestand.
//
// **200 mit Befunden und nicht 204.** Der Rumpf trägt, was liegen blieb — und
// eine Antwort ohne Rumpf hätte dafür keinen Platz (§7, Regel 3).
func (d *Dienst) setLoeschen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	// **Ein leerer Name wird durchgereicht und nicht hier abgewiesen.** Der
	// Bestand vergleicht ihn mit dem heutigen; dass ein leerer nie passt, ist
	// seine Antwort und nicht die dieser Fläche. Eine zweite Prüfung hier
	// gäbe eine zweite Meldung für dieselbe Sache.
	name := anfrage.URL.Query().Get("name")

	befunde, fehler := d.naehte.Loescher.SetLoeschen(anfrage.Context(), setID, name)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"befunde": ohneLeeretexte(befunde)})
}

// ohneLeeretexte macht aus `nil` eine leere Liste.
func ohneLeeretexte(texte []string) []string {
	if texte == nil {
		return []string{}
	}

	return texte
}
