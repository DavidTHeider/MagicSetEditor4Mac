package service

// **Wann der Bestand gesichert wurde, und wie alt das ist** — OPS01
// (#2347),
// über den Vermerken aus M01 (#2345).
//
// # Wann vermerkt wird
//
// **Nach** dem Zurücklesen und **nach** dem Veröffentlichen: erst dann gibt
// es eine Datei unter ihrem Namen, die geprüft ist ([sicherung.PackenGeprueft]).
// Scheitert vorher etwas, bleibt die Zeit, wie sie war. **Scheitert erst der
// Vermerk, bleibt die Datei** — geschrieben ist geschrieben (§7, Regel 3) —,
// und der Fehler sagt beides ([domain.ErrSicherungUnvermerkt]). Andersherum,
// erst vermerken und dann veröffentlichen, nennte ein Absturz dazwischen
// eine Sicherung, die es nicht gibt; so nennt er schlimmstenfalls ein Alter,
// das zu hoch ist.

import (
	"cmp"
	"context"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// Sicherungsbuch führt die Vermerke der Sicherungen.
//
// **Die Schnittstelle steht hier, weil dieser Anwendungsfall sie braucht**
// (§2); erfüllt wird sie von `storage.Sicherungsvermerke`.
type Sicherungsbuch interface {
	Vermerken(ctx context.Context, vermerk domain.Sicherungsvermerk) error

	// Letzter ist der jüngste Vermerk einer Sicherung, die man wegtragen
	// kann — von Hand oder nach Plan.
	Letzter(ctx context.Context) (domain.Sicherungsvermerk, bool, error)

	// Vermerke sind die eines Anlasses, der jüngste zuerst.
	Vermerke(ctx context.Context, anlass domain.Sicherungsanlass) ([]domain.Sicherungsvermerk, error)
}

// MitSicherungsbuch setzt das Buch — für Tests und den Composition Root.
//
// **Ohne Buch wird nichts vermerkt**, und das Alter ist unbekannt: das sind
// die Läufe dieses Pakets. Jede ausgelieferte Fläche setzt es (`internal/app`,
// gehalten von `nahtverdrahtung_test.go`).
func (b *Bestandsweg) MitSicherungsbuch(buch Sicherungsbuch) { b.vermerke = buch }

// Sicherungsstand sagt, wie alt die letzte Sicherung ist.
//
// **Ein Fehler beim Lesen bleibt ein Fehler** und wird nicht zu „unbekannt":
// das hieße, es nie gewusst zu haben.
func (b *Bestandsweg) Sicherungsstand(ctx context.Context) (domain.Sicherungsstand, error) {
	if b.vermerke == nil {
		return domain.SicherungsstandBei(domain.Sicherungsvermerk{}, false, b.uhr()), nil
	}

	letzter, bekannt, fehler := b.vermerke.Letzter(ctx)
	if fehler != nil {
		return domain.Sicherungsstand{}, fehler
	}

	return domain.SicherungsstandBei(letzter, bekannt, b.uhr()), nil
}

// Planvermerk gibt den jüngsten Vermerk des Plans **in seinem Ordner** —
// falsch, wenn dort keiner liegt oder es kein Buch gibt.
//
// **Am Ordner und nicht am Bestand** (Zweitblick über Phase 4, Befund 3): ein
// Vermerk kennt nur den Dateinamen. Wer den Plan von A nach B legte, hatte in
// B keine Sicherung, und der Vermerk aus A ließ B eine Woche lang leer — samt
// einem „zuletzt", das zu A gehörte. Wiedererkannt wird an Name und Größe;
// die Prüfsumme braucht erst das Ausdünnen.
//
// **Lässt sich der Ordner nicht lesen** — ein Datenträger, der fehlt —, oder
// gibt es keinen, gilt der jüngste Vermerk überhaupt: ob er fällig ist, sagt
// dann die Zeit, und dass der Ordner fehlt, sagt der Lauf.
func (b *Bestandsweg) Planvermerk(ctx context.Context, ordner string) (domain.Sicherungsvermerk, bool, error) {
	if b.vermerke == nil {
		return domain.Sicherungsvermerk{}, false, nil
	}

	eigene, fehler := b.vermerke.Vermerke(ctx, domain.SicherungNachPlan)
	if fehler != nil || len(eigene) == 0 {
		return domain.Sicherungsvermerk{}, false, fehler
	}

	eintraege, lesefehler := os.ReadDir(ordner)
	if ordner == "" || lesefehler != nil {
		return eigene[0], true, nil //nolint:nilerr // ein Ordner, der fehlt, ist hier kein Fehler: das sagt der Lauf, siehe oben
	}

	da := make(map[string]int64, len(eintraege))

	for _, eintrag := range eintraege {
		if lage, infofehler := eintrag.Info(); infofehler == nil && eintrag.Type().IsRegular() {
			da[eintrag.Name()] = lage.Size()
		}
	}

	for _, vermerk := range eigene {
		if groesse, ja := da[vermerk.Datei]; ja && groesse == vermerk.Groesse {
			return vermerk, true, nil
		}
	}

	return domain.Sicherungsvermerk{}, false, nil
}

// vermerken hält eine geprüfte Sicherung fest.
func (b *Bestandsweg) vermerken(ctx context.Context, ziel string, abdruck sicherung.Abdruck, wunsch Sicherungswunsch) error {
	if b.vermerke == nil {
		return nil
	}

	vermerk := domain.Sicherungsvermerk{
		Gesichert:      b.uhr(),
		Anlass:         cmp.Or(wunsch.Anlass, domain.SicherungVonHand),
		Flaeche:        wunsch.Flaeche,
		Datei:          filepath.Base(ziel),
		Groesse:        abdruck.Groesse,
		Pruefsumme:     abdruck.Summe,
		Verschluesselt: wunsch.Kennwort != "",
	}

	if fehler := b.vermerke.Vermerken(ctx, vermerk); fehler != nil {
		return domain.MitGrund(domain.ErrSicherungUnvermerkt, fehler,
			"sichere noch einmal, wenn der Bestand wieder frei ist — die eben geschriebene Datei ist gut")
	}

	return nil
}
