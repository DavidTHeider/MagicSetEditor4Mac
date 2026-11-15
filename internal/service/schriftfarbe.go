package service

// Die Schriftfarbe eines Feldes — und warum sie gerechnet wird.
//
// # Der Befund, der diese Datei ausgelöst hat
//
// Bis #382 las
// `feldebene` sie als **Literal**:
//
//	Farbe: farbeAus(feld.Schrift.Farbe)
//
// In der Vorlage steht dort aber im Regelfall ein Skript. Ausgezählt über die
// 190 Stile von `M15-Magic-Pack`, über die 5.950 Feldstile ihrer
// `card style:`-Blöcke: **3.011 nennen eine `font: color:`, und 2.321 davon
// sind Skripte** — sie alle wurden schwarz.
//
// **Die Reichweite dieser Zahl** (§10): gezählt am **Feldstil**, nicht an
// gezeichneten Ebenen. Wie viele Ebenen einer Karte betroffen sind, hängt an
// `visible:` und daran, welche Felder sie führt.
//
// # Woran es aufgefallen ist
//
// Am Fuß der Karte. Illustrator, Copyright und Setcode stehen in
// `magic-modules.mse-include/information/card_fields` mit
//
//	color: { styling.copyright_text_color or else set.copyright_text_color }
//
// und `magic-m15.mse-style` bindet `information/styling_fields` **nicht** ein.
// Es gilt also das Setfeld — ein Farbfeld mit den Wahlfarben `rgb(255,255,255)`
// und `rgb(0,0,0)`, dessen erste Wahl nach unserer eigenen Regel Weiß ist. Auf
// dem Bild stand schwarze Schrift auf dem schwarzen Kartenrand
// (#375).

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// schwarz ist die Farbe, die bleibt, wenn keine zu ermitteln war.
//
// **MSEs Vorgabe**, und deshalb auch die Antwort auf einen Fehlschlag: ein Text
// ohne Farbe wäre gar kein Text.
var schwarz = render.Farbe{A: 255}

// farbangabe wertet eine Farbangabe aus, die ein Literal **oder** ein Skript ist.
//
// Die Reihenfolge ist die von [masszahl]: erst das Literal, dann die Rechnung.
// Ein Stil, der `white` schreibt, soll nicht durch den Auswerter laufen — und
// `rgb(255,255,255)` ist als Literal **und** als Ausdruck lesbar, was ohne diese
// Reihenfolge zwei Wege zu einer Farbe wären.
func farbangabe(
	ctx context.Context, feld, angabe, quelle string, umgebung *msescript.Umgebung,
) (render.Farbe, []domain.Befund) {
	if farbe, ist := gedeuteteFarbe(quelle); ist {
		return farbe, nil
	}

	// **Die leere Angabe ist keine Lücke.** Ein Feld ohne `color:` zeichnet
	// schwarz; das ist MSEs Vorgabe und kein Befund.
	roh := strings.TrimSpace(quelle)

	rumpf, geklammert := skriptrumpf(roh)
	if !geklammert {
		if roh == "" {
			return schwarz, nil
		}

		// **Ein Text, der weder Farbe noch Skript ist, wird benannt.** MSEs
		// `parse_color` warnt an dieser Stelle und nimmt `Color()`; schweigend
		// schwarz zu zeichnen hieße, einen unlesbaren Text als Absicht
		// auszugeben (§8).
		return schwarz, []domain.Befund{{
			Ort:     feld + " → " + angabe,
			Meldung: "die Angabe ist keine Farbe: " + quelle,
		}}
	}

	return gerechneteFarbe(ctx, feld, angabe, rumpf, umgebung)
}

// gerechneteFarbe wertet den Rumpf aus und deutet, was herauskommt.
func gerechneteFarbe(
	ctx context.Context, feld, angabe, rumpf string, umgebung *msescript.Umgebung,
) (render.Farbe, []domain.Befund) {
	wert, err := msescript.AuswertenQuelle(ctx, rumpf, umgebung)
	if err != nil {
		return schwarz, []domain.Befund{{
			Ort:     feld + " → " + angabe,
			Meldung: "die Angabe ließ sich nicht ausrechnen: " + err.Error(),
		}}
	}

	// **Ein Farbwert ist der Normalfall.** `rgb(…)`, ein Farbfeld aus
	// `styling` oder `set` — sie alle kommen als Farbe zurück und nicht als
	// Text, der noch einmal zu deuten wäre.
	if farbe, ist := wert.AlsFarbe(); ist {
		return render.Farbe{R: farbe.R, G: farbe.G, B: farbe.B, A: farbe.A}, nil
	}

	// **Ein Fehlerwert nennt seinen eigenen Grund** — dieselbe Regel wie bei
	// [masszahl]: er sagt, *welcher* Zugriff ins Leere ging.
	if wert.IstFehler() {
		return schwarz, []domain.Befund{{
			Ort:     feld + " → " + angabe,
			Meldung: "die Angabe ergab keine Farbe: " + wert.Text,
		}}
	}

	if farbe, ist := gedeuteteFarbe(wert.AlsText()); ist {
		return farbe, nil
	}

	// **Nichts ist keine Farbe, aber auch kein Fehler.** `{ if … then rgb(…) }`
	// ohne `else` ergibt bei MSE `nil`, und die Vorgabe greift.
	if wert.Art == msescript.ArtNichts {
		return schwarz, nil
	}

	return schwarz, []domain.Befund{{
		Ort:     feld + " → " + angabe,
		Meldung: "die Angabe ergab keine Farbe, sondern " + wert.AlsText(),
	}}
}

// textangabe wertet eine Angabe aus, die ein Text **oder** ein Skript ist.
//
// # Wofür
//
// Für den **Schriftnamen**. Über die 190 Stile von `M15-Magic-Pack` sind
// **664 von 3.119** `font: name:` Skripte, dazu **439 von 581** `italic name:`
// — der rohe Text ging bis
// #393 an die
// Schriftsuche, und die suchte nach einer Schrift namens `{ body_font() }`.
//
// **Die Reichweite der Zahlen** (§10): gezählt am Feldstil, nicht an
// gezeichneten Ebenen.
//
// # Warum der rohe Name bei einem Fehlschlag stehen bleibt
//
// Damit die Meldung der Schriftsuche sagt, **wonach** sie gesucht hat. Ein
// leerer Name führte zu „keine Schrift eingerichtet" — einer Meldung über einen
// anderen Mangel.
func textangabe(
	ctx context.Context, feld, angabe, quelle string, umgebung *msescript.Umgebung,
) (string, []domain.Befund) {
	roh := strings.TrimSpace(quelle)

	rumpf, geklammert := skriptrumpf(roh)
	if !geklammert {
		return roh, nil
	}

	wert, err := msescript.AuswertenQuelle(ctx, rumpf, umgebung)
	if err != nil {
		return roh, []domain.Befund{{
			Ort:     feld + " → " + angabe,
			Meldung: "die Angabe ließ sich nicht ausrechnen: " + err.Error(),
		}}
	}

	if wert.IstFehler() {
		return roh, []domain.Befund{{
			Ort:     feld + " → " + angabe,
			Meldung: "die Angabe ergab keinen Namen: " + wert.Text,
		}}
	}

	// **Nichts ist kein Name, aber auch kein Fehler.** `{ if … then "X" }` ohne
	// `else` ergibt bei MSE `nil`, und dann gilt keine Schrift — dieselbe Regel
	// wie bei der Farbe.
	if wert.Art == msescript.ArtNichts {
		return "", nil
	}

	return strings.TrimSpace(wert.AlsText()), nil
}
