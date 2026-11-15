// **Die Turnierleitung** — Werkstück 4 von Phase 4.7
// (#1294),
// nach Vermerk [0062](../../docs/entscheidungen/0062-was-ein-turnier-hier-ist.md).
//
// # Was diese Schicht dazutut
//
// Der Speicher schreibt Zeilen, `domain` rechnet — hier steht, **was gilt**,
// und das sind drei Dinge, die keine der beiden anderen halten kann (§2):
//
//   - **Eine Runde wird nicht gepaart, solange die vorige offen ist.** Sonst
//     paarte sie nach einem Stand, den es noch nicht gibt.
//   - **Ein Ergebnis gehört zu einer Paarung dieses Turniers.** Ohne die
//     Frage entstünde aus einer falschen Kennung ein Fremdschlüsselfehler,
//     und der sagt dem Aufrufer nichts.
//   - **Ein angemeldetes Deck wird abgezogen, nicht verwiesen.** Die Liste
//     ist ein eingereichtes Dokument (Vermerk 0062, Entscheidung 3), und der
//     Abzug entsteht genau hier — im Speicher wäre es eine fachliche
//     Entscheidung, in `domain` fehlte der Zugriff auf das Deck.

package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Turnierspeicher ist, was die Leitung zum Lesen und Schreiben braucht.
type Turnierspeicher interface {
	Anlegen(ctx context.Context, turnier domain.Turnier) (domain.Turnier, error)
	Turnier(ctx context.Context, kennung int64) (domain.Turnier, error)
	Turniere(ctx context.Context, spiel string) ([]domain.Turnier, error)
	Anmelden(ctx context.Context, turnierID int64, teilnehmer domain.Teilnehmer) (
		domain.Teilnehmer, error)
	Aussteigen(ctx context.Context, teilnehmerID int64, ausgestiegen bool) error
	RundeLegen(ctx context.Context, turnierID int64, erwartet int,
		paarungen []domain.Paarung) (domain.Runde, error)
	ErgebnisSetzen(
		ctx context.Context, fassung, paarungID int64, stand *domain.Ergebnis,
	) error
	Loeschen(ctx context.Context, kennung int64) error

	// Abschluss und Protokoll — DOM-04.AK-04 (#1699), Vermerk 0139.
	Abschliessen(ctx context.Context, turnierID, fassung int64) error
	Wiedereroeffnen(ctx context.Context, turnierID, fassung int64) error
	Protokoll(ctx context.Context, turnierID int64) ([]domain.Protokolleintrag, error)

	// Einlesen legt ein Turnier samt Verlauf an — DOM-04.AK-03 (#1699).
	Einlesen(ctx context.Context, turnier domain.Turnier) (domain.Turnier, error)
}

// Deckabzug ist, was die Leitung über ein angemeldetes Deck braucht.
//
// **Eine eigene Schnittstelle mit einem Verfahren** (§2): die Leitung liest
// ein Deck genau einmal — beim Anmelden — und schreibt keines.
type Deckabzug interface {
	Deck(ctx context.Context, kennung int64) (domain.Deck, error)
}

// Kartenbezeichnung sagt, wie eine eigene Karte heißt.
//
// **Eine eigene Schnittstelle mit einem Verfahren** (§2). Der Abzug einer
// Deckliste ist Text; ohne den Namen bliebe von einer eigenen Karte nichts
// übrig, sobald jemand sie löscht.
type Kartenbezeichnung interface {
	NameZurKarte(ctx context.Context, karteID int64) (string, error)
}

// Turnierleitung ist der Anwendungsfall „ein Turnier führen".
type Turnierleitung struct {
	turniere Turnierspeicher
	decks    Deckabzug
	karten   Kartenbezeichnung

	// formate ist die Naht zu den Formaten — nachgerüstet mit DOM-04
	// (#1699), deshalb ein Setzer und kein weiterer Bauparameter.
	formate Turnierformate
}

