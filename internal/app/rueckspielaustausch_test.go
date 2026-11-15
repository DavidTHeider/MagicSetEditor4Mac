// **Was geprüft wurde, ist was ankommt** —
// #964.
package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Eine Datei, die zwischen Prüfung und Übernahme ausgetauscht wird, kommt
// nicht an.**
//
// # Der Befund
//
// `SicherungZurueckspielen` prüfte die Datei über ihren **Pfad** und öffnete
// danach **denselben Pfad** ein zweites Mal, um zu kopieren. Zwischen beiden
// Öffnungen kann ein anderer Prozess — ein Abgleichdienst, ein Skript — die
// Datei ersetzen.
//
// Die Nutzenden bestätigen Bestand **A**, übernommen wird Bestand **B**. Ist B
// unbrauchbar, dreht der Rückweg es zurück; ist B ein **anderer gültiger**
// Bestand, wird er ohne erneute Rückfrage angenommen — und niemand erfährt es.
//
// **Der Kommentar an der Stelle nannte genau diese Bedrohung** als Grund für
// die zweite Prüfung und band Prüfung und Kopie trotzdem nicht an dasselbe
// Objekt. Angestrichen vom Zweitblick über Phase S, Befund 7.
//
// # Wie der Lauf ihn herstellt
//
// Der Prüfer ist eine Naht. Dieser Lauf hängt einen davor, der die Datei
// **während der Prüfung** durch eine andere gültige ersetzt — genau das
// Fenster, um das es geht. Danach muss der Stand aus der **geprüften** Datei
// dastehen.
func TestEineAusgetauschteSicherungKommtNichtAn(t *testing.T) {
	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")
	geprueft := filepath.Join(ort, "geprueft.db")

	// **Zwei Sicherungen, beide gültig**, mit unterscheidbarem Inhalt.
	anwendung := laufendeAnwendung(t, bestand)

	andere := gesichertMitSprache(t, anwendung, filepath.Join(ort, "andere.db"), "en")
	echte := gesichertMitSprache(t, anwendung, geprueft, "de")

	// **Der laufende Stand ist `en`** — derselbe wie in der ausgetauschten
	// Datei. Ohne das sagte ein `en` am Ende nicht, woher es kommt.
	if antwort := anwendung.EinstellungSetzen(t.Context(), "sprache", "en"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	// **Der Austausch geschieht in dem Augenblick, in dem geprüft wird.**
	var geprueftePfade []string

	anwendung.MitSicherungspruefer(
		func(_ context.Context, pfad string) (app.Sicherungsbefund, error) {
			geprueftePfade = append(geprueftePfade, pfad)

			// **Die Datei, die von außen kommt, wird ausgetauscht.** Wer sie
			// danach noch einmal öffnet, bekommt einen anderen Bestand.
			ersetzeDatei(t, echte, andere)

			return app.Sicherungsbefund{}, nil
		})

	if antwort := anwendung.SicherungZurueckspielen(t.Context(), echte, ""); antwort.Fehler != nil {
		t.Fatalf("zurückspielen: %+v", antwort.Fehler)
	}

	// **Geprüft wurde `de`.** Steht hier `en`, ist der ausgetauschte Bestand
	// angekommen, und die Bestätigung galt einer Datei, die niemand bekam.
	if got := merkmal(t, anwendung); got != "de" {
		t.Errorf("nach dem Zurückspielen steht %q — der ausgetauschte Bestand kam an", got)
	}

	// **Und geprüft wurde nicht die fremde Datei, sondern die Kopie daneben.**
	//
	// Ohne diese Zusicherung bliebe offen, *warum* der Lauf grün ist: ein
	// Programm, das die fremde Datei prüft und zufällig vor dem Austausch
	// kopiert, bestünde ihn ebenso — bis das Zeitfenster einmal anders liegt.
	//
	// **Der Pfad wird nicht benannt, sondern abgegrenzt** (#1162): seit die
	// Sicherung ein Archiv ist, liegt die Kopie im ausgepackten Verzeichnis
	// statt neben dem Bestand. Was die Zusage trägt, ist beides Mal dasselbe —
	// geprüft wurde **einmal**, und nicht das, was von außen kam.
	if len(geprueftePfade) != 1 || geprueftePfade[0] == echte {
		t.Errorf("geprüft wurden %v — erwartet war ein Pfad, und nicht %q",
			geprueftePfade, echte)
	}
}

// **Und die eingehende Kopie bleibt nicht liegen.**
//
// Sie ist Arbeitsvorrat und kein Stand. Bliebe sie stehen, wüchse das
// Datenverzeichnis bei jedem Zurückspielen um einen ganzen Bestand.
func TestDieEingehendeKopieBleibtNichtLiegen(t *testing.T) {
	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")

	anwendung := laufendeAnwendung(t, bestand)
	sicherung := gesichertMitSprache(t, anwendung, filepath.Join(ort, "sicherung.db"), "de")

	if antwort := anwendung.SicherungZurueckspielen(t.Context(), sicherung, ""); antwort.Fehler != nil {
		t.Fatalf("zurückspielen: %+v", antwort.Fehler)
	}

	if _, err := os.Stat(bestand + app.Eingehend); err == nil {
		t.Error("die eingehende Kopie liegt noch da")
	}
}

// gesichertMitSprache setzt die Sprache und schreibt eine Sicherung dahin.
func gesichertMitSprache(t *testing.T, anwendung *app.App, ziel, sprache string) string {
	t.Helper()

	if antwort := anwendung.EinstellungSetzen(t.Context(), "sprache", sprache); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return ziel, nil
	})

	if antwort := anwendung.BestandSichern(t.Context(), "", false); antwort.Fehler != nil {
		t.Fatalf("sichern nach %s: %+v", ziel, antwort.Fehler)
	}

	return ziel
}

