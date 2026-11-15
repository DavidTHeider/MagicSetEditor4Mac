package service

// **Die digitale Deckleihe** — Werkstück G6 von Phase 5.5
// (#1463).
//
// # Die Arbeitsteilung
//
// **`domain` sagt, ob ein Dokument gilt; hier wird es ausgestellt und
// geprüft.** Die Unterschrift braucht `internal/peer`, und `domain` kennt es
// nicht (§2) — also steht die Rechnung dort, wo der Schlüssel liegt.
//
// # Und warum der Abzug mitgespeichert wird
//
// **Weil die Leihe sonst auf etwas zeigte, das verschwinden darf.** Das Issue
// verlangt: „Änderung am ursprünglichen Deck und Löschen des lokalen Ursprungs
// während der Partie ändern den verliehenen Snapshot nicht." Das Dokument
// nennt nur den **Abdruck**; die Bytes liegen daneben, auf beiden Seiten.

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Leihablage ist, was die Leihstelle zum Lesen und Schreiben braucht.
type Leihablage interface {
	LeiheSchreiben(ctx context.Context, stand domain.Leihstand) error
	Leihe(ctx context.Context, kennung string) (domain.Leihstand, error)
	Leihen(ctx context.Context) ([]domain.Leihstand, error)
}

// Leihwunsch ist, was jemand verleihen will.
type Leihwunsch struct {
	// Entleiher ist die **Peerkennung** des Gegenübers.
	Entleiher string

	DeckID int64

	Bindung domain.Leihbindung
	Rechte  []domain.Leihrecht

	// Dauer ist, wie lange sie gilt — ab jetzt.
	//
	// **Eine Dauer und kein Zeitpunkt.** Wer einen Zeitpunkt schickte,
	// schickte ihn aus der Uhr seines Fensters; die Leihe entsteht hier, und
	// hier steht die Uhr, gegen die später geprüft wird.
	Dauer time.Duration

	Modus domain.Leihmodus
}

// Leihstelle stellt Leihen aus, nimmt sie an und legt sie vor.
type Leihstelle struct {
	ablage  Leihablage
	decks   Deckquelle
	ausweis Ausweisstelle
	uhr     func() time.Time

	// schranke ist die Handelsrichtlinie (#2435).
	schranke Handelsschranke
}

// NeueLeihstelle baut sie.
func NeueLeihstelle(
	ablage Leihablage, decks Deckquelle, ausweis Ausweisstelle,
	uhr func() time.Time,
) *Leihstelle {
	return &Leihstelle{ablage: ablage, decks: decks, ausweis: ausweis, uhr: uhr, schranke: ohneEinstellung{}}
}

// MitHandelsschranke setzt die Handelsrichtlinie, die gilt (#2435).
func (s *Leihstelle) MitHandelsschranke(schranke Handelsschranke) *Leihstelle {
	s.schranke = schranke

	return s
}

// Anbieten stellt eine Leihe aus und unterschreibt sie.
//
// **Der Abzug kommt mit zurück**, denn der Aufrufer muss ihn dem Gegenüber
// schicken: das Dokument nennt nur den Abdruck.
func (s *Leihstelle) Anbieten(
	ctx context.Context, wunsch Leihwunsch,
) (domain.Deckleihe, domain.Partiedeck, error) {
	if fehler := artErlaubt(ctx, s.schranke, domain.HandelDeckleihe); fehler != nil {
		return domain.Deckleihe{}, domain.Partiedeck{}, fehler
	}

	ich, fehler := s.ausweis.Ausweis(ctx)
	if fehler != nil {
		return domain.Deckleihe{}, domain.Partiedeck{}, fehler
	}

	deck, fehler := s.decks.Deck(ctx, wunsch.DeckID)
	if fehler != nil {
		return domain.Deckleihe{}, domain.Partiedeck{}, fehler
	}

	abzug := abzugAus(deck)
	jetzt := s.uhr()

	leihe := domain.Deckleihe{
		Kennung:     neueLeihkennung(),
		Verleiher:   string(ich.Kennung()),
		Entleiher:   wunsch.Entleiher,
		Abdruck:     abzug.Abdruck,
		Bindung:     wunsch.Bindung,
		Rechte:      wunsch.Rechte,
		Ausgestellt: jetzt,
		Bis:         jetzt.Add(wunsch.Dauer),
		Modus:       wunsch.Modus,
	}

	if fehler := leihe.Pruefe(); fehler != nil {
		return domain.Deckleihe{}, domain.Partiedeck{}, fehler
	}

	leihe.Unterschrift = hex.EncodeToString(
		ich.Signiere(leihe.Unterschriftsform()))

	stand := domain.Leihstand{Leihe: leihe, Abzug: abzug, Verliehen: true}

	if fehler := s.ablage.LeiheSchreiben(ctx, stand); fehler != nil {
		return domain.Deckleihe{}, domain.Partiedeck{}, fehler
	}

	return leihe, abzug, nil
}

