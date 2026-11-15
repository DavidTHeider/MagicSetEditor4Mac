package service

// **Kleine Bilder für die Kartenliste** — Werkstück 2 von Phase 3.5
// (#1166).
//
// # Was MSE2 tut, nachgelesen
//
// Die Kartenliste von MSE2 ist eine `FilteredImageCardList`
// (`src/gui/set/cards_panel.cpp`, Zeile 37), und die zeigt zu jeder Zeile ein
// Bild. Woher es kommt, steht in `src/gui/control/image_card_list.cpp`,
// Fassung `5deaeee` vom 2020-05-16:
//
//   - `ImageCardList::findImageField()` nimmt das **erste** `ImageField` der
//     `card_fields` des Spiels (Zeilen 42–48).
//   - `OnGetItemImage(pos)` liest dessen `ImageValue.filename`; ist er leer,
//     gibt es kein Bild (Zeilen 92–97).
//   - `CardThumbnailRequest::generate` lädt die **Datei aus dem Set-Paket**
//     und verkleinert zweistufig: `image.Rescale(36, 28)` ohne Kantenglättung,
//     dann `resample(image, 18, 14)` mit (Zeilen 61–70).
//
// **Es ist also die Illustration und nicht die gezeichnete Karte.** Das ist
// die Entscheidung, die `PARITAET.md` als `uncertain` geführt hat, und sie
// fällt hiermit: die Liste zeigt das Bild aus dem Feld, nicht das Ergebnis des
// Planbauers. Eine Karte durch Skriptauswertung und Rasterung zu schicken,
// nur um sie 32 Punkte hoch anzuzeigen, wäre um Größenordnungen teurer — und
// MSE2 tut es aus demselben Grund nicht.
//
// # Was hier anders ist, und warum
//
// **Die Maße.** 18×14 war ein Symbol in einer `wxListCtrl` von 2007; auf einem
// heutigen Bildschirm ist das nicht zu erkennen. Damit ist dieses Werkstück
// `adapted`: die Regel „ein verkleinertes Abbild der Illustration" ist
// übernommen, die Zahl nicht.
//
// **Und eine Stufe statt zweier.** MSEs zwei Schritte sind ein Ersatz für
// einen guten Filter — `Rescale` ohne Glättung, dann `resample` mit.
// [render.ZuschnittAus] hat den Filter, mit dem der Rasterer jedes Bild auf
// sein Feld bringt ([0532]); ihn zweimal zu fahren machte das Bild weicher,
// nicht besser.

import (
	"image"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Vorschaubreite und Vorschauhoehe sind die Schranken eines Listenbildes.
//
// **Als Schranken und nicht als Maße**: das Seitenverhältnis der Quelle
// bleibt, und ein Bild passt **in** dieses Rechteck. Eine feste Größe stauchte
// eine hochkant liegende Illustration.
const (
	Vorschaubreite = 56
	Vorschauhoehe  = 40
)

// Vorschaubildwerk macht kleine Bilder für die Kartenliste.
//
// **Es kennt keine Karte und kein Set.** Was es bekommt, ist ein Dateiname und
// zwei Orte; wer wissen will, welches Feld einer Karte das Bild trägt, fragt
// die Vorlage. Damit bleibt dieser Teil prüfbar, ohne eine Datenbank zu haben.
type Vorschaubildwerk struct {
	bilder render.Bildquelle
}

// NeuesVorschaubildwerk baut das Werk über einer Bildquelle.
func NeuesVorschaubildwerk(bilder render.Bildquelle) *Vorschaubildwerk {
	return &Vorschaubildwerk{bilder: bilder}
}

// Kartenbild gibt das verkleinerte Bild eines Kartenfeldes.
//
// **`nil` ohne Fehler heißt: es gibt keines**, und das ist der Normalfall —
// die meisten Karten eines frischen Sets tragen kein Bild. Ein Fehler wäre
// hier eine Meldung, die niemand braucht und die jede Liste rot färbte.
//
// `ablage` ist der Ort der Bilder, die jemand **in dieser Anwendung**
// eingesetzt hat; er geht dem Set-Paket vor
// ([0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md)) —
// derselbe Vorrang wie beim Zeichnen, und aus demselben Grund: läge das Paket
// vorn, zeigte die Liste nach dem Einsetzen weiter das alte Bild.
func (v *Vorschaubildwerk) Kartenbild(datei, setpaket, ablage string) (image.Image, error) {
	if strings.TrimSpace(datei) == "" {
		return nil, nil
	}

	verweis, gefunden := vorschauverweis(datei, setpaket, ablage)
	if !gefunden {
		return nil, nil
	}

	voll, err := v.bilder.Bild(verweis)
	if err != nil {
		return nil, err
	}

	return verkleinert(voll), nil
}

// vorschauverweis sagt, unter welchem Verweis das Bild zu holen ist.
//
// **Dieselbe Reihenfolge wie [kartenbild]** — Ablage vor Setpaket. Zwei
// Fassungen dieser Frage liefen an dem Tag auseinander, an dem eine dritte
// Quelle dazukommt; deshalb steht der Grund dort und hier nur die Reihenfolge.
func vorschauverweis(datei, setpaket, ablage string) (string, bool) {
	if ort := abgelegt(ablage, datei); ort != "" {
		return msepaket.SetpaketVerweis(ort, datei), true
	}

	if strings.TrimSpace(setpaket) == "" {
		return "", false
	}

	return msepaket.SetpaketVerweis(setpaket, datei), true
}

// verkleinert bringt ein Bild in die Schranken — mit seinem Seitenverhältnis.
func verkleinert(voll image.Image) image.Image {
	masse := voll.Bounds()
	if masse.Empty() {
		return nil
	}

	breite, hoehe := hineinpassend(masse.Dx(), masse.Dy())

	return render.ZuschnittAus(voll, masse, breite, hoehe)
}

// hineinpassend rechnet die Maße aus, die in die Schranken passen.
//
// **Ganzzahlig und mindestens eins.** Ein sehr breites Bild ergäbe sonst eine
// Höhe von null, und ein Bild ohne Fläche ist keines.
func hineinpassend(breite, hoehe int) (zielbreite, zielhoehe int) {
	if breite*Vorschauhoehe > hoehe*Vorschaubreite {
		return Vorschaubreite, max(1, hoehe*Vorschaubreite/breite)
	}

	return max(1, breite*Vorschauhoehe/hoehe), Vorschauhoehe
}
