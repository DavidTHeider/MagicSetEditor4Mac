package app_test

// **Zeichnen belegt einen Renderplatz, an jeder Fläche** — Folgebefund F3 des
// Lastlaufs QA-08 (#2020).
//
// # Was er hält
//
// Beide Composition Roots — das Fenster und die Flächen ohne Fenster
// ([app.Zeichennaht]) — bauen ihr Zeichenwerk über [app.NeuesZeichenwerk].
// Geprüft wird an dem **einen** Haushalt, den auch `GET /grenzen` nennt
// (`service.Maschinenhaushalt`), nicht an einer Attrappe daneben:
//
//   - **Jeder Weg, der eine Karte rastert, tut es mit einem belegten Platz**
//     — Kartenbild (`mse zeichnen`, `schreiben.karte.zeichnen`,
//     `POST /karten/{}/bild`), Druckbogen (`mse bogen`, `schreiben.set.bogen`,
//     `POST /sets/{}/bogen`), Spieltextur und die Rasterung selbst, die auch
//     die Vorschau im Fenster ruft. Gemessen wird in dem Augenblick, in dem der
//     Rasterer eine Schrift sucht, also **mitten in der Rasterung**: führt ein
//     Weg an der Pforte vorbei, ist dort kein Platz belegt. Das ist ein
//     Ereignis und keine Frist — der Lauf hängt nicht an der Geschwindigkeit
//     der Maschine.
//   - **Sind alle Plätze belegt, wartet die Rasterung** und kommt erst durch,
//     wenn einer frei wird; mit der Frist als Beleg, an einem Plan, der ohne
//     Warten in Mikrosekunden fertig wäre.
//
// # Was er nicht hält
//
// Dass ein Platz auch wieder zurückkommt, wenn der Lauf scheitert: das halten
// `internal/render/pforte_test.go` und `internal/service/rasterpforte_test.go`.
// Und das Fenster hat seinen eigenen Lauf (`rasterplatzfenster_intern_test.go`).

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// besetzeDieRenderplaetze nimmt alle Plätze der Klasse rendern, die der
// Maschinenhaushalt hergibt, und gibt eine Rückgabe dafür.
//
// **Alle und nicht zwei**: ein Platz, den ein anderer Lauf in diesem Prozess
// schon hält, ändert nichts daran, dass danach keiner mehr frei ist.
func besetzeDieRenderplaetze(t *testing.T) (freigeben func()) {
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

// platzmesser ist eine Schriftsuche, die nichts findet und dabei zählt, wie
// viele Renderplätze der Maschine gerade belegt sind.
//
// **Sie sitzt in der Rasterung**: der Rasterer fragt sie, während er eine
// Textebene zeichnet — nach der Pforte, vor der Rückgabe des Platzes.
type platzmesser struct {
	mu        sync.Mutex
	belegt    []int
	insgesamt int
}

func (m *platzmesser) Suche(string, bool) (render.Schrift, error) {
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

// abholen gibt die Messungen seit dem letzten Mal und vergisst sie.
func (m *platzmesser) abholen() []int {
	m.mu.Lock()
	defer m.mu.Unlock()

	gemessen := m.belegt
	m.belegt = nil

	return gemessen
}

// probesetEinfuehren führt das Probeset des Repositoriums in den Bestand ein.
func probesetEinfuehren(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	ergebnis, fehler := app.Einfuhrnaht(db).Set(probekontext(t),
		filepath.Join("..", "..", "testdaten", "probe.mse-set"))
	if fehler != nil {
		t.Fatalf("einfahren: %v", fehler)
	}

	return ergebnis.SetID
}

// ersteKarteDesSets gibt die erste Karte eines Sets.
func ersteKarteDesSets(t *testing.T, werk *app.Zeichenwerk, set int64) *domain.CardRow {
	t.Helper()

	karten, fehler := werk.Leser.Karten(probekontext(t), set)
	if fehler != nil || len(karten) == 0 {
		t.Fatalf("Karten lesen: %v (%d)", fehler, len(karten))
	}

	return karten[0]
}

// winzigerPlan ist ein Plan, der ohne Warten in Mikrosekunden fertig ist.
func winzigerPlan() *render.Plan {
	return &render.Plan{Breite: 8, Hoehe: 8, Ebenen: []render.Ebene{{
		Art: render.ArtFlaeche, Deckung: 1,
		Rahmen: render.Rechteck{Breite: 8, Hoehe: 8}, Farbe: render.Farbe{R: 200, A: 255},
	}}}
}

// zeichenwerkDerWurzeln baut das Zeichenwerk auf beiden Wegen, die es gibt.
func zeichenwerkDerWurzeln(t *testing.T) map[string]func(*testing.T) (*app.Zeichenwerk, int64) {
	t.Helper()

	saat, err := filepath.Abs(filepath.Join("..", "..", "saat"))
	if err != nil {
		t.Fatalf("Saatpfad: %v", err)
	}

	return map[string]func(*testing.T) (*app.Zeichenwerk, int64){
		"NeuesZeichenwerk (das Fenster und jeder Test)": func(t *testing.T) (*app.Zeichenwerk, int64) {
			db := offeneProbedatenbank(t)
			set := probesetEinfuehren(t, db)

			return app.NeuesZeichenwerk(storage.NeuerSetLeser(db), []string{saat}, t.TempDir()), set
		},
		"Zeichennaht (mse, mse-rest, mse-mcp)": func(t *testing.T) (*app.Zeichenwerk, int64) {
			paketMitEigenemZuhause(t)

			db := offeneProbedatenbank(t)
			set := probesetEinfuehren(t, db)

			werk, fehler := app.Zeichennaht(db)
			if fehler != nil {
				t.Fatalf("die Zeichennaht: %v", fehler)
			}

			return werk, set
		},
	}
}

func TestJederZeichenwegRastertMitBelegtemRenderplatz(t *testing.T) {
	for name, baue := range zeichenwerkDerWurzeln(t) {
		t.Run(name, func(t *testing.T) {
			werk, set := baue(t)
			karte := ersteKarteDesSets(t, werk, set)
			ort := t.TempDir()

			messer := &platzmesser{insgesamt: domain.Vorgabegrenzen().Parallel[domain.RessourceRendern]}
			werk.Rasterer.MitSchriften(messer)

			plan, _, fehler := werk.Planbauer.Kartenplan(probekontext(t), karte)
			if fehler != nil {
				t.Fatalf("Plan: %v", fehler)
			}

			wege := []struct {
				name string
				ruf  func() error
			}{
				{"Kartenbild", func() error {
					_, err := werk.SchreibeBild(probekontext(t), karte.ID, filepath.Join(ort, "karte.png"), render.GueteMittel)

					return err
				}},
				{"Druckbogen", func() error {
					_, err := werk.LegeBogen(probekontext(t), set, filepath.Join(ort, "bogen.pdf"))

					return err
				}},
				{"die Rasterung selbst (Vorschau, Palette)", func() error {
					_, _, err := werk.Rasterer.Zeichne(probekontext(t), plan, 1, render.GueteMittel)

					return err
				}},
				{"Spieltextur", func() error {
					_, err := werk.Spielbilder.Textur(probekontext(t), karte, service.SeiteVorne, service.GueteMiniatur)

					return err
				}},
			}

			for _, weg := range wege {
				if err := weg.ruf(); err != nil {
					t.Errorf("%s: %v", weg.name, err)
				}

				// **Erst der Beleg, dass der Rasterer bis zur Schrift kam** —
				// sonst sagte „kein Platz belegt" nichts.
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
		})
	}
}

func TestSindAlleRenderplaetzeBelegtWartetDieRasterungBisEinerFreiWird(t *testing.T) {
	for name, baue := range zeichenwerkDerWurzeln(t) {
		t.Run(name, func(t *testing.T) {
			werk, _ := baue(t)

			freigeben := besetzeDieRenderplaetze(t)

			frist, abbrechen := context.WithTimeout(probekontext(t), 150*time.Millisecond)
			defer abbrechen()

			// **Ein Plan, der ohne Warten in Mikrosekunden fertig wäre.** Dass
			// er hier bis zur Frist braucht, ist das Warten und nichts sonst.
			if _, _, err := werk.Rasterer.Zeichne(frist, winzigerPlan(), 1, render.GueteMittel); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("mit besetzten Plätzen: %v — erwartet war das Ende der Frist", err)
			}

			freigeben()

			if _, _, err := werk.Rasterer.Zeichne(probekontext(t), winzigerPlan(), 1, render.GueteMittel); err != nil {
				t.Errorf("mit freien Plätzen: %v", err)
			}
		})
	}
}
