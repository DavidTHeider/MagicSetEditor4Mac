// **Lokale Entsperrung über den zweiten Faktor** — ID-02
// (#1701).
//
// **Eine eigene Datei neben `faktorpruefung.go`**, weil sie eine andere Frage
// beantwortet: dort, ob jemand sich **ausweist**, hier, ob eine **gesperrte**
// Sitzung wieder aufgeht. Die Wege sind dieselben — Sensor und Zettel —, die
// Regeln nicht: zum Entsperren zählt kein Einmalkennwort, weil es kein
// Kennwort ersetzt, und der Sensor heißt hier **lokale Entsperrung**, nicht
// Passkey (AK-02).

package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/kennwort"
)

// grundEntsperren steht im Dialog von Touch ID.
const grundEntsperren = "um diese Anwendung zu entsperren"

// SensorMoeglich sagt, ob dieser Nutzer auf diesem Rechner über den Sensor
// entsperren kann — **ohne ihn zu fragen**.
func (z *Zweiterfaktor) SensorMoeglich(ctx context.Context, nutzerID int64) (bool, error) {
	faktoren, fehler := z.ablage.Faktoren(ctx, nutzerID)
	if fehler != nil {
		return false, fehler
	}

	return hat(faktoren, domain.FaktorBiometrie) && z.sensor.Vorhanden(ctx), nil
}

// SensorPruefen fragt den Sensor, um eine gesperrte Sitzung zu öffnen.
func (z *Zweiterfaktor) SensorPruefen(ctx context.Context, nutzerID int64) error {
	moeglich, fehler := z.SensorMoeglich(ctx, nutzerID)
	if fehler != nil {
		return fehler
	}

	if !moeglich {
		return domain.NewFieldError("sensor", domain.SchluesselLokaleEntsperrungFehlt,
			"Lokale Entsperrung ist für diesen Nutzer auf diesem Rechner nicht eingerichtet.")
	}

	return z.sensor.Bestaetige(ctx, grundEntsperren)
}

// CodeFinden sucht den Code vom Zettel, zu dem dieser Text gehört — **und
// verbraucht ihn nicht**.
//
// **Kein Einmalkennwort**: wer das Kennwort vergessen hat, beweist mit der
// App auf dem Telefon, dass er das Telefon hat, nicht, dass er das Konto
// angelegt hat. Der Zettel ist beim Einrichten ausgegeben worden und damit
// das einzige **bereits autorisierte** lokale Verfahren (AK-03).
//
// **Verbraucht wird er im selben Geschäft wie das neue Kennwort**
// (Zweitblick über ID-02, Befund 5): wer ihn hier verbrauchte, verlöre ihn,
// sobald das Setzen danach scheiterte.
//
// **Im Budget der Fehlversuche** wie jeder Codeweg (SEC-10, [imBudget]).
func (z *Zweiterfaktor) CodeFinden(ctx context.Context, nutzerID int64, code string) (int64, error) {
	return imBudget(ctx, z, nutzerID, func() (int64, error) {
		return z.codeFinden(ctx, nutzerID, code)
	})
}

// codeFinden sucht ohne Budget — für [Zweiterfaktor.ueberDenZettel], das
// schon darin läuft.
func (z *Zweiterfaktor) codeFinden(ctx context.Context, nutzerID int64, code string) (int64, error) {
	frei, fehler := z.ablage.FreieCodes(ctx, nutzerID)
	if fehler != nil {
		return 0, fehler
	}

	for _, einer := range frei {
		if kennwort.Stimmt(einer.Gehuetet, code) == nil {
			return einer.ID, nil
		}
	}

	return 0, domain.NewFieldError("faktor", domain.SchluesselCodeFalsch, "Dieser Code stimmt nicht.")
}
