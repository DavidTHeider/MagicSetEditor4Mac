package app

// **Was an einem Schalter hängt und läuft, endet mit ihm** — OPS-10b
// (#1978).
//
// # Warum
//
// **Ein Lauf, dessen Beenden der Schalter sperrt, wäre danach nicht mehr
// anzuhalten.** Nach `netz` aus weist das Tor `HorchBeenden` ab, und der
// Horchposten hielte seinen Port offen, bis das Programm endet. Welche Läufe
// das sind, sagt die Domäne aus derselben Tafel, die das Tor fragt.
//
// # Zwei Anlässe
//
// **Das Fenster legt den Schalter selbst um** — dann beendet
// [App.schalterWirkt] nach dem Schreiben. **Oder ein anderer Prozess tut
// es** — `mse`, MCP und REST setzen dieselben Einstellungen —, und dann
// erfährt es das Fenster beim nächsten Weg, der den Stand liest: das Tor
// beendet nebenbei, was er sperrt.

import (
	"context"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// laeufeBeenden beendet jeden Lauf, dessen Beenden dieser Stand sperrt.
//
// **Ohne die Rechteprüfung der Bindungen**: wer einen Schalter umlegen
// darf, darf beenden, was daran hängt — und ein Lauf, den das Tor nicht
// mehr erreichen lässt, hat sonst niemanden, der ihn beendet. Jeder Weg ist
// einer, der nichts tut, wenn nichts läuft.
func (a *App) laeufeBeenden(ctx context.Context, stand domain.Schalterstand) {
	if stand.AllesAn() {
		return
	}

	beender := a.laufende()
	laeufe := domain.EingebauteWegetafel().GesperrteLaeufe(stand)

	// **Erst zählen, dann beenden** (Nachlauf über Phase 6.12, N2): ein
	// Start, der vor dem Zählen seine Marke nahm, sieht danach eine andere.
	a.generationen.weiter(laeufe)

	for _, lauf := range laeufe {
		beender[lauf](ctx)
	}
}

// laufende ordnet jedem Lauf zu, wie er endet — dieselben Wege wie hinter
// seiner Beenden-Bindung.
func (a *App) laufende() map[domain.Lauf]func(context.Context) {
	return map[domain.Lauf]func(context.Context){
		domain.LaufPartiebeitritt: a.partiebeitrittBeenden,
		// **Und die Gäste gehen mit** (Zweitblick über Phase 6.12, Z3): das
		// Ende der Freigabe allein schickt niemanden vom Tisch, das Abschalten
		// soll es.
		domain.LaufPartiefreigabe: func(context.Context) {
			a.partiefreigabeBeenden()
			a.partie.runde.Entlasse()
		},
		domain.LaufSichtbarkeit: func(context.Context) { a.sichtbarkeitBeenden() },
		domain.LaufHorchen:      func(context.Context) { a.horchenBeenden() },
		domain.LaufFernhorchen:  func(context.Context) { a.fernhorchenBeenden() },
		domain.LaufIndexaufbau: func(context.Context) {
			a.indexabbruch.ausloesen()
			a.IndexAbbrechen()
		},
		domain.LaufBildauftraege: func(context.Context) { a.bildlaeufe.alleAbbrechen() },
		domain.LaufREST:          func(context.Context) { a.dienstBeenden(domain.DienstREST) },
		domain.LaufMCP:           func(context.Context) { a.dienstBeenden(domain.DienstMCP) },
	}
}

// nachDemStart fragt nach dem Eintragen eines Laufs noch einmal nach seinem
// Schalter — und beendet ihn, wenn er inzwischen gesperrt ist **oder
// dazwischen gesperrt war**. Der Rückgabewert sagt, ob er beendet wurde.
//
// **Zweitblick über Phase 6.12, Z2.** Das Tor fragt vor dem Start, das
// Abschalten beendet, was eingetragen ist; dazwischen liegt der Start selbst.
// Weil das Abschalten **erst schreibt, dann zählt und dann beendet** und
// dieser Weg **erst seine Marke nimmt, dann einträgt und dann fragt**, sieht
// eines von beiden das andere.
//
// **Nachlauf, N2: der Stand allein reichte nicht.** Aus und wieder an,
// während ein Start wartete, ließ ihn weiterlaufen — deshalb die Marke. Und
// beendet wird **nur der eigene Lauf** (`ende`), nicht die ganze Art: eine
// verspätete Nachfrage hätte sonst einen neuen Bildauftrag mitgenommen.
//
// **Ohne eigene Meldung**: der Start antwortet mit dem Stand danach, und der
// sagt „läuft nicht". Den Grund kennt, wer gerade abgeschaltet hat.
func (a *App) nachDemStart(
	ctx context.Context, lauf domain.Lauf, marke uint64, ende func(context.Context),
) bool {
	// **Und nach dem Beginn des Schlusses endet jeder Start selbst**
	// (Zweitblick über A und B, Befund 2): wer vor dem Schluss durchs Tor
	// kam und erst danach eintrug, hätte einen Lauf hinterlassen, den der
	// Schluss nicht mehr sah. Das Tor wird **vor** dem Beenden geschlossen
	// (`Schliessen`), also sieht eines von beiden den Lauf. **Dasselbe
	// während eines Tauschs** (A02): der Bestand darunter ist gleich ein
	// anderer.
	if a.lauf.tor.Load() != torOffen {
		ende(ctx)

		return true
	}

	gesperrt := false

	if a.einstellungen != nil {
		werte, _ := a.einstellungen.Alle(ctx) //nolint:errcheck // ein Schalter sichert nichts — ohne Stand gilt alles als an
		gesperrt = domain.EingebauteWegetafel().Pruefe(domain.FlaecheFenster, lauf.Beendenweg(),
			domain.SchalterstandAus(werte)) != nil
	}

	if !gesperrt && a.laufmarke(lauf) == marke {
		return false
	}

	ende(ctx)

	return true
}

// laufmarke ist die Generation eines Laufs — wie oft ein Abschalten ihn
// beendet hat.
func (a *App) laufmarke(lauf domain.Lauf) uint64 {
	a.generationen.schloss.Lock()
	defer a.generationen.schloss.Unlock()

	return a.generationen.je[lauf]
}

// laufbuch hält, was an Läufen unterwegs ist: die Bildaufträge je Karte und
// die Generation je Lauf.
type laufbuch struct {
	bildlaeufe   bildlaeufe
	generationen laufgenerationen
}

// laufgenerationen zählt je Lauf, wie oft ein Abschalten ihn beendet hat.
type laufgenerationen struct {
	schloss sync.Mutex
	je      map[domain.Lauf]uint64
}

func (g *laufgenerationen) weiter(laeufe []domain.Lauf) {
	g.schloss.Lock()
	defer g.schloss.Unlock()

	if g.je == nil {
		g.je = map[domain.Lauf]uint64{}
	}

	for _, lauf := range laeufe {
		g.je[lauf]++
	}
}

// laufDes nennt den Lauf eines Dienstes.
func laufDes(art domain.Dienstart) domain.Lauf {
	if art == domain.DienstMCP {
		return domain.LaufMCP
	}

	return domain.LaufREST
}

// bildlaufAnmelden meldet einen Bildauftrag an und fragt danach nach `ki`
// — beendet wird dabei nur dieser Auftrag.
func (a *App) bildlaufAnmelden(ctx context.Context, karteID int64) (
	abbrechbar context.Context, fertig func(), fehler error,
) {
	marke := a.laufmarke(domain.LaufBildauftraege)

	abbrechbar, fertig, fehler = a.bildlaeufe.anmelden(ctx, karteID)
	if fehler == nil {
		a.nachDemStart(ctx, domain.LaufBildauftraege, marke, func(context.Context) {
			a.bildlaeufe.abbrechen(karteID)
		})
	}

	return abbrechbar, fertig, fehler
}

// abbruchplatz hält den Abbruch eines Laufs, der gerade läuft.
type abbruchplatz struct {
	schloss   sync.Mutex
	abbrechen context.CancelFunc
}

func (p *abbruchplatz) setzen(abbrechen context.CancelFunc) {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	p.abbrechen = abbrechen
}

func (p *abbruchplatz) ausloesen() {
	p.schloss.Lock()
	abbrechen := p.abbrechen
	p.schloss.Unlock()

	if abbrechen != nil {
		abbrechen()
	}
}
