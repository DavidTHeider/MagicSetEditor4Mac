package app

// **Das Fenster rastert mit einem Renderplatz wie jede andere Fläche** —
// Folgebefund F3 des Lastlaufs QA-08
// (#2020).
//
// Das Fenster setzt sein Zeichenwerk in `zeichenwerkVerdrahten` zusammen, die
// Flächen ohne Fenster in `Zeichennaht` — beide über `NeuesZeichenwerk`, und
// `rasterplatz_test.go` hält diese Seite. **Hier die andere**: die Vorschau
// und die Probevorschau, die das Fenster über `a.rasterer` ruft, dazu
// Bildausfuhr, Druckbogen und Spieltextur an den Feldern des laufenden `App`.
//
// Gemessen wird wie dort: in dem Augenblick, in dem der Rasterer eine Schrift
// sucht, also mitten in der Rasterung, ist genau ein Renderplatz der Maschine
// belegt. Dazu der Beleg, dass die Vorschau bei besetzten Plätzen **wartet
// und abbricht** (`CodeAbgebrochen`, den die Oberfläche nicht zeigt), statt
// zu zeichnen.

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// fensterplatzmesser ist eine Schriftsuche, die nichts findet und dabei zählt,
// wie viele Renderplätze der Maschine gerade belegt sind.
type fensterplatzmesser struct {
	mu        sync.Mutex
	belegt    []int
	insgesamt int
}

func (m *fensterplatzmesser) Suche(string, bool) (render.Schrift, error) {
	var genommen []func()

	for {
		gib, frei := service.Maschinenhaushalt().Versuchen(domain.RessourceRendern)
		if !frei {
			break
		}

		genommen = append(genommen, gib)
	}

	for _, gib := range genommen {
		gib()
	}

	m.mu.Lock()
	m.belegt = append(m.belegt, m.insgesamt-len(genommen))
	m.mu.Unlock()

	return nil, errors.New("die Probe liefert keine Schrift")
}

func (m *fensterplatzmesser) abholen() []int {
	m.mu.Lock()
	defer m.mu.Unlock()

	gemessen := m.belegt
	m.belegt = nil

	return gemessen
}

func besetzeAlleRenderplaetze(t *testing.T) (freigeben func()) {
	t.Helper()

	var gehalten []func()

	for {
		gib, frei := service.Maschinenhaushalt().Versuchen(domain.RessourceRendern)
		if !frei {
			break
		}

		gehalten = append(gehalten, gib)
	}

	if len(gehalten) == 0 {
		t.Fatal("kein einziger Renderplatz war frei — ein anderer Lauf hält sie alle")
	}

	freigeben = func() {
		for _, gib := range gehalten {
			gib()
		}
	}

	t.Cleanup(freigeben)

	return freigeben
}

func TestDasFensterRastertMitBelegtemRenderplatz(t *testing.T) {
	anwendung := mitSaat(t)

	geoeffnet := anwendung.SetOeffnen(t.Context(), saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if len(karten.Karten) == 0 {
		t.Fatal("das Set trägt keine Karte")
	}

	karteID, fehler := strconv.ParseInt(karten.Karten[0].ID, 10, 64)
	if fehler != nil {
		t.Fatalf("Kennung %q: %v", karten.Karten[0].ID, fehler)
	}

	karte, fehler := anwendung.kartenleser.KarteLesen(probekontext(t), karteID)
	if fehler != nil {
		t.Fatalf("Karte lesen: %v", fehler)
	}

	rasterer, ok := anwendung.rasterer.(*render.Rasterer)
	if !ok {
		t.Fatalf("das Fenster rastert mit %T, erwartet war *render.Rasterer", anwendung.rasterer)
	}

	messer := &fensterplatzmesser{insgesamt: domain.Vorgabegrenzen().Parallel[domain.RessourceRendern]}
	rasterer.MitSchriften(messer)

	ort := t.TempDir()

	wege := []struct {
		name string
		ruf  func() error
	}{
		{"Vorschau", func() error {
			if befund := anwendung.Vorschau(t.Context(), karten.Karten[0].ID, 1).Fehler; befund != nil {
				return errors.New(befund.Meldung)
			}

			return nil
		}},
		{"Probevorschau", func() error {
			if befund := anwendung.Probevorschau(t.Context(), karten.Karten[0].ID, "name", "Eisregen", 1).Fehler; befund != nil {
				return errors.New(befund.Meldung)
			}

			return nil
		}},
		{"Bildausfuhr", func() error {
			_, err := anwendung.bildausfuhr.Schreibe(probekontext(t), karteID, ort, "karte.png",
				service.FormatPNG, 0, domain.BildkonfliktUeberschreiben, render.GueteMittel)

			return err
		}},
		{"Druckbogen", func() error {
			_, err := anwendung.bogenleger.Lege(probekontext(t), geoeffnetSetID(t, geoeffnet.SetID), ort+"/bogen.pdf",
				service.StandardBogen(), domain.Kartenauswahl{})

			return err
		}},
		{"Spieltextur", func() error {
			_, err := anwendung.texturen.Textur(probekontext(t), karte, service.SeiteVorne, service.GueteMiniatur)

			return err
		}},
	}

	for _, weg := range wege {
		if err := weg.ruf(); err != nil {
			t.Errorf("%s: %v", weg.name, err)
		}

		gemessen := messer.abholen()
		if len(gemessen) == 0 {
			t.Errorf("%s: die Rasterung hat keine Schrift gesucht — nichts gemessen", weg.name)
		}

		for _, belegt := range gemessen {
			if belegt != 1 {
				t.Errorf("%s: %d Renderplätze belegt mitten in der Rasterung, einer wäre richtig", weg.name, belegt)
			}
		}
	}
}

// geoeffnetSetID macht aus der Kennung des Fensters die des Bestands.
func geoeffnetSetID(t *testing.T, kennung string) int64 {
	t.Helper()

	setID, fehler := strconv.ParseInt(kennung, 10, 64)
	if fehler != nil {
		t.Fatalf("Setkennung %q: %v", kennung, fehler)
	}

	return setID
}

// **Sind alle Plätze belegt, wartet das Fenster** — am Rasterer, den seine
// Vorschau ruft, mit einem Plan, der ohne Warten in Mikrosekunden fertig wäre.
// Was die Vorschau daraus macht, wenn ihr Lauf an der Frist endet — den Abbruch,
// den die Oberfläche nicht zeigt —, halten die Tests der Fehlerübersetzung.
func TestDasFensterWartetBeiBesetztenRenderplaetzen(t *testing.T) {
	anwendung := mitSaat(t)

	freigeben := besetzeAlleRenderplaetze(t)

	frist, abbrechen := context.WithTimeout(t.Context(), 150*time.Millisecond)
	defer abbrechen()

	plan := &render.Plan{Breite: 8, Hoehe: 8, Ebenen: []render.Ebene{{
		Art: render.ArtFlaeche, Deckung: 1,
		Rahmen: render.Rechteck{Breite: 8, Hoehe: 8}, Farbe: render.Farbe{R: 200, A: 255},
	}}}

	if _, _, err := anwendung.rasterer.Zeichne(frist, plan, 1, render.GueteMittel); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("mit besetzten Plätzen: %v — erwartet war das Ende der Frist", err)
	}

	freigeben()

	if _, _, err := anwendung.rasterer.Zeichne(t.Context(), plan, 1, render.GueteMittel); err != nil {
		t.Errorf("mit freien Plätzen: %v", err)
	}
}
