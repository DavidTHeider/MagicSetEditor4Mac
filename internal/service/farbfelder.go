package service

// **Ein Farbfeld trägt eine Farbe, keinen Text.**
//
// # Die Regel, nachgelesen in MSE2
//
//	ColorValue::ColorValue(const ColorFieldP& field)
//	  : Value(field)
//	  , value( !field->initial.isDefault() ? field->initial()
//	         : !field->choices.empty()     ? field->choices[0]->color
//	         :                               Color()
//	         , true)
//	{}
//
// `src/data/field/color.cpp`, Zeilen 82–89, Fassung f9d9356 vom 2021-07-06.
//
// # Was das gekostet hat
//
// `magic.mse-game` rechnet an vier Ebenen je Karte
//
//	color := styling.copyright_text_color or else set.copyright_text_color
//	recolor_image(image, color: color)
//
// und `set field: copyright text color` ist ein Farbfeld mit den Wahlen
// `white` und `black`. Ohne Vorgabe stand dort der leere Text, `recolor_image`
// bekam keine Farbe und meldete es — **vier Befunde auf jeder der vier
// Referenzkarten**, und die Ebene blieb in ihrer Ausgangsfarbe
// (#270).

import (
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// farbfeld ist, was ein Farbfeld über seine Vorgabe hergibt.
//
// **Einer für [domain.Field] und [domain.StyleField]**: die Regel ist
// dieselbe — `initial:`, sonst die erste Wahl, sonst Schwarz —, und zwei
// Kopien davon liefen auseinander, sobald eine von beiden einen Fall dazu
// bekäme.
type farbfeld struct {
	initial    string
	wahlfarben []string
}

// farbkartenfelder sammelt die Farbfelder einer Feldliste.
func farbkartenfelder(felder []domain.Field) map[string]farbfeld {
	gesammelt := make(map[string]farbfeld)

	// **Über den Index und nicht über den Wert**: ein [domain.Field] ist 128
	// Byte groß, und diese Schleife läuft je Karte über 249 Felder.
	for i := range felder {
		if felder[i].Kind != domain.KindColor {
			continue
		}

		name := strings.ReplaceAll(felder[i].Name, " ", "_")
		gesammelt[name] = farbfeld{
			initial: felder[i].Initial, wahlfarben: felder[i].Wahlfarben,
		}
	}

	return gesammelt
}

// farbstilfelder sammelt die Farbfelder einer Stiloptionsliste.
func farbstilfelder(optionen []domain.StyleField) map[string]farbfeld {
	gesammelt := make(map[string]farbfeld)

	for i := range optionen {
		if optionen[i].Kind != domain.KindColor {
			continue
		}

		name := strings.ReplaceAll(optionen[i].Name, " ", "_")
		gesammelt[name] = farbfeld{
			initial: optionen[i].Initial, wahlfarben: optionen[i].Wahlfarben,
		}
	}

	return gesammelt
}

// mitFarbwerten macht aus den Einträgen der Farbfelder Farbwerte.
//
// **Eine Kopie und nicht am Platz.** Die Einträge werden weiterverwendet — der
// Text eines Feldes wird gezeichnet, gemessen und von einem Wertskript
// gelesen —, und ein Farbwert an dieser Stelle machte aus `rgb(0,0,0)` ein
// `rgba(0,0,0,255)` im gezeichneten Text.
func mitFarbwerten(
	eintraege []msescript.Eintrag, farbfelder map[string]farbfeld,
) []msescript.Eintrag {
	if len(farbfelder) == 0 {
		return eintraege
	}

	gewandelt := make([]msescript.Eintrag, len(eintraege))
	copy(gewandelt, eintraege)

	for stelle := range gewandelt {
		feld, ist := farbfelder[gewandelt[stelle].Name]
		if !ist {
			continue
		}

		gewandelt[stelle].Wert = farbfeldwert(gewandelt[stelle].Wert.AlsText(), feld)
	}

	return gewandelt
}

// farbfeldwert gibt die Farbe, die ein Farbfeld mit diesem Text trägt.
//
// Die Reihenfolge ist MSEs: der **Wert**, sonst `initial:`, sonst die **erste
// Wahl**, sonst Schwarz.
//
// **Ein ungedeuteter Text ist kein Grund, den nächsten Schritt zu nehmen.**
// Steht im Set `karminrot`, ist das Feld gefüllt und die Farbe schwarz — MSEs
// `parse_color` warnt und nimmt `Color()`; die Vorgabe des Spiels kommt dann
// nicht mehr zum Zug.
func farbfeldwert(text string, feld farbfeld) msescript.Wert {
	farbe := farbfeldfarbe(text, feld)

	return msescript.Farbwert(msescript.Farbe{R: farbe.R, G: farbe.G, B: farbe.B, A: farbe.A})
}

// farbfeldfarbe gibt dieselbe Farbe für den Rasterer.
func farbfeldfarbe(text string, feld farbfeld) render.Farbe {
	roh := strings.TrimSpace(text)

	if roh == "" {
		roh = strings.TrimSpace(feld.initial)
	}

	if roh == "" && len(feld.wahlfarben) > 0 {
		roh = feld.wahlfarben[0]
	}

	return farbeAus(roh)
}

// kartenabbildung ist `card` — die Felder, mit den Farbfeldern als Farben.
//
// **An einer Stelle und nicht an drei.** `card` wird während des Plans dreimal
// gesetzt — nach den Vorgaben, nach den Vorgabeskripten, nach den Wertskripten
// —, und eine Wandlung, die nur an einer davon steht, gilt für die Skripte
// dazwischen nicht.
func kartenabbildung(
	felder []msescript.Eintrag, beschreibungen []domain.Field, setpaket string,
) msescript.Wert {
	return kartenabbildungMit(felder, beschreibungen, setpaket, "")
}

// kartenabbildungMit gibt dieselbe Abbildung mit einer **Kennung**.
//
// **Sie ist MSEs Zeiger auf die Karte** ([msescript.Wert.Kennung]): zwei
// Karten sind gleich, wenn es dieselbe Karte ist. Ohne sie galten zwei
// strukturgleiche Karten als dieselbe, und
//
//	if card != set.cards.0 then "" else (…)
//
// lief an jeder von ihnen — eine Auszählung zählte still doppelt (Befund 1 des
// Zweitblicks über Phase 4.11).
//
// **Leer heißt: keine Kennung**, und dann entscheidet die Gestalt. So bleibt
// jeder Aufrufer richtig, der keine Karte hat — der Rasterer etwa baut die
// Abbildung auch für eine Attrappe.
func kartenabbildungMit(
	felder []msescript.Eintrag, beschreibungen []domain.Field,
	setpaket, kennung string,
) msescript.Wert {
	return msescript.Wert{
		Art:     msescript.ArtAbbildung,
		Kennung: kennung,
		Eintraege: mitBildwerten(
			mitFarbwerten(felder, farbkartenfelder(beschreibungen)),
			beschreibungen, setpaket),
	}
}

// Skriptkartenkennung ist die Kennung einer Karte in der Skriptumgebung.
//
// **Exportiert, weil ein Lauf von außen sie prüft**: der Zweig „keine Karte"
// ist der des Rasterers, der auch für eine Attrappe baut.
//
// **Nicht `kartenkennung`** — der Name ist in `exportbild.go` vergeben und
// beantwortet dort eine andere Frage.
//
// **Über die Datenbankkennung und nicht über die Stelle in der Liste.** Eine
// Auszählung bekommt die Karten in einer Reihenfolge, die von der Sortierung
// abhängt; die Kennung eines Datensatzes hängt von nichts ab.
func Skriptkartenkennung(karte *domain.CardRow) string {
	if karte == nil {
		return ""
	}

	return "karte:" + strconv.FormatInt(karte.ID, 10)
}

// mitBildwerten macht aus dem Wert eines Bildfeldes ein **Bild**.
//
//	void ImageValue::reflect(GetMember& handler) {}
//	void ImageValue::reflect(GetDefaultMember& handler) {
//	  // convert to ScriptImageP for scripting
//	  handler.handle( (ScriptValueP)make_intrusive<ImageValueToImage>(filename, last_update) );
//	}
//
// `src/data/field/image.cpp`, Zeilen 49–53, Fassung f9d9356 vom 2021-07-06.
// `ImageValueToImage::generate` öffnet die Datei aus `opt.local_package`
// (`src/gfx/generated_image.cpp`, Zeilen 489–496, dazu `local() == true` in
// `generated_image.hpp`, Zeile 382) — und das lokale Paket einer Karte ist ihr
// **Set**.
//
// **Woran es aufgefallen ist:** `magic-m15-mainframe-planeswalker.mse-style`
// schreibt für das Feld `img mirror` schlicht `image: {card.image}`. Als Text
// gelesen wurde daraus der Dateiname `image16`, und der wurde gegen das
// **Stil**paket aufgelöst — wo er nicht liegt und nicht liegen kann. Fünf der
// neun verbleibenden Befunde des kanonischen Laufs waren dieser eine Fall
// (#440).
//
// **Ohne Ort des Set-Pakets bleibt der Text stehen.** Ein Verweis ohne Ort
// führte nirgendwohin; der bisherige Weg endet wenigstens in einer Meldung,
// die die Datei nennt.
func mitBildwerten(
	eintraege []msescript.Eintrag, beschreibungen []domain.Field, setpaket string,
) []msescript.Eintrag {
	if strings.TrimSpace(setpaket) == "" {
		return eintraege
	}

	bildfelder := map[string]bool{}

	for i := range beschreibungen {
		if beschreibungen[i].Kind == domain.KindImage {
			bildfelder[strings.ReplaceAll(beschreibungen[i].Name, " ", "_")] = true
		}
	}

	if len(bildfelder) == 0 {
		return eintraege
	}

	gewandelt := make([]msescript.Eintrag, len(eintraege))
	copy(gewandelt, eintraege)

	for stelle := range gewandelt {
		datei := strings.TrimSpace(gewandelt[stelle].Wert.AlsText())
		if !bildfelder[gewandelt[stelle].Name] || datei == "" {
			continue
		}

		gewandelt[stelle].Wert = msescript.Dateibild(
			msepaket.SetpaketVerweis(setpaket, datei))
		gewandelt[stelle].Wert.Marke = setbildmarke
	}

	return gewandelt
}

// farbfeldZu sucht das Farbfeld dieses Namens unter den Feldern der Vorlage.
//
// **Kartenfelder und Zusatzfelder**: `card stamp border` ist ein
// `extra card field:` des Stils und ein Farbfeld wie jedes andere.
func farbfeldZu(vorlage Vorlage, name string) (farbfeld, bool) {
	for _, felder := range [][]domain.Field{vorlage.Kartenfelder, vorlage.Stil.Zusatzfelder} {
		for i := range felder {
			if felder[i].Kind == domain.KindColor && felder[i].Name == name {
				return farbfeld{
					initial: felder[i].Initial, wahlfarben: felder[i].Wahlfarben,
				}, true
			}
		}
	}

	return farbfeld{}, false
}
