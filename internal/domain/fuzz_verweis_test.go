package domain_test

// **Eine portable Referenz aus fremder Hand** — Werkstück O11 von Phase 5.2
// (#1435),
// Befund 3 des Zweitblicks: das Ziel über die Referenzen fehlte.
//
// # Warum
//
// Ein [domain.Druckverweis] ist das, was einen Bestand **verlässt** — in einer
// Ausfuhr, in einem CSV, in einem Deck, das jemand weitergibt. Er kommt in
// einem anders nummerierten Bestand an, und dort entscheidet [domain.LoeseAuf],
// welchen Druck er meint. Eine falsche Auflösung ist keine Meldung, sondern
// eine **Karte, die jemandem untergeschoben wird**.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - zweimal derselbe Verweis, zwei Auflösungen
//   - **eine Art außerhalb der drei**
//   - **„eindeutig" mit null oder mehr als einem Druck** — das ist die
//     Zusage aus 0086, und ihr Bruch gäbe die falsche Karte ohne Rückfrage
//   - **„mehrdeutig" mit weniger als zwei** und **„unbekannt" mit einem**
//   - **eine Kennung, die nicht trifft, und trotzdem ein Rückfall über einen
//     unvollständigen Namen** — der Rückfall gilt nur ganz
//
//	go test ./internal/domain -run Fuzz -fuzz FuzzLoeseAuf -fuzztime 2m

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// verweissaat sind die Gestalten, mit denen der Fuzzer beginnt.
var verweissaat = []string{
	`{}`,
	`{"Fassung":1}`,
	`{"Fassung":1,"System":"scryfall","Semantik":"druck","Wert":"w-1"}`,
	`{"Fassung":1,"System":"scryfall","Semantik":"druck","Wert":"gibtesnicht"}`,
	`{"Fassung":1,"Name":"Karte 1","Setcode":"m11","Nummer":"1","Sprache":"en"}`,
	`{"Fassung":1,"Name":"Karte 1","Setcode":"m11"}`,
	`{"Fassung":1,"Name":"  ","Setcode":"m11","Nummer":"1"}`,

	// **Eine Kennung, die nicht trifft, neben einem halben Namen** — Befund 2
	// des Nachlaufs zum Zweitblick über Phase 5.2. Genau diese Gestalt nennt
	// der Kopf als Fund, und keine Saat brachte sie mit.
	`{"Fassung":1,"System":"scryfall","Semantik":"druck","Wert":"gibtesnicht",` +
		`"Name":"Karte 1","Setcode":"m11"}`,
	`{"Fassung":1,"System":"scryfall","Semantik":"druck","Wert":"gibtesnicht",` +
		`"Name":"Karte 1","Setcode":"m11","Nummer":"1","Sprache":"en"}`,

	// **Und eine, die trifft, neben einem Namen, der etwas anderes trifft.**
	// Ohne sie sagte der Lauf nur, dass die Kennung **irgendetwas** ergibt —
	// nicht, dass sie den Vorrang behält: `w-1` meint 1 und 2, der Name meint
	// 3. Nachgemessen als Gegenprobe am 2026-09-11.
	`{"Fassung":1,"System":"scryfall","Semantik":"druck","Wert":"w-1",` +
		`"Name":"Karte 3","Setcode":"m11","Nummer":"2","Sprache":"en"}`,
	`{"Fassung":99,"System":"scryfall","Semantik":"druck","Wert":"w-1"}`,
	`{"Fassung":1,"System":"mtgjson","Semantik":"karte","Wert":"w-1"}`,
	`{"Fassung":1,"System":"","Semantik":"","Wert":"w-1"}`,
	`null`,
	`[]`,
}

// kandidaten baut Drucke, unter denen ein Verweis suchen kann.
//
// **Zwei tragen dieselbe Kennung** — sonst gäbe es den Konflikt nie, und die
// Zusage darüber wäre ungeprüft.
func kandidaten() []domain.Druck {
	gebaut := make([]domain.Druck, 0, 4)

	for nummer := range 4 {
		druck := domain.Druck{
			ID:   int64(nummer + 1),
			Name: fmt.Sprintf("Karte %d", nummer+1),
			// **Zwei mit demselben Setcode und derselben Nummer**: auch der
			// Rückfall soll mehrdeutig werden können.
			Setcode: "m11", Nummer: fmt.Sprintf("%d", nummer/2+1), Sprache: "en",
			Kennungen: []domain.Druckkennung{{
				System: domain.SystemScryfall, Semantik: domain.SemantikDruck,
				Wert: fmt.Sprintf("w-%d", nummer/2+1),
			}},
		}

		gebaut = append(gebaut, druck)
	}

	return gebaut
}

