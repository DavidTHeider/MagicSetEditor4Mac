// **Der Umfang der Sammlungswerkzeuge, gemessen statt aufgezählt** —
// Werkstück 2 von Phase 3.4
// (#1165).
//
// # Warum das neben `werkzeugumfang_test.go` steht und nicht darin
//
// Der Wächter dort liest **Namen**. Er hält, dass kein schreibendes Werkzeug
// ohne eine aufgeschriebene Umfangsentscheidung dasteht — und mehr kann er
// nicht: seine elf Gruppen sind Text, den jemand getippt hat, und ein
// Werkzeug, das in die falsche Gruppe einträgt, macht ihn nicht rot.
//
// **Hier wird gerufen.** Die zehnte Zusage lautet »der Aufruf ändert **eine**
// Zeile der Sammlung«, und dieser Lauf sieht nach: drei Zeilen hinein, ein
// Ruf, und die beiden anderen sind danach Feld für Feld dieselben.
//
// Angestrichen vom Zweitblick über Phase 3.4: eine Kategorieliste, die sich
// als Wirkungswächter ausgibt, ist die halbe Zusage — und macht die andere
// unsichtbar (§10).
package mcp_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// dreiZeilen gibt eine Sammlung mit drei unterscheidbaren Exemplaren.
//
// **Drei und nicht zwei.** Bei zweien wäre „die andere ist unberührt" auch
// dann grün, wenn ein Werkzeug immer genau die zweite anfasst; bei dreien
// muss es die richtige treffen.
func dreiZeilen() *sammlungsprobe {
	return &sammlungsprobe{
		zeilen: []mcp.Exemplarsicht{
			{ID: 1, Printing: "druck-eins", Menge: 3, Zustand: "nearMint", Sprache: "en"},
			{ID: 2, Printing: "druck-zwei", Menge: 5, Zustand: "played", Sprache: "de"},
			{ID: 3, Printing: "druck-drei", Menge: 7, Zustand: "good", Sprache: "fr"},
		},
		naechste: 3,
	}
}

// **Jedes der drei Werkzeuge fasst genau eine Zeile an.**
func TestDieSammlungswerkzeugeAendernGenauEineZeile(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		werkzeug   string
		argumente  string
		danach     int
		unberuehrt []int64
	}{
		"anlegen": {
			werkzeug:  "schreiben.exemplar",
			argumente: `{"printing":"druck-vier","menge":1}`,
			danach:    4, unberuehrt: []int64{1, 2, 3},
		},
		"ändern": {
			werkzeug:  "schreiben.exemplaraendern",
			argumente: `{"id":2,"printing":"druck-zwei","menge":9}`,
			danach:    3, unberuehrt: []int64{1, 3},
		},
		"löschen": {
			werkzeug:  "schreiben.exemplarloeschen",
			argumente: `{"id":2}`,
			danach:    2, unberuehrt: []int64{1, 3},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			sammlung := dreiZeilen()
			vorher := merkeZeilen(sammlung)

			if _, err := rufeSammlungswerkzeug(t, sammlung, neuePreisprobe(),
				fall.werkzeug, fall.argumente); err != nil {
				t.Fatalf("%s: %v", fall.werkzeug, err)
			}

			if gezaehlt := len(sammlung.zeilen); gezaehlt != fall.danach {
				t.Errorf("%d Zeilen danach, erwartet %d", gezaehlt, fall.danach)
			}

			// **Und die übrigen sind Feld für Feld dieselben.** Ein
			// Vergleich der Anzahl allein ließe ein Werkzeug durch, das eine
			// Zeile anlegt und dabei eine andere still überschreibt.
			for _, id := range fall.unberuehrt {
				jetzt, steht := zeileMit(sammlung, id)
				if !steht {
					t.Errorf("Zeile %d ist verschwunden", id)

					continue
				}

				if jetzt != vorher[id] {
					t.Errorf("Zeile %d hat sich geändert: %+v statt %+v",
						id, jetzt, vorher[id])
				}
			}
		})
	}
}

// **Und die Gegenprobe: der Vergleich sieht eine Änderung.**
//
// Ein Lauf, der keine Abweichung findet, hat zwei Lesarten — die Werkzeuge
// fassen wirklich nur eine Zeile an, oder der Vergleich sieht nichts. Erst
// diese Zeile trennt sie (§10): dieselbe Prüfung gegen eine Sammlung, in der
// eine Nachbarzeile absichtlich verschoben wurde, muss anschlagen.
func TestDerVergleichSiehtEineFremdeAenderung(t *testing.T) {
	t.Parallel()

	sammlung := dreiZeilen()
	vorher := merkeZeilen(sammlung)

	if err := sammlung.ExemplarAendern(context.Background(), mcp.Exemplarsicht{
		ID: 1, Printing: "druck-eins", Menge: 99, Zustand: "nearMint", Sprache: "en",
	}); err != nil {
		t.Fatalf("ändern: %v", err)
	}

	jetzt, _ := zeileMit(sammlung, 1)
	if jetzt == vorher[1] {
		t.Error("der Vergleich hält zwei verschiedene Zeilen für gleich")
	}
}

// merkeZeilen hält den Stand je Kennung fest.
func merkeZeilen(sammlung *sammlungsprobe) map[int64]mcp.Exemplarsicht {
	gemerkt := map[int64]mcp.Exemplarsicht{}
	for _, zeile := range sammlung.zeilen {
		gemerkt[zeile.ID] = zeile
	}

	return gemerkt
}

// zeileMit sucht eine Zeile.
func zeileMit(sammlung *sammlungsprobe, id int64) (mcp.Exemplarsicht, bool) {
	for _, zeile := range sammlung.zeilen {
		if zeile.ID == id {
			return zeile, true
		}
	}

	return mcp.Exemplarsicht{}, false
}
