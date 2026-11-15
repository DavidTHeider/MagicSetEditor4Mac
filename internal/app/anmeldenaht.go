// **Die Anmeldung über einen Anbieter, verdrahtet** — Werkstück 4 von Phase
// 4.12 (#1351).
//
// # Was hier steht
//
// **Die Verdrahtung und sonst nichts** (§2). Wo die Kennung liegt, wo das
// Geheimnis liegt, welcher Klient benutzt wird, wer den Browser aufmacht — das
// sind Fragen dieser Schicht. **Wer ein Mensch ist**, entscheidet
// `internal/service`; **wie ein OAuth-Fluss geht**, weiß `internal/anmeldung`.

package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anmeldung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Anmeldedienst ist die Naht zu `internal/service`.
type Anmeldedienst interface {
	Herein(ctx context.Context, kennung domain.Fremdkennung) (int64, error)
}

// Anbieterzugang ist, was ein Anbieter an Zugangsdaten braucht.
//
// **Die Kennung ist kein Geheimnis, das Geheimnis schon.** Jene steht in der
// Anmeldeadresse und geht durch jeden Browser; dieses liegt im Schlüsselbund
// ([0056](../../docs/entscheidungen/0056-drei-weitere-kartendienste.md)).
type Anbieterzugang struct {
	Kennung   string
	Geheimnis string
}

// Zugangsquelle gibt die Zugangsdaten eines Anbieters.
type Zugangsquelle func(ctx context.Context, anbieter string) (Anbieterzugang, error)

// Browseroeffner macht eine Adresse im Browser auf.
//
// **Als Naht und nicht als fester Aufruf**: ein Test, der wirklich einen
// Browser öffnete, risse den Rechner auf, und die Anmeldung ließe sich nur von
// Hand prüfen.
type Browseroeffner func(ctx context.Context, adresse string) error

// Anmeldenaht führt den ganzen Fluss.
type Anmeldenaht struct {
	dienst   Anmeldedienst
	zugaenge Zugangsquelle
	oeffne   Browseroeffner
	rufer    *anmeldung.Rufer

	// bau gibt je Anbieter seinen Bauplan.
	//
	// **Als Feld und nicht als feste Zuordnung**, aus demselben Grund wie
	// `oeffnen` in `anwendung.go`: die Endpunkte zeigen auf github.com und
	// accounts.google.com, und ein Lauf, der den **ganzen** Fluss prüfen will,
	// bräuchte sonst das Netz — und einen Menschen, der im Browser klickt.
	//
	// **Was der Wechsel nicht verschiebt**: dass die Adressen stimmen, hält
	// `internal/anmeldung`; hier läuft der Weg **durch** sie.
	bau map[string]func(kennung, geheimnis string) anmeldung.Anbieter

	// horche öffnet den Rückruf-Horcher.
	//
	// **Als Feld und aus demselben Grund wie `lauschen` in
	// `internal/anmeldung`**: ein Horchen, das scheitert, kommt auf einem
	// gesunden Rechner nicht zustande — im Betrieb aber schon, wenn alle Ports
	// belegt sind. Ein Fehlerpfad, den kein Test betritt, ist eine Behauptung
	// darüber, was im Ernstfall geschieht (§3).
	horche func() (*anmeldung.Horcher, error)

	// wartefrist ist, wie lange der Fluss nach dem Öffnen des Browsers auf
	// den Rückruf wartet — BK-07d (#2118). **Zehn Minuten**: genug für ein
	// Anmelden samt zweitem Faktor beim Anbieter, und danach steht der Port
	// der Schleifenadresse nicht länger offen als der Versuch lebt.
	wartefrist time.Duration
}

// Anmeldefrist ist die Vorgabe für [Anmeldenaht.wartefrist].
const Anmeldefrist = 10 * time.Minute

// MitWartefrist setzt, wie lange auf den Rückruf gewartet wird — die Naht
// für Läufe, die nicht zehn Minuten warten wollen.
func (n *Anmeldenaht) MitWartefrist(frist time.Duration) *Anmeldenaht {
	n.wartefrist = frist

	return n
}

