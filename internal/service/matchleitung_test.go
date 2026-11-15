package service_test

// **Eine Turnierpaarung an eine Partie gebunden** — Werkstück G7 von Phase
// 5.5 (#1464).
//
// # Wogegen hier geprüft wird
//
//   - **Ein Bye startet keine Zweispielersession.**
//   - **Der Deckabdruck kommt aus der Partie** und nicht vom Aufrufer.
//   - **Ein Ergebnis wird einmal übernommen**; eine Wiederholung ändert
//     nichts, ein abweichendes erzeugt Konflikt statt Überschreiben.
//   - **Eine inzwischen geänderte Paarung** wird abgewiesen.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// matchprobe hält Bindungen und Meldungen im Arbeitsspeicher.
type matchprobe struct {
	bindungen map[int64]domain.Matchbindung
	meldungen map[int64][]domain.Matchmeldung

	klagt error

	// meldeklage klagt **nur** beim Lesen der Meldungen — derselbe Grund wie
	// bei `setzfehler` am Turnierdoppel.
	meldeklage error
}

func neueMatchprobe() *matchprobe {
	return &matchprobe{
		bindungen: map[int64]domain.Matchbindung{},
		meldungen: map[int64][]domain.Matchmeldung{},
	}
}

func (p *matchprobe) BindungSchreiben(
	_ context.Context, bindung domain.Matchbindung,
) error {
	if p.klagt != nil {
		return p.klagt
	}

	p.bindungen[bindung.PaarungID] = bindung
	delete(p.meldungen, bindung.PaarungID)

	return nil
}

func (p *matchprobe) Bindung(
	_ context.Context, paarungID int64,
) (domain.Matchbindung, error) {
	if p.klagt != nil {
		return domain.Matchbindung{}, p.klagt
	}

	bindung, da := p.bindungen[paarungID]
	if !da {
		return domain.Matchbindung{}, domain.MitGrund(domain.ErrNotFound, nil,
			"an dieser Paarung hängt keine Partie")
	}

	return bindung, nil
}

func (p *matchprobe) MeldungSchreiben(
	_ context.Context, meldung domain.Matchmeldung,
) error {
	if p.klagt != nil {
		return p.klagt
	}

	geblieben := []domain.Matchmeldung{}

	for _, alte := range p.meldungen[meldung.PaarungID] {
		if alte.Von != meldung.Von {
			geblieben = append(geblieben, alte)
		}
	}

	p.meldungen[meldung.PaarungID] = append(geblieben, meldung)

	return nil
}

func (p *matchprobe) Meldungen(
	_ context.Context, paarungID int64,
) ([]domain.Matchmeldung, error) {
	if p.klagt != nil {
		return nil, p.klagt
	}

	if p.meldeklage != nil {
		return nil, p.meldeklage
	}

	return p.meldungen[paarungID], nil
}

// partieprobeFuerMatch gibt eine Partie mit zwei Sitzen und Abzügen.
type partieprobeFuerMatch struct {
	partien map[string]domain.Partie

	// kettenklage klagt **nur** beim Lesen der Kette — derselbe Grund wie
	// bei `setzfehler` am Turnierdoppel: wer die Partie nicht lesen kann,
	// erreicht die Kette nie.
	kettenklage error
}

func (p *partieprobeFuerMatch) Partie(
	_ context.Context, kennung string,
) (domain.Partie, error) {
	partie, da := p.partien[kennung]
	if !da {
		return domain.Partie{}, domain.MitGrund(domain.ErrNotFound, nil,
			"die Partie gibt es nicht")
	}

	return partie, nil
}

// LetzteKette gibt einen Abdruck, der die Paarung erkennbar macht.
func (p *partieprobeFuerMatch) LetzteKette(
	_ context.Context, partieID string,
) (abdruck string, folge int64, fehler error) {
	if p.kettenklage != nil {
		return "", 0, p.kettenklage
	}

	if _, da := p.partien[partieID]; !da {
		return "", 0, domain.MitGrund(domain.ErrNotFound, nil,
			"die Partie gibt es nicht")
	}

	return "abdruck-7", 1, nil
}

