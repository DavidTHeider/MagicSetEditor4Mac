package domain_test

// **Das Regelregister** — PLAY-01
// (#1709).
//
// Hier steht, was ein Eintrag sein muss. **Ob der Aufrufer und die Tests
// wirklich da sind**, prüft `internal/werkzeuge/regelregister_test.go` —
// dafür muss man den Baum lesen, und das ist nicht Sache der Domäne.

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einRegeleintrag gibt einen gültigen Eintrag, den ein Fall verbiegt.
func einRegeleintrag() domain.Regeleintrag {
	return domain.Regeleintrag{
		Nummer:       "508.1a",
		Fassung:      domain.Grundprofilquelle,
		Katalog:      domain.KatalogKampf,
		Bereich:      domain.BereichCombat,
		Beschreibung: "Angreifen darf nur eine ungetappte Kreatur des aktiven Spielers",
		Stand:        domain.RegelDurchgesetzt,
		Stelle:       "domain.Partie.AngreiferErklaeren",
		Aufrufer:     "domain.Partie.AngreiferErklaeren",
		Positiv:      "TestEinAngreiferWirdErklaertUndGetappt",
		Negativ:      "TestWerGetapptIstGreiftNichtAn",
	}
}

// **Was ein Eintrag sein muss** — und jede Abweichung sagt, woran es liegt.
func TestEinRegeleintragWirdGeprueft(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		biegen func(*domain.Regeleintrag)
		teil   string
	}{
		"eine Nummer, die keine ist": {
			func(e *domain.Regeleintrag) { e.Nummer = "508,1a" }, "Nummer",
		},
		"eine alte Fassung": {
			func(e *domain.Regeleintrag) { e.Fassung = "2025-11-14" }, "Fassung",
		},
		"ein Katalog, den es nicht gibt": {
			func(e *domain.Regeleintrag) { e.Katalog = "Mondphasen" }, "Katalog",
		},
		"ein Bereich, den es nicht gibt": {
			func(e *domain.Regeleintrag) { e.Bereich = "wetter" }, "Bereich",
		},
		"ohne Beschreibung": {
			func(e *domain.Regeleintrag) { e.Beschreibung = " " }, "Beschreibung",
		},
		"ein Stand, den es nicht gibt": {
			func(e *domain.Regeleintrag) { e.Stand = "vielleicht" }, "Stand",
		},
		"durchgesetzt ohne Gegenfall": {
			func(e *domain.Regeleintrag) { e.Negativ = "" }, "Gegenfall",
		},
		"durchgesetzt ohne Aufrufer": {
			func(e *domain.Regeleintrag) { e.Aufrufer = "" }, "Aufrufer",
		},
		"durchgesetzt ohne Stelle": {
			func(e *domain.Regeleintrag) { e.Stelle = "" }, "Stelle",
		},
		"manuell ohne Grenze": {
			func(e *domain.Regeleintrag) { e.Stand = domain.RegelManuell }, "Grenze",
		},
		"teilweise ohne Grenze": {
			func(e *domain.Regeleintrag) { e.Stand = domain.RegelTeilweise }, "Grenze",
		},
		"offen ohne Grenze": {
			func(e *domain.Regeleintrag) {
				e.Stand, e.Stelle, e.Aufrufer = domain.RegelOffen, "", ""
				e.Positiv, e.Negativ = "", ""
			}, "Grenze",
		},
		"eine Voraussetzung, die keine Nummer ist": {
			func(e *domain.Regeleintrag) { e.Voraussetzungen = []string{"viel"} }, "Voraussetzung",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			eintrag := einRegeleintrag()
			fall.biegen(&eintrag)

			fehler := eintrag.Pruefe()
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
			}

			if !strings.Contains(fehler.Error(), fall.teil) {
				t.Errorf("die Meldung %q nennt %q nicht", fehler, fall.teil)
			}
		})
	}

	// **Und der gültige Eintrag geht durch** — die Gegenprobe, ohne die jeder
	// Fall oben auch dann rot wäre, wenn `Pruefe` alles abwiese.
	if fehler := einRegeleintrag().Pruefe(); fehler != nil {
		t.Errorf("ein gültiger Eintrag wurde abgewiesen: %v", fehler)
	}
}

// **Jeder Eintrag des Registers ist gültig, und jede Nummer steht einmal.**
func TestDasRegisterIstGueltig(t *testing.T) {
	t.Parallel()

	gesehen := map[string]bool{}

	for _, eintrag := range domain.Regelregister() {
		if fehler := eintrag.Pruefe(); fehler != nil {
			t.Errorf("CR %s: %v", eintrag.Nummer, fehler)
		}

		if gesehen[eintrag.Nummer] {
			t.Errorf("CR %s steht zweimal im Register", eintrag.Nummer)
		}

		gesehen[eintrag.Nummer] = true
	}
}

