package app

// **Die Menüleiste als eigenes Modell** — ENG-02
// (#1727).
//
// # Warum ein eigenes Modell und nicht das der Hülle
//
// **Weil die Hülle wechselt und die Leiste nicht.** Bis Wails 3 baute
// [Menueleiste] unmittelbar Wails' `menu.Menu`, und jeder Lauf darüber las
// dessen Felder — `Items`, `Role`, `Accelerator`, `Click`. Wails 3 gibt einen
// Eintrag nicht mehr so her: sein Rückruf ist nicht auszulösen, ohne dass ein
// Fenster läuft (`MenuItem.handleClick` ist nicht exportiert). Mit einem
// eigenen Modell bleibt die Leiste prüfbar, wie sie es war, und die Hülle
// setzt sie an **einer** Stelle um.
//
// **Und `internal/app` kennt Wails dann gar nicht mehr.** Wails 3 lässt sich
// unter Linux nur mit cgo bauen; eine Anwendungsschicht, die es importierte,
// fiele aus dem Kreuzbau der Pipeline heraus (§12).

// Menue ist eine Menüleiste oder ein Untermenü.
type Menue struct {
	Eintraege []*Menueeintrag
}

// Menueeintrag ist ein Befehl, ein Untermenü, eine Trennlinie oder eine Rolle
// — genau eines davon.
type Menueeintrag struct {
	// Beschriftung ist, was im Menü steht; bei einer Rolle leer.
	Beschriftung string

	// Kuerzel ist die Tastenkombination — `nil`, wenn es keine gibt.
	Kuerzel *Kuerzel

	// Handlung ist, was ein Befehl tut. **Nie `nil` bei einem Befehl**: eine
	// fehlende Handlung wird beim Bauen zu einer, die nichts tut ([rueckruf]).
	Handlung func()

	// Untermenue ist das Menü unter diesem Eintrag.
	Untermenue *Menue

	// Trenner ist eine Trennlinie.
	Trenner bool

	// Gegraut sagt, dass der Eintrag dasteht und nichts tut — UXR-F14
	// (#2029).
	//
	// **Die Handlung bleibt dran.** Gegraut wird, wo es nichts gibt, das der
	// Eintrag umlegen könnte; ob jemand ihn trotzdem auslöst, entscheidet die
	// Hülle (macOS löst einen grauen Eintrag nicht aus), und eine Handlung, die
	// fehlte, wäre der Absturz beim Klick, den [rueckruf] verhindert.
	Gegraut bool

	// Rolle ist ein Eintrag, dessen Handlung macOS stellt — an der
	// Responder-Kette oder an der Anwendung selbst. **Die Beschriftung stellt
	// diese Anwendung** (#255): unter Wails 2 kamen ganze Menüs als Rolle und
	// trugen englische Beschriftungen; Wails 3 lässt einen Eintrag mit Rolle
	// beschriften, ohne dass er seinen Selektor verliert.
	//
	// **An einem Untermenü** heißt sie: macOS führt dieses Menü mit — das
	// Fenstermenü die offenen Fenster.
	Rolle Menuerolle
}

// Menuerolle ist ein Eintrag, dessen Handlung die Plattform stellt.
type Menuerolle int

// Die Rollen, die diese Leiste führt — die Einträge, die macOS in
// Anwendungs-, Bearbeiten- und Fenstermenü erwartet.
const (
	// KeineMenuerolle heißt: ein gewöhnlicher Eintrag.
	KeineMenuerolle Menuerolle = iota

	// Das Anwendungsmenü.
	RolleDienste
	RolleAusblenden
	RolleAndereAusblenden
	RolleAlleEinblenden
	RolleBeenden

	// Das Bearbeitenmenü — an der Responder-Kette: ein Textfeld nimmt zurück,
	// was getippt wurde, und weiß von Karten nichts.
	RolleWiderrufen
	RolleWiederholen
	RolleAusschneiden
	RolleKopieren
	RolleEinsetzen
	RolleLoeschen
	RolleAllesAuswaehlen
	RolleSprechenStarten
	RolleSprechenStoppen

	// Das Fenstermenü.
	RolleMinimieren
	RolleZoomen
	RolleAlleNachVorne

	// RolleFenstermenue steht an einem **Untermenü**: dort führt macOS die
	// offenen Fenster.
	RolleFenstermenue
)

