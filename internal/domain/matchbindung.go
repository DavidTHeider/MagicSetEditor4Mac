package domain

// **Eine Turnierpaarung, an eine Partie gebunden** — Werkstück G7 von Phase
// 5.5 (#1464).
//
// # Wofür es sie gibt
//
// **Damit ein Ergebnis nicht behauptet, sondern belegt wird.** Bisher trägt
// jemand einen Spielstand ein, und das Turnier glaubt ihn — das ist am
// Küchentisch richtig und bleibt es. Wer **hier** spielt, hat einen
// Ereignisstrom mit einer Kette; die Bindung sagt, **welche** Partie zu
// **welcher** Paarung gehört, und die Übernahme verlangt, dass beide Seiten
// dasselbe melden.
//
// # Was daraus ausdrücklich nicht folgt
//
// **Keine Turnierzulassung und kein Schutz gegen Betrug.** Das Issue grenzt es
// selbst ab: „Offizielle Turnierzulassung, vollständiges Strict-Regelwerk und
// manipulationssichere kompetitive Fernturniere sind keine aus Signaturen
// ableitbaren Zusagen." Was hier entsteht, ist die Übereinstimmung zweier
// Meldungen über denselben Strom — nicht mehr.
//
// # Und warum die Fassung mitgeht
//
// **Weil eine Paarung sich ändern darf.** Wer eine Runde neu legt, ändert die
// Fassung des Turniers; eine Meldung, die zu einer alten gehört, gehört zu
// einer Paarung, die es so nicht mehr gibt. Sie wird abgewiesen, statt ein
// Ergebnis auf eine andere Begegnung zu schreiben.

import "fmt"

// Matchsitz ordnet einen Teilnehmer seinem Sitz zu.
//
// **Vier Angaben, und drei davon sind neu gegenüber dem Turnier.** Ein
// [Teilnehmer] ist ein **Name** — das Issue sagt es scharf: „ein bisheriger
// Name ist kein authentifizierter Spieler." Hier kommt dazu, **wo** er sitzt,
// **wer** er im Netz ist und **womit** er antritt.
type Matchsitz struct {
	TeilnehmerID int64

	// Sitz ist die Kennung am Tisch ([Sitz.Kennung]).
	Sitz string

	// Peer ist die Peerkennung — **leer heißt: an diesem Rechner**.
	//
	// **Und das ist der Normalfall am Küchentisch.** Wer neben einem sitzt,
	// hat keine Peerkennung in dieser Partie; ein Feld, das eine verlangte,
	// machte aus dem häufigsten Fall den Ausnahmefall.
	Peer string

	// Deckabdruck ist der Hash des eingefrorenen Abzugs
	// ([Partiedeck.Abdruck]).
	//
	// **Er bindet das Deck an die Begegnung.** Wer mit einem anderen antritt,
	// hat einen anderen Abdruck — und die Bindung passt nicht mehr.
	Deckabdruck string
}

// Matchbindung sagt, welche Partie zu welcher Paarung gehört.
type Matchbindung struct {
	TurnierID int64
	PaarungID int64

	// Fassung ist die des Turniers beim Binden — der Grund steht im Kopf.
	Fassung int64

	PartieID string

	// Sitze sind **genau zwei** — der Grund steht bei [Matchbindung.Pruefe].
	Sitze []Matchsitz

	// Hoechstens ist die Zahl der Spiele; null heißt: keine Grenze.
	Hoechstens int
}

// Pruefe sagt, ob eine Bindung eine ist.
//
// # Warum genau zwei Sitze
//
// **Weil eine Paarung zwei Teilnehmer hat.** Ein Bye hat einen und ist kein
// Match — das Issue verlangt es wörtlich: „Bye startet keine
// Zweispielersession." Drei wären ein Tisch und keine Paarung; wer sie hier
// zuließe, hätte eine Übernahme, die nicht weiß, wessen Zustimmung sie
// braucht.
func (b Matchbindung) Pruefe() error {
	switch {
	case b.TurnierID == 0 || b.PaarungID == 0:
		return fmt.Errorf(
			"%w: eine Bindung braucht Turnier und Paarung", ErrValidation)

	case b.Fassung == 0:
		return fmt.Errorf(
			"%w: eine Bindung ohne Fassung bindet an nichts — eine Paarung "+
				"darf sich ändern", ErrValidation)

	case b.PartieID == "":
		return fmt.Errorf(
			"%w: eine Bindung braucht die Partie, an die sie bindet",
			ErrValidation)

	case len(b.Sitze) != 2:
		return fmt.Errorf(
			"%w: eine Paarung hat zwei Seiten, hier sind es %d — ein Bye ist "+
				"kein Match", ErrValidation, len(b.Sitze))
	}

	return b.pruefeSitze()
}

