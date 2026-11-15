package main

// **Die atomare Gruppe am Dienst** — Werkstück A13.1b von Phase 5.3
// (#1452).
//
// **Die fachliche Zusage prüft `internal/flaechen`.** Hier steht die Frage
// davor: dass die Naht die Wahl übersetzt und den Fehler durchreicht, statt
// still nichts zu tun.

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Ohne Bestand kommt ein Fehler und keine leere Gruppe.**
//
// **Der Unterschied zählt**: „nichts geschrieben" und „die Frage ging
// schief" sind zwei Auskünfte, und ein Dienst, der die erste gäbe, wo die
// zweite gilt, sagte, die Übernahme sei geglückt.
func TestOhneBestandUebernimmtDerDienstKeineGruppe(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	var ohneBestand *sql.DB

	naht := &preise{db: ohneBestand}

	if _, fehler := naht.KIGruppe(probekontext(t), rest.KIGruppenbezug{
		Wahl: []rest.KIGruppenwahlsicht{
			{Kennung: "a1", Felder: []string{"name"}},
		},
	}); fehler == nil {
		t.Error("ohne Bestand wurde übernommen")
	}
}
