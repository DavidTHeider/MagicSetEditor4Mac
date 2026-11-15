// **Freigaben über REST** — Werkstück 11 von Phase 4.12
// (#1351).
//
// **Drei Wege, alle erweiternd** (§16).
//
// # Warum `PUT` auf den Nutzer und nicht `POST` auf die Liste
//
// **Weil der Nutzer der Ort ist.** „Dieser Mensch hat diese Rolle" ist eine
// Aussage über einen Ort, und derselbe Ruf zweimal führt zum selben Zustand —
// ein `POST` legte beim zweiten Mal eine zweite Freigabe an, die es nicht
// geben kann.

package rest

import (
	"context"
	"net/http"
)

// Freigabenstand ist, wer an einem Profil was darf.
type Freigabenstand struct {
	Freigaben []Freigabezeile `json:"freigaben"`

	// Sichtbarkeit ist `privat` oder `bestand`.
	Sichtbarkeit string `json:"sichtbarkeit"`

	// EigeneRolle sagt, was der arbeitende Mensch darf — leer heißt: nichts.
	EigeneRolle string `json:"eigeneRolle"`

	// Nachladefehler sagt: die Änderung **gilt**, nur die Liste danach war
	// nicht zu lesen — etwa nach dem eigenen Herabstufen (§7, Regel 3;
	// #1970). **Erweiternd im Sinne von 0033**: wo nichts fehlt, steht es
	// nicht da.
	Nachladefehler *Befund `json:"nachladefehler,omitempty"`
}

// Freigabezeile ist die Rolle eines Nutzers.
type Freigabezeile struct {
	NutzerID string `json:"nutzerID"`
	Name     string `json:"name"`
	Rolle    string `json:"rolle"`
	Angelegt string `json:"angelegt"`
}

// Freigaberolle ist der Rumpf eines `PUT` — **nur die Rolle**.
//
// **Kennungen stehen im Weg**: der Ort ist das Paar aus Profil und Nutzer.
type Freigaberolle struct {
	Rolle string `json:"rolle"`
}

// Fremdfreigaben führt die Freigaben an Profilen.
type Fremdfreigaben interface {
	Freigaben(ctx context.Context, profilID string) (Freigabenstand, error)
	FreigabeSetzen(ctx context.Context, profilID, nutzerID, rolle string) (
		Freigabenstand, error)
	FreigabeLoeschen(ctx context.Context, profilID, nutzerID string) (
		Freigabenstand, error)
}

// freigaben gibt, wer an einem Profil was darf.
func (d *Dienst) freigaben(schreiben http.ResponseWriter, anfrage *http.Request) {
	stand, fehler := d.naehte.Freigaben.Freigaben(
		anfrage.Context(), anfrage.PathValue("profilID"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, stand)
}

// freigabeSetzen gibt einem Nutzer eine Rolle.
func (d *Dienst) freigabeSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf Freigaberolle

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	stand, fehler := d.naehte.Freigaben.FreigabeSetzen(anfrage.Context(),
		anfrage.PathValue("profilID"), anfrage.PathValue("nutzerID"), rumpf.Rolle)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, stand)
}

// freigabeLoeschen zieht eine Freigabe zurück.
func (d *Dienst) freigabeLoeschen(schreiben http.ResponseWriter, anfrage *http.Request) {
	stand, fehler := d.naehte.Freigaben.FreigabeLoeschen(anfrage.Context(),
		anfrage.PathValue("profilID"), anfrage.PathValue("nutzerID"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, stand)
}
