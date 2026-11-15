// **Ein Fehler beim Zeichnen steht mit seiner Kennung im Protokoll** — F01
// (#2340),
// Befund A-01.
//
// # Warum
//
// Die Fehlerfläche im Fenster nennt eine Kennung. **Sie ist nur etwas wert,
// wenn sie sich wiederfinden lässt**: im Protokoll, das das Diagnosepaket
// einsammelt. Ohne diese Zeile stünde im Fenster eine Zahl, zu der es nichts
// gibt.
//
// # Was die Prüfungen halten
//
// Die Zeile trägt Kennung, Ebene, Bereich und Meldung — **durch das
// schwärzende Protokoll** und nicht an ihm vorbei: ein Merkmal, das niemand
// angemeldet hat, verliert dort seinen Wert, und die Kennung stünde im
// Fenster und nirgends sonst. Eine Kennung in fremder Gestalt und eine
// unbekannte Ebene werden abgewiesen, statt ins Protokoll zu gehen; und eine
// überlange Meldung wird gekürzt — wer das Protokoll füllen kann, verdrängt
// die Zeilen, um derentwillen es da ist.
package app_test

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// imEchtenProtokoll schreibt über das Protokoll, das auch das Fenster
// benutzt — mit Schwärzer — und gibt seinen Inhalt.
func imEchtenProtokoll(t *testing.T, schreiben func()) string {
	t.Helper()

	return protokollLesen(t, func(schreiber *slog.Logger) {
		vorher := slog.Default()
		slog.SetDefault(schreiber)
		defer slog.SetDefault(vorher)

		schreiben()
	})
}

func TestEinFensterfehlerStehtMitKennungImProtokoll(t *testing.T) {
	zeile := imEchtenProtokoll(t, func() {
		antwort := app.NeueApp("").FensterfehlerMelden(t.Context(), app.Fensterfehler{
			Kennung: "7f3a-21c9", Ebene: "bereich", Bereich: "sammlung",
			Meldung: "Cannot read properties of undefined (reading 'name')",
		})
		if antwort.Fehler != nil {
			t.Fatalf("die Meldung wurde abgewiesen: %+v", antwort.Fehler)
		}
	})

	for _, teil := range []string{"fehlerkennung=7f3a-21c9", "art=bereich", "bereich=sammlung", "reading 'name'"} {
		if !strings.Contains(zeile, teil) {
			t.Errorf("im Protokoll fehlt %q:\n%s", teil, zeile)
		}
	}

	// **Die Meldung sagt nur, woher** — welcher Art der Fehler war, sagt
	// `art`; ein Satz über das Zeichnen wäre für einen Ruf falsch (F03).
	if !strings.Contains(zeile, `msg="das Fenster fing einen Fehler"`) {
		t.Errorf("die Zeile trägt nicht die Meldung des Fensters:\n%s", zeile)
	}

	if strings.Contains(zeile, app.Nichterklaert) {
		t.Errorf("ein Merkmal der Zeile ist nicht angemeldet:\n%s", zeile)
	}
}

// **Drei Ebenen mehr als zwei** — F03 (#2342): eine Panik in einer Bindung
// (`ruf`, der Bereich ist die Methode) und ein Fehler ohne Schlüssel, den
// eine Meldung zeigt (`meldung`), schreiben ihre Kennung genauso.
func TestDieEbenenRufUndMeldungWerdenAngenommen(t *testing.T) {
	for _, ebene := range []string{"ruf", "meldung"} {
		t.Run(ebene, func(t *testing.T) {
			zeile := imEchtenProtokoll(t, func() {
				antwort := app.NeueApp("").FensterfehlerMelden(t.Context(), app.Fensterfehler{
					Kennung: "1a2b-3c4d", Ebene: ebene, Bereich: "kartenzeilen",
					Meldung: "Kartenzeilen: panic: runtime error",
				})
				if antwort.Fehler != nil {
					t.Fatalf("abgewiesen: %+v", antwort.Fehler)
				}
			})

			if !strings.Contains(zeile, "art="+ebene) || !strings.Contains(zeile, "fehlerkennung=1a2b-3c4d") {
				t.Errorf("die Zeile trägt Ebene oder Kennung nicht:\n%s", zeile)
			}
		})
	}
}

func TestEinFensterfehlerInFremderGestaltWirdAbgewiesen(t *testing.T) {
	for name, fehler := range map[string]app.Fensterfehler{
		"eine Kennung mit Zeilenumbruch": {Kennung: "7f3a-21c9\nlevel=ERROR", Ebene: "bereich", Bereich: "sammlung"},
		"eine Kennung ohne Bindestrich":  {Kennung: "7f3a21c9", Ebene: "bereich", Bereich: "sammlung"},
		"eine unbekannte Ebene":          {Kennung: "7f3a-21c9", Ebene: "keller", Bereich: "sammlung"},
		"ein Bereich mit Leerzeichen":    {Kennung: "7f3a-21c9", Ebene: "bereich", Bereich: "die sammlung"},
	} {
		t.Run(name, func(t *testing.T) {
			gesammelt := mitProtokoll(t)

			antwort := app.NeueApp("").FensterfehlerMelden(t.Context(), fehler)

			if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeValidierung {
				t.Errorf("angenommen statt abgewiesen: %+v", antwort.Fehler)
			}

			if gesammelt.Len() != 0 {
				t.Errorf("eine abgewiesene Meldung steht im Protokoll:\n%s", gesammelt.String())
			}
		})
	}
}

func TestEineUeberlangeMeldungWirdGekuerzt(t *testing.T) {
	gesammelt := mitProtokoll(t)

	antwort := app.NeueApp("").FensterfehlerMelden(t.Context(), app.Fensterfehler{
		Kennung: "0000-ffff", Ebene: "wurzel", Bereich: "fenster",
		Meldung: strings.Repeat("x", 10*app.FensterfehlerHoechstens),
	})
	if antwort.Fehler != nil {
		t.Fatalf("abgewiesen: %+v", antwort.Fehler)
	}

	if n := strings.Count(gesammelt.String(), "x"); n > app.FensterfehlerHoechstens {
		t.Errorf("%d Zeichen Meldung im Protokoll, erlaubt sind %d", n, app.FensterfehlerHoechstens)
	}
}
