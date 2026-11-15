// **Was ein Set über sich sagt, über jede Fläche ohne Fenster** — Angaben,
// Vorbehalte, Felder und die mitgelieferten Sets
// (#1169).
//
// # Warum die vier zusammen stehen
//
// Sie sind die **Auskunft** über ein Set: was darin steht (Angaben), was beim
// Einfahren nicht gedeutet wurde (Vorbehalte), welche Felder es hat, und
// welche Sets mitgeliefert kommen. Drei davon brauchen ein eingefahrenes Set;
// sie in vier Läufen je neu einzufahren kostete viermal dieselbe Vorrichtung.
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Angaben eines Sets: lesen, setzen, wiederlesen — dreimal.**
//
// # Die Zusage
//
// Jede Fläche setzt **ihre** Angabe an **ihrem** Set und liest sie zurück.
// Danach sind es drei Sets mit drei Titeln, und keine Fläche hat die des
// anderen überschrieben.
func TestDieSetangabenUeberJedeFlaeche(t *testing.T) {
	// **Mit der Saat und nicht nur mit einem Bestand.** Die Angaben eines
	// Sets kommen aus dem `set info style` seines Spiels; ohne das Spielpaket
	// antwortet jede Fläche mit „liegt in keinem Paketverzeichnis" — und der
	// Lauf prüfte dann drei gleiche Fehler statt drei gleicher Auskünfte.
	umgebung := mitEigenerSaat(t)

	// **Drei Sets, eines je Fläche.** Ein gemeinsames träfe sich im Titel,
	// und der Lauf prüfte am Ende, wer zuletzt geschrieben hat.
	//
	// **Angelegt und nicht dreimal eingefahren**: dasselbe `.mse-set` zweimal
	// einzufahren legt nichts Neues an — das ist die Zusage von
	// [TestEinZweitesEinfahrenLegtAnBeidenFlaechenNichtsAn], und der erste
	// Anlauf hier lief genau hinein: drei Kennungen, ein Set, und am Ende
	// stand nur, was der Dienst zuletzt geschrieben hatte.
	cli := neuesSet(t, umgebung, "Für die Zeile")
	mcp := neuesSet(t, umgebung, "Für das Modell")
	rest := neuesSet(t, umgebung, "Für die Schleife")

	mussKommandozeile(t, umgebung, "setangaben", cli, "setzen", "title", "Aus der Zeile")

	if text, istFehler := mcpRuf(t, umgebung, "schreiben.setangabe",
		map[string]any{"setID": zahl(t, mcp), "feld": "title", "wert": "Aus dem Modell"}); istFehler {
		t.Fatalf("schreiben.setangabe: %s", text)
	}

	an := starteDienst(t, umgebung)

	if status, rumpf := an.ruf(t, http.MethodPut, "/sets/"+rest+"/angaben/title",
		`{"wert":"Aus der Schleife"}`); status != http.StatusOK {
		t.Fatalf("PUT angaben: %d (%s)", status, rumpf)
	}

	// **Jede Fläche liest ihre eigene zurück** — über ihren eigenen Weg. Wer
	// alle drei über eine läse, prüfte den Schreibweg und nicht den Lesweg.
	if aus := mussKommandozeile(t, umgebung, "setangaben", cli); !strings.Contains(aus, "Aus der Zeile") {
		t.Errorf("die Kommandozeile liest ihre Angabe nicht zurück:\n%s", aus)
	}

	aus, istFehler := mcpRuf(t, umgebung, "lesen.setangaben", map[string]any{"setID": zahl(t, mcp)})
	if istFehler || !strings.Contains(aus, "Aus dem Modell") {
		t.Errorf("MCP liest seine Angabe nicht zurück: %s", aus)
	}

	status, rumpf := an.ruf(t, http.MethodGet, "/sets/"+rest+"/angaben", "")
	if status != http.StatusOK || !strings.Contains(rumpf, "Aus der Schleife") {
		t.Errorf("der Dienst liest seine Angabe nicht zurück: %d (%s)", status, rumpf)
	}
}

// **Eine Angabe ohne Namen wird überall gleich abgewiesen.**
//
// # Warum dieser Fall und nicht ein fremder Feldname
//
// Ein **fremder** Name ist keiner: die Angaben eines Sets kommen aus dem `set
// info style` seines Spiels, und welche es gibt, weiß nur die Vorlage. Geprüft
// wird die Gestalt (`domain.PruefeFeldname`) — ein leerer Name ergäbe eine
// Zeile, die niemand wiederfindet.
//
// **Das ist der Fehlerfall neben dem Erfolg** (§16), und er muss an allen drei
// Flächen derselbe sein. Der erste Anlauf prüfte hier `gibtesnicht` und war
// rot, weil alle drei ihn richtigerweise annahmen — der Lauf hatte eine Regel
// erwartet, die es nicht gibt.
func TestEineSetangabeOhneNamenWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := eingefahren(t, umgebung)

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "eine Angabe ohne Namen", abweisung{
		kern:   "Ein Feld braucht einen Namen",
		klasse: klasseValidierung,
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"setangaben", setID, "setzen", "   ", "egal")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.setangabe",
				map[string]any{"setID": zahl(t, setID), "feld": "   ", "wert": "egal"})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPut, "/sets/"+setID+"/angaben/%20%20",
				`{"wert":"egal"}`)
		},
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).SetwertSetzen(t.Context(), setID, "   ", "egal", nil).Fehler
		},
	})
}

