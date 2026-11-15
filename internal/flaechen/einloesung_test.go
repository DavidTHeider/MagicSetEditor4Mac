package flaechen_test

// **Eine Einladung genügt, über jede Fläche** — NET-01b.AK-04
// (#1954):
// „Fenster, Kommandozeile und REST zeigen dasselbe."
//
// # Was hier gemessen wird
//
// **Drei Rechner mit je eigenem Bestand und Schlüssel.** Anna horcht in einem
// eigenen Prozess (`mse horchen`) und stellt an der Kommandozeile einlösbar
// aus; Bert nimmt im Dienst an, Carla an der Kommandozeile. Danach kennt Anna
// beide — unbestätigt und unter dem Namen, den sie sich gaben. **Und eine
// eingelöste Einladung ein zweites Mal** heißt drüben „verbraucht".
//
// Das Fenster fährt denselben Weg in `internal/app/einloesung_intern_test.go`.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestEineEinladungGenuegtUeberJedeFlaeche(t *testing.T) {
	anna, bert, carla := stelleAuf(t), stelleAuf(t), stelleAuf(t)
	adresse, protokoll, _ := horcheMitStopp(t, anna)

	einlad := func() string {
		t.Helper()

		return zeilenwert(t, mussKommandozeile(t, anna.umgebung,
			"einladung-erstellen", "Anna", "handel", "--adresse", adresse), "code")
	}

	// **Bert nimmt im Dienst an.**
	fuerBert := einlad()
	bertsDienst := starteDienst(t, bert.umgebung)

	annahme, fehler := json.Marshal(map[string]any{
		"code": fuerBert, "name": "Anna", "meinName": "Bert",
	})
	if fehler != nil {
		t.Fatal(fehler)
	}

	var angenommen struct {
		Gemeldet struct {
			Stand string `json:"stand"`
		} `json:"gemeldet"`
	}

	if gelesen := json.Unmarshal([]byte(mussREST(t, bertsDienst, http.MethodPost,
		"/peer/einladungen/annahme", string(annahme))), &angenommen); gelesen != nil {
		t.Fatal(gelesen)
	}

	if angenommen.Gemeldet.Stand != "gemeldet" {
		t.Errorf("der Dienst meldet %+v — Annas Protokoll: %s", angenommen, protokoll.text())
	}

	// **Carla nimmt an der Kommandozeile an** — dieselbe Meldung, als Zeile.
	if gemeldet := zeilenwert(t, mussKommandozeile(t, carla.umgebung,
		"einladung-annehmen", einlad(), "Anna", "--mein-name", "Carla"), "gemeldet"); gemeldet != "gemeldet" {
		t.Errorf("die Kommandozeile meldet %q", gemeldet)
	}

	// **Anna kennt beide jetzt**, unbestätigt, unter ihren Namen.
	kontakte := mussKommandozeile(t, anna.umgebung, "kontakte")
	for _, erwartet := range []string{
		"unbestaetigt\t" + bert.kennung + "\t" + bert.abdruck + "\tBert",
		"unbestaetigt\t" + carla.kennung + "\t" + carla.abdruck + "\tCarla",
	} {
		if !strings.Contains(kontakte, erwartet) {
			t.Errorf("bei Anna fehlt %q in %q", erwartet, kontakte)
		}
	}

	// **Berts Einladung ein zweites Mal** — Carla nimmt sie an, und Anna sagt
	// „verbraucht". Die Annahme bei Carla bleibt (§7, Regel 3).
	if gemeldet := zeilenwert(t, mussKommandozeile(t, carla.umgebung,
		"einladung-annehmen", fuerBert, "Anna"), "gemeldet"); !strings.HasPrefix(gemeldet, "gescheitert\t") ||
		!strings.Contains(gemeldet, "schon eingelöst") {
		t.Errorf("eine verbrauchte Einladung meldet %q", gemeldet)
	}
}
