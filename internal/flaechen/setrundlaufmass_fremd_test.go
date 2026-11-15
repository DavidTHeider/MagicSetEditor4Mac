//go:build fremdpakete

package flaechen_test

// **Die Maße des Setdatei-Rundlaufs** — F05
// (#2247):
// was das Orakel über ein Set sagt, was `mse` darüber sagt, und welche
// Schlüssel und Dateien eine Setdatei trägt. Der Lauf selbst steht in
// `setrundlauf_fremd_test.go`.

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// kartenwerte liest die Werte jeder Karte, wie `mse karten --json` sie gibt.
func kartenwerte(t *testing.T, umgebung []string, setID string) (kennungen []string, werte []map[string]string) {
	t.Helper()

	var antwort struct {
		Karten []struct {
			ID    string            `json:"id"`
			Werte map[string]string `json:"werte"`
		} `json:"karten"`
	}

	if fehler := json.Unmarshal([]byte(mussMse(t, umgebung, "karten", setID, "--json")), &antwort); fehler != nil {
		t.Fatalf("karten: %v", fehler)
	}

	for _, karte := range antwort.Karten {
		kennungen = append(kennungen, karte.ID)
		werte = append(werte, karte.Werte)
	}

	return kennungen, werte
}

// mussMse ruft `mse` und bricht ab, wenn es scheitert.
func mussMse(t *testing.T, umgebung []string, argumente ...string) string {
	t.Helper()

	aus, fehler, code := kommandozeile(t, umgebung, argumente...)
	if code != 0 {
		t.Fatalf("mse %v: %d (%s)", argumente, code, fehler)
	}

	return aus
}

// spielfelder nennt die Kartenfelder des Spiels, das ein Set nennt — samt
// ihrer Art, denn Bildfelder vergleicht das Orakel nicht als Text.
func spielfelder(t *testing.T, pack, setdatei string) map[string]bool {
	t.Helper()

	spiel := regexp.MustCompile(`(?m)^game:\s*(\S+)`).FindStringSubmatch(setdatei)
	if spiel == nil {
		t.Fatal("die Setdatei nennt kein Spiel")
	}

	paket, fehler := msepaket.Oeffne(filepath.Join(pack, spiel[1]+".mse-game"))
	if fehler != nil {
		t.Fatalf("das Spiel %s: %v", spiel[1], fehler)
	}

	defer func() { _ = paket.Schliessen() }() //nolint:errcheck // nur gelesen

	modell, _ := paket.Spielmodell()
	felder := make(map[string]bool, len(modell.Fields))

	for _, feld := range modell.Fields {
		felder[msepaket.Entkanonisch(feld.Name)] = feld.Kind == "image"
	}

	return felder
}

// paketinhalt liest die Dateien eines Sets — ZIP oder Verzeichnis.
func paketinhalt(t *testing.T, pfad string) map[string][]byte {
	t.Helper()

	inhalt := make(map[string][]byte)

	if info, fehler := os.Stat(pfad); fehler == nil && info.IsDir() {
		eintraege, lesefehler := os.ReadDir(pfad)
		if lesefehler != nil {
			t.Fatalf("%s: %v", pfad, lesefehler)
		}

		// **Über das Verzeichnis als Dateisystem** — ein Name aus ihm kann es
		// nicht verlassen, und gosec sieht keine offene Senke (G304).
		ordner := os.DirFS(pfad)

		for _, eintrag := range eintraege {
			daten, dateifehler := fs.ReadFile(ordner, eintrag.Name())
			if dateifehler != nil {
				t.Fatalf("%s: %v", eintrag.Name(), dateifehler)
			}

			inhalt[eintrag.Name()] = daten
		}

		return inhalt
	}

	archiv, fehler := zip.OpenReader(pfad)
	if fehler != nil {
		t.Fatalf("%s: %v", pfad, fehler)
	}

	defer func() { _ = archiv.Close() }() //nolint:errcheck // nur gelesen

	for _, datei := range archiv.File {
		strom, oeffnefehler := datei.Open()
		if oeffnefehler != nil {
			t.Fatalf("%s: %v", datei.Name, oeffnefehler)
		}

		daten, lesefehler := io.ReadAll(strom)
		_ = strom.Close() //nolint:errcheck // nur gelesen, wie das Archiv darüber

		if lesefehler != nil {
			t.Fatalf("%s: %v", datei.Name, lesefehler)
		}

		inhalt[datei.Name] = daten
	}

	return inhalt
}

// dateienVergleichen zählt die Dateien neben `set`, die in der Abschrift
// fehlen oder andere Bytes tragen.
func dateienVergleichen(vorher, nachher map[string][]byte) (zahl int, fehlen, anders []string) {
	for name, daten := range vorher {
		if name == "set" {
			continue
		}

		zahl++

		neu, da := nachher[name]

		switch {
		case !da:
			fehlen = append(fehlen, name)
		case sha256.Sum256(neu) != sha256.Sum256(daten):
			anders = append(anders, name)
		}
	}

	return zahl, fehlen, anders
}

// schluesselzeile ist ein Schlüssel auf der ersten oder zweiten Ebene.
var schluesselzeile = regexp.MustCompile(`^(\t?)([^\t:][^:]*):`)

// schluesselpfade zählt die Schlüssel einer Setdatei auf den obersten zwei
// Ebenen — `card#3/rule text`, `set info/title`, `styling/magic-m15/…`.
//
// **Tiefer zählt er nicht**: ein mehrzeiliger Wert steht dort, und ihn von
// einem Block zu unterscheiden, verlangte das Schema des Spiels. Beide
// Seiten eines Vergleichs werden gleich gezählt.
func schluesselpfade(datei string) map[string]int {
	pfade := make(map[string]int)
	oben := ""
	zaehler := make(map[string]int)

	for _, zeile := range strings.Split(strings.ReplaceAll(datei, "\r\n", "\n"), "\n") {
		treffer := schluesselzeile.FindStringSubmatch(strings.TrimPrefix(zeile, "\ufeff"))
		if treffer == nil {
			continue
		}

		schluessel := msepaket.Kanonisch(strings.TrimSpace(treffer[2]))

		if treffer[1] == "" {
			oben = fmt.Sprintf("%s#%d", schluessel, zaehler[schluessel])
			zaehler[schluessel]++
			pfade[oben]++

			continue
		}

		pfade[oben+"/"+schluessel]++
	}

	return pfade
}

// verloreneSchluessel nennt, was in `vorher` steht und in `nachher` nicht.
func verloreneSchluessel(vorher, nachher map[string]int) []string {
	var weg []string

	for pfad, zahl := range vorher {
		if nachher[pfad] < zahl {
			weg = append(weg, pfad)
		}
	}

	return weg
}

// hatUnicode sagt, ob ein Text ein Zeichen jenseits von ASCII trägt.
func hatUnicode(text string) bool {
	for _, zeichen := range text {
		if zeichen > 127 {
			return true
		}
	}

	return false
}
