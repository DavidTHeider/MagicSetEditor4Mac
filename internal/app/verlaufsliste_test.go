package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Verlauf als Liste von Schritten an der Fassade** — Werkstück U6 von
// Phase 5.1 (#1417): eine Massenänderung ist ein Schritt mit Umfang, jede
// Zeile trägt Vorher und Nachher, und der jüngste Schritt steht **zuerst** —
// die Seite zeigt, was eben geschah.
func TestDerVerlaufKommtAlsSchritteAn(t *testing.T) {
	probe := &verlaufsprobe{eintraege: []domain.Verlaufseintrag{
		{
			ID: 1, KarteID: 7, Zeitpunkt: "2026-09-09T02:00:00Z", Handlung: domain.HandlungFeld,
			Feld: "name", Vorher: "Alt", GabVorher: true, Nachher: "Neu", GabNachher: true,
			Urheber: domain.UrheberFenster,
		},
		// **Der erste Eintrag einer Gruppe vertritt sie** (#2396) — er trägt
		// deshalb jedes Feld, das die Abbildung zeigen muss.
		{
			ID: 2, KarteID: 8, Zeitpunkt: "2026-09-09T02:01:00Z", Handlung: domain.HandlungFeld,
			Feld: "typ", Vorher: "Zauber", GabVorher: true, Nachher: "Land", GabNachher: true,
			Gruppe: "g1", Zurueckgenommen: true, Urheber: domain.UrheberServer, Herkunft: "https://scryfall.com/x",
		},
		{
			ID: 3, KarteID: 7, Zeitpunkt: "2026-09-09T02:01:00Z", Handlung: domain.HandlungFeld,
			Feld: "typ", Nachher: "Land", GabNachher: true, Gruppe: "g1", Zurueckgenommen: true,
			Urheber: domain.UrheberServer,
		},
	}}

	antwort := mitVerlauf(probe).Verlauf(probekontext(t), "3")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if probe.gefragt != 3 {
		t.Errorf("gefragt wurde nach Set %d", probe.gefragt)
	}

	if len(antwort.Schritte) != 2 {
		t.Fatalf("%d Schritte statt 2: %+v", len(antwort.Schritte), antwort.Schritte)
	}

	juengster := antwort.Schritte[0]
	if juengster.ID != "2" || juengster.Vorgang != "g1" || juengster.Umfang != 2 ||
		juengster.KarteID != "8" || juengster.Feld != "typ" || juengster.Vorher != "Zauber" ||
		!juengster.GabVorher || juengster.Nachher != "Land" || !juengster.GabNachher ||
		!juengster.Zurueckgenommen || juengster.Urheber != "server" ||
		juengster.Herkunft != "https://scryfall.com/x" || juengster.Zeitpunkt != "2026-09-09T02:01:00Z" {
		t.Errorf("jüngster Schritt: %+v", juengster)
	}

	aelterer := antwort.Schritte[1]
	if aelterer.ID != "1" || aelterer.Vorgang != "" || aelterer.Umfang != 1 || aelterer.Handlung != "feld" ||
		aelterer.Vorher != "Alt" || aelterer.Nachher != "Neu" || aelterer.Urheber != "fenster" {
		t.Errorf("älterer Schritt: %+v", aelterer)
	}
}

func TestDerVerlaufWeistEineUnlesbareSetkennungAb(t *testing.T) {
	antwort := mitVerlauf(&verlaufsprobe{}).Verlauf(probekontext(t), "abc")
	if antwort.Fehler == nil || antwort.Fehler.Feld != "setID" {
		t.Errorf("%+v", antwort.Fehler)
	}

	if antwort.Schritte == nil {
		t.Error("die Schritte fehlen — eine leere Liste, kein `null`")
	}
}

func TestDerVerlaufUebersetztDenFehler(t *testing.T) {
	antwort := mitVerlauf(&verlaufsprobe{fehler: domain.ErrNotFound}).Verlauf(probekontext(t), "3")
	if antwort.Fehler == nil || antwort.Fehler.Code != "nichtGefunden" {
		t.Errorf("%+v", antwort.Fehler)
	}
}
