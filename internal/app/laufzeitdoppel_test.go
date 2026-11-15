// Das Doppel der KI-Laufzeit — herausgezogen aus `kilaufzeit_test.go`, weil
// es dort über die Zeilengrenze aus §2 ging (A13.2a, #1452).
//
// **Der Schnitt folgt der Frage**: dort stehen die Läufe über die
// Einrichtung, hier die Vorrichtung, die sie **und** die Läufe über
// Vorschlag, Batch und Einschätzung benutzen.
package app_test

import (
	"context"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// laufzeitdoppel steht für eine Ollama-Laufzeit.
//
// **Es hält ein Schloss** — Werkstück A13.1 von Phase 5.3
// (#1452).
// Ein Batch fragt dieselbe Quelle **gleichzeitig**, und die Zähler darunter
// liefen sonst in den Race-Detektor. Der echte Klient ist nebenläufig
// benutzbar; ein Doppel, das es nicht ist, wäre enger als der Betrieb (§3).
type laufzeitdoppel struct {
	// schloss hält die Zähler, nicht die Antworten: die stehen fest, bevor
	// ein Lauf beginnt.
	schloss sync.Mutex

	// vorZaehler zählt gleichzeitige Rufe — nil heißt: keiner zählt.
	vorZaehler *gleichzeitig

	fassung string
	modelle []domain.KIModell
	befunde []domain.Befund
	fehler  error

	// listenfehler scheitert **nur** beim Lesen der Modelle — die Fassung
	// kommt. Das ist etwas anderes als eine Laufzeit, die gar nicht
	// antwortet, und die Naht unterscheidet es.
	listenfehler error

	gefragt  int
	adressen []string
	schemen  []map[string]any
	antwort  app.KIRohantwort

	// anweisungen hält, welche Systemzeile jeder Ruf mitbrachte — A13.2a
	// (#1452). **Seit sie ein Parameter ist**, ist sie eine Aussage über den
	// Ruf und keine Eigenschaft der Naht: eine Einschätzung, die die
	// Kartenanweisung bekäme, fiele sonst nirgends auf.
	anweisungen []string

	// auftraege hält den **Rumpf** jedes Rufs — A13.2b (#1452).
	//
	// **Ohne ihn war unsichtbar, was das Modell wirklich bekam.** Eine
	// Gegenprobe, die das Format aus der Frage nahm, blieb grün: der Kasten
	// zeigte „Format: Commander", und im Auftrag stand es nicht (§10).
	auftraege []string

	// gesehene hält, was an die Sichtquelle ging — A13.3 (#1452).
	gesehene []app.KISichtwunsch

	// redefehler scheitert **nur** beim Reden — Fassung und Modelle kommen.
	redefehler error
}

func (l *laufzeitdoppel) Fassung(context.Context) (string, error) {
	if l.fehler != nil {
		return "", l.fehler
	}

	return l.fassung, nil
}

func (l *laufzeitdoppel) Modelle(context.Context) (
	[]domain.KIModell, []domain.Befund, error,
) {
	l.schloss.Lock()
	l.gefragt++
	l.schloss.Unlock()

	if l.fehler != nil {
		return nil, nil, l.fehler
	}

	if l.listenfehler != nil {
		return nil, nil, l.listenfehler
	}

	return l.modelle, l.befunde, nil
}

func (l *laufzeitdoppel) Vorschlagen(
	_ context.Context, anweisung, alias, auftrag string, schema map[string]any,
) (app.KIRohantwort, error) {
	if l.vorZaehler != nil {
		l.vorZaehler.an()
		defer l.vorZaehler.aus()

		// **Kurz stehen bleiben, damit Gleichzeitigkeit messbar ist.** Eine
		// Antwort, die sofort kommt, überlappt mit keiner — und der
		// Höchststand wäre eins, gleich welche Schranke gilt (§10).
		time.Sleep(2 * time.Millisecond)
	}

	l.schloss.Lock()
	l.adressen = append(l.adressen, alias)
	l.schemen = append(l.schemen, schema)
	l.anweisungen = append(l.anweisungen, anweisung)
	l.auftraege = append(l.auftraege, auftrag)
	l.schloss.Unlock()

	if l.fehler != nil {
		return app.KIRohantwort{}, l.fehler
	}

	if l.redefehler != nil {
		return app.KIRohantwort{}, l.redefehler
	}

	return l.antwort, nil
}

// Sehen macht das Doppel zur Sichtquelle — Werkstück A13.3 (#1452).
//
// **Es hält das Bild fest**, denn die Zusage „das Bild geht mit" ist sonst
// nicht zu messen: eine Antwort käme auch ohne eines.
func (l *laufzeitdoppel) Sehen(
	_ context.Context, wunsch app.KISichtwunsch,
) (app.KIRohantwort, error) {
	l.schloss.Lock()
	l.gesehene = append(l.gesehene, wunsch)
	l.anweisungen = append(l.anweisungen, wunsch.Anweisung)
	l.schemen = append(l.schemen, wunsch.Schema)
	l.schloss.Unlock()

	if l.fehler != nil {
		return app.KIRohantwort{}, l.fehler
	}

	if l.redefehler != nil {
		return app.KIRohantwort{}, l.redefehler
	}

	return l.antwort, nil
}
