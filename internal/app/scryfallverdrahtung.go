// **Die Verdrahtung zum fremden Dienst.**
//
// Eine eigene Datei, weil es eine eigene Frage ist: `scryfall.go` sagt, **was**
// die Fassade anbietet — hier steht, **woraus** es gebaut wird. Fristen,
// Drossel und Speicherort sind Fragen der Verdrahtung, nicht der Fachlichkeit
// (§2, §16).

package app

import (
	"context"
	"database/sql"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Scryfallnaht baut den Abgleich samt Klient und Zwischenspeicher.
//
// # Warum das hier zusammenkommt und nicht im Dienst
//
// `service.Scryfallabgleich` kennt eine **Fähigkeit** (`Fremdkarten`), nicht
// einen Klienten. Wer den Klienten baut, entscheidet über Fristen, Drossel und
// Speicherort — und das sind Fragen der Verdrahtung, nicht der Fachlichkeit
// (§2, §16).
//
// **Der Verweisleser steht im Aufruf und nicht in dieser Datei** — Werkstück 2
// von Phase 3.6. Er hängt am **Bestand**, und wer den Bestand hat, ist der
// Aufrufer: das Fenster hält ihn über den Start, die Flächen ohne Fenster
// bauen ihn je Aufruf über ihre eigene Datenbank.
func Scryfallnaht(
	wahl Dienstwahl,
	schreiber service.Uebernahmeschreiber, verweise service.Verweisleser,
) (*service.Scryfallabgleich, error) {
	klient, err := Scryfallklient(wahl)
	if err != nil {
		return nil, err
	}

	// **Mit dem Rückfall auf die Sammeldatei** (#888): fällt Scryfall aus,
	// kommt die Antwort aus der abgelegten Datei — mit einem Befund, der sagt,
	// dass sie bis zu einen Tag alt ist (§16, Regel 5).
	//
	// **Der Druckweg bekommt ihn nicht**, und das ist keine Auslassung: die
	// Sammeldatei `oracle_cards` führt **je Karte** einen Eintrag, nicht je
	// Druck (`bulk.go`). Ein Rückfall darauf beantwortete „welcher Druck?" mit
	// „irgendeiner" — und genau das soll der Verweis abstellen. Ohne Netz und
	// ohne Zwischenspeicher scheitert der Abgleich über einen Verweis
	// deshalb, statt still den falschen Druck zu nehmen.
	return service.NeuerScryfallabgleich(
		service.MitBulkrueckfall(klient, Bulkrueckfall()), klient, verweise, schreiber), nil
}

// Scryfallanlagenaht baut den Weg „aus einem Namen eine Karte"
// (#1032).
//
// **Über [Scryfallnaht] und nicht daneben**: derselbe Klient, derselbe
// Zwischenspeicher, dieselbe Drossel. Ein zweiter Klient hätte eine zweite
// Drossel, und zwei Drosseln unter einer Grenze sind keine (#1024).
func Scryfallanlagenaht(
	wahl Dienstwahl,
	schreiber service.Uebernahmeschreiber, anleger service.Kartenanleger,
	verweise service.Verweisleser,
) (*service.Scryfallanlage, error) {
	abgleich, err := Scryfallnaht(wahl, schreiber, verweise)
	if err != nil {
		return nil, err
	}

	return service.NeueScryfallanlage(abgleich, anleger), nil
}

// Scryfallsatznaht baut den Weg „aus einem Satzcode ein Set"
// (#1033).
//
// **Über denselben Klienten wie der Abgleich** — ein zweiter hätte eine zweite
// Drossel, und zwei Drosseln unter einer Grenze sind keine (#1024).
func Scryfallsatznaht(
	wahl Dienstwahl,
	speicher service.SetSpeicher, koepfe service.Kopfliste,
) (*service.Satzimport, error) {
	klient, err := Scryfallklient(wahl)
	if err != nil {
		return nil, err
	}

	return service.NeuerSatzimport(klient, speicher, koepfe), nil
}

// Scryfallklient baut den Klienten samt Frist, Drossel und Zwischenspeicher.
//
// # Warum mehrere Klienten unbedenklich sind und mehrere Drosseln nicht
//
// Die Grenze gehört zur **Gegenstelle**, nicht zur Verbindung. `drosselwerk`
// ist deshalb ein geteiltes Werk: jeder hier gebaute Klient nimmt seine
// Drosseln daraus, und zwei Klienten halten zusammen dieselbe Zahl ein
// (#1024).
//
// **Der Bau stand dreimal gleichlautend da** — beim Abgleich, beim Satzimport
// und bei der Suche. Drei Fassungen derselben Verdrahtung sind drei, von denen
// zwei bei der nächsten Änderung stehen bleiben.
func Scryfallklient(wahl Dienstwahl) (*scryfall.Klient, error) {
	ort, err := Scryfallverzeichnis()
	if err != nil {
		return nil, err
	}

	// **Die Sperre reist am Transport mit und wird vor jeder Anfrage
	// gefragt** (Phase 4.8, [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md)).
	//
	// # Warum nicht hier, beim Bauen
	//
	// **Weil ein Schalter sich ändert.** Der erste Anlauf prüfte an dieser
	// Stelle: wer Scryfall danach abschaltete, sah bis zum Neustart keine
	// Wirkung. Dazu kam ein zweiter Fehler derselben Art — beim Start ist die
	// Datenbank an dieser Stelle **noch nicht migriert**, und die Einstellung
	// war gar nicht zu lesen.
	//
	// # Und warum das trotzdem die eine Stelle ist
	//
	// Hier entsteht der **einzige** Scryfall-Klient dieses Programms; acht
	// Stellen bauten ihn vorher wörtlich gleich, und sieben davon wären an
	// der Sperre vorbeigegangen.
	return scryfall.NeuerKlient(scryfalladresse,
		netz.NeuerKlientMit(netz.UeblicheFristen(), proxyquelleAus(wahl)).
			MitErlaubnis(func(ctx context.Context) error {
				return pruefeNachAussen(ctx, wahl, domain.DienstScryfall)
			}),
		netz.NeuerSpeicher(ort, scryfall.Speicherfrist, time.Now),
		time.Now, Fassung).
		MitDrosselwerk(drosselwerk), nil
}

// Verweisnaht baut den Verweisdienst über einem Bestand — Werkstück 2 von
// Phase 3.6.
func Verweisnaht(wahl Dienstwahl, db *sql.DB) (*service.Kartenverweise, error) {
	klient, err := Scryfallklient(wahl)
	if err != nil {
		return nil, err
	}

	// **Mit Setwache** (#1336): »dieses Set hat keine Verweise« und »dieses
	// Set gibt es nicht« sind zwei Sachverhalte.
	return service.NeueKartenverweise(
		klient, storage.NeueKartenverweise(db), time.Now).
		MitSetwache(storage.NeuerSetLeser(db)), nil
}

// dienstwahlUeber baut die Dienstwahl über einem Bestand.
//
// **Der Rückgabetyp ist der konkrete**, weil sie zwei Fragen beantwortet:
// [Dienstwahl] („darf gefragt werden") **und** [Fragewahl] („beantwortet
// diese Frage"). Ein Rückgabewert vom ersten Typ ließe die zweite fallen.
//
// **Sie steht hier und nicht in jeder Naht.** `internal/schichten` führt auf,
// welche Dateien `internal/storage` kennen dürfen (Composition Root, §2) —
// und diese Aufzählung soll nicht mit jeder Naht wachsen, die einen Schalter
// lesen will. Drei wären es beim ersten Anlauf geworden; der Wächter hat es
// angestrichen, und er hatte recht.
func dienstwahlUeber(db *sql.DB) *Einstellungswahl {
	return NeueEinstellungswahl(storage.NeueEinstellungen(db)).MitRiegel(AutorisierungAus(db))
}

// druckvorrat ist der Zwischenspeicher von Scryfall — **über den einen
// Klienten** (§16) und ohne Anfrage: [scryfall.Klient.KarteImVorrat] fragt
// nicht. Ohne Zuhause gibt es keinen Speicher und damit keine Bilder; die
// Namen aus dem eigenen Bestand bleiben.
func (a *App) druckvorrat() service.Druckvorrat {
	klient, fehler := Scryfallklient(dienstwahlUeber(a.laufendeVerbindung()))
	if fehler != nil {
		return keinDruckvorrat{}
	}

	return klient
}

// keinDruckvorrat kennt keinen Druck.
type keinDruckvorrat struct{}

func (keinDruckvorrat) KarteImVorrat(string) (*scryfall.Karte, bool) { return nil, false }
