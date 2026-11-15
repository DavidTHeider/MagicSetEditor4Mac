package app

// **Eine Turnierpaarung an eine Partie binden** — Werkstück G7 von Phase 5.5
// (#1464).
//
// # Vier Wege, und keiner trägt allein ein Ergebnis
//
// **Binden** sagt, welche Partie zu welcher Paarung gehört. **Melden** ist die
// Aussage **einer** Seite. **Übernehmen** trägt ein — und nur, wenn beide
// dasselbe sagen. **Lesen** zeigt, woran es liegt, wenn noch nichts
// eingetragen ist.
//
// **Der Eintrag von Hand bleibt** (`TurnierErgebnisSetzen`). Wer am
// Küchentisch ein Turnier führt, tippt ein Ergebnis — daran ändert dieses
// Werkstück nichts; es kommt ein zweiter Weg dazu für die, die hier spielen.

import (
	"context"
	"errors"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Matchleitung ist, was die Fassade über Bindungen und Meldungen braucht.
type Matchleitung interface {
	Binden(ctx context.Context, turnierID, paarungID int64, partieID string,
		zuordnung []service.Sitzzuordnung) (domain.Matchbindung, error)
	Melden(ctx context.Context, meldung domain.Matchmeldung) error
	AusDerPartie(ctx context.Context, paarungID int64, von string) error
	Uebernehmen(ctx context.Context, turnierID, paarungID int64) (
		domain.Ergebnis, error)
	Bindung(ctx context.Context, paarungID int64) (domain.Matchbindung, error)
	Meldungen(ctx context.Context, paarungID int64) (
		[]domain.Matchmeldung, error)
}

// MatchsitzDTO ist ein Sitz einer Bindung.
type MatchsitzDTO struct {
	TeilnehmerID string `json:"teilnehmerID"`
	Sitz         string `json:"sitz"`

	// Peer ist die Peerkennung — leer heißt: an diesem Rechner.
	Peer string `json:"peer"`

	// Deckabdruck ist der Hash des eingefrorenen Abzugs.
	Deckabdruck string `json:"deckabdruck"`
}

// MatchmeldungDTO ist, was eine Seite gemeldet hat.
type MatchmeldungDTO struct {
	Von      string      `json:"von"`
	Ergebnis ErgebnisDTO `json:"ergebnis"`

	// Kette ist der letzte Kettenabdruck des Ereignisstroms.
	Kette string `json:"kette"`
}

// MatchAntwort trägt die Bindung einer Paarung und ihre Meldungen.
type MatchAntwort struct {
	// Gebunden sagt, ob an dieser Paarung überhaupt eine Partie hängt.
	Gebunden bool `json:"gebunden"`

	PaarungID string `json:"paarungID"`
	PartieID  string `json:"partieID"`
	Fassung   string `json:"fassung"`

	Sitze     []MatchsitzDTO    `json:"sitze"`
	Meldungen []MatchmeldungDTO `json:"meldungen"`

	// Uebernommen ist das Ergebnis, das übernommen wurde — nur beim
	// Übernehmen.
	Uebernommen *ErgebnisDTO `json:"uebernommen"`

	// Nachladefehler sagt: der Schritt **gilt**, nur der Stand danach war
	// nicht zu lesen (§7, Regel 3; #1970) — dann sind die Listen leer und
	// `fehler` auch.
	Nachladefehler *Fehler `json:"nachladefehler,omitempty"`

	Fehler *Fehler `json:"fehler"`
}

// TurnierMatch gibt die Bindung einer Paarung samt Meldungen.
func (a *App) TurnierMatch(
	ctx context.Context, paarungID string,
) MatchAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungLesen); fehler != nil {
		return leeresMatch(ctx, fehler)
	}

	paarung, befund := benannteKennungAus("paarungID", paarungID)
	if befund != nil {
		return MatchAntwort{
			Sitze: []MatchsitzDTO{}, Meldungen: []MatchmeldungDTO{},
			Fehler: befund,
		}
	}

	bindung, fehler := a.matchnaht().Bindung(ctx, paarung)
	if errors.Is(fehler, domain.ErrNotFound) {
		// **Keine Bindung ist kein Fehler.** Die meisten Paarungen werden
		// von Hand eingetragen; wer hier eine Meldung bekäme, hielte den
		// Normalfall für eine Störung (§6).
		return MatchAntwort{
			PaarungID: paarungID,
			Sitze:     []MatchsitzDTO{}, Meldungen: []MatchmeldungDTO{},
		}
	}

	if fehler != nil {
		return leeresMatch(ctx, fehler)
	}

	meldungen, fehler := a.matchnaht().Meldungen(ctx, paarung)
	if fehler != nil {
		return leeresMatch(ctx, fehler)
	}

	return matchAntwortAus(bindung, meldungen)
}

