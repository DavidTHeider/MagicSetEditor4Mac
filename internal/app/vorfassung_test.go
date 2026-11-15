// **Was 0.5.0 geschrieben hat, liest diese Fassung** — ENG-06.AK-03
// (#1731).
//
// # Woher die Vorrichtungen kommen
//
// **Von der Fassung 0.5.0 selbst** und nicht aus einem Nachbau: die
// Kommandozeile aus dem Tag `v0.5.0` (f498fb4a) hat sie geschrieben — wie,
// steht in `testdaten/fassung-0.5.0/LIESMICH.md`. Ein Nachbau mit dem heutigen
// Code prüfte nur, ob diese Fassung sich selbst lesen kann.
//
// # Was sich seit 0.8.0 geändert hat
//
// **Die Sicherung aus 0.5.0 wird abgewiesen**, und zwar mit Rat: seit es nur
// die Grundlinie gibt (Vermerk 0189), wandert diese Fassung keinen Bestand von
// vor 0.7.0 mehr nach. Der Weg führt über 0.7.0, und dass er dort ankommt, hält
// `grundlinienaufstieg_test.go` an einer Sicherung, die 0.7.0 selbst geschrieben
// hat. Einstellungsdatei und Umschlag gehören nicht zum Schema und werden
// weiter gelesen.
//
// # Was sie nicht prüfen
//
// **Den Handschlag.** Er ist an die Leitung gebunden, über die er läuft, und
// an zwei frisch gewürfelte Zahlen; ein Mitschnitt lässt sich deshalb nicht
// wieder vorlegen — und genau das soll er auch nicht. Belegt ist hier die
// **Unterschrift** eines 0.5.0-Peers, über den Umschlag; der Handschlag
// zwischen zwei Fassungen steht in `docs/FASSUNGEN.md` als Handabnahme.
package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// vorfassung legt eine Vorrichtung aus 0.5.0 in einen Wegwerfort.
//
// **Eine Kopie und nicht die eingecheckte Datei**: ein Zurückspielen darf
// nichts an dem ändern, was der nächste Lauf wieder braucht.
func vorfassung(t *testing.T, name string) string {
	t.Helper()

	return ausFassung(t, "0.5.0", name)
}

// ausFassung legt eine Vorrichtung aus einer früheren Fassung in einen
// Wegwerfort.
func ausFassung(t *testing.T, fassung, name string) string {
	t.Helper()

	//nolint:gosec // ein fester Ort im Baum; Fassung und Name kommen aus diesem Test
	roh, fehler := os.ReadFile(filepath.Join("..", "..", "testdaten", "fassung-"+fassung, name))
	if fehler != nil {
		t.Fatalf("die Vorrichtung aus %s fehlt: %v", fassung, fehler)
	}

	ziel := filepath.Join(t.TempDir(), name)
	if fehler := os.WriteFile(ziel, roh, 0o600); fehler != nil { //nolint:gosec // ziel liegt in t.TempDir()
		t.Fatal(fehler)
	}

	return ziel
}

// **Eine Sicherung aus 0.5.0 wird mit Rat abgewiesen** — und der laufende
// Bestand bleibt, wie er war. Bis 0.7.0 stand hier der Aufstieg aus
// ENG-06.AK-02 an derselben Datei; seit Vermerk 0189 führt er über 0.7.0.
func TestEineSicherungAusNullFuenfWirdMitRatAbgewiesen(t *testing.T) {
	ort := t.TempDir()
	anwendung := laufendeAnwendung(t, filepath.Join(ort, "bestand.db"))

	if antwort := anwendung.EinstellungSetzen(t.Context(), "sprache", "de"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	antwort := anwendung.SicherungZurueckspielen(t.Context(), vorfassung(t, "sicherung.zip"), "")
	if antwort.Fehler == nil || antwort.Fehler.Schluessel != domain.SchluesselSicherungAeltereFassung {
		t.Fatalf("die Sicherung aus 0.5.0: %+v", antwort.Fehler)
	}

	if got := merkmal(t, anwendung); got != "de" {
		t.Errorf("der laufende Bestand hat sich verändert: %q", got)
	}

	if _, err := os.Stat(filepath.Join(ort, app.Rueckspielsicherung)); !os.IsNotExist(err) {
		t.Error("es wurde etwas beiseitegelegt, obwohl gar nicht getauscht wurde")
	}
}

// **Eine Einstellungsdatei aus 0.5.0 wird gelesen und übernommen** — ohne
// Befund, denn sie trägt nichts, was diese Fassung nicht kennt.
func TestEineEinstellungsdateiAusNullFuenfWirdUebernommen(t *testing.T) {
	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))

	roh, fehler := os.ReadFile(vorfassung(t, "einstellungen.json"))
	if fehler != nil {
		t.Fatal(fehler)
	}

	vorschau := anwendung.EinstellungenVorschau(t.Context(), string(roh))
	if vorschau.Fehler != nil {
		t.Fatalf("die Einstellungen aus 0.5.0 ließen sich nicht lesen: %+v", vorschau.Fehler)
	}

	if vorschau.Sicht.Fassung != 1 || len(vorschau.Sicht.Befunde) != 0 {
		t.Errorf("Fassung %d, Befunde %+v", vorschau.Sicht.Fassung, vorschau.Sicht.Befunde)
	}

	uebernahme := anwendung.EinstellungenUebernehmen(t.Context(), string(roh),
		vorschau.Sicht.Revision, nil)
	if uebernahme.Fehler != nil {
		t.Fatalf("übernehmen: %+v", uebernahme.Fehler)
	}

	if uebernahme.Bericht.Uebernommen != 4 {
		t.Errorf("übernommen: %d, erwartet waren die vier Werte der Datei",
			uebernahme.Bericht.Uebernommen)
	}

	if got := merkmal(t, anwendung); got != "en" {
		t.Errorf("die Sprache ist %q, erwartet war %q", got, "en")
	}
}

// **Ein Umschlag, den ein 0.5.0-Peer unterschrieben hat, wird übernommen** —
// die Unterschrift der vorigen Peerfassung trägt in dieser.
func TestEinUmschlagAusNullFuenfWirdUebernommen(t *testing.T) {
	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))
	umschlag := vorfassung(t, "probeset.mseumschlag")

	angesehen := anwendung.UmschlagAnsehen(t.Context(), umschlag)
	if angesehen.Fehler != nil {
		t.Fatalf("der Umschlag aus 0.5.0 ließ sich nicht lesen: %+v", angesehen.Fehler)
	}

	// **Die Kennung, unter der 0.5.0 unterschrieben hat** — sie steht auch in
	// `LIESMICH.md` daneben.
	const urheber = "bafzaajaiaejcbbdwuyds5lqlzi3gg3yx5ooj3nowx3kci5asu5ani6sevore3zsd"
	if angesehen.Urheber != urheber {
		t.Errorf("Urheber %q, erwartet war %q", angesehen.Urheber, urheber)
	}

	uebernommen := anwendung.UmschlagUebernehmen(t.Context(), umschlag)
	if uebernommen.Fehler != nil {
		t.Fatalf("übernehmen: %+v", uebernommen.Fehler)
	}

	sets := anwendung.Sets(t.Context())
	if sets.Fehler != nil || len(sets.Sets) != 1 ||
		sets.Sets[0].Name != "Mein Probeset" || sets.Sets[0].Karten != 3 {
		t.Errorf("nach der Übernahme: %+v", sets)
	}
}
