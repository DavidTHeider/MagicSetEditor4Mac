//go:build fremdpakete

package app

// **Eine Referenz gilt nur mit ihrem Manifest** — QA-03.AK-03
// (#1656).
//
// MSE2 nummeriert gleichnamige Karten eines Laufs durch und überschreibt, was
// es neu schreibt; was ein früherer Lauf hinterließ, rührt es nicht an
// (`src/util/file_utils.cpp`, Zeilen 88–95, Fassung f9d9356 vom 2021-07-06).
// Ein Bild aus einem älteren Stand des Sets läge also unbemerkt neben den
// neuen, und der Vergleich nähme es. `scripts/mse2-referenz.sh` legt deshalb
// je Satz ein Manifest, und hier wird es gehalten:
//
//   - **gegen das heutige Set** — der Hash des Sets steht darin; eine Referenz
//     zu einem anderen Stand ist keine,
//   - **gegen das Verzeichnis** — jedes Bild steht im Manifest, und keines,
//     das darin steht, fehlt.
//
// **Was es nicht hält**: ob ein Bild seit dem Manifest verändert wurde. Dafür
// stehen die Hashes darin; sie bei jedem Lauf über 25.000 Bilder zu rechnen,
// kostete Minuten für einen Fall, der nur von Hand entsteht.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// referenzmanifest ist, was `scripts/mse2-referenz.sh` je Satz schreibt.
type referenzmanifest struct {
	Format    int               `json:"format"`
	Set       string            `json:"set"`
	Programm  string            `json:"programm"`
	Schriften string            `json:"schriften"`
	Karten    int               `json:"karten"`
	Erzeugt   string            `json:"erzeugt"`
	Bilder    map[string]string `json:"bilder"`
}

// manifestbefunde merkt sich je Satz, was die Prüfung sagte — **einmal je
// Prozess**, denn der Hash eines Sets geht über bis zu 227 MB.
var manifestbefunde sync.Map

// pruefeReferenzmanifest hält das Manifest eines Satzes und bricht ab, wenn es
// nicht trägt.
func pruefeReferenzmanifest(t *testing.T, satz string) {
	t.Helper()

	befund, gesehen := manifestbefunde.Load(satz)
	if !gesehen {
		befund, _ = manifestbefunde.LoadOrStore(satz, manifestbefund(satz))
	}

	if text, ist := befund.(string); ist && text != "" {
		t.Fatal(text)
	}
}

// manifestbefund gibt, was an der Referenz eines Satzes nicht stimmt — leer,
// wenn sie trägt.
func manifestbefund(satz string) string {
	ort := mse2ausgabe(satz)
	neu := "`scripts/mse2-referenz.sh " + satz + "` fährt die Referenz neu"

	roh, fehler := os.ReadFile(filepath.Join(ort, "manifest.json")) //nolint:gosec // der Ort der Referenz
	if fehler != nil {
		return satz + ": eine Referenz ohne Manifest — ob ein Bild zu diesem Stand gehört, weiß niemand; " + neu
	}

	var manifest referenzmanifest
	if fehler := json.Unmarshal(roh, &manifest); fehler != nil || manifest.Format != 1 {
		return satz + ": das Manifest ist nicht lesbar; " + neu
	}

	if hash, fehler := sethash(filepath.Join(kanonverzeichnis(), satz+".mse-set")); fehler != nil ||
		hash != manifest.Set {
		return satz + ": die Referenz gehört zu einem anderen Stand des Sets; " + neu
	}

	dateien, _ := filepath.Glob(filepath.Join(ort, "*.png")) //nolint:errcheck // ein festes Muster
	if len(dateien) != len(manifest.Bilder) {
		return satz + ": im Verzeichnis liegen andere Bilder, als das Manifest nennt; " + neu
	}

	for _, datei := range dateien {
		if _, steht := manifest.Bilder[filepath.Base(datei)]; !steht {
			return satz + ": " + filepath.Base(datei) + " steht in keinem Manifest — ein altes Bild?; " + neu
		}
	}

	return ""
}

// sethash gibt den SHA-256 einer Datei in der Form des Manifests.
func sethash(pfad string) (string, error) {
	datei, fehler := os.Open(pfad) //nolint:gosec // ein Satz aus dem Kanon
	if fehler != nil {
		return "", fehler
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	summe := sha256.New()
	if _, fehler := io.Copy(summe, datei); fehler != nil {
		return "", fehler
	}

	return "sha256:" + hex.EncodeToString(summe.Sum(nil)), nil
}

// referenzort gibt den Ort der Referenzbilder eines Satzes — **erst nach der
// Prüfung seines Manifests**. Wer Bilder daraus lädt, nimmt diesen Weg;
// `mse2ausgabe` bleibt für die bloße Frage, ob es sie gibt.
func referenzort(t *testing.T, satz string) string {
	t.Helper()

	pruefeReferenzmanifest(t, satz)

	return mse2ausgabe(satz)
}
