// **Die Testkarten der Werkstatt** — Werkstück U7 von Phase 5.1
// (#1418).
//
// Die Vorschau zeichnet ohne Nennung die erste Karte des ersten Sets, das den
// Stil benutzt. Eine Vorlage bricht aber selten an der ersten Karte: sie
// bricht am langen Text, am fehlenden Bild, an der zweiten Seite. Deshalb
// bietet der Inspector die Karten desselben Sets zur Wahl an — **dieselbe
// Suche** wie bei der Probekarte, damit Vorgabe und Auswahl nie
// auseinanderlaufen.

package app

import (
	"context"
	"strconv"
)

// ProbekarteDTO ist eine wählbare Testkarte.
type ProbekarteDTO struct {
	ID string `json:"id"`

	// Name ist der Wert des Feldes `name` — leer, wenn die Karte keinen hat.
	Name string `json:"name"`

	// Nummer ist die Nummer der Karte im Set — für die Aufschrift ohne Namen.
	Nummer int `json:"nummer"`
}

// ProbekartenAntwort trägt die Testkarten zu einem Paket.
type ProbekartenAntwort struct {
	Karten []ProbekarteDTO `json:"karten"`
	Fehler *Fehler         `json:"fehler"`
}

// Probekarten gibt die Karten des ersten Sets, das diesen Stil benutzt.
//
// **Ohne passendes Set kommt eine leere Liste und kein Fehler**: die
// Werkstatt zeigt dann, wie man zu einem Set kommt — dieselbe Auskunft wie
// unter der Vorschau.
func (a *App) Probekarten(ctx context.Context, paket string) ProbekartenAntwort {
	ctx = a.imProfil(ctx)

	if a.vorlagenzeichner == nil {
		return ProbekartenAntwort{Karten: []ProbekarteDTO{}, Fehler: Uebersetze(ctx, keinVorlagenzeichner())}
	}

	karten, err := a.vorlagenzeichner.Probekarten(ctx, paket)
	if err != nil {
		return ProbekartenAntwort{Karten: []ProbekarteDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	dtos := make([]ProbekarteDTO, 0, len(karten))

	for _, karte := range karten {
		name, _ := karte.Value("name")

		dtos = append(dtos, ProbekarteDTO{
			ID: strconv.FormatInt(karte.ID, 10), Name: name, Nummer: karte.Number,
		})
	}

	return ProbekartenAntwort{Karten: dtos}
}
