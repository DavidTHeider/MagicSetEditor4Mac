package service

// **Das Ergebnis kommt aus der Partie** — Werkstück M17 der Regelmaschine
// (#1618).
//
// # Was sich damit ändert
//
// **G7 hat den Weg gebaut und niemand ging ihn von selbst.** Eine
// [domain.Matchmeldung] verlangt zwei Angaben, und beide tippte bis hierher
// ein Mensch: die Spielstände und den **Kettenabdruck** — eine Zeichenkette
// aus vierundsechzig Zeichen, die im Fenster in einem Textfeld stand. Wer sie
// abschreibt, schreibt sie falsch ab, und dann sagt die Übernahme „ihr habt
// nicht dasselbe gespielt".
//
// **Die Übereinstimmung der zwei Seiten bleibt.** Was hier entsteht, ist eine
// Meldung **je Seite** aus demselben Zustand — nicht eine Übernahme ohne
// Zustimmung. Wer gegen den Tisch meldet, bekommt dasselbe Ergebnis, weil er
// dieselbe Partie liest; wer eine andere Partie liest, bekommt eine andere
// Kette, und genau das soll die Bindung zeigen.
//
// # Eine Partie ist ein Spiel
//
// **Und damit ist eine Meldung eins zu null, null zu eins oder unentschieden
// zu eins.** Ein Match über drei Spiele (`Hoechstens`) braucht drei Partien
// oder einen Neustart nach CR 727 — jenes hängt an einer Karte (Karn
// Liberated) und damit an Effekten, die es hier nicht gibt. **Das ist die
// Grenze und keine Auslassung** (§10): wer heute ein Match über mehrere
// Spiele führt, bindet die Paarung an das Spiel, das zählt, und trägt den
// Rest von Hand ein.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// AusDerPartie meldet für eine Seite das Ergebnis, das die Partie festgestellt
// hat.
//
// **Sie übernimmt nicht** — dasselbe wie bei [Matchleitung.Melden]: was aus
// zwei Meldungen wird, entscheidet [Matchleitung.Uebernehmen].
func (l *Matchleitung) AusDerPartie(
	ctx context.Context, paarungID int64, von string,
) error {
	bindung, fehler := l.match.Bindung(ctx, paarungID)
	if fehler != nil {
		return fehler
	}

	partie, fehler := l.partien.Partie(ctx, bindung.PartieID)
	if fehler != nil {
		return fehler
	}

	if partie.Stand != domain.StandBeendet {
		return fmt.Errorf(
			"%w: die Partie %q läuft noch — ein Ergebnis hat sie erst, wenn "+
				"sie zu Ende ist", domain.ErrValidation, bindung.PartieID)
	}

	turnier, fehler := l.turniere.Turnier(ctx, bindung.TurnierID)
	if fehler != nil {
		return fehler
	}

	paarung, fehler := paarungAus(turnier, paarungID)
	if fehler != nil {
		return fehler
	}

	ergebnis, fehler := ergebnisAus(paarung, bindung, partie.Ausgang)
	if fehler != nil {
		return fehler
	}

	kette, _, fehler := l.partien.LetzteKette(ctx, bindung.PartieID)
	if fehler != nil {
		return fehler
	}

	return l.Melden(ctx, domain.Matchmeldung{
		PaarungID: paarungID, Fassung: bindung.Fassung,
		PartieID: bindung.PartieID, Von: von, Kette: kette,
		Ergebnis: ergebnis,
	})
}

// ergebnisAus macht aus dem Ausgang einer Partie das Ergebnis dieser Paarung.
//
// # Warum die Paarung dazugehört
//
// **Weil „A" und „B" der Paarung gehören und nicht der Partie.**
// [Matchbindung.Sitze] steht in der Reihenfolge, in der der Aufrufer sie
// genannt hat ([sitzeAus]); wer daraus die Seiten läse, drehte das Ergebnis
// um, sobald jemand die Zuordnung andersherum schreibt.
func ergebnisAus(
	paarung domain.Paarung, bindung domain.Matchbindung,
	ausgang domain.Partieausgang,
) (domain.Ergebnis, error) {
	// **CR 104.4a** (Fassung 2026-08-19): verlieren alle Verbliebenen
	// gleichzeitig, ist die Partie unentschieden — und zwar als eigener
	// Ausgang und nicht als „niemand hat gewonnen".
	if ausgang.Unentschieden {
		return domain.Ergebnis{Unentschieden: 1}, nil
	}

	if ausgang.Sieger == "" {
		return domain.Ergebnis{}, fmt.Errorf(
			"%w: die Partie ist beendet, nennt aber weder Sieger noch "+
				"Unentschieden", domain.ErrValidation)
	}

	for _, sitz := range bindung.Sitze {
		if sitz.Sitz != ausgang.Sieger {
			continue
		}

		if sitz.TeilnehmerID == paarung.A {
			return domain.Ergebnis{SpieleA: 1}, nil
		}

		return domain.Ergebnis{SpieleB: 1}, nil
	}

	return domain.Ergebnis{}, fmt.Errorf(
		"%w: %q hat gewonnen und sitzt nicht an dieser Paarung",
		domain.ErrValidation, ausgang.Sieger)
}
