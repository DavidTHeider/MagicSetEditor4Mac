package domain_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// TestDasZielDerAusgabeIstEinGedaechtnisDesArbeitsplatzes hält AP-10 (#2367):
// welches Ziel der Ausgabeassistent je Set zuletzt zeigte, ist eine Frage an
// das Fenster wie die Breite einer Spalte — es reist nur auf Wunsch.
func TestDasZielDerAusgabeIstEinGedaechtnisDesArbeitsplatzes(t *testing.T) {
	t.Parallel()

	if ebene, da := domain.EbeneVon(domain.EinstellungAusgabeziele); !da || ebene != domain.EbeneArbeitsplatz {
		t.Errorf("Ebene %q (%v)", ebene, da)
	}

	if vorgabe := domain.EinstellungsVorgabe(domain.EinstellungAusgabeziele); vorgabe != "" {
		t.Errorf("Vorgabe %q — ohne Wahl ist noch kein Ziel gemerkt", vorgabe)
	}

	if strings.Join(domain.Ausgabeziele(), " ") != "setdatei bilder druckbogen vorlage mtgjson" {
		t.Errorf("die Ziele: %v", domain.Ausgabeziele())
	}
}

func TestGemerkteZiele(t *testing.T) {
	t.Parallel()

	gelesen, fehler := domain.GemerkteZieleAus(`[{"set":"4","ziel":"druckbogen"},{"set":"17","ziel":"bilder"}]`)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if fmt.Sprint(gelesen) != "[{4 druckbogen} {17 bilder}]" {
		t.Errorf("gelesen: %v", gelesen)
	}

	leer, fehler := domain.GemerkteZieleAus("  ")
	if fehler != nil || leer == nil || len(leer) != 0 {
		t.Errorf("leer: %v, %v — leer heißt keines, und als Liste", leer, fehler)
	}
}

func TestGemerkteZieleWeistAb(t *testing.T) {
	t.Parallel()

	zuViele := make([]string, 0, domain.HoechstensGemerkteZiele+1)
	for stelle := range domain.HoechstensGemerkteZiele + 1 {
		zuViele = append(zuViele, fmt.Sprintf(`{"set":"%d","ziel":"bilder"}`, stelle+1))
	}

	for name, wert := range map[string]string{
		"kein JSON":                 `bilder`,
		"ein Objekt statt Liste":    `{"4":"bilder"}`,
		"ein unbekanntes Ziel":      `[{"set":"4","ziel":"pdf"}]`,
		"ein Set ohne Kennung":      `[{"set":"","ziel":"bilder"}]`,
		"eine Kennung ohne Zahl":    `[{"set":"vier","ziel":"bilder"}]`,
		"eine Kennung unter eins":   `[{"set":"0","ziel":"bilder"}]`,
		"dasselbe Set zweimal":      `[{"set":"4","ziel":"bilder"},{"set":"4","ziel":"vorlage"}]`,
		"mehr als die Höchstzahl":   "[" + strings.Join(zuViele, ",") + "]",
		"ein Ziel mit Großbuchstab": `[{"set":"4","ziel":"Bilder"}]`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := domain.PruefeEinstellung(domain.EinstellungAusgabeziele, wert); fehler == nil {
				t.Errorf("%q nahm sie an", wert)
			}
		})
	}
}
