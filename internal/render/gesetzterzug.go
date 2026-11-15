package render

// **Der fertige Umbruch einer Textebene, für Messläufe** —
// #1029,
// Werkstück Ω.3.
//
// # Wozu
//
// Der Vergleich mit MSE2 zeigt, **dass** einzelne Zeilenabstände einen Punkt
// zu groß sind — gemessen an den Tintenbändern in beiden Bildern
// (`app.TestDieZeilenlageGegenMSE2`). Welche Zeile es trifft und woher der
// Punkt kommt, sagt erst der **ungerundete** Satz.
//
// Der steht heute nirgends: `Ebene` trägt `Abschnitte` und `Satz`, also die
// Eingabe, und das Ergebnis lebt und stirbt in [Rasterer.text].
//
// # Warum am Rasterer und nicht als freie Funktion
//
// [SetzeAbschnitte] ist ausgeführt und ließe sich rufen — aber nur mit dem
// **Schriftsatz**, und den baut der Rasterer aus seiner Schriftsuche. Wer ihn
// von außen nachbaute, bekäme einen anderen: dieselbe Klasse wie beim
// Herkunftslauf, wo ein eigener Rasterer weder Bildquelle noch Schriftsuche
// hätte und die Messung auf ein Bild träfe, das so nie entsteht.
//
// **Es gibt dafür einen Vorgänger**: [Rasterer.ZeichneMitHerkunft] ist ebenso
// ein Weg, den nur ein Messlauf geht.

import (
	"fmt"
	"strings"
)

// GesetzterZug gibt den fertigen Umbruch einer Textebene.
//
// **Derselbe Weg wie beim Zeichnen** — Schriftsuche, Schnitte, Bildquelle —,
// damit die Zahl zu dem Bild gehört, das der Rasterer malt.
func (r *Rasterer) GesetzterZug(ebene Ebene) (Satzergebnis, error) {
	if ebene.Art != ArtText {
		return Satzergebnis{}, fmt.Errorf(
			"%w: die Ebene malt %s und keinen Text", ErrPlan, ebene.Art)
	}

	if r.schriften == nil {
		return Satzergebnis{}, fmt.Errorf(
			"%w: für Text ist keine Schriftsuche eingerichtet", ErrPlan)
	}

	name := ebene.Satz.Schrift
	if strings.TrimSpace(name) == "" {
		return Satzergebnis{}, fmt.Errorf(
			"%w: eine Textebene ohne Schriftnamen", ErrPlan)
	}

	schrift, err := r.schriften.Suche(name, false)
	if err != nil {
		return Satzergebnis{}, fmt.Errorf(
			"%w: die Schrift %s ist nicht zu finden: %w", ErrPlan, name, err)
	}

	schriften := r.schnitte(schrift, name, strings.TrimSpace(ebene.Satz.Kursivschrift))

	return SetzeAbschnitte(ebene.Satz, ebene.Rahmen, ebene.Abschnitte,
		schriften, r.quelle), nil
}
