// **Die Unsicherheit überlebt den Prozess** — Befund 10 des Zweitblicks über
// Phase 5.2 (#1435).
//
// # Der Befund
//
// Die Epoche einer Rückkehr ([0088](../../docs/entscheidungen/0088-die-epoche-einer-rueckkehr.md))
// entstand **nach** dem Tausch: Bestand ersetzt, wieder geöffnet, dann
// vermerkt. Scheiterte der Vermerk, lief die Anwendung mit dem neuen Bestand
// weiter — und nur der laufende Aufruf endete mit einer Meldung. Nach einem
// Neustart oder einem weggeklickten Fenster fehlte der Nachweis für die
// Unsicherheit, den 0088 ausdrücklich als **nicht verschwindend** zusagt.
//
// # Die Antwort ist dieselbe wie beim Bildertausch
//
// **Eine Marke vor dem ersten Schritt**, und ein Start, der sie vorfindet,
// holt nach. Das Muster steht in `rueckspielmarke.go` und wird hier nicht neu
// erfunden: dieselbe Stelle auf der Platte, dieselbe atomare Schreibweise,
// derselbe Aufrufer beim Start.
//
// **Und sie gilt auch für die nackte `.db`.** Der Bildertausch hat seine Marke
// nur, wenn Bilder dabei sind; die Unsicherheit über Mengen hängt daran nicht.

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Epochenmarke ist der Name der Marke neben dem Bestand.
const Epochenmarke = "rueckkehr-offen.json"

// epochenstand beschreibt die Epoche, die noch ins Buch gehört.
type epochenstand struct {
	// Quelle ist der Name der Sicherung — er steht später in der Epoche.
	Quelle string `json:"quelle"`

	// Zurueckgespielt ist der Zeitpunkt des Tauschs.
	//
	// **Aus der Marke und nicht von der Uhr beim Nachholen**: die Epoche
	// gehört zum Tausch, und der war gestern.
	Zurueckgespielt time.Time `json:"zurueckgespielt"`

	Datenstand time.Time `json:"datenstand"`

	// Mengen sind die unklaren Bestände — so, wie die Lage sie gezählt hat.
	Mengen []domain.Sammlungsmenge `json:"mengen"`
}

// epochenmarkenpfad gibt den Ort der Marke.
func (b *Bestandsweg) epochenmarkenpfad() string {
	return filepath.Join(filepath.Dir(b.orte.Datenbank), Epochenmarke)
}

// epochenmarkeSchreiben legt die Marke an — **vor** dem Tausch.
//
// **Ohne unklare Mengen entsteht keine** — dieselbe Bedingung wie beim
// Vermerken selbst: eine Rückkehr, die nichts offen lässt, ist eine gewöhnliche
// Wiederherstellung.
func (b *Bestandsweg) epochenmarkeSchreiben(
	lage domain.Ruecksprunglage, quelle string, jetzt time.Time,
) error {
	if len(lage.Unklar) == 0 && !lage.HandpruefungNoetig {
		return nil
	}

	roh, err := json.Marshal(epochenstand{
		Quelle: quelle, Zurueckgespielt: jetzt,
		Datenstand: lage.Datenstand, Mengen: lage.Unklar,
	})
	if err == nil {
		err = schreibeMarke(b.epochenmarkenpfad(), roh)
	}

	if err != nil {
		return fmt.Errorf("die Rückkehrmarke schreiben: %w", err)
	}

	return nil
}

// epochenmarkeEntfernen nimmt sie weg — **nach** dem Vermerk.
func (b *Bestandsweg) epochenmarkeEntfernen() error {
	if fehler := os.Remove(b.epochenmarkenpfad()); fehler != nil &&
		!errors.Is(fehler, os.ErrNotExist) {
		return fmt.Errorf("die Rückkehrmarke wegräumen: %w", fehler)
	}

	return nil
}

// EpocheNachholen trägt eine liegengebliebene Rückkehr ins Buch.
//
// **Beim Start, neben [Bestandsweg.RueckspielVollenden].** Der Rückgabewert
// ist der Satz fürs Fenster — leer heißt „nichts liegengeblieben".
//
// **Ein Fehler hält den Start nicht auf**, und die Marke bleibt dann liegen:
// der nächste Start versucht es wieder. Sie wegzuräumen hieße, die
// Unsicherheit für geklärt zu erklären, ohne dass jemand nachgesehen hat.
func (b *Bestandsweg) EpocheNachholen(ctx context.Context) string {
	roh, err := os.ReadFile(b.epochenmarkenpfad()) // #nosec G304 -- neben dem Bestand
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			slog.Default().ErrorContext(ctx, "die Rückkehrmarke ließ sich nicht lesen",
				"grund", err.Error())

			return "Eine Rückkehrmarke liegt neben dem Bestand und ließ sich " +
				"nicht lesen. Sieh deine Sammlung von Hand nach."
		}

		return ""
	}

	var stand epochenstand
	if fehler := json.Unmarshal(roh, &stand); fehler != nil {
		slog.Default().ErrorContext(ctx, "die Rückkehrmarke ist unlesbar",
			"grund", fehler.Error())

		return "Eine Rückkehrmarke liegt neben dem Bestand und ist unlesbar. " +
			"Sieh deine Sammlung von Hand nach."
	}

	if fehler := b.epocheAusMarke(ctx, stand); fehler != nil {
		slog.Default().ErrorContext(ctx,
			"eine liegengebliebene Rückkehr ließ sich nicht vermerken",
			"grund", fehler.Error(), "datei", stand.Quelle)

		return "Eine zurückgespielte Sicherung ließ sich nicht vermerken: " +
			fehler.Error() + " Sieh deine Sammlung von Hand nach."
	}

	if fehler := b.epochenmarkeEntfernen(); fehler != nil {
		slog.Default().ErrorContext(ctx, "die Rückkehrmarke blieb liegen",
			"grund", fehler.Error())
	}

	return ""
}

// epocheAusMarke schreibt die Epoche, die in der Marke steht.
//
// **Zweimal dieselbe Marke gibt eine Epoche und nicht zwei** — Befund 8 des
// Nachlaufs zum Zweitblick über Phase 5.2. Lässt sich die Marke nach dem
// Vermerken nicht wegräumen, liest der nächste Start sie erneut; `Vermerken`
// legt aber unbedingt an. Der zweite Start machte damit aus **einer**
// Rückkehr zwei Sperren, die beide jemand quittieren müsste.
//
// **Erkannt wird sie an Zeitpunkt und Quelle** — das ist, was die Marke
// festhält, und zwei Rückkehren aus derselben Datei im selben Augenblick gibt
// es nicht.
func (b *Bestandsweg) epocheAusMarke(ctx context.Context, stand epochenstand) error {
	if b.buch == nil {
		return fmt.Errorf("%w: dieser Weg führt kein Buch über Rückkehren",
			domain.ErrConflict)
	}

	vorhandene, lesefehler := b.buch.Alle(ctx)
	if lesefehler != nil {
		return lesefehler
	}

	for _, epoche := range vorhandene {
		if epoche.Quelle == stand.Quelle &&
			epoche.Zurueckgespielt.Equal(stand.Zurueckgespielt) {
			return nil
		}
	}

	_, fehler := b.buch.Vermerken(ctx, domain.Wiederherstellungsepoche{
		Zurueckgespielt: stand.Zurueckgespielt,
		Datenstand:      stand.Datenstand,
		Quelle:          stand.Quelle,
		Klaerung:        domain.KlaerungOffen,
		Mengen:          stand.Mengen,
	})

	return fehler
}
