package storage_test

// **Abschluss und Protokoll eines Turniers im Bestand** — DOM-04.AK-04
// (#1699),
// Vermerk 0139.
//
// **Die zweite Reihe**: die Regel „abgeschlossen nimmt nichts mehr an" steht
// im Dienst, und der Speicher hält sie noch einmal — ein Weg am Dienst vorbei
// schriebe sonst trotzdem.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// gespieltesTurnier legt ein Turnier an, meldet zwei an, legt eine Runde und
// trägt ihr Ergebnis ein. Es gibt die Kennungen von Turnier, Teilnehmer und
// Paarung.
func gespieltesTurnier(t *testing.T, ablage *storage.Turnierablage) (turnierID, teilnehmerID, paarungID int64) {
	t.Helper()

	angelegt, err := ablage.Anlegen(probekontext(t), domain.Turnier{Name: "Freitag", Spiel: "magic"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	namen := []string{"Anna", "Bert"}
	ids := make([]int64, 0, len(namen))

	for _, name := range namen {
		teilnehmer, anmeldefehler := ablage.Anmelden(probekontext(t), angelegt.ID, domain.Teilnehmer{Name: name})
		if anmeldefehler != nil {
			t.Fatalf("%s: %v", name, anmeldefehler)
		}

		ids = append(ids, teilnehmer.ID)
	}

	runde, err := ablage.RundeLegen(probekontext(t), angelegt.ID, 1,
		[]domain.Paarung{{A: ids[0], B: ids[1]}})
	if err != nil {
		t.Fatalf("die Runde: %v", err)
	}

	if err := ablage.ErgebnisSetzen(probekontext(t), fassungVon(t, ablage, angelegt.ID),
		runde.Paarungen[0].ID, &domain.Ergebnis{SpieleA: 2}); err != nil {
		t.Fatalf("das Ergebnis: %v", err)
	}

	return angelegt.ID, ids[1], runde.Paarungen[0].ID
}

func TestEinAbgeschlossenesTurnierNimmtNichtsMehrAn(t *testing.T) {
	ablage := turnierablage(t)
	turnierID, bert, paarung := gespieltesTurnier(t, ablage)

	if err := ablage.Abschliessen(probekontext(t), turnierID, fassungVon(t, ablage, turnierID)); err != nil {
		t.Fatalf("abschließen: %v", err)
	}

	gelesen, err := ablage.Turnier(probekontext(t), turnierID)
	if err != nil || gelesen.Abgeschlossen == "" {
		t.Fatalf("das Turnier trägt kein Datum: %+v, %v", gelesen, err)
	}

	faelle := map[string]func() error{
		"anmelden": func() error {
			_, fehler := ablage.Anmelden(probekontext(t), turnierID, domain.Teilnehmer{Name: "Cleo"})

			return fehler
		},
		"aussteigen": func() error { return ablage.Aussteigen(probekontext(t), bert, true) },
		"runde legen": func() error {
			_, fehler := ablage.RundeLegen(probekontext(t), turnierID, 2,
				[]domain.Paarung{{A: bert}})

			return fehler
		},
		"Ergebnis setzen": func() error {
			return ablage.ErgebnisSetzen(probekontext(t), fassungVon(t, ablage, turnierID), paarung, nil)
		},
	}

	for name, schreiben := range faelle {
		if err := schreiben(); !errors.Is(err, domain.ErrConflict) {
			t.Errorf("%s: %v, erwartet ErrConflict", name, err)
		}
	}

	// **Wiedereröffnet nimmt es wieder an.**
	if err := ablage.Wiedereroeffnen(probekontext(t), turnierID, fassungVon(t, ablage, turnierID)); err != nil {
		t.Fatalf("wiedereröffnen: %v", err)
	}

	if _, err := ablage.Anmelden(probekontext(t), turnierID, domain.Teilnehmer{Name: "Cleo"}); err != nil {
		t.Errorf("anmelden nach dem Wiedereröffnen: %v", err)
	}
}

// **Abschluss und Wiedereröffnen gehen gegen die Fassung** — und nur in die
// Richtung, in der sie etwas ändern.
func TestAbschlussGegenFassungUndStand(t *testing.T) {
	ablage := turnierablage(t)
	turnierID, _, _ := gespieltesTurnier(t, ablage)
	fassung := fassungVon(t, ablage, turnierID)

	if err := ablage.Abschliessen(probekontext(t), turnierID, fassung-1); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("eine alte Fassung: %v, erwartet ErrConflict", err)
	}

	if err := ablage.Wiedereroeffnen(probekontext(t), turnierID, fassung); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("ein offenes wiedereröffnen: %v, erwartet ErrConflict", err)
	}

	if err := ablage.Abschliessen(probekontext(t), 999, 1); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein fehlendes Turnier: %v, erwartet ErrNotFound", err)
	}

	if err := ablage.Abschliessen(probekontext(t), turnierID, fassung); err != nil {
		t.Fatalf("abschließen: %v", err)
	}

	if err := ablage.Abschliessen(probekontext(t), turnierID, fassungVon(t, ablage, turnierID)); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("zweimal abschließen: %v, erwartet ErrConflict", err)
	}
}

