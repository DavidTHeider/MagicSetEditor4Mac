package domain_test

// **Was ein Gegenüber von einem Exemplar sieht** — Werkstück `T3` von Phase
// 5.6 (#1468).
//
// # Warum das über die serialisierte Nachricht geprüft wird
//
// **Weil ein Feldvergleich die Frage nicht beantwortet.** „Trägt die Sicht
// keinen Erwerbspreis?" lässt sich am Typ ablesen; „steht der Erwerbspreis
// nicht in dem, was hinausgeht?" nicht. Zwischen beidem liegt jede Marke, die
// jemand später hinzufügt, jedes eingebettete Feld und jeder Umweg über einen
// Text — und die Abnahme des Issues verlangt ausdrücklich den **tatsächlich
// serialisierten** Stand.
//
// # Und warum gegen die Werte und nicht gegen die Feldnamen
//
// **Weil ein umbenanntes Feld denselben Wert trägt.** Wer nach `"lagerort"`
// suchte, fände `"ort"` nicht; wer nach `"Kiste 4"` sucht, findet beides.

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// vollesExemplar ist ein Exemplar, in dem **jedes** Feld steht.
//
// **Jedes, und das ist der Punkt**: die Abnahme verlangt den Nachweis am
// Ausgangsstück, das alles enthält. Ein Exemplar mit leeren Feldern belegte
// nur, dass Leeres nicht hinausgeht.
func vollesExemplar() domain.Exemplar {
	return domain.Exemplar{
		ID: 4711, Fassung: 9, DruckID: 815,
		Printing: "a1b2c3d4-0000-4000-8000-00000000000a",
		Menge:    3, Zustand: domain.ZustandPlayed, Sprache: "de",
		Finish:       domain.FinishEtched,
		Erwerbsdatum: "2026-08-01", Erwerbspreis: 417, Erwerbswaehrung: "EUR",
		Notiz: "getauscht gegen den Drachen",
		Name:  "Lightning Bolt", Setcode: "lea", Nummer: "161",
		Lagerort: "Kiste 4",
		Herkunft: domain.Herkunft{
			Art: domain.HerkunftProfil, Quelle: "Zweitprofil",
			QuellProfil: 77, Uebernommen: "2026-07-01T10:00:00Z",
		},
	}
}

// **Was privat ist, steht nicht in der Nachricht — gemessen an den Bytes.**
func TestEineSichtkarteTraegtNichtsPrivates(t *testing.T) {
	t.Parallel()

	roh, err := json.Marshal(domain.AlsSichtkarte(vollesExemplar(), "s-7"))
	if err != nil {
		t.Fatalf("kodieren: %v", err)
	}

	nachricht := string(roh)

	for name, wert := range map[string]string{
		"die lokale Kennung":  "4711",
		"die Druckzeile":      "815",
		"die Fassung":         `"fassung"`,
		"der Erwerbspreis":    "417",
		"das Erwerbsdatum":    "2026-08-01",
		"die Währung":         "EUR",
		"die Notiz":           "getauscht",
		"der Lagerort":        "Kiste 4",
		"die Herkunftsquelle": "Zweitprofil",
		"das Quellprofil":     "77",
		"der Übernahmetag":    "2026-07-01",
	} {
		if strings.Contains(nachricht, wert) {
			t.Errorf("%s steht in der Nachricht: %s", name, nachricht)
		}
	}
}

// **Und was hinaus darf, steht auch wirklich drin.**
//
// **Die Gegenprobe zur Zeile darüber** (§10): ohne sie hielte auch eine Sicht,
// die gar nichts trägt — und die wäre für jeden Handel wertlos.
func TestEineSichtkarteTraegtWasSieTragenSoll(t *testing.T) {
	t.Parallel()

	sicht := domain.AlsSichtkarte(vollesExemplar(), "s-7")

	roh, err := json.Marshal(sicht)
	if err != nil {
		t.Fatalf("kodieren: %v", err)
	}

	nachricht := string(roh)

	for name, wert := range map[string]string{
		"die Druckreferenz":              "a1b2c3d4-0000-4000-8000-00000000000a",
		"die Menge":                      "3",
		"der Zustand":                    "played",
		"die Sprache":                    "de",
		"die Ausführung":                 "etched",
		"die lesbare Bezeichnung":        "Lightning Bolt",
		"die sitzungsgebundene Referenz": "s-7",
	} {
		if !strings.Contains(nachricht, wert) {
			t.Errorf("%s fehlt in der Nachricht: %s", name, nachricht)
		}
	}
}

