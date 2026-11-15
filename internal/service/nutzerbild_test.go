package service_test

// **Das Profilbild** — Werkstück 8 von Phase 4.12
// (#1351).

import (
	"errors"
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Bild landet unter `nutzer/` und trägt die Kennung.**
//
// **Neben den Kartenbildern und nicht darunter**: die stehen je Set, und ein
// Profilbild dort verlöre man beim Löschen des Sets — ein Mensch gehört
// keinem Set.
func TestEinProfilbildLandetUnterNutzer(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	name, fehler := ablage.NutzerbildAblegen(7, probebild())
	if fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	// **Kennung und Inhalt im Namen** (Zweitblick über Phase 6.1, Befund 3).
	if !strings.HasPrefix(name, "nutzer-7-") || !strings.HasSuffix(name, ".png") {
		t.Errorf("der Name ist %q", name)
	}

	ort := filepath.Join(wurzel, service.Nutzerbildordner, name)
	if _, err := os.Stat(ort); err != nil {
		t.Errorf("die Datei fehlt: %v", err)
	}
}

// **Zwei Ablagen, zwei Namen — auch für dasselbe Bild** (Nachlauf über
// Phase 6.1, Befund 1).
//
// **Hier stand „dasselbe Bild zweimal ist eine Datei"**, und genau das war
// der Wettlauf: zwei gleichzeitige Wechsel auf dasselbe Bild teilten sich eine
// Datei, und der Verlierer räumte die des Gewinners weg. Jetzt hat jedes
// Ablegen seinen Namen; das alte Bild räumt die Anwendung weg, sobald der
// Bestand gewechselt hat (`internal/app/nutzerbilder.go`), und so liegt nach
// einem Wechsel trotzdem eine Datei da.
func TestJedesAblegenBekommtEinenEigenenNamen(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	erster, fehler := ablage.NutzerbildAblegen(7, probebild())
	if fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	zweiter, zweitfehler := ablage.NutzerbildAblegen(7, probebild())
	if zweitfehler != nil {
		t.Fatalf("ablegen: %v", zweitfehler)
	}

	if erster == zweiter {
		t.Errorf("zweimal derselbe Name %q — zwei Wechsel teilten sich eine Datei", erster)
	}
}

// **Ein Bild lässt sich wegnehmen — und zweimal wegnehmen ist kein Fehler.**
func TestEinProfilbildLaesstSichWegnehmen(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	if _, fehler := ablage.NutzerbildAblegen(7, probebild()); fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	if fehler := ablage.NutzerbildWegnehmen(7); fehler != nil {
		t.Fatalf("wegnehmen: %v", fehler)
	}

	// **Dass dort schon jemand aufgeräumt hat, ist keine schlechte
	// Nachricht** — dieselbe Regel wie bei den Kartenbildern.
	if fehler := ablage.NutzerbildWegnehmen(7); fehler != nil {
		t.Errorf("das zweite Wegnehmen meldete: %v", fehler)
	}
}

// **Ohne Wurzel scheitert das Ablegen — und das Wegnehmen nicht.**
//
// **Weil das eine etwas tun soll und das andere nichts.** Eine Anwendung
// ohne Benutzerverzeichnis kann kein Bild ablegen; dass sie auch keines
// wegzunehmen hat, ist die richtige Antwort und kein Fehler.
func TestOhneWurzelScheitertNurDasAblegen(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage("")

	if _, fehler := ablage.NutzerbildAblegen(7, probebild()); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("gemeldet wurde: %v", fehler)
	}

	if fehler := ablage.NutzerbildWegnehmen(7); fehler != nil {
		t.Errorf("das Wegnehmen meldete: %v", fehler)
	}

	if ort := ablage.Nutzerverzeichnis(); ort != "" {
		t.Errorf("das Verzeichnis ist %q", ort)
	}
}

// **Ein Verzeichnis, das sich nicht anlegen lässt, meldet sich.**
func TestEinUnmoeglichesVerzeichnisMeldetSich(t *testing.T) {
	t.Parallel()

	// **Eine Datei dort, wo das Verzeichnis hin soll** — `MkdirAll`
	// scheitert dann, und zwar auf jedem Betriebssystem.
	wurzel := t.TempDir()
	if err := os.WriteFile(filepath.Join(wurzel, service.Nutzerbildordner),
		[]byte("keine Ablage"), 0o600); err != nil {
		t.Fatalf("die Sperre legen: %v", err)
	}

	if _, fehler := service.NeueBildablage(wurzel).
		NutzerbildAblegen(7, probebild()); fehler == nil {
		t.Error("das Ablegen glückte trotzdem")
	}
}

// **Ein Bild, das sich nicht kodieren lässt, wird gemeldet.**
//
// Ein Bild ohne Ausdehnung ist keines — `png.Encode` weist es ab. Genau
// darauf beruht die Zusage aus dem Kopf von `nutzerbild.go`: das Umkodieren
// **beweist**, dass die gewählte Datei ein Bild ist.
func TestEinUnkodierbaresProfilbildWirdGemeldet(t *testing.T) {
	t.Parallel()

	leer := image.NewNRGBA(image.Rectangle{})
	if _, fehler := service.NeueBildablage(t.TempDir()).
		NutzerbildAblegen(7, leer); fehler == nil {
		t.Error("das leere Bild kam durch")
	}
}

