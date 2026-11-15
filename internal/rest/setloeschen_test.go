// **Ein Set löschen über die Rückschleife** —
// #1218.
//
// Der Weg selbst steht in `alleEndpunkte` und wird von den Wächtern über
// **jeden** Endpunkt mitgefahren — Marke, Herkunft, Kennung, Fehlerweg. Was
// hier steht, sind die zwei Dinge, die nur dieser Weg hat: der **Name als
// Abfragewert** und die **Befunde im Rumpf**.
package rest_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// namensmerker hält fest, mit welchem Namen gelöscht werden sollte.
type namensmerker struct {
	williger

	name    string
	befunde []string
}

func (n *namensmerker) SetLoeschen(_ context.Context, _ int64, name string) (
	[]string, error,
) {
	n.name = name

	return n.befunde, nil
}

// **Der Name kommt aus der Abfrage.**
//
// **Und ein fehlender wird durchgereicht, nicht hier abgewiesen.** Der Bestand
// vergleicht ihn mit dem heutigen; dass ein leerer nie passt, ist seine
// Antwort. Eine zweite Prüfung hier gäbe eine zweite Meldung für dieselbe
// Sache — und zwei Meldungen für einen Fehler sind eine zu viel (§8).
func TestDasSetloeschenLiestDenNamenAusDerAbfrage(t *testing.T) {
	for name, fall := range map[string]struct {
		weg      string
		erwartet string
	}{
		"ein Name steht dabei":  {"/sets/1?name=Mein+Set", "Mein Set"},
		"kein Name steht dabei": {"/sets/1", ""},
		"ein leerer Name":       {"/sets/1?name=", ""},
		"ein Name mit Umlauten": {"/sets/1?name=Gr%C3%BCn", "Grün"},
	} {
		t.Run(name, func(t *testing.T) {
			merker := &namensmerker{}
			dienst := dienstMitLoescher(t, merker)

			antwort := ruf(t, dienst, endpunkt{
				methode: http.MethodDelete, weg: fall.weg, erfolg: http.StatusOK,
			}, gueltig)

			if antwort.Code != http.StatusOK {
				t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
			}

			if merker.name != fall.erwartet {
				t.Errorf("gelöscht werden sollte %q, erwartet war %q",
					merker.name, fall.erwartet)
			}
		})
	}
}

// **Eine unbrauchbare Kennung wird abgewiesen, und nichts wird gelöscht.**
func TestDasSetloeschenWeistEineUnbrauchbareKennungAb(t *testing.T) {
	t.Parallel()

	merker := &namensmerker{}

	antwort := ruf(t, dienstMitLoescher(t, merker), endpunkt{
		methode: http.MethodDelete, weg: "/sets/x?name=Mein+Set",
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Errorf("endete mit %d statt 400: %s", antwort.Code, antwort.Body.String())
	}

	if merker.name != "" {
		t.Error("die Löschnaht wurde trotzdem gerufen")
	}
}

// **Die Befunde stehen im Rumpf, und eine leere Liste ist keine `null`.**
//
// Ein Aufrufer, der über die Länge geht, stolperte sonst über den Unterschied
// zwischen „keine Befunde" und „das Feld gibt es nicht".
func TestDasSetloeschenGibtDieBefundeImRumpf(t *testing.T) {
	for name, fall := range map[string]struct {
		befunde  []string
		erwartet string
	}{
		"keine Befunde": {nil, `"befunde":[]`},
		"ein Befund":    {[]string{"Reste"}, `"befunde":["Reste"]`},
	} {
		t.Run(name, func(t *testing.T) {
			dienst := dienstMitLoescher(t, &namensmerker{befunde: fall.befunde})

			antwort := ruf(t, dienst, endpunkt{
				methode: http.MethodDelete, weg: "/sets/1?name=x", erfolg: http.StatusOK,
			}, gueltig)

			if !strings.Contains(antwort.Body.String(), fall.erwartet) {
				t.Errorf("der Rumpf führt %s nicht: %s",
					fall.erwartet, antwort.Body.String())
			}
		})
	}
}

// dienstMitLoescher baut einen Dienst, dessen Löschnaht die genannte ist.
//
// **Eigens und nicht über `neuerDienst`**: der ganze Zweck ist, ein Double
// einzusetzen, das mitschreibt — jenes nimmt für jede Naht dasselbe willige.
func dienstMitLoescher(t *testing.T, loescher rest.Setloescher) *rest.Dienst {
	t.Helper()

	jetzt := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)

	return rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{
			Leser:            williger{},
			Schreiber:        williger{},
			Ausfuhr:          williger{},
			Abgleich:         williger{},
			Bilder:           williger{},
			Auskunft:         williger{},
			Suche:            williger{},
			Katalog:          williger{},
			Grenzen:          williger{},
			Verweise:         williger{},
			Marken:           williger{},
			Ablage:           williger{},
			Ersetzung:        williger{},
			Liste:            williger{},
			Sammeldaten:      williger{},
			Anlage:           williger{},
			Saetze:           williger{},
			Konsole:          williger{},
			Werkstatt:        williger{},
			Sammlung:         williger{},
			Listeneinfuhr:    williger{},
			Deckeinfuhr:      williger{},
			Sammlungsausfuhr: williger{},
			Sammlungssichten: williger{},
			Kartenskripte:    williger{},
			Auszaehlung:      williger{},
			Parameterhilfen:  williger{},
			Autoersetzungen:  williger{},
			Decks:            williger{},
			Turniere:         williger{},
			Symbolbibliothek: williger{},
			Sprachpakete:     williger{},
			Symbol:           williger{},
			Masse:            williger{},
			Loescher:         loescher,
			Preise:           williger{},
			Wortwahl:         williger{},
			Woerter:          williger{},
			Pack:             williger{},
			Angaben:          williger{},
			Stile:            williger{},
			Zusatz:           williger{},
			Pakete:           williger{},
			Kartenbilder:     williger{},
			Vorlagen:         williger{},
			Probelauf:        williger{},
			Zahlen:           williger{},
			Verlauf:          williger{},
			Einstellungen:    williger{},
			Fassungen:        williger{},
			Anleger:          williger{},
		})
}
