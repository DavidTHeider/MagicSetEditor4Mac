package app_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Vorschau einer Massenänderung an der Fassade** — Werkstück U6 von
// Phase 5.1 (#1417): Kennungen als Zeichenketten, Zeilen wie die Domäne sie
// gibt, und der Fehler übersetzt.
func TestFelderVorschauReichtDieAuswahlDurchUndUebersetztDieZeilen(t *testing.T) {
	probe := &massenprobe{}

	antwort := mitMassenschreiber(probe).FelderVorschau(probekontext(t), "10", []string{"1", "3"}, "typ", "Land")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if len(probe.karten) != 2 || probe.karten[0] != 1 || probe.karten[1] != 3 || probe.feld != "typ" || probe.wert != "Land" {
		t.Errorf("durchgereicht: %v %q %q", probe.karten, probe.feld, probe.wert)
	}

	if antwort.Feld != "typ" || antwort.Wert != "Land" || len(antwort.Zeilen) != 2 {
		t.Fatalf("Antwort: %+v", antwort)
	}

	if antwort.Zeilen[0].KarteID != "1" || antwort.Zeilen[0].Name != "Feuerball" ||
		antwort.Zeilen[0].Alt != "Zauber" || !antwort.Zeilen[0].GabAlt || antwort.Zeilen[0].Neu != "Land" ||
		antwort.Zeilen[0].Ausschluss != "" {
		t.Errorf("erste Zeile: %+v", antwort.Zeilen[0])
	}

	if antwort.Zeilen[1].KarteID != "3" || antwort.Zeilen[1].Ausschluss != domain.AusschlussGleich {
		t.Errorf("zweite Zeile: %+v", antwort.Zeilen[1])
	}
}

func TestFelderVorschauWeistUnlesbareKennungenAb(t *testing.T) {
	probe := &massenprobe{}
	anwendung := mitMassenschreiber(probe)

	if befund := anwendung.FelderVorschau(probekontext(t), "x", []string{"1"}, "typ", "Land").Fehler; befund == nil || befund.Feld != "setID" {
		t.Errorf("Set: %+v", befund)
	}

	if befund := anwendung.FelderVorschau(probekontext(t), "10", []string{"1", "abc"}, "typ", "Land").Fehler; befund == nil || befund.Feld != "id" {
		t.Errorf("Karte: %+v", befund)
	}

	if befund := anwendung.FelderVorschau(probekontext(t), "10", []string{"1"}, " ", "Land").Fehler; befund == nil || befund.Feld != "feld" {
		t.Errorf("Feld: %+v", befund)
	}

	if len(probe.karten) != 0 {
		t.Errorf("gelesen wurde trotzdem: %v", probe.karten)
	}
}

func TestFelderVorschauUebersetztDenFehler(t *testing.T) {
	probe := &massenprobe{fehler: domain.ErrNotFound}

	antwort := mitMassenschreiber(probe).FelderVorschau(probekontext(t), "10", []string{"1"}, "typ", "Land")
	if antwort.Fehler == nil || antwort.Fehler.Code != "nichtGefunden" {
		t.Errorf("Befund: %+v", antwort.Fehler)
	}

	if antwort.Zeilen == nil {
		t.Error("die Zeilen fehlen — eine leere Liste, kein `null`")
	}
}

// **Das Anwenden trägt die Erwartung mit** — je Karte Kennung, Altwert und
// ob es ihn gab; ein Konflikt nennt die abweichenden Karten wie jeder
// Massenfehler.
func TestFelderSetzenMitErwartungReichtDieErwartungDurch(t *testing.T) {
	probe := &massenprobe{}

	antwort := mitMassenschreiber(probe).FelderSetzenMitErwartung(probekontext(t), "10",
		[]app.FeldvorwertDTO{{KarteID: "1", Wert: "Zauber", GabEs: true}, {KarteID: "2"}}, "typ", "Land")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if antwort.Vorgang != "vorgang8" {
		t.Errorf("Vorgang %q", antwort.Vorgang)
	}

	if len(probe.erwartet) != 2 || probe.erwartet[0] != (domain.Feldvorwert{KarteID: 1, Wert: "Zauber", GabEs: true}) ||
		probe.erwartet[1] != (domain.Feldvorwert{KarteID: 2}) {
		t.Errorf("Erwartung: %+v", probe.erwartet)
	}
}

func TestFelderSetzenMitErwartungNenntDieAbweichendenKarten(t *testing.T) {
	probe := &massenprobe{fehler: &domain.BulkError{
		CardIDs: []int64{2, 5},
		Err:     errors.Join(domain.ErrConflict),
	}}

	antwort := mitMassenschreiber(probe).FelderSetzenMitErwartung(probekontext(t), "10",
		[]app.FeldvorwertDTO{{KarteID: "2"}, {KarteID: "5"}}, "typ", "Land")
	if antwort.Fehler == nil || antwort.Fehler.Code != "konflikt" {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if len(antwort.Betroffen) != 2 || antwort.Betroffen[0] != "2" || antwort.Betroffen[1] != "5" {
		t.Errorf("betroffen: %v", antwort.Betroffen)
	}
}

func TestFelderSetzenMitErwartungWeistUnlesbareKennungenAb(t *testing.T) {
	probe := &massenprobe{}
	anwendung := mitMassenschreiber(probe)

	if befund := anwendung.FelderSetzenMitErwartung(probekontext(t), "x", nil, "typ", "Land").Fehler; befund == nil || befund.Feld != "setID" {
		t.Errorf("Set: %+v", befund)
	}

	if befund := anwendung.FelderSetzenMitErwartung(probekontext(t), "10", []app.FeldvorwertDTO{{KarteID: "abc"}}, "typ", "Land").Fehler; befund == nil || befund.Feld != "id" {
		t.Errorf("Karte: %+v", befund)
	}

	if befund := anwendung.FelderSetzenMitErwartung(probekontext(t), "10", []app.FeldvorwertDTO{{KarteID: "1"}}, "", "Land").Fehler; befund == nil || befund.Feld != "feld" {
		t.Errorf("Feld: %+v", befund)
	}

	if probe.erwartet != nil {
		t.Errorf("geschrieben wurde trotzdem: %v", probe.erwartet)
	}
}
