// **Ein Booster aus dem offenen Set** —
// #1072,
// Werkstück 4 von Phase W.
//
// # Die Naht
//
// Drei Dinge treffen sich hier:
//
//   - das **Spielpaket** sagt, welche Packarten es gibt (`magic.mse-game/packs`
//     führt 30);
//   - jede Packart trägt einen **Filter** — einen MSE-Script-Ausdruck über
//     `card`, der sagt, welche Karten in Frage kommen;
//   - das **Set** liefert die Karten.
//
// Der Zieher selbst kennt keinen Filter (#1072): ihn dort auszuwerten hieße,
// den Planbauer und die ganze Umgebung dorthin zu holen. Er kommt hier zum
// Einsatz — an der Stelle, die beides schon hat.

package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Packangebot ist eine Packart, wie die Oberfläche sie anbietet.
type Packangebot struct {
	Name string

	// Waehlbar sagt, ob MSE sie zum Ziehen anbietet — `selectable: true`.
	//
	// **Im Pack sind es genau vier von dreißig**: `booster pack`,
	// `tournament pack`, `fat pack` und `custom pack`. Die übrigen sind
	// Bausteine, die von jenen aus gezogen werden.
	Waehlbar bool

	// Karten ist die Zahl der Karten des Sets, aus denen diese Packart
	// **erreichbar** zieht — die eigenen und die aller Posten, ohne Doppelte.
	//
	// **Eigene hat, wer einen Filter hat**, und sonst niemand — auch eine
	// Packart mit Posten (`src/data/pack.cpp`, Zeilen 110–117, Fassung f9d9356
	// vom 2021-07-06).
	//
	// **Sie steht hier, weil eine leere Packart nichts zieht** und niemand
	// sieht, warum: ein Set ohne seltene Karten gibt kein Booster.
	//
	// **Erreichbar und nicht eigen**: eine Packart mit Posten hat gar keine
	// eigenen ([Packbereich.gefilterteKarten]), und „Booster — 0 Karten" wäre
	// genau die Auskunft, die diese Zahl verhindern soll.
	Karten int
}

// Packbereich zieht Booster aus einem Set.
type Packbereich struct {
	karten Kartensetleser
	planer *Planbauer
	sets   Setkartenquelle

	// speicher hält das zuletzt gerechnete Angebot je Set — der Grund steht
	// bei [packspeicher].
	speicher packspeicher

	// arbeiter ist, wie viele Karten zugleich ausgewertet werden
	// ([Packbereich.MitArbeitern]).
	arbeiter int
}

// Setkartenquelle gibt die Karten eines Sets — und seine eigenen Packarten.
//
// **Beides an einer Schnittstelle**, weil beides aus demselben Leser kommt und
// eine zweite Naht daneben nur einen zweiten Konstruktorparameter kostete.
type Setkartenquelle interface {
	Karten(ctx context.Context, setID int64) ([]*domain.CardRow, error)

	// Setpackarten gibt die Packarten, die das Set selbst mitbringt
	// (#1081).
	Setpackarten(ctx context.Context, setID int64) ([]*domain.PackArt, error)
}

// NeuerPackbereich baut ihn.
func NeuerPackbereich(
	karten Kartensetleser, sets Setkartenquelle, planer *Planbauer,
) *Packbereich {
	return &Packbereich{karten: karten, sets: sets, planer: planer}
}

// Angebot gibt die Packarten des Spiels samt der Zahl passender Karten.
func (p *Packbereich) Angebot(ctx context.Context, setID int64) (
	[]Packangebot, []domain.Befund, error,
) {
	arten, karten, vorlage, befunde, err := p.aufstellung(ctx, setID)
	if err != nil {
		return nil, befunde, err
	}

	gefiltert, filterbefunde, filterfehler := p.gefilterteKarten(ctx, setID, arten, karten, vorlage)
	if filterfehler != nil {
		return nil, append(befunde, filterbefunde...), filterfehler
	}

	angebot := make([]Packangebot, 0, len(arten))

	nachName := make(map[string]msepaket.Pack, len(arten))
	for _, art := range arten {
		nachName[art.Name] = art
	}

	for _, art := range arten {
		angebot = append(angebot, Packangebot{
			Name: art.Name, Waehlbar: art.Waehlbar,
			Karten: erreichbareKarten(art, nachName, gefiltert, map[string]bool{}),
		})
	}

	return angebot, append(befunde, filterbefunde...), nil
}

// Ziehen zieht ein Pack und gibt die Karten in ihrer Ziehreihenfolge.
//
// **Der Same kommt von außen** (#1072): derselbe Same gibt dieselbe Ziehung.
// Ohne ihn wäre kein Test möglich — und niemand könnte ein Ergebnis noch
// einmal herstellen.
func (p *Packbereich) Ziehen(ctx context.Context, setID int64, art string, same uint64) (
	[]*domain.CardRow, []domain.Befund, error,
) {
	arten, karten, vorlage, befunde, err := p.aufstellung(ctx, setID)
	if err != nil {
		return nil, befunde, err
	}

	gefiltert, filterbefunde, filterfehler := p.gefilterteKarten(ctx, setID, arten, karten, vorlage)
	if filterfehler != nil {
		return nil, append(befunde, filterbefunde...), filterfehler
	}

	befunde = append(befunde, filterbefunde...)

	gezogen, ziehfehler := NeuerPackzieher(arten, gefiltert).Ziehen(art, same)

	return gezogen, befunde, ziehfehler
}

