package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **`create directory:` entscheidet, ob eine Vorlage überhaupt schreiben
// darf** — #732.
//
//	void guard_export_info(const String& fun, bool need_template = false) {
//	  if (!export_info() && …) { throw ScriptError(_("Can only use ") + fun + …); }
//	  else if (export_info()->directory_relative.empty()) {
//	    throw ScriptError(_("Can only use ") + fun + _(" when 'create directory' is set to true"));
//	  }
//	}
//
// `src/script/functions/export.cpp`, Zeilen 26–33, Fassung f9d9356 vom
// 2021-07-06. **Jede** schreibende Funktion hängt daran.
//
// Über die 19 Vorlagen des M15-Packs: zehn sagen `true`, vier `false`, fünf
// sagen nichts — und nichts heißt `false`.
func TestOhneVerzeichnisDarfKeineVorlageSchreiben(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, skript string }{
		{"copy_file", `copy_file("x")`},
		{"write_text_file", `write_text_file("x", file: "y")`},
		{"write_set_file", `write_set_file(file: "y")`},
		{"write_image_file", `write_image_file("b.png", file: "y")`},
		{"symbols_to_html", `symbols_to_html("W", symbol_font: "mana")`},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			vorlage := vorlageMitFassung(fall.skript)
			vorlage.Verzeichnis = false

			_, _, err := mitSetsymbol(&beigabenmitschrift{}, &baumrastererprobe{},
				&symbolgestaltprobe{}).
				Fuehre(t.Context(), vorlage, eineKarte(), nil, nil)
			if err == nil {
				t.Fatal("der Aufruf ging durch")
			}

			if !strings.Contains(err.Error(), "create directory") {
				t.Errorf("Meldung: %v", err)
			}
		})
	}
}

// **Und `directory` trägt den relativen Namen.**
//
//	info.directory_relative = fn.GetName() + _("-files");
//	ctx.setVariable(_("directory"), to_script(info.directory_relative));
//
// `src/gui/html_export_window.cpp`, Zeilen 66 und 78, Fassung f9d9356 vom
// 2021-07-06. Die Vorlage baut damit
// ihre URLs: `replace(html, match:"…(href|src)='", replace:"&{directory}/")`
// steht in `magic-spoiler` Zeile 261.
func TestDirectoryStehtInDerUmgebung(t *testing.T) {
	t.Parallel()

	vorlage := vorlageMitFassung(`directory`)
	vorlage.Verzeichnis = true

	ausgabe, _, err := probelauf(service.Ausfuhrmittel{
		Beigaben: &beigabenmitschrift{}, Verzeichnis: "spoiler-files",
	}).
		Fuehre(t.Context(), vorlage, eineKarte(), nil, nil)
	if err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	if ausgabe != "spoiler-files" {
		t.Errorf("directory = %q", ausgabe)
	}
}

// **Ohne `create directory:` gibt es den Namen nicht.**
//
// Nicht „leer": ein leerer `directory`-Wert baute URLs wie `/bild.png`, und die
// zeigten auf die Wurzel des Dateisystems. Der unbekannte Name sagt, was fehlt.
func TestOhneVerzeichnisGibtEsDirectoryNicht(t *testing.T) {
	t.Parallel()

	vorlage := vorlageMitFassung(`directory`)
	vorlage.Verzeichnis = false

	_, _, err := probelauf(service.Ausfuhrmittel{
		Beigaben: &beigabenmitschrift{}, Verzeichnis: "ausfuhr-files",
	}).
		Fuehre(t.Context(), vorlage, eineKarte(), nil, nil)
	if err == nil || !strings.Contains(err.Error(), "directory") {
		t.Fatalf("Fehler = %v", err)
	}
}

// **Eine Vorlage ohne Angabe ist eine ohne Verzeichnis.**
//
// Der Vorgabewert des Feldes ist `false` (`src/data/export_template.hpp`,
// Fassung f9d9356 vom 2021-07-06), und
// fünf der 19 Vorlagen des Packs schreiben nichts hin.
func TestOhneAngabeGiltKeinVerzeichnis(t *testing.T) {
	t.Parallel()

	vorlage := msepaket.Exportvorlage{Skript: `copy_file("x")`, KurzerName: "ohne"}

	_, _, err := probelauf(service.Ausfuhrmittel{
		Beigaben: &beigabenmitschrift{}, Verzeichnis: "ausfuhr-files",
	}).
		Fuehre(t.Context(), vorlage, eineKarte(), nil, nil)
	if err == nil || !strings.Contains(err.Error(), "create directory") {
		t.Fatalf("Fehler = %v", err)
	}
}

// **Auch der indirekte Weg nennt `create directory`** — Befund des Zweitblicks
// durch Codex vom 2026-08-21.
//
//	} else if (is_substr(str, i, _("/sym"))) {
//	  if (!symbols.empty()) { … ret += symbols_to_html(symbols, *symbol_font, symbol_size); }
//
// `to_html`, `src/script/functions/export.cpp`, Zeilen 230–236 und 262–267,
// Fassung f9d9356 vom 2021-07-06: `to_html` mit einer Symbolschrift ruft
// `symbols_to_html`, und dort sitzt `guard_export_info`. **Zwei Stellen** — am
// schließenden `</sym>` und am Ende der Eingabe.
//
// Bis hierher blieb ohne Verzeichnis die Bibliotheksfassung von `to_html` mit
// nil-Wandler stehen. Geschrieben wurde nichts — richtig —, aber die Meldung
// lautete „symbols_to_html steht außerhalb einer Ausfuhr nicht zur Verfügung".
// Das schickt jemanden auf die Suche nach einer fehlenden Ausfuhr statt zu der
// Zeile, die zu ändern ist.
func TestAuchToHTMLNenntCreateDirectory(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, skript string }{
		{
			name:   "am schließenden sym",
			skript: `to_html("Kosten: <sym>2W</sym>!", symbol_font: "mana")`,
		},
		{
			name:   "am Ende der Eingabe",
			skript: `to_html("Kosten: <sym>2W", symbol_font: "mana")`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			vorlage := vorlageMitFassung(fall.skript)
			vorlage.Verzeichnis = false

			_, _, err := mitSetsymbol(&beigabenmitschrift{}, &baumrastererprobe{},
				&symbolgestaltprobe{}).
				Fuehre(t.Context(), vorlage, eineKarte(), nil, nil)
			if err == nil {
				t.Fatal("der Aufruf ging durch")
			}

			if !strings.Contains(err.Error(), "create directory") {
				t.Errorf("Meldung: %v", err)
			}
		})
	}

	// **Ohne Symbolschrift bleibt `to_html` das, was es überall ist** — reine
	// Zeichenkettenarbeit, und `create directory` hat damit nichts zu tun.
	vorlage := vorlageMitFassung(`to_html("<b>x</b>")`)
	vorlage.Verzeichnis = false

	ausgabe, _, err := mitSetsymbol(&beigabenmitschrift{}, &baumrastererprobe{},
		&symbolgestaltprobe{}).
		Fuehre(t.Context(), vorlage, eineKarte(), nil, nil)
	if err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	if ausgabe != "<b>x</b>" {
		t.Errorf("%q", ausgabe)
	}
}
