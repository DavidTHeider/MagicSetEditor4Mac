//go:build fremdpakete

// Der Rundlauf über **bezogene** Pakete.
//
// # Warum das hinter einem Build-Tag steht
//
// `testdaten/` enthält keine fremden Vorlagen (§11,
// [0005](../../docs/entscheidungen/0005-keine-fremden-vorlagen.md)): sie stehen
// unter keiner Lizenz und enthalten geschützte Grafiken. Unsere Testpakete sind
// **nachgebaut** — sie prüfen, dass der Leser tut, was wir für richtig halten,
// nicht, dass er echte Pakete liest.
//
// Dieser Lauf schließt die Lücke, ohne sie ins Repository zu holen. Er läuft
// nicht in der Pipeline, zählt nicht zur Abdeckung und wird von Hand angestoßen:
//
//	MSE_FREMDPAKETE=~/MagicSetEditorPacks/Full-Magic-Pack/data \
//	  go test -tags fremdpakete ./internal/msepaket/ -run Fremdpakete -v
//
// **Vor jeder Freigabe** (`docs/RELEASE.md`): was hier rot wird, ist ein
// Rundlauffehler an echten Daten — und genau der ist der Grund für die Zusage.
package msepaket_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestDerRundlaufUeberFremdpakete(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt — kein Verzeichnis mit bezogenen Paketen")
	}

	eintraege, err := os.ReadDir(wurzel)
	if err != nil {
		t.Fatalf("%s nicht lesbar: %v", wurzel, err)
	}

	geprueft, dateien, uebersprungen, keinPaket := 0, 0, 0, 0

	for _, eintrag := range eintraege {
		// **Die Endung entscheidet, und zwar die letzte.**
		//
		// Der erste Anlauf fragte `strings.Contains(name, ".mse-")`. Neben
		// `magic-mana-small-border.mse-symbol-font` liegt im Pack ein
		// `…​.mse-symbol-font.new` — das Arbeitsstück von jemandem, kein Paket.
		// Es kam durch und ließ den Lauf an einer Meldung scheitern, die
		// stimmte: „trägt keine bekannte Paketendung".
		//
		// Gefragt wird über [msepaket.ArtAusName] und nicht über eine eigene
		// Liste: zwei Aufzählungen derselben sieben Endungen laufen an dem Tag
		// auseinander, an dem eine achte dazukommt.
		if msepaket.ArtAusName(eintrag.Name()) == msepaket.ArtUnbekannt {
			keinPaket++

			continue
		}

		geprueft++

		t.Run(eintrag.Name(), func(t *testing.T) {
			paket, err := msepaket.Oeffne(filepath.Join(wurzel, eintrag.Name()))
			if err != nil {
				t.Fatalf("öffnen: %v", err)
			}

			t.Cleanup(func() {
				if fehler := paket.Schliessen(); fehler != nil {
					t.Errorf("schließen: %v", fehler)
				}
			})

			namen, err := paket.Dateien()
			if err != nil {
				t.Fatalf("aufzählen: %v", err)
			}

			for _, datei := range namen {
				if !istMSEDatei(datei) {
					uebersprungen++

					continue
				}

				roh, err := paket.Datei(datei)
				if err != nil {
					t.Errorf("%s: lesen: %v", datei, err)

					continue
				}

				if !istText(roh) {
					uebersprungen++

					continue
				}

				baum, err := msepaket.Zerteile(datei, roh)
				if err != nil {
					t.Errorf("%s: zerteilen: %v", datei, err)

					continue
				}

				dateien++

				if gebaut := msepaket.Schreibe(baum); !bytes.Equal(gebaut, roh) {
					t.Errorf("der Rundlauf änderte %s/%s:\n%s",
						eintrag.Name(), datei, ersteAbweichung(roh, gebaut))
				}
			}
		})
	}

	if geprueft == 0 {
		t.Fatalf("in %s liegt kein einziges Paket — der Lauf greift ins Leere", wurzel)
	}

	// **Und keine einzige Textdatei ist dasselbe wie kein Paket.**
	//
	// Der erste Anlauf zählte nur die Pakete. Wären alle Dateien übersprungen
	// worden, hätte er „0 Textdateien geprüft" gemeldet und wäre grün gewesen —
	// ein Lauf, der nichts tut und es sagt, ist trotzdem einer, der nichts tut.
	if dateien == 0 {
		t.Fatalf("%d Pakete, aber keine einzige Textdatei geprüft (%d übersprungen)",
			geprueft, uebersprungen)
	}

	// **Was übersprungen wurde, steht dabei.** Eine Zahl, die nur die
	// geprüften nennt, liest sich wie Vollständigkeit.
	t.Logf("%d Pakete, %d Textdateien geprüft, %d übersprungen (keine Textdateien), "+
		"%d Einträge ohne Paketendung", geprueft, dateien, uebersprungen, keinPaket)
}

