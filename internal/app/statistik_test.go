package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

type statistikprobe struct {
	zahlen     domain.SetStats
	belegung   []domain.FieldCount
	verteilung []domain.ValueCount
	setID      int64
	feld       string

	zahlenFehler     error
	belegungFehler   error
	verteilungFehler error
}

func (s *statistikprobe) Kennzahlen(_ context.Context, setID int64) (domain.SetStats, error) {
	s.setID = setID

	return s.zahlen, s.zahlenFehler
}

func (s *statistikprobe) Belegung(_ context.Context, _ int64) ([]domain.FieldCount, error) {
	return s.belegung, s.belegungFehler
}

func (s *statistikprobe) Verteilung(_ context.Context, setID int64, feld string) (
	[]domain.ValueCount, error,
) {
	s.setID, s.feld = setID, feld

	return s.verteilung, s.verteilungFehler
}

func mitStatistiker(probe *statistikprobe) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitStatistiker(probe)

	return anwendung
}

func TestStatistikGibtDieKennzahlen(t *testing.T) {
	probe := &statistikprobe{
		zahlen:   domain.SetStats{Cards: 12, Fields: 4, Empty: 2},
		belegung: []domain.FieldCount{{Field: "name", Cards: 12}, {Field: "typ", Cards: 7}},
	}

	antwort := mitStatistiker(probe).Statistik(probekontext(t), "3")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if probe.setID != 3 {
		t.Errorf("SetID %d", probe.setID)
	}

	if antwort.Statistik.Karten != 12 || antwort.Statistik.Felder != 4 ||
		antwort.Statistik.Leere != 2 {
		t.Errorf("%+v", antwort.Statistik)
	}

	if len(antwort.Statistik.Belegung) != 2 ||
		antwort.Statistik.Belegung[1] != (app.ZahlDTO{Name: "typ", Karten: 7}) {
		t.Errorf("%+v", antwort.Statistik.Belegung)
	}
}

func TestStatistikWeistEineUnlesbareKennungAb(t *testing.T) {
	befund := mitStatistiker(&statistikprobe{}).Statistik(probekontext(t), "abc").Fehler
	if befund == nil || befund.Feld != "id" {
		t.Errorf("%+v", befund)
	}
}

// Beide Abfragen können scheitern, und beide Fehler kommen an.
func TestStatistikMeldetBeideFehler(t *testing.T) {
	fuer := map[string]*statistikprobe{
		"Kennzahlen": {zahlenFehler: errors.New("nicht lesbar")},
		"Belegung":   {belegungFehler: errors.New("nicht lesbar")},
	}

	for name, probe := range fuer {
		t.Run(name, func(t *testing.T) {
			if mitStatistiker(probe).Statistik(probekontext(t), "1").Fehler == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Eine leere Belegung ist `[]`, nicht `null`** (§3, „Zwischen den Prüfebenen").
func TestEineLeereBelegungIstEineListe(t *testing.T) {
	antwort := mitStatistiker(&statistikprobe{}).Statistik(probekontext(t), "1")

	gebaut, err := json.Marshal(antwort.Statistik.Belegung)
	if err != nil {
		t.Fatalf("serialisieren: %v", err)
	}

	if string(gebaut) != "[]" {
		t.Errorf("%s", gebaut)
	}
}

func TestVerteilungGibtDieWerte(t *testing.T) {
	probe := &statistikprobe{
		verteilung: []domain.ValueCount{{Value: "Zauber", Cards: 5, CardIDs: []int64{4, 9}}},
	}

	antwort := mitStatistiker(probe).Verteilung(probekontext(t), "3", "typ")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if probe.setID != 3 || probe.feld != "typ" {
		t.Errorf("%d / %q", probe.setID, probe.feld)
	}

	if len(antwort.Verteilung) != 1 || antwort.Verteilung[0].Name != "Zauber" ||
		antwort.Verteilung[0].Karten != 5 {
		t.Errorf("%+v", antwort.Verteilung)
	}

	// **Und der Wert nennt seine Karten — als Zeichenketten** (UXR-F6, #2013).
	if !slices.Equal(antwort.Verteilung[0].KartenIDs, []string{"4", "9"}) {
		t.Errorf("die Karten lauten %v", antwort.Verteilung[0].KartenIDs)
	}
}

func TestVerteilungMeldetBefunde(t *testing.T) {
	if befund := mitStatistiker(&statistikprobe{}).
		Verteilung(probekontext(t), "abc", "typ").Fehler; befund == nil || befund.Feld != "id" {
		t.Errorf("%+v", befund)
	}

	probe := &statistikprobe{verteilungFehler: errors.New("nicht lesbar")}
	if mitStatistiker(probe).Verteilung(probekontext(t), "1", "typ").Fehler == nil {
		t.Error("der Fehler ging verloren")
	}
}

// Auch eine leere Verteilung ist eine Liste.
func TestEineLeereVerteilungIstEineListe(t *testing.T) {
	antwort := mitStatistiker(&statistikprobe{}).Verteilung(probekontext(t), "1", "typ")

	gebaut, err := json.Marshal(antwort.Verteilung)
	if err != nil {
		t.Fatalf("serialisieren: %v", err)
	}

	if string(gebaut) != "[]" {
		t.Errorf("%s", gebaut)
	}
}
