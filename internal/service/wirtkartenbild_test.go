package service_test

// **Der Wirt beantwortet einen Bildwunsch** — P9
// (#2377),
// Vermerk 0198. Ein Bild gibt es nur für ein Objekt, das der Gast **sehen**
// darf und dessen Deck aus **diesem** Bestand stammt; sonst einen Grund. Die
// Zeile, mit der gezeichnet wird, verlässt den Wirt nie.

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bildnehmer gibt je Bestand eine andere Sicht — am Wirt mit Zeilen, für den
// Gast ohne.
//
// `klagtBei` lässt den n-ten Ruf scheitern — der erste rechnet für den Gast,
// der zweite für den Wirt.
type bildnehmer struct {
	nehmerdoppel
	objekte  []domain.Sichtobjekt
	klagtBei int
	gerufen  int

	// bestaende nennt je Objekt den Gast, in dessen Bestand seine Zeile
	// liegt (Zweitblick über Phase 7, Befund 6) — für dessen Sicht trägt es
	// eine, für jede andere fremde keine.
	bestaende map[string]string
}

func (n *bildnehmer) Partie(
	ctx context.Context, kennung, fuer string,
) (domain.Partiesicht, error) {
	n.gerufen++
	if n.gerufen == n.klagtBei {
		return domain.Partiesicht{}, errors.New("die Partie ließ sich nicht lesen")
	}

	gesehen := make([]domain.Sichtobjekt, 0, len(n.objekte))

	for _, objekt := range n.objekte {
		if bestand := service.BestandssichtAus(ctx); bestand != "" {
			objekt.KarteID = 0
			if n.bestaende[objekt.Kennung] == bestand {
				objekt.KarteID = 77
			}
		}

		gesehen = append(gesehen, objekt)
	}

	return domain.Partiesicht{PartieID: kennung, Fuer: fuer, Objekte: gesehen}, nil
}

// malerdoppel zeichnet und merkt sich, welche Zeile es zeichnen sollte.
type malerdoppel struct {
	gemalt []int64
	klage  error
	gross  bool
}

func (m *malerdoppel) Kartenbild(
	_ context.Context, karte int64, _, _ string,
) (service.Kartenbild, error) {
	m.gemalt = append(m.gemalt, karte)

	if m.klage != nil {
		return service.Kartenbild{}, m.klage
	}

	if m.gross {
		return service.Kartenbild{Png: make([]byte, 2<<20), Breite: 1, Hoehe: 1}, nil
	}

	return service.Kartenbild{Png: []byte{0x89, 'P', 'N', 'G'}, Breite: 300, Hoehe: 420}, nil
}

func einBildwirt(t *testing.T, maler service.Kartenbildmaler) (*service.Partiewirt, *nahtdoppel) {
	t.Helper()

	return einBildwirtMit(t, maler, 0)
}

func einBildwirtMit(t *testing.T, maler service.Kartenbildmaler, klagtBei int) (*service.Partiewirt, *nahtdoppel) {
	t.Helper()

	nehmer := &bildnehmer{objekte: []domain.Sichtobjekt{
		{Kennung: "offen", KarteID: 42, Bild: "b-offen"},
		{Kennung: "geheim", Geheim: true, KarteID: 99},
		{Kennung: "fremd", Bild: "b-fremd"},
	}, klagtBei: klagtBei}
	wirt, naht := einWirt(nehmer)
	wirt = wirt.MitGegenueber("peer-gast")

	if maler != nil {
		wirt = wirt.MitMaler(maler)
	}

	return wirt, naht
}

func erfrage(t *testing.T, wirt *service.Partiewirt, naht *nahtdoppel, objekt string) domain.Bildantwort {
	t.Helper()

	nachricht := eineKontrollnachricht(t, domain.KontrolleBildwunsch,
		domain.Bildwunsch{Objekt: objekt, Seite: "vorne", Guete: "tisch"})
	if fehler := wirt.Angekommen(context.Background(), nachricht); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	return dieBildantwort(t, naht)
}

// dieBildantwort liest, was der Wirt zuletzt als Bild hinausgab.
func dieBildantwort(t *testing.T, naht *nahtdoppel) domain.Bildantwort {
	t.Helper()

	art, inhalt, fehler := domain.LiesKontrolle(naht.letzte(t).Inhalt, "wirt", domain.KontrolleBild)
	if fehler != nil || art != domain.KontrolleBild {
		t.Fatalf("keine Bildantwort: %v", fehler)
	}

	var antwort domain.Bildantwort
	if fehler := json.Unmarshal(inhalt, &antwort); fehler != nil {
		t.Fatalf("die Antwort ist kein Bild: %v", fehler)
	}

	if fehler := antwort.Pruefe(); fehler != nil {
		t.Fatalf("die Antwort hält ihre Gestalt nicht: %v", fehler)
	}

	return antwort
}

