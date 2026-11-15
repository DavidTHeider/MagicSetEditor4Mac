package service

// **Die Angaben eines Feldstils deuten** — Farbe, Ausrichtung, Schrumpfen.
//
// Getrennt vom Planbauen in `stilplan.go`, weil es eine andere Frage ist: dort
// steht, **welche** Ebenen entstehen — hier, **wie** eine Zeichenkette aus einer
// Vorlage zu lesen ist.

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// ausrichtungAus deutet `alignment:`; unbekannt heißt links oben.
//
// **Die Angabe trägt mehrere Wörter** (`bottom shrink-overflow`, `middle
// right`), und nicht jedes davon ist eine Ausrichtung: `shrink-overflow` sagt,
// was bei zu langem Text geschieht, und wird von [stauchtAus] gelesen.
// Gedeutet wird, was wir kennen.

// stauchtAus sagt, ob ein zu breiter Zug waagerecht gestaucht wird.
//
// **Gesucht wird ein Teilwort, kein Wort.** MSE liest die Angabe mit
// `s.find(_("shrink"))` — `shrink-overflow` trifft damit, und `stretch` steht
// als ältere Schreibweise daneben (`src/util/alignment.cpp`, Zeilen 46–49,
// Fassung f9d9356 vom 2021-07-06).
func stauchtAus(angabe string) bool {
	return strings.Contains(angabe, "shrink") || strings.Contains(angabe, "stretch")
}

// nurBeiUeberlauf sagt, ob nur gestaucht und nie gedehnt wird — `overflow`.
func nurBeiUeberlauf(angabe string) bool {
	return strings.Contains(angabe, "overflow")
}

// ausrichtungstext gibt die Ausrichtungsangabe eines Feldes — ausgerechnet.
//
// **Sie darf ein Skript sein.** MSE führt `alignment` als
// `Scriptable<Alignment>` und rechnet sie mit jedem Wechsel neu
// (`TextStyle::update`, `src/data/field/text.cpp`, Zeile 95, Fassung f9d9356
// vom 2021-07-06); `magic-m15` entscheidet damit, ob ein kurzer Regeltext
// mittig steht.
//
// **216 der 1.597 Ausrichtungen im Pack stehen als Skript.** Wer den rohen
// Text als Wortliste läse, fände darin weder `middle` noch `left` verlässlich
// und bekäme die Vorgabe — links oben. Genau daran stand der Regeltext an der
// Oberkante seines Kastens
// (#345).
//
// **Ein Skript, das scheitert, ergibt die leere Angabe** und damit dieselbe
// Vorgabe wie gar keine. Eine erfundene Ausrichtung säße zuverlässig falsch.
func ausrichtungstext(
	ctx context.Context, feld *msepaket.Feldstil, umgebung *msescript.Umgebung,
) string {
	// **Der Skriptblock schlägt den Wert.** Steht `alignment:` mit einem
	// `script:` darunter, ist der Wert des Knotens leer.
	if skript := strings.TrimSpace(feld.AusrichtungSkript); skript != "" {
		return ausgerechnet(ctx, skript, umgebung)
	}

	roh := strings.TrimSpace(feld.Ausrichtung)

	rumpf, geklammert := skriptrumpf(roh)
	if !geklammert {
		return roh
	}

	return ausgerechnet(ctx, rumpf, umgebung)
}

// ausgerechnet wertet einen Skriptrumpf aus; was scheitert, ist leer.
func ausgerechnet(ctx context.Context, rumpf string, umgebung *msescript.Umgebung) string {
	wert, err := msescript.AuswertenQuelle(ctx, rumpf, umgebung)
	if err != nil {
		return ""
	}

	return wert.AlsText()
}

func ausrichtungAus(angabe string) render.Ausrichtung {
	woerter := strings.Fields(angabe)

	waagerecht, senkrecht := "left", "top"

	for _, wort := range woerter {
		switch wort {
		case "left", "center", "right":
			waagerecht = wort

		case "top", "middle", "bottom":
			senkrecht = wort
		}
	}

	return ausrichtungen[senkrecht+" "+waagerecht]
}