// einMatchtisch gibt Turnier, Partie und Leitung — alles zusammengesteckt.
func einMatchtisch(t *testing.T) (
	*service.Matchleitung, *turnierdoppel, *matchprobe,
) {
	t.Helper()

	return einMatchtischSo(t, func(*domain.Partie) {})
}

// einMatchtischMitStrenge gibt einen, dessen Partie eine Strenge verlangt.
func einMatchtischMitStrenge(t *testing.T, grad domain.Strenge) (
	*service.Matchleitung, *turnierdoppel, *matchprobe,
) {
	t.Helper()

	return einMatchtischSo(t, func(p *domain.Partie) {
		p.Regelwerk.Strenge = grad
	})
}

// einMatchtischOhneProfil gibt einen, dessen Partie ein krummes Profil trägt.
//
// **Ein Fehler ist etwas anderes als ein Nein** — der Grund steht bei
// [service.Regelfreigabe.Pruefe].
func einMatchtischOhneProfil(t *testing.T) (
	*service.Matchleitung, *turnierdoppel, *matchprobe,
) {
	t.Helper()

	return einMatchtischSo(t, func(p *domain.Partie) {
		p.Profil = domain.Regelprofil{}
	})
}

// einMatchtischSo baut ihn mit einer Änderung an der Partie.
func einMatchtischSo(t *testing.T, aendere func(*domain.Partie)) (
	*service.Matchleitung, *turnierdoppel, *matchprobe,
) {
	t.Helper()

	leitung, turniere, match, _ := einMatchtischMitProben(t, aendere)

	return leitung, turniere, match
}

// einMatchtischMitProben gibt zusätzlich das Partiedoppel.
//
// **Eine vierte Rückgabe und keine fünfte Vorrichtung**: die Fehlerwege der
// Meldung aus der Partie (M17, #1618) greifen an ihr an, und ein zweiter
// Aufbau daneben liefe auseinander.
func einMatchtischMitProben(t *testing.T, aendere func(*domain.Partie)) (
	*service.Matchleitung, *turnierdoppel, *matchprobe,
	*partieprobeFuerMatch,
) {
	t.Helper()

	turniere := neuesTurnierdoppel()
	turniere.turniere[1] = &domain.Turnier{
		ID: 1, Fassung: 3, Name: "Freitagsrunde", Spiel: "magic",
		Teilnehmer: []domain.Teilnehmer{
			{ID: 10, TurnierID: 1, Name: "Anna", Stelle: 1},
			{ID: 11, TurnierID: 1, Name: "Bert", Stelle: 2},
		},
		Runden: []domain.Runde{{
			ID: 100, TurnierID: 1, Nummer: 1,
			Paarungen: []domain.Paarung{
				{ID: 7, RundeID: 100, A: 10, B: 11, Stelle: 1},
				{ID: 8, RundeID: 100, A: 12, Stelle: 2},
			},
		}},
	}

	abzugA := einMatchabzug("abdruckA")
	abzugB := einMatchabzug("abdruckB")

	partie := domain.Partie{
		ID: "p1", Spiel: "magic", Stand: domain.StandLaeuft, Wirt: "anna",
		Regelwerk: domain.Regelwerk{
			Kennung: "sandkasten", Fassung: domain.Grundprofilfassung,
			Strenge: domain.StrengeSandkasten,
		},
		Profil: domain.Grundprofil("magic"),
		Sitze: []domain.Sitz{
			{Kennung: "anna", Name: "Anna", Deck: abzugA},
			{Kennung: "bert", Name: "Bert", Deck: abzugB},
		},
	}

	aendere(&partie)

	partien := &partieprobeFuerMatch{
		partien: map[string]domain.Partie{"p1": partie},
	}

	match := neueMatchprobe()

	return service.NeueMatchleitung(turniere, partien, match), turniere,
		match, partien
}

// zweiSitze ist die gewöhnliche Zuordnung.
func zweiSitze() []service.Sitzzuordnung {
	return []service.Sitzzuordnung{
		{TeilnehmerID: 10, Sitz: "anna"},
		{TeilnehmerID: 11, Sitz: "bert", Peer: "12D3KooWBert"},
	}
}

