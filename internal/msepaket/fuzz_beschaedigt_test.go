package msepaket_test

// **Ein beschädigtes Paket wird benannt abgelehnt oder benannt gelesen** —
// #1289,
// Werkstück „Fehlerinjektion: abgebrochenes Schreiben, halb geschriebene Sets".
//
// # Die andere Hälfte
//
// `fehlerinjektion_intern_test.go` prüft die **schreibende** Seite: was liegt
// da, wenn der Schreiber abbricht. Hier steht die **lesende**: was macht
// [msepaket.Oeffne] aus dem, was ein abgebrochener Schreiber — unserer oder ein
// fremder — hinterlassen hat.
//
// Beide Formen kommen dran, und sie können nicht dasselbe:
//
//   - als **Archiv** trägt jeder Eintrag eine CRC32. Ein verändertes Byte ist
//     damit *nachweisbar*, und genau das prüft der Lauf unten mit einer
//     **eigenen** Prüfsumme statt mit dem Wort des Lesers.
//   - als **Verzeichnis** ist eine abgeschnittene Textdatei von einer kürzeren
//     heilen nicht zu unterscheiden. Kein Byte darin sagt, dass etwas fehlt.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - zweimal dieselbe Eingabe, zwei Ergebnisse
//   - ein Paket **und** ein Fehler, oder keins von beidem
//   - ein Befund ohne Meldung
//   - **[msepaket.Paket.Datei] gibt Bytes, deren CRC32 dem Archiv
//     widerspricht** — das wäre der stille halbe Stand in Reinform: der Leser
//     meldet Erfolg und liefert etwas anderes, als dasteht
//   - **ein Name aus [msepaket.Paket.Dateien], den [msepaket.Paket.Datei] mit
//     einem leeren Fehler abweist** — abgelehnt ist in Ordnung, stumm nicht
//
// # Was **nicht** zugesichert wird
//
// Dass ein abgeschnittenes **Verzeichnispaket** auffällt. Es fällt nicht auf,
// und das ist keine Nachlässigkeit, sondern die Eigenschaft des Formats: `set`
// ist eine Textdatei ohne Länge und ohne Prüfsumme. Wer diese Zusage haben
// will, braucht ein Archiv — und `docs/FORMATE.md` sagt, dass ein `.mse-set` im
// Betrieb eines ist.
//
//	go test ./internal/msepaket -run Fuzz -fuzz FuzzBeschaedigtesPaket -fuzztime 2m -fuzzminimizetime 5s
//
// **Mit kurzer Minimierzeit, und das ist gemessen** (FMT-06,
// #1695):
// die Vorgabe minimiert jede neue Eingabe bis zu 60 Sekunden, und bei einem
// Paket auf der Platte je Durchlauf hält das die Arbeiter fest. Eine Minute mit
// der Vorgabe: 2.561 Ausführungen und 8 neue Eingaben; zwei Minuten mit 5 s:
// 50.411 und 192. Die Zahlen stehen in `docs/RELEASE.md`.

import (
	"archive/zip"
	"bytes"
	"hash/crc32"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein beschädigtes Paket führt nie zu einem stillen halben Stand.**
func FuzzBeschaedigtesPaket(f *testing.F) {
	heil := heilesArchiv(f)

	f.Add(heil)
	f.Add(heil[:len(heil)/2])
	f.Add(heil[:len(heil)-1])
	f.Add(mitVerdrehtemByte(heil))
	f.Add(mitVerdrehterBeigabe(heil))
	f.Add(mitDoppeltemNamen(f))
	f.Add([]byte("mse version: 2.0.0\ngame: probe\ncard:\n\tname: A\n"))
	f.Add([]byte("mse version: 2.0.0\ngame: probe\ncard:\n\tname: A\ncard:\n\tna"))
	f.Add([]byte{})
	f.Add([]byte("PK\x03\x04"))

	f.Fuzz(func(t *testing.T, roh []byte) {
		alsArchiv(t, roh)
		alsVerzeichnis(t, roh)
	})
}

// alsArchiv legt die Bytes als `.mse-set` hin und liest sie.
func alsArchiv(t *testing.T, roh []byte) {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "probe.mse-set")
	if err := os.WriteFile(ort, roh, 0o600); err != nil {
		t.Fatalf("hinlegen: %v", err)
	}

	if paket := geoeffnet(t, ort); paket != nil {
		defer func() { _ = paket.Schliessen() }() //nolint:errcheck // nur gelesen

		dateienPruefen(t, paket, crcTafel(roh))
		zweimalDieselbenDateien(t, ort)
	}
}

