// **Die Exportvorlage ohne Fenster** —
// #1291.
package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// vorlagenset legt ein Set mit Karten an und gibt Weg und Setkennung.
//
// **Mit der Saat als Paketwurzel** (§11): `probe-liste.mse-export-template`
// liegt dort und schreibt eine Zeile je Karte — sie ist selbst geschrieben,
// und ihr Ergebnis lässt sich nachlesen.
func vorlagenset(t *testing.T) (weg *app.Vorlagenweg, setID int64) {
	t.Helper()

	weg, setID, _ = vorlagensetMitPaketen(t)

	return weg, setID
}

// vorlagensetMitPaketen gibt dazu das Paketverzeichnis, in das ein Test eine
// eigene Vorlage legen kann.
func vorlagensetMitPaketen(t *testing.T) (weg *app.Vorlagenweg, setID int64, pakete string) {
	t.Helper()

	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	quelle, err := filepath.Abs(filepath.Join("..", "..", "saat"))
	if err != nil {
		t.Fatalf("Saatpfad: %v", err)
	}

	ziel := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac", "pakete")
	if kopierfehler := os.CopyFS(ziel, os.DirFS(quelle)); kopierfehler != nil {
		t.Fatalf("Saat kopieren: %v", kopierfehler)
	}

	db := offeneProbedatenbank(t)

	kennung, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Vorlagenset", Game: "probe", Style: "probe-schlicht",
		Cards: []*domain.Card{
			{Values: []domain.Value{{Field: "name", Text: "Erste"}}},
			{Values: []domain.Value{{Field: "name", Text: "Zweite"}}},
		},
	}, nil)
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	weg, nahtfehler := app.VorlagenwegUeber(db, domain.UrheberKommandozeile)
	if nahtfehler != nil {
		t.Fatalf("Vorlagenweg: %v", nahtfehler)
	}

	return weg, kennung, ziel
}

// **Die Vorlagen des Sets werden genannt.**
func TestDerVorlagenwegNenntDieVorlagen(t *testing.T) {
	weg, setID := vorlagenset(t)

	vorlagen, _, fehler := weg.Vorlagen(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("Vorlagen: %v", fehler)
	}

	if len(vorlagen) == 0 {
		t.Fatal("keine Vorlage")
	}

	var gefunden bool

	for _, vorlage := range vorlagen {
		if strings.Contains(vorlage.Paket, "probe-liste") {
			gefunden = true
		}
	}

	if !gefunden {
		t.Errorf("probe-liste fehlt: %+v", vorlagen)
	}
}

// **Und sie läuft — das Ergebnis steht in der Datei.**
func TestDerVorlagenwegSchreibtDasErgebnis(t *testing.T) {
	weg, setID := vorlagenset(t)

	ziel := filepath.Join(t.TempDir(), "ausfuhr.txt")

	lauf, fehler := weg.Ausfuehren(probekontext(t), setID, "probe-liste.mse-export-template",
		ziel, nil)
	if fehler != nil {
		t.Fatalf("ausführen: %v (%v)", fehler, lauf.Befunde)
	}

	if lauf.Pfad != ziel {
		t.Errorf("der Pfad ist %q", lauf.Pfad)
	}

	roh, err := os.ReadFile(ziel) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("die Datei fehlt: %v", err)
	}

	// **Beide Karten stehen darin** — eine Vorlage, die nur die erste
	// schriebe, gäbe auch eine Datei.
	for _, name := range []string{"Erste", "Zweite"} {
		if !strings.Contains(string(roh), name) {
			t.Errorf("%q fehlt in der Ausgabe: %s", name, roh)
		}
	}
}

// **Eine Vorlage, die es nicht gibt, ist ein Nicht-Gefunden.**
func TestDerVorlagenwegWeistEineUnbekannteVorlageAb(t *testing.T) {
	weg, setID := vorlagenset(t)

	_, fehler := weg.Ausfuehren(probekontext(t), setID, "gibtsnicht.mse-export-template",
		filepath.Join(t.TempDir(), "raus.txt"), nil)
	if fehler == nil {
		t.Error("eine unbekannte Vorlage ging durch")
	}
}

