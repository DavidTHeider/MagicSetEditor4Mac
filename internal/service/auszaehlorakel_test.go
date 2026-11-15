//go:build fremdpakete

package service_test

// **Dieselbe Antwort wie MSE2 — gefragt, nicht geraten** —
// #1338.
//
// # Die Abnahme aus dem Issue
//
// > Für ein Set mit `magic.mse-game`: die Dimension „color category" liefert
// > je Karte denselben Wert wie MSE2 (über `magicseteditor --cli`, das Orakel
// > aus §15).
//
// Genau das tut dieser Lauf. Er fragt das **gebaute MSE2** nach der
// Farbgruppe jeder Karte eines kanonischen Sets, rechnet dieselbe Dimension
// hier, und vergleicht Karte für Karte.
//
// **Warum das Orakel und nicht eine eingecheckte Liste**: eine Liste veraltet
// still, sobald das Pack sich ändert (§10). Das Orakel rechnet mit demselben
// Pack, das hier gelesen wird.
//
//	go test -tags fremdpakete ./internal/service/ -run Orakel -v

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

const (
	orakel  = "/Users/david/mse2-quelle/build/magicseteditor"
	packort = "/Users/david/full-magic-pack/data"
)

// kanonischesSet gibt den Pfad eines kanonischen Sets.
func kanonischesSet(t *testing.T, name string) string {
	t.Helper()

	zuhause, fehler := os.UserHomeDir()
	if fehler != nil {
		t.Fatalf("Zuhause: %v", fehler)
	}

	pfad := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac", "kanonische-sets", name)

	if _, fehler := os.Stat(pfad); fehler != nil {
		t.Skipf("%s liegt nicht da: %v", pfad, fehler)
	}

	return pfad
}

// orakelantwort fragt MSE2 nach der Farbgruppe jeder Karte.
func orakelantwort(t *testing.T, setpfad string) []string {
	t.Helper()

	if _, fehler := os.Stat(orakel); fehler != nil {
		t.Skipf("das Orakel ist nicht gebaut: %v", fehler)
	}

	frist, abbrechen := context.WithTimeout(t.Context(), 10*time.Minute)
	defer abbrechen()

	// #nosec G204 -- der Pfad ist eine Konstante dieser Datei, das Set kommt
	// aus dem Zuhause dieses Rechners, und der Lauf steht hinter `fremdpakete`.
	befehl := exec.CommandContext(frist, orakel, "--cli", setpfad, "-q", "--raw")
	befehl.Stdin = strings.NewReader(
		"n := length(set.cards)-1\n" +
			"for x from 0 to n do [ color_category_statistic(card: set.cards[x]) ]\n")

	roh, fehler := befehl.Output()
	if fehler != nil {
		t.Fatalf("das Orakel: %v", fehler)
	}

	// **Die letzte nichtleere Zeile ist die Antwort.** Davor stehen die
	// Fortschrittszahlen, die `--raw` je Ausdruck ausgibt.
	zeilen := strings.Split(strings.TrimSpace(string(roh)), "\n")

	var antwort []string
	if fehler := json.Unmarshal(
		[]byte(zeilen[len(zeilen)-1]), &antwort); fehler != nil {
		t.Fatalf("die Antwort des Orakels ist keine Liste: %v", fehler)
	}

	return antwort
}

