package service_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// paketserver liefert ein Archiv und kennt eine Freigabe.
//
// **Und öffnet dafür die Schleife** (#2000): der Bezug wählt nur öffentlich
// gerichtete Adressen, und dieser Server steht auf `127.0.0.1`. Der benannte
// Schalter öffnet die Schleife und nur sie — darum `t.Setenv`, und darum
// laufen die Fälle mit ihm nicht nebeneinander.
//
// **Der Name steht hier als Wort**, denn `service` kennt `netz` nicht, auch
// nicht im Lauf (`internal/schichten`); er ist `netz.SchleifeUmgebung`.
// Hieße er anders, wählte der Bezug die Schleife nicht, und jeder Fall hier
// würde rot.
func paketserver(t *testing.T) *httptest.Server {
	t.Helper()

	t.Setenv("MSE_NETZ_SCHLEIFE", "an")

	var gepackt bytes.Buffer

	packer := gzip.NewWriter(&gepackt)
	schreiber := tar.NewWriter(packer)

	for name, inhalt := range map[string]string{
		"wurzel/data/probe.mse-game/game": "short name: Probe\n",
		"wurzel/data/mse.exe":             "MZ",
	} {
		if err := schreiber.WriteHeader(&tar.Header{
			Name: name, Mode: 0o644, Size: int64(len(inhalt)), Typeflag: tar.TypeReg,
		}); err != nil {
			t.Fatalf("Kopf schreiben: %v", err)
		}

		if _, err := schreiber.Write([]byte(inhalt)); err != nil {
			t.Fatalf("Inhalt schreiben: %v", err)
		}
	}

	if err := schreiber.Close(); err != nil {
		t.Fatalf("tar schließen: %v", err)
	}

	if err := packer.Close(); err != nil {
		t.Fatalf("gzip schließen: %v", err)
	}

	dienst := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/tags/") {
			if !strings.HasSuffix(r.URL.Path, "3.1.0") {
				w.WriteHeader(http.StatusNotFound)
			}

			return
		}

		if _, err := w.Write(gepackt.Bytes()); err != nil {
			t.Errorf("ausliefern: %v", err)
		}
	}))

	t.Cleanup(dienst.Close)

	return dienst
}

// Ein Paket landet unter seinem Namen im Paketverzeichnis.
func TestEinPaketLandetImPaketverzeichnis(t *testing.T) {
	dienst := paketserver(t)
	wurzel := t.TempDir()

	ergebnis, err := service.NeuerPaketbezieher(wurzel, "MagicSetEditor4Mac/probe").
		Beziehen(t.Context(), dienst.URL, "3.1.0", "data", "M15-Magic-Pack")
	if err != nil {
		t.Fatalf("Beziehen: %v", err)
	}

	if _, err := os.Stat(filepath.Join(wurzel, "M15-Magic-Pack", "probe.mse-game", "game")); err != nil {
		t.Errorf("das Paket fehlt: %v", err)
	}

	if !strings.HasPrefix(ergebnis.Pruefsumme, "sha256:") {
		t.Errorf("Prüfsumme: %q", ergebnis.Pruefsumme)
	}

	if len(ergebnis.Uebergangen) != 1 || ergebnis.Uebergangen[0] != "mse.exe" {
		t.Errorf("Übergangen: %v", ergebnis.Uebergangen)
	}
}

// **Ein gescheiterter Bezug hinterlässt kein halbes Paket.**
//
// Sonst fände das nächste Öffnen ein Verzeichnis unter dem richtigen Namen und
// hielte es für vollständig.
func TestEinGescheiterterBezugHinterlaesstNichts(t *testing.T) {
	dienst := paketserver(t)
	wurzel := t.TempDir()

	_, err := service.NeuerPaketbezieher(wurzel, "MagicSetEditor4Mac/probe").
		Beziehen(t.Context(), dienst.URL, "main", "data", "M15-Magic-Pack")
	if err == nil {
		t.Fatal("kein Fehler")
	}

	eintraege, err := os.ReadDir(wurzel)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(eintraege) != 0 {
		t.Errorf("es liegt etwas herum: %v", eintraege)
	}
}

// Ein zweiter Bezug ersetzt den ersten.
func TestEinZweiterBezugErsetztDenErsten(t *testing.T) {
	dienst := paketserver(t)
	wurzel := t.TempDir()

	bezieher := service.NeuerPaketbezieher(wurzel, "MagicSetEditor4Mac/probe")

	for lauf := range 2 {
		if _, err := bezieher.Beziehen(t.Context(), dienst.URL, "3.1.0", "data", "Probe"); err != nil {
			t.Fatalf("Lauf %d: %v", lauf+1, err)
		}
	}

	eintraege, err := os.ReadDir(wurzel)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// **Die Sperrdatei zählt nicht mit.** Sie bleibt seit dem Nachlauf über
	// die Nacharbeit absichtlich liegen — `flock` hängt an der Inode, und wer
	// die Datei entfernt, lässt die Sperre zerfallen
	// (#1282,
	// Befund 5). Gezählt wird, was dieser Lauf prüft: **das Paket**.
	var pakete []string

	for _, eintrag := range eintraege {
		if !strings.HasSuffix(eintrag.Name(), service.Bezugssperrendung) {
			pakete = append(pakete, eintrag.Name())
		}
	}

	if len(pakete) != 1 {
		t.Errorf("%d Pakete (%v), erwartet eines", len(pakete), pakete)
	}
}

