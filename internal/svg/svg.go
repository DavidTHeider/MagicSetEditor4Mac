// Package svg schreibt eine Karte als SVG aus einem fertigen Zeichenplan.
//
// # Warum es dieses Paket gibt
//
// Die vier anderen Ausgaben entstehen aus **Bildpunkten**: `render` rastert,
// und was herauskommt, wird gepackt. **SVG muss aus dem Plan entstehen** — ein
// Rasterbild in ein `<svg>` zu wickeln wäre der Name ohne die Sache
// (#837).
//
// Damit ist dieses Paket der **zweite Verbraucher** desselben Plans, und genau
// das ist die Naht, die Phase P prüft: trägt die Trennung aus §2 — `render`
// bekommt einen Zeichenplan, keine Skripte — auch dann, wenn jemand anderes
// als der Rasterer ihn liest?
//
// # Was vektoriell wird und was nicht
//
// | Art | Wird | Warum |
// | --- | --- | --- |
// | Fläche | `<rect>`, `<linearGradient>` | rein geometrisch |
// | Text ohne Symbole | `<text>` mit `<tspan>` je Stück | der ganze Gewinn: markierbar, skalierbar, durchsuchbar |
// | Bild, Symbol | eingebettetes Rasterbild | die Quelle **ist** eines |
// | Text **mit** Symbolen, Symbolzug | eingebettetes Rasterbild der ganzen Ebene | siehe unten |
//
// **Die dritte Zeile ist keine Nachlässigkeit, sondern eine Zählung.** Im
// vollen Vorlagenpack liegen **23.966** PNG und JPG in `.mse-style`-Paketen
// und **5.004** PNG in `.mse-symbol-font`-Paketen; `.mse-symbol`-Dateien —
// das vektorielle Format des Setsymbol-Editors — gibt es **21** im ganzen
// Pack. Ein Kartenrahmen ist ein Bild, und daran ändert kein Ausgabeformat
// etwas.
//
// **Die vierte ist eine Grenze mit Grund.** Wo Manasymbole mitten im Regeltext
// stehen, entscheidet erst der Umbruch, wohin sie kommen; sie einzeln zu
// setzen hieße, die Symbolmaschine des Rasterers ein zweites Mal zu bauen —
// und zwei Rechnungen über dieselbe Sache laufen auseinander, sobald eine
// etwas lernt. Solche Ebenen gehen deshalb **durch den Rasterer** und kommen
// als Bild an: dieselben Bildpunkte wie im PNG, an derselben Stelle. Der Preis
// steht daneben — für diese Ebene ist der Text nicht mehr markierbar, und je
// solcher Ebene entsteht ein Bildpuffer in Kartengröße.
//
// # Was der Plan trägt und dieses Paket nicht ausdrückt
//
// **Genau eines**, und es wird je Ebene **gemeldet**: die [render.Ebene.Mischart].
// CSS `mix-blend-mode` kennt `add`, `subtract`, `stamp`, `negation`, `reflect`,
// `shadow` und `glow` nicht — und `shadow` allein steht 667 der 698 Aufrufe im
// vollen Pack (#470).
// Die fünf, die CSS gleichlautend führt, wären übernehmbar; sie zu übernehmen
// und die übrigen nicht ergäbe eine Ausgabe, die mal stimmt und mal nicht,
// ohne dass jemand es sieht. **Keine wird angewandt, jede wird genannt.**
//
// Alles andere steht im SVG: Deckung, Winkel, Maske, Verlauf, Passung, runde
// Ecken, Rand, Bildbaum, Textumbruch, Ausrichtung und Stauchung. Ein Wächter
// hält die Liste vollständig (`internal/werkzeuge/svgebene_test.go`): eine neue
// Eigenschaft an [render.Ebene], die hier niemand liest, macht die Pipeline
// rot.
//
// # Und es gibt keinen Weg zurück
//
// SVG trägt je Ebene ein `data-feld` und damit mehr Struktur als ein PNG —
// **zurücklesen lässt sich daraus trotzdem keine Karte**
// (#841).
// Die Ebenen sind **gezeichnet** und nicht die Werte: aus
// `<text>Feuerball</text>` ließe sich der Name holen, aus einem umgebrochenen
// Regeltext mit einem `<tspan>` je Zeile nicht mehr der Quelltext mit seinen
// `<i>`-Marken. Eine Einfuhr, die drei von zwanzig Feldern trifft, ist
// schlechter als keine.
package svg

