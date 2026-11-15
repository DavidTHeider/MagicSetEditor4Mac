package domain_test

// **Was als Adresse eines Gegenübers durchgeht** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Warum die Form hier geprüft wird und die Erreichbarkeit nicht
//
// **Weil das zwei Fragen sind.** „Sieht das aus wie eine Adresse" ist eine
// Regel und gehört in die Domäne; „antwortet dort jemand" ist eine Messung und
// gehört an die Leitung. Wer beides hier beantworten wollte, hätte eine
// Domäne, die ins Netz greift.
//
// # Und warum leer erlaubt ist
//
// **Weil die meisten Kontakte keine haben.** Der Weg „Adresse von Hand" ist
// der Zusatz für Netze, über die mDNS nicht läuft — ein Kontakt ohne Adresse
// ist der Normalfall und kein Mangel.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineAdresseHatEinenWirtUndEinenPort(t *testing.T) {
	t.Parallel()

	gut := []string{
		"",
		"192.168.1.7:41234",
		"beispiel.invalid:443",
		"[2001:db8::1]:41234",
		"[::1]:1",
		"mein-mac.local:65535",
	}

	for _, adresse := range gut {
		t.Run("gilt: "+adresse, func(t *testing.T) {
			t.Parallel()

			if fehler := domain.PruefeAdresse(adresse); fehler != nil {
				t.Fatalf("%q wurde abgewiesen: %v", adresse, fehler)
			}
		})
	}
}

func TestWasKeineAdresseIstWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	schlecht := map[string]string{
		"ohne Port":            "192.168.1.7",
		"ohne Wirt":            ":41234",
		"Port ist kein Port":   "beispiel.invalid:http",
		"Port ist null":        "beispiel.invalid:0",
		"Port ist zu groß":     "beispiel.invalid:65536",
		"Port hat ein Zeichen": "beispiel.invalid:44 3",
		"nur ein Doppelpunkt":  ":",
		"leerer Port":          "beispiel.invalid:",
		"IPv6 ohne Klammern":   "2001:db8::1:41234",
		"IPv6 halb geklammert": "[2001:db8::1:41234",
		"Klammer im Wirt":      "[[]:41234", // FuzzZerlegeAdresse, QA-06
		"Leerzeichen im Wirt":  "mein mac:443",
		"zu lang":              strings.Repeat("a", 300) + ":443",
		"ein ganzer Satz":      "frag mal David nach der Adresse",
	}

	for name, adresse := range schlecht {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := domain.PruefeAdresse(adresse); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("%q kam durch: %v", adresse, fehler)
			}
		})
	}
}

// **Ein Kontakt mit einer krummen Adresse ist keiner.**
//
// **Die Prüfung steht nicht nur daneben, sie hängt am Kontakt** — sonst
// bliebe sie ein Weg, den ein Aufrufer nehmen kann oder auch nicht.
func TestEinKontaktMitKrummerAdresseIstKeiner(t *testing.T) {
	t.Parallel()

	kontakt := domain.Peerkontakt{
		Kennung: "beispielkennung",
		Stand:   domain.KontaktUnbestaetigt,
		Adresse: "frag mal David",
	}

	if fehler := kontakt.Validate(); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der Kontakt kam durch: %v", fehler)
	}
}

// **Eine Horchadresse darf ohne Wirt dastehen — eine Wähladresse nicht.**
//
// # Der Fund, für den es diesen Lauf gibt
//
// **Es sind zwei Fragen, und eine Prüfung beantwortete beide.** `:41234` ist
// die übliche Form, auf **jeder** Schnittstelle dieses Rechners zu horchen —
// und zugleich keine Adresse, zu der sich wählen ließe. [PruefeAdresse] hat
// sie abgewiesen, während Katalog, Hilfetext und `FAEHIGKEITEN.md` sie
// empfahlen: eine Zusage, die ihr eigener Code nicht hielt.
//
// **Gefunden am 2026-09-14** von `internal/app/peerhorchen_intern_test.go` —
// dem ersten Lauf, der wirklich eine Horchadresse angab.
func TestEineHorchadresseDarfOhneWirtDastehen(t *testing.T) {
	t.Parallel()

	// **Was nur beim Horchen gilt.**
	//
	// **Port 0 gehört dazu**, und er ist der zweite Fund derselben Sorte
	// (2026-09-14): beim Horchen heißt er „such mir einen freien aus", beim
	// Wählen wäre er keiner. Gefunden hat ihn der erste Lauf, der wirklich
	// einen echten Port geöffnet hat.
	for _, adresse := range []string{
		":41234", ":1", ":65535", "127.0.0.1:0", ":0", "beispiel.invalid:0",
	} {
		t.Run("horcht auf "+adresse, func(t *testing.T) {
			t.Parallel()

			if fehler := domain.PruefeHorchadresse(adresse); fehler != nil {
				t.Errorf("%q wurde abgewiesen: %v", adresse, fehler)
			}

			if fehler := domain.PruefeAdresse(adresse); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Errorf("%q taugte zum Wählen: %v", adresse, fehler)
			}
		})
	}

	// **Und was für beide gilt.**
	for _, adresse := range []string{
		"192.168.1.7:41234", "[2001:db8::1]:41234", "mein-mac.local:443",
	} {
		t.Run("beides: "+adresse, func(t *testing.T) {
			t.Parallel()

			for name, fehler := range map[string]error{
				"horchen": domain.PruefeHorchadresse(adresse),
				"wählen":  domain.PruefeAdresse(adresse),
			} {
				if fehler != nil {
					t.Errorf("%s: %q wurde abgewiesen: %v",
						name, adresse, fehler)
				}
			}
		})
	}
}

