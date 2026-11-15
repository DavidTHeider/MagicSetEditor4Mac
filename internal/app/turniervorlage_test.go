package app_test

// **Format und Vorlage eines Turniers an der Naht** — DOM-04.AK-01 und
// AK-03 (#1699).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Das Format geht als Kennung hinein und kommt als Text heraus** — leer
// heißt: keines.
func TestEinTurnierNenntSeinFormat(t *testing.T) {
	anwendung, naht := turnieranwendung(t)

	mit := anwendung.TurnierAnlegen(t.Context(), "Freitag", "magic", "", false, "4")
	if mit.Fehler != nil || naht.angelegt.FormatID != 4 || mit.Turnier.FormatID != "4" {
		t.Errorf("mit Format: %+v, angelegt %+v", mit, naht.angelegt)
	}

	ohne := anwendung.TurnierAnlegen(t.Context(), "Samstag", "magic", "", false, "")
	if ohne.Fehler != nil || naht.angelegt.FormatID != 0 || ohne.Turnier.FormatID != "" {
		t.Errorf("ohne Format: %+v, angelegt %+v", ohne, naht.angelegt)
	}

	krumm := anwendung.TurnierAnlegen(t.Context(), "Sonntag", "magic", "", false, "vier")
	if krumm.Fehler == nil || krumm.Fehler.Feld != "formatID" {
		t.Errorf("eine unbrauchbare Formatkennung: %+v", krumm.Fehler)
	}
}

// **Eine Vorlage geht als Kennung hinein, und die Kopie kommt als DTO
// heraus.**
func TestEinTurnierLaesstSichNachVorlageAnlegen(t *testing.T) {
	anwendung, naht := turnieranwendung(t)

	kopie := anwendung.TurnierKopieren(t.Context(), "3", "Freitag II", true)
	if kopie.Fehler != nil || kopie.Turnier.Name != "Freitag II" || kopie.Turnier.FormatID != "4" {
		t.Fatalf("die Kopie: %+v", kopie)
	}

	if naht.kopiert.quelle != 3 || naht.kopiert.name != "Freitag II" || !naht.kopiert.mitTeilnehmenden {
		t.Errorf("an die Naht ging %+v", naht.kopiert)
	}

	if krumm := anwendung.TurnierKopieren(t.Context(), "drei", "X", false); krumm.Fehler == nil ||
		krumm.Fehler.Feld != "turnierID" {
		t.Errorf("eine unbrauchbare Kennung: %+v", krumm.Fehler)
	}

	naht.fehler = domain.ErrNotFound
	if fehlt := anwendung.TurnierKopieren(t.Context(), "3", "X", false); fehlt.Fehler == nil ||
		fehlt.Fehler.Code != app.CodeNichtGefunden {
		t.Errorf("ein Nahtfehler: %+v", fehlt.Fehler)
	}
}
