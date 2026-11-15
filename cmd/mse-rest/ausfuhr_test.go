// **Die Wache aus [0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)**:
// dieser Dienst schreibt nur, was er nicht überschreibt.
//
// Sie steht hier und nicht in `internal/rest`, weil sie eine Aussage über die
// **Platte** ist — und jenes Paket ist ein Blatt.
package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein freier Ort geht durch — und ist danach reserviert.**
func TestEinFreierOrtGehtDurch(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "gibtsnicht.png")

	aufraeumen, err := reserviereOrt(ziel)
	if err != nil {
		t.Fatalf("ein freier Ort wurde abgewiesen: %v", err)
	}

	// **Die Reservierung ist eine Datei**, kein Vermerk im Kopf. Genau das
	// ist der Unterschied zu `os.Stat` (#1343).
	if _, err := os.Stat(ziel); err != nil {
		t.Errorf("nach der Reservierung liegt nichts am Ziel: %v", err)
	}

	aufraeumen()

	if _, err := os.Stat(ziel); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("das Aufräumen ließ etwas stehen: %v", err)
	}
}

// **Ein belegter Ort wird abgewiesen — als Konflikt.**
//
// `ErrConflict` und nicht `ErrValidation`: der Aufruf war richtig, der Ort ist
// belegt (§8). Ein Gegenüber kann das unterscheiden.
func TestEinBelegterOrtWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "liegt-schon-da.png")
	if err := os.WriteFile(ziel, []byte("x"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	_, fehler := reserviereOrt(ziel)
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("Fehler %v, erwartet war ErrConflict", fehler)
	}

	// **Die Meldung nennt den Ort und den Ausweg.** Eine Abweisung ohne
	// Ausweg ist eine Sackgasse — und ein Modell versucht es dann noch einmal
	// genauso.
	if !strings.Contains(fehler.Error(), ziel) || !strings.Contains(fehler.Error(), "Namen") {
		t.Errorf("die Meldung hilft nicht weiter: %v", fehler)
	}

	// **Und der Inhalt steht noch da.** Eine Abweisung, die das Vorhandene
	// leert, wäre schlimmer als ein Überschreiben.
	roh, err := os.ReadFile(ziel) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil || string(roh) != "x" {
		t.Errorf("der vorhandene Inhalt ist hin: %q, %v", roh, err)
	}
}

// **Auch ein Verzeichnis am Ziel ist belegt.**
//
// Es überschriebe nichts — es ließe das Schreiben scheitern, und zwar mit
// einer Meldung über Dateisystemrechte statt über den Ort.
func TestEinVerzeichnisAmZielIstBelegt(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "verzeichnis")
	if err := os.Mkdir(ziel, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if _, fehler := reserviereOrt(ziel); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("ein Verzeichnis am Ziel ging durch: %v", fehler)
	}
}

// **Ein unanlegbarer Ort ist ein Eingabefehler**, kein Konflikt.
//
// Ein Pfad durch ein Verzeichnis, das es nicht gibt: der Aufruf ist falsch,
// nicht der Ort belegt (§8).
func TestEinUnanlegbarerOrtIstEinEingabefehler(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "gibt-es-nicht", "datei.png")

	_, fehler := reserviereOrt(ziel)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Fehler %v, erwartet war ErrValidation", fehler)
	}

	if !strings.Contains(fehler.Error(), ziel) {
		t.Errorf("die Meldung nennt den Ort nicht: %v", fehler)
	}
}

