// **Zugriffstoken** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkt 13 der Kommentierung vom 2026-09-05.
//
// # Was ein Token hier ist
//
// **Ein Schlüssel zu einem Profil, mit einem Scope und einer Frist.** Wer den
// Dienst auf der Rückschleife automatisch bedienen lässt, braucht etwas, das
// er einem Skript geben kann — und das er wieder wegnehmen kann, ohne sein
// Kennwort zu ändern.
//
// # Warum nur zwei Scopes
//
// **Weil es zwei Fragen gibt und nicht hundert.** Ein Werkzeug liest, oder es
// schreibt; `internal/mcp` trennt das seit jeher (`Lesend`/`Schreibend`), und
// über HTTP sagt es die Methode. Ein feineres Modell wäre eine Aufzählung, die
// bei **jedem** neuen Endpunkt gepflegt sein will — und ein ungepflegter
// Eintrag in einer Rechteliste ist gefährlicher als keiner: er sieht aus wie
// eine Regel und ist eine Lücke.
//
// **Rechte und Rollen über Nutzer hinweg sind etwas anderes** und stehen als
// Werkstück 11 desselben Issues: nicht „was darf dieses Token", sondern „was
// darf dieser Mensch bei den Sachen eines anderen".

package domain

import (
	"slices"
	"strings"
	"time"
)

// Scope sagt, was ein Token darf.
type Scope string

// Die zwei Scopes.
//
// **Deutsch benannt**, weil es eigene Begriffe dieses Programms sind
// ([0015](../../docs/entscheidungen/0015-bezeichner-folgen-dem-vokabular.md))
// — dieselben Wörter, die `internal/mcp` seinen Werkzeugen voranstellt.
const (
	ScopeLesen     Scope = "lesen"
	ScopeSchreiben Scope = "schreiben"
)

// alleScopes ist die Aufzählung — die Reihenfolge ist die Steigerung.
var alleScopes = []Scope{ScopeLesen, ScopeSchreiben}

// Scopes gibt die erlaubten Scopes, in Anzeigereihenfolge.
//
// **Eine Kopie**: ein Aufrufer, der die Liste änderte, änderte die Regel.
func Scopes() []Scope { return slices.Clone(alleScopes) }

// Reicht sagt, ob dieser Scope für den verlangten genügt.
//
// **Schreiben schließt Lesen ein.** Wer schreiben darf, darf lesen — alles
// andere wäre ein Token, das eine Karte ändern kann, ohne sie vorher zu
// sehen.
func (s Scope) Reicht(verlangt Scope) bool {
	if s == ScopeSchreiben {
		return true
	}

	return s == verlangt
}

// Zugriffstoken ist ein Schlüssel zu einem Profil.
type Zugriffstoken struct {
	ID       int64
	ProfilID int64

	// Bezeichnung sagt, wofür es da ist — „Sicherungsskript", „Editor".
	//
	// **Ohne sie ist eine Liste von Token unbrauchbar**: sie sehen einander
	// gleich, und wer eines abschalten will, weiß nicht welches.
	Bezeichnung string

	// Principal sagt, **für wen** dieses Token arbeitet — der Mensch, der es
	// eingerichtet hat, oder die Maschine, die damit läuft.
	//
	// **Warum ein Name und kein Fremdschlüssel auf einen Nutzer**: der
	// Verantwortliche eines Maschinenzugangs ist meist gar kein Nutzer dieser
	// Anwendung — ein Baurechner, ein `cron`-Eintrag, ein MCP-Gegenüber. Ein
	// Verweis auf die Nutzertafel zwänge einen von beiden in eine Gestalt,
	// die er nicht hat, und ein Verweis, der leer bleiben darf, ist wieder
	// keine Angabe (Wanderung 0048).
	//
	// **Er ist Pflicht**, und das ist der ganze Punkt: [Zugriffstoken.Bezeichnung]
	// sagt, **wofür** ein Token da ist, dieses Feld, **wer** dahintersteht.
	// Ein Feld, das leer bleiben darf, bleibt leer.
	Principal string

	Scope Scope

	// Ablauf ist RFC 3339 — leer heißt: läuft nicht ab.
	//
	// **Als Zeichenkette und nicht als [time.Time]**: dieselbe Entscheidung
	// wie bei `Exemplar.Erwerbsdatum`. Ein Zeitpunkt ohne Zone wäre beim
	// nächsten Lesen woanders, und die Null von `time.Time` sähe aus wie ein
	// Datum im Jahr 1.
	Ablauf string

	// Angelegt ist der Zeitpunkt der Anlage, ebenfalls RFC 3339.
	Angelegt string

	// LetzteNutzung ist der Zeitpunkt, an dem dieses Token zuletzt
	// angenommen wurde — leer heißt: noch nie.
	//
	// **Damit ein Zugang, den niemand mehr braucht, sichtbar wird.** Eine
	// Liste von Token sagt sonst nur, dass es sie gibt; sie sagt nicht,
	// welches man abschalten kann, ohne etwas kaputtzumachen.
	LetzteNutzung string

	// VorgaengerBis ist die Frist, bis zu der das **vorige** Geheimnis nach
	// einer Rotation noch gilt — leer heißt: es gibt keines.
	//
	// **Das Geheimnis selbst steht hier so wenig wie das geltende** (siehe
	// [Zugriffstoken.Validate]); hier steht nur, wie lange es noch zählt.
	VorgaengerBis string

	// Sitzbindung bindet das Token an **einen** Sitz einer Partie (#1819) —
	// leer heißt: wie jedes andere Token. **Sie entsteht mit dem Token** und
	// ändert sich nicht; wer einen anderen Sitz braucht, legt ein neues an.
	Sitzbindung Sitzbindung

	// Fassung ist die Bedingung einer Änderung (§7, Regel 2).
	Fassung int64
}

