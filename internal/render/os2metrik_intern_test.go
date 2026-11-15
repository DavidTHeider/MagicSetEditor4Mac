package render

import (
	"encoding/binary"
	"math"
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// **Fehlt die `hhea`-Metrik, gilt `OS/2` typo — nicht der Tintenkasten**
// (#597).
//
// # Der Befund
//
// `MATRIX.TTF` aus `Full-Magic-Pack` nennt in `hhea` asc = 0, desc = 0,
// gap = 0. Der bisherige Rückfall las die **Tintenränder** und ergab 0,957 em
// Zeilenhöhe; **wx rechnet mit 0,860** — gemessen am 2026-08-20 über zwanzig
// Punktgrößen von 9 bis 120, und die Datei nennt in `OS/2` typo genau
// 860/1000.
//
// Für die **Grundlinie** dasselbe: der Tintenkasten gibt 0,779, wx setzt sie
// auf `ceil(0,633 · Punktgröße)` — fünf Größen, 5/5 —, und 633/1000 ist der
// typo-Ascent.
//
// # Warum das eine Anpassung bleibt und keine Portierung
//
// **Im MSE-Quelltext steht diese Regel nicht.** MSE fragt den Schriftdienst
// des Betriebssystems (`wxFont`); die Rückfallkette liegt darunter, in
// wxWidgets und Core Text. Was hier entsteht, bildet **deren** Verhalten nach
// und nicht MSEs Code — Kennzeichnung **adapted** (§15,
// `docs/PORTIERUNGSKARTE.md`).
//
// # Was gemessen ist und was geraten bleibt
//
// **Gemessen** ist, dass wx bei leerer `hhea` die typo-Werte nimmt: für
// `MATRIX.TTF` treffen sie Höhe und Grundlinie, für `mplantin` und
// `beleren-bold` nimmt wx die vorhandene `hhea` (20/20 je Schrift).
//
// **Geraten bleibt der Fall, dass auch `OS/2` nichts hergibt.** Dann greift
// weiter der Tintenkasten, und der ist eine Schätzung — er ist der
// **Tintenumriss** und größer als der typografische Ascent.
func TestOhneHheaGiltDieTypoMetrik(t *testing.T) {
	vorher := masseLesen
	masseLesen = func(*sfnt.Font, *sfnt.Buffer, fixed.Int26_6) (font.Metrics, error) {
		return font.Metrics{}, nil
	}

	t.Cleanup(func() { masseLesen = vorher })

	roh := schriftbytes(t)

	schrift, err := SchriftAusDatei(roh)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// `Andale Mono` nennt in `OS/2` typo asc 1436, desc −412, gap 343 bei
	// em = 2048 — also 2191/2048 Zeilenhöhe und 1436/2048 Grundlinie.
	const (
		erwarteteHoehe      = 2191.0 / 2048
		erwarteteGrundlinie = 1436.0 / 2048
	)

	if math.Abs(schrift.Zeilenhoehe()-erwarteteHoehe) > 1e-9 {
		t.Errorf("Zeilenhöhe %g, erwartet %g — sie muss aus `OS/2` typo kommen",
			schrift.Zeilenhoehe(), erwarteteHoehe)
	}

	if math.Abs(schrift.Grundlinie()-erwarteteGrundlinie) > 1e-9 {
		t.Errorf("Grundlinie %g, erwartet %g — sie muss aus `OS/2` typo kommen",
			schrift.Grundlinie(), erwarteteGrundlinie)
	}
}

// **Was `typoAus` abweist** — die Wege, auf denen der nächste Rückfall dran
// ist.
//
// # Warum das keine Fleißarbeit ist
//
// Eine Schriftdatei ist eine **fremde** Datei (§8), und der Weg hierher geht
// über einen Paketinhalt, den niemand geprüft hat. Ein Verzeichniseintrag, der
// über das Dateiende zeigt, ist kein Absturz, sondern eine Auskunft: diese
// Tabelle gibt es nicht.
func TestWasDieTypoMetrikAbweist(t *testing.T) {
	t.Parallel()

	// kopf baut ein Tabellenverzeichnis mit einem Eintrag.
	kopf := func(name string, stelle, laenge uint32) []byte {
		roh := make([]byte, tabellenkopf+tabelleneintrag)
		binary.BigEndian.PutUint16(roh[4:], 1)
		copy(roh[tabellenkopf:], name)
		binary.BigEndian.PutUint32(roh[tabellenkopf+8:], stelle)
		binary.BigEndian.PutUint32(roh[tabellenkopf+12:], laenge)

		return roh
	}

	faelle := []struct {
		name string
		roh  []byte
	}{
		{name: "gar keine Datei", roh: nil},
		{name: "zu kurz für ein Verzeichnis", roh: make([]byte, 8)},
		{name: "mehr Einträge als Bytes", roh: func() []byte {
			roh := make([]byte, tabellenkopf+4)
			binary.BigEndian.PutUint16(roh[4:], 99)

			return roh
		}()},
		{name: "ein Eintrag zeigt über das Ende", roh: kopf("OS/2", 9000, 100)},
		{name: "kein `head`, nur `OS/2`", roh: kopf("OS/2", 0, 4)},
		{name: "ein Eintrag, der keine von beiden ist", roh: kopf("cmap", 0, 4)},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if _, da := typoAus(fall.roh); da {
				t.Error("angenommen, erwartet war eine Abweisung")
			}
		})
	}
}

