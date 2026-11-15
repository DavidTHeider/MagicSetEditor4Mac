package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bulkprobe ist ein Doppel für die lokale Sammeldatei.
type bulkprobe struct {
	karte   *scryfall.Karte
	befunde []domain.Befund
	fehler  error
	gefragt string
}

func (b *bulkprobe) KarteNachName(_ context.Context, name string) (
	*scryfall.Karte, []domain.Befund, error,
) {
	b.gefragt = name

	return b.karte, b.befunde, b.fehler
}

// netzprobe ist ein Doppel für den Klienten.
type netzprobe struct {
	karte  *scryfall.Karte
	alter  time.Duration
	fehler error
	rufe   int
}

func (n *netzprobe) KarteNachName(context.Context, string) (
	*scryfall.Karte, time.Duration, []domain.Befund, error,
) {
	n.rufe++

	return n.karte, n.alter, nil, n.fehler
}

// **Ohne Netz arbeitet die Anwendung weiter** — §16, Regel 5.
//
// # Warum das ein Rückfall ist und kein Ersatz
//
// Die Sammeldatei ist **12 bis 24 Stunden alt** — das sagt Scryfall selbst:
// „Bulk data is only collected once every 12-24 hours." Sie zuerst zu fragen
// hieße, freiwillig einen alten Stand zu nehmen. Sie **danach** zu fragen
// heißt: lieber ein alter Stand als gar keiner.
func TestDerBulkrueckfall(t *testing.T) {
	t.Parallel()

	t.Run("das Netz gewinnt, solange es antwortet", func(t *testing.T) {
		t.Parallel()

		netz := &netzprobe{karte: &scryfall.Karte{Name: "Aus dem Netz"}}
		lokal := &bulkprobe{karte: &scryfall.Karte{Name: "Aus der Datei"}}

		karte, _, _, fehler := service.MitBulkrueckfall(netz, lokal).KarteNachName(
			t.Context(), "Prüfstein")
		if fehler != nil {
			t.Fatalf("KarteNachName: %v", fehler)
		}

		if karte.Name != "Aus dem Netz" {
			t.Errorf("Karte = %q", karte.Name)
		}

		// **Die Datei wird gar nicht erst gelesen** — sie liegt entpackt bei
		// 400 MB, und ein Durchlauf kostet Sekunden.
		if lokal.gefragt != "" {
			t.Errorf("die Datei wurde gefragt: %q", lokal.gefragt)
		}
	})

	t.Run("die Datei springt ein, wenn das Netz ausfällt", func(t *testing.T) {
		t.Parallel()

		netz := &netzprobe{fehler: errors.New("kein Netz")}
		lokal := &bulkprobe{karte: &scryfall.Karte{Name: "Aus der Datei"}}

		karte, alter, befunde, fehler := service.MitBulkrueckfall(netz, lokal).KarteNachName(
			t.Context(), "Prüfstein")
		if fehler != nil {
			t.Fatalf("KarteNachName: %v", fehler)
		}

		if karte.Name != "Aus der Datei" || lokal.gefragt != "Prüfstein" {
			t.Errorf("Karte = %+v, gefragt = %q", karte, lokal.gefragt)
		}

		// **Ein Befund sagt, dass der Stand aus der Datei kommt.** Ohne ihn
		// sähe eine Auskunft aus der Datei aus wie eine frische — und sie ist
		// bis zu einen Tag alt.
		if len(befunde) == 0 {
			t.Error("der Rückfall blieb ohne Befund")
		}

		// **Und das Alter ist nicht null.** Null hieße „gerade geholt", und
		// das wäre die falsche Auskunft.
		if alter == 0 {
			t.Error("Alter = 0 — das hieße, der Stand sei frisch")
		}
	})

	t.Run("ein fachliches Nein geht nicht in die Datei", func(t *testing.T) {
		t.Parallel()

		// **Eine Karte, die es nicht gibt, gibt es auch in der Datei nicht.**
		// Dort nachzusehen kostete Sekunden für eine Antwort, die feststeht —
		// und schlimmer: eine ältere Datei könnte eine Karte führen, die
		// Scryfall inzwischen zurückgezogen hat.
		netz := &netzprobe{fehler: domain.ErrNotFound}
		lokal := &bulkprobe{karte: &scryfall.Karte{Name: "Aus der Datei"}}

		if _, _, _, fehler := service.MitBulkrueckfall(netz, lokal).KarteNachName(
			t.Context(), "Gibtesnicht"); !errors.Is(fehler, domain.ErrNotFound) {
			t.Fatalf("Fehler = %v", fehler)
		}

		if lokal.gefragt != "" {
			t.Errorf("die Datei wurde gefragt: %q", lokal.gefragt)
		}
	})

	t.Run("ohne Datei kommt der Netzfehler durch", func(t *testing.T) {
		t.Parallel()

		netz := &netzprobe{fehler: errors.New("kein Netz")}
		lokal := &bulkprobe{fehler: errors.New("keine Datei")}

		_, _, _, fehler := service.MitBulkrueckfall(netz, lokal).KarteNachName(
			t.Context(), "Prüfstein")
		if fehler == nil {
			t.Fatal("beides scheiterte, und es kam kein Fehler")
		}

		// **Der Netzfehler steht vorn**, denn er ist der eigentliche Grund;
		// „keine Datei" ist die Folge davon, dass es keinen Rückfall gibt.
		if !strings.Contains(fehler.Error(), "kein Netz") {
			t.Errorf("Fehler = %v — der Netzfehler fehlt darin", fehler)
		}
	})

	t.Run("ohne eingerichteten Rückfall bleibt es beim Netzfehler", func(t *testing.T) {
		t.Parallel()

		netz := &netzprobe{fehler: errors.New("kein Netz")}

		if _, _, _, fehler := service.MitBulkrueckfall(netz, nil).KarteNachName(
			t.Context(), "Prüfstein"); fehler == nil {
			t.Fatal("der Netzfehler kam nicht durch")
		}
	})
}

