package app_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// sammelprobe ist ein Doppel für [app.Sammelquelle].
type sammelprobe struct {
	eintraege []app.Sammeleintrag
	fehler    error
	holfehler error
	geholt    string
}

func (s *sammelprobe) Verzeichnis(context.Context) ([]app.Sammeleintrag, error) {
	return s.eintraege, s.fehler
}

func (s *sammelprobe) Holen(_ context.Context, eintrag app.Sammeleintrag, ziel string) error {
	s.geholt = eintrag.Art
	_ = ziel

	return s.holfehler
}

// **Der Stand sagt auch ohne Netz, was hier liegt** — §16, Regel 5.
func TestSammeldatenstand(t *testing.T) {
	t.Run("nennt die angebotene Größe, wenn Scryfall antwortet", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		naht := app.SammeldatenUeber(t, &sammelprobe{eintraege: []app.Sammeleintrag{
			{Art: "oracle_cards", Bytes: 24531600, Stand: "2026-08-23T21:01:56+00:00"},
			{Art: "all_cards", Bytes: 391700000},
		}}, mussBulkpfad(t))

		stand, fehler := naht.Stand(probekontext(t))
		if fehler != nil {
			t.Fatalf("Stand: %v", fehler)
		}

		if stand.Art != "oracle_cards" || stand.Angeboten != 24531600 {
			t.Errorf("Stand = %+v", stand)
		}

		if stand.Vorhanden {
			t.Error("es liegt noch keine Datei, und der Stand sagt das Gegenteil")
		}
	})

	t.Run("sagt auch ohne Verzeichnis, was liegt", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		naht := app.SammeldatenUeber(t, &sammelprobe{fehler: errors.New("kein Netz")}, mussBulkpfad(t))

		stand, fehler := naht.Stand(probekontext(t))
		if fehler != nil {
			t.Fatalf("Stand: %v — der Ausfall darf die Auskunft nicht anhalten", fehler)
		}

		if stand.Angeboten != 0 || stand.Art != "oracle_cards" {
			t.Errorf("Stand = %+v", stand)
		}
	})

	// **Ohne Netz sagt der Stand, wie alt die liegende Datei ist** (F04,
	// #2246; §16, Satz 5) — bis dahin sagte er nur, dass sie liegt.
	t.Run("nennt ohne Netz, wann die liegende Datei geholt wurde", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		pfad := mussBulkpfad(t)
		if fehler := os.MkdirAll(filepath.Dir(pfad), 0o700); fehler != nil {
			t.Fatal(fehler)
		}

		if fehler := os.WriteFile(pfad, []byte("[]"), 0o600); fehler != nil {
			t.Fatal(fehler)
		}

		geholt := time.Date(2026, time.October, 1, 8, 0, 0, 0, time.UTC)
		if fehler := os.Chtimes(pfad, geholt, geholt); fehler != nil {
			t.Fatal(fehler)
		}

		naht := app.SammeldatenUeber(t, &sammelprobe{fehler: errors.New("kein Netz")}, pfad)

		stand, fehler := naht.Stand(probekontext(t))
		if fehler != nil {
			t.Fatalf("Stand: %v", fehler)
		}

		if !stand.Vorhanden || stand.Geholt != "2026-10-01T08:00:00Z" {
			t.Errorf("Stand = %+v — erwartet vorhanden, geholt 2026-10-01T08:00:00Z", stand)
		}
	})

	t.Run("holt die Datei und meldet danach den Stand", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		probe := &sammelprobe{eintraege: []app.Sammeleintrag{{Art: "oracle_cards"}}}

		if _, fehler := app.SammeldatenUeber(t, probe, mussBulkpfad(t)).Holen(probekontext(t)); fehler != nil {
			t.Fatalf("Holen: %v", fehler)
		}

		if probe.geholt != "oracle_cards" {
			t.Errorf("geholt wurde %q", probe.geholt)
		}
	})

	t.Run("meldet einen Ausfall des Verzeichnisses beim Holen", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		probe := &sammelprobe{fehler: errors.New("kein Netz")}

		if _, fehler := app.SammeldatenUeber(t, probe, mussBulkpfad(t)).Holen(probekontext(t)); fehler == nil {
			t.Fatal("der Ausfall kam nicht durch")
		}
	})

	t.Run("meldet einen Ausfall beim Laden", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		probe := &sammelprobe{
			eintraege: []app.Sammeleintrag{{Art: "oracle_cards"}},
			holfehler: errors.New("abgebrochen"),
		}

		if _, fehler := app.SammeldatenUeber(t, probe, mussBulkpfad(t)).Holen(probekontext(t)); fehler == nil {
			t.Fatal("der Ausfall kam nicht durch")
		}
	})

	t.Run("meldet, wenn die gesuchte Art nicht angeboten wird", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		probe := &sammelprobe{eintraege: []app.Sammeleintrag{{Art: "all_cards"}}}

		if _, fehler := app.SammeldatenUeber(t, probe, mussBulkpfad(t)).Holen(probekontext(t)); !errors.Is(
			fehler, domain.ErrNotFound) {
			t.Fatalf("Fehler = %v", fehler)
		}
	})
}

