package flaechen_test

// **Die Handelsrichtlinie sperrt jede Fläche gleich** — #2435.
//
// Gefunden bei der Erreichbarkeitsinventur von SEC04 (#2357): die
// Einstellungsseite sagte zu `richtlinie.handel`: „Schaltet Teilen,
// Deckleihe, Tausch und Markt zugleich ab", und kein Dienst fragte die
// Schranke. Hier steht sie auf `aus`, und ein Tausch wird an Fenster,
// Kommandozeile und Dienst versucht — jede Fläche weist ihn als
// Validierung ab und nennt den Grund. MCP eröffnet keinen Handel
// (`FAEHIGKEITEN.md`, 0106).
//
// **Die Gegenprobe ist derselbe Weg ohne die Schranke**: am Ende geht der
// Tausch durch, sobald `richtlinie.handel` wieder `an` steht — sonst hieße
// das Rot oben nur, dass irgendetwas abgewiesen wird.

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

func TestDieHandelsrichtlinieSperrtJedeFlaeche(t *testing.T) {
	umgebung := append(mitEigenerSaat(t), schluesselbund.Bundort+"="+probebund(t))
	mussKommandozeile(t, umgebung, "peer-anlegen")
	mussKommandozeile(t, umgebung, "einstellungen", "setzen", "richtlinie.handel", "aus")

	an := starteDienst(t, umgebung)

	var positionen []app.HandelspositionDTO
	if fehler := json.Unmarshal([]byte(diePositionen), &positionen); fehler != nil {
		t.Fatal(fehler)
	}

	abweisungGleich(t, "ein Tausch, während das Profil nicht handelt", abweisung{
		klasse: klasseValidierung,
		kern:   "dieses Profil handelt nicht",
		cli: func() (string, int) {
			aus, fehler, code := mitEingabeUndCode(t, umgebung, diePositionen,
				"handel", "eroeffnen", "", "tausch", dasGegenueberImHandel)

			return aus + fehler, code
		},
		ohneMCP: "MCP eröffnet keinen Handel — eine Zusage gegenüber einem anderen Menschen gibt kein Modell in fremdem Namen (0106)",
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/handel",
				`{"art":"tausch","gegenueber":["`+dasGegenueberImHandel+`"],"positionen":`+diePositionen+`}`)
		},
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).HandelEroeffnen(t.Context(), app.Handelswunsch{
				Art: "tausch", Gegenueber: []string{dasGegenueberImHandel}, Positionen: positionen,
			}).Fehler
		},
	})

	mussKommandozeile(t, umgebung, "einstellungen", "setzen", "richtlinie.handel", "an")
	mitEingabe(t, umgebung, diePositionen, "handel", "eroeffnen", "", "tausch", dasGegenueberImHandel)
}
