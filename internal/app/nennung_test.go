package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// mitBildebene ist ein Planbauer, dessen Plan ein **Bildfeld** zeichnet.
//
// **Der Behelfsplan tut das nicht**, und daran hängt die Illustrationszeile:
// behauptet wird sie nur, wenn ein Bildfeld von dort stammt. Ein Test ohne
// zeichnendes Bildfeld prüfte etwas anderes, als er behauptet.
type mitBildebene struct {
	ohneWortlisten
	ohneWahlbilder
}

func (mitBildebene) Kartenplan(context.Context, *domain.CardRow) (
	*render.Plan, []domain.Befund, error,
) {
	return &render.Plan{
		Breite: 20, Hoehe: 28,
		Ebenen: []render.Ebene{
			{
				Art: render.ArtBild, Feld: "kunst",
				Rahmen: render.Rechteck{Breite: 20, Hoehe: 14},
			},
		},
	}, nil, nil
}

func (mitBildebene) Kartenpalette(context.Context, *domain.CardRow, string) (
	[]service.Paletteneintrag, bool, []domain.Befund, error,
) {
	return nil, false, nil, nil
}

// bildeditor baut die Fassade mit einer Vorlage, die ein **Bildfeld** führt.
//
// **`type: image` und nicht „zeichnet ein Bild"** (#1032): `rarity`, `card
// symbol` und `watermark` zeichnen ebenfalls Bilder — die Illustrationszeile
// hängt an der Feldart, nicht an der Ebene.
func bildeditor(karte karteprobe) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitEditor(&schreibprobe{}, karte, mitBildebene{},
		render.NeuerRasterer(ohneBilder{}))
	anwendung.MitStilwahl(&stilwaehler{felder: []domain.Field{
		{Name: "kunst", Kind: domain.KindImage},

		// **Und eines, das ein Bild zeichnet, ohne ein Bildfeld zu sein** —
		// genau der Fall, an dem die Regel vorher danebenlag.
		{Name: "rarity", Kind: domain.KindChoice},
	}})

	return anwendung
}

// errHerkunftsprobe ist der Fehler, den die Herkunftsnaht im Test wirft.
var errHerkunftsprobe = errors.New("der Bestand ist nicht erreichbar")

// **Die Nennung steht neben der Karte, nicht neben dem Zuschnitt** —
// #1025.
//
// Scryfall verlangt für den Bildausschnitt, nachgelesen am 2026-08-23 unter
// `https://scryfall.com/docs/api`: „list the artist name and copyright
// elsewhere in the same interface presenting the art crop". Die Oberfläche,
// die den Ausschnitt zeigt, ist die **Kartenvorschau** — und sie zeigt ihn
// weiter, wenn der Zuschnittkasten längst zu ist.
//
// **Deshalb reist die Nennung mit der Vorschau** und nicht in einer eigenen
// Bindung: wer das Bild bekommt, bekommt damit, was neben ihm stehen muss, und
// kann nicht vergessen, es zu holen.
func TestDieVorschauTraegtDieNennung(t *testing.T) {
	anwendung := bildeditor(karteprobe{
		karte: &domain.CardRow{ID: 7, Values: []domain.Value{
			{Field: "illustrator", Text: "Christopher Rush"},
		}},
		herkuenfte: map[string]string{
			"kunst": "https://scryfall.com/card/lea/232/black-lotus",
		},
	})

	antwort := anwendung.Vorschau(context.Background(), "7", 1)
	if antwort.Fehler != nil {
		t.Fatalf("Vorschau: %v", antwort.Fehler)
	}

	if len(antwort.Vorschau.Nennungen) != 1 {
		t.Fatalf("%d Nennungen, erwartet eine", len(antwort.Vorschau.Nennungen))
	}

	nennung := antwort.Vorschau.Nennungen[0]

	if !strings.Contains(nennung.Text, "Christopher Rush") ||
		!strings.Contains(nennung.Text, "Wizards of the Coast") {
		t.Errorf("Text = %q, erwartet Künstler **und** Copyright — die Bedingung "+
			"nennt beides", nennung.Text)
	}

	if nennung.Adresse != "https://scryfall.com/card/lea/232/black-lotus" {
		t.Errorf("Adresse = %q — „Users should be able to identify the artist and "+
			"source of the image somehow.\"", nennung.Adresse)
	}
}

