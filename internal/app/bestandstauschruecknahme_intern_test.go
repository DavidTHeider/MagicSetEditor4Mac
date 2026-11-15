package app

// **Der Tausch des Bestands und der Rückweg, wenn das Öffnen scheitert** —
// A05 (#2219).
// Aus `bestandstauschtor_intern_test.go` herausgelöst (§2, Zeilengrenze —
// Restbefund des Nachlaufs über Phase A).

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// **Scheitert das Wiederöffnen, bleibt die Anwendung für den Rückweg da** —
// dieselbe Zusage wie nach einem gescheiterten Start (#962): wer nicht
// öffnen kann, muss eine Sicherung zurückspielen können.
func TestEinGescheitertesWiederoeffnenLaesstDenRueckwegOffen(t *testing.T) {
	a, bestand := tauschbereit(t)

	fehler := a.imTausch(func() error {
		if fehler := a.zumTauschSchliessen(); fehler != nil {
			return fehler
		}

		if fehler := os.WriteFile(bestand, []byte("kein Bestand"), 0o600); fehler != nil {
			t.Fatalf("verderben: %v", fehler)
		}

		return a.wiederOeffnen(t.Context())
	})
	if fehler == nil {
		t.Fatal("ein verdorbener Bestand ließ sich öffnen")
	}

	var tor torfehler

	func() {
		defer func() {
			if grund := recover(); grund != nil {
				if fehler, ist := grund.(error); ist && errors.As(fehler, &tor) {
					return
				}

				panic(grund)
			}
		}()

		NeueBindungen(a).SicherungBeschauen(t.Context(), filepath.Join(t.TempDir(), "fehlt.db"), "")
	}()

	if tor.fehler != nil {
		t.Errorf("nach dem gescheiterten Wiederöffnen weist das Tor den Rückweg ab: %+v", tor.fehler)
	}
}

// **Scheitert das Öffnen nach Daten- und Bildertausch, steht danach der ganze
// alte Stand da — und das Tor öffnet erst dann** (A05, Befunde 3 und 5 des
// Zweitblicks über Phase A). Die Kohorte trägt die Sprache und ein Bild im
// Beiwerk; der Lauf geht durch die echte Sicherung als Archiv.
func TestEinScheiterndesOeffnenNachDemTauschHinterlaesstDenGanzenAltenStand(t *testing.T) {
	a, bestand := tauschbereit(t)
	bilder := filepath.Join(filepath.Dir(bestand), "beiwerk")
	a.MitBilderort(func() (string, error) { return bilder, nil })

	stand := func(sprache, bild string) {
		if antwort := a.EinstellungSetzen(t.Context(), "sprache", sprache); antwort.Fehler != nil {
			t.Fatalf("setzen: %+v", antwort.Fehler)
		}

		if fehler := os.MkdirAll(bilder, 0o700); fehler != nil {
			t.Fatal(fehler)
		}

		if fehler := os.WriteFile(filepath.Join(bilder, "kohorte.png"), []byte(bild), 0o600); fehler != nil {
			t.Fatal(fehler)
		}
	}

	stand("de", "neu")

	sicherung := filepath.Join(filepath.Dir(bestand), "gut.msesicherung")
	a.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) { return sicherung, nil })

	if antwort := a.BestandSichern(t.Context(), "", false); antwort.Fehler != nil {
		t.Fatalf("sichern: %+v", antwort.Fehler)
	}

	stand("en", "alt")

	// **Das erste Öffnen — das des zurückgespielten Stands — scheitert**,
	// und während der Rücknahme fragt das Tor noch einmal.
	echt := a.oeffnen
	versuche := 0
	torWaehrendDerRuecknahme := false
	a.oeffnen = func(ctx context.Context, pfad string) (*sql.DB, error) {
		versuche++
		if versuche == 1 {
			return nil, errors.New("die zurückgespielte Datei taugt nicht")
		}

		torWaehrendDerRuecknahme = torAntwort(t, a) == nil

		return echt(ctx, pfad)
	}

	antwort := a.SicherungZurueckspielen(t.Context(), sicherung, "")
	if antwort.Fehler == nil {
		t.Fatal("das gescheiterte Öffnen wurde als Erfolg gemeldet")
	}

	if !antwort.Unveraendert {
		t.Error("nach einem gelungenen Rückweg sagt die Antwort nicht »unverändert«")
	}

	if torWaehrendDerRuecknahme {
		t.Error("während der Rücknahme ließ das Tor durch")
	}

	roh, _ := os.ReadFile(filepath.Join(bilder, "kohorte.png")) //nolint:errcheck,gosec // aus t.TempDir()
	if sprache := a.Einstellungen(t.Context()); string(roh) != "alt" || spracheAus(sprache) != "en" {
		t.Errorf("danach stehen Sprache %q und Bild %q — erwartet war der ganze alte Stand", spracheAus(sprache), roh)
	}

	if abgewiesen := torAntwort(t, a); abgewiesen != nil {
		t.Errorf("nach der Rücknahme weist das Tor ab: %+v", abgewiesen)
	}
}

