// Derselbe Arbeitsablauf über den Dienst auf der Rückschleife
// (#997).
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **Derselbe Ablauf, dritte Fläche**: einfahren, ändern, zeichnen, sichern.
//
// # Warum gegen das gebaute Programm
//
// `internal/rest` prüft die Wachen über Doubles, `cmd/mse-rest` die
// Verdrahtung. Hier fährt ein **echter Prozess** über eine **echte
// Verbindung** — das ist, was ein Aufrufer bekommt.
func TestDerAblaufUeberRest(t *testing.T) {
	umgebung := eigenerBestand(t)
	ziel := t.TempDir()
	an := starteDienst(t, umgebung)

	// 1. einfahren
	status, rumpf := an.ruf(t, http.MethodPost, "/sets",
		`{"pfad":`+alsJSONText(t, probeset(t))+`}`)
	if status != http.StatusCreated {
		t.Fatalf("einfahren: %d (%s)", status, rumpf)
	}

	setID := ausRest(t, rumpf, "setID")

	// 2. eine Karte finden
	status, rumpf = an.ruf(t, http.MethodGet, "/sets/"+setID+"/karten", "")
	if status != http.StatusOK {
		t.Fatalf("karten: %d (%s)", status, rumpf)
	}

	karteID := ersteKarteRest(t, rumpf)

	// 3. ändern
	status, rumpf = an.ruf(t, http.MethodPut,
		"/sets/"+setID+"/karten/"+karteID+"/felder/name", `{"wert":"Aus der Abnahme"}`)
	if status != http.StatusOK {
		t.Fatalf("feld: %d (%s)", status, rumpf)
	}

	// 4. zeichnen
	bild := filepath.Join(ziel, "karte.png")

	status, rumpf = an.ruf(t, http.MethodPost, "/karten/"+karteID+"/bild",
		`{"ziel":`+alsJSONText(t, bild)+`}`)
	if status != http.StatusCreated {
		t.Fatalf("zeichnen: %d (%s)", status, rumpf)
	}

	istBild(t, bild)

	// 5. sichern
	raus := filepath.Join(ziel, "raus.mse-set")

	status, rumpf = an.ruf(t, http.MethodPost, "/sets/"+setID+"/sicherung",
		`{"ziel":`+alsJSONText(t, raus)+`}`)
	if status != http.StatusCreated {
		t.Fatalf("sichern: %d (%s)", status, rumpf)
	}

	istWiederLesbaresSet(t, umgebung, raus, "Aus der Abnahme")

	// 6. und die Änderung steht wirklich darin
	_, rumpf = an.ruf(t, http.MethodGet, "/sets/"+setID+"/karten", "")
	if !strings.Contains(rumpf, "Aus der Abnahme") {
		t.Error("die Änderung steht nicht im Bestand")
	}
}

// **Und ohne Marke kommt der ganze Ablauf nicht einmal an.**
//
// Die Gegenprobe zum Fall darüber: ohne sie bliebe offen, ob die Wachen im
// gebauten Programm überhaupt hängen — geprüft ist das sonst nur an Doubles.
func TestOhneMarkeGehtUeberRestNichts(t *testing.T) {
	umgebung := eigenerBestand(t)
	an := starteDienst(t, umgebung).ohneMarke()

	faelle := []struct{ methode, weg, rumpf string }{
		{http.MethodGet, "/sets", ""},
		{http.MethodGet, "/sets/1/karten", ""},
		{http.MethodPut, "/sets/1/karten/1/felder/name", `{"wert":"x"}`},
		{http.MethodPost, "/sets", `{"pfad":"/tmp/x.mse-set"}`},
		{http.MethodPost, "/sets/1/sicherung", `{"ziel":"/tmp/y.mse-set"}`},
		{http.MethodPost, "/karten/1/bild", `{"ziel":"/tmp/z.png"}`},
	}

	for _, fall := range faelle {
		t.Run(fall.methode+" "+fall.weg, func(t *testing.T) {
			status, rumpf := an.ruf(t, fall.methode, fall.weg, fall.rumpf)
			if status != http.StatusUnauthorized {
				t.Errorf("Status %d, erwartet war %d (%s)",
					status, http.StatusUnauthorized, rumpf)
			}
		})
	}
}

// ersteKarteRest liest die erste Kartenkennung aus einer REST-Antwort.
func ersteKarteRest(t *testing.T, rumpf string) string {
	t.Helper()

	var gelesen struct {
		Karten []struct {
			ID string `json:"id"`
		} `json:"karten"`
	}

	if err := json.Unmarshal([]byte(rumpf), &gelesen); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, rumpf)
	}

	if len(gelesen.Karten) == 0 {
		t.Fatalf("keine Karte in %s", rumpf)
	}

	return gelesen.Karten[0].ID
}

// alsJSONText schreibt eine Zeichenkette als JSON-Wert.
//
// **Nicht von Hand mit Anführungszeichen.** Ein Pfad aus `t.TempDir()` kann
// Zeichen enthalten, die in JSON entwertet werden müssen.
func alsJSONText(t *testing.T, wert string) string {
	t.Helper()

	roh, err := json.Marshal(wert)
	if err != nil {
		t.Fatalf("%q als JSON: %v", wert, err)
	}

	return string(roh)
}

// **Ein Pfad aus einem verborgenen Winkel erreicht keinen Weg** — am echten
// `mse-rest`, OPS-02.AK-04 (#1719). Die Datei gibt es nicht; gelesen würde
// ohnehin nichts, denn der Vertrag weist ab, bevor ein Weg den Pfad sieht.
func TestEinOffenerServerNimmtKeinenVerborgenenPfad(t *testing.T) {
	heim, fehler := os.UserHomeDir()
	if fehler != nil {
		t.Fatal(fehler)
	}

	an := starteDienst(t, eigenerBestand(t))

	for weg, rumpf := range map[string]string{
		"/sets":                          `{"pfad":` + alsJSONText(t, filepath.Join(heim, ".ssh", "gibt-es-nicht")) + `}`,
		"/bestand/sicherung?pfad=/etc/x": "",
	} {
		methode := http.MethodPost
		if rumpf == "" {
			methode = http.MethodGet
		}

		// **Abgewiesen, bevor ein Weg den Pfad sieht** — der Satz sagt, ob
		// es der verborgene Ordner war oder der Ort außerhalb; kein Weg
		// sagt „nicht gefunden".
		if status, antwort := an.ruf(t, methode, weg, rumpf); status != http.StatusBadRequest ||
			!strings.Contains(antwort, `"validierung"`) || strings.Contains(antwort, "nicht zu finden") {
			t.Errorf("%s: %d %s", weg, status, antwort)
		}
	}

	// **Dass jemand zuhört**: das Probeset aus dem Dienstordner kommt durch.
	if status, antwort := an.ruf(t, http.MethodPost, "/sets", `{"pfad":`+alsJSONText(t, probeset(t))+`}`); status != http.StatusCreated {
		t.Errorf("das Probeset: %d %s", status, antwort)
	}
}
