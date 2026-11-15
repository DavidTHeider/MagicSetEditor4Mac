package app_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// vorlagenprobe gibt vorgegebene Vorlagen und eine vorgegebene Ausgabe.
type vorlagenprobe struct {
	vorlagen []msepaket.Exportvorlage
	befunde  []domain.Befund
	fehler   error

	ausgabe          string
	laufbefunde      []domain.Befund
	lauffehler       error
	gerufeneWahl     map[string]string
	gerufeneSetwerte []domain.Value

	// bildziel, beigaben und symbolschriften halten fest, was der Aufrufer
	// erlaubt hat (#657, #658).
	symbolschriften   service.Symbolschriftquelle
	symbolzeichnungen service.Symbolquelle
	setpaket          string
	verzeichnis       string
	bildziel          string
	beigaben          service.Beigabenschreiber

	// waehrend läuft **in** der Auswertung und bekommt die Mittel dieses
	// Laufs — wie eine Vorlage, die `copy_file` oder `write_image_file` ruft
	// (#2015).
	waehrend func(mittel service.Ausfuhrmittel)
}

func (v *vorlagenprobe) Vorlagen(context.Context, int64) (
	[]msepaket.Exportvorlage, []domain.Befund, error,
) {
	return v.vorlagen, v.befunde, v.fehler
}

// **Die Fähigkeiten kommen mit dem Lauf, und die Attrappe merkt sie sich.**
//
// Sie stehen in der Naht, weil der Aufrufer entscheidet, ob und wohin eine
// Vorlage schreiben darf; hier wird nur festgehalten, **dass** er es tut —
// ohne diesen Merkposten war `MitBildschreiber` seit #216 unerreichbar, und
// kein Test hat es gesehen (#657).
//
// Bis zum 2026-08-21 waren es vier Setter am geteilten Ausführer, und die
// Attrappe fing sie einzeln ab. Seit
// #736 reisen
// sie als **ein Wert je Lauf** — was hier ankommt, gehört damit belegbar zu
// diesem Aufruf und nicht zu irgendeinem davor.
func (v *vorlagenprobe) Fuehre(
	_ context.Context, mittel service.Ausfuhrmittel, _ msepaket.Exportvorlage,
	_ []*domain.CardRow, setwerte []domain.Value, wahl map[string]string,
) (string, []domain.Befund, error) {
	v.gerufeneSetwerte = setwerte
	v.gerufeneWahl = wahl
	v.bildziel = mittel.Bildziel
	v.beigaben = mittel.Beigaben
	v.symbolzeichnungen = mittel.Symbolzeichnungen
	v.setpaket = mittel.Setpaket
	v.verzeichnis = mittel.Verzeichnis

	if v.waehrend != nil {
		v.waehrend(mittel)
	}

	return v.ausgabe, v.laufbefunde, v.lauffehler
}

func (v *vorlagenprobe) MitSymbolen(
	schriften service.Symbolschriftquelle, _ service.Symbolrasterer,
) *service.Vorlagenausfuehrer {
	v.symbolschriften = schriften

	return nil
}

// probevorlage gibt eine Vorlage, wie das Paket sie liefert.
func probevorlage() msepaket.Exportvorlage {
	return msepaket.Exportvorlage{
		Paket:      "probe.mse-export-template",
		VollerName: "Ausfuhr der Probe",
		Dateityp:   "*.txt|*.txt",
		Skript:     `to_string("fertig")`,
		Optionen: []msepaket.Exportoption{
			{Art: "boolean", Name: "mit namen", Vorgabe: "yes"},
		},
	}
}

// **Die Vorlagen erreichen die Oberfläche** — die Naht aus #209.
func TestExportvorlagenErreichenDieOberflaeche(t *testing.T) {
	anwendung := mitVorlagen(t, &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
	})

	antwort := anwendung.Exportvorlagen(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%v", antwort.Fehler)
	}

	if len(antwort.Vorlagen) != 1 || antwort.Vorlagen[0].Name != "Ausfuhr der Probe" {
		t.Fatalf("%+v", antwort.Vorlagen)
	}

	if len(antwort.Vorlagen[0].Optionen) != 1 ||
		antwort.Vorlagen[0].Optionen[0].Auswahl == nil {
		t.Errorf("%+v — die Auswahl darf nie nil sein", antwort.Vorlagen[0].Optionen)
	}
}

