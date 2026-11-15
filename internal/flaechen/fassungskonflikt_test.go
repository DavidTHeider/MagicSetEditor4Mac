// **Eine veraltete Fassung wird an jeder Fläche gleich abgewiesen** — C04
// (#2229).
//
// Die Fassung ist die Zusage, dass niemand einen Stand überschreibt, den er
// nicht gesehen hat (§7, Regel 2). Bis C04 prüften das vier Läufe, jeder an
// höchstens zwei Flächen, und einer nur an der Kommandozeile, obwohl sein
// Name „überall" sagte. Hier geht derselbe Fall über **jede** Fläche, die
// eine Fassung entgegennimmt, und danach steht der Stand noch da.
//
// **Die Kommandozeile nimmt keine Fassung entgegen** — wer tippt, hat den
// vorigen Stand nicht zur Hand; sie liest ihn selbst. Das ist begründet und
// steht so an jedem Fall.
package flaechen_test

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

const kommandozeileLiestSelbst = "die Kommandozeile nimmt keine Fassung entgegen — sie liest den Stand selbst"

func TestEinExemplarMitVeralteterFassungWirdAnJederFlaecheAbgewiesen(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	const druck = "00000000-0000-4000-8000-00000000c041"

	mussKommandozeile(t, umgebung, "exemplar", druck, "1", "played", "de", "nonfoil")
	alt := exemplarFuer(t, umgebung, druck)

	// **Einmal geändert**, damit die gelesene Fassung veraltet ist.
	mussKommandozeile(t, umgebung, "exemplar-aendern", strconv.FormatInt(alt.ID, 10), "2")

	an := starteDienst(t, umgebung)
	rumpf := map[string]any{
		"id": alt.ID, "fassung": alt.Fassung, "printing": druck, "menge": 7,
		"zustand": "played", "sprache": "de", "finish": "nonfoil",
	}

	abweisungGleich(t, "ein Exemplar mit veralteter Fassung", abweisung{
		klasse:            klasseKonflikt,
		kern:              "Exemplar",
		ohneKommandozeile: kommandozeileLiestSelbst,
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.exemplaraendern", rumpf)
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPut, "/sammlung/"+strconv.FormatInt(alt.ID, 10), string(mussJSON(t, rumpf)))
		},
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).ExemplarAendern(t.Context(), app.ExemplarDTO{
				ID: alt.ID, Fassung: alt.Fassung, Printing: druck, Menge: 7,
				Zustand: "played", Sprache: "de", Finish: "nonfoil",
			})
		},
	})

	// **Und danach steht die Änderung von vorhin**, nicht die abgewiesene.
	if danach := exemplarFuer(t, umgebung, druck); danach.Menge != 2 {
		t.Errorf("die Menge ist %d — eine abgewiesene Änderung hat geschrieben", danach.Menge)
	}
}

// **Ein Deck nimmt die Fassung an allen vier Flächen** — auch an der
// Kommandozeile (`deck umbenennen <deck> <fassung> …`).
func TestEinDeckMitVeralteterFassungWirdAnJederFlaecheAbgewiesen(t *testing.T) {
	umgebung := deckumgebung(t)

	deckID := strings.TrimSpace(mussKommandozeile(t, umgebung, "deck", "neu", "Aggro", "magic"))

	// **Ein neues Deck hat die Fassung 1**; einmal umbenannt hat es die 2.
	mussKommandozeile(t, umgebung, "deck", "umbenennen", deckID, "1", "Midrange", "")

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "ein Deck mit veralteter Fassung", abweisung{
		klasse: klasseKonflikt,
		kern:   "Deck",
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung, "deck", "umbenennen", deckID, "1", "Kontrolle", "")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.deck.umbenennen",
				map[string]any{"deckID": zahl(t, deckID), "fassung": 1, "name": "Kontrolle", "notiz": ""})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPut, "/decks/"+deckID, `{"fassung":1,"name":"Kontrolle","notiz":""}`)
		},
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).DeckUmbenennen(t.Context(), deckID, "1", "Kontrolle", "").Fehler
		},
	})

	if aus := mussKommandozeile(t, umgebung, "deck", deckID); !strings.Contains(aus, "Midrange") || strings.Contains(aus, "Kontrolle") {
		t.Errorf("eine abgewiesene Umbenennung hat geschrieben:\n%s", aus)
	}
}
