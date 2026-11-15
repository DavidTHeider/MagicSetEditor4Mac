package service_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// optionsbildvorlage gibt den Stil der Saat samt einem Set, das `title` führt.
//
// **Das Setfeld gehört dazu**: alle 400 gerechneten Wahlbilder des Packs lesen
// `set`, und eine Vorlage ohne Setfelder gäbe „die Abbildung führt kein Glied"
// statt eines Bildes.
func optionsbildvorlage(t *testing.T) service.Vorlage {
	t.Helper()

	return service.Vorlage{
		Stil:      saatstil(t),
		Setfelder: []domain.Field{{Name: "title"}},
	}
}

// **Zu jeder Wahl ein Bild** —
// #1225.
//
// MSE zeigt in der Auswahlliste einer Stiloption je Wahl ein Vorschaubild
// (`DropDownChoiceListBase::generateThumbnailImages`,
// `src/gui/value/choice.cpp`, Zeilen 145–185, Fassung f9d9356 vom
// 2021-07-06). Dieses Fenster zeigte nur die Namen.
func TestJedeWahlEinerStiloptionBekommtIhrBild(t *testing.T) {
	bauer := service.NeuerPlanbauer().MitStilen(
		festerStil{vorlage: optionsbildvorlage(t), da: true})

	bilder, befunde, err := bauer.Optionsbilder(t.Context(), 7, "hintergrund")
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(bilder) != 2 {
		t.Fatalf("%d Bilder: %+v", len(bilder), bilder)
	}

	if bilder[0].Wahl != "hell" || bilder[1].Wahl != "dunkel" {
		t.Errorf("Reihenfolge der Wahlen: %+v", bilder)
	}

	// **Der Pfad ist der des Pakets**, nicht der rohe Dateiname: der Rasterer
	// bekommt einen Baum und keinen Ordner dazu.
	if bilder[0].Baum.Pfad == "" || bilder[0].Baum.Pfad == "hell.png" {
		t.Errorf("Pfad des ersten Bildes: %q", bilder[0].Baum.Pfad)
	}
}

// **Ein gerechnetes Wahlbild wird gerechnet.**
//
// Über `full-magic-pack/data` stehen **400 der 2.707** Wahlbilder als Skript,
// und alle vierhundert lesen `set` (gezählt am 2026-08-28). Wer den Ausdruck
// als Dateinamen durchreichte, suchte eine Datei namens `{ if … }`.
func TestEinGerechnetesWahlbildWirdAusgerechnet(t *testing.T) {
	bauer := service.NeuerPlanbauer().MitStilen(
		festerStil{vorlage: optionsbildvorlage(t), da: true})

	bilder, _, err := bauer.Optionsbilder(t.Context(), 7, "rahmen zeigen")
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	// **Eins und nicht zwei**: `no` erklärt kein Bild, und eine Wahl ohne
	// Eintrag bekommt keines — MSEs `or else nil` (`src/data/field/choice.cpp`,
	// Zeile 171, Fassung f9d9356 vom 2021-07-06).
	if len(bilder) != 1 {
		t.Fatalf("%d Bilder: %+v", len(bilder), bilder)
	}

	// `set.title` ist leer, also greift der `then`-Zweig.
	if bilder[0].Wahl != "yes" || bilder[0].Baum.Pfad == "" {
		t.Fatalf("erstes Bild: %+v", bilder[0])
	}

	if got := bilder[0].Baum.Pfad; got[len(got)-10:] != "dunkel.png" {
		t.Errorf("gerechneter Pfad: %q", got)
	}
}

// **Eine Option ohne Sichtseite gibt keine Bilder — und keinen Befund.**
//
// `magic-m15` erklärt dreizehn Optionen und gibt vieren ein Bild. Ein Befund
// je bildloser Option wären neun Meldungen, die niemand liest (§8).
func TestEineOptionOhneWahlbilderGibtKeine(t *testing.T) {
	bauer := service.NeuerPlanbauer().MitStilen(
		festerStil{vorlage: optionsbildvorlage(t), da: true})

	bilder, befunde, err := bauer.Optionsbilder(t.Context(), 7, "titelfarbe")
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	if len(bilder) != 0 || len(befunde) != 0 {
		t.Fatalf("%d Bilder, %d Befunde", len(bilder), len(befunde))
	}
}

// **Ohne Stil gibt es keine Wahlbilder, und das wird gesagt.**
func TestOhneStilGibtEsKeineWahlbilder(t *testing.T) {
	bilder, befunde, err := service.NeuerPlanbauer().
		Optionsbilder(t.Context(), 7, "hintergrund")
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	if len(bilder) != 0 {
		t.Fatalf("%d Bilder", len(bilder))
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde: %v", len(befunde), befunde)
	}
}

