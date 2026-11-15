// Diese Datei verdrahtet den MCP-Server und hält seine Schreibnaht.
//
// **Eine eigene Datei**, weil sie eine eigene Frage beantwortet: was braucht
// der zweite Prozess zum Schreiben, und über welche fachliche Naht bekommt er
// es? `app.go` daneben beantwortet die des Fensters.
//
// Getrennt am 2026-08-22, weil `app.go` über die 400 Zeilen aus §2 lief.

package app

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Schreibnaht baut die fachliche Schreibnaht für eine Fläche außerhalb des
// Fensters.
//
// **Sie hieß bis zum 2026-08-23 `MCPSchreiber`** und war es nie: sie setzt
// zusammen, was **jede** Fläche braucht, die Karten schreibt. Mit der
// Kommandozeile (#886) gibt es eine zweite, und ein Name, der die erste nennt,
// führte den nächsten Aufrufer in die Irre — oder zu einer zweiten Fassung
// derselben Zusammensetzung. Genau davor warnt der Absatz darunter.
//
// **Dieselbe fachliche Naht wie im Fenster** —
// #710.
//
// Der zweite Composition Root verdrahtete `storage.NeuerKartenschreiber`
// unmittelbar; damit entstand über MCP ein roher `text`-Wert, den das Paket nie
// liest, während dasselbe Feld über das Fenster seit #696 in `rule text` und
// `flavor text` geht. **Derselbe Aufruf, zwei Ergebnisse.** Angestrichen vom
// Zweitblick durch Codex am 2026-08-21.
//
// **Sie steht hier und nicht in `cmd/mse-mcp`**, weil der Schichtenwächter dort
// weder `service` noch `msepaket` zulässt — und das aus gutem Grund: was
// zusammengesetzt wird, gehört an **eine** Stelle, sonst laufen die beiden
// Wurzeln auseinander. Genau das war der Befund.
func Schreibnaht(db *sql.DB, urheber domain.Urheber) (Kartenschreibnaht, error) {
	pakete, paketfehler := Paketverzeichnis()
	bilder, bildfehler := Bilderverzeichnis()

	if err := errors.Join(paketfehler, bildfehler); err != nil {
		return nil, err
	}

	return schreibnahtMit(db, urheber, stilwaehlerMit(db, pakete), nachzugswerk(db, pakete, bilder)), nil
}

// nachzugswerk baut das Zeichenwerk, über das der Feldweg rechnet, was er
// speichert (#2396) — **dasselbe wie im Fenster**: Kommandozeile, MCP und
// Dienst setzen ein Feld, wie das Fenster es setzt (§16).
func nachzugswerk(db *sql.DB, pakete, bilder string) *Zeichenwerk {
	return NeuesZeichenwerk(storage.NeuerSetLeser(db), paketwurzelnFuer(pakete), bilder)
}

// schreibnahtMit baut sie über einem fertigen Stilwähler.
//
// **Damit ein Aufrufer, der den Wähler ohnehin braucht, das
// Paketverzeichnis nur einmal auflöst** (A7, #1446).
func schreibnahtMit(
	db *sql.DB, urheber domain.Urheber, stile Stilwaehler, werk *Zeichenwerk,
) Kartenschreibnaht {
	leser := storage.NeuerSetLeser(db)
	bestand := storage.NeuerKartenschreiber(db).Als(urheber)

	naht := feldnaht{
		Kartenschreiber: bestand,
		weg: service.NeuerFeldweg(leser, stile, bestand).
			MitNachzug(werk.Stile, werk.Schriften, werk.Bildquelle),
	}

	return mcpnaht{
		feldnaht: naht,
		aufsicht: service.NeueKartenaufsicht(leser, bestand).
			MitFeldweg(naht, bestand).
			// **Auch hier** (#1369): ein Sprachmodell löscht eine Karte
			// leichter als ein Mensch, und der Deckplatz wäre still fort.
			MitDeckwache(storage.NeueDeckablage(db)).
			// **Und die Vorlage** (#1390): ohne sie reichte die
			// Massenänderung ein nicht editierbares oder berechnetes Feld
			// ungeprüft bis in den Speicher durch. Das Fenster hat diese
			// Zeile seit #1211; hier hat sie gefehlt, solange es die
			// Fähigkeit kopflos nicht gab.
			MitFeldformen(stile),
	}
}

