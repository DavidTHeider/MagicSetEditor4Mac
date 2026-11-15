// **Die Rollenmatrix über die Flächen** — ID-03.AK-01
// (#1702).
//
// # Was sie fährt
//
// **Rolle × Eigentümer/Fremdprofil × Fläche × Wirkung**, wie das Issue es
// verlangt: vier Lagen — das eigene Profil, ein fremdes als Mitarbeiter, als
// Leser und **ohne** Rolle (ein Eigner ohne Faktor lässt jeden herein) — über
// Kommandozeile, MCP und Dienst, je mit vier Wirkungen: lesen, im Bestand
// schreiben, eine Datei schreiben, etwas hinausschicken.
//
// **Geprüft an der Wirkung** (AK-04), nicht am Fehlertext: ob das Deck
// entstand, ob die Datei daliegt, ob die Gegenstelle gefragt wurde. Und jedes
// Nein hat sein Ja daneben — der Mitarbeiter und das eigene Profil fahren
// dieselben Wege und kommen durch, sonst hieße ein Nein nur, dass der Weg
// nicht ging (§10).
//
// # Was sie nicht fährt
//
//   - **Das Fenster**: es fährt dieselben Lagen in `internal/app`
//     (`ohnerolle_test.go`, `leserausfuhr_test.go`, `leseraussen_test.go`) —
//     ohne gebautes Programm, denn die Bindungen sind die Fläche.
//   - **Tokens**: ihren Umfang hält `autorisierungsmatrix_test.go`.
//   - **Jede Fähigkeit**: je Wirkung ein Weg. Dass jede Fähigkeit eine Klasse
//     hat und wo sie durchgesetzt wird, halten `AUTORISIERUNG.md` und sein
//     Wächter.
package flaechen_test

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// rollenlage ist eine Zeile der Matrix.
type rollenlage struct {
	name  string
	rolle string // leer: keine Freigabe
	fremd bool
	darf  bool // schreiben und hinausschicken
}

func TestDieRollenmatrixUeberDieFlaechen(t *testing.T) {
	t.Parallel()

	for _, lage := range []rollenlage{
		{"eigenes Profil", "", false, true},
		{"fremd als Mitarbeiter", "mitarbeiter", true, true},
		{"fremd als Leser", "leser", true, false},
		{"fremd ohne Rolle", "", true, false},
	} {
		t.Run(lage.name, func(t *testing.T) {
			t.Parallel()

			fahreRollenlage(t, lage)
		})
	}
}

// fahreRollenlage baut den Bestand für eine Lage und fährt die drei Flächen.
func fahreRollenlage(t *testing.T, lage rollenlage) {
	t.Helper()

	var gefragt atomic.Int64

	server := gefragteGegenstelle(t, &gefragt)
	umgebung := append(eigenerBestand(t),
		app.Scryfallumgebung+"="+server.URL, "HOME="+t.TempDir(),
		schluesselbund.Bundort+"="+probebund(t))

	setID := richteLageEin(t, umgebung, lage)
	ziel := t.TempDir()

	// **Die Flächen, und was jede für die vier Wirkungen ruft.** Der MCP-
	// Server und der Dienst starten **nach** dem Einrichten, im Profil der
	// Lage — so wie jemand, der sie neben dem Fenster laufen lässt.
	mcp := starteMCPProzess(t, umgebung)
	an := starteDienst(t, umgebung)

	flaechen := map[string]rollenflaeche{
		"Kommandozeile": kommandozeilenflaeche(t, umgebung),
		"MCP":           mcpflaeche(mcp),
		"REST":          restflaeche(t, an),
	}

	for name, flaeche := range flaechen {
		// **Lesen geht in jeder Lage** — auch ohne Rolle: wer hereinkam,
		// sieht, was das Profil zeigt (Bedrohungsmodell in AUTORISIERUNG.md).
		if gelesen, gut := flaeche.sets(); !gut || !strings.Contains(gelesen, "Probe") {
			t.Errorf("%s: lesen ging nicht (%v): %s", name, gut, gelesen)
		}

		if _, gut := flaeche.deck(name + " Deck"); gut != lage.darf {
			t.Errorf("%s: ein Deck anlegen ging %v, erwartet %v", name, gut, lage.darf)
		}

		datei := filepath.Join(ziel, name+".json")
		if _, gut := flaeche.ausfuhr(setID, datei); gut != lage.darf {
			t.Errorf("%s: die Ausfuhr ging %v, erwartet %v", name, gut, lage.darf)
		}

		if _, fehlt := os.Stat(datei); lage.darf == errors.Is(fehlt, fs.ErrNotExist) {
			t.Errorf("%s: die Datei %s liegt da: %v, erwartet %v", name, datei, fehlt == nil, lage.darf)
		}

		if _, gut := flaeche.suche("probe " + name); gut != lage.darf {
			t.Errorf("%s: die Scryfallsuche ging %v, erwartet %v", name, gut, lage.darf)
		}
	}

	// **Und hinaus ging nur, was durfte** — an der Gegenstelle gezählt.
	//
	// **Über die Lage und nicht je Fläche**: die Antworten werden
	// zwischengespeichert, und nur die erste erlaubte Suche fragt wirklich.
	// Für die verbotenen Lagen ist die Zahl die Zusage — **null** —, für die
	// erlaubten der Beleg, dass die Umlenkung auf die Gegenstelle wirkt.
	if hinaus := gefragt.Load(); (hinaus > 0) != lage.darf {
		t.Errorf("%d Anfragen an Scryfall, erwartet welche: %v", hinaus, lage.darf)
	}

	// **An der Wirkung** (AK-04): im Profil der Lage stehen genau die Decks,
	// die entstehen durften.
	decks := deckeDerLage(t, umgebung)
	if erwartet := map[bool]int{true: 3, false: 0}[lage.darf]; decks != erwartet {
		t.Errorf("%d Decks im Profil der Lage, erwartet %d", decks, erwartet)
	}
}

