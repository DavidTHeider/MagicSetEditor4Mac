package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gedaechtnisquelle gibt einen Setkopf und einen Fehler auf die Setangaben.
type gedaechtnisquelle struct{ fehler error }

func (gedaechtnisquelle) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return []domain.SetHead{{ID: 7, Style: "probe-schlicht", Game: "probe"}}, nil
}

func (q *gedaechtnisquelle) Setwerte(context.Context, int64) ([]domain.Value, error) {
	return nil, q.fehler
}

// **Zwei Leser, zwei Befundlisten — und keine, die der andere ändert.**
//
// # Warum das ohne Kopie schiefgeht
//
// [Stilspeicher.StilFuerSet] gibt die gemerkte Befundliste heraus und hängt den
// Lesefehler der Setangaben an. Hat die Liste **freien Platz**, schreibt
// `append` in das gemerkte Feld. Die Länge im Gedächtnis bleibt dabei, wie sie
// war — was bleibt, ist der geteilte Speicherplatz: der zweite Leser
// überschreibt den Befund, den der erste noch in der Hand hält.
//
// Das ist die stillste Sorte Fehler: die Vorschau zeigt zwei Sets, und beim
// zweiten Klick nennt die erste Meldung einen Grund, der zur anderen Karte
// gehört.
//
// # Warum von innen
//
// Freier Platz entsteht dort, wo eine Liste wächst — und ob die Befunde eines
// bestimmten Stilpakets welchen haben, ist ein Zufall der Paketgröße. Ein Test,
// der von außen darauf hofft, prüft nichts Bestimmtes. Hier wird das Gedächtnis
// **gesetzt**, und zwar mit dem Fall, um den es geht.
func TestEinBefundUeberDieSetangabenBleibtNichtImGedaechtnis(t *testing.T) {
	quelle := &gedaechtnisquelle{fehler: errors.New("Bestand weg")}
	speicher := NeuerStilspeicher(quelle)

	// Eine Liste mit Platz für einen weiteren Eintrag — genau der Fall.
	gemerkte := make([]domain.Befund, 0, 4)
	//
	// **Mit der Zeit der letzten Prüfung**, wie der Speicher selbst merkt
	// (`stilstempel.go`) — ohne sie gälte das Gedächtnis als veraltet und
	// würde neu gelesen, und der Fall, um den es geht, käme nicht zustande.
	//
	// **Und mit dem Kopf, zu dem es gehört** (#2005) — nennt der Bestand einen
	// anderen, gilt es nicht mehr.
	speicher.gemerkt[7] = gemerkterStil{
		befunde: gemerkte, da: true, geprueft: time.Now(), stil: "probe-schlicht", spiel: "probe",
	}

	_, erste, _ := speicher.StilFuerSet(t.Context(), 7)
	if len(erste) != 1 {
		t.Fatalf("der erste Lauf meldet %d Befunde, erwartet einen", len(erste))
	}

	quelle.fehler = errors.New("etwas ganz anderes")

	if _, zweite, _ := speicher.StilFuerSet(t.Context(), 7); len(zweite) != 1 {
		t.Fatalf("der zweite Lauf meldet %d Befunde, erwartet einen", len(zweite))
	}

	if !strings.Contains(erste[0].Meldung, "Bestand weg") {
		t.Errorf("der erste Leser hält jetzt %q — der zweite hat ihm hineingeschrieben",
			erste[0].Meldung)
	}
}
