//go:build netzlauf || dienstdrift

// **Der Lauf gegen den echten Dienst** — Werkstück A14 von Phase 5.3
// (#1453),
// nachgetragen zu A3
// (#1442).
//
// # Warum es ihn erst jetzt gibt
//
// **`docs/HAERTUNG.md` sagte „nie gegen das Original gelaufen"**, und das war
// richtig: geprüft war die Zuordnung gegen eingecheckte, selbst geschriebene
// Beispiele. Was daran fehlte, ist genau das, wogegen Vertragstests nicht
// helfen — ob der Dienst heute noch so antwortet, wie dieses Programm es
// annimmt.
//
// # Was er prüft, das kein anderer Test kann
//
//   - **Die Adresse.** `https://mtgjson.com/api/v5` — steht als Konstante da
//     und war nie gefragt worden.
//   - **Den Dateinamen.** `AllPricesToday.json.gz`, aus der Aufzählung in
//     `holen.go`.
//   - **Das Packformat.** Dieses Programm holt gzip, weil Go das von sich aus
//     kann; ob MTGJSON es unter diesem Namen noch ausliefert, sagt nur ein
//     Ruf.
//   - **Das Strömen.** Die Datei wird nicht in den Speicher gelegt, sondern
//     gelesen und gefiltert — an einer echten Datei mit echter Größe.
//
// # Warum **heute** und nicht neunzig Tage
//
// **`AllPrices.json` wiegt über ein Gigabyte.** Ein Lauf dagegen verbrauchte
// von einer Grenze, die anderen gehört, und sagte über den Vertrag nichts
// mehr als die kleine Datei (§3, §16 Regel 4).
//
// # Warum er nicht in der Pipeline läuft
//
// Er befragt einen fremden Rechner: er scheitert, wenn jemand anderes ein
// Problem hat, und er lädt Daten, die niemand für einen Testlauf bezahlt hat.
//
//	go test -tags netzlauf ./internal/mtgjsonpreise/

package mtgjsonpreise_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjsonpreise"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// echterPreisklient baut ihn gegen den echten Dienst.
//
// **Mit einer langen Stillstandsfrist**: die Datei ist groß, und der Strom
// darf zwischen zwei Bündeln stehen, ohne dass der Lauf ihn aufgibt.
func echterPreisklient() *mtgjsonpreise.Klient {
	return mtgjsonpreise.NeuerKlient("", netz.NeuerKlient(netz.Fristen{
		Verbindung: 15 * time.Second,
		Kopf:       60 * time.Second,
		Stillstand: 5 * time.Minute,
	}), netz.Kenner{Name: "MagicSetEditor4Mac/netzlauf", Annahme: "*/*"},
		mtgjsonpreise.Grenzen{})
}

// **Der Dienst liefert die Datei, und sie lässt sich strömend lesen.**
//
// **Mit leerem Crosswalk**: keine Kennung des Bestands passt, also kommt kein
// Preisstand zurück — und **genau das ist die Zusage** („überspringt, was der
// Bestand nicht kennt, **ohne es zu lesen**"). Was dieser Lauf belegt, ist,
// dass die Datei ankommt, gzip ist und bis zum Ende gelesen wird.
func TestDerEchteMTGJSONDienstLiefertDieTagesdatei(t *testing.T) {
	staende, befunde, fehler := echterPreisklient().Holen(t.Context(),
		string(mtgjsonpreise.UmfangHeute), map[string]string{})
	if fehler != nil {
		t.Fatalf("holen: %v", fehler)
	}

	if len(staende) != 0 {
		t.Errorf("mit leerem Crosswalk kamen %d Stände", len(staende))
	}

	// **Befunde sind kein Fehler** — aber es dürfen keine sein, die von einer
	// Drift sprechen.
	for _, befund := range befunde {
		t.Logf("Befund: %s — %s", befund.Ort, befund.Meldung)
	}
}

