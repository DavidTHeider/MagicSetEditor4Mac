package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Nennung überlebt das Schließen des Zuschnittfensters** —
// #1025.
//
// Scryfall verlangt für den Bildausschnitt, nachgelesen am 2026-08-23 unter
// `https://scryfall.com/docs/api`: „list the artist name and copyright
// elsewhere in the same interface presenting the art crop". Erfüllt war das
// bis hierher **nur, solange der Zuschnitt offen war**; danach stand das Bild
// auf der Karte und der Kasten war fort.
func TestDieNennungEntstehtAusHerkunftUndKuenstler(t *testing.T) {
	karte := &domain.CardRow{Values: []domain.Value{
		{Field: "name", Text: "Black Lotus"},
		{Field: "illustrator", Text: "Christopher Rush"},
	}}

	nennungen := service.Nennungen(karte, map[string]string{
		"kunst": "https://scryfall.com/card/lea/232/black-lotus?utm_source=api",
	}, []string{"kunst", "image"})

	if len(nennungen) != 1 {
		t.Fatalf("%d Nennungen, erwartet eine", len(nennungen))
	}

	if nennungen[0].Text != "Illustration: Christopher Rush. ™ & © Wizards of the Coast" {
		t.Errorf("Text = %q", nennungen[0].Text)
	}

	if nennungen[0].Adresse != "https://scryfall.com/card/lea/232/black-lotus?utm_source=api" {
		t.Errorf("Adresse = %q — die Quelle gehört dazu: „Users should be able to "+
			"identify the artist and source of the image somehow.\"", nennungen[0].Adresse)
	}
}

// **Was nicht von dort kam, bekommt keine fremde Nennung.**
//
// Eine falsche Nennung ist schlimmer als keine: über ein selbst gewähltes Foto
// zu behaupten, es sei ein Bildausschnitt von Scryfall, wäre eine Aussage über
// fremde Rechte, die niemand geprüft hat.
func TestEineEigeneHerkunftBekommtKeineNennung(t *testing.T) {
	karte := &domain.CardRow{Values: []domain.Value{
		{Field: "illustrator", Text: "Selbst gemalt"},
	}}

	for name, herkunft := range map[string]string{
		"eine Datei":          "/Users/probe/bilder/lotus.png",
		"ein fremder Wirt":    "https://scryfall.example.com/card/lea/232",
		"eine andere Seite":   "https://scryfall.com/sets/lea",
		"eine leere Herkunft": "",
	} {
		t.Run(name, func(t *testing.T) {
			if nennungen := service.Nennungen(karte,
				map[string]string{"kunst": herkunft},
				[]string{"kunst"}); len(nennungen) != 0 {
				t.Errorf("%d Nennungen für %q, erwartet keine", len(nennungen), herkunft)
			}
		})
	}
}

// **Mehrere Felder von derselben Karte ergeben eine Nennung, nicht drei.**
//
// Wer Namen, Regeltext und Bild von einer Scryfall-Karte übernimmt, soll nicht
// dreimal denselben Satz lesen. Genannt wird, **woher** es kam — und das ist
// einmal dieselbe Seite.
func TestMehrereFelderDerselbenSeiteErgebenEineNennung(t *testing.T) {
	karte := &domain.CardRow{Values: []domain.Value{
		{Field: "illustrator", Text: "Christopher Rush"},
	}}

	nennungen := service.Nennungen(karte, map[string]string{
		"kunst":     "https://scryfall.com/card/lea/232",
		"name":      "https://scryfall.com/card/lea/232",
		"rule text": "https://scryfall.com/card/lea/232",
	}, []string{"kunst", "image"})

	if len(nennungen) != 1 {
		t.Fatalf("%d Nennungen, erwartet eine", len(nennungen))
	}

	if len(nennungen[0].Felder) != 3 {
		t.Errorf("Felder = %v, erwartet alle drei", nennungen[0].Felder)
	}

	if nennungen[0].Felder[0] != "kunst" {
		t.Errorf("Felder = %v, erwartet sie in fester Reihenfolge — sonst tanzt die "+
			"Anzeige bei jedem Zeichnen", nennungen[0].Felder)
	}
}

