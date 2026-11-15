package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// bogenplan ist, was dem Zeichnen eines Bogens vorausgeht: welche Karten auf
// welches Raster kommen.
//
// # Ein Weg, zwei Ausgänge
//
// **Der echte Lauf und der Probelauf gehen beide hier hindurch**
// (#2015).
// Was vor dem ersten Zeichnen abgewiesen wird, wird in beiden gleich
// abgewiesen — ohne dass ein zweiter Prüfer neben dem Lauf stünde, der
// irgendwann anders urteilte als er.
type bogenplan struct {
	karten  []*domain.CardRow
	raster  bogenraster
	jeSeite int
}

// plane prüft Maße, Set und Auswahl und gibt, was auf den Bogen kommt.
//
// **Ohne den Schreibriegel und ohne den Platz im Haushalt**: beides gehört zum
// Schreiben und schweres Zeichnen, und der Probelauf tut weder noch.
func (b *Bogenleger) plane(
	ctx context.Context, setID int64, masse Bogenmasse, auswahl domain.Kartenauswahl,
) (*bogenplan, error) {
	if err := masse.pruefe(); err != nil {
		return nil, err
	}

	// **Erst die Frage, ob es dieses Set gibt** (#1291). `Karten` gibt für ein
	// leeres Set und für ein Set, das es nicht gibt, dieselbe leere Liste; ohne
	// diese Zeile endete `bogen 999` in einer PDF-Datei mit null Seiten und
	// meldete Erfolg. Am Fenster fiel es nie auf, weil die Kennung dort aus
	// einer Liste kommt — die Fläche ohne Liste ist die, die es zeigt.
	if err := b.setDaSein(ctx, setID); err != nil {
		return nil, err
	}

	alle, err := b.karten.Karten(ctx, setID)
	if err != nil {
		return nil, fmt.Errorf("die Karten lesen: %w", err)
	}

	if fehler := auswahl.Pruefe(alle); fehler != nil {
		return nil, fehler
	}

	karten := auswahl.Anwenden(alle)

	// **Kein Bogen ohne eine einzige Karte** — Befund 6 des Zweitblicks über
	// Phase 4.4 (#1291).
	//
	// # Warum das die vorige Entscheidung umdreht
	//
	// Hier stand seit Phase 3.6: „ein Bogen ohne Karten ist ein Bogen ohne
	// Seiten — und kein Fehler. Ein leeres Set ist eine Lage, kein Versehen."
	// Das erklärt das **Fachereignis** richtig und übersieht das **Artefakt**:
	// eine PDF-Datei mit null Seiten öffnet kein Betrachter. Genau diesen Satz
	// hat dieselbe Phase als Begründung benutzt, um ein **fehlendes** Set
	// abzuweisen — und für das leere stehen lassen. Zwei Wege, ein Artefakt,
	// zwei Antworten.
	//
	// **`ErrValidation` und nicht `ErrNotFound`**: das Set gibt es, es hat nur
	// nichts, was sich legen ließe. Die Meldung sagt, was zu tun ist.
	//
	// # Und warum das **nur** hier gilt
	//
	// Nicht „leer ist ein Fehler", sondern „**das Artefakt ist unbrauchbar**".
	// Die anderen Ausgabewege sind am 2026-08-30 mit einem leeren Set gefahren
	// worden (Nachlauf zu #1291), und sie schreiben etwas Lesbares:
	//
	//   - `mse mtgjson-schreiben` gibt `{"data":{"cards":[],…}}` — gültiges
	//     JSON und eine richtige Antwort.
	//   - `mse vorlagen … ausfuehren` gibt die Kopfzeile der Vorlage und keine
	//     Karte darunter — auch das ist lesbar.
	//
	// Nur das PDF hat die Eigenschaft, dass **null Seiten kein Dokument
	// sind**. Die Zusage hier ist also keine über leere Sets, sondern eine
	// über Dateien, die sich öffnen lassen.
	if len(karten) == 0 {
		return nil, fmt.Errorf("%w: dieses Set hat keine Karte — ein Bogen "+
			"ohne Seiten ist keine Datei, die sich öffnen lässt", domain.ErrValidation)
	}

	raster := masse.raster()

	return &bogenplan{karten: karten, raster: raster, jeSeite: raster.spalten * raster.zeilen}, nil
}
