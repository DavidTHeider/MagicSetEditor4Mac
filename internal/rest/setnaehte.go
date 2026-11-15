// Die Nähte für **Einstellungen und Setangaben** — die Gestalten, in denen
// der Dienst sie nennt.
//
// **Eine eigene Datei**, seit `naehte.go` mit den Parameterhilfen über die
// Zeilengrenze aus §2 ging
// (#1342).
// Der Schnitt folgt der Frage: dort stehen die Nähte **zum Bestand** — Sets,
// Karten, Sammlung —, hier die **um** ein Set herum: was es an Angaben trägt
// und wie die Anwendung eingestellt ist.

package rest

import "context"

// Einstellungen ist, was die zwei Einstellungsendpunkte brauchen.
// **`EinstellungSetzen` und nicht `Setzen`**: [Setangaben] hat schon ein
// `Setzen`, und dieselbe Testvorrichtung erfüllt beide Schnittstellen. Zwei
// gleichnamige Methoden mit verschiedenen Signaturen kann ein Typ nicht haben —
// und ein Double, das sie trennen müsste, wäre zwei Doubles.
type Einstellungen interface {
	Alle(ctx context.Context) ([]Einstellung, error)
	EinstellungSetzen(ctx context.Context, schluessel, wert string) error
}

// Einstellung ist eine Einstellung samt ihren erlaubten Werten.
type Einstellung struct {
	Schluessel string   `json:"schluessel"`
	Wert       string   `json:"wert"`
	Werte      []string `json:"werte"`
	Bereich    string   `json:"bereich"`

	// **Was gilt, wo es hingehört und von wem es kommt** — Werkstück O3 von
	// Phase 5.2
	// (#1427).
	// Dieselben drei Felder wie am Fenster und über MCP: eine Fläche, die sie
	// nicht trüge, gäbe eine andere Antwort auf dieselbe Frage (§16).
	Wirksam  string `json:"wirksam"`
	Ebene    string `json:"ebene"`
	Herkunft string `json:"herkunft"`
	// Wirkung sagt, **wann** ein geänderter Wert gilt — `sofort` oder
	// `neustart` (UX-04.AK-02,
	// #1676).
	// Dieselbe Antwort wie am Fenster: sie hängt daran, wie der Wert
	// **gelesen** wird, nicht daran, wo jemand ihn umlegt (§16).
	Wirkung string `json:"wirkung"`

	// Vorgabe ist der Wert, mit dem die Anwendung anfängt (#1422).
	Vorgabe string `json:"vorgabe"`
}

// Setangaben ist, was die Endpunkte für Setangaben und Vorbehalte brauchen.
type Setangaben interface {
	Angaben(ctx context.Context, setID int64) (
		felder []Feldbeschreibung, werte []Setwert, fehler error)
	Setzen(ctx context.Context, setID int64, feld, wert string) (
		vorher string, hatte bool, fehler error)
	Vorbehalte(ctx context.Context, setID int64) ([]Vorbehalt, error)
	Beispielsets() ([]Beispielset, error)
}

// Feldbeschreibung ist ein Angabenfeld, wie der Dienst es nennt.
//
// **Eine eigene Gestalt und nicht `app.FeldbeschreibungDTO`**: `rest` ist ein
// Blatt und kennt `app` nicht (§2). Die Wiederholung ist der Preis dafür, und
// sie ist der billigere — dieselbe Begründung wie bei [Ausfuhr].
type Feldbeschreibung struct {
	Name       string `json:"name"`
	Aufschrift string `json:"aufschrift"`
	Art        string `json:"art"`
}

// Setwert ist eine gesetzte Angabe.
type Setwert struct {
	Feld string `json:"feld"`
	Wert string `json:"wert"`
}

// Vorbehalt ist ein Block, den das Einfahren nicht gedeutet hat.
type Vorbehalt struct {
	Datei  string `json:"datei"`
	Zeile  int    `json:"zeile"`
	Inhalt string `json:"inhalt"`
}

// Beispielset ist ein mitgeliefertes Set.
type Beispielset struct {
	Name string `json:"name"`
	Pfad string `json:"pfad"`
}
