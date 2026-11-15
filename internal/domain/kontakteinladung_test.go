package domain_test

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

var einladungsjetzt = time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

func eineKontakteinladung() domain.Kontakteinladung {
	return domain.Kontakteinladung{
		Kennung: "12D3KooWAnnaAnnaAnnaAnnaAnnaAnnaAnnaAnnaAnnaAnnaAnna",
		Name:    "Anna",
		Zweck:   domain.ZweckHandel,
		Adresse: "192.168.1.42:41234",
		Ablauf:  einladungsjetzt.Add(domain.KontakteinladungGueltig),
		Nonce:   base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{7}, 16)),
	}
}

// probeUnterschrift ersetzt die Unterschrift — `domain` kennt kein `peer`.
func probeUnterschrift(daten []byte) []byte {
	return append([]byte("unterschrift:"), daten[:8]...)
}

func TestKontakteinladungHinUndZurueck(t *testing.T) {
	t.Parallel()

	einladung := eineKontakteinladung()
	einladung.Vorher = "12D3KooWVorherVorherVorherVorherVorherVorherVorher"

	code, fehler := einladung.Kodiere(probeUnterschrift)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if !strings.HasPrefix(code, "mse4mac-kontakt:1:") {
		t.Fatalf("Marke fehlt: %q", code)
	}

	// **Ein Messenger bricht Zeilen um** — der Code muss das überstehen.
	umbrochen := code[:40] + "\n  " + code[40:90] + " " + code[90:]

	gelesen, unterschrieben, fehler := domain.ZerlegeKontakteinladung(umbrochen)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if gelesen != einladung {
		t.Fatalf("gelesen %+v, geschrieben %+v", gelesen, einladung)
	}

	if !bytes.Equal(unterschrieben.Unterschrift, probeUnterschrift(unterschrieben.Daten)) {
		t.Fatal("die Unterschrift gilt nicht den übergebenen Daten")
	}

	if !bytes.HasPrefix(unterschrieben.Daten, []byte("mse4mac-kontakt/1\n")) {
		t.Fatalf("unterschrieben ohne Vorsatz: %q", unterschrieben.Daten)
	}
}

func TestKontakteinladungZerlegenLehntAb(t *testing.T) {
	t.Parallel()

	gut, fehler := eineKontakteinladung().Kodiere(probeUnterschrift)
	if fehler != nil {
		t.Fatal(fehler)
	}

	rumpf := strings.TrimPrefix(gut, "mse4mac-kontakt:1:")
	daten, unterschrift, _ := strings.Cut(rumpf, ".")
	b64 := base64.RawURLEncoding.EncodeToString

	faelle := []struct {
		name string
		code string
		ziel error
	}{
		{"fremde Marke", "mse4mac:1:" + rumpf, domain.ErrValidation},
		{"ohne Fassung", "mse4mac-kontakt::" + rumpf, domain.ErrValidation},
		{"neuere Fassung", "mse4mac-kontakt:2:" + rumpf, domain.ErrEinladungZuNeu},
		{"ohne Unterschrift", "mse4mac-kontakt:1:" + daten, domain.ErrValidation},
		// **Nach dem Punkt abgeschnitten** — FuzzZerlegeKontakteinladung, QA-06.
		{"leere Unterschrift", "mse4mac-kontakt:1:" + daten + ".", domain.ErrValidation},
		{"leere Daten", "mse4mac-kontakt:1:." + unterschrift, domain.ErrValidation},
		{"Daten kein Base64", "mse4mac-kontakt:1:%%." + unterschrift, domain.ErrValidation},
		{"Unterschrift kein Base64", "mse4mac-kontakt:1:" + daten + ".%%", domain.ErrValidation},
		{"Daten kein JSON", "mse4mac-kontakt:1:" + b64([]byte("{")) + "." + unterschrift, domain.ErrValidation},
		{"leer", "", domain.ErrValidation},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, _, fehler := domain.ZerlegeKontakteinladung(fall.code)
			if !errors.Is(fehler, fall.ziel) {
				t.Fatalf("Fehler %v, erwartet %v", fehler, fall.ziel)
			}
		})
	}
}