// **Eine Bindung nimmt den Abdruck aus der Partie und nicht vom Aufrufer.**
func TestEineBindungNimmtDenAbdruckAusDerPartie(t *testing.T) {
	t.Parallel()

	leitung, _, _ := einMatchtisch(t)

	bindung, fehler := leitung.Binden(t.Context(), 1, 7, "p1", zweiSitze())
	if fehler != nil {
		t.Fatalf("binden: %v", fehler)
	}

	switch {
	case bindung.Fassung != 3:
		t.Errorf("die Fassung ist %d", bindung.Fassung)
	case bindung.Sitze[0].Deckabdruck != "abdruckA":
		t.Errorf("der Abdruck ist %q", bindung.Sitze[0].Deckabdruck)
	case bindung.Sitze[1].Peer != "12D3KooWBert":
		t.Errorf("die Peerkennung ist %q", bindung.Sitze[1].Peer)
	}
}

// **Ein Bye startet keine Zweispielersession.**
//
// # Warum der Satz mitgeprüft wird und nicht nur die Abweisung
//
// **Weil ein Bye auch ohne diese Prüfung abgewiesen würde** — er hat einen
// Sitz, und eine Bindung braucht zwei. Die Abweisung allein belegt den
// Wächter also nicht; eine Gegenprobe, die ihn entschärft, bliebe grün, und
// das ist keine (§10).
//
// **Was er leistet, ist die Auskunft.** Ohne ihn hinge sie davon ab, **wie**
// jemand es versucht: mit einem Sitz stünde „eine Paarung hat zwei Seiten",
// mit zwei stünde „der Teilnehmer 11 gehört nicht zu dieser Paarung" — beides
// richtig und beides am Thema vorbei.
//
// **Deshalb der Versuch mit zwei Sitzen.** Er ist der, bei dem die anderen
// Wächter die falsche Auskunft geben.
func TestEinByeWirdNichtGebunden(t *testing.T) {
	t.Parallel()

	leitung, _, _ := einMatchtisch(t)

	_, fehler := leitung.Binden(t.Context(), 1, 8, "p1", zweiSitze())

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "Bye") {
		t.Fatalf("der Satz nennt den Bye nicht: %v", fehler)
	}
}

// **Ein fremder Teilnehmer, ein fremder Sitz und eine fremde Paarung werden
// abgewiesen.**
func TestEineFremdeBindungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		paarung   int64
		partie    string
		zuordnung []service.Sitzzuordnung
	}{
		"eine Paarung, die es nicht gibt": {
			paarung: 99, partie: "p1", zuordnung: zweiSitze(),
		},
		"eine Partie, die es nicht gibt": {
			paarung: 7, partie: "p9", zuordnung: zweiSitze(),
		},
		"ein fremder Teilnehmer": {
			paarung: 7, partie: "p1",
			zuordnung: []service.Sitzzuordnung{
				{TeilnehmerID: 10, Sitz: "anna"},
				{TeilnehmerID: 99, Sitz: "bert"},
			},
		},
		"ein Sitz, den die Partie nicht hat": {
			paarung: 7, partie: "p1",
			zuordnung: []service.Sitzzuordnung{
				{TeilnehmerID: 10, Sitz: "anna"},
				{TeilnehmerID: 11, Sitz: "cem"},
			},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			leitung, _, _ := einMatchtisch(t)

			if _, fehler := leitung.Binden(t.Context(), 1, fall.paarung,
				fall.partie, fall.zuordnung); fehler == nil {
				t.Fatal("die fremde Bindung kam durch")
			}
		})
	}
}

// **Eine Paarung mit Ergebnis wird nicht mehr gebunden.**
func TestEineEntschiedenePaarungWirdNichtGebunden(t *testing.T) {
	t.Parallel()

	leitung, turniere, _ := einMatchtisch(t)

	ergebnis := &domain.Ergebnis{SpieleA: 2}
	turniere.turniere[1].Runden[0].Paarungen[0].Ergebnis = ergebnis

	if _, fehler := leitung.Binden(
		t.Context(), 1, 7, "p1", zweiSitze()); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("erwartet ErrConflict, bekam %v", fehler)
	}
}
