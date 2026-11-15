// **Format und Vorlage eines Turniers** — DOM-04.AK-01 und AK-03
// (#1699).
//
// # Was eine Vorlage mitbringt, und was nicht
//
// **Einstellungen und Namen**: Spiel, Format, Einzelspiel, Notiz und — auf
// Wunsch — wer beim letzten Mal dabei war. **Nicht** die Decks und ihre
// Abzüge: eine eingereichte Liste ist die Zusage eines Menschen für **ein**
// Turnier (Vermerk 0062, Entscheidung 3), und wer sie in ein neues trägt,
// meldet jemanden mit einer Liste an, die er dort nie eingereicht hat. Nicht
// die Runden und nicht, wer ausgestiegen ist — das ist Verlauf.
//
// Portierungsmarke **new**: MSE2 kennt keine Turniere (§11, gesucht in
// `/Users/david/mse2-quelle/src` nach `tournament`, `pairing`, `round`).

package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Turnierformate ist, was die Leitung über Formate wissen muss: welches eine
// Kennung meint, und — für eine eingelesene Datei (DOM-04.AK-03) — welche es
// in einem Spiel gibt.
type Turnierformate interface {
	Formatnachschlag
	Formate(ctx context.Context, spiel string) ([]domain.Format, error)
}

// MitFormaten setzt die Naht zu den Formaten — und gibt die Leitung zurück,
// damit ein Composition Root sie beim Bauen dazusetzt.
func (l *Turnierleitung) MitFormaten(formate Turnierformate) *Turnierleitung {
	l.formate = formate

	return l
}

// formatPruefen sagt, ob das gewählte Format zu diesem Turnier passt.
//
// **Die Naht liest im Profil** (`Formatablage.Format`): ein Format aus einem
// anderen Profil ist hier eines, das es nicht gibt.
func (l *Turnierleitung) formatPruefen(ctx context.Context, turnier domain.Turnier) error {
	if turnier.FormatID == 0 {
		return nil
	}

	// **Ohne Naht wird keines angenommen** — ungeprüft abgelegt stünde am
	// Turnier womöglich ein Format eines anderen Spiels.
	if err := l.formatnahtDa(); err != nil {
		return err
	}

	format, err := l.formate.Format(ctx, turnier.FormatID)
	if err != nil {
		return err
	}

	if format.Spiel != turnier.Spiel {
		return domain.NewFieldError("formatID", domain.SchluesselTurnierformatSpiel,
			"Das Format »%s« gehört zum Spiel »%s«, dieses Turnier zu »%s«.",
			format.Name, format.Spiel, turnier.Spiel)
	}

	return nil
}

// Kopieren legt ein neues Turnier nach einer Vorlage an — **mit** den Namen
// ihrer Teilnehmenden, wenn `mitTeilnehmenden` es sagt.
//
// **Durch [Turnierleitung.Anlegen]**, damit die Kopie dieselben Prüfungen
// durchläuft wie ein neues Turnier — und ein Format, das inzwischen zu einem
// anderen Spiel gehörte, nicht still mitkäme.
func (l *Turnierleitung) Kopieren(
	ctx context.Context, quelleID int64, name string, mitTeilnehmenden bool,
) (domain.Turnier, error) {
	quelle, err := l.turniere.Turnier(ctx, quelleID)
	if err != nil {
		return domain.Turnier{}, err
	}

	kopie := domain.Turnier{
		Name: name, Spiel: quelle.Spiel, FormatID: quelle.FormatID,
		Einzelspiel: quelle.Einzelspiel, Notiz: quelle.Notiz,
	}

	if mitTeilnehmenden {
		for _, teilnehmer := range quelle.Teilnehmer {
			kopie.Teilnehmer = append(kopie.Teilnehmer, domain.Teilnehmer{Name: teilnehmer.Name})
		}
	}

	return l.Anlegen(ctx, kopie)
}