// ausrichtungen ordnet die Wortpaare des Stils den Werten des Rasterers zu.
var ausrichtungen = map[string]render.Ausrichtung{
	"top left":      render.LinksOben,
	"top center":    render.MitteOben,
	"top right":     render.RechtsOben,
	"middle left":   render.LinksMitte,
	"middle center": render.MitteMitte,
	"middle right":  render.RechtsMitte,
	"bottom left":   render.LinksUnten,
	"bottom center": render.MitteUnten,
	"bottom right":  render.RechtsUnten,
}

// Warum `card_style` nicht in der Umgebung steht
//
// `card_style.<feld>.content_width` fragt, **wie breit der gesetzte Text eines
// Feldes wäre** — 5 Feldstile von `M15-Magic-Pack` rechnen damit. Die Antwort
// hängt an der Schrift, an ihrer Größe und am Umbruch, und all das weiß erst
// `internal/render`: der Planbauer stellt den Plan **auf**, gesetzt wird er
// danach.
//
// **Eine Zahl zu erfinden wäre die schlechtere Antwort.** Sie sähe aus wie eine
// Messung und wäre geraten; ein Feld, das sich danach ausrichtet, säße
// zuverlässig falsch. Abgelehnt wird deshalb **benannt** (§8): der Befund nennt
// Feld, Angabe und den unbekannten Namen.
//
// **Was es bräuchte**, steht damit auch fest: eine Naht, über die der Planbauer
// den Satz nach einer Breite fragen kann, ohne ihn zu zeichnen — und die gibt
// es heute nicht (#195).