// **Zwei gleichzeitige Aufrufe auf dasselbe Ziel: genau einer gewinnt** —
// die Abnahme aus #1343.
//
// # Warum das der Test ist, den `os.Stat` nicht bestehen konnte
//
// Die alte Wache sah nach und schrieb danach. Zwischen beidem lag ein
// Zeitraum, in dem ein zweiter Aufrufer dieselbe Antwort bekam: **frei**.
// Beide gingen durch, und der spätere ersetzte den früheren.
//
// Dieser Test gibt beiden Seiten dieselbe Startlinie über eine
// `sync.WaitGroup` und zählt danach ab. **Er ist nicht beweisend, sondern
// belastend**: er kann den Wettlauf nicht erzwingen, aber er macht ihn
// wahrscheinlich — und über `-race` und viele Durchläufe fällt eine Regression
// auf. Der eigentliche Beleg ist `O_EXCL` selbst, das der Kern entscheidet.
//
// **Die Gegenprobe ist gelaufen** (2026-08-31, Apple M5, `-race`): mit der
// alten Gestalt — `os.Stat`, dann anlegen — gewinnen **im ersten Durchgang
// beide Seiten**. Nicht in einem seltenen, sondern sofort.
func TestZweiGleichzeitigeAufrufeAufDasselbeZiel(t *testing.T) {
	t.Parallel()

	const versuche = 200

	for durchgang := range versuche {
		ziel := filepath.Join(t.TempDir(), "umkaempft.png")

		var (
			los           sync.WaitGroup
			fertig        sync.WaitGroup
			gewonnen      [2]bool
			fehler        [2]error
			schreibfehler [2]error
		)

		los.Add(1)
		fertig.Add(2)

		for seite := range 2 {
			go func() {
				defer fertig.Done()

				los.Wait()

				aufraeumen, err := reserviereOrt(ziel)
				if err != nil {
					fehler[seite] = err

					return
				}

				gewonnen[seite] = true

				// **Der Gewinner füllt seine Reservierung.** Erst damit
				// prüft der Test, wessen Datei am Ende dasteht.
				_ = aufraeumen
				schreibfehler[seite] = os.WriteFile(
					ziel, []byte{byte('A' + seite)}, 0o600)
			}()
		}

		los.Done()
		fertig.Wait()

		sieger := 0

		for seite := range 2 {
			if !gewonnen[seite] {
				continue
			}

			sieger++

			if schreibfehler[seite] != nil {
				t.Fatalf("Durchgang %d: der Gewinner konnte nicht schreiben: %v",
					durchgang, schreibfehler[seite])
			}
		}

		if sieger != 1 {
			t.Fatalf("Durchgang %d: %d Sieger, erwartet war genau einer (%v)",
				durchgang, sieger, fehler)
		}

		for seite := range 2 {
			if gewonnen[seite] {
				continue
			}

			if !errors.Is(fehler[seite], domain.ErrConflict) {
				t.Fatalf("Durchgang %d: der Verlierer bekam %v statt ErrConflict",
					durchgang, fehler[seite])
			}
		}

		// **Und die Datei ist die des Gewinners.**
		roh, err := os.ReadFile(ziel) // #nosec G304 -- Testpfad aus t.TempDir()
		if err != nil {
			t.Fatalf("Durchgang %d: die Datei fehlt: %v", durchgang, err)
		}

		erwartet := "A"
		if gewonnen[1] {
			erwartet = "B"
		}

		if string(roh) != erwartet {
			t.Fatalf("Durchgang %d: am Ziel steht %q, der Gewinner schrieb %q",
				durchgang, roh, erwartet)
		}
	}
}

// **Und die drei Wege gehen wirklich durch die Nähte.**
//
// # Was dieser Test hält
//
// Dass `ausfuhr` die Anwendungsfälle **erreicht** — nicht, was sie tun. Der
// Arbeitsablauf über die ganze Fläche steht in `internal/flaechen`; hier geht
// es um die Verdrahtung, die sonst niemand ausführt.
func TestDieDreiWegeErreichenIhreNaht(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	db, fehler := storage.Oeffnen(probekontext(t), filepath.Join(zuhause, "bestand.db"))
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	if err := storage.Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	aus := ausfuhr{db: db}
	ziel := t.TempDir()

	// **Einfahren eines Pfads, den es nicht gibt** — der Weg wird betreten,
	// und der Fehler kommt aus dem Fachkern, nicht aus der Verdrahtung.
	if _, _, _, err := aus.Einfahren(probekontext(t),
		filepath.Join(ziel, "gibtsnicht.mse-set")); err == nil {
		t.Error("ein Pfad, den es nicht gibt, wurde eingefahren")
	}

	// **Sichern und Zeichnen an einen belegten Ort** — die Wache greift, bevor
	// irgendetwas geschrieben wird.
	belegt := filepath.Join(ziel, "belegt")
	if err := os.WriteFile(belegt, []byte("x"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := aus.Sichern(probekontext(t), 1, belegt); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("Sichern an einen belegten Ort gab %v", err)
	}

	if _, _, _, err := aus.Zeichnen(probekontext(t), 1, belegt, ""); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("Zeichnen an einen belegten Ort gab %v", err)
	}

	// **Und ein Set, das es nicht gibt, an einen freien Ort.**
	if _, err := aus.Sichern(probekontext(t), 4711,
		filepath.Join(ziel, "frei.mse-set")); err == nil {
		t.Error("ein Set, das es nicht gibt, wurde gesichert")
	}

	if _, _, _, err := aus.Zeichnen(probekontext(t), 4711,
		filepath.Join(ziel, "frei.png"), ""); err == nil {
		t.Error("eine Karte, die es nicht gibt, wurde gezeichnet")
	}
}

// **Ohne Paketverzeichnis kommen die Nähte nicht zusammen.**
//
// Der Fehlerweg der Verdrahtung selbst — er wird sonst von keinem Test
// betreten.
func TestOhneZuhauseKommenDieNaehteNichtZusammen(t *testing.T) {
	t.Setenv("HOME", "")

	if _, fehler := app.Setschreibnaht(nil); fehler == nil {
		t.Error("ohne Zuhause kam die Setschreibnaht zusammen")
	}

	aus := ausfuhr{db: nil}
	ziel := filepath.Join(t.TempDir(), "frei")

	if _, err := aus.Sichern(context.Background(), 1, ziel); err == nil { //nolint:usetesting // probekontext(t) endet mit dem Test; hier soll er nicht abgebrochen sein
		t.Error("ohne Zuhause wurde gesichert")
	}

	if _, _, _, err := aus.Zeichnen(context.Background(), 1, ziel, ""); err == nil { //nolint:usetesting // wie oben
		t.Error("ohne Zuhause wurde gezeichnet")
	}
}
