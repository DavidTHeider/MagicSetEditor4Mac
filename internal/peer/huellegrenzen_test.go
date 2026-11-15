package peer_test

// **Die Grenzen der Nachrichtenhülle** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Wogegen hier geprüft wird
//
// **Dass zu Großes und Unbekanntes vor der Fachausführung endet.** Das Issue
// verlangt es wörtlich: „unzulässige/zu große Typen vor Fachausführung
// abweisen". Ein Empfänger, der erst zerteilt und dann misst, hat die Arbeit
// schon getan, gegen die die Grenze schützt.

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// **Drei Protokolle, und ein viertes gibt es nicht.**
func TestNurDieDreiProtokolleGelten(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	for _, gilt := range []string{
		peer.ProtokollPaarung, peer.ProtokollTeilen, peer.ProtokollSpiel,
	} {
		if _, fehler := ich.Verpacke(eineHuelle(func(h *peer.Huelle) {
			h.Absender = string(ich.Kennung())
			h.Protokoll = gilt
		})); fehler != nil {
			t.Errorf("%s kam nicht durch: %v", gilt, fehler)
		}
	}

	// **Und die Fassung steht darin** — ein `/mse4mac/pair/2` ist ein
	// anderes Protokoll und keine Verhandlungssache.
	for _, giltNicht := range []string{
		"", "/mse4mac/pair/2", "/mse4mac/handel/1", "pair",
	} {
		if _, fehler := ich.Verpacke(eineHuelle(func(h *peer.Huelle) {
			h.Absender = string(ich.Kennung())
			h.Protokoll = giltNicht
		})); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%q kam durch: %v", giltNicht, fehler)
		}
	}
}

// **Zu Großes endet vor dem Zerteilen.**
//
// **Vor** und nicht danach: wer erst zerteilt und dann misst, hat die Arbeit
// schon getan, gegen die die Grenze schützt.
func TestZuGrossesEndetVorDemZerteilen(t *testing.T) {
	t.Parallel()

	// **Gültiges JSON, und sonst misst der Lauf nichts.** Der erste Anlauf
	// schickte eine abgeschnittene Zeichenkette; die scheitert auch ohne
	// jede Grenze am Zerteiler, und die Gegenprobe — die Grenze entfernen —
	// blieb grün (§10).
	zuviel := []byte(`{"protokoll":"` + peer.ProtokollPaarung +
		`","art":"` + strings.Repeat("x", peer.Hoechstlaenge) +
		`","id":"i","sitzung":"s","absender":"b","folge":0,` +
		`"rumpf":{},"unterschrift":""}`)

	fehler := peerAuspackeFehler(t, zuviel)

	// **Und die Grenze war es, die gegriffen hat.** Ohne diesen Satz wäre
	// der Lauf auch dann grün, wenn erst der Zerteiler oder die Feldgrenze
	// abgewiesen hätte — also genau dann, wenn die Zusage „vor dem
	// Zerteilen" nicht mehr gilt.
	if !strings.Contains(fehler.Error(), "die Nachricht wiegt") {
		t.Fatalf("abgewiesen hat etwas anderes: %v", fehler)
	}

	// **Und ein zu großer Rumpf endet beim Verpacken** — beim Absender, wo
	// er entsteht.
	ich := eineIdentitaet(t)

	gross := json.RawMessage(`{"was":"` + strings.Repeat("y", peer.Hoechstlaenge) + `"}`)

	if _, fehler := ich.Verpacke(eineHuelle(func(h *peer.Huelle) {
		h.Absender = string(ich.Kennung())
		h.Rumpf = gross
	})); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der zu große Rumpf kam durch: %v", fehler)
	}
}

// **Eine Hülle ohne Pflichtfeld ist keine.**
func TestEineHuelleOhnePflichtfeldIstKeine(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	faelle := map[string]func(*peer.Huelle){
		"ohne Art":      func(h *peer.Huelle) { h.Art = "" },
		"ohne ID":       func(h *peer.Huelle) { h.ID = "" },
		"ohne Sitzung":  func(h *peer.Huelle) { h.Sitzung = "" },
		"ohne Absender": func(h *peer.Huelle) { h.Absender = "" },
		"ohne Rumpf":    func(h *peer.Huelle) { h.Rumpf = nil },

		// **Eine negative Folge gibt es nicht** — sie zählt, und Zählen
		// fängt bei null an.
		"Folge unter null": func(h *peer.Huelle) { h.Folge = -1 },
	}

	for name, aendere := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := ich.Verpacke(eineHuelle(func(h *peer.Huelle) {
				h.Absender = string(ich.Kennung())
				aendere(h)
			})); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("es kam durch: %v", fehler)
			}
		})
	}
}

