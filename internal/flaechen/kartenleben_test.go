// **Der Lebenslauf einer Karte über jede Fläche ohne Fenster** — anlegen,
// kopieren, umordnen, löschen
// (#1169).
//
// # Warum die vier zusammen stehen
//
// Weil sie **einen** Zustand teilen: die Reihenfolge der Karten in einem Set.
// Vier Läufe daneben legten je ein eigenes Set an und sähen nie, dass eine
// Kopie hinter dem Vorbild landet oder ein Löschen die Nummern nachrücken
// lässt. Was hier geprüft wird, ist nicht „der Aufruf kommt durch", sondern
// **dass alle drei Flächen denselben Zustand hinterlassen**.
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"net/http"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Anlegen, kopieren, tauschen, löschen — viermal derselbe Weg.**
//
// # Die Zusage
//
// Jede Fläche fährt denselben Ablauf über ihr **eigenes** Set und kommt zum
// selben Ergebnis: nach dem Anlegen eine Karte mehr, nach dem Kopieren die
// Werte des Vorbilds daneben, nach dem Tausch die vertauschte Reihenfolge,
// nach dem Löschen eine weniger.
//
// **Eigene Sets und nicht ein gemeinsames.** Drei Flächen auf derselben
// Kartenliste träfen sich in der Reihenfolge, und der Lauf prüfte am Ende die
// Summe dreier Eingriffe statt dreimal denselben.
func TestDerKartenlebenslaufUeberJedeFlaeche(t *testing.T) {
	for name, fahre := range map[string]func(*testing.T, []string) kartenstand{
		"Fenster":       kartenlaufFenster,
		"Kommandozeile": kartenlaufCLI,
		"MCP":           kartenlaufMCP,
		"REST":          kartenlaufRest,
	} {
		t.Run(name, func(t *testing.T) {
			umgebung := eigenerBestand(t)

			stand := fahre(t, umgebung)

			if stand.nachAnlegen != stand.vorher+1 {
				t.Errorf("nach dem Anlegen %d Karten, vorher %d",
					stand.nachAnlegen, stand.vorher)
			}

			if stand.nachKopieren != stand.nachAnlegen+1 {
				t.Errorf("nach dem Kopieren %d Karten, davor %d",
					stand.nachKopieren, stand.nachAnlegen)
			}

			// **Der Tausch ändert die Zahl nicht** — er ordnet um. Ohne diese
			// Zeile sagte der Lauf über `tauschen` nur, dass der Aufruf
			// durchkam.
			if stand.nachTausch != stand.nachKopieren {
				t.Errorf("der Tausch änderte die Zahl: %d statt %d",
					stand.nachTausch, stand.nachKopieren)
			}

			if !stand.getauscht {
				t.Error("die ersten beiden Karten stehen nach dem Tausch wie vorher")
			}

			if stand.nachLoeschen != stand.nachTausch-1 {
				t.Errorf("nach dem Löschen %d Karten, davor %d",
					stand.nachLoeschen, stand.nachTausch)
			}

			// **Und die Kopie trägt die Werte des Vorbilds.**
			//
			// Der erste Anlauf zählte nur die Karten; eine Fläche, die eine
			// **leere** Karte anlegte statt zu kopieren, wäre grün gewesen
			// (Befund 4 des Zweitblicks über Phase 3.8). Das Probeset führt
			// eine Karte namens „Doppelt" — nach dem Kopieren gibt es sie
			// zweimal.
			if stand.namenNachKopieren[stand.vorbild] != 2 {
				t.Errorf("der Name %q steht %dmal da — nach einer Kopie erwartet: zweimal",
					stand.vorbild, stand.namenNachKopieren[stand.vorbild])
			}
		})
	}
}

// kartenstand hält, was der Lauf an jeder Station gezählt hat.
type kartenstand struct {
	vorher       int
	nachAnlegen  int
	nachKopieren int
	nachTausch   int
	nachLoeschen int

	// getauscht sagt, ob die ersten beiden Karten ihre Plätze wirklich
	// getauscht haben.
	getauscht bool

	// vorbild ist der Name der Karte, die kopiert wurde, und
	// namenNachKopieren zählt, wie oft jeder Name danach vorkommt.
	vorbild           string
	namenNachKopieren map[string]int
}

