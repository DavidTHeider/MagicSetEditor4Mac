package service_test

// **Liegengebliebene eigene Reste gehen — fremde und frische bleiben** —
// OPS04 (#2350),
// Befund A-27 des Audits 0.9.0; **Zugehörigkeit statt Namensgestalt** seit dem
// Zweitblick über Phase 4, Befunde 1 und 12.

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// altern stellt die Zeit einer Datei zwei Stunden zurück — doppelt so weit
// wie die Restfrist.
func altern(t *testing.T, pfad string) {
	t.Helper()

	damals := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(pfad, damals, damals); err != nil {
		t.Fatal(err)
	}
}

func legeDatei(t *testing.T, pfad string) string {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(pfad), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(pfad, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	return pfad
}

func legeOrdner(t *testing.T, pfad string) string {
	t.Helper()

	legeDatei(t, filepath.Join(pfad, "teil"))

	return pfad
}

// pruefeRaeumen räumt, altert vorher alles außer `frisch` und hält fest, dass
// genau `alt` ging und `bleiben` blieb.
func pruefeRaeumen(t *testing.T, orte service.Restorte, alt, bleiben []string, frisch ...string) {
	t.Helper()

	for _, pfad := range append(slices.Clone(alt), bleiben...) {
		if !slices.Contains(frisch, pfad) {
			altern(t, pfad)
		}
	}

	entfernt, err := service.RestePflegen(orte, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	slices.Sort(entfernt)
	erwartet := slices.Clone(alt)
	slices.Sort(erwartet)

	if !slices.Equal(entfernt, erwartet) {
		t.Errorf("entfernt %v,\nerwartet %v", entfernt, erwartet)
	}

	for _, pfad := range bleiben {
		if _, err := os.Lstat(pfad); err != nil {
			t.Errorf("%s ist fort: %v", pfad, err)
		}
	}
}

// **Neben dem Bestand gehen nur die eigenen Erzeugnisse** (Befund 1): wer
// `MSE_BESTAND` nach `~/Dokumente` zeigt, teilt den Ordner — eine fremde
// `.notizen.2026` oder ein `.cache-2024` sind dort Daten anderer, und kein
// Unterordner gehört diesem Programm, nur weil er neben dem Bestand liegt.
func TestNebenDemBestandGehenNurDieEigenenErzeugnisse(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	bestand := filepath.Join(ordner, "bestand.db")
	legeDatei(t, bestand)

	alt := []string{
		legeOrdner(t, filepath.Join(ordner, ".rueckspiel-31337")),               // Zurückspielen
		legeOrdner(t, filepath.Join(ordner, ".sicherung-987654")),               // Sicherung mit Kennwort
		legeDatei(t, filepath.Join(ordner, ".sicherung-424242.unfertig")),       // das Archiv (Befund 12)
		legeDatei(t, filepath.Join(ordner, ".bestand-vor-0.9.0.db.entsteht-5")), // die Sicherung vor einer Migration
		legeDatei(t, filepath.Join(ordner, ".bestand-vor-0.9.0-0123456789ab.db.entsteht-6")),
		legeDatei(t, filepath.Join(ordner, ".rueckspiel-laeuft.json.123")), // die Rückspielmarke
		legeDatei(t, filepath.Join(ordner, ".rueckkehr-offen.json.9")),     // die Epochenmarke
	}
	bleiben := []string{
		bestand,
		legeDatei(t, filepath.Join(ordner, ".notizen.2026")),              // fremd, auch wenn es so aussieht
		legeOrdner(t, filepath.Join(ordner, ".cache-2024")),               // fremd
		legeDatei(t, filepath.Join(ordner, ".bild.png.123")),              // neben dem Bestand schreibt kein Bild
		legeDatei(t, filepath.Join(ordner, "Projekt", ".backup-1")),       // kein Unterordner wird betreten
		legeDatei(t, filepath.Join(ordner, ".rueckspiel-laeuft.json.v2")), // kein Zufall aus Ziffern
		// **Die Art gehört zum Erzeuger** (Nachlauf über Phase 4, Befund 1):
		// `os.MkdirTemp` legt Ordner an, das Archiv und `SchreibeAtomar`
		// Dateien — ein Eintrag der anderen Art ist nicht von hier.
		legeDatei(t, filepath.Join(ordner, ".rueckspiel-4711")),
		legeOrdner(t, filepath.Join(ordner, ".sicherung-55.unfertig")),
		legeOrdner(t, filepath.Join(ordner, ".rueckkehr-offen.json.8")),
		// **Und nur die Form, die eine Sicherung vor der Migration hat.**
		legeDatei(t, filepath.Join(ordner, ".bestand-vor-notizen.txt.entsteht-1")),
	}

	pruefeRaeumen(t, service.Restorte{Bestand: bestand}, alt, bleiben)
}

// **Im eigenen Beiwerk gilt die allgemeine Gestalt**, oben und eine Ebene
// tiefer — dort schreibt nur diese Anwendung.
func TestImBeiwerkGiltDieAllgemeineGestalt(t *testing.T) {
	t.Parallel()

	beiwerk := filepath.Join(t.TempDir(), "beiwerk")
	unter := filepath.Join(beiwerk, "set-1")

	frisch := legeDatei(t, filepath.Join(beiwerk, ".frisch.png.77"))
	alt := []string{
		legeDatei(t, filepath.Join(beiwerk, ".bild.png.123456789")), // SchreibeAtomar
		legeDatei(t, filepath.Join(unter, ".karte.png.42")),         // eine Ebene tiefer
	}
	bleiben := []string{
		frisch, // jünger als eine Stunde
		legeDatei(t, filepath.Join(beiwerk, ".DS_Store")),
		legeDatei(t, filepath.Join(beiwerk, "bild.png.123")), // ohne Punkt
		legeDatei(t, filepath.Join(unter, "karte.png")),
		legeDatei(t, filepath.Join(unter, "tiefer", ".x.png.1")), // zwei Ebenen tiefer
		legeDatei(t, filepath.Join(beiwerk, ".liste.txt.v2")),
	}

	pruefeRaeumen(t, service.Restorte{Eigene: []string{beiwerk, filepath.Join(beiwerk, "gibts-nicht")}}, alt, bleiben, frisch)
}

// **Im Paketordner nur oben** — darunter liegen bezogene Pakete, und ihr
// Inhalt ist fremd.
func TestImPaketordnerNurOben(t *testing.T) {
	t.Parallel()

	pakete := filepath.Join(t.TempDir(), "pakete")

	alt := []string{legeOrdner(t, filepath.Join(pakete, ".m15-magic-2718"))} // Paketbezug
	bleiben := []string{legeDatei(t, filepath.Join(pakete, "m15.mse-style", ".sicherung-12"))}

	pruefeRaeumen(t, service.Restorte{Flach: []string{pakete}}, alt, bleiben)
}

// **Eine Wurzel, die ein Verweis ist, wird nicht betreten** (Befund 1) —
// `os.ReadDir` folgte ihm, und ein Beiwerk, das auf `~/Bilder` zeigt,
// hätte dort geräumt.
func TestEineWurzelAlsVerweisWirdNichtBetreten(t *testing.T) {
	t.Parallel()

	draussen, ordner := t.TempDir(), t.TempDir()
	fremd := legeDatei(t, filepath.Join(draussen, ".bild.png.1"))
	altern(t, fremd)

	beiwerk := filepath.Join(ordner, "beiwerk")
	if err := os.Symlink(draussen, beiwerk); err != nil {
		t.Fatal(err)
	}

	bestand := filepath.Join(ordner, "bestand.db")
	if err := os.Symlink(filepath.Join(draussen, "bestand.db"), bestand); err != nil {
		t.Fatal(err)
	}

	legeDatei(t, filepath.Join(draussen, ".rueckspiel-1"))
	altern(t, filepath.Join(draussen, ".rueckspiel-1"))

	entfernt, err := service.RestePflegen(service.Restorte{Eigene: []string{beiwerk}, Flach: []string{beiwerk}}, time.Now())
	if err != nil || len(entfernt) != 0 {
		t.Errorf("entfernt %v, Fehler %v", entfernt, err)
	}

	if _, err := os.Stat(fremd); err != nil {
		t.Errorf("hinter dem Verweis geräumt: %v", err)
	}
}

// **Ein Rest, der ein Verweis ist, geht selbst** — sein Ziel bleibt.
func TestEinRestAlsVerweisNimmtSeinZielNichtMit(t *testing.T) {
	t.Parallel()

	beiwerk, draussen := t.TempDir(), t.TempDir()
	ziel := legeDatei(t, filepath.Join(draussen, "wichtig"))

	verweis := filepath.Join(beiwerk, ".sicherung-1")
	if err := os.Symlink(draussen, verweis); err != nil {
		t.Fatal(err)
	}

	if _, err := service.RestePflegen(service.Restorte{Eigene: []string{beiwerk}}, time.Now().Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Lstat(verweis); !os.IsNotExist(err) {
		t.Errorf("der Verweis steht noch: %v", err)
	}

	if inhalt, err := os.ReadFile(ziel); err != nil || string(inhalt) != "x" { //nolint:gosec // eine Datei des Laufs
		t.Errorf("das Ziel des Verweises: %q, %v", inhalt, err)
	}
}

// **Was sich nicht räumen lässt, wird gemeldet** — und der Rest geht weiter:
// ein Ordner, der eine Datei ist, und ein Rest in einem Ordner ohne
// Schreibrecht.
func TestEinUnraeumbarerRestWirdGemeldet(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	datei := legeDatei(t, filepath.Join(ordner, "keinordner"))

	gesperrt := filepath.Join(ordner, "gesperrt")
	rest := legeDatei(t, filepath.Join(gesperrt, ".bild.png.7"))
	altern(t, rest)

	if err := os.Chmod(gesperrt, 0o500); err != nil { //nolint:gosec // nur lesbar, für diesen Lauf
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(gesperrt, 0o700) }) //nolint:errcheck,gosec // Aufräumen

	entfernt, err := service.RestePflegen(service.Restorte{Eigene: []string{datei, gesperrt}}, time.Now())
	if err == nil || len(entfernt) != 0 {
		t.Errorf("entfernt %v, Fehler %v — erwartet zwei Meldungen", entfernt, err)
	}

	if _, lesefehler := os.Stat(rest); lesefehler != nil {
		t.Errorf("der Rest ist fort, obwohl der Ordner gesperrt war: %v", lesefehler)
	}
}
