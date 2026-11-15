package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die drei Funktionen, die selbst schreiben** —
// #657.
//
// `magic-spoiler` ruft alle drei; `magic-cockatrice-v2`, `magic-lackeyccg` und
// `magic-planesculptors` je `write_text_file`. Bis hierher scheiterten sie an
// „der Name … ist hier nicht bekannt".
type beigabenmitschrift struct {
	// gelegt hält je Datei ihren Inhalt.
	gelegt map[string]string

	// gelesen hält die Namen, die `copy_file` aus dem Paket geholt hat.
	gelesen []string

	// sets hält die Namen, unter denen ein Set abgelegt wurde.
	sets []string

	lesefehler   error
	schreibfehte error
}

func (b *beigabenmitschrift) AusVorlage(name string) ([]byte, error) {
	if b.lesefehler != nil {
		return nil, b.lesefehler
	}

	b.gelesen = append(b.gelesen, name)

	return []byte("aus dem Paket"), nil
}

func (b *beigabenmitschrift) Lege(datei string, inhalt []byte) (string, error) {
	if b.schreibfehte != nil {
		return "", b.schreibfehte
	}

	if b.gelegt == nil {
		b.gelegt = map[string]string{}
	}

	b.gelegt[datei] = string(inhalt)

	return datei, nil
}

func (b *beigabenmitschrift) LegeSet(_ context.Context, datei string) (string, error) {
	if b.schreibfehte != nil {
		return "", b.schreibfehte
	}

	b.sets = append(b.sets, datei)

	return datei, nil
}

func TestDieDreiSchreibendenFunktionen(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name    string
		skript  string
		ausgabe string
		pruefe  func(t *testing.T, m *beigabenmitschrift)
	}{
		{
			name:    "copy_file holt aus dem Vorlagenpaket",
			skript:  `copy_file("stil.css")`,
			ausgabe: "stil.css",
			pruefe: func(t *testing.T, m *beigabenmitschrift) {
				t.Helper()

				if strings.Join(m.gelesen, "|") != "stil.css" {
					t.Errorf("gelesen: %q", m.gelesen)
				}

				if m.gelegt["stil.css"] != "aus dem Paket" {
					t.Errorf("gelegt: %q", m.gelegt)
				}
			},
		},
		{
			name:    "write_text_file schreibt mit Byte-Order-Mark",
			skript:  `write_text_file("eine Notiz", file: "notiz.txt")`,
			ausgabe: "notiz.txt",
			pruefe: func(t *testing.T, m *beigabenmitschrift) {
				t.Helper()

				// **Mit BOM** — `tout.WriteString(BYTE_ORDER_MARK)` steht
				// ausdrücklich in `export.cpp`, Zeile 394.
				if m.gelegt["notiz.txt"] != "\ufeffeine Notiz" {
					t.Errorf("gelegt: %q", m.gelegt["notiz.txt"])
				}
			},
		},
		{
			name:    "write_set_file legt das Set daneben",
			skript:  `write_set_file(file: "kopie.mse-set")`,
			ausgabe: "kopie.mse-set",
			pruefe: func(t *testing.T, m *beigabenmitschrift) {
				t.Helper()

				if strings.Join(m.sets, "|") != "kopie.mse-set" {
					t.Errorf("Sets: %q", m.sets)
				}
			},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			mitschrift := &beigabenmitschrift{}

			ausgabe, _, err := probelauf(service.Ausfuhrmittel{
				Beigaben: mitschrift, Verzeichnis: "ausfuhr-files",
			}).
				Fuehre(t.Context(), vorlageMit(fall.skript), eineKarte(), nil, nil)
			if err != nil {
				t.Fatalf("ausführen: %v", err)
			}

			// **Der Name kommt zurück** — die Vorlagen hängen ihn in ihre
			// Ausgabe.
			if ausgabe != fall.ausgabe {
				t.Errorf("Ausgabe %q, erwartet %q", ausgabe, fall.ausgabe)
			}

			fall.pruefe(t, mitschrift)
		})
	}
}

