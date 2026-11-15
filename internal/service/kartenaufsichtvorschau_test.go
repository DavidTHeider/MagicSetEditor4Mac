package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// wertesicht gibt Karten samt Werten — die Vorschau braucht sie.
type wertesicht struct {
	karten map[int64]*domain.CardRow
}

func (w *wertesicht) KarteLesen(_ context.Context, karteID int64) (*domain.CardRow, error) {
	karte, da := w.karten[karteID]
	if !da {
		return nil, domain.ErrNotFound
	}

	return karte, nil
}

func vorschauaufsicht(t *testing.T, felder []domain.Field) (*service.Kartenaufsicht, *massenablage) {
	t.Helper()

	sicht := &wertesicht{karten: map[int64]*domain.CardRow{
		1: {ID: 1, SetID: 10, Values: []domain.Value{
			{Field: "name", Text: "Feuerball"}, {Field: "typ", Text: "Zauber"},
		}},
		2: {ID: 2, SetID: 10, Values: []domain.Value{
			{Field: "name", Text: "Eisblitz"},
		}},
		3: {ID: 3, SetID: 10, Values: []domain.Value{
			{Field: "name", Text: "Land"}, {Field: "typ", Text: "Land"},
		}},
		4: {ID: 4, SetID: 20, Values: []domain.Value{{Field: "name", Text: "Fremd"}}},
	}}
	ablage := &massenablage{}

	return service.NeueKartenaufsicht(sicht, nil).
		MitFeldweg(nil, ablage).
		MitFeldformen(feldformen{felder: felder}), ablage
}

// **Die Vorschau nennt je Karte Name, Alt, Neu und Ausschluss** — in der
// Reihenfolge, in der die Karten genannt wurden (Werkstück U6, #1417).
func TestDieVorschauNenntZielkartenAltNeuUndAusschluesse(t *testing.T) {
	aufsicht, ablage := vorschauaufsicht(t, []domain.Field{
		{Name: "typ", Editierbar: true, Gespeichert: true},
	})

	vorschau, err := aufsicht.FelderVorschau(t.Context(), 10, []int64{3, 1, 2}, "typ", "Land")
	if err != nil {
		t.Fatalf("Vorschau: %v", err)
	}

	if vorschau.Feld != "typ" || vorschau.Wert != "Land" {
		t.Errorf("Kopf: %+v", vorschau)
	}

	soll := []domain.Massenvorschauzeile{
		{KarteID: 3, Name: "Land", Alt: "Land", GabAlt: true, Neu: "Land", Ausschluss: domain.AusschlussGleich},
		{KarteID: 1, Name: "Feuerball", Alt: "Zauber", GabAlt: true, Neu: "Land"},
		{KarteID: 2, Name: "Eisblitz", Neu: "Land"},
	}

	if len(vorschau.Zeilen) != len(soll) {
		t.Fatalf("%d Zeilen statt %d: %+v", len(vorschau.Zeilen), len(soll), vorschau.Zeilen)
	}

	for i := range soll {
		if vorschau.Zeilen[i] != soll[i] {
			t.Errorf("Zeile %d:\nist  %+v\nsoll %+v", i, vorschau.Zeilen[i], soll[i])
		}
	}

	// **Eine Vorschau schreibt nichts.**
	if ablage.geschrieben != 0 {
		t.Errorf("die Vorschau hat geschrieben: %d", ablage.geschrieben)
	}
}

// **Eine fremde Karte macht die ganze Vorschau zum Fehler** — dieselbe
// Setgrenze wie beim Schreiben (0041).
func TestDieVorschauWeistEineFremdeKarteAb(t *testing.T) {
	aufsicht, _ := vorschauaufsicht(t, nil)

	if _, err := aufsicht.FelderVorschau(t.Context(), 10, []int64{1, 4}, "typ", "Land"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet war ErrNotFound", err)
	}
}

