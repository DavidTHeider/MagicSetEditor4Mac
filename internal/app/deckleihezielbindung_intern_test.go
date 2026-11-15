package app

// **Der Dialog für das Ziel einer Leihe** — UXR-F15
// (#2030).
//
// # Was hier geprüft wird und was nicht
//
// **Nicht der native Dialog** — er braucht ein Fenster und steht in
// `.covignore` (`dateiwahl.go`); die Hüllenprobe `uxrf15` fährt ihn am
// Fenster. Geprüft wird, was davor und danach steht: dass der Sichern-Dialog
// mit **Namen und Endung** aufgeht, dass ein leerer Pfad ein **Abbruch** ist
// und kein Fehler, dass ein Fehler des Dialogs ankommt — und dass die Wahl
// **nichts schreibt**. Geschrieben wird erst in [App.DeckleiheAnbieten].

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
)

// **Der Dialog geht als Sichern-Dialog auf — mit Namen und Endung.**
func TestDasLeihzielGehtAlsSichernDialogMitNamenUndEndungAuf(t *testing.T) {
	anwendung := mitSaat(t)

	gewaehlt := filepath.Join(t.TempDir(), "leihe.msepackage")

	var gesehen []Dateiauswahl

	anwendung.MitDateiwahl(func(_ context.Context, auswahl Dateiauswahl) (string, error) {
		gesehen = append(gesehen, auswahl)

		return gewaehlt, nil
	})

	antwort := NeueBindungen(anwendung).DeckleiheZielWaehlen(t.Context())

	if antwort.Fehler != nil || antwort.Abgebrochen {
		t.Fatalf("die Wahl scheiterte: %+v", antwort)
	}

	// **Der Pfad kommt zurück** — das Fenster zeigt ihn, und das Anbieten
	// bekommt ihn mit.
	if antwort.Pfad != gewaehlt {
		t.Errorf("zurück kam %q statt %q", antwort.Pfad, gewaehlt)
	}

	if len(gesehen) != 1 {
		t.Fatalf("der Dialog ging %dmal auf", len(gesehen))
	}

	texte := DialogtexteFuer(anwendung.sprache)
	auswahl := gesehen[0]

	if !auswahl.Speichern {
		t.Error("das Ziel ging als Öffnen-Dialog auf")
	}

	// **Die Endung kommt aus dem Code**, nicht aus einer Abschrift davon.
	if auswahl.Muster != "*"+austausch.Endung {
		t.Errorf("das Muster ist %q", auswahl.Muster)
	}

	if auswahl.Vorschlag != texte.LeiheVorschlag || auswahl.Titel != texte.LeiheAblegen ||
		auswahl.Filter != texte.UmschlagFilter {
		t.Errorf("der Dialog geht als %+v auf", auswahl)
	}
}

// **Die Wahl schreibt nichts** — am gewählten Ort entsteht keine Datei.
func TestDasLeihzielWaehlenSchreibtNichts(t *testing.T) {
	anwendung := mitSaat(t)

	gewaehlt := filepath.Join(t.TempDir(), "leihe.msepackage")

	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) {
		return gewaehlt, nil
	})

	NeueBindungen(anwendung).DeckleiheZielWaehlen(t.Context())

	if _, fehler := os.Stat(gewaehlt); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("am Ziel liegt etwas (%v)", fehler)
	}
}

// **Ein weggeklickter Dialog ist ein Abbruch, kein Fehler** — und er trägt
// keinen Pfad, damit das Fenster den vorigen stehen lässt.
func TestEinWeggeklicktesLeihzielIstEinAbbruch(t *testing.T) {
	anwendung := mitSaat(t)
	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) {
		return "", nil
	})

	antwort := NeueBindungen(anwendung).DeckleiheZielWaehlen(t.Context())

	if !antwort.Abgebrochen {
		t.Error("der Abbruch kam nicht an")
	}

	if antwort.Fehler != nil {
		t.Errorf("ein Abbruch ist kein Fehler (%+v)", antwort.Fehler)
	}

	if antwort.Pfad != "" {
		t.Errorf("ein Abbruch trägt einen Pfad: %q", antwort.Pfad)
	}
}

// **Ein Fehler des Dialogs kommt als Fehler an.**
func TestEinFehlerDesLeihzieldialogsKommtAn(t *testing.T) {
	anwendung := mitSaat(t)
	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) {
		return "", errors.New("kein Fenster")
	})

	antwort := NeueBindungen(anwendung).DeckleiheZielWaehlen(t.Context())

	if antwort.Fehler == nil {
		t.Fatal("der Fehler ging verloren")
	}

	if antwort.Abgebrochen || antwort.Pfad != "" {
		t.Errorf("ein Fehler ist kein Abbruch und trägt keinen Pfad: %+v", antwort)
	}
}

// **In jeder Sprache fragt der Dialog nach `.msepackage`** — der vorgeschlagene
// Name trägt die Endung, die der Umschlag verlangt; sonst schriebe er eine
// Datei, die der Filter beim Annehmen nicht anbietet.
func TestJederLeihvorschlagTraegtDieEndungDesUmschlags(t *testing.T) {
	for _, sprache := range Sprachen {
		t.Run(sprache, func(t *testing.T) {
			texte := DialogtexteFuer(sprache)

			if !strings.HasSuffix(texte.LeiheVorschlag, austausch.Endung) {
				t.Errorf("der Vorschlag %q hat nicht die Endung %q",
					texte.LeiheVorschlag, austausch.Endung)
			}

			if texte.LeiheVorschlag == DialogtexteFuer("de").LeiheVorschlag && sprache != "de" {
				t.Errorf("%q übernimmt den deutschen Vorschlag", sprache)
			}
		})
	}
}