// Kartenschreibnaht ist, was der MCP-Server zum Schreiben braucht.
//
// **Hier deklariert, damit `cmd/mse-mcp` sie nennen kann**, ohne `service` zu
// kennen. Sie ist gleichlautend mit `mcp.Schreiber`; die Zuordnung besorgt Gos
// strukturelle Typisierung.
type Kartenschreibnaht interface {
	FeldSetzenMitVorwert(ctx context.Context, setID, karteID int64, feld, wert string) (
		vorher string, hatte bool, fehler error)

	// FeldSetzenVonAussen setzt ein Feld und hält fest, **woher der Wert kam**
	// (#888).
	//
	// **Sie steht neben der Methode darüber und ersetzt sie nicht**: eine
	// Tipparbeit hat keine Herkunft, und ein Aufruf, der sie mit einer leeren
	// Zeichenkette mitschleifte, wäre eine Einladung, versehentlich eine
	// falsche zu setzen.
	FeldSetzenVonAussen(
		ctx context.Context, setID, karteID int64, feld, wert, herkunft string,
	) (vorher string, hatte bool, fehler error)

	// FelderVonAussen schreibt mehrere übernommene Werte in **einem** Zug.
	//
	// **Sie steht neben der Methode darüber, weil sie etwas anderes zusagt**
	// (#1026):
	// eine Schleife über die einzelne kann halb durch sein, dieser Weg nicht.
	FelderVonAussen(
		ctx context.Context, setID, karteID int64, werte []domain.Value, herkunft string,
	) (vorher string, hatte bool, fehler error)

	// FelderVonAussenMitHerkunft schreibt zusätzlich, **woher jeder Wert
	// kommt** — Werkstück A11 (#1450).
	// **`erwartet` ist der Stand, auf dem die Karte stehen muss** — leer
	// heißt: ohne Bedingung (Befund 2 des Zweitblicks über Phase 5.3).
	FelderVonAussenMitHerkunft(
		ctx context.Context, setID, karteID int64, werte []domain.Value,
		herkunft string, herkuenfte []domain.KIFeldherkunft,
		angaben map[string][]domain.KIBildparameter, erwartet string,
	) (vorher string, hatte bool, fehler error)

	// GruppeVonAussenMitHerkunft schreibt **mehrere Karten in einem Zug** —
	// Werkstück A13.1b (#1452). Jede Karte wird beaufsichtigt, **bevor** die
	// erste geschrieben wird.
	GruppeVonAussenMitHerkunft(
		ctx context.Context, schritte []domain.Kartenschritt) error

	FeldEntfernen(ctx context.Context, setID, karteID int64, feld string) (
		vorher string, hatte bool, stelle *int64, fehler error)
	FeldSetzenAnStelle(ctx context.Context, setID, karteID int64, feld, wert string,
		stelle int64) (vorher string, hatte bool, fehler error)

	// **Und der Lebenszyklus der Karte** — seit
	// #591.
	//
	// Er stand schon in `storage.Kartenschreiber`, den `mcpnaht` einbettet;
	// gefehlt hat nur, dass jemand ihn verlangt. Ein Sprachmodell konnte bis
	// dahin ein Feld setzen und sonst nichts ändern.
	//
	// **Mit dem Set** — seit
	// #801:
	// eine verwechselte Kartenkennung löschte sonst eine Karte in einem Set,
	// das niemand geöffnet hatte.
	KarteAnlegen(ctx context.Context, setID int64) (int64, error)
	KarteKopieren(ctx context.Context, setID, karteID int64) (int64, error)
	KarteLoeschen(ctx context.Context, setID, karteID int64, trotzdem bool) error

	// KartenLoeschen entfernt mehrere in **einem** Zug — dieselbe Zusage wie
	// `FelderVonAussen` darüber, und aus demselben Grund: eine Schleife über
	// die einzelne kann halb durch sein, dieser Weg nicht (Befund 4 des
	// Zweitblicks über Phase 3.7).
	//
	// **Das letzte Argument bestätigt ein Löschen, das einen Deckplatz
	// mitnimmt** (#1369).
	KartenLoeschen(
		ctx context.Context, setID int64, karteIDs []int64, trotzdem bool) error
	KartenTauschen(ctx context.Context, setID, eineID, andereID int64) error

	// FelderSetzen setzt **ein** Feld auf **mehreren** Karten — in einer
	// Transaktion (#1390).
	//
	// # Warum das eine eigene Fähigkeit ist und keine Schleife
	//
	// Weil die Atomarität die Fähigkeit **ist**. Wer ohne Fenster ein Feld
	// über 200 Karten ändert, macht sonst 200 Aufrufe; ein Fehler in der
	// Mitte lässt das Set halb geändert zurück, und der Aufrufer weiß nicht,
	// welche Hälfte.
	//
	// **Der Katalog behauptete das Gegenteil**, und zwar mit einer
	// Begründung, die eine andere Operation beschrieb: „das Fenster schreibt
	// die Eingaben **einer** Karte in einem Zug, weil eine Maske viele Felder
	// hat". Gefunden bei der Prüfung der 24 Flächenlücken (Werkstück E von
	// #1324).
	//
	// **Sie gibt den Vorgang zurück** (#1393). Mit ihm nimmt der Aufrufer
	// genau diese Änderung zurück; „der jüngste Schritt des Sets" trifft
	// ohne Fenster womöglich den eines zweiten Prozesses.
	FelderSetzen(ctx context.Context, setID int64, karteIDs []int64, feld, wert string) (
		vorgang string, fehler error)

	// FelderVorschau sagt, was der Wert auf den Karten hieße — ohne zu
	// schreiben (Werkstück U6, #1417).
	FelderVorschau(ctx context.Context, setID int64, karteIDs []int64, feld, wert string) (
		domain.Massenvorschau, error)

	// FelderSetzenMitErwartung schreibt nur, wenn jede Karte noch auf dem
	// Altwert steht, den die Vorschau gezeigt hat.
	FelderSetzenMitErwartung(ctx context.Context, setID int64, erwartet []domain.Feldvorwert, feld, wert string) (
		vorgang string, fehler error)
}

