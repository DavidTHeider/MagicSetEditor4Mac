//go:build darwin

// **Ein Leser exportiert nicht — auch nicht in eine Datei** — Werkstück O1
// von Phase 5.2 (#1425),
// nach [0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md).
//
// # Der Befund
//
// **`query_only` ersetzt keine Autorisierung.** Der Leser liest über eine
// Verbindung, die nichts annimmt (0073) — aber ein Druckbogen, ein Kartenbild,
// eine Kopie des Sets und MTG JSON schreiben **Dateien**, und die Platte fragt
// den Bestand nicht. Bis zu diesem Werkstück ging jeder dieser Wege durch.
//
// **Darwin, weil der Weg den Schlüsselbund braucht** — wie `nurlesen_test.go`.

package app_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// alsLeserVorEinemSet geht als Leser in ein fremdes Profil, in dem der Eigner
// vorher ein Set geöffnet hat — und gibt das Set zurück.
func alsLeserVorEinemSet(t *testing.T) (*app.App, app.OeffnenAntwort) {
	t.Helper()

	anwendung, _ := gestarteteAnwendung(t)

	ich := anwendung.WerArbeitet(t.Context())
	if ich.Fehler != nil {
		t.Fatalf("wer arbeitet: %+v", ich.Fehler)
	}

	zweiter := anwendung.NutzerAnlegen(t.Context(), "Zweiter", "")
	if zweiter.Fehler != nil {
		t.Fatalf("anlegen: %+v", zweiter.Fehler)
	}

	fremdesProfil := zweiter.Nutzer.Profile[0].ID

	// **Der Eigner öffnet ein Set und vergibt die Freigabe.**
	if wechsel := anwendung.NutzerWechseln(t.Context(), zweiter.Nutzer.ID, ""); wechsel.Fehler != nil {
		t.Fatalf("Nutzer wechseln: %+v", wechsel.Fehler)
	}

	geoeffnet := anwendung.SetOeffnen(t.Context(), probesetVerzeichnis(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("Set öffnen: %+v", geoeffnet.Fehler)
	}

	if gesetzt := anwendung.FreigabeSetzen(t.Context(), app.FreigabeangabenDTO{
		ProfilID: fremdesProfil, NutzerID: ich.NutzerID, Rolle: string(domain.RolleLeser),
	}); gesetzt.Fehler != nil {
		t.Fatalf("Freigabe setzen: %+v", gesetzt.Fehler)
	}

	if zurueck := anwendung.NutzerWechseln(t.Context(), ich.NutzerID, ""); zurueck.Fehler != nil {
		t.Fatalf("zurück: %+v", zurueck.Fehler)
	}

	if hinein := anwendung.ProfilWechseln(t.Context(), fremdesProfil, ""); hinein.Fehler != nil {
		t.Fatalf("als Leser hinein: %+v", hinein.Fehler)
	}

	return anwendung, geoeffnet
}

// probesetVerzeichnis legt ein kleines Set als Paketverzeichnis an.
func probesetVerzeichnis(t *testing.T) string {
	t.Helper()

	quelle := filepath.Join(t.TempDir(), "Probe.mse-set")
	if err := os.MkdirAll(quelle, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\ngame: probe\nstylesheet: probe-schlicht\n" +
		"set info:\n\ttitle: Probe\ncard:\n\tname: Feuerball\n"

	if err := os.WriteFile(filepath.Join(quelle, "set"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return quelle
}

// **Jeder Weg, der eine Datei schreibt, sagt dem Leser Nein — bevor etwas
// auf der Platte liegt.**
func TestEinLeserExportiertNicht(t *testing.T) {
	anwendung, geoeffnet := alsLeserVorEinemSet(t)
	ziel := t.TempDir()

	// **Lesen geht weiter** — sonst bewiese der Lauf eine Anwendung, die
	// gar nichts kann (§10).
	karten := anwendung.Kartenzeilen(t.Context(), geoeffnet.SetID)
	if karten.Fehler != nil || len(karten.Zeilen) == 0 {
		t.Fatalf("der Leser sieht die Karten nicht: %+v", karten.Fehler)
	}

	pfad := func(name string) string { return filepath.Join(ziel, name) }

	faelle := map[string]struct {
		datei  string
		fehler *app.Fehler
	}{
		"Set sichern": func() (f struct {
			datei  string
			fehler *app.Fehler
		},
		) {
			f.datei = pfad("kopie.mse-set")
			f.fehler = anwendung.SetSpeichern(t.Context(), geoeffnet.SetID, f.datei).Fehler

			return f
		}(),
		"Druckbogen": func() (f struct {
			datei  string
			fehler *app.Fehler
		},
		) {
			f.datei = pfad("bogen.pdf")
			f.fehler = anwendung.BogenLegen(t.Context(), geoeffnet.SetID, f.datei, "a4", 0, 300, false, nil).Fehler

			return f
		}(),
		"MTG JSON": func() (f struct {
			datei  string
			fehler *app.Fehler
		},
		) {
			f.datei = pfad("set.json")
			f.fehler = anwendung.MTGJSONSchreiben(t.Context(), geoeffnet.SetID, "magic", f.datei).Fehler

			return f
		}(),
		"Kartenbild": func() (f struct {
			datei  string
			fehler *app.Fehler
		},
		) {
			f.datei = pfad("karte.png")
			f.fehler = anwendung.BildSchreiben(t.Context(), karten.Zeilen[0].ID, ziel, "karte.png", "png", 0, "", "medium").Fehler

			return f
		}(),
		"Exportvorlage": func() (f struct {
			datei  string
			fehler *app.Fehler
		},
		) {
			f.datei = pfad("vorlage.html")
			f.fehler = anwendung.VorlageAusfuehren(t.Context(), geoeffnet.SetID, "probe-vorlage", f.datei, nil, nil).Fehler

			return f
		}(),
	}

	for name, fall := range faelle {
		if fall.fehler == nil {
			t.Errorf("%s: der Leser schrieb", name)

			continue
		}

		if fall.fehler.Code != app.CodeNurLesend {
			t.Errorf("%s: Code %q, erwartet %q — %s", name, fall.fehler.Code, app.CodeNurLesend, fall.fehler.Meldung)
		}

		if _, err := os.Stat(fall.datei); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("%s: die Datei %s liegt trotzdem da (%v)", name, fall.datei, err)
		}
	}
}

// **Der Probelauf liest — ein Leser darf wissen, was er nicht schreiben darf**
// (UXR-F7, #2015).
//
// Derselbe Leser, vor demselben Set, dem eben jeder schreibende Weg Nein sagte:
// der Probelauf des Bogens rechnet, und der der Exportvorlage kommt nicht am
// Riegel zum Stehen — er scheitert an der Vorlage, die es hier nicht gibt, und
// nicht an `nurlesend`.
func TestEinLeserDarfDenProbelaufRufen(t *testing.T) {
	anwendung, geoeffnet := alsLeserVorEinemSet(t)

	bogen := anwendung.BogenProbelauf(t.Context(), geoeffnet.SetID, "a4", 0, 300, false, nil)
	if bogen.Fehler != nil || bogen.Karten == 0 || bogen.Seiten == 0 {
		t.Errorf("Bogen: %+v", bogen)
	}

	vorlage := anwendung.VorlageProbelauf(t.Context(), geoeffnet.SetID, "probe-vorlage", "", nil, nil)
	if vorlage.Fehler != nil && vorlage.Fehler.Code == app.CodeNurLesend {
		t.Errorf("die Vorlage: der Riegel hat den Probelauf gesperrt: %+v", vorlage.Fehler)
	}
}
