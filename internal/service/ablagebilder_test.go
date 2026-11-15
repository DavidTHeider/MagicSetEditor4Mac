// **Eine kopierte Karte nimmt ihre Bilder mit** —
// #1237.
//
// # Der Befund
//
// Die Zwischenablage überträgt **Feldwerte**. Ein Bildfeld trägt aber keinen
// Bildinhalt, sondern einen **paketlokalen Namen**: `image: image1`. Und
// dieser Name ist im Bestand je Set aufgelöst — `<wurzel>/<setID>/image1`.
//
// Wer eine Karte mit `image1` aus Set A nach Set B kopierte, bekam dort einen
// Verweis auf `<wurzel>/B/image1`: eine Datei, die **nicht existiert** oder,
// schlimmer, ein **fremdes Bild**, das zufällig gleich heißt.
//
// # Was die Quellen sagen (§11)
//
//	WITH_DYNAMIC_ARG(clipboard_package, &package);
//
// `serialize_for_clipboard`, `src/data/format/clipboard.cpp`, Zeilen 23–30 —
// MSE bindet das **Quellpaket** ausdrücklich mit, und beim Lesen werden
// paketlokale Namen ins Zielpaket **kopiert** (`src/util/io/package.cpp`,
// Zeile 311). Fassung f9d9356 vom 2021-07-06.
//
// # Warum hier Base64 und kein Paket
//
// MSEs Ablage ist ein Paketkontext; unsere ist **Text** — sie geht durch die
// Zwischenablage des Systems und soll dort auch von MSE2 lesbar bleiben. Ein
// zusätzlicher Block je Bild ist ein Schlüssel, den MSE2 nicht kennt und
// deshalb überspringt (`Reader::unknownKey`). Der Preis ist die Größe: ein
// Kunstbild wiegt Hunderte Kilobyte, und Base64 macht daraus ein Drittel mehr.
package service_test

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bildvorrichtung ist eine Ablage mit einem Bilderverzeichnis daneben.
type bildvorrichtung struct {
	doppel *ablagedoppel
	ablage *service.Kartenablage
	bilder *service.Bildablage
}

// neueBildvorrichtung baut sie über zwei Sets desselben Spiels.
func neueBildvorrichtung(t *testing.T) *bildvorrichtung {
	t.Helper()

	doppel := &ablagedoppel{
		koepfe: []domain.SetHead{
			{ID: 1, Name: "Erstes", Game: "probe"},
			{ID: 2, Name: "Zweites", Game: "probe"},
		},
		karten: map[int64][]*domain.CardRow{},
	}

	bilder := service.NeueBildablage(t.TempDir())

	return &bildvorrichtung{
		doppel: doppel,
		ablage: service.NeueKartenablage(doppel, doppel, doppel).MitBildern(bilder),
		bilder: bilder,
	}
}

// karteMitWert legt eine Karte mit einem Feldwert im **Quellset** an.
//
// **Immer Set 1**: das ist der Gegenstand dieser Läufe — kopiert wird aus 1
// nach 2, und ein Parameter dafür wäre eine Wahl, die es nicht gibt.
func (v *bildvorrichtung) karteMitWert(feld, wert string) {
	v.doppel.karten[1] = append(v.doppel.karten[1], &domain.CardRow{
		ID: 1, SetID: 1, Values: []domain.Value{{Field: feld, Text: wert}},
	})
}

// wertVon gibt den geschriebenen Wert einer eingefügten Karte.
func (v *bildvorrichtung) wertVon(karteID int64, feld string) string {
	for _, wert := range v.doppel.werte[karteID] {
		if wert.Field == feld {
			return wert.Text
		}
	}

	return ""
}

// bildLiegt sagt, ob im Verzeichnis eines Sets eine Datei liegt.
func (v *bildvorrichtung) bildLiegt(setID int64, name string) bool {
	_, err := os.Stat(filepath.Join(v.bilder.Verzeichnis(setID), name))

	return err == nil
}

// bildKarte legt eine Karte mit einem abgelegten Bild an und gibt beides.
func bildKarte(t *testing.T) (vorrichtung *bildvorrichtung, bildname string) {
	t.Helper()

	vor := neueBildvorrichtung(t)

	bild := image.NewRGBA(image.Rect(0, 0, 2, 2))
	bild.Set(0, 0, color.RGBA{R: 200, G: 30, B: 30, A: 255})

	name, err := vor.bilder.Ablegen(1, "image", bild)
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	vor.karteMitWert("image", name)

	return vor, name
}

