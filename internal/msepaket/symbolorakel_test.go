//go:build fremdpakete

// **Der Lauf gegen MSE2 selbst** —
// #890,
// Befund 10 des Zweitblicks über Phase X.
//
// # Warum der Rundlauf daneben nicht genügt
//
// `symbolschreiber_test.go` schreibt mit unserem Schreiber und liest mit
// unserem Leser. **Derselbe Feld-, Hierarchie- oder Formatfehler in beiden
// lässt ihn grün**, während MSE2 die Datei nicht benutzt. Der
// Vollständigkeitswächter daneben prüft die Aufzählungen, nicht die Gestalt
// eines Blocks.
//
// # Was dieser Lauf tut
//
// Er legt ein Set mit einer **gewöhnlichen** Karte an — ihr Seltenheitszeichen
// ist das Setsymbol —, schreibt unsere `.mse-symbol` hinein und lässt **MSE2**
// die Karte rastern. Zweimal, mit **zwei verschiedenen, beide gültigen**
// Zeichnungen unseres Schreibers: einer großen Raute und einem kleinen
// Quadrat.
//
// **Unterscheiden sich die Bilder, hat MSE2 die Punkte gelesen.** Sind sie
// gleich, sagt der Lauf nichts über unsere Datei — und genau das war der
// frühere Anlauf, der mit `--cli` und einem Vanguard-Set alle drei Fassungen
// gleich bekam.
//
// # Warum nicht gegen eine unbrauchbare Datei
//
// Der erste Anlauf stellte unsere Raute einer Datei mit erfundener Teilart
// gegenüber. **MSE2 wirft dafür einen `ParseError` und schreibt gar kein
// Bild** — `ohne` war dann leer, `bytes.Equal(mit, nil)` falsch, und der Lauf
// grün, ohne dass irgendwer die vier Punkte gerastert hätte. Ein Gegenpol, an
// dem das Werkzeug scheitert, belegt nichts über den Fall, den es besteht.
//
// Angestrichen vom Nachlauf des Zweitblicks über Phase X, Befund 10. Deshalb
// verlangt der Lauf **von beiden Seiten** ein Bild.
//
// # Was er braucht
//
//	MSE2_BINAER=~/mse2-quelle/build/magicseteditor \
//	MSE2_DATEN=~/Library/Application\ Support/magicseteditor/data \
//	  go test -tags fremdpakete ./internal/msepaket/ -run Mse2 -v
//
// `MSE2_DATEN` muss die Vorlagenpakete führen — MSE2 sucht sie in seinem
// Benutzerdatenverzeichnis. Ohne die beiden Variablen wird übersprungen und
// nicht geraten (§11).
package msepaket_test

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// TestMse2ZeichnetUnsereSymboldatei ist die Abnahme von Phase X.
func TestMse2ZeichnetUnsereSymboldatei(t *testing.T) {
	binaer := os.Getenv("MSE2_BINAER")
	daten := os.Getenv("MSE2_DATEN")

	if binaer == "" || daten == "" {
		t.Skip("MSE2_BINAER oder MSE2_DATEN nicht gesetzt — MSE2 liegt neben dem Repositorium (§11)")
	}

	// **Über eine Wurzel und nicht über einen zusammengesetzten Pfad** — der
	// Prüfer nennt ein `filepath.Join` über eine Umgebungsvariable zu Recht
	// einen Pfad aus fremder Quelle.
	wurzel, wurzelfehler := os.OpenRoot(daten)
	if wurzelfehler != nil {
		t.Skipf("MSE2_DATEN %s: %v", daten, wurzelfehler)
	}

	t.Cleanup(func() {
		if err := wurzel.Close(); err != nil {
			t.Errorf("die Wurzel schließen: %v", err)
		}
	})

	if _, err := fs.Stat(wurzel.FS(), "magic-m15.mse-style"); err != nil {
		t.Skipf("in %s liegt kein `magic-m15.mse-style`: %v", daten, err)
	}

	// **Eine große Raute und ein kleines Quadrat** — beide von unserem
	// Schreiber, beide gültig, und im Bild nicht zu verwechseln.
	raute := viereck("Raute", [4]msepaket.Vektor{
		{X: 0.5, Y: 0.02},
		{X: 0.98, Y: 0.5},
		{X: 0.5, Y: 0.98},
		{X: 0.02, Y: 0.5},
	})

	quadrat := viereck("Quadrat", [4]msepaket.Vektor{
		{X: 0.4, Y: 0.4},
		{X: 0.6, Y: 0.4},
		{X: 0.6, Y: 0.6},
		{X: 0.4, Y: 0.6},
	})

	erste := gerastert(t, binaer, setMit(t, msepaket.SchreibeSymbolzeichnung(raute)))
	zweite := gerastert(t, binaer, setMit(t, msepaket.SchreibeSymbolzeichnung(quadrat)))

	// **Beide Seiten müssen ein Bild geben.** Fehlt eines, hat MSE2 die Datei
	// abgewiesen, und ein Ungleichheitsvergleich gegen nichts wäre grün, ohne
	// etwas geprüft zu haben.
	if len(erste) == 0 || len(zweite) == 0 {
		t.Fatalf("MSE2 gab %d und %d Bytes — beide Zeichnungen sind gültig und "+
			"müssen rastern", len(erste), len(zweite))
	}

	if bytes.Equal(erste, zweite) {
		t.Error("zwei verschiedene Zeichnungen geben dasselbe Bild — MSE2 liest " +
			"die Punkte nicht, und dieser Lauf sagt nichts über unsere Datei")
	}
}

