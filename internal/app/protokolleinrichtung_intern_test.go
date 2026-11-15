package app

// **Die Weiche folgt den Einstellungen, und die Diagnose sagt, was gilt** —
// OPS-04.AK-01 und AK-02
// (#1721).

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gestartetMitWeiche startet eine Anwendung, der eine Weiche mitgegeben ist
// — die Einstellungen werden vor dem Start in den Bestand geschrieben.
func gestartetMitWeiche(t *testing.T, werte map[string]string) (*App, *Protokollweiche) {
	t.Helper()

	// **Das Zuhause ist der Ordner des Tests** — ein eigener Protokollort
	// wird beim Setzen an der Platte geprüft, und `~` meint dann ihn.
	ordner := t.TempDir()
	t.Setenv("HOME", ordner)

	if fehler := os.Mkdir(filepath.Join(ordner, "Protokolle"), 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	weiche := neueWeicheMit(ordner, &bytes.Buffer{}, &bytes.Buffer{}, func() (string, error) { return ordner, nil })

	t.Cleanup(func() {
		if fehler := weiche.Schliessen(); fehler != nil {
			t.Error(fehler)
		}
	})

	pfad := filepath.Join(ordner, "bestand.db")
	vorher := NeueApp(pfad)
	if fehler := vorher.Starten(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	for schluessel, wert := range werte {
		if antwort := vorher.EinstellungSetzen(t.Context(), schluessel, wert); antwort.Fehler != nil {
			t.Fatalf("%s: %+v", schluessel, antwort.Fehler)
		}
	}

	if fehler := vorher.Beenden(); fehler != nil {
		t.Fatal(fehler)
	}

	anwendung := NeueApp(pfad).MitProtokollweiche(weiche)
	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Error(fehler)
		}
	})

	return anwendung, weiche
}

func TestNachDemStartGiltWasEingestelltIst(t *testing.T) {
	_, weiche := gestartetMitWeiche(t, map[string]string{
		domain.ProtokollStufe: domain.ProtokollstufeWarnung, domain.ProtokollAusgabe: domain.ProtokollAn,
	})

	if stand := weiche.Stand(); stand.Stufe != domain.ProtokollstufeWarnung || !kanal(stand, KanalAusgabe).Aktiv {
		t.Errorf("%+v", stand)
	}
}

// **Die Diagnose nennt jeden Kanal, und einen fehlenden mit Grund** — AK-02:
// „nicht verfügbare Protokollierung ist in Diagnose oder Status sichtbar".
func TestDieDiagnoseNenntDieKanaele(t *testing.T) {
	ordner := t.TempDir()
	keinOrdner := filepath.Join(ordner, "eine-datei")

	if fehler := os.WriteFile(keinOrdner, nil, 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	anwendung, weiche := gestartetMitWeiche(t, nil)
	weiche.Einrichten(domain.ProtokollregelAus(map[string]string{domain.ProtokollOrt: keinOrdner}))

	text := anwendung.Diagnose(t.Context()).Text
	texte := DialogtexteFuer("de")

	for _, teil := range []string{
		texte.DiagnoseKanaele + ":", texte.DiagnoseKanalDatei + " " + texte.DiagnoseKanalFehlt,
		texte.DiagnoseKanalSystem + " " + texte.DiagnoseKanalAn,
		texte.DiagnoseKanalAusgabe + " " + texte.DiagnoseKanalAus,
	} {
		if !strings.Contains(text, teil) {
			t.Errorf("die Diagnose nennt %q nicht:\n%s", teil, text)
		}
	}

	if strings.Contains(text, texte.DiagnoseAusfuehrlich) {
		t.Error("der Hinweis auf »ausführlich« steht ohne Grund da")
	}

	// **Ausführlich sagt sie es** — AK-03: Hinweis und begrenzte Aufbewahrung.
	weiche.Einrichten(domain.ProtokollregelAus(map[string]string{domain.ProtokollStufe: domain.ProtokollstufeAusfuehrlich}))

	if text := anwendung.Diagnose(t.Context()).Text; !strings.Contains(text, texte.DiagnoseAusfuehrlich) {
		t.Errorf("kein Hinweis:\n%s", text)
	}
}

// **Die Diagnose liest die Datei, in die wirklich geschrieben wird** — mit
// einem eigenen Ort nicht die neben dem Bestand.
func TestDieDiagnoseLiestDieDateiDerWeiche(t *testing.T) {
	anwendung, weiche := gestartetMitWeiche(t, map[string]string{domain.ProtokollOrt: "~/Protokolle"})

	pfad, fehler := anwendung.protokolldatei()
	if fehler != nil || pfad != weiche.Stand().Datei || !strings.Contains(pfad, "Protokolle") {
		t.Errorf("%q, %v", pfad, fehler)
	}

	ohne := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	if pfad, fehler := ohne.protokolldatei(); fehler != nil || filepath.Base(pfad) != Protokollname {
		t.Errorf("ohne Weiche: %q, %v", pfad, fehler)
	}
}

// **Ein Lesefehler lässt die Vorgabe stehen** — das Protokoll schreibt
// weiter, und die Zeile darüber sagt, warum es die Einstellungen nicht kennt.
func TestEinLesefehlerLaesstDieVorgabeStehen(t *testing.T) {
	anwendung, weiche := gestartetMitWeiche(t, map[string]string{domain.ProtokollStufe: domain.ProtokollstufeFehler})
	weiche.Einrichten(domain.ProtokollregelAus(nil))

	ctx := anwendung.imProfil(t.Context())
	if fehler := anwendung.db.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	anwendung.protokollEinrichten(ctx)

	if stufe := weiche.Stand().Stufe; stufe != domain.ProtokollstufeInfo {
		t.Errorf("Stufe %q", stufe)
	}
}
