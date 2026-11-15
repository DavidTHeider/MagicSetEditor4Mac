// **Die Setgrenze auf dem Lebenszyklus einer Karte** (#801).
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// kartensicht gibt eine Karte samt ihrem Set.
type kartensicht struct {
	karten map[int64]int64
}

func (k *kartensicht) KarteLesen(_ context.Context, karteID int64) (*domain.CardRow, error) {
	setID, da := k.karten[karteID]
	if !da {
		return nil, domain.ErrNotFound
	}

	return &domain.CardRow{ID: karteID, SetID: setID}, nil
}

// lebenszyklusprotokoll merkt sich, was durchgekommen ist.
type lebenszyklusprotokoll struct {
	angelegt  []int64
	kopiert   []int64
	geloescht []int64
	getauscht [][2]int64
}

func (l *lebenszyklusprotokoll) KarteAnlegen(_ context.Context, setID int64) (int64, error) {
	l.angelegt = append(l.angelegt, setID)

	return 99, nil
}

func (l *lebenszyklusprotokoll) KarteKopieren(_ context.Context, karteID int64) (int64, error) {
	l.kopiert = append(l.kopiert, karteID)

	return 98, nil
}

func (l *lebenszyklusprotokoll) KarteLoeschen(_ context.Context, karteID int64) error {
	l.geloescht = append(l.geloescht, karteID)

	return nil
}

func (l *lebenszyklusprotokoll) KartenLoeschen(_ context.Context, karteIDs []int64) error {
	l.geloescht = append(l.geloescht, karteIDs...)

	return nil
}

func (l *lebenszyklusprotokoll) KartenTauschen(_ context.Context, eine, andere int64) error {
	l.getauscht = append(l.getauscht, [2]int64{eine, andere})

	return nil
}

// aufsicht baut eine Aufsicht über zwei Sets: Karte 1 und 2 in Set 10, Karte 3
// in Set 20.
func aufsicht(t *testing.T) (*service.Kartenaufsicht, *lebenszyklusprotokoll) {
	t.Helper()

	protokoll := &lebenszyklusprotokoll{}
	sicht := &kartensicht{karten: map[int64]int64{1: 10, 2: 10, 3: 20}}

	return service.NeueKartenaufsicht(sicht, protokoll), protokoll
}

// **Eine Karte aus einem anderen Set wird abgewiesen** — je Vorgang.
func TestDieAufsichtWeistEineFremdeKarteAb(t *testing.T) {
	faelle := []struct {
		name string
		ruf  func(a *service.Kartenaufsicht) error
	}{
		{"kopieren", func(a *service.Kartenaufsicht) error {
			_, fehler := a.KarteKopieren(t.Context(), 10, 3)

			return fehler
		}},
		{"löschen", func(a *service.Kartenaufsicht) error {
			return a.KarteLoeschen(t.Context(), 10, 3, false)
		}},
		{"mehrere löschen, eine ist fremd", func(a *service.Kartenaufsicht) error {
			return a.KartenLoeschen(t.Context(), 10, []int64{1, 3}, false)
		}},
		{"tauschen, die zweite ist fremd", func(a *service.Kartenaufsicht) error {
			return a.KartenTauschen(t.Context(), 10, 1, 3)
		}},
		{"tauschen, die erste ist fremd", func(a *service.Kartenaufsicht) error {
			return a.KartenTauschen(t.Context(), 10, 3, 1)
		}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			auge, protokoll := aufsicht(t)

			fehler := fall.ruf(auge)
			if !errors.Is(fehler, domain.ErrNotFound) {
				t.Fatalf("Fehler %v, erwartet war ErrNotFound", fehler)
			}

			// **Und nichts ist durchgekommen.** Ohne diese Zusicherung prüfte
			// der Test den Fehlertext und nicht die Wirkung — der Vorgang
			// könnte gelaufen sein und **danach** gemeldet haben (§10).
			if len(protokoll.kopiert)+len(protokoll.geloescht)+
				len(protokoll.getauscht) != 0 {
				t.Errorf("der Vorgang lief trotzdem: %+v", protokoll)
			}
		})
	}
}