// ersetzeDatei legt den Inhalt von `quelle` an die Stelle von `ziel`.
func ersetzeDatei(t *testing.T, ziel, quelle string) {
	t.Helper()

	inhalt, err := os.ReadFile(quelle) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("die andere Sicherung lesen: %v", err)
	}

	// #nosec G703 -- beide Pfade kommen aus `t.TempDir()` dieses Laufs.
	if fehler := os.WriteFile(ziel, inhalt, 0o600); fehler != nil {
		t.Fatalf("die Sicherung austauschen: %v", fehler)
	}
}

// **Scheitert schon das Kopieren, ist nichts geschehen.**
//
// Der Bestand ist noch offen, die Anwendung läuft weiter, und die Meldung
// nennt die Datei. Das ist die frühestmögliche Stelle zum Scheitern — und
// deshalb die beste (#964).
func TestEineUnlesbareSicherungKommtGarNichtErstAn(t *testing.T) {
	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")

	anwendung := laufendeAnwendung(t, bestand)

	if antwort := anwendung.EinstellungSetzen(t.Context(), "sprache", "de"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	antwort := anwendung.SicherungZurueckspielen(t.Context(), filepath.Join(ort, "gibt-es-nicht.db"), "")
	if antwort.Fehler == nil {
		t.Fatal("eine fehlende Datei kam durch")
	}

	if antwort.Ersetzt {
		t.Error("die Antwort sagt, es sei ersetzt worden")
	}

	// **Und die Anwendung läuft weiter** — mit dem Stand von vorher.
	if got := merkmal(t, anwendung); got != "de" {
		t.Errorf("nach dem gescheiterten Zurückspielen steht %q", got)
	}
}

// **Scheitert der Tausch, läuft die Anwendung mit dem alten Stand weiter.**
//
// Der Fall wird hergestellt, indem am Platz der Rückspielsicherung ein
// **nicht leeres Verzeichnis** liegt: der bisherige Bestand lässt sich dorthin
// nicht beiseitelegen, und damit ist der erste bewegende Schritt schon
// blockiert. Geprüft und kopiert war zu diesem Zeitpunkt alles.
func TestEinGescheiterterTauschLaesstDenAltenStandStehen(t *testing.T) {
	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")

	anwendung := laufendeAnwendung(t, bestand)
	sicherung := gesichertMitSprache(t, anwendung, filepath.Join(ort, "sicherung.db"), "de")

	if antwort := anwendung.EinstellungSetzen(t.Context(), "sprache", "en"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	if err := os.MkdirAll(
		filepath.Join(ort, app.Rueckspielsicherung, "inhalt"), 0o700); err != nil {
		t.Fatalf("das Verzeichnis anlegen: %v", err)
	}

	antwort := anwendung.SicherungZurueckspielen(t.Context(), sicherung, "")
	if antwort.Fehler == nil {
		t.Fatal("der Tausch kam durch, obwohl er nicht möglich war")
	}

	if got := merkmal(t, anwendung); got != "en" {
		t.Errorf("nach dem gescheiterten Tausch steht %q, erwartet war der alte Stand", got)
	}

	// **Und die eingehende Kopie liegt auch dann nicht mehr da.**
	if _, err := os.Stat(bestand + app.Eingehend); err == nil {
		t.Error("die eingehende Kopie liegt nach dem Fehler noch da")
	}
}
