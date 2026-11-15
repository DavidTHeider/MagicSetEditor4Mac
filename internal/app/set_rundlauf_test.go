package app_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Der Rundlauf über den echten Bestand: `.mse-set` → SQLite → `.mse-set`.**
//
// # Warum dieser Test hier steht und nicht in `internal/msepaket`
//
// Der Rundlauf der Fachbibliothek (#128) prüft `Schreibe(Zerteile(x))` im
// Arbeitsspeicher — und genau **dazwischen** liegt die Frage, die
// #48 stellt:
// aus `card_values` kommen die Werte ohne ihre Rohform zurück, Einrückung und
// Abstände stehen dort nirgends.
//
// Dieser Weg geht durch **beide** Schichten und durch die Oberfläche: öffnen,
// wie es ein Mensch tut, speichern, wie es ein Mensch tut, vergleichen. Das ist
// die letzte offene Zusage von #48, und sie ist nur hier zu halten.
func TestEinDurchgereichtesSetIstByteFuerByteDasselbe(t *testing.T) {
	// **Mehr als ein Paket, und eines davon eigenwillig.** Die erste Fassung
	// prüfte nur `probe.mse-set` — und blieb grün, als die Gegenprobe **alle**
	// Rohzeilen der Karten verwarf: `\tname: Feuerball` setzt sich zufällig
	// genau so wieder zusammen. Ein Prüfstück, das seinen Fall nicht enthält,
	// prüft ihn nicht (§10).
	for _, paket := range []string{"probe.mse-set", "eigenwillig.mse-set", "karg.mse-set"} {
		t.Run(paket, func(t *testing.T) {
			anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

			if err := anwendung.Starten(t.Context()); err != nil {
				t.Fatalf("starten: %v", err)
			}

			t.Cleanup(func() {
				if fehler := anwendung.Beenden(); fehler != nil {
					t.Errorf("beenden: %v", fehler)
				}
			})

			quelle, err := filepath.Abs(filepath.Join("..", "..", "testdaten", paket))
			if err != nil {
				t.Fatalf("Pfad: %v", err)
			}

			geoeffnet := anwendung.SetOeffnen(t.Context(), quelle)
			if geoeffnet.Fehler != nil {
				t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
			}

			ziel := filepath.Join(t.TempDir(), "wieder.mse-set")

			gespeichert := anwendung.SetSpeichern(t.Context(), geoeffnet.SetID, ziel)
			if gespeichert.Fehler != nil {
				t.Fatalf("speichern: %+v", gespeichert.Fehler)
			}

			if len(gespeichert.Befunde) != 0 {
				t.Errorf("Befunde: %v, erwartet keine", gespeichert.Befunde)
			}

			vorher := lies(t, filepath.Join(quelle, "set"))
			nachher := lies(t, filepath.Join(ziel, "set"))

			if !bytes.Equal(nachher, vorher) {
				t.Errorf("die Datei ist eine andere:\n%s", unterschied(vorher, nachher))
			}
		})
	}
}

// **Ohne Ziel geht das Speichern dorthin, wo das Set herkommt.**
//
// Das ist ⌘S: kein Dialog, kein neuer Ort. Wer stattdessen jedes Mal fragte,
// hätte kein Speichern, sondern ein „Speichern unter".
func TestOhneZielWirdUeberDieHerkunftGeschrieben(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	// **Eine Kopie**, nicht das Prüfstück selbst: ein Test, der `testdaten/`
	// beschreibt, ändert die Grundlage aller anderen.
	quelle := kopiereVerzeichnis(t, filepath.Join("..", "..", "testdaten", "probe.mse-set"))

	geoeffnet := anwendung.SetOeffnen(t.Context(), quelle)
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	gespeichert := anwendung.SetSpeichern(t.Context(), geoeffnet.SetID, "")
	if gespeichert.Fehler != nil {
		t.Fatalf("speichern: %+v", gespeichert.Fehler)
	}

	if gespeichert.Pfad != quelle {
		t.Errorf("geschrieben nach %q, erwartet %q", gespeichert.Pfad, quelle)
	}
}

// lies gibt eine Datei oder bricht ab.
func lies(t *testing.T, pfad string) []byte {
	t.Helper()

	inhalt, err := os.ReadFile(pfad) // #nosec G304 -- Prüfstück beziehungsweise t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	return inhalt
}

// kopiereVerzeichnis legt ein Paketverzeichnis unter einem eigenen Ort ab.
func kopiereVerzeichnis(t *testing.T, herkunft string) string {
	t.Helper()

	ziel := filepath.Join(t.TempDir(), filepath.Base(herkunft))
	if err := os.MkdirAll(ziel, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	eintraege, err := os.ReadDir(herkunft)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	for _, eintrag := range eintraege {
		inhalt := lies(t, filepath.Join(herkunft, eintrag.Name()))

		if fehler := os.WriteFile(filepath.Join(ziel, eintrag.Name()), inhalt, 0o600); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}
	}

	ort, err := filepath.Abs(ziel)
	if err != nil {
		t.Fatalf("Pfad: %v", err)
	}

	return ort
}

// unterschied nennt die erste abweichende Zeile, mit sichtbarem `\r` und BOM.
//
// **Nicht beide Dateien.** Ein Vergleich, der zweimal siebenhundert Zeilen
// ausgibt, wird nicht gelesen; die eine Zeile, an der es auseinandergeht, wird
// gelesen. Und ohne das Sichtbarmachen stünde zweimal derselbe Text da — der
// Unterschied wäre ein `\r`.
func unterschied(vorher, nachher []byte) string {
	sichtbar := strings.NewReplacer("\t", "\\t", "\r", "\\r", "\n", "\\n", "\ufeff", "<BOM>")

	alt := strings.SplitAfter(string(vorher), "\n")
	neu := strings.SplitAfter(string(nachher), "\n")

	for i := range max(len(alt), len(neu)) {
		links, rechts := "", ""
		if i < len(alt) {
			links = alt[i]
		}

		if i < len(neu) {
			rechts = neu[i]
		}

		if links != rechts {
			return "Zeile " + strconv.Itoa(i+1) +
				":\n  gelesen:      " + sichtbar.Replace(links) +
				"\n  geschrieben:  " + sichtbar.Replace(rechts)
		}
	}

	return "die Zeilen sind gleich — der Unterschied liegt am Dateiende"
}
