// **Ein Feld auf vielen Karten setzen — über jede Fläche**
// (#1390).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Die Zusage, um die es geht
//
// **Alles oder nichts.** Wer ein Feld über viele Karten ändert, will nicht
// erfahren, dass die Hälfte durchging — und schon gar nicht, welche Hälfte
// raten müssen. Eine Karte, die nicht zu diesem Set gehört, lässt den ganzen
// Aufruf scheitern, und danach ist **keine** geändert.
package flaechen_test

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setMitKarten fährt das Probeset ein und gibt Umgebung, Set und Karten.
func setMitKarten(t *testing.T) (umgebung []string, setID string, karteIDs []string) {
	t.Helper()

	umgebung = eigenerBestand(t)

	aus := mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json")
	setID = setkennungAus(t, aus)

	aus = mussKommandozeile(t, umgebung, "karten", setID, "--json")

	var antwort struct {
		Karten []struct {
			ID string `json:"id"`
		} `json:"karten"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, aus)
	}

	if len(antwort.Karten) < 2 {
		t.Fatalf("%d Karten im Probeset, gebraucht werden zwei", len(antwort.Karten))
	}

	for _, karte := range antwort.Karten {
		karteIDs = append(karteIDs, karte.ID)
	}

	return umgebung, setID, karteIDs
}

// feldwert liest den Wert eines Feldes einer Karte über die Kommandozeile.
//
// **Über eine Fläche und nicht über den Bestand.** Der Bestand wäre die
// kürzere Antwort und die schwächere: sie sagte, dass etwas geschrieben
// wurde, nicht dass es zu lesen ist.
//
// **Über `karten` und nicht über `karte`**: `mse karte` ist der Lebenszyklus
// (anlegen, kopieren, löschen), nicht das Lesen einer einzelnen.
func feldwert(t *testing.T, umgebung []string, setID, karteID, feld string) string {
	t.Helper()

	aus := mussKommandozeile(t, umgebung, "karten", setID, "--json")

	var antwort struct {
		Karten []struct {
			ID    string            `json:"id"`
			Werte map[string]string `json:"werte"`
		} `json:"karten"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, aus)
	}

	for _, karte := range antwort.Karten {
		if karte.ID == karteID {
			return karte.Werte[feld]
		}
	}

	t.Fatalf("die Karte %s steht nicht in %s", karteID, aus)

	return ""
}

// zweitesSet legt eine Kopie des Probesets unter einem anderen Pfad an.
//
// **Weil das Set über seinen Pfad erkannt wird.** Zweimal dieselbe Datei zu
// öffnen gibt dieselbe Kennung; für eine Karte, die wirklich fremd ist,
// braucht es eine zweite Datei.
// **Und `probe.mse-set` ist ein Verzeichnis**, kein Archiv — ein
// entpacktes Paket. Kopiert wird deshalb sein Inhalt, nicht die Datei.
func zweitesSet(t *testing.T) string {
	t.Helper()

	ziel := filepath.Join(t.TempDir(), "zweites.mse-set")
	if err := os.MkdirAll(ziel, 0o750); err != nil {
		t.Fatalf("das Ziel anlegen: %v", err)
	}

	// **Gelesen wird durch ein `os.Root`**, nicht über zusammengesetzte
	// Pfade. `gosec` weist die zweite Form als G703 ab, und zu Recht: ein
	// Name aus `ReadDir` kann über einen Symlink aus dem Verzeichnis
	// hinausführen. Dass hier ein Verzeichnis des Repositoriums steht, macht
	// den Zugriff nicht wurzelgebunden — es macht den Angriff
	// unwahrscheinlich. Derselbe Griff wie in `vorlagenbeigaben.go`.
	griff, err := os.OpenRoot(probeset(t))
	if err != nil {
		t.Fatalf("das Probeset öffnen: %v", err)
	}

	defer func() { _ = griff.Close() }() //nolint:errcheck // nur gelesen

	eintraege, err := fs.ReadDir(griff.FS(), ".")
	if err != nil {
		t.Fatalf("das Probeset lesen: %v", err)
	}

	for _, eintrag := range eintraege {
		if eintrag.IsDir() {
			continue
		}

		inhalt, lesefehler := fs.ReadFile(griff.FS(), eintrag.Name())
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", eintrag.Name(), lesefehler)
		}

		if err := os.WriteFile(
			filepath.Join(ziel, eintrag.Name()), inhalt, 0o600,
		); err != nil {
			t.Fatalf("%s schreiben: %v", eintrag.Name(), err)
		}
	}

	return ziel
}

