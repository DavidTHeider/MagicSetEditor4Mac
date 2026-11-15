package app

import (
	"cmp"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/fernleitung"
)

// **Der Fernstand an der Naht** — Werkstück N3 von Phase 5.4
// (#1456).

// Fernzustand sagt, ob dieses Profil über fremde Netze erreichbar ist.
type Fernzustand struct {
	// Laeuft heißt: es wird gerade gehorcht.
	Laeuft bool `json:"laeuft"`

	// Einladung ist, was ein Gegenüber zum Wählen braucht — leer heißt:
	// nichts auszugeben.
	//
	// **Sie trägt eine Fassung** ([domain.NeueEinladung]), und sie stirbt
	// mit dem Horchen: der Serverschlüssel wird nicht aufbewahrt.
	Einladung string `json:"einladung"`

	// Wege sagt je Gegenstelle, wie die Verbindung geht — NET-02.AK-03
	// (#1706).
	// **Leer heißt nicht „niemand"** — siehe Angebunden.
	Wege []FernwegDTO `json:"wege"`

	// Angebunden sagt, wie viele Ströme gerade offen sind — SEC-21
	// (#2162).
	// **Unabhängig von Wege**: Wege nennt die Gegenstellen, die tailcat kennt
	// (seit v0.7.0, K01), Angebunden die offenen Ströme
	// ([fernleitung.Horcher.Angebunden]); sind Ströme offen und Wege leer —
	// zwischen Anbinden und erstem Weg —, ist der Weg unbestimmt, nicht
	// „niemand".
	Angebunden int `json:"angebunden"`
}

// FernwegDTO ist ein Weg zu einer Gegenstelle.
type FernwegDTO struct {
	// Art ist `direkt`, `relais` oder `aufbau` — das letzte heißt: die
	// Gegenstelle ist bekannt, aber noch über keinen Weg erreichbar.
	Art string `json:"art"`

	// Adresse ist beim direkten Weg die Gegenstelle, beim Relais die
	// DERP-Region; beim Aufbau leer.
	Adresse string `json:"adresse"`
}

// fernwege ordnet die Wege: direkt, dann Relais, dann Aufbau.
//
// **Ein direkter Weg schlägt das Relais**, auch wenn tailcat beide nennt:
// die Region bleibt als Heimat stehen, gesprochen wird über die Adresse.
func fernwege(roh []fernleitung.Fernweg) []FernwegDTO {
	wege := make([]FernwegDTO, 0, len(roh))

	for _, weg := range roh {
		switch {
		case weg.Direkt != "":
			wege = append(wege, FernwegDTO{Art: "direkt", Adresse: weg.Direkt})
		case weg.Relais != "":
			wege = append(wege, FernwegDTO{Art: "relais", Adresse: weg.Relais})
		default:
			wege = append(wege, FernwegDTO{Art: "aufbau"})
		}
	}

	rang := map[string]int{"direkt": 0, "relais": 1, "aufbau": 2}
	slices.SortStableFunc(wege, func(a, b FernwegDTO) int {
		return cmp.Or(cmp.Compare(rang[a.Art], rang[b.Art]), cmp.Compare(a.Adresse, b.Adresse))
	})

	return wege
}

// Fernantwort ist der Stand samt Fehler.
type Fernantwort struct {
	Stand  Fernzustand `json:"stand"`
	Fehler *Fehler     `json:"fehler"`
}

// ohneEinladung ist der Stand, wenn es keinen gibt.
func ohneEinladung() Fernzustand { return Fernzustand{Wege: []FernwegDTO{}} }
