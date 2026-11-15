// **Der Zugang zu Scryfall beim Start** — herausgelöst, weil `start.go` mit
// ihm über 400 Zeilen ginge (§2).
//
// Der Schnitt folgt einer Frage: dort steht, **was** beim Start entsteht,
// hier, wie der fremde Dienst verdrahtet wird — Klient, Frist, Drossel und
// Zwischenspeicher (§16).

package app

import (
	"database/sql"
	"log/slog"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// scryfallStarten verdrahtet Abgleich, Bildweg, Auskünfte und Suche.
//
// **Die Aufsicht und der Leser kommen von außen**, weil sie beim Start ohnehin
// entstehen: sie hier ein zweites Mal zu bauen gäbe zwei Urheber im Verlauf
// für dieselbe Handlung.
func (a *App) scryfallStarten(
	db *sql.DB, aufsicht *service.Kartenaufsicht, leser Zeichenleser,
) {
	// **Der Zugang zu Scryfall wird hier gebaut** (#1005): Klient, Frist,
	// Drossel und Zwischenspeicher sind Fragen der Verdrahtung, nicht der
	// Fachlichkeit (§2, §16).
	//
	// **Und er schreibt über dieselbe Aufsicht wie das Fenster.** Ein
	// übernommener Regeltext landet damit in derselben Zeile wie ein
	// getippter — Vorschau, Druckbogen, Ausfuhr und MCP sehen ihn ohne eine
	// weitere Zeile Code.
	//
	// **Ein Fehler hier hält den Start nicht auf.** Wer kein Zuhause hat,
	// bekommt keine Scryfall-Fähigkeit; das Fenster geht trotzdem auf, und
	// die Meldung kommt beim ersten Versuch — dieselbe Haltung wie beim
	// Paketverzeichnis darüber.
	// **Der Verweisleser hängt am Bestand** — Werkstück 2 von Phase 3.6.
	// Derselbe, den [App.KarteVerweisSetzen] benutzt: zwei Ablagen über
	// derselben Tabelle wären zwei Wahrheiten über denselben Verweis.
	verweise := storage.NeueKartenverweise(db)

	// **Eine Wahl für alle Nähte dieses Starts.** Zwei bauten dieselbe Frage
	// zweimal, und eine davon läse irgendwann einen anderen Schalter.
	//
	// **Ohne Kontext seit dem 2026-09-02**: keine Naht liest hier noch etwas
	// (Befund 1 des Zweitblicks über Phase 4.8) — gelesen wird je Anfrage.
	wahl := NeueEinstellungswahl(storage.NeueEinstellungen(db)).MitRiegel(AutorisierungAus(db))

	// **Und sie bleibt an der Anwendung**, damit die Wege, die erst später
	// einen Klienten bauen, denselben Schalter lesen.
	a.MitDienstwahl(wahl)

	if dienst, fehler := Verweisnaht(wahl, db); fehler != nil {
		slog.Default().Error("die Verweise auf echte Drucke ließen sich nicht einrichten",
			"grund", fehler.Error())
	} else {
		a.verweise = dienst
	}

	if abgleich, fehler := Scryfallnaht(wahl, aufsicht, verweise); fehler != nil {
		slog.Default().Error("der Zugang zu Scryfall ließ sich nicht einrichten",
			"grund", fehler.Error())
	} else {
		a.scryfall = abgleich
	}

	// **Und derselbe Zugang legt Karten an** (#1032). Über dieselbe Naht wie
	// der Abgleich — ein zweiter Klient hätte eine zweite Drossel, und zwei
	// Drosseln unter einer Grenze sind keine (#1024).
	if anlage, fehler := Scryfallanlagenaht(wahl, aufsicht, aufsicht, verweise); fehler != nil {
		slog.Default().Error("das Anlegen aus Scryfall ließ sich nicht einrichten",
			"grund", fehler.Error())
	} else {
		a.scryfallanlage = anlage
	}

	a.fassungen = FassungsnahtMit(a.Proxyquelle())

	// **Und der Prüfer, der sagt, ob sie diesen Bestand öffnen kann** —
	// Werkstück O6 von Phase 5.2 (#1430).
	a.MitVorpruefer(VertraeglichkeitsprueferUeber(a, a.datenbankpfad, a.bestandsweg()))

	// **Und ein ganzer Satz** (#1033). Über denselben Speicher wie der
	// MTG-JSON-Import — „eine Quelle, ein Set" gilt für beide, und zwei
	// Speicher kennten einander nicht.
	if saetze, fehler := Scryfallsatznaht(wahl,
		storage.NeuerSetSpeicher(db), leser); fehler != nil {
		slog.Default().Error("das Einlesen von Scryfall-Sätzen ließ sich nicht einrichten",
			"grund", fehler.Error())
	} else {
		a.scryfallsaetze = saetze
	}

	// **Und derselbe Zugang liefert Bilder** (#1005). Sie gehen durch dieselbe
	// Naht wie eine gezogene Datei — `BildEinsetzen` legt ab, und damit sehen
	// Vorschau, Druckbogen und `.mse-set`-Ausfuhr sie ohne eine weitere Zeile.
	if bildweg, fehler := ScryfallBildnaht(wahl); fehler != nil {
		slog.Default().Error("der Bildweg zu Scryfall ließ sich nicht einrichten",
			"grund", fehler.Error())
	} else {
		a.scryfallbilder = bildweg
	}

	// **Und die lesenden Auskünfte** (#888): Regelhinweise, Setangaben und die
	// Symbologie. Sie schreiben nichts und hängen deshalb an keiner Aufsicht.
	if auskunft, fehler := Auskunftsnaht(wahl); fehler != nil {
		slog.Default().Error("die Auskünfte von Scryfall ließen sich nicht einrichten",
			"grund", fehler.Error())
	} else {
		a.auskunft = auskunft
	}

	// **Und die Kartensuche** (#1167). Sie schreibt ebenso wenig — und sie ist
	// der Weg, über den jemand eine Karte überhaupt findet, deren Namen er
	// nicht auswendig kennt.
	// **Kein Fehlerweg mehr.** Die Naht entscheidet nichts; was fehlt, sagt
	// die Suche je Anfrage (Befund 1 des Zweitblicks über Phase 4.8).
	a.MitScryfallsuche(Suchnaht(wahl, schluesselbund.Anmeldebund(), wahl))

	// **Und derselbe Schreiber legt neue Karten an** (#562). Ein zweiter wäre
}