// aufstellung holt Packarten, Karten und Vorlage in einem Zug.
func (p *Packbereich) aufstellung(ctx context.Context, setID int64) (
	[]msepaket.Pack, []*domain.CardRow, Vorlage, []domain.Befund, error,
) {
	karten, err := p.sets.Karten(ctx, setID)
	if err != nil {
		return nil, nil, Vorlage{}, nil, err
	}

	// **Die eigenen zuerst, auch ohne Vorlage** (#1081). Ein Set kann eine
	// Packart nennen, ohne dass sein Stil auffindbar ist — und dann soll die
	// eigene trotzdem gelten. Der erste Anlauf holte sie erst hinter der
	// Vorlagenprüfung, und ein Set ohne Stil zog aus nichts.
	eigene, eigenfehler := p.sets.Setpackarten(ctx, setID)
	if eigenfehler != nil {
		return nil, karten, Vorlage{}, nil, eigenfehler
	}

	vorlage, befunde, da := p.setvorlage(ctx, setID)
	if !da {
		return PackartenMitVorrang(eigene, nil), karten, Vorlage{}, befunde, nil
	}

	return PackartenMitVorrang(eigene, vorlage.Packs), karten, vorlage, befunde, nil
}

// setvorlage gibt den Stil des Sets — leer, wenn es keinen gibt.
func (p *Packbereich) setvorlage(ctx context.Context, setID int64) (
	Vorlage, []domain.Befund, bool,
) {
	if p.planer == nil || p.planer.stile == nil {
		return Vorlage{}, nil, false
	}

	return p.planer.stile.StilFuerSet(ctx, setID)
}

// gefilterteKarten wertet den Filter jeder Packart über jeder Karte aus.
//
// **Eigene Karten hat, wer einen Filter hat** — und sonst niemand.
//
//	// Filter cards
//	if (pack_type.filter) {
//	  FOR_EACH(card, parent.set->cards) { … }
//	}
//
// `src/data/pack.cpp`, Zeilen 110–117, Fassung f9d9356 vom 2021-07-06. Die
// Posten ändern daran **nichts**; `generate_all` hängt danach die eigenen
// Karten an und verlangt zusätzlich die Posten (Zeilen 317–328).
//
// # Zwei Abweichungen standen hier, und beide sahen aus wie gelesene Regeln
//
// **„Eine Packart mit Posten hat keine eigenen Karten."** Das galt nur für
// `select: auto`: dort **wird** eine Postenart zu `all` — aber nur, wenn kein
// Filter dasteht (Zeilen 43–46). Wer beides hat, bekommt bei MSE beides.
// Angestrichen vom Zweitblick durch Codex über Phase W.
//
// **„Ohne Filter und ohne Posten gehört jede Karte dazu."** Bei MSE bleibt
// `cards` dann leer, und `total_weight` wird 0 — eine solche Packart zieht
// **nichts**. Der Satz war erschlossen, nicht nachgelesen.
//
// **Gemessen am 2026-08-24** an den 29 Packarten in
// `M15-Magic-Pack/data/magic.mse-game/packs`: **keine** trägt Filter und
// Posten zugleich, und **keine** trägt weder noch. Beide Abweichungen sind im
// echten Pack folgenlos; sie werden behoben, weil eine erschlossene Regel mit
// einer Fundstelle daneben schlimmer ist als gar keine.
//
// # Die Umgebung steht außen, die Packart innen
//
// Und das ist keine Geschmacksfrage. `magic.mse-game` führt **30** Packarten;
// `Vanguard.mse-set` hat 107 Karten. Wer je Packart über alle Karten liefe,
// baute 3.210 Umgebungen — und die Umgebung ist das Teure: 249 Kartenfelder,
// ihre Wertskripte, das Initskript des Stils.
//
// **So sind es 107.** Gemessen am 2026-08-24 in der laufenden Anwendung: mit
// der Umgebung innen stand die Seite über eine halbe Minute auf „Wird geladen".
//
// **Ein gescheiterter Filter lässt die Packart leer** und bekommt einen Befund
// (§8) — nicht alle Karten. Wer im Zweifel alles durchließe, zöge aus einem
// Ausdruck, den niemand ausrechnen konnte.
//
// **Ein Abbruch ist kein Befund, sondern ein Fehler.** Er sagt nichts über die
// Packart; ihn als ihren Mangel zu führen wäre eine Falschaussage (§7).
func (p *Packbereich) gefilterteKarten(
	ctx context.Context, setID int64,
	arten []msepaket.Pack, karten []*domain.CardRow, vorlage Vorlage,
) (map[string][]*domain.CardRow, []domain.Befund, error) {
	gefiltert := make(map[string][]*domain.CardRow, len(arten))
	for _, art := range arten {
		gefiltert[art.Name] = []*domain.CardRow{}
	}

	mitFilter, zerteilt, gescheitert := zerteilteFilter(arten)

	if len(mitFilter) == 0 || len(karten) == 0 {
		return gefiltert, gescheiterteBefunde(mitFilter, gescheitert, gefiltert), nil
	}

	// **Hier und nicht in [Packbereich.Angebot]**, weil [Packbereich.Ziehen]
	// durch dieselbe Rechnung geht: die Oberfläche ruft erst `Packarten` und
	// dann beim Klick `PackZiehen`, und der zweite Lauf rechnete alles noch
	// einmal. Angestrichen vom Nachlauf über Phase 4.1, Befund 5.3.
	schluessel := p.packschluessel(ctx, arten, karten, vorlage)
	if gemerkt, gemerkteBefunde, da := p.speicher.hole(setID, schluessel); da {
		return gemerkt, gemerkteBefunde, nil
	}

	// **Ein Bauer je Stilbogen** — die Initskripte laufen einmal je Stil statt
	// einmal je Karte. Der Posten war 1.905 von 4.051 ms
	// ([umgebungsbauer], #1304).
	//
	// **Je Stilbogen und nicht je Set**, und das ist nachgelesen: MSE2 hält
	// seine Kontexte in einer Abbildung über den Stilbogen und wählt für jede
	// Karte zuerst `stylesheetForP(card)` (`src/script/script_manager.cpp`,
	// Zeilen 30 und 64, Fassung f9d9356 vom 2021-07-06). Eine Karte darf
	// einen **eigenen** Stil tragen ([Stilquelle.StilFuerKarte]) — dann gilt
	// dessen Initskript und nicht das des Sets. Angestrichen vom Zweitblick
	// über Phase 4.1, Befund 1.
	urteile, abgebrochen := p.filterlauf(ctx, karten, mitFilter, zerteilt, vorlage)
	if abgebrochen != nil {
		return nil, nil, abgebrochen
	}

	for stelle, karte := range karten {
		for art, filter := range mitFilter {
			if urteile[stelle].grund[art] != "" {
				gescheitert[filter.Name] = urteile[stelle].grund[art]
			}

			if urteile[stelle].trifft[art] {
				gefiltert[filter.Name] = append(gefiltert[filter.Name], karte)
			}
		}
	}

	befunde := gescheiterteBefunde(mitFilter, gescheitert, gefiltert)

	p.speicher.lege(setID, schluessel, gefiltert, befunde)

	return gefiltert, befunde, nil
}

