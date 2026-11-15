// Die Naht zum **Zufallspack** für die Flächen ohne Fenster
// (#889,
// Befund 7 des Zweitblicks über Phase W).
//
// # Warum es sie gibt
//
// Bis hierher bestand das Zufallspack **nur als Wails-Bindung**, ohne dass
// irgendwo gestanden hätte, warum. §16 lässt genau zwei Antworten zu: die
// Fähigkeit gibt es an allen vier Flächen — oder es steht begründet, warum
// nicht. Ein Booster ist Text: eine Liste von Kartennamen, herstellbar aus
// einem Samen. Ein Dateiwahldialog, den es ohne Fenster nicht gibt, ist er
// nicht.
//
// **Und der Samen ist der Grund, warum es sich lohnt.** Dieselbe Zahl gibt
// dieselbe Ziehung; ein Sprachmodell oder ein Skript kann damit ein Ergebnis
// herstellen, über das jemand im Fenster spricht.

package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Packweg ist das Zufallspack für die Flächen ohne Fenster.
//
// **Derselbe Dienst wie hinter der Bindung**, nur ohne die DTO-Übersetzung:
// eine Fläche ohne Fenster braucht einen `error`, keinen Fehlerverbund. Was
// beide teilen, ist [service.Packbereich] — und damit die Antwort.
type Packweg struct{ bereich Packquelle }

// PackUeber baut das Zufallspack über einem Bestand.
//
// **Mit dem Zeichenwerk**, denn der Filter jeder Packart ist ein Ausdruck über
// `card` und braucht dieselbe Umgebung, in der auch gezeichnet wird. Ohne sie
// sähe eine Fläche ohne Fenster andere Packarten als das Fenster (§16).
func PackUeber(db *sql.DB) (*Packweg, error) {
	werk, err := Zeichennaht(db)
	if err != nil {
		return nil, err
	}

	leser := storage.NeuerSetLeser(db)

	// **Verteilt wie die Problemprüfung** (#2466) — dieselbe Zahl Arbeiter.
	return &Packweg{bereich: service.NeuerPackbereich(leser, leser, werk.Planbauer).
		MitArbeitern(problemarbeiter())}, nil
}

// Arten gibt die Packarten eines Sets samt den Befunden ihrer Filter.
func (w *Packweg) Arten(ctx context.Context, setID int64) (PackAntwort, error) {
	arten, befunde, err := w.bereich.Angebot(ctx, setID)
	if err != nil {
		return PackAntwort{}, err
	}

	umgesetzt := make([]PackartDTO, 0, len(arten))
	for _, art := range arten {
		umgesetzt = append(umgesetzt,
			PackartDTO{Name: art.Name, Waehlbar: art.Waehlbar, Karten: art.Karten})
	}

	return PackAntwort{Arten: umgesetzt, Befunde: befundtexte(befunde)}, nil
}

// Ziehen zieht ein Pack und gibt die Karten in ihrer Ziehreihenfolge.
//
// **Der Samen kommt als Text**, wie an der Bindung: eine Zahl über 2⁵³
// verlöre in JSON ihre letzten Stellen, und ein Samen, der beim Hinweg schon
// ein anderer ist, wäre keine Wiederholbarkeit.
func (w *Packweg) Ziehen(ctx context.Context, setID int64, art, same string) (
	PackziehungAntwort, error,
) {
	samen, samenfehler := samenAus(same)
	if samenfehler != nil {
		return PackziehungAntwort{}, fmt.Errorf("%w: %s", domain.ErrValidation, samenfehler.Meldung)
	}

	karten, befunde, err := w.bereich.Ziehen(ctx, setID, art, samen)
	if err != nil {
		return PackziehungAntwort{}, err
	}

	return PackziehungAntwort{
		Karten: kartenDTO(karten), Befunde: befundtexte(befunde),
	}, nil
}
