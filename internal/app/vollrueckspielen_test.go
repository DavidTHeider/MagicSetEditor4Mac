// **Was gesichert wurde, kommt zurück — die Bilder eingeschlossen**
// (#1162).
//
// Der Gegenweg zu `bestandsicherung_test.go`. Er ist der, an dem es hängt:
// eine Sicherung, die Bilder trägt und beim Zurückspielen fallen lässt, ist
// keine Sicherung, sondern eine Zusage mit Loch.
package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// mitBildern gibt eine laufende Anwendung, deren Bilder in einem eigenen
// Verzeichnis liegen — und dieses Verzeichnis.
func mitBildern(t *testing.T, ort string) (anwendung *app.App, bilder string) {
	t.Helper()

	anwendung = laufendeAnwendung(t, filepath.Join(ort, "bestand.db"))
	bilder = filepath.Join(ort, "bilder")

	anwendung.MitBilderort(func() (string, error) { return bilder, nil })

	return anwendung, bilder
}

// bildHinlegen schreibt ein Bild unter dem Kartenordner `karte`.
func bildHinlegen(t *testing.T, bilder, karte, name, inhalt string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(bilder, karte), 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(bilder, karte, name), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}
}

// bildLesen gibt den Inhalt eines Bildes oder scheitert benannt.
func bildLesen(t *testing.T, bilder, karte, name string) string {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join(bilder, karte, name)) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("das Bild fehlt: %v", err)
	}

	return string(roh)
}

