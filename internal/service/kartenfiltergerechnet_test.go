package service_test

// **Der Schnellfilter sieht, was die Liste zeigt** — FP-01
// (#2155).
//
// MSE2 durchsucht in `Card::contains` jeden Feldwert, wie er nach den
// Skripten dasteht, und dazu die Notiz (`src/data/card.cpp`, Zeilen 55–61,
// Fassung f9d9356 vom 2021-07-06). Hier sah der Filter nur Gespeichertes:
// `type` aus `magic.mse-game` wird aus Ober- und Untertyp gerechnet, die Liste
// zeigte „Creature — Beast", und weder `type:"Creature — Beast"` noch die
// freie Phrase fanden die Karte — eine Verneinung behielt sie. Der Vertrag
// steht am Filter: frei sieht, was die Liste zeigt; ein genanntes Feld wird
// gerechnet, auch wenn es keine Spalte ist.

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func gerechnetefilterkarten() *listendoppel {
	return &listendoppel{
		spiel: "probe",
		felder: []domain.Field{
			{Name: "name", Kind: domain.KindText},
			{Name: "super type", Kind: domain.KindText},
			{Name: "sub type", Kind: domain.KindText},
			// **Eine sichtbare Spalte mit Skript** — so steht `type` in
			// `magic.mse-game` (`card list visible: true`).
			{
				Name: "type", Kind: domain.KindText, Skript: `card.super_type + " — " + card.sub_type`,
				Listenangaben: domain.Listenangaben{Listenerlaubt: true, Listensichtbar: true},
			},
			// **Und eine unsichtbare**: ein freier Begriff sieht sie nicht, ein
			// genannter Feldbezug schon.
			{Name: "geheim", Kind: domain.KindText, Skript: `"verborgen"`},
		},
		karten: []*domain.CardRow{
			{ID: 1, SetID: 1, Values: []domain.Value{
				{Field: "name", Text: "Grizzly Bears"},
				{Field: "super type", Text: "Creature"},
				{Field: "sub type", Text: "Beast"},
			}},
			{ID: 2, SetID: 1, Values: []domain.Value{
				{Field: "name", Text: "Island"},
				{Field: "super type", Text: "Basic Land"},
				{Field: "sub type", Text: "Island"},
			}},
			// **Gespeichert schlägt gerechnet** — wie in der Liste.
			{ID: 3, SetID: 1, Values: []domain.Value{
				{Field: "name", Text: "Eigenwillig"},
				{Field: "super type", Text: "Creature"},
				{Field: "sub type", Text: "Beast"},
				{Field: "type", Text: "Von Hand"},
			}},
		},
	}
}

func TestDerFilterFindetGerechneteWerte(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		anfrage string
		soll    []int64
	}{
		"das Feld mit der ganzen Phrase":          {`type:"Creature — Beast"`, []int64{1}},
		"die freie Phrase":                        {`"Creature — Beast"`, []int64{1}},
		"die Verneinung ist das Gegenstück":       {`-"Creature — Beast"`, []int64{2, 3}},
		"gespeichert schlägt gerechnet":           {`type:"von hand"`, []int64{3}},
		"der gespeicherte Kontrollfall":           {`"super type":creature`, []int64{1, 3}},
		"ein unsichtbares Feld über seinen Namen": {`geheim:verborgen`, []int64{1, 2, 3}},
		"ein unsichtbares Feld nicht frei":        {`verborgen`, []int64{}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doppel := gerechnetefilterkarten()
			ist, fehler := service.NeuerKartenlistenbau(doppel, doppel, doppel, doppel, doppel).
				Filtern(context.Background(), 1, fall.anfrage)
			if fehler != nil {
				t.Fatal(fehler)
			}

			if !slices.Equal(ist, fall.soll) {
				t.Errorf("Filtern(%q) = %v, erwartet %v", fall.anfrage, ist, fall.soll)
			}
		})
	}
}

// abbrechendeKarten bricht den Lauf ab, sobald die Karten gelesen sind — vor
// der Rechnung der Felder.
type abbrechendeKarten struct {
	*listendoppel

	abbrechen context.CancelFunc
}

func (a abbrechendeKarten) Karten(ctx context.Context, setID int64) ([]*domain.CardRow, error) {
	a.abbrechen()

	return a.listendoppel.Karten(ctx, setID)
}

// **Ein Abbruch ist kein Nichttreffer** — Befund 3 des Zweitblicks über Phase
// G. Bricht der Lauf zwischen dem Lesen der Karten und der Rechnung ab, fehlen
// die gerechneten Werte; als Nichttreffer gelesen, gab die Suche `[]` und ihre
// Verneinung `[1 2 3]` — zwei verschiedene falsche Antworten ohne Fehler.
func TestEinAbbruchBeimFilternIstEinFehler(t *testing.T) {
	t.Parallel()

	for _, anfrage := range []string{`"Creature — Beast"`, `-"Creature — Beast"`} {
		t.Run(anfrage, func(t *testing.T) {
			t.Parallel()

			ctx, abbrechen := context.WithCancel(t.Context())
			defer abbrechen()

			doppel := gerechnetefilterkarten()
			karten := abbrechendeKarten{listendoppel: doppel, abbrechen: abbrechen}

			ist, fehler := service.NeuerKartenlistenbau(doppel, karten, doppel, doppel, doppel).
				Filtern(ctx, 1, anfrage)
			if !errors.Is(fehler, context.Canceled) {
				t.Errorf("Filtern(%q) = %v, %v — erwartet den Abbruch", anfrage, ist, fehler)
			}
		})
	}
}

// **Dieselbe Klasse in der Liste**: abgebrochen zwischen dem Lesen und der
// Rechnung, stand dort eine Zeile mit leeren gerechneten Zellen als Erfolg —
// eine Liste, die jemand für vollständig hält.
func TestEinAbbruchBeimListenaufbauIstEinFehler(t *testing.T) {
	t.Parallel()

	ctx, abbrechen := context.WithCancel(t.Context())
	defer abbrechen()

	doppel := gerechnetefilterkarten()
	karten := abbrechendeKarten{listendoppel: doppel, abbrechen: abbrechen}

	aufbau, fehler := service.NeuerKartenlistenbau(doppel, karten, doppel, doppel, doppel).
		Aufbau(ctx, 1)
	if !errors.Is(fehler, context.Canceled) {
		t.Errorf("Aufbau = %d Zeilen, %v — erwartet den Abbruch", len(aufbau.Zeilen), fehler)
	}
}
