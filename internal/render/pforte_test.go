package render_test

import (
	"context"
	"errors"
	"image"
	"image/color"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Eine Rasterung belegt einen Platz, bevor sie zeichnet** — Folgebefund F3
// des Lastlaufs QA-08
// (#2020).
//
// # Was sie hält
//
//   - **Der Platz ist belegt, solange gezeichnet wird** — gemessen in dem
//     Augenblick, in dem der Rasterer ein Bild von der Quelle holt, also
//     mitten in der Rasterung und nicht davor oder danach.
//   - **Wer keinen bekommt, zeichnet nicht**: kein Griff in die Bildquelle,
//     und der Fehler ist der der Pforte.
//   - **Der Platz geht zurück** — nach dem Erfolg, nach einem Abbruch mitten
//     im Lauf, und über beide Wege, die rastern ([render.Rasterer.Zeichne] und
//     [render.Rasterer.ZeichneMitHerkunft]).
//   - **Ein Plan, der ohnehin abgewiesen wird, wartet nicht**: die Prüfung
//     steht vor der Pforte.
//
// # Was sie nicht hält
//
// **Dass die Pforte den Haushalt der Maschine fragt.** `render` kennt keinen;
// das halten `internal/service` und `internal/app` an ihrer Naht.

// platzprobe ist eine Pforte, die mitzählt und sich anhalten lässt.
type platzprobe struct {
	mu          sync.Mutex
	belegt      int
	freigegeben int
	aktiv       int

	// angefragt meldet, dass jemand an der Pforte steht; einfahrt hält ihn dort
	// fest, bis sie geschlossen wird.
	angefragt chan struct{}
	einfahrt  chan struct{}

	// ablehnen ist der Fehler, mit dem die Pforte abweist.
	ablehnen error

	// dahinter läuft, sobald der Platz vergeben ist — mit dem Kontext des Rufs.
	dahinter func()
}

func (p *platzprobe) Belegen(ctx context.Context) (func(), error) {
	if p.angefragt != nil {
		p.angefragt <- struct{}{}
	}

	if p.einfahrt != nil {
		select {
		case <-p.einfahrt:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if p.ablehnen != nil {
		return nil, p.ablehnen
	}

	p.mu.Lock()
	p.belegt++
	p.aktiv++
	p.mu.Unlock()

	if p.dahinter != nil {
		p.dahinter()
	}

	return func() {
		p.mu.Lock()
		defer p.mu.Unlock()

		p.freigegeben++
		p.aktiv--
	}, nil
}

func (p *platzprobe) stand() (belegt, freigegeben, aktiv int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.belegt, p.freigegeben, p.aktiv
}

// quellprobe ist eine Bildquelle, die merkt, wie viele Plätze beim Abruf
// belegt waren.
type quellprobe struct {
	pforte *platzprobe

	mu       sync.Mutex
	abrufe   int
	aktivBei []int
}

func (q *quellprobe) Bild(string) (image.Image, error) {
	_, _, aktiv := q.pforte.stand()

	q.mu.Lock()
	q.abrufe++
	q.aktivBei = append(q.aktivBei, aktiv)
	q.mu.Unlock()

	return einfarbig(4, 4, color.NRGBA{R: 200, A: 255}), nil
}

// erwarte nimmt einen Wert aus einem Kanal — und scheitert, statt bis zum
// Zeitablauf des ganzen Laufs zu warten, wenn er nie kommt.
func erwarte[T any](t *testing.T, kanal <-chan T, was string) T {
	t.Helper()

	select {
	case wert := <-kanal:
		return wert
	case <-time.After(10 * time.Second):
		t.Fatalf("%s", was)

		var leer T

		return leer
	}
}

func (q *quellprobe) zaehler() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.abrufe
}

// planMitBild ist ein Plan, der die Bildquelle braucht — und damit einen
// Augenblick mitten in der Rasterung hat, an dem man hinsehen kann.
func planMitBild() *render.Plan {
	return &render.Plan{
		Breite: 20, Hoehe: 20,
		Ebenen: []render.Ebene{{
			Art:     render.ArtBild,
			Deckung: 1,
			Rahmen:  render.Rechteck{Breite: 20, Hoehe: 20},
			Bild:    render.Bildbaum{Knotenart: "datei", Texte: map[string]string{"pfad": "/rahmen.png"}},
		}},
	}
}

func rastererMitPforte(pforte *platzprobe) (*render.Rasterer, *quellprobe) {
	quelle := &quellprobe{pforte: pforte}

	return render.NeuerRasterer(quelle).MitPforte(pforte), quelle
}

func TestEineRasterungZeichnetMitBelegtemPlatzUndGibtIhnZurueck(t *testing.T) {
	t.Parallel()

	wege := map[string]func(*render.Rasterer) error{
		"Zeichne": func(r *render.Rasterer) error {
			_, _, err := r.Zeichne(t.Context(), planMitBild(), 1, render.GueteMittel)

			return err
		},
		"ZeichneMitHerkunft": func(r *render.Rasterer) error {
			_, _, _, err := r.ZeichneMitHerkunft(t.Context(), planMitBild(), 1)

			return err
		},
	}

	for name, weg := range wege {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pforte := &platzprobe{}
			rasterer, quelle := rastererMitPforte(pforte)

			if err := weg(rasterer); err != nil {
				t.Fatalf("zeichnen: %v", err)
			}

			// **Mitten im Lauf belegt**: beim Abruf des Bildes stand genau ein
			// Platz auf Rechnung — nicht null davor, nicht mehr danach.
			if quelle.zaehler() == 0 || quelle.aktivBei[0] != 1 {
				t.Errorf("Plätze beim Abruf des Bildes: %v (%d Abrufe)", quelle.aktivBei, quelle.zaehler())
			}

			if belegt, frei, aktiv := pforte.stand(); belegt != 1 || frei != 1 || aktiv != 0 {
				t.Errorf("belegt %d, freigegeben %d, noch belegt %d", belegt, frei, aktiv)
			}
		})
	}
}