// istMSEDatei sagt, ob eine Datei im Paket überhaupt eines unserer Formate ist.
//
// **MSE-Datendateien tragen keine Endung** — `set`, `style`, `game`, `scripts`,
// `masks`. Was eine hat, ist Beiwerk: `README.txt`, `LICENSE`, `desktop.ini`,
// Bilder. Der Rundlauf zerteilte bisher jede Datei und meldete für ein
// Windows-`desktop.ini`, dass die erste Zeile keinen Doppelpunkt trägt — ein
// Befund über eine Datei, die uns nichts angeht.
func istMSEDatei(name string) bool {
	// Beiwerk trägt seinen Namen in Großbuchstaben oder eine Endung.
	nurName := filepath.Base(name)
	if nurName == strings.ToUpper(nurName) {
		return false
	}

	return filepath.Ext(name) == ""
}

// **Die Feststellungen über Stile, geprüft an echten Stilen.**
//
// `docs/FORMATE.md` sagt, wie ein Stil seine Zeichenanweisungen ausdrückt, und
// nennt als Quelle zwei Dateien aus `M15-Magic-Pack`. Dieser Lauf hält die
// Feststellung gegen **alle** Stile eines bezogenen Packs — er ist der einzige
// Ort, an dem die Aussage über ihre Quelle hinaus geprüft wird.
func TestFremdeStile(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("ohne MSE_FREMDPAKETE")
	}

	eintraege, err := os.ReadDir(wurzel)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	var (
		stile    int
		ohneMass int
	)

	for _, eintrag := range eintraege {
		if !strings.HasSuffix(eintrag.Name(), ".mse-style") {
			continue
		}

		paket, fehler := msepaket.Oeffne(filepath.Join(wurzel, eintrag.Name()))
		if fehler != nil {
			t.Errorf("%s öffnen: %v", eintrag.Name(), fehler)

			continue
		}

		stil, befunde := paket.Stilmodell()

		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("%s schließen: %v", eintrag.Name(), fehler)
		}

		stile++

		if len(befunde) > 0 {
			t.Errorf("%s: %v", eintrag.Name(), befunde)
		}

		// **Die Kartengröße samt Auflösung ist die Feststellung, die trägt.**
		// Ohne sie gibt es keine physische Größe — und ein Stil ohne Felder ist
		// keiner.
		if stil.Breite == 0 || stil.Hoehe == 0 || stil.DPI == 0 {
			ohneMass++

			t.Logf("%s: %d × %d bei %d dpi", eintrag.Name(), stil.Breite, stil.Hoehe, stil.DPI)
		}

		if len(stil.Felder) == 0 {
			t.Errorf("%s führt keinen Feldstil", eintrag.Name())
		}
	}

	if stile < 10 {
		t.Errorf("nur %d Stile geprüft — greift der Lauf ins Leere?", stile)
	}

	t.Logf("%d Stile gelesen, %d ohne vollständiges Kartenmaß", stile, ohneMass)
}

// **Die Wortlisten des echten Spielpakets** —
// #1074.
//
// Ein nachgebauter Block prüft, dass der Leser tut, was wir für richtig halten.
// Er prüft **nicht**, dass er die 51 Listen des Packs wirklich bekommt — und
// genau daran ist bei den Schlüsselwörtern eine ganze Kette gescheitert.
//
// Die Zahlen sind am 2026-08-24 am bezogenen `M15-Magic-Pack` gezählt:
//
//	grep -c '^word list:' word_lists   → 51
//	grep -c '^\tword:'    word_lists   → 432
//	grep -c '^\t\tscript:' word_lists  → 77
func TestFremdeWortlisten(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	paket, err := msepaket.Oeffne(filepath.Join(wurzel, "magic.mse-game"))
	if err != nil {
		t.Fatalf("magic.mse-game öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	// **Ohne Einbinden sind es null.** `magic.mse-game/game` holt sie über
	// `include file: word_lists` (Zeile 29); wer nur den Kopf liest, sieht
	// keine einzige Liste — dieselbe Falle, die bei den Schlüsselwörtern eine
	// ganze Kette hat ins Leere laufen lassen
	// (#1064).
	if befunde := paket.ModuleEinbinden(msepaket.NeuerPaketleser(wurzel)); len(befunde) > 0 {
		t.Logf("%d Befunde beim Einbinden", len(befunde))
	}

	listen := paket.Wortlisten()

	woerter, skripte, praefixe, gruppen := 0, 0, 0, 0

	for _, liste := range listen {
		woerter += len(liste.Woerter)

		for _, wort := range liste.Woerter {
			if wort.Skript != "" {
				skripte++
			}

			if wort.IstPraefix {
				praefixe++
			}

			if wort.IstGruppe() {
				gruppen++
			}
		}
	}

	t.Logf("%d Listen, %d Wörter, %d mit Skript, %d Präfixe, %d Gruppen",
		len(listen), woerter, skripte, praefixe, gruppen)

	if len(listen) != 51 {
		t.Errorf("%d Wortlisten, gezählt sind 51", len(listen))
	}

	if woerter != 432 {
		t.Errorf("%d Wörter, gezählt sind 432", woerter)
	}

	if skripte != 77 {
		t.Errorf("%d Wörter mit Skript, gezählt sind 77", skripte)
	}
}
