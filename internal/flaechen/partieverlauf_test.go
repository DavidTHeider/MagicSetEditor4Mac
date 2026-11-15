package flaechen_test

// **Die Rückschau über jede Fläche** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum das einen eigenen Lauf bekommt
//
// **Weil hier dasselbe Geheimnis ein zweites Mal vorbeikommt.** Am Tisch
// kürzt die Projektion je Sitz; im Protokoll danach kürzt `RueckschauFuer`.
// Eine Fläche, die nur die erste hätte, sähe am Tisch nichts und im Verlauf
// alles — und das fiele niemandem auf, denn der Verlauf sieht aus wie eine
// Buchführung und nicht wie ein Blick in fremde Karten.
//
// # Was er hält
//
//   - Derselbe Verlauf kommt an jeder Fläche gleich an.
//   - Was Anna privat sah, steht in Bertas Rückschau an **keiner** Fläche.
//   - Eine unbekannte Partie klagt an jeder Fläche — und zwar fachlich
//     gleich, statt eine leere Rückschau auszugeben.
//
// # Was er nicht hält
//
// **Das Fenster.** Er fährt es nicht (gemessen in `FLAECHENPARITAET.md`) —
// sie prüft `internal/app/partieverlauf_intern_test.go`. Derselbe Satz steht
// über `partie_test.go`, und er steht dort, weil ein Kopf, der vier Flächen
// behauptet und drei fährt, schlimmer ist als keiner (§10).

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// verlaufsereignis ist ein Ereignis, wie **alle drei** Flächen es nennen.
type verlaufsereignis struct {
	Folge       string         `json:"folge"`
	Art         string         `json:"art"`
	Oeffentlich map[string]any `json:"oeffentlich"`
	Privat      map[string]any `json:"privat"`
	Kette       string         `json:"kette"`
}

// **Derselbe Verlauf sieht an jeder Fläche gleich aus.**
func TestDerVerlaufSiehtAnJederFlaecheGleichAus(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	partieID := partieMitEinemBlick(t, umgebung)

	ausCLI := verlaufAusCLI(t, mussKommandozeile(t, umgebung,
		"partie-verlauf", partieID, "anna", "--json"))

	ausMCP := verlaufAusFlach(t, mussMCP(t, umgebung, "lesen.partieverlauf",
		map[string]any{"partieID": partieID, "fuer": "anna"}))

	ausREST := verlaufAusFlach(t, mussREST(t, an, http.MethodGet,
		"/partien/"+partieID+"/verlauf?fuer=anna", ""))

	gleicherVerlauf(t, "die Rückschau desselben Sitzes", ausCLI, ausMCP)
	gleicherVerlauf(t, "die Rückschau desselben Sitzes", ausCLI, ausREST)

	// **Und sie trägt wirklich etwas** — sonst verglichen drei leere Listen
	// drei leere Listen, und der Lauf wäre grün, ohne etwas zu messen (§10).
	if len(ausCLI) == 0 {
		t.Fatal("die Rückschau ist leer — dann prüft dieser Lauf nichts")
	}

	if !strings.Contains(alsText(t, ausCLI), "karten") {
		t.Errorf("Annas eigener Blick fehlt in ihrer Rückschau: %+v", ausCLI)
	}
}

// **Was Anna privat sah, steht in Bertas Rückschau an keiner Fläche.**
//
// **Die Gegenprobe zum Lauf darüber** (§10): dort ist die Zusage „überall
// dasselbe", hier „überall dieselbe Grenze". Eine Fläche, die nur die erste
// hielte, gäbe an allen dreien gleich viel zu viel her.
func TestDieRueckschauKuerztAnJederFlaecheGleich(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	partieID := partieMitEinemBlick(t, umgebung)

	fuerBert := map[string][]verlaufsereignis{
		"Kommandozeile": verlaufAusCLI(t, mussKommandozeile(t, umgebung,
			"partie-verlauf", partieID, "bert", "--json")),
		"MCP": verlaufAusFlach(t, mussMCP(t, umgebung, "lesen.partieverlauf",
			map[string]any{"partieID": partieID, "fuer": "bert"})),
		"REST": verlaufAusFlach(t, mussREST(t, an, http.MethodGet,
			"/partien/"+partieID+"/verlauf?fuer=bert", "")),
	}

	for name, verlauf := range fuerBert {
		// **Bert sieht, *dass* gesucht wurde** — die Reihenfolge der Züge
		// ist der Verlauf der Partie und kein Geheimnis.
		if len(verlauf) == 0 {
			t.Errorf("%s: Bert sieht gar keinen Verlauf", name)

			continue
		}

		for _, ereignis := range verlauf {
			if len(ereignis.Privat) != 0 {
				t.Errorf("%s: Folge %s trägt für Bert eine Ergänzung: %v",
					name, ereignis.Folge, ereignis.Privat)
			}
		}

		// **Und nicht, *was* gefunden wurde.**
		if strings.Contains(alsText(t, verlauf), "karten") {
			t.Errorf("%s: Bert sieht Annas Blick: %+v", name, verlauf)
		}
	}
}