// **Eine Option lässt sich setzen und kommt bei der nächsten Frage zurück.**
func TestEineGesetzteOptionKommtZurueck(t *testing.T) {
	weg, setID := vorlagenset(t)

	vorlagen, _, fehler := weg.Vorlagen(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("Vorlagen: %v", fehler)
	}

	var (
		paket string
		feld  string
	)

	for _, vorlage := range vorlagen {
		if len(vorlage.Optionen) > 0 {
			paket, feld = vorlage.Paket, vorlage.Optionen[0].Name

			break
		}
	}

	if paket == "" {
		t.Skip("keine Vorlage der Saat führt eine Option")
	}

	if err := weg.OptionSetzen(probekontext(t), paket, feld, "geändert"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	danach, _, fehler := weg.Vorlagen(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("Vorlagen danach: %v", fehler)
	}

	for _, vorlage := range danach {
		if vorlage.Paket != paket {
			continue
		}

		for _, option := range vorlage.Optionen {
			if option.Name == feld && option.Wert != "geändert" {
				t.Errorf("der Wert ist %q", option.Wert)
			}
		}
	}
}

// **Ohne Benutzerverzeichnis kommt der Vorlagenweg nicht zustande.**
func TestOhneBenutzerverzeichnisGibtEsKeinenVorlagenweg(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	if _, err := app.VorlagenwegUeber(nil, domain.UrheberKommandozeile); err == nil {
		t.Error("kein Fehler")
	}
}

// **Eine unbrauchbare Setkennung gibt keine Vorlagenliste.**
//
// Der Fehlerzweig in `Vorlagen`. **Nicht mit einer Kennung, die es nicht
// gibt**: die geht heute durch und gibt eine leere Liste — dieselbe Klasse wie
// beim Druckbogen, und sie steht als
// #1336.
// Ein Lauf, der hier einen Fehler erwartete, prüfte eine Zusage, die es nicht
// gibt.
func TestDerVorlagenwegWeistEineUnbrauchbareKennungAb(t *testing.T) {
	// **Über einen geschlossenen Bestand**: `Vorlagenweg` nimmt eine Zahl,
	// und eine Zahl ist immer wohlgeformt — der Fehlerzweig liegt hinter dem
	// Lesen und nicht davor.
	t.Setenv("HOME", t.TempDir())

	geschlossen := offeneProbedatenbank(t)
	if err := geschlossen.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	zu, nahtfehler := app.VorlagenwegUeber(geschlossen, domain.UrheberKommandozeile)
	if nahtfehler != nil {
		t.Fatalf("Vorlagenweg: %v", nahtfehler)
	}

	if _, _, fehler := zu.Vorlagen(probekontext(t), 1); fehler == nil {
		t.Error("ein geschlossener Bestand ging durch")
	}
}

// **Die Ausfuhr ohne Fenster kennt den Stil des Sets** — gefunden in VK-01b
// (#2182).
// Der Vorlagenweg von Kommandozeile, REST und MCP setzte keine Stilwahl, und
// `set.stylesheet` war dort `nil`, während das Fenster ihn kannte (#2039).
// Gemessen am kanonischen `Planechase.mse-set`: 86 Karten, `set.stylesheet`
// `nil` an der Kommandozeile.
func TestDerVorlagenwegKenntDenStilDesSets(t *testing.T) {
	weg, setID, pakete := vorlagensetMitPaketen(t)

	vorlage := filepath.Join(pakete, "probe-stilname.mse-export-template")
	if err := os.MkdirAll(vorlage, 0o750); err != nil {
		t.Fatalf("Vorlage anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\nshort name: Stilname\ngame: probe\n" +
		"file type: *.txt|*.txt\ncreate directory: false\n\nscript:\n" +
		"\tif set.stylesheet == nil then \"nil\" else set.stylesheet.short_name\n"
	if err := os.WriteFile(filepath.Join(vorlage, "export-template"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("Vorlage schreiben: %v", err)
	}

	ziel := filepath.Join(t.TempDir(), "stil.txt")

	lauf, fehler := weg.Ausfuehren(probekontext(t), setID, "probe-stilname.mse-export-template",
		ziel, nil)
	if fehler != nil {
		t.Fatalf("ausführen: %v (%v)", fehler, lauf.Befunde)
	}

	roh, err := os.ReadFile(ziel) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("die Datei fehlt: %v", err)
	}

	if string(roh) != "Schlicht" {
		t.Errorf("set.stylesheet.short_name ist %q, erwartet Schlicht — Befunde %v", roh, lauf.Befunde)
	}
}
