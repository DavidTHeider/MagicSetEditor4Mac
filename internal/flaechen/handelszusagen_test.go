package flaechen_test

// **Ein Handel, bei dem diese Seite etwas weggibt** —
// #1643.
//
// # Was hier gemessen wird
//
// **Dass ein Exemplar wirklich festgehalten wird und der Abgang ankommt.**
// `storage.Reservierungsablage.Festhalten` stand gebaut, geprüft und zu
// 100 % abgedeckt da — und **ohne Aufrufer**. Damit konnte nur die
// **empfangende** Seite verbuchen: wer Karten weggab, bekam „es fehlen N
// Stück an bestätigter Übergabe", richtig gemeldet und ohne einen Weg, es in
// Ordnung zu bringen.
//
// **Der Lauf von `T1` daneben gibt nichts weg** — dort kommt alles von
// drüben, und genau deshalb fiel es dort nicht auf.

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// derAbgangshandel ist der Vorgang, bei dem etwas weggeht.
const derAbgangshandel = "v-geben"

// derZusagenweg ist der Endpunkt, wie er im Katalog steht.
const derZusagenweg = "GET /handel/{vorgang}/zusagen"

func TestEinHandelGibtEtwasWegUndDerAbgangKommtAn(t *testing.T) {
	umgebung := append(mitEigenerSaat(t),
		schluesselbund.Bundort+"="+probebund(t))
	an := starteDienst(t, umgebung)

	ich := zeilenwert(t,
		mussKommandozeile(t, umgebung, "peer-anlegen"), "kennung")

	// **Drei Stück im Bestand**, und zwei davon gehen weg — in der Variante
	// des Tauschs (#2492).
	angelegt := mussKommandozeile(t, umgebung,
		"--json", "exemplar", "bolt-m10", "3", "nearMint", "de", "nonfoil")

	exemplarID := eineKennungAus(t, angelegt)

	positionen := `[{"von":"` + ich +
		`","printing":"bolt-m10","menge":2,"guete":"nearMint",` +
		`"sprache":"de","finish":"nonfoil"}]`

	mitEingabe(t, umgebung, positionen,
		"handel", "eroeffnen", derAbgangshandel, "tausch", dasGegenueberImHandel)

	mussKommandozeile(t, umgebung, "handel", "anbieten", derAbgangshandel)
	mussKommandozeile(t, umgebung, "handel", "zustimmen", derAbgangshandel)
	schrittAmVorgang(t, an, derAbgangshandel, "zustimmen",
		dasGegenueberImHandel)

	// **Ohne Festgehaltenes geht `vorbereiten` nicht** — der Schritt heißt
	// „die Exemplare sind reserviert", und bis zu diesem Werkstück bewegte
	// er den Zustand und hielt nichts fest.
	_, _, code := kommandozeile(t, umgebung,
		"handel", "vorbereiten", derAbgangshandel)
	if code == 0 {
		t.Fatal("vorbereiten ging ohne festgehaltene Exemplare durch")
	}

	mussKommandozeile(t, umgebung, "handel-festhalten",
		derAbgangshandel, exemplarID, "2", "1")

	// **Mehr als dasteht, wird abgewiesen** — ohne Bestandswirkung.
	_, _, code = kommandozeile(t, umgebung, "handel-festhalten",
		derAbgangshandel, exemplarID, "9", "1")
	if code == 0 {
		t.Fatal("neun von drei Stück wurden festgehalten")
	}

	zusagenGleich(t, umgebung, an, 1)

	mussKommandozeile(t, umgebung, "handel", "vorbereiten", derAbgangshandel)

	mussKommandozeile(t, umgebung, "handel", "uebergeben", derAbgangshandel)
	schrittAmVorgang(t, an, derAbgangshandel, "uebergeben",
		dasGegenueberImHandel)

	// **Und jetzt gilt sie als übergeben** — genau daran findet die
	// Verbuchung ihre Abgangszeilen.
	uebergeben := mussKommandozeile(t, umgebung,
		"--json", "handelszusagen", derAbgangshandel)

	if !strings.Contains(uebergeben, `"uebergeben": true`) {
		t.Fatalf("die Zusage gilt nicht als übergeben: %s", uebergeben)
	}

	abgangschluss(t, umgebung, an, ich, exemplarID)
}

