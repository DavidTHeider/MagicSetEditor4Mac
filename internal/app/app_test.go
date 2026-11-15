package app_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Der Wächter über die Verdrahtung.
//
// Bis zum Zweitblick am Ende von Phase 1 gab es sie nicht: der Migrator sagte
// „jeder Start migriert", und die einzige vollständige Folge aus Öffnen und
// Migrieren stand in der Testvorrichtung von `internal/storage`. `main.go`
// startete Wails, ohne sie je zu rufen — eine Zusage, die nur im Test eingelöst
// wird, ist keine.
func TestStartenLegtDieDatenbankAnUndMigriertSie(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "unterordner", "bestand.db")

	anwendung := app.NeueApp(pfad)

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("die Anwendung ließ sich nicht starten: %v", err)
	}
	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("Beenden: %v", fehler)
		}
	})

	if _, err := os.Stat(pfad); err != nil {
		t.Errorf("die Datenbank liegt nicht unter %s: %v", pfad, err)
	}
}

// Ein zweiter Start auf demselben Bestand darf nichts kaputtmachen — das ist
// der Normalfall, nicht die frische Datenbank.
func TestZweiterStartAufDemselbenBestand(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "bestand.db")

	for lauf := 1; lauf <= 2; lauf++ {
		anwendung := app.NeueApp(pfad)

		if err := anwendung.Starten(t.Context()); err != nil {
			t.Fatalf("Lauf %d schlug fehl: %v", lauf, err)
		}
		if err := anwendung.Beenden(); err != nil {
			t.Fatalf("Lauf %d ließ sich nicht beenden: %v", lauf, err)
		}
	}
}

func TestStartenMeldetEinUnbrauchbaresVerzeichnis(t *testing.T) {
	// Eine Datei dort, wo ein Verzeichnis entstehen müsste.
	sperre := filepath.Join(t.TempDir(), "sperre")
	if err := os.WriteFile(sperre, []byte("keine Datenbank"), 0o600); err != nil {
		t.Fatalf("Sperre nicht anlegbar: %v", err)
	}

	if err := app.NeueApp(filepath.Join(sperre, "bestand.db")).Starten(probekontext(t)); err == nil {
		t.Error("ein unbrauchbares Verzeichnis wurde angenommen")
	}
}

func TestBeendenOhneStartIstStill(t *testing.T) {
	if err := app.NeueApp("egal").Beenden(); err != nil {
		t.Errorf("Beenden ohne Start meldete %v", err)
	}
}

func TestBeendenZweimalIstStill(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := anwendung.Beenden(); err != nil {
		t.Fatalf("erstes Beenden: %v", err)
	}
	if err := anwendung.Beenden(); err != nil {
		t.Errorf("zweites Beenden meldete %v", err)
	}
}

func TestDatenbankpfadLiegtImBenutzerverzeichnis(t *testing.T) {
	pfad, err := app.Datenbankpfad()
	if err != nil {
		t.Fatalf("Datenbankpfad: %v", err)
	}

	if filepath.Base(pfad) != "bestand.db" {
		t.Errorf("Dateiname %q, erwartet war bestand.db", filepath.Base(pfad))
	}
	if filepath.Base(filepath.Dir(pfad)) != "MagicSetEditor4Mac" {
		t.Errorf("Verzeichnis %q, erwartet war MagicSetEditor4Mac", filepath.Dir(pfad))
	}
}

func TestDatenbankpfadMeldetEinFehlendesBenutzerverzeichnis(t *testing.T) {
	// os.UserConfigDir liest HOME bzw. XDG_CONFIG_HOME; ohne beides scheitert
	// es. Der Fehlerpfad gehört geprüft: er entscheidet, ob die Anwendung mit
	// einer Meldung endet oder mit einem leeren Pfad weiterläuft.
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	if _, err := app.Datenbankpfad(); err == nil {
		t.Error("ohne Benutzerverzeichnis kam ein Pfad zurück")
	}
}

// Ein gescheiterter Start lässt keine offene Datenbank zurück.
//
// Aus dem Nachlauf des Zweitblicks: `a.db` wurde gesetzt, bevor migriert wurde,
// und ein Fehler dabei kehrte unmittelbar zurück. In `main.lauf` entsteht das
// schließende `defer` erst nach einem geglückten Start — also blieb genau auf
// den neuen Fehlerpfaden eine offene Datenbank zurück.
func TestEinGescheiterterStartLaesstKeineOffeneDatenbank(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "bestand.db")

	// Eine Buchführung in fremder Form: die Migration scheitert danach.
	vorbereiten(t, pfad, `CREATE TABLE migrationen (fremd TEXT)`)

	anwendung := app.NeueApp(pfad)

	if err := anwendung.Starten(t.Context()); err == nil {
		t.Fatal("der Start glückte, obwohl die Buchführung unbrauchbar ist")
	}

	// Ein zweites Beenden muss still bleiben — das ist der Beleg, dass Starten
	// die Datenbank bereits geschlossen und den Verweis geräumt hat.
	if err := anwendung.Beenden(); err != nil {
		t.Errorf("nach dem gescheiterten Start blieb etwas offen: %v", err)
	}
}