// **Eine unbekannte Partie klagt an jeder Fläche.**
//
// **Und das ist keine Selbstverständlichkeit**: der Speicher fragt eine
// Tabelle nach Zeilen, und keine ist keine. Ohne die Existenzfrage in
// `service` antworteten alle drei Flächen auf einen Tippfehler mit „in
// dieser Partie ist noch nichts geschehen" (§8).
func TestEineUnbekanntePartieKlagtAnJederFlaeche(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)

	_, _, code := kommandozeile(t, umgebung,
		"partie-verlauf", "gibt-es-nicht", "anna", "--json")
	if code == 0 {
		t.Error("Kommandozeile: eine unbekannte Partie kam durch")
	}

	if roh, istFehler := mcpRuf(t, umgebung, "lesen.partieverlauf",
		map[string]any{"partieID": "gibt-es-nicht", "fuer": "anna"},
	); !istFehler {
		t.Errorf("MCP: eine unbekannte Partie kam durch: %s", roh)
	}

	if status, antwort := an.ruf(t, http.MethodGet,
		"/partien/gibt-es-nicht/verlauf?fuer=anna", "",
	); status != http.StatusNotFound {
		t.Errorf("REST: Status %d, erwartet 404 (%s)", status, antwort)
	}
}

// partieMitEinemBlick legt eine Partie an, in der Anna einmal hingesehen hat.
//
// **Ein Blick und kein Tappen**: nur er legt eine **private** Ergänzung an,
// und ohne eine prüfte die Rückschau nichts, was es zu kürzen gäbe. **Nach
// dem Beginn** (#1851) — seine Schritte stehen damit auch im Verlauf, und sie
// nennen, wie viele Karten eine Hand hat, und nicht welche.
func partieMitEinemBlick(t *testing.T, umgebung []string) string {
	t.Helper()

	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)
	partieID := spielbereitePartie(t, umgebung, ersterDeck, zweiterDeck).ID

	mussKommandozeile(t, umgebung, "partie-befehl",
		partieID, "anna", "b1", "zone.blick", `{"zone":"bibliothek"}`)

	return partieID
}

// verlaufAusCLI liest die Antwort der Kommandozeile — sie trägt eine Hülle.
func verlaufAusCLI(t *testing.T, roh string) []verlaufsereignis {
	t.Helper()

	var huelle struct {
		Verlauf []verlaufsereignis `json:"verlauf"`
	}

	if err := json.Unmarshal([]byte(roh), &huelle); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	return huelle.Verlauf
}

// verlaufAusFlach liest die Antwort von MCP und Dienst — sie ist die Liste.
//
// **Zwei Leser und nicht einer**, weil die Gestalten verschieden sind und das
// dürfen (§16): gleich sein muss die **Auskunft**, nicht die Hülle. Derselbe
// Satz steht über [partiebild].
func verlaufAusFlach(t *testing.T, roh string) []verlaufsereignis {
	t.Helper()

	var verlauf []verlaufsereignis

	if err := json.Unmarshal([]byte(roh), &verlauf); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	return verlauf
}

// gleicherVerlauf verlangt, dass zwei Flächen dasselbe Protokoll beschreiben.
//
// **Der ganze Verbund und nicht ein Wort darin** — eine Fläche, die die Kette
// wegließe oder eine Ergänzung unterschlüge, käme sonst durch.
func gleicherVerlauf(t *testing.T, was string, eine, andere []verlaufsereignis) {
	t.Helper()

	if !reflect.DeepEqual(eine, andere) {
		t.Errorf("%s: die beiden Flächen sagen Verschiedenes:\n%+v\n%+v",
			was, eine, andere)
	}
}

// alsText macht aus einem Verlauf wieder JSON — für die Frage, ob ein Wort
// darin vorkommt.
func alsText(t *testing.T, verlauf []verlaufsereignis) string {
	t.Helper()

	return string(mussJSON(t, verlauf))
}
