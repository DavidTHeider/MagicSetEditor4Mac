// **Kennzahlen, Rücknahme und Einstellungen für die Flächen ohne Fenster**
// (#1107).
//
// # Drei in einer Datei
//
// Sie teilen dieselbe Eigenschaft: sie brauchen **nur** den Bestand — keinen
// Stilwähler, kein Paketverzeichnis, kein Zuhause. Damit sind sie die
// billigsten Nähte des Projekts, und drei eigene Dateien für je zwölf Zeilen
// wären eine Ordnung, die niemandem beim Lesen hilft (§2: der Schnitt folgt
// der Frage, nicht der Zeilenzahl).

package app

import (
	"context"
	"database/sql"
	"path/filepath"
	"strconv"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Zahlenweg gibt die Kennzahlen eines Sets.
type Zahlenweg struct{ zahlen Statistiker }

// ZahlennahtUeber baut den Weg zu den Kennzahlen über einer Datenbank.
func ZahlennahtUeber(db *sql.DB) *Zahlenweg {
	return &Zahlenweg{zahlen: storage.NeuerSetLeser(db)}
}

// MitZaehler setzt die Naht zu den Kennzahlen — für Tests.
func (w *Zahlenweg) MitZaehler(zahlen Statistiker) *Zahlenweg {
	w.zahlen = zahlen

	return w
}

// Statistik gibt Kartenzahl, Feldzahl, leere Karten und die Belegung je Feld.
func (w *Zahlenweg) Statistik(ctx context.Context, setID int64) (StatistikDTO, error) {
	zahlen, err := w.zahlen.Kennzahlen(ctx, setID)
	if err != nil {
		return StatistikDTO{}, err
	}

	belegung, err := w.zahlen.Belegung(ctx, setID)
	if err != nil {
		return StatistikDTO{}, err
	}

	statistik := StatistikDTO{
		Karten:   zahlen.Cards,
		Felder:   zahlen.Fields,
		Leere:    zahlen.Empty,
		Belegung: make([]ZahlDTO, 0, len(belegung)),
	}

	for _, feld := range belegung {
		statistik.Belegung = append(statistik.Belegung,
			ZahlDTO{Name: feld.Field, Karten: feld.Cards})
	}

	return statistik, nil
}

// Verteilung sagt, wie viele Karten welchen Wert eines Felds tragen.
func (w *Zahlenweg) Verteilung(
	ctx context.Context, setID int64, feld string,
) ([]WertzahlDTO, error) {
	werte, err := w.zahlen.Verteilung(ctx, setID, feld)
	if err != nil {
		return nil, err
	}

	verteilung := make([]WertzahlDTO, 0, len(werte))
	for _, wert := range werte {
		verteilung = append(verteilung, verteilungszeile(wert))
	}

	return verteilung, nil
}

// Verlaufsweg nimmt Änderungen zurück und stellt sie wieder her.
type Verlaufsweg struct{ verlauf Verlaufsfuehrer }

// VerlaufsnahtUeber baut den Weg zum Verlauf über einer Datenbank.
func VerlaufsnahtUeber(db *sql.DB) *Verlaufsweg {
	// **Mit Setwache** (#1336): »Nichts zu bewegen« ist für ein Set ohne
	// Verlauf richtig — für eines, das es nicht gibt, nicht.
	return &Verlaufsweg{verlauf: service.NeuerVerlaufsdienst(
		storage.NeuerVerlaufsleser(db)).MitSetwache(storage.NeuerSetLeser(db))}
}

// Zuruecknehmen macht eine Änderung rückgängig.
//
// **Ohne `vorgang` die jüngste noch geltende des Sets, mit ihm genau diese**
// (#1393).
// Die Kennung kommt aus der Antwort der Massenänderung
// ([MassenAntwort.Vorgang]).
func (w *Verlaufsweg) Zuruecknehmen(
	ctx context.Context, setID int64, vorgang string,
) (BewegungAntwort, error) {
	return bewegt(w.verlauf.Zuruecknehmen(ctx, setID, vorgang))
}

// Wiederherstellen führt eine zurückgenommene Änderung wieder aus.
func (w *Verlaufsweg) Wiederherstellen(
	ctx context.Context, setID int64, vorgang string,
) (BewegungAntwort, error) {
	return bewegt(w.verlauf.Wiederherstellen(ctx, setID, vorgang))
}

// bewegt übersetzt beide Richtungen — sie unterscheiden sich nur im Aufruf.
func bewegt(bewegung service.Bewegung, err error) (BewegungAntwort, error) {
	if err != nil {
		return BewegungAntwort{}, err
	}

	return BewegungAntwort{
		Bewegt:  bewegung.Bewegt,
		KarteID: strconv.FormatInt(bewegung.KarteID, 10),
		Feld:    bewegung.Feld,
		Umfang:  bewegung.Umfang,
		Urheber: string(bewegung.Urheber),
	}, nil
}

// Einstellungsweg liest und schreibt die Einstellungen.
type Einstellungsweg struct {
	einstellungen Einstellungsfuehrer

	// austausch führt sie aus und ein — Werkstück O3 von Phase 5.2
	// (#1427).
	austausch Einstellungsaustauscher
}

// EinstellungsnahtUeber baut den Weg zu den Einstellungen über einer Datenbank.
func EinstellungsnahtUeber(db *sql.DB) *Einstellungsweg {
	return &Einstellungsweg{
		// **Mit der Naht zum Menschen** — Werkstück 9 (#1351), Punkt 9. Ohne
		// sie bekäme die Kommandozeile eine andere Antwort als das Fenster,
		// und §16 verlangt dieselbe.
		einstellungen: einstellungsverwalterUeber(db),

		// **Mit dem Riegel** — Werkstück O3 von Phase 5.2 (#1427): die
		// Übernahme ist `lokalSchreiben`, und die drei Flächen ohne Fenster
		// bekommen denselben Prüfer wie das Fenster (0078).
		austausch: service.NeuerEinstellungsaustausch(
			storage.NeueEinstellungen(db), storage.NeueNutzerablage(db),
			storage.NeueEinstellungen(db), AutorisierungAus(db)),
	}
}

// Alle gibt jede Einstellung samt ihren erlaubten Werten.
func (w *Einstellungsweg) Alle(ctx context.Context) ([]EinstellungDTO, error) {
	// **Über `Stand` und nicht über `Alle`** — Werkstück 9 (#1351), Punkt 9:
	// die Antwort trägt die **Wahl** und das, was **gilt**. Wer nur das
	// zweite hätte, bekäme auf `mse einstellungen` einen Wert, den er nie
	// gewählt hat — und beim nächsten Setzen schriebe er ihn fest.
	stand, err := w.einstellungen.Stand(ctx)
	if err != nil {
		return nil, err
	}

	// **Die Reihe kommt aus der Domäne**, nicht von hier: sonst entschiede
	// diese Naht, welche Einstellungen es gibt, und die Antwort stünde an zwei
	// Stellen — mit dem Fenster als dritter.
	reihe := domain.Einstellungsreihe()

	alle := make([]EinstellungDTO, 0, len(reihe))
	for _, schluessel := range reihe {
		alle = append(alle, EinstellungDTO{
			Schluessel: schluessel,
			// **Ohne Proxykennwort** — S13.NAT-02 (#1706).
			Wert:     domain.Einstellungsanzeige(schluessel, stand[schluessel].Gewaehlt),
			Wirksam:  domain.Einstellungsanzeige(schluessel, stand[schluessel].Wirksam),
			Werte:    domain.ErlaubteEinstellungswerte(schluessel),
			Bereich:  domain.EinstellungsBereich(schluessel),
			Ebene:    ebeneVon(schluessel),
			Herkunft: stand[schluessel].Herkunft,
			Wirkung:  domain.EinstellungsWirkung(schluessel),
			Vorgabe:  domain.EinstellungsVorgabe(schluessel),
		})
	}

	return alle, nil
}

// Setzen schreibt eine Einstellung fort.
func (w *Einstellungsweg) Setzen(ctx context.Context, schluessel, wert string) error {
	return w.einstellungen.Setzen(ctx, schluessel, wert)
}

// DienstablageNeben baut die Ablage des Dienstes neben einem Bestand — dort,
// wo auch seine Sicherungen liegen, und damit innerhalb des Pfadvertrags —,
// mit der Frage an diesen Bestand, ob ein Set auf eine Quelle verweist
// (#1944).
func DienstablageNeben(db *sql.DB, bestand string) *Dienstablage {
	return NeueDienstablage(filepath.Join(filepath.Dir(bestand), "dienst")).
		MitVerweis(storage.NeuerSetLeser(db).HerkunftVerwiesen)
}

// NeueSprachpaketeUeber baut die eigenen Übersetzungen über einem Bestand —
// für das Fenster und für die drei Flächen ohne Fenster derselbe Dienst
// (LOC-02, #1688). **Hier und nicht in `sprachvorlage.go`**: nur die
// benannten Wurzeln kennen `storage` (Schichtenregel).
func NeueSprachpaketeUeber(db *sql.DB) *service.Sprachpakete {
	return service.NeueSprachpakete(storage.NeueSprachpaketablage(db), Sprachvorlage(),
		"MagicSetEditor4Mac "+Fassung, time.Now)
}