func kartenlaufCLI(t *testing.T, umgebung []string) kartenstand {
	t.Helper()

	aus := mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json")
	setID := setkennungAus(t, aus)

	zaehle := func() []string {
		return kartenkennungen(t,
			mussKommandozeile(t, umgebung, "karten", setID, "--json"))
	}

	vorher := zaehle()
	vorbild := kartenname(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"), vorher[0])

	mussKommandozeile(t, umgebung, "karte", setID, "anlegen")
	nachAnlegen := zaehle()

	mussKommandozeile(t, umgebung, "karte", setID, "kopieren", vorher[0])
	nachKopieren := zaehle()
	namen := namenGezaehlt(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))

	mussKommandozeile(t, umgebung, "karte", setID, "tauschen",
		nachKopieren[0], nachKopieren[1])
	nachTausch := zaehle()

	mussKommandozeile(t, umgebung, "karte", setID, "loeschen", nachTausch[0])
	nachLoeschen := zaehle()

	return kartenstand{
		vorher: len(vorher), nachAnlegen: len(nachAnlegen),
		nachKopieren: len(nachKopieren), nachTausch: len(nachTausch),
		nachLoeschen: len(nachLoeschen),
		getauscht:    nachTausch[0] == nachKopieren[1] && nachTausch[1] == nachKopieren[0],

		vorbild: vorbild, namenNachKopieren: namen,
	}
}

func kartenlaufMCP(t *testing.T, umgebung []string) kartenstand {
	t.Helper()

	text, istFehler := mcpRuf(t, umgebung, "schreiben.set.einfahren",
		map[string]any{"pfad": probeset(t)})
	if istFehler {
		t.Fatalf("einfahren: %s", text)
	}

	// **MCP nimmt Kennungen als Zahl** — anders als Kommandozeile und Dienst,
	// die sie als Zeichenkette führen. Auch das ist eine Gestalt und keine
	// Fachlichkeit; hier wird sie einmal umgerechnet statt überall.
	setID := zahl(t, setkennungAus(t, text))
	alsZahl := func(roh string) int64 { return zahl(t, roh) }

	zaehle := func() []string {
		aus, fehlgeschlagen := mcpRuf(t, umgebung, "lesen.karten",
			map[string]any{"setID": setID})
		if fehlgeschlagen {
			t.Fatalf("lesen.karten: %s", aus)
		}

		return kartenkennungen(t, aus)
	}

	rufe := func(werkzeug string, felder map[string]any) {
		t.Helper()

		if aus, fehlgeschlagen := mcpRuf(t, umgebung, werkzeug, felder); fehlgeschlagen {
			t.Fatalf("%s: %s", werkzeug, aus)
		}
	}

	vorher := zaehle()

	rohVorher, leerfehler := mcpRuf(t, umgebung, "lesen.karten", map[string]any{"setID": setID})
	if leerfehler {
		t.Fatalf("lesen.karten: %s", rohVorher)
	}

	vorbild := kartenname(t, rohVorher, vorher[0])

	rufe("schreiben.karte.anlegen", map[string]any{"setID": setID})
	nachAnlegen := zaehle()

	rufe("schreiben.karte.kopieren", map[string]any{"setID": setID, "karteID": alsZahl(vorher[0])})
	nachKopieren := zaehle()

	rohNachKopieren, fehlgeschlagen := mcpRuf(t, umgebung, "lesen.karten",
		map[string]any{"setID": setID})
	if fehlgeschlagen {
		t.Fatalf("lesen.karten: %s", rohNachKopieren)
	}

	namen := namenGezaehlt(t, rohNachKopieren)

	rufe("schreiben.karten.tauschen",
		map[string]any{
			"setID": setID, "eineID": alsZahl(nachKopieren[0]),
			"andereID": alsZahl(nachKopieren[1]),
		})
	nachTausch := zaehle()

	rufe("schreiben.karte.loeschen",
		map[string]any{"setID": setID, "karteID": alsZahl(nachTausch[0])})
	nachLoeschen := zaehle()

	return kartenstand{
		vorher: len(vorher), nachAnlegen: len(nachAnlegen),
		nachKopieren: len(nachKopieren), nachTausch: len(nachTausch),
		nachLoeschen: len(nachLoeschen),
		getauscht:    nachTausch[0] == nachKopieren[1] && nachTausch[1] == nachKopieren[0],

		vorbild: vorbild, namenNachKopieren: namen,
	}
}