// **Und die Fassade übersetzt beides in Megabyte.**
func TestScryfallSammeldatenAnDerFassade(t *testing.T) {
	t.Run("ein Ausfall kommt übersetzt durch", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		vorher := app.MitScryfalladresse("http://unerreichbar.invalid")
		t.Cleanup(func() { app.MitScryfalladresse(vorher) })

		if antwort := app.NeueApp("").ScryfallSammeldatenHolen(
			probekontext(t)); antwort.Fehler == nil {
			t.Error("der Ausfall kam nicht durch")
		}
	})

	// **Und die Bindung reicht die Abrufzeit durch** (F04, #2246) — ohne
	// Netz, wo sie die einzige Altersangabe ist.
	t.Run("nennt ohne Netz, wann die liegende Datei geholt wurde", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		vorher := app.MitScryfalladresse("http://unerreichbar.invalid")
		t.Cleanup(func() { app.MitScryfalladresse(vorher) })

		pfad := mussBulkpfad(t)
		if fehler := os.MkdirAll(filepath.Dir(pfad), 0o700); fehler != nil {
			t.Fatal(fehler)
		}

		if fehler := os.WriteFile(pfad, []byte("[]"), 0o600); fehler != nil {
			t.Fatal(fehler)
		}

		geholt := time.Date(2026, time.October, 1, 8, 0, 0, 0, time.UTC)
		if fehler := os.Chtimes(pfad, geholt, geholt); fehler != nil {
			t.Fatal(fehler)
		}

		antwort := app.NeueApp("").ScryfallSammeldaten(probekontext(t))
		if antwort.Fehler != nil || !antwort.Vorhanden || antwort.Geholt != "2026-10-01T08:00:00Z" {
			t.Errorf("Antwort = %+v — erwartet vorhanden, geholt 2026-10-01T08:00:00Z", antwort)
		}
	})

	t.Run("ohne Zuhause kommt ein Fehler", func(t *testing.T) {
		t.Setenv("HOME", "")
		t.Setenv("MSE_BESTAND", "")

		anwendung := app.NeueApp("")

		if antwort := anwendung.ScryfallSammeldaten(t.Context()); antwort.Fehler == nil {
			t.Error("ohne Zuhause blieb es ohne Fehler")
		}

		if antwort := anwendung.ScryfallSammeldatenHolen(t.Context()); antwort.Fehler == nil {
			t.Error("ohne Zuhause blieb das Holen ohne Fehler")
		}
	})
}

