package service

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Die gerechneten Vorgabewerte einer Karte —
// #563.
//
// # Der Befund
//
// 142 der 249 Felder von `magic.mse-game` tragen ein `default:`, und alle 142
// sind Skript. Die **gezeichnete** Karte trägt sie — der Kanon steht bei null
// Befunden. Der **Editor** zeigte an derselben Stelle ein leeres Feld: wer die
// Karte ansah, sah eine Farbe; wer das Feld ansah, sah nichts.
//
// # Warum hier und nicht in einem zweiten Lauf
//
// Die Werte stehen beim Planbau fertig da: `umgebungAus` gibt sie als
// `felder` zurück, und `Stilplan` hält sie ohnehin in der Hand. Ein zweiter
// Lauf rechnete dasselbe noch einmal — mit demselben Einwand, den
// [render.Plan.Feldflaechen] schon nennt.

// vorgabenAus sammelt die Werte, die **nicht** von der Karte kommen.
//
// **Der Vergleich läuft über den Namen mit Unterstrichen.** In der Umgebung
// heißt ein Feld `card_symbol`, im Paket und im Editor `card symbol`; wer
// beide Formen mischt, findet nichts und meldet jedes Feld als Vorgabe.
func vorgabenAus(
	gerechnet []msescript.Eintrag, felder []domain.Field, karte *domain.CardRow,
) []render.Vorgabe {
	// **Was die Karte selbst führt, ist keine Herleitung.** Ein Hinweis „das
	// ist gerechnet" über einem getippten Wert wäre eine Falschaussage.
	//
	// **„Führt" heißt: mit Inhalt.** Der erste Anlauf zählte jede Zeile — und
	// wer ein Feld leert, hat eine Zeile mit leerem Text. Die **gezeichnete**
	// Karte trug den Vorgabewert weiter (`mitVorgabeskripten` füllt jedes
	// leere Feld), der Editor zeigte ab da nichts mehr: zwei Antworten auf
	// dieselbe Frage, und die im Editor war die falsche.
	//
	// Damit ist zugleich der dritte Abnahmekasten aus #563 eingelöst — „wer
	// ihn leert, bekommt den gerechneten zurück" —, und zwar an **einer**
	// Stelle statt an zweien.
	eigen := make(map[string]bool, len(karte.Values))

	for i := range karte.Values {
		if strings.TrimSpace(karte.Values[i].Text) == "" {
			continue
		}

		eigen[umgebungsname(karte.Values[i].Field)] = true
	}

	// **Die Reihenfolge ist die der Felder des Spiels**, nicht die der
	// Umgebung: der Editor zeigt sie in dieser Ordnung, und eine Antwort, die
	// bei jedem Aufruf anders sortiert ist, lässt die Seite springen.
	stand := make(map[string]string, len(gerechnet))
	for i := range gerechnet {
		stand[gerechnet[i].Name] = gerechnet[i].Wert.AlsText()
	}

	vorgaben := make([]render.Vorgabe, 0, len(felder))

	for i := range felder {
		name := umgebungsname(felder[i].Name)

		if eigen[name] {
			continue
		}

		wert := stand[name]

		// **Ein leerer Wert ist keine Vorgabe.** „Hier steht gerechnet nichts"
		// ist keine Herleitung.
		if wert == "" {
			continue
		}

		vorgaben = append(vorgaben,
			render.Vorgabe{Feld: felder[i].Name, Wert: wert})
	}

	return vorgaben
}

// umgebungsname macht aus `card symbol` das `card_symbol` der Skriptumgebung.
func umgebungsname(feld string) string {
	return strings.ReplaceAll(feld, " ", "_")
}
