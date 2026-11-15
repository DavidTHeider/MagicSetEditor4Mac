package flaechen_test

// **Die Gestalten und die Legendenwahl an jeder Fläche** — Werkstück M12 der
// Regelmaschine
// (#1614).
//
// # Warum dieser Lauf einzeln steht
//
// **Weil CR 704.5j eine Frage stellt, und eine Frage muss ankommen.** Ein
// Wächter zählt, **dass** jede Fläche den Befehl kennt (§16); er sagt nicht,
// dass sie auch **zeigt**, dass gerade jemand gefragt ist. Eine Pflicht, die
// niemand sieht, erfüllt niemand — und die Partie steht still.
//
// **Das Fenster steht nicht dabei**, aus demselben Grund wie nebenan: es
// startet keinen Prozess, den ein Test treiben kann. Der Weg ist derselbe.

import (
	"net/http"
	"testing"
)

// zweiLegenden gibt die Nutzlast für zwei gleichnamige legendäre
// Spielsteine — CR 111.1 mit CR 205.4a.
func zweiLegenden(kennung string) string {
	return `{"kennung":"` + kennung + `","felder":{` +
		`"name":"Zwilling","super type":"Legendary Creature",` +
		`"power":"2","toughness":"2"}}`
}

// **Zwei gleichnamige Legenden stellen an jeder Fläche dieselbe Frage** —
// CR 704.5j (Fassung 2026-08-19).
func TestDieLegendenwahlStehtAnJederFlaeche(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	partieID := spielbereitePartie(t, umgebung, ersterDeck, zweiterDeck).ID

	for _, kennung := range []string{"anna/zwilling/0", "anna/zwilling/1"} {
		mussKommandozeile(t, umgebung, "partie-befehl", partieID, "anna",
			"b-"+kennung, "spieler.spielstein", zweiLegenden(kennung))
	}

	fuer := map[string]func() partiebild{
		"Kommandozeile": func() partiebild {
			return bildAusCLI(t, mussKommandozeile(t, umgebung,
				"partie-zeigen", partieID, "anna", "--json"))
		},

		"MCP": func() partiebild {
			return bildAusFlach(t, mussMCP(t, umgebung, "lesen.partie", map[string]any{
				"partieID": partieID, "fuer": "anna",
			}))
		},

		"REST": func() partiebild {
			return bildAusFlach(t, mussREST(t, an, http.MethodGet,
				"/partien/"+partieID+"?fuer=anna", ""))
		},
	}

	for name, holen := range fuer {
		t.Run(name, func(t *testing.T) {
			bild := holen()

			if anzahl := len(bild.Legendenwahlen); anzahl != 1 {
				t.Fatalf("%s nennt %d offene Wahlen: %v",
					name, anzahl, bild.Legendenwahlen)
			}

			wahl := bild.Legendenwahlen[0]
			if wahl.Sitz != "anna" || wahl.Name != "Zwilling" {
				t.Errorf("%s sagt, %q solle über %q wählen",
					name, wahl.Sitz, wahl.Name)
			}
		})
	}
}

// **Und ein einzelner Spielstein stellt keine** — die Gegenprobe.
//
// **Ohne sie sagte der Lauf darüber wenig**: eine Fläche, die immer eine
// Wahl nennte, bestünde ihn ebenfalls (§10).
func TestEinEinzelnerSpielsteinStelltKeineWahl(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	partieID := spielbereitePartie(t, umgebung, ersterDeck, zweiterDeck).ID

	mussKommandozeile(t, umgebung, "partie-befehl", partieID, "anna",
		"b1", "spieler.spielstein", zweiLegenden("anna/zwilling/0"))

	bild := bildAusCLI(t, mussKommandozeile(t, umgebung,
		"partie-zeigen", partieID, "anna", "--json"))

	if anzahl := len(bild.Legendenwahlen); anzahl != 0 {
		t.Fatalf("es stehen %d Wahlen an: %v", anzahl, bild.Legendenwahlen)
	}
}

// **Die Wahl wird an jeder Fläche beantwortet** — CR 704.5j.
func TestDieLegendenwahlWirdAnJederFlaecheBeantwortet(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	fuer := map[string]func(partieID string){
		"Kommandozeile": func(partieID string) {
			mussKommandozeile(t, umgebung, "partie-befehl", partieID, "anna",
				"w1", "spieler.legende", `{"objekt":"anna/zwilling/0"}`)
		},

		"MCP": func(partieID string) {
			mussMCP(t, umgebung, "schreiben.partie.befehl", map[string]any{
				"partieID": partieID, "absender": "anna", "kennung": "w1",
				"art": "spieler.legende",
				"nutzlast": map[string]any{
					"objekt": "anna/zwilling/0",
				},
			})
		},

		"REST": func(partieID string) {
			mussREST(t, an, http.MethodPost, "/partien/"+partieID+"/befehle",
				`{"absender":"anna","kennung":"w1","art":"spieler.legende",`+
					`"nutzlast":{"objekt":"anna/zwilling/0"}}`)
		},
	}

	for name, stelle := range fuer {
		t.Run(name, func(t *testing.T) {
			partieID := spielbereitePartie(t, umgebung, ersterDeck,
				zweiterDeck).ID

			for _, kennung := range []string{
				"anna/zwilling/0", "anna/zwilling/1",
			} {
				mussKommandozeile(t, umgebung, "partie-befehl", partieID,
					"anna", "b-"+kennung, "spieler.spielstein",
					zweiLegenden(kennung))
			}

			stelle(partieID)

			bild := bildAusCLI(t, mussKommandozeile(t, umgebung,
				"partie-zeigen", partieID, "anna", "--json"))

			if anzahl := len(bild.Legendenwahlen); anzahl != 0 {
				t.Fatalf("%s lässt %d Wahlen offen", name, anzahl)
			}

			// **Der übrige Spielstein ist ganz weg** — CR 704.5d: was
			// den Tisch verlässt, hört auf zu existieren.
			for _, objekt := range bild.Objekte {
				if objekt.Kennung == "anna/zwilling/1" {
					t.Errorf("%s hat den zweiten Zwilling noch: %+v",
						name, objekt)
				}
			}
		})
	}
}
