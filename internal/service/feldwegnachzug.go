package service

// **Der Nachzug am Schreibweg** — #2396
// (Issue (#2396)).
//
// Was gerechnet wird, steht in `wertnachzug.go`; hier steht, **was davon
// gespeichert wird**: das gesetzte Feld mit seinem gerechneten Wert, und jedes
// nachgezogene, das `save value` trägt — so wie MSE beim Sichern schreibt, was
// im Speicher steht (`TextValue`, `REFLECT_NAMELESS(value)` nur bei
// `save_value`, `src/data/field/text.cpp`, Zeilen 185–187, Fassung f9d9356 vom
// 2021-07-06).
//
// **Brechend** (§16, 0033 Regel 7, entschieden mit 0.9.0): wer über `mse`, MCP
// oder REST ein Feld setzt, findet danach den gerechneten Wert im Bestand und
// nicht den übergebenen.

import (
	"context"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// feldnachzug hält, was der Nachzug zum Rechnen braucht.
type feldnachzug struct {
	stile     Kartenstilquelle
	schriften render.Schriftquelle
	bilder    render.Bildquelle
}

// MitNachzug lässt den Feldweg speichern, was gerechnet ist — **wahlfrei**:
// ohne ihn geht der eingegebene Wert in den Bestand, wie bis #2396.
func (f *Feldweg) MitNachzug(
	stile Kartenstilquelle, schriften render.Schriftquelle, bilder render.Bildquelle,
) *Feldweg {
	f.nachzug = &feldnachzug{stile: stile, schriften: schriften, bilder: bilder}

	return f
}

// nachgezogen gibt die Werte, die gespeichert werden — **die gesetzten
// zuerst**, damit der Vorwert, den der Bestand zurückgibt, ihrer ist.
//
// **Ohne Nachzug, ohne Karte oder ohne Vorlage** bleibt, was gesetzt wurde:
// ohne Vorlage weiß niemand, was ein Feld rechnet (#720).
//
// **Und `stand` ist der Stand, aus dem gerechnet wurde** — Karte und
// Setangaben ([domain.Rechenstand]), leer,
// wenn nichts gerechnet wurde. Geschrieben wird auf ihm (Gesamtreview über
// 0.9.0, Befund 2): ein gerechneter Wert hängt am Rest der Karte, und ein
// Schreiber dazwischen ließe ihn zu keinem Stand passen.
func (n *feldnachzug) nachgezogen(
	ctx context.Context, karte *domain.CardRow, werte []domain.Value,
) (gespeichert []domain.Value, stand string) {
	if n == nil || karte == nil {
		return werte, ""
	}

	kopie := mitWerten(karte, werte)

	vorlage, _, gefunden := n.stile.StilFuerKarte(ctx, kopie)
	if !gefunden {
		return werte, ""
	}

	namen := make([]string, 0, len(werte))
	for _, wert := range werte {
		namen = append(namen, strings.ReplaceAll(wert.Field, " ", "_"))
	}

	eintraege, gerechnet := n.rechne(ctx, vorlage, kopie, namen)

	// **Der Stand trägt die Setangaben mit** (#2533): ein Wertskript liest
	// auch `set.*`, und eine Änderung daran macht den Wert so falsch wie eine
	// an der Karte.
	return gespeicherte(eintraege, gerechnet, vorlage.Kartenfelder, werte),
		domain.Rechenstand(karte.Stand(), vorlage.Setangaben)
}

// rechne rechnet die Felder `namen` auf der Karte und zieht nach, wer sie
// liest — für ein gesetztes Feld wie für eine neue Karte (#2516). **Die
// Befunde fallen weg**: geschrieben wird trotzdem, und das Zeichnen meldet
// sie.
func (n *feldnachzug) rechne(
	ctx context.Context, vorlage Vorlage, karte *domain.CardRow, namen []string,
) (eintraege []msescript.Eintrag, gerechnet []string) {
	eintraege, gerechnet, _ = neuerUmgebungsbauer(ctx, vorlage, n.schriften, n.bilder, karte).
		nachgezogen(ctx, namen)

	return eintraege, gerechnet
}

// mitWerten gibt eine Kopie der Karte, in der die gesetzten Werte stehen.
func mitWerten(karte *domain.CardRow, werte []domain.Value) *domain.CardRow {
	kopie := *karte
	kopie.Values = slices.Clone(karte.Values)

	for _, wert := range werte {
		stelle := slices.IndexFunc(kopie.Values, func(v domain.Value) bool { return v.Field == wert.Field })
		if stelle < 0 {
			kopie.Values = append(kopie.Values, wert)

			continue
		}

		kopie.Values[stelle] = wert
	}

	return &kopie
}

// gespeicherte macht aus den gerechneten Einträgen die Werte, die gesichert
// werden: die gesetzten unter ihrem Namen, die nachgezogenen nur mit
// `save value`.
func gespeicherte(
	eintraege []msescript.Eintrag, gerechnet []string, felder []domain.Field,
	werte []domain.Value,
) []domain.Value {
	text := make(map[string]string, len(eintraege))
	for _, eintrag := range eintraege {
		text[eintrag.Name] = eintrag.Wert.AlsText()
	}

	ergebnis := make([]domain.Value, 0, len(gerechnet))

	for _, wert := range werte {
		if gerechneterText, da := text[strings.ReplaceAll(wert.Field, " ", "_")]; da {
			wert.Text = gerechneterText
		}

		ergebnis = append(ergebnis, wert)
	}

	for _, name := range gerechnet[len(werte):] {
		stelle := slices.IndexFunc(felder, func(feld domain.Field) bool {
			return strings.ReplaceAll(feld.Name, " ", "_") == name
		})
		if stelle >= 0 && felder[stelle].Gespeichert {
			ergebnis = append(ergebnis, domain.Value{Field: felder[stelle].Name, Text: text[name], Nachgezogen: true})
		}
	}

	return ergebnis
}
