// **Ein abgebrochenes Schreiben lässt kein halbes Paket zurück** —
// #1289,
// Werkstück „Fehlerinjektion: abgebrochenes Schreiben, halb geschriebene Sets".
//
// # Warum ein Durchlauf über alle Abbruchstellen und nicht ein Fall
//
// `austausch_intern_test.go` prüft [austauschen] für sich: scheitert der zweite
// Zug, holt der dritte zurück. Das ist die **Buchführung** einer Funktion.
//
// Was die Abnahme aus #1289 verlangt, ist etwas anderes: „Beschädigte Daten
// führen zu benannten Teilbefunden oder zu einer klaren Ablehnung — **nie zu
// einem stillen halben Stand**." Das ist eine Aussage über das **Paket auf der
// Platte**, nachdem der ganze Weg abgebrochen wurde — Sperre, Sicherung,
// Nachbardatei, Austausch und Aufräumen zusammen. Ein Weg kann an jeder
// einzelnen Stelle richtig sein und trotzdem als Ganzes etwas hinterlassen.
//
// Deshalb läuft hier **jede erreichbare Abbruchstelle** gegen **beide**
// Paketformen, und geprüft wird nicht der Rückgabewert, sondern was danach im
// Verzeichnis liegt.
//
// # Die drei Gestalten, und warum es genau drei sind
//
// [austauschen] benennt bis zu dreimal um. Ein Zähler, der beim `n`-ten Zug
// abbricht, erreicht davon nur zwei: bricht der **erste** ab, gibt es keinen
// zweiten; gelingt der zweite, gibt es keinen dritten. Der dritte Zug ist die
// **Rücknahme** und findet nur statt, wenn der zweite scheiterte. Also:
//
//	{1}    — der erste Zug scheitert, nichts ist angefasst
//	{2}    — der zweite scheitert, die Rücknahme gelingt
//	{2, 3} — auch die Rücknahme scheitert
//
// # Was hier **nicht** zugesichert wird
//
// Ein Fehlschlag beim **Schreiben** der Nachbardatei. Hier stand, er sei auf
// einer gesunden Platte nicht herbeizuführen — das war eine Aussage über
// `os.CreateTemp` und nicht über den Weg. Der Zweitblick über Phase 4.2 hat es
// angestrichen; seitdem ist die Nachbardatei eine Schnittstelle, und
// `ablegeabbruch_intern_test.go` bricht an jedem ihrer Griffe ab. **Dieser Lauf
// prüft weiterhin nur die Umbenennungen**, und die Aufteilung ist die
// Fragestellung: hier der Tausch, dort das Schreiben.
//
// Und die **Stromausfallsicherheit**. `ablegen` ruft [os.File.Sync], und was
// danach im Journal des Dateisystems steht, entscheidet nicht dieser Lauf.
package msepaket

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// errAbbruch ist der Fehlschlag, den dieser Lauf einspeist.
var errAbbruch = errors.New("das Dateisystem hat abgelehnt")

// abbrechenderUmbenenner benennt um — außer an den genannten Zügen.
//
// **Er tut sonst das Echte.** Ein Umbenenner, der gar nichts täte, prüfte einen
// Weg, den es nicht gibt: die Wirkung auf der Platte ist genau das, was hier
// zur Debatte steht.
type abbrechenderUmbenenner struct {
	scheitertBei []int
	gezaehlt     int
}

func (a *abbrechenderUmbenenner) benenne(alt, neu string) error {
	a.gezaehlt++

	if slices.Contains(a.scheitertBei, a.gezaehlt) {
		return errAbbruch
	}

	return os.Rename(alt, neu)
}

// abbruchstellen sind die drei erreichbaren Gestalten eines Fehlschlags.
var abbruchstellen = []struct {
	name    string
	stellen []int

	// ganz sagt, ob das Paket danach vollständig dasteht.
	ganz bool
}{
	{"der erste Zug scheitert", []int{1}, true},
	{"der zweite scheitert, die Rücknahme gelingt", []int{2}, true},
	{"auch die Rücknahme scheitert", []int{2, 3}, false},
}

