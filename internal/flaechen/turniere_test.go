// **Der ganze Turnierweg — an jeder Fläche** — Werkstück 5 von Phase 4.7
// (#1294),
// nach Vermerk 0062.
//
// # Die Abnahme aus dem Issue
//
//	„Ein vollständiges kleines Turnier — acht Teilnehmer, drei Runden —
//	lässt sich von der Anlage bis zur Endrangliste durchführen,
//	**zwischendurch schließen und wieder aufnehmen**, ohne dass eine Paarung
//	oder ein Ergebnis verloren geht."
//
// **Das Schließen und Wiederaufnehmen ist hier kein Kunstgriff**: jeder Ruf
// an der Kommandozeile ist ein **eigener Prozess**, der den Bestand öffnet und
// wieder schließt. Wer acht Teilnehmer anmeldet und drei Runden spielt, hat
// die Anwendung dabei zwanzigmal beendet und neu gestartet.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

// **Der ganze Weg über die Kommandozeile — acht Teilnehmer, drei Runden.**
func TestDerGanzeTurnierwegAnDerKommandozeile(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)

	turnierID := strings.TrimSpace(mussKommandozeile(t, umgebung,
		"turnier", "neu", "Küchentisch", "magic", "Freitagabend"))

	for _, name := range []string{
		"Ada", "Bert", "Cem", "Dana", "Eike", "Ferry", "Gil", "Hana",
	} {
		mussKommandozeile(t, umgebung, "turnier", "anmelden", turnierID, name)
	}

	// **Drei Runden** — die Zahl aus Appendix E der MTR für 5–8 Spieler.
	for runde := 1; runde <= 3; runde++ {
		// **Erst hinsehen, dann legen.** Der Vorschlag schafft keine
		// Tatsachen; das ist der ganze Unterschied.
		vorschlag := mussKommandozeile(t, umgebung, "turnier", "vorschlag", turnierID)
		if strings.Count(vorschlag, "\n") != 4 {
			t.Fatalf("Runde %d: der Vorschlag nennt %d Paarungen, erwartet 4\n%s",
				runde, strings.Count(vorschlag, "\n"), vorschlag)
		}

		mussKommandozeile(t, umgebung, "turnier", "runde", turnierID)

		for _, paarung := range offenePaarungen(t, umgebung, turnierID) {
			mussKommandozeile(t, umgebung, "turnier", "ergebnis",
				turnierID, paarung, "2", "0")
		}
	}

	rangliste := mussKommandozeile(t, umgebung, "turnier", "rangliste",
		turnierID, "--json")

	var gelesen struct {
		Rangliste []struct {
			Platz       int    `json:"platz"`
			Name        string `json:"name"`
			Matchpunkte int    `json:"matchpunkte"`
		} `json:"rangliste"`
	}

	if err := json.Unmarshal([]byte(rangliste), &gelesen); err != nil {
		t.Fatalf("die Rangliste ist kein gültiges JSON: %v (%s)", err, rangliste)
	}

	if len(gelesen.Rangliste) != 8 {
		t.Fatalf("%d Ränge, erwartet 8", len(gelesen.Rangliste))
	}

	// **Genau einer hat alles gewonnen** — bei acht Teilnehmern und drei
	// Runden im Schweizer System bleibt einer ohne Niederlage.
	if gelesen.Rangliste[0].Matchpunkte != 9 {
		t.Errorf("Platz 1 hat %d Punkte, erwartet 9: %+v",
			gelesen.Rangliste[0].Matchpunkte, gelesen.Rangliste[0])
	}

	// **Und die Plätze sind lückenlos vergeben.**
	for stelle, rang := range gelesen.Rangliste {
		if rang.Platz != stelle+1 {
			t.Errorf("Rang %d trägt Platz %d (%s)", stelle+1, rang.Platz, rang.Name)
		}
	}
}

