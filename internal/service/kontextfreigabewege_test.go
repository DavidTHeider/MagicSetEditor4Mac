package service_test

// **Der eigene Bestand geht nur mit Freigabe hinaus** — Werkstück A13.4 von
// Phase 5.3
// (#1452).
//
// # Was hier gemessen wird
//
// **Dass der Filter der Suche wirklich anders aussieht.** Die Domäne sagt,
// welche Quellen gesucht werden sollen; ob der Dienst danach fragt, sagt erst
// der Suchwunsch, der bei der Quelle ankommt.
//
// **Und dass die Auslassung in der Antwort steht.** Ein Korpus, in dem nicht
// gesucht wurde, erklärt, warum etwas fehlt — „nichts gefunden" erklärt es
// nicht (§8).

import (
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ohne Freigabe fragt der Dienst nicht nach eigenen Karten.**
func TestOhneFreigabeFragtDieRechercheNichtNachEigenen(t *testing.T) {
	t.Parallel()

	quelle := &recherchequellendoppel{stand: einAktiverStand()}

	ergebnis, fehler := einRecherchedienst(quelle, &ausgangsdoppel{}).
		Recherchieren(t.Context(), domain.Rechercheanfrage{
			Fall: domain.FallKarten, Frage: "Drache", Grenze: 3,
		}, "")
	if fehler != nil {
		t.Fatalf("recherchieren: %v", fehler)
	}

	if slices.Contains(quelle.gesehen.Filter.Quellen,
		domain.QuelleEigeneKarte) {
		t.Errorf("gefragt wurde nach %v", quelle.gesehen.Filter.Quellen)
	}

	// **Und die Antwort sagt es.**
	if !nenntEigeneKarten(ergebnis.Einschraenkungen) {
		t.Errorf("die Einschränkungen sind %v", ergebnis.Einschraenkungen)
	}
}

// **Mit Freigabe fragt er danach — und schweigt darüber.**
//
// **Die Gegenprobe zum Lauf darüber**: ohne sie sagte er nur, dass die Liste
// kürzer ist, und nicht, dass die Freigabe etwas bewirkt.
func TestMitFreigabeFragtDieRechercheNachEigenen(t *testing.T) {
	t.Parallel()

	quelle := &recherchequellendoppel{stand: einAktiverStand()}

	ergebnis, fehler := einRecherchedienst(quelle, &ausgangsdoppel{}).
		Recherchieren(t.Context(), domain.Rechercheanfrage{
			Fall: domain.FallKarten, Frage: "Drache", Grenze: 3,
			Freigegeben: []domain.Retrievalquelle{domain.QuelleEigeneKarte},
		}, "")
	if fehler != nil {
		t.Fatalf("recherchieren: %v", fehler)
	}

	if !slices.Contains(quelle.gesehen.Filter.Quellen,
		domain.QuelleEigeneKarte) {
		t.Errorf("gefragt wurde nach %v", quelle.gesehen.Filter.Quellen)
	}

	if nenntEigeneKarten(ergebnis.Einschraenkungen) {
		t.Errorf("die Freigabe stand da und wurde trotzdem beanstandet: %v",
			ergebnis.Einschraenkungen)
	}
}

// nenntEigeneKarten sagt, ob eine Einschränkung **die fehlende Freigabe** für
// den eigenen Bestand nennt.
//
// **Und nicht bloß den Namen der Quelle.** „Steht nicht im Index" ist eine
// andere Auskunft als „ist nicht freigegeben", und ein Lauf, der beide für
// dasselbe hielte, wäre grün, sobald irgendeine davon zutrifft.
func nenntEigeneKarten(saetze []string) bool {
	for _, satz := range saetze {
		if strings.Contains(satz, string(domain.QuelleEigeneKarte)) &&
			strings.Contains(satz, "nicht freigegeben") {
			return true
		}
	}

	return false
}
