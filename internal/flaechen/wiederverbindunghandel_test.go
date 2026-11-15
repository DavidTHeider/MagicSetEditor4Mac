package flaechen_test

// **Der Handel nach einem Ausfall: gemessen und verglichen** — D01
// (#2232).
//
// QA-10 hält fest, **dass** eine Nachricht nach einem Ausfall im Ausgang
// bleibt und mit dem nächsten Aufruf hinübergeht (`netzfaelle_test.go`).
// Hier stehen die drei Dinge, die D01 dazu verlangt:
//
//   - **die Zeit** vom Startereignis — der Weg ist wieder frei, und Bert ruft
//     die Zustellung — bis der Vorgang bei Anna dasteht, gegen eine Frist von
//     10 s;
//   - **der fachliche Vergleich beider Endstände** — Annas und Berts
//     `handel stand` nennen denselben Vorgang mit derselben Revision,
//     demselben **Abdruck**, denselben Parteien, Positionen und Zustimmungen;
//     verschieden darf nur sein, was aus der eigenen Sicht spricht (`ich`,
//     `lage`, `naechster`, `moegliche`, `phase`);
//   - **keine Doppelung** — der Ausgang räumt erst, wenn ein Mensch die
//     Ankunft bestätigt; eine zweite Zustellung trägt dieselbe Nachricht noch
//     einmal, und Annas Stand bleibt, wie er war.
//
// **Was er nicht zeigt**: zwei Rechner — der Vermittler ist ein Prozess auf
// dieser Maschine.

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// handelsstand ist, was beide Seiten über denselben Vorgang sagen.
type handelsstand struct {
	Vorgang      string          `json:"vorgang"`
	Revision     int             `json:"revision"`
	Abdruck      string          `json:"abdruck"`
	Art          string          `json:"art"`
	Initiator    string          `json:"initiator"`
	Parteien     []string        `json:"parteien"`
	Zustand      string          `json:"zustand"`
	Positionen   json.RawMessage `json:"positionen"`
	Zustimmungen []string        `json:"zustimmungen"`
	Ende         bool            `json:"ende"`
}

func standBei(t *testing.T, rechner einRechner) handelsstand {
	t.Helper()

	roh := mussKommandozeile(t, rechner.umgebung, "--json", "handel", "stand", derDrahtvorgang)

	var stand handelsstand
	if fehler := json.Unmarshal([]byte(roh), &stand); fehler != nil {
		t.Fatalf("den Stand lesen: %v\n%s", fehler, roh)
	}

	return stand
}

func TestNachDemAusfallGleichenSichBeideHandelsstaende(t *testing.T) {
	anna, bert := zweiImLabor(t)

	adresse, protokoll := horcheImHintergrund(t, anna)
	weg := neuerVermittler(t, adresse)

	mussKommandozeile(t, bert.umgebung, "kontakt-adresse", anna.kennung, weg.adresse())
	einHandelImAusgang(t, anna, bert)

	weg.sperren(true)

	if code, _ := zustellen(t, bert, anna); code == 0 {
		t.Fatal("die Zustellung meldete Erfolg, obwohl der Weg fort war")
	}

	nichtsBei(t, anna)

	// **Startereignis**: der Weg ist frei, und Bert stellt zu.
	beginn := time.Now()
	weg.sperren(false)

	if code, klage := zustellen(t, bert, anna); code != 0 {
		t.Fatalf("nach dem Ausfall: %d %s", code, klage)
	}

	// **Endereignis**: der Vorgang steht bei Anna.
	warteAufDenVorgang(t, anna, protokoll)

	dauer := time.Since(beginn)
	t.Logf("gemessen: Handel nach dem Ausfall — bei Anna nach %v", dauer.Round(time.Millisecond))

	if dauer > 10*time.Second {
		t.Errorf("der Abgleich brauchte %v", dauer)
	}

	beiAnna, beiBert := standBei(t, anna), standBei(t, bert)
	if beiAnna.Vorgang == "" || beiAnna.Abdruck == "" || !reflect.DeepEqual(beiAnna, beiBert) {
		t.Errorf("die Stände gleichen sich nicht:\nAnna %+v\nBert %+v", beiAnna, beiBert)
	}

	// **Keine Doppelung**: dieselbe Nachricht noch einmal ändert drüben
	// nichts.
	if code, klage := zustellen(t, bert, anna); code != 0 {
		t.Fatalf("die zweite Zustellung: %d %s", code, klage)
	}

	if danach := standBei(t, anna); !reflect.DeepEqual(danach, beiAnna) {
		t.Errorf("die zweite Zustellung änderte Annas Stand:\nvorher %+v\nnachher %+v", beiAnna, danach)
	}
}
