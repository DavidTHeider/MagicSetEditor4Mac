package app_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

type massenprobe struct {
	karten   []int64
	erwartet []domain.Feldvorwert
	feld     string
	wert     string
	fehler   error
}

func (m *massenprobe) FelderVorschau(
	_ context.Context, _ int64, karten []int64, feld, wert string,
) (domain.Massenvorschau, error) {
	m.karten, m.feld, m.wert = karten, feld, wert

	if m.fehler != nil {
		return domain.Massenvorschau{}, m.fehler
	}

	return domain.Massenvorschau{Feld: feld, Wert: wert, Zeilen: []domain.Massenvorschauzeile{
		{KarteID: 1, Name: "Feuerball", Alt: "Zauber", GabAlt: true, Neu: wert},
		{KarteID: 3, Name: "Land", Alt: wert, GabAlt: true, Neu: wert, Ausschluss: domain.AusschlussGleich},
	}}, nil
}

func (m *massenprobe) FelderSetzenMitErwartung(
	_ context.Context, _ int64, erwartet []domain.Feldvorwert, feld, wert string,
) (string, error) {
	m.erwartet, m.feld, m.wert = erwartet, feld, wert

	return "vorgang8", m.fehler
}

func mitMassenschreiber(probe *massenprobe) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitMassenschreiber(probe)

	return anwendung
}
