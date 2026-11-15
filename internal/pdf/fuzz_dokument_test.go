package pdf_test

// **Der PDF-Schreiber gegen entartete Maße** —
// #1289.
//
// # Warum, obwohl die Eingabe nicht von außen kommt
//
// Weil sie **halb** von außen kommt. Die Seitengröße wählt ein Mensch, das
// Kartenmaß aber steht in einem fremden Stilpaket, und der Bogenleger rechnet
// beides in Punkte um. Eine Karte mit `card width: 1e309` ist ein gültiger
// Eintrag in einem `.mse-style` — und was daraus wird, entscheidet sich hier.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - **gültige Struktur**: was herauskommt, beginnt mit `%PDF-`, endet mit
//     `%%EOF` — und **seine Verweise lösen auf**: `startxref` zeigt auf die
//     Tabelle, die Tabelle nennt so viele Einträge wie der `trailer`, und jede
//     Stelle darin zeigt auf den Kopf des Objekts, das sie meint. Ein PDF, das
//     kein Leser öffnet, ist kein Druckbogen
//
//     Der erste Anlauf suchte fünf Zeichenketten — `%PDF-`, `%%EOF`,
//     `trailer`, `/Root`, `xref` — und nannte das „gültige Struktur". Eine
//     Datei mit falschen Stellen in der Tabelle hätte sie alle getragen; der
//     Kopf versprach mehr, als der Rumpf hielt (Zweitblick über Phase 4.2)
//   - **kein `NaN` und kein `Inf` im Inhaltsstrom** — beides ist in einem PDF
//     ein Syntaxfehler, und `zahl` sagt selbst, dass sie ihn verhindern soll
//   - zweimal dieselbe Eingabe, zwei Ergebnisse
//
// **Nicht zugesichert**: dass die Karte an der richtigen Stelle steht. Das ist
// der Bogenleger, und der misst gegen ein Lineal (`docs/RELEASE.md`,
// Handprüfung 5).
//
// **Und nicht, dass ein fremder Leser die Datei öffnet.** Auflösende Verweise
// sind die Bedingung dafür, nicht der Beweis. Den führt `dokument_test.go` über
// Quartz — mit echten Maßen und einem Aufruf von `sips` je Fall, und deshalb
// nicht je Fuzzausführung.
//
//	go test ./internal/pdf -run Fuzz -fuzz FuzzDokument -fuzztime 2m

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/pdf"
)

// **Der Schreiber gibt immer ein wohlgeformtes PDF.**
func FuzzDokument(f *testing.F) {
	// **Die Saat sind echte Maße** — A4 in Punkten, eine Magic-Karte in
	// Millimetern, und daneben das, was weh tut.
	f.Add(595.28, 841.89, 63.0, 88.0, 1.0)
	f.Add(0.0, 0.0, 0.0, 0.0, 0.0)
	f.Add(-1.0, -1.0, -1.0, -1.0, -1.0)
	f.Add(math.MaxFloat64, math.MaxFloat64, math.MaxFloat64, math.MaxFloat64, 1.0)
	f.Add(math.SmallestNonzeroFloat64, 1.0, 1.0, 1.0, 1.0)
	f.Add(math.Inf(1), math.Inf(-1), math.NaN(), 1.0, 0.5)

	f.Fuzz(func(t *testing.T, breite, hoehe, kartenbreite, kartenhoehe, staerke float64) {
		roh := gebaut(breite, hoehe, kartenbreite, kartenhoehe, staerke)

		wohlgeformt(t, roh)
		verweisePruefen(t, roh)
		ohneUnzahl(t, roh)

		// **Zweimal dasselbe Maß, dasselbe PDF.**
		if nochmal := gebaut(breite, hoehe, kartenbreite, kartenhoehe, staerke); !bytes.Equal(roh, nochmal) {
			t.Error("dieselben Maße gaben zwei verschiedene Dokumente")
		}
	})
}

