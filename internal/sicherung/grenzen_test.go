// **Die Grenzen des Auspackens** — Befund 5 des Zweitblicks über #1162.
//
// Bis dahin war die einzige Schranke die Angabe, die **im Archiv selbst**
// steht: `Groesse` aus dem Manifest. Eine Sicherung, die eine gewaltige Zahl
// hineinschreibt, bekam sie geglaubt — und `io.ReadAll` versuchte, so viel in
// den Arbeitsspeicher zu legen.
//
// **Die Reichweite dieser Läufe**: sie prüfen, dass die Grenzen greifen, nicht
// dass sie richtig gewählt sind. Ob 64 MiB Manifest großzügig oder knapp ist,
// entscheidet niemand hier.
package sicherung_test

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// **Ein Teil, der mehr ankündigt als erlaubt, wird abgewiesen — vor dem
// Lesen.**
func TestEinZuGrossAngekuendigterTeilWirdAbgewiesen(t *testing.T) {
	pfad := gepacktesArchiv(t)

	umbauen(t, pfad, func(name string, roh []byte) ([]byte, bool) {
		if name != sicherung.Manifestname {
			return roh, true
		}

		return []byte(strings.Replace(string(roh), `"groesse": 23`,
			`"groesse": 1099511627776`, 1)), true
	})

	_, fehler := sicherung.Entpacken(pfad, t.TempDir())
	if fehler == nil {
		t.Fatal("die angekündigte Größe von einem Terabyte kam durch")
	}

	if !strings.Contains(fehler.Error(), "zu groß") {
		t.Errorf("die Meldung sagt nicht, was zu groß ist: %v", fehler)
	}
}

// **Ein Manifest über der Grenze wird abgewiesen.**
//
// Es ist der eine Eintrag, dessen Größe niemand vorher kennt — er *ist* die
// Auskunft darüber, wie groß etwas sein darf.
func TestEinZuGrossesManifestWirdAbgewiesen(t *testing.T) {
	pfad := gepacktesArchiv(t)

	umbauen(t, pfad, func(name string, roh []byte) ([]byte, bool) {
		if name != sicherung.Manifestname {
			return roh, true
		}

		// **Gültiges JSON, nur zu viel davon.** Ein Haufen Nullbytes ginge
		// schon am Zerteiler kaputt, und dann prüfte der Lauf den falschen
		// Zweig (§10).
		fuellung := `{"fassung":"0.3.0","fuellung":"` +
			strings.Repeat("x", int(sicherung.HoechstensManifest)+1) + `"}`

		return []byte(fuellung), true
	})

	_, fehler := sicherung.Entpacken(pfad, t.TempDir())
	if fehler == nil {
		t.Fatal("das übergroße Manifest kam durch")
	}

	if !strings.Contains(fehler.Error(), "zu groß") {
		t.Errorf("die Meldung sagt nicht, was zu groß ist: %v", fehler)
	}
}

// **Und ein Teil, dessen Prüfsumme nicht stimmt, kommt nicht an seinem Ort
// an.**
//
// Diese Zusage hing bis zum Zweitblick daran, dass der ganze Teil erst in den
// Arbeitsspeicher ging. Seit er **strömt**, hängt sie an einer Nachbardatei
// und einem Umbenennen — und dieser Lauf ist der, der es festhält.
func TestEinTeilMitFalscherSummeKommtNichtAn(t *testing.T) {
	pfad := gepacktesArchiv(t)

	umbauen(t, pfad, func(name string, roh []byte) ([]byte, bool) {
		if name != "beiwerk/4/bild.png" {
			return roh, true
		}

		return []byte("etwas ganz anderes, gleiche Länge!!"), true
	})

	ziel := t.TempDir()

	if _, fehler := sicherung.Entpacken(pfad, ziel); fehler == nil {
		t.Fatal("der veränderte Teil kam durch")
	}

	eintraege, fehler := os.ReadDir(filepath.Join(ziel, "beiwerk", "4"))
	if fehler == nil && len(eintraege) != 0 {
		t.Errorf("am Ort des Teils liegt etwas: %v", eintraege)
	}
}

// gepacktesArchiv legt eine gültige Sicherung an und gibt ihren Pfad.
func gepacktesArchiv(t *testing.T) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "grenze"+sicherung.Endung)

	if fehler := sicherung.Packen(pfad, probeteile(t), sicherung.Angaben{
		Fassung: "0.3.0",
	}); fehler != nil {
		t.Fatalf("Packen: %v", fehler)
	}

	return pfad
}

