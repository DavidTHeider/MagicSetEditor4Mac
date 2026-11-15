package msepaket_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Pakete liegen nicht dort, wo der Bezug sie abgelegt hat.**
//
// # Der Befund, aus dem diese Datei entstanden ist
//
// `Bildquelle` und `Paketleser` fügen Wurzel und Paketnamen zusammen — sie
// erwarten die Pakete **unmittelbar** unter der Wurzel. Bezogen wird aber ein
// ganzes Repositorium, und `M15-Magic-Pack` legt seine 190 Stile unter
// `M15-Magic-Pack/data/` ab.
//
// Nachgemessen am Bestand dieses Rechners: unter `…/pakete` liegt genau ein
// Eintrag, und das ist ein Verzeichnis ohne ein einziges Paket darin. **Damit
// fand die laufende Anwendung kein Bild und kein Modul** — beides gehört zu
// #206.
func TestPaketwurzelnFindetPaketeInDerTiefe(t *testing.T) {
	wurzel := t.TempDir()

	// Wie ein bezogenes Repositorium aussieht: zwei Ebenen bis zum Paket.
	tief := filepath.Join(wurzel, "M15-Magic-Pack", "data")
	lege(t, filepath.Join(tief, "magic-m15.mse-style"))
	lege(t, filepath.Join(tief, "magic.mse-game"))

	gefunden := msepaket.Paketwurzeln(wurzel)

	if !slices.Contains(gefunden, tief) {
		t.Errorf("%v enthält %q nicht — die Pakete sind unauffindbar", gefunden, tief)
	}
}

// **Die Wurzel selbst zählt mit**, wenn die Pakete gleich darunter liegen.
//
// So sieht `saat/` aus, und so sähe ein von Hand entpacktes Paket aus.
func TestPaketwurzelnNimmtDieWurzelSelbst(t *testing.T) {
	wurzel := t.TempDir()
	lege(t, filepath.Join(wurzel, "probe-schlicht.mse-style"))

	if gefunden := msepaket.Paketwurzeln(wurzel); !slices.Contains(gefunden, wurzel) {
		t.Errorf("%v enthält die Wurzel %q nicht", gefunden, wurzel)
	}
}

// **In ein Paket wird nicht hineingestiegen.**
//
// Ein Stilpaket führt Unterverzeichnisse, und eines davon könnte zufällig auf
// `.mse-style` enden. Wer dort weitersucht, meldet eine Wurzel, die keine ist —
// und zahlt bei 190 Paketen mit Tausenden Dateien für das Nachsehen.
func TestPaketwurzelnSteigtNichtInEinPaket(t *testing.T) {
	wurzel := t.TempDir()
	paket := filepath.Join(wurzel, "aussen.mse-style")
	lege(t, filepath.Join(paket, "innen.mse-style"))

	if gefunden := msepaket.Paketwurzeln(wurzel); slices.Contains(gefunden, paket) {
		t.Errorf("%v nennt %q — dort ist gar keine Wurzel, sondern ein Paket", gefunden, paket)
	}
}

// **Ein Verzeichnis ohne Paket ist keine Wurzel.**
//
// Ohne diese Zusage gäbe jede Suche die halbe Platte zurück, und `Bildquelle`
// liefe für jedes fehlende Bild durch alle davon.
func TestPaketwurzelnMeldetNurWasPaketeTraegt(t *testing.T) {
	wurzel := t.TempDir()
	lege(t, filepath.Join(wurzel, "leer", "auch-leer"))

	if gefunden := msepaket.Paketwurzeln(wurzel); len(gefunden) != 0 {
		t.Errorf("%v — dort liegt kein einziges Paket", gefunden)
	}
}

// **Eine leere Wurzel gibt nichts**, statt im Arbeitsverzeichnis zu suchen.
//
// Dieselbe Falle wie bei [msepaket.NeueBildquelle]: `filepath.Join("", …)`
// macht daraus einen relativen Pfad.
func TestPaketwurzelnGehtNichtInsArbeitsverzeichnis(t *testing.T) {
	if gefunden := msepaket.Paketwurzeln("", "   "); len(gefunden) != 0 {
		t.Errorf("%v — aus einer leeren Angabe darf keine Wurzel werden", gefunden)
	}
}

// lege legt ein Verzeichnis samt Eltern an.
func lege(t *testing.T, pfad string) {
	t.Helper()

	if err := os.MkdirAll(pfad, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}
}

// **Was tiefer als [msepaket.GrenzePakettiefe] liegt, wird nicht gefunden.**
//
// Der Test steht hier, damit die Grenze eine Zusage ist und nicht eine Zahl:
// ohne ihn ließe sich `GrenzePakettiefe` auf 1 setzen, und kein Lauf würde rot.
// Was er zugleich sagt: **wer ein Paket tiefer ablegt, bekommt keine Meldung**,
// sondern ein leeres Ergebnis.
func TestPaketwurzelnGehtNichtBeliebigTief(t *testing.T) {
	wurzel := t.TempDir()

	teile := make([]string, 0, msepaket.GrenzePakettiefe+2)
	teile = append(teile, wurzel)

	for range msepaket.GrenzePakettiefe + 1 {
		teile = append(teile, "tiefer")
	}

	tief := filepath.Join(teile...)
	lege(t, filepath.Join(tief, "magic-m15.mse-style"))

	if gefunden := msepaket.Paketwurzeln(wurzel); len(gefunden) != 0 {
		t.Errorf("%v — %d Ebenen tief soll nicht mehr gefunden werden",
			gefunden, msepaket.GrenzePakettiefe+1)
	}
}

// **Ein Verzeichnis ohne Leserecht überspringt seinen Ast.**
//
// Es liegt in jedem Heimatverzeichnis — `~/Library` führt mehrere. Wer daran
// abbräche, verlöre die Pakete daneben mit; und ein Fehlerzweig, den kein Test
// betritt, ist eine Behauptung darüber, was im Ernstfall geschieht (§3).
func TestPaketwurzelnUeberspringtWasNichtLesbarIst(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("als root ist jedes Verzeichnis lesbar")
	}

	wurzel := t.TempDir()

	// Ein Paket, das gefunden werden **soll** — daneben eine verschlossene Tür.
	offen := filepath.Join(wurzel, "z-offen")
	lege(t, filepath.Join(offen, "magic-m15.mse-style"))

	// **Vor dem lesbaren Ast.** `WalkDir` geht lexikalisch; stünde `offen`
	// davor, bliebe der Test auch grün, wenn der Lauf am Lesefehler abbräche.
	zu := filepath.Join(wurzel, "a-verschlossen")
	lege(t, filepath.Join(zu, "drinnen"))

	if err := os.Chmod(zu, 0o000); err != nil {
		t.Fatalf("verschließen: %v", err)
	}

	// **Wieder aufschließen, sonst kann `t.TempDir` nicht aufräumen.** Ein
	// Fehler dabei gehört gemeldet: er hinterließe ein Verzeichnis, das der
	// nächste Lauf nicht löschen kann.
	t.Cleanup(func() {
		if err := os.Chmod(zu, 0o700); err != nil { // #nosec G302 -- ein Verzeichnis braucht x
			t.Errorf("aufschließen: %v", err)
		}
	})

	gefunden := msepaket.Paketwurzeln(wurzel)

	if !slices.Contains(gefunden, offen) {
		t.Errorf("%v — das lesbare Paket daneben ist mit verlorengegangen", gefunden)
	}
}
