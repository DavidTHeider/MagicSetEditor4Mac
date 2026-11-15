package service

// **Ein Paket, das als Verweis im Paketordner liegt** — gefunden bei C04
// (#2229),
// als das Fenster zum ersten Mal neben den anderen Flächen sicherte.
//
// `filepath.WalkDir` folgt keinem Verweis: der Paketname wurde zu **einem**
// Teil, dessen Quelle ein Verzeichnis ist, und die Sicherung brach mit „is a
// directory" ab — im Fenster als „unerwarteter Fehler". **Die Regel davor
// bleibt** (`sicherungspakete.go`): mit reist nur, was im verwalteten Ordner
// liegt. Ein Verweis dorthin ist ein zweiter Name für ein Paket des Ordners
// und reist unter ihm mit; einer nach außen trägt fremde Dateien und reist
// nicht mit — wie ein Paket aus einer MSE2-Installation.

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

func TestEinPaketAlsVerweisHaeltDieSicherungNichtAuf(t *testing.T) {
	bau := neuerWegbau(t)
	ordner := filepath.Join(bau.ort, "pakete")

	eigen := paketAnlegen(t, ordner, "eigen.mse-game")
	draussen := paketAnlegen(t, t.TempDir(), "draussen.mse-game")

	zweitname := filepath.Join(ordner, "zweitname.mse-game")
	nachAussen := filepath.Join(ordner, "aussen.mse-game")

	for verweis, ziel := range map[string]string{zweitname: eigen, nachAussen: draussen} {
		if fehler := os.Symlink(ziel, verweis); fehler != nil {
			t.Fatal(fehler)
		}
	}

	bau.weg.MitPaketen(festerKatalog{
		{Art: "mse-game", Name: "eigen", Pfad: eigen},
		{Art: "mse-game", Name: "zweitname", Pfad: zweitname},
		{Art: "mse-game", Name: "aussen", Pfad: nachAussen},
		// **Seit dem Lesen des Katalogs verschwunden** — es reist nicht mit
		// und steht nicht als „enthalten" da.
		{Art: "mse-game", Name: "weg", Pfad: filepath.Join(ordner, "weg.mse-game")},
	}, func() (string, error) { return ordner, nil })

	ziel := filepath.Join(t.TempDir(), "s"+sicherung.Endung)
	if fehler := bau.weg.Sichern(t.Context(), ziel); fehler != nil {
		t.Fatalf("sichern: %v", fehler)
	}

	schau, fehler := bau.weg.Beschauen(t.Context(), ziel)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if !slices.Equal(schau.PaketeEnthalten, []string{"eigen.mse-game", "zweitname.mse-game"}) {
		t.Errorf("enthalten %v — der Verweis im Ordner reist mit, der nach außen und das verschwundene nicht", schau.PaketeEnthalten)
	}

	// **Und der Zweitname kommt als Paket zurück**, mit seinem Inhalt.
	if raeumfehler := os.RemoveAll(ordner); raeumfehler != nil {
		t.Fatal(raeumfehler)
	}

	if _, fehler := bau.weg.Zurueckspielen(t.Context(), ziel); fehler != nil {
		t.Fatalf("zurückspielen: %v", fehler)
	}

	if inhalt, lesefehler := os.ReadFile(filepath.Join(ordner, "zweitname.mse-game", "game")); lesefehler != nil || //nolint:gosec // aus t.TempDir()
		string(inhalt) != "name: eigen.mse-game" {
		t.Errorf("der Zweitname kam nicht zurück: %q, %v", inhalt, lesefehler)
	}
}
