package service_test

// **Was geschieht, wenn Schreiben scheitert** — OPS-09b
// (#1961):
// es bleibt kein halber Stand, und der Fehler kommt an.

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine plist, die sich nicht schreiben lässt, ruft keinen Befehl.**
func TestEineUnschreibbarePlistLaedtNichts(t *testing.T) {
	t.Parallel()

	ziel, gerufen := einAgentenziel(t)
	if fehler := os.MkdirAll(filepath.Dir(ziel.Datei), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	sperren(t, filepath.Dir(ziel.Datei), 0o500)

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); fehler == nil || len(*gerufen) != 0 {
		t.Errorf("%v %v", fehler, *gerufen)
	}
}

// **Eine plist, die sich nach dem Entladen nicht löschen lässt, sagt es.**
func TestEineUnloeschbarePlistSagtEs(t *testing.T) {
	t.Parallel()

	ziel, _ := einAgentenziel(t)
	ausgang(t)(service.SchnittstelleEinrichten(t.Context(), ziel))
	sperren(t, filepath.Dir(ziel.Datei), 0o500)

	if _, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); fehler == nil {
		t.Error("das Löschen scheiterte still")
	}
}

// **Ohne Gedächtnis wird nicht eingerichtet** — sonst ginge der Stand davor
// verloren, und widerrufen könnte ihn nie zurücklegen.
func TestOhneGedaechtnisKeinEintrag(t *testing.T) {
	t.Parallel()

	ziel, datei := einDesktopziel(t)
	if fehler := os.MkdirAll(filepath.Dir(ziel.Gedaechtnis), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	sperren(t, filepath.Dir(ziel.Gedaechtnis), 0o500)

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); fehler == nil {
		t.Fatal("eingerichtet ohne Gedächtnis")
	}

	if _, fehler := os.Stat(datei); !os.IsNotExist(fehler) {
		t.Errorf("die Datei entstand trotzdem: %v", fehler)
	}
}

// **Lässt sich der Stand davor nicht zurücklegen, sagt widerrufen es** —
// und das Gedächtnis bleibt, damit ein zweiter Versuch ihn noch findet.
func TestEinUnzuruecklegbarerStandSagtEs(t *testing.T) {
	t.Parallel()

	for name, vorher := range map[string]string{"die Datei fehlte": "", "die Datei war da": "{\"a\": 1}\n"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ziel, datei := einDesktopziel(t)
			if fehler := os.MkdirAll(filepath.Dir(datei), 0o700); fehler != nil {
				t.Fatal(fehler)
			}

			if vorher != "" {
				schreibeDatei(t, datei, vorher)
			}

			ausgang(t)(service.SchnittstelleEinrichten(context.Background(), ziel))
			sperren(t, filepath.Dir(datei), 0o500)

			if _, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); fehler == nil {
				t.Fatal("das Zurücklegen scheiterte still")
			}

			if _, fehler := os.Stat(ziel.Gedaechtnis); fehler != nil {
				t.Errorf("das Gedächtnis ging: %v", fehler)
			}
		})
	}
}

// **Ein Gedächtnis, das sich nicht löschen lässt, sagt es** — zurückgelegt
// ist die Datei trotzdem.
func TestEinUnloeschbaresGedaechtnisSagtEs(t *testing.T) {
	t.Parallel()

	ziel, datei := einDesktopziel(t)

	ausgang(t)(service.SchnittstelleEinrichten(t.Context(), ziel))
	sperren(t, filepath.Dir(ziel.Gedaechtnis), 0o500)

	aus, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel)
	if fehler == nil {
		t.Error("das Löschen des Gedächtnisses scheiterte still")
	}

	if _, nichtDa := os.Stat(datei); !os.IsNotExist(nichtDa) || aus.Stand != service.SchnittstelleOffen {
		t.Errorf("zurückgelegt? %+v %v", aus, nichtDa)
	}
}
