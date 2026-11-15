//go:build darwin

// **Der Weg zum Schlüsselbund** — Werkstück 14 von Phase 4.12
// (#1351).
//
// # Warum ein eigener Bund und nicht der echte
//
// **§3, wörtlich**: ein Test, der nach draußen greift, tut es auch — und ist
// dann grün, **weil** er wirklich funktioniert hat. Hier wäre der Schaden
// größer als sonst: geschrieben würde in den Anmeldebund des Menschen, der
// diesen Lauf startet.
//
// `MSE_SCHLUESSELBUND` gibt es dafür seit Werkstück 7; `t.Setenv` setzt es für
// die Dauer eines Laufs.

package storage_test

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// eigenerBund legt einen Schlüsselbund an und lenkt die Anwendung dorthin um.
func eigenerBund(t *testing.T) {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "probe.keychain")

	//nolint:gosec // das Programm steht als Literal, der Ort liegt in t.TempDir()
	if aus, err := exec.CommandContext(t.Context(), "/usr/bin/security",
		"create-keychain", "-p", "leer", ort).CombinedOutput(); err != nil {
		t.Fatalf("den Probebund anlegen: %v — %s", err, aus)
	}

	t.Cleanup(func() {
		// **Ein eigener Kontext**: der des Laufs ist beim Aufräumen schon
		// abgebrochen, und dann räumte nichts ab.
		ctx, fertig := context.WithTimeout(context.Background(), 10*time.Second)
		defer fertig()

		//nolint:gosec,errcheck // dito
		_ = exec.CommandContext(ctx, "/usr/bin/security",
			"delete-keychain", ort).Run()
	})

	// **Aufgeschlossen, damit `security` nicht nach einem Kennwort fragt** —
	// sonst wartet es, bis die Frist es erschlägt.
	//
	//nolint:gosec,errcheck // dito
	_ = exec.CommandContext(t.Context(), "/usr/bin/security",
		"unlock-keychain", "-p", "leer", ort).Run()

	t.Setenv(schluesselbund.Bundort, ort)
}

// **Der Schlüssel geht in den Bund und kommt wieder heraus — und wieder weg.**
//
// Die drei Schritte sind der ganze Lebenszyklus: einschalten legt ab,
// ausschalten räumt weg, und dazwischen liest jedes Öffnen.
func TestDerBestandsschluesselGehtInDenBundUndWiederHeraus(t *testing.T) {
	eigenerBund(t)

	schluessel := storage.NeuerSchluessel()

	if err := storage.SchluesselSchreiben(t.Context(), schluessel); err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	gelesen, err := schluesselbund.Anmeldebund().
		Lies(t.Context(), storage.Schluesselbunddienst, storage.Schluesselkonto)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen != schluessel {
		t.Errorf("gelesen wurde %q, abgelegt war %q", gelesen, schluessel)
	}

	if fehler := storage.SchluesselLoeschen(t.Context()); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	// **Und danach ist er wirklich fort** — ohne diese Zeile wäre „gelöscht"
	// eine Behauptung.
	if _, fehler := schluesselbund.Anmeldebund().
		Lies(t.Context(), storage.Schluesselbunddienst,
			storage.Schluesselkonto); fehler == nil {
		t.Error("der Schlüssel liegt noch im Bund")
	}
}

// **Ein verschlüsselter Bestand geht mit dem Schlüssel aus dem Bund auf** —
// über den **echten** Weg, ohne die Naht zu ersetzen.
//
// Der Lauf daneben in `bestandsschluessel_intern_test.go` legt die Naht um und
// prüft damit die **Entscheidung**; dieser hier prüft den **Weg** durch
// `security` (§10: zwei grüne Seiten einer Naht sind kein Beleg für die Naht).
func TestOeffnenNimmtDenSchluesselAusDemEchtenBund(t *testing.T) {
	eigenerBund(t)

	ort := t.TempDir()
	klar := filepath.Join(ort, "klar.db")
	chiffre := filepath.Join(ort, "chiffre.db")

	bestandMit(t, klar, "Feuerball")

	schluessel := storage.NeuerSchluessel()

	quelle, err := storage.Oeffnen(t.Context(), klar)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := storage.Verschluesseln(
		t.Context(), quelle, chiffre, schluessel); fehler != nil {
		t.Fatalf("verschlüsseln: %v", fehler)
	}

	_ = quelle.Close() //nolint:errcheck // Test

	if fehler := storage.SchluesselSchreiben(t.Context(), schluessel); fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	t.Cleanup(func() {
		_ = storage.SchluesselLoeschen(context.Background()) //nolint:errcheck // Test
	})

	zurueck, err := storage.Oeffnen(t.Context(), chiffre)
	if err != nil {
		t.Fatalf("verschlüsselt öffnen: %v", err)
	}

	t.Cleanup(func() { _ = zurueck.Close() }) //nolint:errcheck // Test

	var wert string
	if fehler := zurueck.QueryRowContext(t.Context(),
		`SELECT wert FROM probe`).Scan(&wert); fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if wert != "Feuerball" {
		t.Errorf("Wert %q, erwartet war „Feuerball\"", wert)
	}
}
