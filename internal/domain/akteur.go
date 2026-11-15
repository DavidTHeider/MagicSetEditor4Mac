// **Der wirksame Akteur** — Werkstück O1 von Phase 5.2
// (#1425),
// nach [Vermerk 0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md).
//
// # Wozu
//
// Bis hierher wurde an genau zwei Stellen gefragt, wer etwas darf: beim
// Wechsel in ein fremdes Profil und beim Vergeben einer Freigabe (0070). Alles
// andere hielt der Treiber — `query_only` für den Leser (0073). Das reicht,
// solange jede Wirkung eine Zeile im Bestand ist. **Es reicht nicht mehr**,
// sobald eine Wirkung **draußen** liegt: eine Datei auf der Platte, ein Auftrag
// an einen fremden Dienst, ein Angebot an einen Partner, ein Kauf. Nichts
// davon hält ein `PRAGMA`.
//
// # Fünf Achsen, eine Schnittmenge
//
// Ein Akteur ist, was ein Aufruf über sich weiß: **wer** arbeitet (der
// Nutzer, 0074), **wo** (das Profil und sein Eigner), **als was** (die Rolle
// an diesem Profil, 0070), **womit** (ein Token mit seinem Umfang, 0068 —
// oder keines) und **worüber** (die Fläche). Was er darf, ist die
// **Schnittmenge**: die engste Achse gewinnt, und ein Token erweitert nie,
// was die Rolle nicht hergibt.
//
// # Portierungsmarke: **new** — und hier ist gesucht worden
//
// MSE2 kennt keine Nutzer (0070, 0073 mit derselben Suche: zwei Treffer auf
// `role`/`permission`/`readonly` in `/Users/david/mse2-quelle/src`, beide
// unbeteiligt). Es gibt kein Altverhalten, das zu portieren wäre.

package domain

import "fmt"

// Wirkung ist die Klasse dessen, was ein Aufruf bewirkt — von innen nach
// außen. **Die Reihenfolge ist die Aussage**: jede Klasse schließt die davor
// ein, und wer eine spätere darf, darf die früheren.
type Wirkung string

const (
	// WirkungLesen: der Bestand wird gelesen, sonst nichts.
	WirkungLesen Wirkung = "lokalLesen"

	// WirkungSchreiben: der Bestand oder die Platte dieses Rechners ändert
	// sich — eine Karte, ein Set, eine Datei, ein Bild.
	WirkungSchreiben Wirkung = "lokalSchreiben"

	// WirkungExternEmpfangen: dieser Rechner wird von außen **erreichbar** —
	// ein Port geht auf, eine Einladung wird ausgegeben, eine Ankündigung
	// geht ins Netz. Werkstück N7 von Phase 5.4
	// (#1568).
	//
	// # Warum es eine eigene Klasse braucht
	//
	// **Weil keine der fünf sie traf.** [WirkungSchreiben] meint den eigenen
	// Bestand, [WirkungExternSenden] einen Ruf an einen **bekannten**
	// Empfänger. Ein offener Port ist beides nicht: nichts wird geschrieben,
	// nichts geht hinaus — und trotzdem kann danach jemand von sich aus
	// anklopfen. `docs/AUTORISIERUNG.md` sagte bei „Für andere erreichbar
	// sein" ausdrücklich, dass die Klasse weniger sagt, als geschieht.
	//
	// # Warum sie **vor** dem Senden steht
	//
	// **Weil diese Ordnung die Reichweite der Daten misst.** Ein
	// Kartentext, der zu einem Sprachmodell geht, ist fort; ein Port, der
	// aufgeht, bewegt kein Byte. Was er schafft, ist eine **Gelegenheit** —
	// und die hält ein zweites Tor: ohne bestätigten Kontakt und ohne
	// freigegebenen Zweck kommt niemand hindurch (N1, #1454).
	//
	// **Was die Stelle in der Reihe damit nicht sagt**: dass ein offener
	// Port harmlos wäre. Sie sagt, dass er nichts preisgibt. Wer das Risiko
	// begrenzen will, begrenzt es am zweiten Tor und nicht an dieser Zahl.
	//
	// **Und die Rollengrenze bleibt, wo sie war**: [rollenreichweite] nennt
	// Klassen und keine Zahlen, und ein Mitarbeiter reicht weiterhin bis
	// [WirkungExternSenden] — also über diese hier hinaus. Das Issue
	// verlangt es: „Kein Verhalten ändert sich dadurch."
	WirkungExternEmpfangen Wirkung = "externEmpfangen"

	// WirkungExternSenden: Daten verlassen den Rechner, ohne dort etwas zu
	// ändern — eine Anfrage an Scryfall, ein Kartentext an ein Sprachmodell.
	WirkungExternSenden Wirkung = "externSenden"

	// WirkungExternSchreiben: bei einem fremden Dienst ändert sich etwas —
	// ein Listing, eine Bestellung, eine Nachricht an einen Partner.
	WirkungExternSchreiben Wirkung = "externSchreiben"

	// WirkungVerbindlich: ein Geschäft wird geschlossen — Geld, ein Tausch,
	// eine Zusage, die man nicht zurücknimmt.
	WirkungVerbindlich Wirkung = "verbindlich"
)

// reichweite ist die Ordnung der Wirkungen.
//
// **Berechnet und nicht abgeschrieben** (N7, #1568): eine zweite Aufzählung
// neben [Wirkungen] liefe beim nächsten Einschub auseinander — und genau der
// hat gerade stattgefunden.
var reichweite = ordnungDerWirkungen()

