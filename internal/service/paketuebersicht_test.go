package service_test

// **Die Übersicht über die Pakete** — DOM-01.AK-01
// (#1696):
// je Paket, woher es stammt, wer es braucht und wie viele Sets es benutzen.

import (
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieUebersichtNenntHerkunftBedarfUndNutzung(t *testing.T) {
	verwaltet := t.TempDir()
	katalog := &ablage{pakete: []domain.Paket{
		{Art: "mse-game", Name: "magic", Pfad: "/saat/magic.mse-game"},
		{
			Art: "mse-style", Name: "magic-m15", Pfad: filepath.Join(verwaltet, "magic-m15.mse-style"),
			Braucht: []domain.Bedarf{{Paket: "magic.mse-game"}},
		},
		{Art: "mse-style", Name: "magic-alt", Pfad: "/bezug/magic-alt.mse-style", Pruefsumme: "sha256:1"},
	}}

	verwalter := service.NeuerPaketverwalter(katalog, verwaltet, verwaltet).
		MitSetkoepfen(setkoepfe{koepfe: []domain.SetHead{
			// **Die Kurzform und der volle Name** — beide meinen dasselbe Paket
			// (`stilpaketnamen`).
			{Game: "magic", Style: "m15"},
			{Game: "magic", Style: "magic-m15"},
			{Game: "magic", Style: ""},
		}})

	zeilen, err := verwalter.Uebersicht(t.Context())
	if err != nil {
		t.Fatalf("Übersicht: %v", err)
	}

	nach := map[string]service.Paketzeile{}
	for _, zeile := range zeilen {
		nach[zeile.Paket.Dateiname()] = zeile
	}

	spiel, m15, alt := nach["magic.mse-game"], nach["magic-m15.mse-style"], nach["magic-alt.mse-style"]

	if spiel.Herkunftsart != domain.PaketMitgeliefert || spiel.Benutzt != 3 ||
		!slices.Equal(spiel.GebrauchtVon, []string{"magic-m15.mse-style"}) {
		t.Errorf("das Spiel: %+v", spiel)
	}

	if m15.Herkunftsart != domain.PaketEigen || m15.Benutzt != 2 || len(m15.GebrauchtVon) != 0 {
		t.Errorf("der Stil m15: %+v", m15)
	}

	if alt.Herkunftsart != domain.PaketBezogen || alt.Benutzt != 0 {
		t.Errorf("der bezogene Stil: %+v", alt)
	}
}

// **Ohne Setköpfe gibt es keine Nutzung, aber eine Übersicht** — und ein
// Fehler beim Lesen der Köpfe oder des Katalogs kommt durch.
func TestDieUebersichtOhneUndMitFehlern(t *testing.T) {
	katalog := &ablage{pakete: []domain.Paket{{Art: "mse-game", Name: "magic", Pfad: "/saat/magic.mse-game"}}}

	zeilen, err := service.NeuerPaketverwalter(katalog, "").Uebersicht(t.Context())
	if err != nil || len(zeilen) != 1 || zeilen[0].Benutzt != 0 {
		t.Errorf("ohne Setköpfe: %+v, %v", zeilen, err)
	}

	klemmt := errors.New("klemmt")

	if _, err := service.NeuerPaketverwalter(katalog, "").MitSetkoepfen(setkoepfe{fehler: klemmt}).
		Uebersicht(t.Context()); !errors.Is(err, klemmt) {
		t.Errorf("die Köpfe klemmen: %v", err)
	}

	if _, err := service.NeuerPaketverwalter(&ablage{lesefehler: klemmt}, "").Uebersicht(t.Context()); !errors.Is(err, klemmt) {
		t.Errorf("der Katalog klemmt: %v", err)
	}
}
