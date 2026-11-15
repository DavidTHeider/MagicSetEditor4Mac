// **Gemeinsam oder gar nicht** — Werkstück 1 von
// #1162.
//
// # Warum es eine Marke braucht
//
// Der Tausch hat zwei Schritte in **zwei Verzeichnissen**: `bestand.db` und
// `beiwerk/`. Für gemeldete Fehler gibt es einen Rückweg — er läuft im
// Aufrufer und ist geprüft. Für einen **Abbruch** läuft er nicht: stirbt der
// Prozess dazwischen, steht die neue Datenbank neben den alten Bildern.
//
// **Unteilbar geht es nicht.** Zwei Umbenennungen in zwei Verzeichnissen
// zusammen auszuführen kann kein Dateisystem; wer es behauptete, machte eine
// Zusage, die er nicht halten kann.
//
// **Was geht, ist Zuendeführen.** Vor dem ersten Schritt entsteht eine Marke,
// die den zweiten vollständig beschreibt; nach dem zweiten verschwindet sie.
// Findet ein Start sie vor, war ein Zurückspielen unterwegs — und er führt es
// zu Ende, statt einen gemischten Stand stehen zu lassen.
//
// **Es braucht keine Rückfrage.** Was zu tun ist, steht in der Marke, und es
// ist genau das, was die Nutzenden bereits bestätigt haben.

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Rueckspielmarke ist der Name der Marke neben dem Bestand.
const Rueckspielmarke = "rueckspiel-laeuft.json"

// rueckspielstand beschreibt den **zweiten** Schritt vollständig.
//
// **Der erste steht nicht darin.** Er ist beim Schreiben der Marke noch nicht
// geschehen und danach unteilbar erledigt; was offenbleiben kann, ist allein
// der Bildertausch.
type rueckspielstand struct {
	// Quelle ist die gewählte Sicherung — für das Protokoll, nicht für die
	// Arbeit.
	Quelle string `json:"quelle"`

	// Beiwerk ist der Ort der neuen Bilder im Auszug.
	//
	// **Leer bei [BilderGeleert]**: dann gibt es keine, und das ist die
	// Absicht und kein Mangel.
	Beiwerk string `json:"beiwerk"`

	// Bilder ist die **Absicht** — [BilderErsetzt] oder [BilderGeleert].
	//
	// **Ohne sie ließe sich beim Leeren „noch nicht geschehen" nicht von
	// „schon geschehen" unterscheiden**: in beiden Fällen liegt im Auszug
	// kein `beiwerk/`, und ein Start räumte dann ein zweites Mal.
	Bilder string `json:"bilder"`
}

// markenpfad gibt den Ort der Marke.
func (b *Bestandsweg) markenpfad() string {
	return filepath.Join(filepath.Dir(b.orte.Datenbank), Rueckspielmarke)
}

// markeSchreiben legt die Marke an — **vor** dem ersten Schritt.
func (b *Bestandsweg) markeSchreiben(stand rueckspielstand) error {
	// **Entwerfen und Schreiben auf einem Weg.** `json.Marshal` kann an
	// diesem Verbund aus drei Zeichenketten nicht scheitern; ein eigener
	// Zweig dafür wäre eine Behauptung darüber, was im Ernstfall geschieht,
	// die kein Test betreten kann (§3).
	roh, err := json.Marshal(stand)
	if err == nil {
		err = schreibeMarke(b.markenpfad(), roh)
	}

	if err != nil {
		return fmt.Errorf("die Rückspielmarke schreiben: %w", err)
	}

	return nil
}

// markeEntfernen nimmt die Marke weg — **nach** dem zweiten Schritt.
func (b *Bestandsweg) markeEntfernen() error {
	if fehler := os.Remove(b.markenpfad()); fehler != nil &&
		!errors.Is(fehler, os.ErrNotExist) {
		return fmt.Errorf("die Rückspielmarke wegräumen: %w", fehler)
	}

	return nil
}

// markeAufraeumen nimmt die Marke weg und schreibt auf, wenn sie liegen blieb.
//
// **Eine Stelle für zwei Wege** — das Zurückspielen und der Neuanfang (O4,
// #1428) haben denselben `defer`, denselben Zweig und denselben Satz. Zwei
// Kopien davon hießen zwei Zweige, von denen ein Lauf nur einen betritt (§3).
func (b *Bestandsweg) markeAufraeumen(ctx context.Context) {
	if markenfehler := b.markeEntfernen(); markenfehler != nil {
		slog.Default().ErrorContext(ctx, "die Rückspielmarke blieb liegen",
			"grund", markenfehler.Error())
	}
}

