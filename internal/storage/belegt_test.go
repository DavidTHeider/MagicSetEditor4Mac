package storage_test

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein belegter Bestand meldet sich als belegt, nicht als SQLite**
// (#886).
//
// # Warum das ein eigener Fehler ist
//
// Gemessen am 2026-08-23 an der Kommandozeile gegen ein offenes Fenster: hält
// das Fenster eine Schreibsperre **länger** als `busy_timeout` (5 s), endet
// `mse feld …` mit
//
//	Migration beginnen: database is locked (5) (SQLITE_BUSY)
//
// Das ist eine rohe SQLite-Meldung an der Oberfläche, und §8 verbietet sie
// ausdrücklich. Sie sagt außerdem das Falsche: es ist keine Migration kaputt,
// und der Aufrufer muss nichts reparieren — er muss **warten**.
//
// # Wo sie auftaucht, und wo dieser Test hinsieht
//
// An **zwei** Stellen: in der Migrationstransaktion beim Öffnen (so entstand
// die Meldung oben) und beim ersten Schreibversuch danach. Nachgemessen am
// selben Tag: bei bereits migriertem Bestand kommt `Oeffnen` **durch** und der
// Fehler fällt erst am Schreiben an.
//
// Geprüft wird deshalb der Schreibweg — er lässt sich hier verlässlich
// herstellen. Die Migrationsstelle ist mit derselben Abbildung versorgt
// (`migrator.go` reicht sie seit dem 2026-08-23 durch `uebersetze`), und
// **das** ist der Satz daneben: dieser Test belegt sie nicht.
//
// # Warum er fünf Sekunden dauert
//
// So lange wartet `busy_timeout`, und den herabzusetzen hieße, etwas anderes
// zu prüfen als das, was im Betrieb läuft.
//
// # Warum nicht [domain.ErrConflict]
//
// Der heißt in diesem Programm „dein Stand ist überholt", und die Oberfläche
// **lädt darauf nach** (§7, Regel 3). Hier ist kein Stand überholt; nachladen
// liefe in dieselbe Sperre. Der richtige Rat ist ein anderer, also ist es ein
// anderer Fehler.
func TestEinBelegterBestandMeldetSichAlsBelegt(t *testing.T) {
	t.Parallel()

	ort := filepath.Join(t.TempDir(), "bestand.db")

	// Ein erster Zugriff legt den Bestand an und migriert ihn.
	erster, fehler := storage.Oeffnen(probekontext(t), ort)
	if fehler != nil {
		t.Fatalf("erstes Öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = erster.Close() })

	// **Die Sperre wird gehalten, nicht nachgestellt.** Ein selbst gebauter
	// `sqlite.Error` bewiese nur, dass die Abbildung eine Konstante trifft —
	// nicht, dass dieser Weg sie erreicht (§10).
	sperre, fehler := erster.BeginTx(probekontext(t), nil)
	if fehler != nil {
		t.Fatalf("Sperre nehmen: %v", fehler)
	}

	if _, schreibfehler := sperre.ExecContext(probekontext(t),
		"CREATE TABLE belegt (x INTEGER)"); schreibfehler != nil {
		t.Fatalf("schreiben: %v", schreibfehler)
	}

	t.Cleanup(func() { _ = sperre.Rollback() })

	// Der zweite Zugriff — wie ihn die Kommandozeile neben dem Fenster nimmt.
	ctx, ende := context.WithTimeout(probekontext(t), 30*time.Second)
	t.Cleanup(ende)

	zweiter, fehler := storage.Oeffnen(ctx, ort)
	if fehler != nil {
		pruefe(t, fehler)

		return
	}

	t.Cleanup(func() { _ = zweiter.Close() })

	_, fehler = storage.NeuerSetSpeicher(zweiter).SetAnlegen(ctx,
		&domain.Set{Name: "zweiter", Game: "magic"}, nil)
	if fehler == nil {
		t.Fatal("der zweite Zugriff schrieb, obwohl der erste sperrt")
	}

	pruefe(t, fehler)
}

// pruefe hält die Zusage an **einer** Stelle — der Fehler darf an zwei
// entstehen, und beide sollen dasselbe melden.
func pruefe(t *testing.T, fehler error) {
	t.Helper()

	if !errors.Is(fehler, domain.ErrBelegt) {
		t.Errorf("Fehler %v ist kein %v", fehler, domain.ErrBelegt)
	}

	// **Und die rohe Meldung bleibt drin, aber nicht vorn.** Wer sie zum
	// Nachsehen braucht, findet sie; wer die Meldung zeigt, zeigt den Satz
	// aus der Domäne (§8).
	if !strings.HasPrefix(fehler.Error(), domain.ErrBelegt.Error()) &&
		!strings.Contains(fehler.Error(), domain.ErrBelegt.Error()) {
		t.Errorf("die Meldung nennt den Grund nicht: %q", fehler)
	}
}