// MitAnbietern tauscht die Baupläne — für Läufe gegen eine Gegenstelle.
func (n *Anmeldenaht) MitAnbietern(
	bau map[string]func(kennung, geheimnis string) anmeldung.Anbieter,
) *Anmeldenaht {
	n.bau = bau

	return n
}

// NeueAnmeldenaht baut sie.
func NeueAnmeldenaht(
	dienst Anmeldedienst, zugaenge Zugangsquelle,
	oeffne Browseroeffner, rufer *anmeldung.Rufer,
) *Anmeldenaht {
	return &Anmeldenaht{
		dienst: dienst, zugaenge: zugaenge, oeffne: oeffne, rufer: rufer,
		bau: anbieterbau, horche: anmeldung.Horche, wartefrist: Anmeldefrist,
	}
}

// anbieterbau gibt für jeden Namen seinen Bauplan.
//
// **Eine Zuordnung und keine Verzweigung im Fluss** (§16): die vier
// unterscheiden sich in Adressen und zwei Schaltern, nicht im Weg.
var anbieterbau = map[string]func(kennung, geheimnis string) anmeldung.Anbieter{
	"github":    anmeldung.GitHub,
	"google":    anmeldung.Google,
	"microsoft": anmeldung.Microsoft,
	"x":         anmeldung.X,
}

// Anmelden führt den ganzen Fluss und gibt den Nutzer.
//
// # Die Reihenfolge ist die Zusage
//
// Horcher auf, Adresse bauen, Browser auf, warten, prüfen, tauschen, deuten.
// **Der Horcher steht vor der Adresse**, weil die Rückadresse seinen Port
// trägt; **die Prüfung steht vor dem Tausch**, weil ein Rückruf mit fremdem
// `state` gar nicht erst eingelöst werden darf.
func (n *Anmeldenaht) Anmelden(ctx context.Context, anbieter string) (int64, error) {
	bau, kennt := n.bau[anbieter]
	if !kennt {
		return 0, fmt.Errorf("%w: %q ist keiner der Anbieter %v",
			domain.ErrValidation, anbieter, domain.Anbieter())
	}

	zugang, zugangsfehler := n.zugaenge(ctx, anbieter)
	if zugangsfehler != nil {
		return 0, zugangsfehler
	}

	horcher, horchfehler := n.horche()
	if horchfehler != nil {
		return 0, horchfehler
	}

	defer func() { _ = horcher.Schliesse() }() //nolint:errcheck // der Fehler steht unten

	lauf, laufsfehler := anmeldung.Beginne(
		bau(zugang.Kennung, zugang.Geheimnis), horcher.Adresse)
	if laufsfehler != nil {
		return 0, laufsfehler
	}

	// **Erst den `state` nennen, dann den Browser öffnen** (SEC-18): sonst
	// käme ein schneller Rückruf vor der Erwartung an und zählte nicht.
	horcher.Erwarte(lauf.State)

	if oeffnungsfehler := n.oeffne(ctx, lauf.Adresse); oeffnungsfehler != nil {
		return 0, oeffnungsfehler
	}

	// **Mit eigener Frist** (BK-07d, #2118): ohne sie wartete der Horcher, bis
	// der Ruf selbst abbrach. Ein Ablauf ist kein Abbruch — er bekommt seinen
	// eigenen Fehler und damit seinen eigenen Rat.
	wartekontext, aufhoeren := context.WithTimeout(ctx, n.wartefrist)
	defer aufhoeren()

	werte, wartefehler := horcher.Warte(wartekontext)
	if wartefehler != nil {
		if ctx.Err() == nil && errors.Is(wartefehler, context.DeadlineExceeded) {
			return 0, fmt.Errorf("%w (nach %v)", domain.ErrAnmeldungAbgelaufen, n.wartefrist)
		}

		return 0, wartefehler
	}

	if pruefung := lauf.PruefeRueckruf(werte); pruefung != nil {
		return 0, pruefung
	}

	kennung, einloesefehler := lauf.Einloesen(ctx, n.rufer, werte.Get("code"))
	if einloesefehler != nil {
		return 0, einloesefehler
	}

	return n.dienst.Herein(ctx, kennung)
}
