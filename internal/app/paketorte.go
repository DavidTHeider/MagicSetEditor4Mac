package app

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Wo die Pakete liegen** — eine Frage für sich, getrennt vom Lebenszyklus der
// Anwendung in `app.go`.

// mitVerwaltetemOrt stellt den verwalteten Ort vor die gefundenen Wurzeln.
//
// **Vorn**, denn die Reihenfolge ist die Aussage: was jemand installiert hat,
// gilt vor dem, was ein Bezug mitbrachte, und beides vor der Saat (#58).
// **Ohne Dopplung**, weil ein Ort, der Pakete trägt, ohnehin gefunden wird.
func mitVerwaltetemOrt(ort string, gefunden []string) []string {
	if strings.TrimSpace(ort) == "" {
		return gefunden
	}

	wurzeln := make([]string, 0, len(gefunden)+1)
	wurzeln = append(wurzeln, ort)

	for _, wurzel := range gefunden {
		if wurzel != ort {
			wurzeln = append(wurzeln, wurzel)
		}
	}

	return wurzeln
}

// abbrechen schließt die halb geöffnete Datenbank und gibt den Grund zurück.
//
// Ein Schließfehler wird angehängt, nicht verworfen: der eigentliche Befund ist
// der übergebene, aber eine Datenbank, die sich auch nicht schließen lässt, ist
// die zweite Hälfte der Nachricht.

// paketwurzelnFuer gibt die Wurzeln, die eine Naht über `pakete` durchsucht —
// **der verwaltete Ort immer, auch leer**, davor und danach, was
// [msepaket.Paketwurzeln] findet.
//
// **Eine Stelle für alle Nähte** — FMT-02, S10.NAT-03
// (#1691).
// Der Composition Root hielt die Regel seit dem Zweitblick über Phase 7; sechs
// Nähte daneben riefen `Paketwurzeln` ohne sie, und wer ihre Wurzeln beim
// Start in einem frischen Zuhause festlegte, sah jedes später angelegte Paket
// erst nach einem Neustart. Gemessen am nativen Fenster: die Werkstatt zeichnete
// eine eigene Kopie als Platzhalter, während das Set sie richtig zeigte.
//
// **Gemerkt und je Ruf geprüft** (PERF-02,
// #2169):
// die Nähte fragen je Ruf, und ein neues Paket gilt trotzdem beim nächsten —
// wann das Gemerkte gilt, steht bei [msepaket.Wurzelsuche].
func paketwurzelnFuer(pakete string) []string {
	return mitVerwaltetemOrt(pakete, wurzelsuche.Paketwurzeln(pakete, Vorlagenverzeichnis(), Saatverzeichnis()))
}

// wurzelsuche ist die Wurzelsuche dieses Prozesses.
var wurzelsuche = msepaket.NeueWurzelsuche()
