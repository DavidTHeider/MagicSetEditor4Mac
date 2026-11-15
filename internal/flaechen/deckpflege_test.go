// **Der ganze Deckweg — an jeder Fläche** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei**, damit `decks_test.go` unter der Zeilengrenze aus §2
// bleibt. Der Schnitt folgt der Frage: dort steht, dass **etwas entsteht**,
// hier, dass der **ganze Ablauf** über jede Fläche trägt.
//
// # Der Ablauf
//
// Deck anlegen → umbenennen → Karte legen → Format anlegen → Deck darauf
// stellen → prüfen → Bestand abgleichen → Platz nehmen → Deck löschen.
//
// **Neun Schritte, drei Flächen, dasselbe Ergebnis.** Das ist die Zusage aus
// §16, letzter Absatz — und mehr als „die Fähigkeit ist da".
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// **Der ganze Weg über die Kommandozeile.**
func TestDerGanzeDeckwegAnDerKommandozeile(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)

	deckID := strings.TrimSpace(mussKommandozeile(t, umgebung,
		"deck", "neu", "Aggro", "magic"))

	// **Die Fassung kommt aus dem Lesen** — sie ist die Bedingung.
	fassung := deckfassung(t, mussKommandozeile(t, umgebung, "deck", deckID, "--json"))

	mussKommandozeile(t, umgebung, "deck", "umbenennen", deckID, fassung, "Midrange")

	platzID := strings.TrimSpace(mussKommandozeile(t, umgebung, "deck", "legen",
		deckID, druckkennung, "4"))

	formatID := strings.TrimSpace(mussKommandozeile(t, umgebung,
		"format", "neu", "konstruiert", "Konstruiert", "magic"))

	mussKommandozeile(t, umgebung, "deck", "format", deckID, formatID)

	// **Ein Deck mit vier Karten ist in einem 60-Karten-Format unzulässig** —
	// und das ist ein fachliches Nein, kein Fehler.
	_, _, code := kommandozeile(t, umgebung, "deck", "pruefen", deckID)
	if code != 1 {
		t.Errorf("prüfen gab %d, erwartet 1 (unzulässig)", code)
	}

	// **Und was fehlt, fehlt** — die Sammlung ist leer.
	_, _, code = kommandozeile(t, umgebung, "deck", "bestand", deckID)
	if code != 1 {
		t.Errorf("bestand gab %d, erwartet 1 (es fehlt)", code)
	}

	mussKommandozeile(t, umgebung, "deck", "nehmen", deckID, platzID)
	mussKommandozeile(t, umgebung, "deck", "loeschen", deckID)
	mussKommandozeile(t, umgebung, "format", "loeschen", formatID)
}

// **Derselbe Weg über MCP.**
func TestDerGanzeDeckwegUeberMCP(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)

	deckID := kennungAusMCP(t, umgebung, "schreiben.deck.anlegen",
		map[string]any{"name": "Aggro", "spiel": "magic"})

	mussMCP(t, umgebung, "schreiben.deck.umbenennen", map[string]any{
		"deckID": deckID, "fassung": 1, "name": "Midrange",
	})

	platz := mussMCP(t, umgebung, "schreiben.deckplatz.legen", map[string]any{
		"deckID": deckID, "printing": druckkennung,
		"menge": 4, "bereich": "hauptdeck",
	})

	formatID := kennungAusMCP(t, umgebung, "schreiben.format.ausbauart",
		map[string]any{"bauart": "konstruiert", "name": "Konstruiert", "spiel": "magic"})

	mussMCP(t, umgebung, "schreiben.deck.format",
		map[string]any{"deckID": deckID, "formatID": formatID})

	// **Die Prüfung glückt und sagt „unzulässig"** — ein Fehlschlag wäre
	// etwas anderes als ein Nein.
	pruefeUnzulaessig(t, mussMCP(t, umgebung, "lesen.deck.pruefung",
		map[string]any{"deckID": deckID}))

	mussMCP(t, umgebung, "lesen.deck.bestand", map[string]any{"deckID": deckID})

	mussMCP(t, umgebung, "schreiben.deckplatz.nehmen", map[string]any{
		"deckID": deckID, "platzID": platzkennung(t, platz),
	})

	mussMCP(t, umgebung, "schreiben.deck.loeschen", map[string]any{"deckID": deckID})
	mussMCP(t, umgebung, "schreiben.format.loeschen",
		map[string]any{"formatID": formatID})
	mussMCP(t, umgebung, "lesen.formatbauarten", map[string]any{})
}

