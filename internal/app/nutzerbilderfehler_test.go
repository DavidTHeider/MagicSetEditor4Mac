package app_test

// **Die Fehlerwege des Profilbildes** — Werkstück 8 von Phase 4.12
// (#1351).
//
// **Eine eigene Datei neben `nutzerbilder_test.go`**, damit beide unter der
// Zeilengrenze aus §2 bleiben. Der Schnitt folgt der Frage: dort steht, was
// glückt, hier, was schiefgeht — und dass dabei kein halber Stand bleibt.

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Eine schweigende Platte erreicht jede der drei Bindungen.**
func TestDasProfilbildReichtEinenLesefehlerDurch(t *testing.T) {
	anwendung := mitBilderort(t)
	anwendung.MitNutzern(&nutzerfuehrer{fehler: errors.New("die Platte schweigt")})

	if antwort := anwendung.Nutzerbilder(t.Context()); antwort.Fehler == nil {
		t.Error("die Tabelle kam trotzdem")
	}

	if antwort := anwendung.NutzerbildSetzen(
		t.Context(), "1", "1", bilddatei(t, 10, 10)); antwort.Fehler == nil {
		t.Error("das Setzen glückte trotzdem")
	}

	if antwort := anwendung.NutzerbildEntfernen(
		t.Context(), "1", "1"); antwort.Fehler == nil {
		t.Error("das Entfernen glückte trotzdem")
	}
}

// **Ein Nutzer, den der Bestand nicht führt, ist „nicht gefunden".**
//
// **Und nicht „ungültig".** Wer die falsche Kennung schickt, sucht am
// falschen Ort; wer eine gültige schickt, deren Nutzer gelöscht wurde, lädt
// neu. Die Oberfläche tut dafür etwas anderes (§8).
func TestEinFehlenderNutzerHatKeinProfilbild(t *testing.T) {
	anwendung := mitBilderort(t)
	anwendung.MitNutzern(&nutzerfuehrer{})

	antwort := anwendung.NutzerbildSetzen(t.Context(), "1", "1", bilddatei(t, 10, 10))
	if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeNichtGefunden {
		t.Errorf("%+v", antwort.Fehler)
	}

	entfernt := anwendung.NutzerbildEntfernen(t.Context(), "1", "1")
	if entfernt.Fehler == nil || entfernt.Fehler.Code != app.CodeNichtGefunden {
		t.Errorf("%+v", entfernt.Fehler)
	}
}

// **Eine Ablage, in die sich nicht schreiben lässt, meldet sich.**
func TestEineGesperrteAblageMeldetSichBeimProfilbild(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	wurzel := t.TempDir()
	anwendung.MitBilderort(func() (string, error) { return wurzel, nil })

	//nolint:gosec // absichtlich: lesbar, nicht schreibbar
	if err := os.MkdirAll(filepath.Join(wurzel, "nutzer"), 0o500); err != nil {
		t.Fatalf("die Sperre legen: %v", err)
	}

	t.Cleanup(func() {
		//nolint:gosec // ein Verzeichnis braucht x, sonst lässt es sich nicht räumen
		if fehler := os.Chmod(filepath.Join(wurzel, "nutzer"), 0o750); fehler != nil {
			t.Errorf("Rechte zurücksetzen: %v", fehler)
		}
	})

	if antwort := anwendung.NutzerbildSetzen(
		t.Context(), "1", "1", bilddatei(t, 10, 10)); antwort.Fehler == nil {
		t.Error("das gesperrte Verzeichnis kam durch")
	}
}

// **Ein Wegräumen, das scheitert, nimmt das Entfernen nicht zurück** (§7,
// Regel 3).
//
// **Hier stand die umgekehrte Zusage** — „ein Wegnehmen, das scheitert, lässt
// den Bestand in Ruhe" —, und sie gehörte zur Reihenfolge Datei vor Bestand.
// Die hat der Zweitblick über Phase 6.1 umgedreht (Befund 3): jetzt steht der
// Bestand vorn, und eine Datei, die sich danach nicht wegräumen lässt, bleibt
// ohne Verweis liegen und steht im Protokoll.
func TestEinGescheitertesWegraeumenNimmtDasEntfernenDesBildesNichtZurueck(t *testing.T) {
	anwendung, ort := mitOrt(t)

	gesetzt := anwendung.NutzerbildSetzen(t.Context(), "1", "1", bilddatei(t, 10, 10))
	if gesetzt.Fehler != nil {
		t.Fatalf("%+v", gesetzt.Fehler)
	}

	// **Ein Verzeichnis an der Stelle der Datei** — `os.Remove` weist es ab.
	datei := filepath.Join(ort, gesetzt.Nutzer.Bild)
	if err := os.Remove(datei); err != nil {
		t.Fatalf("die Datei wegnehmen: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(datei, "inhalt"), 0o750); err != nil {
		t.Fatalf("die Sperre legen: %v", err)
	}

	entfernt := anwendung.NutzerbildEntfernen(t.Context(), "1", gesetzt.Nutzer.Fassung)
	if entfernt.Fehler != nil {
		t.Fatalf("das Entfernen ist geschrieben und meldet trotzdem %+v", entfernt.Fehler)
	}

	if bilder := anwendung.Nutzerbilder(t.Context()); bilder.Bilder["1"] != "" {
		t.Error("das Bild steht noch im Bestand")
	}
}

// **Eine veraltete Fassung weist auch das Entfernen ab.**
func TestEineVeralteteFassungWeistDasEntfernenAb(t *testing.T) {
	anwendung := mitBilderort(t)

	antwort := anwendung.NutzerbildEntfernen(t.Context(), "1", "4711")
	if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeKonflikt {
		t.Errorf("%+v", antwort.Fehler)
	}
}

// **Ohne Bilderort scheitert auch das Entfernen.**
func TestOhneBilderortScheitertDasEntfernen(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	anwendung.MitBilderort(func() (string, error) {
		return "", errors.New("kein Benutzerverzeichnis")
	})

	if antwort := anwendung.NutzerbildEntfernen(t.Context(), "1", "1"); antwort.Fehler == nil {
		t.Error("es ging ohne Ort")
	}
}

// **Wer schon ein Bild hatte, behält es, wenn das Schreiben abgewiesen
// wird.**
//
// # Warum das die Gegenprobe zu [TestEineVeralteteFassungIstEinKonflikt] ist
//
// Dort lag **keines**, und die abgelegte Datei kommt weg — sie hätte keinen
// Verweis. Hier lag **eines**, und es bleibt: der Eintrag im Bestand lautet
// vorher wie nachher `nutzer-1.png`, und wer nach einem Konflikt wegnähme,
// löschte das Bild, das jemand schon hatte.
func TestEinVorhandenesProfilbildUeberlebtEinenKonflikt(t *testing.T) {
	anwendung := mitBilderort(t)

	gesetzt := anwendung.NutzerbildSetzen(t.Context(), "1", "1", bilddatei(t, 40, 40))
	if gesetzt.Fehler != nil {
		t.Fatalf("%+v", gesetzt.Fehler)
	}

	if antwort := anwendung.NutzerbildSetzen(
		t.Context(), "1", "1", bilddatei(t, 40, 40)); antwort.Fehler == nil {
		t.Fatal("die veraltete Fassung kam durch")
	}

	if bilder := anwendung.Nutzerbilder(t.Context()); bilder.Bilder["1"] == "" {
		t.Error("das vorhandene Bild wurde mitgenommen")
	}
}
