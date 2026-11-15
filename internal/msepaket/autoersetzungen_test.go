package msepaket_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Die Regeln eines Spiels werden gelesen** —
// #897.
//
// Bis zum 2026-08-23 gab es den Typ [msepaket.AutoErsetzung] und einen
// Ersetzer darüber — und **keine Stelle, die sie aus einem Paket liest**. Die
// Regeln standen in jedem Spielpaket und erreichten die Anwendung nie.
func TestDieAutoErsetzungenEinesSpielsWerdenGelesen(t *testing.T) {
	paket := spielMitBlock(t, `auto replace:
	match: (C)
	replace: ©
auto replace:
	match: AE
	replace: Æ
	whole word: false
auto replace:
	match: TM
	replace: ™
	whole word: false
`)

	regeln := paket.AutoErsetzungen()

	if len(regeln) != 3 {
		t.Fatalf("%d Regeln, erwartet drei: %+v", len(regeln), regeln)
	}

	if regeln[0].Muster != "(C)" || regeln[0].Ersatz != "©" {
		t.Errorf("die erste Regel ist %+v", regeln[0])
	}

	// **`whole word` steht standardmäßig auf `true`** — `AutoReplace::AutoReplace`,
	// `src/data/word_list.cpp`, Zeilen 38–42, Fassung f9d9356 vom 2021-07-06.
	if !regeln[0].GanzesWort {
		t.Error("ohne Angabe muss GanzesWort wahr sein")
	}

	if regeln[1].GanzesWort {
		t.Error("»whole word: false« wurde nicht gelesen")
	}
}

// **Eine Regel ohne Muster ist abgeschaltet** — `after_reading(AutoReplace&)`,
// `src/data/word_list.cpp`, Zeilen 51–55, Fassung f9d9356 vom 2021-07-06:
// „if (ar.match.empty()) ar.enabled = false".
//
// Sie wird deshalb gar nicht erst aufgenommen: eine abgeschaltete Regel, die
// mitreist, wäre ein Zustand, den niemand einschalten kann.
func TestEineRegelOhneMusterWirdUebergangen(t *testing.T) {
	paket := spielMitBlock(t, `auto replace:
	replace: ©
auto replace:
	match: TM
	replace: ™
`)

	regeln := paket.AutoErsetzungen()

	if len(regeln) != 1 {
		t.Fatalf("%d Regeln, erwartet eine: %+v", len(regeln), regeln)
	}

	if regeln[0].Muster != "TM" {
		t.Errorf("die falsche Regel blieb übrig: %+v", regeln[0])
	}
}

// **Eine abgeschaltete Regel wird nicht aufgenommen** —
// #914.
//
// `AutoReplace` liest `enabled` aus der Datei
// (`IMPLEMENT_REFLECTION_NO_SCRIPT`, `src/data/word_list.cpp`, Zeilen 44–49,
// Fassung f9d9356 vom 2021-07-06), und `tryAutoReplace` prüft es, bevor eine
// Regel greift (`src/gui/value/text.cpp`, Zeile 1038).
//
// **Hier stand, `enabled` sei der Schalter der Einstellungen und nicht des
// Pakets.** Das war erfunden: `use_auto_replace` ist der globale Schalter
// (`src/data/settings.hpp`, Zeile 83), `enabled` der je Regel — und
// `GameSettings::initDefaults` schiebt die Regelzeiger des **Spiels** in seine
// Liste (`src/data/settings.cpp`, Zeilen 88–104), sodass der Zustand des
// Pakets mitreist. Angestrichen vom Zweitblick über Phase R.
func TestEineAbgeschalteteRegelWirdNichtAufgenommen(t *testing.T) {
	paket := spielMitBlock(t, `auto replace:
	match: (C)
	replace: ©
	enabled: false
auto replace:
	match: TM
	replace: ™
`)

	regeln := paket.AutoErsetzungen()

	if len(regeln) != 1 {
		t.Fatalf("%d Regeln, erwartet eine: %+v", len(regeln), regeln)
	}

	if regeln[0].Muster != "TM" {
		t.Errorf("die abgeschaltete Regel blieb übrig: %+v", regeln[0])
	}
}

// **Ohne Angabe ist eine Regel eingeschaltet** — `AutoReplace::AutoReplace`,
// `src/data/word_list.cpp`, Zeilen 38–42, Fassung f9d9356 vom 2021-07-06.
//
// Der Gegenbeleg zum Fall darüber: ohne ihn ginge eine Prüfung durch, die
// **jede** Regel abwiese.
func TestOhneAngabeIstEineRegelEingeschaltet(t *testing.T) {
	paket := spielMitBlock(t, `auto replace:
	match: (C)
	replace: ©
`)

	if regeln := paket.AutoErsetzungen(); len(regeln) != 1 {
		t.Errorf("%d Regeln, erwartet eine", len(regeln))
	}
}

