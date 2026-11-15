package bezug_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
)

// rohesArchiv baut ein `tar.gz` ohne die Annahme, dass alles unter einer Wurzel
// liegt — für die Fälle, die ein wohlgeformtes Archiv nicht hergibt.
func rohesArchiv(t *testing.T, bauen func(*tar.Writer)) []byte {
	t.Helper()

	var gepackt bytes.Buffer

	packer := gzip.NewWriter(&gepackt)
	schreiber := tar.NewWriter(packer)

	bauen(schreiber)

	// **Ohne `Close` auf dem tar-Schreiber**, wo der Fall es verlangt: dann
	// fehlt der Abschluss, und der Leser läuft in ein unerwartetes Ende.
	_ = schreiber.Flush() //nolint:errcheck // der Fall verlangt den fehlenden Abschluss

	if err := packer.Close(); err != nil {
		t.Fatalf("gzip schließen: %v", err)
	}

	return gepackt.Bytes()
}

// Ein Eintrag ohne Verzeichnis gehört zu keinem Paket.
//
// GitHub packt alles unter eine Wurzel; ein Archiv aus anderer Hand muss das
// nicht. Ein Eintrag ohne `/` hat kein Innen und wird übergangen.
func TestEinEintragOhneVerzeichnisWirdUebergangen(t *testing.T) {
	inhalt := rohesArchiv(t, func(w *tar.Writer) {
		schreibeEintrag(t, w, &tar.Header{
			Name: "einzeln.txt", Mode: 0o644, Size: 4, Typeflag: tar.TypeReg,
		}, "vier")

		schreibeEintrag(t, w, &tar.Header{
			Name: "wurzel/data/gut.txt", Mode: 0o644, Size: 3, Typeflag: tar.TypeReg,
		}, "gut")

		if err := w.Close(); err != nil {
			t.Fatalf("tar schließen: %v", err)
		}
	})

	dienst := server(t, inhalt, "3.1.0")
	ziel := t.TempDir()

	if _, err := bezug.Hole(t.Context(), quelle(dienst, "3.1.0"), ziel, "MagicSetEditor4Mac/probe"); err != nil {
		t.Fatalf("Hole: %v", err)
	}

	if _, err := os.Stat(filepath.Join(ziel, "einzeln.txt")); err == nil {
		t.Error("der Eintrag ohne Verzeichnis ist ausgepackt worden")
	}
}

// Was weder Datei noch Verzeichnis noch Verweis ist, wird übergangen.
//
// Ein `tar` kann Gerätedateien und Warteschlangen führen. Sie anzulegen wäre
// eine Wirkung, die niemand angefordert hat.
func TestEineWarteschlangeWirdUebergangen(t *testing.T) {
	inhalt := rohesArchiv(t, func(w *tar.Writer) {
		schreibeEintrag(t, w, &tar.Header{
			Name: "wurzel/data/rohr", Mode: 0o644, Typeflag: tar.TypeFifo,
		}, "")

		if err := w.Close(); err != nil {
			t.Fatalf("tar schließen: %v", err)
		}
	})

	dienst := server(t, inhalt, "3.1.0")
	ziel := t.TempDir()

	ergebnis, err := bezug.Hole(t.Context(), quelle(dienst, "3.1.0"), ziel, "MagicSetEditor4Mac/probe")
	if err != nil {
		t.Fatalf("Hole: %v", err)
	}

	if len(ergebnis.Uebergangen) != 0 {
		t.Errorf("Übergangen: %v — eine Warteschlange ist kein Programm", ergebnis.Uebergangen)
	}

	if _, err := os.Stat(filepath.Join(ziel, "rohr")); err == nil {
		t.Error("die Warteschlange ist angelegt worden")
	}
}

// Ein abgeschnittenes Archiv wird gemeldet, nicht halb ausgepackt.
func TestEinAbgeschnittenesArchivWirdGemeldet(t *testing.T) {
	inhalt := rohesArchiv(t, func(w *tar.Writer) {
		// Ein Kopf, der 100 Byte ankündigt, und nichts dahinter.
		schreibeEintrag(t, w, &tar.Header{
			Name: "wurzel/data/kurz.txt", Mode: 0o644, Size: 100, Typeflag: tar.TypeReg,
		}, "zu kurz")
	})

	dienst := server(t, inhalt, "3.1.0")

	if _, err := bezug.Hole(t.Context(), quelle(dienst, "3.1.0"), t.TempDir(), "MagicSetEditor4Mac/probe"); err == nil {
		t.Error("kein Fehler — ein abgeschnittenes Archiv ist durchgegangen")
	}
}