// Annehmen nimmt eine fremde Leihe samt ihrem Abzug an.
//
// # Was hier geprüft wird, bevor irgendetwas liegen bleibt
//
// **Unterschrift, Gestalt, Abdruck und Empfänger** — in dieser Reihenfolge,
// und **vor** dem Schreiben. Das Issue verlangt es für die ganze
// Austauschschicht: „ein verweigerter Import lässt weder aktive Leihe noch
// halbe dauerhafte Objekte zurück."
//
// **Und sie übernimmt nichts in die Sammlung.** Eine Leihe ist eine
// Nutzungsfreigabe; der Abzug liegt neben ihr und nicht darin. „Die Leihe
// importiert nicht automatisch eine dauerhafte Kopie in die Sammlung" steht
// wörtlich im Issue.
func (s *Leihstelle) Annehmen(
	ctx context.Context, leihe domain.Deckleihe, abzug domain.Partiedeck,
) error {
	if fehler := artErlaubt(ctx, s.schranke, domain.HandelDeckleihe); fehler != nil {
		return fehler
	}

	ich, fehler := s.ausweis.Ausweis(ctx)
	if fehler != nil {
		return fehler
	}

	if fehler := PruefeLeihe(leihe); fehler != nil {
		return fehler
	}

	if leihe.Entleiher != string(ich.Kennung()) {
		return fmt.Errorf("%w: diese Leihe gilt %s und nicht diesem Profil",
			domain.ErrValidation, leihe.Entleiher)
	}

	// **Der Abzug muss der sein, den das Dokument meint.** Sonst käme ein
	// echtes Dokument mit fremden Karten an, und die Unterschrift stimmte —
	// sie steht über dem Abdruck und nicht über den Bytes.
	if abzug.Abdruck != leihe.Abdruck {
		return fmt.Errorf(
			"%w: der mitgeschickte Abzug gehört nicht zu dieser Leihe",
			domain.ErrValidation)
	}

	if fehler := abzug.Pruefe(); fehler != nil {
		return fehler
	}

	return s.ablage.LeiheSchreiben(ctx,
		domain.Leihstand{Leihe: leihe, Abzug: abzug})
}

// Widerrufen nimmt eine Leihe zurück.
//
// **Nur der Verleiher kann das**, und auf diesem Rechner heißt das: nur wer
// sie ausgestellt hat. Ein Entleiher, der sie hier als widerrufen markierte,
// beschriebe seinen eigenen Stand — und der Verleiher erführe es nie.
func (s *Leihstelle) Widerrufen(ctx context.Context, kennung string) error {
	stand, fehler := s.ablage.Leihe(ctx, kennung)
	if fehler != nil {
		return fehler
	}

	if !stand.Verliehen {
		return fmt.Errorf(
			"%w: diese Leihe hat jemand anderes ausgestellt — zurücknehmen "+
				"kann sie nur er", domain.ErrConflict)
	}

	stand.Widerrufen = true

	return s.ablage.LeiheSchreiben(ctx, stand)
}

// Leihen gibt alles, was dieser Rechner an Leihen weiß.
func (s *Leihstelle) Leihen(ctx context.Context) ([]domain.Leihstand, error) {
	return s.ablage.Leihen(ctx)
}

// PruefeLeihe hält Unterschrift und Gestalt eines Dokuments.
//
// **Die Unterschrift zuerst.** Was nicht echt ist, wird nicht gelesen — und
// eine Fehlermeldung über ein Feld eines gefälschten Dokuments sagte etwas
// über den Fälscher aus.
func PruefeLeihe(leihe domain.Deckleihe) error {
	unterschrift, fehler := hex.DecodeString(leihe.Unterschrift)
	if fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"die Unterschrift der Leihe ist kein Hex")
	}

	if fehler := peer.Pruefe(peer.Kennung(leihe.Verleiher),
		leihe.Unterschriftsform(), unterschrift); fehler != nil {
		return fehler
	}

	return leihe.Pruefe()
}

// Leihkennungslaenge ist die Zahl der Hexzeichen einer Leihkennung.
//
// **Zweiunddreißig, also sechzehn Byte.** Sie muss zwischen zwei Rechnern
// eindeutig sein, ohne dass jemand sie abstimmt — eine fortlaufende Zahl
// gäbe zwei Leihen dieselbe.
const Leihkennungslaenge = 32

// neueLeihkennung würfelt eine.
//
// **Ohne Fehlerrückgabe**, wie an den anderen Stellen dieses Repositoriums,
// die würfeln: `crypto/rand.Read` „never returns an error, and always fills b
// entirely" — ein Zweig dafür wäre einer, den kein Lauf betreten kann, und
// sähe aus wie eine geprüfte Zusage (§10).
func neueLeihkennung() string {
	roh := make([]byte, Leihkennungslaenge/2)

	_, _ = rand.Read(roh) //nolint:errcheck // siehe oben

	return hex.EncodeToString(roh)
}
