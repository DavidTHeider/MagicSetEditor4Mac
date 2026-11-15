package mcpnaehte

// **Die KI-Naht des MCP-Servers reicht ihre Fehler durch** — angestrichen
// beim Bau von A13.4
// (#1452).
//
// # Was hier gemessen wird
//
// **Dass eine Naht über einem leeren Bestand einen Fehler gibt und keine
// leere Sicht.** Für ein Sprachmodell ist der Unterschied alles: „es gibt
// nichts" liest es als Antwort, „die Frage ging schief" als Aufforderung,
// etwas anderes zu tun.
//
// **Und was er nicht deckt**: den Erfolgsfall. Der braucht ein Modell, und
// was er belegte, steht in `internal/flaechen` und hinter `netzlauf`.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Jeder KI-Weg des Servers meldet sich über einem leeren Bestand.**
//
// **Tabellengetrieben**, weil vier Wege dieselbe Zusage haben — und eine
// Schleife über sie hält sie auch für den fünften, der dazukommt.
func TestJedeKINahtDesServersMeldetSichAufLeeremBestand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	naht := &preise{db: frischerBestand(t)}

	wege := map[string]func() (any, error){
		"schreiben.ki.vorschlag": func() (any, error) {
			return naht.KIVorschlag(probekontext(t), mcp.KIVorschlagsbezug{
				KarteID: 4, Anbieter: "ollama", Alias: "gemma4:8b",
				Auftrag: "Schreibe.",
			})
		},
		"schreiben.ki.einschaetzung": func() (any, error) {
			return naht.KIEinschaetzung(probekontext(t),
				mcp.KIEinschaetzungsbezug{
					KarteID: 4, Art: "regel", Anbieter: "ollama",
					Alias: "gemma4:8b", Frage: "Wie blockt sie?",
				})
		},
		"schreiben.ki.erkennung": func() (any, error) {
			return naht.KIErkennung(probekontext(t), mcp.KIErkennungsbezug{
				Pfad: "/gibtesnicht.png", Anbieter: "ollama",
				Alias: "gemma4:8b",
			})
		},
	}

	for name, fahren := range wege {
		t.Run(name, func(t *testing.T) {
			if _, fehler := fahren(); fehler == nil {
				t.Error("über einem leeren Bestand kam kein Fehler")
			}
		})
	}
}

// **Und der Vorschlag reicht seine Belegfrage weiter.**
//
// **Ohne diesen Lauf sagte der obige nur, dass etwas scheitert** — nicht,
// dass die zwei Felder aus A13.4 an der Anwendung ankommen. Ein Feld, das
// die Naht fallen ließe, sähe an der Absage genauso aus.
func TestDerVorschlagDesServersTraegtSeineBelegfrage(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	naht := &preise{db: frischerBestand(t)}

	if _, fehler := naht.KIVorschlag(probekontext(t), mcp.KIVorschlagsbezug{
		KarteID: 4, Anbieter: "ollama", Alias: "gemma4:8b",
		Auftrag: "Schreibe.", Belegfrage: "trampelschaden",
		Freigegeben: []string{"eigenekarte"},
	}); fehler == nil {
		t.Error("über einem leeren Bestand kam kein Fehler")
	}
}