// NeueTurnierleitung baut sie.
func NeueTurnierleitung(
	turniere Turnierspeicher, decks Deckabzug, karten Kartenbezeichnung,
) *Turnierleitung {
	return &Turnierleitung{turniere: turniere, decks: decks, karten: karten}
}

// Anlegen legt ein Turnier an.
func (l *Turnierleitung) Anlegen(ctx context.Context, turnier domain.Turnier) (
	domain.Turnier, error,
) {
	if err := turnier.Pruefe(); err != nil {
		return domain.Turnier{}, err
	}

	if err := l.formatPruefen(ctx, turnier); err != nil {
		return domain.Turnier{}, err
	}

	return l.turniere.Anlegen(ctx, turnier)
}

// Turnier gibt ein Turnier samt Verlauf.
func (l *Turnierleitung) Turnier(ctx context.Context, kennung int64) (
	domain.Turnier, error,
) {
	return l.turniere.Turnier(ctx, kennung)
}

// Turniere listet die Turniere eines Spiels — leer heißt: alle.
func (l *Turnierleitung) Turniere(ctx context.Context, spiel string) (
	[]domain.Turnier, error,
) {
	return l.turniere.Turniere(ctx, spiel)
}

// Loeschen nimmt ein Turnier samt seinem Verlauf heraus.
func (l *Turnierleitung) Loeschen(ctx context.Context, kennung int64) error {
	return l.turniere.Loeschen(ctx, kennung)
}

// Anmelden trägt einen Teilnehmer ein — **samt dem Abzug seines Decks**.
//
// **Der Abzug entsteht hier und nicht im Speicher.** Ob eine Deckliste ein
// eingereichtes Dokument ist, ist eine fachliche Frage; der Speicher legt sie
// nur ab.
//
// **Ein Deck, das es nicht gibt, ist ein Nein** — und keine stille Anmeldung
// ohne Liste. Wer ein Deck nennt, meint eines.
func (l *Turnierleitung) Anmelden(
	ctx context.Context, turnierID int64, teilnehmer domain.Teilnehmer,
) (domain.Teilnehmer, error) {
	turnier, err := l.turniere.Turnier(ctx, turnierID)
	if err != nil {
		return domain.Teilnehmer{}, err
	}

	if turnier.IstAbgeschlossen() {
		return domain.Teilnehmer{}, domain.TurnierIstAbgeschlossen(turnier.Name)
	}

	if teilnehmer.Name == "" {
		return domain.Teilnehmer{}, domain.NewFieldError("name",
			domain.SchluesselPflichtfeld, "Ein Teilnehmer braucht einen Namen.")
	}

	if teilnehmer.DeckID != 0 {
		deck, deckfehler := l.decks.Deck(ctx, teilnehmer.DeckID)
		if deckfehler != nil {
			return domain.Teilnehmer{}, deckfehler
		}

		// **Ein Turnier hat ein Spiel, ein Deck auch.**
		//
		// Bis zum Nachlauf über Phase 4.7 (Befund 3) hat niemand die beiden
		// verglichen: ein Deck aus einem anderen Spiel ließ sich anmelden,
		// und sein Abzug stand daneben, als gehöre er dazu. Die Formatprüfung
		// dahinter kann das nicht auffangen — sie prüft Karten gegen ein
		// Format, nicht das Spiel gegen das Turnier.
		if deck.Spiel != turnier.Spiel {
			return domain.Teilnehmer{}, domain.NewFieldError("deckID",
				domain.SchluesselFalschesSpiel,
				"Das Deck gehört zum Spiel »%s«, dieses Turnier zu »%s«.",
				deck.Spiel, turnier.Spiel)
		}

		abzug, abzugsfehler := l.abziehen(ctx, deck)
		if abzugsfehler != nil {
			return domain.Teilnehmer{}, abzugsfehler
		}

		teilnehmer.Deckabzug = abzug
	}

	return l.turniere.Anmelden(ctx, turnierID, teilnehmer)
}

