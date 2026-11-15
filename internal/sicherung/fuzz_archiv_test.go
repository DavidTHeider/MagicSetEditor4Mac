package sicherung_test

// **Eine Sicherung aus fremder Hand** — Werkstück O11 von Phase 5.2
// (#1435),
// Befund 3 des Zweitblicks: das Ziel über das Sicherungsmanifest fehlte.
//
// # Warum
//
// Eine Sicherung ist eine **Datei, die jemand mitbringt** — von einem anderen
// Rechner, aus einer Wolke, aus einem Ticket. Sie wird ausgepackt, und beim
// Auspacken entstehen Dateien: das ist der Weg, auf dem ein ZIP am meisten
// Schaden anrichten kann. Ein Eintrag namens `../../.zshrc` schriebe außerhalb
// des Verzeichnisses, das der Aufrufer genannt hat.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - **eine Datei außerhalb des genannten Verzeichnisses** — der Einbruch
//   - zweimal dieselbe Datei, zwei Ergebnisse
//   - **eine Beschreibung und ein Fehler**, oder keins von beidem
//
//	go test ./internal/sicherung -run Fuzz -fuzz FuzzEntpacken -fuzztime 2m -fuzzminimizetime 5s
//
// **Mit kurzer Minimierzeit, und das ist gemessen** (FMT-06,
// #1695): die Vorgabe minimiert jede neue Eingabe bis zu
// 60 Sekunden, und bei einem Archiv auf der Platte je Durchlauf hält das die
// Arbeiter fest. Zwei Minuten mit der Vorgabe: 2.786 Ausführungen und 12 neue
// Eingaben; mit 5 s: 23.854 und 168. Die Zahlen stehen in `docs/RELEASE.md`.

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// archivsaat baut die Gestalten, mit denen der Fuzzer beginnt.
//
// **Jede Angriffssaat bringt ihr Manifest mit** — Befund 1 des Nachlaufs zum
// Zweitblick über Phase 5.2. Vorher enthielten sie **nur** den bösartigen
// Eintrag, und der hieß obendrein `manifest.json`: [sicherung.Entpacken]
// endete beim Manifestlesen, bevor je ein Dateiname benutzt wurde. Die Saat
// prüfte einen Weg, den sie nie betrat.
//
// **Und der Name steht im Manifest**, nicht nur im Archiv: ausgepackt wird,
// was das Manifest führt (`EntpackenNur`) — ein ZIP-Eintrag, den niemand
// nennt, wird nie angefasst.
func archivsaat(f *testing.F) [][]byte {
	f.Helper()

	saat := make([][]byte, 0, 8)
	saat = append(saat,
		[]byte{},
		[]byte("PK\x05\x06"+string(make([]byte, 18))),
		[]byte("kein ZIP"),
		gebuendelt(f, "ohne.txt", "ein ganz gewöhnlicher Teil"),
	)

	for _, name := range []string{
		"../ausbruch.txt",
		"/absolut.txt",
		"a/../../ausbruch.md",
		"..",
		"a/b/../../../drueber.txt",
	} {
		saat = append(saat, gebuendelt(f, name, "hier sollte nichts landen"))
	}

	return saat
}

// gebuendelt baut eine Sicherung, deren Manifest genau diesen Teil führt.
func gebuendelt(f *testing.F, name, inhalt string) []byte {
	f.Helper()

	summe := sha256.Sum256([]byte(inhalt))

	manifest, fehler := json.Marshal(sicherung.Beschreibung{
		Angaben: sicherung.Angaben{Fassung: "0.5.0", Erstellt: time.Unix(0, 0).UTC()},
		Teile: []sicherung.Teilangabe{{
			Name:    name,
			Groesse: int64(len(inhalt)),
			Summe:   hex.EncodeToString(summe[:]),
		}},
	})
	if fehler != nil {
		f.Fatalf("das Manifest bauen: %v", fehler)
	}

	var puffer bytes.Buffer

	schreiber := zip.NewWriter(&puffer)

	for eintragsname, eintragsinhalt := range map[string]string{
		sicherung.Manifestname: string(manifest),
		name:                   inhalt,
	} {
		eintrag, anlegefehler := schreiber.Create(eintragsname)
		if anlegefehler != nil {
			// **Ein Name, den `archive/zip` nicht schreibt, ist kein Fehlschlag
			// dieses Laufs** — die Saat entfällt, und der Fuzzer findet
			// vergleichbare Gestalten selbst.
			continue
		}

		if _, schreibfehler := eintrag.Write([]byte(eintragsinhalt)); schreibfehler != nil {
			f.Fatalf("%s schreiben: %v", eintragsname, schreibfehler)
		}
	}

	if fehler := schreiber.Close(); fehler != nil {
		f.Fatalf("%s schließen: %v", name, fehler)
	}

	return puffer.Bytes()
}