// **Ohne Künstlernamen bleibt das Copyright.**
//
// Die Bedingung nennt beides; ein Feldwert, der von einer Kartenseite stammt,
// ohne dass die Karte einen Illustrator führt, hat trotzdem einen
// Rechteinhaber. Der Satz sagt dann, was bekannt ist, statt gar nichts.
func TestOhneKuenstlerBleibtDasCopyright(t *testing.T) {
	nennungen := service.Nennungen(&domain.CardRow{},
		map[string]string{"kunst": "https://scryfall.com/card/lea/232"},
		[]string{"kunst"})

	if len(nennungen) != 1 {
		t.Fatalf("%d Nennungen, erwartet eine", len(nennungen))
	}

	if nennungen[0].Text != "™ & © Wizards of the Coast" {
		t.Errorf("Text = %q, erwartet den Rechtevermerk ohne Illustrationszeile",
			nennungen[0].Text)
	}
}

// **Zwei Herkünfte kommen in fester Reihenfolge.**
//
// Eine Karte kann Werte von zwei Scryfall-Karten tragen — die Vorderseite von
// der einen, ein Regeltext von der anderen. Eine Map gibt sie in wechselnder
// Reihenfolge heraus, und die Anzeige tanzte dann bei jedem Zeichnen.
func TestZweiHerkuenfteStehenInFesterReihenfolge(t *testing.T) {
	karte := &domain.CardRow{Values: []domain.Value{
		{Field: "illustrator", Text: "Christopher Rush"},
	}}

	nennungen := service.Nennungen(karte, map[string]string{
		"kunst":     "https://scryfall.com/card/lea/232",
		"rule text": "https://scryfall.com/card/lea/161",
	}, []string{"kunst", "image"})

	if len(nennungen) != 2 {
		t.Fatalf("%d Nennungen, erwartet zwei", len(nennungen))
	}

	if nennungen[0].Adresse >= nennungen[1].Adresse {
		t.Errorf("Reihenfolge %q vor %q — erwartet aufsteigend nach Adresse",
			nennungen[0].Adresse, nennungen[1].Adresse)
	}
}

// **Ohne Bild von dort wird keine Illustration behauptet** —
// gefunden bei der Sichtprüfung am 2026-08-24, nachdem #1025 schon grün war.
//
// # Was geschehen war
//
// Das Bild kam von Scryfall, der Künstler wurde mitgeschrieben — **zwei**
// Zeilen mit derselben Herkunft. Danach wurde über das Bildfeld ein eigener
// Wert geschrieben; seine Herkunftszeile fiel, die des Künstlers blieb, und
// die Karte trug weiter „Illustration: Chris Rahn" neben einem Bild, das von
// woanders kam.
//
// **Kein Test konnte das zeigen**, weil jeder von ihnen eine einzige
// Herkunft setzte. Erst die laufende Anwendung hatte zwei.
//
// # Die Regel
//
// Die Illustrationszeile hängt daran, dass ein **Bildfeld** von dort stammt —
// nicht daran, dass die Karte einen Illustrator führt. Was von dort kam, wird
// weiter genannt; behauptet wird nur, was zutrifft.
func TestOhneBildfeldWirdKeineIllustrationBehauptet(t *testing.T) {
	karte := &domain.CardRow{Values: []domain.Value{
		{Field: "illustrator", Text: "Chris Rahn"},
	}}

	herkuenfte := map[string]string{"illustrator": "https://scryfall.com/card/vma/4"}

	nennungen := service.Nennungen(karte, herkuenfte, nil)
	if len(nennungen) != 1 {
		t.Fatalf("%d Nennungen, erwartet eine — was von dort kam, wird genannt",
			len(nennungen))
	}

	if nennungen[0].Text != "™ & © Wizards of the Coast" {
		t.Errorf("Text = %q — ohne Bild von dort darf keine Illustration "+
			"behauptet werden", nennungen[0].Text)
	}

	if nennungen[0].Kuenstler != "" {
		t.Errorf("Kuenstler = %q, erwartet keinen: die Angabe gehört zum Bild",
			nennungen[0].Kuenstler)
	}

	// **Die Gegenprobe im selben Lauf**: dieselbe Karte, dieselbe Herkunft —
	// und ein Bildfeld dazu. Ohne sie sagte der Test nur, dass hier nichts
	// steht, und nicht, dass es an der richtigen Bedingung hängt.
	herkuenfte["image"] = "https://scryfall.com/card/vma/4"

	mitBild := service.Nennungen(karte, herkuenfte, []string{"image"})
	if mitBild[0].Text != "Illustration: Chris Rahn. ™ & © Wizards of the Coast" {
		t.Errorf("mit Bildfeld: Text = %q, erwartet die Illustrationszeile",
			mitBild[0].Text)
	}
}
