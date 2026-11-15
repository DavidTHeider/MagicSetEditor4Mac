package service_test

// **Die Vorrichtung der Datenpflege** — Werkstück O8 von Phase 5.2
// (#1432).
//
// **Eine eigene Datei**, weil `datenpflege_test.go` mit ihr über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort steht, **was**
// geprüft wird, hier, **womit**.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// pflegedoppel spielt Bestand: es merkt sich, was gelesen und geschrieben wurde.
type pflegedoppel struct {
	staende          [][]domain.Pflegebefund
	gelesen          int
	geschrieben      map[string]string
	gelaufen         []domain.Pflegezeile
	gesichert        int
	lesefehler       error
	schreibfehler    error
	sicherungsfehler error
}

func (p *pflegedoppel) lesen(context.Context, time.Time) ([]domain.Pflegebefund, error) {
	if p.lesefehler != nil {
		return nil, p.lesefehler
	}

	stand := p.staende[min(p.gelesen, len(p.staende)-1)]
	p.gelesen++

	return stand, nil
}

func (p *pflegedoppel) schreiben(
	_ context.Context, _ time.Time, revisionen map[string]string,
) ([]domain.Pflegezeile, error) {
	p.geschrieben = revisionen

	return p.gelaufen, p.schreibfehler
}

func (p *pflegedoppel) sichern(context.Context) (string, error) {
	p.gesichert++

	return "/wo/die/sicherung/liegt.db", p.sicherungsfehler
}

func pflege(doppel *pflegedoppel) *service.Datenpflege {
	weg := service.NeueDatenpflege(doppel.lesen, doppel.schreiben, doppel.sichern)
	weg.MitUhr(func() time.Time { return pflegezeit })

	return weg
}

func befund(art string, klasse domain.Pflegeklasse, revision string) domain.Pflegebefund {
	return domain.Pflegebefund{Art: art, Klasse: klasse, Anzahl: 1, Revision: revision}
}