// **Unter dem endgültigen Namen steht nie eine halbe Sicherung.**
//
// Befund 6 des Zweitblicks: `Packen` legte die `.msesicherung` an und schrieb
// unmittelbar hinein. Ein Absturz, ein erzwungenes Beenden, ein Stromausfall —
// und die Datei blieb unter dem richtigen Namen liegen, mit halbem Inhalt.
// Beim nächsten Mal fiel es doppelt auf: die Sicherung war unbrauchbar, **und**
// ein neuer Versuch unter demselben Namen scheiterte am belegten Ziel.
//
// **Was dieser Lauf zeigt und was nicht**: er zeigt, dass ein gescheitertes
// Packen weder das Ziel noch eine Nachbardatei zurücklässt. Dass der Name im
// Augenblick des Absturzes frei ist, folgt aus `os.Link` — und das zeigt kein
// Test in diesem Prozess.
func TestEinGescheitertesPackenLaesstNichtsLiegen(t *testing.T) {
	ort := t.TempDir()
	ziel := filepath.Join(ort, "halb"+sicherung.Endung)

	// **Ein Verzeichnis als Quelle**: `os.Open` gelingt, das Lesen nicht.
	unlesbar := filepath.Join(ort, "verzeichnis")
	if err := os.Mkdir(unlesbar, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if fehler := sicherung.Packen(ziel,
		[]sicherung.Teil{{Name: "bestand.db", Quelle: unlesbar}},
		sicherung.Angaben{Fassung: "0.3.0"}); fehler == nil {
		t.Fatal("das gescheiterte Packen meldete Erfolg")
	}

	eintraege, fehler := os.ReadDir(ort)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	for _, eintrag := range eintraege {
		if eintrag.Name() != "verzeichnis" {
			t.Errorf("nach dem Fehlschlag liegt %q da", eintrag.Name())
		}
	}
}

// **Ein Archiv mit zu vielen ZIP-Einträgen wird abgewiesen — auch wenn das
// Manifest kurz ist.**
//
// Befund 3 des Nachlaufs: gezählt wurde die Liste **im Manifest**. Ein Archiv
// mit einem kurzen Manifest und Millionen nicht aufgeführter Einträge ging
// daran vorbei.
//
// **Die Reichweite**: der Lauf zeigt, dass die Zählung am ZIP-Verzeichnis
// hängt und nicht am Manifest. Er sagt **nichts** darüber, was `archive/zip`
// schon beim Öffnen tut — das Verzeichnis liegt dann bereits im Speicher, und
// diese Grenze ist keine von diesem Paket.
func TestEinArchivMitZuVielenEintraegenWirdAbgewiesen(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "viele"+sicherung.Endung)

	datei, err := os.Create(pfad) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	schreiber := zip.NewWriter(datei)

	// **Ein kurzes Manifest** — es führt nichts auf.
	manifest, err := schreiber.Create(sicherung.Manifestname)
	if err != nil {
		t.Fatalf("eintragen: %v", err)
	}

	if _, fehler := manifest.Write([]byte(`{"fassung":"0.3.0"}`)); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	// **Und sehr viele Einträge daneben**, die es nicht nennt.
	for i := 0; i <= sicherung.HoechstensEintraege; i++ {
		if _, fehler := schreiber.CreateHeader(&zip.FileHeader{
			Name: "f" + strconv.Itoa(i), Method: zip.Store,
		}); fehler != nil {
			t.Fatalf("eintragen: %v", fehler)
		}
	}

	if fehler := errors.Join(schreiber.Close(), datei.Close()); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	_, fehler := sicherung.Entpacken(pfad, t.TempDir())
	if fehler == nil {
		t.Fatal("das Archiv mit zu vielen Einträgen kam durch")
	}

	if !strings.Contains(fehler.Error(), "Einträge") {
		t.Errorf("die Meldung sagt nicht, was zu viel ist: %v", fehler)
	}
}

// **Teile, die einzeln zulässig sind und zusammen zu schwer, werden
// abgewiesen.**
//
// Sie füllten sonst das Laufwerk, auf dem der laufende Bestand liegt — jeder
// für sich unter der Grenze, alle zusammen weit darüber.
func TestZuSchwereTeileZusammenWerdenAbgewiesen(t *testing.T) {
	pfad := gepacktesArchiv(t)

	umbauen(t, pfad, func(name string, roh []byte) ([]byte, bool) {
		if name != sicherung.Manifestname {
			return roh, true
		}

		var beschreibung sicherung.Beschreibung
		if fehler := json.Unmarshal(roh, &beschreibung); fehler != nil {
			t.Fatalf("Manifest lesen: %v", fehler)
		}

		// **Je Teil zulässig, zusammen darüber.** Fünf zu je 15 GiB sind
		// einzeln unter der Grenze und zusammen 75 GiB.
		beschreibung.Teile = nil
		for i := range 5 {
			beschreibung.Teile = append(beschreibung.Teile, sicherung.Teilangabe{
				Name:    "beiwerk/" + strconv.Itoa(i) + "/gross.png",
				Groesse: 15 << 30,
			})
		}

		gebaut, fehler := json.Marshal(beschreibung)
		if fehler != nil {
			t.Fatalf("Manifest schreiben: %v", fehler)
		}

		return gebaut, true
	})

	_, fehler := sicherung.Entpacken(pfad, t.TempDir())
	if fehler == nil {
		t.Fatal("die zusammen zu schweren Teile kamen durch")
	}

	if !strings.Contains(fehler.Error(), "zusammen") {
		t.Errorf("die Meldung nennt den Grund nicht: %v", fehler)
	}
}
