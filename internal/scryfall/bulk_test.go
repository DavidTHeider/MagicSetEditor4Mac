// **Bulk-Daten für den Betrieb ohne Netz** —
// Abnahmekriterium aus #888.
package scryfall_test

import (
	"bytes"
	"compress/gzip"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

func TestBulkverzeichnis(t *testing.T) {
	t.Parallel()

	t.Run("liest die Einträge", func(t *testing.T) {
		t.Parallel()

		server, _ := gegenstelleMit(t, string(beispiel(t, "bulkverzeichnis.json")))
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		eintraege, _, fehler := klient.Bulkverzeichnis(t.Context())
		if fehler != nil {
			t.Fatalf("Bulkverzeichnis: %v", fehler)
		}

		if len(eintraege) != 2 {
			t.Fatalf("%d Einträge", len(eintraege))
		}

		erster := eintraege[0]

		// **Die Adresse heißt `jsonl_download_uri`, nicht `download_uri`.**
		//
		// Die Dokumentation nennt beide; die **echte** Antwort führte am
		// 2026-08-24 nur die erste. Wer die andere läse, bekäme eine leere
		// Adresse und einen Abruf, der ins Nichts geht.
		if !strings.HasSuffix(erster.Adresse, ".jsonl.gz") {
			t.Errorf("Adresse = %q", erster.Adresse)
		}

		if erster.Art != "oracle_cards" || erster.Groesse != 24531600 {
			t.Errorf("Eintrag = %+v", erster)
		}

		// **Der Zeitpunkt bleibt eine Zeichenkette**, wie er kommt — mit Zone.
		if !strings.HasPrefix(erster.Stand, "2026-08-23T21:01:56") {
			t.Errorf("Stand = %q", erster.Stand)
		}
	})

	t.Run("weist einen Eintrag ohne Adresse ab", func(t *testing.T) {
		t.Parallel()

		server, _ := gegenstelleMit(t, `{"object":"list","data":[
			{"object":"bulk_data","type":"oracle_cards","name":"Ohne Adresse"}]}`)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		if _, _, fehler := klient.Bulkverzeichnis(t.Context()); !errors.Is(
			fehler, domain.ErrValidation) {
			t.Fatalf("Fehler = %v, erwartet ErrValidation", fehler)
		}
	})

	t.Run("erträgt ein unbekanntes Zusatzfeld", func(t *testing.T) {
		t.Parallel()

		server, _ := gegenstelleMit(t, `{"object":"list","data":[
			{"object":"bulk_data","type":"oracle_cards","name":"Neu",
			 "jsonl_download_uri":"https://beispiel.invalid/x.jsonl.gz",
			 "compressed_size":1,"voellig_neu":{"tief":1}}]}`)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		if _, _, fehler := klient.Bulkverzeichnis(t.Context()); fehler != nil {
			t.Fatalf("das neue Feld hat gebrochen: %v", fehler)
		}
	})
}

// gezipptesJSONL packt Zeilen als `.jsonl.gz`.
func gezipptesJSONL(t *testing.T, zeilen ...string) []byte {
	t.Helper()

	var gepackt bytes.Buffer

	schreiber := gzip.NewWriter(&gepackt)
	for _, zeile := range zeilen {
		if _, fehler := schreiber.Write([]byte(zeile + "\n")); fehler != nil {
			t.Fatalf("packen: %v", fehler)
		}
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	return gepackt.Bytes()
}

func TestBulkHolen(t *testing.T) {
	t.Parallel()

	inhalt := gezipptesJSONL(t,
		`{"object":"card","id":"a","name":"Lightning Bolt","artist":"Christopher Rush"}`,
		`{"object":"card","id":"b","name":"Counterspell"}`)

	t.Run("legt die Datei ab", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(
			func(schreiben http.ResponseWriter, anfrage *http.Request) {
				if anfrage.Header.Get("User-Agent") == "" {
					t.Error("die Datei wurde ohne User-Agent geholt")
				}

				if _, fehler := schreiben.Write(inhalt); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			}))
		t.Cleanup(server.Close)

		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))
		ziel := filepath.Join(t.TempDir(), "oracle.jsonl.gz")

		if fehler := klient.BulkHolen(t.Context(), scryfall.Bulkeintrag{
			Art: "oracle_cards", Adresse: server.URL + "/oracle.jsonl.gz",
		}, ziel); fehler != nil {
			t.Fatalf("BulkHolen: %v", fehler)
		}

		abgelegt, lesefehler := os.ReadFile(ziel) //nolint:gosec // Pfad aus t.TempDir()
		if lesefehler != nil {
			t.Fatalf("die Datei fehlt: %v", lesefehler)
		}

		if !bytes.Equal(abgelegt, inhalt) {
			t.Error("die Bytes sind verändert angekommen")
		}
	})

	// **Nach dem Ablegen bleibt keine Nebendatei liegen.**
	//
	// # Was dieser Lauf hält — und was nicht
	//
	// Er hält, dass am Ende **genau eine** Datei unter dem gewünschten Namen
	// liegt. Er fände eine vergessene Umbenennung: dann stünde dort
	// `oracle.jsonl.gz.teil`.
	//
	// **Er hält nicht, dass es zwei Schritte sind.** Die Gegenprobe
	// `neben := ziel` blieb hier grün — `os.Rename(ziel, ziel)` glückt. Das
	// belegt der Lauf über ein Ziel, das ein **Verzeichnis** ist: dort
	// verschiebt dieselbe Änderung den Fehler vom Umbenennen ins Ablegen.
	t.Run("hinterlässt keine Nebendatei", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(
			func(schreiben http.ResponseWriter, _ *http.Request) {
				if _, fehler := schreiben.Write(inhalt); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			}))
		t.Cleanup(server.Close)

		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))
		ordner := t.TempDir()

		if fehler := klient.BulkHolen(t.Context(), scryfall.Bulkeintrag{
			Adresse: server.URL + "/x.jsonl.gz",
		}, filepath.Join(ordner, "oracle.jsonl.gz")); fehler != nil {
			t.Fatalf("BulkHolen: %v", fehler)
		}

		eintraege, lesefehler := os.ReadDir(ordner)
		if lesefehler != nil {
			t.Fatalf("den Ordner lesen: %v", lesefehler)
		}

		// **Die Ladeabsprache zählt nicht mit** (#1163). Sie bleibt liegen:
		// eine Sperrdatei zu löschen, während ein anderer sie hält, gäbe
		// jedem einen eigenen Inode.
		namen := ohneLadeabsprache(eintraege)
		if len(namen) != 1 || namen[0] != "oracle.jsonl.gz" {
			t.Errorf("im Ordner liegt %v, erwartet nur die fertige Datei", namen)
		}
	})

	t.Run("meldet einen Ausfall und schreibt gar nichts", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(
			func(schreiben http.ResponseWriter, _ *http.Request) {
				schreiben.WriteHeader(http.StatusInternalServerError)
			}))
		t.Cleanup(server.Close)

		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))
		ordner := t.TempDir()
		ziel := filepath.Join(ordner, "oracle.jsonl.gz")

		if fehler := klient.BulkHolen(t.Context(), scryfall.Bulkeintrag{
			Adresse: server.URL + "/x.jsonl.gz",
		}, ziel); fehler == nil {
			t.Fatal("der Ausfall kam nicht durch")
		}

		// **Gar nichts** — auch keine Nebendatei. Der Abruf scheitert, bevor
		// geschrieben wird; was hier stünde, käme aus einem früheren Lauf.
		eintraege, lesefehler := os.ReadDir(ordner)
		if lesefehler != nil {
			t.Fatalf("den Ordner lesen: %v", lesefehler)
		}

		if namen := ohneLadeabsprache(eintraege); len(namen) != 0 {
			t.Errorf("es liegen %d Dateien im Ordner: %v", len(namen), namen)
		}
	})

	t.Run("weist eine leere Adresse ab", func(t *testing.T) {
		t.Parallel()

		klient := klientFuer(t, "http://unerreichbar.invalid",
			netz.NeuerKlient(netz.UeblicheFristen()))

		if fehler := klient.BulkHolen(t.Context(), scryfall.Bulkeintrag{},
			filepath.Join(t.TempDir(), "x")); !errors.Is(fehler, domain.ErrValidation) {
			t.Fatalf("Fehler = %v", fehler)
		}
	})
}

