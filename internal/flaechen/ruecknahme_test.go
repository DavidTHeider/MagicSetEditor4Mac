package flaechen_test

// **Die Rücknahme an jeder Fläche ohne Fenster** — P8
// (#2376),
// Vermerk 0197: derselbe Kern, dieselbe Antwort, dasselbe Nein. Das Fenster
// geht über die Bindung `PartieBefehl` (`internal/app`).

import (
	"encoding/json"
	"net/http"
	"testing"
)

// roheNutzlast macht aus der Nutzlast der Kommandozeile die des MCP.
func roheNutzlast(t *testing.T, roh string) map[string]any {
	t.Helper()

	nutzlast := map[string]any{}
	if fehler := json.Unmarshal([]byte(roh), &nutzlast); fehler != nil {
		t.Fatalf("die Nutzlast %q ist kein JSON: %v", roh, fehler)
	}

	return nutzlast
}

// ruecknahmeflaeche stellt einen Befehl und gibt Antwort, Abweisung und ob
// abgewiesen wurde.
type ruecknahmeflaeche struct {
	name    string
	stellen func(partieID, absender, kennung, art, nutzlast string) (string, bool)
}

func ruecknahmeflaechen(t *testing.T, umgebung []string, an dienst) []ruecknahmeflaeche {
	t.Helper()

	return []ruecknahmeflaeche{
		{"Kommandozeile", func(partieID, absender, kennung, art, nutzlast string) (string, bool) {
			aus, klage, code := kommandozeile(t, umgebung, "partie-befehl",
				partieID, absender, kennung, art, nutzlast, "--json")
			if code != 0 {
				return klage, true
			}

			return aus, false
		}},
		{"MCP", func(partieID, absender, kennung, art, nutzlast string) (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.partie.befehl", map[string]any{
				"partieID": partieID, "absender": absender, "kennung": kennung,
				"art": art, "nutzlast": roheNutzlast(t, nutzlast),
			})
		}},
		{"REST", func(partieID, absender, kennung, art, nutzlast string) (string, bool) {
			status, rumpf := an.ruf(t, http.MethodPost, "/partien/"+partieID+"/befehle",
				`{"absender":"`+absender+`","kennung":"`+kennung+`","art":"`+art+
					`","nutzlast":`+nutzlast+`}`)

			return rumpf, status >= 400
		}},
	}
}

// **Der eigene Tappzug wird zurückgenommen, eine Rücknahme nicht** — an
// jeder Fläche, mit demselben Satz.
func TestDieRuecknahmeWirktAnJederFlaecheGleich(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	for _, flaeche := range ruecknahmeflaechen(t, umgebung, an) {
		bild := spielbereitePartie(t, umgebung, ersterDeck, zweiterDeck)
		objekt := objektIn(t, bild, "anna", "bibliothek")

		if roh, abgewiesen := flaeche.stellen(bild.ID, "anna", "b1", "objekt.tappen", `{"objekt":"`+objekt+`"}`); abgewiesen {
			t.Fatalf("%s: das Tappen wurde abgewiesen: %s", flaeche.name, roh)
		}

		roh, abgewiesen := flaeche.stellen(bild.ID, "anna", "z1", "spieler.zuruecknehmen", `{}`)
		if abgewiesen {
			t.Fatalf("%s: die Rücknahme wurde abgewiesen: %s", flaeche.name, roh)
		}

		ausgang := befehlAus(t, flaeche.name, roh)
		if len(ausgang.Ereignisse) != 2 || ausgang.Ereignisse[0].Art != "spieler.zuruecknehmen" ||
			ausgang.Ereignisse[1].Art != "objekt.enttappen" {
			t.Errorf("%s: %+v", flaeche.name, ausgang.Ereignisse)
		}

		if ausgang.Fassung != naechsteFassung(t, naechsteFassung(t, bild.Fassung)) {
			t.Errorf("%s: Fassung %q", flaeche.name, ausgang.Fassung)
		}

		// **Dieselbe Kennung ist derselbe Wunsch** — dieselbe Antwort.
		nochmal, _ := flaeche.stellen(bild.ID, "anna", "z1", "spieler.zuruecknehmen", `{}`)
		if befehlAus(t, flaeche.name, nochmal).Fassung != ausgang.Fassung {
			t.Errorf("%s: die Wiederholung wirkte", flaeche.name)
		}

		meldung, abgewiesen := flaeche.stellen(bild.ID, "anna", "z2", "spieler.zuruecknehmen", `{}`)
		nurFachlich(t, flaeche.name, abgewiesen, meldung, "eine Rücknahme nimmt man nicht zurück")
	}
}

// **Nach einem fremden Folgezug weist jede Fläche ab** — und ändert nichts.
func TestDieRuecknahmeNachFremdemZugWeistJedeFlaecheAb(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	for _, flaeche := range ruecknahmeflaechen(t, umgebung, an) {
		bild := spielbereitePartie(t, umgebung, ersterDeck, zweiterDeck)

		for _, schritt := range [][2]string{{"anna", "b1"}, {"bert", "b2"}} {
			if roh, abgewiesen := flaeche.stellen(bild.ID, schritt[0], schritt[1], "spieler.leben", `{"um":-1}`); abgewiesen {
				t.Fatalf("%s: %s wurde abgewiesen: %s", flaeche.name, schritt[1], roh)
			}
		}

		meldung, abgewiesen := flaeche.stellen(bild.ID, "anna", "z1", "spieler.zuruecknehmen", `{}`)
		nurFachlich(t, flaeche.name, abgewiesen, meldung, "ein anderer Sitz gehandelt")

		nachher := bildAusCLI(t, mussKommandozeile(t, umgebung, "partie-zeigen", bild.ID, "anna", "--json"))
		if nachher.Fassung != naechsteFassung(t, naechsteFassung(t, bild.Fassung)) {
			t.Errorf("%s: die Abweisung änderte die Fassung auf %q", flaeche.name, nachher.Fassung)
		}
	}
}