// mcpnaht schickt **das Setzen eines Feldes** über den Feldweg, **den
// Lebenszyklus** über die Kartenaufsicht und alles Übrige unmittelbar an den
// Bestand.
//
// **Eingebettet und nicht abgeschrieben**: `FeldEntfernen` und
// `FeldSetzenAnStelle` sind Rücknahmewege und gehen an dieselbe Stelle wie
// bisher; nur das Setzen hat eine fachliche Frage davor.
//
// **Der Lebenszyklus ging bis zum 2026-08-22 ebenfalls unmittelbar dorthin** —
// und damit an `service` vorbei. Eine Setgrenze, die dort geprüft werden soll,
// konnte es nicht, weil der Weg nicht dort entlangführte
// ([0041](../../docs/entscheidungen/0041-die-setgrenze-liegt-im-dienst.md)).
// Die Aufsicht steht deshalb **vor** dem eingebetteten Schreiber: ihre vier
// Methoden tragen eine Setkennung mehr und verdecken die eingebetteten nicht,
// sondern ersetzen sie für jeden, der die Naht über `Kartenschreibnaht` nimmt.
type mcpnaht struct {
	feldnaht

	aufsicht *service.Kartenaufsicht
}

// bestandsnaht bindet `storage.Sichern` an die Schnittstelle aus
// `bestandsicherung.go`.
//
// **Hier und nicht dort**: `storage` zu importieren ist den beiden Composition
// Roots vorbehalten (§2), und `bestandsicherung.go` ist keiner.
type bestandsnaht struct{ db *sql.DB }

