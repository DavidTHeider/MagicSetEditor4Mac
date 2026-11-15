package app_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Eine MTG-JSON-Quelle wird geöffnet wie ein Setpaket.
func TestMTGJSONOeffnenLegtDasSetAn(t *testing.T) {
	probe := &importprobe{ergebnis: &service.Ergebnis{SetID: 5, Karten: 2}}

	antwort := mitImporteur(probe).MTGJSONOeffnen(probekontext(t), "/tmp/PRB.json", "magic")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if probe.pfad != "/tmp/PRB.json" || probe.spiel != "magic" {
		t.Errorf("geöffnet wurde %q für %q", probe.pfad, probe.spiel)
	}

	if antwort.SetID != "5" || antwort.Karten != 2 {
		t.Errorf("%+v", antwort)
	}
}

// **Was der Importeur über eine schon offene Quelle sagt, reicht die Fassade
// durch.**
//
// # Warum hier nicht mehr die Regel geprüft wird
//
// „Eine Quelle, ein Set" stand bis zum 2026-08-23 in dieser Fassade und galt
// damit **nur** für das Fenster: die Kommandozeile ging am Importeur
// unmittelbar vorbei und legte dasselbe Set ein zweites Mal an
// (#886).
//
// Die Regel steht jetzt im Fachkern und ist dort belegt
// (`internal/service/schonoffen_test.go`). Hier bleibt die Frage der Fassade:
// **kommt die Auskunft an?**
func TestMTGJSONOeffnenReichtSchonOffenDurch(t *testing.T) {
	probe := &importprobe{ergebnis: &service.Ergebnis{SetID: 3, Karten: 2, SchonOffen: true}}

	antwort := mitImporteur(probe).MTGJSONOeffnen(probekontext(t), "/tmp/PRB.json", "magic")

	if !antwort.SchonOffen || antwort.SetID != "3" {
		t.Errorf("%+v — erwartet war das bereits offene Set", antwort)
	}
}

// Ein leerer Pfad ist der Abbruch des Dateidialogs — kein Fehler.
func TestMTGJSONOeffnenMitLeeremPfadIstEinAbbruch(t *testing.T) {
	probe := &importprobe{}

	antwort := mitImporteur(probe).MTGJSONOeffnen(probekontext(t), "", "magic")
	if antwort.Fehler != nil || !antwort.Abgebrochen {
		t.Errorf("%+v", antwort)
	}
}

// Ein Fehler des Importeurs kommt als übersetzter Befund an.
func TestMTGJSONOeffnenMeldetBefunde(t *testing.T) {
	faelle := map[string]*importprobe{
		"das Spiel hat keine Zuordnung": {fehler: domain.NewFieldError(
			"spiel", domain.SchluesselValidierung, "keine Zuordnung")},
		"der Bestand streikt": {fehler: errors.New("die Platte ist voll")},
	}

	for name, probe := range faelle {
		t.Run(name, func(t *testing.T) {
			if mitImporteur(probe).MTGJSONOeffnen(probekontext(t), "/tmp/x.json", "magic").Fehler == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// Die Oberfläche erfährt, für welche Spiele es eine Zuordnung gibt.
//
// **Damit die Wahl nicht hier fällt.** `internal/app` trifft keine fachlichen
// Entscheidungen (§2); ein fest eingetragenes »magic« wäre genau das. Die Liste
// kommt aus der Fachbibliothek, und die Oberfläche zeigt sie.
func TestMTGJSONSpieleNenntDieBekanntenZuordnungen(t *testing.T) {
	spiele := app.NeueApp("").MTGJSONSpiele()

	if len(spiele) == 0 {
		t.Fatal("keine einzige Zuordnung")
	}

	gefunden := false

	for _, spiel := range spiele {
		if spiel == "magic" {
			gefunden = true
		}
	}

	if !gefunden {
		t.Errorf("Spiele: %v — »magic« fehlt", spiele)
	}
}

// Die Befunde reisen mit — mit ihrer Stelle in der Ausdrucksweise von JSON.
func TestMTGJSONOeffnenNenntDieBefunde(t *testing.T) {
	probe := &importprobe{ergebnis: &service.Ergebnis{
		SetID: 1, Karten: 2,
		Befunde: []domain.Befund{
			{Datei: "PRB.json", Ort: "PRB → Karte 1 → uuid", Meldung: "unbekanntes Feld"},
		},
	}}

	antwort := mitImporteur(probe).MTGJSONOeffnen(probekontext(t), "/tmp/PRB.json", "magic")

	if len(antwort.Befunde) != 1 {
		t.Fatalf("%+v", antwort.Befunde)
	}

	if antwort.Befunde[0] != "PRB.json:PRB → Karte 1 → uuid: unbekanntes Feld" {
		t.Errorf("%q", antwort.Befunde[0])
	}
}

// **Ein Fehler des Importeurs erreicht die Oberfläche.**
//
// Er entstand früher hier — beim Nachsehen, ob die Quelle schon offen ist.
// Jetzt entsteht er im Fachkern; geprüft wird, dass er nicht verlorengeht.
func TestMTGJSONOeffnenMeldetEinenFehlerDesImporteurs(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitImporteur(&importprobe{fehler: errors.New("nicht lesbar")})

	if anwendung.MTGJSONOeffnen(t.Context(), "/tmp/x.json", "magic").Fehler == nil {
		t.Error("der Fehler ging verloren")
	}
}
