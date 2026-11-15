package service_test

// **Ein vorgezeigtes Token aus fremder Hand** — Werkstück O11 von Phase 5.2
// (#1435),
// Befund 3 des Zweitblicks: das Tokenziel fehlte.
//
// # Warum
//
// Das Token ist der **Schlüssel** zu diesem Bestand für jeden Prozess ohne
// Fenster. Wer es vorzeigt, ist per Annahme nicht vertrauenswürdig — und die
// Zusage aus O2 ([0080](../../docs/entscheidungen/0080-maschinenzugaenge-und-widerrufbare-sitzungen.md))
// ist enger als „abgewiesen": **jede** Abweisung außer der am Scope sieht
// gleich aus. „Diese Kennung gibt es nicht" sagte einem Suchenden, welche es
// gibt; „abgelaufen" sagte ihm, dass er beinahe richtig lag.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - zweimal dieselbe Eingabe, zwei Ergebnisse
//   - **ein angenommener Klartext, der nicht der echte ist** — das wäre der
//     Einbruch
//   - **zwei verschiedene Meldungen** auf zwei falsche Klartexte: jede
//     Unterscheidung ist ein Orakel
//   - **ein angenommenes Token ohne die verlangte Reichweite**
//
//	go test ./internal/service -run Fuzz -fuzz FuzzTokenPruefen -fuzztime 2m

import (
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// tokensaat sind die Gestalten, mit denen der Fuzzer beginnt.
var tokensaat = []string{
	"",
	"mse",
	"mse_",
	"mse_1",
	"mse_1_",
	"mse_1_geheim",
	"mse_0_geheim",
	"mse_-1_geheim",
	"mse_99_geheim",
	"mse_1_geheim_mehr",
	"MSE_1_geheim",
	" mse_1_geheim ",
	"mse_9223372036854775808_geheim",
	"mse_1_" + strings.Repeat("a", 500),
	"mse_０_geheim",
}

// **Kein fremder Klartext kommt durch, und keine Abweisung verrät etwas.**
func FuzzTokenPruefen(f *testing.F) {
	for _, saat := range tokensaat {
		f.Add(saat)
	}

	jetzt := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)

	doppel := neuesTokendoppel()
	verwaltung := service.NeueTokenverwaltung(doppel).
		MitUhr(func() time.Time { return jetzt })

	_, echter, fehler := verwaltung.Anlegen(f.Context(), domain.Zugriffstoken{
		ProfilID: 1, Principal: "Rechner", Bezeichnung: "Probe",
		Scope: domain.ScopeSchreiben,
	})
	if fehler != nil {
		f.Fatalf("das Probetoken anlegen: %v", fehler)
	}

	f.Add(echter)

	// **Eine Meldung, an der sich jede weitere messen lässt.**
	_, erste := verwaltung.Pruefe(f.Context(), "mse_1_falsch", domain.ScopeLesen)
	if erste == nil {
		f.Fatal("ein falsches Geheimnis kam durch — der Lauf greift daneben")
	}

	f.Fuzz(func(t *testing.T, klartext string) {
		eines, ergebnis := verwaltung.Pruefe(t.Context(), klartext, domain.ScopeLesen)

		_, zweites := verwaltung.Pruefe(t.Context(), klartext, domain.ScopeLesen)
		if (ergebnis == nil) != (zweites == nil) {
			t.Fatalf("derselbe Klartext einmal so und einmal anders: %v gegen %v",
				ergebnis, zweites)
		}

		if ergebnis != nil {
			if zweites.Error() != ergebnis.Error() {
				t.Errorf("derselbe Klartext gab zwei Meldungen: %q gegen %q",
					ergebnis, zweites)
			}

			// **Jede Abweisung außer der am Scope sieht gleich aus.**
			if !errorsIstErlaubnis(ergebnis) && ergebnis.Error() != erste.Error() {
				t.Errorf("eine Abweisung unterscheidet sich von der anderen: "+
					"%q gegen %q", ergebnis, erste)
			}

			return
		}

		if klartext != echter {
			t.Fatalf("ein fremder Klartext kam durch: %q", klartext)
		}

		if !eines.Scope.Reicht(domain.ScopeLesen) {
			t.Errorf("angenommen ohne Reichweite: %+v", eines)
		}
	})
}

// errorsIstErlaubnis sagt, ob die Abweisung die am Scope ist.
//
// **Sie darf sich unterscheiden**, und zwar mit Grund: wer ein gültiges Token
// hat, ist kein Fremder — ihm zu sagen, dass es nur liest, ist eine Auskunft
// an einen Berechtigten.
func errorsIstErlaubnis(fehler error) bool {
	return strings.Contains(fehler.Error(), "darf")
}
