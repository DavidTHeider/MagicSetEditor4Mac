package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Vorschauzeile einer Massenänderung** — Werkstück U6 von Phase 5.1
// (#1417):
// je Karte Name, Alt und Neu, und ein Ausschluss, wo nichts zu tun ist.
func TestVorschauzeileNenntAltNeuUndAusschluss(t *testing.T) {
	faelle := []struct {
		name  string
		karte domain.CardRow
		feld  string
		wert  string
		soll  domain.Massenvorschauzeile
	}{
		{
			name: "ein Feld mit anderem Wert wird geändert",
			karte: domain.CardRow{ID: 7, Values: []domain.Value{
				{Field: "name", Text: "Feuerball"}, {Field: "typ", Text: "Zauber"},
			}},
			feld: "typ", wert: "Verzauberung",
			soll: domain.Massenvorschauzeile{
				KarteID: 7, Name: "Feuerball", Alt: "Zauber", GabAlt: true, Neu: "Verzauberung",
			},
		},
		{
			name: "ein Feld, das die Karte nicht hat, kommt hinzu",
			karte: domain.CardRow{ID: 8, Values: []domain.Value{
				{Field: "name", Text: "Eisblitz"},
			}},
			feld: "typ", wert: "Zauber",
			soll: domain.Massenvorschauzeile{
				KarteID: 8, Name: "Eisblitz", GabAlt: false, Neu: "Zauber",
			},
		},
		{
			name: "derselbe Wert ist ausgeschlossen",
			karte: domain.CardRow{ID: 9, Values: []domain.Value{
				{Field: "name", Text: "Blitz"}, {Field: "typ", Text: "Zauber"},
			}},
			feld: "typ", wert: "Zauber",
			soll: domain.Massenvorschauzeile{
				KarteID: 9, Name: "Blitz", Alt: "Zauber", GabAlt: true, Neu: "Zauber",
				Ausschluss: domain.AusschlussGleich,
			},
		},
		{
			name:  "eine Karte ohne Namen heißt nach ihrer Kennung",
			karte: domain.CardRow{ID: 10},
			feld:  "typ", wert: "Land",
			soll: domain.Massenvorschauzeile{KarteID: 10, Name: "", Neu: "Land"},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if ist := domain.Vorschauzeile(&fall.karte, fall.feld, fall.wert); ist != fall.soll {
				t.Errorf("\nist  %+v\nsoll %+v", ist, fall.soll)
			}
		})
	}
}

// **Die Erwartung einer Vorschau** ist, was beim Anwenden mitgeht: je Karte
// der Altwert, den die Vorschau gezeigt hat — und ob es ihn gab.
func TestErwartungAusEinerVorschau(t *testing.T) {
	vorschau := domain.Massenvorschau{Feld: "typ", Wert: "Land", Zeilen: []domain.Massenvorschauzeile{
		{KarteID: 1, Alt: "Zauber", GabAlt: true, Neu: "Land"},
		{KarteID: 2, GabAlt: false, Neu: "Land"},
		{KarteID: 3, Alt: "Land", GabAlt: true, Neu: "Land", Ausschluss: domain.AusschlussGleich},
	}}

	erwartet := vorschau.Erwartung()

	soll := []domain.Feldvorwert{
		{KarteID: 1, Wert: "Zauber", GabEs: true},
		{KarteID: 2, Wert: "", GabEs: false},
	}

	if len(erwartet) != len(soll) {
		t.Fatalf("%d Erwartungen statt %d: %+v", len(erwartet), len(soll), erwartet)
	}

	for i := range soll {
		if erwartet[i] != soll[i] {
			t.Errorf("Zeile %d: %+v statt %+v", i, erwartet[i], soll[i])
		}
	}
}
