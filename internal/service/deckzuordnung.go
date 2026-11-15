package service

// **Bedarf und Zuordnung** — Werkstück U9 von Phase 5.1
// (#1420).
//
// # Zwei Fragen, die man nicht verwechseln darf
//
// *Was braucht dieses Deck, und was davon habe ich?* — das rechnet
// [Habenabgleich], und **er reserviert nichts**: zwei Decks mit denselben
// drei Blitzen sind beide vollständig. *Welche Exemplare gehören in dieses
// Deck?* — das entscheidet eine **Zuordnung**, und sie nimmt sie aus dem
// Freien: das zweite Deck sieht sie dann nicht mehr.
//
// Beides steht nebeneinander in [Zuordnungszeile], und genau das verlangt die
// Abnahme: zwei erklärbar verschiedene Werte, keiner davon falsch.
//
// # Der Vertrag mit dem Handeln
//
// Reservierungen kommen aus T2 (#1467),
// das in dieser Fassung nicht gebaut ist. Die Rechnung kennt sie als Zahl
// ([Reservierungen]); ohne Naht ist sie null. So steht der Vertrag in der
// Rechnung und nicht als Annahme im Kopf — und T2 füllt eine Naht statt eine
// Formel zu ändern.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sammelliste"
)

// Zuordnungsablage ist, was die Pflege vom Bestand braucht.
type Zuordnungsablage interface {
	Setzen(ctx context.Context, zuordnung domain.Deckzuordnung) (domain.Deckzuordnung, error)
	Loesen(ctx context.Context, deckID, id int64) error
	JeDeck(ctx context.Context, deckID int64) ([]domain.Deckzuordnung, error)
	JeExemplar(ctx context.Context, exemplarID int64) ([]domain.Deckzuordnung, error)
	ZugeordnetJeExemplar(ctx context.Context) (map[int64]int, error)
}

// Reservierungen sagt, was ein Handelsvorgang festhält (T2, #1467).
type Reservierungen interface {
	ReserviertJeExemplar(ctx context.Context) (map[int64]int, error)
}

// Exemplarnachschlag holt ein Exemplar nach seiner Kennung.
type Exemplarnachschlag interface {
	Einzeln(ctx context.Context, id int64) (*domain.Exemplar, error)
}

// Zuordnungszeile ist eine Zeile des Abgleichs samt ihrer Zuordnungen.
type Zuordnungszeile struct {
	Habenzeile

	// Reserviert, Zugeordnet und Frei beziehen sich auf **alle** Exemplare
	// dieses Drucks, nicht nur auf die dieses Decks.
	Reserviert int
	Zugeordnet int
	Frei       int

	Zuordnungen []Zuordnungsangabe
}

// Zuordnungsangabe ist eine Zuordnung, wie eine Fläche sie zeigt.
type Zuordnungsangabe struct {
	domain.Deckzuordnung

	// Bezeichnung ist, wie die Sammlung das Exemplar nennt.
	Bezeichnung string

	// Abweichend sagt, dass das Exemplar einen **anderen** Druck trägt als
	// der Platz — ausdrücklich gewählt, nie über Namensgleichheit.
	Abweichend bool
}

// Zuordnungsstand ist der Bedarf eines Decks samt seinen Zuordnungen.
type Zuordnungsstand struct {
	Haben  Habenstand
	Zeilen []Zuordnungszeile
}

// Deckbezug sagt, welches Deck ein Exemplar benutzt.
type Deckbezug struct {
	DeckID int64
	Name   string
	Menge  int
}

// Deckzuordnungspflege ist der Anwendungsfall „welche Exemplare gehören in
// dieses Deck?".
type Deckzuordnungspflege struct {
	decks      Deckspeicher
	ablage     Zuordnungsablage
	sammlung   Exemplarnachschlag
	abgleich   *Habenabgleich
	reserviert Reservierungen
}

// NeueDeckzuordnungspflege baut sie.
func NeueDeckzuordnungspflege(
	decks Deckspeicher, ablage Zuordnungsablage,
	sammlung Exemplarnachschlag, abgleich *Habenabgleich,
) *Deckzuordnungspflege {
	return &Deckzuordnungspflege{decks: decks, ablage: ablage, sammlung: sammlung, abgleich: abgleich}
}

// MitReservierungen setzt die Naht zum Handeln (T2).
//
// **Nachgerüstet und darum wahlfrei** — dieselbe Bauform wie
// `Sammlungsausfuhr.MitHerkunft`: ohne sie ist reserviert null, und das ist
// heute die Wahrheit.
func (d *Deckzuordnungspflege) MitReservierungen(reserviert Reservierungen) *Deckzuordnungspflege {
	d.reserviert = reserviert

	return d
}

