package app_test

// **Was die Antwort über die Anbieter sagt** — Werkstück A8 von Phase 5.3
// (#1447)
// und A9 (#1448),
// abgetrennt, weil `kianthropic_test.go` über die Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Dort steht, **wer** antwortet und ob er darf; hier, **wie** die Antwort
// aussieht — die Adresse, der Anbieter am Vorschlag, und dass keine Liste
// darin `null` ist.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Anthropic nennt seine Adresse, statt sie leer zu lassen.**
//
// # Warum das ein Befund der Sichtprüfung ist
//
// **Leer heißt im Fenster „die Vorgabeadresse auf diesem Rechner"** — und
// damit stand am 2026-09-12 unter »KI-Modelle«: „Anbieter anthropic: Gefragt
// wird die Vorgabeadresse auf diesem Rechner." Das ist das Gegenteil dessen,
// was zutrifft.
//
// Gefunden hat es die **Sichtprüfung am laufenden Fenster** und kein Test:
// die Go-Seite trug ein leeres Feld, und leer war für Ollama richtig.
func TestAnthropicNenntSeineAdresse(t *testing.T) {
	t.Parallel()

	antwort := zweianbieter(t, map[string]string{
		derSchalter(): domain.DienstAn,
	}, &laufzeitdoppel{}, &laufzeitdoppel{
		fassung: "anthropic-version 2023-06-01",
	}).KIModelle(probekontext(t))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	for _, anbieter := range antwort.Anbieter {
		if anbieter.Name != "anthropic" {
			continue
		}

		if anbieter.Adresse != "https://api.anthropic.com" {
			t.Fatalf("die Adresse ist %q", anbieter.Adresse)
		}

		return
	}

	t.Fatal("anthropic steht nicht in der Liste")
}

// **Keine Liste dieser Antwort ist jemals `null`.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil ihn kein anderer findet.** Die Frontendtests gehen über
// `test/api.ts` und laufen nie durch die JSON-Serialisierung; die Go-Tests
// sehen die Struct-Marken nicht. Ein `nil`-Slice wird zu `null`, und
// `null.length` beendet die Seite — bei grüner Pipeline (§3, „Zwischen den
// Prüfebenen").
//
// **Gefunden am 2026-09-12 in der Sichtprüfung am laufenden Fenster**: ein
// abgeschalteter Anbieter trug `messbar: nil`, und die Einstellungsseite
// stürzte beim Aufschlagen des Bereichs ab. Kein Test war rot.
func TestKeineListeDerKIAntwortIstNull(t *testing.T) {
	t.Parallel()

	// **Alle Anbieter aus**, denn genau dann entstehen die leeren Listen.
	antwort := zweianbieter(t, map[string]string{},
		&laufzeitdoppel{}, &laufzeitdoppel{}).KIModelle(probekontext(t))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	roh, fehler := json.Marshal(antwort)
	if fehler != nil {
		t.Fatalf("verpacken: %v", fehler)
	}

	for _, feld := range []string{
		"anbieter", "befunde", "modelle", "messbar", "gemessen",
	} {
		if strings.Contains(string(roh), `"`+feld+`":null`) {
			t.Errorf("%q steht als null in der Antwort:\n%s", feld, roh)
		}
	}
}
