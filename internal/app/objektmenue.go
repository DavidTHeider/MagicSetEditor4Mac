package app

import (
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Handlungsliste des gewählten Objekts in der Menüleiste** — GESTE
// (#2358),
// Befund B-01; Artboard `Kontext09`.
//
// # Eine Liste, drei Orte
//
// Die Oberfläche baut **je Objekt eine** Handlungsliste — die Karte, den
// Handel —, und sie erscheint im Rechtsklick, unter ⇧F10, im Überlauf an der
// Zeile und hier, als Menü mit dem Namen des Objekts
// (`haken/useHandlungsliste.ts`). Go bekommt nur, was ein Menü braucht:
// Titel, die Einträge mit Namen und ob sie gesperrt sind, und die **Epoche**
// des Objekts.
//
// # Und es entscheidet nichts
//
// Ein Klick meldet [EreignisObjekthandlung] mit Objekt, Epoche und Schlüssel
// ins Fenster — derselbe Weg wie jeder andere Eintrag ([App.Fensterhandlungen]).
// **Ob die Epoche noch gilt, prüft das Fenster**: ein Klick, der nach einem
// Wechsel der Karte ankommt, bewirkt dort nichts (§7, Regel 2). Go führt
// keine zweite Wahrheit über das, was gewählt ist; es merkt sich die letzte
// Meldung, wie beim Darstellungsstand.
//
// # Ohne Kürzel
//
// **Ein Kürzel in der Leiste griffe auch in einem Textfeld**: ⌘⌫ löschte die
// Karte statt eines Wortes. Welche Kürzel die Karte bekommt, entscheidet AP-7
// (#2364) mit
// einem Lauf, der sie im Textfeld drückt.

// Objekthandlungen ist, was das Fenster über das gewählte Objekt meldet —
// leer, wenn keines gewählt ist.
type Objekthandlungen struct {
	Titel     string           `json:"titel"`
	Art       string           `json:"art"`
	Kennung   string           `json:"kennung"`
	Epoche    int              `json:"epoche"`
	Eintraege []Objekthandlung `json:"eintraege"`
}

// Objekthandlung ist ein Eintrag der Liste.
type Objekthandlung struct {
	Schluessel string `json:"schluessel"`
	Name       string `json:"name"`
	// Gesperrt steht grau da — für einen Leser oder während sein Vorgang
	// läuft. **Er verschwindet nicht**: was fehlt, kann niemand vermissen.
	Gesperrt bool `json:"gesperrt"`

	// Kuerzel nennt die Taste in der Leiste — leer, wenn es keine gibt, sonst
	// ein Wort aus [objektkuerzel] (AP-7, #2364). **Die Oberfläche sagt, welche
	// Handlung welches trägt**; Go setzt nur, was es kennt.
	Kuerzel string `json:"kuerzel,omitempty"`
}

// objektkuerzel sind die Kürzel, die ein Objektmenü tragen darf — die aus dem
// Entwurf `Kontext09`. **Eine feste Liste**: ein Wort, das Go nicht kennt,
// wäre eine Taste, die niemand geprüft hat.
var objektkuerzel = map[string]Kuerzel{
	"befehl+return":    {Taste: "return"},
	"befehl+d":         {Taste: "d"},
	"befehl+backspace": {Taste: "backspace"},
	"bild auf":         {Taste: "page up", OhneBefehl: true},
	"bild ab":          {Taste: "page down", OhneBefehl: true},
}

// Objektausloesung ist, was ein Klick im Objektmenü ins Fenster meldet.
type Objektausloesung struct {
	Art        string `json:"art"`
	Kennung    string `json:"kennung"`
	Epoche     int    `json:"epoche"`
	Schluessel string `json:"schluessel"`
}

// hoechstensObjekthandlungen ist die Grenze einer Liste. Die längste, die
// gebaut ist — „Diese Karte" —, hat elf Einträge; ein Menü mit mehr als
// vierzig liest niemand.
const hoechstensObjekthandlungen = 40

// ObjekthandlungenMelden nimmt die Handlungsliste des gewählten Objekts
// entgegen und baut die Leiste neu — **nur, wenn sich etwas geändert hat**:
// die Oberfläche meldet bei jedem Zeichnen, und eine neue Leiste kostet
// (#2312).
func (a *App) ObjekthandlungenMelden(meldung Objekthandlungen) SchreibAntwort {
	if fehler := objektmeldungPruefen(meldung); fehler != nil {
		return SchreibAntwort{Fehler: Nutzerfehler(fehler)}
	}

	a.menuesperre.Lock()
	defer a.menuesperre.Unlock()

	if gleicheObjektmeldung(a.objekthandlungen, meldung) {
		return SchreibAntwort{}
	}

	a.objekthandlungen = meldung
	a.objekthandlungen.Eintraege = slices.Clone(meldung.Eintraege)
	a.menueAufbauen()

	return SchreibAntwort{}
}

// objektmeldungPruefen weist ab, woraus kein Menü wird.
func objektmeldungPruefen(meldung Objekthandlungen) error {
	if len(meldung.Eintraege) == 0 {
		return nil
	}

	pruefung := func(regel bool, satz string, werte ...any) error {
		if regel {
			return nil
		}

		return domain.NewFieldError("eintraege", domain.SchluesselValidierung, satz, werte...)
	}

	gesehen := map[string]bool{}
	tasten := map[string]bool{}
	for _, eintrag := range meldung.Eintraege {
		if fehler := pruefung(eintrag.Schluessel != "" && !gesehen[eintrag.Schluessel],
			"Ein Eintrag hat keinen oder einen doppelten Schlüssel (%q).", eintrag.Schluessel); fehler != nil {
			return fehler
		}
		gesehen[eintrag.Schluessel] = true

		if eintrag.Kuerzel == "" {
			continue
		}

		_, bekannt := objektkuerzel[eintrag.Kuerzel]
		if fehler := pruefung(bekannt && !tasten[eintrag.Kuerzel],
			"Ein Kürzel ist unbekannt oder doppelt (%q).", eintrag.Kuerzel); fehler != nil {
			return fehler
		}
		tasten[eintrag.Kuerzel] = true
	}

	if fehler := pruefung(len(meldung.Eintraege) <= hoechstensObjekthandlungen,
		"Ein Menü mit %d Einträgen — höchstens %d.", len(meldung.Eintraege), hoechstensObjekthandlungen); fehler != nil {
		return fehler
	}

	return pruefung(strings.TrimSpace(meldung.Titel) != "", "Ein Menü braucht einen Titel.")
}

func gleicheObjektmeldung(vorher, jetzt Objekthandlungen) bool {
	return vorher.Titel == jetzt.Titel && vorher.Art == jetzt.Art && vorher.Kennung == jetzt.Kennung &&
		vorher.Epoche == jetzt.Epoche && slices.Equal(vorher.Eintraege, jetzt.Eintraege)
}

// objektmenueEinsetzen setzt das Menü des Objekts **vor das Fenstermenü** —
// dorthin, wo unter macOS die eigenen Menüs einer Anwendung stehen, gleich
// hinter *Set*. MSE2 setzt sein Kartenmenü ebenso hinter Bearbeiten ein
// (`mb->Insert(2, menuCard, …)`, `src/gui/set/cards_panel.cpp`, Zeile 208,
// Fassung f9d9356 vom 2021-07-06).
func objektmenueEinsetzen(leiste *Menue, meldung Objekthandlungen, melden func(Objektausloesung)) {
	if len(meldung.Eintraege) == 0 {
		return
	}

	unter := &Menue{}
	for _, eintrag := range meldung.Eintraege {
		ausloesung := Objektausloesung{
			Art: meldung.Art, Kennung: meldung.Kennung, Epoche: meldung.Epoche, Schluessel: eintrag.Schluessel,
		}
		befehl := &Menueeintrag{
			Beschriftung: eintrag.Name,
			Gegraut:      eintrag.Gesperrt,
			Handlung:     func() { melden(ausloesung) },
		}
		if kuerzel, gibt := objektkuerzel[eintrag.Kuerzel]; gibt {
			befehl.Kuerzel = &kuerzel
		}
		unter.Eintraege = append(unter.Eintraege, befehl)
	}

	stelle := slices.IndexFunc(leiste.Eintraege, func(eintrag *Menueeintrag) bool {
		return eintrag.Rolle == RolleFenstermenue
	})
	if stelle < 0 {
		stelle = len(leiste.Eintraege)
	}

	leiste.Eintraege = slices.Insert(leiste.Eintraege, stelle, &Menueeintrag{Beschriftung: meldung.Titel, Untermenue: unter})
}