// **Dasselbe Feld auf zwei Karten — über Kommandozeile, MCP und REST.**
func TestEinFeldAufVielenKartenUeberJedeFlaeche(t *testing.T) {
	for name, setze := range map[string]func(*testing.T, []string, string, []string){
		"Kommandozeile": func(t *testing.T, umgebung []string, setID string, karteIDs []string) {
			t.Helper()

			argumente := append(
				[]string{"felder-setzen", setID, "artist", "Aus der Fläche"}, karteIDs...)

			if aus := mussKommandozeile(t, umgebung, argumente...); !strings.Contains(
				aus, "2 Karten") {
				t.Errorf("die Ausgabe lautet %q", aus)
			}
		},
		"MCP": func(t *testing.T, umgebung []string, setID string, karteIDs []string) {
			t.Helper()

			text, istFehler := mcpRuf(t, umgebung, "schreiben.felder", map[string]any{
				"setID":    zahl(t, setID),
				"karteIDs": zahlen(t, karteIDs),
				"feld":     "artist",
				"wert":     "Aus der Fläche",
			})
			if istFehler {
				t.Fatalf("schreiben.felder: %s", text)
			}
		},
		"REST": func(t *testing.T, umgebung []string, setID string, karteIDs []string) {
			t.Helper()

			an := starteDienst(t, umgebung)

			status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/felder",
				`{"karteIDs":`+alsJSONZahlen(t, karteIDs)+
					`,"feld":"artist","wert":"Aus der Fläche"}`)
			if status != http.StatusOK {
				t.Fatalf("POST: %d (%s)", status, rumpf)
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			umgebung, setID, karteIDs := setMitKarten(t)
			zwei := karteIDs[:2]

			setze(t, umgebung, setID, zwei)

			for _, karteID := range zwei {
				if wert := feldwert(t, umgebung, setID, karteID, "artist"); wert != "Aus der Fläche" {
					t.Errorf("Karte %s trägt %q", karteID, wert)
				}
			}
		})
	}
}

// **Und eine fremde Karte lässt überall den ganzen Aufruf scheitern.**
//
// Die Gegenprobe: ohne sie wäre der Lauf darüber auch dann grün, wenn eine
// Fläche in einer Schleife schriebe und die Hälfte stehen ließe.
func TestEineFremdeKarteLaesstDenGanzenAufrufScheitern(t *testing.T) {
	// **Eine Karte aus einem zweiten Set, nicht eine erfundene Kennung.**
	//
	// Hier stand `999999`, und die Gegenprobe hat es angestrichen: mit einer
	// Kennung, die es **gar nicht** gibt, scheitert schon der Speicher. Der
	// Lauf war grün, auch als die Naht die Aufsicht umging — er prüfte also
	// nicht die Setgrenze, sondern nur, dass eine unbekannte Zeile auffällt.
	//
	// Eine Karte, die es **gibt** und die zu einem **anderen** Set gehört,
	// unterscheidet die beiden: sie kommt durch den Speicher und muss an der
	// Aufsicht scheitern
	// ([0041](../../docs/entscheidungen/0041-die-setgrenze-liegt-im-dienst.md)).

	for name, setze := range map[string]func(*testing.T, []string, string, []string) bool{
		"Kommandozeile": func(t *testing.T, umgebung []string, setID string, karteIDs []string) bool {
			t.Helper()

			argumente := append(
				[]string{"felder-setzen", setID, "artist", "Verboten"}, karteIDs...)

			_, _, code := kommandozeile(t, umgebung, argumente...)

			return code != 0
		},
		"MCP": func(t *testing.T, umgebung []string, setID string, karteIDs []string) bool {
			t.Helper()

			_, istFehler := mcpRuf(t, umgebung, "schreiben.felder", map[string]any{
				"setID":    zahl(t, setID),
				"karteIDs": zahlen(t, karteIDs),
				"feld":     "artist",
				"wert":     "Verboten",
			})

			return istFehler
		},
		"REST": func(t *testing.T, umgebung []string, setID string, karteIDs []string) bool {
			t.Helper()

			an := starteDienst(t, umgebung)

			status, _ := an.ruf(t, http.MethodPost, "/sets/"+setID+"/felder",
				`{"karteIDs":`+alsJSONZahlen(t, karteIDs)+
					`,"feld":"artist","wert":"Verboten"}`)

			return status != http.StatusOK
		},
	} {
		t.Run(name, func(t *testing.T) {
			umgebung, setID, karteIDs := setMitKarten(t)

			// **Ein zweites Set aus einer Kopie unter anderem Pfad.**
			//
			// Zweimal dieselbe Datei zu öffnen gibt **dasselbe** Set —
			// gemessen, nicht vermutet: beide Male kommt Kennung 1 zurück.
			// Der erste Anlauf tat genau das, und die „fremde" Karte lag
			// damit im selben Set. Der Lauf war grün und prüfte nichts.
			zweites := mussKommandozeile(t, umgebung, "oeffnen", zweitesSet(t), "--json")
			fremdeKarte := ersteKarte(t, mussKommandozeile(t, umgebung,
				"karten", setkennungAus(t, zweites), "--json"))

			mitFremder := []string{karteIDs[0], fremdeKarte}

			if !setze(t, umgebung, setID, mitFremder) {
				t.Fatal("der Aufruf ging durch, erwartet war eine Abweisung")
			}

			// **Und die erste Karte ist unberührt.** Das ist die eigentliche
			// Zusage: nicht dass es scheitert, sondern dass nichts halb
			// geschrieben ist.
			if wert := feldwert(t, umgebung, setID, karteIDs[0], "artist"); wert == "Verboten" {
				t.Error("die erste Karte wurde geändert, obwohl der Aufruf scheiterte")
			}
		})
	}
}

