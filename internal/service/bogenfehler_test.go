// Die Fehlerwege des Bogenlegers.
//
// **Eine eigene Datei** (AGENTS.md §2), und der Schnitt folgt der Frage: dort
// steht, was einen Bogen ausmacht, hier, was geschieht, wenn keiner entsteht.
package service_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Maße, mit denen kein Bogen entstehen kann, werden vorher abgewiesen.**
//
// Vorher heißt: bevor gerastert wird. Wer erst dreihundert Karten zeichnet und
// dann merkt, dass keine auf die Seite passt, hat eine Minute verschenkt.
func TestUnmoeglicheBogenmasse(t *testing.T) {
	faelle := map[string]func(m *service.Bogenmasse){
		"eine Karte ohne Breite":  func(m *service.Bogenmasse) { m.Kartenbreite = 0 },
		"eine Karte ohne Höhe":    func(m *service.Bogenmasse) { m.Kartenhoehe = 0 },
		"ein negativer Anschnitt": func(m *service.Bogenmasse) { m.Anschnitt = -1 },
		"ein negativer Rand":      func(m *service.Bogenmasse) { m.Rand = -1 },
		"keine Auflösung":         func(m *service.Bogenmasse) { m.DPI = 0 },
		"eine Auflösung über der Grenze": func(m *service.Bogenmasse) {
			m.DPI = service.GrenzeDPI + 1
		},
		"eine Karte, die größer ist als das Papier": func(m *service.Bogenmasse) {
			m.Kartenbreite = 300
		},
		"ein Rand, der die Seite füllt": func(m *service.Bogenmasse) { m.Rand = 140 },
	}

	for name, aendern := range faelle {
		t.Run(name, func(t *testing.T) {
			masse := service.StandardBogen()
			aendern(&masse)

			ziel := filepath.Join(t.TempDir(), "bogen.pdf")

			_, err := bogenleger(t, 1).Lege(t.Context(), 7, ziel, masse, domain.Kartenauswahl{})
			if err == nil {
				t.Fatal("kein Fehler")
			}

			if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Fehler: %v, erwartet ErrValidation", err)
			}

			if _, fehler := os.Stat(ziel); !errors.Is(fehler, os.ErrNotExist) {
				t.Error("es wurde trotzdem eine Datei angelegt")
			}
		})
	}
}