func (b bestandsnaht) Sichern(ctx context.Context, ziel string) error {
	return sichernUeber(b.db)(ctx, ziel)
}

// Einfuhrnaht baut den Weg, auf dem ein Setpaket in den Bestand kommt.
//
// **Aus demselben Grund exportiert wie [Schreibnaht]**: was zusammengesetzt
// wird, gehört an **eine** Stelle. Die Kommandozeile (#886) darf `service`
// nicht kennen — der Schichtenwächter lässt es nicht zu, und das aus gutem
// Grund: sonst hätte der dritte Composition Root seine eigene Fassung davon,
// und drei Wurzeln liefen auseinander.
func Einfuhrnaht(db *sql.DB) *service.Importeur {
	// **Mit der Kopfquelle**: „eine Quelle, ein Set" ist eine fachliche Regel
	// und gilt für jede Fläche.
	return service.NeuerImporteur(storage.NeuerSetSpeicher(db)).
		MitKoepfen(storage.NeuerSetLeser(db))
}

// Setschreibnaht baut den Weg, auf dem ein Set wieder in eine Datei kommt.
//
// **Mit Bildablage und Stilauflösung**, wie im Fenster: sonst fehlten die
// eingesetzten Bilder, oder der Stil stünde in einer Form da, die MSE2 nicht
// findet (#710; #2247, Befund 2) — eine andere Datei als die des Fensters.
func Setschreibnaht(db *sql.DB) (*service.Setschreiber, error) {
	pakete, paketfehler := Paketverzeichnis()
	ort, bildfehler := Bilderverzeichnis()

	if fehler := errors.Join(paketfehler, bildfehler); fehler != nil {
		return nil, fehler
	}

	leser := storage.NeuerSetLeser(db)

	return service.NeuerSetschreiber(leser, leser).
		MitBildablage(service.NeueBildablage(ort)).
		MitStilaufloesung(service.StilaufloesungUeber(paketwurzelnFuer(pakete)...)).
		MitPruefer(AutorisierungAus(db)), nil
}

// Zeichennaht gibt das Zeichenwerk für eine Fläche ohne Fenster.
//
// **Sie bestimmt die Wurzeln selbst**, weil eine Fläche ohne Fenster keinen
// verwalteten Ort kennt: es gibt kein „zuletzt gewähltes Verzeichnis".
//
// **Und sie behält es je Bestand** (#2005), solange Orte und Paketwurzeln
// dieselben sind — wann es neu entsteht und wer den Rest sieht, steht in
// `zeichenvorrat.go`. Die Orte fragt sie trotzdem je Ruf: ohne Zuhause gibt
// es auch kein behaltenes Werk.
func Zeichennaht(db *sql.DB) (*Zeichenwerk, error) {
	// **Beide Orte auf einmal, ein Zweig.** Sie kommen aus derselben Quelle —
	// dem Benutzerverzeichnis —, und ein zweites `if` darauf wäre ein Zweig,
	// den kein Aufruf betreten kann: scheitert der eine, scheitert der andere
	// mit (§3). Beide Fehler zu melden ist obendrein die bessere Auskunft.
	pakete, paketfehler := Paketverzeichnis()

	bilder, bildfehler := Bilderverzeichnis()

	if fehler := errors.Join(paketfehler, bildfehler); fehler != nil {
		return nil, fehler
	}

	wurzeln := paketwurzelnFuer(pakete)

	return zeichenvorrat.werk(db, werkstand(wurzeln, bilder, Schriftwurzel()), func() *Zeichenwerk {
		return NeuesZeichenwerk(storage.NeuerSetLeser(db), wurzeln, bilder).
			MitPruefer(AutorisierungAus(db))
	}), nil
}

