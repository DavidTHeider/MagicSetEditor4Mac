package domain

// **Was dieses Programm wirklich durchsetzt** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Wofür es die Matrix gibt
//
// **Damit „streng" keine Zusage ist, die beim ersten Sonderfall bricht.** Das
// Issue sagt es zweimal: „Eine Liste von Namen ist kein Beleg für vollständige
// Regeldurchsetzung", und „Keine Vollständigkeitsbehauptung für sämtliche
// Magic-Karten". Wer Strenge anbietet, ohne zu sagen, **worüber**, verkauft
// eine Engine, die es nicht gibt.
//
// # Warum „unbekannt" ein eigener Zustand ist
//
// **Weil es der häufigste ist und der gefährlichste.** Das Issue verlangt ihn
// wörtlich: „Unbekannt/ununterstützt ist ein eigener Zustand, niemals still
// erlaubt oder gültig." Ein Wahrheitswert kennt nur ja und nein — und ein
// Bereich, den niemand geprüft hat, wäre dann entweder fälschlich zugesagt
// oder fälschlich verboten.
//
// **Die Vorgabe ist deshalb `unbekannt`**: ein Bereich, für den kein Eintrag
// dasteht, ist nicht geprüft. Eine leere Matrix trägt gar nichts — und sagt
// das.
//
// # Die dreizehn Bereiche
//
// **Sie stehen im Issue und sind in der CR nachgeschlagen** — Nummer und
// Fassung, wie §11 es verlangt. Abgerufen am 2026-09-15 aus
// `MagicCompRules 20260819.txt`; die Nummern stammen aus dem
// Inhaltsverzeichnis dieser Datei und nicht aus dem Gedächtnis.

import (
	"fmt"
	"slices"
)

// Regelbereich ist ein Feld der Unterstützungsmatrix.
type Regelbereich string

const (
	// BereichZonen — wo eine Karte liegt (CR 400).
	BereichZonen Regelbereich = "zonen"

	// BereichZugstruktur — Phasen und Abschnitte eines Zuges (CR 500).
	BereichZugstruktur Regelbereich = "zugstruktur"

	// BereichStapel — was wartet, bis es verrechnet wird (CR 405).
	BereichStapel Regelbereich = "stapel"

	// BereichPrioritaet — wer wann handeln darf (CR 117).
	BereichPrioritaet Regelbereich = "prioritaet"

	// BereichCombat — der Kampfabschnitt (CR 506).
	BereichCombat Regelbereich = "combat"

	// BereichZustandsbasiert — was von selbst geschieht (CR 704).
	BereichZustandsbasiert Regelbereich = "zustandsbasiert"

	// BereichSpielsteine — Tokens (CR 111).
	BereichSpielsteine Regelbereich = "spielsteine"

	// BereichMarken — Counters (CR 122).
	BereichMarken Regelbereich = "marken"

	// BereichAnhaengsel — was an etwas anderem hängt; Ausrüstung
	// (CR 301) und Auren (CR 303).
	BereichAnhaengsel Regelbereich = "anhaengsel"

	// BereichTypen — die Typzeile (CR 205).
	BereichTypen Regelbereich = "typen"

	// BereichMana — Vorrat, Manafähigkeiten und Bezahlen (CR 106, 118, 605).
	//
	// **Getrennt von `wechselwirkungen`**, obwohl beides beim Wirken
	// zusammenkommt: Mana entsteht und vergeht auch ohne einen einzigen
	// Zauber — ein Land tappt, ein Abschnitt endet —, und ein Bereich, der
	// erst mit dem Stapel etwas aussagte, wäre bis dahin leer.
	BereichMana Regelbereich = "mana"

	// BereichWechselwirkungen — das Verrechnen von Zaubern und Fähigkeiten
	// (CR 608) und das Kopieren (CR 707).
	BereichWechselwirkungen Regelbereich = "wechselwirkungen"

	// BereichCommander — die Variante aus CR 903 (M19,
	// #1620).
	//
	// **Ein eigener Bereich und keine Grenze anderswo**: CR 903 ändert
	// **mehrere** Bereiche zugleich — Startleben, Zonen, Stapelkosten und
	// eine zustandsbasierte Aktion. Wer sie dort einzeln unterbrächte,
	// hätte die Variante an vier Stellen und nirgends ganz.
	BereichCommander Regelbereich = "commander"
)

// Regelbereiche gibt alle zwölf, in der Reihenfolge des Issues.
func Regelbereiche() []Regelbereich {
	return []Regelbereich{
		BereichZonen, BereichZugstruktur, BereichStapel, BereichPrioritaet,
		BereichCombat, BereichZustandsbasiert, BereichSpielsteine,
		BereichMarken, BereichAnhaengsel, BereichTypen,
		BereichMana, BereichWechselwirkungen, BereichCommander,
	}
}

// IstRegelbereich sagt, ob ein Wert einer davon ist.
func IstRegelbereich(bereich Regelbereich) bool {
	return slices.Contains(Regelbereiche(), bereich)
}

