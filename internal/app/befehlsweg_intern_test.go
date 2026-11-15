package app

// **Die Kommandozeile aus den Einstellungen** — OPS-03
// (#1720).
// Gegen ein echtes Dateisystem unter `t.TempDir()`: Verknüpfungen, fremde
// Dateien und ein PATH, der einen gleichnamigen Befehl vorn führt.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buendel legt ein Programmverzeichnis an, wie es in einem Bündel steht.
func buendel(t *testing.T) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "MagicSetEditor4Mac.app", "Contents", "MacOS")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatal(err)
	}

	// **Ein Bündel, das es gibt, sagt, wem es gehört** (Nachlauf über Phase
	// 4, Befund 9) — ohne `Info.plist` wäre es keines von hier.
	legeAn(t, filepath.Join(filepath.Dir(ort), "Info.plist"),
		"<plist><dict><key>CFBundleIdentifier</key>\n\t<string>"+Kennung+"</string></dict></plist>")

	for _, befehl := range befehlsnamen {
		if err := os.WriteFile(filepath.Join(ort, befehl), []byte("#!/bin/sh\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	return ort
}

func stande(v Befehlsvorschau) string {
	teile := make([]string, 0, len(v.Eintraege))
	for _, eintrag := range v.Eintraege {
		teile = append(teile, eintrag.Name+"="+eintrag.Stand)
	}

	return strings.Join(teile, " ")
}

func TestDieVorschauNenntZielBefehlUndPfad(t *testing.T) {
	ort := buendel(t)
	ziel := filepath.Join(t.TempDir(), "bin")

	v := befehlsvorschau(ziel, ort, "/usr/bin:/bin")

	if v.Verzeichnis != ziel || v.ImPfad || !strings.Contains(v.Pfadzeile, ziel) {
		t.Errorf("Verzeichnis %q, im Pfad %v, Zeile %q", v.Verzeichnis, v.ImPfad, v.Pfadzeile)
	}

	if stande(v) != "mse=frei mse-mcp=frei" {
		t.Errorf("Stände: %s", stande(v))
	}

	if v.Pruefbefehl != "mse meta" || v.Eintraege[0].Quelle != filepath.Join(ort, "mse") {
		t.Errorf("Prüfbefehl %q, Quelle %q", v.Pruefbefehl, v.Eintraege[0].Quelle)
	}

	if im := befehlsvorschau(ziel, ort, "/usr/bin:"+ziel); !im.ImPfad || im.Pfadzeile != "" {
		t.Errorf("im Pfad: %v, Zeile %q", im.ImPfad, im.Pfadzeile)
	}
}

// **Installieren legt Verknüpfungen, Entfernen nimmt nur eigene** — ein
// fremder gleichnamiger Befehl wird erkannt und nie überschrieben.
func TestInstallierenUndEntfernenFassenNurEigenesAn(t *testing.T) {
	ort := buendel(t)
	ziel := filepath.Join(t.TempDir(), "bin")

	if err := os.MkdirAll(ziel, 0o750); err != nil {
		t.Fatal(err)
	}

	fremd := filepath.Join(ziel, "mse-mcp")
	if err := os.WriteFile(fremd, []byte("fremd"), 0o600); err != nil {
		t.Fatal(err)
	}

	v, err := befehleInstallieren(befehlsvorschau(ziel, ort, ""))
	if err != nil {
		t.Fatal(err)
	}

	if stande(v) != "mse=eigen mse-mcp=fremd" {
		t.Errorf("nach dem Installieren: %s", stande(v))
	}

	if roh, lesefehler := os.ReadFile(fremd); lesefehler != nil || string(roh) != "fremd" { //nolint:gosec // G304: eine Datei dieses Tests
		t.Error("der fremde Befehl wurde überschrieben")
	}

	v, err = befehleEntfernen(v)
	if err != nil {
		t.Fatal(err)
	}

	if stande(v) != "mse=frei mse-mcp=fremd" {
		t.Errorf("nach dem Entfernen: %s", stande(v))
	}
}

// **Nach dem Verschieben des Bündels ist die Verknüpfung veraltet** — sie
// zeigt ins Leere; Installieren ersetzt sie, Entfernen nimmt sie weg.
func TestEineVeralteteVerknuepfungWirdErkanntUndErsetzt(t *testing.T) {
	alt := buendel(t)
	neu := buendel(t)
	ziel := t.TempDir()

	if err := os.Symlink(filepath.Join(alt, "mse"), filepath.Join(ziel, "mse")); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(filepath.Dir(filepath.Dir(alt))); err != nil {
		t.Fatal(err)
	}

	v := befehlsvorschau(ziel, neu, "")
	if stande(v) != "mse=veraltet mse-mcp=frei" {
		t.Fatalf("vorher: %s", stande(v))
	}

	v, err := befehleInstallieren(v)
	if err != nil || stande(v) != "mse=eigen mse-mcp=eigen" {
		t.Fatalf("nach dem Installieren: %s, %v", stande(v), err)
	}

	// **Eine Verknüpfung auf ein anderes Programm ist fremd**, auch wenn sie
	// eine Verknüpfung ist.
	anderes := filepath.Join(ziel, "mse-mcp")
	if err := os.Remove(anderes); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink("/usr/bin/true", anderes); err != nil {
		t.Fatal(err)
	}

	if stande(befehlsvorschau(ziel, neu, "")) != "mse=eigen mse-mcp=fremd" {
		t.Errorf("fremde Verknüpfung: %s", stande(befehlsvorschau(ziel, neu, "")))
	}
}

// **Ein gleichnamiger Befehl vorn im PATH verdeckt den eigenen** — die
// Vorschau nennt ihn, statt still einen anderen laufen zu lassen.
func TestEinGleichnamigerBefehlVornImPfadWirdGenannt(t *testing.T) {
	ort := buendel(t)
	vorn := t.TempDir()
	ziel := t.TempDir()

	if err := os.WriteFile(filepath.Join(vorn, "mse"), []byte("#!/bin/sh\n"), 0o700); err != nil { //nolint:gosec // ein Prüfbefehl muss ausführbar sein
		t.Fatal(err)
	}

	v := befehlsvorschau(ziel, ort, vorn+":"+ziel)
	if v.Eintraege[0].Verdeckt != filepath.Join(vorn, "mse") || v.Eintraege[1].Verdeckt != "" {
		t.Errorf("verdeckt: %q, %q", v.Eintraege[0].Verdeckt, v.Eintraege[1].Verdeckt)
	}
}

// **Ein Verzeichnis, das sich nicht anlegen lässt, ist ein Fehler** und kein
// stiller Rückfall an einen anderen Ort.
func TestEinUnbrauchbaresZielIstEinFehler(t *testing.T) {
	ort := buendel(t)
	datei := filepath.Join(t.TempDir(), "keinverzeichnis")

	if err := os.WriteFile(datei, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := befehleInstallieren(befehlsvorschau(filepath.Join(datei, "bin"), ort, "")); err == nil {
		t.Error("ein Ziel unter einer Datei wurde angenommen")
	}
}

// **Die Fassade**: leer liest, »installieren« und »entfernen« schreiben,
// alles andere ist ein Fehler — und ohne Programm oder Zuhause gibt es eine
// Meldung statt einer halben Vorschau.
func TestDieBefehlszeileAmFenster(t *testing.T) {
	ort := buendel(t)
	heim := t.TempDir()
	programm := func() (string, error) { return ort, nil }
	zuhause := func() (string, error) { return heim, nil }
	kaputt := func() (string, error) { return "", os.ErrNotExist }
	ctx := t.Context()

	if v := befehlszeileNeben(ctx, "", programm, zuhause, ""); v.Fehler != nil || stande(v) != "mse=frei mse-mcp=frei" ||
		v.Verzeichnis != filepath.Join(heim, ".local", "bin") {
		t.Errorf("lesen: %+v", v)
	}

	if v := befehlszeileNeben(ctx, "installieren", programm, zuhause, ""); v.Fehler != nil || stande(v) != "mse=eigen mse-mcp=eigen" {
		t.Errorf("installieren: %s, %+v", stande(v), v.Fehler)
	}

	if v := befehlszeileNeben(ctx, "entfernen", programm, zuhause, ""); v.Fehler != nil || stande(v) != "mse=frei mse-mcp=frei" {
		t.Errorf("entfernen: %s, %+v", stande(v), v.Fehler)
	}

	for name, v := range map[string]Befehlsvorschau{
		"unbekannte Handlung": befehlszeileNeben(ctx, "loeschen", programm, zuhause, ""),
		"kein Programm":       befehlszeileNeben(ctx, "", kaputt, zuhause, ""),
		"kein Zuhause":        befehlszeileNeben(ctx, "", programm, kaputt, ""),
	} {
		if v.Fehler == nil {
			t.Errorf("%s: kein Fehler", name)
		}
	}
}

// **Am Fenster liest sie unter dem Zuhause** — hier einem eigenen, damit kein
// Lauf das Verzeichnis dieses Rechners ansieht.
func TestDieBefehlszeileLiestUnterDemZuhause(t *testing.T) {
	heim := t.TempDir()
	t.Setenv("HOME", heim)

	// **zsh liest dabei die Einrichtung unter diesem Zuhause** — also keine;
	// die des Rechners bleibt unberührt.
	v := NeueApp("").Befehlszeile(t.Context(), "")
	if v.Fehler != nil || v.Verzeichnis != filepath.Join(heim, ".local", "bin") {
		t.Errorf("Verzeichnis %q, Fehler %+v", v.Verzeichnis, v.Fehler)
	}
}

// **Ein fremdes Bündel mit ähnlichem Namen ist nicht das eigene** (Zweitblick
// über Phase 4, Befund 8): `MacOS` und `MagicSetEditor4Mac*.app` genügten,
// und eine Verknüpfung dorthin galt als veraltet — „Alles widerrufen" nahm
// sie. Gibt es das Bündel, sagt sein `Info.plist`, wem es gehört.
func TestEinFremdesBuendelMitAehnlichemNamenIstFremd(t *testing.T) {
	neu := buendel(t)
	ziel := t.TempDir()

	for kennung, erwartet := range map[string]string{
		"com.fremd.kopie":              "mse=fremd mse-mcp=frei",
		"com.wails.MagicSetEditor4Mac": "mse=veraltet mse-mcp=frei",
	} {
		kopie := filepath.Join(t.TempDir(), "MagicSetEditor4Mac Kopie.app", "Contents")
		legeAn(t, filepath.Join(kopie, "MacOS", "mse"), "#!/bin/sh\n")
		legeAn(t, filepath.Join(kopie, "Info.plist"),
			"<plist><dict><key>CFBundleIdentifier</key>\n\t<string>"+kennung+"</string></dict></plist>")

		verweis := filepath.Join(ziel, "mse")
		_ = os.Remove(verweis) //nolint:errcheck // beim ersten Mal gibt es ihn nicht

		if err := os.Symlink(filepath.Join(kopie, "MacOS", "mse"), verweis); err != nil {
			t.Fatal(err)
		}

		if stand := stande(befehlsvorschau(ziel, neu, "")); stand != erwartet {
			t.Errorf("%s: %s, erwartet %s", kennung, stand, erwartet)
		}
	}
}

// **Ein umbenanntes eigenes Bündel bleibt eigen** (Nachlauf über Phase 4,
// Befund 9) — die Kennung entscheidet, nicht der Name; nur für ein Bündel,
// das fort ist, bleibt der Name der einzige Anhalt.
func TestEinUmbenanntesEigenesBuendelBleibtEigen(t *testing.T) {
	neu := buendel(t)
	ziel := t.TempDir()

	umbenannt := filepath.Join(t.TempDir(), "Umbenannt.app", "Contents")
	legeAn(t, filepath.Join(umbenannt, "MacOS", "mse"), "#!/bin/sh\n")
	legeAn(t, filepath.Join(umbenannt, "Info.plist"),
		"<plist><dict><key>CFBundleIdentifier</key>\n\t<string>"+Kennung+"</string></dict></plist>")

	if err := os.Symlink(filepath.Join(umbenannt, "MacOS", "mse"), filepath.Join(ziel, "mse")); err != nil {
		t.Fatal(err)
	}

	if stand := stande(befehlsvorschau(ziel, neu, "")); stand != "mse=veraltet mse-mcp=frei" {
		t.Errorf("Stand %s", stand)
	}
}
