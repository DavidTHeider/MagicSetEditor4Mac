package service_test

// **Der Wirt fragt einen Gast nach dem Bild seiner Karte** — P9b
// (#2377).
// Bringt ein Gast sein eigenes Deck mit, zeigen dessen Kennungen in seinen
// Bestand; zeichnen kann es nur er. Die Frage geht über seinen Platz am Tisch
// hinaus, die Antwort kommt über den Wirt dieses Sitzes zurück — beide
// teilen eine Post.

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

var bosBildwunsch = domain.Bildwunsch{Objekt: "bo/k9/0", Seite: "vorne", Guete: "tisch"}

// eineGastantwort ist die Bildantwort, wie sie vom Gast kommt.
func eineGastantwort(t *testing.T, antwort domain.Bildantwort) domain.Partienachricht {
	t.Helper()

	nachricht := eineKontrollnachricht(t, domain.KontrolleBild, antwort)
	nachricht.Von = "bo"

	return nachricht
}

// warteAufBildwunsch wartet, bis über den Platz ein Bildwunsch hinausging.
func warteAufBildwunsch(t *testing.T, doppel *platzdoppel) domain.Bildwunsch {
	t.Helper()

	frist := time.Now().Add(5 * time.Second)
	for time.Now().Before(frist) {
		for _, nachricht := range doppel.alle() {
			art, inhalt, fehler := domain.LiesKontrolle(nachricht.Inhalt, "wirt", domain.KontrolleBildwunsch)
			if fehler != nil || art != domain.KontrolleBildwunsch {
				continue
			}

			var wunsch domain.Bildwunsch
			if json.Unmarshal(inhalt, &wunsch) == nil {
				return wunsch
			}
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatal("es ging kein Bildwunsch hinaus")

	return domain.Bildwunsch{}
}

func TestDerWirtFragtDenGastNachSeinemBild(t *testing.T) {
	t.Parallel()

	post := &service.Bildpost{}
	runde := &service.Partierunde{}
	platz, doppel := einPlatz("bo")
	runde.Setze(platz.MitBildpost(post))

	wirt, _ := einWirt(&nehmerdoppel{})
	wirt = wirt.MitBildpost(platz.Bildpost())

	ctx, fertig := context.WithTimeout(context.Background(), 5*time.Second)
	defer fertig()

	ergebnis := make(chan domain.Bildantwort, 1)
	go func() {
		antwort, fehler := runde.Bild(ctx, "bo", bosBildwunsch)
		if fehler != nil {
			t.Errorf("Bild: %v", fehler)
		}

		ergebnis <- antwort
	}()

	if wunsch := warteAufBildwunsch(t, doppel); wunsch != bosBildwunsch {
		t.Fatalf("hinaus ging %+v", wunsch)
	}

	if fehler := wirt.Angekommen(context.Background(), eineGastantwort(t,
		domain.Bildantwort{Wunsch: bosBildwunsch, Png: []byte{1}, Breite: 3, Hoehe: 4})); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	if antwort := <-ergebnis; len(antwort.Png) != 1 || antwort.Hoehe != 4 {
		t.Errorf("die Antwort: %+v", antwort)
	}
}

func TestOhneGastAmSitzFragtNiemand(t *testing.T) {
	t.Parallel()

	ohnePost, _ := einPlatz("cora")

	runde := &service.Partierunde{}
	runde.Setze(ohnePost)

	for _, sitz := range []string{"bo", "cora"} {
		if _, fehler := runde.Bild(context.Background(), sitz, bosBildwunsch); !errors.Is(fehler, domain.ErrConflict) {
			t.Errorf("%s: es kam %v", sitz, fehler)
		}
	}
}

func TestEineKlagendeLeitungFragtNicht(t *testing.T) {
	t.Parallel()

	klage := errors.New("abgerissen")
	platz, doppel := einPlatz("bo")
	doppel.klage = klage

	runde := &service.Partierunde{}
	runde.Setze(platz.MitBildpost(&service.Bildpost{}))

	if _, fehler := runde.Bild(context.Background(), "bo", bosBildwunsch); !errors.Is(fehler, klage) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Eine späte oder fremde Antwort reißt nichts ab** — aber eine falsche
// Gestalt wird abgewiesen.
func TestDerWirtNimmtBildantwortenNurInGestalt(t *testing.T) {
	t.Parallel()

	wirt, _ := einWirt(&nehmerdoppel{})
	wirt = wirt.MitBildpost(&service.Bildpost{})

	if fehler := wirt.Angekommen(context.Background(), eineGastantwort(t,
		domain.Bildantwort{Wunsch: bosBildwunsch, Grund: "zu spät"})); fehler != nil {
		t.Errorf("die späte Antwort riss ab: %v", fehler)
	}

	if fehler := wirt.Angekommen(context.Background(), eineGastantwort(t,
		domain.Bildantwort{Wunsch: bosBildwunsch})); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("eine Antwort ohne Bild und Grund: %v", fehler)
	}
}

// **Der Wirt fragt nach dem Objekt, nicht nach dem Sitz** — und nur nach
// einem, das er selbst offen sieht.
func TestDerWirtFragtDenEignerEinesOffenenObjekts(t *testing.T) {
	t.Parallel()

	leser := &bildnehmer{objekte: []domain.Sichtobjekt{
		{Kennung: "bo/k9/0", Bild: "b1", Eigner: "bo"},
		{Kennung: "verdeckt", Geheim: true, Eigner: "bo"},
		{Kennung: "geliehen", Bild: "b2", Eigner: "bo"},
		{Kennung: "cys", Bild: "b3", Eigner: "cy"},
	}, bestaende: map[string]string{"bo/k9/0": "peer-bo"}}

	post := &service.Bildpost{}
	runde := &service.Partierunde{}
	platz, doppel := einPlatz("bo")
	runde.Setze(platz.MitBildpost(post).MitPeer("peer-bo"))
	runde.Kann("bo", []string{domain.FaehigkeitBild})

	ctx, fertig := context.WithTimeout(context.Background(), 5*time.Second)
	defer fertig()

	fertigGefragt := make(chan error, 1)
	go func() {
		_, fehler := runde.BildVomEigner(ctx, leser, "p1", "wirt", bosBildwunsch)
		fertigGefragt <- fehler
	}()

	wunsch := warteAufBildwunsch(t, doppel)

	bowirt, _ := einWirt(&nehmerdoppel{})
	if fehler := bowirt.MitBildpost(post).Angekommen(context.Background(),
		eineGastantwort(t, domain.Bildantwort{Wunsch: wunsch, Grund: "verdeckt"})); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	if fehler := <-fertigGefragt; fehler != nil {
		t.Errorf("BildVomEigner: %v", fehler)
	}

	for _, objekt := range []string{"verdeckt", "weg"} {
		anders := bosBildwunsch
		anders.Objekt = objekt

		if _, fehler := runde.BildVomEigner(ctx, leser, "p1", "wirt", anders); !errors.Is(fehler, domain.ErrConflict) {
			t.Errorf("%s: es kam %v", objekt, fehler)
		}
	}

	// **Ein Deck aus einem dritten Bestand** (Befund 6): ein Grund, und Bo
	// wird nicht gefragt.
	vorher := len(doppel.alle())
	geliehen := bosBildwunsch
	geliehen.Objekt = "geliehen"

	if antwort, fehler := runde.BildVomEigner(ctx, leser, "p1", "wirt", geliehen); fehler != nil || antwort.Grund == "" {
		t.Errorf("geliehen: %+v, %v", antwort, fehler)
	}

	// **An einem Sitz ohne Gast** ebenso.
	leer := bosBildwunsch
	leer.Objekt = "cys"

	if antwort, fehler := runde.BildVomEigner(ctx, leser, "p1", "wirt", leer); fehler != nil || antwort.Grund == "" {
		t.Errorf("ohne Gast: %+v, %v", antwort, fehler)
	}

	if len(doppel.alle()) != vorher {
		t.Error("Bo wurde nach einer fremden Leihe gefragt")
	}

	// **Eine Bildkennung, die nicht die des Objekts ist** (R2), bekommt
	// keines — sonst läge das Bild unter der Kennung eines anderen.
	falsch := bosBildwunsch
	falsch.Bild = "b2"
	gefragt := len(doppel.alle())

	if _, fehler := runde.BildVomEigner(ctx, leser, "p1", "wirt", falsch); !errors.Is(fehler, domain.ErrConflict) || len(doppel.alle()) != gefragt {
		t.Errorf("fremde Bildkennung: es kam %v, gefragt %d", fehler, len(doppel.alle())-gefragt)
	}

	// **Lässt sich die Sicht des Gastes nicht lesen, wird er nicht gefragt.**
	zweiteKlagt := &bildnehmer{objekte: leser.objekte, bestaende: leser.bestaende, klagtBei: 2}
	if antwort, fehler := runde.BildVomEigner(ctx, zweiteKlagt, "p1", "wirt", bosBildwunsch); fehler != nil || antwort.Grund == "" {
		t.Errorf("unlesbare Gastsicht: %+v, %v", antwort, fehler)
	}

	// **Und eine Sicht, die sich nicht lesen lässt, kommt als Fehler zurück.**
	klagend := &bildnehmer{klagtBei: 1}
	if _, fehler := runde.BildVomEigner(ctx, klagend, "p1", "wirt", bosBildwunsch); fehler == nil {
		t.Error("eine unlesbare Sicht ging durch")
	}
}
