package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **`symbols_to_html` — Manasymbole als Bilder neben der Ausgabe** —
// #658.
//
// # Was es tut
//
// Es zerlegt eine Kostenkette in Symbole, rastert jedes **einmal**, legt es als
// PNG neben die Ausgabe und gibt `<img>`-Marken zurück.
//
//	String filename = symbol_font.name() + _("-") + clean_filename(sym.text) + _(".png");
//	map<String,wxSize>::iterator it = ei.exported_images.find(filename);
//	if (it == ei.exported_images.end()) { … }
//
// `src/script/functions/export.cpp`, Zeilen 183–205, Fassung f9d9356 vom
// 2021-07-06.
//
// # Die Merkliste gehört zum Lauf, nicht zum Ausführer
//
// Bei MSE steht sie in `ExportInfo`, und die lebt für **eine** Ausfuhr. Der
// `Vorlagenausfuehrer` dieses Programms wird einmal gebaut und für jede Ausfuhr
// wiederverwendet; eine Merkliste an ihm hielte Maße aus einem Lauf, dessen
// Verzeichnis längst ein anderes ist.

// Symbolschriftquelle gibt eine Symbolschrift nach ihrem Namen.
//
// **Ohne Dateiendung**, so wie die Vorlagen sie schreiben: `magic-mana-small`,
// nicht `magic-mana-small.mse-symbol-font`.
type Symbolschriftquelle interface {
	Symbolschrift(name string) (msepaket.Symbolschrift, bool)
}

// Symbolrasterer rastert ein einzelnes Symbol in ein eigenes Bild.
//
// **`*render.Rasterer` erfüllt sie**, ohne sie zu kennen — Gos strukturelle
// Typisierung, dieselbe Richtung wie bei den Repositories (§2).
type Symbolrasterer interface {
	SymbolAlsBild(symbol render.Symbol, masse render.Symbolmasse) (
		bild *image.NRGBA, befund string,
	)
}

// MitSymbolen gibt dem Ausführer die Fähigkeit, Symbolbilder zu schreiben.
//
// **Ohne sie gibt es den Namen nicht** — dieselbe Regel wie bei
// `write_image_file` (#216) und den drei Geschwistern (#657). Eine Vorlage, die
// `symbols_to_html` ruft, bekommt dann „ist hier nicht bekannt".
//
// **Und `to_html` bekommt seinen Rückruf**: dieselbe Umsetzung, ein Weg
// (#723). Ohne diese Fähigkeit bleibt ein `<sym>`-Block dort Text im `<span>`.
func (v *Vorlagenausfuehrer) MitSymbolen(
	schriften Symbolschriftquelle, rasterer Symbolrasterer,
) *Vorlagenausfuehrer {
	v.symbolschriften = schriften
	v.symbolrasterer = rasterer

	return v
}

// ausfuhrlauf hält, was während **einer** Ausfuhr über die Platte gegangen ist.
//
// **Ein Lauf für beide Schreiber.** `write_image_file` und `symbols_to_html`
// teilen sich bei MSE die Liste `ei.exported_images`; wer schon eine Datei
// dieses Namens geschrieben hat, schreibt nicht noch einmal (`export.cpp`,
// Zeilen 406–409, Fassung f9d9356 vom 2021-07-06).
type ausfuhrlauf struct {
	ausfuehrer *Vorlagenausfuehrer

	// mittel sind die Fähigkeiten und Orte **dieser** Ausfuhr — ein Wert, kein
	// Verweis (#736).
	mittel Ausfuhrmittel

	umgebung *msescript.Umgebung

	// abgelegt zählt, was dieser Lauf schon auf die Platte geschrieben hat —
	// #750.
	//
	// **Für die Meldung, nicht für eine Rücknahme.** Aufgeräumt wird nichts:
	// MSE schreibt ebenso inkrementell (`src/script/functions/export.cpp`,
	// Zeile 368, Fassung f9d9356 vom 2021-07-06), und §15.4 verbietet die
	// stillschweigende Änderung. Was sich ändern lässt, ist die Auskunft
	// darüber.
	abgelegt int

	// gelegt ist `ei.exported_images` — je Dateiname seine Maße.
	gelegt map[string][2]int

	// befunde sind die Meldungen, die beim Rastern angefallen sind und trotzdem
	// ein Bild ergeben haben.
	//
	// **Ein Befund und kein Fehler** (§8): eine noch nicht gebaute Verknüpfung
	// zeichnet das Ausgangsbild, und eine Beigabe, die zu neun Zehnteln stimmt,
	// ist brauchbar — eine fehlende ist es nicht. Verworfen wurden sie bis zum
	// 2026-08-21; angestrichen vom Zweitblick durch Codex.
	befunde []domain.Befund

	// packen macht aus einem Bild ein PNG.
	//
	// **Eine Naht und kein Aufruf**, aus demselben Grund wie bei [App.packen]:
	// mit einem `bytes.Buffer` kann `png.Encode` nicht scheitern, und der
	// Fehlerzweig wäre eine Behauptung darüber, was im Ernstfall geschieht
	// (§3). Nil heißt `png.Encode`.
	packen func(io.Writer, image.Image) error
}

