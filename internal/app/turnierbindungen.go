package app

// **Die Turnierbindungen** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, damit `turnier.go` bei den DTOs bleibt. Der Schnitt
// folgt der Frage: dort steht, **welche Gestalt** eine Antwort hat, hier,
// **wie sie entsteht**.

import (
	"context"
	"math"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// TurniereAntwort trägt eine Liste.
type TurniereAntwort struct {
	Turniere []TurnierDTO `json:"turniere"`
	Fehler   *Fehler      `json:"fehler"`
}

// TurnierAntwort trägt ein einzelnes Turnier.
type TurnierAntwort struct {
	Turnier *TurnierDTO `json:"turnier"`
	Fehler  *Fehler     `json:"fehler"`
}

// TeilnehmerAntwort trägt einen Teilnehmer.
type TeilnehmerAntwort struct {
	Teilnehmer *TeilnehmerDTO `json:"teilnehmer"`
	Fehler     *Fehler        `json:"fehler"`
}

// RundeAntwort trägt eine Runde samt ihren Paarungen.
type RundeAntwort struct {
	Runde  *RundeDTO `json:"runde"`
	Fehler *Fehler   `json:"fehler"`
}

// PaarungenAntwort trägt einen Vorschlag — **ohne dass er gelegt wäre**.
type PaarungenAntwort struct {
	Paarungen []PaarungDTO `json:"paarungen"`
	Fehler    *Fehler      `json:"fehler"`
}

// RanglisteAntwort trägt den Stand eines Turniers.
type RanglisteAntwort struct {
	Rangliste []RangDTO `json:"rangliste"`
	Fehler    *Fehler   `json:"fehler"`
}

// MitTurniernaht setzt die Naht — für Tests und den Composition Root.
func (a *App) MitTurniernaht(naht Turnierleitung) { a.turniere = naht }

// Turniere gibt die Turniere eines Spiels — ein leeres Spiel heißt: alle.
func (a *App) Turniere(ctx context.Context, spiel string) TurniereAntwort {
	ctx = a.imProfil(ctx)

	turniere, err := a.turniere.Turniere(ctx, spiel)
	if err != nil {
		return TurniereAntwort{Turniere: []TurnierDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	alle := make([]TurnierDTO, 0, len(turniere))
	for _, eines := range turniere {
		alle = append(alle, alsTurnier(eines))
	}

	return TurniereAntwort{Turniere: alle}
}

// Turnier gibt ein Turnier samt Teilnehmern und Verlauf.
func (a *App) Turnier(ctx context.Context, turnierID string) TurnierAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("turnierID", turnierID)
	if befund != nil {
		return TurnierAntwort{Fehler: befund}
	}

	turnier, err := a.turniere.Turnier(ctx, kennung)
	if err != nil {
		return TurnierAntwort{Fehler: Uebersetze(ctx, err)}
	}

	gelesen := alsTurnier(turnier)

	return TurnierAntwort{Turnier: &gelesen}
}

// TurnierAnlegen legt ein Turnier an — `formatID` leer heißt: keines.
func (a *App) TurnierAnlegen(
	ctx context.Context, name, spiel, notiz string, einzelspiel bool, formatID string,
) TurnierAntwort {
	ctx = a.imProfil(ctx)

	var format int64

	if formatID != "" {
		kennung, befund := benannteKennungAus("formatID", formatID)
		if befund != nil {
			return TurnierAntwort{Fehler: befund}
		}

		format = kennung
	}

	angelegt, err := a.turniere.Anlegen(ctx, domain.Turnier{
		Name: name, Spiel: spiel, Notiz: notiz, Einzelspiel: einzelspiel, FormatID: format,
	})
	if err != nil {
		return TurnierAntwort{Fehler: Uebersetze(ctx, err)}
	}

	gelesen := alsTurnier(angelegt)

	return TurnierAntwort{Turnier: &gelesen}
}

// TurnierLoeschen nimmt ein Turnier samt seinem Verlauf heraus.
func (a *App) TurnierLoeschen(ctx context.Context, turnierID string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("turnierID", turnierID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.turniere.Loeschen(ctx, kennung))}
}

// TurnierAnmelden trägt einen Teilnehmer ein — samt dem Abzug seines Decks.
func (a *App) TurnierAnmelden(
	ctx context.Context, turnierID, name, deckID string,
) TeilnehmerAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("turnierID", turnierID)
	if befund != nil {
		return TeilnehmerAntwort{Fehler: befund}
	}

	teilnehmer := domain.Teilnehmer{Name: name}

	if deckID != "" {
		deck, deckbefund := benannteKennungAus("deckID", deckID)
		if deckbefund != nil {
			return TeilnehmerAntwort{Fehler: deckbefund}
		}

		teilnehmer.DeckID = deck
	}

	angemeldet, err := a.turniere.Anmelden(ctx, kennung, teilnehmer)
	if err != nil {
		return TeilnehmerAntwort{Fehler: Uebersetze(ctx, err)}
	}

	gelesen := alsTeilnehmer(angemeldet)

	return TeilnehmerAntwort{Teilnehmer: &gelesen}
}

