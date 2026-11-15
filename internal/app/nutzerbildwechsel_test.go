package app_test

// **Ein abgewiesenes Schreiben lässt das Bild, wie es war** — Befund 3 des
// Zweitblicks über Phase 6.1.
//
// # Der Befund
//
// Profilbild und Abzeichen lagen unter einem **festen** Namen
// (`nutzer-7.png`), und die Datei wurde geschrieben, **bevor** der Bestand
// die Fassung prüfte. Wurde das Schreiben abgewiesen, stand trotzdem das neue
// Bild da — unter dem alten Verweis, mit einer Fehlermeldung darüber. Das
// Entfernen löschte die Datei vor der Prüfung: abgewiesen, und das Bild war
// trotzdem fort.
//
// # Was hier gilt
//
// **Geprüft wird am Inhalt**, nicht daran, dass „irgendein" Bild da ist — der
// alte Lauf daneben war grün, weil er nur das fragte.

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// farbbild legt ein einfarbiges Bild ab und gibt seinen Pfad.
func farbbild(t *testing.T, farbe color.NRGBA) string {
	t.Helper()

	bild := image.NewNRGBA(image.Rect(0, 0, 12, 12))
	for x := range 12 {
		for y := range 12 {
			bild.SetNRGBA(x, y, farbe)
		}
	}

	pfad := filepath.Join(t.TempDir(), "farbe.png")

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

var (
	rot  = color.NRGBA{R: 220, A: 255}
	blau = color.NRGBA{B: 220, A: 255}
)

// mitOrt gibt eine Anwendung samt dem Verzeichnis ihrer Profilbilder.
func mitOrt(t *testing.T) (anwendung *app.App, ort string) {
	t.Helper()

	anwendung = angelaufeneAnwendung(t)
	wurzel := t.TempDir()
	anwendung.MitBilderort(func() (string, error) { return wurzel, nil })

	return anwendung, filepath.Join(wurzel, "nutzer")
}

// dateienIn zählt, was im Verzeichnis liegt.
func dateienIn(t *testing.T, ort string) []string {
	t.Helper()

	eintraege, err := os.ReadDir(ort)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("lesen: %v", err)
	}

	namen := make([]string, 0, len(eintraege))
	for _, eintrag := range eintraege {
		namen = append(namen, eintrag.Name())
	}

	return namen
}

func TestEinAbgewiesenesProfilbildLaesstDasAlteStehen(t *testing.T) {
	anwendung, ort := mitOrt(t)

	if gesetzt := anwendung.NutzerbildSetzen(t.Context(), "1", "1", farbbild(t, rot)); gesetzt.Fehler != nil {
		t.Fatalf("%+v", gesetzt.Fehler)
	}

	vorher := anwendung.Nutzerbilder(t.Context()).Bilder["1"]

	// **Mit der Fassung von vorhin** — sie ist seit dem Setzen überholt.
	abgewiesen := anwendung.NutzerbildSetzen(t.Context(), "1", "1", farbbild(t, blau))
	if abgewiesen.Fehler == nil || abgewiesen.Fehler.Code != app.CodeKonflikt {
		t.Fatalf("erwartet war ein Konflikt, gekommen ist %+v", abgewiesen.Fehler)
	}

	if nachher := anwendung.Nutzerbilder(t.Context()).Bilder["1"]; nachher != vorher {
		t.Error("das Bild hat sich geändert, obwohl das Schreiben abgewiesen wurde")
	}

	if namen := dateienIn(t, ort); len(namen) != 1 {
		t.Errorf("im Ordner liegen %v — erwartet war das eine Bild", namen)
	}
}

func TestEinAbgewiesenesEntfernenLaesstDasBildStehen(t *testing.T) {
	anwendung, _ := mitOrt(t)

	if gesetzt := anwendung.NutzerbildSetzen(t.Context(), "1", "1", farbbild(t, rot)); gesetzt.Fehler != nil {
		t.Fatalf("%+v", gesetzt.Fehler)
	}

	vorher := anwendung.Nutzerbilder(t.Context()).Bilder["1"]

	if entfernt := anwendung.NutzerbildEntfernen(t.Context(), "1", "1"); entfernt.Fehler == nil {
		t.Fatal("die überholte Fassung kam durch")
	}

	if nachher := anwendung.Nutzerbilder(t.Context()).Bilder["1"]; nachher != vorher {
		t.Error("das Bild ist fort, obwohl das Entfernen abgewiesen wurde")
	}
}

