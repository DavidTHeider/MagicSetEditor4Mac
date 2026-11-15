package bezug_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
)

// eintrag ist eine Datei im gebauten Archiv.
type eintrag struct {
	name    string
	inhalt  string
	rechte  int64
	verweis string
}

// archiv baut ein `tar.gz`, wie GitHub es ausliefert: alles unter **einem**
// Wurzelverzeichnis, dessen Name den Stand trägt.
func archiv(t *testing.T, wurzel string, eintraege []eintrag) []byte {
	t.Helper()

	var gepackt bytes.Buffer

	packer := gzip.NewWriter(&gepackt)
	schreiber := tar.NewWriter(packer)

	for _, e := range eintraege {
		kopf := &tar.Header{
			Name:     wurzel + "/" + e.name,
			Mode:     e.rechte,
			Size:     int64(len(e.inhalt)),
			Typeflag: tar.TypeReg,
		}

		if e.verweis != "" {
			kopf.Typeflag = tar.TypeSymlink
			kopf.Linkname = e.verweis
			kopf.Size = 0
		}

		if kopf.Mode == 0 {
			kopf.Mode = 0o644
		}

		if err := schreiber.WriteHeader(kopf); err != nil {
			t.Fatalf("Kopf schreiben: %v", err)
		}

		if kopf.Typeflag == tar.TypeReg {
			if _, err := schreiber.Write([]byte(e.inhalt)); err != nil {
				t.Fatalf("Inhalt schreiben: %v", err)
			}
		}
	}

	if err := schreiber.Close(); err != nil {
		t.Fatalf("tar schließen: %v", err)
	}

	if err := packer.Close(); err != nil {
		t.Fatalf("gzip schließen: %v", err)
	}

	return gepackt.Bytes()
}

// probeeintraege stellen das nach, was `M15-Magic-Pack` wirklich enthält:
// Paketdaten unter `data/` und Windows-Programme daneben.
func probeeintraege() []eintrag {
	return []eintrag{
		{name: "README.md", inhalt: "Eine Sammlung."},
		{name: "magicseteditor.exe", inhalt: "MZ ein Windows-Programm", rechte: 0o755},
		{name: "mse.com", inhalt: "auch eines"},
		{name: "data/probe.mse-game/game", inhalt: "short name: Probe\n"},
		{name: "data/probe.mse-style/style", inhalt: "game: probe\n"},
		{name: "Magic - Fonts/mplantin.ttf", inhalt: "eine Schrift"},
	}
}

// server gibt einen Testserver, der Archiv und Marken ausliefert.
func server(t *testing.T, inhalt []byte, marken ...string) *httptest.Server {
	t.Helper()

	dienst := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if stand, ist := strings.CutPrefix(r.URL.Path, "/marke/"); ist {
			for _, marke := range marken {
				if marke == stand {
					w.WriteHeader(http.StatusOK)

					return
				}
			}

			w.WriteHeader(http.StatusNotFound)

			return
		}

		if _, err := w.Write(inhalt); err != nil {
			t.Errorf("ausliefern: %v", err)
		}
	}))

	t.Cleanup(dienst.Close)

	return dienst
}

// quelle baut eine Bezugsquelle über den Testserver.
func quelle(dienst *httptest.Server, stand string) bezug.Repositorium {
	return bezug.Repositorium{
		Adresse:          dienst.URL + "/archiv",
		Marken:           dienst.URL + "/marke/",
		Stand:            stand,
		Unterverzeichnis: "data",
	}
}

// **Ein Paket wird über einen Stand bezogen, nicht über einen Zweig.**
//
// Ein Zweig ist keine Fassung: `main` ändert sich, und eine Prüfsumme darauf
// wäre eine Momentaufnahme. Bezogen wird eine Freigabe oder ein Commit.
func TestEinPaketWirdUeberEinenStandBezogen(t *testing.T) {
	dienst := server(t, archiv(t, "M15-Magic-Pack-3.1.0", probeeintraege()), "3.1.0")
	ziel := t.TempDir()

	ergebnis, err := bezug.Hole(t.Context(), quelle(dienst, "3.1.0"), ziel, "MagicSetEditor4Mac/probe")
	if err != nil {
		t.Fatalf("Hole: %v", err)
	}

	// Das Unterverzeichnis ist die Paketwurzel: was darunter lag, liegt jetzt
	// unmittelbar im Ziel.
	if _, err := os.Stat(filepath.Join(ziel, "probe.mse-game", "game")); err != nil {
		t.Errorf("das Paket fehlt im Ziel: %v", err)
	}

	if ergebnis.Pruefsumme == "" {
		t.Error("keine Prüfsumme — was bezogen wurde, ist nicht nachweisbar")
	}

	if ergebnis.Stand != "3.1.0" {
		t.Errorf("Stand: %q", ergebnis.Stand)
	}
}

