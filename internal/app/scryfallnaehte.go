// **Die Nähte zu Scryfall** — Klient, Drossel, Zwischenspeicher, Rückfall.
//
// **Aus `scryfall.go` herausgelöst**, als es über 400 Zeilen ging (§2). Der
// Schnitt folgt der Frage: dort stehen die **Bindungen** — Kennungen,
// Fehlerübersetzung, DTO —, hier die **Composition Roots**, die sie mit einem
// Klienten versorgen.

package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// ScryfallBildnaht baut den Bildweg samt Klient und Zwischenspeicher.
//
// **Derselbe Klient wie beim Feldabgleich**, und das ist kein Zufall: er
// bringt Drossel und Zwischenspeicher mit, und ein zweiter daneben hätte
// beides doppelt — zwei Drosseln unter derselben Grenze halten sie nicht ein.
func ScryfallBildnaht(wahl Dienstwahl) (Bildquelle, error) {
	// **Beide auf einmal, ein Zweig** — dieselbe Überlegung wie in
	// [Zeichennaht]: sie scheitern an derselben Ursache, dem fehlenden
	// Benutzerverzeichnis. Ein zweites `if` darauf wäre ein Zweig, den kein
	// Aufruf betreten kann (§3).
	ort, ortfehler := Scryfallverzeichnis()

	klient, klientfehler := Scryfallklient(wahl)

	if err := errors.Join(ortfehler, klientfehler); err != nil {
		return nil, err
	}

	// **Ein eigener Ordner neben dem Antwortspeicher.** Bilder sind groß und
	// haben eine andere Frist als die JSON-Antworten; sie im selben Verzeichnis
	// abzulegen hieße, dass ein Aufräumen des einen das andere trifft.
	ordner := filepath.Join(ort, "bilder")
	if fehler := os.MkdirAll(ordner, 0o700); fehler != nil {
		return nil, fmt.Errorf("den Bildordner anlegen: %w", fehler)
	}

	return bildweg{weg: service.NeuerBildweg(
		service.MitBulkrueckfall(klient, Bulkrueckfall()), klient, ordner)}, nil
}

// bildweg übersetzt die Bildart und führt an [service.Bildweg].
//
// **Die Übersetzung steht hier und nicht in der Fassade**: `internal/scryfall`
// gehört zum Composition Root, und die Aufzählung in `internal/schichten` hält
// die Kante auf diese Datei begrenzt (§2).
type bildweg struct{ weg *service.Bildweg }

// Holen nimmt das Wort einer Fläche und gibt das Bild.
func (b bildweg) Holen(ctx context.Context, name, art string) (
	*service.Bilduebernahme, error,
) {
	gewaehlt, gueltig := bildartAus(art)
	if !gueltig {
		return nil, fmt.Errorf(
			"%w: %q ist keine bekannte Bildart — bekannt sind klein, normal und ausschnitt",
			domain.ErrValidation, art)
	}

	return b.weg.Holen(ctx, name, gewaehlt)
}

// bildartAus liest die Bildart aus dem Wort, das eine Fläche schickt.
func bildartAus(wort string) (scryfall.Bildart, bool) {
	switch scryfall.Bildart(wort) {
	case scryfall.BildKlein:
		return scryfall.BildKlein, true
	case scryfall.BildNormal:
		return scryfall.BildNormal, true
	case scryfall.BildAusschnitt:
		return scryfall.BildAusschnitt, true
	default:
		return "", false
	}
}

