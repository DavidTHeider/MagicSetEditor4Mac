package service_test

// **Das Doppel der Tokenablage** — Werkstück 10 von Phase 4.12
// (#1351).
//
// **Es steht eigens hier und nicht neben den Fällen** — derselbe Grund wie bei
// `kennungsdoppel`: die Testdatei hält Verhaltensklassen zusammen, ein Doppel
// darunter machte sie zu der einen Stelle, an der jede Änderung stattfindet
// (§2).

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// tokendoppel ist eine Ablage im Arbeitsspeicher.
//
// **Sie hütet nicht selbst.** Was hineingereicht wird, legt sie ab, wie sie es
// bekommt — das Hüten ist die Entscheidung des Dienstes, und ein Doppel, das
// sie nachbaute, prüfte den Nachbau.
type tokendoppel struct {
	// abgelegt ist, was der Bestand kennt, nach Kennung.
	abgelegt map[int64]domain.Zugriffstoken

	// gehuetet ist die gehütete Form je Kennung.
	gehuetet map[int64]string

	// vorgaenger ist die gehütete Form des vorigen Geheimnisses je Token —
	// Werkstück O2 von Phase 5.2 (#1426).
	vorgaenger map[int64]string

	// vermerkfehler ist ein Fehler, den nur das Vermerken der Nutzung gibt.
	vermerkfehler error

	// naechste ist die Kennung, die das nächste Anlegen bekommt.
	naechste int64

	// kaputt ist der Fehler, den jeder Weg stattdessen gibt.
	kaputt error

	// gesehen ist die gehütete Form, die das letzte Anlegen zu sehen bekam.
	//
	// **Damit die Zusage „im Bestand steht kein Klartext" an der Naht geprüft
	// wird und nicht am Ergebnis**: der Rückgabewert von `Anlegen` **soll** den
	// Klartext tragen — was nicht sein darf, ist, dass er auch hier ankommt.
	gesehen string
}

func neuesTokendoppel() *tokendoppel {
	return &tokendoppel{
		abgelegt:   map[int64]domain.Zugriffstoken{},
		gehuetet:   map[int64]string{},
		vorgaenger: map[int64]string{},
		naechste:   1,
	}
}

func (t *tokendoppel) TokenAnlegen(
	_ context.Context, neu domain.Zugriffstoken, gehuetet string,
) (domain.Zugriffstoken, error) {
	if t.kaputt != nil {
		return domain.Zugriffstoken{}, t.kaputt
	}

	neu.ID = t.naechste
	neu.Fassung = 1
	t.naechste++
	t.abgelegt[neu.ID] = neu
	t.gehuetet[neu.ID] = gehuetet
	t.gesehen = gehuetet

	return neu, nil
}

func (t *tokendoppel) Token(_ context.Context, profilID int64) (
	[]domain.Zugriffstoken, error,
) {
	if t.kaputt != nil {
		return nil, t.kaputt
	}

	gefunden := []domain.Zugriffstoken{}
	for kennung := int64(1); kennung < t.naechste; kennung++ {
		if eines, da := t.abgelegt[kennung]; da && eines.ProfilID == profilID {
			gefunden = append(gefunden, eines)
		}
	}

	return gefunden, nil
}

func (t *tokendoppel) TokenMitKennung(_ context.Context, tokenID int64) (
	domain.Zugriffstoken, domain.Gehuetetes, error,
) {
	if t.kaputt != nil {
		return domain.Zugriffstoken{}, domain.Gehuetetes{}, t.kaputt
	}

	eines, da := t.abgelegt[tokenID]
	if !da {
		return domain.Zugriffstoken{}, domain.Gehuetetes{}, fmt.Errorf(
			"%w: dieses Token gibt es nicht", domain.ErrNotFound)
	}

	return eines, domain.Gehuetetes{
		Aktuell: t.gehuetet[tokenID], Vorgaenger: t.vorgaenger[tokenID],
	}, nil
}

// TokenRotieren tauscht das Geheimnis und hebt das vorige bis zur Frist auf.
func (t *tokendoppel) TokenRotieren(
	_ context.Context, tokenID int64, gehuetet, vorgaengerBis string,
) (domain.Zugriffstoken, error) {
	if t.kaputt != nil {
		return domain.Zugriffstoken{}, t.kaputt
	}

	eines, da := t.abgelegt[tokenID]
	if !da {
		return domain.Zugriffstoken{}, fmt.Errorf(
			"%w: dieses Token gibt es nicht", domain.ErrNotFound)
	}

	t.vorgaenger[tokenID] = t.gehuetet[tokenID]
	t.gehuetet[tokenID] = gehuetet
	eines.VorgaengerBis = vorgaengerBis
	eines.Fassung++
	t.abgelegt[tokenID] = eines

	return eines, nil
}

// TokenNutzungVermerken schreibt die letzte Nutzung an.
func (t *tokendoppel) TokenNutzungVermerken(
	_ context.Context, tokenID int64, zeitpunkt string,
) error {
	if t.vermerkfehler != nil {
		return t.vermerkfehler
	}

	eines, da := t.abgelegt[tokenID]
	if !da {
		return fmt.Errorf("%w: dieses Token gibt es nicht", domain.ErrNotFound)
	}

	eines.LetzteNutzung = zeitpunkt
	t.abgelegt[tokenID] = eines

	return nil
}

// TokenAlleLoeschen zieht die Token eines Profils zurück.
func (t *tokendoppel) TokenAlleLoeschen(_ context.Context, profilID int64) (int64, error) {
	if t.kaputt != nil {
		return 0, t.kaputt
	}

	var gezogen int64

	for kennung, eines := range t.abgelegt {
		if eines.ProfilID == profilID {
			delete(t.abgelegt, kennung)
			gezogen++
		}
	}

	return gezogen, nil
}

func (t *tokendoppel) TokenAendern(
	_ context.Context, geaendert domain.Zugriffstoken,
) (domain.Zugriffstoken, error) {
	if t.kaputt != nil {
		return domain.Zugriffstoken{}, t.kaputt
	}

	t.abgelegt[geaendert.ID] = geaendert

	return geaendert, nil
}

func (t *tokendoppel) TokenLoeschen(_ context.Context, tokenID int64) error {
	if t.kaputt != nil {
		return t.kaputt
	}

	delete(t.abgelegt, tokenID)

	return nil
}
