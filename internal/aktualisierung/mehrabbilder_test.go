package aktualisierung_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Zwei Abbilder an einer Freigabe sind keine Wahl, sondern ein Fehler** —
// I06 (#2264):
// die Aktualisierung liest die **richtige** Freigabedatei. Hier nahm die
// Schleife über die Beigaben stillschweigend das letzte `.dmg`; welches
// geladen wurde, entschied die Reihenfolge, in der GitHub sie nannte.
func TestZweiAbbilderWerdenNichtGeladen(t *testing.T) {
	t.Parallel()

	dienst, _ := gegenstelle(t, `{"tag_name":"v9.9.9","assets":[
		{"name":"MagicSetEditor4Mac-9.9.9.dmg","size":1,"url":"a"},
		{"name":"MagicSetEditor4Mac-9.9.9-alt.dmg","size":1,"url":"b"},
		{"name":"pruefsummen.txt","size":1,"url":"c"},
		{"name":"pruefsummen.txt.asc","size":1,"url":"d"}]}`, http.StatusOK)

	klient := probeklient(t, dienst.URL)

	freigabe, err := klient.NeuesteFreigabe(t.Context())
	if err != nil {
		t.Fatalf("holen: %v", err)
	}

	// **Melden geht**, wie ohne Abbild: die Fassung ist bekannt.
	if freigabe.Fassung != "9.9.9" {
		t.Errorf("Fassung = %q", freigabe.Fassung)
	}

	_, err = klient.Laden(t.Context(), freigabe, t.TempDir(), nil)
	if !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "-alt.dmg") ||
		!strings.Contains(err.Error(), "9.9.9.dmg") {
		t.Errorf("Laden: %v — erwartet ein Konflikt, der beide Abbilder nennt", err)
	}
}
