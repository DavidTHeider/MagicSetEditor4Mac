package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Der ganze Weg, einmal echt**: eine echte Datenbank, das echte Protokoll.
//
// Was hier geprüft wird, ist die **Naht** — dass der zweite Composition Root
// verdrahtet ist, dass er migriert und dass eine Sitzung von `initialize` bis
// `tools/call` durchläuft. Genau an dieser Stelle stand in Phase 5 der Fehler,
// den kein Test sah: eine Fassade, die die Anwendung nie erreichte (#130).
func TestEineSitzungVonAnfangBisEnde(t *testing.T) {
	// `app.Datenbankpfad` liest das Benutzerverzeichnis; ein eigenes je Test.
	t.Setenv("HOME", t.TempDir())

	sitzung := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize"}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"lesen.sets","arguments":{}}}`,
	}, "\n") + "\n"

	var aus strings.Builder

	if err := lauf(probekontext(t), nil, strings.NewReader(sitzung), &aus); err != nil {
		t.Fatalf("lauf: %v", err)
	}

	zeilen := strings.Split(strings.TrimSpace(aus.String()), "\n")
	if len(zeilen) != 3 {
		t.Fatalf("%d Antworten, erwartet drei (die Meldung bekommt keine):\n%s", len(zeilen), aus.String())
	}

	for i, muss := range []string{"protocolVersion", "schreiben.feld", "content"} {
		if !strings.Contains(zeilen[i], muss) {
			t.Errorf("Antwort %d: %s", i+1, zeilen[i])
		}
	}

	// Und die Datenbank ist dabei entstanden — der Server migriert selbst.
	if !strings.Contains(zeilen[2], "[]") {
		t.Errorf("erwartet war ein leerer Bestand: %s", zeilen[2])
	}

	// **Und er meldet die Fassung der Anwendung** (#896).
	//
	// Hier und nicht nur in `internal/mcp`: dort steht, dass die Vorstellung
	// die **gegebene** Fassung nennt; ob sie ihm gegeben wird, entscheidet
	// dieser Composition Root. Genau die Naht, für die es diesen Test gibt.
	if !strings.Contains(zeilen[0], `"version":"`+app.Fassung+`"`) &&
		!strings.Contains(zeilen[0], `"version": "`+app.Fassung+`"`) {
		t.Errorf("die Vorstellung nennt nicht die Fassung %s: %s", app.Fassung, zeilen[0])
	}
}

// Ohne Benutzerverzeichnis wird gemeldet statt geraten.
func TestOhneBenutzerverzeichnisMeldetDerServer(t *testing.T) {
	t.Setenv("HOME", "")

	if err := lauf(probekontext(t), nil, strings.NewReader(""), &strings.Builder{}); err == nil {
		t.Error("kein Fehler")
	}
}

// Ein Ort, an dem sich keine Datenbank anlegen lässt, wird gemeldet.
func TestEinUnbrauchbarerOrtWirdGemeldet(t *testing.T) {
	// Eine Datei dort, wo das Verzeichnis hingehört: `Oeffnen` scheitert daran.
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	if err := legeDateiAn(t, filepath.Join(zuhause, "Library")); err != nil {
		t.Skipf("der Ort ließ sich nicht blockieren: %v", err)
	}

	if err := lauf(probekontext(t), nil, strings.NewReader(""), &strings.Builder{}); err == nil {
		t.Error("kein Fehler")
	}
}

// legeDateiAn legt eine Datei an, wo ein Verzeichnis erwartet wird.
func legeDateiAn(t *testing.T, pfad string) error {
	t.Helper()

	return os.WriteFile(pfad, []byte("kein Verzeichnis"), 0o600)
}

// **Was der Server ändert, steht unter seinem Namen im Verlauf.**
//
// Seit dem MCP-Server schreiben zwei Prozesse auf dieselbe Datei
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)); eine
// Rücknahme, die nicht sagt, wessen Änderung sie umkehrt, ist eine Vermutung
// (#52). `Als` gibt es dafür — und ohne diesen Test wäre es eine Zusage, die
// niemand einlöst: der Server schriebe unter „fenster", und kein Lauf würde rot.
func TestWasDerServerAendertStehtUnterSeinemNamen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	pfad, err := app.Datenbankpfad()
	if err != nil {
		t.Fatalf("Pfad: %v", err)
	}

	if fehler := os.MkdirAll(filepath.Dir(pfad), 0o700); fehler != nil {
		t.Fatalf("Verzeichnis: %v", fehler)
	}

	// Ein Set mit einer Karte, angelegt wie das Fenster es täte.
	db, err := storage.Oeffnen(probekontext(t), "file:"+pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe",
		Cards: []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Feuerball"}}}},
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	aendern := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":` +
		`{"name":"schreiben.feld","arguments":{"setID":` +
		strconv.FormatInt(setID, 10) +
		`,"karteID":` +
		strconv.FormatInt(karten[0].ID, 10) + `,"feld":"name","wert":"Eisball"}}}`

	sitzung := `{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n" + aendern + "\n"

	if fehler := lauf(probekontext(t), nil, strings.NewReader(sitzung), &strings.Builder{}); fehler != nil {
		t.Fatalf("lauf: %v", fehler)
	}

	nachher, fehler := storage.Oeffnen(probekontext(t), "file:"+pfad)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = nachher.Close() })

	verlauf, fehler := storage.NeuerVerlaufsleser(nachher).Verlauf(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("Verlauf: %v", fehler)
	}

	if len(verlauf) != 1 {
		t.Fatalf("%d Einträge, erwartet einer", len(verlauf))
	}

	if verlauf[0].Urheber != domain.UrheberServer {
		t.Errorf("Urheber %q, erwartet %q", verlauf[0].Urheber, domain.UrheberServer)
	}
}