// **Der ganze Weg: Naht, Klient, Gegenstelle, Platte.**
//
// # Warum das ein eigener Lauf ist
//
// Die Läufe darüber gehen über ein Doppel und prüfen die Fassade. Keiner
// zeigt, dass [app.Sammelquellenaht] den Klienten richtig zusammensetzt — und
// genau dort saß in Phase 5 der Fehler, den kein Test sah (#130).
func TestDerGanzeSammeldatenwegGehtBisZurPlatte(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	inhalt := einePackung(t,
		`{"object":"card","id":"a","name":"Prüfstein","artist":"Wer auch immer"}`)

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			if strings.HasSuffix(anfrage.URL.Path, ".jsonl.gz") {
				sende(t, schreiben, inhalt)

				return
			}

			schreiben.Header().Set("Content-Type", "application/json")
			sende(t, schreiben, []byte(`{"object":"list","data":[
				{"object":"bulk_data","type":"oracle_cards","name":"Oracle Cards",
				 "jsonl_download_uri":"http://`+anfrage.Host+`/oracle.jsonl.gz",
				 "compressed_size":123,"updated_at":"2026-08-23T21:01:56+00:00"}]}`))
		}))
	t.Cleanup(gegenstelle.Close)

	vorher := app.MitScryfalladresse(gegenstelle.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	naht, fehler := app.Sammeldatennaht(app.ImmerAn{})
	if fehler != nil {
		t.Fatalf("Sammeldatennaht: %v", fehler)
	}

	stand, holfehler := naht.Holen(probekontext(t))
	if holfehler != nil {
		t.Fatalf("Holen: %v", holfehler)
	}

	if !stand.Vorhanden || stand.Bytes == 0 {
		t.Errorf("Stand = %+v", stand)
	}

	// **Und der Rückfall findet die Karte darin** — das ist der Punkt des
	// ganzen Werkstücks: ohne Netz kommt die Auskunft von der Platte.
	karte, _, lesefehler := app.Bulkrueckfall().KarteNachName(probekontext(t), "Prüfstein")
	if lesefehler != nil {
		t.Fatalf("KarteNachName: %v", lesefehler)
	}

	if karte.Name != "Prüfstein" {
		t.Errorf("Karte = %+v", karte)
	}
}

// **Ein Ausfall der Gegenstelle kommt über die Naht durch.**
func TestEinAusfallKommtUeberDieSammeldatennahtDurch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusInternalServerError)
		}))
	t.Cleanup(gegenstelle.Close)

	vorher := app.MitScryfalladresse(gegenstelle.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	naht, fehler := app.Sammeldatennaht(app.ImmerAn{})
	if fehler != nil {
		t.Fatalf("Sammeldatennaht: %v", fehler)
	}

	if _, holfehler := naht.Holen(probekontext(t)); holfehler == nil {
		t.Error("der Ausfall kam nicht durch")
	}

	// **Der Stand kommt trotzdem** — er liest zuerst die Platte.
	if _, standfehler := naht.Stand(probekontext(t)); standfehler != nil {
		t.Errorf("Stand: %v — der Ausfall darf ihn nicht anhalten", standfehler)
	}
}

// **Ohne Zuhause kommt keine Naht.**
func TestOhneZuhauseKommtKeineSammeldatennaht(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("MSE_BESTAND", "")

	if _, fehler := app.Sammeldatennaht(app.ImmerAn{}); fehler == nil {
		t.Error("ohne Zuhause kam eine Naht")
	}

	if _, fehler := app.Sammelquellenaht(app.ImmerAn{}); fehler == nil {
		t.Error("ohne Zuhause kam eine Quelle")
	}
}

// mussBulkpfad gibt den Ort der Sammeldatei oder bricht ab.
func mussBulkpfad(t *testing.T) string {
	t.Helper()

	pfad, fehler := app.Bulkpfad()
	if fehler != nil {
		t.Fatalf("Bulkpfad: %v", fehler)
	}

	return pfad
}