// **Die Referenz kommt von außen und nicht aus dem Bestand.**
//
// **Sonst wäre sie die Datenbankkennung mit einem anderen Namen.** Das Issue
// sagt es wörtlich: eine **opake, sitzungsgebundene** Referenz statt der
// lokalen Kennung — zwei Sitzungen dürfen für dasselbe Stück nicht dieselbe
// Zeichenkette hören, sonst ist sie über Sitzungen hinweg wiedererkennbar.
func TestDieReferenzIstNichtDieKennung(t *testing.T) {
	t.Parallel()

	erste := domain.AlsSichtkarte(vollesExemplar(), "s-7")
	zweite := domain.AlsSichtkarte(vollesExemplar(), "s-8")

	if erste.Referenz == zweite.Referenz {
		t.Errorf("beide Sitzungen hören %q", erste.Referenz)
	}

	if strings.Contains(erste.Referenz, "4711") {
		t.Errorf("die Referenz trägt die Kennung: %q", erste.Referenz)
	}
}

// **Eine Ausführung ohne Angabe ist nonfoil — auch nach draußen.**
//
// **Sonst stünde dort eine leere Zeichenkette**, und das Gegenüber hätte ein
// viertes Wort für etwas, wofür es drei gibt (T2, 0119).
func TestEineSichtkarteBeantwortetDenNullwert(t *testing.T) {
	t.Parallel()

	nackt := domain.Exemplar{Printing: "abc", Menge: 1}

	if wie := domain.AlsSichtkarte(nackt, "s-1").Finish; wie != domain.FinishNonfoil {
		t.Errorf("die Ausführung ist %q", wie)
	}
}

// **Eine abgelaufene Freigabe zeigt nichts.**
func TestEineAbgelaufeneFreigabeZeigtNichts(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)

	freigabe := domain.Sammlungsfreigabe{
		Empfaenger: "peer-1", Umfang: domain.UmfangSammlung,
		Zweck: domain.ZweckHandel, Ablauf: jetzt.Add(-time.Minute),
	}

	if freigabe.Gilt(jetzt) {
		t.Error("die abgelaufene Freigabe gilt noch")
	}

	// **Ohne Ablauf gilt sie** — der Nullwert heißt „unbefristet" und nicht
	// „schon vorbei".
	unbefristet := freigabe
	unbefristet.Ablauf = time.Time{}

	if !unbefristet.Gilt(jetzt) {
		t.Error("die unbefristete Freigabe gilt nicht")
	}
}

// **Ein Wunsch sagt, dass er einer ist** — P11
// (#2379).
// Ohne das Kennzeichen sah eine freigegebene Wunschzeile beim Gegenüber aus
// wie ein Angebot: es meinte, bekommen zu können, was hier gesucht wird.
func TestEineWunschzeileTraegtIhrKennzeichen(t *testing.T) {
	t.Parallel()

	wunsch := vollesExemplar()
	wunsch.Wunsch = true

	if !domain.AlsSichtkarte(wunsch, "s-1").Wunsch {
		t.Error("die Wunschzeile kam ohne Kennzeichen hinaus")
	}

	// **Und ein Angebot trägt es nicht** — auch nicht als `false` im Draht:
	// eine ältere Fassung kennt das Feld nicht und braucht es nicht.
	roh, err := json.Marshal(domain.AlsSichtkarte(vollesExemplar(), "s-2"))
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(roh), "wunsch") {
		t.Errorf("ein Angebot nennt das Kennzeichen: %s", roh)
	}
}
