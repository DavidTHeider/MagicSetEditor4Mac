package domain_test

// **Zugriffstoken** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkt 13.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gutesToken gibt eines, das durchgeht.
func gutesToken() domain.Zugriffstoken {
	return domain.Zugriffstoken{
		ProfilID:    1,
		Bezeichnung: "Sicherungsskript",
		Principal:   "Nachtrechner",
		Scope:       domain.ScopeLesen,
	}
}

// **Ein Token braucht ein Profil, einen Namen und einen Scope.**
func TestEinZugriffstokenBrauchtSeineDrei(t *testing.T) {
	t.Parallel()

	if err := gutesToken().Validate(); err != nil {
		t.Fatalf("das gute Token wurde abgewiesen: %v", err)
	}

	faelle := map[string]func(*domain.Zugriffstoken){
		"ohne Profil":       func(z *domain.Zugriffstoken) { z.ProfilID = 0 },
		"ohne Bezeichnung":  func(z *domain.Zugriffstoken) { z.Bezeichnung = "" },
		"mit Rändern":       func(z *domain.Zugriffstoken) { z.Bezeichnung = " Skript " },
		"mit fremdem Scope": func(z *domain.Zugriffstoken) { z.Scope = "alles" },
		"ohne Scope":        func(z *domain.Zugriffstoken) { z.Scope = "" },
	}

	for name, verbiege := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			eines := gutesToken()
			verbiege(&eines)

			if fehler := eines.Validate(); !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("gemeldet wurde: %v", fehler)
			}
		})
	}
}

// **Ein Ablauf ist ein Zeitpunkt oder gar nichts.**
//
// **Leer heißt „läuft nicht ab"** — und das ist eine Antwort und kein
// fehlender Wert: wer ein Token für ein Skript anlegt, das jede Nacht läuft,
// will keines, das nach einer Woche stillsteht.
func TestEinAblaufIstEinZeitpunktOderNichts(t *testing.T) {
	t.Parallel()

	ohne := gutesToken()
	if err := ohne.Validate(); err != nil {
		t.Errorf("ohne Ablauf abgewiesen: %v", err)
	}

	mit := gutesToken()
	mit.Ablauf = "2026-12-24T18:00:00Z"

	if err := mit.Validate(); err != nil {
		t.Errorf("mit Ablauf abgewiesen: %v", err)
	}

	krumm := gutesToken()
	krumm.Ablauf = "Heiligabend"

	if fehler := krumm.Validate(); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("gemeldet wurde: %v", fehler)
	}
}

// **Abgelaufen ist, was vor jetzt liegt.**
//
// **Die Uhr kommt herein und steht nicht in der Domäne** (§16): ein Lauf mit
// echter Wartezeit flackert, und eine Domäne, die `time.Now` ruft, lässt sich
// nicht an einem Stichtag prüfen.
func TestEinAbgelaufenesTokenSagtEs(t *testing.T) {
	t.Parallel()

	stichtag := time.Date(2026, time.September, 5, 12, 0, 0, 0, time.UTC)

	faelle := map[string]struct {
		ablauf string
		hin    bool
	}{
		"ohne Ablauf":     {"", false},
		"morgen":          {"2026-09-06T12:00:00Z", false},
		"gestern":         {"2026-09-04T12:00:00Z", true},
		"genau jetzt":     {"2026-09-05T12:00:00Z", true},
		"unlesbares Wort": {"Heiligabend", true},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			eines := gutesToken()
			eines.Ablauf = fall.ablauf

			if eines.Abgelaufen(stichtag) != fall.hin {
				t.Errorf("%q gilt als abgelaufen: %v", fall.ablauf, !fall.hin)
			}
		})
	}
}

// **Schreiben schließt Lesen ein — und nicht umgekehrt.**
//
// **Zwei Fragen und nicht hundert**: ein Werkzeug liest oder es schreibt.
// Ein feineres Modell wäre eine Aufzählung, die bei jedem neuen Endpunkt
// gepflegt sein will, und ein ungepflegter Eintrag in einer Rechteliste ist
// gefährlicher als keiner.
func TestSchreibenSchliesstLesenEin(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		hat, braucht domain.Scope
		reicht       bool
	}{
		"lesen für lesen":         {domain.ScopeLesen, domain.ScopeLesen, true},
		"lesen für schreiben":     {domain.ScopeLesen, domain.ScopeSchreiben, false},
		"schreiben für lesen":     {domain.ScopeSchreiben, domain.ScopeLesen, true},
		"schreiben für schreiben": {domain.ScopeSchreiben, domain.ScopeSchreiben, true},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fall.hat.Reicht(fall.braucht) != fall.reicht {
				t.Errorf("%q reicht für %q: %v", fall.hat, fall.braucht, !fall.reicht)
			}
		})
	}
}

// **Die Scopeliste ist eine Kopie** — wer sie änderte, änderte die Regel.
func TestDieScopelisteIstEineKopie(t *testing.T) {
	t.Parallel()

	erste := domain.Scopes()
	if len(erste) != 2 {
		t.Fatalf("es sind %d", len(erste))
	}

	erste[0] = "alles"

	if domain.Scopes()[0] != domain.ScopeLesen {
		t.Error("die Liste ließ sich von außen ändern")
	}
}
