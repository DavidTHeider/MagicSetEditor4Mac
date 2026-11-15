package peer_test

// **Derselbe Inhalt in anderer Darstellung** — QA-FZ01
// (#2091).
//
// Der Fuzzlauf zum Handschlag fand `d7e6246289f77b4d`: das ehrliche Angebot,
// in der Unterschrift ein großes `D` und ein Leerzeichen dahinter. Angenommen
// — und das ist richtig, denn die Unterschrift gilt den **dekodierten**
// Bytes ([peer.Huelle.Pruefe] liest sie mit `hex.DecodeString`). Rot war das
// Orakel des Fuzzziels, das die gelesene Hülle samt Hex-Zeichenkette verglich.
//
// Hier steht der Vertrag, den beide halten: **gleicher Inhalt ist gleich,
// gleich wie er geschrieben ist** — und alles andere bleibt abgewiesen,
// auch eine zweite Annahme desselben Angebots in anderer Schreibung.

import (
	"regexp"
	"strings"
	"testing"
)

// umgeschrieben gibt die Nachricht mit großgeschriebener Unterschrift und
// Leerraum — derselbe Inhalt in anderer Darstellung.
func umgeschrieben(t *testing.T, ehrlich []byte) []byte {
	t.Helper()

	return []byte(" " + strings.Replace(string(ehrlich), `"unterschrift":"`, `"unterschrift" : "`, 1) + " \n")
}

// unterschriftIn findet die Hex-Unterschrift einer Nachricht.
var unterschriftIn = regexp.MustCompile(`"unterschrift":"([0-9a-fA-F]+)"`)

// ersetzt tauscht genau eine Stelle der Nachricht — und nur sie. **Nicht über
// ein Neuschreiben**: `json.Marshal` ordnete den Rumpf um, und die
// Unterschrift bräche an einer Änderung, die der Fall gar nicht meint.
func ersetzt(t *testing.T, ehrlich []byte, alt, neu string) []byte {
	t.Helper()

	if strings.Count(string(ehrlich), alt) != 1 {
		t.Fatalf("%q steht nicht genau einmal in der Nachricht", alt)
	}

	return []byte(strings.Replace(string(ehrlich), alt, neu, 1))
}

// unterschrift gibt die Hex-Unterschrift der Nachricht.
func unterschrift(t *testing.T, ehrlich []byte) string {
	t.Helper()

	treffer := unterschriftIn.FindSubmatch(ehrlich)
	if treffer == nil {
		t.Fatal("keine Unterschrift")
	}

	return string(treffer[1])
}

// gross schreibt die Hex-Unterschrift einer Nachricht groß.
func gross(t *testing.T, ehrlich []byte) []byte {
	t.Helper()

	u := unterschrift(t, ehrlich)

	return ersetzt(t, ehrlich, u, strings.ToUpper(u))
}

func TestDasOrakelVergleichtDenInhaltNichtDieSchreibung(t *testing.T) {
	t.Parallel()

	for stufe, ehrlich := range saatnachrichten(t) {
		for name, anders := range map[string][]byte{
			"großes Hex":           gross(t, ehrlich),
			"Leerraum":             umgeschrieben(t, ehrlich),
			"großes Hex, Leerraum": umgeschrieben(t, gross(t, ehrlich)),
		} {
			if !gleicherInhalt(anders, ehrlich) {
				t.Errorf("Stufe %d, %s: derselbe Inhalt gilt als anderer", stufe, name)
			}
		}

		geaendert := ersetzt(t, ehrlich, `"absender":"`, `"absender":"x`)
		if gleicherInhalt(geaendert, ehrlich) {
			t.Errorf("Stufe %d: ein anderer Kopf gilt als derselbe Inhalt", stufe)
		}
	}
}

func TestDasAngebotInAndererSchreibungGiltUndNurEinmal(t *testing.T) {
	t.Parallel()

	ich, du := beide(t)
	wunsch, bereit := wunschUndBereitschaft()

	_, angebot, fehler := ich.Anbieten(wunsch, &festerZufall{stand: 7})
	if fehler != nil {
		t.Fatal(fehler)
	}

	// **Die andere Schreibung gilt** — sie trägt dieselbe Unterschrift.
	eigene := liste()
	if _, _, fehler := du.Entgegennehmen(umgeschrieben(t, gross(t, angebot)), bereit, &festerZufall{stand: 200}, eigene); fehler != nil {
		t.Fatalf("das ehrliche Angebot in anderer Schreibung: %v", fehler)
	}

	// **Und nur einmal**: die ehrliche Schreibung danach ist dasselbe Angebot.
	if _, _, fehler := du.Entgegennehmen(angebot, bereit, &festerZufall{stand: 201}, eigene); fehler == nil {
		t.Fatal("dasselbe Angebot wurde in zwei Schreibungen zweimal angenommen")
	}
}

func TestEinAndererInhaltBleibtAbgewiesen(t *testing.T) {
	t.Parallel()

	ich, du := beide(t)
	wunsch, bereit := wunschUndBereitschaft()

	_, angebot, fehler := ich.Anbieten(wunsch, &festerZufall{stand: 7})
	if fehler != nil {
		t.Fatal(fehler)
	}

	_, fremdes, fehler := du.Anbieten(wunsch, &festerZufall{stand: 8})
	if fehler != nil {
		t.Fatal(fehler)
	}

	absender := regexp.MustCompile(`"absender":"([^"]+)"`)
	u := unterschrift(t, angebot)
	letzte := map[bool]string{true: "1", false: "0"}[u[len(u)-1] == '0']

	for name, nachricht := range map[string][]byte{
		"ein Kopffeld":         ersetzt(t, angebot, `"folge":0`, `"folge":1`),
		"ein Rumpffeld":        ersetzt(t, angebot, `"name":"A"`, `"name":"B"`),
		"ungültiges Hex":       ersetzt(t, angebot, u, "zz"+u[2:]),
		"falsche Unterschrift": ersetzt(t, angebot, u, u[:len(u)-1]+letzte),
		"fremde Identität": ersetzt(t, angebot, string(absender.Find(angebot)),
			string(absender.Find(fremdes))),
	} {
		if _, _, fehler := du.Entgegennehmen(nachricht, bereit, &festerZufall{stand: 200}, liste()); fehler == nil {
			t.Errorf("%s: angenommen", name)
		}
	}
}