// kodiere ist die Kodierung dieses Laufs.
func (l *ausfuhrlauf) kodiere(ziel io.Writer, bild image.Image) error {
	if l.packen != nil {
		return l.packen(ziel, bild)
	}

	return png.Encode(ziel, bild)
}

// wandler ist der Rückruf, den `to_html` und `symbols_to_html` teilen.
func (l *ausfuhrlauf) wandler(
	ctx context.Context, zug, schriftname string, groesse float64,
) (string, error) {
	if strings.TrimSpace(schriftname) == "" {
		return "", fmt.Errorf("%w: symbols_to_html braucht ein Argument `symbol_font:`",
			domain.ErrValidation)
	}

	schrift, da := l.ausfuehrer.symbolschriften.Symbolschrift(schriftname)
	if !da {
		return "", fmt.Errorf("%w: die Symbolschrift %s liegt nicht vor",
			domain.ErrNotFound, schriftname)
	}

	treffer, _ := zerlegeInSymbole(zug, schrift)

	// **Die Maße stehen einmal für den ganzen Zug.** `symbol_font_size` gilt
	// je Aufruf, der Rest kommt aus dem Paket — dieselbe Zusammensetzung wie
	// bei [symbolmasse], nur ohne Feldstil: eine eigene Datei hat weder Rahmen
	// noch Ausrichtung.
	masse := render.Symbolmasse{
		Groesse:           groesse,
		Bildgroesse:       schrift.Bildgroesse,
		AbstandWaagerecht: schrift.AbstandWaagerecht,
		AbstandSenkrecht:  schrift.AbstandSenkrecht,
	}

	symbole := symbolbilder(ctx, treffer, schriftname, l.umgebung)

	var gebaut strings.Builder

	for i := range symbole {
		marke, err := l.eineMarke(schriftname, symbole[i], masse)
		if err != nil {
			return "", err
		}

		gebaut.WriteString(marke)
	}

	return gebaut.String(), nil
}

// eineMarke legt ein Symbol ab, sofern nötig, und gibt seine `<img>`-Marke.
func (l *ausfuhrlauf) eineMarke(
	schriftname string, symbol render.Symbol, masse render.Symbolmasse,
) (string, error) {
	datei := schriftname + "-" + sauberergDateiname(symbol.Text) + ".png"

	grosse, schonda := l.gelegt[datei]
	if !schonda {
		bild, befund := l.ausfuehrer.symbolrasterer.SymbolAlsBild(symbol, masse)
		if bild == nil {
			return "", fmt.Errorf("%w: das Symbol %s: %s",
				domain.ErrNotFound, symbol.Text, befund)
		}

		l.melde(datei, befund)

		var puffer bytes.Buffer

		if err := l.kodiere(&puffer, bild); err != nil {
			return "", err
		}

		if _, err := l.mittel.Beigaben.Lege(datei, puffer.Bytes()); err != nil {
			return "", err
		}

		grosse = [2]int{bild.Bounds().Dx(), bild.Bounds().Dy()}
		l.gelegt[datei] = grosse
		l.abgelegt++
	}

	return "<img src='" + datei + "' alt='" + htmlMaskiert(symbol.Text) +
		"' width='" + strconv.Itoa(grosse[0]) +
		"' height='" + strconv.Itoa(grosse[1]) + "'>", nil
}

