package service_test

// **Der Gast zeichnet das Bild seiner eigenen Karte** — P9b
// (#2377).
// Fragt der Wirt, zeichnet der Gast nur, was in seiner letzten Sicht steht
// und in **seinen** Bestand zeigt — eine Zeile, die die Sicht nicht nennt,
// gibt es für ihn nicht.

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einMalenderGast sitzt an einem Tisch, dessen Sicht drei Objekte nennt.
func einMalenderGast(t *testing.T, maler service.Kartenbildmaler) (*service.Partiegast, *nahtdoppel) {
	t.Helper()

	gast, naht := einGast(&nehmerdoppelGast{})
	if maler != nil {
		gast.MitMaler(maler)
	}

	bild := domain.Tischbild{Sicht: domain.Partiesicht{PartieID: "p1", Fuer: "bo", Objekte: []domain.Sichtobjekt{
		{Kennung: "eigen", KarteID: 7, Bild: "b-eigen"},
		{Kennung: "geheim", Geheim: true},
		{Kennung: "fremd", Bild: "b-fremd"},
	}}, Erlaubt: []string{}}

	if fehler := gast.Angekommen(t.Context(), eineWirtsnachricht(t, domain.ArtKontrolle,
		mussVerpacken(t, domain.KontrolleTischbild, bild))); fehler != nil {
		t.Fatalf("Tischbild: %v", fehler)
	}

	return gast, naht
}

// gefragtWird lässt den Wirt nach einem Objekt fragen und liest die Antwort.
func gefragtWird(t *testing.T, gast *service.Partiegast, naht *nahtdoppel, objekt string) domain.Bildantwort {
	t.Helper()

	if fehler := gast.Angekommen(t.Context(), eineWirtsnachricht(t, domain.ArtKontrolle,
		mussVerpacken(t, domain.KontrolleBildwunsch, domain.Bildwunsch{Objekt: objekt, Seite: "vorne", Guete: "tisch"}))); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	art, inhalt, fehler := domain.LiesKontrolle(naht.letzte(t).Inhalt, "gast", domain.KontrolleBild)
	if fehler != nil || art != domain.KontrolleBild {
		t.Fatalf("keine Bildantwort: %v", fehler)
	}

	var antwort domain.Bildantwort
	if fehler := json.Unmarshal(inhalt, &antwort); fehler != nil {
		t.Fatalf("die Antwort ist kein Bild: %v", fehler)
	}

	return antwort
}

func TestDerGastZeichnetSeineEigeneKarte(t *testing.T) {
	t.Parallel()

	maler := &malerdoppel{}
	gast, naht := einMalenderGast(t, maler)

	antwort := gefragtWird(t, gast, naht, "eigen")

	if len(antwort.Png) == 0 || antwort.Breite != 300 || len(maler.gemalt) != 1 || maler.gemalt[0] != 7 {
		t.Errorf("die Antwort: %d Bytes, gezeichnet %v", len(antwort.Png), maler.gemalt)
	}
}

func TestDerGastZeichnetNurWasSeinIst(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		objekt string
		maler  *malerdoppel
	}{
		"verdeckt":                {"geheim", &malerdoppel{}},
		"aus einem fremden Deck":  {"fremd", &malerdoppel{}},
		"gibt es nicht":           {"weg", &malerdoppel{}},
		"ohne Maler":              {"eigen", nil},
		"der Maler scheitert":     {"eigen", &malerdoppel{klage: errors.New("kein Stil")}},
		"zu groß für die Leitung": {"eigen", &malerdoppel{gross: true}},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var maler service.Kartenbildmaler
			if fall.maler != nil {
				maler = fall.maler
			}

			gast, naht := einMalenderGast(t, maler)

			if antwort := gefragtWird(t, gast, naht, fall.objekt); antwort.Grund == "" || len(antwort.Png) != 0 {
				t.Errorf("die Antwort: Grund %q, %d Bytes", antwort.Grund, len(antwort.Png))
			}
		})
	}
}

func TestEinUnlesbarerWunschDesWirtsWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	gast, _ := einMalenderGast(t, &malerdoppel{})

	for _, nutzlast := range []any{map[string]int{"objekt": 5}, domain.Bildwunsch{}} {
		if fehler := gast.Angekommen(context.Background(), eineWirtsnachricht(t, domain.ArtKontrolle,
			mussVerpacken(t, domain.KontrolleBildwunsch, nutzlast))); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%v: es kam %v", nutzlast, fehler)
		}
	}
}
