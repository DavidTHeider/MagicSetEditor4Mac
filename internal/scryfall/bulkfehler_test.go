// **Die Fehlerwege der Sammeldateien** —
// #888.
//
// **Eine eigene Datei, weil `bulk_test.go` mit ihr über 400 Zeilen ginge**
// (§2). Der Schnitt folgt der Frage: dort steht, was **geht**, hier, was
// passiert, wenn es nicht geht.
package scryfall_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// **Die Fehlerwege der Sammeldateien.**
//
// # Warum sie zusammenstehen
//
// Sie beantworten dieselbe Frage — *was passiert, wenn etwas nicht geht* —
// und jeder von ihnen wäre einzeln ein Absatz mit drei Zeilen Vorrichtung.
func TestBulkfehlerwege(t *testing.T) {
	t.Parallel()

	t.Run("das Verzeichnis meldet einen Ausfall", func(t *testing.T) {
		t.Parallel()

		klient := klientFuer(t, "http://unerreichbar.invalid",
			netz.NeuerKlient(netz.UeblicheFristen()))

		if _, _, fehler := klient.Bulkverzeichnis(t.Context()); fehler == nil {
			t.Fatal("der Ausfall kam nicht durch")
		}
	})

	t.Run("das Verzeichnis meldet einen unlesbaren Rumpf", func(t *testing.T) {
		t.Parallel()

		server, _ := gegenstelleMit(t, `{kein json`)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		if _, _, fehler := klient.Bulkverzeichnis(t.Context()); !errors.Is(
			fehler, domain.ErrValidation) {
			t.Fatalf("Fehler = %v", fehler)
		}
	})

	t.Run("ein unerreichbarer Abruf kommt durch", func(t *testing.T) {
		t.Parallel()

		klient := klientFuer(t, "http://unerreichbar.invalid",
			netz.NeuerKlient(netz.UeblicheFristen()))

		if fehler := klient.BulkHolen(t.Context(), scryfall.Bulkeintrag{
			Adresse: "http://unerreichbar.invalid/x.jsonl.gz",
		}, filepath.Join(t.TempDir(), "x")); fehler == nil {
			t.Fatal("der Ausfall kam nicht durch")
		}
	})

	t.Run("ein unbeschreibbares Ziel wird gemeldet", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(
			func(schreiben http.ResponseWriter, _ *http.Request) {
				// **Eine gültige Sammeldatei** — seit #1027 prüft
				// `BulkHolen` den Inhalt, **bevor** es schreibt. Mit
				// „egal" käme dieser Lauf am Schreibweg gar nicht mehr an
				// und prüfte einen anderen Zweig als seinen.
				if _, fehler := schreiben.Write(gezipptesJSONL(t,
					`{"object":"card","id":"a","name":"Lightning Bolt"}`)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			}))
		t.Cleanup(server.Close)

		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		// **Ein Ziel in einem Ordner, den es nicht gibt.**
		ziel := filepath.Join(t.TempDir(), "gibtesnicht", "oracle.jsonl.gz")

		if fehler := klient.BulkHolen(t.Context(), scryfall.Bulkeintrag{
			Adresse: server.URL + "/x.jsonl.gz",
		}, ziel); fehler == nil {
			t.Fatal("das unbeschreibbare Ziel blieb ohne Fehler")
		}
	})

	// **Ein Ziel, das ein Verzeichnis ist, scheitert beim Umbenennen.**
	//
	// Der Schreibvorgang glückt — die Nebendatei entsteht neben dem
	// Verzeichnis —, und erst `os.Rename` scheitert. Das ist der einzige Weg,
	// diesen Zweig zu betreten, und er ist nicht erfunden: ein Ordner mit dem
	// Namen einer Sammeldatei kann durch eine frühere Fassung entstanden sein.
	t.Run("ein Ziel, das ein Verzeichnis ist", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(
			func(schreiben http.ResponseWriter, _ *http.Request) {
				// **Eine gültige Sammeldatei** — seit #1027 prüft
				// `BulkHolen` den Inhalt, **bevor** es schreibt. Mit
				// „egal" käme dieser Lauf am Schreibweg gar nicht mehr an
				// und prüfte einen anderen Zweig als seinen.
				if _, fehler := schreiben.Write(gezipptesJSONL(t,
					`{"object":"card","id":"a","name":"Lightning Bolt"}`)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			}))
		t.Cleanup(server.Close)

		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		ziel := filepath.Join(t.TempDir(), "oracle.jsonl.gz")
		if fehler := os.Mkdir(ziel, 0o750); fehler != nil {
			t.Fatalf("das Verzeichnis anlegen: %v", fehler)
		}

		fehler := klient.BulkHolen(t.Context(), scryfall.Bulkeintrag{
			Adresse: server.URL + "/x.jsonl.gz",
		}, ziel)
		if fehler == nil {
			t.Fatal("das Verzeichnis als Ziel blieb ohne Fehler")
		}

		// **„einsetzen" und nicht mehr „umbenennen"** (#1163): Schreiben,
		// Schließen und Umbenennen sind zu einem Zweig geworden, weil die
		// beiden ersten an einer eben angelegten Datei nicht scheitern können
		// (§3). Die Meldung nennt deshalb den Schritt, der es kann.
		if !strings.Contains(fehler.Error(), "einsetzen") {
			t.Errorf("Fehler = %v — er kommt aus dem falschen Schritt", fehler)
		}
	})

	t.Run("eine zu lange Zeile wird gemeldet", func(t *testing.T) {
		t.Parallel()

		// **Über der Zeilengrenze.** Eine Datei mit einer 5-MB-Zeile ist keine
		// Sammeldatei von Scryfall; ein Kartenobjekt wiegt wenige Kilobyte.
		lang := `{"object":"card","name":"` + strings.Repeat("x", 5<<20) + `"}`

		ort := filepath.Join(t.TempDir(), "lang.jsonl.gz")
		if fehler := os.WriteFile(ort, gezipptesJSONL(t, lang), 0o600); fehler != nil {
			t.Fatalf("ablegen: %v", fehler)
		}

		if _, _, fehler := scryfall.BulkKarteNachName(ort, "X"); fehler == nil {
			t.Fatal("die zu lange Zeile blieb ohne Fehler")
		}
	})
}