// **Ohne die Fähigkeit gibt es die Namen nicht.**
//
// Dieselbe Regel wie bei `write_image_file` (#216): eine Vorlage, die sie ruft,
// bekommt „ist hier nicht bekannt" — und nicht eine Datei an einem Ort, den
// niemand gewählt hat.
func TestOhneBeigabenGibtEsDieNamenNicht(t *testing.T) {
	t.Parallel()

	for _, skript := range []string{
		`copy_file("x")`,
		`write_text_file("x", file: "y")`,
		`write_set_file(file: "y")`,
	} {
		t.Run(skript, func(t *testing.T) {
			t.Parallel()

			_, _, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
				Fuehre(t.Context(), vorlageMit(skript), eineKarte(), nil, nil)
			if err == nil {
				t.Fatal("die Vorlage lief ohne die Fähigkeit durch")
			}

			if !strings.Contains(err.Error(), "nicht bekannt") {
				t.Errorf("Meldung: %v", err)
			}
		})
	}
}

// **Ein fehlendes Argument wird benannt, nicht erraten.**
func TestEinFehlendesArgumentWirdBenannt(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, skript string }{
		{"write_text_file ohne file", `write_text_file("x")`},
		{"write_text_file ohne Text", `write_text_file(file: "y")`},
		{"write_set_file ohne file", `write_set_file()`},
		{"copy_file ohne Namen", `copy_file()`},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := probelauf(service.Ausfuhrmittel{
				Beigaben: &beigabenmitschrift{}, Verzeichnis: "ausfuhr-files",
			}).
				Fuehre(t.Context(), vorlageMit(fall.skript), eineKarte(), nil, nil)
			if err == nil {
				t.Fatal("der Aufruf ging durch")
			}
		})
	}
}

// **Ein Fehler beim Lesen oder Schreiben kommt durch.**
//
// Anders als bei einer Karte (§8): eine halbe Ausfuhr ist eine, die jemand für
// vollständig hält.
func TestEinFehlerBeimBeilegenKommtDurch(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name       string
		skript     string
		mitschrift *beigabenmitschrift
	}{
		{
			name:       "copy_file kann nicht lesen",
			skript:     `copy_file("x")`,
			mitschrift: &beigabenmitschrift{lesefehler: errors.New("kein x")},
		},
		{
			name:       "copy_file kann nicht schreiben",
			skript:     `copy_file("x")`,
			mitschrift: &beigabenmitschrift{schreibfehte: errors.New("voll")},
		},
		{
			name:       "write_text_file kann nicht schreiben",
			skript:     `write_text_file("x", file: "y")`,
			mitschrift: &beigabenmitschrift{schreibfehte: errors.New("voll")},
		},
		{
			name:       "write_set_file kann nicht schreiben",
			skript:     `write_set_file(file: "y")`,
			mitschrift: &beigabenmitschrift{schreibfehte: errors.New("voll")},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := probelauf(service.Ausfuhrmittel{
				Beigaben: fall.mitschrift, Verzeichnis: "ausfuhr-files",
			}).
				Fuehre(t.Context(), vorlageMit(fall.skript), eineKarte(), nil, nil)
			if err == nil {
				t.Fatal("der Fehler blieb liegen")
			}
		})
	}
}

// **Und `write_text_file` nimmt genau einen Text.**
func TestWriteTextFileNimmtGenauEinenText(t *testing.T) {
	t.Parallel()

	_, _, err := probelauf(service.Ausfuhrmittel{
		Beigaben: &beigabenmitschrift{}, Verzeichnis: "ausfuhr-files",
	}).
		Fuehre(t.Context(), vorlageMit(`write_text_file("a", "b", file: "y")`),
			eineKarte(), nil, nil)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler = %v, erwartet ErrValidation", err)
	}
}
