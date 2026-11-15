//go:build fremdpakete

// **Was Planbau und Rastern kosten, je Güte** —
// #1029,
// Werkstück Ω.5.
//
// # Warum getrennt und nicht als ein Weg
//
// `BenchmarkWegRastern` misst, was ein Tastendruck kostet — Plan **und**
// Rasterung, durch `Vorschau`, so wie das Fenster es tut. Das ist die Zahl,
// die jemand wartet, und sie bleibt die wichtigste.
//
// Sie sagt aber nicht, **wo** die Zeit liegt. Über zwanzig Karten gemessen
// trägt der Planbau 55 Prozent und die Rasterung 45; eine Güte fasst nur die
// zweite Hälfte an. Wer eine Beschleunigung gegen die erste Zahl belegt,
// belegt sie gegen eine Summe, in der die Hälfte gar nicht bewegt werden
// kann.
//
// # Wie sie gefahren werden
//
// Wie die anderen — dasselbe Kommando, dieselbe Liste in `docs/RELEASE.md`.
package app

import (
	"context"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Die Schwellen, und woran sie gemessen sind.
//
// Gemessen am 2026-08-25 auf einem Apple-Silicon-Rechner mit `-benchtime 5x`,
// über die **erste Karte** aus `Black.mse-set`:
//
//	Planbau    47 ms
//	fast       10 ms
//	medium     18 ms
//	high       20 ms
//	ultra      33 ms
//
// **Neu gemessen am 2026-10-15 mit den Schriften des Packs** (#2416, Mac17,3,
// je drei Läufe): bis dahin startete [messanwendung] ohne Schriftwurzel, und
// jede Karte trug „die Schrift Relay-Medium ist nicht zu finden" — gemessen
// war eine Karte ohne Text. Mit Schriften:
//
//	Planbau    22–23 ms
//	fast       13 ms        (Schwelle 15 ms — nah an ihr)
//	medium     20–21 ms
//	high       23 ms
//	ultra      36 ms
//
// Jede Schwelle hält; [keineFehlendeSchrift] hält jetzt, dass es so bleibt.
//
// # Warum das andere Zahlen sind als in Ω.4
//
// Dort stand `fast` bei 17,5 und `medium` bei 23,9 ms. **Beides stimmt, und
// beides misst etwas anderes**: dort zwanzig **verschiedene** Karten, hier
// dieselbe fünfmal. Der Ebenenspeicher hält Rahmen und Illustration nach dem
// ersten Lauf; wer dieselbe Karte wiederholt, bezahlt das Abtasten nicht noch
// einmal.
//
// **Und das ist hier die richtige Messung**, denn genau das tut das Fenster
// beim Tippen: dieselbe Karte, wieder und wieder. Die Zahl aus Ω.4 gilt für
// einen Stapellauf über ein Set.
//
// # Wie eng sie liegen, und warum
//
// **Jede Schwelle liegt unter den Kosten der nächstlangsameren Güte.** Das ist
// die Rückschrittsklasse, gegen die es sie gibt: eine Güte, die still die
// Abtastung ihrer Nachbarin nimmt, ist keine eigene mehr. Der erste Anlauf
// setzte `fast` auf 30 ms — sie hätte den Rückfall auf `medium` (18 ms) nicht
// gesehen und wäre damit ein Wächter über nichts gewesen.
//
// **Was sie nicht halten**: `medium` bei 26 ms fängt einen Rückfall auf
// `ultra` (33 ms), **nicht** einen auf `high` (20 ms) — dafür läge die
// Schwelle bei 19 ms und damit 5 Prozent über dem Messwert, was unter Last
// flackerte. Dieselbe Lücke hat `high` gegenüber `medium`. Enger geht es nur
// mit einem Wächter, der öfter grundlos rot ist als er einen Fehler findet.
//
// **Und der Warmlauf verdeckt einen Teil des Hebels.** Der Lauf davor füllt
// den Ebenenspeicher; was danach gemessen wird, ist die Abtastung, die
// **nicht** gespeichert wird — Masken und Symbole —, nicht die der Rahmen- und
// Illustrationsebene. Angestrichen vom Zweitblick über Phase Ω.
//
// **Gemessen bleibt der Rückfall trotzdem sichtbar**: entschärft man
// `GueteSchnell` zu bilinear, misst dieser Lauf 20 ms gegen die Schwelle von
// 15 und wird rot. Der ungespeicherte Rest reicht also — aber er ist der
// Grund, und nicht die 78 Prozent, die weiter oben stehen.
const (
	// schwellePlanbau gilt für **einen** Kartenplan.
	//
	// **Ohne Güte**: der Plan ist für alle vier derselbe, und genau das ist
	// die Zusage des Werkstücks. Eine Güte, die ihn bewegte, änderte den
	// Zeilenumbruch.
	schwellePlanbau = 68 * time.Millisecond

	// schwelleRasternSchnell gilt für `fast` — unter `medium`.
	schwelleRasternSchnell = 15 * time.Millisecond

	// schwelleRasternMittel gilt für `medium` — unter `ultra`.
	schwelleRasternMittel = 26 * time.Millisecond

	// schwelleRasternHoch gilt für `high` — unter `ultra`.
	schwelleRasternHoch = 29 * time.Millisecond

	// schwelleRasternHoechst gilt für `ultra`.
	schwelleRasternHoechst = 47 * time.Millisecond
)

// schwelleJeGuete ordnet jeder Güte ihre Schwelle zu.
//
// **Eine Abbildung und keine Reihung.** Ein Feld, das nach `int(guete)`
// indiziert, ordnet sich um, sobald jemand eine Konstante verschiebt — und die
// Schwelle von `fast` läge dann bei `ultra`.
var schwelleJeGuete = map[render.Guete]time.Duration{
	render.GueteSchnell: schwelleRasternSchnell,
	render.GueteMittel:  schwelleRasternMittel,
	render.GueteHoch:    schwelleRasternHoch,
	render.GueteHoechst: schwelleRasternHoechst,
}

// **Einen Kartenplan bauen.**
func BenchmarkWegPlanbau(b *testing.B) {
	anwendung, karteID := messkarte(b)

	planZurKarte(b, anwendung, karteID)

	for b.Loop() {
		plan := planZurKarte(b, anwendung, karteID)

		// **Und es sind Ebenen drin.** Ein Planbauer, der einen leeren Plan
		// gibt, ist der schnellste — und ohne diese Zeile der beste Messwert.
		if len(plan.Ebenen) == 0 {
			b.Fatal("ein Plan ohne Ebenen")
		}
	}

	haelt(b, "einen Kartenplan bauen", schwellePlanbau)
}

// **Eine Karte rastern, je Güte.**
func BenchmarkWegRasternJeGuete(b *testing.B) {
	anwendung, karteID := messkarte(b)

	rasterer, ok := anwendung.rasterer.(*render.Rasterer)
	if !ok {
		b.Skip("kein Rasterer")
	}

	plan := planZurKarte(b, anwendung, karteID)

	for _, guete := range render.Gueten {
		b.Run(guete.Name(), func(b *testing.B) {
			// **Einmal vorab, je Güte.** Der Ebenenspeicher hält je Güte
			// eigene Bilder (der Schlüssel trägt sie); ohne diesen Lauf
			// bezahlte die erste Wiederholung das Abtasten für alle weiteren.
			if _, _, err := rasterer.Zeichne(
				context.Background(), plan, 1, guete); err != nil {
				b.Fatalf("zeichnen: %v", err)
			}

			for b.Loop() {
				bild, _, err := rasterer.Zeichne(context.Background(), plan, 1, guete)
				if err != nil {
					b.Fatalf("zeichnen: %v", err)
				}

				if bild == nil || bild.Bounds().Empty() {
					b.Fatal("kein Bild")
				}
			}

			haelt(b, "eine Karte rastern ("+guete.Name()+")", schwelleJeGuete[guete])
		})
	}
}

// messkarte gibt eine Anwendung mit eingefahrenem Set und eine Kartenkennung.
func messkarte(b *testing.B) (anwendung *App, karteID string) {
	b.Helper()

	anwendung = messanwendung(b)
	setID := eingefahren(b, anwendung, "Black.mse-set")

	zeilen := anwendung.Kartenzeilen(context.Background(), setID)
	if zeilen.Fehler != nil {
		b.Fatalf("Kartenzeilen: %+v", zeilen.Fehler)
	}

	if len(zeilen.Zeilen) == 0 {
		b.Fatal("ein Set ohne Karten")
	}

	// **Mit allen Schriften** (#2416): gemessen wird die Karte, die ein Mensch
	// sieht, nicht eine ohne Text.
	vorschau := anwendung.Vorschau(context.Background(), zeilen.Zeilen[0].ID, 1)
	if vorschau.Fehler != nil {
		b.Fatalf("Vorschau: %+v", vorschau.Fehler)
	}

	keineFehlendeSchrift(b, vorschau.Vorschau.Befunde)

	return anwendung, zeilen.Zeilen[0].ID
}