// ordnungDerWirkungen nummeriert sie in ihrer Reihenfolge, ab eins.
//
// **Ab eins und nicht ab null**, weil [Akteur.Darf] die Null als „gibt es
// nicht" liest.
func ordnungDerWirkungen() map[Wirkung]int {
	ordnung := make(map[Wirkung]int, len(Wirkungen()))
	for stelle, wirkung := range Wirkungen() {
		ordnung[wirkung] = stelle + 1
	}

	return ordnung
}

// Wirkungen gibt alle sechs, von innen nach außen.
func Wirkungen() []Wirkung {
	return []Wirkung{
		WirkungLesen, WirkungSchreiben, WirkungExternEmpfangen,
		WirkungExternSenden, WirkungExternSchreiben, WirkungVerbindlich,
	}
}

// Reicht sagt, ob diese Wirkung mindestens die verlangte umfasst.
func (w Wirkung) Reicht(verlangt Wirkung) bool {
	return reichweite[w] >= reichweite[verlangt] && reichweite[verlangt] > 0
}

// BrauchtNachweis sagt, ob eine Wirkung einen [Nachweis] verlangt — über die
// Rolle hinaus.
//
// **Was draußen etwas ändert oder verbindlich ist, braucht immer einen.**
// Eine Rolle sagt, was jemand grundsätzlich darf; ein Nachweis sagt, dass er
// **dieses eine** wollte — mit Zweck, Ziel und Stand. Ein Feld `bestaetigt`,
// das der Aufrufer selbst setzt, wäre keiner: es beweist nur, dass er das
// Feld kennt.
func (w Wirkung) BrauchtNachweis() bool {
	return reichweite[w] >= reichweite[WirkungExternSchreiben]
}

// Dienstmodus ist die Fläche, über die ein Aufruf kommt.
type Dienstmodus string

const (
	ModusFenster       Dienstmodus = "fenster"
	ModusKommandozeile Dienstmodus = "kommandozeile"
	ModusMCP           Dienstmodus = "mcp"
	ModusDienst        Dienstmodus = "dienst"
)

// Dienstmodi gibt alle vier.
func Dienstmodi() []Dienstmodus {
	return []Dienstmodus{ModusFenster, ModusKommandozeile, ModusMCP, ModusDienst}
}

// Akteur ist, was ein Aufruf über sich weiß.
type Akteur struct {
	// NutzerID ist, wer arbeitet (0074) — nicht der Eigner des Profils.
	NutzerID int64

	// ProfilID ist das gezeigte Profil, EignerID sein Eigner.
	ProfilID int64
	EignerID int64

	// Rolle ist die wirksame Rolle des Nutzers an diesem Profil; der Eigner
	// ist Verwalter ohne Zeile im Bestand (0070).
	Rolle Rolle

	// NurLesend sagt, dass die Verbindung `query_only` ist (0073).
	NurLesend bool

	// Token ist der Umfang des Zugriffstokens, über das der Aufruf kam — leer
	// am Fenster, an der Kommandozeile und mit der Prozessmarke.
	Token Scope

	Modus Dienstmodus

	// Revision zählt die Stände von Zeiger und Freigaben: wer eine ältere
	// trägt, hat einen Akteur gelesen, der so nicht mehr gilt.
	Revision int64
}

// rollenreichweite sagt, wie weit eine Rolle ohne Nachweis reicht.
//
// **Der Mitarbeiter schickt hinaus, der Verwalter ändert draußen** — und
// beides braucht ab [WirkungExternSchreiben] zusätzlich den Nachweis.
var rollenreichweite = map[Rolle]Wirkung{
	RolleLeser:       WirkungLesen,
	RolleMitarbeiter: WirkungExternSenden,
	RolleVerwalter:   WirkungVerbindlich,
}

// tokenreichweite sagt, wie weit ein Token reicht: **nie über den Bestand
// hinaus.** Ein Token ist ein Nachweis für eine Maschine, und was draußen
// wirkt, entscheidet ein Mensch am Fenster (0068).
var tokenreichweite = map[Scope]Wirkung{
	ScopeLesen:     WirkungLesen,
	ScopeSchreiben: WirkungSchreiben,
}

// Darf sagt, ob dieser Akteur eine Wirkung auslösen darf — ohne Nachweis.
//
// **Die Absage nennt die Achse**, die begrenzt hat: wer bis hierher kommt,
// ist kein Fremder, und ohne diesen Satz suchte er den Fehler bei sich.
func (a Akteur) Darf(wirkung Wirkung) error {
	if reichweite[wirkung] == 0 {
		return fmt.Errorf("%w: die Wirkung %q gibt es nicht", ErrErlaubnis, string(wirkung))
	}

	if a.NurLesend && wirkung != WirkungLesen {
		return fmt.Errorf("%w: in diesem Profil wird nur gelesen", ErrNurLesend)
	}

	rolle, bekannt := rollenreichweite[a.Rolle]
	if !bekannt {
		return fmt.Errorf("%w: ohne Rolle an diesem Profil geht nichts", ErrErlaubnis)
	}

	if !rolle.Reicht(wirkung) {
		return fmt.Errorf("%w: die Rolle %q reicht nicht bis %q",
			ErrErlaubnis, string(a.Rolle), string(wirkung))
	}

	if a.Token != "" {
		if grenze, gilt := tokenreichweite[a.Token]; !gilt || !grenze.Reicht(wirkung) {
			return fmt.Errorf("%w: das Token reicht nicht bis %q — ein Token "+
				"erweitert nie, was die Rolle hergibt, und wirkt nie nach draußen",
				ErrErlaubnis, string(wirkung))
		}
	}

	return nil
}
