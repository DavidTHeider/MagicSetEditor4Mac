package netz_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Ein Klient mit Erlaubnis behält seine Fristen** — gefunden von D02
// (#2233)
// am nativen Fenster.
//
// Die Kartensuche meldete „»127.0.0.1:…« hat nicht binnen {sekunden}
// Sekunden geantwortet": der Platzhalter blieb stehen, weil der Fehler keine
// Frist trug. `MitErlaubnis` kopierte den Klienten Feld für Feld und ließ die
// Fristen aus — und genau so bauen Scryfall und die Kartendienste ihren
// Klienten. Die Kopffrist selbst griff; nur sagen konnte es niemand.
func TestEinKlientMitErlaubnisBehaeltSeineFristen(t *testing.T) {
	t.Setenv(netz.SchleifeUmgebung, "an")

	fristen := netz.Fristen{Verbindung: time.Second, Kopf: 300 * time.Millisecond, Stillstand: 600 * time.Millisecond}
	klient := netz.NeuerKlient(fristen).MitErlaubnis(func(context.Context) error { return nil })

	stumm := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	t.Cleanup(stumm.Close)

	_, fehler := netz.Hole(t.Context(), klient, nil, netz.Kenner{Name: "test", Annahme: "*/*"}, stumm.URL+"/karten", 1<<20)

	var gegenstelle *domain.Gegenstellenfehler
	if !errors.Is(fehler, domain.ErrKeineAntwort) || !errors.As(fehler, &gegenstelle) {
		t.Fatalf("der Fehler %v ist keine ausbleibende Antwort", fehler)
	}

	if gegenstelle.Frist != fristen.Kopf {
		t.Errorf("der Fehler nennt die Frist %v, erwartet die Kopffrist %v", gegenstelle.Frist, fristen.Kopf)
	}
}
