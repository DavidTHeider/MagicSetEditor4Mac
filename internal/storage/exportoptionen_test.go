package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Optionen einer Exportvorlage werden je Vorlage behalten** —
// #492.
//
//	IndexMap<FieldP,ValueP>& exportOptionsFor(const ExportTemplate& export_template);
//
// `src/data/settings.hpp`, Zeilen 180–186, Fassung f9d9356 vom 2021-07-06 —
// **je Vorlage**, nicht je Set. Bis hierher lagen sie im Zustand der Seite und
// waren beim nächsten Öffnen fort.
//
// # Warum eine eigene Tabelle und nicht `settings`
//
// `domain.PruefeEinstellung` urteilt über eine **feste** Liste aus Schlüssel
// und Wert. Optionsfelder sind dynamisch: über `full-magic-pack` führen 21 von
// 22 Vorlagen zusammen **96** davon, und eine eigene Vorlage bringt weitere
// mit. Geprüft werden sie gegen die Felder ihrer Vorlage, nicht gegen eine
// Aufzählung.
func TestEineExportoptionUeberlebtDenNeustart(t *testing.T) {
	ablage := storage.NeueExportoptionen(frischeDatenbank(t))

	if err := ablage.Setzen(probekontext(t), "magic-forum.mse-export-template",
		"forum", "mtgsalvation"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	gelesen, err := ablage.Fuer(probekontext(t), "magic-forum.mse-export-template")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen["forum"] != "mtgsalvation" {
		t.Errorf("%v", gelesen)
	}
}

// **Je Vorlage, und das ist die Zusage.** Zwei Vorlagen mit einem Feld
// desselben Namens — `creator` steht in mehreren — teilen sich den Wert
// nicht.
func TestZweiVorlagenTeilenSichIhreOptionenNicht(t *testing.T) {
	ablage := storage.NeueExportoptionen(frischeDatenbank(t))

	for vorlage, wert := range map[string]string{
		"magic-forum.mse-export-template": "Anna",
		"magic-mws.mse-export-template":   "Bert",
	} {
		if err := ablage.Setzen(probekontext(t), vorlage, "creator", wert); err != nil {
			t.Fatalf("setzen: %v", err)
		}
	}

	eine, err := ablage.Fuer(probekontext(t), "magic-forum.mse-export-template")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if eine["creator"] != "Anna" {
		t.Errorf("%v", eine)
	}
}

// **Zweimal gesetzt heißt einmal gespeichert.**
func TestEineExportoptionWirdUeberschrieben(t *testing.T) {
	ablage := storage.NeueExportoptionen(frischeDatenbank(t))

	for _, wert := range []string{"mse", "mtgsalvation"} {
		if err := ablage.Setzen(probekontext(t), "v", "forum", wert); err != nil {
			t.Fatalf("setzen: %v", err)
		}
	}

	gelesen, err := ablage.Fuer(probekontext(t), "v")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen) != 1 || gelesen["forum"] != "mtgsalvation" {
		t.Errorf("%v", gelesen)
	}
}

// **Ohne Wahl ist die Abbildung leer, nicht nil.**
func TestOhneWahlIstDieOptionsabbildungLeer(t *testing.T) {
	gelesen, err := storage.NeueExportoptionen(frischeDatenbank(t)).
		Fuer(probekontext(t), "gibtesnicht")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen == nil || len(gelesen) != 0 {
		t.Errorf("%v", gelesen)
	}
}
