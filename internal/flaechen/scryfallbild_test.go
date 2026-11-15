// Bild und Sammeldaten über jede Fläche
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

// **Ein Kartenbild einsetzen — die Abweisung über jede Fläche.**
//
// # Warum hier die Abweisung steht und nicht der Erfolg
//
// Der Bildweg braucht die **Fläche**, auf die zugeschnitten wird, und die
// rechnet der Planbauer aus der Vorlage aus (`app.feldgroesse`,
// `render.Plan.Feldflaechen`). Eine Vorlage liegt in diesem Repositorium
// nicht — fremde Pakete werden nicht eingecheckt (§11), und `saat/` führt das
// Spiel »probe«, für das es keine Scryfall-Zuordnung gibt.
//
// **Was dieser Lauf also hält**: dass alle drei Flächen denselben Fall
// gleich beantworten — sie weisen ab, keine schreibt still ein halbes
// Bildfeld, und keine bricht auf eine andere Art ab.
//
// **Was er nicht hält**: den Erfolgsweg. Der läuft hinter dem Tag
// `fremdpakete` gegen ein bezogenes Paket, und im Fenster ist er von Hand
// nachgesehen worden (#1025). Ein Lauf, der so täte, als prüfte er ihn, wäre
// schlimmer als keiner.
func TestEinKartenbildOhneVorlageWirdUeberallAbgewiesen(t *testing.T) {
	gegenstelle := scryfallgegenstelle(t)

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		setID, karteID := geleerteKarte(t, umgebung)

		aus, fehler, code := kommandozeile(t, umgebung, "scryfall-bild", setID, karteID,
			"image", "Erste Probe")
		if code == 0 {
			t.Fatalf("das Bild kam ohne Vorlage durch: %s", aus)
		}

		if !strings.Contains(aus+fehler, "Bild") {
			t.Errorf("die Meldung sagt nicht, worum es ging: %s%s", aus, fehler)
		}

		bildfeldLeerPruefen(t, umgebung, setID)
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		setID, karteID := geleerteKarte(t, umgebung)

		text, istFehler := mcpRuf(t, umgebung, "schreiben.scryfall.bild",
			map[string]any{
				"setID": zahl(t, setID), "karteID": zahl(t, karteID),
				"feld": "image", "name": "Erste Probe", "art": "normal",
			})
		if !istFehler {
			t.Fatalf("das Bild kam ohne Vorlage durch: %s", text)
		}

		bildfeldLeerPruefen(t, umgebung, setID)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		setID, karteID := geleerteKarte(t, umgebung)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost,
			"/sets/"+setID+"/karten/"+karteID+"/scryfall/bild",
			`{"feld":"image","name":"Erste Probe","art":"normal"}`)
		if status == http.StatusOK || status == http.StatusCreated {
			t.Fatalf("das Bild kam ohne Vorlage durch: %s", rumpf)
		}

		bildfeldLeerPruefen(t, umgebung, setID)
	})
}

// **Und eine Bildart, die es nicht gibt, wird überall abgewiesen.**
func TestEineUnbekannteBildartWirdUeberallAbgewiesen(t *testing.T) {
	gegenstelle := scryfallgegenstelle(t)

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		setID, karteID := geleerteKarte(t, umgebung)

		aus, _, code := kommandozeile(t, umgebung, "scryfall-bild", setID, karteID,
			"image", "Erste Probe", "--art", "gibtsnicht")
		if code == 0 {
			t.Fatalf("die unbekannte Bildart kam durch: %s", aus)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		setID, karteID := geleerteKarte(t, umgebung)

		text, istFehler := mcpRuf(t, umgebung, "schreiben.scryfall.bild",
			map[string]any{
				"setID": zahl(t, setID), "karteID": zahl(t, karteID),
				"feld": "image", "name": "Erste Probe", "art": "gibtsnicht",
			})
		if !istFehler {
			t.Fatalf("die unbekannte Bildart kam durch: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		setID, karteID := geleerteKarte(t, umgebung)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost,
			"/sets/"+setID+"/karten/"+karteID+"/scryfall/bild",
			`{"feld":"image","name":"Erste Probe","art":"gibtsnicht"}`)
		if status == http.StatusOK || status == http.StatusCreated {
			t.Fatalf("die unbekannte Bildart kam durch: %s", rumpf)
		}
	})
}

// **Die Sammeldatei holen — über jede Fläche.**
//
// # Was sie ändert
//
// **Keine Karte und kein Set** — sie legt eine Datei ab, aus der die
// Scryfall-Wege ohne Netz antworten. Geprüft ist deshalb die abgelegte Größe,
// die jede Fläche zurückmeldet.
func TestDieSammeldateiHolenUeberJedeFlaeche(t *testing.T) {
	gegenstelle := scryfallgegenstelle(t)

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)

		aus := mussKommandozeile(t, umgebung, "scryfall-sammeldaten", "holen", "--json")
		if !strings.Contains(aus, "oracle_cards") {
			t.Errorf("die Antwort nennt die Art nicht: %s", aus)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)

		text, istFehler := mcpRuf(t, umgebung, "schreiben.scryfall.sammeldaten",
			map[string]any{})
		if istFehler {
			t.Fatalf("schreiben.scryfall.sammeldaten: %s", text)
		}

		if !strings.Contains(text, "oracle_cards") {
			t.Errorf("die Antwort nennt die Art nicht: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := scryfallumgebung(t, gegenstelle)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/scryfall/sammeldaten", "")
		if status != http.StatusOK {
			t.Fatalf("POST /scryfall/sammeldaten: %d (%s)", status, rumpf)
		}

		if !strings.Contains(rumpf, "oracle_cards") {
			t.Errorf("die Antwort nennt die Art nicht: %s", rumpf)
		}
	})
}

// bildfeldLeerPruefen hält fest, dass eine Abweisung nichts geschrieben hat.
//
// **Das ist der Teil, den die bloße Fehlermeldung nicht sagt**: eine Fläche,
// die abwiese und trotzdem den Künstler nach `illustrator` schriebe, hätte
// einen halben Stand hinterlassen.
func bildfeldLeerPruefen(t *testing.T, umgebung []string, setID string) {
	t.Helper()

	aus := mussKommandozeile(t, umgebung, "karten", setID, "--json")

	if bild := feldwertAus(t, aus, "image"); bild != "" {
		t.Errorf("das Bildfeld trägt %q, obwohl der Aufruf abgewiesen wurde", bild)
	}
}
