package app

// **Die Protokollweiche** — OPS-04
// (#1721):
// drei Kanäle, unabhängig wählbar; ein Kanal, der ausfällt, nimmt die
// anderen nicht mit (AK-02); Stufe, Format, Filter, Ort und Drehung aus den
// Einstellungen (AK-01). Das Systemprotokoll ist hier ein Puffer — dass
// `logger` das Unified Log erreicht, belegt `protokoll_darwin_test.go`.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// weicheMitPuffern baut eine Weiche über einem Bestandsordner, deren System-
// und Ausgabekanal Puffer sind.
func weicheMitPuffern(t *testing.T, ordner string) (weiche *Protokollweiche, system, ausgabe *bytes.Buffer) {
	t.Helper()

	system, ausgabe = &bytes.Buffer{}, &bytes.Buffer{}
	weiche = neueWeicheMit(ordner, system, ausgabe, func() (string, error) { return ordner, nil })

	t.Cleanup(func() {
		if fehler := weiche.Schliessen(); fehler != nil {
			t.Error(fehler)
		}
	})

	return weiche, system, ausgabe
}

func dateiinhalt(t *testing.T, pfad string) string {
	t.Helper()

	roh, fehler := os.ReadFile(pfad) //nolint:gosec // eine Datei dieses Tests
	if fehler != nil {
		t.Fatal(fehler)
	}

	return string(roh)
}

func kanal(stand Protokollstand, name string) Protokollkanal {
	for _, eintrag := range stand.Kanaele {
		if eintrag.Name == name {
			return eintrag
		}
	}

	return Protokollkanal{}
}

// **Die Vorgabe ist, was es bisher gab** — Datei neben dem Bestand und das
// Systemprotokoll, Stufe »info«, keine Standardausgabe.
func TestDieWeicheBeginntMitDerVorgabe(t *testing.T) {
	ordner := t.TempDir()
	weiche, system, ausgabe := weicheMitPuffern(t, ordner)

	weiche.Logger().Info("der Bestand ist offen", "anzahl", 3)
	weiche.Logger().Debug("ein Schritt")

	stand := weiche.Stand()
	if !kanal(stand, KanalDatei).Aktiv || !kanal(stand, KanalSystem).Aktiv || kanal(stand, KanalAusgabe).Aktiv ||
		stand.Stufe != domain.ProtokollstufeInfo || stand.Datei != filepath.Join(ordner, Protokollname) {
		t.Errorf("%+v", stand)
	}

	if datei := dateiinhalt(t, stand.Datei); !strings.Contains(datei, "der Bestand ist offen") || strings.Contains(datei, "ein Schritt") {
		t.Errorf("Datei: %s", datei)
	}

	if !strings.Contains(system.String(), "der Bestand ist offen") || ausgabe.Len() != 0 {
		t.Errorf("System %q, Ausgabe %q", system.String(), ausgabe.String())
	}
}

// **Eingerichtet gilt die Regel** — auch für einen Schreiber, der vorher
// mit `With` gebunden wurde; er hängt an der Weiche und nicht am Kanal.
func TestEineEingerichteteRegelGiltAuchFuerGebundeneSchreiber(t *testing.T) {
	ordner := t.TempDir()
	weiche, system, ausgabe := weicheMitPuffern(t, ordner)
	gebunden := weiche.Logger().With("anzahl", 7)

	weiche.Einrichten(domain.ProtokollregelAus(map[string]string{
		domain.ProtokollDatei: domain.ProtokollAus, domain.ProtokollSystem: domain.ProtokollAus,
		domain.ProtokollAusgabe: domain.ProtokollAn, domain.ProtokollFormat: domain.ProtokollformatJSON,
		domain.ProtokollStufe: domain.ProtokollstufeWarnung,
	}))

	gebunden.Info("zu leise")
	gebunden.Warn("laut genug")

	zeilen := strings.Split(strings.TrimSpace(ausgabe.String()), "\n")
	if len(zeilen) != 1 {
		t.Fatalf("Ausgabe: %q", ausgabe.String())
	}

	var satz map[string]any
	if fehler := json.Unmarshal([]byte(zeilen[0]), &satz); fehler != nil || satz["msg"] != "laut genug" || satz["anzahl"] != "7" {
		t.Errorf("%v: %q", fehler, zeilen[0])
	}

	if strings.Contains(system.String(), "laut genug") {
		t.Error("das abgeschaltete Systemprotokoll bekam die Zeile")
	}

	if stand := weiche.Stand(); kanal(stand, KanalDatei).Aktiv || stand.Datei != "" || stand.Format != domain.ProtokollformatJSON {
		t.Errorf("%+v", stand)
	}
}