// matchNachDemSchreiben gibt den Stand nach einem **geglückten** Schritt.
//
// **Geschrieben ist geschrieben** (§7, Regel 3;
// #1970):
// scheitert das Lesen danach, steht sein Fehler in
// [MatchAntwort.Nachladefehler] und nicht in `fehler` — gebunden, gemeldet
// oder übernommen ist trotzdem, und eine Wiederholung liefe gegen diesen
// Stand.
func (a *App) matchNachDemSchreiben(
	ctx context.Context, paarungID string,
) MatchAntwort {
	antwort := a.TurnierMatch(ctx, paarungID)
	antwort.Nachladefehler, antwort.Fehler = antwort.Fehler, nil

	return antwort
}

// TurnierMatchBinden sagt, dass diese Partie zu dieser Paarung gehört.
func (a *App) TurnierMatchBinden(
	ctx context.Context, turnierID, paarungID, partieID string,
	zuordnung []MatchsitzDTO,
) MatchAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return leeresMatch(ctx, fehler)
	}

	turnier, paarung, befund := zweiKennungen(turnierID, paarungID)
	if befund != nil {
		return MatchAntwort{
			Sitze: []MatchsitzDTO{}, Meldungen: []MatchmeldungDTO{},
			Fehler: befund,
		}
	}

	sitze, befund := zuordnungAus(zuordnung)
	if befund != nil {
		return MatchAntwort{
			Sitze: []MatchsitzDTO{}, Meldungen: []MatchmeldungDTO{},
			Fehler: befund,
		}
	}

	if _, fehler := a.matchnaht().Binden(
		ctx, turnier, paarung, partieID, sitze); fehler != nil {
		return leeresMatch(ctx, fehler)
	}

	return a.matchNachDemSchreiben(ctx, paarungID)
}

// TurnierMatchMelden trägt das Ergebnis **einer** Seite ein.
func (a *App) TurnierMatchMelden(
	ctx context.Context, paarungID, von, kette string, stand ErgebnisDTO,
) MatchAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return leeresMatch(ctx, fehler)
	}

	paarung, befund := benannteKennungAus("paarungID", paarungID)
	if befund != nil {
		return MatchAntwort{
			Sitze: []MatchsitzDTO{}, Meldungen: []MatchmeldungDTO{},
			Fehler: befund,
		}
	}

	bindung, fehler := a.matchnaht().Bindung(ctx, paarung)
	if fehler != nil {
		return leeresMatch(ctx, fehler)
	}

	if fehler := a.matchnaht().Melden(ctx, domain.Matchmeldung{
		PaarungID: paarung, Fassung: bindung.Fassung,
		PartieID: bindung.PartieID, Von: von, Kette: kette,
		Ergebnis: domain.Ergebnis{
			SpieleA:       stand.SpieleA,
			SpieleB:       stand.SpieleB,
			Unentschieden: stand.Unentschieden,
		},
	}); fehler != nil {
		return leeresMatch(ctx, fehler)
	}

	return a.matchNachDemSchreiben(ctx, paarungID)
}

// TurnierMatchAusDerPartie meldet für eine Seite das, was die Partie
// festgestellt hat — M17
// (#1618).
//
// **Zwei Angaben weniger als `TurnierMatchMelden`**: Spielstand und
// Kettenabdruck kommen aus der beendeten Partie. Übrig bleibt, **wer** meldet
// — und das ist die Zustimmung, die G7 verlangt.
func (a *App) TurnierMatchAusDerPartie(
	ctx context.Context, paarungID, von string,
) MatchAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return leeresMatch(ctx, fehler)
	}

	paarung, befund := benannteKennungAus("paarungID", paarungID)
	if befund != nil {
		return MatchAntwort{
			Sitze: []MatchsitzDTO{}, Meldungen: []MatchmeldungDTO{},
			Fehler: befund,
		}
	}

	if fehler := a.matchnaht().AusDerPartie(ctx, paarung, von); fehler != nil {
		return leeresMatch(ctx, fehler)
	}

	return a.matchNachDemSchreiben(ctx, paarungID)
}