// sauberergDateiname ist MSEs `clean_filename`.
//
//	bool is_filename_char(Char c) {
//	  return isAlnum(c) || c == _(' ') || c == _('_') || c == _('-') || c == _('.');
//	}
//	bool is_filename_char_trim(Char c) { return c == _(' ') || c == _('.'); }
//
// `src/util/file_utils.cpp`, Zeilen 47–72, Fassung f9d9356 vom 2021-07-06.
//
// **`isAlnum` ist bei wxWidgets unicodefähig** — ein `é` bleibt stehen. Was
// wegfällt, ist der Schrägstrich in einem Code wie `T/2`; sonst schriebe die
// Vorlage in ein Unterverzeichnis, das niemand gemeint hat.
func sauberergDateiname(text string) string {
	var gebaut strings.Builder

	for _, zeichen := range text {
		if !unicode.IsLetter(zeichen) && !unicode.IsDigit(zeichen) &&
			zeichen != ' ' && zeichen != '_' && zeichen != '-' && zeichen != '.' {
			continue
		}

		// **Führende Leerzeichen und Punkte fallen weg.**
		if gebaut.Len() == 0 && (zeichen == ' ' || zeichen == '.') {
			continue
		}

		gebaut.WriteRune(zeichen)
	}

	sauber := strings.TrimRight(gebaut.String(), " .")
	if sauber == "" {
		return "no-name"
	}

	return sauber
}

// htmlMaskiert ist MSEs `html_escape`.
//
//	if (c == ESCAPED_LANGLE || c == _('<')) ret += _("&lt;");
//	else if (c == _('>')) ret += _("&gt;");
//	else if (c == _('&')) ret += _("&amp;");
//	else if (c == _('\'')) ret += _("&#39;");
//	else if (c == _('\"')) ret += _("&quot;");
//	else if (c >= 0x80) ret += String(_("&#")) << (int)c << _(';');
//
// `src/script/functions/export.cpp`, Zeilen 158–178, Fassung f9d9356 vom
// 2021-07-06.
//
// **Strenger als der Weg in `to_html`**, und das mit Grund: hier steht der Text
// in einem **Attribut**, und ein Hochkomma darin beendete es.
func htmlMaskiert(text string) string {
	var gebaut strings.Builder

	for _, zeichen := range text {
		switch {
		case zeichen == '\u0001' || zeichen == '<':
			gebaut.WriteString("&lt;")

		case zeichen == '>':
			gebaut.WriteString("&gt;")

		case zeichen == '&':
			gebaut.WriteString("&amp;")

		case zeichen == '\'':
			gebaut.WriteString("&#39;")

		case zeichen == '"':
			gebaut.WriteString("&quot;")

		case zeichen >= 0x80:
			gebaut.WriteString("&#" + strconv.Itoa(int(zeichen)) + ";")

		default:
			gebaut.WriteRune(zeichen)
		}
	}

	return gebaut.String()
}

// symbolfunktion baut die eingebaute Funktion `symbols_to_html`.
//
//	SCRIPT_PARAM_N(String, _("symbol_font"), font_name);
//	SCRIPT_OPTIONAL_PARAM_(double, symbol_font_size);
//	if (symbol_font_size <= 0) symbol_font_size = 12;
//
// `export.cpp`, Zeilen 287–294. **`symbol_font:` ist Pflicht** — anders als bei
// `to_html`, wo ohne Schrift der Text im `<span>` bleibt.
func (l *ausfuhrlauf) symbolfunktion() msescript.Eingebaut {
	return func(ctx context.Context, ruf msescript.Ruf) (msescript.Wert, error) {
		if err := ruf.EineEingabe("symbols_to_html"); err != nil {
			return msescript.Nichts, err
		}

		schrift := ruf.AusRufOderUmgebung("symbol_font").AlsText()

		groesse, unlesbar := ruf.AusRufOderUmgebung("symbol_font_size").AlsZahl()
		if unlesbar != nil || groesse <= 0 {
			groesse = msescript.HTMLVorgabegroesse
		}

		marken, err := l.wandler(ctx, ruf.Stellung[0].AlsText(), schrift, groesse)
		if err != nil {
			return msescript.Nichts, err
		}

		return msescript.Text(marken), nil
	}
}

// melde nimmt eine Meldung auf, die trotz eines Bildes angefallen ist.
func (l *ausfuhrlauf) melde(datei, befund string) {
	if befund == "" {
		return
	}

	l.befunde = append(l.befunde, domain.Befund{
		Ort:     datei,
		Meldung: befund,
		Art:     domain.ArtPortierung,
	})
}
