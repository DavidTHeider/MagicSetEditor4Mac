//go:build fremdpakete

// **Wie viel Tinte unsere Glyphen setzen, gemessen gegen MSE2**
// (#446).
//
// # Die Zahl, die hier überprüft wird
//
// Am 2026-08-18 über 44 Karten aus `White` gemessen: **132.090 gegen 155.689**
// — unsere Glyphen decken **85 %** dessen, was MSE2 auf diesem Mac deckt, und
// der Unterschied sitzt fast ganz in der obersten Deckungsstufe. MSE2 hatte
// dort **63 % mehr** volldeckende Bildpunkte.
//
// **`docs/ROADMAP.md` führte sie bis zum 2026-08-20 als überholt** — die
// fehlenden 15 % seien nicht die Glättung gewesen, sondern der fehlende
// **Ausweich auf eine zweite Schrift**, gebaut seit
// #515.
//
// **Nachgemessen war sie das nicht.** Über **671** Karten aus `White` statt 44
// standen danach **86,3 %**, und die Verteilung hatte dieselbe Gestalt wie
// damals: in den unteren Stufen gleichauf oder darüber, in der obersten bei
// **61,9 %** von MSE2. Der Ausweich hatte die Lücke nicht geschlossen.
//
// **Geschlossen hat sie die Glättung** (Vermerk 0130,
// #1806).
// Gemessen am 2026-09-23 über dieselben 671 Karten: **99,0 %**, in der
// obersten Stufe **96,9 %** von MSE2 — und das Kanonmittel fiel dabei von
// 14,465 auf 13,792 (`docs/SCHRIFTUNTERSUCHUNG.md`). Die fehlende Tinte saß an
// den Kanten der Stämme, und dort legt die Nachbildung sie hin.
//
// # Wie gemessen wird
//
// Dieselbe Fläche wie damals: **Zeilen 340–460, Spalten 36–340** — der
// Regeltextkasten einer M15-Karte im Hochformat. Deckung eines Bildpunkts ist
// `1 − Helligkeit`, mit der Helligkeit nach Rec. 601 (0,299 / 0,587 / 0,114),
// summiert über die Fläche, dazu die Verteilung in zehn Stufen.
//
// **Die Wahl der Graurechnung verschiebt die Zahl um weniger als einen
// Punkt**: über `color.GrayModel` kommen 85,9 % statt 86,3 % heraus. Sie ist
// also keine Erklärung für die Lücke.
//
// **Die Fläche ist fest und nicht aus den Feldflächen genommen**, damit die
// Zahl mit der von damals vergleichbar bleibt. Wer sie ändert, misst etwas
// anderes (§10).
//
// # Und die Lücke lässt sich nicht wegrechnen
//
// **Am 2026-08-20 probiert und verworfen.** Eine Gammakorrektur auf die
// Glyphendeckung — `deckung^0,70` — bringt die Tinte auf **99,0 %** und damit
// zur Deckung mit MSE2. Der Abstand wird davon **schlechter**:
//
//	Abstand ganz        20,692  →  20,774
//	Ebenenart `Text`    81,325  →  82,128
//	schlechteste Karte  40,297  →  41,490
//
// Die 99,0 % sind über `color.GrayModel` erhoben, also mit derselben Rechnung
// wie die 85,9 % oben — die beiden gehören zusammen (§10).
//
// **Die fehlende Tinte als Menge ist also kein Hebel auf den Abstand.** Sie
// sitzt nicht nur in geringerer Menge, sondern an anderer Stelle — was zur
// Erklärung „Kantenglättung von macOS" passt und nicht zu „unsere Glyphen sind
// zu dünn". **Ihr Ort ist einer**: die Verdickung des Umrisses aus 0130
// bringt dieselben 99 % und senkt den Abstand, statt ihn zu heben.
//
// # Was er nicht sagt
//
//   - **Nicht, ob die Referenz die richtige ist.** Sie entsteht über
//     wxWidgets → Core Text auf **diesem** Mac, und der Urheber hat
//     angemerkt, dass MSE2 die Schrift unter Windows anders setzt. Die
//     Nachbildung aus 0130 gleicht an die Glättung von macOS an — mit Absicht
//     und entschieden, aber an eine Plattform.
//   - **Nicht, woher ein Unterschied kommt.** Er meldet Deckung, nicht Ursache.
package app

