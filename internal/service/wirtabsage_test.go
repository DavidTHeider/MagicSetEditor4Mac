package service_test

// **Ein abgewiesener Wunsch beendet die Sitzung nicht** —
// #2478,
// gefunden bei P9b mit zwei Anwendungen: der Kern wies einen Zug an der
// Fassungsprüfung ab, `befehlen` gab den Fehler zurück, und `Bediene` legte
// die Leitung nieder. Bert saß danach an keinem Tisch mehr.
//
// **Eine fachliche Abweisung geht als Grund an den Gast** — am Tischbild,
// das er ohnehin liest. Was kein Urteil des Kerns ist, endet weiter
// (`TestWasDerDienstKlagtWirdNichtVerkuendet`).

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinAbgewiesenerWunschBeendetDieSitzungNicht(t *testing.T) {
	t.Parallel()

	faelle := map[string]error{
		"ungültig":       fmt.Errorf("%w: dieser Zug ist nicht erlaubt", domain.ErrValidation),
		"überholt":       fmt.Errorf("%w: die Fassung gilt nicht mehr", domain.ErrConflict),
		"nicht gefunden": fmt.Errorf("%w: dieses Objekt liegt nicht am Tisch", domain.ErrNotFound),
	}

	for name, abweisung := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wirt, naht := einWirt(&nehmerdoppel{klagt: abweisung})

			if fehler := wirt.Angekommen(context.Background(), einSpielwunsch(t,
				domain.Spielbefehl{Kennung: "b1", Art: "objekt.tappen"})); fehler != nil {
				t.Fatalf("die Sitzung endete: %v", fehler)
			}

			bild := dasLetzteTischbild(t, naht)
			if bild.Absage == nil || bild.Absage.Befehl != "b1" || bild.Absage.Art != "objekt.tappen" ||
				!strings.Contains(abweisung.Error(), bild.Absage.Grund) || bild.Absage.Grund == "" {
				t.Errorf("die Absage: %+v", bild.Absage)
			}
		})
	}
}

// dasLetzteTischbild liest, was der Wirt zuletzt als Tischbild schickte.
func dasLetzteTischbild(t *testing.T, naht *nahtdoppel) domain.Tischbild {
	t.Helper()

	gesendet := naht.letzte(t)

	art, inhalt, fehler := domain.LiesKontrolle(gesendet.Inhalt, gesendet.Von, domain.KontrolleTischbild)
	if fehler != nil || art != domain.KontrolleTischbild {
		t.Fatalf("kein Tischbild: %v", fehler)
	}

	var bild domain.Tischbild
	if fehler := json.Unmarshal(inhalt, &bild); fehler != nil {
		t.Fatalf("unlesbar: %v", fehler)
	}

	return bild
}

// **Lässt sich das Bild danach nicht rechnen, endet sie doch** — dann klemmt
// dieser Rechner, und der Gast liest einen Abriss statt eines falschen Grundes.
func TestOhneBildEndetAuchDieAbsage(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("das Bild ließ sich nicht rechnen")
	wirt, _ := einWirt(&nehmerdoppel{
		klagt:     fmt.Errorf("%w: dieser Zug ist nicht erlaubt", domain.ErrValidation),
		bildklage: kaputt,
	})

	if fehler := wirt.Angekommen(context.Background(), einSpielwunsch(t,
		domain.Spielbefehl{Kennung: "b1", Art: "objekt.tappen"})); !errors.Is(fehler, kaputt) {
		t.Errorf("es kam %v", fehler)
	}
}