// gebaut legt einen Bogen mit einer Karte und zwei Linien.
func gebaut(breite, hoehe, kartenbreite, kartenhoehe, staerke float64) []byte {
	dokument := pdf.NeuesDokument()
	seite := dokument.NeueSeite(breite, hoehe)

	seite.Bild(einBild(), 0, 0, kartenbreite, kartenhoehe)
	seite.Linie(0, 0, breite, hoehe, staerke, 0.5)
	seite.Linie(kartenbreite, kartenhoehe, 0, 0, staerke, 0)

	return dokument.Bytes()
}

// wohlgeformt sichert die Gestalt zu, an der ein Leser ein PDF erkennt.
func wohlgeformt(t *testing.T, roh []byte) {
	t.Helper()

	text := string(roh)

	if !strings.HasPrefix(text, "%PDF-") {
		t.Errorf("das Dokument beginnt nicht mit %%PDF-: %q", kopf(text))
	}

	if !strings.HasSuffix(strings.TrimRight(text, "\n"), "%%EOF") {
		t.Errorf("das Dokument endet nicht mit %%%%EOF: %q", schwanz(text))
	}

	for _, muss := range []string{"trailer", "/Root", "xref"} {
		if !strings.Contains(text, muss) {
			t.Errorf("dem Dokument fehlt %q", muss)
		}
	}
}

// verweisePruefen geht die Querverweistabelle durch und löst sie auf.
//
// **Das ist der Unterschied zwischen „sieht aus wie ein PDF" und „ein Leser
// findet seine Objekte".** Ein Leser springt über `startxref` in die Tabelle,
// liest dort je Objekt eine Stelle und springt sie an; stimmt eine davon nicht,
// bricht er ab — und die fünf Zeichenketten aus [wohlgeformt] stünden trotzdem
// alle da.
//
// **Ein vollständiger Zerteiler ist das nicht**, und das gehört gesagt: Ströme,
// Längen, Wörterbücher und der Inhalt der Objekte bleiben ungelesen. Geprüft
// wird die **Adressierung** — und die vollständig: Stelle, Erzeugung und
// Kennzeichen. Was ein Objekt an seiner Stelle *enthält*, ist nicht geprüft.
func verweisePruefen(t *testing.T, roh []byte) {
	t.Helper()

	text := string(roh)

	tabelle, gefunden := stelleNach(t, text, "startxref\n")
	if !gefunden {
		return
	}

	if tabelle < 0 || tabelle+4 > len(roh) || string(roh[tabelle:tabelle+4]) != "xref" {
		t.Errorf("startxref zeigt auf %d — dort steht kein `xref`", tabelle)

		return
	}

	angesagt, gefunden := stelleNach(t, text, "trailer\n<< /Size ")
	if !gefunden {
		return
	}

	// Der Kopf der Tabelle lautet `xref\n0 <Anzahl>\n`.
	geführt, gefunden := stelleNach(t, text[tabelle:], "xref\n0 ")
	if !gefunden {
		return
	}

	if geführt != angesagt {
		t.Errorf("die Tabelle führt %d Einträge, der trailer sagt %d",
			geführt, angesagt)

		return
	}

	stellenPruefen(t, roh, tabelle, geführt)
}

// stellenPruefen springt jede Stelle der Tabelle an.
func stellenPruefen(t *testing.T, roh []byte, tabelle, anzahl int) {
	t.Helper()

	// `xref\n0 <anzahl>\n` und danach der freie Eintrag mit seinen 20 Byte.
	kopf := tabelle + len("xref\n0 \n") + len(strconv.Itoa(anzahl))
	erster := kopf + eintragslaenge

	for nummer := 1; nummer < anzahl; nummer++ {
		anfang := erster + (nummer-1)*eintragslaenge
		if anfang+eintragslaenge > len(roh) {
			t.Errorf("die Tabelle bricht bei Objekt %d ab", nummer)

			return
		}

		eintrag := string(roh[anfang : anfang+eintragslaenge])

		// **Die Gestalt gehört dazu, nicht nur die Zahl.**
		//
		// Der erste Anlauf las die ersten zehn Byte als Stelle und sah den Rest
		// nicht an. Ein Eintrag, dessen Kennzeichen von `n` auf `f` steht, gilt
		// einem Leser als **frei** — das Objekt ist dann nicht adressiert, und
		// die Stelle davor ändert daran nichts. Angestrichen vom Nachlauf über
		// Phase 4.2: geprüft war „die Stelle zeigt zufällig auf einen
		// Objektkopf" und nicht „der Eintrag adressiert dieses Objekt als
		// benutzt".
		if !benutzterEintrag.MatchString(eintrag) {
			t.Errorf("der Eintrag für Objekt %d ist kein benutzter: %q",
				nummer, eintrag)

			return
		}

		stelle, err := strconv.Atoi(eintrag[:10])
		if err != nil {
			t.Errorf("der Eintrag für Objekt %d ist keine Zahl: %q",
				nummer, eintrag[:10])

			return
		}

		erwartet := strconv.Itoa(nummer) + " 0 obj"
		if stelle < 0 || stelle+len(erwartet) > len(roh) ||
			string(roh[stelle:stelle+len(erwartet)]) != erwartet {
			t.Errorf("der Verweis auf Objekt %d zeigt auf %d — dort steht nicht %q",
				nummer, stelle, erwartet)

			return
		}
	}
}