func TestKontakteinladungPruefe(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name    string
		aendern func(*domain.Kontakteinladung)
		ziel    error
	}{
		{"gültig", func(*domain.Kontakteinladung) {}, nil},
		{"ohne Adresse gültig", func(e *domain.Kontakteinladung) { e.Adresse = "" }, nil},
		{"ohne Kennung", func(e *domain.Kontakteinladung) { e.Kennung = "" }, domain.ErrValidation},
		{"ohne Namen", func(e *domain.Kontakteinladung) { e.Name = "  " }, domain.ErrValidation},
		{"Name zu lang", func(e *domain.Kontakteinladung) { e.Name = strings.Repeat("ä", 65) }, domain.ErrValidation},
		{"Steuerzeichen im Namen", func(e *domain.Kontakteinladung) { e.Name = "An\u202ena" }, domain.ErrValidation},
		{"fremder Zweck", func(e *domain.Kontakteinladung) { e.Zweck = "alles" }, domain.ErrValidation},
		{"kaputte Adresse", func(e *domain.Kontakteinladung) { e.Adresse = "nirgends" }, domain.ErrValidation},
		{"Nonce zu kurz", func(e *domain.Kontakteinladung) { e.Nonce = "AAAA" }, domain.ErrValidation},
		{"Nonce kein Base64", func(e *domain.Kontakteinladung) { e.Nonce = strings.Repeat("%", 22) }, domain.ErrValidation},
		{"abgelaufen", func(e *domain.Kontakteinladung) { e.Ablauf = einladungsjetzt }, domain.ErrValidation},
		{"zu lange gültig", func(e *domain.Kontakteinladung) {
			e.Ablauf = einladungsjetzt.Add(domain.KontakteinladungHoechstens + time.Second)
		}, domain.ErrValidation},
		{"Vorher ist sie selbst", func(e *domain.Kontakteinladung) { e.Vorher = e.Kennung }, domain.ErrValidation},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			einladung := eineKontakteinladung()
			fall.aendern(&einladung)

			fehler := einladung.Pruefe(einladungsjetzt)
			if fall.ziel == nil && fehler != nil || fall.ziel != nil && !errors.Is(fehler, fall.ziel) {
				t.Fatalf("Pruefe = %v, erwartet %v", fehler, fall.ziel)
			}
		})
	}
}

func TestKontakteinladungKodiereOhneJSON(t *testing.T) {
	t.Parallel()

	// **Ein ungültiges Byte ginge durch `json.Marshal`** — es ersetzte es
	// still, und unterschrieben wäre ein anderer Name als der gewünschte.
	einladung := eineKontakteinladung()
	einladung.Name = "\xff"

	if _, fehler := einladung.Kodiere(probeUnterschrift); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Kodiere = %v", fehler)
	}
}

func TestKontakteinladungKodiereOhneDarstellbarenAblauf(t *testing.T) {
	t.Parallel()

	// **Ein Jahr jenseits von 9999 hat keine RFC-3339-Form** — `json.Marshal`
	// scheitert daran, und der Fehler geht hinaus statt eines halben Codes.
	einladung := eineKontakteinladung()
	einladung.Ablauf = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)

	if _, fehler := einladung.Kodiere(probeUnterschrift); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Kodiere = %v", fehler)
	}
}

// **Ein lesbarer Name** ist einer, wie ihn eine Einladung tragen darf — und
// dieselbe Regel gilt für Namen, die ein Fremder ohne Prüfung schickt (QA-06).
func TestEinLesbarerName(t *testing.T) {
	t.Parallel()

	for name, lesbar := range map[string]bool{
		"Anna": true, "": false, "   ": false, "A\u202Enna": false,
		"A\u0007": false, strings.Repeat("a", 65): false,
	} {
		if domain.LesbarerName(name) != lesbar {
			t.Errorf("%q: erwartet %v", name, lesbar)
		}
	}
}