// **Scheitert auch der Rückweg, sagt die Antwort es** — die Rückkehrmeldung
// aus A05 darf nur „unverändert" sagen, wenn es stimmt (Nachlauf über
// Phase A). Hier lässt sich das Beiwerk der Sicherung nicht wegräumen.
func TestEinGescheiterterRueckwegIstNichtUnveraendert(t *testing.T) {
	a, bestand := tauschbereit(t)
	bilder := filepath.Join(filepath.Dir(bestand), "beiwerk")
	a.MitBilderort(func() (string, error) { return bilder, nil })

	if fehler := os.MkdirAll(bilder, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	sicherung := filepath.Join(filepath.Dir(bestand), "gut.msesicherung")
	a.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) { return sicherung, nil })

	if antwort := a.BestandSichern(t.Context(), "", false); antwort.Fehler != nil {
		t.Fatalf("sichern: %+v", antwort.Fehler)
	}

	echt := a.oeffnen
	versuche := 0
	a.oeffnen = func(ctx context.Context, pfad string) (*sql.DB, error) {
		versuche++
		if versuche == 1 {
			unleerbar(t, filepath.Join(bilder, "sperre"))

			return nil, errors.New("die zurückgespielte Datei taugt nicht")
		}

		return echt(ctx, pfad)
	}

	antwort := a.SicherungZurueckspielen(t.Context(), sicherung, "")
	if antwort.Fehler == nil {
		t.Fatal("das gescheiterte Öffnen wurde als Erfolg gemeldet")
	}

	if antwort.Unveraendert {
		t.Error("nach einem gescheiterten Rückweg sagt die Antwort »unverändert«")
	}
}

// unleerbar legt ein Verzeichnis mit einer Datei an, das sich nicht leeren
// lässt.
func unleerbar(t *testing.T, ort string) {
	t.Helper()

	if fehler := os.MkdirAll(ort, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.WriteFile(filepath.Join(ort, "bleibt.png"), []byte("bleibt"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	//nolint:gosec // G302: ein Verzeichnis braucht das Ausführungsrecht, um betreten zu werden
	if fehler := os.Chmod(ort, 0o500); fehler != nil {
		t.Fatal(fehler)
	}

	//nolint:errcheck,gosec // t.TempDir räumt sonst nicht; ein Verzeichnis braucht x
	t.Cleanup(func() { _ = os.Chmod(ort, 0o700) })
}

// spracheAus liest die Spracheinstellung aus einer Antwort.
func spracheAus(antwort EinstellungenAntwort) string {
	for _, eintrag := range antwort.Einstellungen {
		if eintrag.Schluessel == "sprache" {
			return eintrag.Wert
		}
	}

	return ""
}
