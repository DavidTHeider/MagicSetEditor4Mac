package flaechen_test

// **Eine Partie über jede Fläche ohne Fenster** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Was hier gemessen wird
//
// **Dass dieselbe Handlung überall dasselbe tut** (§16). Ein Wächter zählt,
// **dass** eine Fläche eine Fähigkeit hat
// (`internal/werkzeuge/flaechenlauf_test.go`); dass sie dasselbe tut, hält
// erst dieser Lauf.
//
// # Und warum alle drei denselben Bestand benutzen
//
// **Weil dieselben zwei Decks dreimal an den Tisch kommen.** Damit steht
// nebenbei die Zusage aus dem Issue: ein Deck wird **eingefroren** und nicht
// verbraucht — es liegt nach drei Partien unverändert da, und die drei
// Partien sind bis auf ihre Kennung ununterscheidbar.
//
// # Was er nicht hält
//
// **Das Fenster.** Er fährt es nicht (gemessen in `FLAECHENPARITAET.md`).
// Sie prüft `internal/app/partiebindungen_test.go`. Derselbe Satz steht in
// `deckbereich_test.go` und in `pack_test.go`, und er steht dort, weil ein
// Kopf, der vier Flächen behauptet und drei fährt, schlimmer ist als keiner
// (§10).

import (
	"net/http"
	"testing"
)

// partiebild ist die Auskunft, die **alle drei** Flächen geben.
//
// # Warum nicht die ganze Antwort
//
// **Weil die Gestalten verschieden sind und das dürfen** (§16, und derselbe
// Satz steht über [fuerAlleDrei]): die Kommandozeile gibt die Antwort der
// Naht — `{"partie":…,"erlaubt":…}` —, MCP und Dienst geben einen flachen
// Verbund. Was gleich sein **muss**, ist die Auskunft.
//
// **Was hier fehlt, fehlt benannt**: der Deckabdruck am Sitz und
// `verdeckt`, `eigner`, `ordnung`, `haengtAn`, `haufen`, `zaehler` am Objekt
// stehen nur am Fenster. Wer sie hier zusicherte, sicherte etwas zu, das
// zwei der drei Flächen gar nicht sagen.
type partiebild struct {
	ID      string `json:"id"`
	Fassung string `json:"fassung"`
	Spiel   string `json:"spiel"`
	Stand   string `json:"stand"`
	Fuer    string `json:"fuer"`
	Wirt    string `json:"wirt"`

	Sitze   []bildsitz   `json:"sitze"`
	Objekte []bildobjekt `json:"objekte"`
	Erlaubt []string     `json:"erlaubt"`

	// Handlungen sind je Objekt, was der Sitz daran tun kann — **an jeder
	// Fläche dieselben** (P3,
	// #2371).
	Handlungen []bildhandlung `json:"handlungen"`

	// Mischpflichten sind die verborgenen Zonen, die jemand angesehen und
	// seitdem nicht gemischt hat — **an jeder Fläche dieselben**.
	Mischpflichten []bildmischpflicht `json:"mischpflichten"`

	// Zug ist, wo die Partie im Zug steht — **an jeder Fläche derselbe**
	// (G9, #1592).
	Zug bildzug `json:"zug"`

	// Legendenwahlen sind die offenen Fragen aus CR 704.5j — **an jeder
	// Fläche dieselben** (M12,
	// #1614).
	Legendenwahlen []bildlegendenwahl `json:"legendenwahlen"`
}

// bildhandlung ist die Handlungsauskunft zu einem Objekt.
type bildhandlung struct {
	Objekt     string   `json:"objekt"`
	Handlungen []string `json:"handlungen"`
	Primaer    string   `json:"primaer"`
	Zweit      string   `json:"zweit"`
}

// bildlegendenwahl ist eine offene Wahl aus CR 704.5j.
type bildlegendenwahl struct {
	Sitz string `json:"sitz"`
	Name string `json:"name"`
}

// bildzug ist der Zugstand, wie ihn jede Fläche nennt.
type bildzug struct {
	Nummer     int    `json:"nummer"`
	Aktiv      string `json:"aktiv"`
	Phase      string `json:"phase"`
	Schritt    string `json:"schritt"`
	Prioritaet string `json:"prioritaet"`

	// Angriffe und Bloecke sind der Kampfstand — CR 508.1, CR 509.1
	// (#1623).
	Angriffe []bildpaar `json:"angriffe"`
	Bloecke  []bildpaar `json:"bloecke"`
}

// bildpaar ist eine Kreatur und das, wogegen sie steht.
type bildpaar struct {
	Kreatur   string `json:"kreatur"`
	Ziel      string `json:"ziel"`
	Angreifer string `json:"angreifer"`
}

// bildmischpflicht ist eine offene Mischpflicht.
type bildmischpflicht struct {
	Sitz string `json:"sitz"`
	Zone string `json:"zone"`
}

// bildsitz ist ein Platz am Tisch, wie ihn jede Fläche nennt.
type bildsitz struct {
	Kennung string `json:"kennung"`
	Name    string `json:"name"`
	Karten  int    `json:"karten"`

	// Leben, Vergiftung und Manavorrat machen den Sitz zum Spieler — und
	// stehen seit `M3` an **jeder** Fläche
	// (#1605).
	Leben      int    `json:"leben"`
	Vergiftung int    `json:"vergiftung"`
	Manavorrat string `json:"manavorrat"`
}

