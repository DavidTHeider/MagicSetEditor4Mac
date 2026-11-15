//go:build darwin

// **Die Diagnose ist rollengefiltert** — Werkstück O9 von Phase 5.2
// (#1433).
//
// Die Abnahme aus dem Issue lautet „fremde Profile und Geheimnisse sind nicht
// sichtbar". Das Protokoll gehört dem **Prozess**: wer als Gast in einem
// fremden Profil arbeitet, läse darin die Spuren seines Gastgebers.
package app_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinGastSiehtDieProtokollzeilenNicht(t *testing.T) {
	anwendung := alsLeserHinein(t)

	texte := app.DialogtexteFuer("de")

	diagnose := anwendung.Diagnose(t.Context())
	if diagnose.Fehler != nil {
		t.Fatalf("Diagnose: %+v", diagnose.Fehler)
	}

	switch {
	case !strings.Contains(diagnose.Text, texte.DiagnoseProtokollVerdeckt):
		t.Errorf("dem Gast fehlt der Satz, warum er nichts sieht:\n%s", diagnose.Text)

	// **Der Rest bleibt stehen** — Fassung, Rechnerart und der Ort des
	// Bestands sind eine Auskunft über die Installation und über niemanden.
	case !strings.Contains(diagnose.Text, "MagicSetEditor4Mac "):
		t.Errorf("dem Gast fehlt auch der Kopf:\n%s", diagnose.Text)
	case !strings.Contains(diagnose.Text, texte.DiagnoseBestand):
		t.Errorf("dem Gast fehlt der Ort des Bestands:\n%s", diagnose.Text)
	}
}

// **Und der Eigner sieht sie.** Ohne diese Hälfte wäre der Test oben auch
// dann grün, wenn niemand mehr etwas sähe.
func TestDerVerwalterSiehtDieProtokollzeilen(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	texte := app.DialogtexteFuer("de")

	diagnose := anwendung.Diagnose(t.Context())
	if diagnose.Fehler != nil {
		t.Fatalf("Diagnose: %+v", diagnose.Fehler)
	}

	if strings.Contains(diagnose.Text, texte.DiagnoseProtokollVerdeckt) {
		t.Errorf("dem Eigner wurde das Protokoll verdeckt:\n%s", diagnose.Text)
	}
}

// **Wer nicht sagen kann, wer da ist, verdeckt.** Ein Lesefehler beim Akteur
// ist kein Freibrief: er heißt „unbekannt", und unbekannt sieht nichts.
func TestEinLesefehlerBeimAkteurVerdecktDasProtokoll(t *testing.T) {
	db := frischerBestand(t)
	anwendung := app.NeueApp("")
	anwendung.MitAutorisierung(service.NeueAutorisierung(
		stolperzeiger{fehler: errors.New("der Zeiger klemmt")},
		service.NeueFreigabenverwaltung(
			storage.NeueFreigabenablage(db), storage.NeueNutzerablage(db))))

	diagnose := anwendung.Diagnose(probekontext(t))
	if !strings.Contains(diagnose.Text,
		app.DialogtexteFuer("de").DiagnoseProtokollVerdeckt) {
		t.Errorf("bei unklarem Akteur stand das Protokoll trotzdem da:\n%s",
			diagnose.Text)
	}
}