// viereck baut eine geschlossene Form aus vier geraden Strecken.
func viereck(name string, ecken [4]msepaket.Vektor) msepaket.Symbolzeichnung {
	punkte := make([]msepaket.Kontrollpunkt, 0, len(ecken))
	for _, ecke := range ecken {
		punkte = append(punkte, msepaket.Kontrollpunkt{Stelle: ecke})
	}

	return msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{{
		Art: msepaket.TeilForm, Name: name,
		Kombination: msepaket.KombinationVereinen, Punkte: punkte,
	}}}
}

// setMit legt ein `.mse-set` mit einer Karte und dieser Symboldatei an.
func setMit(t *testing.T, symbol []byte) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "Probe.mse-set")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Eine gewöhnliche Karte, kein Avatar.** Das Setsymbol kommt über das
	// Seltenheitsfeld auf die Karte (`magic-modules.mse-include/rarities/choice_images`),
	// und ein Kartenstil ohne dieses Feld zeichnet es nicht.
	//
	// **Daran ist der erste Anlauf gescheitert**: er nahm das kanonische
	// `Vanguard`, dessen Karten alle Avatare sind — dort gaben unsere Datei,
	// eine unbrauchbare und gar keine dasselbe Bild, und die Messung sagte
	// nichts über die Datei. Am Stilnamen lag es **nicht**; `m15-altered`
	// zeichnet es über eine gewöhnliche Karte genauso.
	inhalt := "mse_version: 2.1.2\ngame: magic\ngame_version: 2020-04-25\n" +
		"stylesheet: m15\nstylesheet_version: 2020-04-25\n" +
		"set_info:\n\tsymbol: symbol1.mse-symbol\n" +
		"card:\n\tname: Probekarte\n\trarity: common\n"

	if err := os.WriteFile(filepath.Join(ort, "set"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ort, "symbol1.mse-symbol"), symbol, 0o600); err != nil {
		t.Fatalf("Symbol schreiben: %v", err)
	}

	return ort
}

// gerastert lässt MSE2 die Karte des Sets als Bild schreiben.
func gerastert(t *testing.T, binaer, set string) []byte {
	t.Helper()

	ziel := t.TempDir()

	// #nosec G204,G702 -- der Pfad kommt aus MSE2_BINAER, gesetzt von dem, der
	// den Lauf fährt; dieser Lauf steht hinter `fremdpakete` und läuft nicht in
	// der Pipeline (§3)
	lauf := exec.CommandContext(t.Context(), binaer, "--export-images", set, "karte.png")
	lauf.Dir = ziel

	if _, err := lauf.CombinedOutput(); err != nil {
		t.Logf("MSE2 endete mit %v — das Bild zählt, nicht der Rückgabewert", err)
	}

	bild, err := os.ReadFile(filepath.Join(ziel, "karte.png")) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		return nil
	}

	return bild
}
