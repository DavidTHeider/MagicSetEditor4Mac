package service_test

// Was eine Exportvorlage über ihr eigenes Paket hinaus braucht (#216):
// eingebundene Module und das Initskript ihres Spiels.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein `include file:` in einer Exportvorlage wird aufgelöst** (#216).
//
// # Was hier fehlte
//
// `vorlageAus` öffnete das Paket und griff sofort nach `Exportmodell()`. Der
// Weg über `ModuleEinbinden`, den `stilquelle.go` für Stile und Spiele längst
// geht, fehlte — und damit stand die Zeile `include file: script` unaufgelöst
// im Skript. Der Zerteiler las `include` als Namen und meldete an Spalte 9
// „hier hört der Ausdruck auf, aber der Text geht weiter".
//
// Gemessen an `M15-Magic-Pack`: zwei der neunzehn Vorlagen — `magic-lackeybot`
// und `magic-namecheck` — beginnen ihr Skript genau so.
func TestEinIncludeInEinerExportvorlageWirdAufgeloest(t *testing.T) {
	wurzel := t.TempDir()
	paket := filepath.Join(wurzel, "probe-mitmodul.mse-export-template")
	legeVerzeichnis(t, paket)

	// Das Modul liegt **im eigenen Paket** und ohne Schrägstrich davor — so
	// schreibt es `magic.mse-game`, und so steht es in beiden Vorlagen oben.
	if err := os.WriteFile(filepath.Join(paket, "teilskript"),
		[]byte("gruss := \"aus dem Modul\"\n"), 0o600); err != nil {
		t.Fatalf("Modul schreiben: %v", err)
	}

	inhalt := "mse version: 2.0.0\ngame: probe\nshort name: MitModul\n" +
		"full name: Vorlage mit Modul\nfile type: *.txt|*.txt\n" +
		"script:\n\tinclude file: teilskript\n\tgruss\n"

	if err := os.WriteFile(filepath.Join(paket, "export-template"),
		[]byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Game: "probe"}}}

	vorlagen, befunde, err := service.NeueVorlagenliste(quelle, wurzel).
		Vorlagen(t.Context(), 7)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(vorlagen) != 1 {
		t.Fatalf("%d Vorlagen, erwartet 1; Befunde: %v", len(vorlagen), befunde)
	}

	// **Die Gegenprobe geht über das Ergebnis, nicht über den Text.** Dass das
	// Modul dasteht, sagt noch nicht, dass es auch zerteilt wird.
	ausgabe, _, fehler := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), vorlagen[0], nil, nil, nil)
	if fehler != nil {
		t.Fatalf("ausführen: %v", fehler)
	}

	if ausgabe != "aus dem Modul" {
		t.Errorf("Ausgabe %q — das Modul ist nicht eingebunden", ausgabe)
	}
}

// **Ohne Spielpaket gibt es kein Spielskript — und keinen Befund.**
//
// Eine Vorlage ohne installiertes Spielpaket ist der Normalfall einer frischen
// Einrichtung; wer sie ausführt, bekommt den unbekannten Namen genannt, und
// *der* sagt, was fehlt (#216).
func TestOhneSpielpaketBleibtDasSpielskriptLeer(t *testing.T) {
	wurzel := vorlagenwurzel(t)

	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Game: "probe"}}}

	vorlagen, _, err := service.NeueVorlagenliste(quelle, wurzel).Vorlagen(t.Context(), 7)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(vorlagen) != 1 {
		t.Fatalf("%d Vorlagen", len(vorlagen))
	}

	if vorlagen[0].Spielskript != "" {
		t.Errorf("Spielskript %q — es gibt kein Spielpaket", vorlagen[0].Spielskript)
	}
}

// **Mit Spielpaket steht sein Initskript in der Vorlage.**
func TestMitSpielpaketKommtDasSpielskriptMit(t *testing.T) {
	wurzel := vorlagenwurzel(t)

	spiel := filepath.Join(wurzel, "probe.mse-game")
	legeVerzeichnis(t, spiel)

	if err := os.WriteFile(filepath.Join(spiel, "game"),
		[]byte("mse version: 2.0.0\nshort name: Probe\ninit script:\n\tgruss := \"vom Spiel\"\n"),
		0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Game: "probe"}}}

	vorlagen, _, err := service.NeueVorlagenliste(quelle, wurzel).Vorlagen(t.Context(), 7)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(vorlagen) != 1 || vorlagen[0].Spielskript == "" {
		t.Fatalf("%d Vorlagen, Spielskript %q", len(vorlagen), vorlagen[0].Spielskript)
	}

	// **Die Gegenprobe geht über das Ergebnis**: dass das Skript dasteht, sagt
	// noch nicht, dass die Vorlage seine Namen sieht.
	ausgabe, _, fehler := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), msepaket.Exportvorlage{
			KurzerName: "P", Skript: "gruss", Spielskript: vorlagen[0].Spielskript,
		}, nil, nil, nil)
	if fehler != nil {
		t.Fatalf("ausführen: %v", fehler)
	}

	if ausgabe != "vom Spiel" {
		t.Errorf("%q", ausgabe)
	}
}

// **Das Spielpaket wird in jeder Wurzel gesucht, nicht nur in der ersten.**
//
// Die Vorlage kann aus einem bezogenen Paketverzeichnis kommen und ihr Spiel
// aus der mitgelieferten Saat — oder umgekehrt (#216).
func TestDasSpielpaketDarfInEinerZweitenWurzelLiegen(t *testing.T) {
	vorlagen := vorlagenwurzel(t)
	spiele := t.TempDir()

	spiel := filepath.Join(spiele, "probe.mse-game")
	legeVerzeichnis(t, spiel)

	if err := os.WriteFile(filepath.Join(spiel, "game"),
		[]byte("mse version: 2.0.0\nshort name: Probe\ninit script:\n\tgruss := \"zweite Wurzel\"\n"),
		0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Game: "probe"}}}

	gefunden, _, err := service.NeueVorlagenliste(quelle, vorlagen, spiele).
		Vorlagen(t.Context(), 7)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gefunden) != 1 || gefunden[0].Spielskript == "" {
		t.Fatalf("%d Vorlagen, Spielskript %q", len(gefunden), gefunden[0].Spielskript)
	}
}
