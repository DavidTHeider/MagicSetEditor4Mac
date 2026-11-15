package flaechen_test

// **Spielbefehle und der Stand über jede Fläche ohne Fenster** — Werkstück G1
// von Phase 5.5
// (#1458).
//
// **Aus `partie_test.go` herausgelöst**, damit beide Dateien unter der
// Zeilengrenze aus §2 bleiben. Der Schnitt folgt der Frage: dort, **wie** eine
// Partie entsteht und aussieht; hier, **was** sich an ihr ändern lässt.
//
// # Und warum die Fehlerwege hier so viel Platz bekommen
//
// **Weil §16 „dieselben fachlichen Fehler" verlangt und nicht „dieselbe
// Fähigkeit".** Der Wächter daneben zählt, dass eine Fläche einen Weg hat;
// dass sie beim selben Irrtum dasselbe Nein gibt, steht nur hier.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// befehlsbild ist, was aus einem Spielbefehl wurde — an jeder Fläche.
//
// **Die Kommandozeile gibt denselben flachen Verbund**: `partie-befehl`
// schreibt die Antwort der Naht, und die trägt `fassung` und `ereignisse`
// nebeneinander. Für die Partie selbst ist das anders, und der Grund steht
// über [partiebild].
type befehlsbild struct {
	Fassung    string `json:"fassung"`
	Ereignisse []struct {
		Folge string `json:"folge"`
		Art   string `json:"art"`
		Kette string `json:"kette"`
	} `json:"ereignisse"`
}

// **Derselbe Befehl wirkt an jeder Fläche — und zweimal gestellt nur einmal.**
//
// **Die Wiederholung gehört in denselben Lauf.** Sie ist keine zweite
// Fähigkeit, sondern die Zusage der ersten: „mit derselben Kennung geschieht
// nichts ein zweites Mal" (§7, Regel 4). Wer sie danebenstellte, prüfte sie
// an einer Partie, die den ersten Befehl gar nicht gesehen hat.
func TestEinSpielbefehlWirktAnJederFlaecheUndNurEinmal(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	for _, fall := range []struct {
		name    string
		stellen func(partieID, objekt string) string
	}{
		{"Kommandozeile", func(partieID, objekt string) string {
			return mussKommandozeile(t, umgebung, "partie-befehl",
				partieID, "anna", "b1", "objekt.tappen",
				`{"objekt":"`+objekt+`"}`, "--json")
		}},
		{"MCP", func(partieID, objekt string) string {
			return mussMCP(t, umgebung, "schreiben.partie.befehl",
				map[string]any{
					"partieID": partieID, "absender": "anna",
					"kennung": "b1", "art": "objekt.tappen",
					"nutzlast": map[string]any{"objekt": objekt},
				})
		}},
		{"REST", func(partieID, objekt string) string {
			return mussREST(t, an, http.MethodPost,
				"/partien/"+partieID+"/befehle",
				`{"absender":"anna","kennung":"b1","art":"objekt.tappen",`+
					`"nutzlast":{"objekt":"`+objekt+`"}}`)
		}},
	} {
		bild := spielbereitePartie(t, umgebung, ersterDeck, zweiterDeck)
		partieID := bild.ID
		objekt := objektIn(t, bild, "anna", "bibliothek")

		// **Eine Fassung weiter als nach dem Beginn** — seine Schritte heben
		// sie selbst (#1851); bis dahin stand hier „2", eine nach der
		// frischen Partie.
		danach := naechsteFassung(t, bild.Fassung)

		erst := befehlAus(t, fall.name, fall.stellen(partieID, objekt))

		if erst.Fassung != danach {
			t.Errorf("%s: Fassung %q nach dem ersten Befehl, erwartet %s",
				fall.name, erst.Fassung, danach)
		}

		if len(erst.Ereignisse) != 1 ||
			erst.Ereignisse[0].Art != "objekt.tappen" {
			t.Errorf("%s: %+v", fall.name, erst.Ereignisse)
		}

		// **Jedes Ereignis trägt den Abdruck seines Vorgängers** — angehängt
		// und nicht überschrieben ([0044]). Eine leere Kette wäre ein Strom,
		// an dem sich nichts nachrechnen lässt.
		if erst.Ereignisse[0].Kette == "" {
			t.Errorf("%s: das Ereignis trägt keine Kette", fall.name)
		}

		// **Dieselbe Kennung ein zweites Mal** — derselbe Wunsch, dieselbe
		// Antwort, kein zweiter Zug.
		nochmal := befehlAus(t, fall.name, fall.stellen(partieID, objekt))

		if nochmal.Fassung != erst.Fassung {
			t.Errorf("%s: die Wiederholung hob die Fassung von %q auf %q",
				fall.name, erst.Fassung, nochmal.Fassung)
		}

		pruefeGenauEinGetapptes(t, fall.name, umgebung, partieID, danach)
	}
}

