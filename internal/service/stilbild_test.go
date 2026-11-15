package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bildstil baut einen Stil von Hand — ohne Paket auf der Platte.
//
// **Von Hand und nicht aus `saat/`**, weil hier die Deutung geprüft wird und
// nicht das Lesen: ein Prüfstück je Fall im Dateisystem machte aus neun Fällen
// neun Pakete, und keiner davon wäre neben seinem Test zu lesen.
func bildstil(felder ...msepaket.Feldstil) msepaket.Stil {
	return msepaket.Stil{
		Paket: "probe.mse-style", Breite: 375, Hoehe: 523, Felder: felder,
	}
}

// gedeutet baut den Plan und verlangt, dass es keinen Fehler gab.
func gedeutet(t *testing.T, stil msepaket.Stil) (*render.Plan, []domain.Befund) {
	t.Helper()

	return gedeutetMit(t, stilkarte("rot"), stil)
}

// gedeutetMit baut den Plan aus einer eigens gefüllten Karte.
func gedeutetMit(
	t *testing.T, karte *domain.CardRow, stil msepaket.Stil,
) (*render.Plan, []domain.Befund) {
	t.Helper()

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	return plan, befunde
}

// ebenen gibt die Ebenen ohne den weißen Grund.
func ebenen(t *testing.T, plan *render.Plan) []render.Ebene {
	t.Helper()

	if len(plan.Ebenen) == 0 {
		t.Fatal("der Plan trägt nicht einmal den Grund")
	}

	return plan.Ebenen[1:]
}

// flaeche ist die volle Kartenfläche als Feldstil-Angabe.
func flaeche(feld string) msepaket.Feldstil {
	return msepaket.Feldstil{
		Feld: feld, Links: "0", Oben: "0", Breite: "375", Hoehe: "523",
	}
}

// **Ein Feld mit `image:` kommt als Bildebene in den Plan.**
//
// # Was der einfache Dateiname bedeutet
//
// `image: frame.png` meint eine Datei im **eigenen** Paket — so steht es
// 26-mal in den Stilen von `M15-Magic-Pack`. Der Plan trägt einen Paketpfad
// und keinen Dateisystempfad (`render.Ebene.Bild`), also wird der Name um sein
// Paket ergänzt. Wer ihn roh weiterreichte, suchte ihn in der Paketwurzel.
func TestEinBildfeldKommtAlsBildebeneInDenPlan(t *testing.T) {
	feld := flaeche("card color")
	feld.Darstellung = "image"
	feld.Bild = "rahmen.png"

	plan, befunde := gedeutet(t, bildstil(feld))

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	gezeichnet := ebenen(t, plan)
	if len(gezeichnet) != 1 {
		t.Fatalf("%d Ebenen, erwartet eine", len(gezeichnet))
	}

	if gezeichnet[0].Art != render.ArtBild {
		t.Errorf("Art %q, erwartet %q", gezeichnet[0].Art, render.ArtBild)
	}

	if gezeichnet[0].Bild.Pfad != "/probe.mse-style/rahmen.png" {
		t.Errorf("Bild %q", gezeichnet[0].Bild.Pfad)
	}

	if gezeichnet[0].Rahmen.Breite != 375 || gezeichnet[0].Rahmen.Hoehe != 523 {
		t.Errorf("Rahmen %+v", gezeichnet[0].Rahmen)
	}
}

// **Ein Bildskript wird ausgewertet, und die Karte steht ihm zur Verfügung.**
//
// 356 der 358 `image:`-Angaben des Packs sind Skripte. Ein Leser, der nur den
// Dateinamen kennte, deutete zwei davon.
func TestEinBildskriptWirdAusgewertet(t *testing.T) {
	feld := flaeche("card color")
	feld.Darstellung = "image"
	feld.Bild = `{ "rahmen_" + card.farbe + ".png" }`

	plan, befunde := gedeutet(t, bildstil(feld))

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if bild := ebenen(t, plan)[0].Bild; bild.Pfad != "/probe.mse-style/rahmen_rot.png" {
		t.Errorf("Bild %q", bild.Pfad)
	}
}

// **Ein Pfad, der mit `/` beginnt, meint ein anderes Paket und bleibt.**
//
// `mask: /magic-genevensis-00-main.mse-style/pride_mask.png` steht so im Pack.
// Wer ihm das eigene Paket voranstellte, baute einen Pfad, den es nirgends
// gibt.
func TestEinAbsoluterPfadMeintEinAnderesPaket(t *testing.T) {
	feld := flaeche("card color")
	feld.Darstellung = "image"
	feld.Bild = "/anderes.mse-include/rahmen.png"

	plan, _ := gedeutet(t, bildstil(feld))

	if bild := ebenen(t, plan)[0].Bild; bild.Pfad != "/anderes.mse-include/rahmen.png" {
		t.Errorf("Bild %q", bild.Pfad)
	}
}

