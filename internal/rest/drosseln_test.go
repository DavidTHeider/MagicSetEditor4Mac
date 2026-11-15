package rest_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Aufrufer ohne Marke kann den mit Marke nicht aussperren**
// (Zweitblick über Phase U).
//
// # Der Befund
//
// Die Drossel stand als **ein** Zähler vor der Markenprüfung. Damit konnte
// jeder lokale Prozess 60 Anfragen ohne Marke schicken und danach war der
// legitime Aufrufer für eine Minute draußen — ohne dass der Angreifer je eine
// Marke gebraucht hätte.
//
// # Warum die Drossel trotzdem vor der Marke steht
//
// Weil sie genau deren Prüfung schützt: eine Drossel dahinter zählte nur die
// erfolgreichen Versuche und ließe das Raten ungezählt.
//
// **Die Lösung sind zwei Budgets**, nicht eine andere Reihenfolge: wer keine
// gültige Marke vorzeigt, zahlt aus dem einen; wer eine hat, aus dem anderen.
func TestEinFremderKannDenBerechtigtenNichtAussperren(t *testing.T) {
	t.Parallel()

	dienst := neuerDienst(t)
	ziel := endpunkt{methode: http.MethodGet, weg: "/sets"}

	// Das Budget der Fehlversuche wird ausgeschöpft.
	for range 61 {
		ruf(t, dienst, ziel, ohneMarke)
	}

	if antwort := ruf(t, dienst, ziel, ohneMarke); antwort.Code != http.StatusTooManyRequests {
		t.Errorf("der Fremde wurde nicht gedrosselt: %d", antwort.Code)
	}

	// **Und der Berechtigte kommt trotzdem durch.**
	if antwort := ruf(t, dienst, ziel, gueltig); antwort.Code != http.StatusOK {
		t.Errorf("der Berechtigte wurde ausgesperrt: %d (%s)",
			antwort.Code, antwort.Body.String())
	}
}

// **Und umgekehrt sperrt der Berechtigte den Fremden nicht aus.**
//
// Ohne diesen Fall wäre auch eine Bauart grün, die dem Fremden ein eigenes,
// aber unbegrenztes Budget gäbe.
func TestAuchMitMarkeGiltEineGrenze(t *testing.T) {
	t.Parallel()

	// **Seit O5 ein Guthaben** (#1429): drei Einheiten, eine je Sekunde nach.
	dienst := mitHaushalt(t, neuerDienst(t), engeGrenzen(), festeUhr)
	ziel := endpunkt{methode: http.MethodGet, weg: "/sets"}

	for range 3 {
		if antwort := ruf(t, dienst, ziel, gueltig); antwort.Code != http.StatusOK {
			t.Fatalf("im Guthaben und doch %d", antwort.Code)
		}
	}

	antwort := ruf(t, dienst, ziel, gueltig)
	if antwort.Code != http.StatusTooManyRequests {
		t.Errorf("der vierte Aufruf mit Marke kam durch: %d", antwort.Code)
	}

	// **Die Wartezeit ist echt**: eine Einheit je Sekunde, also eine Sekunde.
	if antwort.Header().Get("Retry-After") != "1" {
		t.Errorf("Retry-After: %q", antwort.Header().Get("Retry-After"))
	}
}

// engeGrenzen sind Grenzen, an denen ein Test in wenigen Rufen ankommt.
func engeGrenzen() domain.Haushaltsgrenzen {
	grenzen := domain.Vorgabegrenzen()
	grenzen.Principal = domain.Kontingent{Burst: 3, Nachfuellen: 60, Fenster: time.Minute}

	return grenzen
}
