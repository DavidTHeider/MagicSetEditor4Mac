package flaechen_test

// **Gesetzt ist gesetzt — an jeder Fläche** — §7, Regel 3
// (#1970).
//
// # Was hier gemessen wird
//
// Ein Format, dessen offenes Sideboard weder ja noch nein ist, lässt sich
// nicht lesen. **Das Setzen einer Deckvorgabe fragt es nicht, das Nachlesen
// des ganzen Stands schon** — ohne Auslöser, nur mit einer krummen Zeile.
// Jede Fläche sagt dann: gesetzt, und warum der Stand fehlt — als
// `unbekannt`, denn ein Feld, das der Bestand nicht wandeln kann, ist kein
// Befund über eine Eingabe (§8).

import (
	"net/http"
	"strings"
	"testing"
)

func TestGesetztIstGesetztAnJederFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	imBestand(t, umgebung, `INSERT INTO decks (profil_id, name, spiel)
		SELECT id, 'Mono Rot', 'magic' FROM profile WHERE id = 1`)
	imBestand(t, umgebung, `INSERT INTO formate (profil_id, name, spiel, sideboard_offen)
		VALUES (1, 'Krumm', 'magic', 2)`)

	deck := ausDemBestand(t, umgebung, `SELECT CAST(id AS TEXT) FROM decks WHERE name = 'Mono Rot'`)

	aus, fehlerstrom, code := mitEingabeUndCode(t, umgebung, `["`+deck+`"]`,
		"vorgabe-setzen", "1", "deck", "--json")
	zeile := nachgeladenAus(t, aus)

	if code != 0 || !strings.Contains(fehlerstrom, "geschrieben") ||
		zeile.Nachladefehler == nil || zeile.Nachladefehler.Code != "unbekannt" {
		t.Fatalf("Kommandozeile: %d, %s, %q", code, aus, fehlerstrom)
	}

	status, rumpf := an.ruf(t, http.MethodPut, "/profile/1/vorgaben/deck",
		`{"werte":["`+deck+`"]}`)
	if dienst := nachgeladenAus(t, rumpf); status != http.StatusOK ||
		dienst.Nachladefehler == nil || *dienst.Nachladefehler != *zeile.Nachladefehler {
		t.Fatalf("Dienst: %d, %s", status, rumpf)
	}

	if werkzeug := nachgeladenAus(t, mussMCP(t, umgebung, "schreiben.profilvorgabe",
		map[string]any{"profilID": "1", "art": "deck", "werte": []string{deck}},
	)); werkzeug.Nachladefehler == nil || *werkzeug.Nachladefehler != *zeile.Nachladefehler {
		t.Fatalf("MCP: %+v", werkzeug)
	}

	if gesetzt := ausDemBestand(t, umgebung,
		`SELECT wert FROM profilvorgaben WHERE profil_id = 1 AND art = 'deck'`); gesetzt != deck {
		t.Errorf("gesetzt ist %q", gesetzt)
	}
}