// richteLageEin legt Bob an, öffnet ein Set im Profil der Lage und stellt
// den Zeiger dorthin; es gibt die Kennung des Sets.
//
// **`vorbereiten` läuft als Eigentümer** des Profils der Lage, gleich nach
// dem Set: was eine Lage schon braucht — ein Gegenüber, ein Deck —, legt der
// an, dem das Profil gehört, und nicht der, dessen Rolle geprüft wird.
func richteLageEin(t *testing.T, umgebung []string, lage rollenlage, vorbereiten ...func(setID string)) string {
	t.Helper()

	ich := eignerkennung(t, mussKommandozeile(t, umgebung, "profil", "--json"))

	if !lage.fremd {
		setID := setkennungAus(t, mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))
		for _, schritt := range vorbereiten {
			schritt(setID)
		}

		return setID
	}

	bob, bobsProfil := profilEinesFremden(t,
		mussKommandozeile(t, umgebung, "nutzer-neu", "Bob", "--json"))

	mussKommandozeile(t, umgebung, "nutzer-wechseln", bob)

	setID := setkennungAus(t, mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))
	for _, schritt := range vorbereiten {
		schritt(setID)
	}

	if lage.rolle != "" {
		mussKommandozeile(t, umgebung, "freigabe-setzen", bobsProfil, ich, lage.rolle)
	}

	mussKommandozeile(t, umgebung, "nutzer-wechseln", ich)
	mussKommandozeile(t, umgebung, "profil", bobsProfil)

	return setID
}

// deckeDerLage zählt die Decks im gezeigten Profil.
func deckeDerLage(t *testing.T, umgebung []string) int {
	t.Helper()

	var gelesen struct {
		Decks []json.RawMessage `json:"decks"`
	}

	roh := mussKommandozeile(t, umgebung, "decks", "--json")
	if fehler := json.Unmarshal([]byte(roh), &gelesen); fehler != nil {
		t.Fatalf("die Decks sind kein JSON: %v (%s)", fehler, roh)
	}

	return len(gelesen.Decks)
}

// rollenflaeche sind die vier Wege einer Fläche; jeder sagt, ob er glückte.
type rollenflaeche struct {
	sets    func() (string, bool)
	deck    func(name string) (string, bool)
	ausfuhr func(setID, datei string) (string, bool)
	suche   func(frage string) (string, bool)
}

func kommandozeilenflaeche(t *testing.T, umgebung []string) rollenflaeche {
	t.Helper()

	ruf := func(argumente ...string) (string, bool) {
		aus, fehler, code := kommandozeile(t, umgebung, argumente...)

		return aus + fehler, code == 0
	}

	return rollenflaeche{
		sets: func() (string, bool) { return ruf("sets", "--json") },
		deck: func(name string) (string, bool) { return ruf("deck", "neu", name, "magic") },
		ausfuhr: func(setID, datei string) (string, bool) {
			return ruf("mtgjson-schreiben", setID, "magic", datei)
		},
		suche: func(frage string) (string, bool) { return ruf("scryfall-suche", frage) },
	}
}

func mcpflaeche(mcp *mcpprozess) rollenflaeche {
	ruf := func(werkzeug string, argumente map[string]any) (string, bool) {
		text, istFehler := mcp.ruf(werkzeug, argumente)

		return text, !istFehler
	}

	return rollenflaeche{
		sets: func() (string, bool) { return ruf("lesen.sets", map[string]any{}) },
		deck: func(name string) (string, bool) {
			return ruf("schreiben.deck.anlegen", map[string]any{"name": name, "spiel": "magic"})
		},
		ausfuhr: func(setID, datei string) (string, bool) {
			return ruf("schreiben.mtgjson.schreiben",
				map[string]any{"setID": json.Number(setID), "spiel": "magic", "ziel": datei})
		},
		suche: func(frage string) (string, bool) {
			return ruf("lesen.scryfall.suche", map[string]any{"frage": frage})
		},
	}
}

func restflaeche(t *testing.T, an dienst) rollenflaeche {
	t.Helper()

	ruf := func(methode, weg, rumpf string) (string, bool) {
		status, antwort := an.ruf(t, methode, weg, rumpf)

		return antwort, status >= http.StatusOK && status < http.StatusMultipleChoices
	}

	return rollenflaeche{
		sets: func() (string, bool) { return ruf(http.MethodGet, "/sets", "") },
		deck: func(name string) (string, bool) {
			return ruf(http.MethodPost, "/decks", `{"name":"`+name+`","spiel":"magic"}`)
		},
		ausfuhr: func(setID, datei string) (string, bool) {
			rumpf, fehler := json.Marshal(map[string]string{"spiel": "magic", "ziel": datei})
			if fehler != nil {
				t.Fatalf("Rumpf bauen: %v", fehler)
			}

			return ruf(http.MethodPost, "/sets/"+setID+"/mtgjson", string(rumpf))
		},
		suche: func(frage string) (string, bool) {
			return ruf(http.MethodGet, "/scryfall/suche?frage="+url.QueryEscape(frage), "")
		},
	}
}
