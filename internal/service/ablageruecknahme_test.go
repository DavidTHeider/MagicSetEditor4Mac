package service_test

// **Ein gescheitertes Einfügen lässt nichts liegen** —
// #1278,
// Befund 2.
//
// **Eine eigene Datei**, weil `ablagebilder_test.go` mit ihr über 400 Zeilen
// ginge (§2). Der Schnitt folgt der Frage: dort steht, **was mitgeht**, hier,
// **was nicht liegen bleibt**, wenn es schiefgeht.

import (
	"errors"
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Scheitert das Anlegen, bleibt kein Bild liegen** —
// #1278,
// Befund 2.
//
// # Der halbe Stand, den es zu verhindern gilt
//
// `Einfuegen` schreibt **erst alle Bilder** und legt **danach** die Karten an —
// aus gutem Grund: ein Wert, der auf ein Bild zeigt, bekommt beim Anlegen den
// neuen Namen, und wäre die Datei noch nicht da, zeigte er einen Augenblick
// lang ins Leere (#1237).
//
// Der Preis war ein halber Stand: scheitert das Anlegen, lag die Datei schon
// im Bilderverzeichnis des Zielsets — ohne Karte, die auf sie zeigt. Sie
// wanderte später als Beiwerk in jede Sicherung, und niemand konnte sagen,
// wozu sie gehört. §8 und §10 fragen beide danach: **kein halber Stand**.
func TestEinGescheitertesEinfuegenLaesstKeinBildLiegen(t *testing.T) {
	t.Parallel()

	vor, _ := bildKarte(t)

	inhalt, err := vor.ablage.Kopieren(t.Context(), 1, []int64{1})
	if err != nil {
		t.Fatalf("Kopieren: %v", err)
	}

	vorher := vor.bilderImZielset(t)

	// **Das Anlegen scheitert, nicht das Ablegen.** Genau die Reihenfolge, in
	// der die Bilder schon liegen und die Karte noch nicht da ist.
	vor.doppel.anlegefehler = errors.New("kein Platz im Bestand")

	if _, fehler := vor.ablage.Einfuegen(t.Context(), 2, inhalt); fehler == nil {
		t.Fatal("Einfuegen hat nicht gemeldet, dass es gescheitert ist")
	}

	nachher := vor.bilderImZielset(t)

	if len(nachher) != len(vorher) {
		t.Errorf("im Zielset liegen %d Dateien statt %d — ein gescheitertes "+
			"Einfügen hat Bilder ohne Karte hinterlassen: %v",
			len(nachher), len(vorher), nachher)
	}
}

// **Und das gilt auch, wenn das Füllen der Werte scheitert.**
//
// Dort ist die Karte schon da — sie bleibt stehen, das ist die Zusage von
// `anlegen`. Die Bilder aber gehören zu einem Einfügen, das nicht durchkam.
func TestEinGescheitertesSchreibenLaesstKeinBildLiegen(t *testing.T) {
	t.Parallel()

	vor, _ := bildKarte(t)

	inhalt, err := vor.ablage.Kopieren(t.Context(), 1, []int64{1})
	if err != nil {
		t.Fatalf("Kopieren: %v", err)
	}

	vorher := vor.bilderImZielset(t)

	vor.doppel.schreibfehler = errors.New("der Bestand nimmt nichts an")

	if _, fehler := vor.ablage.Einfuegen(t.Context(), 2, inhalt); fehler == nil {
		t.Fatal("Einfuegen hat nicht gemeldet, dass es gescheitert ist")
	}

	if nachher := vor.bilderImZielset(t); len(nachher) != len(vorher) {
		t.Errorf("im Zielset liegen %d Dateien statt %d: %v",
			len(nachher), len(vorher), nachher)
	}
}

// bilderImZielset gibt die Dateinamen im Bilderverzeichnis des Zielsets.
//
// **Immer Set 2**, wie [bildvorrichtung.karteMitWert] immer Set 1 nimmt:
// kopiert wird aus 1 nach 2, und ein Parameter dafür wäre eine Wahl, die es
// nicht gibt.
func (v *bildvorrichtung) bilderImZielset(t *testing.T) []string {
	t.Helper()

	eintraege, err := os.ReadDir(v.bilder.Verzeichnis(2))
	if err != nil {
		// **Kein Verzeichnis heißt: keine Dateien.** Vor dem ersten Ablegen
		// gibt es keines, und das ist der Normalfall dieses Laufs.
		return nil
	}

	namen := make([]string, 0, len(eintraege))
	for _, eintrag := range eintraege {
		namen = append(namen, eintrag.Name())
	}

	return namen
}

// **Ein Teilerfolg behält die Bilder seiner geglückten Karten** —
// #1282,
// Befund 2.
//
// # Warum die erste Fassung zu viel wegnahm
//
// Karten und Bilder haben **verschiedene Rücknahmegrenzen**: eine Karte, die
// entstanden ist, bleibt stehen — das ist die Zusage von `anlegen`. Die
// Rücknahme entfernte trotzdem **alle** neu geschriebenen Bilder, auch die,
// auf die eine geglückte Karte zeigt.
//
// Die Läufe dazu ließen den Fehler stets schon bei der **ersten** Karte
// eintreten; der Fall „die erste ist da, die zweite nicht" kam in keinem vor.
// Deshalb hat das Doppel jetzt `abSchreibzug`.
func TestEinTeilerfolgBehaeltSeineBilder(t *testing.T) {
	t.Parallel()

	vor, _ := bildKarte(t)

	// **Eine zweite Karte mit einem zweiten Bild.** Sie ist die, die scheitert.
	zweites := zweitesBild(t, vor)
	vor.doppel.karten[1] = append(vor.doppel.karten[1], &domain.CardRow{
		ID: 2, SetID: 1, Values: []domain.Value{{Field: "image", Text: zweites}},
	})

	inhalt, err := vor.ablage.Kopieren(t.Context(), 1, []int64{1, 2})
	if err != nil {
		t.Fatalf("Kopieren: %v", err)
	}

	// **Erst der zweite Schreibzug scheitert.**
	vor.doppel.schreibfehler = errors.New("der Bestand nimmt nichts mehr an")
	vor.doppel.abSchreibzug = 2

	einfuegung, fehler := vor.ablage.Einfuegen(t.Context(), 2, inhalt)
	if fehler == nil {
		t.Fatal("Einfuegen hat nicht gemeldet, dass es gescheitert ist")
	}

	if einfuegung == nil || len(einfuegung.Karten) == 0 {
		t.Fatal("keine Karte angelegt — dann prüft dieser Lauf nichts")
	}

	// **Die erste Karte steht, und ihr Bild muss stehen.**
	erste := vor.wertVon(einfuegung.Karten[0], "image")
	if erste == "" {
		t.Fatal("die erste Karte trägt keinen Bildwert")
	}

	if !vor.bildLiegt(2, erste) {
		t.Errorf("die Datei %q der geglückten Karte ist weg — die Rücknahme "+
			"hat mehr entfernt, als zu diesem Fehlschlag gehört", erste)
	}

	// **Und das Bild der gescheiterten Karte liegt nicht herum.**
	if len(vor.bilderImZielset(t)) != 1 {
		t.Errorf("im Zielset liegen %v — erwartet genau die eine Datei der "+
			"geglückten Karte", vor.bilderImZielset(t))
	}
}

// zweitesBild legt ein zweites Bild im Quellset an und gibt seinen Namen.
func zweitesBild(t *testing.T, vor *bildvorrichtung) string {
	t.Helper()

	bild := image.NewRGBA(image.Rect(0, 0, 3, 3))
	bild.Set(1, 1, color.RGBA{G: 200, A: 255})

	name, err := vor.bilder.Ablegen(1, "image", bild)
	if err != nil {
		t.Fatalf("zweites Bild ablegen: %v", err)
	}

	return name
}
