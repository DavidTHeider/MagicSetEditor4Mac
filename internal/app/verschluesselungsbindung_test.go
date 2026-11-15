//go:build darwin

// **Die Verschlüsselung am Fenster** — Werkstück 14 von Phase 4.12
// (#1351).
//
// # Warum über die **gestartete** Anwendung und nicht über ein Doppel
//
// **Weil `Starten` die Naht baut.** Ein eingesetztes Doppel prüfte die
// Übersetzung und ließe offen, ob die Verdrahtung steht — und genau die ist
// hier neu: `a.bestandschiffre` wird bei jedem Start neu gesetzt, weil die
// Verbindung wechselt. Zwei grüne Seiten einer Naht sind kein Beleg für die
// Naht (§10).

package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// gestarteteAnwendung gibt eine Anwendung über einem eigenen Bestand.
func gestarteteAnwendung(t *testing.T) (anwendung *app.App, bestandsort string) {
	t.Helper()

	t.Setenv("HOME", t.TempDir())
	eigenerBund(t)

	bestand, err := app.Datenbankpfad()
	if err != nil {
		t.Fatalf("Datenbankpfad: %v", err)
	}

	if fehler := os.MkdirAll(filepath.Dir(bestand), 0o700); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	anwendung = app.NeueApp(bestand)
	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatalf("starten: %v", fehler)
	}

	return anwendung, bestand
}

// **Der Stand kommt zurück, das Umschalten wirkt, und beides sagt dasselbe.**
func TestDieBindungSchaltetDieVerschluesselungUndSagtDenStand(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	if antwort := anwendung.Verschluesselungsstand(t.Context()); antwort.Fehler != nil ||
		antwort.Verschluesselt {
		t.Fatalf("Stand %+v, erwartet war Klartext ohne Fehler", antwort)
	}

	an := anwendung.VerschluesselungUmschalten(t.Context(), true)
	if an.Fehler != nil {
		t.Fatalf("einschalten: %+v", an.Fehler)
	}

	// **Die Antwort trägt den neuen Stand selbst** (§7, Regel 3) — die Seite
	// muss nicht nachladen, und zwischen Antwort und Nachladen läge ein Stand,
	// den niemand kennt.
	if !an.Verschluesselt {
		t.Error("die Antwort meldet Klartext, obwohl eingeschaltet wurde")
	}

	// **Und die Platte sagt dasselbe** — der Beleg an der Datei und nicht an
	// der Auskunft über sie.
	if ist, fehler := storage.IstVerschluesselt(bestand); fehler != nil || !ist {
		t.Errorf("die Datei sagt %v (%v)", ist, fehler)
	}

	aus := anwendung.VerschluesselungUmschalten(t.Context(), false)
	if aus.Fehler != nil {
		t.Fatalf("ausschalten: %+v", aus.Fehler)
	}

	if aus.Verschluesselt {
		t.Error("die Antwort meldet verschlüsselt, obwohl ausgeschaltet wurde")
	}
}

// **Und ein zweites Einschalten kommt als Fehler an der Fläche an** — nicht
// als Absturz und nicht als stiller Erfolg.
func TestEinZweitesEinschaltenKommtAlsFehlerAmFensterAn(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	if erste := anwendung.VerschluesselungUmschalten(t.Context(), true); erste.Fehler != nil {
		t.Fatalf("einschalten: %+v", erste.Fehler)
	}

	t.Cleanup(func() {
		anwendung.VerschluesselungUmschalten(t.Context(), false) //nolint:errcheck // Aufräumen
	})

	zweite := anwendung.VerschluesselungUmschalten(t.Context(), true)

	if zweite.Fehler == nil {
		t.Fatal("zweimal eingeschaltet")
	}

	if zweite.Fehler.Code != app.CodeValidierung {
		t.Errorf("Code %q, erwartet war %q", zweite.Fehler.Code, app.CodeValidierung)
	}
}

// **Eine Anwendung, die nie gestartet wurde, sagt das — statt abzustürzen.**
//
// **Der Weg ist echt und kein Kunstgriff**: `Starten` baut die Naht, und ein
// Fenster, dessen Start scheiterte, steht genau so da. Die Wiederherstellung
// über eine Sicherung läuft in diesem Zustand ab (`Startbefund`), und die
// Einstellungsseite fragt trotzdem.
func TestOhneStartSagtDieBindungWasFehltStattAbzustuerzen(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	antwort := anwendung.Verschluesselungsstand(t.Context())

	if antwort.Fehler == nil {
		t.Fatal("ohne Start kam eine Auskunft ohne Fehler")
	}

	if antwort.Fehler.Code != app.CodeValidierung {
		t.Errorf("Code %q, erwartet war %q", antwort.Fehler.Code, app.CodeValidierung)
	}

	if antwort.Verschluesselt {
		t.Error("bei einem Fehler soll nichts behauptet werden")
	}
}