func TestDerWirtZeichnetEinSichtbaresObjektMitSeinerZeile(t *testing.T) {
	t.Parallel()

	maler := &malerdoppel{}
	wirt, naht := einBildwirt(t, maler)

	antwort := erfrage(t, wirt, naht, "offen")

	if len(antwort.Png) == 0 || antwort.Breite != 300 || antwort.Wunsch.Objekt != "offen" {
		t.Errorf("die Antwort: %+v", antwort)
	}

	if len(maler.gemalt) != 1 || maler.gemalt[0] != 42 {
		t.Errorf("gezeichnet wurde %v", maler.gemalt)
	}
}

func TestDerWirtGibtNurMitGrundKeinBild(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		objekt string
		maler  *malerdoppel
		mit    bool
	}{
		"geheim für den Gast":      {"geheim", &malerdoppel{}, true},
		"nicht aus diesem Bestand": {"fremd", &malerdoppel{}, true},
		"gibt es nicht":            {"weg", &malerdoppel{}, true},
		"ohne Maler":               {"offen", nil, false},
		"der Maler scheitert":      {"offen", &malerdoppel{klage: errors.New("kein Stil")}, true},
		"zu groß für die Leitung":  {"offen", &malerdoppel{gross: true}, true},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var maler service.Kartenbildmaler
			if fall.mit {
				maler = fall.maler
			}

			wirt, naht := einBildwirt(t, maler)
			antwort := erfrage(t, wirt, naht, fall.objekt)

			if antwort.Grund == "" || len(antwort.Png) != 0 {
				t.Errorf("die Antwort: Grund %q, %d Bytes", antwort.Grund, len(antwort.Png))
			}

			if fall.mit && fall.objekt != "offen" && len(fall.maler.gemalt) != 0 {
				t.Errorf("gezeichnet wurde trotzdem %v", fall.maler.gemalt)
			}
		})
	}
}

// **Ein Bild an einen Wirt, der nie fragt, ist eine Anmaßung** — fragen kann
// er erst mit einer Post (P9b, `partierundebild_test.go`).
func TestDerWirtWeistEinUngefragtesBildAb(t *testing.T) {
	t.Parallel()

	wirt, _ := einBildwirt(t, &malerdoppel{})

	nachricht := eineKontrollnachricht(t, domain.KontrolleBild, domain.Bildantwort{
		Wunsch: domain.Bildwunsch{Objekt: "o", Seite: "vorne", Guete: "tisch"},
		Grund:  "x",
	})
	if fehler := wirt.Angekommen(context.Background(), nachricht); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Ist der Tisch nicht zu lesen, gibt es einen Grund statt eines Bildes** —
// für den Gast gerechnet wie für den Wirt.
func TestEinUnlesbarerTischGibtKeinBild(t *testing.T) {
	t.Parallel()

	for name, klagtBei := range map[string]int{"für den Gast": 1, "für den Wirt": 2} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			maler := &malerdoppel{}
			wirt, naht := einBildwirtMit(t, maler, klagtBei)
			antwort := erfrage(t, wirt, naht, "offen")

			if antwort.Grund != "der Tisch ist gerade nicht zu lesen" || len(maler.gemalt) != 0 {
				t.Errorf("die Antwort: %+v, gezeichnet %v", antwort, maler.gemalt)
			}
		})
	}
}

// **Ein Bildwunsch, der keiner ist, bekommt keine Antwort**, sondern einen
// Fehler — und es geht nichts hinaus.
func TestEinUnlesbarerBildwunschWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]any{
		"falsche Gestalt": map[string]int{"objekt": 5},
		"ohne Objekt":     domain.Bildwunsch{Seite: "vorne", Guete: "tisch"},
	}

	for name, nutzlast := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wirt, naht := einBildwirt(t, &malerdoppel{})

			fehler := wirt.Angekommen(context.Background(),
				eineKontrollnachricht(t, domain.KontrolleBildwunsch, nutzlast))
			if !errors.Is(fehler, domain.ErrValidation) || len(naht.gesendet) != 0 {
				t.Errorf("es kam %v, hinaus gingen %d", fehler, len(naht.gesendet))
			}
		})
	}
}

// **Die Bildkennung gehört zum Objekt** — Nachlauf zum Zweitblick über
// Phase 7, Rest R2: ein Wunsch nennt die Kennung, unter der der Gast das Bild
// ablegen wird, und der Wirt zeichnet nur, wenn sie die des Objekts ist.
// Sonst läge das Bild eines Objekts unter der Kennung eines anderen.
func TestEineFremdeBildkennungBekommtKeinBild(t *testing.T) {
	t.Parallel()

	maler := &malerdoppel{}
	wirt, naht := einBildwirt(t, maler)

	if fehler := wirt.Angekommen(context.Background(), eineKontrollnachricht(t, domain.KontrolleBildwunsch,
		domain.Bildwunsch{Objekt: "offen", Bild: "b-fremd", Seite: "vorne", Guete: "tisch"})); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	if antwort := dieBildantwort(t, naht); antwort.Grund == "" || len(maler.gemalt) != 0 {
		t.Errorf("die Antwort: %+v, gemalt %v", antwort, maler.gemalt)
	}
}
