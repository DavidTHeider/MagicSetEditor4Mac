package domain_test

// **Der KI-Auftrag und sein Vorschlag** — Werkstück A6 von Phase 5.3
// (#1445).

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// derZeitpunkt ist die Uhr, gegen die die Läufe unten rechnen.
var derZeitpunkt = time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)

// einAuftrag baut einen gültigen Auftrag.
func einAuftrag() domain.KIAuftrag {
	return domain.KIAuftrag{
		Kennung:       "auftrag-1",
		ProfilID:      7,
		Faehigkeit:    domain.KITextgenerierung,
		Anbieter:      "ollama",
		Alias:         "beispiel:3b",
		Ort:           domain.KIOrtLokal,
		Zielkarte:     42,
		Zielrevision:  "kartenstand-abc",
		Schemakennung: einSchema().Kennung(),
		Kostenposten:  "kosten-1",
		Angelegt:      derZeitpunkt,
		Ablauf:        derZeitpunkt.Add(time.Hour),
	}
}

// **Ein vollständiger Auftrag geht durch.**
func TestEinVollstaendigerAuftragGehtDurch(t *testing.T) {
	t.Parallel()

	if befunde := einAuftrag().Validate(); len(befunde) != 0 {
		t.Errorf("der Auftrag wurde abgewiesen: %v", befunde)
	}
}

// **Jedes Stück des Auftrags ist Pflicht.**
//
// Das Issue zählt sie auf: „KI-Job erhält Profil, Zielrevision,
// Schema-/Paketversion, Ablauf, Budgetreservierung, Abbruch und konsumierbare
// Vorschlags-ID." Fehlt eines, ist der Auftrag nicht an den Stand gebunden,
// gegen den er läuft — und ein Vorschlag daraus träfe irgendetwas.
func TestJedesStueckDesAuftragsIstPflicht(t *testing.T) {
	t.Parallel()

	for name, kaputt := range map[string]func(*domain.KIAuftrag){
		"ohne Kennung":       func(a *domain.KIAuftrag) { a.Kennung = "" },
		"ohne Profil":        func(a *domain.KIAuftrag) { a.ProfilID = 0 },
		"ohne Zielkarte":     func(a *domain.KIAuftrag) { a.Zielkarte = 0 },
		"ohne Zielrevision":  func(a *domain.KIAuftrag) { a.Zielrevision = "" },
		"ohne Schemakennung": func(a *domain.KIAuftrag) { a.Schemakennung = "" },
		"ohne Kostenposten":  func(a *domain.KIAuftrag) { a.Kostenposten = "" },
		"ohne Ablauf":        func(a *domain.KIAuftrag) { a.Ablauf = time.Time{} },
		"ohne Anbieter":      func(a *domain.KIAuftrag) { a.Anbieter = "" },
		"ohne Alias":         func(a *domain.KIAuftrag) { a.Alias = "" },
		"mit erfundener Fähigkeit": func(a *domain.KIAuftrag) {
			a.Faehigkeit = "hellsehen"
		},
		"mit erfundenem Ort": func(a *domain.KIAuftrag) { a.Ort = "irgendwo" },
		"mit einem Ablauf vor der Anlage": func(a *domain.KIAuftrag) {
			a.Ablauf = a.Angelegt.Add(-time.Minute)
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			auftrag := einAuftrag()
			kaputt(&auftrag)

			if befunde := auftrag.Validate(); len(befunde) == 0 {
				t.Errorf("%s kam durch", name)
			}
		})
	}
}

// **Ein abgelaufener Auftrag ist abgelaufen.**
//
// **Genau am Ablauf noch nicht**: eine Frist, die im selben Augenblick endet,
// in dem sie gilt, ist an einer Uhr mit Millisekunden ein Zufall.
func TestEinAbgelaufenerAuftragIstAbgelaufen(t *testing.T) {
	t.Parallel()

	auftrag := einAuftrag()

	if auftrag.Abgelaufen(derZeitpunkt) {
		t.Error("der frische Auftrag gilt als abgelaufen")
	}

	if auftrag.Abgelaufen(auftrag.Ablauf) {
		t.Error("genau am Ablauf gilt er schon als abgelaufen")
	}

	if !auftrag.Abgelaufen(auftrag.Ablauf.Add(time.Nanosecond)) {
		t.Error("eine Nanosekunde danach gilt er nicht als abgelaufen")
	}
}