// **Eine Karte aus dem genannten Set kommt durch.**
func TestDieAufsichtLaesstDieEigeneKarteDurch(t *testing.T) {
	auge, protokoll := aufsicht(t)

	if _, fehler := auge.KarteKopieren(t.Context(), 10, 1); fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	if fehler := auge.KarteLoeschen(t.Context(), 10, 2, false); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	if fehler := auge.KartenTauschen(t.Context(), 10, 1, 2); fehler != nil {
		t.Fatalf("tauschen: %v", fehler)
	}

	if _, fehler := auge.KarteAnlegen(t.Context(), 10); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if len(protokoll.kopiert) != 1 || len(protokoll.geloescht) != 1 ||
		len(protokoll.getauscht) != 1 || len(protokoll.angelegt) != 1 {
		t.Errorf("nicht jeder Vorgang kam an: %+v", protokoll)
	}
}

// **Eine Karte, die es gar nicht gibt, ist derselbe Befund.**
func TestDieAufsichtWeistEineUnbekannteKarteAb(t *testing.T) {
	auge, _ := aufsicht(t)

	if fehler := auge.KarteLoeschen(t.Context(), 10, 4711, false); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet war ErrNotFound", fehler)
	}
}

// feldprobe merkt sich, was auf dem Feldweg durchkam.
type feldprobe struct {
	gesetzt    []int64
	entfernt   []int64
	zurueck    []int64
	massen     [][]int64
	herkuenfte []string

	// feldherkuenfte hält, was eine KI-Übernahme mitschickte (A11, #1450),
	// erwartet, welcher Stand als Bedingung ankam (Befund 2).
	feldherkuenfte []domain.KIFeldherkunft
	erwartet       []string
}

func (f *feldprobe) FeldSetzenVonAussen(
	_ context.Context, karteID int64, _, _, herkunft string,
) (vorher string, hatte bool, fehler error) {
	f.gesetzt = append(f.gesetzt, karteID)
	f.herkuenfte = append(f.herkuenfte, herkunft)

	return "", false, nil
}

// FelderVonAussen zählt **einen** Vorgang je Zug — so, wie der Bestand ihn
// führt (#1026).
// FelderVonAussenMitHerkunft ist derselbe Weg samt Feldherkunft (A11, #1450).
func (f *feldprobe) FelderVonAussenMitHerkunft(
	ctx context.Context, karteID int64, werte []domain.Value,
	_ string, herkuenfte []domain.KIFeldherkunft,
	_ map[string][]domain.KIBildparameter, erwartet string,
) (vorher string, hatte bool, fehler error) {
	f.feldherkuenfte = append(f.feldherkuenfte, herkuenfte...)
	f.erwartet = append(f.erwartet, erwartet)

	return f.FelderVonAussen(ctx, karteID, werte, "ki")
}

func (f *feldprobe) FelderVonAussen(
	_ context.Context, karteID int64, _ []domain.Value, herkunft string,
) (vorher string, hatte bool, fehler error) {
	f.gesetzt = append(f.gesetzt, karteID)
	f.herkuenfte = append(f.herkuenfte, herkunft)

	return "", false, nil
}

func (f *feldprobe) FeldSetzenMitVorwert(
	_ context.Context, karteID int64, _, _ string,
) (vorher string, hatte bool, fehler error) {
	f.gesetzt = append(f.gesetzt, karteID)

	return "", false, nil
}

func (f *feldprobe) FeldEntfernen(_ context.Context, karteID int64, _ string) (
	vorher string, hatte bool, stelle *int64, fehler error,
) {
	f.entfernt = append(f.entfernt, karteID)

	return "", false, nil, nil
}

func (f *feldprobe) FeldSetzenAnStelle(
	_ context.Context, karteID int64, _, _ string, _ int64,
) (vorher string, hatte bool, fehler error) {
	f.zurueck = append(f.zurueck, karteID)

	return "", false, nil
}

