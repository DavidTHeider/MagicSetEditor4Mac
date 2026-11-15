package app

// **Die Naht zum Suchindex** — Werkstück A15 von Phase 5.3
// (#1514).
//
// **Ein Lauf im Paket**, weil Öffner und Einbetterquelle unexportiert sind:
// sie sind Verdrahtung und keine Fähigkeit, und sie zu exportieren, nur damit
// ein Test sie sieht, machte aus einer inneren Naht eine Fläche.

import (
	"compress/gzip"
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einAufbauauftrag nennt eine Sammeldatei.
func einAufbauauftrag(pfad string) service.Aufbauauftrag {
	return service.Aufbauauftrag{
		Modell: "all-minilm", Quelle: pfad, Revision: "1",
		Abgerufen: time.Now(),
	}
}

// **Eine Sammeldatei, die es nicht gibt, geht nicht auf.**
//
// **Der Öffner ist die Naht zu `internal/scryfall`** — ohne diesen Lauf
// stünde er da, ohne je gerufen worden zu sein.
func TestEineFehlendeSammeldateiGehtNichtAuf(t *testing.T) {
	t.Parallel()

	oeffner := sammeldateioeffner{}

	if _, _, fehler := oeffner.Oeffnen(t.Context(), einAufbauauftrag(
		filepath.Join(t.TempDir(), "gibtesnicht.jsonl.gz"))); fehler == nil {
		t.Fatal("die fehlende Sammeldatei ging auf")
	}
}

// **Eine Datei, die kein gzip ist, ebenso — mit einem Satz, der es sagt.**
func TestEineSammeldateiOhneGzipGehtNichtAuf(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "keingzip.jsonl.gz")

	if fehler := os.WriteFile(pfad, []byte("das ist kein gzip"),
		0o600); fehler != nil {
		t.Fatalf("die Probe schreiben: %v", fehler)
	}

	_, _, fehler := sammeldateioeffner{}.Oeffnen(t.Context(),
		einAufbauauftrag(pfad))
	if fehler == nil || !strings.Contains(fehler.Error(), "gzip") {
		t.Fatalf("die Probe ging auf: %v", fehler)
	}
}

// **Ohne Einstellungen gibt es keine Adresse — und keinen Sturz.**
//
// **Der Klient nimmt dann die Vorgabeadresse auf diesem Rechner**, und das
// ist richtig: wer nichts eingestellt hat, meint das lokale Ollama.
func TestOhneEinstellungenGibtEsKeineKIAdresse(t *testing.T) {
	t.Parallel()

	if adresse := NeueApp("").kiadresse(t.Context()); adresse != "" {
		t.Errorf("die Adresse ist %q", adresse)
	}
}

// **Ein unlesbarer Einstellungsstand gibt ebenfalls keine Adresse.**
func TestEinUnlesbarerStandGibtKeineKIAdresse(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(unwilligeEinstellungen{})

	if adresse := anwendung.kiadresse(t.Context()); adresse != "" {
		t.Errorf("die Adresse ist %q", adresse)
	}
}

// **Die Einbetterquelle reicht den Befund durch, statt zu stürzen.**
//
// **Ohne erreichbare Laufzeit gibt es keinen Raum** — und das ist ein
// Fehler und kein Befund: es kam gar nichts (§10).
func TestDieEinbetterquelleReichtDenFehlerDurch(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")

	quelle := einbetterquelle{bauen: anwendung.retrievallaufzeit}

	if _, _, fehler := quelle.Einbetter(t.Context(),
		"gibtesnicht"); fehler == nil {
		t.Fatal("die unerreichbare Laufzeit kam durch")
	}
}

