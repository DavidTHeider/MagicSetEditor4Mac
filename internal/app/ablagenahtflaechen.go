// **Die Zwischenablage für die Flächen ohne Fenster** — Werkstück 1 von
// Phase 3.7
// (#1168).
//
// # Ohne Fenster gibt es keine Zwischenablage — und das ist kein Grund
//
// §16 sagt: dieselbe Fähigkeit, überall dieselbe Antwort — „oder es steht
// begründet, warum nicht (etwa ein Dateiwahldialog, den es ohne Fenster nicht
// gibt)". **Die Zwischenablage ist nicht dieser Fall.** Was das Fenster
// hineinlegt, ist **Text**; ohne Fenster ist derselbe Text die Ausgabe
// beziehungsweise die Eingabe.
//
//	mse ablage-kopieren 1 11 12 > karten.txt
//	mse ablage-einfuegen 2 < karten.txt
//
// Damit trägt der Weg auch dorthin, wo das Fenster gar nicht hinkommt: in eine
// Datei, durch eine Pipe, über eine Maschinengrenze.

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Fremdablage ist das Kopieren und Einfügen für eine Fläche ohne Fenster.
type Fremdablage struct{ quelle Ablagequelle }

// Ablageflaechennaht baut sie über einem Bestand.
//
// **Über [Schreibnaht] und nicht daneben**: sie baut Aufsicht, Feldweg und
// Stilwähler an **einer** Stelle. Zwei Wurzeln liefen auseinander, und genau
// das war der Befund, der `Schreibnaht` hierher gebracht hat.
func Ablageflaechennaht(db *sql.DB, urheber domain.Urheber) (*Fremdablage, error) {
	naht, err := Schreibnaht(db, urheber)
	if err != nil {
		return nil, err
	}

	leser := storage.NeuerSetLeser(db)

	// **Die Naht wird unmittelbar gereicht und nicht zugesichert.**
	//
	// Hier stand eine Typzusicherung mit Wache — nach dem Muster von
	// #591,
	// wo eine ohne Wache eine Fähigkeit stillschweigend verlor. **Sie war hier
	// falsch angewandt**: [Kartenschreibnaht] verspricht `KarteAnlegen`,
	// `FelderVonAussen` und `KarteLoeschen` selbst, also traf die Zusicherung
	// **immer** zu — ihr Fehlerzweig war einer, den kein Lauf betreten kann,
	// und damit eine Behauptung (§3). Gefunden beim Versuch, ihn zu prüfen:
	// es ließ sich keine Naht bauen, die ihn auslöst.
	//
	// **Was die Zusage jetzt hält, ist der Übersetzer.** Nähme jemand
	// `KarteLoeschen` aus [Kartenschreibnaht], schlüge diese Zeile fehl —
	// nicht erst im Betrieb, sondern beim Bauen.
	return &Fremdablage{quelle: service.NeueKartenablage(leser, leser, naht).
		MitLoescher(naht)}, nil
}

// MitAblagequelle setzt die Naht — für Tests.
func (f *Fremdablage) MitAblagequelle(quelle Ablagequelle) *Fremdablage {
	f.quelle = quelle

	return f
}

// Kopieren gibt die genannten Karten als Text.
func (f *Fremdablage) Kopieren(
	ctx context.Context, setID int64, karteIDs []int64,
) (string, error) {
	return f.quelle.Kopieren(ctx, setID, karteIDs)
}

// Ausschneiden gibt sie als Text und entfernt sie aus dem Set.
//
// # Ohne Fenster gibt es kein Klemmbrett, in das man erst legen könnte
//
// Am Fenster muss das Ablegen glücken, bevor eine Karte fällt (Befund 2 des
// Nachlaufs über Phase 3.7) — dort ist es ein eigener, fehlbarer Schritt.
// Hier ist der Weg nach draußen die **Antwort selbst**: der Aufrufer bekommt
// den Text zurück und schreibt ihn, wohin er will.
//
// **Was das kostet, gehört ausgesprochen**: bricht die Ausgabe danach ab —
// eine geschlossene Standardausgabe, eine abgerissene Verbindung —, sind die
// Karten fort und der Text ist niemandem zugegangen. Das ist die Grenze
// dieser Naht und keine Nachlässigkeit: sie kann nicht wissen, ob ihr
// Aufrufer die Antwort je liest.
//
// **Wer sie schließen will**, lässt sich den Inhalt erst über
// `lesen.ablage.kopieren` geben und löscht danach selbst.
func (f *Fremdablage) Ausschneiden(
	ctx context.Context, setID int64, karteIDs []int64,
) (string, error) {
	return f.quelle.Ausschneiden(ctx, setID, karteIDs,
		func(string) error { return nil })
}

// Einfuegen legt die Karten aus einem Text im Set an.
//
// **Dieselbe Gestalt wie am Fenster** ([EinfuegeAntwort]), ohne den Fehler:
// die Flächen ohne Fenster tragen ihn als `error` (§8).
func (f *Fremdablage) Einfuegen(
	ctx context.Context, setID int64, inhalt string,
) (karten, befunde []string, fehler error) {
	ergebnis, fehler := f.quelle.Einfuegen(ctx, setID, inhalt)
	if ergebnis == nil {
		return []string{}, []string{}, fehler
	}

	return kennungstexte(ergebnis.Karten), befundtexte(ergebnis.Befunde), fehler
}