// vorbereiten legt eine Datenbank an und führt darauf Anweisungen aus.
func vorbereiten(t *testing.T, pfad string, anweisungen ...string) {
	t.Helper()

	db, err := storage.Oeffnen(probekontext(t), pfad)
	if err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}
	defer func() {
		if fehler := db.Close(); fehler != nil {
			t.Errorf("vorbereiten: %v", fehler)
		}
	}()

	for _, anweisung := range anweisungen {
		if _, fehler := db.ExecContext(probekontext(t), anweisung); fehler != nil {
			t.Fatalf("vorbereiten: %v", fehler)
		}
	}
}

// frischeAppdatenbank gibt eine migrierte Datenbank über den öffentlichen Weg.
//
// Über `App.Starten` und nicht über `storage.Oeffnen` unmittelbar: so prüft
// jeder Integrationstest nebenbei mit, dass die Verdrahtung tut, was sie zusagt.
func frischeAppdatenbank(t *testing.T) *sql.DB {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "bestand.db")

	anwendung := app.NeueApp(pfad)
	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("Beenden: %v", fehler)
		}
	})

	db, err := storage.Oeffnen(probekontext(t), "file:"+pfad)
	if err != nil {
		t.Fatalf("Datenbank öffnen: %v", err)
	}
	t.Cleanup(func() {
		if fehler := db.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	return db
}

// **Der Datenstand kommt über die Fassade** (#156).
//
// Er ist die Naht, über die das Fenster bemerkt, dass der MCP-Server
// geschrieben hat. Ein Fehler darin ist ein Fehler wie jeder andere — er geht
// als `Fehler` hinaus, nicht als Wails-Zeichenkette (§8).
func TestDerDatenstandKommtUeberDieFassade(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "bestand.db")

	anwendung := app.NeueApp(pfad)
	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("Start: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("Beenden: %v", fehler)
		}
	})

	antwort := anwendung.Datenstand(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	// **Zweimal derselbe Stand**, denn niemand hat dazwischen geschrieben.
	zweite := anwendung.Datenstand(t.Context())
	if zweite.Stand != antwort.Stand {
		t.Errorf("%d gegen %d — ohne Schreibvorgang", zweite.Stand, antwort.Stand)
	}
}

// **Ein fertig migrierter Bestand mit verdorbenem Inhalt wird beim Start
// abgewiesen** — #938.
//
// # Warum es diesen Fall braucht
//
// Seit #938 prüft `storage.Migriere` den Bestand **innerhalb** der
// Migrationstransaktion, und zwar bei **jedem** Start — auch wenn nichts zu
// migrieren ist. Eine Abweisung ist dort eine Rücknahme statt eines
// festgeschriebenen Schadens.
//
// **Die zweite Prüfung in `Starten` ist damit weggefallen**: sie war ein
// Zweig, den kein Aufruf mehr betreten konnte (§3). Dieser Fall belegt, dass
// die verbliebene ihn abdeckt.
func TestEinFertigMigrierterBestandMitVerdorbenemInhaltWirdAbgewiesen(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "bestand.db")

	// Erst regulär anlegen: danach sind alle Migrationen verbucht.
	erste := app.NeueApp(pfad)
	if err := erste.Starten(probekontext(t)); err != nil {
		t.Fatalf("erster Start: %v", err)
	}

	if err := erste.Beenden(); err != nil {
		t.Fatalf("beenden: %v", err)
	}

	// Dann den Inhalt verderben — das Schema bleibt, die Migration hat nichts
	// mehr zu tun.
	//
	// **Gelöscht und nicht verfälscht**: ein CHECK im Schema lässt eine falsche
	// Form gar nicht erst zu. Genau deshalb prüft `PruefeBestand` den
	// **Inhalt** und nicht das Schema — ein CHECK gilt nur für das, was durch
	// ihn hindurchgeht.
	vorbereiten(t, pfad, `DELETE FROM bestand WHERE schluessel = 'angelegt_am'`)

	zweite := app.NeueApp(pfad)

	err := zweite.Starten(t.Context())
	if err == nil {
		t.Fatal("der verdorbene Bestand wurde angenommen")
	}

	if !strings.Contains(err.Error(), "angelegt_am fehlt") {
		t.Errorf("die Meldung nennt den verdorbenen Bestand nicht: %v", err)
	}

	if fehler := zweite.Beenden(); fehler != nil {
		t.Errorf("nach dem gescheiterten Start blieb etwas offen: %v", fehler)
	}
}