// Die Fehler, die aus den Nähten kommen.
func TestBogenfehlerAusDenNaehten(t *testing.T) {
	faelle := map[string]func(t *testing.T) *service.Bogenleger{
		"der Bestand antwortet nicht": func(t *testing.T) *service.Bogenleger {
			t.Helper()

			bestand := karten("Eine")
			bestand.kartenfehler = errors.New("die Datenbank ist zu")

			return service.NeuerBogenleger(
				kartenDouble{leserDouble: bestand}, planDouble{}, render.NeuerRasterer(nil))
		},
		// **Die Setliste ist die erste Naht** (#1291) — seit der Leger fragt,
		// ob es das Set überhaupt gibt.
		"die Setliste antwortet nicht": func(t *testing.T) *service.Bogenleger {
			t.Helper()

			bestand := karten("Eine")
			bestand.fehler = errors.New("die Datenbank ist zu")

			return service.NeuerBogenleger(
				kartenDouble{leserDouble: bestand}, planDouble{}, render.NeuerRasterer(nil))
		},
		"der Plan lässt sich nicht bauen": func(t *testing.T) *service.Bogenleger {
			t.Helper()

			return service.NeuerBogenleger(
				kartenDouble{leserDouble: karten("Eine")},
				planDouble{fehler: errors.New("kein Stil")}, render.NeuerRasterer(nil))
		},
		"der Plan ist nicht zeichenbar": func(t *testing.T) *service.Bogenleger {
			t.Helper()

			return service.NeuerBogenleger(
				kartenDouble{leserDouble: karten("Eine")},
				planDouble{breite: -1, hoehe: -1}, render.NeuerRasterer(nil))
		},
	}

	for name, bauen := range faelle {
		t.Run(name, func(t *testing.T) {
			_, err := bauen(t).Lege(t.Context(), 7,
				filepath.Join(t.TempDir(), "bogen.pdf"), service.StandardBogen(), domain.Kartenauswahl{})
			if err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **Ein Ziel, das sich nicht anlegen lässt, wird gemeldet.**
func TestEinUnbeschreibbaresBogenzielWirdGemeldet(t *testing.T) {
	faelle := map[string]func(t *testing.T) string{
		"ein Verzeichnis, das nicht entstehen kann": func(t *testing.T) string {
			t.Helper()

			imWeg := filepath.Join(t.TempDir(), "imweg")
			if err := os.WriteFile(imWeg, []byte("x"), 0o600); err != nil {
				t.Fatalf("anlegen: %v", err)
			}

			return filepath.Join(imWeg, "tiefer", "bogen.pdf")
		},
		"eine Datei, die ein Verzeichnis ist": func(t *testing.T) string {
			t.Helper()

			ziel := filepath.Join(t.TempDir(), "bogen.pdf")
			if err := os.MkdirAll(ziel, 0o750); err != nil {
				t.Fatalf("anlegen: %v", err)
			}

			return ziel
		},
	}

	for name, ziel := range faelle {
		t.Run(name, func(t *testing.T) {
			if _, err := bogenleger(t, 1).
				Lege(t.Context(), 7, ziel(t), service.StandardBogen(), domain.Kartenauswahl{}); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **Ein Set ohne Karten gibt eine Absage und keine unlesbare Datei.**
//
// # Diese Zusage stand bis zum 2026-08-30 andersherum
//
// „Ein Bogen ohne Karten ist ein Bogen ohne Seiten — und kein Fehler. Ein
// leeres Set ist eine Lage, kein Versehen." Der Satz erklärt das
// **Fachereignis** richtig und übersieht das **Artefakt**: eine PDF-Datei mit
// null Seiten öffnet kein Betrachter.
//
// Aufgefallen ist es, weil dieselbe Phase genau diesen Satz — „die ist
// verwirrender als ein Fehler" — benutzt hat, um ein **fehlendes** Set
// abzuweisen, und ihn für das **leere** stehen ließ. Zwei Wege, ein Artefakt,
// zwei Antworten (Befund 6 des Zweitblicks über Phase 4.4,
// #1291).
func TestEinSetOhneKartenGibtEineAbsage(t *testing.T) {
	ziel := filepath.Join(t.TempDir(), "leer.pdf")

	_, err := bogenleger(t, 0).
		Lege(t.Context(), 7, ziel, service.StandardBogen(), domain.Kartenauswahl{})

	// **`ErrValidation` und nicht `ErrNotFound`**: das Set gibt es, es hat nur
	// nichts zu legen. Ein Skript soll das von „dieses Set gibt es nicht"
	// unterscheiden können (§8).
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler: %v, erwartet ErrValidation", err)
	}

	// **Und es bleibt keine Datei zurück.** Eine unlesbare abzuweisen und sie
	// trotzdem zu schreiben wäre die schlechteste der drei Möglichkeiten.
	if _, fehler := os.Stat(ziel); !errors.Is(fehler, os.ErrNotExist) {
		t.Error("es wurde trotzdem eine Datei angelegt")
	}
}

// **Eine Karte ohne Namen wird im Befund trotzdem benannt.**
//
// „Karte 3" ist eine Auskunft; ein leerer Ort ist keine.
func TestEineKarteOhneNamenHeisstImBefundNachIhrerNummer(t *testing.T) {
	leger := service.NeuerBogenleger(
		kartenDouble{leserDouble: karten("")},
		planDouble{breite: 400, hoehe: 400}, render.NeuerRasterer(nil))

	lauf, err := leger.Lege(t.Context(), 7,
		filepath.Join(t.TempDir(), "bogen.pdf"), service.StandardBogen(), domain.Kartenauswahl{})
	if err != nil {
		t.Fatalf("legen: %v", err)
	}

	if len(lauf.Befunde) == 0 || lauf.Befunde[0].Ort != "Karte 1" {
		t.Errorf("Befunde: %v", lauf.Befunde)
	}
}

// Ein abgebrochener Lauf schreibt keinen Bogen.
func TestEinAbgebrochenerBogenlaufSchreibtNicht(t *testing.T) {
	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	if _, err := bogenleger(t, 1).Lege(ctx, 7, ziel, service.StandardBogen(), domain.Kartenauswahl{}); err == nil {
		t.Fatal("kein Fehler")
	}

	if _, err := os.Stat(ziel); !errors.Is(err, os.ErrNotExist) {
		t.Error("es wurde trotzdem geschrieben")
	}
}

// **Ein Set, das es nicht gibt, gibt keinen Bogen, sondern einen Fehler.**
//
// # Und das ist nicht dasselbe wie ein leeres Set
//
// **Seit dem 2026-08-30 geben beide eine Absage** — aber eine verschiedene: ein
// Set **ohne Karten** gibt `ErrValidation` (es gibt das Set, es hat nur
// nichts), eine Kennung **ohne Set** gibt `ErrNotFound`. Ein Skript soll das
// unterscheiden können (§8).
//
// **Hier stand, das leere Set gebe eine leere Datei, und die Entscheidung
// bleibe** — sie ist mit Befund 6 des Zweitblicks über Phase 4.4 umgedreht
// worden, und der Satz blieb einen Commit lang stehen (Befund 6 des
// Nachlaufs). Ein Kommentar, der den vorigen Vertrag erklärt, ist schlimmer
// als keiner.
//
// **Warum es zwei Prüfungen braucht**: `Karten` gibt für ein leeres Set und
// für ein Set, das es nicht gibt, dieselbe leere Liste — deshalb fragt der
// Leger vorher nach den Setköpfen.
//
// # Gefunden an der Kommandozeile
//
// Beim Legen des Bogens ohne Fenster
// (#1291):
// `mse bogen 999 /tmp/b.pdf` meldete Erfolg und schrieb 0 Seiten. Am Fenster
// fiel es nie auf, weil die Kennung dort aus einer Liste kommt — die Fläche
// ohne Liste ist die, die es zeigt.
func TestBogenFuerEinSetDasEsNichtGibt(t *testing.T) {
	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	_, err := bogenleger(t, 3).Lege(
		t.Context(), 999, ziel, service.StandardBogen(), domain.Kartenauswahl{})
	if err == nil {
		t.Fatal("kein Fehler")
	}

	// **Nicht gefunden und nicht Validierung**: die Kennung ist wohlgeformt,
	// sie zeigt nur auf nichts.
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler: %v, erwartet ErrNotFound", err)
	}

	if _, fehler := os.Stat(ziel); !errors.Is(fehler, os.ErrNotExist) {
		t.Error("es wurde trotzdem eine Datei angelegt")
	}
}
