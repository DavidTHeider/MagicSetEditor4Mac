package domain_test

// **Der wirksame Akteur und die fünf Wirkungsklassen** — Werkstück O1 von
// Phase 5.2 (#1425).
//
// # Was hier geprüft wird
//
// **Rechte sind eine Schnittmenge, keine Vereinigung.** Fünf Achsen — Rolle,
// Nur-Lesen, Token, Dienstmodus, Nachweis —, und die engste gewinnt. Ein Token
// erweitert nie, was die Rolle nicht hergibt; ein Verwalter mit einem
// Lesetoken liest.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func akteur(rolle domain.Rolle) domain.Akteur {
	return domain.Akteur{
		NutzerID: 1, ProfilID: 7, EignerID: 2, Rolle: rolle,
		Modus: domain.ModusFenster,
	}
}

// **Die Rolle ist die erste Achse.**
func TestDieRolleBegrenztDieWirkung(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		rolle   domain.Rolle
		wirkung domain.Wirkung
		darf    bool
	}{
		"Leser liest":                        {domain.RolleLeser, domain.WirkungLesen, true},
		"Leser schreibt nicht":               {domain.RolleLeser, domain.WirkungSchreiben, false},
		"Leser schickt nichts hinaus":        {domain.RolleLeser, domain.WirkungExternSenden, false},
		"Mitarbeiter schreibt":               {domain.RolleMitarbeiter, domain.WirkungSchreiben, true},
		"Mitarbeiter schickt hinaus":         {domain.RolleMitarbeiter, domain.WirkungExternSenden, true},
		"Mitarbeiter schreibt draußen nicht": {domain.RolleMitarbeiter, domain.WirkungExternSchreiben, false},
		"Verwalter schreibt draußen":         {domain.RolleVerwalter, domain.WirkungExternSchreiben, true},
		"Verwalter schließt ab":              {domain.RolleVerwalter, domain.WirkungVerbindlich, true},
		"ohne Rolle nichts":                  {"", domain.WirkungLesen, false},
		"erfundene Wirkung nie":              {domain.RolleVerwalter, "zaubern", false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := akteur(fall.rolle).Darf(fall.wirkung)
			if (fehler == nil) != fall.darf {
				t.Fatalf("Darf(%q) bei %q: %v, erwartet darf=%v",
					fall.wirkung, fall.rolle, fehler, fall.darf)
			}

			if fehler != nil && !errors.Is(fehler, domain.ErrErlaubnis) {
				t.Errorf("die Absage ist kein ErrErlaubnis: %v", fehler)
			}
		})
	}
}

// **Ein Token erweitert nie** — es engt ein, was die Rolle hergibt.
func TestEinTokenEngtEinUndErweitertNie(t *testing.T) {
	t.Parallel()

	verwalter := akteur(domain.RolleVerwalter)
	verwalter.Modus = domain.ModusDienst

	lesetoken := verwalter
	lesetoken.Token = domain.ScopeLesen

	schreibtoken := verwalter
	schreibtoken.Token = domain.ScopeSchreiben

	faelle := map[string]struct {
		wer     domain.Akteur
		wirkung domain.Wirkung
		darf    bool
	}{
		"Lesetoken liest":                       {lesetoken, domain.WirkungLesen, true},
		"Lesetoken schreibt nicht":              {lesetoken, domain.WirkungSchreiben, false},
		"Schreibtoken schreibt":                 {schreibtoken, domain.WirkungSchreiben, true},
		"Schreibtoken schickt nichts hinaus":    {schreibtoken, domain.WirkungExternSenden, false},
		"Schreibtoken schreibt draußen nicht":   {schreibtoken, domain.WirkungExternSchreiben, false},
		"Schreibtoken schließt nicht ab":        {schreibtoken, domain.WirkungVerbindlich, false},
		"Leser mit Schreibtoken schreibt nicht": {func() domain.Akteur { a := akteur(domain.RolleLeser); a.Token = domain.ScopeSchreiben; return a }(), domain.WirkungSchreiben, false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := fall.wer.Darf(fall.wirkung)
			if (fehler == nil) != fall.darf {
				t.Fatalf("Darf(%q): %v, erwartet darf=%v", fall.wirkung, fehler, fall.darf)
			}
		})
	}
}

// **Nur-Lesen schlägt jede Rolle** — der Treiber lässt ohnehin nichts durch,
// und die Antwort soll dasselbe sagen, bevor er es tut.
func TestNurLesenSchlaegtJedeRolle(t *testing.T) {
	t.Parallel()

	wer := akteur(domain.RolleVerwalter)
	wer.NurLesend = true

	if fehler := wer.Darf(domain.WirkungLesen); fehler != nil {
		t.Fatalf("Lesen abgewiesen: %v", fehler)
	}

	fehler := wer.Darf(domain.WirkungSchreiben)
	if !errors.Is(fehler, domain.ErrNurLesend) {
		t.Fatalf("Schreiben nur lesend: %v, erwartet ErrNurLesend", fehler)
	}
}