// alsVerzeichnis legt die Bytes als Hauptdatei eines Verzeichnispakets hin.
//
// **Dieselben Bytes in der anderen Form.** Ein Leser, der beide Gestalten
// desselben Gegenstands verschieden behandelt, schreibt jede Auswertung zweimal
// — und die seltener benutzte fällt zuerst um (`paket.go`, [msepaket.Oeffne]).
func alsVerzeichnis(t *testing.T, roh []byte) {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "probe.mse-set")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ort, "set"), roh, 0o600); err != nil {
		t.Fatalf("hinlegen: %v", err)
	}

	if paket := geoeffnet(t, ort); paket != nil {
		defer func() { _ = paket.Schliessen() }() //nolint:errcheck // nur gelesen

		dateienPruefen(t, paket, nil)
		zweimalDieselbenDateien(t, ort)
	}
}

// geoeffnet öffnet ein Paket und hält die Zusagen über die Antwort selbst.
//
// Gibt `nil`, wenn es abgelehnt wurde — eine Ablehnung ist kein Fund.
func geoeffnet(t *testing.T, ort string) *msepaket.Paket {
	t.Helper()

	paket, fehler := msepaket.Oeffne(ort)

	if fehler != nil {
		if paket != nil {
			t.Errorf("abgelehnt **und** ein Paket zurück: %v", fehler)
		}

		return nil
	}

	if paket == nil {
		t.Fatal("kein Fehler und kein Paket — das ist keine Antwort")
	}

	for _, befund := range paket.Befunde {
		if befund.Meldung == "" {
			t.Error("ein Befund ohne Meldung — er sagt nicht, was los ist")
		}
	}

	// **Zweimal derselbe Ort, dieselbe Antwort.**
	nochmal, nochmalFehler := msepaket.Oeffne(ort)
	if nochmalFehler != nil {
		t.Fatalf("dasselbe Paket wurde beim zweiten Mal abgelehnt: %v", nochmalFehler)
	}

	defer func() { _ = nochmal.Schliessen() }() //nolint:errcheck // nur gelesen

	// **Die Befunde und die Dateiliste, nicht ihre Längen.** Der erste Anlauf
	// verglich zwei Zahlen; damit durfte dasselbe Paket beim zweiten Öffnen
	// andere Befunde über andere Stellen melden, solange es gleich viele waren
	// (Zweitblick über Phase 4.2).
	if !reflect.DeepEqual(paket.Befunde, nochmal.Befunde) {
		t.Errorf("dasselbe Paket gab zwei Befundlisten: %d gegen %d Einträge",
			len(paket.Befunde), len(nochmal.Befunde))
	}

	ersteNamen, ersterFehler := paket.Dateien()
	zweiteNamen, zweiterFehler := nochmal.Dateien()

	if (ersterFehler == nil) != (zweiterFehler == nil) {
		t.Errorf("dasselbe Paket zählte einmal auf und einmal nicht: %v / %v",
			ersterFehler, zweiterFehler)
	}

	if !slices.Equal(ersteNamen, zweiteNamen) {
		t.Errorf("dasselbe Paket gab zwei Dateilisten: %v gegen %v",
			ersteNamen, zweiteNamen)
	}

	return paket
}

// dateienPruefen liest jede Datei und hält sie gegen die Prüfsummen.
//
// **`tafel` ist `nil`, wo es keine gibt** — ein Verzeichnispaket trägt keine.
// Dann bleibt die schwächere Zusage: abgewiesen wird benannt, nicht stumm.
func dateienPruefen(t *testing.T, paket *msepaket.Paket, tafel map[string]uint32) {
	t.Helper()

	namen, fehler := paket.Dateien()
	if fehler != nil {
		// **Eine Teilliste ist gemeldet, und das ist die Zusage** — sie darf
		// nur nicht stumm sein.
		//
		// Hierher kommt unter anderem das Archiv mit zwei gleichnamigen
		// Einträgen: Gos ZIP-Dateisystem weist es beim Gehen mit
		// „duplicate entries in zip file" ab, gemessen am 2026-08-29. Die
		// Prüfsummentafel wird für so ein Archiv also gar nicht erst befragt
		// — was sie für doppelte Namen tut, ist trotzdem festgelegt und nicht
		// dem Zufall überlassen (siehe [crcTafel]).
		if fehler.Error() == "" {
			t.Error("die Dateien ließen sich stumm nicht aufzählen")
		}

		return
	}

	for _, name := range namen {
		inhalt, lesefehler := paket.Datei(name)
		if lesefehler != nil {
			if lesefehler.Error() == "" {
				t.Errorf("%q wurde stumm abgewiesen", name)
			}

			continue
		}

		erwartet, da := tafel[name]
		if !da {
			continue
		}

		if gefunden := crc32.ChecksumIEEE(inhalt); gefunden != erwartet {
			t.Errorf("%q kam mit CRC %08x zurück, das Archiv nennt %08x",
				name, gefunden, erwartet)
		}
	}
}