// **Was kein JSON ist, wird abgewiesen — und zwar als Befund.**
func TestWasKeinJSONIstWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, roh := range map[string][]byte{
		"leer":          {},
		"kein JSON":     []byte("hallo"),
		"halbe Klammer": []byte(`{"protokoll":`),
		"krummer Rumpf": []byte(`{"protokoll":"/mse4mac/pair/1","art":"a","id":"i","sitzung":"s","absender":"b","folge":0,"rumpf":{,},"unterschrift":"ab"}`),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := peer.Auspacke(roh); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("es kam durch: %v", fehler)
			}
		})
	}
}

// **Eine Unterschrift, die keine ist, prüft nicht.**
func TestEineUnterschriftDieKeineIstPrueftNicht(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	roh, fehler := ich.Verpacke(eineHuelle(func(h *peer.Huelle) {
		h.Absender = string(ich.Kennung())
	}))
	if fehler != nil {
		t.Fatalf("Verpacke: %v", fehler)
	}

	zurueck, fehler := peer.Auspacke(roh)
	if fehler != nil {
		t.Fatalf("Auspacke: %v", fehler)
	}

	for name, unterschrift := range map[string]string{
		"leer":     "",
		"kein Hex": "keinhex",
		"zu kurz":  "abcd",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			kopie := zurueck
			kopie.Unterschrift = unterschrift

			if fehler := kopie.Pruefe(); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("es kam durch: %v", fehler)
			}
		})
	}
}

// **Und eine Hülle, die erst mit ihrem Kopf zu groß wird, endet auch.**
//
// # Der Fund
//
// **Zwei Grenzen sind zwei Prüfungen.** Der frühe Blick auf den Rumpf spart
// das Unterschreiben von Unsinn; er sagt aber nichts über die **Nachricht**,
// und die ist es, die über die Leitung geht. Ein Rumpf von genau
// [peer.Hoechstlaenge] kommt an ihm vorbei und ist mit Kopf darüber.
func TestEineHuelleDieErstMitKopfZuGrossWirdEndetAuch(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	// **Genau [peer.Hoechstlaenge] Byte**: zwei Anführungszeichen und der
	// Rest Füllung.
	amRand := json.RawMessage(`"` + strings.Repeat("y", peer.Hoechstlaenge-2) + `"`)

	if _, fehler := ich.Verpacke(eineHuelle(func(h *peer.Huelle) {
		h.Absender = string(ich.Kennung())
		h.Rumpf = amRand
	})); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die Hülle am Rand kam durch: %v", fehler)
	}
}

// **Ein zu langes Kopffeld ist keins.**
//
// # Warum die Grenze überhaupt dasteht
//
// **Weil der Kopf sonst die ganze Nachricht wäre.** Ohne sie trüge eine
// Sitzungskennung von einem Megabyte durch dieselbe Prüfung wie eine von
// zwanzig Zeichen — und der Empfänger hielte einen Kopf im Speicher, den
// niemand je lesen wollte.
func TestEinZuLangesKopffeldIstKeins(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)
	zuLang := strings.Repeat("z", peer.Hoechstwort+1)

	for name, aendere := range map[string]func(*peer.Huelle){
		"Art":     func(h *peer.Huelle) { h.Art = zuLang },
		"ID":      func(h *peer.Huelle) { h.ID = zuLang },
		"Sitzung": func(h *peer.Huelle) { h.Sitzung = zuLang },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := ich.Verpacke(eineHuelle(func(h *peer.Huelle) {
				h.Absender = string(ich.Kennung())
				aendere(h)
			})); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("es kam durch: %v", fehler)
			}
		})
	}
}

// **Ein Rumpf, der kein JSON ist, geht nicht hinaus.**
//
// **Beim Absender und nicht beim Empfänger**: die kanonische Form entsteht
// beim Verpacken, und was sich nicht verdichten lässt, hat keine.
func TestEinRumpfDerKeinJSONIstGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	for name, rumpf := range map[string]json.RawMessage{
		"krumm": json.RawMessage(`{,}`),
		"null":  json.RawMessage(`null`),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := ich.Verpacke(eineHuelle(func(h *peer.Huelle) {
				h.Absender = string(ich.Kennung())
				h.Rumpf = rumpf
			})); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("es kam durch: %v", fehler)
			}
		})
	}
}

// peerAuspackeFehler verlangt einen Befund und gibt ihn zurück.
func peerAuspackeFehler(t *testing.T, roh []byte) error {
	t.Helper()

	_, fehler := peer.Auspacke(roh)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam durch: %v", fehler)
	}

	return fehler
}

// **Und eine Hülle ohne Kopf prüft gar nicht erst ihre Unterschrift.**
//
// **Die billigste Prüfung zuerst**: eine Ed25519-Prüfung über einen Rumpf,
// der ohnehin nicht angenommen würde, ist Arbeit für einen Absender, der
// keine verdient hat.
func TestEineHuelleOhneKopfPrueftGarNicht(t *testing.T) {
	t.Parallel()

	if fehler := (peer.Huelle{}).Pruefe(); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("die leere Hülle kam durch: %v", fehler)
	}
}