// **Und derselbe Weg über den Dienst.**
func TestDerGanzeDeckwegUeberREST(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)

	deckID := kennungAusREST(t, an, "/decks", `{"name":"Aggro","spiel":"magic"}`)

	mussREST(t, an, http.MethodPut, "/decks/"+deckID,
		`{"fassung":1,"name":"Midrange","notiz":""}`)

	platz := mussREST(t, an, http.MethodPut, "/decks/"+deckID+"/plaetze",
		`{"printing":"`+druckkennung+`","menge":4,"bereich":"hauptdeck"}`)

	formatID := kennungAusREST(t, an, "/formate",
		`{"bauart":"konstruiert","name":"Konstruiert","spiel":"magic"}`)

	mussREST(t, an, http.MethodPut, "/decks/"+deckID+"/format",
		`{"formatID":`+formatID+`}`)

	pruefeUnzulaessig(t, mussREST(t, an, http.MethodGet,
		"/decks/"+deckID+"/pruefung", ""))

	mussREST(t, an, http.MethodGet, "/decks/"+deckID+"/bestand", "")
	mussREST(t, an, http.MethodGet, "/formatbauarten", "")

	mussREST(t, an, http.MethodDelete,
		"/decks/"+deckID+"/plaetze/"+platzkennungText(t, platz), "")
	mussREST(t, an, http.MethodDelete, "/decks/"+deckID, "")
	mussREST(t, an, http.MethodDelete, "/formate/"+formatID, "")
}

// druckkennung ist eine Scryfall-Kennung in der richtigen Gestalt.
//
// **Selbst erfunden und nicht bezogen** (§11): geprüft wird die Gestalt des
// Wegs, nicht der Inhalt einer fremden Antwort.
const druckkennung = "a1b2c3d4-0000-4000-8000-00000000000a"

// mussMCP ruft ein Werkzeug und bricht ab, wenn es nicht glückt.
func mussMCP(
	t *testing.T, umgebung []string, werkzeug string, argumente map[string]any,
) string {
	t.Helper()

	roh, istFehler := mcpRuf(t, umgebung, werkzeug, argumente)
	if istFehler {
		t.Fatalf("%s: %s", werkzeug, roh)
	}

	return roh
}

// mussREST ruft einen Weg und bricht ab, wenn er nicht glückt.
func mussREST(t *testing.T, an dienst, methode, weg, rumpf string) string {
	t.Helper()

	status, antwort := an.ruf(t, methode, weg, rumpf)
	if status < 200 || status >= 300 {
		t.Fatalf("%s %s: %d (%s)", methode, weg, status, antwort)
	}

	return antwort
}

