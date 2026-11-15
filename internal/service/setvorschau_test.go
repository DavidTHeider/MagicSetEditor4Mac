package service_test

// **Ein Set vorab ansehen** — FMT-03.AK-01
// (#1692).
//
// Die Vorschau liest die Datei wie das Einfahren und sagt, was es täte —
// **ohne** in den Bestand zu schreiben. Eine Datei, die das Einfahren
// abwiese, weist sie mit derselben Meldung ab.

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieVorschauSagtWasDasEinfahrenTaete(t *testing.T) {
	speicher := &speicherDouble{}

	vorschau, err := service.NeuerImporteur(speicher).
		Vorschau(t.Context(), filepath.Join(testdaten, "probe.mse-set"))
	if err != nil {
		t.Fatalf("Vorschau: %v", err)
	}

	if vorschau.Name != "Mein Probeset" || vorschau.Spiel != "probe" || vorschau.Stil != "probe-stil" ||
		vorschau.Karten != 3 || vorschau.SchonOffen != 0 {
		t.Errorf("Vorschau: %+v", vorschau)
	}

	// **Was das Einfahren melden würde, meldet sie schon jetzt** — der
	// unverstandene Block reist roh mit und ist kein Befund.
	if vorschau.Befunde == nil {
		t.Error("die Befunde sind nil statt leer")
	}

	if speicher.set != nil {
		t.Errorf("die Vorschau hat geschrieben: %q", speicher.set.Name)
	}
}

// **Liegt die Quelle schon im Bestand, sagt sie es** — das Einfahren gäbe
// dann das vorhandene Set zurück, statt ein zweites anzulegen.
func TestDieVorschauNenntDasSchonOffeneSet(t *testing.T) {
	speicher := &setspeicherprobe{}
	importeur := neuerImporteurMitKoepfen(t, speicher)
	pfad := filepath.Join(testdaten, "probe.mse-set")

	if _, err := importeur.Set(t.Context(), pfad); err != nil {
		t.Fatal(err)
	}

	vorschau, err := importeur.Vorschau(t.Context(), pfad)
	if err != nil {
		t.Fatal(err)
	}

	if vorschau.SchonOffen != 1 || vorschau.Karten != 3 || speicher.laeufe != 1 {
		t.Errorf("Vorschau %+v nach %d Läufen", vorschau, speicher.laeufe)
	}
}

func TestDieVorschauWeistAbWasDasEinfahrenAbwiese(t *testing.T) {
	for _, fall := range []struct {
		name   string
		pfad   string
		klasse error
	}{
		{"ein Spiel statt eines Sets", filepath.Join(testdaten, "probe.mse-game"), domain.ErrValidation},
		{"ein Set ohne Spiel", filepath.Join(testdaten, "ohnespiel.mse-set"), domain.ErrValidation},
		{"eine Datei, die es nicht gibt", filepath.Join(t.TempDir(), "gibtsnicht.mse-set"), nil},
	} {
		t.Run(fall.name, func(t *testing.T) {
			speicher := &speicherDouble{}

			_, err := service.NeuerImporteur(speicher).Vorschau(t.Context(), fall.pfad)
			if err == nil || (fall.klasse != nil && !errors.Is(err, fall.klasse)) {
				t.Errorf("Fehler %v", err)
			}

			// **Und dieselbe Meldung wie das Einfahren** — sonst sagte die
			// Vorschau „geht" und das Einfahren danach „geht nicht".
			_, einfahrfehler := service.NeuerImporteur(speicher).Set(t.Context(), fall.pfad)
			if einfahrfehler == nil || einfahrfehler.Error() != err.Error() {
				t.Errorf("Vorschau %v, Einfahren %v", err, einfahrfehler)
			}
		})
	}
}

// **Ein Bestand, der nicht zu lesen ist, ist ein Fehler** — die Vorschau
// sagte sonst „neu", und das Einfahren legte ein zweites Set an.
func TestEineVorschauOhneLesbarenBestandScheitert(t *testing.T) {
	_, err := service.NeuerImporteur(&speicherDouble{}).MitKoepfen(kaputteKoepfe{}).
		Vorschau(t.Context(), filepath.Join(testdaten, "probe.mse-set"))
	if err == nil {
		t.Error("kein Fehler")
	}
}
