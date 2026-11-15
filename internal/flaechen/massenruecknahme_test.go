// **Eine Massenänderung lässt sich zurücknehmen — an jeder Fläche** —
// DOM-03.AK-03 (#1698).
//
// # Der Ablauf
//
// Drei Exemplare in „Kiste 4" → Lagerort „Ordner Rot" für alle → der
// Bericht nennt je Zeile den Lagerort **vorher** und die **neue** Fassung →
// mit genau diesen Angaben zurück nach „Kiste 4".
//
// **Dieselbe Antwort an jeder Fläche** (§16): die Rücknahme ist kein eigener
// Weg, sondern dieselbe Massenänderung mit den Angaben aus dem Bericht.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

func TestEineMassenaenderungLaesstSichUeberallZuruecknehmen(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung, zeilen := dreiExemplare(t)

		vorher := pruefeVorher(t, mussKommandozeile(t, umgebung,
			append([]string{"masse-aendern", "lagerort=Ordner Rot", "--json"}, zeilen...)...))

		mussKommandozeile(t, umgebung,
			append([]string{"masse-aendern", "lagerort=Kiste 4"}, vorher...)...)
		pruefeZurueckInKiste(t, mussKommandozeile(t, umgebung, "sammlung", "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung, zeilen := dreiExemplare(t)

		vorher := pruefeVorher(t, mussMCP(t, umgebung, "schreiben.sammlung.masse", map[string]any{
			"zeilen": alsZeilenobjekte(t, zeilen), "lagerort": "Ordner Rot",
		}))

		mussMCP(t, umgebung, "schreiben.sammlung.masse", map[string]any{
			"zeilen": alsZeilenobjekte(t, vorher), "lagerort": "Kiste 4",
		})
		pruefeZurueckInKiste(t, mussKommandozeile(t, umgebung, "sammlung", "--json"))
	})

	t.Run("REST", func(t *testing.T) {
		umgebung, zeilen := dreiExemplare(t)
		an := starteDienst(t, umgebung)

		vorher := pruefeVorher(t, mussREST(t, an, http.MethodPost, "/sammlung/masse",
			string(mussJSON(t, map[string]any{
				"zeilen": alsZeilenobjekte(t, zeilen), "lagerort": "Ordner Rot",
			}))))

		mussREST(t, an, http.MethodPost, "/sammlung/masse", string(mussJSON(t, map[string]any{
			"zeilen": alsZeilenobjekte(t, vorher), "lagerort": "Kiste 4",
		})))
		pruefeZurueckInKiste(t, mussKommandozeile(t, umgebung, "sammlung", "--json"))
	})
}

// pruefeVorher liest den Stand vorher aus dem Bericht und gibt die Zeilen
// als `id:fassung` für die Rücknahme.
func pruefeVorher(t *testing.T, roh string) []string {
	t.Helper()

	var bericht struct {
		Vorher []struct {
			ID       int64  `json:"id"`
			Fassung  int64  `json:"fassung"`
			Lagerort string `json:"lagerort"`
		} `json:"vorher"`
	}

	if err := json.Unmarshal([]byte(roh), &bericht); err != nil {
		t.Fatalf("der Bericht ist kein gültiges JSON: %v (%s)", err, roh)
	}

	if len(bericht.Vorher) != 3 {
		t.Fatalf("vorher nennt %d Zeilen, erwartet 3 (%s)", len(bericht.Vorher), roh)
	}

	zeilen := make([]string, 0, len(bericht.Vorher))

	for _, stand := range bericht.Vorher {
		if stand.Lagerort != "Kiste 4" || stand.Fassung != 2 {
			t.Errorf("vorher %+v, erwartet Kiste 4 mit Fassung 2", stand)
		}

		zeilen = append(zeilen, strconv.FormatInt(stand.ID, 10)+":"+strconv.FormatInt(stand.Fassung, 10))
	}

	return zeilen
}

// pruefeZurueckInKiste hält, dass alle drei wieder in „Kiste 4" liegen.
func pruefeZurueckInKiste(t *testing.T, roh string) {
	t.Helper()

	var sammlung struct {
		Exemplare []struct {
			Lagerort string `json:"lagerort"`
		} `json:"exemplare"`
	}

	if err := json.Unmarshal([]byte(roh), &sammlung); err != nil {
		t.Fatalf("die Sammlung lesen: %v (%s)", err, roh)
	}

	for _, exemplar := range sammlung.Exemplare {
		if exemplar.Lagerort != "Kiste 4" {
			t.Errorf("ein Exemplar liegt in %q — die Rücknahme kam nicht an", exemplar.Lagerort)
		}
	}
}
