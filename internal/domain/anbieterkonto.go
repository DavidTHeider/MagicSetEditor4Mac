// **Ein Anbieterkonto** — Werkstück 4 von Phase 4.12
// (#1351),
// nach Vermerk [0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md).
//
// # Was hier steht und was nicht
//
// **Hier steht, wer jemand bei einem Anbieter ist** — der Anbieter und seine
// Kennung für diesen Menschen. **Kein Zugangswort, kein Auffrischungswort,
// kein Geheimnis**: was zu hüten wäre, liegt im Schlüsselbund, und der Grund
// ist derselbe wie bei [Nutzer] — die Bestandsdatei ist die, die man
// weitergibt.
//
// **Und kein Wort davon, wie man dorthin kommt.** OAuth, PKCE, `state` und
// `nonce` sind Sache von `internal/anmeldung`; eine Domäne, die den Fluss
// kennte, wäre an ihn gebunden.

package domain

import (
	"fmt"
	"slices"
	"strings"
)

// Anbieterkonto verbindet einen Nutzer mit seiner Kennung bei einem Anbieter.
type Anbieterkonto struct {
	ID       int64
	NutzerID int64

	// Anbieter ist einer aus [Anbieter] — kleingeschrieben und ohne Zierrat.
	Anbieter string

	// Subjekt ist die Kennung, die der Anbieter für diesen Menschen führt.
	//
	// **Und nicht seine Adresse.** Wer bei GitHub die Mailadresse tauscht, ist
	// derselbe Mensch; eine Adresse als Schlüssel machte aus ihm einen
	// zweiten, und der erste hätte alle Daten. Bei OIDC ist es `sub`, bei
	// GitHub und X die numerische Kennung des Kontos.
	Subjekt string

	// Anzeigename ist, wie der Anbieter diesen Menschen nennt — leer erlaubt.
	//
	// **Nur zum Anzeigen.** Er ändert sich, wenn jemand ihn ändert, und nichts
	// hängt daran.
	Anzeigename string
}

// Fremdkennung ist, was ein Anbieter über einen Menschen hergibt.
//
// # Warum sie hier steht und nicht im Anmeldepaket
//
// **Weil sie ein fachlicher Begriff ist und kein Detail des Flusses.** Sie
// entsteht am Ende einer Anmeldung und wird von `internal/service` gedeutet:
// kennen wir diesen Menschen schon? Stünde sie in `internal/anmeldung`, müsste
// `service` den Fluss kennen, um seine **Antwort** zu lesen — und die
// Schichtenregel sagt zu Recht Nein (§2).
//
// **Sie ist ein [Anbieterkonto] ohne Nutzer**, und das ist Absicht: wer sie
// hat, weiß noch nicht, wem sie gehört. Ein `Anbieterkonto` mit `NutzerID: 0`
// wäre ein Typ, der über seine eigene Gestalt lügt.
type Fremdkennung struct {
	Anbieter    string
	Subjekt     string
	Anzeigename string
}

// anbieter ist die geschlossene Liste.
//
// **Geschlossen und kleingeschrieben.** Wer „Github" schreibt, meint denselben
// Anbieter — zwei Zeilen dafür wären zwei Menschen. Eine offene Liste hätte
// dasselbe Problem und dazu keinen Ort, an dem jemand entscheidet.
//
// **Apple fehlt mit Grund** (Vermerk 0066): der Weg verlangt einen
// Developer-Account, eine Services ID und eine App im App Store. Dieses
// Projekt hat keinen Account und soll keinen bekommen
// ([0003](../../docs/entscheidungen/0003-keine-notarisierung.md)). **X steht an
// seiner Stelle**, auf Anweisung des Auftraggebers vom 2026-09-04.
//
// **OpenAI fehlt anders** — nicht abgelehnt, sondern *nicht belegt*: ein Weg,
// auf dem sich eine fremde Desktop-Anwendung als Klient registriert, war am
// 2026-09-04 nicht zu finden. Der Unterschied gehört stehen zu bleiben.
var anbieter = []string{"github", "google", "microsoft", "x"}

// Anbieter gibt die Liste.
func Anbieter() []string { return slices.Clone(anbieter) }

// AnbieterKennung gibt den Einstellungsschlüssel für die `client_id` eines
// Anbieters.
//
// # Warum das hier steht und nicht in `internal/app`
//
// **Weil `PruefeEinstellung` ihn kennen muss.** Er stand in
// `app.AnbieterKennungsschluessel`, die Domäne kannte ihn nicht — und wies ihn
// deshalb ab. Die Anmeldung **las** eine Einstellung, die sich nicht
// **setzen** ließ: vier Anbieter, ein fertiger OAuth-Lauf, und kein Weg
// hinein. Gefunden vom Auftraggeber am 2026-09-06 an der laufenden Anwendung.
//
// **Ein Schlüssel gehört dorthin, wo über ihn geurteilt wird** (§10: die
// Zusage gehört an die Stelle, an der sie eingelöst wird).
func AnbieterKennung(anbieter string) string {
	return "anbieter." + anbieter + ".kennung"
}

// KenntAnbieter sagt, ob dieser Name einer ist.
func KenntAnbieter(name string) bool { return slices.Contains(anbieter, name) }

// Validate prüft ein Konto.
func (k Anbieterkonto) Validate() error {
	if k.NutzerID == 0 {
		return fmt.Errorf(
			"%w: ein Anbieterkonto gehört einem Nutzer", ErrValidation)
	}

	if strings.TrimSpace(k.Anbieter) == "" {
		return fmt.Errorf("%w: ein Anbieterkonto nennt seinen Anbieter", ErrValidation)
	}

	if !KenntAnbieter(k.Anbieter) {
		return fmt.Errorf("%w: %q ist keiner der Anbieter %v",
			ErrValidation, k.Anbieter, anbieter)
	}

	if strings.TrimSpace(k.Subjekt) == "" {
		return fmt.Errorf(
			"%w: ein Anbieterkonto nennt die Kennung, die der Anbieter führt",
			ErrValidation)
	}

	return nil
}
