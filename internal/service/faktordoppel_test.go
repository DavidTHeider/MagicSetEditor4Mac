package service_test

// **Die Doppel des zweiten Faktors** — Werkstück 7 von Phase 4.12
// (#1351).
//
// **Sie stehen eigens hier**, damit die Testdatei daneben die
// Verhaltensklassen zusammenhält und nicht die eine Stelle wird, an der jede
// Änderung stattfindet (§2).

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// faktordoppel ist Faktorablage und Geheimnisablage in einem.
//
// **Beides zusammen, weil beides zusammengehört**: ein Faktor ohne sein
// Geheimnis ist keiner, und zwei Doppel hielten zwei Stände desselben Nutzers.
type faktordoppel struct {
	faktoren  []domain.Faktor
	geheimnis string
	codes     []domain.Wiederherstellungscode
	naechste  int64
	kaputt    error

	// geheimnisfehler und verbrauchfehler treffen **einen** Weg.
	//
	// **Getrennt von `kaputt`, und das ist der Punkt**: wer alles zugleich
	// kaputt macht, betritt den Weg dahinter nie und prüft die
	// Unterscheidung nicht, um die es geht.
	geheimnisfehler error
	verbrauchfehler error
	lesefehler      error
	setzfehler      error

	// **Der Haushalt** (SEC-08 bis SEC-10, #2149–#2151): der zuletzt
	// verbrauchte Schritt, die gezählten Versuche und die Sperre — mit
	// denselben Regeln wie die Ablage, damit ein Lauf hier dasselbe sieht.
	schritt        int64
	versuche       int
	gesperrtBis    time.Time
	haushaltfehler error
	gelungenfehler error
	schrittfehler  error
	anlegefehler   error
}

func (f *faktordoppel) SchrittVerbrauchen(_ context.Context, _, schritt int64) (bool, error) {
	if f.schrittfehler != nil {
		return false, f.schrittfehler
	}

	if schritt <= f.schritt {
		return false, nil
	}

	f.schritt = schritt

	return true, nil
}

func (f *faktordoppel) FaktorversuchBeginnen(_ context.Context, _ int64, jetzt time.Time) (time.Time, error) {
	if f.haushaltfehler != nil {
		return time.Time{}, f.haushaltfehler
	}

	if jetzt.Before(f.gesperrtBis) {
		return f.gesperrtBis, nil
	}

	f.versuche++

	if dauer := domain.Faktorsperre(f.versuche); dauer > 0 {
		f.gesperrtBis = jetzt.Add(dauer)
	}

	return time.Time{}, nil
}

func (f *faktordoppel) FaktorversuchGelungen(context.Context, int64) error {
	if f.gelungenfehler != nil {
		return f.gelungenfehler
	}

	f.versuche, f.gesperrtBis = 0, time.Time{}

	return nil
}

func (f *faktordoppel) Faktoren(context.Context, int64) ([]domain.Faktor, error) {
	return f.faktoren, f.kaputt
}

func (f *faktordoppel) FaktorAnlegen(_ context.Context, neu domain.Faktor) (
	domain.Faktor, error,
) {
	if f.kaputt != nil {
		return domain.Faktor{}, f.kaputt
	}

	if f.anlegefehler != nil {
		return domain.Faktor{}, f.anlegefehler
	}

	f.naechste++
	neu.ID = f.naechste
	f.faktoren = append(f.faktoren, neu)
	f.schritt = neu.LetzterSchritt

	return neu, nil
}

func (f *faktordoppel) FaktorLoeschen(
	_ context.Context, _ int64, art domain.Faktorart,
) error {
	if f.kaputt != nil {
		return f.kaputt
	}

	geblieben := f.faktoren[:0]

	for _, einer := range f.faktoren {
		if einer.Art != art {
			geblieben = append(geblieben, einer)
		}
	}

	f.faktoren = geblieben

	return nil
}

func (f *faktordoppel) CodesSetzen(
	_ context.Context, _ int64, gehuetete []string,
) error {
	if f.kaputt != nil {
		return f.kaputt
	}

	if f.setzfehler != nil {
		return f.setzfehler
	}

	f.codes = nil

	for _, gehuetet := range gehuetete {
		f.naechste++
		f.codes = append(f.codes,
			domain.Wiederherstellungscode{ID: f.naechste, Gehuetet: gehuetet})
	}

	return nil
}

func (f *faktordoppel) FreieCodes(context.Context, int64) (
	[]domain.Wiederherstellungscode, error,
) {
	if f.lesefehler != nil {
		return nil, f.lesefehler
	}

	return f.codes, f.kaputt
}

func (f *faktordoppel) CodeVerbrauchen(_ context.Context, codeID int64) error {
	if f.kaputt != nil {
		return f.kaputt
	}

	if f.verbrauchfehler != nil {
		return f.verbrauchfehler
	}

	geblieben := f.codes[:0]

	for _, einer := range f.codes {
		if einer.ID != codeID {
			geblieben = append(geblieben, einer)
		}
	}

	if len(geblieben) == len(f.codes) {
		return fmt.Errorf("%w: diesen Code gibt es nicht mehr", domain.ErrNotFound)
	}

	f.codes = geblieben

	return nil
}

// **Das Geheimnis** — im Betrieb der Schlüsselbund, hier ein Feld.
func (f *faktordoppel) Setze(_ context.Context, _ int64, geheimnis string) error {
	if f.kaputt != nil {
		return f.kaputt
	}

	f.geheimnis = geheimnis

	return nil
}

func (f *faktordoppel) Hole(context.Context, int64) (string, error) {
	if f.kaputt != nil {
		return "", f.kaputt
	}

	if f.geheimnisfehler != nil {
		return "", f.geheimnisfehler
	}

	if f.geheimnis == "" {
		return "", fmt.Errorf(
			"%w: für diesen Nutzer liegt kein Geheimnis", domain.ErrNotFound)
	}

	return f.geheimnis, nil
}

func (f *faktordoppel) Nimm(context.Context, int64) error {
	if f.kaputt != nil {
		return f.kaputt
	}

	f.geheimnis = ""

	return nil
}

// sensordoppel ist ein Sensor, der sagt, was der Lauf will.
type sensordoppel struct {
	vorhanden bool
	fehler    error

	// gefragt ist der Grund, mit dem er zuletzt gefragt wurde — die Probe
	// darauf, dass der Satz bis in den Dialog reist.
	gefragt string
}

func (s *sensordoppel) Vorhanden(context.Context) bool { return s.vorhanden }

func (s *sensordoppel) Bestaetige(_ context.Context, grund string) error {
	s.gefragt = grund

	return s.fehler
}