// gescheiterteBefunde nennt die Packarten, deren Filter nicht durchlief.
//
// **Eine Meldung je Packart und nicht je Karte.** Ein Filter, der an einer
// Karte scheitert, scheitert an allen — und 107 gleichlautende Zeilen wären
// eine Liste, die niemand liest (§8).
//
// **Und die Packart bleibt leer.** Was der Filter bis dahin durchgelassen hat,
// wäre ein halber Stand: gezogen würde aus dem, was zufällig vor dem Fehler
// kam.
func gescheiterteBefunde(
	mitFilter []msepaket.Pack, gescheitert map[string]string,
	gefiltert map[string][]*domain.CardRow,
) []domain.Befund {
	var befunde []domain.Befund

	for _, art := range mitFilter {
		grund, kaputt := gescheitert[art.Name]
		if !kaputt {
			continue
		}

		gefiltert[art.Name] = nil

		befunde = append(befunde, domain.Befund{
			Ort:     art.Name,
			Meldung: "der Filter ließ sich nicht ausrechnen: " + grund,
		})
	}

	return befunde
}

// erreichbareKarten zählt die Karten, aus denen eine Packart ziehen kann.
//
// **Die eigenen und die aller Posten, ohne Doppelte** — zwei Posten dürfen
// dieselbe Packart nennen, und dann wäre die Summe zu groß.
//
// **Der Kreis wird abgeschnitten, nicht durchlaufen.** `Packzieher.staende`
// weist ihn ab; hier ist die Frage nur, wie viele Karten es sind, und eine
// endlose Rekursion wäre die falsche Antwort darauf.
func erreichbareKarten(
	art msepaket.Pack, nachName map[string]msepaket.Pack,
	gefiltert map[string][]*domain.CardRow, gesehen map[string]bool,
) int {
	return len(kartenmenge(art, nachName, gefiltert, gesehen))
}

// kartenmenge sammelt die Kennungen der erreichbaren Karten.
func kartenmenge(
	art msepaket.Pack, nachName map[string]msepaket.Pack,
	gefiltert map[string][]*domain.CardRow, gesehen map[string]bool,
) map[int64]bool {
	menge := map[int64]bool{}

	if gesehen[art.Name] {
		return menge
	}

	gesehen[art.Name] = true

	for _, karte := range gefiltert[art.Name] {
		menge[karte.ID] = true
	}

	for _, posten := range art.Posten {
		unten, da := nachName[posten.Name]
		if !da {
			continue
		}

		for kennung := range kartenmenge(unten, nachName, gefiltert, gesehen) {
			menge[kennung] = true
		}
	}

	return menge
}
