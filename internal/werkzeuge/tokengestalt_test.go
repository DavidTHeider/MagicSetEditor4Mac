package werkzeuge_test

// **Die Gestalt eines Zugriffstokens steht an zwei Stellen** — Werkstück 10
// von Phase 4.12
// (#1351).
//
// `internal/service` **baut** ein Token, `internal/rest` **erkennt** eines.
// Beide müssen dieselbe Gestalt meinen, und keines darf das andere kennen:
// `rest` ist eine Fläche und importiert nur `domain` (§2).
//
// **Zwei Zeichenketten in zwei Paketen, die zusammenpassen müssen, laufen beim
// nächsten Umbenennen auseinander** — und zwar lautlos: der Dienst gäbe
// weiterhin Token aus, und der Zugang hielte jedes für eine falsche Marke.
// Dieser Wächter steht deshalb hier, wo beide erreichbar sind.
//
// # Was er nicht prüft
//
// **Ob ein echtes Token durch den echten Zugang kommt.** Das ist eine Aussage
// über zwei laufende Prozesse und steht im Flächenlauf; hier steht nur, dass
// die beiden Konstanten dasselbe sagen.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieBeidenTokenmarkenPassenZusammen(t *testing.T) {
	t.Parallel()

	if rest.Tokenvorspann != service.Tokenmarke+"_" {
		t.Errorf("rest.Tokenvorspann = %q, service.Tokenmarke = %q — "+
			"der Zugang erkennt nicht, was der Dienst ausgibt",
			rest.Tokenvorspann, service.Tokenmarke)
	}
}

// **Und ein gebautes Token trägt den Vorspann, den der Zugang sucht.**
//
// **Der Fall, ohne den der Lauf darüber nur zwei Konstanten vergliche.** Er
// baut eines über den echten Weg — und dabei fällt auf, wenn jemand die
// Zusammensetzung ändert, ohne die Konstante anzufassen.
func TestEinGebautesTokenTraegtDenVorspann(t *testing.T) {
	t.Parallel()

	verwaltung := service.NeueTokenverwaltung(&tokenablage{})

	_, klartext, fehler := verwaltung.Anlegen(t.Context(), tokenprobe())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if !strings.HasPrefix(klartext, rest.Tokenvorspann) {
		t.Errorf("das gebaute Token %q beginnt nicht mit %q",
			klartext, rest.Tokenvorspann)
	}
}
