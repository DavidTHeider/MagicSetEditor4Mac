package service_test

// **Der Wirt reicht ein Bild weiter** — P9b
// (#2377).
// Fragt ein Gast nach einer Karte aus dem Deck eines **anderen** Gastes,
// zeichnet sie keiner von beiden: der Wirt fragt den, dem das Deck gehört,
// und gibt seine Antwort weiter. **Nie zurück an den Fragenden selbst** — der
// wartete auf sich.

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einWeiterreichenderWirt hat einen zweiten Gast, Cora, am Tisch; ihr gehört
// das Objekt `coras`.
func einWeiterreichenderWirt(t *testing.T) (*service.Partiewirt, *nahtdoppel, *platzdoppel, *service.Bildpost, *service.Partierunde) {
	t.Helper()

	nehmer := &bildnehmer{objekte: []domain.Sichtobjekt{
		{Kennung: "coras", Bild: "b-coras", Eigner: "cora"},
		{Kennung: "eigenes", Bild: "b-eigenes", Eigner: "gast"},
		// **Geliehen aus einem dritten Bestand** — an Coras Sitz, aber
		// ihre Zeile liegt bei niemandem am Tisch.
		{Kennung: "geliehen", Bild: "b-geliehen", Eigner: "cora"},
	}, bestaende: map[string]string{"coras": "peer-cora"}}

	post := &service.Bildpost{}
	runde := &service.Partierunde{}
	platz, doppel := einPlatz("cora")
	runde.Setze(platz.MitBildpost(post).MitPeer("peer-cora"))
	runde.Kann("cora", []string{domain.FaehigkeitBild})

	wirt, naht := einWirt(nehmer)

	return wirt.MitGegenueber("peer-gast").MitMaler(&malerdoppel{}).MitRunde(runde), naht, doppel, post, runde
}