// TurnierAussteigen meldet einen Teilnehmer ab — oder wieder an.
func (a *App) TurnierAussteigen(
	ctx context.Context, turnierID, teilnehmerID string, ausgestiegen bool,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	turnier, befund := benannteKennungAus("turnierID", turnierID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	kennung, teilnehmerbefund := benannteKennungAus("teilnehmerID", teilnehmerID)
	if teilnehmerbefund != nil {
		return SchreibAntwort{Fehler: teilnehmerbefund}
	}

	return SchreibAntwort{
		Fehler: Uebersetze(ctx,
			a.turniere.Aussteigen(ctx, turnier, kennung, ausgestiegen)),
	}
}

// TurnierPaarungsvorschlag zeigt, wie die nächste Runde aussähe.
//
// **Ohne sie zu legen** — ein Turnierleiter soll hinsehen können, bevor er
// Tatsachen schafft.
func (a *App) TurnierPaarungsvorschlag(
	ctx context.Context, turnierID string,
) PaarungenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("turnierID", turnierID)
	if befund != nil {
		return PaarungenAntwort{Paarungen: []PaarungDTO{}, Fehler: befund}
	}

	plan, err := a.turniere.Paarungsvorschlag(ctx, kennung)
	if err != nil {
		return PaarungenAntwort{Paarungen: []PaarungDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	return PaarungenAntwort{Paarungen: alsPaarungen(plan)}
}

// TurnierRundeLegen paart die nächste Runde und legt sie an.
func (a *App) TurnierRundeLegen(ctx context.Context, turnierID string) RundeAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("turnierID", turnierID)
	if befund != nil {
		return RundeAntwort{Fehler: befund}
	}

	runde, err := a.turniere.RundeLegen(ctx, kennung)
	if err != nil {
		return RundeAntwort{Fehler: Uebersetze(ctx, err)}
	}

	gelegt := alsRunde(runde)

	return RundeAntwort{Runde: &gelegt}
}

// TurnierErgebnisSetzen trägt ein Ergebnis ein.
//
// **`nil` heißt: doch nicht gespielt.** Berichtigen ist am Turniertisch der
// Normalfall.
//
// **Die Fassung reist mit** (§7, Regel 2). Zwei Leute am selben Tisch, beide
// haben die Runde offen: wer als Zweiter einträgt, überschriebe sonst die
// Berichtigung des Ersten (Befund 7 des Nachlaufs über Phase 4.7).
func (a *App) TurnierErgebnisSetzen(
	ctx context.Context, turnierID, fassung, paarungID string, stand *ErgebnisDTO,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("turnierID", turnierID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	stufe, stufenbefund := benannteKennungAus("fassung", fassung)
	if stufenbefund != nil {
		return SchreibAntwort{Fehler: stufenbefund}
	}

	paarung, paarungsbefund := benannteKennungAus("paarungID", paarungID)
	if paarungsbefund != nil {
		return SchreibAntwort{Fehler: paarungsbefund}
	}

	var ergebnis *domain.Ergebnis

	if stand != nil {
		ergebnis = &domain.Ergebnis{
			SpieleA:       stand.SpieleA,
			SpieleB:       stand.SpieleB,
			Unentschieden: stand.Unentschieden,
		}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx,
		a.turniere.ErgebnisSetzen(ctx, kennung, stufe, paarung, ergebnis))}
}

// TurnierRangliste rechnet den Stand eines Turniers.
func (a *App) TurnierRangliste(ctx context.Context, turnierID string) RanglisteAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("turnierID", turnierID)
	if befund != nil {
		return RanglisteAntwort{Rangliste: []RangDTO{}, Fehler: befund}
	}

	rangliste, err := a.turniere.Rangliste(ctx, kennung)
	if err != nil {
		return RanglisteAntwort{Rangliste: []RangDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	alle := make([]RangDTO, 0, len(rangliste))
	for _, rang := range rangliste {
		alle = append(alle, alsRang(rang))
	}

	return RanglisteAntwort{Rangliste: alle}
}

// quotenstellen ist, auf wie viele Nachkommastellen eine Quote gerundet wird.
//
// **Vier** (Vermerk 0062, Entscheidung 6). Die MTR zeigt drei — sie sagt aber
// nicht, wann gerundet wird, und zwei Teilnehmer, die sich in der vierten
// Stelle unterscheiden, sind nicht gleichauf.
const quotenstellen = 10000

// gerundet macht aus einer gerechneten Quote eine angezeigte.
//
// **Erst hier**, nicht in `domain`: dort wird gerechnet, und eine gerundete
// Zwischenzahl machte die Rangfolge von der Anzeige abhängig.
func gerundet(quote float64) float64 {
	return math.Round(quote*quotenstellen) / quotenstellen
}
