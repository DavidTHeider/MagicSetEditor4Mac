package domain_test

// **Die Liste der Handelsvorgänge: finden, ordnen, eingrenzen** —
// TRADE-01.AK-02 (#1715).

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

var tag = time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

// zeile baut eine Zeile — `vor` Tage vor dem Stichtag geändert.
func zeile(vorgang string, lage domain.Handelslage, name string, vor int) domain.Handelsvorgangzeile {
	return domain.Handelsvorgangzeile{
		Angebot:    domain.Handelsangebot{Vorgang: vorgang},
		Lage:       lage,
		Gegenueber: []domain.Handelspartner{{Kennung: "peer-" + vorgang, Name: name}},
		Karten:     []string{"Lightning Bolt"},
		Marken:     []string{"messe"},
		Angelegt:   tag.AddDate(0, 0, -10-vor),
		Geaendert:  tag.AddDate(0, 0, -vor),
	}
}

func vorgaenge(zeilen []domain.Handelsvorgangzeile) []string {
	namen := make([]string, 0, len(zeilen))
	for _, z := range zeilen {
		namen = append(namen, z.Angebot.Vorgang)
	}

	return namen
}

func TestDerFilterGrenztEin(t *testing.T) {
	t.Parallel()

	archiviert := zeile("alt", domain.LageAbgeschlossen, "Turfer", 30)
	archiviert.Archiviert = true
	ohneKarte := zeile("leer", domain.LageWartet, "Ümit", 2)
	ohneKarte.Karten, ohneKarte.Marken = nil, nil

	alle := []domain.Handelsvorgangzeile{
		zeile("v1", domain.LageDran, "Turfer", 1), ohneKarte,
		zeile("v3", domain.LageAbgebrochen, "anna", 5), archiviert,
	}

	for _, fall := range []struct {
		name     string
		filter   domain.Handelsfilter
		erwartet []string
	}{
		{"ohne Angaben alles außer dem Archiv", domain.Handelsfilter{}, []string{"v1", "leer", "v3"}},
		{"mit dem Archiv", domain.Handelsfilter{Archiv: true}, []string{"v1", "leer", "v3", "alt"}},
		{"nach Gegenüber", domain.Handelsfilter{Partner: "peer-v3"}, []string{"v3"}},
		{"nach Lage", domain.Handelsfilter{Lagen: []domain.Handelslage{domain.LageDran, domain.LageWartet}}, []string{"v1", "leer"}},
		{"Suche nach Karte ohne Groß und Klein", domain.Handelsfilter{Suche: "BOLT"}, []string{"v1", "v3"}},
		{"Suche nach dem Namen des Gegenübers", domain.Handelsfilter{Suche: "ümit"}, []string{"leer"}},
		{"Suche nach der Kennung", domain.Handelsfilter{Suche: "v3"}, []string{"v3"}},
		{"nach Marke in Normalform", domain.Handelsfilter{Marke: "  Messe "}, []string{"v1", "v3"}},
		{"seit einem Tag", domain.Handelsfilter{Seit: tag.AddDate(0, 0, -2)}, []string{"v1", "leer"}},
		{"bis zu einem Tag", domain.Handelsfilter{Bis: tag.AddDate(0, 0, -2)}, []string{"leer", "v3"}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if gesehen := vorgaenge(fall.filter.Gesehen(alle)); !slices.Equal(gesehen, fall.erwartet) {
				t.Errorf("gesehen %v, erwartet %v", gesehen, fall.erwartet)
			}
		})
	}
}

func TestDieOrdnungDerListe(t *testing.T) {
	t.Parallel()

	alle := []domain.Handelsvorgangzeile{
		zeile("b", domain.LageWartet, "Zora", 1),
		zeile("a", domain.LageAbgebrochen, "Ängla", 3),
		zeile("c", domain.LageDran, "Björn", 2),
		zeile("d", domain.LageDran, "", 4),
	}

	for _, fall := range []struct {
		name     string
		filter   domain.Handelsfilter
		erwartet []string
	}{
		{"zuletzt geändert zuerst", domain.Handelsfilter{}, []string{"b", "c", "a", "d"}},
		{"zuletzt angelegt zuerst", domain.Handelsfilter{Ordnung: domain.OrdnungAngelegt}, []string{"b", "c", "a", "d"}},
		{"was zu tun ist zuerst", domain.Handelsfilter{Ordnung: domain.OrdnungLage}, []string{"c", "d", "b", "a"}},
		// **Ohne Namen zählt die Kennung** — „peer-d" steht nach „Björn".
		{"nach Gegenüber, deutsch", domain.Handelsfilter{Ordnung: domain.OrdnungPartner}, []string{"a", "c", "d", "b"}},
		// **Im Schwedischen steht »Ä« hinter »Z«** — die Ordnung ist die der Sprache.
		{"nach Gegenüber, schwedisch", domain.Handelsfilter{Ordnung: domain.OrdnungPartner, Sprache: "sv"}, []string{"c", "d", "b", "a"}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if gesehen := vorgaenge(fall.filter.Gesehen(alle)); !slices.Equal(gesehen, fall.erwartet) {
				t.Errorf("gesehen %v, erwartet %v", gesehen, fall.erwartet)
			}
		})
	}

	if vorgaenge(alle)[0] != "b" {
		t.Error("die gelesene Liste wurde umgeordnet")
	}
}

func TestDerFilterWeistAb(t *testing.T) {
	t.Parallel()

	lang := make([]rune, 201)
	for i := range lang {
		lang[i] = 'x'
	}

	for _, fall := range []struct {
		name   string
		filter domain.Handelsfilter
		feld   string
	}{
		{"eine unbekannte Lage", domain.Handelsfilter{Lagen: []domain.Handelslage{"irgendwo"}}, "lagen"},
		{"eine unbekannte Ordnung", domain.Handelsfilter{Ordnung: "zufall"}, "ordnung"},
		{"eine zu lange Suche", domain.Handelsfilter{Suche: string(lang)}, "suche"},
		{"eine ungültige Marke", domain.Handelsfilter{Marke: string(lang)}, "marke"},
		{"ein Ende vor dem Anfang", domain.Handelsfilter{Seit: tag, Bis: tag.AddDate(0, 0, -1)}, "bis"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			var feldfehler *domain.FieldError
			if fehler := fall.filter.Pruefe(); !errors.As(fehler, &feldfehler) || feldfehler.Field != fall.feld {
				t.Errorf("Fehler ist %v, erwartet am Feld %q", fehler, fall.feld)
			}
		})
	}

	if fehler := (domain.Handelsfilter{Lagen: domain.Handelslagen(), Ordnung: domain.OrdnungLage}).Pruefe(); fehler != nil {
		t.Errorf("ein gültiger Filter: %v", fehler)
	}
}

// **Bei gleicher Änderung entscheidet die Kennung**, und ein Vorgang ohne
// Gegenüber steht beim Ordnen nach Namen vorn — sonst spränge die Liste.
func TestDieOrdnungBleibtBeiGleichstandStehen(t *testing.T) {
	t.Parallel()

	ohne := zeile("a-ohne", domain.LageDran, "", 1)
	ohne.Gegenueber = nil

	alle := []domain.Handelsvorgangzeile{
		zeile("z", domain.LageDran, "Björn", 1), zeile("m", domain.LageDran, "Björn", 1), ohne,
	}

	if gesehen := vorgaenge(domain.Handelsfilter{Ordnung: domain.OrdnungPartner}.Gesehen(alle)); !slices.Equal(
		gesehen, []string{"a-ohne", "m", "z"}) {
		t.Errorf("gesehen %v", gesehen)
	}
}