// RueckspielVollenden führt ein abgebrochenes Zurückspielen zu Ende.
//
// **Beim Start und genau einmal je Prozess.** [App.Starten] läuft auch
// **während** eines Zurückspielens, unmittelbar nach dem Tausch; liefe die
// Vollendung dort erneut, arbeitete sie gegen den Vorgang, der sie ausgelöst
// hat.
//
// **Ein Fehler hält den Start nicht auf.** Die Anwendung ist benutzbar; was
// fehlt, sind Bilder — und ein Programm, das deswegen nicht startet, nimmt den
// Nutzenden auch den Weg, es von Hand zu richten.
//
// **Aber er wird gemeldet.** Stünde er nur im Protokoll, öffnete das Fenster
// danach wortlos die schon getauschte Datenbank: neue Daten, alte Bilder, und
// niemand erführe davon. Der Rückgabewert ist der Satz fürs Fenster — leer
// heißt „nichts liegengeblieben"
// (#1170).
func (b *Bestandsweg) RueckspielVollenden() string {
	roh, err := os.ReadFile(b.markenpfad()) // #nosec G304 -- neben dem Bestand
	if err != nil {
		// **Keine Marke ist der Normalfall** und keine Meldung wert.
		if !errors.Is(err, os.ErrNotExist) {
			slog.Default().Error("die Rückspielmarke ließ sich nicht lesen",
				"grund", err.Error())

			return "Eine Rückspielmarke liegt neben dem Bestand und ließ sich " +
				"nicht lesen. Die Bilder könnten zu einem anderen Stand gehören " +
				"als die Daten."
		}

		return ""
	}

	var stand rueckspielstand
	if fehler := json.Unmarshal(roh, &stand); fehler != nil {
		slog.Default().Error("die Rückspielmarke ist unlesbar",
			"grund", fehler.Error())

		return "Eine Rückspielmarke liegt neben dem Bestand und ist unlesbar. " +
			"Die Bilder könnten zu einem anderen Stand gehören als die Daten."
	}

	if fehler := b.bildertauschVollenden(stand); fehler != nil {
		slog.Default().Error(
			"ein abgebrochenes Zurückspielen ließ sich nicht vollenden",
			"grund", fehler.Error(), "ort", stand.Quelle)

		// **Die Marke bleibt liegen.** Der nächste Start versucht es wieder;
		// sie wegzuräumen hieße, den halben Stand für endgültig zu erklären.
		return "Ein abgebrochenes Zurückspielen ließ sich nicht abschließen: " +
			fehler.Error() + " Die Bilder gehören möglicherweise noch zum " +
			"vorigen Stand."
	}

	slog.Default().Info("ein abgebrochenes Zurückspielen wurde vollendet",
		"ort", stand.Quelle, "art", stand.Bilder)

	if fehler := b.markeEntfernen(); fehler != nil {
		slog.Default().Error("die Rückspielmarke blieb liegen",
			"grund", fehler.Error())
	}

	return ""
}

// bildertauschVollenden holt den zweiten Schritt nach — oder stellt fest, dass
// er schon geschehen ist.
func (b *Bestandsweg) bildertauschVollenden(stand rueckspielstand) error {
	if stand.Bilder == BilderGeleert {
		ziel, err := b.orte.Bilder()
		if err != nil {
			return fmt.Errorf("das Bilderverzeichnis finden: %w", err)
		}

		if _, statfehler := os.Stat(ziel); errors.Is(statfehler, os.ErrNotExist) {
			// **Schon geleert.** Der Abbruch lag zwischen Tausch und Marke.
			return nil
		}

		_, fehler := b.beiwerkErsetzen("")

		return fehler
	}

	if _, err := os.Stat(stand.Beiwerk); errors.Is(err, os.ErrNotExist) {
		// **Fehlt der Auszug, entscheidet der Elternordner.**
		//
		// `beiwerkErsetzen` benennt das `beiwerk/` darin um und lässt den
		// Zwischenplatz stehen; wegzuräumen ist er die Aufgabe des Aufrufers.
		// Steht er noch, **ist getauscht worden** — das war der Fall, den die
		// alte Fassung meinte.
		//
		// Sind beide weg, weiß niemand etwas: der Zwischenplatz kann
		// weggeräumt worden sein, ohne dass je getauscht wurde. Bis zum
		// 2026-08-28 galt auch das als erledigt — **ohne Gegenbeweis**.
		if _, platzfehler := os.Stat(filepath.Dir(stand.Beiwerk)); platzfehler == nil {
			return nil
		}

		return fmt.Errorf("%w: der Auszug %q ist samt seinem Zwischenplatz "+
			"verschwunden, und damit ist nicht zu sehen, ob die Bilder je "+
			"getauscht wurden", domain.ErrConflict, stand.Beiwerk)
	}

	_, fehler := b.beiwerkErsetzen(stand.Beiwerk)

	return fehler
}