// **Die Vorbehalte und die Feldnamen stehen an jeder Fläche.**
//
// # Warum beide in einem Lauf
//
// Beides ist eine **reine Auskunft** über dasselbe eingefahrene Set, und
// beides hat dieselbe Zusage: die drei Flächen nennen dasselbe. Ein Lauf je
// Auskunft führte dieselbe Vorrichtung zweimal.
func TestVorbehalteUndFelderUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := eingefahren(t, umgebung)

	an := starteDienst(t, umgebung)

	// **Die Feldnamen: `name` steht in jeder Antwort.** Das Probeset hat es,
	// und es ist der Name, unter dem eine Karte in jeder Liste erscheint.
	fuerAlleDrei(t, "die Feldnamen", "name",
		func() string { return mussKommandozeile(t, umgebung, "felder", setID) },
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.felder",
				map[string]any{"setID": zahl(t, setID)})
			if istFehler {
				t.Fatalf("lesen.felder: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet, "/sets/"+setID+"/felder", "")
			if status != http.StatusOK {
				t.Fatalf("GET felder: %d (%s)", status, rumpf)
			}

			return rumpf
		})

	// **Die Vorbehalte: das Probeset hat keine.** Die Zusage ist deshalb
	// „alle drei antworten, und keine meldet einen Befund" — und genau das ist
	// die Form, in der §3 vor `omitempty` warnt: eine **leere** Liste muss
	// als leere Liste ankommen und nicht als Fehler.
	for name, hole := range map[string]func() (int, string){
		"Kommandozeile": func() (int, string) {
			aus, fehler, code := kommandozeile(t, umgebung, "vorbehalte", setID)

			return code, aus + fehler
		},
		"MCP": func() (int, string) {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.vorbehalte",
				map[string]any{"setID": zahl(t, setID)})
			if istFehler {
				return 1, aus
			}

			return 0, aus
		},
		"REST": func() (int, string) {
			status, rumpf := an.ruf(t, http.MethodGet, "/sets/"+setID+"/vorbehalte", "")
			if status != http.StatusOK {
				return 1, rumpf
			}

			return 0, rumpf
		},
	} {
		code, antwort := hole()
		if code != 0 {
			t.Errorf("%s meldet einen Fehler auf die Vorbehalte: %s", name, antwort)
		}
	}
}

// **Die mitgelieferten Sets antworten an jeder Fläche — und zwar leer.**
//
// # Warum leer die richtige Antwort ist
//
// Sie liegen im **Bündel**, und ein Lauf aus dem Quellbaum hat keines. „Ohne
// Bündel eine leere Liste und kein Fehler" steht so an `app.Beispielsets`:
// eine Meldung „keine Beispielsets" bei jedem Start wäre eine, die niemand
// mehr liest.
//
// # Und genau deshalb ist er wertvoll
//
// **Es ist der Fall, vor dem §3 warnt.** Eine leere Liste, die als `null`
// ankommt oder ganz verschwindet, sieht an jeder einzelnen Prüfebene richtig
// aus: der Go-Test sieht die Struct-Marke nicht, der Frontend-Test geht nicht
// durch die Serialisierung. Hier geht sie durch — dreimal, über drei echte
// Prozesse.
func TestDieBeispielsetsUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	an := starteDienst(t, umgebung)

	for name, hole := range map[string]func() (string, bool){
		"Kommandozeile": func() (string, bool) {
			aus, fehler, code := kommandozeile(t, umgebung, "beispielsets", "--json")

			return aus + fehler, code == 0
		},
		"MCP": func() (string, bool) {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.beispielsets", map[string]any{})

			return aus, !istFehler
		},
		"REST": func() (string, bool) {
			status, rumpf := an.ruf(t, http.MethodGet, "/beispielsets", "")

			return rumpf, status == http.StatusOK
		},
	} {
		antwort, gut := hole()
		if !gut {
			t.Errorf("%s meldet einen Fehler auf die Beispielsets: %s", name, antwort)

			continue
		}

		// **Wirklich leer, und wirklich eine Liste.**
		//
		// Der erste Anlauf prüfte, dass ein `[` dasteht und kein `null` —
		// damit wäre `["beliebig"]` durchgekommen (Befund 4 des Zweitblicks
		// über Phase 3.8). Gelesen wird deshalb, und gezählt wird auf null.
		if laenge := listenlaenge(t, antwort, "sets"); laenge != 0 {
			t.Errorf("%s nennt %d Beispielsets — im Quellbaum gibt es keine:\n%s",
				name, laenge, antwort)
		}
	}
}

// eingefahren fährt das Probeset ein und gibt seine Kennung.
func eingefahren(t *testing.T, umgebung []string) string {
	t.Helper()

	return setkennungAus(t,
		mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))
}

// neuesSet legt ein leeres Set an und gibt seine Kennung.
func neuesSet(t *testing.T, umgebung []string, name string) string {
	t.Helper()

	mussKommandozeile(t, umgebung, "set-neu", name, "probe", "probe-schlicht")

	zeilen := setzeilenAus(t, mussKommandozeile(t, umgebung, "sets", "--json"))
	for _, zeile := range zeilen {
		if zeile.Name == name {
			return zeile.ID
		}
	}

	t.Fatalf("das angelegte Set %q steht nicht in der Liste", name)

	return ""
}