func kartenlaufRest(t *testing.T, umgebung []string) kartenstand {
	t.Helper()

	// **Eingefahren wird über die Kommandozeile.** Der Dienst hält den
	// Bestand, solange er läuft; ein Set davor einzufahren ist billiger als
	// ein zweiter Weg hinein — und es ist derselbe Bestand.
	setID := setkennungAus(t,
		mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))

	an := starteDienst(t, umgebung)

	zaehle := func() []string {
		status, rumpf := an.ruf(t, http.MethodGet, "/sets/"+setID+"/karten", "")
		if status != http.StatusOK {
			t.Fatalf("GET karten: %d (%s)", status, rumpf)
		}

		return kartenkennungen(t, rumpf)
	}

	rufe := func(methode, weg, rumpf string, erwartet int) {
		t.Helper()

		if status, kam := an.ruf(t, methode, weg, rumpf); status != erwartet {
			t.Fatalf("%s %s: %d (%s)", methode, weg, status, kam)
		}
	}

	vorher := zaehle()

	statusVor, rohVorher := an.ruf(t, http.MethodGet, "/sets/"+setID+"/karten", "")
	if statusVor != http.StatusOK {
		t.Fatalf("GET karten: %d (%s)", statusVor, rohVorher)
	}

	vorbild := kartenname(t, rohVorher, vorher[0])

	rufe(http.MethodPost, "/sets/"+setID+"/karten", "", http.StatusCreated)
	nachAnlegen := zaehle()

	rufe(http.MethodPost, "/sets/"+setID+"/karten/"+vorher[0]+"/kopie", "",
		http.StatusCreated)
	nachKopieren := zaehle()

	statusNach, rohNachKopieren := an.ruf(t, http.MethodGet, "/sets/"+setID+"/karten", "")
	if statusNach != http.StatusOK {
		t.Fatalf("GET karten: %d (%s)", statusNach, rohNachKopieren)
	}

	namen := namenGezaehlt(t, rohNachKopieren)

	rufe(http.MethodPost, "/sets/"+setID+"/karten/tausch",
		`{"eine":"`+nachKopieren[0]+`","andere":"`+nachKopieren[1]+`"}`, http.StatusOK)
	nachTausch := zaehle()

	// **200 und nicht 204** — die Antwort trägt den Verlaufsstand für die
	// Rücknahme; ein leerer Rumpf nähme dem Aufrufer den Weg zurück.
	rufe(http.MethodDelete, "/sets/"+setID+"/karten/"+nachTausch[0], "",
		http.StatusOK)
	nachLoeschen := zaehle()

	return kartenstand{
		vorher: len(vorher), nachAnlegen: len(nachAnlegen),
		nachKopieren: len(nachKopieren), nachTausch: len(nachTausch),
		nachLoeschen: len(nachLoeschen),
		getauscht:    nachTausch[0] == nachKopieren[1] && nachTausch[1] == nachKopieren[0],

		vorbild: vorbild, namenNachKopieren: namen,
	}
}

// **Eine fremde Kartenkennung wird überall gleich abgewiesen.**
//
// Sie ist der Fehlerfall, den §16 neben den Erfolg stellt: „dieselben
// fachlichen Fehler". Ohne ihn sagte der Lauf darüber nur, dass der Weg
// hinein an drei Flächen offen ist.
func TestEineFremdeKarteWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := eigenerBestand(t)

	setID := setkennungAus(t,
		mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))

	const fremd = "999999"

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "eine Karte, die es nicht gibt", abweisung{
		kern:   "die Karte 999999",
		klasse: klasseNichtGefunden,
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"karte", setID, "loeschen", fremd)

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.karte.loeschen",
				map[string]any{"setID": zahl(t, setID), "karteID": 999999})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodDelete, "/sets/"+setID+"/karten/"+fremd, "")
		},
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).KarteLoeschen(t.Context(), setID, fremd, false).Fehler
		},
	})
}