// **Ein gerechnetes `image:` gilt, wenn kein `choice images:` dasteht.**
//
//	if (style().choice_images.empty() && style().image.isScripted()) {
//	  … ctx.setVariable(_("input"), to_script(name)); …
//	}
//
// `src/gui/value/choice.cpp`, Zeilen 149–160, Fassung f9d9356 vom 2021-07-06.
// Das ist der Weg, den man weglässt — und dann bliebe die Auswahlliste genau
// dort ohne Bilder, wo der Stil sie **rechnet**.
func TestEinGerechnetesBildGiltJeWahl(t *testing.T) {
	bilder, _, err := optionsbauer(t, msepaket.Feldstil{
		Feld: "hintergrund",
		Bild: `{ input + ".png" }`,
	}).Optionsbilder(t.Context(), 7, "hintergrund")
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	if len(bilder) != 2 {
		t.Fatalf("%d Bilder: %+v", len(bilder), bilder)
	}

	// **`input` trägt den Wahlnamen** — ohne ihn rechnete jede Wahl dasselbe.
	if bilder[0].Baum.Pfad != "/probe-schlicht.mse-style/hell.png" ||
		bilder[1].Baum.Pfad != "/probe-schlicht.mse-style/dunkel.png" {
		t.Errorf("Pfade: %q, %q", bilder[0].Baum.Pfad, bilder[1].Baum.Pfad)
	}
}

// **Ein festes `image:` gilt nicht.**
//
// `style().image.isScripted()` — ein Dateiname stünde für **jede** Wahl gleich
// da, und eine Reihe aus dreimal demselben Bild sagt nichts.
func TestEinFestesBildGiltNichtJeWahl(t *testing.T) {
	bilder, _, err := optionsbauer(t, msepaket.Feldstil{
		Feld: "hintergrund",
		Bild: "hell.png",
	}).Optionsbilder(t.Context(), 7, "hintergrund")
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	if len(bilder) != 0 {
		t.Fatalf("%d Bilder: %+v", len(bilder), bilder)
	}
}

// **Eine Option, die der Stil nicht erklärt, hat keine Wahlen.**
//
// Der `styling style:`-Block darf einen Namen führen, den `styling field:`
// nicht kennt — dann gibt es nichts zu zeigen, und das ist kein Befund.
func TestEineUnbekannteOptionHatKeineWahlen(t *testing.T) {
	bilder, befunde, err := optionsbauer(t, msepaket.Feldstil{
		Feld: "gibt es nicht",
		Bild: `{ input + ".png" }`,
	}).Optionsbilder(t.Context(), 7, "gibt es nicht")
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	if len(bilder) != 0 || len(befunde) != 0 {
		t.Fatalf("%d Bilder, %d Befunde", len(bilder), len(befunde))
	}
}

// **Ohne Stil für dieses Set gibt es nichts** — und die Befunde des Suchens
// reisen mit.
func TestOhneStilFuerDasSetGibtEsKeineWahlbilder(t *testing.T) {
	bauer := service.NeuerPlanbauer().MitStilen(festerStil{
		befunde: []domain.Befund{{Meldung: "das Stilpaket fehlt"}},
	})

	bilder, befunde, err := bauer.Optionsbilder(t.Context(), 7, "hintergrund")
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	if len(bilder) != 0 || len(befunde) != 1 {
		t.Fatalf("%d Bilder, %d Befunde: %v", len(bilder), len(befunde), befunde)
	}
}

// **Ein abgebrochener Lauf zeichnet nicht mehr** (§7).
func TestEinAbgebrochenerLaufBautKeineWahlbilder(t *testing.T) {
	abbruch, brich := context.WithCancel(t.Context())
	brich()

	if _, _, err := optionsbauer(t, msepaket.Feldstil{Feld: "hintergrund"}).
		Optionsbilder(abbruch, 7, "hintergrund"); err == nil {
		t.Fatal("kein Fehler")
	}
}

// optionsbauer gibt einen Planbauer, dessen Stil den genannten Optionsstil
// führt — **und keinen anderen**.
//
// **Von Hand gebaut und nicht aus der Saat**: die Saat zeigt den Weg über eine
// echte Datei ([TestJedeWahlEinerStiloptionBekommtIhrBild]); hier geht es um
// die Verzweigung dahinter, und dafür ist ein Feldstil im Test die kürzere
// Wahrheit.
func optionsbauer(t *testing.T, sicht msepaket.Feldstil) *service.Planbauer {
	t.Helper()

	stil := saatstil(t)
	stil.Optionsstile = []msepaket.Feldstil{sicht}

	return service.NeuerPlanbauer().MitStilen(festerStil{
		vorlage: service.Vorlage{
			Stil:      stil,
			Setfelder: []domain.Field{{Name: "title"}},
		},
		da: true,
	})
}
