package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// Bulkquelle ist eine abgelegte Sammeldatei, in der gesucht werden kann.
type Bulkquelle interface {
	KarteNachName(ctx context.Context, name string) (
		*scryfall.Karte, []domain.Befund, error)
}

// bulkrueckfall fragt erst das Netz und dann die Datei.
type bulkrueckfall struct {
	netz  Fremdkarten
	lokal Bulkquelle
}

// MitBulkrueckfall legt eine abgelegte Sammeldatei hinter einen Klienten.
//
// # Warum hinter und nicht davor
//
// Die Sammeldatei ist **12 bis 24 Stunden alt** — Scryfall sagt es selbst:
// „Bulk data is only collected once every 12-24 hours." Sie zuerst zu fragen
// hieße, freiwillig einen alten Stand zu nehmen; sie danach zu fragen heißt:
// lieber ein alter Stand als gar keiner. Das ist §16, Regel 5 — „Ohne Netz
// arbeitet die Anwendung weiter. Der letzte gute Stand bleibt, und er sagt,
// wie alt er ist."
//
// # Warum es ein Umschlag ist und kein Feld im Klienten
//
// Weil jeder Verbraucher von [Fremdkarten] davon hat, ohne es zu wissen: der
// Feldabgleich, der Bildweg und die Regelauskunft fragen alle über dieselbe
// Naht. Ein Feld im Klienten hätte dieselbe Wirkung und läge in einem Blatt,
// das von der Ablage auf der Platte nichts wissen soll (§2).
//
// **`nil` als Datei ist erlaubt**: dann verhält sich der Umschlag wie der
// Klient allein. Wer keine Sammeldatei geholt hat, soll die Anwendung
// benutzen können.
func MitBulkrueckfall(netz Fremdkarten, lokal Bulkquelle) Fremdkarten {
	return &bulkrueckfall{netz: netz, lokal: lokal}
}

// bulkalter ist das Alter, das eine Antwort aus der Sammeldatei bekommt.
//
// # Warum 24 Stunden und keine Messung
//
// Gemessen werden könnte das Alter der **Datei**, und das wäre falsch: die
// Datei kann eine Minute alt sein und Daten von gestern enthalten. Scryfall
// sammelt „once every 12-24 hours"; die obere Zahl ist die ehrliche.
//
// **Null wäre die falsche Auskunft** — sie hieße „gerade geholt".
const bulkalter = 24 * time.Hour

// KarteNachName fragt das Netz und fällt auf die Datei zurück.
func (b *bulkrueckfall) KarteNachName(ctx context.Context, name string) (
	*scryfall.Karte, time.Duration, []domain.Befund, error,
) {
	karte, alter, befunde, fehler := b.netz.KarteNachName(ctx, name)
	if fehler == nil {
		return karte, alter, befunde, nil
	}

	// **Ein fachliches Nein geht nicht in die Datei.** Eine Karte, die es
	// nicht gibt, gibt es auch dort nicht — und eine ältere Datei könnte eine
	// führen, die Scryfall zurückgezogen hat.
	if b.lokal == nil || errors.Is(fehler, domain.ErrNotFound) {
		return nil, 0, nil, fehler
	}

	ausDatei, dateibefunde, dateifehler := b.lokal.KarteNachName(ctx, name)
	if dateifehler != nil {
		// **Der Netzfehler steht vorn**, denn er ist der Grund; dass es
		// keinen Rückfall gibt, ist die Folge.
		return nil, 0, nil, fmt.Errorf("%w (und die Sammeldatei half nicht: %w)",
			fehler, dateifehler)
	}

	hinweis := domain.Befund{
		Ort:     "Sammeldatei",
		Art:     domain.ArtUmgebung,
		Meldung: rueckfallgrund(fehler),
	}

	return ausDatei, bulkalter, append([]domain.Befund{hinweis}, dateibefunde...), nil
}

// rueckfallgrund sagt, **warum** die Antwort aus der Datei kommt.
//
// # Warum nicht ein Satz für alle
//
// Weil „Scryfall war nicht erreichbar" bei einem 429 **falsch** ist: Scryfall
// war bestens erreichbar und hat gesagt „zu schnell". Wer den Satz liest, sucht
// den Fehler bei seiner Verbindung statt bei der Drossel dieses Programms — und
// die 429 ist genau das Zeichen, dass sie nicht reicht.
//
// Eigener Befund und Teil von Befund 2 des Zweitblicks über Phase V
// (#888).
func rueckfallgrund(fehler error) string {
	const woher = "die Angaben stammen aus der abgelegten Sammeldatei und sind " +
		"bis zu einen Tag alt."

	if errors.Is(fehler, domain.ErrGedrosselt) {
		return "Scryfall hat zu viele Anfragen gemeldet und wartet ab; " + woher
	}

	return "Scryfall war nicht erreichbar; " + woher
}
