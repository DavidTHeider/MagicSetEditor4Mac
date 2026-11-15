package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// **Die Aufnahme hält dieselbe Grenze wie der Draht** — und bricht ebenso ab.
func TestDieAufnahmeBrichtAnIhrerGrenzeAb(t *testing.T) {
	defer func() {
		if grund := recover(); grund != http.ErrAbortHandler { //nolint:errorlint // ein Wert aus recover
			t.Errorf("abgebrochen mit %v", grund)
		}
	}()

	_, _ = neueAufnahme(2).Write([]byte("abc")) //nolint:errcheck // bricht ab

	t.Error("nicht abgebrochen")
}

// fristschreiber merkt sich, ob jemand seine Lesefrist setzt.
type fristschreiber struct {
	http.ResponseWriter

	gesetzt bool
}

func (f *fristschreiber) SetReadDeadline(time.Time) error {
	f.gesetzt = true

	return nil
}

// **Die Hüllen des Dienstes lassen den Fristruf durch** (FMT-03.AK-03) —
// sonst verlängerte keine Übertragung ihre Frist, und nach zehn Sekunden
// bräche sie ab.
func TestDieHuellenLassenDieFristDurch(t *testing.T) {
	innen := &fristschreiber{ResponseWriter: httptest.NewRecorder()}
	huelle := &statusmitschrift{ResponseWriter: begrenzt(innen, 10)}

	if err := http.NewResponseController(huelle).SetReadDeadline(time.Now()); err != nil || !innen.gesetzt {
		t.Errorf("%v, gesetzt %v", err, innen.gesetzt)
	}
}