// **Der Katalog aus AK-03 ist ganz da** — jede der sieben Gruppen mit
// mindestens zwei Regeln, die durchgesetzt oder teilweise belegt sind.
func TestJedeGruppeDesKatalogsIstBelegt(t *testing.T) {
	t.Parallel()

	belegt := map[domain.Regelkatalog]int{}

	for _, eintrag := range domain.Regelregister() {
		if eintrag.Stand == domain.RegelDurchgesetzt ||
			eintrag.Stand == domain.RegelTeilweise {
			belegt[eintrag.Katalog]++
		}
	}

	for _, katalog := range domain.Regelkataloge() {
		if belegt[katalog] < 2 {
			t.Errorf("die Gruppe %q hat %d belegte Regeln, erwartet mindestens zwei",
				katalog, belegt[katalog])
		}
	}
}

// **Ein Bereich, den die Matrix `automatisch` nennt, hat mindestens eine
// durchgesetzte Regel im Register** (AK-02) — sonst wäre der Grad eine
// Behauptung ohne Lauf.
func TestEinAutomatischerBereichHatEineDurchgesetzteRegel(t *testing.T) {
	t.Parallel()

	durchgesetzt := map[domain.Regelbereich]bool{}

	for _, eintrag := range domain.Regelregister() {
		if eintrag.Stand == domain.RegelDurchgesetzt {
			durchgesetzt[eintrag.Bereich] = true
		}
	}

	// **Jeder automatische Bereich beider Profile** — und nicht eine Auswahl:
	// eine Liste, die nur die Bereiche aus AK-03 hielte, ließe Mana, Marken
	// und Commander unbelegt, ohne dass es jemand sähe (§10).
	for _, profil := range []domain.Regelprofil{
		domain.Grundprofil("magic"), domain.Commanderprofil("magic"),
	} {
		for _, eintrag := range profil.Matrix.Eintraege {
			if eintrag.Grad == domain.UnterstuetzungAutomatisch &&
				!durchgesetzt[eintrag.Bereich] {
				t.Errorf("%s steht auf automatisch, und keine Regel des "+
					"Registers ist dort durchgesetzt", eintrag.Bereich)
			}
		}
	}
}

// **Was nicht durchgesetzt wird, sehen die Spielenden** — PLAY-01.AK-04:
// „Nicht unterstützte Regeln erhalten sichtbare manuelle Behandlung".
//
// **Sichtbar ist die Matrix, nicht das Register.** Ihre Grenzen zeigen alle
// vier Flächen über `Regelfreigabe`; das Register liest kein Mensch am Tisch.
// Eine Regel, die hier `teilweise`, `manuell` oder `offen` steht, muss
// deshalb in einer Grenze ihres Bereichs genannt sein — sonst wüsste es nur,
// wer den Quelltext liest.
//
// **Eine Unterregel deckt ihre Regel**: `601.2` ist genannt, wenn die Grenze
// `601.2b` nennt, denn die Lücke liegt dann in einem benannten Schritt.
// Umgekehrt nicht — `510.1c` verlangt genau diese Nummer.
func TestJedeNichtDurchgesetzteRegelStehtInDenGrenzen(t *testing.T) {
	t.Parallel()

	grenzen := map[domain.Regelbereich]string{}

	for _, profil := range []domain.Regelprofil{
		domain.Grundprofil("magic"), domain.Commanderprofil("magic"),
	} {
		for _, eintrag := range profil.Matrix.Eintraege {
			grenzen[eintrag.Bereich] += " " + strings.Join(eintrag.Grenzen, " ")
		}
	}

	for _, eintrag := range domain.Regelregister() {
		if eintrag.Stand == domain.RegelDurchgesetzt {
			continue
		}

		if !genannt(eintrag.Nummer).MatchString(grenzen[eintrag.Bereich]) {
			t.Errorf("CR %s ist %s, und keine Grenze von %q nennt sie — "+
				"die Spielenden sähen nicht, dass sie es selbst halten",
				eintrag.Nummer, eintrag.Stand, eintrag.Bereich)
		}
	}
}

// genannt trifft eine Regelnummer — und nach einer Ziffer auch ihre
// Unterregeln.
func genannt(nummer string) *regexp.Regexp {
	danach := `([^0-9]|$)`
	if letztes := nummer[len(nummer)-1]; letztes < '0' || letztes > '9' {
		danach = `([^0-9a-z]|$)`
	}

	return regexp.MustCompile(`(^|[^0-9.])` + regexp.QuoteMeta(nummer) + danach)
}