// **Ein 429 sagt etwas anderes als „kein Netz"** —
// eigener Befund E-1 und Teil von Befund 2 des Zweitblicks, 2026-08-24.
//
// # Der Befund
//
// Der Rückfall griff bei **jedem** Fehler außer `ErrNotFound` — also auch bei
// `domain.ErrGedrosselt`. Der Befund lautete dann „Scryfall war nicht
// erreichbar". Bei einem 429 war Scryfall **bestens** erreichbar; es hat
// gesagt „zu schnell". Wer die Meldung liest, sucht den Fehler bei seiner
// Verbindung statt bei der Drossel dieses Programms — und die 429 ist genau
// das Zeichen, dass sie nicht reicht.
func TestEin429SagtEtwasAnderesAlsKeinNetz(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		fehler   error
		enthaelt string
	}{
		"kein Netz": {
			errors.New("dial tcp: no route to host"), "nicht erreichbar",
		},
		"zu viele Anfragen": {
			domain.ErrGedrosselt, "zu viele Anfragen",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			netzquelle := &netzprobe{fehler: fall.fehler}
			lokal := &bulkprobe{karte: &scryfall.Karte{Name: "Aus der Datei"}}

			_, _, befunde, fehler := service.MitBulkrueckfall(
				netzquelle, lokal).KarteNachName(t.Context(), "Prüfstein")
			if fehler != nil {
				t.Fatalf("KarteNachName: %v", fehler)
			}

			if len(befunde) == 0 {
				t.Fatal("der Rückfall blieb ohne Befund")
			}

			if !strings.Contains(befunde[0].Meldung, fall.enthaelt) {
				t.Errorf("Meldung = %q, erwartet einen Hinweis auf %q",
					befunde[0].Meldung, fall.enthaelt)
			}
		})
	}
}
