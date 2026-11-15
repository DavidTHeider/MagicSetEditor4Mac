package app

// **Bereiche und Funktionen abschalten** — OPS-10
// (#1908,
// Vermerk [0158](../../docs/entscheidungen/0158-ein-schalter-verbirgt-er-sichert-nicht.md)).
//
// # Eine Wache für vier Flächen
//
// **Dieselbe Frage an derselben Stelle**: das Fenster fragt sie im Tor von
// [Bindungen.app], MCP vor dem Werkzeug, REST vor dem Endpunkt und die
// Kommandozeile vor dem Befehl — und jede bekommt dieselbe Abweisung.
//
// # Warum ein Lesefehler durchlässt
//
// **Weil ein Schalter nichts sichert.** Ein Bestand, der gerade nicht
// antwortet, soll das Fenster nicht leer machen; der Weg selbst stößt dann
// auf denselben Fehler und sagt ihn.

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"maps"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// SchalterDTO ist ein Schalter mit seinem wirksamen Stand.
type SchalterDTO struct {
	Kennung string `json:"kennung"`
	Art     string `json:"art"`
	Reife   string `json:"reife"`

	// Vorgabe ist `an` oder `aus` — was ohne eigene Wahl gilt.
	Vorgabe string `json:"vorgabe"`

	// An ist, was gerade gilt.
	An bool `json:"an"`
}

// SchalterAntwort ist das Verzeichnis samt Stand.
type SchalterAntwort struct {
	Schalter []SchalterDTO `json:"schalter"`
	Fehler   *Fehler       `json:"fehler"`
}

// werteleser ist, was die Wache von den Einstellungen braucht.
type werteleser interface {
	Alle(ctx context.Context) (map[string]string, error)
}

// Schalterwache prüft einen Weg gegen die Schalter.
type Schalterwache struct {
	einstellungen werteleser

	// gelesen hört jeden gelesenen Stand — das Fenster beendet damit, was
	// ein anderer Prozess abgeschaltet hat (OPS-10b, #1978).
	gelesen func(context.Context, domain.Schalterstand)
}

// SchalterwacheUeber baut die Wache über einem Bestand — für die Flächen
// ohne Fenster. Der Kontext jeder Frage trägt das Profil.
func SchalterwacheUeber(db *sql.DB) Schalterwache {
	return Schalterwache{einstellungen: einstellungsverwalterUeber(db)}
}

// Pruefe weist einen Weg ab, dessen Fähigkeiten alle abgeschaltet sind.
func (w Schalterwache) Pruefe(ctx context.Context, flaeche domain.Flaeche, weg string) error {
	return w.Tor(flaeche, "", nil)(ctx, weg)
}

// Tor gibt die Frage einer Fläche: erst ihr eigener Schalter, wenn sie einen
// hat — `dienst` für REST, `werkzeugserver` für MCP über HTTP —, dann der Weg.
//
// **Ist der eigene aus, endet die Fläche** (OPS-10.AK-03): `beenden` hält
// den Prozess an. Ein Dienst, der nur abwiese, käme beim Einschalten von
// selbst wieder.
//
// **Erst die Tafel, dann der Bestand**: ein Weg des Kerns auf einer Fläche
// ohne eigenen Schalter kostet keinen Lesevorgang. **Ohne Einstellungen ist
// alles an** — eine Anwendung, der niemand sie gesetzt hat, kennt keinen
// Schalter.
func (w Schalterwache) Tor(
	flaeche domain.Flaeche, eigener string, beenden func(),
) func(ctx context.Context, weg string) error {
	tafel := domain.EingebauteWegetafel()

	return func(ctx context.Context, weg string) error {
		if (eigener == "" && len(tafel[flaeche][weg]) == 0) || w.einstellungen == nil {
			return nil
		}

		werte, fehler := w.einstellungen.Alle(ctx)
		if fehler != nil {
			return nil //nolint:nilerr // ein Schalter sichert nichts — der Weg nennt den Fehler selbst
		}

		stand := domain.SchalterstandAus(werte)
		if w.gelesen != nil {
			w.gelesen(ctx, stand)
		}

		if eigener != "" && stand.Aus(eigener) {
			beenden()

			return &domain.Abschaltung{Schalter: eigener}
		}

		return tafel.Pruefe(flaeche, weg, stand)
	}
}

// Stand gibt das Verzeichnis mit dem, was gerade gilt.
func (w Schalterwache) Stand(ctx context.Context) SchalterAntwort {
	antwort := SchalterAntwort{Schalter: []SchalterDTO{}}

	var werte map[string]string

	if w.einstellungen != nil {
		var fehler error
		if werte, fehler = w.einstellungen.Alle(ctx); fehler != nil {
			antwort.Fehler = Uebersetze(ctx, fehler)

			return antwort
		}
	}

	stand := domain.SchalterstandAus(werte)

	for _, schalter := range domain.Schalterverzeichnis() {
		vorgabe := domain.SchalterAus
		if schalter.Vorgabe {
			vorgabe = domain.SchalterAn
		}

		antwort.Schalter = append(antwort.Schalter, SchalterDTO{
			Kennung: schalter.Kennung, Art: string(schalter.Art), Reife: string(schalter.Reife),
			Vorgabe: vorgabe, An: !stand.Aus(schalter.Kennung),
		})
	}

	return antwort
}