// Ein Zweig wird abgewiesen, und die Meldung sagt warum.
func TestEinZweigWirdAbgewiesen(t *testing.T) {
	dienst := server(t, archiv(t, "M15-Magic-Pack-main", probeeintraege()), "3.1.0")

	_, err := bezug.Hole(t.Context(), quelle(dienst, "main"), t.TempDir(), "MagicSetEditor4Mac/probe")
	if err == nil {
		t.Fatal("kein Fehler — ein Zweig ist als Stand durchgegangen")
	}

	if !strings.Contains(err.Error(), "main") {
		t.Errorf("Fehler: %q — der Stand ist nicht benannt", err)
	}
}

// Ein Commit-Hash ist ein Stand, auch ohne Marke.
func TestEinCommitIstEinStand(t *testing.T) {
	const commit = "0123456789abcdef0123456789abcdef01234567"

	dienst := server(t, archiv(t, "M15-Magic-Pack-"+commit, probeeintraege()))

	if _, err := bezug.Hole(t.Context(), quelle(dienst, commit), t.TempDir(), "MagicSetEditor4Mac/probe"); err != nil {
		t.Fatalf("Hole: %v", err)
	}
}

// **Ausführbare Dateien werden nicht ausgepackt.**
//
// Die Regel steht im Code, nicht in einem Kommentar: `M15-Magic-Pack` liefert
// `magicseteditor.exe`, `mse.exe` und einen Deinstallierer mit — ausführbaren
// Code unbekannter Herkunft.
func TestAusfuehrbaresWirdNichtAusgepackt(t *testing.T) {
	eintraege := append(probeeintraege(),
		eintrag{name: "data/mse.exe", inhalt: "MZ"},
		eintrag{name: "data/hilfe.sh", inhalt: "#!/bin/sh", rechte: 0o755},
		eintrag{name: "data/a/build.bat", inhalt: "@echo"},
		eintrag{name: "data/b/build.bat", inhalt: "@echo"},
	)

	dienst := server(t, archiv(t, "wurzel", eintraege), "3.1.0")
	ziel := t.TempDir()

	ergebnis, err := bezug.Hole(t.Context(), quelle(dienst, "3.1.0"), ziel, "MagicSetEditor4Mac/probe")
	if err != nil {
		t.Fatalf("Hole: %v", err)
	}

	for _, name := range []string{"mse.exe", "hilfe.sh"} {
		if _, err := os.Stat(filepath.Join(ziel, name)); err == nil {
			t.Errorf("%s ist ausgepackt worden", name)
		}
	}

	// **Und es wird gesagt, nicht verschwiegen** — **mit Pfad im Paket**. Wer
	// nicht erfährt, dass etwas fehlt, sucht später den Fehler im Paket; und
	// der Echtlauf am 2026-10-04 (#2049) meldete fünfmal „build.bat", ohne zu
	// sagen, welche.
	erwartet := []string{"mse.exe", "hilfe.sh", "a/build.bat", "b/build.bat"}
	if !slices.Equal(ergebnis.Uebergangen, erwartet) {
		t.Errorf("Übergangen: %v, erwartet %v", ergebnis.Uebergangen, erwartet)
	}
}

// Ein Archiv kann nicht aus dem Zielverzeichnis ausbrechen (#56).
func TestEinArchivKannNichtAusbrechen(t *testing.T) {
	faelle := map[string]eintrag{
		"ein Weg mit ..":        {name: "data/../../draussen.txt", inhalt: "hier nicht"},
		"ein Verweis nach oben": {name: "data/verweis", verweis: "../../../etc/passwd"},
	}

	for name, boeser := range faelle {
		t.Run(name, func(t *testing.T) {
			dienst := server(t, archiv(t, "wurzel", []eintrag{boeser}), "3.1.0")
			ziel := t.TempDir()

			_, err := bezug.Hole(t.Context(), quelle(dienst, "3.1.0"), ziel, "MagicSetEditor4Mac/probe")
			if err == nil {
				t.Fatal("kein Fehler")
			}

			if _, err := os.Stat(filepath.Join(filepath.Dir(ziel), "draussen.txt")); err == nil {
				t.Error("es ist etwas außerhalb des Ziels entstanden")
			}
		})
	}
}

// **Vor dem Bezug steht, wessen Inhalte geholt werden.**
//
// **Als Angaben, nicht als Satz** (§9): der Satz steht in den Sprachkatalogen
// der Oberfläche, hier stehen die Werte, die er nennt. Ein deutscher Fließtext
// von hier wäre in der englischen Oberfläche deutsch — der Zweitblick über
// Phase 9 hat es angestrichen.
func TestDieHerkunftNenntQuelleUndStand(t *testing.T) {
	woher := bezug.Woher(bezug.Repositorium{
		Adresse: "https://github.com/MagicSetEditorPacks/M15-Magic-Pack",
		Stand:   "3.1.0",
	})

	if woher.Quelle != "MagicSetEditorPacks/M15-Magic-Pack" {
		t.Errorf("Quelle: %q", woher.Quelle)
	}

	if woher.Stand != "3.1.0" {
		t.Errorf("Stand: %q", woher.Stand)
	}
}
