// **Einstellungen ausführen und einlesen, über den Dienst** — Werkstück O3
// von Phase 5.2
// (#1427).
//
// # Drei Endpunkte und ihre Verben
//
//   - `GET /einstellungen/ausfuhr` — sie lesen; die gewünschten Abschnitte
//     stehen in der Abfrage (`?abschnitt=nutzer&abschnitt=profil`).
//   - `POST /einstellungen/vorschau` — sagen, was ein Text ändern würde.
//     **`POST`, obwohl nichts geschieht**: der Text steht im Rumpf, und ein
//     `GET` mit Rumpf ist ein Vertrag, den Zwischenstationen nicht halten
//     müssen.
//   - `POST /einstellungen/einfuhr` — ihn übernehmen.
//
// # Warum die Übernahme kein `PUT` ist
//
// **Weil sie nicht idempotent ist.** Sie verlangt die Revision des Stands, den
// jemand gesehen hat; derselbe Ruf ein zweites Mal trifft eine andere und
// bekommt einen Konflikt. Ein `PUT` verspricht das Gegenteil.

package rest

import (
	"context"
	"net/http"
)

// Einstellungsaustausch ist, was die drei Endpunkte brauchen.
type Einstellungsaustausch interface {
	Ausfuhr(ctx context.Context, abschnitte []string) (string, error)
	Vorschau(ctx context.Context, roh []byte) (Einstellungssicht, error)
	Uebernehmen(
		ctx context.Context, roh []byte, revision string, nur []string,
	) (Einstellungsbericht, error)
}

// Einstellungsdiffzeile ist eine Zeile des Unterschieds.
type Einstellungsdiffzeile struct {
	Ebene      string `json:"ebene"`
	Schluessel string `json:"schluessel"`
	Vorher     string `json:"vorher"`
	Nachher    string `json:"nachher"`
	Herkunft   string `json:"herkunft"`
}

// Einstellungsbefund ist, was an einer Zeile nicht stimmt.
type Einstellungsbefund struct {
	Ebene      string `json:"ebene"`
	Schluessel string `json:"schluessel"`
	Grund      string `json:"grund"`
}

// Einstellungssicht ist, was eine Datei ändern würde.
type Einstellungssicht struct {
	Fassung      int                     `json:"fassung"`
	Revision     string                  `json:"revision"`
	Zeilen       []Einstellungsdiffzeile `json:"zeilen"`
	Befunde      []Einstellungsbefund    `json:"befunde"`
	Unveraendert int                     `json:"unveraendert"`
}

// Einstellungsbericht ist, was eine Übernahme bewirkt hat.
type Einstellungsbericht struct {
	Uebernommen int                     `json:"uebernommen"`
	Revision    string                  `json:"revision"`
	Zeilen      []Einstellungsdiffzeile `json:"zeilen"`
	Befunde     []Einstellungsbefund    `json:"befunde"`
}

// einstellungsausfuhr gibt die Einstellungen als Austauschtext.
func (d *Dienst) einstellungsausfuhr(schreiben http.ResponseWriter, anfrage *http.Request) {
	inhalt, fehler := d.naehte.Austausch.Ausfuhr(
		anfrage.Context(), anfrage.URL.Query()["abschnitt"])
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"inhalt": inhalt})
}

// einstellungsvorschau sagt, was ein Text ändern würde.
func (d *Dienst) einstellungsvorschau(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Inhalt string `json:"inhalt"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	sicht, fehler := d.naehte.Austausch.Vorschau(
		anfrage.Context(), []byte(rumpf.Inhalt))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, sicht)
}

// einstellungseinfuhr übernimmt die gewählten Zeilen.
func (d *Dienst) einstellungseinfuhr(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Inhalt   string   `json:"inhalt"`
		Revision string   `json:"revision"`
		Nur      []string `json:"nur"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	bericht, fehler := d.naehte.Austausch.Uebernehmen(
		anfrage.Context(), []byte(rumpf.Inhalt), rumpf.Revision, rumpf.Nur)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, bericht)
}