// bildobjekt ist ein Ding auf dem Tisch, wie es jede Fläche nennt.
type bildobjekt struct {
	Kennung     string `json:"kennung"`
	Beherrscher string `json:"beherrscher"`
	Zone        string `json:"zone"`
	Karte       string `json:"karte"`
	Name        string `json:"name"`
	Getappt     bool   `json:"getappt"`
	Geheim      bool   `json:"geheim"`
}

// **Dieselben zwei Decks werden an jeder Fläche zu derselben Partie.**
//
// **Verglichen wird der ganze Verbund und nicht ein Wort darin.** Eine
// Fläche, die den Wirt wegließe oder eine Karte weniger in die Bibliothek
// legte, käme sonst durch — das ist der Befund aus Phase 3.8, der
// [gleicheAuskunft] hervorgebracht hat.
func TestEinePartieEntstehtAnJederFlaecheGleich(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	ausCLI := bildAusCLI(t, mussKommandozeile(t, umgebung, "partie-anlegen",
		"magic", "anna:Anna:"+ersterDeck, "bert:Bert:"+zweiterDeck, "--json"))

	ausMCP := bildAusFlach(t, mussMCP(t, umgebung, "schreiben.partie.anlegen",
		map[string]any{
			"spiel": "magic",
			"sitze": sitzwuensche(ersterDeck, zweiterDeck),
		}))

	ausREST := bildAusFlach(t, mussREST(t, an, http.MethodPost, "/partien",
		`{"spiel":"magic","sitze":`+
			string(mussJSON(t, sitzwuensche(ersterDeck, zweiterDeck)))+`}`))

	// **Die Kennungen sind verschieden, und das ist die Zusage.** Drei
	// Partien aus denselben Decks sind drei Partien; wären ihre Kennungen
	// gleich, überschriebe die zweite die erste.
	verschiedeneKennungen(t, ausCLI.ID, ausMCP.ID, ausREST.ID)

	gleicheBilderZweierPartien(t, "die frisch angelegte Partie", ausCLI, ausMCP)
	gleicheBilderZweierPartien(t, "die frisch angelegte Partie", ausCLI, ausREST)

	pruefeFrischePartie(t, ausCLI)
}

// **Dieselbe Partie sieht an jeder Fläche gleich aus.**
//
// **Ein Lauf über *eine* Partie**, nicht über drei gleich gebaute: der Lauf
// darüber prüft das Anlegen, dieser das Lesen. Wer beides in einen legte,
// bekäme bei einem roten Lauf zwei Verdächtige.
func TestDieselbePartieSiehtAnJederFlaecheGleichAus(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	partieID := bildAusCLI(t, mussKommandozeile(t, umgebung, "partie-anlegen",
		"magic", "anna:Anna:"+ersterDeck, "bert:Bert:"+zweiterDeck, "--json")).ID

	ausCLI := bildAusCLI(t, mussKommandozeile(t, umgebung,
		"partie-zeigen", partieID, "anna", "--json"))

	ausMCP := bildAusFlach(t, mussMCP(t, umgebung, "lesen.partie",
		map[string]any{"partieID": partieID, "fuer": "anna"}))

	ausREST := bildAusFlach(t, mussREST(t, an, http.MethodGet,
		"/partien/"+partieID+"?fuer=anna", ""))

	gleichesBild(t, "dieselbe Partie für denselben Sitz", ausCLI, ausMCP)
	gleichesBild(t, "dieselbe Partie für denselben Sitz", ausCLI, ausREST)

	if ausCLI.ID != partieID {
		t.Errorf("gelesen wurde %q, angelegt war %q", ausCLI.ID, partieID)
	}
}

// **Und eine Bibliothek bleibt an jeder Fläche geheim — auch die eigene.**
//
// **Ein eigener Lauf, weil es die Zusage ist, an der am meisten hängt.** Wer
// die Reihenfolge seiner Bibliothek kennt, kennt den nächsten Zug, ohne eine
// Karte zu lesen. **Auch für den Wirt**: seine Projektion ist dieselbe wie
// die jedes anderen Sitzes, und der Vollzustand hat an keiner der drei
// Flächen einen Weg.
func TestDieBibliothekBleibtAnJederFlaecheGeheim(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	partieID := bildAusCLI(t, mussKommandozeile(t, umgebung, "partie-anlegen",
		"magic", "anna:Anna:"+ersterDeck, "bert:Bert:"+zweiterDeck, "--json")).ID

	// **Anna ist der Wirt** — der erste Sitz, wo keiner genannt ist.
	fuer := map[string]partiebild{
		"Kommandozeile": bildAusCLI(t, mussKommandozeile(t, umgebung,
			"partie-zeigen", partieID, "anna", "--json")),
		"MCP": bildAusFlach(t, mussMCP(t, umgebung, "lesen.partie",
			map[string]any{"partieID": partieID, "fuer": "anna"})),
		"REST": bildAusFlach(t, mussREST(t, an, http.MethodGet,
			"/partien/"+partieID+"?fuer=anna", "")),
	}

	for name, bild := range fuer {
		offen := 0

		for _, objekt := range bild.Objekte {
			if objekt.Zone != "bibliothek" {
				t.Errorf("%s: das Objekt %q liegt in %q statt in der Bibliothek",
					name, objekt.Kennung, objekt.Zone)
			}

			if !objekt.Geheim || objekt.Karte != "" || objekt.Name != "" {
				offen++
			}
		}

		if offen > 0 {
			t.Errorf("%s: %d von %d Bibliothekskarten liegen offen da — "+
				"auch der Wirt darf sie nicht sehen",
				name, offen, len(bild.Objekte))
		}
	}
}