// **Und ein bekannter Druck kommt als Preisstand zurück.**
//
// # Woher die Kennung kommt
//
// **Aus der Kartendatei desselben Anbieters** — `internal/mtgjson` holt sie im
// Lauf daneben. Eine hier eingetragene Kennung wäre eine Behauptung über
// fremde Daten, die morgen nicht mehr stimmt; diese hier ist am selben Tag
// aus derselben Quelle gelesen.
//
// **Ohne Kennung wird übersprungen**, und dann sagt der Lauf es und endet —
// eine leere Zusicherung wäre schlimmer als keine.
func TestEinBekannterDruckKommtAlsPreisstand(t *testing.T) {
	kennung := eineEchteDruckkennung(t)

	staende, _, fehler := echterPreisklient().Holen(t.Context(),
		string(mtgjsonpreise.UmfangHeute),
		map[string]string{kennung.uuid: kennung.scryfall})
	if fehler != nil {
		t.Fatalf("holen: %v", fehler)
	}

	if len(staende) == 0 {
		t.Fatalf("zu %s kam kein Preisstand — hat MTGJSON die Karte heute "+
			"nicht bepreist, oder liest dieses Programm falsch?", kennung.uuid)
	}

	eine := staende[0]

	if eine.Printing != kennung.scryfall {
		t.Errorf("die Zuordnung ging daneben: %+v", eine)
	}

	// **Die sieben Teile einer Preisreihe** (0093) — hier zählt, dass sie
	// überhaupt gefüllt sind: eine Reihe ohne Währung oder Markt ordnet
	// nichts.
	if eine.Quelle == "" || eine.Markt == "" || eine.Art == "" ||
		eine.Waehrung == "" {
		t.Errorf("die Reihe ist unvollständig: %+v", eine)
	}

	t.Logf("gelesen: %d Stände, erster %+v", len(staende), eine)
}

// druckkennung ist ein Druck, wie MTGJSON ihn führt.
type druckkennung struct{ uuid, scryfall string }

// eineEchteDruckkennung holt eine aus einer Setdatei desselben Anbieters.
//
// # Warum sie geholt und nicht eingetragen wird
//
// **Weil eine eingetragene Kennung eine Behauptung über fremde Daten wäre**,
// die morgen nicht mehr stimmt — und dann wäre der Lauf rot, ohne dass sich
// an diesem Programm etwas geändert hätte (§10).
//
// **`LEA` ist bewusst gewählt**: das erste Set von 1993, ein paar hundert
// Kilobyte, und wenn es einmal verschwände, wäre das eine Nachricht über den
// Dienst und keine über diesen Lauf.
func eineEchteDruckkennung(t *testing.T) druckkennung {
	t.Helper()

	anfrage, fehler := http.NewRequestWithContext(t.Context(),
		http.MethodGet, mtgjsonpreise.Vorgabeadresse+"/LEA.json", http.NoBody)
	if fehler != nil {
		t.Fatalf("die Anfrage bauen: %v", fehler)
	}

	anfrage.Header.Set("User-Agent", "MagicSetEditor4Mac/netzlauf")

	antwort, fehler := http.DefaultClient.Do(anfrage) //nolint:bodyclose // unten
	if fehler != nil {
		t.Fatalf("die Setdatei holen: %v", fehler)
	}

	defer func() { _ = antwort.Body.Close() }() //nolint:errcheck // gelesen ist gelesen

	if antwort.StatusCode != http.StatusOK {
		t.Fatalf("die Setdatei gab %d", antwort.StatusCode)
	}

	var gelesen struct {
		Daten struct {
			Karten []struct {
				UUID      string `json:"uuid"`
				Kennungen struct {
					Scryfall string `json:"scryfallId"`
				} `json:"identifiers"`
			} `json:"cards"`
		} `json:"data"`
	}

	if fehler := json.NewDecoder(antwort.Body).Decode(&gelesen); fehler != nil {
		t.Fatalf("die Setdatei lesen: %v", fehler)
	}

	for _, karte := range gelesen.Daten.Karten {
		if karte.UUID != "" && karte.Kennungen.Scryfall != "" {
			return druckkennung{uuid: karte.UUID, scryfall: karte.Kennungen.Scryfall}
		}
	}

	t.Fatalf("in LEA.json stand keine Karte mit UUID und Scryfall-Kennung")

	return druckkennung{}
}
