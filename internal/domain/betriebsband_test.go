package domain_test

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

const gib = uint64(1) << 30

// **Knapp unter 2 GiB oder 5 %, wieder gut erst über 2,5 GiB und 6 %** —
// sonst meldete sich der Hinweis bei jedem Abruf neu, wenn der Platz um die
// Schwelle pendelt.
func TestDerPlatzIstKnappMitHysterese(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		frei, gesamt uint64
		vorher       bool
		erwartet     bool
	}{
		"viel frei":                                {100 * gib, 500 * gib, false, false},
		"unter 2 GiB auf kleiner Platte":           {gib + gib/2, 20 * gib, false, true},
		"unter 5 % auf großer Platte":              {20 * gib, 500 * gib, false, true},
		"knapp bleibt knapp unter 2,5 GiB":         {2*gib + gib/4, 20 * gib, true, true},
		"dieselbe Lage ohne Vorgeschichte ist gut": {2*gib + gib/4, 20 * gib, false, false},
		"knapp bleibt knapp unter 6 %":             {28 * gib, 500 * gib, true, true},
		"5,6 % ohne Vorgeschichte ist gut":         {28 * gib, 500 * gib, false, false},
		"knapp wird gut über beiden Grenzen":       {3 * gib, 40 * gib, true, false},
		"ohne Größe ist nichts knapp":              {0, 0, false, false},
	} {
		if knapp := domain.PlatzKnapp(fall.frei, fall.gesamt, fall.vorher); knapp != fall.erwartet {
			t.Errorf("%s: knapp %v, erwartet %v", name, knapp, fall.erwartet)
		}
	}
}

// **Beendet heißt: eingerichtet, nicht laufend, und nicht von selbst** — ein
// Dienst, dessen Schalter aus ist, endet mit 0 und bleibt aus (OPS-10).
func TestDerDienstIstBeendet(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		stand    domain.Skriptdienstlage
		erwartet bool
	}{
		"nicht eingerichtet":      {domain.Skriptdienstlage{}, false},
		"läuft":                   {domain.Skriptdienstlage{Eingerichtet: true, Laeuft: true}, false},
		"endete mit 0":            {domain.Skriptdienstlage{Eingerichtet: true}, false},
		"endete mit einem Fehler": {domain.Skriptdienstlage{Eingerichtet: true, LetzterAusgang: 1}, true},
		"Frage nicht beantwortet": {domain.Skriptdienstlage{Eingerichtet: true, Unbekannt: true, LetzterAusgang: 1}, false},
	} {
		if beendet := fall.stand.Beendet(); beendet != fall.erwartet {
			t.Errorf("%s: beendet %v, erwartet %v", name, beendet, fall.erwartet)
		}
	}
}

// **Je Lage genau ein Hinweis, und jeder nennt sein Ziel.**
func TestDasBetriebsbandNenntJedeLageEinmalMitZiel(t *testing.T) {
	t.Parallel()

	alles := domain.Betriebslage{
		PlatzKnapp: true, FreiBytes: gib, SicherungAlt: true, SicherungTage: 20,
		DienstBeendet: true, DienstAusgang: 2, Webneustarts: 3,
	}

	hinweise := domain.Betriebshinweise(alles)
	if len(hinweise) != 4 {
		t.Fatalf("%d Hinweise: %+v", len(hinweise), hinweise)
	}

	gesehen := map[domain.Betriebshinweisart]bool{}
	for _, hinweis := range hinweise {
		if gesehen[hinweis.Art] || hinweis.Ziel == "" {
			t.Errorf("doppelt oder ohne Ziel: %+v", hinweis)
		}

		gesehen[hinweis.Art] = true
	}

	if leer := domain.Betriebshinweise(domain.Betriebslage{}); len(leer) != 0 {
		t.Errorf("ohne Lage %+v", leer)
	}
}

// **Ein Dienst, der immer wieder abstürzt, flattert nicht** — launchd startet
// ihn nach einem Fehler neu (`KeepAlive`, `SuccessfulExit` falsch), und wer
// ihn in den Millisekunden seines Laufs fragte, nähme den Hinweis zurück, nur
// damit er beim nächsten Abruf wiederkommt. **Zurück geht er erst**, wenn der
// Dienst zwischen zwei Abrufen durchgelaufen ist, oder sauber endete.
func TestEinAbstuerzenderDienstBleibtBeendet(t *testing.T) {
	t.Parallel()

	abgestuerzt := domain.Skriptdienstlage{Eingerichtet: true, LetzterAusgang: 1, Laeufe: 4}
	kurzWieder := domain.Skriptdienstlage{Eingerichtet: true, Laeuft: true, LetzterAusgang: 1, Laeufe: 5}
	durchgelaufen := domain.Skriptdienstlage{Eingerichtet: true, Laeuft: true, LetzterAusgang: 1, Laeufe: 5}

	for name, fall := range map[string]struct {
		jetzt, vorher domain.Skriptdienstlage
		warBeendet    bool
		gelaufen      time.Duration
		erwartet      bool
	}{
		"beendet ohne Vorgeschichte":           {abgestuerzt, domain.Skriptdienstlage{}, false, 0, true},
		"läuft ohne Vorgeschichte":             {kurzWieder, abgestuerzt, false, 0, false},
		"neu gestartet seit dem letzten Abruf": {kurzWieder, abgestuerzt, true, 0, true},
		"zwischen zwei Abrufen durchgelaufen":  {durchgelaufen, kurzWieder, true, domain.Erholfrist, false},
		// **Derselbe Lauf, aber nur kurz gesehen** (Zweitblick über Phase 4,
		// Befund 11) — zwei Abrufe kurz hintereinander lagen sonst beide in
		// einem Absturzlauf.
		"derselbe Lauf, kürzer als die Frist": {durchgelaufen, kurzWieder, true, domain.Erholfrist - time.Second, true},
		"sauber geendet":                      {domain.Skriptdienstlage{Eingerichtet: true, Laeufe: 6}, abgestuerzt, true, 0, false},
		"nicht zu fragen nimmt nichts zurück": {domain.Skriptdienstlage{Eingerichtet: true, Unbekannt: true}, abgestuerzt, true, 0, true},
		"nicht zu fragen meldet nichts Neues": {domain.Skriptdienstlage{Eingerichtet: true, Unbekannt: true}, abgestuerzt, false, 0, false},
		"ausgebaut nimmt den Hinweis zurück":  {domain.Skriptdienstlage{}, abgestuerzt, true, 0, false},
	} {
		if beendet := domain.DienstBeendet(fall.jetzt, fall.vorher, fall.warBeendet, fall.gelaufen); beendet != fall.erwartet {
			t.Errorf("%s: beendet %v, erwartet %v", name, beendet, fall.erwartet)
		}
	}
}