func (f *feldprobe) FelderSetzen(_ context.Context, karteIDs []int64, _, _ string) (
	string, error,
) {
	f.massen = append(f.massen, karteIDs)

	return "vorgang7", nil
}

func (f *feldprobe) FelderSetzenMitErwartung(
	_ context.Context, erwartet []domain.Feldvorwert, _, _ string,
) (string, error) {
	karteIDs := make([]int64, 0, len(erwartet))
	for _, vorwert := range erwartet {
		karteIDs = append(karteIDs, vorwert.KarteID)
	}

	f.massen = append(f.massen, karteIDs)

	return "vorgang8", nil
}

// **Auch der Feldweg trägt die Setgrenze** (#808).
//
// Der Zweitblick über Phase N hat es angestrichen: [0041] sagte eine Regel für
// jede Kartenänderung zu und hielt sie für vier von sieben Wegen. „Späteres
// Werkstück" ist eine Umfangsverschiebung und keine fachliche Begründung.
func TestDieAufsichtHaeltAuchDenFeldweg(t *testing.T) {
	protokoll := &lebenszyklusprotokoll{}
	felder := &feldprobe{}
	sicht := &kartensicht{karten: map[int64]int64{1: 10, 2: 10, 3: 20}}
	auge := service.NeueKartenaufsicht(sicht, protokoll).MitFeldweg(felder, felder)

	t.Run("eine fremde Karte kommt nicht durch", func(t *testing.T) {
		_, _, fehler := auge.FeldSetzenMitVorwert(t.Context(), 10, 3, "name", "x")
		if !errors.Is(fehler, domain.ErrNotFound) {
			t.Fatalf("Fehler %v, erwartet war ErrNotFound", fehler)
		}

		if len(felder.gesetzt) != 0 {
			t.Errorf("geschrieben wurde trotzdem: %v", felder.gesetzt)
		}
	})

	t.Run("die eigene kommt durch", func(t *testing.T) {
		if _, _, fehler := auge.FeldSetzenMitVorwert(
			t.Context(), 10, 1, "name", "x"); fehler != nil {
			t.Fatalf("eigene Karte: %v", fehler)
		}
	})

	// **Auch die beiden Rücknahmewege** — eine Aufsicht, die nur den Hinweg
	// hält, lässt den Rückweg offen.
	t.Run("entfernen weist eine fremde Karte ab", func(t *testing.T) {
		_, _, _, fehler := auge.FeldEntfernen(t.Context(), 10, 3, "name")
		if !errors.Is(fehler, domain.ErrNotFound) {
			t.Fatalf("Fehler %v, erwartet war ErrNotFound", fehler)
		}

		if len(felder.entfernt) != 0 {
			t.Errorf("entfernt wurde trotzdem: %v", felder.entfernt)
		}
	})

	t.Run("zurücksetzen an die Stelle weist sie ebenso ab", func(t *testing.T) {
		_, _, fehler := auge.FeldSetzenAnStelle(t.Context(), 10, 3, "name", "x", 0)
		if !errors.Is(fehler, domain.ErrNotFound) {
			t.Fatalf("Fehler %v, erwartet war ErrNotFound", fehler)
		}

		if len(felder.zurueck) != 0 {
			t.Errorf("gesetzt wurde trotzdem: %v", felder.zurueck)
		}
	})

	t.Run("und lassen die eigene durch", func(t *testing.T) {
		if _, _, _, fehler := auge.FeldEntfernen(
			t.Context(), 10, 1, "name"); fehler != nil {
			t.Fatalf("entfernen: %v", fehler)
		}

		if _, _, fehler := auge.FeldSetzenAnStelle(
			t.Context(), 10, 1, "name", "x", 0); fehler != nil {
			t.Fatalf("zurücksetzen: %v", fehler)
		}
	})

	// **Und der Zug über mehrere Felder** (#1026). Er ist ein eigener Weg an
	// der Aufsicht vorbei, sobald ihn jemand vergisst — dieselbe Lücke, die
	// #808 für die vier Wege darüber gefunden hat.
	t.Run("mehrere Felder weisen eine fremde Karte ab", func(t *testing.T) {
		vorher := len(felder.gesetzt)

		_, _, fehler := auge.FelderVonAussen(t.Context(), 10, 3,
			[]domain.Value{{Field: "name", Text: "x"}}, "quelle")
		if !errors.Is(fehler, domain.ErrNotFound) {
			t.Fatalf("Fehler %v, erwartet war ErrNotFound", fehler)
		}

		if len(felder.gesetzt) != vorher {
			t.Errorf("geschrieben wurde trotzdem: %v", felder.gesetzt)
		}
	})

	t.Run("und lassen die eigene durch", func(t *testing.T) {
		if _, _, fehler := auge.FelderVonAussen(t.Context(), 10, 1,
			[]domain.Value{{Field: "name", Text: "x"}}, "quelle"); fehler != nil {
			t.Fatalf("mehrere Felder: %v", fehler)
		}
	})
}

