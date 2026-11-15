package app_test

// **Das Abzeichen eines Profils am Fenster** — Werkstück 9 von Phase 4.12
// (#1351),
// Punkt 13.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ein gesetztes Abzeichen kommt als Bild zurück — je Profil.**
//
// **Und es liegt neben dem Profilbild, nicht darin**: die Antwort trägt zwei
// Tabellen, eine je Kennungsart. Ein Abzeichen unter einer Nutzerkennung wäre
// beim Zeichnen nicht vom Gesicht zu unterscheiden.
func TestEinAbzeichenKommtAlsBildZurueck(t *testing.T) {
	anwendung := mitBilderort(t)

	gesetzt := anwendung.ProfilbadgeSetzen(t.Context(), "1", "1", bilddatei(t, 20, 20))
	if gesetzt.Fehler != nil {
		t.Fatalf("%+v", gesetzt.Fehler)
	}

	// **Kennung und Inhalt im Namen** — der Inhalt, damit ein abgewiesenes
	// Schreiben das alte Abzeichen nicht überschreibt (Zweitblick über Phase
	// 6.1, Befund 3).
	if !strings.HasPrefix(gesetzt.Profil.Badge, "profil-1-") ||
		!strings.HasSuffix(gesetzt.Profil.Badge, ".png") {
		t.Errorf("der Name ist %q", gesetzt.Profil.Badge)
	}

	antwort := anwendung.Nutzerbilder(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if !strings.HasPrefix(antwort.Badges["1"], "data:image/png;base64,") {
		t.Errorf("der Eintrag ist %q", antwort.Badges["1"])
	}

	// **Und das Gesicht bleibt leer** — es wurde keines gesetzt.
	if len(antwort.Bilder) != 0 {
		t.Errorf("es kam ein Profilbild: %v", antwort.Bilder)
	}
}

// **Es lässt sich entfernen — und die Angaben bleiben stehen.**
//
// **Die zweite Hälfte ist der Punkt**: `ProfilAendern` **setzt** und
// verschmilzt nicht. Wer nur das Abzeichen schriebe und die übrigen Felder
// leer ließe, löschte Subslogan und beide Beschreibungen mit.
func TestEinAbzeichenLaesstSichEntfernenUndDieAngabenBleiben(t *testing.T) {
	anwendung := mitBilderort(t)

	beschrieben := anwendung.ProfilAendern(t.Context(), app.ProfilangabenDTO{
		ID: "1", Name: "Mein Profil", Subslogan: "Nur für Turniere", Fassung: "1",
	})
	if beschrieben.Fehler != nil {
		t.Fatalf("%+v", beschrieben.Fehler)
	}

	gesetzt := anwendung.ProfilbadgeSetzen(
		t.Context(), "1", beschrieben.Profil.Fassung, bilddatei(t, 20, 20))
	if gesetzt.Fehler != nil {
		t.Fatalf("%+v", gesetzt.Fehler)
	}

	entfernt := anwendung.ProfilbadgeEntfernen(
		t.Context(), "1", gesetzt.Profil.Fassung)
	if entfernt.Fehler != nil {
		t.Fatalf("%+v", entfernt.Fehler)
	}

	if entfernt.Profil.Badge != "" {
		t.Errorf("der Bestand führt noch %q", entfernt.Profil.Badge)
	}

	if entfernt.Profil.Subslogan != "Nur für Turniere" {
		t.Errorf("der Subslogan ging mit: %q", entfernt.Profil.Subslogan)
	}
}

// **Kennung, Fassung und Pfad werden geprüft.**
func TestDasAbzeichenWeistKaputteAngabenAb(t *testing.T) {
	anwendung := mitBilderort(t)
	pfad := bilddatei(t, 10, 10)

	faelle := map[string]struct{ profilID, fassung, pfad string }{
		"keine Kennung":      {"keine Zahl", "1", pfad},
		"keine Fassung":      {"1", "später", pfad},
		"kein Pfad":          {"1", "1", ""},
		"kein lesbares Bild": {"1", "1", keinBild(t)},
		"veraltete Fassung":  {"1", "4711", pfad},
		"fehlendes Profil":   {"4711", "1", pfad},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			if antwort := anwendung.ProfilbadgeSetzen(
				t.Context(), fall.profilID, fall.fassung, fall.pfad,
			); antwort.Fehler == nil {
				t.Errorf("die Angabe kam durch: %+v", antwort.Profil)
			}
		})
	}

	// **Und das Entfernen prüft dieselben drei.**
	for name, fall := range map[string]struct{ profilID, fassung string }{
		"keine Kennung":     {"keine Zahl", "1"},
		"keine Fassung":     {"1", "später"},
		"fehlendes Profil":  {"4711", "1"},
		"veraltete Fassung": {"1", "4711"},
	} {
		t.Run("entfernen: "+name, func(t *testing.T) {
			if antwort := anwendung.ProfilbadgeEntfernen(
				t.Context(), fall.profilID, fall.fassung); antwort.Fehler == nil {
				t.Error("die Angabe kam durch")
			}
		})
	}
}

// **Ohne Bilderort scheitern beide Wege.**
func TestOhneBilderortScheitertDasAbzeichen(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	anwendung.MitBilderort(func() (string, error) {
		return "", errors.New("kein Benutzerverzeichnis")
	})

	if antwort := anwendung.ProfilbadgeSetzen(
		t.Context(), "1", "1", bilddatei(t, 10, 10)); antwort.Fehler == nil {
		t.Error("das Setzen ging ohne Ort")
	}

	if antwort := anwendung.ProfilbadgeEntfernen(
		t.Context(), "1", "1"); antwort.Fehler == nil {
		t.Error("das Entfernen ging ohne Ort")
	}
}