// alleinnaht führt `storage.AlleinAmBestand` an die Naht dieser Schicht.
func alleinnaht(ctx context.Context, pfad string) (func() error, error) {
	// **Zwei Fragen und eine Naht** (#1163, Befund Z6-02).
	//
	// `AlleinAmBestand` fragt **SQLite** und sieht damit auch ein `sqlite3`
	// im Terminal — aber es lässt wieder los. `Alleinnehmen` fragt die
	// **Teilnehmer** und hält über den Tausch. Keine von beiden ersetzt die
	// andere, und getrennte Nähte hießen zwei Aufrufe, von denen ein
	// Aufrufer einen vergessen kann.
	if fehler := storage.AlleinAmBestand(ctx, pfad); fehler != nil {
		return nil, alsDienstHalter(ctx, pfad, fehler)
	}

	return storage.Alleinanspruch(pfad)
}

// sicherungTaugt sagt, ob eine vorhandene Datei als Sicherung durchgeht.
//
// **Dieselbe Prüfung wie beim Zurückspielen** — eine zweite wäre eine zweite
// Fassung derselben Frage.
func sicherungTaugt(ctx context.Context, pfad string) error {
	_, err := storage.PruefeSicherung(ctx, pfad)

	return err
}

// sicherungsnaht führt `storage.PruefeSicherung` an die Naht dieser Schicht.
//
// **Eine Umformung und keine Entscheidung** (§2): dieselben Angaben, in dem
// Typ, den `internal/app` selbst benennt.
func sicherungsnaht(ctx context.Context, pfad string) (Sicherungsbefund, error) {
	befund, err := storage.PruefeSicherung(ctx, pfad)
	if err != nil {
		return Sicherungsbefund{}, err
	}

	return Sicherungsbefund{
		AngelegtAm:  befund.AngelegtAm,
		Schemastand: befund.Schemastand,
		Sets:        befund.Sets,
		Karten:      befund.Karten,
	}, nil
}

// feldnaht schickt **das Setzen** über den Feldweg und die beiden
// Rücknahmewege unmittelbar an den Bestand.
//
// **Eine eigene Naht, weil beide Composition Roots sie brauchen** (#808): das
// Fenster reicht sie an die Kartenaufsicht, der MCP-Server ebenso. Sie stand
// bis dahin in `mcpnaht` und war damit für das Fenster nicht zu haben.

type feldnaht struct {
	*storage.Kartenschreiber

	weg *service.Feldweg
}

func (f feldnaht) FeldSetzenMitVorwert(
	ctx context.Context, karteID int64, feld, wert string,
) (vorher string, hatte bool, fehler error) {
	return f.weg.FeldSetzenMitVorwert(ctx, karteID, feld, wert)
}

// FeldSetzenVonAussen geht denselben Weg und trägt die Herkunft mit.
func (f feldnaht) FeldSetzenVonAussen(
	ctx context.Context, karteID int64, feld, wert, herkunft string,
) (vorher string, hatte bool, fehler error) {
	return f.weg.FeldSetzenVonAussen(ctx, karteID, feld, wert, herkunft)
}

// FelderVonAussen schreibt mehrere übernommene Werte in einem Zug (#1026).
func (f feldnaht) FelderVonAussen(
	ctx context.Context, karteID int64, werte []domain.Value, herkunft string,
) (vorher string, hatte bool, fehler error) {
	return f.weg.FelderVonAussen(ctx, karteID, werte, herkunft)
}

// FelderVonAussenMitHerkunft schreibt Werte samt ihrer Herkunft in einem Zug
// (A11, #1450).
func (f feldnaht) FelderVonAussenMitHerkunft(
	ctx context.Context, karteID int64, werte []domain.Value,
	herkunft string, herkuenfte []domain.KIFeldherkunft,
	angaben map[string][]domain.KIBildparameter, erwartet string,
) (vorher string, hatte bool, fehler error) {
	return f.weg.FelderVonAussenMitHerkunft(
		ctx, karteID, werte, herkunft, herkuenfte, angaben, erwartet)
}

// GruppeSetzenMitHerkunft geht denselben Weg — jede Karte gerechnet (#2396).
func (f feldnaht) GruppeSetzenMitHerkunft(ctx context.Context, schritte []domain.Kartenschritt) error {
	return f.weg.GruppeSetzenMitHerkunft(ctx, schritte)
}
