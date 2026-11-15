package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine eigene Übersetzung an der Kommandozeile** — LOC-02
// (#1688).
func TestEineEigeneUebersetzungGehtDurchDieKommandozeile(t *testing.T) {
	ort := frischerBestand(t)

	if leer := mussLaufen(t, ort, "sprachpaket"); !strings.Contains(leer, "sprachpaket anlegen") {
		t.Errorf("ohne Paket: %q", leer)
	}

	mussLaufen(t, ort, "sprachpaket", "anlegen", "pt-BR", "Português do Brasil", "en")
	mussLaufen(t, ort, "sprachpaket", "setze", "pt-BR", "--lizenz=CC0-1.0", "--stand=maschinell",
		"allgemein.arbeitsplatz.keinSet=Nenhum set aberto")

	if liste := mussLaufen(t, ort, "sprachpaket", "liste"); !strings.Contains(liste, "pt-BR") ||
		!strings.Contains(liste, "maschinell") || !strings.Contains(liste, "1/") {
		t.Errorf("die Liste: %q", liste)
	}

	if gezeigt := mussLaufen(t, ort, "sprachpaket", "zeige", "pt-BR"); !strings.Contains(gezeigt, `"fehlendePfade"`) {
		t.Errorf("gezeigt: %q", gezeigt)
	}

	if text := mussLaufen(t, ort, "sprachpaket", "ausgeben", "pt-BR"); !strings.Contains(text, "mse4mac/sprachpaket/1") {
		t.Errorf("als Text: %q", text)
	}

	datei := filepath.Join(t.TempDir(), "pt.json")
	if aus := mussLaufen(t, ort, "sprachpaket", "ausgeben", "pt-BR", datei); !strings.Contains(aus, datei) {
		t.Errorf("in die Datei: %q", aus)
	}

	if _, err := os.Stat(datei); err != nil {
		t.Fatalf("die Datei: %v", err)
	}

	mussLaufen(t, ort, "sprachpaket", "loeschen", "pt-BR")

	if aus := mussLaufen(t, ort, "sprachpaket", "einlesen", datei); !strings.Contains(aus, "pt-BR") {
		t.Errorf("eingelesen: %q", aus)
	}

	// **Eine vorhandene nur mit `--ersetzen`.**
	if code, _, _ := fahreAn(t, ort, "sprachpaket", "einlesen", datei); code != CodeFachlich {
		t.Errorf("doppelt eingelesen gab %d", code)
	}

	mussLaufen(t, ort, "sprachpaket", "einlesen", datei, "--ersetzen")

	if json := mussLaufen(t, ort, "sprachpaket", "liste", "--json"); !strings.Contains(json, `"kennung": "pt-BR"`) {
		t.Errorf("als JSON: %q", json)
	}
}

// **Was nicht passt, ist ein Bedienfehler; was der Dienst abweist, ein
// fachlicher** — und keiner schreibt etwas.
func TestDieKommandozeileWeistUnbrauchbareSprachbefehleAb(t *testing.T) {
	ort := frischerBestand(t)

	for _, fall := range []struct {
		argumente []string
		code      int
	}{
		{[]string{"sprachpaket", "gibtsnicht"}, CodeBedienung},
		{[]string{"sprachpaket", "anlegen", "pt-BR"}, CodeBedienung},
		{[]string{"sprachpaket", "zeige"}, CodeBedienung},
		{[]string{"sprachpaket", "setze"}, CodeBedienung},
		{[]string{"sprachpaket", "setze", "pt-BR", "ohneGleich"}, CodeBedienung},
		{[]string{"sprachpaket", "setze", "pt-BR", "--farbe=rot"}, CodeBedienung},
		{[]string{"sprachpaket", "loeschen"}, CodeBedienung},
		{[]string{"sprachpaket", "ausgeben"}, CodeBedienung},
		{[]string{"sprachpaket", "einlesen"}, CodeBedienung},
		{[]string{"sprachpaket", "einlesen", "--ersetzen"}, CodeBedienung},
		{[]string{"sprachpaket", "anlegen", "ja", "Japanisch", "en"}, CodeFachlich},
		{[]string{"sprachpaket", "zeige", "xx"}, CodeFachlich},
		{[]string{"sprachpaket", "setze", "xx", "--stand=entwurf"}, CodeFachlich},
		{[]string{"sprachpaket", "loeschen", "xx"}, CodeFachlich},
		{[]string{"sprachpaket", "ausgeben", "xx"}, CodeFachlich},
		{[]string{"sprachpaket", "ausgeben", "xx", filepath.Join(t.TempDir(), "x.json")}, CodeFachlich},
		{[]string{"sprachpaket", "einlesen", filepath.Join(t.TempDir(), "fehlt.json")}, CodeFachlich},
	} {
		if code, _, _ := fahreAn(t, ort, fall.argumente...); code != fall.code {
			t.Errorf("%v gab %d, erwartet %d", fall.argumente, code, fall.code)
		}
	}
}

// **Ein Bestand, der die Liste nicht hergibt, ist ein fachlicher Fehler** —
// und keine leere Liste.
func TestEineUnlesbareSprachlisteIstEinFehler(t *testing.T) {
	ort := frischerBestand(t)
	mussLaufen(t, ort, "sprachpaket")

	db, err := storage.Oeffnen(probekontext(t), ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if _, err := db.ExecContext(probekontext(t), `DROP TABLE sprachpakete`); err != nil {
		t.Fatal(err)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	if code, _, _ := fahreAn(t, ort, "sprachpaket"); code != CodeFachlich {
		t.Errorf("gab %d, erwartet %d", code, CodeFachlich)
	}
}