// **Ein Feld mit `mask:` bekommt seine Maske** — als Bild wie als Text.
//
// Die Maske gehört an **beide**: `border color` trägt sie ohne Bild, `card
// color` mit einem. Ein Bauer, der sie nur an Bildebenen hängte, ließe die
// Hälfte der 352 Vorkommen fallen.
func TestEineMaskeKommtAnDieEbene(t *testing.T) {
	bild := flaeche("card color")
	bild.Darstellung = "image"
	bild.Bild = "rahmen.png"
	bild.Maske = "rand_maske.png"

	text := flaeche("name")
	text.Schrift = msepaket.Schrift{Name: "Helvetica", Groesse: "20"}
	text.Maske = "{ \"text_maske.png\" }"

	plan, befunde := gedeutet(t, bildstil(bild, text))

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	gezeichnet := ebenen(t, plan)
	if len(gezeichnet) != 2 {
		t.Fatalf("%d Ebenen, erwartet zwei", len(gezeichnet))
	}

	if gezeichnet[0].Maske.Pfad != "/probe.mse-style/rand_maske.png" {
		t.Errorf("Maske der Bildebene: %q", gezeichnet[0].Maske.Pfad)
	}

	// **Auch aus einem Skript.** 174 der 352 Masken sind welche.
	if gezeichnet[1].Maske.Pfad != "/probe.mse-style/text_maske.png" {
		t.Errorf("Maske der Textebene: %q", gezeichnet[1].Maske.Pfad)
	}
}

// **Ein Bildskript, das nichts ergibt, zeichnet nichts — und meldet nichts.**
//
// `image: {if styling.overlay == "" then nil else styling.overlay + "/overlay.png"}`
// steht 99-mal im Pack. Der leere Fall ist dort die **Regel**, nicht der
// Fehler: keine Überlagerung gewählt heißt keine Ebene. Ein Befund dafür
// machte aus jeder zweiten Karte eine Karte mit Vorbehalten.
func TestEinLeeresBildskriptZeichnetNichts(t *testing.T) {
	feld := flaeche("overlay")
	feld.Darstellung = "image"
	feld.Bild = `{ if card.farbe == "blau" then "rahmen.png" else "" }`

	plan, befunde := gedeutet(t, bildstil(feld))

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if gezeichnet := ebenen(t, plan); len(gezeichnet) != 0 {
		t.Errorf("%d Ebenen, erwartet keine: %+v", len(gezeichnet), gezeichnet)
	}
}

// **Ergibt `image:` nichts, gilt `default:`.**
//
// Die beiden sind nicht dasselbe: `image:` rechnet, `default:` fängt auf. Im
// Pack führt das Kunstfeld `default: {default_image(card.card_color)}` — ein
// Ersatzbild für die Karte, die noch keines hat.
func TestOhneBildGiltDasStandardbild(t *testing.T) {
	feld := flaeche("image")
	feld.Darstellung = "image"
	feld.Bild = `{ "" }`
	feld.Standardbild = "ersatz.png"

	plan, befunde := gedeutet(t, bildstil(feld))

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if bild := ebenen(t, plan)[0].Bild; bild.Pfad != "/probe.mse-style/ersatz.png" {
		t.Errorf("Bild %q", bild.Pfad)
	}
}

// **`render style: hidden` heißt: gar nicht zeichnen.**
//
// Viermal im Pack, und jedes Mal an einem `card color` einer Flip-Karte. Ein
// Befund wäre hier falsch: das Feld ist nicht unverstanden, es ist verborgen.
func TestEinVerborgenesFeldWirdNichtGezeichnet(t *testing.T) {
	feld := flaeche("card color")
	feld.Darstellung = "hidden"
	feld.Bild = "rahmen.png"

	plan, befunde := gedeutet(t, bildstil(feld))

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if gezeichnet := ebenen(t, plan); len(gezeichnet) != 0 {
		t.Errorf("%d Ebenen, erwartet keine", len(gezeichnet))
	}
}

// **Ein Bildskript, das sich nicht ausrechnen lässt, ist ein Befund** (§8) —
// und die Ebene bleibt weg, statt auf einen erfundenen Pfad zu zeigen.
func TestEinUnausrechenbaresBildskriptIstEinBefund(t *testing.T) {
	feld := flaeche("card color")
	feld.Darstellung = "image"
	feld.Bild = "{ nichtsdergleichen() }"

	plan, befunde := gedeutet(t, bildstil(feld))

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet einer: %v", len(befunde), befunde)
	}

	if befunde[0].Ort != "card color → image" {
		t.Errorf("Ort %q", befunde[0].Ort)
	}

	if gezeichnet := ebenen(t, plan); len(gezeichnet) != 0 {
		t.Errorf("%d Ebenen, erwartet keine", len(gezeichnet))
	}
}

