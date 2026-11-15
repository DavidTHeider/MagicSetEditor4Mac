// **Massenvorgänge an jeder Fläche** — Werkstück 6 von Phase 4.5
// (#1292).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Die Zusage, um die es geht
//
// **Jede Zeile trägt ihre Fassung** (§7, Regel 2). Eine, die inzwischen jemand
// anderes geändert hat, wird **übergangen** statt überschrieben — und der
// Bericht sagt, welche. Ohne diese Zusage überschriebe ein Massenvorgang
// fünfzig fremde Änderungen auf einmal.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

// dreiExemplare legt drei an und gibt Umgebung und Zeilen als `id:fassung`.
func dreiExemplare(t *testing.T) (umgebung, zeilen []string) {
	t.Helper()

	umgebung = eigenerBestand(t)

	for zaehler, zustand := range []string{"nearMint", "played", "good"} {
		var angelegt struct {
			ID int64 `json:"id"`
		}

		roh := mussKommandozeile(t, umgebung, "exemplar",
			"druck-"+strconv.Itoa(zaehler), "1", zustand, "en", "nonfoil",
			"", "", "", "", "", "", "", "Kiste 4", "besitz", "--json")

		if err := json.Unmarshal([]byte(roh), &angelegt); err != nil {
			t.Fatalf("die Antwort lesen: %v (%s)", err, roh)
		}

		zeilen = append(zeilen, strconv.FormatInt(angelegt.ID, 10)+":1")
	}

	return umgebung, zeilen
}

// **Ein Lagerort für drei Zeilen — an jeder Fläche.**
func TestEinMassenvorgangLaeuftUeberallGleich(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung, zeilen := dreiExemplare(t)

		argumente := append([]string{"masse-aendern", "lagerort=Ordner Rot"}, zeilen...)

		if aus := mussKommandozeile(t, umgebung, argumente...); !strings.Contains(
			aus, "3 geändert") {
			t.Errorf("die Ausgabe lautet %q", aus)
		}

		pruefeLagerort(t, mussKommandozeile(t, umgebung, "sammlung", "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung, zeilen := dreiExemplare(t)

		roh, istFehler := mcpRuf(t, umgebung, "schreiben.sammlung.masse",
			map[string]any{
				"zeilen": alsZeilenobjekte(t, zeilen), "lagerort": "Ordner Rot",
			})
		if istFehler {
			t.Fatalf("schreiben.sammlung.masse: %s", roh)
		}

		pruefeLagerort(t, mussKommandozeile(t, umgebung, "sammlung", "--json"))
	})

	t.Run("REST", func(t *testing.T) {
		umgebung, zeilen := dreiExemplare(t)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/sammlung/masse",
			string(mussJSON(t, map[string]any{
				"zeilen": alsZeilenobjekte(t, zeilen), "lagerort": "Ordner Rot",
			})))
		if status != http.StatusOK {
			t.Fatalf("POST /sammlung/masse: %d (%s)", status, rumpf)
		}

		pruefeLagerort(t, mussKommandozeile(t, umgebung, "sammlung", "--json"))
	})
}

// pruefeLagerort hält, dass alle drei Zeilen umgezogen sind.
func pruefeLagerort(t *testing.T, roh string) {
	t.Helper()

	var sammlung struct {
		Exemplare []struct {
			Lagerort string `json:"lagerort"`
		} `json:"exemplare"`
	}

	if err := json.Unmarshal([]byte(roh), &sammlung); err != nil {
		t.Fatalf("die Sammlung lesen: %v (%s)", err, roh)
	}

	if len(sammlung.Exemplare) != 3 {
		t.Fatalf("%d Exemplare, erwartet 3", len(sammlung.Exemplare))
	}

	for _, exemplar := range sammlung.Exemplare {
		if exemplar.Lagerort != "Ordner Rot" {
			t.Errorf("ein Exemplar liegt in %q", exemplar.Lagerort)
		}
	}
}

// **Eine veraltete Fassung wird übergangen — an jeder Fläche.**
func TestEineVeralteteFassungWirdUeberallUebergangen(t *testing.T) {
	t.Parallel()

	umgebung, zeilen := dreiExemplare(t)

	// **Die mittlere bekommt eine Fassung, die es nicht gibt.**
	verstellt := make([]string, len(zeilen))
	copy(verstellt, zeilen)
	verstellt[1] = strings.Split(zeilen[1], ":")[0] + ":99"

	argumente := append([]string{"masse-aendern", "lagerort=Kiste 9"}, verstellt...)

	aus := mussKommandozeile(t, umgebung, argumente...)

	if !strings.Contains(aus, "2 geändert") {
		t.Errorf("die Ausgabe lautet %q", aus)
	}

	// **Und die übergangene Zeile steht im Bericht** — sonst weiß niemand,
	// welche es war.
	if !strings.Contains(aus, "!") {
		t.Errorf("kein Befund zur übergangenen Zeile: %q", aus)
	}
}