// offenePaarungen gibt die Kennungen der Paarungen ohne Ergebnis.
//
// **Byes fallen weg** — sie sind schon ein Ergebnis (2–0, Appendix C) und
// warten auf nichts.
func offenePaarungen(t *testing.T, umgebung []string, turnierID string) []string {
	t.Helper()

	roh := mussKommandozeile(t, umgebung, "turnier", turnierID, "--json")

	var gelesen struct {
		Turnier struct {
			Runden []struct {
				Paarungen []struct {
					ID       string `json:"id"`
					B        string `json:"b"`
					Ergebnis *struct {
						SpieleA int `json:"spieleA"`
					} `json:"ergebnis"`
				} `json:"paarungen"`
			} `json:"runden"`
		} `json:"turnier"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("das Turnier ist kein gültiges JSON: %v (%s)", err, roh)
	}

	var offen []string

	for _, runde := range gelesen.Turnier.Runden {
		for _, paarung := range runde.Paarungen {
			if paarung.B != "" && paarung.Ergebnis == nil {
				offen = append(offen, paarung.ID)
			}
		}
	}

	return offen
}

// **Und derselbe Ablauf über MCP.**
func TestDerGanzeTurnierwegUeberMCP(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)

	turnierID := kennungAusMCP(t, umgebung, "schreiben.turnier.anlegen",
		map[string]any{"name": "Küchentisch", "spiel": "magic"})

	for _, name := range []string{"Ada", "Bert", "Cem"} {
		mussMCP(t, umgebung, "schreiben.turnier.anmelden",
			map[string]any{"turnierID": turnierID, "name": name})
	}

	// **Drei Teilnehmer heißen zwei Paarungen** — eine davon ein Bye.
	vorschlag := mussMCP(t, umgebung, "schreiben.turnier.vorschlag",
		map[string]any{"turnierID": turnierID})

	var paarungen []struct {
		B string `json:"b"`
	}

	if err := json.Unmarshal([]byte(vorschlag), &paarungen); err != nil {
		t.Fatalf("der Vorschlag ist kein gültiges JSON: %v (%s)", err, vorschlag)
	}

	if len(paarungen) != 2 || paarungen[1].B != "" {
		t.Fatalf("der Vorschlag ist %+v — erwartet zwei, die zweite ein Bye",
			paarungen)
	}

	mussMCP(t, umgebung, "schreiben.turnier.runde",
		map[string]any{"turnierID": turnierID})

	// **Das Ergebnis eintragen und wieder zurücknehmen.**
	//
	// **Die Kennung wird zur Zahl.** MCP nimmt Kennungen als Zahlen, die
	// Antwort gibt sie als Text — dieselbe Naht, zwei Gestalten, und das ist
	// Absicht: JavaScript kennt nur `float64`, MCP nicht.
	paarung, wandelfehler := strconv.ParseInt(
		ersteOffenePaarungMCP(t, umgebung, turnierID), 10, 64)
	if wandelfehler != nil {
		t.Fatalf("die Paarungskennung: %v", wandelfehler)
	}

	mussMCP(t, umgebung, "schreiben.turnier.ergebnis", map[string]any{
		"turnierID": turnierID, "paarungID": paarung, "spieleA": 2, "spieleB": 0,
	})

	// **Zwei mit drei Punkten**: der Sieger und der mit dem Bye.
	//
	// **Gelesen und nicht in der Zeichenkette gesucht.** MCP setzt sein JSON
	// mit Einrückung; ein `strings.Contains` auf `"matchpunkte":3` fände
	// nichts und läse sich wie ein Befund.
	if mitDreiPunkten(t, umgebung, turnierID) != 2 {
		t.Errorf("nicht zwei Teilnehmer mit drei Punkten")
	}

	mussMCP(t, umgebung, "schreiben.turnier.ergebnis", map[string]any{
		"turnierID": turnierID, "paarungID": paarung, "zuruecknehmen": true,
	})

	// **Nach der Rücknahme bleibt nur der Bye** — der ist ein Ergebnis und
	// keine Eintragung.
	if mitDreiPunkten(t, umgebung, turnierID) != 1 {
		t.Errorf("nach der Rücknahme steht der Sieg noch da")
	}

	// **Abmelden über `schreiben.turnier.aussteigen`** — und er bleibt in
	// der Rangliste stehen.
	teilnehmer := kennungAusMCP(t, umgebung, "schreiben.turnier.anmelden",
		map[string]any{"turnierID": turnierID, "name": "Dana"})

	mussMCP(t, umgebung, "schreiben.turnier.aussteigen",
		map[string]any{
			"turnierID": turnierID, "teilnehmerID": teilnehmer,
			"ausgestiegen": true,
		})

	stand := mussMCP(t, umgebung, "lesen.turnier.rangliste",
		map[string]any{"turnierID": turnierID})

	if !strings.Contains(stand, "Dana") {
		t.Errorf("der Ausgestiegene fehlt in der Rangliste: %s", stand)
	}

	// **Und `schreiben.turnier.loeschen` nimmt den Verlauf mit.**
	mussMCP(t, umgebung, "schreiben.turnier.loeschen",
		map[string]any{"turnierID": turnierID})

	if uebrig := mussMCP(t, umgebung, "lesen.turniere",
		map[string]any{}); strings.Contains(uebrig, "Küchentisch") {
		t.Errorf("das Turnier steht noch da: %s", uebrig)
	}
}

// mitDreiPunkten zählt die Teilnehmer mit drei Matchpunkten.
func mitDreiPunkten(t *testing.T, umgebung []string, turnierID int64) int {
	t.Helper()

	roh := mussMCP(t, umgebung, "lesen.turnier.rangliste",
		map[string]any{"turnierID": turnierID})

	var rangliste []struct {
		Matchpunkte int `json:"matchpunkte"`
	}

	if err := json.Unmarshal([]byte(roh), &rangliste); err != nil {
		t.Fatalf("die Rangliste ist kein gültiges JSON: %v (%s)", err, roh)
	}

	var mit int

	for _, rang := range rangliste {
		if rang.Matchpunkte == 3 {
			mit++
		}
	}

	return mit
}

// ersteOffenePaarungMCP gibt die erste Paarung ohne Ergebnis.
func ersteOffenePaarungMCP(t *testing.T, umgebung []string, turnierID int64) string {
	t.Helper()

	roh := mussMCP(t, umgebung, "lesen.turnier",
		map[string]any{"turnierID": turnierID})

	var turnier struct {
		Runden []struct {
			Paarungen []struct {
				ID       string    `json:"id"`
				B        string    `json:"b"`
				Ergebnis *struct{} `json:"ergebnis"`
			} `json:"paarungen"`
		} `json:"runden"`
	}

	if err := json.Unmarshal([]byte(roh), &turnier); err != nil {
		t.Fatalf("das Turnier ist kein gültiges JSON: %v (%s)", err, roh)
	}

	for _, runde := range turnier.Runden {
		for _, paarung := range runde.Paarungen {
			if paarung.B != "" && paarung.Ergebnis == nil {
				return paarung.ID
			}
		}
	}

	t.Fatalf("keine offene Paarung: %s", roh)

	return ""
}

// **Und über REST.**
func TestDerGanzeTurnierwegUeberREST(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)

	turnierID := kennungAusREST(t, an, "/turniere",
		`{"name":"Küchentisch","spiel":"magic"}`)

	for _, name := range []string{"Ada", "Bert"} {
		mussREST(t, an, http.MethodPost, "/turniere/"+turnierID+"/teilnehmer",
			`{"name":"`+name+`"}`)
	}

	// **Der Vorschlag ist ein `GET`** — er legt nichts an.
	mussREST(t, an, http.MethodGet, "/turniere/"+turnierID+"/vorschlag", "")

	runde := mussREST(t, an, http.MethodPost, "/turniere/"+turnierID+"/runden", "")

	var gelegt struct {
		Paarungen []struct {
			ID string `json:"id"`
		} `json:"paarungen"`
	}

	if err := json.Unmarshal([]byte(runde), &gelegt); err != nil {
		t.Fatalf("die Runde ist kein gültiges JSON: %v (%s)", err, runde)
	}

	if len(gelegt.Paarungen) != 1 {
		t.Fatalf("%d Paarungen, erwartet eine", len(gelegt.Paarungen))
	}

	mussREST(t, an, http.MethodPut,
		"/turniere/"+turnierID+"/paarungen/"+gelegt.Paarungen[0].ID,
		`{"spieleA":2,"spieleB":0}`)

	rangliste := mussREST(t, an, http.MethodGet,
		"/turniere/"+turnierID+"/rangliste", "")

	if !strings.Contains(rangliste, `"matchpunkte":3`) {
		t.Errorf("die Rangliste nennt keine drei Punkte: %s", rangliste)
	}

	// **Und der Ausstieg lässt ihn in der Rangliste stehen.**
	var stand []struct {
		TeilnehmerID string `json:"teilnehmerID"`
	}

	if err := json.Unmarshal([]byte(rangliste), &stand); err != nil {
		t.Fatalf("die Rangliste ist kein gültiges JSON: %v (%s)", err, rangliste)
	}

	mussREST(t, an, http.MethodPut,
		"/turniere/"+turnierID+"/teilnehmer/"+stand[1].TeilnehmerID,
		`{"ausgestiegen":true}`)

	nachher := mussREST(t, an, http.MethodGet,
		"/turniere/"+turnierID+"/rangliste", "")

	if !strings.Contains(nachher, `"ausgestiegen":true`) {
		t.Errorf("der Ausgestiegene fehlt in der Rangliste: %s", nachher)
	}

	// **Und das Löschen nimmt den Verlauf mit.**
	mussREST(t, an, http.MethodDelete, "/turniere/"+turnierID, "")

	uebrig := mussREST(t, an, http.MethodGet, "/turniere", "")
	if strings.Contains(uebrig, "Küchentisch") {
		t.Errorf("das Turnier steht noch da: %s", uebrig)
	}
}