import (
	"image"
	"os"
	"path/filepath"
	"testing"
)

// deckungsflaeche ist die Fläche, über die #446 gemessen hat.
var deckungsflaeche = image.Rect(36, 340, 340, 460)

// deckungsstufen ist die Zahl der Stufen, in die die Deckung fällt.
const deckungsstufen = 10

// deckungsstand sammelt die Deckung über eine Menge Karten.
type deckungsstand struct {
	stufen [deckungsstufen]float64
	summe  float64
}

func (d *deckungsstand) nimm(bild image.Image, flaeche image.Rectangle) {
	versatz := bild.Bounds().Min

	for y := flaeche.Min.Y; y < flaeche.Max.Y; y++ {
		for x := flaeche.Min.X; x < flaeche.Max.X; x++ {
			// **Über `RGBA()` und nicht über eine Typzusicherung** —
			// `color.GrayModel.Convert` gibt einen `color.Color`, und die
			// Zusicherung darauf wäre eine Zusage ohne Prüfung.
			rot, gruen, blau, _ := bild.At(versatz.X+x, versatz.Y+y).RGBA()
			grau := (299*float64(rot) + 587*float64(gruen) + 114*float64(blau)) /
				1000 / 65535

			deckung := 1 - grau
			d.summe += deckung

			stufe := int(deckung * deckungsstufen)
			if stufe >= deckungsstufen {
				stufe = deckungsstufen - 1
			}

			d.stufen[stufe] += deckung
		}
	}
}

// **Die Deckung unserer Glyphen gegen die von MSE2.**
func TestWieVielTinteUnsereGlyphenSetzen(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	if _, err := os.Stat(mse2ausgabe("White")); err != nil {
		t.Skip("keine MSE2-Ausgabe — scripts/mse2-referenz.sh White")
	}

	anwendung := gestartetMit(t, func() string { return wurzel })

	geoeffnet := anwendung.SetOeffnen(t.Context(),
		filepath.Join(kanonverzeichnis(), "White.mse-set"))
	if geoeffnet.Fehler != nil {
		t.Fatalf("White öffnen: %+v", geoeffnet.Fehler)
	}

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if karten.Fehler != nil {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	var unser, ihrer deckungsstand

	verluste := ausfaelle{}
	gezaehlt := 0

	for _, karte := range verteilteStichprobe(karten.Karten,
		stichprobengroesse(t, 1)) {
		name := kartenname(karte)
		if name == "" {
			verluste.nimm("die Karte hat keinen Namen")

			continue
		}

		referenz, err := ladeBild(filepath.Join(referenzort(t, "White"),
			referenzdatei(name)))
		if err != nil {
			verluste.nimm("kein Referenzbild", name)

			continue
		}

		vorschau := anwendung.Vorschau(t.Context(), karte.ID, 1)
		if vorschau.Fehler != nil {
			verluste.nimm("die Vorschau scheiterte", name)

			continue
		}

		unser.nimm(bildAus(t, anwendung, vorschau.Vorschau.Bild), deckungsflaeche)
		ihrer.nimm(referenz, deckungsflaeche)

		gezaehlt++
	}

	if gezaehlt == 0 {
		t.Skip("keine vergleichbare Karte")
	}

	t.Logf("%d Karten aus White, Fläche %v", gezaehlt, deckungsflaeche)

	for stufe := range deckungsstufen {
		t.Logf("  %.1f–%.1f  wir %10.0f   MSE2 %10.0f",
			float64(stufe)/deckungsstufen, float64(stufe+1)/deckungsstufen,
			unser.stufen[stufe], ihrer.stufen[stufe])
	}

	t.Logf("  Gesamte Dunkeldeckung: wir %.0f, MSE2 %.0f — %.1f %%",
		unser.summe, ihrer.summe, 100*unser.summe/ihrer.summe)

	verluste.melde(t, gezaehlt)
}
