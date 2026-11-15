// **Die Verschlüsselung an- und ausschalten** — Werkstück 14 von Phase 4.12
// (#1351),
// nach [0072](../../docs/entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md).
//
// # Warum das hier steht und nicht in `storage`
//
// Weil es ein **Lebenszyklus** ist und keine Speicherfrage: schließen, prüfen,
// umschreiben, tauschen, öffnen — und bei jedem Schritt die Frage, was gilt,
// wenn er nicht glückt. `storage` weiß, wie eine Chiffre geht; **wann** sie
// angelegt wird, entscheidet der Anwendungsfall (§2).
//
// Es ist derselbe Weg wie beim Zurückspielen, und die Kommentare dort gelten
// Wort für Wort: erst schließen, dann fragen, ob wir allein sind; den
// Alleinanspruch ausdrücklich zurückgeben, bevor neu geöffnet wird; und die
// Anwendung bekommt ihre Datenbank auf **jedem** Weg hinaus zurück.

package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Umschreibname ist die Datei, in die umgeschrieben wird, bevor getauscht wird.
//
// **Neben dem Bestand und nicht in `/tmp`.** Ein Tausch über eine
// Dateisystemgrenze ist kein Umbenennen, sondern ein Kopieren — und dann gibt
// es einen Augenblick, in dem beide Hälften unterwegs sind.
const Umschreibname = "bestand-umschreiben.db"

// Bestandschiffre ist die Verschlüsselung, wie dieser Anwendungsfall sie
// braucht.
//
// **Die Schnittstelle steht hier, weil der Anwendungsfall sie braucht** (§2).
// Erfüllt wird sie von `internal/storage`; verdrahtet im Composition Root.
type Bestandschiffre interface {
	// IstVerschluesselt sagt, ob die Datei eine Chiffre trägt.
	IstVerschluesselt(pfad string) (bool, error)

	// NeuerSchluessel würfelt einen.
	//
	// **Ohne Fehler**: Würfeln kann nicht scheitern, und ein Rückgabewert, den
	// niemand je füllt, wäre ein Zweig ohne Fall (§3). Warum das für
	// `crypto/rand` nachgelesen und nicht angenommen ist, steht bei
	// `storage.NeuerSchluessel`.
	NeuerSchluessel() string

	// Verschluesseln schreibt den **geöffneten** Bestand verschlüsselt an
	// einen Ort.
	Verschluesseln(ctx context.Context, ziel, schluessel string) error

	// Entschluesseln schreibt ihn im Klartext an einen Ort.
	Entschluesseln(ctx context.Context, ziel string) error

	// SchluesselSchreiben legt den Schlüssel dort ab, wo er hingehört.
	SchluesselSchreiben(ctx context.Context, schluessel string) error

	// SchluesselLoeschen nimmt ihn wieder heraus.
	SchluesselLoeschen(ctx context.Context) error
}

// MitBestandschiffre setzt die Naht — für Tests und den Composition Root.
func (b *Bestandsweg) MitBestandschiffre(chiffre Bestandschiffre) {
	b.chiffre = chiffre
}

// Verschluesselungsstand sagt, ob der Bestand verschlüsselt auf der Platte
// liegt.
func (b *Bestandsweg) Verschluesselungsstand() (bool, error) {
	if b.chiffre == nil {
		return false, fmt.Errorf(
			"%w: diese Fläche kennt die Verschlüsselung nicht", domain.ErrValidation)
	}

	return b.chiffre.IstVerschluesselt(b.orte.Datenbank)
}

// VerschluesselungEinschalten schreibt den Bestand verschlüsselt neu.
//
// # Der Schlüssel liegt im Bund, bevor die Datei ihn braucht
//
// **Das ist die eine Regel, an der Daten hängen.** Zwischen „die Datei ist
// verschlüsselt" und „der Schlüssel liegt im Bund" darf kein Augenblick
// liegen, in dem das erste gilt und das zweite nicht — dort verliert jemand
// alles, und zwar endgültig.
//
// Andersherum ist die Lücke harmlos: ein Schlüssel im Bund, während die Datei
// noch Klartext ist, stört niemanden. `storage.schluesselFuer` fragt den Bund
// gar nicht erst, solange der Dateikopf die Klartextmarke trägt.
//
// **Deshalb bleibt der Schlüssel auch liegen, wenn es schiefgeht.** Ihn
// wegzuräumen hieße, in dem Augenblick zu löschen, in dem am wenigsten
// feststeht, was auf der Platte liegt.
func (b *Bestandsweg) VerschluesselungEinschalten(ctx context.Context) (bool, error) {
	return b.umschalten(ctx, true)
}

// VerschluesselungAusschalten schreibt ihn wieder im Klartext.
//
// **Es gibt ihn, weil eine Entscheidung, die sich nicht zurücknehmen lässt,
// keine ist.** Wer den Bestand weitergeben oder auf einen anderen Rechner
// tragen will, kommt sonst nicht mehr heran.
//
// **Der Schlüssel fällt erst nach dem Tausch** — dieselbe Regel von der
// anderen Seite: er muss so lange da sein, wie die Datei ihn brauchen könnte.
func (b *Bestandsweg) VerschluesselungAusschalten(ctx context.Context) (bool, error) {
	return b.umschalten(ctx, false)
}

