package flaechen_test

// **Ein Zauber geht an jeder Fläche auf den Stapel** — Werkstück M4 der
// Regelmaschine
// (#1606).
//
// # Was dieser Lauf hält und was nicht
//
// **Er hält den Weg, nicht die Regel.** Was beim Wirken geschieht, misst
// `internal/domain`; hier geht es darum, dass der Befehl über jede Fläche
// ankommt — **samt seiner Nutzlast**. Die ist bei `objekt.wirken` eine
// **Liste** (`ziele`), und eine Liste ist die Stelle, an der eine Fläche
// gern anders aussieht als die andere (§16).

import (
	"net/http"
	"testing"
)

// **Derselbe Zauber, dieselben Ziele, an jeder Fläche.**
func TestEinZauberGehtAnJederFlaecheAufDenStapel(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	fuer := map[string]func(partieID, objekt string){
		"Kommandozeile": func(partieID, objekt string) {
			mussKommandozeile(t, umgebung, "partie-befehl", partieID, "anna",
				"b2", "objekt.wirken",
				`{"objekt":"`+objekt+`","ziele":["bert"]}`)
		},

		"MCP": func(partieID, objekt string) {
			mussMCP(t, umgebung, "schreiben.partie.befehl", map[string]any{
				"partieID": partieID, "absender": "anna", "kennung": "b2",
				"art": "objekt.wirken",
				"nutzlast": map[string]any{
					"objekt": objekt, "ziele": []any{"bert"},
				},
			})
		},

		"REST": func(partieID, objekt string) {
			mussREST(t, an, http.MethodPost, "/partien/"+partieID+"/befehle",
				`{"absender":"anna","kennung":"b2","art":"objekt.wirken",`+
					`"nutzlast":{"objekt":"`+objekt+`","ziele":["bert"]}}`)
		},
	}

	for name, stelle := range fuer {
		t.Run(name, func(t *testing.T) {
			bild := spielbereitePartie(t, umgebung, ersterDeck, zweiterDeck)

			// **Eine Karte von der Hand** — die Starthand hat sie über den
			// gewöhnlichen Weg dorthin gebracht (CR 103.5). Gewirkt wird von
			// dort (CR 601.2a). Bis #1851 schob der Lauf sie erst aus der
			// Bibliothek hinüber, weil ohne Beginn keine Hand da war.
			stelle(bild.ID, objektIn(t, bild, "anna", "hand"))

			danach := bildAusCLI(t, mussKommandozeile(t, umgebung,
				"partie-zeigen", bild.ID, "anna", "--json"))

			aufDemStapel := 0

			for _, eines := range danach.Objekte {
				if eines.Zone == "stapel" {
					aufDemStapel++
				}
			}

			if aufDemStapel != 1 {
				t.Fatalf("%d Objekte auf dem Stapel — diese Fläche legt "+
					"keinen Zauber hinauf", aufDemStapel)
			}
		})
	}
}
