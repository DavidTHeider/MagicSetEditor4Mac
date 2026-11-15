package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Stilplan macht aus einer Karte und einem Stil einen Zeichenplan.
//
// # Die Naht, die AGENTS.md §2 verlangt
//
// `render` sieht kein Skript, `msescript` kennt keine Karte, `msepaket` wertet
// nichts aus. Hier laufen die drei zusammen: der Stil sagt **wo**, das Skript
// rechnet **wie viel**, und heraus kommt ein Plan.
//
// # Was er heute deutet — und was nicht
//
// Gedeutet sind **Textfelder** (Stelle, Größe, Schrift, Schriftgröße, Farbe),
// **Bildfelder** (`image:`, `default:`, `render style: hidden`) und **Masken**
// — je an der Ebene, zu der sie gehören.
//
// Seit #275 auch
// **Symbolschriften** — der Text wird in Codes zerlegt, in der Reihenfolge, in
// der die Schrift sie führt, und der Zug wird zu einer Ebene.
//
// **Nicht gedeutet sind die Wahlbilder des Spiels:** ein Feld, dessen Bild aus
// einer Auswahl des `.mse-game` kommt, hat im Stil nur `render style: image`
// stehen und sonst nichts. Ein Feld, das eines verlangt, wird **benannt** statt
// stillschweigend weggelassen (§8) — und mit seinem eigenen Grund, nicht mit
// einem Sammelsatz.
//
// **Und Symbolschriften sind nur zur Hälfte gedeutet.** Die Fassungen trennt
// seit dem 2026-08-17 die Bedingung (`symbolfassungen.go`), aber ein `image:`,
// das ein **Skript** ist, wird nicht ausgewertet: das Symbol nennt dann kein
// Bild und fällt beim Lesen des Pakets heraus. In `magic-mana-large` sind das
// die farbigen V-Symbole
// (#293).
// Ebenso wenig gedeutet sind Codes, die reguläre Ausdrücke sind. Der Satz steht
// hier, weil ein Kommentar, der die halbe Deutung als ganze ausgibt, den Rest
// unsichtbar macht (§10).
func (p *Planbauer) Stilplan(
	ctx context.Context, karte *domain.CardRow, vorlage Vorlage,
) (*render.Plan, []domain.Befund, error) {
	stil := vorlage.Stil

	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	breite, hoehe := float64(stil.Breite), float64(stil.Hoehe)
	if breite <= 0 || hoehe <= 0 {
		return nil, nil, fmt.Errorf("%w: der Stil nennt keine Kartengröße", domain.ErrValidation)
	}

	// **Ein weißer Grund, und der ist unsere Zutat.** Der Stil zeichnet seinen
	// Hintergrund über ein Bild (`card color` mit `render style: image`) — und
	// dessen Datei liegt im Stilpaket, das auf diesem Rechner fehlen kann. Ohne
	// Grund wäre die Karte dann durchsichtig, und eine durchsichtige Karte
	// sieht im Druck aus wie eine leere.
	plan := &render.Plan{
		Breite: breite, Hoehe: hoehe, DPI: stil.DPI, Stil: stil.Paket,
		Ebenen: []render.Ebene{{
			Art:     render.ArtFlaeche,
			Rahmen:  render.Rechteck{Breite: breite, Hoehe: hoehe},
			Farbe:   render.Farbe{R: 255, G: 255, B: 255, A: 255},
			Deckung: 1,
		}},
	}

	umgebung, gerechnet, befunde := umgebungAus(ctx, karte, vorlage, p.schriften, p.bilder)

	// **Was gerechnet wurde und nicht von der Karte kommt** (#563). Die Werte
	// stehen hier fertig da; ein zweiter Lauf rechnete dasselbe noch einmal.
	plan.Vorgaben = vorgabenAus(gerechnet, vorlage.Kartenfelder, karte)

	// **Welche Fassung eines Symbols gilt, steht erst jetzt fest.** Die
	// Bedingungen fragen nach Funktionen, die das Initskript des Stils anlegt,
	// und das läuft in [umgebungAus] als Letztes.
	//
	// **Damit misst die erste Stufe gegen die ungefilterte Schrift.**
	// `card_style.<feld>.content_width` entsteht in [umgebungAus] und zählt
	// deshalb noch jede Fassung. Für die Breite macht das keinen Unterschied —
	// die Fassungen eines Codes tragen denselben Code —, und der Satz steht
	// hier, damit niemand die Reihenfolge später für beliebig hält.
	schriften, fassungsbefunde := aktiveSymbolschriften(ctx, vorlage.Symbolschriften, umgebung)
	vorlage.Symbolschriften = schriften
	befunde = append(befunde, fassungsbefunde...)

	// **Kartenfelder und Zusatzfelder in einer Reihenfolge** — der Grund steht
	// bei [alleFeldstile].
	zusatz := zusatzwerte(umgebung)

	for _, feld := range alleFeldstile(stil.Felder, stil.Zusatzstile) {
		werte := gerechnet
		if feld.zusatz {
			werte = zusatz
		}

		feldebenen, flaeche, feldbefunde := feldebene(
			ctx, feld.stil, werte, vorlage, breite, hoehe, umgebung,
			// **Die Ablage zuerst** (#890): eine gezeichnete `.mse-symbol`
			// liegt neben dem Set und kommt erst beim Sichern ins Paket. Ohne
			// diesen Schritt zeigte die Karte MSEs Standardquadrat, bis
			// jemand gespeichert hat — und niemand wüsste, warum.
			SymbolquelleAusAblage(vorlage.Bilderverzeichnis, symbolquelle(p.bilder),
				vorlage.Setpaket), p.bilder)

		// **Der Feldname wird hier gesetzt und nicht an jeder Bauzeile.**
		//
		// Ein Feldstil ergibt je nach Art eine Bild-, Flächen-, Text- oder
		// Symbolebene, und mehrere davon zugleich; sie alle einzeln zu
		// beschriften hieße, den Namen an fünf Stellen zu wiederholen — und an
		// der sechsten zu vergessen. Hier steht er einmal, für alles, was
		// dieser Feldstil hergibt
		// (#246).
		for stelle := range feldebenen {
			feldebenen[stelle].Feld = feld.stil.Feld
		}

		// **Und die Fläche des Feldes, ob es zeichnet oder nicht.** Sie ist
		// die Auskunft an die Oberfläche, welche Felder auf der Karte liegen
		// ([render.Feldflaeche]).
		if flaeche != nil {
			plan.Feldflaechen = append(plan.Feldflaechen,
				render.Feldflaeche{Feld: feld.stil.Feld, Rahmen: *flaeche})
		}

		befunde = append(befunde, feldbefunde...)
		plan.Ebenen = append(plan.Ebenen, feldebenen...)
	}

	return plan, befunde, nil
}
