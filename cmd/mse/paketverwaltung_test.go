package main

// **Pakete einbinden, entfernen und aus MSE2 übernehmen — ohne Fenster**
// (#1291).
//
// Zwei Zeilen des Katalogs standen mit drei Gedankenstrichen und einer
// erhobenen Hand da. Die Hand meint einen Verzeichnisdialog — und
// `App.PaketInstallieren` wie `App.VorlagenUebernehmen` nehmen einen Pfad
// entgegen.
//
// **Wer einen Rechner aus einem Skript einrichtet**, hat kein Fenster, und
// `paket-beziehen` allein reicht nicht: ein Paket, das schon auf der Platte
// liegt, will eingebunden und nicht heruntergeladen werden.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// eigenesZuhause gibt dem Lauf ein leeres Zuhause und einen leeren Bestand.
//
// **Ohne das greift der Test nach draußen** (§3), und zwar wirklich: der erste
// Lauf dieser Datei hat drei Pakete in das **echte** Paketverzeichnis dieses
// Rechners gelegt — `Paketverzeichnis()` liegt neben der Datenbank, und die
// liegt unter `HOME`. Der Lauf war dabei grün, weil er funktioniert hat.
//
// Dieselbe Gestalt wie #888,
// nur ohne Netz: eine Naht, die das Zuhause anfasst, tut es im Test auch —
// wenn nichts sie daran hindert.
func eigenesZuhause(t *testing.T) string {
	t.Helper()

	t.Setenv("HOME", t.TempDir())

	return frischerBestand(t)
}

// paketquelle legt ein winziges, selbst geschriebenes Paket an (§11).
func paketquelle(t *testing.T, name string) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\nshort name: Probe\nfull name: Probeteil\nversion: 2026-08-30\n"
	if err := os.WriteFile(filepath.Join(ort, "style"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return ort
}

// **`paket-installieren` bindet ein Paket von der Platte ein.**
func TestPaketInstallierenBindetEin(t *testing.T) {
	ort := eigenesZuhause(t)
	quelle := paketquelle(t, "probe-eingebunden.mse-style")

	code, aus, fehler := fahreAn(t, ort, "paket-installieren", quelle, "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var eingebunden struct {
		Paket paketDTO `json:"paket"`
	}

	if err := json.Unmarshal([]byte(aus), &eingebunden); err != nil {
		t.Fatalf("Antwort lesen: %v (%s)", err, aus)
	}

	if eingebunden.Paket.Name == "" {
		t.Errorf("das Paket hat keinen Namen: %s", aus)
	}

	// **Und es steht danach im Katalog** — nicht nur in der Antwort.
	_, liste, _ := fahreAn(t, ort, "pakete")
	if !strings.Contains(liste, "probe-eingebunden") {
		t.Errorf("das Paket fehlt im Katalog: %s", liste)
	}
}

// **`paket-entfernen` nimmt es wieder heraus.**
func TestPaketEntfernenNimmtHeraus(t *testing.T) {
	ort := eigenesZuhause(t)
	quelle := paketquelle(t, "probe-wegdamit.mse-style")

	if code, _, fehler := fahreAn(t, ort, "paket-installieren", quelle); code != CodeGut {
		t.Fatalf("einbinden: %d (%s)", code, fehler)
	}

	if code, _, fehler := fahreAn(t, ort,
		"paket-entfernen", "mse-style", "probe-wegdamit"); code != CodeGut {
		t.Fatalf("entfernen: %d (%s)", code, fehler)
	}

	_, liste, _ := fahreAn(t, ort, "pakete")
	if strings.Contains(liste, "probe-wegdamit") {
		t.Errorf("das Paket steht noch im Katalog: %s", liste)
	}
}

// **`mse2-orte` nennt die üblichen Orte einer MSE2-Installation.**
//
// **Vorgeschlagen, nicht gelesen** (#55): was dort liegt, erfährt die
// Anwendung erst, wenn jemand einen Ort wählt.
func TestMSE2OrteNenntVorschlaege(t *testing.T) {
	ort := eigenesZuhause(t)

	code, aus, fehler := fahreAn(t, ort, "mse2-orte", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var orte struct {
		Orte []string `json:"orte"`
	}

	if err := json.Unmarshal([]byte(aus), &orte); err != nil {
		t.Fatalf("Antwort lesen: %v (%s)", err, aus)
	}

	// **Eine leere Liste ist eine gültige Antwort** — auf einem Rechner ohne
	// MSE2 gibt es keinen Vorschlag. Was nicht geht, ist `null`.
	if orte.Orte == nil {
		t.Errorf("die Orte sind nil statt einer leeren Liste: %s", aus)
	}
}

// **`mse2-uebernehmen` holt die Vorlagen aus einem Verzeichnis.**
func TestMSE2UebernehmenHoltVorlagen(t *testing.T) {
	ort := eigenesZuhause(t)

	quelle := paketquelle(t, "probe-uebernahme.mse-style")

	code, aus, fehler := fahreAn(t, ort,
		"mse2-uebernehmen", filepath.Dir(quelle), "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var uebernahme struct {
		Uebernommen []paketDTO `json:"uebernommen"`
		Uebergangen []string   `json:"uebergangen"`
		Befunde     []string   `json:"befunde"`
	}

	if err := json.Unmarshal([]byte(aus), &uebernahme); err != nil {
		t.Fatalf("Antwort lesen: %v (%s)", err, aus)
	}

	if len(uebernahme.Uebernommen) == 0 {
		t.Errorf("es wurde nichts übernommen: %s", aus)
	}

	// **Die drei Listen sind nie `nil`** (§3).
	if uebernahme.Uebergangen == nil || uebernahme.Befunde == nil {
		t.Errorf("eine Liste ist nil: %s", aus)
	}
}

// **Und die Bedienfehler.**
func TestPaketverwaltungBedienfehler(t *testing.T) {
	ort := eigenesZuhause(t)

	faelle := map[string][]string{
		"installieren ohne Quelle": {"paket-installieren"},
		"entfernen ohne Namen":     {"paket-entfernen", "style"},
		"übernehmen ohne Ort":      {"mse2-uebernehmen"},
		"orte mit einem Argument":  {"mse2-orte", "zuviel"},
	}

	for name, aufruf := range faelle {
		t.Run(name, func(t *testing.T) {
			if code, _, _ := fahreAn(t, ort, aufruf...); code != CodeBedienung {
				t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
			}
		})
	}
}
