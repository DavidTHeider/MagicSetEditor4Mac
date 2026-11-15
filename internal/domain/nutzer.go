// **Nutzer und Profil** — Werkstück 1 von Phase 4.12
// (#1351),
// nach Vermerk [0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md).
//
// # Zwei Begriffe und warum es zwei sind
//
// **Ein Nutzer ist die Person, ein Profil eine Arbeitsumgebung dieser Person.**
// Wer für ein Turnier eine eigene Sammlung führen will und daneben seine
// eigene, braucht kein zweites Konto — er braucht ein zweites Profil.
//
// # Was hier **nicht** steht
//
// **Kein Kennwort, kein Geheimnis, kein Merkmal eines zweiten Faktors.** Was
// hier liegt, ist nichts, das man hüten müsste; Geheimnisse liegen im
// Schlüsselbund (`internal/schluesselbund`), und zwar aus demselben Grund wie
// die Zugangsdaten fremder Dienste: die Bestandsdatei ist die, die man
// weitergibt.
//
// **Und keine Trennung.** Dass ein Profil die Daten eines anderen nicht sieht,
// ist eine Frage an den Bestand und nicht an die Domäne — Werkstück 2.

package domain

import (
	"fmt"
	"strings"
	"time"
)

// Nutzer ist die Person.
type Nutzer struct {
	ID int64

	// Name ist der Anzeigename. Er ist das Einzige, was ein Nutzer haben muss.
	Name string

	// Email ist die Adresse, unter der dieser Nutzer sich anmeldet — leer
	// heißt: er tut es anders (Werkstück 6, #1351).
	Email string

	// Zweitemail ist die Adresse für die Sicherheitsabfrage — leer erlaubt.
	Zweitemail string

	// Telefon ist die Rufnummer — leer heißt: keine.
	Telefon string

	// Anonym sagt, ob dieser Nutzer ausdrücklich ohne Angaben angelegt wurde.
	//
	// **Ein Vermerk und kein leeres Feld.** Ein Nutzer ohne Adresse könnte
	// auch der erste sein; wer anonym angelegt wurde, soll es sagen können —
	// die Oberfläche zeigte ihn sonst wie einen, der seine Angaben vergessen
	// hat.
	Anonym bool

	// Kennwort ist die **gehütete** Form (`internal/kennwort`) — nie das
	// Kennwort selbst. Leer heißt: dieser Nutzer hat keines.
	//
	// **Sie steht in der Entität und wird nicht geprüft.** Was ein taugliches
	// Kennwort ist, entscheidet `internal/kennwort`; was hier liegt, ist das
	// Ergebnis, und ein Hash hat keine fachliche Form.
	Kennwort string

	// Die fünf Namensteile — Werkstück 8 (#1351), alle **freiwillig**.
	//
	// **Sie sind nicht der Anzeigename.** Der ist eindeutig und Pflicht;
	// diese beschreiben einen Menschen, und die Anwendung leitet den einen
	// nicht aus den anderen ab. Warum die Anrede freier Text ist, steht in
	// `kontoangaben.go`.
	Anrede       string
	Namenszusatz string
	Vorname      string
	Weiterenamen string
	Nachname     string

	// Die drei freiwilligen Texte — Punkte 14 bis 16 der Kommentierung.
	Slogan   string
	Kurzebio string
	Langebio string

	// Fassung zählt die Änderungen — Werkstück 8 (#1351).
	//
	// **Sie schließt die offene Stelle aus Werkstück 5**: wer gleichzeitig an
	// zwei Flächen einen Namen ändert, verlor bis dahin eine der beiden
	// Änderungen, und niemand merkte es (§7, Regel 2 über die Flächen
	// hinweg).
	//
	// **Sie beginnt bei 1**: eine Null sähe aus wie „nicht gesetzt", und ein
	// Aufrufer, der sie vergisst, träfe zufällig das Richtige.
	Fassung int64

	// Bild ist der Dateiname im Bildverzeichnis — leer heißt: keines.
	//
	// **Ein Name und kein Pfad.** Wohin er zeigt, entscheidet die Anwendung;
	// stünde hier ein Pfad, entschiede es der Bestand, und ein weitergegebener
	// Bestand zeigte auf Verzeichnisse eines fremden Rechners.
	Bild string
}

// Validate prüft einen Nutzer.
func (n Nutzer) Validate() error {
	// **Die Kennungen sind freiwillig, aber wenn sie dastehen, taugen sie.**
	// Ein leeres Feld heißt „hat er nicht" — ein krummes hieße „hat er, und
	// es ist falsch", und das erführe niemand (Werkstück 6, #1351).
	for feld, wert := range map[string]string{
		"email":      n.Email,
		"zweitemail": n.Zweitemail,
	} {
		if wert == "" {
			continue
		}

		if err := PruefeEmail(wert); err != nil {
			return fmt.Errorf("%w (%s)", err, feld)
		}
	}

	if n.Telefon != "" {
		if err := PruefeTelefon(n.Telefon); err != nil {
			return err
		}
	}

	if err := pruefeAnzeigename("name", n.Name, "Ein Nutzer braucht einen Namen."); err != nil {
		return err
	}

	// **Die freiwilligen Angaben** — Werkstück 8 (#1351). Sie stehen in
	// `kontoangaben.go`, damit diese Datei unter der Zeilengrenze aus §2
	// bleibt.
	if err := n.pruefeAngaben(); err != nil {
		return err
	}

	// **Ein Bild ist wahlfrei, aber kein Weg.** Dieselbe Frage wie beim
	// Paketnamen (#1350):
	// der Wert wird an ein Verzeichnis gehängt, und ein `../` darin läse eine
	// Datei außerhalb.
	//
	// **Seit Werkstück 9 über [pruefeDateiname]** — dasselbe gilt für das
	// Abzeichen eines Profils, und die Bedingung steht an **einer** Stelle.
	return pruefeDateiname("bild", n.Bild)
}