// **Und ein Verzeichnis, in das sich nicht schreiben lässt, ebenso.**
//
// `MkdirAll` glückt auf einem vorhandenen Verzeichnis auch dann, wenn es nur
// lesbar ist — der Fehler fällt erst beim Schreiben an.
//
// **Das Verzeichnis muss leer sein, und das ist der Unterschied zu
// [Bildablage.Ablegen].** Dort trägt jeder Name einen laufenden Zähler, und
// eine **neue** Datei braucht das Schreibrecht am Verzeichnis. Hier trägt der
// Name die Kennung: liegt das Bild schon da, überschreibt `WriteFile` es über
// das Recht an der **Datei** und käme durch. Der erste Anlauf dieses Tests
// legte deshalb vorher ein Bild ab — und war grün, ohne den Zweig zu
// betreten.
func TestEinNurLesbaresNutzerverzeichnisWirdGemeldet(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())

	//nolint:gosec // absichtlich: lesbar, nicht schreibbar
	if err := os.MkdirAll(ablage.Nutzerverzeichnis(), 0o500); err != nil {
		t.Fatalf("die Sperre legen: %v", err)
	}

	t.Cleanup(func() {
		//nolint:gosec // ein Verzeichnis braucht x, sonst lässt es sich nicht räumen
		if fehler := os.Chmod(ablage.Nutzerverzeichnis(), 0o750); fehler != nil {
			t.Errorf("Rechte zurücksetzen: %v", fehler)
		}
	})

	if _, fehler := ablage.NutzerbildAblegen(7, probebild()); fehler == nil {
		t.Error("das nur lesbare Verzeichnis kam durch")
	}
}

// **Ein Wegnehmen, das an etwas anderem als am Fehlen scheitert, meldet
// sich.**
//
// **Der Unterschied ist die ganze Zusage**: „nicht da" wird geschluckt, alles
// andere nicht. Ein Verzeichnis an der Stelle der Datei ist nicht leer, und
// `os.Remove` weist es ab — ein Fehler, der nicht `IsNotExist` ist.
func TestEinGescheitertesWegnehmenMeldetSich(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	for _, name := range []string{"nutzer-7.png", "nutzer-7-abcdefabcdef.png"} {
		sperre := filepath.Join(wurzel, service.Nutzerbildordner, name)
		if err := os.MkdirAll(filepath.Join(sperre, "inhalt"), 0o750); err != nil {
			t.Fatalf("die Sperre legen: %v", err)
		}
	}

	fehler := ablage.NutzerbildWegnehmen(7)
	if fehler == nil {
		t.Fatal("das Wegnehmen glückte trotzdem")
	}

	// **Beide Namen, der alte feste und der mit Inhalt**, und beide melden
	// sich — sonst verdeckte der erste Fehler den zweiten.
	if strings.Count(fehler.Error(), "das Bild entfernen") != 2 {
		t.Errorf("gemeldet: %v", fehler)
	}
}

// **Ein Bild lässt sich unter seinem Namen wegnehmen — und nur dort.**
//
// Nach einem Wechsel geht das alte, nach einem abgewiesenen das neue; ein
// Name, der aus dem Ordner führte, wird abgewiesen, statt dort zu löschen.
func TestEinBildLaesstSichNamentlichWegnehmen(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	name, fehler := ablage.NutzerbildAblegen(7, probebild())
	if fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	if fehler := ablage.BildWegnehmen(name); fehler != nil {
		t.Fatalf("wegnehmen: %v", fehler)
	}

	if _, err := os.Stat(filepath.Join(wurzel, service.Nutzerbildordner, name)); !os.IsNotExist(err) {
		t.Errorf("die Datei liegt noch: %v", err)
	}

	// **Zweimal ist kein Fehler** — dieselbe Regel wie beim Wegnehmen je Kennung.
	if fehler := ablage.BildWegnehmen(name); fehler != nil {
		t.Errorf("das zweite Wegnehmen: %v", fehler)
	}

	// **Ohne Wurzel gibt es nichts wegzunehmen** — und kein `os.Remove` auf
	// einen Namen, der dann relativ zum Arbeitsverzeichnis gälte.
	if fehler := service.NeueBildablage("").BildWegnehmen(name); fehler != nil {
		t.Errorf("ohne Wurzel: %v", fehler)
	}

	for _, kaputt := range []string{"", "../bestand.db", "unter/nutzer-7.png", ".versteckt"} {
		if fehler := ablage.BildWegnehmen(kaputt); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%q: %v, erwartet war eine Abweisung", kaputt, fehler)
		}
	}
}

// **Das Löschen eines Nutzers nimmt jedes seiner Bilder** — das alte unter
// dem festen Namen und jedes mit Inhalt im Namen, und nur seine.
func TestNutzerbildWegnehmenNimmtJedesBildDesNutzers(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)
	ort := filepath.Join(wurzel, service.Nutzerbildordner)

	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	for _, name := range []string{"nutzer-7.png", "nutzer-7-aaaaaaaaaaaa.png", "nutzer-70-bbbbbbbbbbbb.png", "profil-7.png"} {
		if err := os.WriteFile(filepath.Join(ort, name), []byte("x"), 0o600); err != nil {
			t.Fatalf("anlegen: %v", err)
		}
	}

	if fehler := ablage.NutzerbildWegnehmen(7); fehler != nil {
		t.Fatalf("wegnehmen: %v", fehler)
	}

	eintraege, err := os.ReadDir(ort)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	uebrig := make([]string, 0, len(eintraege))
	for _, eintrag := range eintraege {
		uebrig = append(uebrig, eintrag.Name())
	}

	if strings.Join(uebrig, " ") != "nutzer-70-bbbbbbbbbbbb.png profil-7.png" {
		t.Errorf("übrig sind %v — erwartet waren die Bilder der anderen", uebrig)
	}
}
