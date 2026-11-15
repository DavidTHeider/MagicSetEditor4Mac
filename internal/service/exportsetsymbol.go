package service

import (
	"bytes"
	"fmt"
	"image"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Das Setsymbol als Bild** —
// #660.
//
// # Was hier dazukommt
//
// `symbol_variation` gibt es beim **Zeichnen** einer Karte längst
// (`stilumgebung.go`, #323); dort steht ein Stil daneben, aus dem die Fassungen
// kommen. In der Ausfuhr gibt es keinen — die Fassungen reisen deshalb mit der
// [msepaket.Exportvorlage], gelesen aus dem Spielpaket.
//
// # Und `write_image_file` bekommt seine zweite Gestalt
//
//	ScriptObject<CardP>* card = dynamic_cast<ScriptObject<CardP>*>(input.get());
//	if (card) { … } else { image = input->toImage()->generateConform(options); }
//
// `src/script/functions/export.cpp`, Zeilen 414–421, Fassung f9d9356 vom
// 2021-07-06. Bis hierher nahm die Funktion eine **Karte**; beide Vorlagen, die
// das Setsymbol schreiben, geben ihr einen Bildwert.

// Baumrasterer führt einen Bildbaum aus.
//
// **Zusammen mit [Symbolrasterer] erfüllt `*render.Rasterer` beide** — es ist
// derselbe Rasterer mit derselben Bildquelle. Zwei Schnittstellen statt einer,
// weil zwei Fähigkeiten getrennt erlaubt werden: `symbols_to_html` braucht
// Symbolschriften, `symbol_variation` braucht eine Symbolzeichnung.
type Baumrasterer interface {
	BildAusBaum(baum render.Bildbaum, breite, hoehe int) (bild *image.NRGBA, befund string)
}

// symbolfassung baut die eingebaute Funktion `symbol_variation`.
//
// **Dieselbe Umsetzung wie beim Zeichnen**: [symbolfassungsfrage] gibt den
// Bildwert, und der Baum daraus ist der, den der Rasterer erwartet. Eine
// zweite Fassung daneben liefe auseinander (§10).
func (l *ausfuhrlauf) symbolfassung(fassungen []msepaket.Symbolfassung) msescript.Wert {
	return symbolfassungsfrage(fassungen)
}

// bildwertSchreiben rastert einen Bildwert und legt ihn ab.
func (l *ausfuhrlauf) bildwertSchreiben(
	wert msescript.Wert, datei string, breite, hoehe int,
) (string, error) {
	baum, ist := wert.AlsBild()
	if !ist {
		return "", fmt.Errorf(
			"%w: write_image_file bekam weder eine Karte noch ein Bild", domain.ErrValidation)
	}

	if l.mittel.Symbolzeichnungen == nil {
		return "", fmt.Errorf(
			"%w: write_image_file kann hier nur Karten schreiben", domain.ErrValidation)
	}

	// **Der Verweis trägt den Ort des Sets.** `set.symbol` nennt eine Datei
	// **im Set**; ohne den Ort suchte die Quelle sie in der Paketwurzel und
	// fände sie nie — dieselbe Regel wie bei den Bildfeldern
	// (#380).
	gerastert := alsZeichenbaum(baum, "", setsymbolquelle{
		quelle:   l.mittel.Symbolzeichnungen,
		setpaket: l.mittel.Setpaket,
	})

	bild, befund := l.ausfuehrer.baumrasterer().BildAusBaum(gerastert, breite, hoehe)
	if bild == nil {
		return "", fmt.Errorf("%w: das Bild für %s: %s", domain.ErrNotFound, datei, befund)
	}

	l.melde(datei, befund)

	var puffer bytes.Buffer

	if err := l.kodiere(&puffer, bild); err != nil {
		return "", err
	}

	if _, err := l.mittel.Beigaben.Lege(datei, puffer.Bytes()); err != nil {
		return "", err
	}

	l.gelegt[datei] = [2]int{bild.Bounds().Dx(), bild.Bounds().Dy()}

	return datei, nil
}

// setsymbolquelle hängt den Ort des Sets an jeden Verweis.
//
// **Hier und nicht in [setangaben]**: dort ist `set.symbol` ein Feldwert, wie
// er in der Datei steht, und jede Vorlage, die ihn als Text liest, bekäme sonst
// einen Verweis zu sehen.
type setsymbolquelle struct {
	quelle   Symbolquelle
	setpaket string
}

func (s setsymbolquelle) Symbolgestalt(verweis string) msepaket.Symbolzeichnung {
	name := strings.TrimSpace(verweis)

	// **Ein Verweis, der seinen Ort schon trägt, bleibt** — und ein Pfad, der
	// mit `/` beginnt, meint ein anderes Paket.
	if _, _, ist := msepaket.ZerlegeSetpaketverweis(name); ist || strings.HasPrefix(name, "/") {
		return s.quelle.Symbolgestalt(name)
	}

	return s.quelle.Symbolgestalt(msepaket.SetpaketVerweis(s.setpaket, name))
}

// setsymbolSetzen legt `symbol_variation` in die Umgebung, wenn beides da ist.
//
// **Beides**: die Fassungen aus dem Spielpaket und eine Quelle für die
// Zeichnung. Fehlt eine von beiden, gibt es den Namen nicht — und die Vorlage
// scheitert mit „ist hier nicht bekannt" statt mit einem leeren Bild.
func (l *ausfuhrlauf) setsymbolSetzen(
	umgebung *msescript.Umgebung, fassungen []msepaket.Symbolfassung,
) {
	if l.mittel.Symbolzeichnungen == nil || len(fassungen) == 0 {
		return
	}

	umgebung.Setze("symbol_variation", l.symbolfassung(fassungen))
}

// baumrasterer gibt den Rasterer als [Baumrasterer], wenn er einer ist.
//
// **Dieselbe Bauform wie [symbolquelle]**: der Ausführer bekommt einen
// [Symbolrasterer]; ob der auch Bäume rastert, entscheidet sein Typ. Wer es
// nicht tut, bekommt hier eine Attrappe, die es benennt — und nicht einen
// Absturz.
func (v *Vorlagenausfuehrer) baumrasterer() Baumrasterer {
	if kann, ist := v.symbolrasterer.(Baumrasterer); ist {
		return kann
	}

	return kannKeineBaeume{}
}

type kannKeineBaeume struct{}

func (kannKeineBaeume) BildAusBaum(
	render.Bildbaum, int, int,
) (bild *image.NRGBA, befund string) {
	return nil, "dieser Rasterer führt keine Bildbäume aus"
}

// ganzeAngabe liest ein benanntes Argument als ganze Zahl; fehlt es, ist es 0.
//
//	SCRIPT_OPTIONAL_PARAM_(int, width);
//	SCRIPT_OPTIONAL_PARAM_(int, height);
//
// `export.cpp`, Zeilen 412–413. **Null heißt „nicht angegeben"** und wird von
// `conform_image` als „so lassen" gelesen.
func ganzeAngabe(ruf msescript.Ruf, name string) int {
	wert, da := ruf.Benannter(name)
	if !da {
		return 0
	}

	zahl, err := wert.AlsZahl()
	if err != nil {
		return 0
	}

	return int(zahl)
}
