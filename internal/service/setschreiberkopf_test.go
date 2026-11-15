package service_test

// **Spiel und Stil im Kopf eines geschriebenen Sets** — F05
// (#2247),
// Befund 2 des Zweitblicks über Phase G, zweiter Teil.
//
// Zwei Fehler, gefunden beim Nachlesen des Befunds:
//
//  1. **Ein gewechselter Stil kam nicht in die Datei.** `fuellen` glich
//     Karten, Schlüsselwörter, Setangaben und Stiloptionen an, aber nicht
//     `game:` und `stylesheet:` — ein Set aus einer Datei behielt beim
//     Sichern seinen alten Stil (gemessen mit `mse stile 1 wechseln`).
//  2. **Der Name stand in der falschen Form.** MSE2 schreibt die Kurzform und
//     sucht `<spiel>-<kurzform>.mse-style` (`src/data/stylesheet.cpp`,
//     `stylesheetName` und `byGameAndName`, Zeilen 26–60, Fassung f9d9356
//     vom 2021-07-06). Ein hier gewählter Stil stand als voller Paketname da,
//     und MSE2 fand ihn nicht; der Stilblock bekam `probe-probe-…`.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// pakete ist eine Auflösung über eine feste Liste von Stilpaketen.
func pakete(namen ...string) service.Stilaufloesung {
	return func(stil, spiel string) (string, bool) {
		for _, kandidat := range []string{spiel + "-" + stil, stil} {
			for _, name := range namen {
				if name == kandidat {
					return name, true
				}
			}
		}

		return "", false
	}
}

func geschriebenerKopf(t *testing.T, bestand quellenDouble, aufloesung service.Stilaufloesung) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "wieder.mse-set")

	schreiber := service.NeuerSetschreiber(bestand, &ortsschreiber{})
	if aufloesung != nil {
		schreiber = schreiber.MitStilaufloesung(aufloesung)
	}

	if _, fehler := schreiber.Schreibe(t.Context(), 7, ort); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	gelesen, fehler := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	return string(gelesen)
}

func TestDerKopfFolgtDemBestand(t *testing.T) {
	quelle := "game: probe\nstylesheet: kunst\nset info:\n\ttitle: Mein Probeset\ncard:\n\tname: Feuerball\n"

	for _, fall := range []struct {
		name, spiel, stil string
		aufloesung        service.Stilaufloesung
		kopf              string
		stilblock         string
	}{
		{"ein gewechselter Stil in Kurzform", "probe", "schlicht", nil, "game: probe\nstylesheet: schlicht\n", "\tprobe-schlicht:\n"},
		{"ein voller Paketname wird zur Kurzform", "probe", "probe-schlicht", pakete("probe-schlicht"), "game: probe\nstylesheet: schlicht\n", "\tprobe-schlicht:\n"},
		{"eine Kurzform mit Spiel davor bleibt", "probe", "probe-x", pakete("probe-probe-x"), "game: probe\nstylesheet: probe-x\n", "\tprobe-probe-x:\n"},
		{"ohne Auflösung bleibt der Wert, wie er steht", "probe", "probe-schlicht", nil, "game: probe\nstylesheet: probe-schlicht\n", "\tprobe-probe-schlicht:\n"},
		{"ein Spielwechsel steht im Kopf", "andere", "schlicht", nil, "game: andere\nstylesheet: schlicht\n", "\tandere-schlicht:\n"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			bestand := mitQuelle()
			bestand.koepfe[0].Game, bestand.koepfe[0].Style = fall.spiel, fall.stil
			bestand.stilwerte = []domain.Value{{Field: "chop top", Text: "7"}}
			bestand.quelle = &domain.Quelle{Datei: "set", Inhalt: []byte(quelle)}

			geschrieben := geschriebenerKopf(t, bestand, fall.aufloesung)

			if !strings.HasPrefix(geschrieben, fall.kopf) {
				t.Errorf("Kopf:\n%s", geschrieben)
			}

			if !strings.Contains(geschrieben, "styling:\n"+fall.stilblock) {
				t.Errorf("Stilblock nicht unter %q:\n%s", strings.TrimSpace(fall.stilblock), geschrieben)
			}
		})
	}
}

// **Was der Bestand nicht nennt, bleibt, wie es steht** — ein Set ohne Stil
// im Bestand verliert seine Zeile nicht, und eine gleiche bleibt Byte für
// Byte (`game:  probe` mit zwei Leerzeichen).
func TestEinUnveraenderterKopfBleibtByteFuerByte(t *testing.T) {
	bestand := mitQuelle()
	bestand.quelle = &domain.Quelle{Datei: "set", Inhalt: []byte("game:  probe\nstylesheet: kunst\ncard:\n\tname: Feuerball\n")}

	if geschrieben := geschriebenerKopf(t, bestand, pakete("probe-kunst")); !strings.HasPrefix(geschrieben, "game:  probe\nstylesheet: kunst\n") {
		t.Errorf("Kopf:\n%s", geschrieben)
	}
}