// pruefeSitze hält die beiden Sitze gegeneinander.
func (b Matchbindung) pruefeSitze() error {
	gesehen := map[string]bool{}
	teilnehmer := map[int64]bool{}

	for _, sitz := range b.Sitze {
		switch {
		case sitz.TeilnehmerID == 0 || sitz.Sitz == "":
			return fmt.Errorf(
				"%w: ein Sitz braucht seinen Teilnehmer und seine Kennung",
				ErrValidation)

		case sitz.Deckabdruck == "":
			return fmt.Errorf(
				"%w: der Sitz %q tritt ohne eingefrorenes Deck an",
				ErrValidation, sitz.Sitz)

		case gesehen[sitz.Sitz]:
			return fmt.Errorf("%w: der Sitz %q steht zweimal",
				ErrValidation, sitz.Sitz)

		case teilnehmer[sitz.TeilnehmerID]:
			return fmt.Errorf("%w: der Teilnehmer %d sitzt zweimal",
				ErrValidation, sitz.TeilnehmerID)
		}

		gesehen[sitz.Sitz] = true
		teilnehmer[sitz.TeilnehmerID] = true
	}

	return nil
}

// Matchmeldung ist ein gemeldetes Ergebnis samt Nachweis.
type Matchmeldung struct {
	PaarungID int64
	Fassung   int64
	PartieID  string

	Ergebnis Ergebnis

	// Kette ist der letzte Kettenabdruck des Ereignisstroms.
	//
	// # Warum er der Nachweis ist
	//
	// **Weil er den ganzen Verlauf zusammenfasst.** Zwei Seiten, die dieselbe
	// Kette melden, haben denselben Strom gesehen; eine, die eine andere
	// meldet, hat etwas anderes gespielt — und das ist ein Streit und kein
	// Ergebnis.
	//
	// **Was er nicht ist**: ein Beweis, dass richtig gespielt wurde. Er sagt,
	// dass beide dasselbe **meinen**.
	Kette string

	// Von ist der Sitz, der meldet.
	Von string
}

// Nimmt sagt, ob diese Bindung diese Meldung annimmt.
func (b Matchbindung) Nimmt(meldung Matchmeldung) error {
	switch {
	case meldung.PaarungID != b.PaarungID:
		return fmt.Errorf("%w: diese Meldung gehört zur Paarung %d",
			ErrValidation, meldung.PaarungID)

	case meldung.Fassung != b.Fassung:
		return fmt.Errorf(
			"%w: diese Meldung gehört zur Fassung %d, gebunden ist %d — die "+
				"Paarung hat sich geändert", ErrConflict,
			meldung.Fassung, b.Fassung)

	case meldung.PartieID != b.PartieID:
		return fmt.Errorf("%w: diese Meldung gehört zur Partie %q",
			ErrValidation, meldung.PartieID)

	case meldung.Kette == "":
		return fmt.Errorf(
			"%w: eine Meldung ohne Nachweis ist eine Behauptung",
			ErrValidation)

	case !b.sitzt(meldung.Von):
		return fmt.Errorf("%w: %q sitzt nicht an dieser Paarung",
			ErrValidation, meldung.Von)
	}

	return meldung.Ergebnis.Pruefe(b.Hoechstens)
}

// sitzt sagt, ob diese Sitzkennung zu dieser Bindung gehört.
func (b Matchbindung) sitzt(wer string) bool {
	for _, sitz := range b.Sitze {
		if sitz.Sitz == wer {
			return true
		}
	}

	return false
}

// Uebernahme gibt das Ergebnis, wenn beide Seiten dasselbe melden.
//
// # Die drei Ausgänge
//
//   - **Fertig**: beide Sitze haben gemeldet, und beide meinen dasselbe.
//   - **Noch nicht**: es fehlt eine Seite. Das ist kein Fehler des Turniers,
//     sondern eine Auskunft — und deshalb kein [ErrConflict].
//   - **Streit** ([ErrConflict]): beide haben gemeldet und meinen
//     Verschiedenes. Das Issue verlangt genau das: „Abweichende
//     Peerbestätigungen … erzeugen Konflikt statt Überschreiben."
//
// **Der Streit wird nicht aufgelöst.** Wer entscheidet, ist die
// Turnierleitung — ein Mensch, der beide fragt. Ein Programm, das sich für
// eine Seite entschiede, entschiede ohne Grundlage.
func (b Matchbindung) Uebernahme(
	meldungen []Matchmeldung,
) (Ergebnis, error) {
	jeSitz := map[string]Matchmeldung{}

	for _, meldung := range meldungen {
		if fehler := b.Nimmt(meldung); fehler != nil {
			return Ergebnis{}, fehler
		}

		// **Die letzte gilt.** Wer zweimal meldet, hat sich beim ersten Mal
		// vertippt; eine Übernahme, die die erste hielte, hielte die
		// widerrufene.
		jeSitz[meldung.Von] = meldung
	}

	if len(jeSitz) < len(b.Sitze) {
		return Ergebnis{}, fmt.Errorf(
			"%w: es haben %d von %d Seiten gemeldet", ErrValidation,
			len(jeSitz), len(b.Sitze))
	}

	erste := jeSitz[b.Sitze[0].Sitz]

	for _, sitz := range b.Sitze[1:] {
		andere := jeSitz[sitz.Sitz]

		if andere.Ergebnis != erste.Ergebnis || andere.Kette != erste.Kette {
			return Ergebnis{}, fmt.Errorf(
				"%w: die Seiten melden Verschiedenes — %q und %q müssen sich "+
					"einigen, bevor etwas eingetragen wird", ErrConflict,
				b.Sitze[0].Sitz, sitz.Sitz)
		}
	}

	return erste.Ergebnis, nil
}
