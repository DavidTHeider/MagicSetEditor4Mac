package app

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// **Eine Exportvorlage darf Dateien danebenlegen** —
// #657.
//
// # Der Befund
//
// `write_image_file` gibt es seit
// #216 samt
// seiner Naht — und **niemand hat sie gesetzt**. `MitBildschreiber` kam in
// keiner Zeile Produktivcode vor, nur in Tests. In der laufenden Anwendung
// scheiterte damit jede der fünf Vorlagen, die die Funktion rufen, an „der
// Name `write_image_file` ist hier nicht bekannt": die Fähigkeit war gebaut
// und unerreichbar (§10).
//
// # Was hier geprüft wird
//
// Der Weg **durch** die Fassade, nicht der Ausführer für sich: eine Vorlage im
// Paketverzeichnis, ein Set aus der Saat, `VorlageAusfuehren` — und danach
// liegt die Datei da.
func TestEineVorlageLegtDateienInIhrVerzeichnis(t *testing.T) {
	// **Die Vorlage liegt im Paketverzeichnis, nicht in der Saat.** Sie prüft
	// vier Funktionen und gehört damit zum Wächter, nicht in das, was mit der
	// Anwendung ausgeliefert wird (§11).
	pakete := t.TempDir()
	beigabenvorlage(t, pakete)

	anwendung := mitSaatUndPaketen(t, pakete)

	geoeffnet := anwendung.SetOeffnen(t.Context(), saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "liste.txt")

	antwort := anwendung.VorlageAusfuehren(
		probekontext(t), geoeffnet.SetID, "probe-beigaben.mse-export-template", ziel, nil, nil)
	if antwort.Fehler != nil {
		t.Fatalf("ausführen: %+v (%v)", antwort.Fehler, antwort.Befunde)
	}

	// **In `<name>-files`, nicht daneben** (#732). MSE legt das Verzeichnis
	// an und nennt es der Vorlage als `directory`; die HTML-Ausgaben des
	// Packs bauen ihre URLs damit.
	neben := filepath.Join(filepath.Dir(ziel), "liste-files")

	faelle := []struct{ datei, warum string }{
		{"karte.png", "write_image_file legt ein Kartenbild daneben"},
		{"notiz.txt", "write_text_file schreibt eine Zeichenkette"},
		{"beigabe.txt", "copy_file holt eine Datei aus dem Vorlagenpaket"},
		{"kopie.mse-set", "write_set_file legt das Set daneben"},
	}

	for _, fall := range faelle {
		t.Run(fall.datei, func(t *testing.T) {
			if _, err := os.Stat(filepath.Join(neben, fall.datei)); err != nil {
				t.Errorf("%s: %v", fall.warum, err)
			}
		})
	}

	// **Und die Namen stehen in der Ausgabe.** Jede der vier Funktionen gibt
	// den Dateinamen zurück; die Vorlagen hängen ihn in ein `<img src="…">`
	// oder in eine Zeile.
	roh, err := os.ReadFile(ziel) //nolint:gosec // eigenes Temp
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if string(roh) != "karte.png|notiz.txt|beigabe.txt|kopie.mse-set" {
		t.Errorf("die Ausgabe lautet %q", string(roh))
	}
}

// beigabenvorlage legt eine Exportvorlage an, die alle vier Wege benutzt.
func beigabenvorlage(t *testing.T, wurzel string) {
	t.Helper()

	ort := filepath.Join(wurzel, "probe-beigaben.mse-export-template")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Eine Datei, die `copy_file` holt** — sie liegt im Vorlagenpaket und
	// soll neben der Ausgabe wieder auftauchen.
	if err := os.WriteFile(
		filepath.Join(ort, "beigabe.txt"), []byte("aus der Vorlage"), 0o600,
	); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	inhalt := `mse version: 2.0.0
short name: Beigaben
full name: Vorlage mit Beigaben
version: 2026-08-21
game: probe
file type: *.txt|*.txt
create directory: true

script:
	write_image_file(cards.0, file: "karte.png") + "|" +
	write_text_file("eine Notiz", file: "notiz.txt") + "|" +
	copy_file("beigabe.txt") + "|" +
	write_set_file(file: "kopie.mse-set")
`

	if err := os.WriteFile(
		filepath.Join(ort, "export-template"), []byte(inhalt), 0o600,
	); err != nil {
		t.Fatalf("schreiben: %v", err)
	}
}

