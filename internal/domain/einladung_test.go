package domain_test

// **Die Einladung trägt ihre Fassung** — Werkstück N3 von Phase 5.4
// (#1456).
//
// # Wofür
//
// **Weil das Drahtformat unter uns sich ändern darf.** tailcat sagt nichts zu:
// „types, functions, and the wire format may all change". Bricht es, kommt
// gar keine Verbindung zustande — der Handschlag aus N1 läuft nie, und ein
// Mensch liest „nicht erreichbar". Ein Verbindungsfehler, der in Wahrheit ein
// Fassungsfehler ist, ist die stille Form, vor der §8 warnt.
//
// **Also trägt die Einladung die Fassung, nicht der Transport** — und sie
// wird geprüft, **bevor** irgendetwas ins Netz geht.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineEinladungTraegtFassungUndAdresse(t *testing.T) {
	t.Parallel()

	einladung := domain.NeueEinladung("tc1abc")

	if !strings.HasPrefix(string(einladung), "mse4mac:") {
		t.Errorf("die Einladung nennt das Programm nicht: %q", einladung)
	}

	adresse, fehler := domain.ZerlegeEinladung(einladung)
	if fehler != nil {
		t.Fatalf("zerlegen: %v", fehler)
	}

	if adresse != "tc1abc" {
		t.Errorf("die Adresse kam als %q zurück", adresse)
	}
}

func TestEineNeuereEinladungWirdBenanntAbgewiesen(t *testing.T) {
	t.Parallel()

	_, fehler := domain.ZerlegeEinladung("mse4mac:99:tc1abc")
	if !errors.Is(fehler, domain.ErrEinladungZuNeu) {
		t.Errorf("erwartet war ErrEinladungZuNeu, kam: %v", fehler)
	}
}

// **Und alles andere ist eine krumme Einladung** — kein Grund, es ins Netz zu
// tragen.
func TestEineKrummeEinladungIstEinBefund(t *testing.T) {
	t.Parallel()

	for name, roh := range map[string]string{
		"leer":               "",
		"ohne Marke":         "tc1abc",
		"fremde Marke":       "andereapp:1:tc1abc",
		"ohne Fassung":       "mse4mac:tc1abc",
		"keine Zahl":         "mse4mac:eins:tc1abc",
		"ohne Adresse":       "mse4mac:1:",
		"Fassung unter eins": "mse4mac:0:tc1abc",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := domain.ZerlegeEinladung(
				domain.Einladung(roh)); !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("erwartet war ErrValidation, kam: %v", fehler)
			}
		})
	}
}

// **Eine Adresse mit Doppelpunkten bleibt heil.** tailcat trägt sie base64-nah,
// aber die Zusage gehört geprüft und nicht angenommen.
func TestEineAdresseMitDoppelpunktenBleibtHeil(t *testing.T) {
	t.Parallel()

	adresse := "tc1:abc:def"

	zurueck, fehler := domain.ZerlegeEinladung(domain.NeueEinladung(adresse))
	if fehler != nil {
		t.Fatalf("zerlegen: %v", fehler)
	}

	if zurueck != adresse {
		t.Errorf("zurück kam %q statt %q", zurueck, adresse)
	}
}

// **Ein Wähler muss die beiden Wege auseinanderhalten können.**
func TestIstEinladungTrenntDieBeidenWege(t *testing.T) {
	t.Parallel()

	for text, erwartet := range map[string]bool{
		"mse4mac:1:tc1abc":     true,
		"mse4mac:99:tc1abc":    true,
		"192.168.178.51:41234": false,
		"[::1]:41234":          false,
		":41234":               false,
		"":                     false,
		"mse4mac":              false,
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()

			if domain.IstEinladung(text) != erwartet {
				t.Errorf("IstEinladung(%q) gab %v", text, !erwartet)
			}
		})
	}
}

// **Und eine Einladung mit falscher Fassung wird trotzdem als eine erkannt** —
// sonst liefe sie in den Adressweg und käme als „krumme Adresse" zurück
// statt als „neuere Fassung".
func TestEineNeuereEinladungIstEineEinladung(t *testing.T) {
	t.Parallel()

	if !domain.IstEinladung("mse4mac:99:tc1abc") {
		t.Error("eine neuere Einladung wurde nicht als Einladung erkannt")
	}
}
