package storage

// **Jeder Fehlerweg des Kennwortwechsels meldet sich** — ID-02 (#1701).

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerKennwortwechselMeldetSeineFehler(t *testing.T) {
	if fehler := NeueNutzerablage(offeneDatenbank(t)).KennwortSetzen(probekontext(t), 1, "x"); fehler == nil {
		t.Error("ohne Tafel kam der Wechsel durch")
	}

	if fehler := NeueNutzerablage(offeneDatenbank(t)).KennwortMitCodeSetzen(probekontext(t), 1, 1, "x", domain.SitzungAngemeldet); fehler == nil {
		t.Error("ohne Tafel kam die Wiederherstellung durch")
	}

	// **Der Code ist abgestempelt, das Kennwort scheitert** — das Geschäft
	// meldet es, und nichts davon bleibt (Nachlauf über ID-02, Befund 2).
	zweites, err := sql.Open("stolper-zweites-schreiben", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = zweites.Close() })

	zweiterSchreibzaehler = 0

	if fehler := NeueNutzerablage(zweites).KennwortMitCodeSetzen(probekontext(t), 1, 1, "x", domain.SitzungAngemeldet); fehler == nil {
		t.Error("das gescheiterte Kennwort ging verloren")
	}

	db, err := sql.Open("stolper-zaehlen", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	zweiterSchreibzaehler = 0

	if fehler := NeueNutzerablage(db).KennwortSetzen(probekontext(t), 1, "x"); fehler == nil {
		t.Error("der Zählfehler ging verloren")
	}
}