// **Eine `head`-Tabelle ohne Größe und eine Metrik ohne Höhe.**
//
// **Getrennt vom Test darüber**, weil hier ein vollständiges Verzeichnis mit
// **zwei** Tabellen nötig ist — die Abweisung passiert erst an den Zahlen.
func TestEineTypoMetrikOhneMassWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	// zwei baut ein Verzeichnis mit `head` und `OS/2` und legt beide dahinter.
	zwei := func(em uint16, aufstieg, abstieg, zwischenraum int16) []byte {
		const anfang = tabellenkopf + 2*tabelleneintrag

		roh := make([]byte, anfang+emStelle+2+typoStelle+6)
		binary.BigEndian.PutUint16(roh[4:], 2)

		copy(roh[tabellenkopf:], "head")
		binary.BigEndian.PutUint32(roh[tabellenkopf+8:], anfang)
		binary.BigEndian.PutUint32(roh[tabellenkopf+12:], emStelle+2)

		os2 := anfang + emStelle + 2
		copy(roh[tabellenkopf+tabelleneintrag:], "OS/2")
		binary.BigEndian.PutUint32(roh[tabellenkopf+tabelleneintrag+8:], uint32(os2))
		binary.BigEndian.PutUint32(roh[tabellenkopf+tabelleneintrag+12:], typoStelle+6)

		binary.BigEndian.PutUint16(roh[anfang+emStelle:], em)
		binary.BigEndian.PutUint16(roh[os2+typoStelle:], uint16(aufstieg))
		binary.BigEndian.PutUint16(roh[os2+typoStelle+2:], uint16(abstieg))
		binary.BigEndian.PutUint16(roh[os2+typoStelle+4:], uint16(zwischenraum))

		return roh
	}

	faelle := []struct {
		name string
		roh  []byte
		da   bool
	}{
		{name: "`unitsPerEm` ist null", roh: zwei(0, 700, -200, 0)},
		{name: "der Aufstieg ist null", roh: zwei(1000, 0, -200, 0)},
		{name: "die Höhe geht nicht auf", roh: zwei(1000, 700, 800, 0)},
		{name: "und eine brauchbare Metrik", roh: zwei(1000, 700, -200, 100), da: true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if _, da := typoAus(fall.roh); da != fall.da {
				t.Errorf("angenommen = %v, erwartet %v", da, fall.da)
			}
		})
	}
}