func TestDieAufloesungSuchtInDenPaketwurzeln(t *testing.T) {
	wurzel := t.TempDir()
	paket := filepath.Join(wurzel, "probe-schlicht.mse-style")

	if fehler := os.MkdirAll(paket, 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.WriteFile(filepath.Join(paket, "style"), []byte("mse version: 2.0.0\ngame: probe\n"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	aufloesung := service.StilaufloesungUeber(t.TempDir(), wurzel)

	for _, stil := range []string{"schlicht", "probe-schlicht"} {
		if name, gefunden := aufloesung(stil, "probe"); !gefunden || name != "probe-schlicht" {
			t.Errorf("%s: %q, %v", stil, name, gefunden)
		}
	}

	if _, gefunden := aufloesung("fehlt", "probe"); gefunden {
		t.Error("ein Paket, das es nicht gibt, wurde gefunden")
	}
}

func TestDerKopfInSonderfaellen(t *testing.T) {
	for _, fall := range []struct {
		name, quelle, stil string
		aufloesung         service.Stilaufloesung
		kopf               string
	}{
		{"ein Stil, den keine Wurzel kennt, bleibt, wie er steht", "game: probe\nstylesheet: alt\ncard:\n\tname: Feuerball\n", "fehlt", pakete(), "game: probe\nstylesheet: fehlt\n"},
		{"ein Paket ohne Spiel davor steht, wie es heißt", "game: probe\nstylesheet: alt\ncard:\n\tname: Feuerball\n", "fremd", pakete("fremd"), "game: probe\nstylesheet: fremd\n"},
		{"eine fehlende Zeile kommt hinter das Spiel", "mse version: 2.0.0\ngame: probe\ncard:\n\tname: Feuerball\n", "schlicht", nil, "mse version: 2.0.0\ngame: probe\nstylesheet: schlicht\n"},
		{"ohne Spiel kommt sie an den Anfang", "card:\n\tname: Feuerball\n", "schlicht", nil, "stylesheet: schlicht\ncard:\n"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			bestand := mitQuelle()
			bestand.koepfe[0].Style = fall.stil
			bestand.quelle = &domain.Quelle{Datei: "set", Inhalt: []byte(fall.quelle)}

			if geschrieben := geschriebenerKopf(t, bestand, fall.aufloesung); !strings.HasPrefix(geschrieben, fall.kopf) {
				t.Errorf("Kopf:\n%s", geschrieben)
			}
		})
	}
}

// **Ist der Stil aufgelöst, gilt nur der Block seines Pakets** — Nachlauf
// über Phase G, zweiter Teil, Befund 1. Ein vorhandener Block unter einem
// anderen Namen (`magic-magic-m15`, wie ihn der Schreiber vor F05 anlegte)
// gewann gegen den aufgelösten, und die Option wirkte nach dem Wiederöffnen
// nicht: gelesen wird sie unter dem Namen des Pakets ([setstilwerte]).
func TestEinAufgeloesterStilSchreibtNurInSeinenBlock(t *testing.T) {
	bestand := mitQuelle()
	bestand.koepfe[0].Game, bestand.koepfe[0].Style = "magic", "magic-m15"
	bestand.stilwerte = []domain.Value{{Field: "chop top", Text: "7"}}
	bestand.quelle = &domain.Quelle{Datei: "set", Inhalt: []byte(
		"game: magic\nstylesheet: magic-m15\ncard:\n\tname: Feuerball\nstyling:\n\tmagic-magic-m15:\n\t\tchop top: alt\n")}

	geschrieben := geschriebenerKopf(t, bestand, pakete("magic-m15"))

	if !strings.Contains(geschrieben, "\tmagic-m15:\n\t\tchop_top: 7\n") {
		t.Errorf("die Option steht nicht unter dem aufgelösten Paket:\n%s", geschrieben)
	}

	if !strings.Contains(geschrieben, "\tmagic-magic-m15:\n\t\tchop top: alt\n") {
		t.Errorf("der fremde Block hat sich verändert:\n%s", geschrieben)
	}
}

// **Eine ergänzte Kopfzeile klebt nicht an der letzten** — Nachlauf über
// Phase G, zweiter Teil, Befund 2: endet die Datei ohne Zeilenende mit
// `game: probe`, wurde daraus `game: probestylesheet: schlicht`.
func TestEineErgaenzteKopfzeileKlebtNichtAnDerLetzten(t *testing.T) {
	for _, quelle := range []string{"game: probe", "mse version: 2.0.0\r\ngame: probe"} {
		bestand := mitQuelle()
		bestand.koepfe[0].Style = "schlicht"
		bestand.quelle = &domain.Quelle{Datei: "set", Inhalt: []byte(quelle)}

		geschrieben := geschriebenerKopf(t, bestand, nil)

		if !strings.Contains(geschrieben, "game: probe\r\nstylesheet: schlicht") &&
			!strings.Contains(geschrieben, "game: probe\nstylesheet: schlicht") {
			t.Errorf("aus %q wurde:\n%q", quelle, geschrieben)
		}
	}
}