// crcTafel liest die Prüfsummen, die das Archiv selbst führt.
//
// **Der eigene Leser und nicht der geprüfte.** Ein Oracle, das durch dieselbe
// Tür geht wie das Geprüfte, prüft nichts.
func crcTafel(roh []byte) map[string]uint32 {
	archiv, err := zip.NewReader(bytes.NewReader(roh), int64(len(roh)))
	if err != nil {
		return nil
	}

	// **Erst zählen, dann eintragen.** Ein doppelter Name bekommt keine
	// Prüfsumme: welchen Eintrag ein Leser dann nimmt, ist nicht festgelegt,
	// und eine Zusage darüber wäre erfunden.
	//
	// Der erste Anlauf schrieb dafür eine `0` in die Tafel — und `0` ist selbst
	// eine mögliche CRC. `dateienPruefen` las sie als echte Erwartung und
	// meldete für genau die Eingabeklasse einen Fund, die der Kopf ausdrücklich
	// ausnimmt. Angestrichen vom Zweitblick über Phase 4.2; ein Sentinel aus dem
	// Wertebereich ist keiner.
	gezaehlt := map[string]int{}
	for _, eintrag := range archiv.File {
		if !eintrag.FileInfo().IsDir() {
			gezaehlt[eintrag.Name]++
		}
	}

	tafel := make(map[string]uint32, len(archiv.File))

	for _, eintrag := range archiv.File {
		if gezaehlt[eintrag.Name] == 1 {
			tafel[eintrag.Name] = eintrag.CRC32
		}
	}

	return tafel
}

// zweimalDieselbenDateien sichert zu, dass zwei Öffnungen dasselbe hergeben.
//
// **Die Antworten selbst und nicht ihre Namen.** [geoeffnet] vergleicht Befunde
// und Dateiliste; das ließ zu, dass derselbe Ort beim zweiten Öffnen **andere
// Bytes** liefert, solange Namen und Befunde gleich bleiben. Angestrichen vom
// Nachlauf über Phase 4.2 — zwei Zähler waren durch zwei Listen ersetzt worden,
// die vollständige Antwort war es nicht.
func zweimalDieselbenDateien(t *testing.T, ort string) {
	t.Helper()

	erste, zweite := dateiantworten(t, ort), dateiantworten(t, ort)

	if len(erste) != len(zweite) {
		t.Errorf("dasselbe Paket gab %d und dann %d Dateiantworten",
			len(erste), len(zweite))

		return
	}

	for name, antwort := range erste {
		andere, da := zweite[name]
		if !da {
			t.Errorf("%q kam nur beim ersten Öffnen", name)

			continue
		}

		if antwort.meldung != andere.meldung {
			t.Errorf("%q gab zwei Antworten: %q gegen %q",
				name, antwort.meldung, andere.meldung)

			continue
		}

		if !bytes.Equal(antwort.inhalt, andere.inhalt) {
			t.Errorf("%q gab zweimal verschiedene Bytes: %d gegen %d",
				name, len(antwort.inhalt), len(andere.inhalt))
		}
	}
}

// dateiantwort ist, was [msepaket.Paket.Datei] für einen Namen hergibt.
type dateiantwort struct {
	inhalt []byte

	// meldung ist die Ablehnung, leer bei Erfolg — der Text und nicht der
	// Fehler: zwei Fehlerwerte sind nie gleich, zwei Meldungen schon.
	meldung string
}

// dateiantworten liest jede Datei eines frisch geöffneten Pakets.
func dateiantworten(t *testing.T, ort string) map[string]dateiantwort {
	t.Helper()

	paket, fehler := msepaket.Oeffne(ort)
	if fehler != nil {
		t.Fatalf("dasselbe Paket ließ sich nicht wieder öffnen: %v", fehler)
	}

	defer func() { _ = paket.Schliessen() }() //nolint:errcheck // nur gelesen

	namen, listenfehler := paket.Dateien()
	if listenfehler != nil {
		return nil
	}

	antworten := make(map[string]dateiantwort, len(namen))

	for _, name := range namen {
		inhalt, lesefehler := paket.Datei(name)
		if lesefehler != nil {
			antworten[name] = dateiantwort{meldung: lesefehler.Error()}

			continue
		}

		antworten[name] = dateiantwort{inhalt: inhalt}
	}

	return antworten
}