// **Keine Auflösung behauptet mehr, als sie weiß.**
func FuzzLoeseAuf(f *testing.F) {
	for _, saat := range verweissaat {
		f.Add([]byte(saat))
	}

	unter := kandidaten()

	f.Fuzz(func(t *testing.T, roh []byte) {
		var verweis domain.Druckverweis

		if json.Unmarshal(roh, &verweis) != nil {
			// Was kein Verweis ist, erreicht die Auflösung nie: davor steht
			// der Leser des Formats, das ihn trägt.
			return
		}

		ergebnis := domain.LoeseAuf(unter, verweis)

		if nochmal := domain.LoeseAuf(unter, verweis); !reflect.DeepEqual(ergebnis, nochmal) {
			t.Fatalf("derselbe Verweis gab zwei Auflösungen: %+v gegen %+v",
				ergebnis, nochmal)
		}

		switch ergebnis.Art {
		case domain.AufloesungEindeutig:
			if len(ergebnis.DruckIDs) != 1 {
				t.Errorf("»eindeutig« mit %d Drucken: %+v",
					len(ergebnis.DruckIDs), ergebnis)
			}

		case domain.AufloesungMehrdeutig:
			if len(ergebnis.DruckIDs) < 2 {
				t.Errorf("»mehrdeutig« mit %d Drucken: %+v",
					len(ergebnis.DruckIDs), ergebnis)
			}

		case domain.AufloesungUnbekannt:
			if len(ergebnis.DruckIDs) != 0 {
				t.Errorf("»unbekannt« mit %d Drucken: %+v",
					len(ergebnis.DruckIDs), ergebnis)
			}

		default:
			t.Errorf("eine vierte Art: %q", ergebnis.Art)
		}

		pruefeRueckfall(t, unter, verweis, ergebnis)
	})
}

// pruefeRueckfall löst die Zusage aus dem Kopf ein.
//
// **Befund 2 des Nachlaufs zum Zweitblick über Phase 5.2**: der Kopf nannte
// „eine Kennung, die nicht trifft, und trotzdem ein Rückfall über einen
// unvollständigen Namen" ausdrücklich als Fund — und der Rumpf verglich
// ausschließlich Ergebnisart und Trefferzahl. Eine falsche eindeutige
// Auflösung über einen halben Namen wäre grün geblieben, solange sie genau
// eine Kennung liefert.
//
// **Zwei Aussagen, und beide gehören dazu**: ohne vollständige Angabe und ohne
// Kennungstreffer darf nichts herauskommen — und **trifft** eine Kennung, ist
// das Ergebnis genau ihre Menge und nicht die des Rückfalls.
func pruefeRueckfall(
	t *testing.T, unter []domain.Druck,
	verweis domain.Druckverweis, ergebnis domain.Aufloesung,
) {
	t.Helper()

	for _, kennung := range ergebnis.DruckIDs {
		if !slices.ContainsFunc(unter, func(druck domain.Druck) bool {
			return druck.ID == kennung
		}) {
			t.Errorf("die Auflösung nennt den Druck %d, den es nicht gibt", kennung)
		}
	}

	ueberKennung := kennungstreffer(unter, verweis)

	if len(ueberKennung) > 0 {
		if !slices.Equal(ueberKennung, ergebnis.DruckIDs) {
			t.Errorf("die Kennung traf %v, aufgelöst wurde %v",
				ueberKennung, ergebnis.DruckIDs)
		}

		return
	}

	// **Der Rückfall gilt nur ganz.**
	if strings.TrimSpace(verweis.Name) != "" &&
		strings.TrimSpace(verweis.Setcode) != "" &&
		strings.TrimSpace(verweis.Nummer) != "" {
		return
	}

	if ergebnis.Art != domain.AufloesungUnbekannt {
		t.Errorf("ohne treffende Kennung und ohne volle Angabe kam %+v", ergebnis)
	}
}

// kennungstreffer gibt die Drucke, die die Kennung des Verweises tragen.
//
// **Nachgebaut und nicht aufgerufen**: `LoeseAuf` ist das, was geprüft wird —
// dieselbe Auskunft von ihm zu holen prüfte nichts (§10).
func kennungstreffer(unter []domain.Druck, verweis domain.Druckverweis) []int64 {
	kennung, hat := verweis.Kennung()
	if !hat {
		return nil
	}

	var treffer []int64

	for _, druck := range unter {
		for _, eigene := range druck.Kennungen {
			if eigene.System == kennung.System &&
				eigene.Semantik == kennung.Semantik && eigene.Wert == kennung.Wert {
				treffer = append(treffer, druck.ID)

				break
			}
		}
	}

	return treffer
}
