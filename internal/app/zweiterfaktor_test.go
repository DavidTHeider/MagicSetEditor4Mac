// **Der zweite Faktor über die Fassade** — Werkstück 7 von Phase 4.12
// (#1351).
package app_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Faktoren eines Nutzers kommen samt der Auskunft über den Rechner an.**
//
// **In einer Antwort**, weil die Seite beides nebeneinander zeigt: was
// eingerichtet ist und was sich einrichten ließe (§7, Regel 2).
func TestDieFaktorenKommenMitDerRechnerauskunftAn(t *testing.T) {
	anwendung := faktoranwendung(t)

	antwort := anwendung.Faktoren(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	// **Nie `nil`, sondern eine leere Liste** (§3).
	if antwort.Faktoren == nil {
		t.Error("die Liste ist nil")
	}
}

// **Eine unbrauchbare Kennung wird abgewiesen — und die Liste bleibt eine.**
func TestEineUnbrauchbareKennungAmFaktorWirdAbgewiesen(t *testing.T) {
	anwendung := faktoranwendung(t)

	wege := map[string]func() *app.Fehler{
		"lesen": func() *app.Fehler {
			antwort := anwendung.Faktoren(t.Context(), "keine Zahl")
			if antwort.Faktoren == nil {
				t.Error("die Liste ist nil")
			}

			return antwort.Fehler
		},
		"einrichten": func() *app.Fehler {
			antwort := anwendung.FaktorEinrichten(t.Context(), "keine Zahl")
			if antwort.Codes == nil {
				t.Error("die Codes sind nil")
			}

			return antwort.Fehler
		},
		"bestätigen": func() *app.Fehler {
			return anwendung.FaktorBestaetigen(t.Context(), "keine Zahl", "123456").Fehler
		},
		"biometrie einrichten": func() *app.Fehler {
			return anwendung.BiometrieEinrichten(t.Context(), "keine Zahl").Fehler
		},
		"entfernen": func() *app.Fehler {
			return anwendung.FaktorEntfernen(t.Context(), "keine Zahl", "totp").Fehler
		},
	}

	for name, weg := range wege {
		t.Run(name, func(t *testing.T) {
			if fehler := weg(); fehler == nil {
				t.Error("die unbrauchbare Kennung kam durch")
			}
		})
	}
}

// **Der ganze Weg: einrichten, bestätigen, wechseln, entfernen.**
//
// **Über einen echten Bestand**, denn das ist die Zusage: ein Wechsel, den
// eine Prüfung abweist, findet nicht statt.
func TestDerGanzeFaktorwegAmFenster(t *testing.T) {
	anwendung := faktoranwendung(t)

	eingerichtet := anwendung.FaktorEinrichten(t.Context(), "1")
	if eingerichtet.Fehler != nil {
		t.Fatalf("%+v", eingerichtet.Fehler)
	}

	if !strings.HasPrefix(eingerichtet.Adresse, "otpauth://totp/") {
		t.Errorf("die Adresse ist %q", eingerichtet.Adresse)
	}

	// **Das Geheimnis geht genau einmal hinaus** — und dann nie wieder.
	roh, marshalfehler := json.Marshal(anwendung.Faktoren(t.Context(), "1"))
	if marshalfehler != nil {
		t.Fatalf("marshal: %v", marshalfehler)
	}

	if strings.Contains(string(roh), eingerichtet.Geheimnis) {
		t.Errorf("das Geheimnis steht in der Übersicht: %s", roh)
	}

	if bestaetigt := anwendung.FaktorBestaetigen(
		t.Context(), "1", faktorcode(t, eingerichtet.Geheimnis)); bestaetigt.Fehler != nil {
		t.Fatalf("%+v", bestaetigt.Fehler)
	}

	if antwort := anwendung.Faktoren(t.Context(), "1"); len(antwort.Faktoren) != 1 {
		t.Fatalf("eingetragen ist %+v", antwort.Faktoren)
	}

	// **Ein Wechsel ohne Nachweis findet jetzt nicht mehr statt.**
	if gewechselt := anwendung.ProfilWechseln(
		t.Context(), "1", ""); gewechselt.Fehler == nil {
		t.Error("der Wechsel ging ohne zweiten Faktor durch")
	}

	if entfernt := anwendung.FaktorEntfernen(
		t.Context(), "1", "totp"); entfernt.Fehler != nil {
		t.Fatalf("%+v", entfernt.Fehler)
	}

	// **Und danach geht er wieder.**
	if gewechselt := anwendung.ProfilWechseln(
		t.Context(), "1", ""); gewechselt.Fehler != nil {
		t.Errorf("%+v", gewechselt.Fehler)
	}
}
