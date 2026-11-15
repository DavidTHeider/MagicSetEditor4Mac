package rest

// **Alles rund um den Stil eines Sets** — herausgelöst aus `naehte.go`,
// weil die Datei sonst über die Zeilengrenze aus §2 ginge (Werkstück T5 von
// Phase 5.6,
// #1470).
//
// **Der Schnitt folgt der Frage und nicht der Zahl**: hier steht, was der
// Dienst über den **Stil** weiß — seit DOM-02 (#1697) auch die Naht der
// Stilwahl, die bis dahin am Ende von `naehte.go` stand; dort die übrigen
// Nähte.

import "context"

// Stil ist ein wählbarer Stil.
type Stil struct {
	Paket        string `json:"paket"`
	Spiel        string `json:"spiel"`
	Version      string `json:"version"`
	EigenesSpiel bool   `json:"eigenesSpiel"`
	Gewaehlt     bool   `json:"gewaehlt"`
}

// Wechselfolgen sagt, was ein Wechsel kostet.
type Wechselfolgen struct {
	UngedeuteteFelder []string `json:"ungedeuteteFelder"`
	VerloreneOptionen []string `json:"verloreneOptionen"`
	Spielwechsel      bool     `json:"spielwechsel"`
}

// Stiloption ist eine Option des gewählten Stils samt Wert.
//
// **`Gesetzt` und `Vorgabe` sind zweierlei**: jenes sagt, ob der Wert aus dem
// Bestand kommt, dieses, was `default:` rechnet.
type Stiloption struct {
	Name         string `json:"name"`
	Art          string `json:"art"`
	Beschreibung string `json:"beschreibung"`
	Wert         string `json:"wert"`
	Gesetzt      bool   `json:"gesetzt"`
	HatVorgabe   bool   `json:"hatVorgabe"`
	Vorgabe      string `json:"vorgabe"`
}

// Stilwahl ist, was die sechs Endpunkte der Stilwahl brauchen.
//
// **Der Wechsel bleibt zweistufig** — erst `Folgen`, dann `Wechseln`. Ein
// Wechsel kann Feldwerte ungedeutet lassen und Stiloptionen verlieren; im
// Fenster steht dazwischen eine Rückfrage, und eine Fläche ohne Rückfrage muss
// es wenigstens **fragen** können (§16, Regel 6).
type Stilwahl interface {
	Stile(ctx context.Context, setID int64) ([]Stil, error)
	Folgen(ctx context.Context, setID int64, paket string) (Wechselfolgen, error)
	Wechseln(ctx context.Context, setID int64, paket string) error
	Optionen(ctx context.Context, setID int64) ([]Stiloption, error)
	OptionSetzen(ctx context.Context, setID int64, feld, wert string) error
	OptionZuruecksetzen(ctx context.Context, setID int64, feld string) error
}