// **Die Vorschau einer Massenänderung über jede Fläche** — Werkstück U6 von
// Phase 5.1 (#1417): dieselben Zeilen aus der Zeile, über MCP und im Dienst,
// und keine schreibt.
func TestDieVorschauEinerMassenaenderungUeberJedeFlaeche(t *testing.T) {
	for name, vorschau := range map[string]func(*testing.T, []string, string, []string) string{
		"Kommandozeile": func(t *testing.T, umgebung []string, setID string, karteIDs []string) string {
			t.Helper()

			argumente := append(
				[]string{"felder-vorschau", setID, "artist", "Vorgesehen", "--json"}, karteIDs...)

			return mussKommandozeile(t, umgebung, argumente...)
		},
		"MCP": func(t *testing.T, umgebung []string, setID string, karteIDs []string) string {
			t.Helper()

			text, istFehler := mcpRuf(t, umgebung, "lesen.felder.vorschau", map[string]any{
				"setID":    zahl(t, setID),
				"karteIDs": zahlen(t, karteIDs),
				"feld":     "artist",
				"wert":     "Vorgesehen",
			})
			if istFehler {
				t.Fatalf("lesen.felder.vorschau: %s", text)
			}

			return text
		},
		"REST": func(t *testing.T, umgebung []string, setID string, karteIDs []string) string {
			t.Helper()

			an := starteDienst(t, umgebung)
			status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/felder/vorschau",
				`{"karteIDs":`+alsJSONZahlen(t, karteIDs)+
					`,"feld":"artist","wert":"Vorgesehen"}`)
			if status != http.StatusOK {
				t.Fatalf("POST: %d (%s)", status, rumpf)
			}

			return rumpf
		},
	} {
		t.Run(name, func(t *testing.T) {
			umgebung, setID, karteIDs := setMitKarten(t)
			zwei := karteIDs[:2]

			text := vorschau(t, umgebung, setID, zwei)

			// Jede Karte steht mit dem neuen Wert in der Antwort …
			if !strings.Contains(text, "Vorgesehen") || strings.Count(text, `"neu"`) != 2 {
				t.Errorf("die Antwort nennt nicht zwei Zeilen mit dem neuen Wert: %s", text)
			}

			// … und keine trägt ihn im Bestand.
			for _, karteID := range zwei {
				if wert := feldwert(t, umgebung, setID, karteID, "artist"); wert == "Vorgesehen" {
					t.Errorf("Karte %s wurde geschrieben", karteID)
				}
			}
		})
	}
}