// **Ein Feld, das sich nicht über viele Karten setzen lässt, hat keine
// Vorschau** — dieselbe Regel wie beim Schreiben (#720).
func TestDieVorschauKenntDieselbenFeldregeln(t *testing.T) {
	aufsicht, _ := vorschauaufsicht(t, []domain.Field{
		{Name: "typ", Editierbar: false, Gespeichert: true},
	})

	_, err := aufsicht.FelderVorschau(t.Context(), 10, []int64{1}, "typ", "Land")

	var feldfehler *domain.FieldError
	if !errors.As(err, &feldfehler) || feldfehler.Key != domain.SchluesselNichtEditierbar {
		t.Errorf("Fehler %v, erwartet war ein Feldbefund „nicht editierbar“", err)
	}
}

// **Das Anwenden mit Erwartung geht denselben Weg wie das Schreiben** —
// Feldregeln, Setgrenze, dann die Ablage.
func TestDasAnwendenMitErwartungHaeltDieSetgrenze(t *testing.T) {
	aufsicht, ablage := vorschauaufsicht(t, []domain.Field{
		{Name: "typ", Editierbar: true, Gespeichert: true},
	})

	if _, err := aufsicht.FelderSetzenMitErwartung(t.Context(), 10,
		[]domain.Feldvorwert{{KarteID: 1, Wert: "Zauber", GabEs: true}, {KarteID: 4}},
		"typ", "Land"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet war ErrNotFound", err)
	}

	if ablage.geschrieben != 0 {
		t.Errorf("geschrieben wurde trotzdem: %d", ablage.geschrieben)
	}

	vorgang, err := aufsicht.FelderSetzenMitErwartung(t.Context(), 10,
		[]domain.Feldvorwert{{KarteID: 1, Wert: "Zauber", GabEs: true}}, "typ", "Land")
	if err != nil {
		t.Fatalf("anwenden: %v", err)
	}

	if vorgang != "vorgang8" || ablage.geschrieben != 1 {
		t.Errorf("Vorgang %q, geschrieben %d", vorgang, ablage.geschrieben)
	}
}

// **Das Anwenden kennt dieselben Feldregeln wie die Vorschau** — ein Feld,
// das sich nicht editieren lässt, wird abgewiesen, bevor eine Karte
// gelesen ist; und eine Karte aus einem anderen Set kommt nicht durch.
func TestDasAnwendenMitErwartungKenntFeldregelnUndFremdeKarten(t *testing.T) {
	gesperrt, ablage := vorschauaufsicht(t, []domain.Field{
		{Name: "typ", Editierbar: false, Gespeichert: true},
	})

	_, err := gesperrt.FelderSetzenMitErwartung(t.Context(), 10,
		[]domain.Feldvorwert{{KarteID: 1, Wert: "Zauber", GabEs: true}}, "typ", "Land")

	var feldfehler *domain.FieldError
	if !errors.As(err, &feldfehler) || feldfehler.Key != domain.SchluesselNichtEditierbar {
		t.Errorf("Fehler %v, erwartet war ein Feldbefund „nicht editierbar“", err)
	}

	offen, _ := vorschauaufsicht(t, []domain.Field{
		{Name: "typ", Editierbar: true, Gespeichert: true},
	})

	if _, fremd := offen.FelderSetzenMitErwartung(t.Context(), 10,
		[]domain.Feldvorwert{{KarteID: 4, Wert: "Fremd", GabEs: true}}, "typ", "Land"); !errors.Is(fremd, domain.ErrNotFound) {
		t.Errorf("die fremde Karte kam durch: %v", fremd)
	}

	if _, unbekannt := offen.FelderVorschau(t.Context(), 10, []int64{99}, "typ", "Land"); !errors.Is(unbekannt, domain.ErrNotFound) {
		t.Errorf("eine unbekannte Karte gab eine Vorschau: %v", unbekannt)
	}

	if ablage.geschrieben != 0 {
		t.Errorf("geschrieben wurde trotzdem: %d", ablage.geschrieben)
	}
}