// **Wer an der Pforte steht, zeichnet nicht** — auch kein Stück davon.
func TestEineRasterungWartetAnDerPforteUndZeichnetErstDanach(t *testing.T) {
	t.Parallel()

	pforte := &platzprobe{angefragt: make(chan struct{}), einfahrt: make(chan struct{})}
	rasterer, quelle := rastererMitPforte(pforte)

	fertig := make(chan error, 1)

	go func() {
		_, _, err := rasterer.Zeichne(t.Context(), planMitBild(), 1, render.GueteMittel)
		fertig <- err
	}()

	// **Erst der Beleg, dass jemand an der Pforte steht** — sonst sagte das
	// Nein darunter nichts: ein Lauf, der gar nicht gestartet ist, hat auch
	// noch nichts gezeichnet.
	erwarte(t, pforte.angefragt, "die Rasterung nicht an der Pforte angekommen")

	if abrufe := quelle.zaehler(); abrufe != 0 {
		t.Fatalf("vor der Pforte wurde gezeichnet: %d Abrufe", abrufe)
	}

	close(pforte.einfahrt)

	if err := erwarte(t, fertig, "die Rasterung nach der Einfahrt nicht fertig"); err != nil {
		t.Fatalf("nach der Einfahrt: %v", err)
	}

	if quelle.zaehler() == 0 {
		t.Error("nach der Einfahrt wurde nicht gezeichnet")
	}
}

// **Ein endender Kontext holt den Wartenden von der Pforte** — er zeichnet
// nicht, und es bleibt kein Platz auf seinem Namen zurück.
func TestEinEndenderKontextBeendetDasWartenAnDerPforte(t *testing.T) {
	t.Parallel()

	pforte := &platzprobe{angefragt: make(chan struct{}), einfahrt: make(chan struct{})}
	rasterer, quelle := rastererMitPforte(pforte)

	ctx, abbrechen := context.WithCancel(t.Context())
	fertig := make(chan error, 1)

	go func() {
		_, _, err := rasterer.Zeichne(ctx, planMitBild(), 1, render.GueteMittel)
		fertig <- err
	}()

	erwarte(t, pforte.angefragt, "die Rasterung nicht an der Pforte angekommen")

	abbrechen()

	if err := erwarte(t, fertig, "der Abbruch beendet das Warten nicht"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Fehler %v, erwartet war context.Canceled", err)
	}

	if quelle.zaehler() != 0 {
		t.Errorf("%d Abrufe, obwohl der Lauf abgebrochen war", quelle.zaehler())
	}

	if belegt, _, _ := pforte.stand(); belegt != 0 {
		t.Errorf("%d Plätze auf den Namen eines Abgebrochenen", belegt)
	}
}

