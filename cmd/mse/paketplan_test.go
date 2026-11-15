package main

// **Plan, Ersetzen und Zurück — ohne Fenster** — DOM-01.AK-02 und AK-03
// (#1696).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fassungsquelle legt ein Stilpaket in einer bestimmten Fassung an.
func fassungsquelle(t *testing.T, name, version string) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\nshort name: Probe\nversion: " + version + "\n"
	if err := os.WriteFile(filepath.Join(ort, "style"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return ort
}

// **Planen, ersetzen, zurück — und der Katalog sieht es jedes Mal.**
func TestPlanErsetzenUndZurueckOhneFenster(t *testing.T) {
	ort := eigenesZuhause(t)

	if code, _, fehler := fahreAn(t, ort, "paket-installieren",
		fassungsquelle(t, "probe-tausch.mse-style", "2026-08-30")); code != CodeGut {
		t.Fatalf("einbinden: %d (%s)", code, fehler)
	}

	neu := fassungsquelle(t, "probe-tausch.mse-style", "2026-09-01")

	code, aus, fehler := fahreAn(t, ort, "paket-plan", "einbinden", neu, "--json")
	if code != CodeGut {
		t.Fatalf("Plan: %d (%s)", code, fehler)
	}

	var geplant struct {
		Plan struct {
			Vorgang, Von, Zu string
			Moeglich         bool
		} `json:"plan"`
	}

	if err := json.Unmarshal([]byte(aus), &geplant); err != nil ||
		geplant.Plan.Vorgang != "ersetzen" || geplant.Plan.Von != "2026-08-30" ||
		geplant.Plan.Zu != "2026-09-01" || !geplant.Plan.Moeglich {
		t.Fatalf("der Plan: %+v (%v, %s)", geplant, err, aus)
	}

	if code, aus, fehler := fahreAn(t, ort, "paket-ersetzen", neu); code != CodeGut ||
		!strings.Contains(aus, "2026-09-01") {
		t.Fatalf("ersetzen: %d %q (%s)", code, aus, fehler)
	}

	// **Der Plan als Text** nennt Vorgang und Fassungen.
	if code, aus, _ := fahreAn(t, ort, "paket-plan", "zurueck", "mse-style", "probe-tausch"); code != CodeGut ||
		!strings.Contains(aus, "zurueck") || !strings.Contains(aus, "2026-08-30") {
		t.Errorf("der Plan zurück: %d %q", code, aus)
	}

	if code, aus, fehler := fahreAn(t, ort, "paket-zurueck", "mse-style", "probe-tausch", "--json"); code != CodeGut ||
		!strings.Contains(aus, `"version": "2026-08-30"`) {
		t.Fatalf("zurück: %d %q (%s)", code, aus, fehler)
	}

	if _, liste, _ := fahreAn(t, ort, "pakete"); !strings.Contains(liste, "2026-08-30") {
		t.Errorf("der Katalog kennt die vorige Fassung nicht: %s", liste)
	}
}

// **Ein lokal geändertes Paket braucht `--bestaetigt`.**
func TestErsetzenOhneBestaetigungLaesstDieAenderungStehen(t *testing.T) {
	ort := eigenesZuhause(t)

	code, aus, fehler := fahreAn(t, ort, "paket-installieren",
		fassungsquelle(t, "probe-eigen.mse-style", "2026-08-30"), "--json")
	if code != CodeGut {
		t.Fatalf("einbinden: %d (%s)", code, fehler)
	}

	var eingebunden struct {
		Paket paketDTO `json:"paket"`
	}

	if err := json.Unmarshal([]byte(aus), &eingebunden); err != nil {
		t.Fatalf("Antwort: %v (%s)", err, aus)
	}

	if err := os.WriteFile(filepath.Join(eingebunden.Paket.Pfad, "style"),
		[]byte("mse version: 2.0.0\nversion: 2026-08-30\n# eine eigene, längere Zeile\n"), 0o600); err != nil {
		t.Fatalf("bearbeiten: %v", err)
	}

	neu := fassungsquelle(t, "probe-eigen.mse-style", "2026-09-01")

	if code, _, fehler := fahreAn(t, ort, "paket-ersetzen", neu); code != CodeFachlich || fehler == "" {
		t.Errorf("ohne Bestätigung: %d (%s)", code, fehler)
	}

	if code, _, fehler := fahreAn(t, ort, "paket-ersetzen", "--bestaetigt", neu); code != CodeGut {
		t.Errorf("mit Bestätigung: %d (%s)", code, fehler)
	}
}

// **Falsche Wörter enden als Bedienfehler**, bevor der Bestand geöffnet
// wird — und ein fachlicher Fehler als fachlicher.
func TestPlanErsetzenUndZurueckWeisenFalscheWoerterAb(t *testing.T) {
	ort := eigenesZuhause(t)

	for name, argumente := range map[string][]string{
		"Plan ohne Vorgang":           {"paket-plan"},
		"Plan eines Unbekannten":      {"paket-plan", "wegzaubern", "mse-style", "x"},
		"Einbinden ohne Quelle":       {"paket-plan", "einbinden"},
		"Entfernen ohne Namen":        {"paket-plan", "entfernen", "mse-style"},
		"Ersetzen ohne Quelle":        {"paket-ersetzen"},
		"Ersetzen mit fremder Fahne":  {"paket-ersetzen", "/q", "--gewaltsam"},
		"Ersetzen mit zwei Quellen":   {"paket-ersetzen", "/q", "/r"},
		"Zurück mit nur einem Wort":   {"paket-zurueck", "mse-style"},
		"Ersetzen mit leerer Quelle":  {"paket-ersetzen", ""},
		"Plan mit Quelle aus Nichts":  {"paket-plan", "einbinden", filepath.Join(t.TempDir(), "fehlt.mse-style")},
		"Zurück ohne vorige Fassung":  {"paket-zurueck", "mse-style", "gibtsnicht"},
		"Plan zu einem fehlenden":     {"paket-plan", "entfernen", "mse-style", "gibtsnicht"},
		"Ersetzen ohne Vorgänger":     {"paket-ersetzen", fassungsquelle(t, "neu.mse-style", "1")},
		"Zurück zu einem Unbekannten": {"paket-zurueck", "mse-style", "unbekannt"},
	} {
		code, _, fehler := fahreAn(t, ort, argumente...)
		if code == CodeGut || fehler == "" {
			t.Errorf("%s: %d (%s)", name, code, fehler)
		}
	}
}

// **Was einem Set an Paketen fehlt** — DOM-01.AK-04.
func TestPaketbedarfNenntDenFehlendenStil(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	saatBereitstellen(t, zuhause)

	ort := frischerBestand(t)

	// **Ein Stil, der da war und gegangen ist** — ein Set lässt sich nur mit
	// einem Stil anlegen, den es gibt.
	stil := filepath.Join(t.TempDir(), "probe-weg.mse-style")
	if err := os.MkdirAll(stil, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(stil, "style"),
		[]byte("mse version: 2.0.0\ngame: probe\nshort name: Weg\n"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if code, _, fehler := fahreAn(t, ort, "paket-installieren", stil); code != CodeGut {
		t.Fatalf("einbinden: %d (%s)", code, fehler)
	}

	_, aus, fehler := fahreAn(t, ort, "set-neu", "Bedarfset", "probe", "probe-weg", "--json")

	var angelegt setAnlegeAntwort
	if err := json.Unmarshal([]byte(aus), &angelegt); err != nil {
		t.Fatalf("Set anlegen: %v (%s)", err, fehler)
	}

	if code, _, entfernfehler := fahreAn(t, ort, "paket-entfernen", "mse-style", "probe-weg"); code != CodeGut {
		t.Fatalf("entfernen: %d (%s)", code, entfernfehler)
	}

	code, aus, fehler := fahreAn(t, ort, "paketbedarf", angelegt.ID, "--json")
	if code != CodeGut || !strings.Contains(aus, `"braucht": "probe-weg.mse-style"`) {
		t.Fatalf("Bedarf: %d %s (%s)", code, aus, fehler)
	}

	if code, aus, _ := fahreAn(t, ort, "paketbedarf", angelegt.ID); code != CodeGut ||
		!strings.Contains(aus, "fehlt\tBedarfset braucht probe-weg.mse-style") {
		t.Errorf("als Text: %d %q", code, aus)
	}

	for name, argumente := range map[string][]string{
		"ohne Set":          {"paketbedarf"},
		"keine Zahl":        {"paketbedarf", "zwei"},
		"ein fehlendes Set": {"paketbedarf", "999"},
	} {
		if code, _, fehler := fahreAn(t, ort, argumente...); code == CodeGut || fehler == "" {
			t.Errorf("%s: %d (%s)", name, code, fehler)
		}
	}
}