// **Nach einem abgebrochenen Schreiben ist das Paket ganz — oder die Meldung
// sagt, wo der vorige Stand liegt.**
func TestEinAbgebrochenesSchreibenLaesstKeinHalbesPaketZurueck(t *testing.T) {
	t.Parallel()

	for _, form := range paketformen() {
		for _, abbruch := range abbruchstellen {
			t.Run(form.name+", "+abbruch.name, func(t *testing.T) {
				t.Parallel()

				ort := form.bauen(t)
				vorher := paketstand(t, ort)

				umbenenner := &abbrechenderUmbenenner{scheitertBei: abbruch.stellen}

				// **Über [echteHilfen] und nicht über ein frisches Literal**:
				// ein Griff, den niemand setzt, ist `nil`, und der Lauf stürbe
				// an einer Panik statt an seiner Zusicherung.
				hilfen := echteHilfen()
				hilfen.umbenennen = umbenenner.benenne

				fehler := dateiSchreiben(context.Background(), ort, form.hauptdatei,
					[]byte("mse version: 2.0.0\ngame: anders\n"),
					summeVon(vorher[form.hauptdatei]), hilfen)

				if !errors.Is(fehler, errAbbruch) {
					t.Fatalf("der Abbruch wurde nicht durchgereicht: %v", fehler)
				}

				if abbruch.ganz {
					standVergleichen(t, vorher, paketstand(t, ort))
				} else {
					meldungPruefen(t, fehler, ort, form.hauptdatei, vorher)
				}

				ohneNachbardatei(t, ort)
			})
		}
	}
}

// meldungPruefen sichert zu, dass die Ablehnung sagt, wo der Stand liegt.
//
// **Das ist die zweite Hälfte der Abnahme.** Wenn das Paket nach drei
// gescheiterten Zügen nicht vollständig ist, darf das kein *stiller* halber
// Stand sein: die Meldung nennt den Ort, an dem der vorige Inhalt liegt, und
// erst damit ist sie eine klare Ablehnung.
func meldungPruefen(t *testing.T, fehler error, ort, hauptdatei string, vorher map[string][]byte) {
	t.Helper()

	if !strings.Contains(fehler.Error(), Sicherungsendung) {
		t.Errorf("die Meldung nennt den Rückweg nicht: %v", fehler)
	}

	if _, statfehler := os.Stat(ort + Sicherungsendung); statfehler != nil {
		t.Errorf("die Meldung nennt einen Rückweg, den es nicht gibt: %v", statfehler)

		return
	}

	// **Und dort steht der vorige Stand, nicht irgendetwas.**
	//
	// Der erste Anlauf prüfte nur, dass es den Rückweg **gibt**. Ein Rückweg,
	// der auf eine leere oder eine schon überschriebene Datei zeigt, wäre
	// genau der stille halbe Stand, gegen den dieser Lauf steht — er sähe nur
	// aus wie eine Rettung. Angestrichen vom Zweitblick über Phase 4.2.
	gerettet := gerettetesLesen(t, ort, hauptdatei)
	if !bytes.Equal(gerettet, vorher[hauptdatei]) {
		t.Errorf("unter %q steht nicht der vorige Stand: %q gegen %q",
			ort+Sicherungsendung, gerettet, vorher[hauptdatei])
	}
}

// gerettetesLesen holt die Hauptdatei aus der Sicherung.
//
// **Beide Formen haben sie an verschiedenen Orten**: beim Verzeichnispaket
// liegt sie als Datei **in** `<ort>.bak`, beim Archiv **ist** `<ort>.bak` das
// vorige Archiv. Das ist keine Unsauberkeit, sondern folgt daraus, wohin
// `austauschen` das Ziel schiebt — beim Verzeichnis die einzelne Datei, beim
// Archiv das ganze Paket.
func gerettetesLesen(t *testing.T, ort, hauptdatei string) []byte {
	t.Helper()

	sicherung := ort + Sicherungsendung

	angaben, err := os.Stat(sicherung)
	if err != nil {
		t.Fatalf("die Sicherung ansehen: %v", err)
	}

	if angaben.IsDir() {
		inhalt, lesefehler := os.ReadFile(filepath.Join(sicherung, hauptdatei)) // #nosec G304 -- aus t.TempDir()
		if lesefehler != nil {
			t.Fatalf("die gesicherte Datei lesen: %v", lesefehler)
		}

		return inhalt
	}

	// **Direkt ins Archiv und nicht über [Oeffne]**: `probe.mse-style.bak`
	// trägt keine Paketendung mehr, und genau das ist Absicht — [ArtAusName]
	// soll die Sicherung nicht für ein Paket halten.
	archiv, oeffnefehler := zip.OpenReader(sicherung)
	if oeffnefehler != nil {
		t.Fatalf("die gesicherte Archivform öffnen: %v", oeffnefehler)
	}

	defer func() { _ = archiv.Close() }() //nolint:errcheck // nur gelesen

	inhalt, lesefehler := gelesen(&archiv.Reader, hauptdatei)
	if lesefehler != nil {
		t.Fatalf("die gesicherte Datei lesen: %v", lesefehler)
	}

	return inhalt
}

