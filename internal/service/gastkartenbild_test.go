package service_test

// **Der Gast fragt nach einem Kartenbild** — P9
// (#2377),
// Vermerk 0198. Er schickt einen Bildwunsch und wartet auf **die** Antwort,
// die zu ihm passt — nicht auf irgendeine. Antwortet der Wirt nicht in der
// Frist, sagt er es; eine Antwort, die danach kommt, reißt nichts ab.

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

var einBildwunsch = domain.Bildwunsch{Objekt: "o1", Seite: "vorne", Guete: "tisch"}

// eineBildantwort verpackt die Antwort des Wirts.
func eineBildantwort(t *testing.T, antwort domain.Bildantwort) domain.Partienachricht {
	t.Helper()

	inhalt, fehler := domain.VerpackeKontrolle(domain.KontrolleBild, antwort)
	if fehler != nil {
		t.Fatalf("verpacken: %v", fehler)
	}

	return domain.Partienachricht{
		Art: domain.ArtKontrolle, Partie: "p1", Epoche: 1, Folge: 1,
		Von: "wirt", Inhalt: inhalt,
	}
}

// warteAufGesendetes wartet, bis die Naht `n` Nachrichten hat.
func warteAufGesendetes(t *testing.T, naht *nahtdoppel, n int) {
	t.Helper()

	frist := time.Now().Add(5 * time.Second)
	for time.Now().Before(frist) {
		naht.schloss.Lock()
		da := len(naht.gesendet)
		naht.schloss.Unlock()

		if da >= n {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatalf("es ging nichts hinaus")
}

func TestDerGastBekommtDasErfragteBild(t *testing.T) {
	t.Parallel()

	gast, naht := einGast(&nehmerdoppelGast{})

	ergebnis := make(chan domain.Bildantwort, 1)
	go func() {
		antwort, fehler := gast.Bild(context.Background(), einBildwunsch)
		if fehler != nil {
			t.Errorf("Bild: %v", fehler)
		}

		ergebnis <- antwort
	}()

	warteAufGesendetes(t, naht, 1)

	art, _, fehler := domain.LiesKontrolle(naht.letzte(t).Inhalt, "gast", domain.KontrolleBildwunsch)
	if fehler != nil || art != domain.KontrolleBildwunsch {
		t.Fatalf("hinaus ging kein Bildwunsch: %v", fehler)
	}

	// **Eine Antwort auf einen anderen Wunsch zählt nicht.**
	anders := einBildwunsch
	anders.Objekt = "o2"
	if fehler := gast.Angekommen(context.Background(), eineBildantwort(t,
		domain.Bildantwort{Wunsch: anders, Grund: "anderes"})); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	if fehler := gast.Angekommen(context.Background(), eineBildantwort(t,
		domain.Bildantwort{Wunsch: einBildwunsch, Png: []byte{1}, Breite: 3, Hoehe: 4})); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	select {
	case antwort := <-ergebnis:
		if len(antwort.Png) != 1 || antwort.Breite != 3 {
			t.Errorf("die Antwort: %+v", antwort)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("der Gast wartet noch")
	}
}

// **Zwei, die dasselbe fragen, bekommen beide die Antwort.**
func TestZweiGleicheWuenscheBekommenBeideDieAntwort(t *testing.T) {
	t.Parallel()

	gast, naht := einGast(&nehmerdoppelGast{})

	// **Mit Frist**: bekäme einer nichts, hinge der Lauf sonst bis zum
	// Ende der ganzen Stufe.
	ctx, fertig := context.WithTimeout(context.Background(), 5*time.Second)
	defer fertig()

	var gruppe sync.WaitGroup
	for range 2 {
		gruppe.Go(func() {
			if antwort, fehler := gast.Bild(ctx, einBildwunsch); fehler != nil || antwort.Grund != "verdeckt" {
				t.Errorf("Bild: %+v, %v", antwort, fehler)
			}
		})
	}

	warteAufGesendetes(t, naht, 2)

	if fehler := gast.Angekommen(context.Background(), eineBildantwort(t,
		domain.Bildantwort{Wunsch: einBildwunsch, Grund: "verdeckt"})); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	gruppe.Wait()
}

// **Ohne Antwort in der Frist sagt der Gast es** — und eine Antwort, die
// danach kommt, ist kein Fehler der Leitung.
func TestOhneAntwortGibtDerGastNachDerFristAuf(t *testing.T) {
	t.Parallel()

	gast, _ := einGast(&nehmerdoppelGast{})

	ctx, fertig := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer fertig()

	if _, fehler := gast.Bild(ctx, einBildwunsch); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}

	if fehler := gast.Angekommen(context.Background(), eineBildantwort(t,
		domain.Bildantwort{Wunsch: einBildwunsch, Grund: "zu spät"})); fehler != nil {
		t.Errorf("die späte Antwort riss ab: %v", fehler)
	}
}

func TestEinUngueltigerWunschGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	gast, naht := einGast(&nehmerdoppelGast{})

	if _, fehler := gast.Bild(context.Background(), domain.Bildwunsch{}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}

	if len(naht.gesendet) != 0 {
		t.Error("der Wunsch ging trotzdem hinaus")
	}
}

func TestEineFalscheBildantwortWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	gast, _ := einGast(&nehmerdoppelGast{})

	if fehler := gast.Angekommen(context.Background(), eineBildantwort(t,
		domain.Bildantwort{Wunsch: einBildwunsch})); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Klagt die Leitung, wartet niemand.**
func TestEineKlagendeLeitungWartetNicht(t *testing.T) {
	t.Parallel()

	klage := errors.New("abgerissen")
	naht := &nahtdoppel{klage: klage}
	gast := service.NeuerPartiegast("p1",
		service.NeuePartiesitzung("p1", "gast", naht, service.NeueReihenfolge(1)),
		&nehmerdoppelGast{})

	if _, fehler := gast.Bild(context.Background(), einBildwunsch); !errors.Is(fehler, klage) {
		t.Fatalf("es kam %v", fehler)
	}
}

// leerendeNaht zählt, wie oft hinausgegeben wurde.
type leerendeNaht struct {
	nahtdoppel
	geleert int
}

func (n *leerendeNaht) Leere(context.Context) error {
	n.schloss.Lock()
	defer n.schloss.Unlock()

	n.geleert++

	return nil
}

// **Der Wunsch geht sofort hinaus** — die Sitzung legt nur in die Schlange,
// und ein Wunsch, der dort wartet, während sein Absender auf die Antwort
// wartet, käme nie an.
func TestDerBildwunschGehtSofortHinaus(t *testing.T) {
	t.Parallel()

	naht := &leerendeNaht{}
	gast := service.NeuerPartiegast("p1",
		service.NeuePartiesitzung("p1", "gast", naht, service.NeueReihenfolge(1)),
		&nehmerdoppelGast{})

	ctx, fertig := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer fertig()

	_, _ = gast.Bild(ctx, einBildwunsch) //nolint:errcheck // hier zählt nur das Hinausgeben

	naht.schloss.Lock()
	defer naht.schloss.Unlock()

	if naht.geleert != 1 {
		t.Errorf("hinausgegeben %d-mal", naht.geleert)
	}
}

// **Wer schickt und wer empfängt, sind zwei Gäste derselben Sitzung** — der
// eine je Ruf des Fensters, der andere in der Empfangsschleife. Die Post
// teilen sie.
func TestZweiGaesteTeilenIhreBildpost(t *testing.T) {
	t.Parallel()

	post := &service.Bildpost{}
	schickt, naht := einGast(&nehmerdoppelGast{})
	empfaengt, _ := einGast(&nehmerdoppelGast{})
	schickt = schickt.MitBildpost(post)
	empfaengt = empfaengt.MitBildpost(post)

	ctx, fertig := context.WithTimeout(context.Background(), 5*time.Second)
	defer fertig()

	ergebnis := make(chan error, 1)
	go func() {
		_, fehler := schickt.Bild(ctx, einBildwunsch)
		ergebnis <- fehler
	}()

	warteAufGesendetes(t, naht, 1)

	if fehler := empfaengt.Angekommen(context.Background(), eineBildantwort(t,
		domain.Bildantwort{Wunsch: einBildwunsch, Grund: "verdeckt"})); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	if fehler := <-ergebnis; fehler != nil {
		t.Errorf("der schickende Gast bekam nichts: %v", fehler)
	}
}

// **Gibt einer von zweien auf, bekommt der andere die Antwort trotzdem** —
// das Austragen nimmt nur den einen.
func TestWerAufgibtNimmtDemAnderenNichtsWeg(t *testing.T) {
	t.Parallel()

	gast, naht := einGast(&nehmerdoppelGast{})

	ctx, fertig := context.WithTimeout(context.Background(), 5*time.Second)
	defer fertig()

	bleibt := make(chan domain.Bildantwort, 1)
	go func() {
		antwort, _ := gast.Bild(ctx, einBildwunsch) //nolint:errcheck // geprüft wird die Antwort
		bleibt <- antwort
	}()

	warteAufGesendetes(t, naht, 1)

	kurz, aus := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer aus()

	if _, fehler := gast.Bild(kurz, einBildwunsch); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der Ungeduldige bekam %v", fehler)
	}

	if fehler := gast.Angekommen(context.Background(), eineBildantwort(t,
		domain.Bildantwort{Wunsch: einBildwunsch, Grund: "verdeckt"})); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	if antwort := <-bleibt; antwort.Grund != "verdeckt" {
		t.Errorf("der Gebliebene bekam %+v", antwort)
	}
}

// **Eine Bildantwort, die keine ist, wird abgewiesen** — nicht zugestellt.
func TestEineUnlesbareBildantwortWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	gast, _ := einGast(&nehmerdoppelGast{})

	inhalt, fehler := domain.VerpackeKontrolle(domain.KontrolleBild, map[string]int{"png": 5})
	if fehler != nil {
		t.Fatalf("verpacken: %v", fehler)
	}

	if fehler := gast.Angekommen(context.Background(), domain.Partienachricht{
		Art: domain.ArtKontrolle, Partie: "p1", Epoche: 1, Folge: 1, Von: "wirt", Inhalt: inhalt,
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}
