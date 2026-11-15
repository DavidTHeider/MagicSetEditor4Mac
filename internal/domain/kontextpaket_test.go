package domain_test

// **Das Kontextpaket** — Werkstück A16 von Phase 5.3
// (#1515).

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einKontextmanifest baut ein vollständiges.
func einKontextmanifest() domain.Kontextmanifest {
	jetzt := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)

	return domain.Kontextmanifest{
		PaketID: "kp-1", Aufgabe: "Erkläre Trampelschaden",
		Zielrevision: "rev-1",
		Zielanbieter: "ollama", Zielmodell: "gemma3:270m",
		Laufzeit: "ollama 0.33.3", Zaehlweise: "zeichen/4",
		Kontextgrenze: 8192, Ausgabereserve: 1024, Verbraucht: 300,
		Kontextpolicy: domain.Kontextpolicyfassung,
		ProfilID:      1,
		Erstellt:      jetzt, Laeuftab: jetzt.Add(domain.Kontextfrist),
	}
}

// **Ein Manifest ohne Budget ist keines.**
func TestEinManifestOhneBudgetIstKeines(t *testing.T) {
	t.Parallel()

	for name, aendern := range map[string]func(*domain.Kontextmanifest){
		"ohne Paketkennung": func(m *domain.Kontextmanifest) { m.PaketID = "" },
		"ohne Aufgabe":      func(m *domain.Kontextmanifest) { m.Aufgabe = "" },
		"ohne Zielmodell":   func(m *domain.Kontextmanifest) { m.Zielmodell = "" },
		"ohne Anbieter":     func(m *domain.Kontextmanifest) { m.Zielanbieter = "" },
		"ohne Zählweise":    func(m *domain.Kontextmanifest) { m.Zaehlweise = "" },
		"ohne Policy":       func(m *domain.Kontextmanifest) { m.Kontextpolicy = "" },
		"ohne Profil":       func(m *domain.Kontextmanifest) { m.ProfilID = 0 },
		"ohne Fenster":      func(m *domain.Kontextmanifest) { m.Kontextgrenze = 0 },
		"negative Reserve":  func(m *domain.Kontextmanifest) { m.Ausgabereserve = -1 },
		"Reserve füllt alles": func(m *domain.Kontextmanifest) {
			m.Ausgabereserve = m.Kontextgrenze
		},
		"zu viele Belege": func(m *domain.Kontextmanifest) {
			m.Verbraucht = m.Kontextgrenze
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			manifest := einKontextmanifest()
			aendern(&manifest)

			if fehler := manifest.Pruefe(); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("%s kam durch: %v", name, fehler)
			}
		})
	}

	if fehler := einKontextmanifest().Pruefe(); fehler != nil {
		t.Fatalf("das vollständige Manifest: %v", fehler)
	}
}

// **Ein anderes Zielmodell ist ein anderes Paket.**
//
// **Das Issue sagt es scharf**: „gleiche Aufgabe/Zielrevision genügt nicht zur
// Wiederverwendung." Die Zählweise hängt am Modell, und eine Zahl von einem
// anderen zu übernehmen hieße, ein Budget zu behaupten.
func TestEinAnderesZielmodellIstEinAnderesPaket(t *testing.T) {
	t.Parallel()

	eines := einKontextmanifest()

	for name, aendern := range map[string]func(*domain.Kontextmanifest){
		"anderer Anbieter": func(m *domain.Kontextmanifest) { m.Zielanbieter = "anthropic" },
		"anderes Modell":   func(m *domain.Kontextmanifest) { m.Zielmodell = "ein-anderes" },
		"andere Laufzeit":  func(m *domain.Kontextmanifest) { m.Laufzeit = "ollama 0.34.0" },
		"andere Zählweise": func(m *domain.Kontextmanifest) { m.Zaehlweise = "tiktoken" },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anderes := einKontextmanifest()
			aendern(&anderes)

			if eines.PasstZu(anderes) {
				t.Fatalf("%s galt als dasselbe Ziel", name)
			}
		})
	}

	// **Und dieselbe Bindung bei anderer Aufgabe passt** — die Aufgabe
	// gehört zum Paket, nicht zur Budgetierung.
	andereAufgabe := einKontextmanifest()
	andereAufgabe.Aufgabe = "etwas anderes"

	if !eines.PasstZu(andereAufgabe) {
		t.Error("dieselbe Bindung galt als andere")
	}
}

// **Ein Paket läuft ab.**
func TestEinKontextpaketLaeuftAb(t *testing.T) {
	t.Parallel()

	manifest := einKontextmanifest()

	if manifest.Abgelaufen(manifest.Erstellt) {
		t.Error("das frische Paket galt als abgelaufen")
	}

	// **Genau zur Frist ist es um** — eine Sekunde davor nicht.
	if manifest.Abgelaufen(manifest.Laeuftab.Add(-time.Second)) {
		t.Error("kurz vor der Frist galt es als abgelaufen")
	}

	if !manifest.Abgelaufen(manifest.Laeuftab) {
		t.Error("zur Frist galt es noch")
	}
}

// **Das Verfügbare ist das Fenster ohne die Reserve.**
func TestDasVerfuegbareIstDasFensterOhneReserve(t *testing.T) {
	t.Parallel()

	if verfuegbar := einKontextmanifest().Verfuegbar(); verfuegbar != 7168 {
		t.Fatalf("verfügbar sind %d", verfuegbar)
	}
}

// **Die Belegkennungen zählen ab eins.**
func TestDieBelegkennungenZaehlenAbEins(t *testing.T) {
	t.Parallel()

	if erste := domain.BelegkennungFuer(0); erste != "B1" {
		t.Errorf("die erste ist %q", erste)
	}

	manifest := einKontextmanifest()
	manifest.Belege = []domain.Beleg{{Kennung: "B1"}, {Kennung: "B2"}}

	kennungen := manifest.Kennungen()
	if len(kennungen) != 2 || kennungen[1] != "B2" {
		t.Fatalf("die Kennungen sind %v", kennungen)
	}
}

// **Der Belegabdruck geht über den Ausschnitt.**
//
// **Über den Ausschnitt und nicht über den Abschnitt**: er sagt, ob **das**
// noch dasteht, was das Modell gesehen hat.
func TestDerBelegabdruckGehtUeberDenAusschnitt(t *testing.T) {
	t.Parallel()

	abdruck := domain.Belegabdruck("ein Ausschnitt")

	if !strings.HasPrefix(abdruck, "sha256:") || len(abdruck) != 71 {
		t.Fatalf("der Abdruck ist %q", abdruck)
	}

	if domain.Belegabdruck("ein anderer") == abdruck {
		t.Error("zwei Ausschnitte gaben denselben Abdruck")
	}
}
