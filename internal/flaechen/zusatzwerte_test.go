package flaechen_test

// **Ein Zusatzwert geht über Kommandozeile, MCP und Dienst** (#2298).
//
// # Was er hält
//
// **Derselbe Ablauf an drei Flächen über einem Bestand**: jede setzt den
// Wert von `vermerk`, dem Zusatzfeld des Saatstils, und die **nächste** liest
// ihn — so zeigt sich, dass alle drei denselben Block unter `extra_data:`
// schreiben und lesen. Und **dasselbe Nein**: ein gerechnetes Zusatzfeld
// (`grund`, `save value: false`) lässt sich an keiner setzen.
//
// # Was er nicht hält
//
// **Das Fenster** — dieser Lauf startet Prozesse; die Bindung prüft
// `internal/app`.

import (
	"net/http"
	"strings"
	"testing"
)

func TestEinZusatzwertGehtUeberDreiFlaechen(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := leeresSet(t, umgebung, "Zusatz")

	karte := struct{ ID string }{ID: kennungAusFeld(t, mussKommandozeile(t, umgebung,
		"karte", setID, "anlegen", "--json"))}

	an := starteDienst(t, umgebung)
	ort := "/sets/" + setID + "/karten/" + karte.ID + "/zusatzfelder"
	kennungen := map[string]any{"setID": zahl(t, setID), "karteID": zahl(t, karte.ID)}

	mitFeld := func(feld, wert string) map[string]any {
		return map[string]any{"setID": kennungen["setID"], "karteID": kennungen["karteID"], "feld": feld, "wert": wert}
	}

	// **Die Kommandozeile setzt, MCP liest.**
	mussKommandozeile(t, umgebung, "zusatzfeld", setID, karte.ID, "setzen", "vermerk", "von der Kommandozeile")

	if antwort := mussMCP(t, umgebung, "lesen.zusatzfelder", kennungen); !strings.Contains(antwort, "von der Kommandozeile") {
		t.Errorf("MCP liest %s", antwort)
	}

	// **MCP setzt, der Dienst liest.**
	mussMCP(t, umgebung, "schreiben.zusatzwert", mitFeld("vermerk", "über MCP"))

	if antwort := mussREST(t, an, http.MethodGet, ort, ""); !strings.Contains(antwort, "über MCP") {
		t.Errorf("REST liest %s", antwort)
	}

	// **Der Dienst setzt, die Kommandozeile liest.**
	mussREST(t, an, http.MethodPut, ort+"/vermerk", `{"wert":"vom Dienst"}`)

	if aus := mussKommandozeile(t, umgebung, "zusatzfeld", setID, karte.ID); !strings.Contains(aus, "vom Dienst") {
		t.Errorf("die Kommandozeile liest %s", aus)
	}

	// **Und dasselbe Nein an allen drei.**
	if _, _, code := kommandozeile(t, umgebung, "zusatzfeld", setID, karte.ID, "setzen", "grund", "x"); code == 0 {
		t.Error("die Kommandozeile setzte ein gerechnetes Feld")
	}

	if text, istFehler := mcpRuf(t, umgebung, "schreiben.zusatzwert", mitFeld("grund", "x")); !istFehler {
		t.Errorf("MCP setzte ein gerechnetes Feld: %s", text)
	}

	if status, antwort := an.ruf(t, http.MethodPut, ort+"/grund", `{"wert":"x"}`); status != http.StatusBadRequest {
		t.Errorf("REST setzte ein gerechnetes Feld: %d %s", status, antwort)
	}

	// **Ein fehlender Wert ist keine Aussage** — an keiner Fläche leert er
	// das Feld (Befund 5 des Nachlaufs über E/F).
	ohneWertBleibtEr(t, umgebung, an, setID, karte.ID, ort, kennungen)
}

// ohneWertBleibtEr schickt das Setzen ohne Wert über alle drei Flächen und sieht
// danach nach, dass der vorige Wert noch steht.
func ohneWertBleibtEr(t *testing.T, umgebung []string, an dienst, setID, karteID, ort string, kennungen map[string]any) {
	t.Helper()

	if _, _, code := kommandozeile(t, umgebung, "zusatzfeld", setID, karteID, "setzen", "vermerk"); code == 0 {
		t.Error("die Kommandozeile setzte ohne Wert")
	}

	ohneWert := map[string]any{"setID": kennungen["setID"], "karteID": kennungen["karteID"], "feld": "vermerk"}
	if text, istFehler := mcpRuf(t, umgebung, "schreiben.zusatzwert", ohneWert); !istFehler {
		t.Errorf("MCP setzte ohne Wert: %s", text)
	}

	for _, rumpf := range []string{`{}`, `{"wert":null}`} {
		if status, antwort := an.ruf(t, http.MethodPut, ort+"/vermerk", rumpf); status != http.StatusBadRequest ||
			!strings.Contains(antwort, `"validierung"`) {
			t.Errorf("REST mit %s: %d %s", rumpf, status, antwort)
		}
	}

	if aus := mussKommandozeile(t, umgebung, "zusatzfeld", setID, karteID); !strings.Contains(aus, "vom Dienst") {
		t.Errorf("nach den Rufen ohne Wert steht %s", aus)
	}
}
