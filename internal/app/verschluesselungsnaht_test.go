//go:build darwin

// **Die Verschlüsselung über die Naht ohne Fenster** — Werkstück 14 von Phase
// 4.12 (#1351).
//
// # Warum das hier steht und nicht nur im Flächenlauf
//
// **Weil der Flächenlauf gebaute Programme startet** und damit nichts zur
// Abdeckung dieses Pakets beiträgt. Er belegt, dass die Fähigkeit an jeder
// Fläche **da** ist; hier wird der Weg selbst betreten.
//
// # Und warum ein eigener Schlüsselbund
//
// **§3**: dieser Lauf legt einen Schlüssel ab. Ohne die Umlenkung landete er
// im Anmeldebund des Menschen, der ihn startet — und der Lauf wäre grün,
// **weil** er wirklich funktioniert hat.

package app_test

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// eigenerBund lenkt den Schlüsselbund für die Dauer eines Laufs um.
func eigenerBund(t *testing.T) {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "probe.keychain")

	//nolint:gosec // das Programm steht als Literal, der Ort liegt in t.TempDir()
	if aus, err := exec.CommandContext(t.Context(), "/usr/bin/security",
		"create-keychain", "-p", "leer", ort).CombinedOutput(); err != nil {
		t.Fatalf("den Probebund anlegen: %v — %s", err, aus)
	}

	t.Cleanup(func() {
		ctx, fertig := context.WithTimeout(context.Background(), 10*time.Second)
		defer fertig()

		//nolint:gosec,errcheck // ein Probebund in t.TempDir()
		_ = exec.CommandContext(ctx, "/usr/bin/security",
			"delete-keychain", ort).Run()
	})

	//nolint:gosec,errcheck // aufschließen, damit `security` nicht wartet
	_ = exec.CommandContext(t.Context(), "/usr/bin/security",
		"unlock-keychain", "-p", "leer", ort).Run()

	t.Setenv(schluesselbund.Bundort, ort)
}

// **Der ganze Weg: an, nachsehen, aus.**
//
// **Und der Bestand lebt danach weiter** — das ist der Teil, den ein Lauf über
// die Chiffre allein nicht zeigt: nach dem Tausch hält die Naht eine **neue**
// Verbindung, und wer die alte behielte, griffe auf eine Datei, die es nicht
// mehr gibt.
func TestDieVerschluesselungGehtUeberDieNahtHinUndZurueck(t *testing.T) {
	eigenerBund(t)

	db, ort := offenerBestand(t)

	naht := app.BestandsnahtUeber(&db, ort)

	if an, fehler := naht.Verschluesselungsstand(); fehler != nil || an {
		t.Fatalf("Stand %v (%v), erwartet war Klartext", an, fehler)
	}

	an, fehler := naht.VerschluesselungUmschalten(probekontext(t), true)
	if fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	if !an {
		t.Fatal("nach dem Einschalten gilt der Bestand als Klartext")
	}

	// **Die Datei trägt die Klartextmarke nicht mehr** — der Beleg an der
	// Platte und nicht an der Auskunft über sie.
	if ist, pruefefehler := storage.IstVerschluesselt(ort); pruefefehler != nil || !ist {
		t.Errorf("die Datei sagt %v (%v)", ist, pruefefehler)
	}

	// **Und der Bestand ist weiter zu lesen**, über die **neue** Verbindung.
	var zahl int
	if lesefehler := db.QueryRowContext(probekontext(t),
		`SELECT count(*) FROM sets`).Scan(&zahl); lesefehler != nil {
		t.Fatalf("nach dem Einschalten lesen: %v", lesefehler)
	}

	if _, fehler := naht.VerschluesselungUmschalten(probekontext(t), false); fehler != nil {
		t.Fatalf("ausschalten: %v", fehler)
	}

	if zurueck, pruefefehler := storage.IstVerschluesselt(ort); pruefefehler != nil || zurueck {
		t.Errorf("nach dem Ausschalten sagt die Datei %v (%v)", zurueck, pruefefehler)
	}

	if lesefehler := db.QueryRowContext(probekontext(t),
		`SELECT count(*) FROM sets`).Scan(&zahl); lesefehler != nil {
		t.Fatalf("nach dem Ausschalten lesen: %v", lesefehler)
	}
}

// **Zweimal einschalten wird abgewiesen** — der Weg durch die Naht, nicht nur
// durch den Dienst.
func TestZweimalEinschaltenGehtAuchUeberDieNahtNicht(t *testing.T) {
	eigenerBund(t)

	db, ort := offenerBestand(t)

	naht := app.BestandsnahtUeber(&db, ort)

	if _, fehler := naht.VerschluesselungUmschalten(probekontext(t), true); fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	t.Cleanup(func() {
		// **Wieder ausschalten**, damit der Schlüssel nicht mit dem Bund
		// verschwindet, während die Datei ihn noch braucht.
		_, _ = naht.VerschluesselungUmschalten( //nolint:errcheck // Aufräumen
			context.Background(), false)
	})

	if _, fehler := naht.VerschluesselungUmschalten(probekontext(t), true); fehler == nil {
		t.Error("zweimal eingeschaltet")
	}
}