// **Eine leere Horchadresse ist keine.**
//
// **Anders als beim Kontakt**: dort heißt leer „ich kenne keine", und das ist
// eine Aussage. Hier hieße es „horche nirgends", und dafür gibt es das
// Beenden.
func TestEineLeereHorchadresseIstKeine(t *testing.T) {
	t.Parallel()

	if fehler := domain.PruefeHorchadresse(""); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}

	// **Die Gegenprobe daneben**: beim Kontakt ist dieselbe Eingabe gültig.
	if fehler := domain.PruefeAdresse(""); fehler != nil {
		t.Fatalf("die leere Kontaktadresse wurde abgewiesen: %v", fehler)
	}
}

// **Und eine krumme bleibt krumm — an beiden.**
func TestEineKrummeAdresseTaugtFuerKeinesVonBeiden(t *testing.T) {
	t.Parallel()

	// **`beispiel.invalid:0` steht hier nicht** — es ist beim Horchen
	// gültig (ein Wirt und ein freier Port) und nur beim Wählen keine
	// Adresse. Der erste Anlauf hatte es hier, und der Lauf war zu Recht rot.
	for _, adresse := range []string{
		"frag mal David", "beispiel.invalid:http",
		"2001:db8::1:41234", ":", "mein mac:443",
	} {
		t.Run(adresse, func(t *testing.T) {
			t.Parallel()

			for name, fehler := range map[string]error{
				"horchen": domain.PruefeHorchadresse(adresse),
				"wählen":  domain.PruefeAdresse(adresse),
			} {
				if !errors.Is(fehler, domain.ErrValidation) {
					t.Errorf("%s: %q kam durch: %v", name, adresse, fehler)
				}
			}
		})
	}
}

// **Eine geprüfte Adresse lässt sich zerlegen.**
//
// **In der Domäne und nicht über `net.SplitHostPort`**: `internal/app` darf
// das Netz nicht erreichen (§2), und die Form steht hier ohnehin schon.
func TestEineGeprüfteAdresseLaesstSichZerlegen(t *testing.T) {
	t.Parallel()

	for adresse, erwartet := range map[string]struct {
		wirt string
		port int
	}{
		"192.168.1.7:41234":   {"192.168.1.7", 41234},
		"[2001:db8::1]:41234": {"2001:db8::1", 41234},
		":41234":              {"", 41234},
		"127.0.0.1:0":         {"127.0.0.1", 0},
	} {
		t.Run(adresse, func(t *testing.T) {
			t.Parallel()

			wirt, port, fehler := domain.ZerlegeAdresse(adresse)
			if fehler != nil {
				t.Fatalf("ZerlegeAdresse: %v", fehler)
			}

			if wirt != erwartet.wirt || port != erwartet.port {
				t.Fatalf("es kamen %q und %d", wirt, port)
			}
		})
	}
}

// **Und eine krumme lässt sich nicht zerlegen.**
func TestEineKrummeAdresseLaesstSichNichtZerlegen(t *testing.T) {
	t.Parallel()

	for _, adresse := range []string{"", "krumm", "wirt:keinPort"} {
		t.Run(adresse, func(t *testing.T) {
			t.Parallel()

			if _, _, fehler := domain.ZerlegeAdresse(adresse); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("es kam %v", fehler)
			}
		})
	}
}

// **Eine Einladung ist auch eine Adresse** — N3
// (#1456).
//
// **Der Lauf gehört hierher und nicht nur zum Wähler**: gefunden hat die
// Lücke ein Lauf in `internal/app`, weil der Schreibweg eine Einladung abwies,
// während der Wähler längst auf den Fernweg abbog. Ein Paket, das trägt, und
// ein Weg, der nicht durchgeht (§10).
func TestEineEinladungGiltAlsAdresse(t *testing.T) {
	t.Parallel()

	if fehler := domain.PruefeAdresse("mse4mac:1:tc1abc"); fehler != nil {
		t.Errorf("eine gültige Einladung wurde abgewiesen: %v", fehler)
	}
}

// **Und eine krumme Einladung wird als Einladung abgewiesen, nicht als
// Adresse.**
//
// **Der Unterschied ist der Satz, den ein Mensch liest**: „das ist keine
// Adresse — sie braucht einen Wirt und einen Port" schickte ihn auf die Suche
// nach einem Tippfehler, wo die Gegenseite schlicht neuer ist.
func TestEineZuNeueEinladungWirdAlsEinladungAbgewiesen(t *testing.T) {
	t.Parallel()

	fehler := domain.PruefeAdresse("mse4mac:99:tc1abc")
	if !errors.Is(fehler, domain.ErrEinladungZuNeu) {
		t.Errorf("erwartet war ErrEinladungZuNeu, kam: %v", fehler)
	}
}