// Ein Name, der aus dem Verzeichnis führt, wird abgewiesen.
func TestEinAusbrechenderNameWirdAbgewiesen(t *testing.T) {
	wurzel := t.TempDir()

	for _, name := range []string{"", "..", ".", "unter/tief"} {
		t.Run(name, func(t *testing.T) {
			_, err := service.NeuerPaketbezieher(wurzel, "MagicSetEditor4Mac/probe").
				Beziehen(t.Context(), "https://example.test", "3.1.0", "data", name)
			if err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// Ein Paketverzeichnis, das es nicht gibt, wird gemeldet.
func TestEinFehlendesPaketverzeichnisWirdGemeldet(t *testing.T) {
	_, err := service.NeuerPaketbezieher(filepath.Join(t.TempDir(), "gibtsnicht"), "MagicSetEditor4Mac/probe").
		Beziehen(t.Context(), "https://example.test", "3.1.0", "data", "Probe")
	if err == nil {
		t.Error("kein Fehler")
	}
}

// **Die Herkunft steht vor dem Bezug** — als Angaben, nicht als Satz (§9).
func TestDieHerkunftStehtVorDemBezug(t *testing.T) {
	quelle, stand := service.NeuerPaketbezieher(t.TempDir(), "MagicSetEditor4Mac/probe").
		Herkunft("https://github.com/MagicSetEditorPacks/M15-Magic-Pack", "3.1.0")

	if quelle != "MagicSetEditorPacks/M15-Magic-Pack" || stand != "3.1.0" {
		t.Errorf("Herkunft: %q, %q", quelle, stand)
	}
}

// Ein Platz, der sich nicht räumen lässt, hält den Bezug an.
//
// **Zwei eigene Wege**: das Räumen des alten Paketes und das Umbenennen des
// neuen. Wer nur einen prüft, hat den anderen nie ausgeführt — und ein Bezug,
// der still am Umbenennen scheitert, meldet Erfolg über ein Paket, das nicht da
// ist.
func TestEinUnraeumbaresPaketHaeltDenBezugNichtMehrAn(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("als root lässt sich jedes Verzeichnis räumen")
	}

	dienst := paketserver(t)
	wurzel := t.TempDir()

	// Ein Paketverzeichnis, das schon da ist und dessen Inhalt sich nicht
	// löschen lässt: das Elternverzeichnis gibt kein Schreibrecht her.
	alt := filepath.Join(wurzel, "Probe")
	if err := os.MkdirAll(filepath.Join(alt, "innen"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.Chmod(alt, 0o500); err != nil { // #nosec G302 -- Verzeichnis
		t.Fatalf("Rechte setzen: %v", err)
	}

	t.Cleanup(func() {
		// **Am beiseitegelegten Ort**, denn dorthin ist er umbenannt worden.
		for _, ort := range []string{alt, alt + ".vorher"} {
			_ = os.Chmod(ort, 0o700) //nolint:errcheck,gosec // nur fürs Aufräumen
		}
	})

	// **Und der Bezug gelingt trotzdem.**
	//
	// # Was hier vorher stand
	//
	// „kein Fehler" — der Bezug **musste** scheitern, weil er das bisherige
	// Paket mit `os.RemoveAll` wegräumte und das hier nicht geht.
	//
	// # Warum die Zusage sich geändert hat
	//
	// Seit dem 2026-08-28 wird der bisherige Stand **beiseitegelegt** statt
	// gelöscht (#1170, Befund 2 des zweiten Freigabeblicks): erst umbenennen,
	// dann das neue Paket einsetzen, dann den alten wegräumen. Umbenennen
	// braucht Schreibrecht am **Elternverzeichnis**, und das ist da.
	//
	// Das ist eine bewusste Änderung der Semantik (§15, Satz 4): ein Paket,
	// dessen Inhalt sich nicht löschen lässt, hält den Bezug nicht mehr auf.
	// Der Rest bleibt als `Probe.vorher` liegen und **steht im Protokoll**.
	ergebnis, err := service.NeuerPaketbezieher(wurzel, "MagicSetEditor4Mac/probe").
		Beziehen(t.Context(), dienst.URL, "3.1.0", "data", "Probe")
	if err != nil {
		t.Fatalf("Beziehen: %v", err)
	}

	if _, statfehler := os.Stat(filepath.Join(ergebnis.Ort, "probe.mse-game")); statfehler != nil {
		t.Errorf("das bezogene Paket steht nicht am Ort: %v", statfehler)
	}

	// **Der bisherige Stand liegt daneben und ist nicht verloren.**
	if _, statfehler := os.Stat(filepath.Join(alt+".vorher", "innen")); statfehler != nil {
		t.Errorf("der bisherige Stand ist weg: %v", statfehler)
	}
}

// **Eine leere Paketwurzel bezieht nichts — sie löscht auch nichts.**
//
// # Woher diese Prüfung kommt
//
// Aus dem Zweitblick über Phase 9, und sie ist der schwerste Befund gewesen.
// Der Composition Root verwarf den Fehler von `Paketverzeichnis()` und baute
// den Bezieher mit **leerer** Wurzel. Daraus wird `filepath.Join("", name)` —
// ein **relativer** Pfad —, und `Beziehen` räumt ihn vor dem Umbenennen mit
// `os.RemoveAll`. Wer im falschen Verzeichnis stünde, verlöre ein
// gleichnamiges Verzeichnis darin.
//
// Der Weg dahin war real: eine Datei namens `pakete` an der Stelle des
// Verzeichnisses genügt, damit `Paketverzeichnis()` scheitert.
func TestEineLeerePaketwurzelWirdAbgewiesen(t *testing.T) {
	arbeitsverzeichnis := t.TempDir()

	// Ein Verzeichnis, das ein unbedachter Lauf löschen würde.
	opfer := filepath.Join(arbeitsverzeichnis, "Probe")
	if err := os.Mkdir(opfer, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Eine Marke darin**, nicht nur das Verzeichnis: ein geglückter Bezug
	// legte es neu an, und die Prüfung auf sein Dasein wäre grün geblieben,
	// obwohl der Inhalt fort ist.
	marke := filepath.Join(opfer, "gehoert-jemandem.txt")
	if err := os.WriteFile(marke, []byte("nicht löschen"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	vorher, err := os.Getwd()
	if err != nil {
		t.Fatalf("Arbeitsverzeichnis: %v", err)
	}

	if err := os.Chdir(arbeitsverzeichnis); err != nil {
		t.Fatalf("wechseln: %v", err)
	}

	t.Cleanup(func() {
		if fehler := os.Chdir(vorher); fehler != nil {
			t.Errorf("zurückwechseln: %v", fehler)
		}
	})

	dienst := paketserver(t)

	if _, err := service.NeuerPaketbezieher("", "MagicSetEditor4Mac/probe").
		Beziehen(t.Context(), dienst.URL, "3.1.0", "data", "Probe"); err == nil {
		t.Error("kein Fehler — eine leere Wurzel ist durchgegangen")
	}

	if _, err := os.Stat(marke); err != nil {
		t.Errorf("der Inhalt des fremden Verzeichnisses ist fort: %v", err)
	}
}

// Eine Wurzel, die kein absoluter Pfad ist, wird ebenfalls abgewiesen.
//
// Dieselbe Klasse: der Ort, an dem gelöscht wird, muss feststehen und darf nicht
// davon abhängen, wo der Prozess gerade steht.
func TestEineRelativePaketwurzelWirdAbgewiesen(t *testing.T) {
	if _, err := service.NeuerPaketbezieher("pakete", "MagicSetEditor4Mac/probe").
		Beziehen(t.Context(), "https://example.test", "3.1.0", "data", "Probe"); err == nil {
		t.Error("kein Fehler")
	}
}

// **Ein blockierter Platz für den bisherigen Stand hält den Bezug an.**
//
// `paketEinsetzen` legt den bisherigen Stand nach `<ort>.vorher`. Liegt dort
// schon ein **nicht leeres** Verzeichnis, lehnt `rename` ab — und das ist der
// eine Weg, auf dem der Einsetzfehler im Bezug selbst ankommt.
//
// **Er hat den Platz des alten Laufs eingenommen**, der über ein unräumbares
// Verzeichnis ging: seit dem 2026-08-28 wird nicht mehr geräumt, sondern
// umbenannt (#1170, Befund 3 des zweiten Freigabeblicks).
func TestEinBlockierterRueckwegHaeltDenBezugAn(t *testing.T) {
	dienst := paketserver(t)
	wurzel := t.TempDir()

	for _, ort := range []string{
		filepath.Join(wurzel, "Probe"),
		filepath.Join(wurzel, "Probe.vorher", "imweg"),
	} {
		if err := os.MkdirAll(ort, 0o750); err != nil {
			t.Fatalf("anlegen: %v", err)
		}
	}

	_, err := service.NeuerPaketbezieher(wurzel, "MagicSetEditor4Mac/probe").
		Beziehen(t.Context(), dienst.URL, "3.1.0", "data", "Probe")
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(err.Error(), "an seinen Platz") {
		t.Errorf("die Meldung %q sagt nicht, woran es lag", err.Error())
	}
}
