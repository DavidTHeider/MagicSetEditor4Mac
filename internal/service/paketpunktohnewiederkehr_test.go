// **Nach dem Punkt ohne Wiederkehr hält kein Abbruch mehr auf** — der Nachlauf
// über Phase 3.8, Befund 1
// (#1169).
//
// # Der Befund
//
// `Installieren` kopiert das Paket, **bevor** der Katalog mit dem weiterhin
// abbrechbaren `ctx` abgeglichen wird. `Entfernen` löscht es, bevor
// `PaketEntfernen(ctx, …)` läuft. `Uebernehmen` kopiert Hunderte, bevor
// `Vermerken(ctx, …)` bucht.
//
// Bricht der Kontext dazwischen ab, meldet der Vorgang einen **Fehler**,
// obwohl die Platte längst geändert ist: ein installiertes Paket fehlt im
// Katalog, ein gelöschtes steht noch darin.
//
// # Die Regel dahinter
//
// §7, Regel 3 — **geschrieben ist geschrieben**. Was nach dem Schreiben kommt,
// ist Bestenfalls-Arbeit und darf den geglückten Zug nicht zurücknehmen. Für
// den Bestandsweg ist das seit Phase 3.8b so; für den Paketverwalter war es
// das nicht.
package service_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// abgebrochen gibt einen Kontext, der schon vorbei ist.
func abgebrochen(t *testing.T) context.Context {
	t.Helper()

	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	return ctx
}

func TestEinAbbruchNachDemKopierenHaeltDieInstallationNichtAuf(t *testing.T) {
	t.Parallel()

	wurzel, katalog := t.TempDir(), &ablage{}
	quelle := t.TempDir()
	paket(t, quelle, "neu.mse-style", "mse version: 2.0.0\n")
	verwalter := service.NeuerPaketverwalter(katalog, wurzel, wurzel)

	gelesen, fehler := verwalter.Installieren(abgebrochen(t),
		filepath.Join(quelle, "neu.mse-style"))
	if fehler != nil {
		t.Fatalf("Installieren: %v — der Abbruch kam nach dem Kopieren", fehler)
	}

	if gelesen.Pfad == "" {
		t.Error("das Paket hat keinen Ort")
	}

	if _, err := os.Stat(filepath.Join(wurzel, "neu.mse-style")); err != nil {
		t.Errorf("das Paket liegt nicht im Verzeichnis: %v", err)
	}

	if len(katalog.pakete) == 0 {
		t.Error("der Katalog kennt es nicht — genau der halbe Stand, um den es geht")
	}
}

func TestEinAbbruchNachDemLoeschenHaeltDasEntfernenNichtAuf(t *testing.T) {
	t.Parallel()

	wurzel, katalog := t.TempDir(), &ablage{}
	quelle := t.TempDir()
	paket(t, quelle, "neu.mse-style", "mse version: 2.0.0\n")
	verwalter := service.NeuerPaketverwalter(katalog, wurzel, wurzel)

	if _, fehler := verwalter.Installieren(t.Context(),
		filepath.Join(quelle, "neu.mse-style")); fehler != nil {
		t.Fatalf("vorbereiten: %v", fehler)
	}

	if fehler := verwalter.Entfernen(abgebrochen(t), "mse-style", "neu", ""); fehler != nil {
		t.Fatalf("Entfernen: %v — der Abbruch kam nach dem Löschen", fehler)
	}

	if _, err := os.Stat(filepath.Join(wurzel, "neu.mse-style")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("das Paket liegt noch da: %v", err)
	}

	if len(katalog.entfernt) == 0 {
		t.Error("der Katalog führt es noch — genau der halbe Stand, um den es geht")
	}
}

// **Und dasselbe für die Übernahme aus einer MSE2-Installation.**
//
// Sie kopiert Hunderte, bevor sie bucht; ein Abbruch dazwischen ließe sie
// sämtlich außerhalb des Katalogs zurück.
func TestEinAbbruchNachDemUebernehmenHaeltEsNichtAuf(t *testing.T) {
	t.Parallel()

	wurzel, katalog := t.TempDir(), &ablage{}
	quelle := t.TempDir()
	paket(t, quelle, "fremd.mse-style", "mse version: 2.0.0\n")
	verwalter := service.NeuerPaketverwalter(katalog, wurzel, wurzel)

	ergebnis, fehler := verwalter.Uebernehmen(abgebrochen(t), quelle)
	if fehler != nil {
		t.Fatalf("Uebernehmen: %v — der Abbruch kam nach dem Kopieren", fehler)
	}

	if _, err := os.Stat(filepath.Join(wurzel, "fremd.mse-style")); err != nil {
		t.Errorf("das Paket liegt nicht im Verzeichnis: %v", err)
	}

	if len(katalog.pakete) == 0 {
		t.Error("der Katalog kennt es nicht — genau der halbe Stand, um den es geht")
	}

	if len(ergebnis.Uebernommen) == 0 {
		t.Error("das Ergebnis nennt kein übernommenes Paket")
	}
}
