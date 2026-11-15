package app_test

// **Das Profilbild am Fenster** — Werkstück 8 von Phase 4.12
// (#1351).

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// bilddatei schreibt ein PNG und gibt seinen Pfad.
func bilddatei(t *testing.T, breite, hoehe int) string {
	t.Helper()

	bild := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))
	bild.SetNRGBA(0, 0, color.NRGBA{R: 200, A: 255})

	pfad := filepath.Join(t.TempDir(), "gesicht.png")

	datei, err := os.Create(pfad) // #nosec G304 -- ein Pfad aus t.TempDir()
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if fehler := png.Encode(datei, bild); fehler != nil {
		t.Fatalf("kodieren: %v", fehler)
	}

	if fehler := datei.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	return pfad
}

// mitBilderort gibt eine Anwendung, deren Bilder in einem eigenen
// Verzeichnis liegen.
//
// **Nicht im echten Zuhause** (§3): ein Test, der dorthin schreibt, sieht den
// Stand dieses Rechners und hinterlässt seinen eigenen.
func mitBilderort(t *testing.T) *app.App {
	t.Helper()

	anwendung := angelaufeneAnwendung(t)
	wurzel := t.TempDir()
	anwendung.MitBilderort(func() (string, error) { return wurzel, nil })

	return anwendung
}

// **Ein gesetztes Profilbild kommt als Bild zurück.**
//
// **Das ist die ganze Zusage dieses Werkstücks**: ein Dateiname im Bestand
// ist kein Bild — sichtbar wird er erst, wenn ihn jemand als Daten-URI
// abholen kann.
func TestEinProfilbildKommtAlsBildZurueck(t *testing.T) {
	anwendung := mitBilderort(t)

	gesetzt := anwendung.NutzerbildSetzen(t.Context(), "1", "1", bilddatei(t, 40, 40))
	if gesetzt.Fehler != nil {
		t.Fatalf("%+v", gesetzt.Fehler)
	}

	if gesetzt.Nutzer == nil || gesetzt.Nutzer.Bild == "" {
		t.Fatalf("Nutzer %+v", gesetzt.Nutzer)
	}

	// **Die Fassung zählt mit** — sonst schriebe der nächste Ruf gegen einen
	// Stand, den es nicht mehr gibt (§7, Regel 2).
	if gesetzt.Nutzer.Fassung != "2" {
		t.Errorf("die Fassung ist %q", gesetzt.Nutzer.Fassung)
	}

	antwort := anwendung.Nutzerbilder(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if !strings.HasPrefix(antwort.Bilder["1"], "data:image/png;base64,") {
		t.Errorf("der Eintrag ist %q", antwort.Bilder["1"])
	}
}

// **Wer kein Bild hat, steht nicht in der Tabelle.**
//
// **Eine Tabelle und keine Liste** — dieselbe Bauform wie [App.Kartenbilder]:
// ein leerer Eintrag wäre die teurere Art, „keines" zu sagen.
func TestOhneProfilbildStehtNiemandInDerTabelle(t *testing.T) {
	antwort := mitBilderort(t).Nutzerbilder(t.Context())

	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Bilder == nil {
		t.Fatal("die Tabelle ist nil — eine leere Antwort wäre keine")
	}

	if len(antwort.Bilder) != 0 {
		t.Errorf("es stehen %d Einträge darin", len(antwort.Bilder))
	}
}

// **Ein Profilbild lässt sich entfernen — und die Datei geht mit.**
func TestEinProfilbildLaesstSichAmFensterEntfernen(t *testing.T) {
	anwendung := mitBilderort(t)

	gesetzt := anwendung.NutzerbildSetzen(t.Context(), "1", "1", bilddatei(t, 40, 40))
	if gesetzt.Fehler != nil {
		t.Fatalf("%+v", gesetzt.Fehler)
	}

	entfernt := anwendung.NutzerbildEntfernen(t.Context(), "1", gesetzt.Nutzer.Fassung)
	if entfernt.Fehler != nil {
		t.Fatalf("%+v", entfernt.Fehler)
	}

	if entfernt.Nutzer.Bild != "" {
		t.Errorf("der Bestand führt noch %q", entfernt.Nutzer.Bild)
	}

	if antwort := anwendung.Nutzerbilder(t.Context()); len(antwort.Bilder) != 0 {
		t.Errorf("es steht noch etwas in der Tabelle: %v", antwort.Bilder)
	}
}

// **Kennung und Fassung werden geprüft, und der Pfad auch.**
func TestDasProfilbildWeistKaputteAngabenAb(t *testing.T) {
	anwendung := mitBilderort(t)
	pfad := bilddatei(t, 10, 10)

	faelle := map[string]struct{ nutzerID, fassung, pfad string }{
		"keine Kennung":      {nutzerID: "keine Zahl", fassung: "1", pfad: pfad},
		"keine Fassung":      {nutzerID: "1", fassung: "später", pfad: pfad},
		"kein Pfad":          {nutzerID: "1", fassung: "1", pfad: ""},
		"nur Leerzeichen":    {nutzerID: "1", fassung: "1", pfad: "   "},
		"die Datei fehlt":    {nutzerID: "1", fassung: "1", pfad: "/gibt/es/nicht.png"},
		"kein lesbares Bild": {nutzerID: "1", fassung: "1", pfad: keinBild(t)},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			antwort := anwendung.NutzerbildSetzen(
				t.Context(), fall.nutzerID, fall.fassung, fall.pfad)
			if antwort.Fehler == nil {
				t.Errorf("die Angabe kam durch: %+v", antwort.Nutzer)
			}
		})
	}
}

