package mcp

// **Die Zeile einer Partienliste** — ausgelagert aus `partie.go`, als die
// Felder für den eigenen Sitz dazukamen (UXR-F16, #2031) und die Datei über die
// Zeilengrenze schoben (§2). Der Schnitt folgt der Frage: hier steht, wie eine
// Liste eine Partie zeigt, dort, wie ein Sitz sie sieht.

// Partiezeilenangaben ist eine Partie, wie eine Liste sie zeigt.
//
// **Ohne Objekte** — eine Liste ist keine Sicht, und was in einer Partie
// liegt, rechnet die Projektion je Sitz.
type Partiezeilenangaben struct {
	ID      string `json:"id"`
	Fassung string `json:"fassung"`

	Spiel string `json:"spiel"`
	Stand string `json:"stand"`

	Sitze []Zeilensitzangaben `json:"sitze"`

	// Objekte ist, **wie viele** Dinge in dieser Partie liegen.
	Objekte int `json:"objekte"`

	// Zug ist der wievielte Zug — null heißt: es geht nicht zugweise zu.
	Zug int `json:"zug"`

	// EigenerSitz ist der Sitz, der an das Profil dieses Zugangs gebunden
	// ist — **leer heißt: keiner**, das Profil sieht hier zu (UXR-F16,
	// #2031).
	//
	// **Erweiternd** im Sinne von
	// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md).
	EigenerSitz string `json:"eigenerSitz"`

	// Dran sagt, ob die Partie gerade auf den eigenen Sitz wartet — wegen
	// Priorität oder einer offenen Entscheidung. Gerechnet hat es `service`;
	// hier steht dasselbe wie am Fenster, an der Kommandozeile und im Dienst.
	Dran bool `json:"dran"`
}

// Zeilensitzangaben ist ein Platz, wie eine Liste ihn zeigt.
type Zeilensitzangaben struct {
	Kennung string `json:"kennung"`
	Name    string `json:"name"`
}
