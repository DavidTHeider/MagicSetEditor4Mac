package app_test

// **Abschluss und Protokoll an der Naht** — DOM-04.AK-04
// (#1699).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestAbschlussUndWiedereroeffnenGehenGegenDieFassung(t *testing.T) {
	anwendung, naht := turnieranwendung(t)

	if antwort := anwendung.TurnierAbschliessen(t.Context(), "7", "3"); antwort.Fehler != nil ||
		naht.abschluss.turnier != 7 || naht.abschluss.fassung != 3 || !naht.abschluss.zu {
		t.Errorf("abschließen: %+v, an der Naht %+v", antwort, naht.abschluss)
	}

	if antwort := anwendung.TurnierWiedereroeffnen(t.Context(), "7", "4"); antwort.Fehler != nil ||
		naht.abschluss.fassung != 4 || naht.abschluss.zu {
		t.Errorf("wiedereröffnen: %+v, an der Naht %+v", antwort, naht.abschluss)
	}

	for name, ruf := range map[string]func(string, string) app.SchreibAntwort{
		"abschließen": func(turnier, fassung string) app.SchreibAntwort {
			return anwendung.TurnierAbschliessen(t.Context(), turnier, fassung)
		},
		"wiedereröffnen": func(turnier, fassung string) app.SchreibAntwort {
			return anwendung.TurnierWiedereroeffnen(t.Context(), turnier, fassung)
		},
	} {
		if antwort := ruf("sieben", "1"); antwort.Fehler == nil || antwort.Fehler.Feld != "turnierID" {
			t.Errorf("%s mit krummem Turnier: %+v", name, antwort.Fehler)
		}

		if antwort := ruf("7", "eins"); antwort.Fehler == nil || antwort.Fehler.Feld != "fassung" {
			t.Errorf("%s mit krummer Fassung: %+v", name, antwort.Fehler)
		}
	}

	naht.fehler = domain.ErrConflict
	if antwort := anwendung.TurnierAbschliessen(t.Context(), "7", "3"); antwort.Fehler == nil ||
		antwort.Fehler.Code != app.CodeKonflikt {
		t.Errorf("ein Konflikt: %+v", antwort.Fehler)
	}
}

func TestDasProtokollKommtAlsDTO(t *testing.T) {
	anwendung, naht := turnieranwendung(t)

	antwort := anwendung.TurnierProtokoll(t.Context(), "7")
	if antwort.Fehler != nil || len(antwort.Eintraege) != 1 {
		t.Fatalf("das Protokoll: %+v", antwort)
	}

	eintrag := antwort.Eintraege[0]
	if eintrag.Zeit != "2026-09-27T12:00:00Z" || eintrag.Vorgang != "ergebnis" ||
		eintrag.Angaben != "2:0:0" || eintrag.PaarungID != "12" {
		t.Errorf("der Eintrag: %+v", eintrag)
	}

	if krumm := anwendung.TurnierProtokoll(t.Context(), "sieben"); krumm.Fehler == nil ||
		krumm.Fehler.Feld != "turnierID" || krumm.Eintraege == nil {
		t.Errorf("krumme Kennung: %+v", krumm)
	}

	naht.fehler = domain.ErrNotFound
	if fehlt := anwendung.TurnierProtokoll(t.Context(), "7"); fehlt.Fehler == nil ||
		fehlt.Fehler.Code != app.CodeNichtGefunden || fehlt.Eintraege == nil {
		t.Errorf("ein Nahtfehler: %+v", fehlt)
	}
}

// **Ein abgeschlossenes Turnier trägt sein Datum im DTO.**
func TestEinAbgeschlossenesTurnierNenntSeinDatum(t *testing.T) {
	anwendung, naht := turnieranwendung(t)
	naht.turnier = domain.Turnier{ID: 7, Name: "Freitag", Abgeschlossen: "2026-09-27T12:00:00Z"}

	if antwort := anwendung.Turnier(t.Context(), "7"); antwort.Turnier == nil ||
		antwort.Turnier.Abgeschlossen != "2026-09-27T12:00:00Z" {
		t.Errorf("das Turnier: %+v", antwort.Turnier)
	}
}