// **Ein fremder Absender wird an jeder Fläche mit demselben Satz abgewiesen.**
//
// **Der Fehler wird benannt und nicht nur gezählt**: „irgendein Misserfolg"
// ließe einen Weg durchgehen, der an der Objektkennung scheitert statt am
// Absender. Derselbe Satz steht über `deckbereich_test.go`.
func TestEinFremderAbsenderWirdAnJederFlaecheAbgewiesen(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	partieID := frischePartie(t, umgebung, ersterDeck, zweiterDeck)
	objekt := erstesObjekt(t, umgebung, partieID)

	const grund = "sitzt nicht an diesem Tisch"

	t.Run("Kommandozeile", func(t *testing.T) {
		_, klage, code := kommandozeile(t, umgebung, "partie-befehl",
			partieID, "carla", "b9", "objekt.tappen",
			`{"objekt":"`+objekt+`"}`)

		nurFachlich(t, "die Kommandozeile", code != 0, klage, grund)
	})

	t.Run("MCP", func(t *testing.T) {
		roh, istFehler := mcpRuf(t, umgebung, "schreiben.partie.befehl",
			map[string]any{
				"partieID": partieID, "absender": "carla",
				"kennung": "b9", "art": "objekt.tappen",
				"nutzlast": map[string]any{"objekt": objekt},
			})

		nurFachlich(t, "MCP", istFehler, roh, grund)
	})

	t.Run("REST", func(t *testing.T) {
		status, rumpf := an.ruf(t, http.MethodPost,
			"/partien/"+partieID+"/befehle",
			`{"absender":"carla","kennung":"b9","art":"objekt.tappen",`+
				`"nutzlast":{"objekt":"`+objekt+`"}}`)

		nurFachlich(t, "der Dienst", status >= 400, rumpf, grund)
	})
}

// **Eine beendete Partie nimmt an jeder Fläche keine Befehle mehr an.**
//
// **Und das Beenden selbst geht über jede Fläche** — sonst prüfte dieser
// Lauf drei Verweigerungen und einen Weg, sie herbeizuführen.
func TestEineBeendetePartieNimmtAnJederFlaecheKeineBefehleMehr(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	for _, fall := range []struct {
		name    string
		beenden func(partieID string) partiebild
	}{
		{"Kommandozeile", func(partieID string) partiebild {
			return bildAusCLI(t, mussKommandozeile(t, umgebung,
				"partie-stand", partieID, "anna", "beenden", "--json"))
		}},
		{"MCP", func(partieID string) partiebild {
			return bildAusFlach(t, mussMCP(t, umgebung,
				"schreiben.partie.stand", map[string]any{
					"partieID": partieID, "fuer": "anna", "stand": "beenden",
				}))
		}},
		{"REST", func(partieID string) partiebild {
			return bildAusFlach(t, mussREST(t, an, http.MethodPost,
				"/partien/"+partieID+"/stand",
				`{"fuer":"anna","stand":"beenden"}`))
		}},
	} {
		partieID := frischePartie(t, umgebung, ersterDeck, zweiterDeck)
		objekt := erstesObjekt(t, umgebung, partieID)

		nach := fall.beenden(partieID)

		if nach.Stand != "beendet" {
			t.Errorf("%s: Stand %q nach dem Beenden", fall.name, nach.Stand)
		}

		// **Und ein Sitz, der nichts mehr darf, bekommt auch keine Liste**
		// — `erlaubt` ist leer, statt eine Einladung auszusprechen, die der
		// nächste Ruf ablehnen müsste.
		if len(nach.Erlaubt) != 0 {
			t.Errorf("%s: erlaubt %v an einer beendeten Partie",
				fall.name, nach.Erlaubt)
		}

		_, klage, code := kommandozeile(t, umgebung, "partie-befehl",
			partieID, "anna", "b1", "objekt.tappen",
			`{"objekt":"`+objekt+`"}`)

		nurFachlich(t, fall.name, code != 0, klage,
			"sie nimmt keine Befehle an")
	}
}

