package app

// **Die Verdrahtung der Deckwege** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei neben `start.go`**, aus demselben Grund wie
// `scryfallstart.go`: dort steht, **was** beim Start entsteht, hier, wie die
// Deckwege verdrahtet werden (§2).

import (
	"database/sql"
	"log/slog"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// deckwegeStarten hängt Deckbau, Formatpflege, Prüfung und Abgleich ein.
//
// **Vier Nähte aus drei Ablagen.** Der Deckbau braucht die Formate zum Lesen,
// die Prüfung braucht Scryfall und die Verweise, der Abgleich die Sammlung —
// und alle vier gehen über **dieselben** Ablagen wie der Rest der Anwendung.
//
// **Der Zeichenleser wird hier nicht gebraucht** und steht trotzdem in der
// Signatur: `start.go` reicht ihn an jede Startfunktion, und eine Ausnahme für
// diese eine wäre eine Stelle, an der man nachsehen muss.
func (a *App) deckwegeStarten(db *sql.DB, _ Zeichenleser) {
	var (
		decks     = storage.NeueDeckablage(db)
		formate   = storage.NeueFormatablage(db)
		setzeiger = storage.NeuerSetLeser(db)
	)

	pflege := service.NeueFormatpflege(formate)

	a.MitFormatnaht(pflege)

	// **Der Deckbau kennt die Namen eigener Karten von Anfang an** (#1376):
	// sie stehen im Bestand und brauchen kein Netz. Die Namen echter Drucke
	// kommen weiter unten dazu, wenn ein Klient zustande kommt.
	bau := service.NeuerDeckbau(decks, setzeiger, formate).
		MitKartennamen(setzeiger)

	a.MitDecknaht(bau)

	// **Aus einem Set ein Deck** (FMT-04, #1693) — dieselbe Ablage, derselbe
	// Zug wie bei der Übernahme aus einem fremden Profil.
	a.MitSetdeck(service.NeuesSetdeck(setzeiger, decks))

	// **Turniere hängen an keiner Gegenstelle** (Phase 4.7, #1294): sie
	// lesen Decks und Karten aus dem eigenen Bestand. Deshalb steht diese
	// Naht **vor** dem Scryfall-Zweig — der kann abbrechen, und ein Turnier
	// ohne Netz ist immer noch ein Turnier.
	a.MitTurniernaht(service.NeueTurnierleitung(
		storage.NeueTurnierablage(db), decks, storage.NeuerSetLeser(db)).
		MitFormaten(formate))

	// **Und Partien ebenso** (G1, #1458):
	// sie lesen Decks aus dem eigenen Bestand und schreiben in ihn. Aus
	// demselben Grund wie die Turniere stehen sie **vor** dem
	// Scryfall-Zweig — eine Partie ohne Netz ist immer noch eine Partie.
	a.MitPartieaufbau(service.NeuerPartieaufbau(decks).
		MitBlaettern(storage.NeuerSetLeser(db)).
		MitLeihen(a.leihstelle()))

	// **Und die Brücke zwischen beiden** (G7,
	// #1464):
	// sie steht hier, weil sie Turnier **und** Partie braucht — und beide
	// sind eben entstanden.
	a.MitMatchnaht(service.NeueMatchleitung(
		storage.NeueTurnierablage(db),
		storage.NeuePartieablage(db),
		storage.NeueMatchablage(db)))
	a.MitPartienaht(service.NeuePartieleitung(
		storage.NeuePartieablage(db),
		service.NeuerSandkasten(service.NeuerZufall())))

	verweise := storage.NeueKartenverweise(db)

	// **Der Abgleich hängt an keiner Gegenstelle** und wird deshalb immer
	// gesetzt: er liest die eigene Sammlung.
	sammlung := storage.NeuerSammlungsspeicher(db)
	abgleich := service.NeuerHabenabgleich(decks, sammlung, verweise)

	a.MitDeckabgleich(abgleich)

	// **Und was fehlt, auf die Wunschliste** — DOM-03.AK-02 (#1698). Die
	// Namen kommen vom Deckbau, der sie seit #1376 einträgt.
	a.MitFehlkartenwunsch(service.NeuerFehlkartenwunsch(abgleich, bau, sammlung))

	// **Und die Zuordnung ebenso** — Werkstück U9 von Phase 5.1 (#1420),
	// **samt der Naht zum Handeln** seit `T2` (#1467, 0120): was ein
	// Handelsvorgang festhält, ist hier nicht mehr null.
	a.MitDeckzuordnung(service.NeueDeckzuordnungspflege(
		decks, storage.NeueDeckzuordnungen(db), sammlung, abgleich).
		MitReservierungen(storage.NeueReservierungsablage(db)))

	// **Die Prüfung braucht Scryfall, und der ist wahlfrei.** Wer kein Zuhause
	// hat, bekommt keinen Klienten — dann bleibt diese Naht leer, und
	// `DeckPruefen` meldet es beim ersten Versuch statt beim Start. Dieselbe
	// Haltung wie bei den übrigen Scryfall-Wegen.
	//
	// **Und seit Phase 4.8 gilt dasselbe für den Schalter**: ist Scryfall
	// abgeschaltet, entsteht hier kein Klient — und damit geht kein Aufruf
	// hinaus (0063, Entscheidung 1).
	klient, fehler := Scryfallklient(NeueEinstellungswahl(storage.NeueEinstellungen(db)).MitRiegel(AutorisierungAus(db)))
	if fehler != nil {
		slog.Default().Error("die Formatprüfung ließ sich nicht einrichten",
			"grund", fehler.Error())

		return
	}

	// **Dieselbe Naht wie an der zweiten Wurzel** (#1622) — es sind zwei
	// Composition Roots (§2), und eine verdrahtete allein wäre wieder der
	// Fall aus #1376: das Fenster richtig, die drei anderen Flächen falsch.
	a.MitDeckpruefung(service.NeueFormatpruefung(
		decks, formate, klient, verweise).MitKartenfeldern(setzeiger))

	// **Und die Namen der Drucke** (#1376) — derselbe Klient, dieselbe
	// Sammelabfrage wie die Formatprüfung. Ohne ihn bleibt die Kennung stehen
	// (§16, Regel 5), und die Namen eigener Karten stehen trotzdem da.
	bau.MitDrucknamen(klient)
}