// deckfassung liest die Fassung aus einer Deckantwort.
func deckfassung(t *testing.T, roh string) string {
	t.Helper()

	var gelesen struct {
		Deck struct {
			Fassung string `json:"fassung"`
		} `json:"deck"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	return gelesen.Deck.Fassung
}

// platzkennung liest die Kennung eines gelegten Platzes.
func platzkennung(t *testing.T, roh string) int64 {
	t.Helper()

	return zahl(t, platzkennungText(t, roh))
}

// platzkennungText tut dasselbe und gibt sie als Text.
func platzkennungText(t *testing.T, roh string) string {
	t.Helper()

	var gelesen struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	return gelesen.ID
}

// pruefeUnzulaessig sichert zu, dass die Prüfung ein Nein gibt.
//
// **Vier Karten in einem Format mit Mindestgröße 60** — der Größenbefund
// steht, und `zulaessig` ist falsch.
func pruefeUnzulaessig(t *testing.T, roh string) {
	t.Helper()

	var gelesen struct {
		Urteil struct {
			Zulaessig bool `json:"zulaessig"`
			Groessen  []struct {
				Feld string `json:"feld"`
			} `json:"groessen"`
		} `json:"urteil"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	if gelesen.Urteil.Zulaessig {
		t.Errorf("vier Karten galten in einem 60-Karten-Format als zulässig: %s", roh)
	}

	if len(gelesen.Urteil.Groessen) == 0 {
		t.Errorf("kein Größenbefund: %s", roh)
	}
}

// deckumgebung gibt eine Umgebung mit **eigenem Zuhause und umgelenkter
// Adresse**.
//
// # Warum beides
//
// **§3, seit #888**: „Zwei Dinge gehören an jeden solchen Test:
// `t.Setenv(\"HOME\", t.TempDir())` … **und** eine umgelenkte Adresse. Beides
// einzeln reicht nicht."
//
// `deck pruefen` fragt Scryfall nach jedem Druck. Ohne die Umlenkung ginge
// jeder dieser Läufe wirklich hinaus — und wäre grün, weil er funktioniert
// hat.
//
// **Die Gegenstelle kennt den erfundenen Druck nicht**, und das ist richtig:
// die Prüfung meldet ihn dann als `ungeprüft`, und die Zusage dieses Laufs ist
// der **Größenbefund**, den sie ohne jede Gegenstelle rechnet.
func deckumgebung(t *testing.T) []string {
	t.Helper()

	return scryfallumgebung(t, scryfallgegenstelle(t))
}

// **Ein Format lässt sich an jeder Fläche fortschreiben.**
//
// **Ein eigener Lauf**, weil `format setzen` etwas kann, das `format neu`
// nicht kann: eine Zahl ändern. Und weil er die eine Falle zeigt, die dieser
// Weg hat — **null heißt bei drei Zahlen „keine Grenze"**, und wer eine davon
// vergisst, hat sie aufgehoben.
func TestEinFormatLaesstSichUeberallFortschreiben(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := deckumgebung(t)

		formatID := strings.TrimSpace(mussKommandozeile(t, umgebung,
			"format", "neu", "konstruiert", "Konstruiert", "magic"))

		mussKommandozeile(t, umgebung, "format", "setzen",
			formatID, "1", "Konstruiert", "magic", "40", "0", "15", "4")

		pruefeMindestgroesse(t,
			mussKommandozeile(t, umgebung, "formate", "--json"), 40)
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := deckumgebung(t)

		formatID := kennungAusMCP(t, umgebung, "schreiben.format.ausbauart",
			map[string]any{"bauart": "konstruiert", "name": "Konstruiert", "spiel": "magic"})

		// **Nur die eine Zahl** — was nicht gesagt wird, bleibt stehen.
		mussMCP(t, umgebung, "schreiben.format.setzen", map[string]any{
			"id": formatID, "fassung": 1, "mindestgroesse": 40,
		})

		roh := mussMCP(t, umgebung, "lesen.formate", map[string]any{})

		pruefeMindestgroesse(t, `{"formate":`+roh+`}`, 40)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := deckumgebung(t)
		an := starteDienst(t, umgebung)

		formatID := kennungAusREST(t, an, "/formate",
			`{"bauart":"konstruiert","name":"Konstruiert","spiel":"magic"}`)

		mussREST(t, an, http.MethodPut, "/formate/"+formatID,
			`{"fassung":1,"mindestgroesse":40}`)

		rumpf := mussREST(t, an, http.MethodGet, "/formate", "")

		pruefeMindestgroesse(t, `{"formate":`+rumpf+`}`, 40)
	})
}

// pruefeMindestgroesse sichert die Untergrenze — und dass der Rest steht.
//
// **Beides**: eine Änderung, die nebenbei das Sideboard aufhebt, wäre grün,
// wenn nur die geänderte Zahl geprüft würde.
func pruefeMindestgroesse(t *testing.T, roh string, erwartet int) {
	t.Helper()

	var gelesen struct {
		Formate []struct {
			Name              string `json:"name"`
			Mindestgroesse    int    `json:"mindestgroesse"`
			Sideboardgroesse  int    `json:"sideboardgroesse"`
			HoechstzahlGleich int    `json:"hoechstzahlGleich"`
		} `json:"formate"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	if len(gelesen.Formate) != 1 {
		t.Fatalf("%d Formate, erwartet eines (%s)", len(gelesen.Formate), roh)
	}

	format := gelesen.Formate[0]

	if format.Mindestgroesse != erwartet {
		t.Errorf("Mindestgröße %d, erwartet %d", format.Mindestgroesse, erwartet)
	}

	// **Und die Bauart steht noch** — CR 100.4a und 100.2a.
	if format.Sideboardgroesse != 15 || format.HoechstzahlGleich != 4 {
		t.Errorf("die übrigen Zahlen sind mitgegangen: %+v", format)
	}

	if format.Name != "Konstruiert" {
		t.Errorf("Name %q", format.Name)
	}
}