// **Ein schon vorliegender Stand wird nicht noch einmal geholt.**
//
// Der Fall entsteht, wenn zwei Prozesse dieselbe Sammeldatei wollen: einer
// lädt, der andere wartet — und findet sie danach vor
// (#1163,
// Befund Z7-03).
func TestEinSchonVorliegenderStandWirdNichtGeholt(t *testing.T) {
	t.Parallel()

	var gefragt int

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt++

			if _, err := schreiben.Write(gezipptesJSONL(t, `{"object":"card","id":"a","name":"Eins"}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(server.Close)

	klient := scryfall.NeuerKlient(server.URL, netz.NeuerKlient(netz.UeblicheFristen()),
		nil, time.Now, "prüfung")

	ziel := filepath.Join(t.TempDir(), "oracle.jsonl.gz")

	eintrag := scryfall.Bulkeintrag{Adresse: server.URL + "/x.jsonl.gz"}

	if fehler := klient.BulkHolen(t.Context(), eintrag, ziel); fehler != nil {
		t.Fatalf("erstes Holen: %v", fehler)
	}

	angaben, err := os.Stat(ziel)
	if err != nil {
		t.Fatalf("die Sammeldatei fehlt: %v", err)
	}

	// **Der Eintrag nennt Größe und Zeitpunkt**, und die Datei ist danach
	// entstanden — damit ist der Stand vorhanden.
	eintrag.Groesse = angaben.Size()
	eintrag.Stand = angaben.ModTime().Add(-time.Second).Format(time.RFC3339)

	if fehler := klient.BulkHolen(t.Context(), eintrag, ziel); fehler != nil {
		t.Fatalf("zweites Holen: %v", fehler)
	}

	if gefragt != 1 {
		t.Errorf("die Gegenstelle wurde %dmal gefragt — verlangt ist einmal", gefragt)
	}
}

// **Ohne Ladeabsprache lädt jeder selbst.**
//
// Sie ist eine Verbesserung und keine Bedingung: ein doppelter Ladevorgang ist
// besser als kein Ergebnis (§8).
func TestOhneLadeabspracheLaedtJederSelbst(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, err := schreiben.Write(gezipptesJSONL(t,
				`{"object":"card","id":"a","name":"Eins"}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(server.Close)

	ort := t.TempDir()
	ziel := filepath.Join(ort, "oracle.jsonl.gz")

	// **Ein Verzeichnis am Ort der Nachbardatei**: `dateisperre.Oeffnen`
	// scheitert.
	if err := os.Mkdir(ziel+scryfall.Ladeendung, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	klient := scryfall.NeuerKlient(server.URL, netz.NeuerKlient(netz.UeblicheFristen()),
		nil, time.Now, "prüfung")

	if fehler := klient.BulkHolen(t.Context(), scryfall.Bulkeintrag{
		Adresse: server.URL + "/x.jsonl.gz",
	}, ziel); fehler != nil {
		t.Fatalf("holen ohne Ladeabsprache: %v", fehler)
	}

	if _, err := os.Stat(ziel); err != nil {
		t.Errorf("die Sammeldatei fehlt: %v", err)
	}
}

// **Und wenn ein anderer sie zu lange hält, ebenso.**
func TestNachDerLadefristLaedtJederSelbst(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, err := schreiben.Write(gezipptesJSONL(t,
				`{"object":"card","id":"a","name":"Eins"}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(server.Close)

	ziel := filepath.Join(t.TempDir(), "oracle.jsonl.gz")

	griff, err := dateisperre.Oeffnen(ziel + scryfall.Ladeendung)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := griff.Allein(); fehler != nil {
		t.Fatalf("allein nehmen: %v", fehler)
	}

	t.Cleanup(func() { _ = griff.Schliessen() }) //nolint:errcheck // Aufräumen

	alteFrist, alterTakt := netz.Ladefrist, netz.Ladetakt
	netz.Ladefrist, netz.Ladetakt = 40*time.Millisecond, 10*time.Millisecond

	t.Cleanup(func() { netz.Ladefrist, netz.Ladetakt = alteFrist, alterTakt })

	klient := scryfall.NeuerKlient(server.URL, netz.NeuerKlient(netz.UeblicheFristen()),
		nil, time.Now, "prüfung")

	if fehler := klient.BulkHolen(t.Context(), scryfall.Bulkeintrag{
		Adresse: server.URL + "/x.jsonl.gz",
	}, ziel); fehler != nil {
		t.Fatalf("holen nach der Frist: %v", fehler)
	}

	if _, statfehler := os.Stat(ziel); statfehler != nil {
		t.Errorf("die Sammeldatei fehlt: %v", statfehler)
	}
}

// **Ein anderer Stand gleicher Größe wird geholt.**
//
// # Der Befund des Nachlaufs
//
// `schonDa` verglich nur die **Größe** — und übersah den Zeitpunkt, den der
// Eintrag längst führt. Zwei Fassungen gleicher gepackter Größe wären dauerhaft
// verwechselt worden: der Bulk-Stand bliebe alt, ohne dass jemand es merkt.
func TestEinNeuererStandGleicherGroesseWirdGeholt(t *testing.T) {
	t.Parallel()

	var gefragt int

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt++

			if _, err := schreiben.Write(gezipptesJSONL(t,
				`{"object":"card","id":"a","name":"Eins"}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(server.Close)

	klient := scryfall.NeuerKlient(server.URL, netz.NeuerKlient(netz.UeblicheFristen()),
		nil, time.Now, "prüfung")

	ziel := filepath.Join(t.TempDir(), "oracle.jsonl.gz")
	eintrag := scryfall.Bulkeintrag{Adresse: server.URL + "/x.jsonl.gz"}

	if fehler := klient.BulkHolen(t.Context(), eintrag, ziel); fehler != nil {
		t.Fatalf("erstes Holen: %v", fehler)
	}

	angaben, err := os.Stat(ziel)
	if err != nil {
		t.Fatalf("die Sammeldatei fehlt: %v", err)
	}

	// **Gleiche Größe, aber ein Stand von morgen.** Die Datei ist damit älter
	// als das, was der Eintrag ansagt.
	eintrag.Groesse = angaben.Size()
	eintrag.Stand = angaben.ModTime().Add(time.Hour).Format(time.RFC3339)

	if fehler := klient.BulkHolen(t.Context(), eintrag, ziel); fehler != nil {
		t.Fatalf("zweites Holen: %v", fehler)
	}

	if gefragt != 2 {
		t.Errorf("die Gegenstelle wurde %dmal gefragt — der neuere Stand wurde "+
			"für den alten gehalten", gefragt)
	}
}

// **Ein unlesbarer Zeitpunkt heißt: nicht da.**
//
// Lieber einmal zu viel geladen als einen alten Stand für den neuen gehalten.
func TestEinUnlesbarerStandGiltAlsNichtDa(t *testing.T) {
	t.Parallel()

	var gefragt int

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt++

			if _, err := schreiben.Write(gezipptesJSONL(t,
				`{"object":"card","id":"a","name":"Eins"}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(server.Close)

	klient := scryfall.NeuerKlient(server.URL, netz.NeuerKlient(netz.UeblicheFristen()),
		nil, time.Now, "prüfung")

	ziel := filepath.Join(t.TempDir(), "oracle.jsonl.gz")
	eintrag := scryfall.Bulkeintrag{Adresse: server.URL + "/x.jsonl.gz"}

	if fehler := klient.BulkHolen(t.Context(), eintrag, ziel); fehler != nil {
		t.Fatalf("erstes Holen: %v", fehler)
	}

	angaben, err := os.Stat(ziel)
	if err != nil {
		t.Fatalf("die Sammeldatei fehlt: %v", err)
	}

	eintrag.Groesse = angaben.Size()
	eintrag.Stand = "morgen"

	if fehler := klient.BulkHolen(t.Context(), eintrag, ziel); fehler != nil {
		t.Fatalf("zweites Holen: %v", fehler)
	}

	if gefragt != 2 {
		t.Errorf("die Gegenstelle wurde %dmal gefragt — ein unlesbarer Stand "+
			"galt als vorhanden", gefragt)
	}
}
