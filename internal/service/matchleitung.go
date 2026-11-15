package service

// **Eine Turnierpaarung an eine Partie binden** — Werkstück G7 von Phase 5.5
// (#1464).
//
// # Warum das die bestehende Turnierleitung erweitert und keine neue ist
//
// **Weil das Issue es verlangt**: „Bestehende Turnierleitung erweitern, keine
// neue Turnierverwaltung." Was hier dazukommt, ist die **Brücke** zwischen
// einer Paarung und einer Partie — Anmeldung, Runden, Rangliste und
// Ergebniseintrag bleiben, wo sie sind, und ein Turnier ohne Partien ändert
// sich um nichts.
//
// # Und warum das Ergebnis nicht einfach eingetragen wird
//
// **Weil eine Meldung keine Zustimmung ist.** Wer am Küchentisch ein Ergebnis
// einträgt, ist die Turnierleitung — ein Mensch, der beide gefragt hat. Wer
// über eine Leitung meldet, ist **eine Seite**. Deshalb braucht die Übernahme
// beide, und sie schreibt erst, wenn beide dasselbe sagen.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Matchablage hält Bindungen und Meldungen.
type Matchablage interface {
	BindungSchreiben(ctx context.Context, bindung domain.Matchbindung) error
	Bindung(ctx context.Context, paarungID int64) (domain.Matchbindung, error)

	MeldungSchreiben(ctx context.Context, meldung domain.Matchmeldung) error
	Meldungen(ctx context.Context, paarungID int64) (
		[]domain.Matchmeldung, error)
}

// Partiestand ist, was die Matchleitung über eine Partie braucht.
//
// **Ein Verfahren** (§2): sie liest eine Partie und schreibt keine.
type Partiestand interface {
	Partie(ctx context.Context, kennung string) (domain.Partie, error)

	// LetzteKette gibt den Abdruck des letzten Ereignisses dieser Partie.
	//
	// **Dieselbe Zusage wie an [Partiespeicher.LetzteKette]** — und
	// derselbe Aufrufer erfüllt beide. Sie steht hier ein zweites Mal, weil
	// eine Schnittstelle sagt, was **dieser** Verbraucher braucht (§2): die
	// Matchleitung liest eine Partie und ihre Kette, und sie schreibt
	// keine.
	LetzteKette(ctx context.Context, partieID string) (string, int64, error)
}

// Sitzzuordnung ist, wer an welchem Sitz für welchen Teilnehmer spielt.
//
// **Drei Angaben, und die dritte ist der Punkt des Issues**: „Teilnehmername,
// lokales Teilnehmerobjekt, Peeridentität und Sitz ausdrücklich zuordnen; ein
// bisheriger Name ist kein authentifizierter Spieler."
type Sitzzuordnung struct {
	TeilnehmerID int64
	Sitz         string

	// Peer ist die Peerkennung — leer heißt: an diesem Rechner.
	Peer string
}

// Matchleitung bindet Paarungen an Partien und übernimmt Ergebnisse.
type Matchleitung struct {
	turniere Turnierspeicher
	partien  Partiestand
	match    Matchablage
	freigabe *Regelfreigabe
}

// NeueMatchleitung baut sie.
func NeueMatchleitung(
	turniere Turnierspeicher, partien Partiestand, match Matchablage,
) *Matchleitung {
	return &Matchleitung{
		turniere: turniere, partien: partien, match: match,
		freigabe: NeueRegelfreigabe(),
	}
}

// Binden sagt, dass diese Partie zu dieser Paarung gehört.
//
// # Was hier alles geprüft wird, bevor etwas liegen bleibt
//
//  1. **Die Paarung gibt es, und sie gehört zu diesem Turnier.**
//  2. **Sie ist kein Bye** — „Bye startet keine Zweispielersession."
//  3. **Sie hat noch kein Ergebnis** — was entschieden ist, wird nicht
//     nachgespielt.
//  4. **Die beiden Teilnehmer sind die der Paarung** — wer sonst antritt,
//     spielt ein anderes Match.
//  5. **Die Sitze gibt es in der Partie**, und ihr **eingefrorener Abzug**
//     wird aus ihr genommen. Ein Deckabdruck, den der Aufrufer mitbrächte,
//     wäre eine Behauptung; dieser hier ist der, mit dem wirklich gespielt
//     wird.
//  6. **Die Strenge der Partie wird von der Matrix getragen** — die
//     Unterstützungsmatrix aus G4 bestimmt den zulässigen Modus, und geprüft
//     wird sie von [Regelfreigabe], nicht ein zweites Mal hier (§10).
func (l *Matchleitung) Binden(
	ctx context.Context, turnierID, paarungID int64, partieID string,
	zuordnung []Sitzzuordnung,
) (domain.Matchbindung, error) {
	turnier, turnierfehler := l.turniere.Turnier(ctx, turnierID)
	if turnierfehler != nil {
		return domain.Matchbindung{}, turnierfehler
	}

	paarung, paarungsfehler := paarungAus(turnier, paarungID)
	if paarungsfehler != nil {
		return domain.Matchbindung{}, paarungsfehler
	}

	if fehler := bindbar(paarung); fehler != nil {
		return domain.Matchbindung{}, fehler
	}

	partie, partiefehler := l.partien.Partie(ctx, partieID)
	if partiefehler != nil {
		return domain.Matchbindung{}, partiefehler
	}

	if fehler := l.traegtDieStrenge(partie); fehler != nil {
		return domain.Matchbindung{}, fehler
	}

	sitze, sitzfehler := sitzeAus(paarung, partie, zuordnung)
	if sitzfehler != nil {
		return domain.Matchbindung{}, sitzfehler
	}

	bindung := domain.Matchbindung{
		TurnierID: turnierID, PaarungID: paarungID,
		Fassung: turnier.Fassung, PartieID: partieID,
		Sitze: sitze, Hoechstens: turnier.HoechstensSpiele(),
	}

	if fehler := bindung.Pruefe(); fehler != nil {
		return domain.Matchbindung{}, fehler
	}

	if fehler := l.match.BindungSchreiben(ctx, bindung); fehler != nil {
		return domain.Matchbindung{}, fehler
	}

	return bindung, nil
}