// **Die Gegenprobe**: mit der richtigen Fassung wechselt das Bild, und das
// alte bleibt nicht als Datei ohne Verweis liegen.
func TestEinNeuesProfilbildErsetztDasAlte(t *testing.T) {
	anwendung, ort := mitOrt(t)

	erstes := anwendung.NutzerbildSetzen(t.Context(), "1", "1", farbbild(t, rot))
	if erstes.Fehler != nil {
		t.Fatalf("%+v", erstes.Fehler)
	}

	vorher := anwendung.Nutzerbilder(t.Context()).Bilder["1"]

	zweites := anwendung.NutzerbildSetzen(t.Context(), "1", erstes.Nutzer.Fassung, farbbild(t, blau))
	if zweites.Fehler != nil {
		t.Fatalf("%+v", zweites.Fehler)
	}

	if nachher := anwendung.Nutzerbilder(t.Context()).Bilder["1"]; nachher == vorher {
		t.Error("das Bild hat sich nicht geändert")
	}

	if namen := dateienIn(t, ort); len(namen) != 1 {
		t.Errorf("im Ordner liegen %v — das alte Bild blieb liegen", namen)
	}

	entfernt := anwendung.NutzerbildEntfernen(t.Context(), "1", zweites.Nutzer.Fassung)
	if entfernt.Fehler != nil {
		t.Fatalf("%+v", entfernt.Fehler)
	}

	if namen := dateienIn(t, ort); len(namen) != 0 {
		t.Errorf("nach dem Entfernen liegen %v", namen)
	}
}

func TestEinAbgewiesenesAbzeichenLaesstDasAlteStehen(t *testing.T) {
	anwendung, ort := mitOrt(t)

	erstes := anwendung.ProfilbadgeSetzen(t.Context(), "1", "1", farbbild(t, rot))
	if erstes.Fehler != nil {
		t.Fatalf("%+v", erstes.Fehler)
	}

	vorher := anwendung.Nutzerbilder(t.Context()).Badges["1"]

	if abgewiesen := anwendung.ProfilbadgeSetzen(t.Context(), "1", "1", farbbild(t, blau)); abgewiesen.Fehler == nil {
		t.Fatal("die überholte Fassung kam durch")
	}

	if entfernt := anwendung.ProfilbadgeEntfernen(t.Context(), "1", "1"); entfernt.Fehler == nil {
		t.Fatal("das Entfernen mit überholter Fassung kam durch")
	}

	if nachher := anwendung.Nutzerbilder(t.Context()).Badges["1"]; nachher != vorher {
		t.Error("das Abzeichen hat sich geändert, obwohl beides abgewiesen wurde")
	}

	zweites := anwendung.ProfilbadgeSetzen(t.Context(), "1", erstes.Profil.Fassung, farbbild(t, blau))
	if zweites.Fehler != nil {
		t.Fatalf("%+v", zweites.Fehler)
	}

	if namen := dateienIn(t, ort); len(namen) != 1 {
		t.Errorf("im Ordner liegen %v — das alte Abzeichen blieb liegen", namen)
	}
}

// wettlauffuehrer ist der Verlierer eines Wettlaufs: er hat den Stand von
// vorhin gelesen, und sein Schreiben kommt zu spät.
type wettlauffuehrer struct {
	nutzerfuehrer
	vorher string
}

func (f *wettlauffuehrer) Uebersicht(context.Context) ([]service.Nutzerzeile, error) {
	return []service.Nutzerzeile{{Nutzer: domain.Nutzer{ID: 1, Name: "Anna", Bild: f.vorher, Fassung: 1}}}, nil
}

func (f *wettlauffuehrer) NutzerAendern(context.Context, domain.Nutzer) (domain.Nutzer, error) {
	return domain.Nutzer{}, fmt.Errorf("%w: der Stand ist überholt", domain.ErrUeberholt)
}

// **Zwei gleichzeitige Wechsel auf dasselbe Bild** — Befund 1 des Nachlaufs
// über Phase 6.1.
//
// Der Gewinner hat sein Bild abgelegt und den Verweis gewechselt; der
// Verlierer legt **dasselbe** Bild ab, scheitert an der Fassung und räumt
// sein neues Bild weg. Hieß das Bild nach seinem Inhalt, räumte er damit die
// Datei des Gewinners weg — der Verweis zeigte danach ins Leere. Nachgestellt
// ohne echte Gleichzeitigkeit: der Gewinner ist schon durch, der Verlierer
// kommt mit dem Stand von vorhin.
func TestDerVerliererEinesWettlaufsRaeumtNurSeinBildWeg(t *testing.T) {
	anwendung, ort := mitOrt(t)
	gesicht := farbbild(t, rot)

	ablage := service.NeueBildablage(filepath.Dir(ort))

	gelesen, err := os.Open(gesicht) // #nosec G304 -- ein Pfad aus t.TempDir()
	if err != nil {
		t.Fatal(err)
	}

	bild, _, err := image.Decode(gelesen)
	_ = gelesen.Close() //nolint:errcheck // nur gelesen

	if err != nil {
		t.Fatal(err)
	}

	gewinner, err := ablage.NutzerbildAblegen(1, bild)
	if err != nil {
		t.Fatal(err)
	}

	anwendung.MitNutzern(&wettlauffuehrer{vorher: "nutzer-1-vorher.png"})

	if antwort := anwendung.NutzerbildSetzen(t.Context(), "1", "1", gesicht); antwort.Fehler == nil {
		t.Fatal("der Verlierer kam durch")
	}

	if _, err := os.Stat(filepath.Join(ort, gewinner)); err != nil {
		t.Errorf("das Bild des Gewinners ist fort: %v", err)
	}
}