// Validate prüft ein Token — **nicht** sein Geheimnis.
//
// **Das Geheimnis steht hier gar nicht.** Es wird gehütet, sobald es entsteht,
// und die Domäne sieht nur, was danach übrig bleibt: wem es gehört, wofür es
// da ist und wie lange.
func (z Zugriffstoken) Validate() error {
	if z.ProfilID == 0 {
		return NewFieldError("profilID", SchluesselPflichtfeld,
			"Ein Token gehört zu einem Profil.")
	}

	if err := pruefeAnzeigename("bezeichnung", z.Bezeichnung,
		"Ein Token braucht eine Bezeichnung — sonst sind zwei nicht zu "+
			"unterscheiden."); err != nil {
		return err
	}

	if err := pruefeAnzeigename("principal", z.Principal,
		"Ein Token nennt, für wen es arbeitet — ein Mensch oder eine "+
			"Maschine."); err != nil {
		return err
	}

	if !slices.Contains(alleScopes, z.Scope) {
		return NewFieldError("scope", SchluesselValidierung,
			"%q ist kein Scope; erlaubt sind »lesen« und »schreiben«.", z.Scope)
	}

	if err := z.Sitzbindung.Validate(); err != nil {
		return err
	}

	return z.pruefeAblauf()
}

// pruefeAblauf weist einen Ablauf ab, der kein Zeitpunkt ist.
func (z Zugriffstoken) pruefeAblauf() error {
	if strings.TrimSpace(z.Ablauf) == "" {
		return nil
	}

	if _, err := time.Parse(time.RFC3339, z.Ablauf); err != nil {
		return NewFieldError("ablauf", SchluesselValidierung,
			"Ein Ablauf ist ein Zeitpunkt in RFC 3339; %q geht nicht.", z.Ablauf)
	}

	return nil
}

// Abgelaufen sagt, ob dieses Token zum genannten Zeitpunkt vorbei ist.
//
// # Warum ein unlesbarer Ablauf als abgelaufen gilt
//
// **Weil die andere Lesart ein Schlüssel wäre, der nie stumpf wird.** Ein
// Wert, den niemand deuten kann, steht nur da, wenn ihn jemand von Hand in
// den Bestand geschrieben hat — und dann ist „gilt für immer" die schlechtere
// von zwei Auslegungen. Die Prüfung beim Schreiben lässt ihn ohnehin nicht
// durch.
//
// **Die Uhr kommt herein** (§16): ein Lauf mit echter Wartezeit flackert.
func (z Zugriffstoken) Abgelaufen(jetzt time.Time) bool {
	if strings.TrimSpace(z.Ablauf) == "" {
		return false
	}

	bis, err := time.Parse(time.RFC3339, z.Ablauf)
	if err != nil {
		return true
	}

	return !jetzt.Before(bis)
}