// **Der Name fällt zurück**: voller Name, sonst kurzer, sonst Paket.
func TestDerNameEinerVorlageFaelltZurueck(t *testing.T) {
	for _, fall := range []struct {
		name     string
		vorlage  msepaket.Exportvorlage
		erwartet string
	}{
		{"voller Name", msepaket.Exportvorlage{
			Paket: "p", KurzerName: "kurz", VollerName: "voll",
		}, "voll"},
		{"nur kurzer", msepaket.Exportvorlage{Paket: "p", KurzerName: "kurz"}, "kurz"},
		{"nur Paket", msepaket.Exportvorlage{Paket: "p"}, "p"},
		{"gar nichts", msepaket.Exportvorlage{}, ""},
	} {
		t.Run(fall.name, func(t *testing.T) {
			anwendung := mitVorlagen(t, &vorlagenprobe{
				vorlagen: []msepaket.Exportvorlage{fall.vorlage},
			})

			antwort := anwendung.Exportvorlagen(t.Context(), "1")
			if antwort.Vorlagen[0].Name != fall.erwartet {
				t.Errorf("%q, erwartet %q", antwort.Vorlagen[0].Name, fall.erwartet)
			}
		})
	}
}

// **Eine unlesbare Kennung ist ein Fehler und kein Absturz.**
func TestExportvorlagenMitKrummerKennung(t *testing.T) {
	anwendung := mitVorlagen(t, &vorlagenprobe{})

	if antwort := anwendung.Exportvorlagen(t.Context(), "keine Zahl"); antwort.Fehler == nil {
		t.Error("kein Fehler")
	}
}

// **Ein Lesefehler kommt übersetzt heraus.**
func TestExportvorlagenMeldetEinenLesefehler(t *testing.T) {
	anwendung := mitVorlagen(t, &vorlagenprobe{fehler: errors.New("Bestand kaputt")})

	if antwort := anwendung.Exportvorlagen(t.Context(), "1"); antwort.Fehler == nil {
		t.Error("kein Fehler")
	}
}

// **Die Befunde reisen mit.**
func TestExportvorlagenReichtBefundeDurch(t *testing.T) {
	anwendung := mitVorlagen(t, &vorlagenprobe{
		befunde: []domain.Befund{{Meldung: "eine Vorlage ohne Skript"}},
	})

	if antwort := anwendung.Exportvorlagen(t.Context(), "1"); len(antwort.Befunde) != 1 {
		t.Errorf("%v", antwort.Befunde)
	}
}

// **Eine Vorlage schreibt ihre Ausgabe.**
func TestEineVorlageSchreibtIhreAusgabe(t *testing.T) {
	probe := &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
		ausgabe:  "fertig",
	}

	anwendung := mitVorlagen(t, probe)
	ziel := filepath.Join(t.TempDir(), "ausfuhr.txt")

	antwort := anwendung.VorlageAusfuehren(t.Context(), "1",
		"probe.mse-export-template", ziel, map[string]string{"mit namen": "no"}, nil)
	if antwort.Fehler != nil {
		t.Fatalf("%v", antwort.Fehler)
	}

	inhalt, err := os.ReadFile(ziel) // #nosec G304 -- t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if string(inhalt) != "fertig" {
		t.Errorf("%q", inhalt)
	}

	if probe.gerufeneWahl["mit namen"] != "no" {
		t.Errorf("%v — die Wahl kam nicht an", probe.gerufeneWahl)
	}
}

// **Ohne Ziel ist es abgebrochen, kein Fehler.**
func TestOhneZielIstDieVorlagenausfuhrAbgebrochen(t *testing.T) {
	anwendung := mitVorlagen(t, &vorlagenprobe{})

	antwort := anwendung.VorlageAusfuehren(t.Context(), "1", "p", "", nil, nil)
	if !antwort.Abgebrochen || antwort.Fehler != nil {
		t.Errorf("%+v", antwort)
	}
}

// **Eine unbekannte Vorlage ist ein Fehler.**
func TestEineUnbekannteVorlageWirdGemeldet(t *testing.T) {
	anwendung := mitVorlagen(t, &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
	})

	antwort := anwendung.VorlageAusfuehren(t.Context(), "1", "gibtesnicht",
		filepath.Join(t.TempDir(), "a.txt"), nil, nil)
	if antwort.Fehler == nil {
		t.Error("kein Fehler")
	}
}

// **Eine krumme Kennung beim Ausführen ebenso.**
func TestVorlageAusfuehrenMitKrummerKennung(t *testing.T) {
	anwendung := mitVorlagen(t, &vorlagenprobe{})

	antwort := anwendung.VorlageAusfuehren(t.Context(), "keine Zahl", "p",
		filepath.Join(t.TempDir(), "a.txt"), nil, nil)
	if antwort.Fehler == nil {
		t.Error("kein Fehler")
	}
}