// keinBild schreibt eine Datei, die keine Bilddatei ist.
func keinBild(t *testing.T) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "gesicht.png")
	if err := os.WriteFile(pfad, []byte("kein Bild"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return pfad
}

// **Auch das Entfernen prüft Kennung und Fassung.**
func TestDasEntfernenWeistKaputteAngabenAb(t *testing.T) {
	anwendung := mitBilderort(t)

	if antwort := anwendung.NutzerbildEntfernen(
		t.Context(), "keine Zahl", "1"); antwort.Fehler == nil {
		t.Error("die Kennung kam durch")
	}

	if antwort := anwendung.NutzerbildEntfernen(
		t.Context(), "1", "später"); antwort.Fehler == nil {
		t.Error("die Fassung kam durch")
	}
}

// **Eine veraltete Fassung ist ein Konflikt — und das Bild bleibt liegen.**
//
// **Die zweite Hälfte ist der Punkt.** Ein abgewiesenes Schreiben, das
// trotzdem die Datei getauscht hätte, wäre ein halber Stand (§10).
func TestEineVeralteteFassungIstEinKonflikt(t *testing.T) {
	anwendung := mitBilderort(t)

	antwort := anwendung.NutzerbildSetzen(t.Context(), "1", "4711", bilddatei(t, 40, 40))
	if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeKonflikt {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if bilder := anwendung.Nutzerbilder(t.Context()); len(bilder.Bilder) != 0 {
		t.Errorf("die Datei blieb liegen: %v", bilder.Bilder)
	}
}

// **Ohne Bilderort scheitert das Setzen.**
func TestOhneBilderortScheitertDasProfilbild(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	anwendung.MitBilderort(func() (string, error) {
		return "", errors.New("kein Benutzerverzeichnis")
	})

	if antwort := anwendung.NutzerbildSetzen(
		t.Context(), "1", "1", bilddatei(t, 10, 10)); antwort.Fehler == nil {
		t.Error("es ging ohne Ort")
	}

	if antwort := anwendung.Nutzerbilder(t.Context()); antwort.Fehler == nil {
		t.Error("die Tabelle kam ohne Ort")
	}
}
