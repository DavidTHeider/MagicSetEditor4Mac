// **Die Nähte der Sammlung** — Werkstück 2 und 3 von Phase 3.4
// (#1165).
//
// **Eine eigene Datei und trotzdem Composition Root** — derselbe Schnitt wie
// bei `werkzeugnaehte.go` und aus demselben Grund (§2, Zeilengrenze).
//
// # Beides über denselben Bestand
//
// Sammlung und Preise liegen in **einer** Datenbank
// ([0052](../../docs/entscheidungen/0052-die-sammlung-ist-nicht-das-set.md)):
// zwei Dateien hießen zwei Sperren, zwei Sicherungswege und eine Sicherung aus
// Phase 3.1, die beide kennen müsste.

package app

import (
	"database/sql"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// sammlungsnaehte verdrahtet die Sammlung und die Preise.
//
// **Die Sammlung geht immer, die Preise nicht.** Exemplare zu führen braucht
// kein Netz und kein Zuhause außer dem Bestand; Preise zu holen braucht einen
// Zwischenspeicher, und ohne den bleibt die Naht `nil` — die Bindungen sagen
// es dann selbst, statt dass die Seite leer bleibt (§8).
func sammlungsnaehte(a *App, db *sql.DB) {
	a.MitSammlung(service.NeueSammlungspflege(storage.NeuerSammlungsspeicher(db)))

	// **Das Adressbuch im selben Bestand** (ID-04, #1703) — Kontakte gehören
	// dem Profil wie die Sammlung.
	a.MitAdressbuch(service.NeuesAdressbuch(storage.NeueAdressbuchablage(db)))

	// **Derselbe Dienst wie an den drei Flächen ohne Fenster** (Befund 1 des
	// Zweitblicks über Phase 5.2): der Ausschnitt wird einmal gerechnet, und
	// das Fenster liest ihn über dieselbe Naht wie `mse sammlung`.
	a.MitSammlungsfenster(service.NeuesSammlungsfenster(storage.NeuerSammlungsspeicher(db)))

	// **Die Ausfuhr steht vor dem Zuhause** (#1292): sie braucht keines, und
	// eine Anwendung ohne Scryfall-Verzeichnis soll ihre Sammlung trotzdem
	// ausgeben können (§16, Regel 5).
	a.MitSammlungsausfuhr(AusfuhrUeber(db))

	// **Und die vier Fragen** (#1292) — auch sie ohne Netz.
	a.MitSammlungssichten(SichtenUeber(db))

	// **Und die Massenvorgänge** (#1292). Sie brauchen den Bestand; die
	// Preise darin brauchen ein Zuhause, und ohne das sagt der Schub es selbst.
	a.MitMasse(MassennahtUeber(db))

	// **Derselbe Klient wie Auskunft, Abgleich und Bildweg**, und über
	// dieselbe eine Stelle. Er bringt beide Drosseln mit; ein zweiter daneben
	// hätte sie doppelt, und zwei Drosseln unter derselben Grenze halten sie
	// nicht ein.
	//
	// **Eine eigene Konstruktion wäre ein Loch**: an der einen Stelle wird auch
	// der Dienstschalter geprüft (Phase 4.8).
	klient, err := Scryfallklient(NeueEinstellungswahl(storage.NeueEinstellungen(db)).MitRiegel(AutorisierungAus(db)))
	if err != nil {
		return
	}

	a.MitPreisen(service.NeuePreispflege(storage.NeuerPreisspeicher(db),
		storage.NeuerSammlungsspeicher(db), klient, time.Now))

	// **Und die Listeneinfuhr** (#1292). Sie hängt an denselben zwei Quellen
	// — dem Bestand und Scryfall — und damit an demselben Zuhause.
	//
	// **Mit der Sammelsuche** (#1348): eine Zeile mit Setcode geht in einem
	// Schub von bis zu 75 hinaus statt einzeln. Derselbe Klient beantwortet
	// beide Fragen; zwei hätten zwei Drosseln unter derselben Grenze.
	deutung := service.NeueListeneinfuhr(klient, storage.NeuerSammlungsspeicher(db)).
		MitSammelsuche(klient)

	a.MitListeneinfuhr(&Listeneinfuhrnaht{einfuhr: deutung})

	// **Und die Decklisteneinfuhr** (T8, #1473) — dieselbe Deutung, ein
	// anderes Ziel: sie legt Plätze in ein Deck und keinen Besitz in die
	// Sammlung.
	a.MitDeckeinfuhr(&Deckeinfuhrnaht{
		einfuhr: service.NeueDeckeinfuhr(deutung, storage.NeueDeckablage(db)),
	})
}