// nurFachlich verlangt ein Nein **mit** seinem Grund.
func nurFachlich(t *testing.T, wer string, abgewiesen bool, meldung, grund string) {
	t.Helper()

	switch {
	case !abgewiesen:
		t.Errorf("%s ließ es durch: %s", wer, meldung)
	case !strings.Contains(meldung, grund):
		t.Errorf("%s nennt den Grund nicht — %q fehlt in %q",
			wer, grund, meldung)
	}
}

// pruefeGenauEinGetapptes zählt nach, was der Befehl bewirkt hat, und
// verlangt die Fassung, die er hinterlassen hat.
//
// **Gezählt und nicht gesucht.** „Irgendwo steht getappt" wäre auch dann
// grün, wenn die Wiederholung ein zweites Objekt getappt hätte — und genau
// das ist die Frage dieses Laufs.
func pruefeGenauEinGetapptes(
	t *testing.T, wer string, umgebung []string, partieID, fassung string,
) {
	t.Helper()

	bild := bildAusCLI(t, mussKommandozeile(t, umgebung,
		"partie-zeigen", partieID, "anna", "--json"))

	getappt := 0

	for _, objekt := range bild.Objekte {
		if objekt.Getappt {
			getappt++
		}
	}

	if getappt != 1 {
		t.Errorf("%s: %d getappte Objekte, erwartet genau eines",
			wer, getappt)
	}

	if bild.Fassung != fassung {
		t.Errorf("%s: die Partie steht bei Fassung %q, erwartet %s",
			wer, bild.Fassung, fassung)
	}
}

// frischePartie legt eine Partie aus den beiden Decks an.
//
// **Über die Kommandozeile**, weil sie hier Vorbereitung ist und nicht das
// Geprüfte — derselbe Grund wie bei [zweiDecks].
func frischePartie(t *testing.T, umgebung []string, erster, zweiter string) string {
	t.Helper()

	return bildAusCLI(t, mussKommandozeile(t, umgebung, "partie-anlegen",
		"magic", "anna:Anna:"+erster, "bert:Bert:"+zweiter, "--json")).ID
}

// erstesObjekt gibt die Kennung eines Objekts, das Anna beherrscht.
//
// **Ihres und nicht irgendeines**: ein Befehl auf ein fremdes Objekt wäre
// unter dem Sandkasten zwar erlaubt, prüfte hier aber eine andere Frage.
func erstesObjekt(t *testing.T, umgebung []string, partieID string) string {
	t.Helper()

	bild := bildAusCLI(t, mussKommandozeile(t, umgebung,
		"partie-zeigen", partieID, "anna", "--json"))

	for _, objekt := range bild.Objekte {
		if objekt.Beherrscher == "anna" {
			return objekt.Kennung
		}
	}

	t.Fatalf("die Partie %s hat kein Objekt von Anna: %+v", partieID, bild.Objekte)

	return ""
}

// befehlAus liest die Antwort auf einen Spielbefehl.
func befehlAus(t *testing.T, wer, roh string) befehlsbild {
	t.Helper()

	var bild befehlsbild

	if err := json.Unmarshal([]byte(roh), &bild); err != nil {
		t.Fatalf("%s: die Antwort ist kein gültiges JSON: %v (%s)",
			wer, err, roh)
	}

	return bild
}