// ausbrecher sagt, ob das Manifest einen Teil führt, der hinausführt.
//
// **Aus dem Eingabewert gelesen und nicht aus der Antwort** — Befund 1 des
// Nachlaufs: nur so ist prüfbar, dass ein Ausbruch **abgewiesen** wird, statt
// zufällig an einer anderen Prüfung zu scheitern.
func ausbrecher(roh []byte) bool {
	archiv, fehler := zip.NewReader(bytes.NewReader(roh), int64(len(roh)))
	if fehler != nil {
		return false
	}

	offen, fehler := archiv.Open(sicherung.Manifestname)
	if fehler != nil {
		return false
	}

	defer func() { _ = offen.Close() }() //nolint:errcheck // nur gelesen

	inhalt, fehler := io.ReadAll(io.LimitReader(offen, 1<<20))
	if fehler != nil {
		return false
	}

	var beschreibung sicherung.Beschreibung
	if json.Unmarshal(inhalt, &beschreibung) != nil {
		return false
	}

	for _, teil := range beschreibung.Teile {
		sauber := path.Clean(teil.Name)
		if path.IsAbs(teil.Name) || sauber == ".." ||
			strings.HasPrefix(sauber, "../") {
			return true
		}
	}

	return false
}

// **Auspacken schreibt nur dorthin, wo es soll.**
func FuzzEntpacken(f *testing.F) {
	for _, saat := range archivsaat(f) {
		f.Add(saat)
	}

	f.Fuzz(func(t *testing.T, roh []byte) {
		heim := t.TempDir()
		ziel := filepath.Join(heim, "hinein")
		daneben := filepath.Join(heim, "daneben")

		for _, ordner := range []string{ziel, daneben} {
			if fehler := os.MkdirAll(ordner, 0o750); fehler != nil {
				t.Fatalf("%s anlegen: %v", ordner, fehler)
			}
		}

		pfad := filepath.Join(heim, "sicherung.zip")
		if fehler := os.WriteFile(pfad, roh, 0o600); fehler != nil {
			t.Fatalf("die Sicherung schreiben: %v", fehler)
		}

		beschreibung, fehler := sicherung.Entpacken(pfad, ziel)

		nochmal, zweiterFehler := sicherung.Entpacken(pfad, ziel)
		if (fehler == nil) != (zweiterFehler == nil) {
			t.Fatalf("dieselbe Sicherung einmal so und einmal anders: %v gegen %v",
				fehler, zweiterFehler)
		}

		if fehler == nil && !reflect.DeepEqual(beschreibung, nochmal) {
			t.Errorf("dieselbe Sicherung gab zwei Beschreibungen")
		}

		// **Ein Ausbruch wird abgewiesen** — Befund 1 des Nachlaufs zum
		// Zweitblick über Phase 5.2.
		//
		// **Das ist die Zusage aus dem Kopf**, und sie ist am **Eingabewert**
		// zu prüfen: die Kontrolle darunter sähe einen Ausbruch nur, wenn er
		// zufällig im Heim landet. Ein absoluter Name schriebe ganz woanders,
		// und dort sieht niemand nach.
		if ausbrecher(roh) && fehler == nil {
			t.Errorf("ein Teil, der hinausführt, kam durch")
		}

		nichtsDaneben(t, heim, daneben)
	})
}

// nichtsDaneben prüft, dass außerhalb des Ziels nichts entstanden ist.
//
// **Auch im Heim und nicht nur daneben** — Befund 1 des Nachlaufs: der Rumpf
// meldete unerwartete Einträge ausschließlich für `daneben`; ein
// `heim/ausbruch.txt` lief durch die Schleife und wurde nie beanstandet.
func nichtsDaneben(t *testing.T, heim, daneben string) {
	t.Helper()

	erlaubt := map[string]bool{
		"hinein": true, "daneben": true, "sicherung.zip": true,
	}

	for _, ordner := range []string{daneben, heim} {
		eintraege, lesefehler := os.ReadDir(ordner)
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", ordner, lesefehler)
		}

		for _, eintrag := range eintraege {
			if ordner == heim && erlaubt[eintrag.Name()] {
				continue
			}

			t.Errorf("etwas landete in %s: %s", ordner, eintrag.Name())
		}
	}
}