// Profil ist eine Arbeitsumgebung eines Nutzers.
type Profil struct {
	ID       int64
	NutzerID int64
	Name     string

	// Fassung zählt die Änderungen — dieselbe Frage wie bei [Nutzer.Fassung].
	Fassung int64

	// **Die freiwilligen Angaben** — Werkstück 9 (#1351), Punkte 10 bis 12.
	//
	// **`Subslogan` und nicht `Slogan`**: er steht **unter** dem des
	// Menschen und ersetzt ihn nicht.
	Subslogan         string
	Kurzebeschreibung string
	Langebeschreibung string

	// **Technisch heißt: für die maschinelle Verarbeitung** — Werkstück 10
	// (#1351), Punkt 12.
	//
	// **Eine Marke und keine eigene Entität.** Ein technisches Profil hat
	// einen Nutzer, einen Namen und eine Fassung wie jedes andere; was es
	// unterscheidet, ist, **wofür** es da ist. Eine zweite Tabelle daneben
	// hätte jede Abfrage über Profile verdoppelt.
	//
	// **Wozu es gut ist**: wer einen Dienst automatisch bedienen lässt, will
	// nicht, dass er in **seinem** Profil schreibt — was dort entsteht,
	// entsteht nicht von seiner Hand.
	Technisch bool

	// Sichtbarkeit ist `privat` oder `bestand` — Werkstück 11 (#1351).
	//
	// **Sie steht am Profil und nicht in einer Freigabe**, weil sie eine
	// Aussage über das Profil ist und nicht über einen Nutzer: „das darf jeder
	// sehen" braucht keine Zeile je Mensch.
	//
	// **Leer gilt als `privat`** — die vorsichtige Antwort.
	Sichtbarkeit string

	// Ablauf ist RFC 3339 — leer heißt: läuft nicht ab.
	//
	// **Nur an einem technischen Profil.** Ein Mensch hört nicht auf, dieser
	// Mensch zu sein; ein Datum an seiner Arbeitsumgebung wäre eine Zusage,
	// die niemand einlösen will.
	//
	// **Und es löscht nichts.** Die Kommentierung sagt „nach einer
	// konfigurierbaren Zeit **löschbar**" — nicht „gelöscht". Ein Bestand,
	// der hinter dem Rücken Zeilen entfernt, nimmt jemandem seine Daten,
	// während er nicht hinsieht.
	Ablauf string

	// Badge ist der Dateiname des Abzeichens — leer heißt: keines.
	//
	// **Es liegt über dem Nutzerfoto, rechts unten** (Punkt 13): dieselbe
	// Person in einer anderen Rolle. Zwei Bilder nebeneinander wären zwei
	// Menschen.
	Badge string
}

// Validate prüft ein Profil.
func (p Profil) Validate() error {
	if fehler := Sichtbarkeit(p.Sichtbarkeit).Pruefe(); fehler != nil {
		return fehler
	}

	if err := pruefeAnzeigename("name", p.Name, "Ein Profil braucht einen Namen."); err != nil {
		return err
	}

	if p.NutzerID == 0 {
		return NewFieldError("nutzerID", SchluesselPflichtfeld,
			"Ein Profil gehört zu einem Nutzer.")
	}

	return p.pruefeProfilangaben()
}

// Abgelaufen sagt, ob dieses Profil zum genannten Zeitpunkt vorbei ist.
//
// **Dieselbe Regel wie bei [Zugriffstoken.Abgelaufen]** — und aus demselben
// Grund an derselben Stelle: ein unlesbarer Ablauf gilt als abgelaufen, denn
// die andere Lesart wäre eine Frist, die nie greift.
func (p Profil) Abgelaufen(jetzt time.Time) bool {
	return Zugriffstoken{Ablauf: p.Ablauf}.Abgelaufen(jetzt)
}

// pruefeAnzeigename weist einen leeren Namen ab — und einen mit Rändern.
//
// # Warum Ränder abgewiesen und nicht beschnitten werden
//
// **Weil zwei Zeilen sonst denselben sichtbaren Namen tragen.** „David" und
// „David " stünden untereinander in der Profilwahl, und wer sie unterscheiden
// soll, sieht nichts. Stillschweigend zu beschneiden hieße außerdem, eine
// Eingabe zu ändern, ohne es zu sagen (§8) — der Aufrufer bekäme etwas
// anderes zurück, als er gab.
func pruefeAnzeigename(feld, wert, leer string) error {
	if strings.TrimSpace(wert) == "" {
		return NewFieldError(feld, SchluesselPflichtfeld, "%s", leer)
	}

	if wert != strings.TrimSpace(wert) {
		return NewFieldError(feld, SchluesselValidierung,
			"Ein Name trägt keine Leerzeichen am Rand; %q geht nicht.", wert)
	}

	return nil
}