func TestEineKopierteKarteNimmtIhrBildMit(t *testing.T) {
	t.Parallel()

	vor, name := bildKarte(t)

	inhalt, err := vor.ablage.Kopieren(t.Context(), 1, []int64{1})
	if err != nil {
		t.Fatalf("Kopieren: %v", err)
	}

	if !strings.Contains(inhalt, "image_data") {
		t.Fatalf("der Inhalt trägt keine Bilddaten:\n%s", inhalt)
	}

	einfuegung, err := vor.ablage.Einfuegen(t.Context(), 2, inhalt)
	if err != nil {
		t.Fatalf("Einfuegen: %v", err)
	}

	// **Im Zielset liegt eine eigene Datei** — nicht der Verweis auf die des
	// Quellsets.
	neu := vor.wertVon(einfuegung.Karten[0], "image")
	if neu == "" {
		t.Fatal("die eingefügte Karte hat kein Bild")
	}

	if !vor.bildLiegt(2, neu) {
		t.Errorf("im Zielset liegt keine Datei %q — der Verweis zeigt ins Leere", neu)
	}

	// **Und die Quelle bleibt unberührt.**
	if !vor.bildLiegt(1, name) {
		t.Errorf("das Bild des Quellsets ist weg")
	}
}

// **Ein Bild, das es nicht mehr gibt, hält das Kopieren nicht auf** (§8). Der
// Verweis geht mit, das Bild fehlt, und das steht als Befund daneben.
func TestEinFehlendesBildHaeltDasKopierenNichtAuf(t *testing.T) {
	t.Parallel()

	vor := neueBildvorrichtung(t)
	vor.karteMitWert("image", "image9")

	inhalt, err := vor.ablage.Kopieren(t.Context(), 1, []int64{1})
	if err != nil {
		t.Fatalf("Kopieren: %v", err)
	}

	if !strings.Contains(inhalt, "image9") {
		t.Errorf("der Verweis fehlt im Inhalt:\n%s", inhalt)
	}
}

// **Ein Verweis ohne Bilddaten bleibt stehen, wie er ist.** Das ist der Inhalt
// aus MSE2 oder aus einer älteren Fassung dieser Anwendung — er soll weiter
// einfügbar sein.
func TestEinInhaltOhneBilddatenBleibtEinfuegbar(t *testing.T) {
	t.Parallel()

	vor := neueBildvorrichtung(t)

	einfuegung, err := vor.ablage.Einfuegen(t.Context(), 2,
		"game_name: probe\ncard:\n\timage: image1\n")
	if err != nil {
		t.Fatalf("Einfuegen: %v", err)
	}

	if got := vor.wertVon(einfuegung.Karten[0], "image"); got != "image1" {
		t.Errorf("der Wert ist %q, erwartet \"image1\"", got)
	}
}

// **Ohne Bildablage geht nur der Verweis mit.** Die Flächen haben kein
// Bilderverzeichnis, und dort ist das richtig — besser als gar kein Kopieren.
func TestOhneBildablageGehtNurDerVerweisMit(t *testing.T) {
	t.Parallel()

	vor := neueBildvorrichtung(t)
	vor.karteMitWert("image", "image1")

	ohne := service.NeueKartenablage(vor.doppel, vor.doppel, vor.doppel)

	inhalt, err := ohne.Kopieren(t.Context(), 1, []int64{1})
	if err != nil {
		t.Fatalf("Kopieren: %v", err)
	}

	if strings.Contains(inhalt, "image_data") {
		t.Errorf("es wurden Bilddaten angehängt, ohne Bildablage:\n%s", inhalt)
	}
}

// **Ein Wert, der nicht wie ein Bildname aussieht, wird nicht gesucht.**
func TestEinTextwertWirdNichtFuerEinBildGehalten(t *testing.T) {
	t.Parallel()

	vor := neueBildvorrichtung(t)
	vor.karteMitWert("name", "Lightning Bolt")

	inhalt, err := vor.ablage.Kopieren(t.Context(), 1, []int64{1})
	if err != nil {
		t.Fatalf("Kopieren: %v", err)
	}

	if strings.Contains(inhalt, "image_data") {
		t.Errorf("»Lightning Bolt« wurde für einen Bildnamen gehalten:\n%s", inhalt)
	}
}

// **Und ein leerer Wert ebenso wenig.**
func TestEinLeererWertWirdUebergangen(t *testing.T) {
	t.Parallel()

	vor := neueBildvorrichtung(t)
	vor.karteMitWert("image", "")

	if _, err := vor.ablage.Kopieren(t.Context(), 1, []int64{1}); err != nil {
		t.Fatalf("Kopieren: %v", err)
	}
}