// Unterstuetzung sagt, wie weit dieses Programm einen Bereich trägt.
type Unterstuetzung string

const (
	// UnterstuetzungUnbekannt heißt: **niemand hat es geprüft**.
	//
	// **Die Vorgabe**, und zwar mit Absicht: was nicht eingetragen ist, ist
	// nicht zugesagt. Ein Bereich in diesem Zustand verhindert Strenge und
	// steht in Sandkasten und Begleitet als benannte Lücke da.
	UnterstuetzungUnbekannt Unterstuetzung = "unbekannt"

	// UnterstuetzungManuell heißt: **die Menschen machen es selbst**.
	//
	// Das Programm hält nichts auf und warnt auch nicht — es weiß darüber
	// nichts. Der Unterschied zu `unbekannt` ist, dass jemand **entschieden**
	// hat, es nicht zu bauen.
	UnterstuetzungManuell Unterstuetzung = "manuell"

	// UnterstuetzungBegleitet heißt: **das Programm warnt**.
	UnterstuetzungBegleitet Unterstuetzung = "begleitet"

	// UnterstuetzungAutomatisch heißt: **das Programm setzt es durch**.
	//
	// Nur wer hier steht, trägt [StrengeStreng].
	UnterstuetzungAutomatisch Unterstuetzung = "automatisch"
)

// Unterstuetzungsgrade gibt alle vier, von schwach nach stark.
func Unterstuetzungsgrade() []Unterstuetzung {
	return []Unterstuetzung{
		UnterstuetzungUnbekannt, UnterstuetzungManuell,
		UnterstuetzungBegleitet, UnterstuetzungAutomatisch,
	}
}

// IstUnterstuetzung sagt, ob ein Wert einer der vier ist.
func IstUnterstuetzung(grad Unterstuetzung) bool {
	return slices.Contains(Unterstuetzungsgrade(), grad)
}

// Matrixeintrag ist die Aussage über **einen** Bereich.
type Matrixeintrag struct {
	Bereich Regelbereich
	Grad    Unterstuetzung

	// Quelle belegt die Aussage — Nummer und Fassung.
	//
	// **Pflicht ab `begleitet`**: wer warnt oder abweist, tut es nach einer
	// Regel, und die steht hier. Für `unbekannt` und `manuell` gibt es
	// nichts zu belegen — dort **verbietet** die Prüfung sie sogar, damit
	// eine Quelle nicht wie eine Zusage aussieht.
	Quelle Regelquelle

	// Grenzen nennt, was ausdrücklich **nicht** abgedeckt ist.
	//
	// **Das Issue verlangt sie**: „eigene positive/negative Grenzfälle". Ein
	// Bereich, der als `automatisch` dasteht und keine Grenze nennt,
	// behauptet Vollständigkeit — und die gibt es nicht.
	Grenzen []string

	// NurVariante sagt, dass dieser Bereich **einer Spielvariante** gehört
	// und nur dort zählt
	// ([0123](../../docs/entscheidungen/0123-streng-zaehlt-nur-was-zur-variante-gehoert.md),
	// #1626).
	//
	// **Wozu**: eine Sandkastenpartie ist keine Commanderpartie. Führte sie
	// `commander` als offene Lücke, sagte sie einem Menschen, hier fehle
	// Arbeit — und es fehlt keine. Eine Lücke, die niemand schließen soll,
	// ist keine.
	//
	// **Wie ein Profil sagt, dass es die Variante spielt**: es führt den
	// Bereich auf `begleitet` oder `automatisch`. Steht er auf `manuell`,
	// ist er nicht zuständig und zählt nicht mit.
	NurVariante bool
}

// Regelmatrix ist die Aussage über alle Bereiche.
type Regelmatrix struct {
	Eintraege []Matrixeintrag
}

// Grad gibt die Unterstützung eines Bereichs — `unbekannt`, wo keine steht.
func (m Regelmatrix) Grad(bereich Regelbereich) Unterstuetzung {
	for _, eintrag := range m.Eintraege {
		if eintrag.Bereich == bereich {
			return eintrag.Grad
		}
	}

	return UnterstuetzungUnbekannt
}

// Eintrag gibt den Eintrag eines Bereichs — und ob es einen gibt.
func (m Regelmatrix) Eintrag(bereich Regelbereich) (Matrixeintrag, bool) {
	for _, eintrag := range m.Eintraege {
		if eintrag.Bereich == bereich {
			return eintrag, true
		}
	}

	return Matrixeintrag{}, false
}

