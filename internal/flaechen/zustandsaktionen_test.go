package flaechen_test

// **Die zustandsbasierten Aktionen laufen an jeder Fläche** — Werkstück M5
// der Regelmaschine
// (#1607).
//
// # Warum dieser Lauf einzeln steht
//
// **Weil die Frage lautet: setzt der Regelautomatismus überall durch?** Ein
// Wächter zählt, **dass** jede Fläche eine Fähigkeit hat (§16); er sagt
// nichts darüber, ob sie dasselbe **tut**. Hier wird es getan: an jeder der
// drei Flächen ohne Fenster fällt derselbe Sitz auf null Leben, und an jeder
// endet dieselbe Partie mit demselben Grund.
//
// **Das Fenster steht nicht dabei**, und zwar aus demselben Grund wie im Lauf
// nebenan: es startet keinen Prozess, den ein Test treiben kann. Der Weg ist
// derselbe — `PartieBefehl` an der Wails-Naht ruft dieselbe
// `service.Partieleitung` wie die drei hier.

import (
	"net/http"
	"testing"
)

// **Null Leben beendet die Partie — an der Kommandozeile, über MCP und über
// REST.**
//
// CR 704.5a mit CR 104.2a (Fassung 2026-08-19): wer auf null fällt, verliert;
// wer übrig bleibt, gewinnt.
func TestNullLebenBeendetDiePartieAnJederFlaeche(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	// **Drei Partien und nicht eine** — jede Fläche braucht eine eigene,
	// denn die erste ist nach dem Befehl zu Ende und nimmt keinen zweiten
	// an. Derselbe Grund wie beim Anlegen nebenan.
	fuer := map[string]func(partieID string){
		"Kommandozeile": func(partieID string) {
			mussKommandozeile(t, umgebung, "partie-befehl", partieID, "bert",
				"b1", "spieler.leben", `{"um":-20}`)
		},

		"MCP": func(partieID string) {
			mussMCP(t, umgebung, "schreiben.partie.befehl", map[string]any{
				"partieID": partieID, "absender": "bert", "kennung": "b1",
				"art": "spieler.leben", "nutzlast": map[string]any{"um": -20},
			})
		},

		"REST": func(partieID string) {
			mussREST(t, an, http.MethodPost, "/partien/"+partieID+"/befehle",
				`{"absender":"bert","kennung":"b1","art":"spieler.leben",`+
					`"nutzlast":{"um":-20}}`)
		},
	}

	for name, stelle := range fuer {
		t.Run(name, func(t *testing.T) {
			partieID := spielbereitePartie(t, umgebung, ersterDeck,
				zweiterDeck).ID

			stelle(partieID)

			danach := bildAusCLI(t, mussKommandozeile(t, umgebung,
				"partie-zeigen", partieID, "anna", "--json"))

			if danach.Stand != "beendet" {
				t.Fatalf("der Stand ist %q — der Automatismus greift an "+
					"dieser Fläche nicht", danach.Stand)
			}
		})
	}
}

// **Und ein Leben mehr beendet nichts** — die Gegenprobe.
//
// **Ohne sie sagte der Lauf darüber wenig**: „beendet" wäre auch dann
// richtig, wenn die Partie aus einem anderen Grund endete. Hier fällt
// derselbe Sitz auf **eins** statt auf null, und derselbe Weg lässt die
// Partie laufen (§10: eine Gegenprobe, die nichts ändert, ist keine).
func TestEinLebenMehrBeendetAnKeinerFlaecheEtwas(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	partieID := spielbereitePartie(t, umgebung, ersterDeck, zweiterDeck).ID

	mussKommandozeile(t, umgebung, "partie-befehl", partieID, "bert",
		"b1", "spieler.leben", `{"um":-19}`)

	danach := bildAusCLI(t, mussKommandozeile(t, umgebung,
		"partie-zeigen", partieID, "anna", "--json"))

	if danach.Stand != "laeuft" {
		t.Fatalf("der Stand ist %q — bei einem Leben endet nichts",
			danach.Stand)
	}
}
