package flaechen_test

// **Angelegt und gewechselt ist geschehen — an jeder Fläche** — §7, Regel 3
// (#1970).
//
// # Was hier gemessen wird
//
// Ein Auslöser verbiegt nach dem Anlegen oder nach einem Wechsel des Stands
// den Zustand der Partie ([imBestand]); geschrieben ist dann, und nur die
// Sicht danach scheitert. **Jede Fläche sagt es und nennt die Kennung** —
// beim Anlegen ist sie der einzige Weg, die eben entstandene Partie
// wiederzufinden, statt eine zweite anzulegen.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// partieOhneSicht ist, was eine Fläche nach einem Nachladefehler sagt — flach
// (Dienst, MCP) oder unter `partie` (Kommandozeile).
type partieOhneSicht struct {
	ID     string `json:"id"`
	Partie struct {
		ID string `json:"id"`
	} `json:"partie"`
	Nachladefehler *struct {
		Code string `json:"code"`
	} `json:"nachladefehler"`
}

// ohneSicht liest eine solche Antwort und gibt die Kennung.
func ohneSicht(t *testing.T, flaeche, roh string) string {
	t.Helper()

	var gelesen partieOhneSicht
	if fehler := json.Unmarshal([]byte(roh), &gelesen); fehler != nil {
		t.Fatalf("%s: kein JSON: %v\n%s", flaeche, fehler, roh)
	}

	kennung := gelesen.ID + gelesen.Partie.ID
	if gelesen.Nachladefehler == nil || gelesen.Nachladefehler.Code != "validierung" ||
		kennung == "" {
		t.Fatalf("%s: %s", flaeche, roh)
	}

	return kennung
}

func TestEinePartieGiltAnJederFlaecheAuchOhneSicht(t *testing.T) {
	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	erster, zweiter := zweiDecks(t, umgebung)

	// **Drei Partien für die Wechsel**, vor dem Auslöser angelegt.
	gewechselt := []string{
		frischePartie(t, umgebung, erster, zweiter),
		frischePartie(t, umgebung, erster, zweiter),
		frischePartie(t, umgebung, erster, zweiter),
	}

	imBestand(t, umgebung, `CREATE TRIGGER krumme_partie AFTER INSERT ON partien
BEGIN UPDATE partien SET zustand = 'krumm' WHERE id = NEW.id; END`)
	imBestand(t, umgebung, `CREATE TRIGGER krummer_stand AFTER UPDATE OF stand ON partien
BEGIN UPDATE partien SET zustand = 'krumm' WHERE id = NEW.id; END`)

	sitze := string(mussJSON(t, sitzwuensche(erster, zweiter)))

	aus, fehlerstrom, code := kommandozeile(t, umgebung, "partie-anlegen",
		"magic", "anna:Anna:"+erster, "bert:Bert:"+zweiter, "--json")
	if code != 0 || !strings.Contains(fehlerstrom, "geschrieben") {
		t.Fatalf("Kommandozeile: %d, %q", code, fehlerstrom)
	}

	angelegt := []string{
		ohneSicht(t, "Kommandozeile", aus),
		ohneSicht(t, "MCP", mussMCP(t, umgebung, "schreiben.partie.anlegen",
			map[string]any{"spiel": "magic", "sitze": sitzwuensche(erster, zweiter)})),
		ohneSicht(t, "Dienst", mussREST(t, an, http.MethodPost, "/partien",
			`{"spiel":"magic","sitze":`+sitze+`}`)),
	}

	for _, kennung := range angelegt {
		if stand := ausDemBestand(t, umgebung,
			`SELECT stand FROM partien WHERE id = ?`, kennung); stand != "laeuft" {
			t.Errorf("die angelegte Partie %s steht bei %q", kennung, stand)
		}
	}

	aus, fehlerstrom, code = kommandozeile(t, umgebung,
		"partie-stand", gewechselt[0], "anna", "pausieren", "--json")
	if code != 0 || !strings.Contains(fehlerstrom, "geschrieben") {
		t.Fatalf("Kommandozeile, Stand: %d, %q", code, fehlerstrom)
	}

	ohneSicht(t, "Kommandozeile, Stand", aus)
	ohneSicht(t, "MCP, Stand", mussMCP(t, umgebung, "schreiben.partie.stand",
		map[string]any{"partieID": gewechselt[1], "fuer": "anna", "stand": "pausieren"}))
	ohneSicht(t, "Dienst, Stand", mussREST(t, an, http.MethodPost,
		"/partien/"+gewechselt[2]+"/stand", `{"fuer":"anna","stand":"pausieren"}`))

	for _, kennung := range gewechselt {
		if stand := ausDemBestand(t, umgebung,
			`SELECT stand FROM partien WHERE id = ?`, kennung); stand != "pausiert" {
			t.Errorf("die Partie %s steht bei %q", kennung, stand)
		}
	}
}