// einProbebestand öffnet einen frischen Bestand auf der Platte.
//
// **Eine eigene Vorrichtung**: die in `scryfallnaht_test.go` liegt im
// Testpaket `app_test` und ist von hier nicht erreichbar.
func einProbebestand(t *testing.T) *sql.DB {
	t.Helper()

	db, fehler := storage.Oeffnen(t.Context(),
		filepath.Join(t.TempDir(), "bestand.db"))
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // der Test endet ohnehin

	if fehler := storage.Migriere(t.Context(), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	return db
}

// imProbeprofil trägt das erste Profil — das, das die Wanderung 0034 anlegt.
func imProbeprofil(t *testing.T) context.Context {
	t.Helper()

	return storage.MitProfil(t.Context(), 1)
}

// unwilligeEinstellungen antworten auf nichts.
type unwilligeEinstellungen struct{ Einstellungsfuehrer }

func (unwilligeEinstellungen) Stand(context.Context) (
	map[string]service.Wahl, error,
) {
	return nil, os.ErrClosed
}

// **Eine Sammeldatei, die aufgeht, gibt einen Korpus und einen Schließer.**
//
// **Die Gegenprobe zu den beiden Läufen darüber**: sie zeigen, dass etwas
// scheitert; ohne diesen zeigte keiner, dass überhaupt etwas gelingt.
func TestEineGueltigeSammeldateiGehtAuf(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "probe.jsonl.gz")

	datei, fehler := os.Create(pfad) //nolint:gosec // ein Pfad aus diesem Test
	if fehler != nil {
		t.Fatalf("die Probe anlegen: %v", fehler)
	}

	packer := gzip.NewWriter(datei)

	if _, geschrieben := packer.Write([]byte(
		`{"object":"card","id":"a","oracle_id":"o","name":"Probe","lang":"en",` +
			`"layout":"normal","type_line":"Instant","cmc":1,"set":"tst",` +
			`"set_name":"Test","collector_number":"1","rarity":"common",` +
			"\n")); geschrieben != nil {
		t.Fatalf("die Probe schreiben: %v", geschrieben)
	}

	if zu := errors.Join(packer.Close(), datei.Close()); zu != nil {
		t.Fatalf("die Probe schließen: %v", zu)
	}

	quelle, schliessen, fehler := sammeldateioeffner{}.Oeffnen(t.Context(),
		einAufbauauftrag(pfad))
	if fehler != nil {
		t.Fatalf("Oeffnen: %v", fehler)
	}

	defer schliessen()

	staende, fehler := quelle.Staende(t.Context())
	if fehler != nil || len(staende) != 1 {
		t.Fatalf("die Stände sind %+v (%v)", staende, fehler)
	}
}

// **Die Anwendung für die Flächen ohne Fenster kennt den Index.**
//
// **Und zwar über dieselben Wege wie das Fenster** (§16): ohne Einstellungen
// zeigte der Aufbau hier auf einen anderen Port als dort.
func TestDieRetrievalanwendungKenntDenIndex(t *testing.T) {
	t.Parallel()

	stand := RetrievalAnwendungUeber(einProbebestand(t)).Indexstand(t.Context())
	if stand.Fehler != nil {
		t.Fatalf("Indexstand: %+v", stand.Fehler)
	}

	// **Kein Index ist eine Auskunft und kein Fehler.**
	if stand.Aktiv || stand.Grund == "" {
		t.Fatalf("der Stand ist %+v", stand)
	}
}

// **Und das Fenster bekommt sie über die Verdrahtung.**
func TestDieVerdrahtungHaengtDenIndexAnsFenster(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = einProbebestand(t)
	anwendung.retrievalVerdrahten()

	if anwendung.retrieval == nil {
		t.Fatal("der Suchindex hängt nicht am Fenster")
	}

	if stand := anwendung.Indexstand(t.Context()); stand.Aktiv {
		t.Errorf("der Stand ist %+v", stand)
	}
}

// **Ein Aufbau am verdrahteten Fenster nimmt die Sperre neben dem Bestand.**
//
// # Warum dieser Lauf hier steht
//
// **Er ist der einzige, der die Verdrahtung wirklich durchläuft.** Der Riegel
// wird **vor** dem Modell genommen ([service.Aufbaustoff]); damit kommt der
// Weg über den Bestandspfad auch dann an, wenn keine Laufzeit antwortet — und
// das ist der Normalfall auf einem Rechner ohne Ollama.
//
// **Zu sehen ist es an der Sperrdatei**: sie liegt neben dem Bestand und
// bleibt liegen (so wie `bulk.go` es mit seiner hält — eine Sperre auf einer
// Datei, die verschwindet, ist keine).
func TestEinAufbauNimmtDieSperreNebenDemBestand(t *testing.T) {
	t.Parallel()

	ort := filepath.Join(t.TempDir(), "bestand.db")

	db, fehler := storage.Oeffnen(t.Context(), ort)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // der Test endet ohnehin

	if fehler := storage.Migriere(t.Context(), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	anwendung := NeueApp(ort)
	anwendung.db = db
	anwendung.retrievalVerdrahten()

	// **Er scheitert an der fehlenden Laufzeit** — und das ist in Ordnung:
	// geprüft wird der Weg **davor**.
	if antwort := anwendung.IndexAufbauen(t.Context(), "gibtesnicht",
		"/gibt/es/nicht.jsonl.gz", "1", "", false); antwort.Fehler == nil {
		t.Fatal("der Aufbau kam ohne Laufzeit durch")
	}

	if _, fehler := os.Stat(ort + Indexbauendung); fehler != nil {
		t.Fatalf("die Sperrdatei liegt nicht neben dem Bestand: %v", fehler)
	}
}

// **Und über die Fläche ohne Fenster geht derselbe Weg.**
func TestEinAufbauUeberDieFlaecheNimmtDenselbenWeg(t *testing.T) {
	t.Parallel()

	if antwort := RetrievalAnwendungUeber(einProbebestand(t)).IndexAufbauen(
		t.Context(), "gibtesnicht", "/gibt/es/nicht.jsonl.gz", "1", "",
		false); antwort.Fehler == nil {
		t.Fatal("der Aufbau kam ohne Laufzeit durch")
	}
}