// **MSE kennt sechs Schreibweisen, nicht zwei** —
// #930.
//
// `Reader::handle(bool&)`, `src/util/io/reader.cpp`, Zeilen 324–332, Fassung
// f9d9356 vom 2021-07-06:
//
//	if (v==_("true") || v==_("1") || v==_("yes"))       b = true;
//	else if (v==_("false") || v==_("0") || v==_("no"))  b = false;
//	else warning(…)
//
// #914 hat nur
// `false` gelesen — den genannten Fall, nicht die Klasse (§10). Angestrichen
// vom Nachlauf des Zweitblicks über Phase R.
func TestAlleSechsSchreibweisenEinesBoolwerts(t *testing.T) {
	faelle := map[string]bool{"true": true, "1": true, "yes": true, "false": false, "0": false, "no": false}

	for wert, an := range faelle {
		t.Run("enabled: "+wert, func(t *testing.T) {
			regeln := spielMitBlock(t, "auto replace:\n\tmatch: (C)\n\treplace: ©\n\tenabled: "+wert+"\n").
				AutoErsetzungen()

			if (len(regeln) == 1) != an {
				t.Errorf("%d Regeln bei enabled: %s — erwartet war %v", len(regeln), wert, an)
			}
		})

		t.Run("whole word: "+wert, func(t *testing.T) {
			regeln := spielMitBlock(t, "auto replace:\n\tmatch: (C)\n\treplace: ©\n\twhole word: "+wert+"\n").
				AutoErsetzungen()

			if len(regeln) != 1 {
				t.Fatalf("%d Regeln", len(regeln))
			}

			if regeln[0].GanzesWort != an {
				t.Errorf("GanzesWort ist %v bei whole word: %s", regeln[0].GanzesWort, wert)
			}
		})
	}
}

// **Ein Wert, den MSE nicht kennt, ist ein Befund** — ebenda, der `else`-Zweig
// mit `warning(…)`.
//
// **Und die Regel bleibt an.** MSEs `warning` hält den Leser nicht auf; der
// Wert bleibt, was er vor dem Lesen war — bei `AutoReplace` also `true`
// (`src/data/word_list.cpp`, Zeilen 38–42, Fassung f9d9356 vom 2021-07-06).
func TestEinUnbekannterBoolwertWirdBenannt(t *testing.T) {
	paket := spielMitBlock(t, "auto replace:\n\tmatch: (C)\n\treplace: ©\n\tenabled: vielleicht\n")

	regeln, befunde := paket.AutoErsetzungenMitBefunden()

	if len(regeln) != 1 {
		t.Errorf("%d Regeln — ein unbekannter Wert darf die Regel nicht abschalten", len(regeln))
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet einer", len(befunde))
	}

	if !strings.Contains(befunde[0].Meldung, "vielleicht") {
		t.Errorf("der Befund nennt den Wert nicht: %s", befunde[0].Meldung)
	}
}

// **Und ein Spiel ohne Block gibt keine Regeln**, nicht eine leere Regel.
func TestEinSpielOhneBlockGibtKeineRegeln(t *testing.T) {
	if regeln := spielMitBlock(t, "").AutoErsetzungen(); len(regeln) != 0 {
		t.Errorf("%d Regeln, erwartet keine: %+v", len(regeln), regeln)
	}
}

// **Ein mehrzeiliger Wert kommt mit seinem Leerraum** — der sichtbarste Fall
// im echten Paket ist `match:` mit einem umschlossenen Bindestrich.
func TestEinMehrzeiligerWertBehaeltSeinenLeerraum(t *testing.T) {
	paket := spielMitBlock(t, "auto replace:\n\tmatch:\n\t\t - \n\treplace:\n\t\t — \n")

	regeln := paket.AutoErsetzungen()
	if len(regeln) != 1 {
		t.Fatalf("%d Regeln, erwartet eine", len(regeln))
	}

	if regeln[0].Muster != " - " {
		t.Errorf("das Muster ist %q, erwartet %q", regeln[0].Muster, " - ")
	}

	if regeln[0].Ersatz != " — " {
		t.Errorf("der Ersatz ist %q, erwartet %q", regeln[0].Ersatz, " — ")
	}
}

// spielMitBlock baut ein Spielpaket mit dem gegebenen Inhalt.
func spielMitBlock(t *testing.T, inhalt string) *msepaket.Paket {
	t.Helper()

	verzeichnis := filepath.Join(t.TempDir(), "probe.mse-game")
	if err := os.MkdirAll(verzeichnis, 0o750); err != nil {
		t.Fatalf("Verzeichnis anlegen: %v", err)
	}

	kopf := "mse version: 2.0.0\nshort name: probe\n"
	if err := os.WriteFile(filepath.Join(verzeichnis, "game"),
		[]byte(kopf+inhalt), 0o600); err != nil {
		t.Fatalf("game schreiben: %v", err)
	}

	paket, err := msepaket.Oeffne(verzeichnis)
	if err != nil {
		t.Fatalf("Paket öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	return paket
}
