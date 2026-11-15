package domain_test

// **Was eine Rückkehr kostet, und was sie unklar lässt** — Werkstück O7 von
// Phase 5.2 (#1431).

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Verlust ist die Differenz, und nur nach oben.** Was seit der Sicherung
// **dazugekommen** ist, verschwindet; was weniger geworden ist, kommt zurück —
// das ist kein Verlust, sondern der Zweck.
func TestDerVerlustIstWasSeitDerSicherungDazukam(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name    string
		zeile   domain.Vergleichszeile
		verlust int
		unklar  bool
	}{
		{"zwölf Karten seither", domain.Vergleichszeile{Gegenstand: "karten", InSicherung: 300, Jetzt: 312}, 12, false},
		{"nichts geändert", domain.Vergleichszeile{Gegenstand: "sets", InSicherung: 5, Jetzt: 5}, 0, false},
		{"seither gelöscht — kein Verlust", domain.Vergleichszeile{Gegenstand: "sets", InSicherung: 7, Jetzt: 5}, 0, false},
		{"die Sicherung kennt die Tabelle nicht", domain.Vergleichszeile{Gegenstand: "decks", InSicherung: -1, Jetzt: 3}, 0, true},
		{"der Bestand sagt nichts", domain.Vergleichszeile{Gegenstand: "decks", InSicherung: 3, Jetzt: -1}, 0, true},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if verlust := fall.zeile.Verlust(); verlust != fall.verlust {
				t.Errorf("Verlust ist %d, erwartet war %d", verlust, fall.verlust)
			}

			if unklar := fall.zeile.Unbekannt(); unklar != fall.unklar {
				t.Errorf("Unbekannt ist %v, erwartet war %v", unklar, fall.unklar)
			}
		})
	}
}

// **Eine Rückkehr macht Mengen unklar — nicht falsch.**
//
// Das ist die dritte Antwort aus dem Issue: ein historisches Backup kann
// Karten enthalten, die längst abgegeben sind. Sie stehen wieder da, und
// niemand weiß, ob sie noch jemandem gehören.
func TestWasNachEinerRueckkehrWiederDastehtIstUnklarUndNichtVerfuegbar(t *testing.T) {
	t.Parallel()

	sperre := domain.NeueHandelssperre([]domain.Sammlungsmenge{
		{Schluessel: "scryfall:abc", Zustand: "nm", Sprache: "en", Menge: 3},
	})

	for _, fall := range []struct {
		name       string
		schluessel string
		zustand    string
		sprache    string
		besitz     int
		frei       int
	}{
		{"drei von vier sind unklar", "scryfall:abc", "nm", "en", 4, 1},
		{"mehr unklar als besessen gibt null, nicht negativ", "scryfall:abc", "nm", "en", 2, 0},
		{"ein anderer Zustand ist nicht betroffen", "scryfall:abc", "lp", "en", 4, 4},
		{"eine andere Sprache ist nicht betroffen", "scryfall:abc", "nm", "de", 4, 4},
		{"ein anderer Druck ist nicht betroffen", "scryfall:xyz", "nm", "en", 4, 4},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			frei := sperre.Freigegeben(domain.Sammlungsmenge{
				Schluessel: fall.schluessel, Zustand: fall.zustand, Sprache: fall.sprache,
			}, fall.besitz)

			if frei != fall.frei {
				t.Errorf("frei sind %d von %d, erwartet war %d", frei, fall.besitz, fall.frei)
			}
		})
	}
}

// **Foil und nicht-Foil sind zwei Gegenstände.** Wer das zusammenwirft, gibt
// eine Karte frei, die es so nicht gibt.
func TestFoilUndNichtFoilSindZweiMengen(t *testing.T) {
	t.Parallel()

	sperre := domain.NeueHandelssperre([]domain.Sammlungsmenge{
		{Schluessel: "scryfall:abc", Zustand: "nm", Sprache: "en", Finish: domain.FinishFoil, Menge: 2},
	})

	if frei := sperre.Freigegeben(domain.Sammlungsmenge{
		Schluessel: "scryfall:abc", Zustand: "nm", Sprache: "en",
	}, 2); frei != 2 {
		t.Errorf("die nicht-Foil-Zeile wurde mitgesperrt: %d frei von 2", frei)
	}
}

// **Ohne offene Epoche ist nichts gesperrt** — und das ist der Normalfall.
func TestOhneEpocheIstAllesFrei(t *testing.T) {
	t.Parallel()

	sperre := domain.NeueHandelssperre(nil)

	if frei := sperre.Freigegeben(domain.Sammlungsmenge{Schluessel: "scryfall:abc"}, 9); frei != 9 {
		t.Errorf("ohne Epoche waren nur %d von 9 frei", frei)
	}

	if sperre.Offen() {
		t.Error("eine leere Sperre meldet sich als offen")
	}
}

// **Und eine offene Epoche sagt, wie viel sie betrifft** — die Zahl gehört in
// die Oberfläche, denn „irgendetwas ist unklar" ist keine Auskunft.
func TestEineOffeneSperreNenntIhreMenge(t *testing.T) {
	t.Parallel()

	sperre := domain.NeueHandelssperre([]domain.Sammlungsmenge{
		{Schluessel: "a", Menge: 3},
		{Schluessel: "b", Menge: 4},
	})

	if !sperre.Offen() {
		t.Error("die Sperre meldet sich nicht als offen")
	}

	if summe := sperre.Menge(); summe != 7 {
		t.Errorf("die Sperre betrifft %d Stück, erwartet waren 7", summe)
	}
}

// **Eine quittierte Epoche sperrt nichts mehr** — und sie **verschwindet
// nicht**: sie steht mit ihrem Zeitpunkt da, damit später nachlesbar ist,
// dass jemand nachgesehen hat.
func TestEineQuittierteEpocheSperrtNichtMehrUndBleibtLesbar(t *testing.T) {
	t.Parallel()

	epoche := domain.Wiederherstellungsepoche{
		ID:              7,
		Zurueckgespielt: time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC),
		Datenstand:      time.Date(2026, 9, 1, 8, 30, 0, 0, time.UTC),
		Quelle:          "bestand-2026-09-01.msesicherung",
		Klaerung:        domain.KlaerungOffen,
	}

	if !epoche.Sperrt() {
		t.Error("eine offene Epoche sperrt nicht")
	}

	epoche.Klaerung = domain.KlaerungQuittiert
	epoche.GeklaertAm = time.Date(2026, 9, 10, 13, 0, 0, 0, time.UTC)

	if epoche.Sperrt() {
		t.Error("eine quittierte Epoche sperrt weiter")
	}

	if epoche.Quelle == "" || epoche.Datenstand.IsZero() {
		t.Error("die quittierte Epoche hat ihre Angaben verloren")
	}
}

// **Und sie verjährt nicht.** Die Unsicherheit verschwindet nicht mit der
// Zeit und nicht beim nächsten Start; das Issue verlangt es ausdrücklich.
func TestEineEpocheVerjaehrtNicht(t *testing.T) {
	t.Parallel()

	epoche := domain.Wiederherstellungsepoche{
		Zurueckgespielt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		Klaerung:        domain.KlaerungOffen,
	}

	if !epoche.Sperrt() {
		t.Error("eine sechs Jahre alte Epoche hat sich selbst erledigt")
	}
}