// Ein Ziel, an dem schon ein Verzeichnis dieses Namens liegt, wird gemeldet.
func TestEinBelegterNameWirdGemeldet(t *testing.T) {
	dienst := server(t, archiv(t, "wurzel", []eintrag{
		{name: "data/README", inhalt: "eine Datei"},
	}), "3.1.0")

	ziel := t.TempDir()
	if err := os.Mkdir(filepath.Join(ziel, "README"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if _, err := bezug.Hole(t.Context(), quelle(dienst, "3.1.0"), ziel, "MagicSetEditor4Mac/probe"); err == nil {
		t.Error("kein Fehler")
	}
}

// Eine Adresse, aus der sich keine Anfrage bauen lässt, wird gemeldet.
func TestEineUnbrauchbareAdresseWirdGemeldet(t *testing.T) {
	dienst := server(t, archiv(t, "wurzel", probeeintraege()), "3.1.0")

	faelle := map[string]bezug.Repositorium{
		"eine Markenadresse mit Steuerzeichen": {
			Adresse: dienst.URL + "/archiv",
			Marken:  "http://beispiel.test/\x7f/",
			Stand:   "3.1.0",
		},
		"eine Archivadresse mit Steuerzeichen": {
			Adresse: "http://beispiel.test/\x7f",
			Marken:  dienst.URL + "/marke/",
			Stand:   "3.1.0",
		},
		"eine Archivadresse ohne Gegenstelle": {
			Adresse: "http://127.0.0.1:1/archiv",
			Marken:  dienst.URL + "/marke/",
			Stand:   "3.1.0",
		},
	}

	for name, quelle := range faelle {
		t.Run(name, func(t *testing.T) {
			if _, err := bezug.Hole(t.Context(), quelle, t.TempDir(), "MagicSetEditor4Mac/probe"); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// Eine Antwort, die mitten im Körper abbricht, wird gemeldet.
//
// Die Gegenstelle kündigt eine Länge an und schließt die Verbindung vorher. Wer
// das übergeht, packt ein halbes Archiv aus und hält es für ein ganzes.
func TestEineAbbrechendeAntwortWirdGemeldet(t *testing.T) {
	dienst := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/marke/") {
			return
		}

		verbindung, _, err := http.NewResponseController(w).Hijack()
		if err != nil {
			t.Errorf("übernehmen: %v", err)

			return
		}

		defer func() { _ = verbindung.Close() }() //nolint:errcheck // die Verbindung wird gerade abgeschnitten

		// Ein Kopf, der 4096 Byte ankündigt, und acht dahinter.
		if _, err := verbindung.Write([]byte(
			"HTTP/1.1 200 OK\r\nContent-Length: 4096\r\n\r\nzu wenig")); err != nil {
			t.Errorf("schreiben: %v", err)
		}
	}))

	dienst.Config.ErrorLog = keinLog()

	t.Cleanup(dienst.Close)

	if _, err := bezug.Hole(t.Context(), quelle(dienst, "3.1.0"), t.TempDir(), "MagicSetEditor4Mac/probe"); err == nil {
		t.Error("kein Fehler")
	}
}

// Ein Archiv mit einem unlesbaren Eintragskopf wird gemeldet.
//
// **Ein anderer Fall als ein abgeschnittener Eintrag**: dort bricht der Inhalt
// ab, hier die Struktur. Wer nur den einen prüft, hat den anderen nie
// ausgeführt.
func TestEinUnlesbarerEintragskopfWirdGemeldet(t *testing.T) {
	var gepackt bytes.Buffer

	packer := gzip.NewWriter(&gepackt)

	schreiber := tar.NewWriter(packer)
	schreibeEintrag(t, schreiber, &tar.Header{
		Name: "wurzel/data/gut.txt", Mode: 0o644, Size: 3, Typeflag: tar.TypeReg,
	}, "gut")

	if err := schreiber.Flush(); err != nil {
		t.Fatalf("tar leeren: %v", err)
	}

	// Ein Block, der kein Kopf ist — und keine Nullen, sonst gälte er als Ende.
	if _, err := packer.Write(bytes.Repeat([]byte("X"), 512)); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if err := packer.Close(); err != nil {
		t.Fatalf("gzip schließen: %v", err)
	}

	dienst := server(t, gepackt.Bytes(), "3.1.0")

	if _, err := bezug.Hole(t.Context(), quelle(dienst, "3.1.0"), t.TempDir(), "MagicSetEditor4Mac/probe"); err == nil {
		t.Error("kein Fehler")
	}
}

// schreibeEintrag legt einen Eintrag ins Archiv.
func schreibeEintrag(t *testing.T, w *tar.Writer, kopf *tar.Header, inhalt string) {
	t.Helper()

	if err := w.WriteHeader(kopf); err != nil {
		t.Fatalf("Kopf schreiben: %v", err)
	}

	if inhalt != "" {
		if _, err := w.Write([]byte(inhalt)); err != nil {
			t.Fatalf("Inhalt schreiben: %v", err)
		}
	}
}

// keinLog schluckt die Meldung, die `httptest` beim Abbruch schreibt.
//
// Sie ist erwartet und gehört nicht in die Ausgabe eines grünen Laufs — §3
// verlangt sie pristine.
func keinLog() *log.Logger { return log.New(io.Discard, "", 0) }

// **Das Wurzelverzeichnis des Archivs ist kein Ausbruch.**
//
// # Woher dieser Test kommt
//
// Aus einem echten Bezug. GitHub schreibt in sein Archiv **zuerst einen
// Verzeichniseintrag** für die Wurzel — `M15-Magic-Pack-3f7f8ce…/`, mit
// Schrägstrich am Ende und ohne Inhalt. Für ihn ist der Rest hinter dem ersten
// Schrägstrich **leer**, und die Wegprüfung verglich damit das Zielverzeichnis
// mit sich selbst: `ziel` beginnt nicht mit `ziel/`, also „zeigt aus dem
// Zielverzeichnis hinaus".
//
// Der Bezug scheiterte damit an **jedem** GitHub-Archiv — an genau dem Fall,
// für den er gebaut ist. Kein Test sah es: die Archive der Prüfung entstehen
// aus einer Liste von Dateien, und ein Verzeichniseintrag steht darin nicht.
//
// Das ist dieselbe Klasse wie die drei Befunde aus Phase 9: **echte Daten
// finden, was selbst gebaute nicht finden können.**
func TestDasWurzelverzeichnisIstKeinAusbruch(t *testing.T) {
	inhalt := rohesArchiv(t, func(w *tar.Writer) {
		// Genau so, wie GitHub es schreibt: der Wurzeleintrag zuerst.
		schreibeEintrag(t, w, &tar.Header{
			Name: "wurzel-3f7f8ce/", Mode: 0o755, Typeflag: tar.TypeDir,
		}, "")

		schreibeEintrag(t, w, &tar.Header{
			Name: "wurzel-3f7f8ce/data/", Mode: 0o755, Typeflag: tar.TypeDir,
		}, "")

		schreibeEintrag(t, w, &tar.Header{
			Name: "wurzel-3f7f8ce/data/gut.txt", Mode: 0o644, Size: 3, Typeflag: tar.TypeReg,
		}, "gut")

		if err := w.Close(); err != nil {
			t.Fatalf("tar schließen: %v", err)
		}
	})

	dienst := server(t, inhalt, "3.1.0")
	ziel := t.TempDir()

	if _, err := bezug.Hole(t.Context(), quelle(dienst, "3.1.0"), ziel, "MagicSetEditor4Mac/probe"); err != nil {
		t.Fatalf("Hole: %v", err)
	}

	// `quelle` fordert das Unterverzeichnis `data` an; die Datei landet deshalb
	// unmittelbar im Ziel.
	gelesen, err := os.ReadFile(filepath.Join(ziel, "gut.txt")) // #nosec G304 -- t.TempDir()
	if err != nil {
		t.Fatalf("die Datei unter dem Wurzeleintrag fehlt: %v", err)
	}

	if string(gelesen) != "gut" {
		t.Errorf("Inhalt %q", gelesen)
	}
}