// **Eine Karte ohne fremde Herkunft trägt keine Nennung.**
//
// Über ein selbst gewähltes Bild zu behaupten, es sei ein Ausschnitt von
// Scryfall, wäre eine Aussage über fremde Rechte, die niemand geprüft hat.
func TestEineEigeneKarteTraegtKeineNennung(t *testing.T) {
	anwendung := editor(&schreibprobe{}, karteprobe{
		karte: &domain.CardRow{ID: 7, Values: []domain.Value{
			{Field: "illustrator", Text: "Selbst gemalt"},
		}},
	})

	antwort := anwendung.Vorschau(context.Background(), "7", 1)
	if antwort.Fehler != nil {
		t.Fatalf("Vorschau: %v", antwort.Fehler)
	}

	if len(antwort.Vorschau.Nennungen) != 0 {
		t.Errorf("%d Nennungen, erwartet keine: %v",
			len(antwort.Vorschau.Nennungen), antwort.Vorschau.Nennungen)
	}
}

// **Ein Fehler beim Lesen der Herkünfte hält die Vorschau nicht auf.**
//
// Eine Karte, die nicht mehr gezeichnet wird, weil eine Nebenangabe fehlt,
// wäre der schlechtere Tausch (§8) — die Vorschau ist das Hauptobjekt. Der
// Fehler wird trotzdem gemeldet, als Befund neben dem Bild: sonst stünde ein
// Ausschnitt ohne Nennung da, **ohne dass es jemand merkt**.
func TestEinFehlerBeiDenHerkuenftenMeldetSichAlsBefund(t *testing.T) {
	anwendung := editor(&schreibprobe{}, karteprobe{
		karte:           &domain.CardRow{ID: 7},
		herkunftsfehler: errHerkunftsprobe,
	})

	antwort := anwendung.Vorschau(context.Background(), "7", 1)
	if antwort.Fehler != nil {
		t.Fatalf("Vorschau: %v — sie soll trotzdem zeichnen", antwort.Fehler)
	}

	if antwort.Vorschau == nil {
		t.Fatal("keine Vorschau")
	}

	gefunden := false

	for _, befund := range antwort.Vorschau.Befunde {
		if strings.Contains(befund, "Herk") {
			gefunden = true
		}
	}

	if !gefunden {
		t.Errorf("Befunde = %v, erwartet einen über die Herkünfte", antwort.Vorschau.Befunde)
	}
}

// **Ein Feld, das ein Bild zeichnet, ist noch kein Bildfeld** —
// gefunden an der laufenden Anwendung, nachdem alles grün war (#1032).
//
// # Was geschehen war
//
// Die Karte trug **zwei** Nennungen aus zwei Scryfall-Karten: die eine für
// das Bild, die andere für eine Textübernahme. Beide behaupteten dieselbe
// Illustration — denn die Textübernahme hatte `rarity` gesetzt, und `rarity`
// zeichnet ein Bild (das Seltenheitssymbol).
//
// Die Regel las die **Ebenen des Zeichenplans** und beantwortete damit eine
// andere Frage als die gestellte. `type: image` steht im Spielpaket und meint
// genau das, was ein Mensch als Bild in die Karte legt.
func TestEineTextuebernahmeBehauptetKeineIllustration(t *testing.T) {
	anwendung := bildeditor(karteprobe{
		karte: &domain.CardRow{ID: 7, Values: []domain.Value{
			{Field: "illustrator", Text: "Christopher Rush"},
		}},
		herkuenfte: map[string]string{
			// **`rarity` zeichnet ein Bild und ist keines.**
			"rarity": "https://scryfall.com/card/msc/806",
		},
	})

	antwort := anwendung.Vorschau(context.Background(), "7", 1)
	if antwort.Fehler != nil {
		t.Fatalf("Vorschau: %v", antwort.Fehler)
	}

	if len(antwort.Vorschau.Nennungen) != 1 {
		t.Fatalf("%d Nennungen, erwartet eine", len(antwort.Vorschau.Nennungen))
	}

	if antwort.Vorschau.Nennungen[0].Text != "™ & © Wizards of the Coast" {
		t.Errorf("Text = %q — eine Textübernahme darf keine Illustration "+
			"behaupten", antwort.Vorschau.Nennungen[0].Text)
	}
}