// Menuerollen gibt jede Rolle eines Eintrags — ohne [KeineMenuerolle] und
// ohne die Rolle eines Untermenüs.
//
// **Eine Aufzählung für die Läufe und für die Hülle**: jede Rolle muss eine
// Beschriftung bekommen und in Wails eine Entsprechung haben, und eine neue
// Rolle soll beides erzwingen, statt still ohne eines dazustehen.
func Menuerollen() []Menuerolle {
	rollen := make([]Menuerolle, 0, RolleFenstermenue-1)
	for rolle := RolleDienste; rolle < RolleFenstermenue; rolle++ {
		rollen = append(rollen, rolle)
	}

	return rollen
}

// Kuerzel ist eine Tastenkombination **mit ⌘**.
//
// **Immer mit ⌘**, weil jede der Leiste es so hat — und eine Angabe, die es
// abschaltbar machte, wäre ein Feld, das niemand setzt.
type Kuerzel struct {
	// Taste ist ein Zeichen, klein geschrieben — `s`, `,`.
	Taste string

	// Umschalt nimmt ⇧ dazu, Wahl ⌥.
	Umschalt bool
	Wahl     bool

	// OhneBefehl lässt ⌘ weg — für ⇞ und ⇟ im Menü „Karte“ (AP-7, #2364,
	// Entwurf `Kontext09`). Jedes andere Kürzel trägt ⌘.
	OhneBefehl bool
}

// Befehlstaste gibt ⌘ und eine Taste.
func Befehlstaste(taste string) *Kuerzel { return &Kuerzel{Taste: taste} }

// Untermenue hängt ein Untermenü an und gibt es zurück.
func (m *Menue) Untermenue(beschriftung string) *Menue {
	unter := &Menue{}
	m.Eintraege = append(m.Eintraege, &Menueeintrag{Beschriftung: beschriftung, Untermenue: unter})

	return unter
}

// Befehl hängt einen Eintrag an, der etwas tut.
func (m *Menue) Befehl(beschriftung string, kuerzel *Kuerzel, handlung func()) {
	m.Eintraege = append(m.Eintraege,
		&Menueeintrag{Beschriftung: beschriftung, Kuerzel: kuerzel, Handlung: handlung})
}

// BefehlWenn hängt einen Eintrag an, der nur gilt, solange `aktiv` ist —
// sonst steht er gegraut da (UXR-F14, #2029).
func (m *Menue) BefehlWenn(aktiv bool, beschriftung string, kuerzel *Kuerzel, handlung func()) {
	m.Eintraege = append(m.Eintraege, &Menueeintrag{
		Beschriftung: beschriftung, Kuerzel: kuerzel, Handlung: handlung, Gegraut: !aktiv,
	})
}

// Trennlinie hängt eine Trennlinie an.
func (m *Menue) Trennlinie() {
	m.Eintraege = append(m.Eintraege, &Menueeintrag{Trenner: true})
}

// Rolleneintrag hängt einen Eintrag an, dessen Handlung die Plattform stellt
// — beschriftet von dieser Anwendung (#255).
func (m *Menue) Rolleneintrag(rolle Menuerolle, beschriftung string) {
	m.Eintraege = append(m.Eintraege, &Menueeintrag{Beschriftung: beschriftung, Rolle: rolle})
}

// Rollenuntermenue hängt ein Untermenü an, das die Plattform mitführt, und
// gibt es zurück.
func (m *Menue) Rollenuntermenue(rolle Menuerolle, beschriftung string) *Menue {
	unter := &Menue{}
	m.Eintraege = append(m.Eintraege,
		&Menueeintrag{Beschriftung: beschriftung, Untermenue: unter, Rolle: rolle})

	return unter
}