// abziehen macht aus einem Deck die eingereichte Liste.
//
// **Text und keine Verweise** (Vermerk 0062, Entscheidung 3): was hier
// entsteht, überlebt das Löschen jeder Karte.
func (l *Turnierleitung) abziehen(ctx context.Context, deck domain.Deck) (
	[]domain.Abzugsplatz, error,
) {
	abzug := make([]domain.Abzugsplatz, 0, len(deck.Plaetze))

	for stelle, platz := range deck.Plaetze {
		eine := domain.Abzugsplatz{
			// **Bei einem echten Druck bleibt die Bezeichnung leer**: wie er
			// heißt, weiß nur Scryfall, und die Kennung ist die ehrliche
			// Antwort. Die Fläche kann sie auflösen; eine hier erfundene
			// Bezeichnung wäre für immer falsch.
			Printing: platz.Printing,
			Menge:    platz.Menge,
			Bereich:  platz.Bereich,
			Stelle:   stelle + 1,
		}

		if platz.KarteID != 0 {
			name, err := l.karten.NameZurKarte(ctx, platz.KarteID)
			if err != nil {
				return nil, err
			}

			// **Eine namenlose Karte bekommt ihre Kennung.** Der Abzug
			// verlangt Text — und „Karte 7" ist eine Auskunft, eine leere
			// Zeile ist keine.
			eine.Bezeichnung = name
			if name == "" {
				eine.Bezeichnung = fmt.Sprintf("Karte %d", platz.KarteID)
			}
		}

		abzug = append(abzug, eine)
	}

	return abzug, nil
}

// Aussteigen meldet einen Teilnehmer ab — oder wieder an.
//
// **Er bleibt in der Rangliste** (Appendix C): seine Matchquote geht weiter
// in die Gegnerquote jedes Gegners ein, den er hatte.
//
// **Und er gehört zu diesem Turnier.** Die Kennung eines Teilnehmers ist
// global; ohne diese Frage änderte eine Anfrage unter Turnier A jemanden aus
// Turnier B — genau wie beim Ergebnis, wo die Grenze von Anfang an stand
// (Befund 6 des Zweitblicks über Phase 4.7).
func (l *Turnierleitung) Aussteigen(
	ctx context.Context, turnierID, teilnehmerID int64, ausgestiegen bool,
) error {
	turnier, err := l.turniere.Turnier(ctx, turnierID)
	if err != nil {
		return err
	}

	if turnier.IstAbgeschlossen() {
		return domain.TurnierIstAbgeschlossen(turnier.Name)
	}

	if !gehoertZumTurnier(turnier, teilnehmerID) {
		return fmt.Errorf(
			"%w: der Teilnehmer %d gehört nicht zu diesem Turnier",
			domain.ErrNotFound, teilnehmerID)
	}

	return l.turniere.Aussteigen(ctx, teilnehmerID, ausgestiegen)
}

// gehoertZumTurnier sagt, ob ein Teilnehmer zu diesem Turnier gehört.
func gehoertZumTurnier(turnier domain.Turnier, teilnehmerID int64) bool {
	for _, teilnehmer := range turnier.Teilnehmer {
		if teilnehmer.ID == teilnehmerID {
			return true
		}
	}

	return false
}

// Rangliste rechnet den Stand eines Turniers.
func (l *Turnierleitung) Rangliste(ctx context.Context, turnierID int64) (
	[]domain.Rang, error,
) {
	turnier, err := l.turniere.Turnier(ctx, turnierID)
	if err != nil {
		return nil, err
	}

	return domain.Rangliste(turnier), nil
}

// Paarungsvorschlag zeigt, wie die nächste Runde aussähe — **ohne sie zu
// legen**.
//
// **Getrennt vom Legen**, damit ein Turnierleiter hinsehen kann, bevor er
// Tatsachen schafft.
func (l *Turnierleitung) Paarungsvorschlag(ctx context.Context, turnierID int64) (
	[]domain.Paarung, error,
) {
	turnier, err := l.turniere.Turnier(ctx, turnierID)
	if err != nil {
		return nil, err
	}

	return domain.Paarungsplan(turnier), nil
}

