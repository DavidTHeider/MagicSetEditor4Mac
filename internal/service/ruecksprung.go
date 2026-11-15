package service

// **Was eine Rückkehr kostet, und was sie unklar lässt** — Werkstück O7 von
// Phase 5.2 (#1431).
//
// # Warum das neben [Bestandsweg.Beschauen] steht und nicht darin
//
// **Weil es eine andere Frage ist.** `Beschauen` beantwortet „darf ich diese
// Datei übernehmen, und was liegt darin?"; hier steht „was kostet es, wenn ich
// es tue?" — vier Zahlen nebeneinander und die Sammlung Gegenstand für
// Gegenstand.
//
// # Und warum die Epoche hier entsteht und nicht im Bestand
//
// Der Vergleich braucht **beide** Seiten, und die eine gibt es nach dem Tausch
// nicht mehr. Er läuft deshalb **davor**; geschrieben wird die Epoche danach,
// in den zurückgespielten Bestand hinein.

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Bestandsschau liest den Stand einer **fremden** Bestandsdatei.
//
// **Die Schnittstelle steht hier, weil dieser Anwendungsfall sie braucht**
// (§2) — erfüllt von `storage.BestandsstandAus`.
type Bestandsschau func(ctx context.Context, pfad string) (domain.Bestandsstand, error)

// Eigenstand liest den Stand des **laufenden** Bestands.
//
// **Zwei Nähte und nicht eine mit einem Pfad.** Der laufende Bestand liegt
// hinter einer offenen Verbindung, die die Fassade hält; ihn über seinen
// Dateinamen ein zweites Mal zu öffnen hieße, an derselben Datei zweimal zu
// hängen — und genau das weist `Alleinpruefung` beim Tausch ab.
type Eigenstand func(ctx context.Context) domain.Bestandsstand

// Rueckkehrbuch hält die Epochen — erfüllt von `storage.Wiederherstellungsbuch`.
type Rueckkehrbuch interface {
	Vermerken(ctx context.Context, epoche domain.Wiederherstellungsepoche) (int64, error)
	Offene(ctx context.Context) ([]domain.Wiederherstellungsepoche, error)
	Alle(ctx context.Context) ([]domain.Wiederherstellungsepoche, error)
	Quittieren(ctx context.Context, kennung int64, wann time.Time) error
}

// unbekannterStand ist die Vorgabe der zwei Leser: **„ich kann nicht
// nachsehen"**.
//
// # Warum die Vorgabe nicht „nichts gefunden" ist
//
// Weil beides gleich aussieht und das Gegenteil bedeutet. Ein Weg, dem
// niemand die Naht gesetzt hat, meldete sonst „keine unklaren Mengen" — und
// das ist die eine Auskunft, die niemals falsch sein darf. So meldet er
// stattdessen `HandpruefungNoetig`, und die Epoche steht mit ihrem Grund da.
func unbekannterStand() domain.Bestandsstand {
	return domain.Bestandsstand{Sets: -1, Karten: -1, Decks: -1, Exemplare: -1}
}

// MitRuecksprungschau setzt die drei Nähte des Vergleichs.
//
// **`nil` für einen der beiden Leser lässt die Vorgabe stehen** — sie sagt
// „ich kann nicht nachsehen", und das ist eine Aussage. Für das Buch gibt es
// keine Vorgabe: ohne eines wird nicht zurückgespielt.
func (b *Bestandsweg) MitRuecksprungschau(
	schau Bestandsschau, eigen Eigenstand, buch Rueckkehrbuch,
) {
	if schau != nil {
		b.schau = schau
	}

	if eigen != nil {
		b.eigen = eigen
	}

	b.buch = buch
}

// MitUhr setzt die Uhr — für Läufe, die den Zeitpunkt einer Epoche prüfen.
func (b *Bestandsweg) MitUhr(uhr func() time.Time) { b.uhr = uhr }

// Ruecksprunglage stellt eine Sicherung neben den laufenden Bestand.
//
// **Sie ändert nichts.** Wer wissen will, was ein Zurückspielen kostet,
// bekommt die Auskunft, ohne dass etwas geschieht — dieselbe Zusage wie bei
// [Bestandsweg.Beschauen].
func (b *Bestandsweg) Ruecksprunglage(ctx context.Context, pfad string) (
	domain.Ruecksprunglage, error,
) {
	ausgepackt, auspackfehler := b.auspacken(ctx, pfad, b.orte.Datenbank, Bestandsname)
	if auspackfehler != nil {
		return domain.Ruecksprunglage{}, auspackfehler
	}

	defer ausgepackt.Wegraeumen()

	return b.lageAus(ctx, ausgepackt.Bestand)
}

// lageAus vergleicht eine **ausgepackte** Bestandsdatei mit dem laufenden Stand.
func (b *Bestandsweg) lageAus(ctx context.Context, bestand string) (
	domain.Ruecksprunglage, error,
) {
	sicherung, fehler := b.schau(ctx, bestand)
	if fehler != nil {
		return domain.Ruecksprunglage{}, fehler
	}

	return domain.Ruecksprungvergleich(sicherung, b.eigen(ctx)), nil
}

