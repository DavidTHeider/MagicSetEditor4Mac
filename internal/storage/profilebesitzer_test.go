package storage_test

// **Die Profile dessen, dem ein Profil gehört** — UXR-F21
// (#2041).
//
// **Kommandozeile, MCP und Dienst kennen nur das Profil, in dem sie
// arbeiten** — keinen Nutzer, den ein Zeiger nennte. Die Frage „gibt es das
// Profil, an das ein Sitz gebunden ist?" geht deshalb von dort aus: dieselben
// Profile wie die des Besitzers, und keine eines anderen Nutzers.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieProfileDesBesitzersSindNurSeine(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	ctx := probekontext(t)

	// Der erste Nutzer bekommt ein zweites Profil, der zweite hat sein
	// erstes.
	erste, err := ablage.Nutzer(ctx)
	if err != nil || len(erste) != 1 {
		t.Fatalf("der erste Nutzer: %v, %d", err, len(erste))
	}

	if _, anlegefehler := ablage.ProfilAnlegen(ctx, domain.Profil{NutzerID: erste[0].ID, Name: "Turnier"}); anlegefehler != nil {
		t.Fatalf("das zweite Profil: %v", anlegefehler)
	}

	fremder, err := ablage.NutzerAnlegen(ctx, domain.Nutzer{Name: "Zweite"})
	if err != nil {
		t.Fatalf("der zweite Nutzer: %v", err)
	}

	eigene, err := ablage.Profile(ctx, erste[0].ID)
	if err != nil || len(eigene) != 2 {
		t.Fatalf("die Profile des ersten: %v, %d", err, len(eigene))
	}

	// **Von jedem seiner Profile aus dieselben zwei.**
	for _, ausgang := range eigene {
		gelesen, lesefehler := ablage.ProfileDesBesitzersVon(ctx, ausgang.ID)
		if lesefehler != nil {
			t.Fatalf("von %d aus: %v", ausgang.ID, lesefehler)
		}

		if len(gelesen) != 2 || gelesen[0].NutzerID != erste[0].ID ||
			gelesen[1].NutzerID != erste[0].ID {
			t.Errorf("von %d aus: %+v", ausgang.ID, gelesen)
		}
	}

	// **Und vom Profil des anderen aus nur dessen eines** — nie das eines
	// fremden Nutzers.
	gelesen, err := ablage.ProfileDesBesitzersVon(ctx, 987654)
	if err != nil || len(gelesen) != 0 {
		t.Errorf("ein Profil, das es nicht gibt, hat keine Geschwister: %v, %+v", err, gelesen)
	}

	fremdeProfile, err := ablage.Profile(ctx, fremder.ID)
	if err != nil || len(fremdeProfile) != 1 {
		t.Fatalf("die Profile des zweiten: %v, %d", err, len(fremdeProfile))
	}

	gelesen, err = ablage.ProfileDesBesitzersVon(ctx, fremdeProfile[0].ID)
	if err != nil || len(gelesen) != 1 || gelesen[0].NutzerID != fremder.ID {
		t.Errorf("vom fremden Profil aus: %v, %+v", err, gelesen)
	}
}