// **Jede Änderung schreibt eine Zeile — in der Reihenfolge, in der sie
// geschah.**
func TestDasProtokollNenntJedeAenderung(t *testing.T) {
	ablage := turnierablage(t)
	turnierID, bert, paarung := gespieltesTurnier(t, ablage)

	schritte := []func() error{
		func() error { return ablage.Aussteigen(probekontext(t), bert, true) },
		func() error { return ablage.Aussteigen(probekontext(t), bert, false) },
		func() error {
			return ablage.ErgebnisSetzen(probekontext(t), fassungVon(t, ablage, turnierID), paarung, nil)
		},
		func() error {
			return ablage.ErgebnisSetzen(probekontext(t), fassungVon(t, ablage, turnierID), paarung,
				&domain.Ergebnis{SpieleA: 2, SpieleB: 1})
		},
		func() error {
			return ablage.Abschliessen(probekontext(t), turnierID, fassungVon(t, ablage, turnierID))
		},
		func() error {
			return ablage.Wiedereroeffnen(probekontext(t), turnierID, fassungVon(t, ablage, turnierID))
		},
	}

	for stelle, schritt := range schritte {
		if err := schritt(); err != nil {
			t.Fatalf("Schritt %d: %v", stelle+1, err)
		}
	}

	protokoll, err := ablage.Protokoll(probekontext(t), turnierID)
	if err != nil {
		t.Fatalf("das Protokoll: %v", err)
	}

	erwartet := []struct {
		vorgang domain.Turniervorgang
		angaben string
	}{
		{domain.TurnierAngelegt, "Freitag"},
		{domain.TurnierAngemeldet, "Anna"},
		{domain.TurnierAngemeldet, "Bert"},
		{domain.TurnierRundeGelegt, "1"},
		{domain.TurnierErgebnis, "2:0:0"},
		{domain.TurnierAusgestiegen, "Bert"},
		{domain.TurnierEingestiegen, "Bert"},
		{domain.TurnierZurueckgenommen, ""},
		{domain.TurnierErgebnis, "2:1:0"},
		{domain.TurnierAbgeschlossen, ""},
		{domain.TurnierWiedereroeffnet, ""},
	}

	if len(protokoll) != len(erwartet) {
		t.Fatalf("%d Einträge, erwartet %d: %+v", len(protokoll), len(erwartet), protokoll)
	}

	for stelle, eintrag := range protokoll {
		if eintrag.Vorgang != erwartet[stelle].vorgang || eintrag.Angaben != erwartet[stelle].angaben ||
			eintrag.Zeit == "" {
			t.Errorf("Eintrag %d: %+v, erwartet %+v", stelle+1, eintrag, erwartet[stelle])
		}

		// **Ein Ergebnis nennt seine Paarung**, damit die Fläche die Namen
		// dazu findet.
		if (eintrag.Vorgang == domain.TurnierErgebnis || eintrag.Vorgang == domain.TurnierZurueckgenommen) &&
			eintrag.PaarungID != paarung {
			t.Errorf("Eintrag %d nennt die Paarung %d, erwartet %d", stelle+1, eintrag.PaarungID, paarung)
		}
	}
}

// **Ein gelöschtes Turnier nimmt sein Protokoll mit** — und das Protokoll
// eines Turniers, das es nicht gibt, ist nicht gefunden.
func TestDasProtokollGehtMitDemTurnier(t *testing.T) {
	ablage := turnierablage(t)
	turnierID, _, _ := gespieltesTurnier(t, ablage)

	if err := ablage.Loeschen(probekontext(t), turnierID); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	if _, err := ablage.Protokoll(probekontext(t), turnierID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}
}