import (
	"context"
	"fmt"
	"image"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Bildbaumquelle löst einen Bildbaum in ein fertiges Bild auf.
//
// **`*render.Rasterer` erfüllt sie.** Damit gehen Umfärbung, Mischbild und
// Symbolzeichnung durch **dieselbe** Maschine wie beim Rastern; ein eigener
// Weg hierher wäre die zweite Rechnung über dieselbe Sache.
type Bildbaumquelle interface {
	BildAusBaum(baum render.Bildbaum, breite, hoehe int) (*image.NRGBA, string)
}

// Rasterquelle zeichnet einen Plan.
//
// **Gebraucht für die Ebenen, die SVG nicht ausdrückt** — der Symbolzug und
// ein Textzug mit Symbolen. Sie werden als **eigener Plan mit einer Ebene**
// gezeichnet und als Bild eingebettet; das gibt dieselben Bildpunkte an
// derselben Stelle, ohne die Symbolmaschine ein zweites Mal zu bauen.
type Rasterquelle interface {
	Zeichne(
		ctx context.Context, plan *render.Plan, vergroesserung float64,
		guete render.Guete,
	) (
		*image.NRGBA, []render.Befund, error,
	)
}

// Quellen sind die Mittel, mit denen aus einem Plan ein SVG wird.
//
// **Dieselben drei, die auch der Rasterer bekommt**, und das ist Absicht: die
// Composition Root reicht die vorhandenen Stücke weiter, statt zweite
// anzulegen. Zwei Schriftsuchen wären zwei Antworten auf die Frage, wie breit
// ein Text ist.
//
// **Jede darf fehlen.** Dann bleibt weg, was sie getragen hätte — und es wird
// **gemeldet**, nicht verschwiegen: eine SVG ohne Kartenrahmen ist keine
// Karte, und ein Aufrufer, der davon nichts erfährt, verteilt sie.
type Quellen struct {
	Baeume    Bildbaumquelle
	Raster    Rasterquelle
	Schriften render.Schriftquelle

	// Symbole ist die Bildquelle, mit der [render.SetzeAbschnitte] die
	// Symbole eines Textzugs **misst**. Gezeichnet werden sie über
	// [Quellen.Raster].
	Symbole render.Bildquelle

	// Vergroesserung ist die Auflösung der eingebetteten Rasterbilder — `1`
	// heißt „so groß wie der Plan".
	//
	// **Sie gilt nur für sie.** Flächen und Text bleiben vektoriell und
	// skalieren mit; ein Kartenrahmen ist ein Bild, und wer eine SVG in
	// Druckgröße will, braucht ihn in Druckauflösung. Null heißt 1.
	Vergroesserung float64
}

// AusPlan schreibt den Plan als SVG.
//
// **Drei Rückgaben und nicht zwei.** Ein fehlendes Bild, eine fehlende Schrift
// und eine Mischart, die hier niemand anwendet, sind **Befunde** und kein
// Abbruch (§8) — aber sie verschwinden nicht. Bis zum 2026-08-22 gab dieser
// Weg `(string, error)`, ließ eine unlesbare Ebene kommentarlos weg, und ein
// Test namens `TestOhneBildquelleBleibtEinBefund` prüfte, dass **kein**
// `<image>` vorkommt. Eine SVG ohne Kartenrahmen entstand damit erfolgreich
// (#852).
func AusPlan(ctx context.Context, plan *render.Plan, quellen Quellen) (
	string, []render.Befund, error,
) {
	if plan == nil {
		return "", nil, fmt.Errorf("%w: es gibt keinen Plan", render.ErrPlan)
	}

	schreiber := &schreiber{quellen: quellen}

	var aus strings.Builder

	// **`width` in Zielgröße, `viewBox` in Vorlagenpunkten.** Damit ist die
	// SVG so breit, wie der Aufrufer sie bestellt hat, und ihre inneren
	// Stellen bleiben die des Plans — Flächen und Text skalieren mit, die
	// eingebetteten Bilder liegen in [Quellen.Vergroesserung] vor.
	//
	// Ohne diese Trennung müsste jede Stelle im Dokument umgerechnet werden,
	// und dann stünde die Rechnung an vierzig Stellen statt an einer.
	fmt.Fprintf(&aus,
		`<svg xmlns="http://www.w3.org/2000/svg" `+
			`width="%s" height="%s" viewBox="0 0 %s %s">`,
		zahl(plan.Breite*schreiber.vergroesserung()),
		zahl(plan.Hoehe*schreiber.vergroesserung()),
		zahl(plan.Breite), zahl(plan.Hoehe))

	// **Die Nummer aus dem ursprünglichen Plan.** Sie ist die einzige Angabe,
	// die eine Ebene eindeutig macht — ein Feld kann mehrere ergeben.
	for nummer, ebene := range plan.Ebenen {
		if err := ctx.Err(); err != nil {
			return "", nil, err
		}

		schreiber.ebene(ctx, &aus, plan, nummer, ebene)
	}

	// **Die Vorlagen stehen am Ende, nicht am Anfang.** Ein `<defs>` darf
	// überall im Baum stehen und wird über seine Kennung gefunden, gleich wo
	// es steht; erst hier ist bekannt, welche Verläufe und Masken überhaupt
	// gebraucht wurden.
	if vorlagen := schreiber.vorlagen.String(); vorlagen != "" {
		aus.WriteString("<defs>" + vorlagen + "</defs>")
	}

	aus.WriteString("</svg>")

	return aus.String(), schreiber.befunde, nil
}

// schreiber sammelt Vorlagen und Befunde über den ganzen Lauf.
type schreiber struct {
	quellen  Quellen
	vorlagen strings.Builder
	befunde  []render.Befund
	naechste int
}

// melde hängt einen Befund an.
func (s *schreiber) melde(nummer int, ebene render.Ebene, meldung string) {
	s.befunde = append(s.befunde, render.Befund{
		Ebene:   nummer,
		Feld:    ebene.Feld,
		Meldung: meldung,
		Art:     render.ArtAus(meldung),
	})
}

// name gibt eine im Dokument eindeutige Kennung.
func (s *schreiber) name(vorsilbe string) string {
	s.naechste++

	return fmt.Sprintf("%s%d", vorsilbe, s.naechste)
}

// ebene schreibt eine einzelne Ebene samt ihrer Hülle.
//
// # Dieselben zwei Wächter wie im Rasterer
//
//	if deckung <= 0 { return "" }
//	if !Sichtbar(ebene.Rahmen) { return "" }
//
// `internal/render/rasterer.go`. **Deckung 0 ist unsichtbar, nicht
// undurchsichtig** — der Nullwert eines Go-Verbunds ist 0, und eine Ebene, die
// jemand ohne Deckung hinschreibt, soll nichts zeichnen statt alles zu
// überdecken.
func (s *schreiber) ebene(
	ctx context.Context, aus *strings.Builder, plan *render.Plan,
	nummer int, ebene render.Ebene,
) {
	if ebene.Deckung <= 0 || !render.Sichtbar(ebene.Rahmen) {
		return
	}

	if ebene.Mischart != "" && ebene.Mischart != "normal" {
		s.melde(nummer, ebene, "die Mischart "+ebene.Mischart+
			" wird im SVG nicht angewandt")
	}

	// **Ein schräger Winkel wird hier gedreht und im Bild nicht.** SVG kann
	// jeden Winkel, `render.ebeneGedreht` nur die geraden; wer beide Ausgaben
	// derselben Karte nebeneinanderlegt, soll wissen, warum sie sich
	// unterscheiden (§15, Punkt 7). Im vollen Pack stehen schräge Winkel
	// viermal gegen 142 gerade.
	if ebene.Winkel != 0 && !gerade(ebene.Winkel) {
		s.melde(nummer, ebene, "ein Winkel von "+zahl(ebene.Winkel)+
			" Grad ist nicht gerade — das SVG dreht ihn, das Bild nicht")
	}

	huelle := s.huelle(nummer, ebene)

	// **Der Rahmen rückt an den Ursprung, wenn gedreht wird.** MSE gibt jedem
	// Feld einen eigenen gedrehten Zeichenkontext, und darin läuft alles, als
	// wäre nichts gedreht (`src/render/value/viewer.cpp`, Zeilen 44–46,
	// Fassung f9d9356 vom 2021-07-06). Ohne diesen Schritt wirkte die
	// Verschiebung zweimal.
	innen := ebene
	if ebene.Winkel != 0 {
		innen.Rahmen.Links, innen.Rahmen.Oben = 0, 0
	}

	aus.WriteString(huelle.auf)
	s.inhalt(ctx, aus, plan, nummer, ebene, innen)
	aus.WriteString(huelle.zu)
}

// inhalt schreibt, was eine Ebene ihrer Art nach zeichnet.
func (s *schreiber) inhalt(
	ctx context.Context, aus *strings.Builder, plan *render.Plan,
	nummer int, ebene, innen render.Ebene,
) {
	switch ebene.Art {
	case render.ArtFlaeche:
		s.flaeche(aus, innen)

	case render.ArtBild, render.ArtSymbol:
		s.bild(aus, nummer, ebene, innen)

	case render.ArtText:
		s.text(ctx, aus, plan, nummer, ebene, innen)

	case render.ArtSymbolzug:
		s.alsRaster(ctx, aus, plan, nummer, ebene, innen)
	}
}