// **Löschen in Masse — an jeder Fläche.**
func TestLoeschenInMasseLaeuftUeberallGleich(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung, zeilen := dreiExemplare(t)

		argumente := append([]string{"masse-loeschen"}, zeilen...)

		if aus := mussKommandozeile(t, umgebung, argumente...); !strings.Contains(
			aus, "3 geändert") {
			t.Errorf("die Ausgabe lautet %q", aus)
		}

		sammlungIstLeer(t, umgebung)
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung, zeilen := dreiExemplare(t)

		roh, istFehler := mcpRuf(t, umgebung, "schreiben.sammlung.masseloeschen",
			map[string]any{"zeilen": alsZeilenobjekte(t, zeilen)})
		if istFehler {
			t.Fatalf("schreiben.sammlung.masseloeschen: %s", roh)
		}

		sammlungIstLeer(t, umgebung)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung, zeilen := dreiExemplare(t)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/sammlung/masse/loeschen",
			string(mussJSON(t, map[string]any{
				"zeilen": alsZeilenobjekte(t, zeilen),
			})))
		if status != http.StatusOK {
			t.Fatalf("POST /sammlung/masse/loeschen: %d (%s)", status, rumpf)
		}

		sammlungIstLeer(t, umgebung)
	})
}

// sammlungIstLeer liest die Sammlung an derselben Stelle — **die Wirkung an
// jeder Fläche**, nicht nur ihr Ja: hier prüfte nur die Kommandozeile, ob
// gelöscht war (Zweitblick über Phase C, Befund 3).
func sammlungIstLeer(t *testing.T, umgebung []string) {
	t.Helper()

	if sammlung := mussKommandozeile(t, umgebung, "sammlung", "--json"); !strings.Contains(sammlung, `"exemplare": []`) {
		t.Errorf("die Sammlung ist nicht leer: %s", sammlung)
	}
}

// **Der Preisschub gibt es an jeder Fläche — und sagt, was bleibt.**
//
// **Gegen die Gegenstelle**, nicht gegen Scryfall: die Umgebung lenkt die
// Adresse um und verlegt das Zuhause (§3).
func TestDerPreisschubLaeuftUeberallGleich(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := scryfallumgebung(t, scryfallgegenstelle(t))

		mussKommandozeile(t, umgebung, "exemplar", "probe-druck", "1")

		if aus := mussKommandozeile(t, umgebung,
			"preise-alle", "0"); !strings.Contains(aus, "offen") {
			t.Errorf("die Ausgabe nennt den Rest nicht: %q", aus)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := scryfallumgebung(t, scryfallgegenstelle(t))

		mussKommandozeile(t, umgebung, "exemplar", "probe-druck", "1")

		roh, istFehler := mcpRuf(t, umgebung, "schreiben.sammlung.preisschub",
			map[string]any{"hoechstensStunden": 0})
		if istFehler {
			t.Fatalf("schreiben.sammlung.preisschub: %s", roh)
		}

		if !strings.Contains(roh, "fertig") {
			t.Errorf("die Antwort nennt den Stand nicht: %s", roh)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := scryfallumgebung(t, scryfallgegenstelle(t))

		mussKommandozeile(t, umgebung, "exemplar", "probe-druck", "1")

		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/sammlung/masse/preise",
			`{"hoechstensStunden":0}`)
		if status != http.StatusOK {
			t.Fatalf("POST /sammlung/masse/preise: %d (%s)", status, rumpf)
		}
	})
}

// alsZeilenobjekte macht aus `id:fassung` die Objekte, die MCP und REST wollen.
func alsZeilenobjekte(t *testing.T, zeilen []string) []any {
	t.Helper()

	gebaut := make([]any, 0, len(zeilen))

	for _, zeile := range zeilen {
		kennung, fassung, _ := strings.Cut(zeile, ":")

		id, kennungsfehler := strconv.ParseInt(kennung, 10, 64)
		stand, fassungsfehler := strconv.ParseInt(fassung, 10, 64)

		if kennungsfehler != nil || fassungsfehler != nil {
			t.Fatalf("%q ist keine Zeile", zeile)
		}

		gebaut = append(gebaut, map[string]any{"id": id, "fassung": stand})
	}

	return gebaut
}