func TestDerWirtReichtDasBildEinesAnderenGastesWeiter(t *testing.T) {
	t.Parallel()

	wirt, naht, cora, post, _ := einWeiterreichenderWirt(t)

	fertig := make(chan error, 1)
	go func() {
		fertig <- wirt.Angekommen(context.Background(), eineKontrollnachricht(t, domain.KontrolleBildwunsch,
			domain.Bildwunsch{Objekt: "coras", Seite: "vorne", Guete: "tisch"}))
	}()

	wunsch := warteAufBildwunsch(t, cora)

	// **Coras Antwort kommt über den Wirt ihres Sitzes** — er teilt die Post.
	coraswirt, _ := einWirt(&nehmerdoppel{})
	if fehler := coraswirt.MitBildpost(post).Angekommen(context.Background(), eineGastantwort(t,
		domain.Bildantwort{Wunsch: wunsch, Png: []byte{1, 2}, Breite: 5, Hoehe: 7})); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	select {
	case fehler := <-fertig:
		if fehler != nil {
			t.Fatalf("Angekommen: %v", fehler)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("der Gast wartet noch")
	}

	if antwort := bisZurBildantwort(t, naht); len(antwort.Png) != 2 || antwort.Hoehe != 7 || antwort.Wunsch.Objekt != "coras" {
		t.Errorf("die Antwort: %+v", antwort)
	}
}

// **Nach der eigenen Karte fragt der Wirt nicht zurück** — ein Gast, der
// danach fragt, bekommt einen Grund.
func TestNachSeinerEigenenKarteFragtDerWirtNichtZurueck(t *testing.T) {
	t.Parallel()

	wirt, naht, cora, _, _ := einWeiterreichenderWirt(t)

	if antwort := erfrage(t, wirt, naht, "eigenes"); antwort.Grund == "" {
		t.Errorf("die Antwort: %+v", antwort)
	}

	if len(cora.alle()) != 0 {
		t.Error("Cora wurde gefragt")
	}
}

// **Antwortet der andere Gast nicht, sagt es der Wirt** — mit einem Grund.
func TestOhneAntwortDesAnderenGastesKommtEinGrund(t *testing.T) {
	t.Parallel()

	wirt, naht, _, _, _ := einWeiterreichenderWirt(t)

	ctx, fertig := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer fertig()

	if fehler := wirt.Angekommen(ctx, eineKontrollnachricht(t, domain.KontrolleBildwunsch,
		domain.Bildwunsch{Objekt: "coras", Seite: "vorne", Guete: "tisch"})); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	if antwort := bisZurBildantwort(t, naht); antwort.Grund == "" {
		t.Errorf("die Antwort: %+v", antwort)
	}
}

// **Ein Deck aus einem dritten Bestand zeichnet niemand am Tisch** —
// Zweitblick über Phase 7, Befund 6. Gefragt wird nur ein Gast, dessen
// eigener Bestand die Zeile hält: nur er hat sein Deck mitgebracht, und nur
// wer mitbringt, ist neu genug für einen Bildwunsch. Ein Gast vor 0.9.0 an
// einem vorbelegten Sitz bekäme sonst eine Kontrollart, die er nicht kennt,
// und seine Sitzung endete.
func TestEinDeckAusEinemDrittenBestandFragtNiemanden(t *testing.T) {
	t.Parallel()

	wirt, naht, cora, _, _ := einWeiterreichenderWirt(t)

	if antwort := erfrage(t, wirt, naht, "geliehen"); antwort.Grund == "" {
		t.Errorf("die Antwort: %+v", antwort)
	}

	if len(cora.alle()) != 0 {
		t.Error("Cora wurde gefragt")
	}
}

// **Das Weiterreichen hält die Schleife nicht auf** — Zweitblick über
// Phase 7, Befund 5. Fragt Gast A nach einer Karte von B und zugleich B nach
// einer von A, wartete jede Wirtsschleife in `Angekommen` — und die Antwort,
// auf die sie wartet, muss durch die jeweils andere. Beide liefen bis zur
// Frist. Deshalb kehrt `Angekommen` zurück, bevor Cora antwortet.
func TestDasWeiterreichenHaeltDieSchleifeNichtAuf(t *testing.T) {
	t.Parallel()

	wirt, naht, cora, post, _ := einWeiterreichenderWirt(t)

	zurueck := make(chan error, 1)
	go func() {
		zurueck <- wirt.Angekommen(context.Background(), eineKontrollnachricht(t, domain.KontrolleBildwunsch,
			domain.Bildwunsch{Objekt: "coras", Seite: "vorne", Guete: "tisch"}))
	}()

	select {
	case fehler := <-zurueck:
		if fehler != nil {
			t.Fatalf("Angekommen: %v", fehler)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("die Schleife wartet auf Cora")
	}

	wunsch := warteAufBildwunsch(t, cora)

	coraswirt, _ := einWirt(&nehmerdoppel{})
	if fehler := coraswirt.MitBildpost(post).Angekommen(context.Background(), eineGastantwort(t,
		domain.Bildantwort{Wunsch: wunsch, Png: []byte{1, 2}, Breite: 5, Hoehe: 7})); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	if antwort := bisZurBildantwort(t, naht); len(antwort.Png) != 2 || antwort.Wunsch.Objekt != "coras" {
		t.Errorf("die Antwort: %+v", antwort)
	}
}

// bisZurBildantwort wartet, bis der Wirt ein Bild hinausgab — weitergereicht
// wird nebenher.
func bisZurBildantwort(t *testing.T, naht *nahtdoppel) domain.Bildantwort {
	t.Helper()

	frist := time.Now().Add(5 * time.Second)
	for time.Now().Before(frist) {
		if naht.anzahl() > 0 {
			return dieBildantwort(t, naht)
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatal("es ging kein Bild hinaus")

	return domain.Bildantwort{}
}

// anzahl zählt, was hinausging.
func (n *nahtdoppel) anzahl() int {
	n.schloss.Lock()
	defer n.schloss.Unlock()

	return len(n.gesendet)
}

// klemmendeLeitung nimmt an, gibt aber nichts hinaus, bis die Frist abläuft —
// ein Gast, der nicht mehr liest und die Verbindung offen hält.
type klemmendeLeitung struct {
	nahtdoppel
	zu atomic.Bool
}

func (n *klemmendeLeitung) Leere(ctx context.Context) error {
	<-ctx.Done()

	return ctx.Err()
}

func (n *klemmendeLeitung) Schliesse() error {
	n.zu.Store(true)

	return nil
}

// **Auch der Versand des Weitergereichten hat eine Frist, und ein Fehler
// legt die Leitung nieder** — Nachlauf zum Zweitblick über Phase 7, Rest R3.
// Hier stand der Kontext der Sitzung ohne Frist, und der Fehler verschwand.
func TestEinKlemmenderVersandLegtDieLeitungNieder(t *testing.T) {
	t.Parallel()

	_, _, _, _, runde := einWeiterreichenderWirt(t)
	naht := &klemmendeLeitung{}
	sitzung := service.NeuePartiesitzung("p1", "wirt", naht, service.NeueReihenfolge(1))
	wirt := service.NeuerPartiewirt("p1", "gast", &bildnehmer{objekte: []domain.Sichtobjekt{
		{Kennung: "coras", Bild: "b-coras", Eigner: "cora"},
	}, bestaende: map[string]string{"coras": "peer-cora"}}, sitzung).
			MitGegenueber("peer-gast").MitRunde(runde)

	// **Cora antwortet nie**, die Frage endet mit der Frist des Aufrufers —
	// und der Versand des Grundes klemmt.
	ctx, fertig := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer fertig()

	if fehler := wirt.Angekommen(ctx, eineKontrollnachricht(t, domain.KontrolleBildwunsch,
		domain.Bildwunsch{Objekt: "coras", Seite: "vorne", Guete: "tisch"})); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	frist := time.Now().Add(5 * time.Second)
	for !naht.zu.Load() && time.Now().Before(frist) {
		time.Sleep(time.Millisecond)
	}

	if !naht.zu.Load() {
		t.Error("die Leitung blieb offen")
	}
}

// **Eine Naht ohne Leitung hat nichts niederzulegen** — Schliesse tut dann
// nichts und klagt nicht.
func TestEineNahtOhneLeitungSchliesstStill(t *testing.T) {
	t.Parallel()

	sitzung := service.NeuePartiesitzung("p1", "wirt", &nahtdoppel{}, service.NeueReihenfolge(1))
	if fehler := sitzung.Schliesse(); fehler != nil {
		t.Errorf("es kam %v", fehler)
	}
}
