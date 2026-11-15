//go:build fremdpakete

// **Der Referenzbestand: woraus eine Referenz besteht, und wogegen gemessen wird.**
//
// Getrennt von der Messung selbst (`mse2goldbild_test.go`), weil es eine andere
// Frage ist: dort steht, *wie* verglichen wird — hier, was eine Referenz
// ausmacht und wie man ihr ansieht, dass sie sich geändert hat.
package app_test

import (
	"crypto/sha256"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// bildfensterAus liest `bildfenster: links oben breite hoehe` aus der Referenz.
//
// **Die Referenz sagt, wo ihre Illustration liegt.** Sie ist ohnehin von Hand
// beschrieben, und wo das Bild sitzt, gehört zu ihrer Beschreibung. Eine
// Heuristik über den Zeichenplan wäre eine zweite, **wandernde** Wahrheit —
// warum, steht im Kopf von `mse2goldbild_test.go`.
//
// Fehlt die Angabe, wird **nichts** ausgespart und beide Zahlen sind gleich.
// Das ist kein Fehler: eine Referenz ohne Illustration gibt es.
func bildfensterAus(t *testing.T, knoten *msepaket.Knoten) image.Rectangle {
	t.Helper()

	if knoten == nil {
		return image.Rectangle{}
	}

	var links, oben, breite, hoehe int
	if _, err := fmt.Sscanf(knoten.Wert, "%d %d %d %d", &links, &oben, &breite, &hoehe); err != nil {
		t.Fatalf("`bildfenster:` erwartet vier Zahlen, hier steht %q: %v", knoten.Wert, err)
	}

	if breite <= 0 || hoehe <= 0 {
		t.Fatalf("`bildfenster:` hat die Größe %dx%d", breite, hoehe)
	}

	return image.Rect(links, oben, links+breite, oben+hoehe)
}

// referenzkarte liest die Kartendaten neben dem Referenzbild.
//
// **In MSEs eigener Syntax**, gelesen mit unserem eigenen Zerteiler. Ein
// eigenes Format daneben wäre eine zweite Sprache, die niemand prüft — und der
// Zerteiler ist ohnehin die Stelle, die das Format beherrscht.
func referenzkarte(t *testing.T, pfad string) (*domain.CardRow, string, *msepaket.Baum) {
	t.Helper()

	inhalt, err := os.ReadFile(pfad) //nolint:gosec // ein Pfad aus einer Umgebungsvariable des Laufs
	if err != nil {
		t.Fatalf("Kartendaten lesen: %v", err)
	}

	baum, err := msepaket.Zerteile(filepath.Base(pfad), inhalt)
	if err != nil {
		t.Fatalf("Kartendaten zerteilen: %v", err)
	}

	stil := ""
	if knoten := baum.Ersten("stil"); knoten != nil {
		stil = knoten.Wert
	}

	if stil == "" {
		t.Fatalf("%s nennt keinen `stil:`", pfad)
	}

	karte := &domain.CardRow{ID: 1, Number: 1}

	for _, knoten := range baum.Wurzel {
		if knoten.Zeilenart != msepaket.ArtEintrag ||
			knoten.Schluessel == "stil" || knoten.Schluessel == "bildfenster" {
			continue
		}

		karte.Values = append(karte.Values, domain.Value{
			Field: knoten.Schluessel, Text: feldwert(knoten),
		})
	}

	return karte, stil, baum
}

// feldwert liest den Wert eines Feldes — einzeilig oder mehrzeilig.
//
// **[msepaket.Knoten.Skripttext] gibt nur die Kinder**, nicht den eigenen Wert.
// Für `name: Feuerball` ergäbe das die leere Zeichenkette, und die Karte
// zeichnete keinen Namen — ohne einen einzigen Befund, denn ein leeres Feld ist
// kein Fehler. Genau das ist beim ersten Lauf passiert.
func feldwert(knoten *msepaket.Knoten) string {
	if len(knoten.Kinder) == 0 {
		return knoten.Wert
	}

	return knoten.Skripttext()
}

// bestandsausweis protokolliert, **wogegen** gemessen wurde.
//
// # Warum eine Zahl ohne ihren Bestand nichts wert ist
//
// Dieser Lauf ist ein **Verlaufsmaß**: die Phasen B bis F werden daran
// gemessen, ob sie den Abstand senken. Ein Verlauf über einen veränderlichen
// Korpus vergleicht Äpfel mit Birnen — wer eine Referenz austauscht, hebt oder
// senkt die Zahl, ohne eine Zeile Code zu ändern.
//
// **Und der Bestand selbst ist nicht geprüft.** Am 2026-08-17 nachgemessen: die
// Referenzbilder sind nicht mit der Vorlage entstanden, die die `.karte`-Dateien
// daneben nennen — die schwarzen Randbreiten (17 gegen 15 in jedem Rahmenbild
// des Moduls), das farbige untere Band auf allen 27 Bildern und der graue
// `uncommon`-Verlauf am Seltenheitssymbol sagen es unabhängig voneinander. Die
// `.karte` sind **Rekonstruktionen aus den Bildern**; Stylesheet, Stiloptionen
// und Setangaben von MSE2 stehen nirgends.
//
// **Der Abstand ist damit ein Verlaufsmaß und keine Aussage über Parität**
// (#375, §10).
//
// Deshalb steht vor jeder Messung, **welche** Referenzen mit welchem Inhalt
// gelaufen sind. Angestrichen vom Zweitblick über Phase A: der Lauf nannte
// weder Prüfsumme noch Fassung, und die Bilder ohne Kartendaten wurden still
// übergangen.
func bestandsausweis(t *testing.T, referenzen string, eintraege []os.DirEntry) {
	t.Helper()

	var paare, allein []string

	for _, eintrag := range eintraege {
		name, istBild := strings.CutSuffix(eintrag.Name(), ".jpg")
		if !istBild {
			continue
		}

		daten := filepath.Join(referenzen, name+".karte")
		if _, err := os.Stat(daten); err != nil { //nolint:gosec // ein Pfad aus MSE_FREMDPAKETE oder $HOME — dieser Lauf misst gegen bezogene Pakete (§3)
			allein = append(allein, name)

			continue
		}

		paare = append(paare, fmt.Sprintf("%s %s+%s", name,
			pruefsumme(t, filepath.Join(referenzen, eintrag.Name())), pruefsumme(t, daten)))
	}

	sort.Strings(paare)
	sort.Strings(allein)

	t.Logf("Bestand: %d Referenzen mit Kartendaten, %d Bilder ohne", len(paare), len(allein))

	for _, zeile := range paare {
		t.Logf("  %s", zeile)
	}

	// **Die Bilder ohne Kartendaten werden benannt, nicht verschwiegen.** Sie
	// sind keine Fehler — es fehlt nur ihre Beschreibung. Wer sie nicht sieht,
	// hält den Bestand für vollständig.
	if len(allein) > 0 {
		t.Logf("ohne .karte und deshalb ungemessen: %s", strings.Join(allein, ", "))
	}
}

// pruefsumme gibt die ersten Stellen des SHA-256 einer Datei.
//
// **Kurz, aber nicht kurzgeschlossen:** zwölf Stellen unterscheiden jede
// Referenz von jeder anderen; die volle Summe machte das Protokoll unlesbar,
// ohne mehr auszusagen.
func pruefsumme(t *testing.T, pfad string) string {
	t.Helper()

	inhalt, err := os.ReadFile(pfad) //nolint:gosec // ein Pfad aus einer Umgebungsvariable des Laufs
	if err != nil {
		t.Fatalf("Prüfsumme von %s: %v", pfad, err)
	}

	return fmt.Sprintf("%x", sha256.Sum256(inhalt))[:12]
}