// TurnierMatchUebernehmen trägt das Ergebnis ins Turnier ein — **einmal**.
func (a *App) TurnierMatchUebernehmen(
	ctx context.Context, turnierID, paarungID string,
) MatchAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return leeresMatch(ctx, fehler)
	}

	turnier, paarung, befund := zweiKennungen(turnierID, paarungID)
	if befund != nil {
		return MatchAntwort{
			Sitze: []MatchsitzDTO{}, Meldungen: []MatchmeldungDTO{},
			Fehler: befund,
		}
	}

	stand, fehler := a.matchnaht().Uebernehmen(ctx, turnier, paarung)
	if fehler != nil {
		return leeresMatch(ctx, fehler)
	}

	antwort := a.matchNachDemSchreiben(ctx, paarungID)
	antwort.Uebernommen = &ErgebnisDTO{
		SpieleA:       stand.SpieleA,
		SpieleB:       stand.SpieleB,
		Unentschieden: stand.Unentschieden,
	}

	return antwort
}

// zweiKennungen deutet Turnier und Paarung.
func zweiKennungen(turnierID, paarungID string) (
	turnier, paarung int64, befund *Fehler,
) {
	turnier, befund = benannteKennungAus("turnierID", turnierID)
	if befund != nil {
		return 0, 0, befund
	}

	paarung, befund = benannteKennungAus("paarungID", paarungID)
	if befund != nil {
		return 0, 0, befund
	}

	return turnier, paarung, nil
}

// zuordnungAus deutet die Teilnehmerkennungen.
func zuordnungAus(
	zuordnung []MatchsitzDTO,
) ([]service.Sitzzuordnung, *Fehler) {
	sitze := make([]service.Sitzzuordnung, 0, len(zuordnung))

	for _, eine := range zuordnung {
		teilnehmer, befund := benannteKennungAus(
			"teilnehmerID", eine.TeilnehmerID)
		if befund != nil {
			return nil, befund
		}

		sitze = append(sitze, service.Sitzzuordnung{
			TeilnehmerID: teilnehmer, Sitz: eine.Sitz, Peer: eine.Peer,
		})
	}

	return sitze, nil
}

// matchAntwortAus bildet Bindung und Meldungen ab.
func matchAntwortAus(
	bindung domain.Matchbindung, meldungen []domain.Matchmeldung,
) MatchAntwort {
	sitze := make([]MatchsitzDTO, 0, len(bindung.Sitze))
	for _, sitz := range bindung.Sitze {
		sitze = append(sitze, MatchsitzDTO{
			TeilnehmerID: strconv.FormatInt(sitz.TeilnehmerID, 10),
			Sitz:         sitz.Sitz,
			Peer:         sitz.Peer,
			Deckabdruck:  sitz.Deckabdruck,
		})
	}

	gemeldet := make([]MatchmeldungDTO, 0, len(meldungen))
	for _, meldung := range meldungen {
		gemeldet = append(gemeldet, MatchmeldungDTO{
			Von:   meldung.Von,
			Kette: meldung.Kette,
			Ergebnis: ErgebnisDTO{
				SpieleA:       meldung.Ergebnis.SpieleA,
				SpieleB:       meldung.Ergebnis.SpieleB,
				Unentschieden: meldung.Ergebnis.Unentschieden,
			},
		})
	}

	return MatchAntwort{
		Gebunden:  true,
		PaarungID: strconv.FormatInt(bindung.PaarungID, 10),
		PartieID:  bindung.PartieID,
		Fassung:   strconv.FormatInt(bindung.Fassung, 10),
		Sitze:     sitze,
		Meldungen: gemeldet,
	}
}

// leeresMatch gibt eine Antwort mit leeren Listen und übersetztem Fehler.
func leeresMatch(ctx context.Context, fehler error) MatchAntwort {
	return MatchAntwort{
		Sitze: []MatchsitzDTO{}, Meldungen: []MatchmeldungDTO{},
		Fehler: Uebersetze(ctx, fehler),
	}
}