// **Unlesbare Bilddaten sind ein Befund, kein Abbruch** (§8).
func TestUnlesbareBilddatenSindEinBefund(t *testing.T) {
	t.Parallel()

	vor := neueBildvorrichtung(t)

	einfuegung, err := vor.ablage.Einfuegen(t.Context(), 2,
		"game_name: probe\nimage_data: image1 !!!keinbase64!!!\ncard:\n\timage: image1\n")
	if err != nil {
		t.Fatalf("Einfuegen: %v", err)
	}

	if len(einfuegung.Befunde) == 0 {
		t.Error("die unlesbaren Daten wurden verschwiegen")
	}
}

// **Ein Block ohne Trennzeichen wird übergangen.**
func TestEinBlockOhneNamenWirdUebergangen(t *testing.T) {
	t.Parallel()

	vor := neueBildvorrichtung(t)

	if _, err := vor.ablage.Einfuegen(t.Context(), 2,
		"game_name: probe\nimage_data: nurEinWort\ncard:\n\tname: X\n"); err != nil {
		t.Fatalf("Einfuegen: %v", err)
	}
}

// **Ein Bild, das sich nicht ablegen lässt, ist ein Befund.**
//
// Erzwungen über ein Bilderverzeichnis, das eine **Datei** ist: darunter lässt
// sich kein Verzeichnis anlegen.
func TestEinUnablegbaresBildIstEinBefund(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	if err := os.WriteFile(filepath.Join(wurzel, "2"), []byte("keine Wurzel"), 0o600); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	vor := neueBildvorrichtung(t)
	vor.ablage = service.NeueKartenablage(vor.doppel, vor.doppel, vor.doppel).
		MitBildern(service.NeueBildablage(wurzel))

	einfuegung, err := vor.ablage.Einfuegen(t.Context(), 2,
		"game_name: probe\nimage_data: image1 aGFsbG8=\ncard:\n\timage: image1\n")
	if err != nil {
		t.Fatalf("Einfuegen: %v", err)
	}

	if len(einfuegung.Befunde) == 0 {
		t.Error("der Ablegefehler wurde verschwiegen")
	}
}

// **Ohne Wurzel gibt es kein Ablegen** — und eine klare Meldung.
func TestOhneWurzelLassenSichKeineBytesAblegen(t *testing.T) {
	t.Parallel()

	if _, err := service.NeueBildablage("").Bytes(1, "image1", []byte("x")); err == nil {
		t.Error("ohne Wurzel wurde abgelegt")
	}
}

// **Ein Name ganz aus Ziffern behält seinen Stamm.**
func TestEinNameAusZiffernBehaeltSeinenStamm(t *testing.T) {
	t.Parallel()

	name, err := service.NeueBildablage(t.TempDir()).Bytes(1, "123", []byte("x"))
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	if !strings.HasPrefix(name, "123") {
		t.Errorf("der Name ist %q, erwartet einen Stamm »123«", name)
	}
}

// **Eine Bildablage ohne Wurzel hängt nichts an.** Sie steht so da, wenn
// niemand ein Verzeichnis eingerichtet hat.
func TestEineBildablageOhneWurzelHaengtNichtsAn(t *testing.T) {
	t.Parallel()

	vor := neueBildvorrichtung(t)
	vor.karteMitWert("image", "image1")
	vor.ablage = service.NeueKartenablage(vor.doppel, vor.doppel, vor.doppel).
		MitBildern(service.NeueBildablage(""))

	inhalt, err := vor.ablage.Kopieren(t.Context(), 1, []int64{1})
	if err != nil {
		t.Fatalf("Kopieren: %v", err)
	}

	if strings.Contains(inhalt, "image_data") {
		t.Errorf("ohne Wurzel wurden Bilddaten angehängt:\n%s", inhalt)
	}
}

// **Dieselbe Datei geht einmal mit, nicht je Karte.** Wer zehn Karten mit
// demselben Rahmenbild kopiert, legt es einmal in die Zwischenablage.
func TestDieselbeDateiGehtNurEinmalMit(t *testing.T) {
	t.Parallel()

	vor, name := bildKarte(t)
	vor.doppel.karten[1] = append(vor.doppel.karten[1], &domain.CardRow{
		ID: 2, SetID: 1, Values: []domain.Value{{Field: "image", Text: name}},
	})

	inhalt, err := vor.ablage.Kopieren(t.Context(), 1, []int64{1, 2})
	if err != nil {
		t.Fatalf("Kopieren: %v", err)
	}

	if zahl := strings.Count(inhalt, "image_data"); zahl != 1 {
		t.Errorf("%d Bildblöcke, erwartet 1", zahl)
	}
}