// Auskunftsnaht baut die Auskunft samt Klient und Zwischenspeicher.
//
// **Derselbe Klient wie Abgleich und Bildweg.** Er bringt beide Drosseln mit;
// ein zweiter daneben hätte sie doppelt, und zwei Drosseln unter derselben
// Grenze halten sie nicht ein.
func Auskunftsnaht(wahl Dienstwahl) (Auskunftsquelle, error) {
	// **Über die eine Stelle** (Phase 4.8) — siehe oben. Das Verzeichnis
	// prüft sie mit; ein eigener Ruf hier wäre eine zweite Antwort auf
	// dieselbe Frage.
	klient, klientfehler := Scryfallklient(wahl)
	if klientfehler != nil {
		return nil, klientfehler
	}

	// **Die Auskunft bekommt ihn ebenso.** Sie fragt zuerst nach der Karte,
	// und ohne Netz käme sie sonst nicht bis zu den Regelhinweisen.
	return service.NeueAuskunft(auskunftMitRueckfall{
		Fremdkarten: service.MitBulkrueckfall(klient, Bulkrueckfall()),
		Klient:      klient,
	}), nil
}

// auskunftMitRueckfall verbindet den Kartenweg mit Rückfall und den Rest.
//
// # Warum ein Verbund und kein Umschlag
//
// [service.MitBulkrueckfall] kennt nur [service.Fremdkarten] — eine Methode.
// Die Auskunft braucht vier. Der Verbund setzt sie zusammen: der Kartenweg
// kommt aus dem Umschlag, die anderen drei unverändert vom Klienten.
//
// **Rulings und Symbologie haben keinen Rückfall**, und das ist Absicht: die
// Sammeldatei `oracle_cards` führt Karten, keine Regelhinweise. Es gäbe eine
// eigene Datei dafür (`rulings`, 5,4 MB) — sie zu holen ist eine Entscheidung
// für ein späteres Werkstück und keine, die hier nebenbei fällt.
type auskunftMitRueckfall struct {
	service.Fremdkarten

	*scryfall.Klient
}

// KarteNachName kommt aus dem Umschlag mit Rückfall.
//
// **Ausdrücklich benannt**, weil beide eingebetteten Typen sie führen — Go
// wählte sonst keine und der Bau bräche.
func (a auskunftMitRueckfall) KarteNachName(ctx context.Context, name string) (
	*scryfall.Karte, time.Duration, []domain.Befund, error,
) {
	return a.Fremdkarten.KarteNachName(ctx, name)
}

// bulkdatei liest Karten aus der abgelegten Sammeldatei.
//
// **Sie prüft beim Lesen, ob es sie gibt** — und nicht beim Bauen. Wer sie
// nach dem Start holt, soll die Anwendung nicht neu starten müssen.
type bulkdatei struct{ pfad string }

// KarteNachName sucht in der Sammeldatei.
func (b bulkdatei) KarteNachName(_ context.Context, name string) (
	*scryfall.Karte, []domain.Befund, error,
) {
	if _, fehler := os.Stat(b.pfad); fehler != nil {
		return nil, nil, fmt.Errorf(
			"%w: es liegt keine Sammeldatei bereit (»mse scryfall-sammeldaten holen«)",
			domain.ErrNotFound)
	}

	return scryfall.BulkKarteNachName(b.pfad, name)
}

// Bulkrueckfall gibt die lokale Sammeldatei als Rückfallquelle.
//
// **`nil` ist kein Fehler**: wer kein Zuhause hat, bekommt keinen Rückfall,
// und die Anwendung läuft trotzdem — dieselbe Haltung wie beim Paketort.
func Bulkrueckfall() service.Bulkquelle {
	pfad, fehler := Bulkpfad()
	if fehler != nil {
		return nil
	}

	return bulkdatei{pfad: pfad}
}

// speicherzahlen übersetzt die Zahlen der Zwischenspeicher — OPS-05 (#1722).
// **Hier**, weil diese Datei `internal/netz` kennen darf; die Metriken
// selbst nicht.
func speicherzahlen() SpeicherDTO {
	stand := netz.Speicherzahlen()

	return SpeicherDTO{
		Treffer: stand.Treffer, Geholt: stand.Geholt,
		Veraltet: stand.Veraltet, Fehlgeschlagen: stand.Fehlgeschlagen,
	}
}
