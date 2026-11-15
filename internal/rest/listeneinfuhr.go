// **Die Listeneinfuhr über den Dienst** — Werkstück 2 von Phase 4.5
// (#1292).
//
// Einordnung nach §16: **erweiternd** — zwei Wege kommen hinzu, keiner ändert
// seine Bedeutung.

package rest

import (
	"context"
	"net/http"
)

// Listeneinfuhrwerkstatt ist, was der Dienst zum Einlesen einer Liste braucht.
//
// **Die Methoden heißen `ListeDeuten` und `ListeEinlesen`** und nicht bloß
// `Deuten` und `Einlesen`. In diesem Paket erfüllt **ein** Doppel jede Naht;
// beide kurzen Namen sind schon vergeben — `Einlesen` an den Scryfall-Satz,
// `Uebernehmen` an den Abgleich. Ein Name je Fähigkeit statt je Naht ist hier
// keine Umständlichkeit, sondern die Bedingung dafür, dass die Nähte
// nebeneinander bestehen.
type Listeneinfuhrwerkstatt interface {
	ListeDeuten(ctx context.Context, text string) (any, error)
	ListeEinlesen(ctx context.Context, zeilen []Listeneinfuhrzeile, zusammenfuehrung string) (any, error)

	// ListeVorschauen sagt, was das Einlesen täte — DOM-03.AK-01 (#1698).
	ListeVorschauen(ctx context.Context, zeilen []Listeneinfuhrzeile) (any, error)
}

// Listeneinfuhrzeile ist eine Zeile mit entschiedenem Druck.
type Listeneinfuhrzeile struct {
	Printing string `json:"printing"`
	Menge    int    `json:"menge"`
	Zustand  string `json:"zustand"`
	Sprache  string `json:"sprache"`

	// Foil ist der alte Wahrheitswert — abgekündigt seit T2 (0119); beim
	// Schreiben entscheidet Finish, wenn es dasteht.
	Foil bool `json:"foil"`

	// Finish ist die Ausführung: `nonfoil`, `foil` oder `etched`.
	Finish string `json:"finish"`

	Name    string `json:"name"`
	Setcode string `json:"setcode"`
	Nummer  string `json:"nummer"`
}

// listeDeuten liest eine Liste und sagt, welche Drucke dazu passen.
//
// **`POST` und nicht `GET`, obwohl es nichts schreibt.** Eine Liste von
// dreihundert Zeilen passt in keine Adresse; ein `GET` mit Rumpf ist zwar
// erlaubt, wird aber von Zwischenspeichern und Werkzeugen unterschiedlich
// behandelt. Dass es nichts ändert, steht in der Erklärung — nicht im Verb.
func (d *Dienst) listeDeuten(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Text string `json:"text"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	deutung, fehler := d.naehte.Listeneinfuhr.ListeDeuten(anfrage.Context(), rumpf.Text)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, deutung)
}

// listeEinlesen nimmt die genannten Zeilen in die Sammlung auf.
func (d *Dienst) listeEinlesen(schreiben http.ResponseWriter, anfrage *http.Request) {
	// **`zusammenfuehrung` ist freiwillig** — erweiternd im Sinne von 0033
	// (DOM-03.AK-01, #1698); ohne sie wird addiert wie bisher.
	var rumpf struct {
		Zeilen           []Listeneinfuhrzeile `json:"zeilen"`
		Zusammenfuehrung string               `json:"zusammenfuehrung"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	bericht, fehler := d.naehte.Listeneinfuhr.ListeEinlesen(anfrage.Context(), rumpf.Zeilen,
		rumpf.Zusammenfuehrung)

	// **Ein Bericht samt Fehler geht samt Fehler hinaus** (#1291, Nachlauf):
	// was geschrieben ist, ist geschrieben (§7, Regel 3).
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, bericht)
}

// listeVorschauen sagt, was das Einlesen der Zeilen täte — **ohne zu
// schreiben**; `POST` aus demselben Grund wie das Deuten (DOM-03.AK-01).
func (d *Dienst) listeVorschauen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Zeilen []Listeneinfuhrzeile `json:"zeilen"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	vorschau, fehler := d.naehte.Listeneinfuhr.ListeVorschauen(anfrage.Context(), rumpf.Zeilen)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, vorschau)
}
