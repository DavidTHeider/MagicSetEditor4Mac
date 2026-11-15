package rest_test

// **Modus und Abwahl an der Regelfreigabe** — PLAY-02.AK-01 und AK-02
// (#1710).
//
// **Eine eigene Datei**, weil sie eine eigene Naht braucht: die
// Vorrichtungen daneben antworten, hier wird **mitgeschrieben**, womit
// gerufen wurde — geprüft wird das Argument und nicht der bloße Ruf (§10).

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// merkendeRegelnaht hält fest, womit die Naht gerufen wurde.
type merkendeRegelnaht struct {
	williger

	modus      string
	abgewaehlt []string
}

func (m *merkendeRegelnaht) Regelfreigabe(
	_ context.Context, _, _ string, _ int, modus string, abgewaehlt []string,
) (rest.Regelfreigabeangaben, error) {
	m.modus, m.abgewaehlt = modus, abgewaehlt

	return rest.Regelfreigabeangaben{}, nil
}

// dienstMitRegelnaht baut einen Dienst über einer merkenden Naht.
func dienstMitRegelnaht(t *testing.T, doppel *merkendeRegelnaht) *rest.Dienst {
	t.Helper()

	jetzt := time.Date(2026, time.September, 22, 12, 0, 0, 0, time.UTC)

	return rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{Regelfreigabe: doppel})
}

// **Modus und Abwahl kommen als Abfrageparameter an** — PLAY-02
// (#1710).
//
// **Kommagetrennt, und leere Stücke fallen weg**: „zonen,,combat" sind zwei
// Bereiche und nicht drei (§8).
func TestDieFreigabeLiestModusUndAbwahl(t *testing.T) {
	t.Parallel()

	gesehen := &merkendeRegelnaht{}

	ruf(t, dienstMitRegelnaht(t, gesehen), endpunkt{
		methode: http.MethodGet,
		weg: "/regelfreigabe?spiel=magic&strenge=sandkasten&sitze=2" +
			"&modus=frei&ohne=zonen,,%20combat%20",
	}, gueltig)

	if gesehen.modus != "frei" {
		t.Errorf("der Modus kam als %q an", gesehen.modus)
	}

	if len(gesehen.abgewaehlt) != 2 ||
		gesehen.abgewaehlt[0] != "zonen" || gesehen.abgewaehlt[1] != "combat" {
		t.Errorf("die Abwahl kam als %v an", gesehen.abgewaehlt)
	}
}

// **Ohne beides kommt nichts an** — und nicht eine leere Zeichenkette in
// einer Liste.
func TestOhneAngabeKommtKeineAbwahlAn(t *testing.T) {
	t.Parallel()

	gesehen := &merkendeRegelnaht{}

	ruf(t, dienstMitRegelnaht(t, gesehen), endpunkt{
		methode: http.MethodGet,
		weg:     "/regelfreigabe?spiel=magic&strenge=sandkasten&sitze=2",
	}, gueltig)

	if gesehen.modus != "" || gesehen.abgewaehlt != nil {
		t.Errorf("es kam %q und %v an", gesehen.modus, gesehen.abgewaehlt)
	}
}