// Stand gibt den Bedarf eines Decks samt seinen Zuordnungen.
func (d *Deckzuordnungspflege) Stand(ctx context.Context, deckID int64) (Zuordnungsstand, error) {
	haben, err := d.abgleich.Fuer(ctx, deckID)
	if err != nil {
		return Zuordnungsstand{}, err
	}

	deck, deckfehler := d.decks.Deck(ctx, deckID)
	if deckfehler != nil {
		return Zuordnungsstand{}, deckfehler
	}

	zuordnungen, zuordnungsfehler := d.ablage.JeDeck(ctx, deckID)
	if zuordnungsfehler != nil {
		return Zuordnungsstand{}, zuordnungsfehler
	}

	belegt, belegtfehler := d.belegung(ctx)
	if belegtfehler != nil {
		return Zuordnungsstand{}, belegtfehler
	}

	stand := Zuordnungsstand{Haben: haben, Zeilen: make([]Zuordnungszeile, 0, len(haben.Zeilen))}

	for _, zeile := range haben.Zeilen {
		gebaut, baufehler := d.zeile(ctx, deck, zeile, zuordnungen, belegt)
		if baufehler != nil {
			return Zuordnungsstand{}, baufehler
		}

		stand.Zeilen = append(stand.Zeilen, gebaut)
	}

	return stand, nil
}

// belegung sind Reservierungen und Zuordnungen je Exemplar.
type belegung struct {
	reserviert map[int64]int
	zugeordnet map[int64]int
}

func (d *Deckzuordnungspflege) belegung(ctx context.Context) (belegung, error) {
	zugeordnet, err := d.ablage.ZugeordnetJeExemplar(ctx)
	if err != nil {
		return belegung{}, err
	}

	belegt := belegung{reserviert: map[int64]int{}, zugeordnet: zugeordnet}

	// **Ein `nil` als Naht ist so gut wie keine** — `MitReservierungen(nil)`
	// nimmt sie heraus, und der Vergleich fängt beides.
	if d.reserviert == nil {
		return belegt, nil
	}

	reserviert, reservierfehler := d.reserviert.ReserviertJeExemplar(ctx)
	if reservierfehler != nil {
		return belegung{}, reservierfehler
	}

	belegt.reserviert = reserviert

	return belegt, nil
}

// zeile baut eine Zeile des Stands: Bedarf, Belegung und die Zuordnungen der
// Plätze, die diesen Druck nennen.
func (d *Deckzuordnungspflege) zeile(
	ctx context.Context, deck domain.Deck, haben Habenzeile,
	zuordnungen []domain.Deckzuordnung, belegt belegung,
) (Zuordnungszeile, error) {
	zeile := Zuordnungszeile{Habenzeile: haben, Zuordnungen: []Zuordnungsangabe{}}
	schluessel := domain.Druckschluessel(haben.DruckID, haben.Printing)

	for _, zuordnung := range zuordnungen {
		platz, gibts := platzAus(deck, zuordnung.PlatzID)
		if !gibts || platz.Druckschluessel() != schluessel {
			continue
		}

		exemplar, err := d.sammlung.Einzeln(ctx, zuordnung.ExemplarID)
		if err != nil {
			return Zuordnungszeile{}, err
		}

		zeile.Zugeordnet += zuordnung.Menge
		zeile.Zuordnungen = append(zeile.Zuordnungen, Zuordnungsangabe{
			Deckzuordnung: zuordnung,
			Bezeichnung:   exemplar.Bezeichnung(),
			Abweichend:    domain.Abweichend(platz, *exemplar),
		})
	}

	// **Frei rechnet über die Exemplare dieses Drucks**, nicht über die
	// Zuordnungen dieses Decks: ein anderes Deck hält dieselben Stücke fest.
	verfuegbar := domain.Verfuegbarkeit{Besitz: haben.Vorhanden}

	for _, exemplarID := range haben.ExemplarIDs {
		verfuegbar.Reserviert += belegt.reserviert[exemplarID]
		verfuegbar.Zugeordnet += belegt.zugeordnet[exemplarID]
	}

	zeile.Reserviert = verfuegbar.Reserviert
	zeile.Frei = verfuegbar.Frei()

	return zeile, nil
}

// platzAus sucht einen Platz im Deck.
func platzAus(deck domain.Deck, platzID int64) (domain.Deckplatz, bool) {
	for _, platz := range deck.Plaetze {
		if platz.ID == platzID {
			return platz, true
		}
	}

	return domain.Deckplatz{}, false
}

// Fehlkarten gibt, was zu besorgen wäre — als Liste, die die Einfuhr liest.
//
// **Dieselbe Gestalt wie die Sammlungsausfuhr** (`sammelliste.SchreibeListe`):
// wer sie in einen Deckbaukasten oder einen Marktplatz einfügt, braucht keine
// Übersetzung. Ohne Namen steht die Kennung — sonst wäre die Zeile leer.
func (d *Deckzuordnungspflege) Fehlkarten(ctx context.Context, deckID int64) (string, error) {
	haben, err := d.abgleich.Fuer(ctx, deckID)
	if err != nil {
		return "", err
	}

	zeilen := make([]sammelliste.Ausfuhrzeile, 0, len(haben.Zeilen))

	for _, zeile := range haben.Zeilen {
		if zeile.Fehlend == 0 {
			continue
		}

		zeilen = append(zeilen, sammelliste.Ausfuhrzeile{
			Printing: zeile.Printing, Menge: zeile.Fehlend, Name: zeile.Bezeichnung,
		})
	}

	return sammelliste.SchreibeListe(zeilen), nil
}