// **Ein Ort, an dem keine Datei entstehen kann, nimmt nur die Datei weg** —
// AK-02: das Systemprotokoll schreibt weiter, und der Stand sagt, warum die
// Datei fehlt.
func TestEinAusgefallenerDateikanalNimmtDieAnderenNichtMit(t *testing.T) {
	ordner := t.TempDir()
	keinOrdner := filepath.Join(ordner, "eine-datei")

	if fehler := os.WriteFile(keinOrdner, nil, 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	weiche, system, _ := weicheMitPuffern(t, ordner)
	weiche.Einrichten(domain.ProtokollregelAus(map[string]string{domain.ProtokollOrt: keinOrdner}))
	weiche.Logger().Info("trotzdem da")

	datei := kanal(weiche.Stand(), KanalDatei)
	if datei.Aktiv || datei.Grund == "" {
		t.Errorf("Datei: %+v", datei)
	}

	if !strings.Contains(system.String(), "trotzdem da") {
		t.Errorf("System: %q", system.String())
	}
}

// **Ausführlich mit Filter**: eine Zeile unter »info« kommt nur aus den
// genannten Komponenten; ab »info« kommt jede.
func TestDerKomponentenfilterHaeltNurDieAusfuehrlichenZeilen(t *testing.T) {
	for komponenten, erwartet := range map[string]bool{"app": true, "service,storage": false} {
		ordner := t.TempDir()
		weiche, system, _ := weicheMitPuffern(t, ordner)
		weiche.Einrichten(domain.ProtokollregelAus(map[string]string{
			domain.ProtokollStufe: domain.ProtokollstufeAusfuehrlich, domain.ProtokollKomponenten: komponenten,
		}))

		weiche.Logger().Debug("ein Schritt")
		weiche.Logger().Info("ein Ereignis")

		// **Gelesen wird die Datei**: die ausführliche Zeile geht seit dem
		// Nachlauf über Phase 6.4 (N5) nicht ins Systemprotokoll.
		roh, fehler := os.ReadFile(filepath.Join(ordner, Protokollname)) // #nosec G304 -- aus t.TempDir()
		if fehler != nil {
			t.Fatal(fehler)
		}

		if strings.Contains(string(roh), "ein Schritt") != erwartet || !strings.Contains(system.String(), "ein Ereignis") {
			t.Errorf("%s: %q", komponenten, roh)
		}

		if !weiche.Stand().Ausfuehrlich {
			t.Error("der Stand sagt nicht, dass es ausführlich ist")
		}
	}
}

// **Die Drehung hält die eingestellte Zahl von Vorfassungen** — und einen
// eigenen Ort unter dem Zuhause.
func TestDieDrehungHaeltDieVorfassungen(t *testing.T) {
	ordner := t.TempDir()
	ablage := filepath.Join(ordner, "Protokolle")

	if fehler := os.Mkdir(ablage, 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	for name, inhalt := range map[string]string{
		Protokollname: strings.Repeat("x", 1<<20), Protokollname + ".1": "eins",
	} {
		if fehler := os.WriteFile(filepath.Join(ablage, name), []byte(inhalt), 0o600); fehler != nil {
			t.Fatal(fehler)
		}
	}

	weiche, _, _ := weicheMitPuffern(t, ordner)
	weiche.Einrichten(domain.ProtokollregelAus(map[string]string{
		domain.ProtokollOrt: "~/Protokolle", domain.ProtokollGrenze: "1", domain.ProtokollAufbewahrung: "2",
	}))

	if dateiinhalt(t, filepath.Join(ablage, Protokollname+".2")) != "eins" ||
		len(dateiinhalt(t, filepath.Join(ablage, Protokollname+".1"))) != 1<<20 {
		t.Error("die Vorfassungen stehen nicht, wo sie hingehören")
	}

	if weiche.Stand().Datei != filepath.Join(ablage, Protokollname) {
		t.Errorf("Datei: %s", weiche.Stand().Datei)
	}
}

// **Ohne Zuhause gibt es keinen eigenen Ort** — dann bleibt die Datei neben
// dem Bestand, und der Stand sagt es.
func TestOhneZuhauseBleibtDieDateiNebenDemBestand(t *testing.T) {
	ordner := t.TempDir()
	weiche := neueWeicheMit(ordner, &bytes.Buffer{}, &bytes.Buffer{}, func() (string, error) { return "", os.ErrNotExist })

	t.Cleanup(func() {
		if fehler := weiche.Schliessen(); fehler != nil {
			t.Error(fehler)
		}
	})

	weiche.Einrichten(domain.ProtokollregelAus(map[string]string{domain.ProtokollOrt: "~/Protokolle"}))

	if datei := kanal(weiche.Stand(), KanalDatei); datei.Aktiv || datei.Grund == "" {
		t.Errorf("%+v", datei)
	}
}

// **Ohne Ort des Bestands gibt es keine Datei** — `main` baut die Weiche
// dann so, wenn er sich nicht bestimmen ließ; die anderen Kanäle schreiben.
func TestOhneBestandsortSchreibenDieAnderenKanaele(t *testing.T) {
	system := &bytes.Buffer{}
	weiche := neueWeicheMit("", system, &bytes.Buffer{}, func() (string, error) { return t.TempDir(), nil })

	t.Cleanup(func() {
		if fehler := weiche.Schliessen(); fehler != nil {
			t.Error(fehler)
		}
	})

	weiche.Logger().Info("noch da")

	if datei := kanal(weiche.Stand(), KanalDatei); datei.Aktiv || datei.Grund == "" || !strings.Contains(system.String(), "noch da") {
		t.Errorf("%+v, %q", datei, system.String())
	}

	// **Zweimal schließen ist einmal schließen.**
	if fehler := weiche.Schliessen(); fehler != nil {
		t.Error(fehler)
	}
}