// **Ein verlorenes Bild kommt aus der Sicherung zurück.**
//
// Der Lauf stellt den Fall her, um den es geht: gesichert wird mit Bild,
// danach ist es fort — gelöscht, verlorener Ordner, neuer Rechner —, und das
// Zurückspielen holt es wieder.
func TestDasZurueckspielenHoltDieBilderZurueck(t *testing.T) {
	ort := t.TempDir()
	ziel := filepath.Join(ort, "voll.msesicherung")

	anwendung, bilder := mitBildern(t, ort)
	bildHinlegen(t, bilder, "7", "kunst.png", "das gesicherte Bild")

	gesichertMitSprache(t, anwendung, ziel, "de")

	// **Danach ist alles fort** — der Bilderordner ebenso wie die Sprache.
	if err := os.RemoveAll(bilder); err != nil {
		t.Fatalf("wegräumen: %v", err)
	}

	if antwort := anwendung.EinstellungSetzen(t.Context(), "sprache", "en"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	if antwort := anwendung.SicherungZurueckspielen(t.Context(), ziel, ""); antwort.Fehler != nil {
		t.Fatalf("zurückspielen: %+v", antwort.Fehler)
	}

	if got := bildLesen(t, bilder, "7", "kunst.png"); got != "das gesicherte Bild" {
		t.Errorf("das Bild kam anders zurück: %q", got)
	}

	// **Und der Bestand dazu.** Ein Zurückspielen, das die Bilder holt und die
	// Datenbank stehen lässt, hinterließe Karten ohne Bild und Bilder ohne
	// Karte.
	if got := merkmal(t, anwendung); got != "de" {
		t.Errorf("der Bestand steht auf %q", got)
	}
}

// **Ein Bild, das die Sicherung nicht kennt, ist danach fort.**
//
// Zurückspielen heißt **ersetzen** und nicht **ergänzen**. Bliebe ein später
// hinzugekommenes Bild liegen, stünde am Ende ein Stand da, den es nie gab —
// und niemand könnte sagen, welcher es ist.
func TestEinSpaeteresBildUeberlebtDasZurueckspielenNicht(t *testing.T) {
	ort := t.TempDir()
	ziel := filepath.Join(ort, "voll.msesicherung")

	anwendung, bilder := mitBildern(t, ort)
	bildHinlegen(t, bilder, "7", "kunst.png", "das gesicherte Bild")

	gesichertMitSprache(t, anwendung, ziel, "de")

	bildHinlegen(t, bilder, "8", "spaeter.png", "kam danach")

	if antwort := anwendung.SicherungZurueckspielen(t.Context(), ziel, ""); antwort.Fehler != nil {
		t.Fatalf("zurückspielen: %+v", antwort.Fehler)
	}

	if _, err := os.Stat(filepath.Join(bilder, "8", "spaeter.png")); !os.IsNotExist(err) {
		t.Errorf("das spätere Bild steht noch da: %v", err)
	}

	if got := bildLesen(t, bilder, "7", "kunst.png"); got != "das gesicherte Bild" {
		t.Errorf("das gesicherte Bild kam anders zurück: %q", got)
	}
}

// **Der ersetzte Bilderstand liegt beiseite und ist nicht fort.**
//
// Dieselbe Zusage wie für den Bestand ([app.Rueckspielsicherung]): wer sich
// vertan hat, soll den alten Stand noch haben.
func TestDerErsetzteBilderstandLiegtBeiseite(t *testing.T) {
	ort := t.TempDir()
	ziel := filepath.Join(ort, "voll.msesicherung")

	anwendung, bilder := mitBildern(t, ort)
	bildHinlegen(t, bilder, "7", "kunst.png", "das gesicherte Bild")

	gesichertMitSprache(t, anwendung, ziel, "de")

	bildHinlegen(t, bilder, "8", "spaeter.png", "kam danach")

	if antwort := anwendung.SicherungZurueckspielen(t.Context(), ziel, ""); antwort.Fehler != nil {
		t.Fatalf("zurückspielen: %+v", antwort.Fehler)
	}

	beiseite := filepath.Join(filepath.Dir(bilder), app.Beiwerkbeiseite)

	if got := bildLesen(t, beiseite, "8", "spaeter.png"); got != "kam danach" {
		t.Errorf("der alte Bilderstand liegt nicht beiseite: %q", got)
	}
}

// **Eine Sicherung aus 0.2.0 lässt die Bilder stehen.**
//
// Sie ist eine nackte `.db` und sagt über Bilder **nichts**. Sie zu leeren
// hieße, aus „weiß nichts darüber" ein „da war nichts" zu machen — und ein
// Zurückspielen, das die Bilder eines Jahres wegwirft, wäre der teuerste
// Datenverlust, den dieses Programm anrichten kann.
func TestEineAlteSicherungLaesstDieBilderStehen(t *testing.T) {
	ort := t.TempDir()

	anwendung, bilder := mitBildern(t, ort)
	bildHinlegen(t, bilder, "7", "kunst.png", "steht seit je da")

	// **Eine nackte Datenbank**, wie 0.1.0 und 0.2.0 sie schrieben.
	//
	// Hergestellt aus einer echten Sicherung statt nachgebaut: was hier liegt,
	// ist Zeichen für Zeichen das, was die alte Fassung geschrieben hat (§3).
	ausgepackt := t.TempDir()

	if _, err := sicherung.Entpacken(
		gesichertMitSprache(t, anwendung, filepath.Join(ort, "voll.msesicherung"), "de"),
		ausgepackt); err != nil {
		t.Fatalf("die alte Gestalt herstellen: %v", err)
	}

	alt := filepath.Join(ausgepackt, app.Bestandsname)

	if antwort := anwendung.EinstellungSetzen(t.Context(), "sprache", "en"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	if antwort := anwendung.SicherungZurueckspielen(t.Context(), alt, ""); antwort.Fehler != nil {
		t.Fatalf("zurückspielen: %+v", antwort.Fehler)
	}

	if got := bildLesen(t, bilder, "7", "kunst.png"); got != "steht seit je da" {
		t.Errorf("das Bild hat sich verändert: %q", got)
	}

	if _, err := os.Stat(filepath.Join(ort, app.Beiwerkbeiseite)); !os.IsNotExist(err) {
		t.Error("es wurde ein Bilderstand beiseitegelegt, obwohl keiner ersetzt wurde")
	}
}