// **Der Dialog gibt den Pfad — und ein Abbruch ist kein Fehler.**
func TestDerAbzeichendialogLegtAbOderBrichtAb(t *testing.T) {
	anwendung := mitBilderort(t)
	pfad := bilddatei(t, 20, 20)

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return pfad, nil
	})

	gesetzt := anwendung.ProfilbadgeWaehlen(t.Context(), "1", "1")
	if gesetzt.Fehler != nil {
		t.Fatalf("%+v", gesetzt.Fehler)
	}

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", nil
	})

	abgebrochen := anwendung.ProfilbadgeWaehlen(
		t.Context(), "1", gesetzt.Profil.Fassung)
	if abgebrochen.Fehler != nil || abgebrochen.Profil != nil {
		t.Errorf("%+v — %+v", abgebrochen.Fehler, abgebrochen.Profil)
	}

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", errors.New("kein Fenster")
	})

	if antwort := anwendung.ProfilbadgeWaehlen(
		t.Context(), "1", gesetzt.Profil.Fassung); antwort.Fehler == nil {
		t.Error("der Fehler ging verloren")
	}
}

// **Ein vorhandenes Abzeichen überlebt einen Konflikt.**
//
// **Die Gegenprobe zu „veraltete Fassung"** in [TestDasAbzeichenWeistKaputteAngabenAb]:
// dort lag keines, und die abgelegte Datei kommt weg. Hier lag eines, und es
// bleibt — der Eintrag lautet vorher wie nachher `profil-1.png`.
func TestEinVorhandenesAbzeichenUeberlebtEinenKonflikt(t *testing.T) {
	anwendung := mitBilderort(t)

	gesetzt := anwendung.ProfilbadgeSetzen(t.Context(), "1", "1", bilddatei(t, 20, 20))
	if gesetzt.Fehler != nil {
		t.Fatalf("%+v", gesetzt.Fehler)
	}

	if antwort := anwendung.ProfilbadgeSetzen(
		t.Context(), "1", "1", bilddatei(t, 20, 20)); antwort.Fehler == nil {
		t.Fatal("die veraltete Fassung kam durch")
	}

	if bilder := anwendung.Nutzerbilder(t.Context()); bilder.Badges["1"] == "" {
		t.Error("das vorhandene Abzeichen wurde mitgenommen")
	}
}

// **Eine gesperrte Ablage und ein gescheitertes Wegnehmen melden sich.**
func TestDasAbzeichenMeldetWasAnDerPlatteScheitert(t *testing.T) {
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

	if antwort := anwendung.ProfilbadgeSetzen(
		t.Context(), "1", "1", bilddatei(t, 10, 10)); antwort.Fehler == nil {
		t.Error("das gesperrte Verzeichnis kam durch")
	}
}

// **Ein Wegräumen, das scheitert, nimmt das Entfernen nicht zurück** (§7,
// Regel 3) — seit dem Zweitblick über Phase 6.1 steht der Bestand vorn.
//
// Hier stand „ein gescheitertes Wegnehmen meldet sich": die Datei ging zuerst,
// und ihr Scheitern hielt das Entfernen auf. Jetzt ist das Abzeichen im
// Bestand fort, die Antwort sagt es, und die Datei, die liegen bleibt, zeigt
// niemand mehr an.
func TestEinGescheitertesWegraeumenNimmtDasEntfernenNichtZurueck(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	wurzel := t.TempDir()
	anwendung.MitBilderort(func() (string, error) { return wurzel, nil })

	gesetzt := anwendung.ProfilbadgeSetzen(t.Context(), "1", "1", bilddatei(t, 10, 10))
	if gesetzt.Fehler != nil {
		t.Fatalf("%+v", gesetzt.Fehler)
	}

	// **Ein Verzeichnis an der Stelle der Datei** — `os.Remove` weist es ab.
	datei := filepath.Join(wurzel, "nutzer", gesetzt.Profil.Badge)
	if err := os.Remove(datei); err != nil {
		t.Fatalf("die Datei wegnehmen: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(datei, "inhalt"), 0o750); err != nil {
		t.Fatalf("die Sperre legen: %v", err)
	}

	entfernt := anwendung.ProfilbadgeEntfernen(t.Context(), "1", gesetzt.Profil.Fassung)
	if entfernt.Fehler != nil {
		t.Fatalf("das Entfernen ist geschrieben und meldet trotzdem %+v", entfernt.Fehler)
	}

	if bilder := anwendung.Nutzerbilder(t.Context()); bilder.Badges["1"] != "" {
		t.Error("das Abzeichen steht noch im Bestand")
	}
}

// **Eine schweigende Platte erreicht beide Wege.**
func TestDasAbzeichenReichtEinenLesefehlerDurch(t *testing.T) {
	anwendung := mitBilderort(t)
	anwendung.MitNutzern(&nutzerfuehrer{fehler: errors.New("die Platte schweigt")})

	if antwort := anwendung.ProfilbadgeSetzen(
		t.Context(), "1", "1", bilddatei(t, 10, 10)); antwort.Fehler == nil {
		t.Error("das Setzen glückte trotzdem")
	}

	if antwort := anwendung.ProfilbadgeEntfernen(
		t.Context(), "1", "1"); antwort.Fehler == nil {
		t.Error("das Entfernen glückte trotzdem")
	}
}