// umschalten ist der gemeinsame Weg beider Richtungen.
//
// # Der Stand am Ende wird **gelesen** und nicht behauptet
//
// **Weil die Fläche ihn sonst nachladen müsste** (§7, Regel 3) — und zwischen
// Antwort und Nachladen läge ein Stand, den niemand kennt. Gelesen wird er
// hier und nicht beim Aufrufer: dort wäre es ein zweiter Ruf mit einem eigenen
// Fehlerweg, den kein Fall je betritt.
func (b *Bestandsweg) umschalten(ctx context.Context, an bool) (bool, error) {
	ist, err := b.Verschluesselungsstand()
	if err != nil {
		return false, err
	}

	if ist == an {
		return ist, fmt.Errorf("%w: der Bestand ist %s",
			domain.ErrValidation, standsatz(ist))
	}

	umweg := filepath.Join(filepath.Dir(b.orte.Datenbank), Umschreibname)

	// **Ein Rest von einem abgebrochenen Lauf steht im Weg.** `VACUUM INTO`
	// weigert sich, in eine Datei zu schreiben, die es schon gibt.
	if fehler := os.Remove(umweg); fehler != nil && !errors.Is(fehler, os.ErrNotExist) {
		return ist, fmt.Errorf(
			"den Rest eines früheren Umschreibens wegräumen: %w", fehler)
	}

	// **Umgeschrieben wird, solange die Verbindung offen ist** — sie ist die
	// Quelle. Erst danach beginnt der Lebenszyklus.
	if fehler := b.umschreiben(ctx, umweg, an); fehler != nil {
		return ist, fehler
	}

	defer os.Remove(umweg) //nolint:errcheck // ein Rest stört den nächsten Lauf, nicht diesen

	if fehler := b.tauschen(ctx, umweg, an); fehler != nil {
		// **Der Stand wird auch im Fehlerfall gelesen**, denn der Tausch kann
		// zur Hälfte stattgefunden haben — und dann ist die Frage „was liegt
		// jetzt da?" die einzige, die zählt. Scheitert auch das Lesen, gilt
		// der Fehler des Tauschs: er ist der, der zuerst geschah.
		jetzt, lesefehler := b.Verschluesselungsstand()
		if lesefehler != nil {
			return ist, fehler
		}

		return jetzt, fehler
	}

	return b.Verschluesselungsstand()
}

// umschreiben legt die neue Fassung des Bestands daneben.
func (b *Bestandsweg) umschreiben(ctx context.Context, umweg string, an bool) error {
	if !an {
		return b.chiffre.Entschluesseln(ctx, umweg)
	}

	schluessel := b.chiffre.NeuerSchluessel()

	// **Erst in den Bund, dann in die Datei.** Der Grund steht bei
	// [Bestandsweg.VerschluesselungEinschalten].
	if fehler := b.chiffre.SchluesselSchreiben(ctx, schluessel); fehler != nil {
		return fehler
	}

	return b.chiffre.Verschluesseln(ctx, umweg, schluessel)
}

// tauschen setzt die neue Fassung an die Stelle des Bestands.
//
// **Wort für Wort der Weg aus [Bestandsweg.Zurueckspielen]**, und die
// Begründungen stehen dort: erst schließen, dann nach dem Alleinsein fragen;
// den Anspruch ausdrücklich zurückgeben, bevor neu geöffnet wird; und die
// Anwendung bekommt ihre Datenbank auf jedem Weg hinaus zurück.
func (b *Bestandsweg) tauschen(ctx context.Context, umweg string, an bool) error {
	if fehler := b.schliessen(); fehler != nil {
		return fmt.Errorf("den Bestand schließen: %w", fehler)
	}

	anspruch, err := b.allein(ctx, b.orte.Datenbank)
	if err != nil {
		return errors.Join(err, b.oeffnen(context.WithoutCancel(ctx)))
	}

	zurueckgeben := einmalig(anspruch)
	defer zurueckgeben()

	zurueck, fehler := bestandErsetzen(umweg, b.orte.Datenbank)
	if fehler != nil {
		return errors.Join(fehler, b.wiederOeffnen(ctx, zurueckgeben))
	}

	if oeffnungsfehler := b.wiederOeffnen(ctx, zurueckgeben); oeffnungsfehler != nil {
		return b.zurueckDrehen(ctx, zurueck, zurueckgeben, oeffnungsfehler)
	}

	// **Und jetzt erst darf der Schlüssel fallen.** Vorher stünde ein
	// Augenblick offen, in dem die Datei noch verschlüsselt ist und der
	// Schlüssel schon fort.
	if !an {
		if loeschfehler := b.chiffre.SchluesselLoeschen(ctx); loeschfehler != nil {
			return fmt.Errorf(
				"der Bestand liegt wieder im Klartext, aber der alte Schlüssel "+
					"blieb im Schlüsselbund liegen: %w", loeschfehler)
		}
	}

	return nil
}

// standsatz sagt, wie der Bestand liegt.
func standsatz(verschluesselt bool) string {
	if verschluesselt {
		return "schon verschlüsselt"
	}

	return "gar nicht verschlüsselt"
}
