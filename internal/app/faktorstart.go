// **Was der zweite Faktor an Plattform braucht** — Werkstück 7 von Phase 4.12
// (#1351).
//
// # Zwei Nähte, die nur hier zusammenkommen
//
// **Wo ein TOTP-Geheimnis liegt** und **wer den Sensor fragt** sind Fragen der
// Plattform, nicht der Fachlichkeit. `internal/service` kennt sie als
// [service.Geheimnisse] und [service.Sensor]; **hier** stehen die Antworten —
// der Schlüsselbund von macOS und `LocalAuthentication` (§2).

package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/biometrie"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Faktordienst ist der Name, unter dem die Geheimnisse im Bund liegen.
//
// **Ein eigener neben [Anbieterdienst]**, damit ein `security`-Aufruf, der
// alle Einträge dieses Dienstes räumt, nicht beide Sorten mitnimmt.
const Faktordienst = "MagicSetEditor4Mac-Faktor"

// faktorgeheimnisse legt TOTP-Geheimnisse im Schlüsselbund ab.
//
// **Nicht im Bestand**, aus demselben Grund wie die Zugangsdaten fremder
// Dienste: die Bestandsdatei ist die, die man weitergibt, und wer das
// Geheimnis hat, macht jeden Code.
//
// **Was das kostet, steht daneben**: ein Bestand auf einem neuen Mac findet
// sein Geheimnis nicht — und genau dafür gibt es die Wiederherstellungscodes
// (`service.Zweiterfaktor`).
type faktorgeheimnisse struct {
	bund Geheimnisspeicher

	// kennung liefert die Kennung **dieses** Bestands.
	//
	// **Ohne sie kollidieren zwei Bestände auf demselben Rechner**: jeder hat
	// einen Nutzer 1, und der zweite überschriebe das Geheimnis des ersten.
	// Gefunden am 2026-09-05 durch den Flächenlauf, nicht überlegt (§10).
	//
	// # Warum eine Funktion und keine Zeichenkette
	//
	// **Weil sie beim Verdrahten noch nicht dasteht.** `App.starten` ruft
	// `verdrahten(db)` **vor** `aufDenHeutigenStand` — die Nähte entstehen an
	// einer Datenbank, die noch nicht migriert ist. Ein Bestand, der zum
	// ersten Mal geöffnet wird, hat die Tafel `bestandskennung` (Wanderung
	// 0039) zu diesem Zeitpunkt nicht, und das Lesen scheitert.
	kennung func() (string, error)
}

// Setze schreibt das Geheimnis eines Nutzers.
func (f faktorgeheimnisse) Setze(
	ctx context.Context, nutzerID int64, geheimnis string,
) error {
	konto, fehler := f.konto(nutzerID)
	if fehler != nil {
		return fehler
	}

	return f.bund.Schreibe(ctx, Faktordienst, konto, geheimnis)
}

// Hole liest es zurück.
func (f faktorgeheimnisse) Hole(ctx context.Context, nutzerID int64) (string, error) {
	konto, fehler := f.konto(nutzerID)
	if fehler != nil {
		return "", fehler
	}

	return f.bund.Lies(ctx, Faktordienst, konto)
}

// Nimm löscht es.
func (f faktorgeheimnisse) Nimm(ctx context.Context, nutzerID int64) error {
	konto, fehler := f.konto(nutzerID)
	if fehler != nil {
		return fehler
	}

	return f.bund.Loesche(ctx, Faktordienst, konto)
}

// konto ist der Schlüssel im Bund — Bestand und Nutzer.
//
// **Beides und nicht nur der Nutzer**: jeder Bestand auf diesem Rechner hat
// einen Nutzer 1.
//
// **Und Kennungen, keine Namen**: die ändern sich, und ein Geheimnis, das
// danach unter einem anderen Schlüssel läge, wäre verloren.
func (f faktorgeheimnisse) konto(nutzerID int64) (string, error) {
	bestand, fehler := f.kennung()
	if fehler != nil {
		return "", fmt.Errorf("die Bestandskennung für den Schlüsselbund: %w",
			fehler)
	}

	// **Ein Punkt als Trenner und kein Doppelpunkt**: der Schlüsselbund
	// nimmt nur Buchstaben, Ziffern, Punkt, Strich und Unterstrich
	// (`schluesselbund`, nachgemessen dort).
	return bestand + "." + strconv.FormatInt(nutzerID, 10), nil
}

// Die Nähte, die dieser Verbund erfüllt — geprüft beim Übersetzen.
var _ service.Geheimnisse = faktorgeheimnisse{}

// MitSensor setzt die Biometrie — für Läufe und den Composition Root.
//
// **Ein Sensor und eine Auskunft in einem Ruf**, weil beide dieselbe Frage
// von zwei Seiten stellen: der Dienst **prüft**, die Seite **fragt nach**. Wer
// sie getrennt setzte, könnte einen Rechner bauen, der prüft, ohne es zu
// können.
func (a *App) MitSensor(
	geber service.Sensor, auskunft func(context.Context) biometrie.Auskunft,
) {
	a.sensor = geber
	a.biometrie = auskunft
}

// sensorOderRechner gibt den gesetzten Sensor — oder den dieses Rechners.
//
// **Und setzt die Auskunft gleich mit**, wenn keine dasteht: sonst meldete die
// Seite „keine Biometrie" auf einem Mac, der sie hat.
func (a *App) sensorOderRechner() service.Sensor {
	if a.sensor == nil {
		a.sensor = rechnersensor
	}

	if a.biometrie == nil {
		a.biometrie = biometrie.Stand
	}

	return a.sensor
}