// Vorlage ist, was eine Karte zum Zeichnen braucht: ihr Stil und ihr Spiel.
//
// **Zwei Pakete und nicht eines.** Ein Stilpaket bringt Stelle, Größe und
// Schrift mit; die Funktionen, mit denen es rechnet, stehen zu einem großen
// Teil im **Spielpaket** — `magic.mse-game` legt sie über `init script:` an,
// und jeder Stil dieses Spiels benutzt sie
// (#195).
type Vorlage struct {
	Stil msepaket.Stil

	// Setangaben sind die Felder des `set info:`-Blocks — Titel, Symbol,
	// Rahmenfarbe, Copyright.
	//
	// # Warum sie hierher gehören
	//
	// `stilumgebung.go` sagte seit Phase 6: „`set` bleibt leer, und das ist
	// benannt." Ehrlich war das, einlösbar wurde es erst mit dem Leser für
	// `set_values` — geschrieben lag der Bestand seit Phase 2 da.
	//
	// Gezählt über `M15-Magic-Pack`: **420 Zugriffe auf `set.symbol`**, 345 auf
	// `set.border_color`, 101 auf `set.mainframe_rarity_name`. Nach
	// `stylesheet.card_width` der meistgelesene Name des Packs
	// (#321).
	Setangaben []domain.Value

	// Stilwerte sind die setweiten Stiloptionen, die **in dieser Anwendung**
	// gesetzt wurden — über dem Block `styling:` der Datei ([Stilwertquelle]).
	Stilwerte []domain.Value

	// Setfelder sind die `set field:`-Blöcke des Spielpakets.
	//
	// **Für die Vorgabewerte**, dieselbe Klasse wie bei [Vorlage.Kartenfelder]
	// (#278): ein Setfeld, das das Set nicht ausgefüllt hat, steht mit der
	// Vorgabe des Spiels da — und wenn es keine gibt, leer statt gar nicht.
	Setfelder []domain.Field

	// Kartenfelder sind die `card field:`-Blöcke des Spielpakets.
	//
	// **Für die Vorgabewerte** (`kartenvorgaben.go`): ein Feld, das die Karte
	// nicht führt, trägt den Wert des Spiels und nicht `nil`.
	Kartenfelder []domain.Field

	// Symbolschriften sind die `.mse-symbol-font`-Pakete, die dieser Stil
	// nennt — nach ihrem Namen.
	//
	// **Vorab geladen und nicht auf Zuruf**, weil der Planbauer keine Pakete
	// öffnet: er bekommt ein Modell und rechnet. Wer hier eine Paketwurzel
	// hineinreichte, machte aus dem Planbauer einen zweiten Paketleser.
	Symbolschriften map[string]msepaket.Symbolschrift

	// Symbolfassungen sind die `variation:`-Blöcke des **Spielpakets** — die
	// Grundfassungen des Setsymbols. Die der Vorlage stehen an
	// [msepaket.Stil.Symbolfassungen] und gewinnen bei gleichem Namen
	// (#323).
	Symbolfassungen []msepaket.Symbolfassung

	// Kartenskripte sind die `add cards script:`-Blöcke des Spielpakets
	// (#1337).
	Kartenskripte []Kartenskriptangabe

	// Auszaehldimensionen sind die `statistics dimension:`-Blöcke
	// (#1338).
	Auszaehldimensionen []Auszaehlangabe

	// Spielskript ist das `init script:` des Spielpakets, samt seinen Modulen.
	//
	// **Leer ist erlaubt und heißt: kein Spielpaket zur Hand.** Dann entstehen
	// die Feldebenen, die ohne Spielfunktionen auskommen, und die anderen
	// werden benannt (§8).
	Spielskript string

	// Setpaket ist der **Ort** des Set-Pakets auf der Platte.
	//
	// # Wofür
	//
	// Für das Bild der Karte. Ein Kartenfeld der Art `image` trägt einen
	// Dateinamen, und diese Datei liegt im Set — nicht in einer Vorlage:
	//
	//	auto image_file = getLocalPackage().openIn(value().filename);
	//
	// `ImageValueViewer::draw`, `src/render/value/image.cpp`, Zeilen 30–33,
	// Fassung f9d9356 vom 2021-07-06.
	//
	// # Warum ein Ort und kein Name
	//
	// Ein Stilpaket wird über die Paketwurzeln **gesucht**; ein Set liegt da,
	// wo jemand es abgelegt hat. Der Ort kommt aus [domain.SetHead.Origin] und
	// reist als [msepaket.SetpaketVerweis] bis zur Bildquelle
	// (#380).
	//
	// **Leer ist erlaubt und heißt: der Ort ist unbekannt.** Dann gilt das
	// Ersatzbild, und die Karte bekommt einen Befund — eine Illustration, die
	// wortlos verschwindet, sähe aus wie eine, die es nicht gibt.
	// Wortlisten sind die `word list:`-Blöcke des Spielpakets
	// (#1074).
	//
	// **Roh und nicht vorbereitet** — anders als [Vorlage.Schluesselwoerter].
	// Ein Schlüsselwort trägt einen regulären Ausdruck, der übersetzt werden
	// muss; eine Wortliste trägt Namen und Skripte, und die Skripte gelten je
	// **Karte**: `all_sub_types` sammelt aus dem Set und gibt bei jeder Karte
	// dasselbe erst dann, wenn niemand etwas ändert.
	Wortlisten []msepaket.Wortliste

	// Rohwoerter, Modi und Parameterarten sind die Schlüsselwörter des Spiels
	// in der Gestalt, in der sie **angezeigt** werden
	// (#1062).
	Rohwoerter     []msepaket.Schluesselwort
	Modi           []msepaket.Schluesselwortmodus
	Parameterarten []msepaket.Schluesselwortparameterart

	// Packs sind die `pack type:`-Blöcke des Spiels — die Booster
	// (#1072).
	Packs []msepaket.Pack

	// Schluesselwoerter ist die vorbereitete Wortliste des Spiels
	// (#1061).
	//
	// **Vorbereitet und nicht roh**: 334 Ausdrücke je Vorschau zu übersetzen
	// wäre bei jedem Tastendruck spürbar.
	Schluesselwoerter Wortliste

	Setpaket string

	// Bilderverzeichnis ist der Ort der Bilder, die jemand **in dieser
	// Anwendung** eingesetzt hat — leer, solange es keine gibt.
	//
	// # Warum neben [Vorlage.Setpaket] und nicht statt seiner
	//
	// [0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md):
	// die Bilder eines geöffneten Sets liegen in seinem Herkunftspaket und
	// nirgends sonst. Sie in die Ablage zu kopieren wären bei
	// `Multicolor.mse-set` 229 MB, nur damit sie an einem zweiten Ort noch
	// einmal stehen.
	//
	// **Die Ablage schlägt das Paket** ([kartenbild]): wer ein Bild einsetzt,
	// ersetzt damit das, was vorher dastand.
	Bilderverzeichnis string
}