// **Ein Lesefehler beim Suchen der Vorlage kommt heraus.**
func TestVorlageAusfuehrenMeldetEinenLesefehler(t *testing.T) {
	anwendung := mitVorlagen(t, &vorlagenprobe{fehler: errors.New("kaputt")})

	antwort := anwendung.VorlageAusfuehren(t.Context(), "1", "p",
		filepath.Join(t.TempDir(), "a.txt"), nil, nil)
	if antwort.Fehler == nil {
		t.Error("kein Fehler")
	}
}

// **Warum eine Vorlage nicht durchläuft, steht als Befund.**
//
// Der Satz des Skriptfehlers ist eine Aussage über **ihre** Vorlage; ohne ihn
// stünde nur „Die Eingabe ist nicht gültig" (#209).
func TestEineGescheiterteVorlageNenntDenGrund(t *testing.T) {
	anwendung := mitVorlagen(t, &vorlagenprobe{
		vorlagen:   []msepaket.Exportvorlage{probevorlage()},
		lauffehler: errors.New(`der Name "write_image_file" ist hier nicht bekannt`),
	})

	antwort := anwendung.VorlageAusfuehren(t.Context(), "1",
		"probe.mse-export-template", filepath.Join(t.TempDir(), "a.txt"), nil, nil)
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if !hatBefundMit(antwort.Befunde, "write_image_file") {
		t.Errorf("%v — der Grund fehlt", antwort.Befunde)
	}
}

// **Ein unschreibbares Ziel wird gemeldet.**
func TestEinUnschreibbaresZielWirdGemeldet(t *testing.T) {
	anwendung := mitVorlagen(t, &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
	})

	antwort := anwendung.VorlageAusfuehren(t.Context(), "1",
		"probe.mse-export-template", filepath.Join(t.TempDir(), "fehlt", "a.txt"), nil, nil)
	if antwort.Fehler == nil {
		t.Error("kein Fehler")
	}
}

// **Der Dateityp wird auf sein erstes echtes Muster gelesen.**
//
// MSE schreibt ihn als `*.xml|*.xml|*.*|*.*` — abwechselnd Beschriftung und
// Muster. Ein Dialog, der „alle Dateien" vorschlägt, hilft niemandem.
func TestDasErsteMusterEinesDateityps(t *testing.T) {
	for _, fall := range []struct{ dateityp, erwartet string }{
		{"*.xml|*.xml|*.*|*.*", "*.xml"},
		{"Text|*.txt", "*.txt"},
		{"*.*|*.*", "*.*"},
		{"", "*.*"},
	} {
		t.Run(fall.dateityp, func(t *testing.T) {
			if gelesen := app.ErstesMuster(fall.dateityp); gelesen != fall.erwartet {
				t.Errorf("%q, erwartet %q", gelesen, fall.erwartet)
			}
		})
	}
}

// **Und der Vorschlagsname trägt seine Endung.**
func TestDerVorschlagTraegtSeineEndung(t *testing.T) {
	if app.EndungVon("*.txt") != ".txt" || app.EndungVon("*.*") != "" {
		t.Errorf("%q / %q", app.EndungVon("*.txt"), app.EndungVon("*.*"))
	}
}

// hatBefundMit sagt, ob ein Befund den Teiltext führt.
func hatBefundMit(befunde []string, teil string) bool {
	for _, befund := range befunde {
		if strings.Contains(befund, teil) {
			return true
		}
	}

	return false
}

// mitVorlagen baut eine Anwendung mit vorgegebenen Vorlagen und Karten.
func mitVorlagen(t *testing.T, probe *vorlagenprobe) *app.App {
	t.Helper()

	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if err := anwendung.Beenden(); err != nil {
			t.Errorf("beenden: %v", err)
		}
	})

	// **Zwei Sets, und sie heißen 1 und 2** — die Läufe nennen sie so. Bis
	// #1815 ging eine Ausfuhr auch über ein Set, das es nicht gab: seine Karten
	// waren eine leere Liste statt „nicht gefunden", und diese Läufe führten
	// ein Set aus, das nie angelegt worden war.
	for _, erwartet := range []string{"1", "2"} {
		antwort := anwendung.SetAnlegen(t.Context(), "Vorlagenset "+erwartet, "probe", "probe-schlicht")
		if antwort.Fehler != nil || antwort.ID != erwartet {
			t.Fatalf("anlegen: %+v", antwort)
		}
	}

	anwendung.MitVorlagen(probe)

	return anwendung
}