// DecksMitExemplar sagt, welche Decks ein Exemplar benutzen — der Rückweg.
func (d *Deckzuordnungspflege) DecksMitExemplar(ctx context.Context, exemplarID int64) (
	[]Deckbezug, error,
) {
	zuordnungen, err := d.ablage.JeExemplar(ctx, exemplarID)
	if err != nil {
		return nil, err
	}

	bezuege := make([]Deckbezug, 0, len(zuordnungen))

	for _, zuordnung := range zuordnungen {
		deck, deckfehler := d.decks.Deck(ctx, zuordnung.DeckID)
		if deckfehler != nil {
			return nil, deckfehler
		}

		bezuege = append(bezuege, Deckbezug{
			DeckID: deck.ID, Name: deck.Name, Menge: zuordnung.Menge,
		})
	}

	return bezuege, nil
}

// Zuordnen legt Exemplare in einen Platz — oder sagt, warum nicht.
func (d *Deckzuordnungspflege) Zuordnen(
	ctx context.Context, deckID, platzID, exemplarID int64, menge int,
) (domain.Deckzuordnung, error) {
	deck, err := d.decks.Deck(ctx, deckID)
	if err != nil {
		return domain.Deckzuordnung{}, err
	}

	platz, gibts := platzAus(deck, platzID)
	if !gibts {
		return domain.Deckzuordnung{}, fmt.Errorf(
			"%w: Platz %d liegt nicht in Deck %d", domain.ErrNotFound, platzID, deckID)
	}

	if !platz.Zuordenbar() {
		return domain.Deckzuordnung{}, domain.NewFieldError("platzID", domain.SchluesselValidierung,
			"Dieser Platz nennt eine eigene Karte ohne Verweis — es gibt kein Exemplar, das ihr entspräche.")
	}

	exemplar, exemplarfehler := d.sammlung.Einzeln(ctx, exemplarID)
	if exemplarfehler != nil {
		return domain.Deckzuordnung{}, exemplarfehler
	}

	if !exemplar.Zuordenbar() {
		return domain.Deckzuordnung{}, domain.NewFieldError("exemplarID", domain.SchluesselValidierung,
			"Diese Zeile ist ein Wunsch und kein Besitz — sie lässt sich keinem Deck geben.")
	}

	if fehler := d.passt(ctx, exemplar, platzID, menge); fehler != nil {
		return domain.Deckzuordnung{}, fehler
	}

	return d.ablage.Setzen(ctx, domain.Deckzuordnung{
		DeckID: deckID, PlatzID: platzID, ExemplarID: exemplarID, Menge: menge,
	})
}

// passt prüft, ob die Menge noch ins Freie geht.
//
// **Die eigene Zuordnung zählt nicht gegen sich selbst**: wer eine
// bestehende Zuordnung erhöht, gibt die neue Gesamtmenge an, und die alte ist
// schon in der Summe.
func (d *Deckzuordnungspflege) passt(
	ctx context.Context, exemplar *domain.Exemplar, platzID int64, menge int,
) error {
	if menge < 1 {
		return domain.NewFieldError("menge", domain.SchluesselValidierung,
			"Die Menge ist %d — mindestens eines muss es sein.", menge)
	}

	belegt, err := d.belegung(ctx)
	if err != nil {
		return err
	}

	eigene, eigenefehler := d.eigene(ctx, exemplar.ID, platzID)
	if eigenefehler != nil {
		return eigenefehler
	}

	verfuegbar := domain.Verfuegbarkeit{
		Besitz:     exemplar.Menge,
		Reserviert: belegt.reserviert[exemplar.ID],
		Zugeordnet: belegt.zugeordnet[exemplar.ID] - eigene,
	}

	if !verfuegbar.Erlaubt(menge) {
		return fmt.Errorf(
			"%w: von %s sind %d frei — %d gehen nicht",
			domain.ErrConflict, exemplar.Bezeichnung(), verfuegbar.Frei(), menge)
	}

	return nil
}

// eigene ist die Menge, die dieser Platz von diesem Exemplar schon hält.
func (d *Deckzuordnungspflege) eigene(ctx context.Context, exemplarID, platzID int64) (int, error) {
	zuordnungen, err := d.ablage.JeExemplar(ctx, exemplarID)
	if err != nil {
		return 0, err
	}

	for _, zuordnung := range zuordnungen {
		if zuordnung.PlatzID == platzID {
			return zuordnung.Menge, nil
		}
	}

	return 0, nil
}

// Loesen nimmt eine Zuordnung aus ihrem Deck.
func (d *Deckzuordnungspflege) Loesen(ctx context.Context, deckID, id int64) error {
	return d.ablage.Loesen(ctx, deckID, id)
}
