package service

// **Wie ein Gast zurückkommt** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum der Stand nicht im Gast steht
//
// **Weil ein Gast je Ruf neu gebaut wird und ein Stand das überleben muss.**
// Genau daran ist der Zeigerstand in diesem Werkstück schon einmal
// gescheitert: ein Zähler im Verbund fing bei jedem Ruf wieder bei eins an.
// Die Anwendung hält ihn, weil sie die Lebensdauer hat.
//
// # Und warum die Folge der Partie zählt und nicht die der Leitung
//
// **Weil eine Leitung bei jedem Abriss neu anfängt.** `Partienachricht.Folge`
// gehört der **Sitzung**; nach einem Neuverbinden steht sie wieder bei eins.
// Was den Spielstand beschreibt, ist [domain.Ereignis.Folge] — sie gehört der
// Partie und überlebt jede Leitung.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Standhalter hält, wo ein Gast im Strom steht — über einen Abriss hinaus.
type Standhalter interface {
	// Merke trägt ein, was dieser Gast zuletzt gesehen hat.
	Merke(folge int64, abdruck string)

	// Stand gibt es zurück — null und leer, wenn er noch nie dabei war.
	Stand() (folge int64, abdruck string)
}

// Kehrezurueck sagt dem Wirt, wo dieser Gast stehengeblieben ist.
//
// **Auch beim ersten Mal.** Ein Gast, der noch nie dabei war, schickt Folge
// null — und bekommt darauf denselben Weg wie jeder andere. Ein Sonderfall
// „erstes Mal" wäre ein zweiter Weg zum selben Ziel, und er liefe beim ersten
// Unterschied auseinander (§16).
func (g *Partiegast) Kehrezurueck(ctx context.Context) error {
	folge, abdruck := int64(0), ""

	if g.halter != nil {
		folge, abdruck = g.halter.Stand()
	}

	// **Hier steht kein Fehlerzweig, und das ist eine Aussage.** Ein
	// [domain.Wiederaufnahmewunsch] trägt ausschließlich Zeichenketten und
	// Zahlen — seit P9b auch einen Abzug, der aus nichts anderem besteht;
	// `encoding/json` klagt darüber nie, und die Art ist eine
	// Konstante. Gehalten wird die Voraussetzung von
	// `internal/domain/sichtgestalt_test.go` (§10).
	inhalt, _ := domain.VerpackeKontrolle( //nolint:errcheck // der Grund steht darüber
		domain.KontrolleRueckkehr, domain.Wiederaufnahmewunsch{
			Partie:      g.partie,
			Epoche:      domain.Partieepoche,
			Folge:       folgeZahl(folge),
			Kette:       abdruck,
			Mitgebracht: g.mitbringen(folge),
			Kann:        []string{domain.FaehigkeitBild},
		})

	return g.sitzung.Sende(ctx, domain.ArtKontrolle, inhalt)
}

// abzug nimmt die Antwort des Wirts entgegen.
//
// # Warum ein Abbruch hier zum Fehler wird
//
// **Weil er die Leitung niederlegen soll.** Das Issue verlangt „blockieren
// statt weiterzuspielen": ein Gast, der nach einem Abbruch weiterzöge, spielte
// an einem Tisch, dessen Geschichte er nicht teilt.
//
// # Und warum nur der Abzug etwas tut
//
// **Weil die anderen drei Wege sich selbst erklären.** Bei `weiter` fehlt
// nichts, bei `Ereignisse` kommt das Fehlende gleich hinterher und trägt
// seinen eigenen Stand mit. Nur der **Abzug** springt — und was dabei zu tun
// ist, kann der Gast nicht aus dem Strom ableiten, weil er ihn nicht bekommt.
func (g *Partiegast) abzug(ctx context.Context, roh []byte) error {
	var antwort domain.Rueckkehr

	if fehler := json.Unmarshal(roh, &antwort); fehler != nil {
		return fmt.Errorf("%w: der Inhalt ist keine Rückkehr: %s",
			domain.ErrValidation, fehler.Error())
	}

	if antwort.Art == domain.WegAbbruch {
		return Absage{Grund: antwort.Grund}
	}

	if antwort.Art != domain.WegAbzug {
		return nil
	}

	// **Beim Abzug springt der Stand, und der Abdruck kommt mit.** Ihn selbst
	// zu rechnen ginge nicht: der Gast hat die Ereignisse dazu nicht — und
	// mit einer leeren Kette bekäme er beim nächsten Mal einen Abbruch.
	if g.halter != nil {
		g.halter.Merke(antwort.Stand, antwort.Abdruck)
	}

	return g.nimmBild(ctx, antwort.Bild)
}

// MitMitgebracht nennt, womit sich dieser Gast setzt, wenn sein Sitz wartet
// (P9b, #2377).
func (g *Partiegast) MitMitgebracht(mit *domain.Mitgebracht) *Partiegast {
	g.mitgebracht = mit

	return g
}

// mitbringen gibt das Mitgebrachte **nur beim ersten Mal** — wer schon dabei
// war, sitzt mit seinem Deck, und der Wirt fragt nicht noch einmal.
func (g *Partiegast) mitbringen(folge int64) *domain.Mitgebracht {
	if folge != 0 {
		return nil
	}

	return g.mitgebracht
}

// ErrWirtLehntAb ist die Absage eines Wirts mit Grund (P9c, #2377) — das
// Fenster des Gastes zeigt sie, statt nur von einer abgerissenen Leitung zu
// sprechen.
var ErrWirtLehntAb = errors.New("der Wirt nimmt diesen Stand nicht an")

// Absage trägt den Satz, mit dem der Wirt ablehnt — **allein**, damit das
// Fenster ihn unter seinen eigenen setzen kann statt hinter eine Kette aus
// Präfixen. Sie ist ein Konflikt und eine [ErrWirtLehntAb].
type Absage struct {
	Grund string
}

func (a Absage) Error() string {
	return domain.ErrConflict.Error() + ": " + ErrWirtLehntAb.Error() + ": " + a.Grund
}

// Unwrap macht sie für [errors.Is] zu beidem.
func (a Absage) Unwrap() []error {
	return []error{domain.ErrConflict, ErrWirtLehntAb}
}