// **Was hinausgeht oder verbindlich ist, braucht einen Nachweis** — auch für
// den Verwalter am Fenster. Ein selbstgesetztes „bestätigt" ist keiner.
func TestWasHinausgehtBrauchtEinenNachweis(t *testing.T) {
	t.Parallel()

	faelle := map[domain.Wirkung]bool{
		domain.WirkungLesen:           false,
		domain.WirkungSchreiben:       false,
		domain.WirkungExternSenden:    false,
		domain.WirkungExternSchreiben: true,
		domain.WirkungVerbindlich:     true,
	}

	for wirkung, erwartet := range faelle {
		if wirkung.BrauchtNachweis() != erwartet {
			t.Errorf("%q: BrauchtNachweis=%v, erwartet %v", wirkung, wirkung.BrauchtNachweis(), erwartet)
		}
	}
}

// **Die fünf Klassen haben eine Reihenfolge**, und die Absage nennt die
// Achse, die begrenzt hat — sonst sucht jemand den Fehler bei sich.
func TestDieAbsageNenntDieAchse(t *testing.T) {
	t.Parallel()

	leser := akteur(domain.RolleLeser)
	fehler := leser.Darf(domain.WirkungSchreiben)
	if fehler == nil || !contains(fehler.Error(), "Rolle") {
		t.Errorf("die Absage nennt die Rolle nicht: %v", fehler)
	}

	token := akteur(domain.RolleVerwalter)
	token.Token = domain.ScopeLesen
	fehler = token.Darf(domain.WirkungSchreiben)
	if fehler == nil || !contains(fehler.Error(), "Token") {
		t.Errorf("die Absage nennt das Token nicht: %v", fehler)
	}

	if !domain.WirkungVerbindlich.Reicht(domain.WirkungLesen) ||
		domain.WirkungLesen.Reicht(domain.WirkungSchreiben) {
		t.Error("die Reihenfolge der Wirkungen stimmt nicht")
	}

	if len(domain.Wirkungen()) != 6 || len(domain.Dienstmodi()) != 4 {
		t.Errorf("sechs Wirkungen und vier Modi erwartet, gefunden %d und %d",
			len(domain.Wirkungen()), len(domain.Dienstmodi()))
	}
}

// **Die sechste Wirkung steht vor dem Senden und die Rollengrenze bleibt** —
// Werkstück N7 von Phase 5.4
// (#1568).
//
// # Was hier gemessen wird
//
// **Dass der Einschub kein Verhalten geändert hat.** Das Issue verlangt es
// wörtlich, und eine Ordnung, die über eine Zahlentabelle liefe, hätte
// genau hier eine stille Verschiebung erzeugt: `externSchreiben` stand auf
// 4 und steht jetzt auf 5.
func TestDieSechsteWirkungAendertKeineRollengrenze(t *testing.T) {
	t.Parallel()

	// **Ein Mitarbeiter darf weiterhin erreichbar machen.** Vorher galt es
	// über `lokalSchreiben`, jetzt über die eigene Klasse — und beides Mal
	// reicht seine Rolle.
	mitarbeiter := domain.Akteur{
		NutzerID: 1, ProfilID: 1, Rolle: domain.RolleMitarbeiter,
		Modus: domain.ModusFenster,
	}

	if fehler := mitarbeiter.Darf(domain.WirkungExternEmpfangen); fehler != nil {
		t.Errorf("ein Mitarbeiter darf nicht mehr erreichbar machen: %v", fehler)
	}

	// **Ein Leser darf es nicht** — die untere Grenze steht auch weiterhin.
	leser := domain.Akteur{
		NutzerID: 1, ProfilID: 1, Rolle: domain.RolleLeser,
		Modus: domain.ModusFenster,
	}

	if leser.Darf(domain.WirkungExternEmpfangen) == nil {
		t.Error("ein Leser darf erreichbar machen")
	}

	// **Und sie steht vor dem Senden**: wer Daten hinausschicken darf, darf
	// auch einen Port öffnen — das ist die Aussage der Reihe, und sie ist
	// der Grund für die Stelle.
	if !domain.WirkungExternSenden.Reicht(domain.WirkungExternEmpfangen) {
		t.Error("das Senden reicht nicht bis zum Empfangen")
	}

	if domain.WirkungExternEmpfangen.Reicht(domain.WirkungExternSenden) {
		t.Error("das Empfangen reicht bis zum Senden")
	}

	// **Kein Nachweis** — ein offener Port ist ein Schalter im Fenster und
	// kein Geschäft mit einem Dritten.
	if domain.WirkungExternEmpfangen.BrauchtNachweis() {
		t.Error("das Empfangen verlangt einen Nachweis")
	}

	// **Die Gegenprobe zur berechneten Ordnung**: die beiden Klassen
	// dahinter behalten ihre Zusagen. Eine Zahlentabelle daneben hätte sie
	// beim Einschub verloren, ohne dass es jemand gesehen hätte.
	if !domain.WirkungExternSchreiben.BrauchtNachweis() ||
		!domain.WirkungVerbindlich.BrauchtNachweis() {
		t.Error("eine der beiden äußeren Klassen verlangt keinen Nachweis mehr")
	}
}

func contains(text, teil string) bool {
	return teil != "" && len(text) >= len(teil) && indexOf(text, teil) >= 0
}

func indexOf(text, teil string) int {
	for i := 0; i+len(teil) <= len(text); i++ {
		if text[i:i+len(teil)] == teil {
			return i
		}
	}

	return -1
}