// Schalter gibt die Schalter des Fensters.
func (a *App) Schalter(ctx context.Context) SchalterAntwort {
	ctx = a.imProfil(ctx)

	return Schalterwache{einstellungen: a.einstellungen}.Stand(ctx)
}

// Befehlsweg sucht zu einem Befehl seinen Weg in der Tafel — den längsten,
// der passt. Ein Wort in spitzen Klammern steht für jedes.
//
// **Wörtlich vor Platzhalter**: `mse deck neu` ist ein anderer Befehl als
// `mse deck <id>`, auch wenn beide drei Wörter haben.
func Befehlsweg(befehl string, rest []string) (string, bool) {
	woerter := append([]string{"mse", befehl}, rest...)
	bester, bestePunkte := "", 0

	for weg := range domain.EingebauteWegetafel()[domain.FlaecheCLI] {
		if punkte := passt(strings.Fields(weg), woerter); punkte > bestePunkte {
			bester, bestePunkte = weg, punkte
		}
	}

	return bester, bester != ""
}

// passt gibt, wie gut ein Weg auf die Wörter passt — null, wenn gar nicht.
// Je Wort zählt ein wörtliches doppelt und ein Platzhalter einfach.
func passt(weg, woerter []string) int {
	if len(weg) > len(woerter) {
		return 0
	}

	punkte := 0

	for i, wort := range weg {
		switch {
		case wort == woerter[i]:
			punkte += 2
		case strings.HasPrefix(wort, "<") && strings.HasSuffix(wort, ">"):
			punkte++
		default:
			return 0
		}
	}

	return punkte
}

// abschaltung übersetzt die Abweisung — oder gibt nil, wenn es keine ist.
func abschaltung(fehler error) *Fehler {
	var abgeschaltet *domain.Abschaltung
	if !errors.As(fehler, &abgeschaltet) {
		return nil
	}

	uebersetzt := mitSatz(CodeAbgeschaltet, domain.SchluesselBereichAbgeschaltet, fehler)
	uebersetzt.Werte = map[string]string{"schalter": abgeschaltet.Schalter}

	return uebersetzt
}

// dienstAbgeschaltet übersetzt einen abgeschalteten Dienst — oder gibt nil
// (I05, #2263). **Derselbe Code wie ein abgeschalteter Bereich**, denn die
// Flächen fragen nach der Klasse: es ist abgeschaltet, und jemand kann es
// einschalten. Der Schlüssel ist ein eigener, weil der Ort ein anderer ist —
// Einstellungen › Dienste. Bis dahin stand `ErrDienstAus` in keiner
// Übersetzung, und der Rat kam als „unerwarteter Fehler" an.
func dienstAbgeschaltet(fehler error) *Fehler {
	var aus *domain.AbgeschalteterDienst
	if !errors.As(fehler, &aus) {
		return nil
	}

	uebersetzt := mitSatz(CodeAbgeschaltet, domain.SchluesselDienstAus, fehler)
	uebersetzt.Werte = map[string]string{"dienst": string(aus.Dienst)}

	return uebersetzt
}

// schalterWirkt beendet, was mit einem Schalter endet (OPS-10.AK-03,
// OPS-10b #1978). **Erst nach dem Schreiben**: ein Schalter, der nicht
// gespeichert wurde, beendet nichts (§7, Regel 3).
//
// **Nur „aus“ wirkt.** Eingeschaltet startet keiner von selbst, und ein
// Schalter, der angeht, sperrt nichts, was vorher offen war.
//
// **Der geschriebene Wert gilt, auch wenn der Stand nicht zu lesen ist** —
// ein Lesefehler darf keinen Port offen halten, den danach niemand mehr
// schließen kann. Die übrigen Schalter zählen dann als an.
func (a *App) schalterWirkt(ctx context.Context, schluessel, wert string) {
	if wert != domain.SchalterAus || !domain.IstSchaltereinstellung(schluessel) {
		return
	}

	werte := map[string]string{}

	alle, fehler := a.einstellungen.Alle(ctx)
	if fehler != nil {
		slog.WarnContext(ctx, "der Stand der Schalter ließ sich nicht lesen", "grund", fehler)
	}

	maps.Copy(werte, alle)
	werte[schluessel] = wert

	a.laeufeBeenden(ctx, domain.SchalterstandAus(werte))
}