// RundeLegen paart die nächste Runde und legt sie an.
//
// **Nicht, solange die vorige offen ist.** Eine Runde, die nach einem
// unvollständigen Stand paart, stellt die falschen Tische zusammen — und das
// merkt niemand, bis die Rangliste am Ende nicht stimmt.
func (l *Turnierleitung) RundeLegen(ctx context.Context, turnierID int64) (
	domain.Runde, error,
) {
	turnier, err := l.turniere.Turnier(ctx, turnierID)
	if err != nil {
		return domain.Runde{}, err
	}

	if turnier.IstAbgeschlossen() {
		return domain.Runde{}, domain.TurnierIstAbgeschlossen(turnier.Name)
	}

	if offen := offeneErgebnisse(turnier); offen > 0 {
		return domain.Runde{}, fmt.Errorf(
			"%w: %d Paarung(en) haben noch kein Ergebnis — eine neue Runde "+
				"paarte nach einem Stand, den es noch nicht gibt",
			domain.ErrValidation, offen)
	}

	plan := domain.Paarungsplan(turnier)
	if len(plan) == 0 {
		return domain.Runde{}, fmt.Errorf(
			"%w: dieses Turnier hat niemanden, der noch spielt",
			domain.ErrValidation)
	}

	// **Die erwartete Nummer geht mit.** Der Plan ist aus diesem Stand
	// gerechnet; wer inzwischen eine Runde gelegt hat, hat einen anderen —
	// und die Transaktion weist das ab (Befund 8 des Zweitblicks).
	return l.turniere.RundeLegen(ctx, turnierID, len(turnier.Runden)+1, plan)
}

// offeneErgebnisse zählt **alle** Paarungen ohne Ergebnis.
//
// **Alle und nicht nur die der letzten Runde.** Der erste Anlauf sah nur die
// letzte an und begründete das damit, eine ältere offene Runde sei „ein
// Befund für den Turnierleiter, aber kein Grund anzuhalten". Die Begründung
// trägt nicht: eine Paarung ohne Ergebnis zählt in der Rangliste **gar
// nicht**, und die nächste Runde paart dann nach einem unvollständigen Stand
// — genau das, wogegen die Prüfung steht (Befund 7 des Zweitblicks über Phase
// 4.7).
//
// **Was es kostet, gehört gesagt**: wer ein Ergebnis nachtragen will und
// solange weiterspielen, kann das nicht mehr. Der Preis ist eine Rangliste,
// die stimmt.
//
// **Ein Bye zählt nicht als offen** — er *ist* ein Ergebnis (Appendix C).
func offeneErgebnisse(turnier domain.Turnier) int {
	var offen int

	for _, runde := range turnier.Runden {
		for _, paarung := range runde.Paarungen {
			if !paarung.IstBye() && paarung.Ergebnis == nil {
				offen++
			}
		}
	}

	return offen
}

// ErgebnisSetzen trägt ein Ergebnis ein — oder nimmt es zurück.
//
// **Die Fassung kommt vom Aufrufer** und nicht aus dem Lesen hier: eine, die
// dieser Weg sich selbst holte, stimmte immer und schützte nichts (§7,
// Regel 2).
func (l *Turnierleitung) ErgebnisSetzen(
	ctx context.Context, turnierID, fassung, paarungID int64,
	stand *domain.Ergebnis,
) error {
	turnier, err := l.turniere.Turnier(ctx, turnierID)
	if err != nil {
		return err
	}

	if turnier.IstAbgeschlossen() {
		return domain.TurnierIstAbgeschlossen(turnier.Name)
	}

	if !gehoertDazu(turnier, paarungID) {
		return fmt.Errorf(
			"%w: die Paarung %d gehört nicht zu diesem Turnier",
			domain.ErrNotFound, paarungID)
	}

	if stand != nil {
		if err := stand.Pruefe(turnier.HoechstensSpiele()); err != nil {
			return err
		}
	}

	return l.turniere.ErgebnisSetzen(ctx, fassung, paarungID, stand)
}
