// **Was geschieht, wenn das Schreiben scheitert** —
// #1162.
//
// # Warum innen
//
// Weil der Fall von außen nur über eine volle Platte herzustellen wäre. Ein
// `io.Writer`, der nach n Bytes aufhört, trifft jede Stelle, an der dieses
// Paket schreibt — und ohne ihn wären die Zweige, die den Fehler melden, eine
// Behauptung darüber, was im Ernstfall geschieht (§3).
package sicherung

import (
	"crypto/rand"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// klemmenderSchreiber nimmt `bis` Bytes und scheitert danach.
type klemmenderSchreiber struct {
	bis int
	hat int
}

var errKlemmt = errors.New("die Platte ist voll")

func (k *klemmenderSchreiber) Write(teil []byte) (int, error) {
	frei := k.bis - k.hat
	if frei <= 0 {
		return 0, errKlemmt
	}

	if len(teil) <= frei {
		k.hat += len(teil)

		return len(teil), nil
	}

	k.hat = k.bis

	return frei, errKlemmt
}

// **Ein Schreibfehler wird gemeldet, gleich wo er auftritt.**
//
// Die Grenzen sind so gewählt, dass sie **verschiedene** Stellen treffen: den
// Eintragskopf, den Inhalt eines Teils, und das Manifest ganz am Ende.
func TestEinSchreibfehlerWirdGemeldet(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()

	// **Unkomprimierbar**, damit das Schreiben wirklich stattfindet. Ein
	// Teil aus lauter `x` schrumpft im ZIP auf ein paar Byte, und dann fällt
	// die Grenze nie — der Test prüfte den Puffer und nicht den Weg.
	quelle := filepath.Join(wurzel, "teil")
	if err := os.WriteFile(quelle, zufall(t, 512<<10), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	teile := []Teil{{Name: "teil", Quelle: quelle}}

	// **Die Grenzen werden abgeschritten, nicht geraten.**
	//
	// Ein ZIP schreibt in Stufen — Eintragskopf, gepackter Inhalt, Manifest,
	// Inhaltsverzeichnis am Ende — und `archive/zip` puffert dazwischen. Eine
	// Handvoll gewählter Grenzen trifft mal die eine Stufe und mal die andere,
	// und welche, ändert sich mit jeder Zeile hier. Ein Abschreiten trifft
	// **alle**.
	ganz := gesamtgroesse(t, teile)

	grenzen := []int{ganz - 1, ganz - 40}
	for bis := 0; bis < ganz; bis += ganz / 32 {
		grenzen = append(grenzen, bis)
	}

	// **Und am Ende dicht** — das ist der Nachtrag vom 2026-09-15.
	//
	// Hier stand, ein Abschreiten in 32 Stufen treffe **alle** Stufen des
	// Archivs. Es traf sie nicht: Manifest und Inhaltsverzeichnis stehen
	// **hinter** dem gepackten Teil, zusammen ein paar hundert Byte in einem
	// Archiv von einem halben Megabyte. Ob eine Stufe von 16 KB darin landet,
	// ist Glück — und das Glück hat mit dem Aufstieg auf Go 1.27 aufgehört:
	// `archive/zip` puffert anders, die alten Grenzen fielen daneben, und
	// zwei Zweige in `archivschreiben.go` standen als unbetreten da.
	//
	// **Ein Kopf, der mehr verspricht als sein Rumpf hält** (§10) — und er
	// hat es ein halbes Jahr lang getan, ohne dass es auffiel, weil es
	// zufällig passte.
	for bis := ganz - schwanz; bis < ganz; bis += schwanzschritt {
		if bis > 0 {
			grenzen = append(grenzen, bis)
		}
	}

	for _, bis := range grenzen {
		t.Run(itoa(bis), func(t *testing.T) {
			t.Parallel()

			if fehler := hineinschreiben(&klemmenderSchreiber{bis: bis}, Manifestname, teile, Angaben{}); fehler == nil {
				t.Errorf("bei %d Byte kam das Schreiben durch", bis)
			}
		})
	}
}

// Das dichte Ende: wie weit zurück und in welchen Schritten.
//
// **Vier Kilobyte in Schritten von sechzehn** — 256 Grenzen, und sie kosten
// zusammen weniger als eine Sekunde, weil jede nur bis zum Abbruch schreibt.
// Das Fenster von Manifest und Inhaltsverzeichnis ist größer als sechzehn
// Byte; damit liegt mindestens eine Grenze darin, gleich wie `archive/zip`
// puffert.
const (
	schwanz        = 4 << 10
	schwanzschritt = 16
)

// gesamtgroesse sagt, wie viele Bytes ein vollständiges Archiv braucht.
func gesamtgroesse(t *testing.T, teile []Teil) int {
	t.Helper()

	var zaehler zaehlenderSchreiber
	if fehler := hineinschreiben(&zaehler, Manifestname, teile, Angaben{}); fehler != nil {
		t.Fatalf("messen: %v", fehler)
	}

	return zaehler.hat
}

// zaehlenderSchreiber nimmt alles und merkt sich, wie viel es war.
type zaehlenderSchreiber struct{ hat int }

func (z *zaehlenderSchreiber) Write(teil []byte) (int, error) {
	z.hat += len(teil)

	return len(teil), nil
}

// zufall gibt unkomprimierbare Bytes.
func zufall(t *testing.T, wie int) []byte {
	t.Helper()

	roh := make([]byte, wie)
	if _, err := rand.Read(roh); err != nil {
		t.Fatalf("Zufall: %v", err)
	}

	return roh
}

// itoa gibt eine kleine Zahl als Text — ohne `strconv` für einen Testnamen.
func itoa(zahl int) string {
	if zahl == 0 {
		return "0"
	}

	var stellen []byte
	for zahl > 0 {
		stellen = append([]byte{byte('0' + zahl%10)}, stellen...)
		zahl /= 10
	}

	return string(stellen)
}

// **Eine Datei, die es nicht gibt, ist kein Archiv.**
func TestEineFehlendeDateiIstKeinArchiv(t *testing.T) {
	t.Parallel()

	if IstArchiv(filepath.Join(t.TempDir(), "gibtsnicht")) {
		t.Error("eine fehlende Datei gilt als Archiv")
	}
}

// **Und eine zu kurze ebenso.**
//
// **Ohne diesen Fall wäre `io.ReadFull` ein Zweig, den niemand betritt** —
// und eine leere Datei ist genau das, was eine abgebrochene Sicherung
// hinterlässt.
func TestEineZuKurzeDateiIstKeinArchiv(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "kurz")
	if err := os.WriteFile(pfad, []byte("PK"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if IstArchiv(pfad) {
		t.Error("zwei Bytes gelten als Archiv")
	}
}

// **Ein Manifest, das nicht anzulegen ist** — der Nachtrag vom 2026-09-15.
//
// # Warum es diesen Lauf braucht
//
// **Weil der Weg über den klemmenden Schreiber aufgehört hat zu tragen.** Mit
// dem Aufstieg auf Go 1.27 puffert `archive/zip` das kleine Manifest bis zum
// `Close`; der Fehler kommt dann dort an und nicht in `manifestSchreiben`.
// Gemessen am 2026-09-15: von **524.734** Grenzen trifft **keine einzige**
// die Manifeststelle — vorher taten es einige, und zwar zufällig.
//
// **Eine Abdeckung, die an einer Pufferung hängt, ist keine Zusage.** Dieser
// Lauf hängt an nichts: `archive/zip` weist einen zweiten Eintrag mit
// demselben Namen ab, und `hineinschreiben` nimmt den Manifestnamen als
// Argument entgegen.
func TestEinManifestDasNichtAnzulegenIstWirdGemeldet(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()

	quelle := filepath.Join(wurzel, "teil")
	if err := os.WriteFile(quelle, []byte("inhalt"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	teile := []Teil{{Name: "teil", Quelle: quelle}}

	// **Ein Zeitpunkt, den JSON nicht schreiben kann.**
	//
	// `time.Time.MarshalJSON` gibt einen Fehler, sobald das Jahr außerhalb
	// von [0, 9999] liegt — das steht so in der Standardbibliothek und
	// hängt an keiner Pufferung.
	//
	// **Der erste Anlauf ging über einen doppelten Eintragsnamen**, in der
	// Annahme, `archive/zip` weise ihn ab. Es tut es nicht: geprüft wird
	// die Zeigergleichheit des `FileHeader`, nicht der Name. Eine Annahme
	// über fremdes Verhalten, die nicht nachgelesen war (§10) — der Lauf
	// war rot und hat sie widerlegt.
	zuWeit := Angaben{Erstellt: time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)}

	fehler := hineinschreiben(io.Discard, Manifestname, teile, zuWeit)
	if fehler == nil {
		t.Fatal("ein unschreibbarer Zeitpunkt kam durch")
	}

	if !strings.Contains(fehler.Error(), "Manifest") {
		t.Errorf("der Fehler nennt das Manifest nicht: %v", fehler)
	}
}