// kartenUndAngabenAusSet liest die `card:`-Blöcke und den `set info:`-Block.
func kartenUndAngabenAusSet(t *testing.T, setpfad string) (
	[]*domain.CardRow, []domain.Value,
) {
	t.Helper()

	paket, fehler := msepaket.Oeffne(setpfad)
	if fehler != nil {
		t.Fatalf("%s öffnen: %v", setpfad, fehler)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	bloecke := paket.Kopf.Alle("card")

	karten := make([]*domain.CardRow, 0, len(bloecke))

	for stelle, block := range bloecke {
		werte := make([]domain.Value, 0, len(block.Kinder))

		for _, zeile := range block.Kinder {
			if zeile.Schluessel == "" {
				continue
			}

			// **Der Feldname mit Leerzeichen und nicht mit Unterstrich.**
			// In der Datei steht `styling_data:`, im Bestand `styling data`
			// — `service.Einfahren` übersetzt das, und ohne diesen Schritt
			// fände `kartenstilwerte` die Stilwerte der Karte nicht.
			werte = append(werte, domain.Value{
				Field: msepaket.Entkanonisch(zeile.Schluessel),
				Text:  blockwert(zeile),
			})
		}

		karten = append(karten, &domain.CardRow{
			ID: int64(stelle + 1), SetID: 1, Number: stelle, Values: werte,
		})
	}

	angaben := []domain.Value{}
	if block := paket.Kopf.Ersten("set info"); block != nil {
		for _, zeile := range block.Kinder {
			if zeile.Schluessel == "" {
				continue
			}

			angaben = append(angaben, domain.Value{
				Field: zeile.Schluessel, Text: zeile.Wert,
			})
		}
	}

	return karten, angaben
}

// **Karte für Karte dieselbe Farbgruppe wie MSE2.**
func TestDieFarbgruppeStimmtMitDemOrakelUeberein(t *testing.T) {
	setpfad := kanonischesSet(t, "White.mse-set")

	paket, fehler := msepaket.Oeffne(filepath.Join(packort, "magic.mse-game"))
	if fehler != nil {
		t.Fatalf("magic.mse-game öffnen: %v", fehler)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if fehler := paket.ModuleEinbinden(
		msepaket.NeuerPaketleser(packort)); fehler != nil {
		t.Fatalf("Module einbinden: %v", fehler)
	}

	spiel, _ := paket.Spielmodell()

	var farbgruppe msepaket.Auszaehldimension

	for _, eine := range paket.Auszaehldimensionen() {
		if eine.Name == "color category" {
			farbgruppe = eine
		}
	}

	karten, setangaben := kartenUndAngabenAusSet(t, setpfad)
	vonMSE := orakelantwort(t, setpfad)

	if len(karten) != len(vonMSE) {
		t.Fatalf("%d Karten hier, %d beim Orakel", len(karten), len(vonMSE))
	}

	// **Ein echter Stilspeicher, kein Doppel.** Nur er kennt den Weg von
	// `card.stylesheet` zum Stilpaket — und genau davon hängt `card.shape` ab.
	speicher := service.NeuerStilspeicher(&setkopfprobe{
		kopf:  domain.SetHead{ID: 1, Game: "magic", Style: "m15-altered"},
		werte: setangaben,
	}, packort)

	begonnen := time.Now()

	// **Ohne `split list:` und mit `show empty:`** — verglichen wird der
	// **rohe** Wert des Skripts, denn das ist es, was das Orakel gibt. Das
	// Zerlegen und das Weglassen prüft `auszaehlung_test.go` in `domain`.
	angabe := service.Auszaehlangabe{
		Name: farbgruppe.Name, Skript: farbgruppe.Skript, LeereZeigen: true,
	}

	vorlage := magicvorlage(spiel)

	// **Die Setangaben gehören dazu.** `check_2_statistic` liest
	// `set.count_secondary_faces` (`statistics_script`, Zeile 1195) — eine
	// **Setangabe**. Ohne sie sagte die Dimension bei jeder zweiseitigen
	// Karte `White` statt `White,White`: 21 von 3.761 Karten, gemessen am
	// 2026-08-31.
	vorlage.Setangaben = setangaben

	werte, befunde := service.Auszaehlwerte(
		t.Context(), karten, vorlage,
		func(karte *domain.CardRow) service.Vorlage {
			eigene, _, da := speicher.StilFuerKarte(t.Context(), karte)
			if !da {
				return vorlage
			}

			// **Das Spielskript und die Setfelder kommen aus dem Spiel** —
			// der Stilspeicher hat sie hier nicht, weil dieser Lauf das
			// Spielpaket selbst geöffnet hat.
			eigene.Spielskript = vorlage.Spielskript
			eigene.Setfelder = vorlage.Setfelder
			eigene.Kartenfelder = vorlage.Kartenfelder
			eigene.Setangaben = setangaben

			return eigene
		}, angabe)

	t.Logf("%d Karten in %s gerechnet", len(karten), time.Since(begonnen).Round(time.Millisecond))

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	if len(werte) != len(vonMSE) {
		t.Fatalf("%d Werte hier, %d beim Orakel", len(werte), len(vonMSE))
	}

	abweichungen := 0

	for stelle := range werte {
		if werte[stelle] != vonMSE[stelle] {
			abweichungen++

			if abweichungen <= 10 {
				t.Errorf("Karte %d: hier %q, MSE2 %q",
					stelle, werte[stelle], vonMSE[stelle])
			}
		}
	}

	if abweichungen > 0 {
		t.Errorf("%d von %d Karten weichen ab", abweichungen, len(werte))
	}
}

// setkopfprobe ist die Kopfquelle, die ein Stilspeicher braucht.
type setkopfprobe struct {
	kopf  domain.SetHead
	werte []domain.Value
}

func (s *setkopfprobe) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return []domain.SetHead{s.kopf}, nil
}

func (s *setkopfprobe) Setwerte(context.Context, int64) ([]domain.Value, error) {
	return s.werte, nil
}

// blockwert gibt den Text eines Eintrags — auch wenn er ein Block ist.
//
// **`styling_data:` ist ein Block**, und sein Inhalt entscheidet über die
// Gestalt der Karte: `dfc_type: modal with standard flags` macht aus einer
// zweiseitigen Karte eine modale. Wer nur den Wert der Kopfzeile läse, bekäme
// die leere Zeichenkette und damit die Vorgabe des Stils — gemessen am
// 2026-08-31: acht von 3.761 Karten wichen deshalb vom Orakel ab.
//
// **Der Betrieb liest das richtig** (`service.Einfahren`); nur diese
// Vorrichtung war zu einfach. Eine Vorrichtung, die weniger kann als der
// Betrieb, misst nicht den Betrieb (§3).
func blockwert(zeile *msepaket.Knoten) string {
	if zeile.Wert != "" || len(zeile.Kinder) == 0 {
		return zeile.Wert
	}

	zeilen := make([]string, 0, len(zeile.Kinder))
	for _, kind := range zeile.Kinder {
		if kind.Schluessel == "" {
			continue
		}

		zeilen = append(zeilen, kind.Schluessel+": "+kind.Wert)
	}

	return strings.Join(zeilen, "\n")
}