// traegtDieStrenge hält die Strenge der Partie gegen ihre Matrix.
//
// **Gegen die der Partie und nicht gegen die von heute**: eine Partie friert
// ihr Profil beim Anfang ein (G4), und eine Bindung, die die heutige Matrix
// befragte, könnte eine laufende Begegnung nachträglich verbieten.
func (l *Matchleitung) traegtDieStrenge(partie domain.Partie) error {
	freigabe, fehler := l.freigabe.Pruefe(Freigabewunsch{
		Profil:  partie.Profil,
		Strenge: partie.Regelwerk.Strenge,
		Sitze:   len(partie.Sitze),
	})
	if fehler != nil {
		return fehler
	}

	if !freigabe.Erlaubt {
		return fmt.Errorf(
			"%w: diese Partie trägt die Strenge %q nicht: %s",
			domain.ErrValidation, partie.Regelwerk.Strenge,
			freigabe.Hindernisse[0])
	}

	return nil
}

// Melden trägt das Ergebnis einer Seite ein.
//
// **Es wird nicht übernommen.** Eine Meldung ist eine Aussage; was daraus
// wird, entscheidet [Matchleitung.Uebernehmen] — und die verlangt beide
// Seiten.
func (l *Matchleitung) Melden(
	ctx context.Context, meldung domain.Matchmeldung,
) error {
	bindung, fehler := l.match.Bindung(ctx, meldung.PaarungID)
	if fehler != nil {
		return fehler
	}

	if fehler := bindung.Nimmt(meldung); fehler != nil {
		return fehler
	}

	return l.match.MeldungSchreiben(ctx, meldung)
}

// Bindung gibt die Bindung einer Paarung.
func (l *Matchleitung) Bindung(
	ctx context.Context, paarungID int64,
) (domain.Matchbindung, error) {
	return l.match.Bindung(ctx, paarungID)
}

// Meldungen gibt, was die Seiten gemeldet haben.
func (l *Matchleitung) Meldungen(
	ctx context.Context, paarungID int64,
) ([]domain.Matchmeldung, error) {
	return l.match.Meldungen(ctx, paarungID)
}