// **Die Massenänderung weist ab, bevor sie die Hälfte geschrieben hat.**
//
// Die Reihenfolge ist der Punkt: `FelderSetzen` bekommt eine Liste, und eine
// fremde Kennung **darin** darf nicht dazu führen, dass die davorstehenden
// schon geschrieben sind. Dieselbe Überlegung steht in `app.FelderSetzen` für
// unlesbare Kennungen — sie gilt für fremde genauso.
func TestDieAufsichtPruefteJedeKarteDerMassenaenderungVorher(t *testing.T) {
	protokoll := &lebenszyklusprotokoll{}
	felder := &feldprobe{}
	sicht := &kartensicht{karten: map[int64]int64{1: 10, 2: 10, 3: 20}}
	auge := service.NeueKartenaufsicht(sicht, protokoll).MitFeldweg(felder, felder)

	_, fehler := auge.FelderSetzen(t.Context(), 10, []int64{1, 2, 3}, "name", "x")
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("Fehler %v, erwartet war ErrNotFound", fehler)
	}

	if len(felder.massen) != 0 {
		t.Errorf("geschrieben wurde trotzdem: %v", felder.massen)
	}

	if _, fehler := auge.FelderSetzen(t.Context(), 10, []int64{1, 2}, "name", "x"); fehler != nil {
		t.Fatalf("die eigenen Karten: %v", fehler)
	}

	if len(felder.massen) != 1 {
		t.Errorf("%d Massenläufe, erwartet war einer", len(felder.massen))
	}
}

// **Ein Zug über mehrere Karten prüft jede, bevor die erste fällt** — Befund 4
// des Zweitblicks über Phase 3.7.
//
// # Warum die Reihenfolge zählt
//
// Der Speicher hält „ganz oder gar nicht" in einer Transaktion. Prüfte die
// Aufsicht **während** des Löschens, stünde die Zusage auf dem Kopf: eine
// fremde Kennung mitten in der Liste ließe die Karten davor schon gelöscht
// zurück. Der Lauf zeigt es an der Wirkung — der Zyklus wird gar nicht erst
// gerufen.
func TestEinZugPrueftJedeKarteVorDerErsten(t *testing.T) {
	sicht, protokoll := aufsicht(t)

	if err := sicht.KartenLoeschen(t.Context(), 10, []int64{1, 3}, false); err == nil {
		t.Fatal("kein Fehler, erwartet einen")
	}

	if len(protokoll.geloescht) != 0 {
		t.Fatalf("gelöscht wurden %v, erwartet nichts", protokoll.geloescht)
	}
}

// **Und wenn alle dazugehören, geht der ganze Zug hinaus.**
func TestEinZugReichtAlleKennungenWeiter(t *testing.T) {
	sicht, protokoll := aufsicht(t)

	if err := sicht.KartenLoeschen(t.Context(), 10, []int64{1, 2}, false); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	if len(protokoll.geloescht) != 2 {
		t.Fatalf("gelöscht wurden %v, erwartet beide", protokoll.geloescht)
	}
}