// **Was die Pforte abweist, kommt unverändert zurück.**
func TestDieAbweisungDerPforteIstDerFehlerDerRasterung(t *testing.T) {
	t.Parallel()

	voll := errors.New("kein Platz")
	pforte := &platzprobe{ablehnen: voll}
	rasterer, quelle := rastererMitPforte(pforte)

	bild, befunde, err := rasterer.Zeichne(t.Context(), planMitBild(), 1, render.GueteMittel)
	if !errors.Is(err, voll) || bild != nil || befunde != nil {
		t.Errorf("bild %v, befunde %v, Fehler %v", bild, befunde, err)
	}

	if quelle.zaehler() != 0 {
		t.Errorf("%d Abrufe trotz Abweisung", quelle.zaehler())
	}
}

// **Ein Lauf, der mittendrin endet, gibt den Platz zurück** — sonst verlöre
// jeder Abbruch einen der zwei Plätze, bis keiner mehr übrig ist.
func TestEinAbbruchMitteninGibtDenPlatzZurueck(t *testing.T) {
	t.Parallel()

	ctx, abbrechen := context.WithCancel(t.Context())
	pforte := &platzprobe{dahinter: abbrechen}
	rasterer, quelle := rastererMitPforte(pforte)

	if _, _, err := rasterer.Zeichne(ctx, planMitBild(), 1, render.GueteMittel); !errors.Is(err, context.Canceled) {
		t.Fatalf("Fehler %v, erwartet war context.Canceled", err)
	}

	if belegt, frei, aktiv := pforte.stand(); belegt != 1 || frei != 1 || aktiv != 0 {
		t.Errorf("belegt %d, freigegeben %d, noch belegt %d", belegt, frei, aktiv)
	}

	if quelle.zaehler() != 0 {
		t.Errorf("%d Abrufe nach dem Abbruch", quelle.zaehler())
	}
}

// **Ein Plan, der ohnehin abgewiesen wird, steht nicht an.** Die Prüfung des
// Plans und der Vergrößerung gehört vor die Pforte: wer ohnehin scheitert,
// soll keinen Platz belegen — und keinen Wartenden aufhalten.
func TestEinUnzeichenbarerPlanFragtDiePforteNicht(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		plan           *render.Plan
		vergroesserung float64
	}{
		"kein Plan":               {nil, 1},
		"Plan ohne Maß":           {&render.Plan{Hoehe: 20}, 1},
		"Vergrößerung null":       {planMitBild(), 0},
		"Vergrößerung negativ":    {planMitBild(), -1},
		"Vergrößerung keine Zahl": {planMitBild(), math.NaN()},
		"Bild ohne Bildpunkte":    {planMitBild(), 0.01},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pforte := &platzprobe{}
			rasterer, _ := rastererMitPforte(pforte)

			if _, _, err := rasterer.Zeichne(t.Context(), fall.plan, fall.vergroesserung, render.GueteMittel); !errors.Is(err, render.ErrPlan) {
				t.Fatalf("Fehler %v, erwartet war render.ErrPlan", err)
			}

			if belegt, _, _ := pforte.stand(); belegt != 0 {
				t.Errorf("%d Plätze für einen Plan, der scheitert", belegt)
			}
		})
	}
}

// **Ohne Pforte ist jede Rasterung frei** — so wie bisher, und so bleibt es
// für jeden, der keinen Haushalt kennt.
func TestOhnePforteZeichnetEinRastererWieBisher(t *testing.T) {
	t.Parallel()

	quelle := &quellprobe{pforte: &platzprobe{}}

	if _, _, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), planMitBild(), 1, render.GueteMittel); err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if quelle.zaehler() == 0 {
		t.Error("es wurde nicht gezeichnet")
	}
}
