package service_test

import (
	"context"
	"errors"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Zeichnen belegt einen Renderplatz, und wer keinen bekommt, wartet** —
// Folgebefund F3 des Lastlaufs QA-08
// (#2020).
//
// # Was sie hält
//
//   - **Die Pforte belegt die Klasse `rendern` und keine andere.**
//   - **Ein Haushalt mit einem Platz lässt zwei Rufe nicht zugleich zeichnen**:
//     der zweite wartet, bis der erste freigibt, und ein endender Kontext
//     holt ihn aus der Schlange, ohne dass ein Platz verloren geht.
//   - **Der ganze Weg**: ein Kartenbild über [service.Bildausfuhr] und den
//     echten Rasterer, nicht über eine Attrappe — die Zusage hängt an der
//     Naht zwischen beiden, und eine Attrappe an dieser Stelle bestätigte nur
//     sich selbst.
//
// # Was sie nicht hält
//
// **Dass `internal/app` die Pforte an den Rasterer hängt**, den alle Flächen
// bauen — das hält `internal/app/rasterplatz_test.go`.

// haushaltMitPlaetzen baut einen Haushalt mit `anzahl` Plätzen der Klasse rendern.
func haushaltMitPlaetzen(anzahl int) *service.Haushalt {
	grenzen := domain.Vorgabegrenzen()
	grenzen.Parallel[domain.RessourceRendern] = anzahl

	return service.NeuerHaushalt(grenzen)
}

func pforteMit(haushalt *service.Haushalt) *service.Rasterpforte {
	pforte := service.NeueRasterpforte()
	pforte.MitHaushalt(haushalt)

	return pforte
}

func TestDiePforteBelegtDieKlasseRendernUndKeineAndere(t *testing.T) {
	t.Parallel()

	haushalt := haushaltMitPlaetzen(1)

	freigeben, err := pforteMit(haushalt).Belegen(t.Context())
	if err != nil {
		t.Fatalf("belegen: %v", err)
	}

	if _, frei := haushalt.Versuchen(domain.RessourceRendern); frei {
		t.Error("die Pforte hat keinen Platz der Klasse rendern belegt")
	}

	for _, klasse := range []domain.Ressourcenklasse{
		domain.RessourceAusfuhr, domain.RessourceEinfuhr, domain.RessourceBild,
	} {
		gib, frei := haushalt.Versuchen(klasse)
		if !frei {
			t.Errorf("die Pforte hat auch einen Platz der Klasse %s belegt", klasse)

			continue
		}

		gib()
	}

	freigeben()

	if gib, frei := haushalt.Versuchen(domain.RessourceRendern); !frei {
		t.Error("der Platz kam nicht zurück")
	} else {
		gib()
	}
}

// **Der zweite wartet, bis der erste freigibt** — und bekommt dann den Platz.
func TestEineZweiteRasterungWartetBisDieErsteFreigibt(t *testing.T) {
	t.Parallel()

	pforte := pforteMit(haushaltMitPlaetzen(1))

	erster, err := pforte.Belegen(t.Context())
	if err != nil {
		t.Fatalf("der erste: %v", err)
	}

	// **Wer wartet, kommt nicht durch** — der Beleg ist die Frist: ohne
	// Pforte, die wartet, wäre er sofort hindurch.
	frist, abbrechen := context.WithTimeout(t.Context(), 30*time.Millisecond)
	defer abbrechen()

	if _, err := pforte.Belegen(frist); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("der zweite kam durch oder scheiterte anders: %v", err)
	}

	// **Und er bekommt ihn, sobald der erste freigibt.**
	zweiter := make(chan error, 1)

	go func() {
		gib, err := pforte.Belegen(t.Context())
		if err == nil {
			gib()
		}

		zweiter <- err
	}()

	erster()

	select {
	case err := <-zweiter:
		if err != nil {
			t.Errorf("der zweite nach der Freigabe: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("der zweite bekam nach der Freigabe keinen Platz")
	}
}

// **Ein endender Kontext verliert keinen Platz.** Wer in der Schlange steht
// und aufgibt, hat nichts genommen — sonst schrumpfte der Haushalt mit jedem
// Abbruch, bis kein Platz mehr da wäre.
func TestEinEndenderKontextVerliertKeinenPlatz(t *testing.T) {
	t.Parallel()

	haushalt := haushaltMitPlaetzen(1)
	pforte := pforteMit(haushalt)

	erster, err := pforte.Belegen(t.Context())
	if err != nil {
		t.Fatalf("der erste: %v", err)
	}

	for range 10 {
		ctx, abbrechen := context.WithCancel(t.Context())
		abbrechen()

		if _, err := pforte.Belegen(ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("ein beendeter Kontext bekam %v", err)
		}
	}

	erster()

	gib, frei := haushalt.Versuchen(domain.RessourceRendern)
	if !frei {
		t.Fatal("nach zehn Abbrüchen ist der Platz weg")
	}

	gib()
}

// blockierendeQuelle ist eine Bildquelle, an der man eine Rasterung anhalten
// kann: sie meldet jeden Abruf und lässt ihn erst weiter, wenn der Test es sagt.
type blockierendeQuelle struct {
	mu     sync.Mutex
	aktiv  int
	hoechs int
	abrufe int

	// betreten meldet jeden Abruf (gepuffert, damit ein ungewollter Abruf
	// nicht schon am Melden hängt); weiter hält ihn fest, bis es geschlossen ist.
	betreten chan struct{}
	weiter   chan struct{}
}

func (q *blockierendeQuelle) Bild(string) (image.Image, error) {
	q.mu.Lock()
	q.aktiv++
	q.abrufe++
	q.hoechs = max(q.hoechs, q.aktiv)
	q.mu.Unlock()

	defer func() {
		q.mu.Lock()
		q.aktiv--
		q.mu.Unlock()
	}()

	q.betreten <- struct{}{}

	// **Nie ewig**: ein Ruf, der hier ungewollt landet, soll den Lauf
	// scheitern lassen, statt ihn bis zum Zeitablauf aufzuhalten.
	select {
	case <-q.weiter:
	case <-time.After(3 * time.Second):
	}

	bild := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for x := range 4 {
		for y := range 4 {
			bild.SetNRGBA(x, y, color.NRGBA{R: 200, A: 255})
		}
	}

	return bild, nil
}

func (q *blockierendeQuelle) stand() (hoechstens, abrufe int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.hoechs, q.abrufe
}

// bildplan ist ein Planbauer, dessen Karte ein Bild aus der Quelle braucht —
// und damit eine Stelle mitten in der Rasterung hat.
type bildplan struct{}

func (bildplan) Kartenplan(context.Context, *domain.CardRow) (*render.Plan, []domain.Befund, error) {
	return &render.Plan{
		Breite: 20, Hoehe: 20,
		Ebenen: []render.Ebene{{
			Art:     render.ArtBild,
			Deckung: 1,
			Rahmen:  render.Rechteck{Breite: 20, Hoehe: 20},
			Bild:    render.Bildbaum{Knotenart: "datei", Texte: map[string]string{"pfad": "/rahmen.png"}},
		}},
	}, nil, nil
}

// kartenbild schreibt die erste Karte als PNG in `ort`.
func kartenbild(
	ctx context.Context, ausfuhr *service.Bildausfuhr, ort, datei string,
) error {
	_, err := ausfuhr.Schreibe(ctx, 1, ort, datei, service.FormatPNG, 0,
		domain.BildkonfliktUeberschreiben, render.GueteMittel)

	return err
}

func erwartet[T any](t *testing.T, kanal <-chan T, was string) T {
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

// **Mit einem Platz zeichnet immer nur einer** — über den echten Weg vom
// Kartenbild zum Rasterer.
func TestMitEinemPlatzZeichnetNurEinKartenbildZugleich(t *testing.T) {
	t.Parallel()

	quelle := &blockierendeQuelle{betreten: make(chan struct{}, 8), weiter: make(chan struct{})}
	rasterer := render.NeuerRasterer(quelle).MitPforte(pforteMit(haushaltMitPlaetzen(1)))
	ausfuhr := service.NeueBildausfuhr(kartenDouble{leserDouble: karten("Feuerball")}, bildplan{}, rasterer)
	ort := t.TempDir()

	erster := make(chan error, 1)

	go func() { erster <- kartenbild(t.Context(), ausfuhr, ort, "eins.png") }()

	// **Der erste ist mitten in der Rasterung** und hält den Platz.
	erwartet(t, quelle.betreten, "der erste Ruf kam nicht bis zur Bildquelle")

	// **Der zweite wartet**: seine Frist läuft ab, und er hat die Bildquelle
	// nie gesehen.
	frist, abbrechen := context.WithTimeout(t.Context(), 30*time.Millisecond)
	defer abbrechen()

	if err := kartenbild(frist, ausfuhr, ort, "zwei.png"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("der zweite Ruf zeichnete zugleich oder scheiterte anders: %v", err)
	}

	if _, abrufe := quelle.stand(); abrufe != 1 {
		t.Errorf("%d Abrufe der Bildquelle, einer wäre richtig", abrufe)
	}

	// **Gibt der erste frei, zeichnet der nächste.**
	close(quelle.weiter)

	if err := erwartet(t, erster, "der erste Ruf wurde nicht fertig"); err != nil {
		t.Fatalf("der erste Ruf: %v", err)
	}

	if err := kartenbild(t.Context(), ausfuhr, ort, "drei.png"); err != nil {
		t.Fatalf("der Ruf nach der Freigabe: %v", err)
	}

	if hoechstens, _ := quelle.stand(); hoechstens != 1 {
		t.Errorf("zeitweise zeichneten %d zugleich", hoechstens)
	}

	for _, name := range []string{"eins.png", "drei.png"} {
		if _, err := os.Stat(filepath.Join(ort, name)); err != nil {
			t.Errorf("%s fehlt: %v", name, err)
		}
	}

	// **Und der zweite hat nichts hinterlassen** — er wurde abgewiesen, bevor
	// er etwas schrieb.
	if _, err := os.Stat(filepath.Join(ort, "zwei.png")); err == nil {
		t.Error("der abgewiesene Ruf hat trotzdem eine Datei geschrieben")
	}
}

// **Mit zwei Plätzen zeichnen zwei zugleich** — der Beleg, dass das Warten
// oben an der Grenze liegt und nicht an der Vorrichtung: dieselben Rufe, ein
// Platz mehr, und beide sind gleichzeitig in der Rasterung.
func TestMitZweiPlaetzenZeichnenZweiKartenbilderZugleich(t *testing.T) {
	t.Parallel()

	quelle := &blockierendeQuelle{betreten: make(chan struct{}, 8), weiter: make(chan struct{})}
	rasterer := render.NeuerRasterer(quelle).MitPforte(pforteMit(haushaltMitPlaetzen(2)))
	ausfuhr := service.NeueBildausfuhr(kartenDouble{leserDouble: karten("Feuerball")}, bildplan{}, rasterer)
	ort := t.TempDir()

	ergebnisse := make(chan error, 3)

	for _, datei := range []string{"eins.png", "zwei.png", "drei.png"} {
		go func() { ergebnisse <- kartenbild(t.Context(), ausfuhr, ort, datei) }()
	}

	// **Zwei sind drin, ohne dass einer ging** — der dritte wartet.
	erwartet(t, quelle.betreten, "der erste Ruf kam nicht bis zur Bildquelle")
	erwartet(t, quelle.betreten, "der zweite Ruf kam nicht bis zur Bildquelle")

	if hoechstens, _ := quelle.stand(); hoechstens != 2 {
		t.Errorf("gleichzeitig in der Rasterung: %d, erwartet 2", hoechstens)
	}

	// Danach geht es reihum weiter: der dritte kommt, sobald einer frei ist.
	close(quelle.weiter)

	for range 3 {
		if err := erwartet(t, ergebnisse, "ein Ruf wurde nicht fertig"); err != nil {
			t.Errorf("ein Ruf scheiterte: %v", err)
		}
	}

	if hoechstens, _ := quelle.stand(); hoechstens != 2 {
		t.Errorf("zeitweise zeichneten %d zugleich", hoechstens)
	}
}
