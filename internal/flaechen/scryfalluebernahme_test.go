// Übernehmen, Bild und Sammeldaten über jede Fläche
// (#1028).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"
)

// **Kartendaten übernehmen — über jede Fläche.**
//
// Der Ablauf ist an jeder Fläche derselbe: Set anlegen, Karte darin finden,
// ihren Regeltext leeren, dann von Scryfall übernehmen — und danach steht er
// wieder da.
func TestKartendatenUebernehmenUeberJedeFlaeche(t *testing.T) {
	gegenstelle := scryfallgegenstelle(t)

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		setID, karteID := geleerteKarte(t, umgebung)

		mussKommandozeile(t, umgebung, "scryfall", setID, karteID, "Erste Probe",
			"--uebernehmen", "rule text", "--json")
		regeltextPruefen(t, umgebung, setID)
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		setID, karteID := geleerteKarte(t, umgebung)

		text, istFehler := mcpRuf(t, umgebung, "schreiben.scryfall.uebernehmen",
			map[string]any{
				"setID": zahl(t, setID), "karteID": zahl(t, karteID),
				"name": "Erste Probe", "felder": []string{"rule text"},
			})
		if istFehler {
			t.Fatalf("schreiben.scryfall.uebernehmen: %s", text)
		}

		regeltextPruefen(t, umgebung, setID)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		setID, karteID := geleerteKarte(t, umgebung)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost,
			"/sets/"+setID+"/karten/"+karteID+"/scryfall",
			`{"name":"Erste Probe","felder":["rule text"]}`)
		if status != http.StatusOK {
			t.Fatalf("POST scryfall: %d (%s)", status, rumpf)
		}

		regeltextPruefen(t, umgebung, setID)
	})
}

// **Ein Feld, das nicht im Vorschlag steht, wird überall abgewiesen.**
//
// # Warum gerade dieser Fall
//
// Er ist die **Grenze** der Übernahme (§16, Regel 6): „Nichts wird still
// übernommen." Stünde sie nur an einer Fläche, könnte ein Aufrufer über eine
// andere jeden Wert setzen und ihn mit „von Scryfall" beschriften.
func TestEinFeldAusserhalbDesVorschlagsWirdUeberallAbgewiesen(t *testing.T) {
	gegenstelle := scryfallgegenstelle(t)

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		setID, karteID := geleerteKarte(t, umgebung)

		aus, _, code := kommandozeile(t, umgebung, "scryfall", setID, karteID,
			"Erste Probe", "--uebernehmen", "gibtsnicht")
		if code == 0 {
			t.Fatalf("das fremde Feld kam durch: %s", aus)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		setID, karteID := geleerteKarte(t, umgebung)

		text, istFehler := mcpRuf(t, umgebung, "schreiben.scryfall.uebernehmen",
			map[string]any{
				"setID": zahl(t, setID), "karteID": zahl(t, karteID),
				"name": "Erste Probe", "felder": []string{"gibtsnicht"},
			})
		if !istFehler {
			t.Fatalf("das fremde Feld kam durch: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		setID, karteID := geleerteKarte(t, umgebung)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost,
			"/sets/"+setID+"/karten/"+karteID+"/scryfall",
			`{"name":"Erste Probe","felder":["gibtsnicht"]}`)
		if status == http.StatusOK {
			t.Fatalf("das fremde Feld kam durch: %s", rumpf)
		}
	})
}

// geleerteKarte gibt ein Magic-Set und eine Karte darin, deren Regeltext leer
// ist.
//
// **Geleert und nicht neu**, damit die Übernahme etwas zu tun hat: ein Feld,
// das schon den Wert von Scryfall trägt, steht nicht im Vorschlag — und dann
// prüfte der Lauf die Abweisung statt der Übernahme.
func geleerteKarte(t *testing.T, umgebung []string) (setID, karteID string) {
	t.Helper()

	setID = magicset(t, umgebung)

	karten := kartenzeilenAus(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))
	if len(karten) == 0 {
		t.Fatal("das Set hat keine Karten")
	}

	mussKommandozeile(t, umgebung, "feld", setID, karten[0].ID, "rule text", "")

	return setID, karten[0].ID
}

// regeltextPruefen sieht nach, ob der Regeltext von Scryfall dasteht.
//
// **Mit dem Symbolzeichen**, denn genau das unterscheidet einen übernommenen
// Regeltext von einem getippten: `{T}` wird zu `<sym>T</sym>` (#1013).
func regeltextPruefen(t *testing.T, umgebung []string, setID string) {
	t.Helper()

	// **Über den entschlüsselten Wert und nicht über die rohe Ausgabe.** JSON
	// schreibt `<` als `\u003c`; ein `strings.Contains` auf `<sym>` fand
	// deshalb nichts, obwohl der Wert dastand — gesehen, nicht überlegt.
	aus := mussKommandozeile(t, umgebung, "karten", setID, "--json")

	wert := feldwertAus(t, aus, "rule text")
	if !strings.Contains(wert, "<sym>T</sym>") {
		t.Errorf("der Regeltext ist %q — das Symbolzeichen von Scryfall fehlt", wert)
	}
}