// ohneNachbardatei sichert zu, dass keine Nachbardatei liegen bleibt.
//
// **Sie wäre der halbe Stand in seiner sichtbarsten Form**: eine
// `.mse-schreibt-…` im Paketverzeichnis zählt [Paket.Dateien] mit, und bei
// einer `.mse-symbol-font`, deren Dateien alle etwas bedeuten, ist das kein
// Schönheitsfehler — derselbe Grund, aus dem die Sicherung neben das Paket
// wandert und nicht hinein.
func ohneNachbardatei(t *testing.T, ort string) {
	t.Helper()

	for _, wo := range []string{ort, filepath.Dir(ort)} {
		eintraege, err := os.ReadDir(wo)
		if err != nil {
			continue // ein Archiv ist kein Verzeichnis; der Nachbar daneben zählt
		}

		for _, eintrag := range eintraege {
			if strings.HasPrefix(eintrag.Name(), ".mse-schreibt-") {
				t.Errorf("in %q liegt die Nachbardatei %q noch", wo, eintrag.Name())
			}
		}
	}
}

// paketstand liest jede Datei eines Pakets.
//
// **Über [Oeffne] und nicht über [os.ReadDir]**: was zählt, ist der Stand, den
// ein Leser sieht — für ein Archiv ist das der Eintrag, nicht die Datei.
func paketstand(t *testing.T, ort string) map[string][]byte {
	t.Helper()

	paket, err := Oeffne(ort)
	if err != nil {
		t.Fatalf("das Paket ließ sich nach dem Abbruch nicht öffnen: %v", err)
	}

	defer func() { _ = paket.Schliessen() }() //nolint:errcheck // nur gelesen

	namen, err := paket.Dateien()
	if err != nil {
		t.Fatalf("die Dateien ließen sich nicht aufzählen: %v", err)
	}

	stand := make(map[string][]byte, len(namen))

	for _, name := range namen {
		inhalt, lesefehler := paket.Datei(name)
		if lesefehler != nil {
			t.Fatalf("%s ließ sich nicht lesen: %v", name, lesefehler)
		}

		stand[name] = inhalt
	}

	return stand
}

// standVergleichen sichert zu, dass sich nichts verändert hat.
func standVergleichen(t *testing.T, vorher, nachher map[string][]byte) {
	t.Helper()

	for name, alt := range vorher {
		neu, da := nachher[name]
		if !da {
			t.Errorf("%s fehlt nach dem Abbruch", name)

			continue
		}

		if !bytes.Equal(alt, neu) {
			t.Errorf("%s hat sich verändert: %q → %q", name, alt, neu)
		}
	}

	for name := range nachher {
		if _, da := vorher[name]; !da {
			t.Errorf("nach dem Abbruch liegt %q im Paket, das vorher nicht da war", name)
		}
	}
}

// paketformen gibt die beiden Gestalten eines Pakets.
func paketformen() []paketform {
	return []paketform{
		{name: "Verzeichnis", hauptdatei: "style", bauen: verzeichnisPaket},
		{name: "Archiv", hauptdatei: "style", bauen: archivPaket, archiv: true},
	}
}

// paketform ist eine der beiden Gestalten samt ihrem Bauer.
type paketform struct {
	name       string
	hauptdatei string
	bauen      func(*testing.T) string

	// archiv sagt, ob diese Form ein Archiv offen hält — nur dann gibt es
	// einen Griff, der es wieder freigibt.
	archiv bool
}

// paketinhalt ist der Inhalt beider Probepakete.
//
// **Zwei Dateien und ein Unterordner**: ein Paket mit einer einzigen Datei
// verbirgt, ob ein Abbruch die **anderen** Einträge mitnimmt — und beim Archiv
// wird beim Schreiben jeder von ihnen übernommen.
var paketinhalt = map[string]string{
	"style":              "mse version: 2.0.0\ngame: probe\n",
	"unterordner/skript": "1 + 1\n",
}

// verzeichnisPaket legt ein Verzeichnispaket an.
func verzeichnisPaket(t *testing.T) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "probe.mse-style")
	if err := os.MkdirAll(filepath.Join(ort, "unterordner"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	for name, inhalt := range paketinhalt {
		pfad := filepath.Join(ort, filepath.FromSlash(name))
		if err := os.WriteFile(pfad, []byte(inhalt), 0o600); err != nil {
			t.Fatalf("schreiben: %v", err)
		}
	}

	return ort
}

// archivPaket legt ein Archivpaket an.
func archivPaket(t *testing.T) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "probe.mse-style")

	datei, err := os.Create(ort) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	packer := zip.NewWriter(datei)

	for name, inhalt := range paketinhalt {
		strom, fehler := packer.Create(name)
		if fehler != nil {
			t.Fatalf("Eintrag: %v", fehler)
		}

		if _, fehler := strom.Write([]byte(inhalt)); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}
	}

	if fehler := errors.Join(packer.Close(), datei.Close()); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	return ort
}

// summeVon gibt die SHA-256 eines Inhalts, wie [DateiSchreiben] sie erwartet.
func summeVon(inhalt []byte) string {
	roh := sha256.Sum256(inhalt)

	return hex.EncodeToString(roh[:])
}
