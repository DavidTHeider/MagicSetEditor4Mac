// **Die Scryfall-Bindungen reichen durch.**
//
// Eine eigene Datei, weil `bindungenlesen_test.go` mit ihnen über 400 Zeilen
// ginge (§2). Der Schnitt folgt der Frage: dort die Bindungen des Bestands,
// hier die zum fremden Dienst.

package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ohne eingerichteten Zugang ist jeder dieser Wege ein Befund** — und die
// Bindung reicht ihn durch, statt ihn zu verschlucken.
func TestDieScryfallbindungenReichenDurch(t *testing.T) {
	bindungen := bindungsprobe(t).Bindungen

	// **ScryfallBild ohne Zugang**: die Bindung reicht durch, und ohne
	// eingerichteten Zugang ist das dort ein Befund
	// (#1005).
	t.Run("ScryfallBild", func(t *testing.T) {
		if antwort := bindungen.ScryfallBild(t.Context(),
			"1", "kunst", "Prüfstein", "normal",
		); antwort.Fehler == nil {
			t.Error("der fehlende Zugang kam durch")
		}
	})

	// **ScryfallBildEinsetzen ohne Datei**: dasselbe eine Stufe später.
	t.Run("ScryfallBildEinsetzen", func(t *testing.T) {
		if antwort := bindungen.ScryfallBildEinsetzen(t.Context(),
			"1", "1", "kunst", "", "Wer auch immer", "https://scryfall.invalid/x",
			app.AuswahlDTO{Breite: 10, Hoehe: 10},
		); antwort.Fehler == nil {
			t.Error("der leere Pfad kam durch")
		}
	})

	// **ScryfallSatzEinlesen ohne Zugang** (#1033): dieselbe Gestalt.
	t.Run("ScryfallSatzEinlesen", func(t *testing.T) {
		if antwort := bindungen.ScryfallSatzEinlesen(t.Context(),
			"lea", "magic",
		); antwort.Fehler == nil {
			t.Error("der fehlende Zugang kam durch")
		}
	})

	// **ScryfallKarteAnlegen ohne Zugang** (#1032): dieselbe Gestalt.
	t.Run("ScryfallKarteAnlegen", func(t *testing.T) {
		if antwort := bindungen.ScryfallKarteAnlegen(t.Context(),
			"1", "Prüfstein", "", "",
		); antwort.Fehler == nil {
			t.Error("der fehlende Zugang kam durch")
		}
	})
}