// mitSaatUndPaketen startet eine Anwendung mit der Saat **und** einem eigenen
// Paketverzeichnis.
func mitSaatUndPaketen(t *testing.T, pakete string) *App {
	t.Helper()

	wurzel, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Wurzel: %v", err)
	}

	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.paketort = func() (string, error) { return pakete, nil }
	anwendung.saatort = func() string { return filepath.Join(wurzel, "saat") }

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	return anwendung
}

// **Ein Pfad statt eines Namens wird abgewiesen** —
// #657.
//
// Die Vorlage darf eine Datei **benennen**, nicht einen Ort **wählen**.
//
//	if (!ei.allow_writes_outside) {
//	  … throw ScriptError(_("Not a relative filename: ") + rel_name);
//	}
//
// `get_export_full_path`, `src/script/functions/export.cpp`, Zeilen 350–367,
// Fassung f9d9356 vom 2021-07-06. **Abgewiesen, nicht bereinigt**: wer aus
// `../../x` ein `....x` machte, schriebe eine Datei, die niemand gemeint hat.
func TestEineVorlageKannNichtAusDemAusgabeverzeichnisAusbrechen(t *testing.T) {
	pakete := t.TempDir()
	beigabenvorlage(t, pakete)

	anwendung := mitSaatUndPaketen(t, pakete)

	geoeffnet := anwendung.SetOeffnen(t.Context(), saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	daneben := t.TempDir()
	ziel := filepath.Join(t.TempDir(), "liste.txt")

	for _, name := range []string{
		filepath.Join(daneben, "entwischt.txt"),
		"../entwischt.txt",
		"unterordner/entwischt.txt",
		"",
		".",
		"..",
	} {
		t.Run(name, func(t *testing.T) {
			ausbrecher(t, pakete, name)

			antwort := anwendung.VorlageAusfuehren(
				probekontext(t), geoeffnet.SetID,
				"probe-ausbruch.mse-export-template", ziel, nil, nil)
			if antwort.Fehler == nil {
				t.Errorf("der Name %q ging durch", name)
			}

			if _, err := os.Stat(filepath.Join(daneben, "entwischt.txt")); err == nil {
				t.Errorf("die Datei liegt außerhalb: %s", daneben)
			}
		})
	}
}

// ausbrecher legt eine Vorlage an, die unter dem gegebenen Namen schreibt.
func ausbrecher(t *testing.T, wurzel, name string) {
	t.Helper()

	ort := filepath.Join(wurzel, "probe-ausbruch.mse-export-template")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\nshort name: Ausbruch\nfull name: Ausbruch\n" +
		"version: 2026-08-21\ngame: probe\nfile type: *.txt|*.txt\n\n" +
		"script:\n\twrite_text_file(\"weg\", file: " + strconv.Quote(name) + ")\n"

	if err := os.WriteFile(
		filepath.Join(ort, "export-template"), []byte(inhalt), 0o600,
	); err != nil {
		t.Fatalf("schreiben: %v", err)
	}
}

// **Ein Fehler beim Lesen oder Schreiben kommt durch.**
func TestEinFehlerBeimBeilegenErreichtDieFassade(t *testing.T) {
	pakete := t.TempDir()

	ort := filepath.Join(pakete, "probe-fehlt.mse-export-template")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\nshort name: Fehlt\nfull name: Fehlt\n" +
		"version: 2026-08-21\ngame: probe\nfile type: *.txt|*.txt\n\n" +
		"script:\n\tcopy_file(\"gibtsnicht.txt\")\n"

	if err := os.WriteFile(
		filepath.Join(ort, "export-template"), []byte(inhalt), 0o600,
	); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	anwendung := mitSaatUndPaketen(t, pakete)

	geoeffnet := anwendung.SetOeffnen(t.Context(), saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	antwort := anwendung.VorlageAusfuehren(
		probekontext(t), geoeffnet.SetID, "probe-fehlt.mse-export-template",
		filepath.Join(t.TempDir(), "liste.txt"), nil, nil)
	if antwort.Fehler == nil {
		t.Fatal("die fehlende Datei blieb ohne Fehler")
	}
}