// **Was noch nicht gedeutet ist, wird beim Namen genannt** — und jedes mit
// seinem eigenen Grund.
//
// Ein Sammelsatz („Bilder, Masken und Symbolschriften sind noch nicht
// gedeutet") sagt bei drei verschiedenen Lagen dasselbe und bei keiner, was
// fehlt. Zwei davon sind seit diesem Werkstück gedeutet; die dritte nicht, und
// **das** ist die Auskunft, die jemand braucht.
//
// **Der Fall `render style: image` ohne Bildquelle ist seit
// #506
// keiner mehr**: MSE zeichnet dort nichts und sagt nichts
// (`draw_choice_viewer`, `src/render/value/choice.cpp`, Zeilen 74–92,
// Fassung f9d9356 vom 2021-07-06). Die
// Vorlage meint einen Platzhalter, keine Lücke.
func TestWasNichtGedeutetIstWirdBenannt(t *testing.T) {
	symbole := flaeche("casting cost")
	symbole.Symbolschrift = msepaket.Symbolangabe{Name: "magic-mana-large", Groesse: "15"}

	stumm := flaeche("border color")

	// **Mit Werten, denn ein Feld ohne Wert wird ohnehin nicht gezeichnet.**
	// Ein Befund über ein Feld, das leer bliebe, meldete eine Lücke, die
	// niemand sieht.
	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "casting cost", Text: "2R"},
		{Field: "border color", Text: "black"},
	}}

	_, befunde := gedeutetMit(t, karte, bildstil(symbole, stumm))

	if len(befunde) != 2 {
		t.Fatalf("%d Befunde, erwartet zwei: %v", len(befunde), befunde)
	}

	faelle := []struct {
		ort  string
		wort string
	}{
		{"casting cost", "magic-mana-large"},
		{"border color", "Schrift"},
	}

	for stelle, fall := range faelle {
		if befunde[stelle].Ort != fall.ort {
			t.Errorf("Befund %d: Ort %q, erwartet %q", stelle, befunde[stelle].Ort, fall.ort)
		}

		if !strings.Contains(befunde[stelle].Meldung, fall.wort) {
			t.Errorf("Befund %d: %q nennt %q nicht", stelle, befunde[stelle].Meldung, fall.wort)
		}
	}
}

// **Eine offene geschweifte Klammer ist kein Dateiname.**
//
// Sie ist der eine Fall, in dem beide naheliegenden Deutungen falsch sind:
// auswerten geht nicht, und durchreichen ergäbe den Pfad
// `/probe.mse-style/{ if card.farbe` — die Meldung käme dann aus der
// Bildquelle und hieße „nicht gefunden", wo „kaputt" richtig ist.
func TestEineOffeneKlammerIstEinBefundUndKeinDateiname(t *testing.T) {
	feld := flaeche("card color")
	feld.Darstellung = "image"
	feld.Bild = `{ if card.farbe`

	plan, befunde := gedeutet(t, bildstil(feld))

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet einer: %v", len(befunde), befunde)
	}

	if !strings.Contains(befunde[0].Meldung, "endet nicht darauf") {
		t.Errorf("Meldung %q", befunde[0].Meldung)
	}

	if gezeichnet := ebenen(t, plan); len(gezeichnet) != 0 {
		t.Errorf("%d Ebenen, erwartet keine: %+v", len(gezeichnet), gezeichnet)
	}
}

// **Ohne den Ort des Sets wird das eigene Bild gemeldet, nicht überdeckt.**
//
// # Woran das Feld hier zu erkennen ist
//
// An `default:` — denn diese Vorlage führt kein Spielpaket, und ohne das ist
// die **Art** des Feldes unbekannt (`istBildfeld`). Nachgezählt: über alle 190
// Stile des Packs steht `default:` an **genau zwei** Feldnamen — `image` und
// `image 2` — und **nie** neben einem `image:`.
//
// # Warum ein Befund und trotzdem eine Ebene
//
// Das Kunstbild liegt im **Set**-Paket, und wo dieses Set liegt, weiß diese
// Vorlage nicht ([service.Vorlage.Setpaket] ist leer). Das Ersatzbild wortlos
// zu zeichnen hieße, das Bild der Karte still verschwinden zu lassen — und eine
// Karte, die falsch aussieht, ohne dass etwas fehlt, ist der schlimmste der
// drei Fälle. Gezeichnet wird trotzdem (§8).
//
// **Mit Ort steht dort die Illustration** — das prüft `setbild_test.go`.
func TestEinEigenesKartenbildWirdGemeldetStattUeberdeckt(t *testing.T) {
	feld := flaeche("image")
	feld.Standardbild = "ersatz.png"

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "image", Text: "meine_kunst.jpg"},
	}}

	plan, befunde := gedeutetMit(t, karte, bildstil(feld))

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet einer: %v", len(befunde), befunde)
	}

	if befunde[0].Ort != "image" ||
		!strings.Contains(befunde[0].Meldung, "meine_kunst.jpg") {
		t.Errorf("Befund %+v", befunde[0])
	}

	gezeichnet := ebenen(t, plan)
	if len(gezeichnet) != 1 || gezeichnet[0].Bild.Pfad != "/probe.mse-style/ersatz.png" {
		t.Errorf("Ebenen %+v", gezeichnet)
	}
}
