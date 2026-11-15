package app

// **Die Naht zu Scrydex' Preisen** — Werkstück A5 von Phase 5.3
// (#1444).
//
// Hier wird der Klient gebaut und die Gestalt von Scrydex auf die von `domain`
// gebracht. **Eine fachliche Entscheidung fällt hier keine** (§2): welche
// Fassung zu welchem Druck gehört, entscheidet `service`.

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scrydex"
)

// scrydexpreise bringt den Klienten an [service.Fremdpreisquelle].
type scrydexpreise struct{ klient *scrydex.Klient }

// Familie ist der Name, unter dem Preise von Scrydex laufen.
func (s scrydexpreise) Familie() domain.Preisquelle { return scrydex.Familie }

// System ist das Kennungssystem, unter dem Scrydex seine Karten führt.
func (s scrydexpreise) System() domain.Kennungssystem { return domain.SystemScrydex }

// Fassungen sucht die Karten, die zu einem Druck passen könnten.
//
// **Gesucht wird nach dem Namen, gefiltert nach Set und Nummer.** Scrydex'
// Abfrage ist über `name:"…"` dokumentiert; ein Filter auf den Setcode ist es
// **nicht**, und einen zu erfinden wäre genau der Fehler, den §15 verbietet.
// Also wird der dokumentierte Weg gegangen und das Übrige hier eingegrenzt.
func (s scrydexpreise) Fassungen(ctx context.Context, druck domain.Druck) (
	[]domain.Fremdfassung, error,
) {
	karten, fehler := s.klient.Karten(ctx, druck.Name)
	if fehler != nil {
		return nil, fehler
	}

	var passende []domain.Fremdfassung

	for _, karte := range karten {
		if !strings.EqualFold(karte.Setcode, druck.Setcode) {
			continue
		}

		// **Die Nummer grenzt weiter ein, wo beide eine haben.** Wo eine
		// fehlt, wird nicht weggefiltert: eine fehlende Angabe ist keine
		// Abweichung, und wer sie so behandelte, fände nie etwas.
		if druck.Nummer != "" && karte.Nummer != "" && karte.Nummer != druck.Nummer {
			continue
		}

		passende = append(passende, domain.Fremdfassung{
			Kennung: karte.Kennung, Name: karte.Name,
		})
	}

	return passende, nil
}

// Verlauf gibt die Preisstände einer Fassung.
func (s scrydexpreise) Verlauf(
	ctx context.Context, fassung, alias string, wunsch domain.Angebotswunsch,
) ([]domain.Preisstand, []domain.Befund, error) {
	punkte, befunde, fehler := s.klient.Preisverlauf(ctx, fassung,
		scrydex.Verlaufswunsch{
			Ausfuehrung: string(wunsch.Finish),
			Zustand:     einzigerZustand(wunsch.Zustaende),
		})
	if fehler != nil {
		return nil, nil, fehler
	}

	var staende []domain.Preisstand

	for _, punkt := range punkte {
		for _, stand := range punkt.AlsStaende(alias) {
			// **Der Wunsch wirkt auch auf die Antwort**, und das ist kein
			// doppelter Weg: Scrydex nimmt nur **einen** Zustand in der
			// Abfrage, und wer zwei nennt, bekäme sonst alle (§16, Drift).
			if passtZumWunsch(stand, wunsch) {
				staende = append(staende, stand)
			}
		}
	}

	return staende, befunde, nil
}

// einzigerZustand gibt den Zustand, wenn es genau einen gibt.
//
// **Scrydex nimmt einen und nicht mehrere** (nachgelesen am 2026-09-11):
// `condition` ist ein Wert. Wer zwei wünscht, bekommt sie nicht über die
// Abfrage — sie werden hier ausgesiebt.
func einzigerZustand(zustaende []string) string {
	if len(zustaende) != 1 {
		return ""
	}

	return zustaende[0]
}

// passtZumWunsch siebt einen Stand gegen den Wunsch.
func passtZumWunsch(stand domain.Preisstand, wunsch domain.Angebotswunsch) bool {
	if wunsch.Finish != "" && wunsch.Finish != stand.Finish {
		return false
	}

	if len(wunsch.Zustaende) == 0 {
		return true
	}

	for _, zustand := range wunsch.Zustaende {
		if strings.EqualFold(zustand, stand.Zustand) {
			return true
		}
	}

	return false
}

// scrydexpreiseUeber baut die Quelle — oder sagt, was ihr fehlt.
func scrydexpreiseUeber(ctx context.Context, wahl Dienstwahl, bund Geheimnisleser) (
	scrydexpreise, error,
) {
	zugang, fehler := scrydexzugang(ctx, bund)
	if fehler != nil {
		return scrydexpreise{}, fehler
	}

	if fehlt := zugang.Fehlt(); fehlt != "" {
		return scrydexpreise{}, domain.ZugangFehlt(domain.DienstScrydex, fehlt)
	}

	return scrydexpreise{klient: scrydex.NeuerKlient(scrydexadresse,
		dienstklient(wahl, domain.DienstScrydex), zugang).
		MitDrossel(drosselwerk.Fuer(scrydexadresse, scrydex.Abstand))}, nil
}