// Luecken gibt die Bereiche, die [StrengeStreng] nicht tragen.
//
// **In der Reihenfolge von [Regelbereiche]** und nicht in der der Einträge:
// eine Liste, deren Reihenfolge von der Eingabe abhängt, liest sich bei jedem
// Aufruf anders.
func (m Regelmatrix) Luecken() []Regelbereich {
	luecken := []Regelbereich{}

	for _, bereich := range Regelbereiche() {
		if m.Grad(bereich) == UnterstuetzungAutomatisch {
			continue
		}

		// **Ein Bereich einer Variante, die dieses Profil nicht spielt, ist
		// keine Lücke** ([0123](../../docs/entscheidungen/0123-streng-zaehlt-nur-was-zur-variante-gehoert.md)).
		// Er steht dann auf `manuell`, und das ist eine
		// Nichtzuständigkeit und kein Rückstand.
		if m.nurFremdeVariante(bereich) {
			continue
		}

		luecken = append(luecken, bereich)
	}

	return luecken
}

// nurFremdeVariante sagt, ob dieser Bereich einer Variante gehört, die
// dieses Profil nicht spielt.
func (m Regelmatrix) nurFremdeVariante(bereich Regelbereich) bool {
	for _, eintrag := range m.Eintraege {
		if eintrag.Bereich == bereich {
			return eintrag.NurVariante &&
				eintrag.Grad != UnterstuetzungBegleitet &&
				eintrag.Grad != UnterstuetzungAutomatisch
		}
	}

	return false
}

// TraegtStrenge sagt, ob die Matrix einen Strengegrad deckt.
//
// # Die Regel, und warum sie so und nicht anders lautet
//
//   - **Sandkasten** trägt immer — er prüft nichts (#1458).
//
//   - **Begleitet** trägt, sobald **ein** Bereich `begleitet` oder
//     `automatisch` ist — also etwas, worüber das Programm wirklich etwas
//     sagt.
//
//     **`manuell` genügt ausdrücklich nicht.** Der erste Anlauf ließ es
//     zählen, mit der Begründung, eine Entscheidung sei auch eine Auskunft.
//     Sie ist eine — nur keine, die **warnt**: `manuell` heißt, dass die
//     Menschen es selbst machen und das Programm schweigt. Eine Strenge
//     „begleitet", unter der nichts begleitet, wäre eine Zusage ohne
//     Einlösung.
//
//   - **Streng** trägt nur, wenn **jeder** Bereich `automatisch` ist. Das
//     Issue lässt keine Zwischenstufe zu: „Strict und Tournament nur
//     freischalten, wenn die gewählte Karten-/Regelkombination von einer
//     maschinenlesbaren Unterstützungsmatrix abgedeckt ist."
func (m Regelmatrix) TraegtStrenge(grad Strenge) bool {
	switch grad {
	case StrengeSandkasten:
		return true

	case StrengeBegleitet:
		for _, bereich := range Regelbereiche() {
			grad := m.Grad(bereich)
			if grad == UnterstuetzungBegleitet ||
				grad == UnterstuetzungAutomatisch {
				return true
			}
		}

		return false

	case StrengeStreng:
		return len(m.Luecken()) == 0
	}

	return false
}

// Pruefe sagt, ob die Matrix in sich stimmt.
func (m Regelmatrix) Pruefe() error {
	gesehen := map[Regelbereich]bool{}

	for _, eintrag := range m.Eintraege {
		if !IstRegelbereich(eintrag.Bereich) {
			return fmt.Errorf("%w: %q ist kein Regelbereich",
				ErrValidation, eintrag.Bereich)
		}

		if gesehen[eintrag.Bereich] {
			return fmt.Errorf("%w: der Bereich %q steht zweimal",
				ErrValidation, eintrag.Bereich)
		}

		gesehen[eintrag.Bereich] = true

		if fehler := eintrag.pruefe(); fehler != nil {
			return fehler
		}
	}

	return nil
}

// pruefe hält die Bedingungen eines einzelnen Eintrags.
func (e Matrixeintrag) pruefe() error {
	if !IstUnterstuetzung(e.Grad) {
		return fmt.Errorf("%w: %q ist keine Unterstützung", ErrValidation, e.Grad)
	}

	belegt := e.Grad == UnterstuetzungBegleitet ||
		e.Grad == UnterstuetzungAutomatisch

	if !belegt {
		// **Keine Quelle, wo nichts zugesagt ist.** Eine Regelnummer neben
		// `unbekannt` läse sich wie ein Beleg für etwas, das niemand gebaut
		// hat.
		if len(e.Quelle.Nummern) > 0 {
			return fmt.Errorf(
				"%w: der Bereich %q ist %q und nennt trotzdem eine Quelle — "+
					"eine Nummer neben einer Nichtzusage sieht aus wie ein Beleg",
				ErrValidation, e.Bereich, e.Grad)
		}

		return nil
	}

	if fehler := e.Quelle.Pruefe(); fehler != nil {
		return fmt.Errorf("der Bereich %q ist %q: %w", e.Bereich, e.Grad, fehler)
	}

	// **Wer durchsetzt, nennt seine Grenzen** (#1461).
	if e.Grad == UnterstuetzungAutomatisch && len(e.Grenzen) == 0 {
		return fmt.Errorf(
			"%w: der Bereich %q wird durchgesetzt und nennt keine Grenze — "+
				"das wäre eine Vollständigkeitsbehauptung",
			ErrValidation, e.Bereich)
	}

	return nil
}