// Uebernehmen trägt das Ergebnis ein — **einmal**.
//
// # Die vier Ausgänge
//
//   - **Eingetragen**: beide Seiten meldeten dasselbe, und die Paarung hatte
//     noch kein Ergebnis.
//   - **Schon da** (kein Fehler): dasselbe Ergebnis steht schon — eine
//     zweite Meldung ändert nichts. „Doppelte Meldungen idempotent."
//   - **Streit** ([domain.ErrConflict]): die Seiten melden Verschiedenes, die
//     Paarung hat sich geändert, oder es steht **ein anderes** Ergebnis da.
//     „Konflikt statt Überschreiben."
//   - **Noch nicht**: es fehlt eine Seite.
//
// **Und die Übernahme geschieht auf Zuruf.** Das Issue verlangt es:
// „Übernahme nach expliziter Turnierverantwortlichen-Freigabe." Wer zwei
// übereinstimmende Meldungen automatisch einträge, hätte die Freigabe
// abgeschafft.
func (l *Matchleitung) Uebernehmen(
	ctx context.Context, turnierID, paarungID int64,
) (domain.Ergebnis, error) {
	bindung, fehler := l.match.Bindung(ctx, paarungID)
	if fehler != nil {
		return domain.Ergebnis{}, fehler
	}

	meldungen, fehler := l.match.Meldungen(ctx, paarungID)
	if fehler != nil {
		return domain.Ergebnis{}, fehler
	}

	stand, fehler := bindung.Uebernahme(meldungen)
	if fehler != nil {
		return domain.Ergebnis{}, fehler
	}

	turnier, fehler := l.turniere.Turnier(ctx, turnierID)
	if fehler != nil {
		return domain.Ergebnis{}, fehler
	}

	paarung, fehler := paarungAus(turnier, paarungID)
	if fehler != nil {
		return domain.Ergebnis{}, fehler
	}

	// **Das vorhandene Ergebnis zuerst, und das ist keine Reihenfolgefrage.**
	//
	// Das Eintragen selbst **erhöht die Fassung** des Turniers. Wer die
	// Fassungsprüfung davorsetzte, bekäme beim zweiten Ruf einen Konflikt
	// über genau die Änderung, die der erste geschrieben hat — aus „schon
	// erledigt" würde „jemand war schneller". Das Issue verlangt das
	// Gegenteil: „doppelte Meldungen idempotent."
	if paarung.Ergebnis != nil {
		return *paarung.Ergebnis, schonEingetragen(*paarung.Ergebnis, stand)
	}

	if turnier.Fassung != bindung.Fassung {
		return domain.Ergebnis{}, fmt.Errorf(
			"%w: dieses Turnier steht bei Fassung %d, gebunden wurde bei %d "+
				"— die Paarung hat sich inzwischen geändert",
			domain.ErrConflict, turnier.Fassung, bindung.Fassung)
	}

	if fehler := l.turniere.ErgebnisSetzen(
		ctx, turnier.Fassung, paarungID, &stand); fehler != nil {
		return domain.Ergebnis{}, fehler
	}

	return stand, nil
}

// schonEingetragen entscheidet zwischen „nichts zu tun" und „Streit".
//
// **Dasselbe Ergebnis ist kein Fehler.** Wer zweimal übernimmt, hat zweimal
// geklickt; wer dabei einen Fehler bekäme, glaubte, es sei etwas schiefgegangen.
//
// **Ein anderes ist einer.** Jemand hat von Hand eingetragen, oder die Seiten
// haben sich nach der ersten Übernahme umentschieden — beides ist eine
// Entscheidung für einen Menschen und nicht für dieses Programm.
func schonEingetragen(steht, gemeldet domain.Ergebnis) error {
	if steht == gemeldet {
		return nil
	}

	return fmt.Errorf(
		"%w: an dieser Paarung steht schon ein Ergebnis (%d:%d:%d), gemeldet "+
			"ist (%d:%d:%d) — ein vorhandenes wird nicht überschrieben",
		domain.ErrConflict, steht.SpieleA, steht.SpieleB, steht.Unentschieden,
		gemeldet.SpieleA, gemeldet.SpieleB, gemeldet.Unentschieden)
}

// paarungAus sucht die Paarung im Turnier.
func paarungAus(
	turnier domain.Turnier, paarungID int64,
) (domain.Paarung, error) {
	for _, runde := range turnier.Runden {
		for _, paarung := range runde.Paarungen {
			if paarung.ID == paarungID {
				return paarung, nil
			}
		}
	}

	return domain.Paarung{}, fmt.Errorf(
		"%w: die Paarung %d gehört nicht zu diesem Turnier",
		domain.ErrNotFound, paarungID)
}

// bindbar sagt, ob an dieser Paarung überhaupt gespielt wird.
func bindbar(paarung domain.Paarung) error {
	if paarung.IstBye() {
		return fmt.Errorf(
			"%w: ein Bye ist ein Ergebnis und keine Begegnung — es gibt "+
				"niemanden, gegen den gespielt würde", domain.ErrValidation)
	}

	if paarung.Ergebnis != nil {
		return fmt.Errorf(
			"%w: an dieser Paarung steht schon ein Ergebnis",
			domain.ErrConflict)
	}

	return nil
}

// sitzeAus ordnet die Teilnehmer ihren Sitzen zu — und holt den Abzug.
func sitzeAus(
	paarung domain.Paarung, partie domain.Partie, zuordnung []Sitzzuordnung,
) ([]domain.Matchsitz, error) {
	sitze := make([]domain.Matchsitz, 0, len(zuordnung))

	for _, eine := range zuordnung {
		if eine.TeilnehmerID != paarung.A && eine.TeilnehmerID != paarung.B {
			return nil, fmt.Errorf(
				"%w: der Teilnehmer %d gehört nicht zu dieser Paarung",
				domain.ErrValidation, eine.TeilnehmerID)
		}

		sitz := partie.SitzVon(eine.Sitz)
		if sitz == nil {
			return nil, fmt.Errorf(
				"%w: den Sitz %q gibt es in dieser Partie nicht",
				domain.ErrValidation, eine.Sitz)
		}

		sitze = append(sitze, domain.Matchsitz{
			TeilnehmerID: eine.TeilnehmerID, Sitz: eine.Sitz,
			Peer: eine.Peer, Deckabdruck: sitz.Deck.Abdruck,
		})
	}

	return sitze, nil
}
