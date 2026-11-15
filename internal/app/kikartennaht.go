package app

// **Wie die Naht die Zielkarte liest und schreibt** — herausgezogen aus
// `kiwerkzeug.go`, als es mit dem Schloss aus dem Zweitblick über Phase 5.3
// (Befund 1) über die Zeilengrenze aus §2 ging.
//
// **Der Schnitt folgt der Frage**: dort steht, *welches Werk gerade gilt*,
// hier, *was es mit der Karte tut* — den Stand lesen, die übernommenen Werte
// samt Herkunft schreiben, und die Umrechnung dazwischen.

import (
	"context"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kikarten liest und schreibt die Zielkarte.
type kikarten struct{ app *App }

// Revision gibt den gerechneten Stand der Karte ([domain.CardRow.Stand]).
func (k kikarten) Revision(ctx context.Context, karte int64) (string, error) {
	if k.app.kartenleser == nil {
		return "", fehlenderBestand()
	}

	zeile, fehler := k.app.kartenleser.KarteLesen(ctx, karte)
	if fehler != nil {
		return "", fehler
	}

	return zeile.Stand(), nil
}

// FelderSetzen schreibt die übernommenen Werte.
//
// **Mit Herkunft** (§16, Satz 6): „Nichts wird still übernommen … und trägt
// danach seine Herkunft." Wer später auf das Feld sieht, soll lesen können,
// dass es nicht getippt wurde.
func (k kikarten) FelderSetzen(
	ctx context.Context, karte int64, felder map[string][]string,
	herkunft []domain.KIFeldherkunft, erwartet string,
) error {
	if k.app.kartenleser == nil || k.app.schreiber == nil {
		return fehlenderBestand()
	}

	zeile, fehler := k.app.kartenleser.KarteLesen(ctx, karte)
	if fehler != nil {
		return fehler
	}

	werte := alsWerte(felder)

	herkunft = mitAbdruckDesGeschriebenen(herkunft, werte)

	// **In einem Zug** — zwischen zwei Aufrufen läge ein Stand, in dem die
	// Karte die Hälfte eines Vorschlags trägt.
	//
	// **Und die Herkunft geht mit** (A11, #1450): nicht nur die Marke `ki`
	// am Feld, sondern wer geantwortet hat, worauf und wann. Beides in
	// derselben Transaktion — scheitert das zweite, steht auch das erste
	// nicht da.
	// **Ohne Bildangaben** (A12, #1451): ein Textvorschlag hat keine, und
	// eine leere Karte mitzugeben ist ehrlicher als ein zweiter Weg.
	// **Mit dem Stand als Bedingung** (Befund 2 des Zweitblicks über Phase
	// 5.3): geprüft wird er dort, wo geschrieben wird, und nicht hier.
	_, _, fehler = k.app.schreiber.FelderVonAussenMitHerkunft(
		ctx, zeile.SetID, karte, werte, kiherkunft, herkunft, nil, erwartet)

	return fehler
}

// kiherkunft steht an jedem Feld, das aus einem Vorschlag kommt.
const kiherkunft = "ki"

// alsWerte macht aus den Feldern eines Vorschlags die Werte des Bestands.
//
// **Sortiert** — eine Karte hat keine Reihenfolge, und ein Verlauf, dessen
// Zeilen bei jedem Lauf anders stehen, ist nicht zu vergleichen. **Seit
// A13.1b** steht sie hier und nicht in [kikarten.FelderSetzen]: die Gruppe
// braucht dieselbe Umrechnung, und zwei Fassungen liefen beim ersten
// Sonderfall auseinander.
func alsWerte(felder map[string][]string) []domain.Value {
	werte := make([]domain.Value, 0, len(felder))

	for name, einzelne := range felder {
		werte = append(werte, domain.Value{
			Field: name, Text: zusammengefasst(einzelne),
		})
	}

	slices.SortFunc(werte, func(einer, anderer domain.Value) int {
		return strings.Compare(einer.Field, anderer.Field)
	})

	return werte
}

// zusammengefasst macht aus mehreren Werten den einen Text des Bestands.
//
// **Mit Komma und Leerzeichen** — so schreibt MSE ein `multiple choice`, und
// der Bestand führt einen Feldwert als Text. Ein eigenes Trennzeichen hier
// machte den Wert für jede andere Fläche unlesbar.
func zusammengefasst(werte []string) string { return strings.Join(werte, ", ") }

// mitAbdruckDesGeschriebenen setzt den Abdruck auf **das, was im Feld landet**.
//
// # Warum er hier noch einmal entsteht
//
// **Weil `service` von den Teilen spricht und der Bestand einen Text hält.**
// Ein Mehrfachfeld kommt als `["a","b"]` und steht danach als `a, b` da; der
// Abdruck über die Teile passte nie zu dem, was jemand später im Feld liest.
//
// **Und dafür ist er da** (A11, #1450): er sagt, ob das Feld **noch** das
// trägt, was übernommen wurde. Ein Abdruck, der von Anfang an nicht passt,
// sagte bei jeder Karte „von Hand geändert" — und niemand glaubte ihm mehr.
func mitAbdruckDesGeschriebenen(
	herkunft []domain.KIFeldherkunft, werte []domain.Value,
) []domain.KIFeldherkunft {
	geschrieben := make(map[string]string, len(werte))
	for _, wert := range werte {
		geschrieben[wert.Field] = wert.Text
	}

	gestempelt := make([]domain.KIFeldherkunft, 0, len(herkunft))

	for _, einer := range herkunft {
		einer.Bezug = domain.KIInhaltsbezug(
			[]string{geschrieben[einer.Feld]})
		gestempelt = append(gestempelt, einer)
	}

	return gestempelt
}