// eineKennungAus liest die Zahl hinter `"id"` aus einer JSON-Antwort.
//
// **Als Zeichenkette und nicht als Zahl**: sie geht als Argument wieder
// hinaus, und `float64` machte aus einer 7 eine `7e+00`.
func eineKennungAus(t *testing.T, roh string) string {
	t.Helper()

	var gelesen struct {
		ID int64 `json:"id"`
	}

	if fehler := json.Unmarshal([]byte(roh), &gelesen); fehler != nil {
		t.Fatalf("die Antwort ist kein JSON: %v\n%s", fehler, roh)
	}

	if gelesen.ID == 0 {
		t.Fatalf("die Antwort nennt keine Kennung: %s", roh)
	}

	return strconv.FormatInt(gelesen.ID, 10)
}

// zusagenGleich liest die Zusagen über beide Flächen und zählt sie.
//
// **Beide und nicht eine** (§16): eine Fläche, die sich selbst bestätigte,
// bestätigte nichts.
func zusagenGleich(
	t *testing.T, umgebung []string, an dienst, erwartet int,
) {
	t.Helper()

	aus := mussKommandozeile(t, umgebung,
		"--json", "handelszusagen", derAbgangshandel)

	if gezaehlt := strings.Count(aus, `"exemplarId"`); gezaehlt != erwartet {
		t.Errorf("die Kommandozeile nennt %d Zusagen: %s", gezaehlt, aus)
	}

	ueberDenDienst := mussREST(t, an, http.MethodGet,
		strings.ReplaceAll(pfadAus(derZusagenweg), "{vorgang}",
			derAbgangshandel), "")

	if gezaehlt := strings.Count(
		ueberDenDienst, `"exemplarId"`); gezaehlt != erwartet {
		t.Errorf("der Dienst nennt %d Zusagen: %s", gezaehlt, ueberDenDienst)
	}
}

// abgangschluss führt den Handel zu Ende und prüft den Abgang.
func abgangschluss(
	t *testing.T, umgebung []string, an dienst, ich, exemplarID string,
) {
	t.Helper()

	mussKommandozeile(t, umgebung, "handel", "verbuchen", derAbgangshandel,
		eineFreigabeFuer(t, umgebung, derAbgangshandel, "verbuchen", ""))
	mussKommandozeile(t, umgebung, "handel", "abgleichen", derAbgangshandel)

	for _, wer := range []string{ich, dasGegenueberImHandel} {
		mussKommandozeile(t, umgebung, "handel", "bestaetigen",
			derAbgangshandel, wer, eineFreigabeFuer(
				t, umgebung, derAbgangshandel, "bestaetigen", wer))
	}

	// **Und im Journal steht ein Abgang mit seiner Besitzzeile.**
	//
	// **Das ist die Abnahme dieses Werkstücks**: ohne die Zusage fände
	// `Verbuchungsstelle.abgaengeFuer` nichts und bräche mit „es fehlen
	// zwei Stück an bestätigter Übergabe" ab.
	journal := mussKommandozeile(t, umgebung,
		"--json", "handelsjournal", derAbgangshandel)

	if !strings.Contains(journal, `"art": "abgang"`) {
		t.Errorf("im Journal steht kein Abgang: %s", journal)
	}

	if !strings.Contains(journal, `"exemplarId": `+exemplarID) {
		t.Errorf("der Abgang nennt seine Besitzzeile nicht: %s", journal)
	}

	// **Und danach ist der Vorgang abgeschlossen** — die Verbuchung räumt
	// die Zusagen mit, und eine, die stehenbliebe, hielte Bestand für einen
	// Handel fest, den es nicht mehr gibt.
	zusagenGleich(t, umgebung, an, 0)
}