// lageOderHandpruefung vergleicht — oder sagt, dass es nicht ging.
//
// **Der Unterschied zu [Bestandsweg.Ruecksprunglage]**: dort hat jemand nach
// dem Vergleich gefragt, und eine Antwort, die keine ist, wäre eine Lüge. Hier
// ist er die Nebenrechnung eines Tauschs, der ohnehin stattfindet.
func (b *Bestandsweg) lageOderHandpruefung(
	ctx context.Context, bestand string,
) domain.Ruecksprunglage {
	lage, fehler := b.lageAus(ctx, bestand)
	if fehler != nil {
		slog.Default().ErrorContext(ctx,
			"der Stand der Sicherung ließ sich für den Vergleich nicht lesen",
			"grund", fehler.Error(), "ort", bestand)

		return domain.Ruecksprungvergleich(unbekannterStand(), unbekannterStand())
	}

	return lage
}

// epocheVermerken schreibt die Rückkehr in den **zurückgespielten** Bestand.
//
// **Danach und nicht davor**: vorher gehörte sie in eine Datei, die der Tausch
// gleich ersetzt. Der Aufrufer hat die Lage von vorher in der Hand — sie ist
// die einzige Stelle, an der beide Seiten noch dastanden.
//
// **Und ohne Mengen entsteht keine Epoche.** Eine Rückkehr, die nichts unklar
// lässt, ist eine gewöhnliche Wiederherstellung; ein Eintrag darüber wäre eine
// Sperre ohne Gegenstand, die jemand quittieren müsste, damit sie verschwindet.
//
// **Es sei denn, niemand konnte nachsehen**: dann steht die Epoche **ohne**
// Mengen da und sagt genau das — das Issue verlangt die sichtbare Unsicherheit
// bei fehlenden Nachweisen.
func (b *Bestandsweg) epocheVermerken(
	ctx context.Context, lage domain.Ruecksprunglage, quelle string, jetzt time.Time,
) error {
	if len(lage.Unklar) == 0 && !lage.HandpruefungNoetig {
		return nil
	}

	// **Ohne Buch wird nicht zurückgespielt, sondern gemeldet.** Ein Weg, dem
	// niemand das Buch gesetzt hat, hat die Zusage nicht — und sie still
	// fallen zu lassen wäre schlimmer, als sie nie gegeben zu haben (§10).
	if b.buch == nil {
		return fmt.Errorf("%w: dieser Weg führt kein Buch über Rückkehren",
			domain.ErrConflict)
	}

	_, fehler := b.buch.Vermerken(ctx, domain.Wiederherstellungsepoche{
		Zurueckgespielt: jetzt,
		Datenstand:      lage.Datenstand,
		Quelle:          quelle,
		Klaerung:        domain.KlaerungOffen,
		Mengen:          lage.Unklar,
	})

	return fehler
}

// OffeneRueckkehren gibt die Rückkehren, über die noch niemand entschieden hat.
func (b *Bestandsweg) OffeneRueckkehren(ctx context.Context) (
	[]domain.Wiederherstellungsepoche, error,
) {
	return b.buch.Offene(ctx)
}

// Rueckkehren gibt jede Rückkehr — auch die quittierten.
func (b *Bestandsweg) Rueckkehren(ctx context.Context) (
	[]domain.Wiederherstellungsepoche, error,
) {
	return b.buch.Alle(ctx)
}

// RueckkehrQuittieren erklärt eine Rückkehr für nachgesehen.
//
// **Es bucht nichts.** Wer quittiert, sagt „ich habe nachgesehen" — nicht „zieh
// diese Karten ab". Eine Anwendung, die daraus eine Buchung machte, verwandelte
// eine Unsicherheit in eine Behauptung, und zwar in eine, die niemand mehr
// rückgängig macht.
func (b *Bestandsweg) RueckkehrQuittieren(ctx context.Context, kennung int64) error {
	return b.buch.Quittieren(ctx, kennung, b.uhr())
}

// Handelssperre gibt die Sperre aus allen offenen Rückkehren.
//
// **Eine Stelle für die Frage „darf das raus?"** — sie ist der Vertrag, an dem
// T2 und T6 hängen (Phase 5.6). Wer eine Menge als verfügbar meldet, fragt
// hier, oder er meldet Bestände, über die nichts feststeht.
func (b *Bestandsweg) Handelssperre(ctx context.Context) (domain.Handelssperre, error) {
	offene, fehler := b.buch.Offene(ctx)
	if fehler != nil {
		return domain.Handelssperre{}, fehler
	}

	var mengen []domain.Sammlungsmenge
	for _, epoche := range offene {
		mengen = append(mengen, epoche.Mengen...)
	}

	return domain.NeueHandelssperre(mengen), nil
}