// benutzterEintrag ist die Gestalt eines Eintrags über ein benutztes Objekt.
//
// **Zehn Stellen, ein Leerzeichen, fünf Stellen Erzeugung, `n`, ein
// Leerzeichen, ein Zeilenumbruch** — so schreibt `aufbau.go` sie, und so
// verlangt sie ein Leser. `f` stünde für einen freien Eintrag.
var benutzterEintrag = regexp.MustCompile(`^\d{10} \d{5} n [\r\n]$`)

// eintragslaenge ist die Länge eines Eintrags in der Querverweistabelle.
//
// **Zwanzig Byte, auf das Byte festgelegt** — zehn Stellen, ein Leerzeichen,
// fünf Stellen, ein Leerzeichen, ein Buchstabe, ein Leerzeichen, ein
// Zeilenumbruch. `aufbau.go` schreibt sie so, und ein Leser findet die
// Einträge über diese Länge.
const eintragslaenge = 20

// stelleNach liest die Zahl, die unmittelbar hinter einer Marke steht.
func stelleNach(t *testing.T, text, marke string) (int, bool) {
	t.Helper()

	anfang := strings.LastIndex(text, marke)
	if anfang < 0 {
		t.Errorf("dem Dokument fehlt %q", marke)

		return 0, false
	}

	rest := text[anfang+len(marke):]

	ende := 0
	for ende < len(rest) && rest[ende] >= '0' && rest[ende] <= '9' {
		ende++
	}

	if ende == 0 {
		t.Errorf("hinter %q steht keine Zahl: %q", marke, kopf(rest))

		return 0, false
	}

	zahl, err := strconv.Atoi(rest[:ende])
	if err != nil {
		t.Errorf("hinter %q steht keine lesbare Zahl: %q", marke, rest[:ende])

		return 0, false
	}

	return zahl, true
}

// ohneUnzahl sichert zu, dass keine Unzahl in den Inhaltsstrom gerät.
//
// **`NaN` und `+Inf` sind in einem PDF ein Syntaxfehler**, und `pdf.zahl` sagt
// selbst, dass sie ihn verhindern soll: „ein `1e-05` aus `%v` wäre in einem
// Inhaltsstrom ein Syntaxfehler, den kein Leser meldet — er zeichnet nur
// nichts."
func ohneUnzahl(t *testing.T, roh []byte) {
	t.Helper()

	for _, unzahl := range []string{"NaN", "Inf", "+Inf", "-Inf", "e+", "e-"} {
		if bytes.Contains(roh, []byte(unzahl)) {
			t.Errorf("im Dokument steht %q — das liest kein PDF-Leser", unzahl)
		}
	}
}

// einBild gibt ein winziges Bild — der Inhalt spielt keine Rolle.
func einBild() image.Image {
	bild := image.NewRGBA(image.Rect(0, 0, 2, 2))
	bild.Set(0, 0, color.RGBA{R: 255, A: 255})
	bild.Set(1, 1, color.RGBA{B: 255, A: 255})

	return bild
}

func kopf(text string) string {
	if len(text) > 40 {
		return text[:40]
	}

	return text
}

func schwanz(text string) string {
	if len(text) > 40 {
		return text[len(text)-40:]
	}

	return text
}
