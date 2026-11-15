package app_test

// **Eine gleiche Adresse ist kein Konto** — ID-05.AK-01
// (#1704).
//
// Ein Anbieter, dessen Anzeigename die Adresse eines Nutzers dieses Rechners
// ist, bekommt **einen neuen** Nutzer — nicht den mit der Adresse, und nicht
// dessen Profil. Ein Konto hängt an Anbieter **und** Subjekt, und die
// Fremdkennung trägt gar keine Adresse (`domain.Fremdkennung`).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEineGleicheAdresseUebernimmtKeinKonto(t *testing.T) {
	db := offeneProbedatenbank(t)
	ablage := storage.NeueNutzerablage(db)

	ada, fehler := ablage.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Ada", Email: "ada@example.org"})
	if fehler != nil {
		t.Fatal(fehler)
	}

	herein, fehler := service.NeueAnmeldung(ablage).Herein(probekontext(t),
		domain.Fremdkennung{Anbieter: "github", Subjekt: "4711", Anzeigename: "ada@example.org"})
	if fehler != nil {
		t.Fatal(fehler)
	}

	if herein == ada.ID {
		t.Fatal("die gleiche Adresse führte zu Adas Nutzer")
	}

	konten, fehler := ablage.Konten(probekontext(t), ada.ID)
	if fehler != nil || len(konten) != 0 {
		t.Errorf("Ada trägt jetzt ein fremdes Konto: %+v %v", konten, fehler)
	}
}
