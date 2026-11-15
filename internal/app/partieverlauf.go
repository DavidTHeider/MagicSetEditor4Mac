package app

// **Der Verlauf einer Partie** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum er eine eigene Datei bekommt
//
// **Weil er eine eigene Frage beantwortet** (§2): `partiewege.go` führt, was
// eine Partie **tut** — anlegen, befehlen, pausieren. Hier steht, was sie
// **war**, und das ist eine Rückschau mit eigenem Anlass: ein Ereignisstrom
// liegt vollständig vor, während am Tisch gar nicht erst ankommt, was nicht
// hingehört.
//
// # Warum dieselbe Gestalt wie beim Befehl
//
// **Weil es dasselbe Ereignis ist.** Der erste Anlauf gab hier eine zweite
// Gestalt aus — `Folge` als Zahl statt als Zeichenkette, `Privat` als Liste
// statt als Karte —, und das Fenster hätte für denselben Gegenstand zwei
// Typen gebraucht. Ein Ereignis sieht an dieser Naht überall gleich aus
// ([EreignisDTO]), und an den anderen Flächen auch (§16).

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// PartieverlaufAntwort trägt den Strom.
//
// **Nicht `VerlaufAntwort`** — die gibt es schon für den Verlauf einer Karte
// (`verlaufsliste.go`), und zwei gleichnamige Antworten an derselben Naht
// wären zwei verschiedene Dinge unter einem Namen.
type PartieverlaufAntwort struct {
	Verlauf []EreignisDTO `json:"verlauf"`

	Fehler *Fehler `json:"fehler"`
}

// PartieVerlauf gibt den Ereignisstrom, wie `fuer` ihn sehen darf.
//
// **Die Sicht ist ein Parameter und keine Annahme.** Wer sie aus dem Profil
// ableitete, ließe einen Zuschauer den Verlauf eines Sitzes sehen, sobald
// beide dasselbe Profil benutzen — und beim Spiel gegen sich selbst ist das
// der Normalfall (G8,
// #1465).
func (a *App) PartieVerlauf(
	ctx context.Context, kennung, fuer string,
) PartieverlaufAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungLesen); fehler != nil {
		return leererVerlauf(ctx, fehler)
	}

	strom, fehler := a.partienaht().Verlauf(ctx, kennung, fuer)
	if fehler != nil {
		return leererVerlauf(ctx, fehler)
	}

	return PartieverlaufAntwort{Verlauf: verlaufsereignisseAus(strom)}
}

// leererVerlauf gibt eine Antwort mit leerer Liste und übersetztem Fehler.
//
// **Eine leere Liste und kein `nil`** — der Fehlerweg trägt dieselbe Zusage
// wie der Erfolgsweg. Genau hier ist sie in Phase 5.4 zweimal gebrochen.
func leererVerlauf(ctx context.Context, fehler error) PartieverlaufAntwort {
	return PartieverlaufAntwort{
		Verlauf: []EreignisDTO{},
		Fehler:  Uebersetze(ctx, fehler),
	}
}

// verlaufsereignisseAus bildet den **schon gefilterten** Strom ab.
//
// **Es filtert nicht noch einmal.** `service` hat `domain.RueckschauFuer`
// gerechnet, bevor der Strom hier ankommt; eine zweite Stelle mit derselben
// Regel liefe beim ersten neuen Feld auseinander. Der Unterschied zu
// [ereignisseAus] ist genau dieser — dort kommt der Strom ungefiltert an.
func verlaufsereignisseAus(strom []domain.Ereignis) []EreignisDTO {
	verlauf := make([]EreignisDTO, 0, len(strom))

	for _, ereignis := range strom {
		dto := EreignisDTO{
			Folge:       strconv.FormatInt(ereignis.Folge, 10),
			Art:         ereignis.Art,
			Oeffentlich: leereKarteStattNil(ereignis.Oeffentlich),
			Kette:       ereignis.Kette,
		}

		// **Höchstens eine, denn sie ist empfängergebunden** — die Rückschau
		// hat nur die eigene stehen lassen.
		if len(ereignis.Privat) > 0 {
			dto.Privat = ereignis.Privat[0].Inhalt
		}

		dto.Privat = leereKarteStattNil(dto.Privat)

		verlauf = append(verlauf, dto)
	}

	return verlauf
}

// leereKarteStattNil macht aus `nil` eine leere Karte.
//
// **Durch die Naht wird `nil` zu `null`**, und eine Oberfläche, die darüber
// liefe, bekäme einen Fehler statt einer leeren Anzeige.
func leereKarteStattNil(inhalt map[string]any) map[string]any {
	if inhalt == nil {
		return map[string]any{}
	}

	return inhalt
}
