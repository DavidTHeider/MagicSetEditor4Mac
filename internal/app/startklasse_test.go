//go:build darwin

// **Was beim Start schiefging, hat einen Namen** — Werkstück 3 von Phase 4.13
// (#1352).
//
// # Der Befund
//
// **Am 2026-09-07 an der laufenden Anwendung gemessen**, vier Fälle:
//
//	die Datei fehlt   → der Start glückt, der Bestand ist leer
//	kein Bestand      → "journal_mode: sqlite3: file is not a database"
//	aus der Zukunft   → ein vollständiger deutscher Satz — ohne nächsten Schritt
//	zweiter Start     → glückt, und das ist richtig ([0017](../../docs/entscheidungen/0017-die-prozesse-bemerken-einander.md))
//
// **Der zweite ist ein roher SQLite-Text im Fenster** (§8), der dritte sagt
// nicht, was zu tun ist (Werkstück 2 dieser Phase). Beide bekommen einen
// Schlüssel, und der Katalog sagt den Satz.

package app_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine fehlende Datei ist kein Fehler** — sie ist der erste Start.
func TestEineFehlendeBestandsdateiIstDerErsteStart(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatalf("der erste Start scheiterte: %v", fehler)
	}

	t.Cleanup(func() { _ = anwendung.Beenden() }) //nolint:errcheck // Aufräumen

	befund := anwendung.Startbefund()
	if !befund.Offen {
		t.Errorf("der Bestand ist nicht offen: %+v", befund)
	}

	if befund.Schluessel != "" {
		t.Errorf("ein geglückter Start trägt einen Schlüssel: %+v", befund)
	}
}

// **Eine Datei, die kein Bestand ist, wird benannt — nicht roh gemeldet.**
func TestEineHalbeDateiTraegtIhrenSchluessel(t *testing.T) {
	t.Parallel()

	ort := filepath.Join(t.TempDir(), "bestand.db")
	if err := os.WriteFile(ort, []byte("kein SQLite"), 0o600); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	anwendung := app.NeueApp(ort)

	if fehler := anwendung.Starten(t.Context()); fehler == nil {
		t.Fatal("eine Datei, die kein Bestand ist, wurde geöffnet")
	}

	befund := anwendung.Startbefund()
	if befund.Schluessel != string(domain.SchluesselBestandUnbrauchbar) {
		t.Errorf("Schlüssel %q, erwartet war %q",
			befund.Schluessel, domain.SchluesselBestandUnbrauchbar)
	}
}

// **Und eine aus der Zukunft ebenso.**
func TestEinBestandAusDerZukunftTraegtSeinenSchluessel(t *testing.T) {
	t.Parallel()

	ort := filepath.Join(t.TempDir(), "bestand.db")

	// **Erst ein gültiger Bestand, dann die Zeile aus der Zukunft.**
	vorbereitet := app.NeueApp(ort)
	if fehler := vorbereitet.Starten(t.Context()); fehler != nil {
		t.Fatalf("vorbereiten: %v", fehler)
	}

	if fehler := vorbereitet.Beenden(); fehler != nil {
		t.Fatalf("vorbereiten: %v", fehler)
	}

	zukunftEintragen(t, ort)

	anwendung := app.NeueApp(ort)

	if fehler := anwendung.Starten(t.Context()); fehler == nil {
		t.Fatal("ein Bestand aus der Zukunft wurde geöffnet")
	}

	befund := anwendung.Startbefund()
	if befund.Schluessel != string(domain.SchluesselBestandNeuereFassung) {
		t.Errorf("Schlüssel %q, erwartet war %q",
			befund.Schluessel, domain.SchluesselBestandNeuereFassung)
	}
}

// **Ein zweiter Start neben einem laufenden geht** — und das ist eine Zusage
// und kein Mangel ([0017](../../docs/entscheidungen/0017-die-prozesse-bemerken-einander.md)):
// MCP-Server, Dienst und Fenster arbeiten an derselben Datei, und sie bemerken
// einander über `PRAGMA data_version`.
func TestEinZweiterStartNebenEinemLaufendenGeht(t *testing.T) {
	t.Parallel()

	ort := filepath.Join(t.TempDir(), "bestand.db")

	erste := app.NeueApp(ort)
	if fehler := erste.Starten(t.Context()); fehler != nil {
		t.Fatalf("der erste Start scheiterte: %v", fehler)
	}

	t.Cleanup(func() { _ = erste.Beenden() }) //nolint:errcheck // Aufräumen

	zweite := app.NeueApp(ort)
	if fehler := zweite.Starten(t.Context()); fehler != nil {
		t.Fatalf("der zweite Start scheiterte: %v", fehler)
	}

	t.Cleanup(func() { _ = zweite.Beenden() }) //nolint:errcheck // Aufräumen

	if !zweite.Startbefund().Offen {
		t.Errorf("der zweite Bestand ist nicht offen: %+v", zweite.Startbefund())
	}
}

// **Und ein Bestand ohne Schlüssel wird als solcher benannt.**
//
// **Der eigene Zweig ist der Punkt**: ohne ihn stünde „diese Datei ist kein
// Bestand" über einer Datei, die in Ordnung ist — und jemand ersetzte sie. Der
// Schlüssel lässt sich nicht wiederherstellen; das Ersetzen machte den Verlust
// endgültig (Werkstück 14 von Phase 4.12).
func TestEinBestandOhneSchluesselTraegtSeinenSchluessel(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	if geschaltet := anwendung.VerschluesselungUmschalten(
		t.Context(), true); geschaltet.Fehler != nil {
		t.Fatalf("verschlüsseln: %+v", geschaltet.Fehler)
	}

	if fehler := anwendung.Beenden(); fehler != nil {
		t.Fatalf("beenden: %v", fehler)
	}

	// **Der Schlüssel kommt aus dem Bund weg** — genau der Fall, den es zu
	// benennen gilt: die Datei ist da und in Ordnung, nur unlesbar.
	//
	//nolint:gosec,errcheck // ein Probebund in t.TempDir(); geht es schief,
	// bleibt der Schlüssel liegen und der Lauf darunter wird rot
	_ = exec.CommandContext(t.Context(), "/usr/bin/security",
		"delete-generic-password", "-s", storage.Schluesselbunddienst,
		"-a", storage.Schluesselkonto, os.Getenv(schluesselbund.Bundort)).Run()

	zweite := app.NeueApp(bestand)

	if fehler := zweite.Starten(t.Context()); fehler == nil {
		t.Fatal("ein Bestand ohne Schlüssel wurde geöffnet")
	}

	befund := zweite.Startbefund()
	if befund.Schluessel != string(domain.SchluesselSchluesselFehlt) {
		t.Errorf("Schlüssel %q, erwartet war %q",
			befund.Schluessel, domain.SchluesselSchluesselFehlt)
	}
}