// **Der Leser findet eine Karte ohne Netz.**
//
// # Warum zeilenweise und nicht mit einem `json.Unmarshal`
//
// Die Datei `all_cards` wiegt **392 MB gepackt** (gemessen am 2026-08-24 am
// Verzeichnis von Scryfall). Sie in den Arbeitsspeicher zu laden, um eine
// Karte zu finden, wäre auf einem Rechner mit 8 GB ein Ausfall — und die
// Antwort steht meist in den ersten Zeilen.
//
// **JSONL macht das möglich**: eine Karte je Zeile, kein umschließendes Array.
func TestBulkleser(t *testing.T) {
	t.Parallel()

	ort := filepath.Join(t.TempDir(), "oracle.jsonl.gz")
	if fehler := os.WriteFile(ort, gezipptesJSONL(t,
		`{"object":"card","id":"a","name":"Lightning Bolt","artist":"Christopher Rush"}`,
		``,
		`{"object":"card","id":"b","name":"Counterspell","artist":"Mark Poole"}`,
	), 0o600); fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	t.Run("findet eine Karte nach Namen", func(t *testing.T) {
		t.Parallel()

		karte, befunde, fehler := scryfall.BulkKarteNachName(ort, "counterspell")
		if fehler != nil {
			t.Fatalf("BulkKarteNachName: %v", fehler)
		}

		if len(befunde) != 0 {
			t.Errorf("Befunde = %+v", befunde)
		}

		// **Ohne Rücksicht auf Groß- und Kleinschreibung** — wie
		// `/cards/named?exact=`, das ebenso sucht.
		if karte.Name != "Counterspell" || karte.Kuenstler != "Mark Poole" {
			t.Errorf("Karte = %+v", karte)
		}
	})

	t.Run("meldet eine Karte, die nicht darin steht", func(t *testing.T) {
		t.Parallel()

		if _, _, fehler := scryfall.BulkKarteNachName(ort, "Gibtesnicht"); !errors.Is(
			fehler, domain.ErrNotFound) {
			t.Fatalf("Fehler = %v, erwartet ErrNotFound", fehler)
		}
	})

	t.Run("meldet eine Datei, die es nicht gibt", func(t *testing.T) {
		t.Parallel()

		if _, _, fehler := scryfall.BulkKarteNachName(
			filepath.Join(t.TempDir(), "fehlt.jsonl.gz"), "X"); fehler == nil {
			t.Fatal("die fehlende Datei blieb ohne Fehler")
		}
	})

	t.Run("meldet eine Datei, die kein gzip ist", func(t *testing.T) {
		t.Parallel()

		falsch := filepath.Join(t.TempDir(), "kaputt.jsonl.gz")
		if fehler := os.WriteFile(falsch, []byte("kein gzip"), 0o600); fehler != nil {
			t.Fatalf("ablegen: %v", fehler)
		}

		if _, _, fehler := scryfall.BulkKarteNachName(falsch, "X"); fehler == nil {
			t.Fatal("die falsche Datei blieb ohne Fehler")
		}
	})

	t.Run("überspringt eine unlesbare Zeile mit Befund", func(t *testing.T) {
		t.Parallel()

		// **§8: 300 lesbare Karten und zwei unlesbare ergeben 300 Karten.**
		// Eine kaputte Zeile in einer 400-MB-Datei darf nicht die ganze Datei
		// unbrauchbar machen.
		gemischt := filepath.Join(t.TempDir(), "gemischt.jsonl.gz")
		if fehler := os.WriteFile(gemischt, gezipptesJSONL(t,
			`{kein json`,
			`{"object":"card","id":"b","name":"Counterspell"}`,
		), 0o600); fehler != nil {
			t.Fatalf("ablegen: %v", fehler)
		}

		karte, befunde, fehler := scryfall.BulkKarteNachName(gemischt, "Counterspell")
		if fehler != nil {
			t.Fatalf("BulkKarteNachName: %v", fehler)
		}

		if karte.Name != "Counterspell" {
			t.Errorf("Karte = %+v", karte)
		}

		// **Die Stelle steht in `Ort`, nicht in der Meldung** — `String()`
		// setzt beide zusammen, und so liest sie ein Mensch.
		if len(befunde) != 1 || !strings.Contains(befunde[0].String(), "Zeile 1") {
			t.Errorf("Befunde = %+v", befunde)
		}
	})
}

// ohneLadeabsprache gibt die Namen ohne die Nachbardatei des Ladens.
func ohneLadeabsprache(eintraege []os.DirEntry) []string {
	namen := make([]string, 0, len(eintraege))

	for _, einer := range eintraege {
		if !strings.HasSuffix(einer.Name(), scryfall.Ladeendung) {
			namen = append(namen, einer.Name())
		}
	}

	return namen
}
