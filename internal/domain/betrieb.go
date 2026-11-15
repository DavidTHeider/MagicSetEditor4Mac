package domain

// **Der Betriebszustand: lebt, bereit, und was fremd ist** — Werkstück O9 von
// Phase 5.2 (#1433).
//
// # Die eine Regel, für die es diesen Typ gibt
//
// **Ein fremder Dienst macht diese Anwendung nicht unbereit.** Das ist die
// Abnahme des Issues und zugleich §16, Satz 5: ohne Netz arbeitet die
// Anwendung weiter, der letzte gute Stand bleibt, und er sagt, wie alt er ist.
// Wer beides in **eine** Zahl rechnete, machte aus einem fremden Ausfall einen
// eigenen — und ein Betriebswerkzeug, das darauf hört, würde einen Editor
// neu starten, weil Scryfall gerade nicht antwortet.
//
// **Deshalb sind es zwei Felder und nicht ein Zustand**: [Betriebszustand.Grund]
// ist der **eigene** Grund, [Betriebszustand.Fremde] eine Auskunft.
//
// # Warum der Grund eine Aufzählung ist
//
// **Weil er an vier Flächen ankommt.** Ein Satz wäre viermal ein anderer Satz;
// eine Aufzählung ist an jeder dieselbe Marke, und den Satz dazu sucht sich
// jede Fläche in ihrem Katalog (§9).

// Bereitschaftsgrund sagt, **warum** dieser Prozess nicht bereit ist.
//
// **Leer heißt bereit** — nicht „unbekannt". Wer einen Grund hat, nennt ihn.
type Bereitschaftsgrund string

const (
	// GrundBestandZu heißt: es ist kein Bestand geöffnet.
	//
	// **Der Normalfall beim Start**, und kein Fehler: das Fenster öffnet ihn,
	// sobald es weiß, welches Profil gemeint ist.
	GrundBestandZu Bereitschaftsgrund = "bestandZu"

	// GrundWanderungLaeuft heißt: die Migrationen laufen gerade.
	//
	// **Ausdrücklich nicht bereit** — das verlangt die Abnahme des Issues.
	// Wer währenddessen schriebe, schriebe gegen ein Schema, das es gleich
	// nicht mehr gibt.
	GrundWanderungLaeuft Bereitschaftsgrund = "wanderungLaeuft"

	// GrundBestandUnlesbar heißt: der Bestand ist da und antwortet nicht.
	//
	// **Das ist der ernste Fall** — eine gesperrte, beschädigte oder
	// verschlüsselte Datei ohne Schlüssel.
	GrundBestandUnlesbar Bereitschaftsgrund = "bestandUnlesbar"
)

// Fremdzustand ist, was über einen fremden Dienst bekannt ist.
//
// **Er trägt keinen Zeitpunkt.** Wann zuletzt gefragt wurde, gehört zur
// Auskunft des Dienstkatalogs (A1) und nicht hierher; dieser Typ beantwortet
// eine Frage: *hindert er uns?* — und die Antwort ist immer nein.
type Fremdzustand struct {
	Dienst     Dienst `json:"dienst"`
	Erreichbar bool   `json:"erreichbar"`

	// Grund steht nur, wenn er nicht erreichbar ist — und er ist für Menschen.
	Grund string `json:"grund"`
}

// Betriebszustand sagt, ob dieser Prozess lebt und ob er arbeiten kann.
type Betriebszustand struct {
	// Lebt heißt: der Prozess antwortet. Wer diese Antwort liest, hat den
	// Beweis schon in der Hand.
	Lebt bool `json:"lebt"`

	// Grund ist der **eigene** Grund, nicht bereit zu sein — leer, wenn er es
	// ist.
	Grund Bereitschaftsgrund `json:"grund"`

	// Fremde ist, was über die fremden Dienste bekannt ist — **eine Auskunft
	// und nie ein Grund**.
	Fremde []Fremdzustand `json:"fremde"`
}

// Bereit sagt, ob dieser Prozess arbeiten kann.
//
// **Gerechnet und nicht gespeichert**: ein zweites Feld daneben liefe gegen
// den Grund auseinander, und dann gäbe es zwei Antworten auf eine Frage.
func (b Betriebszustand) Bereit() bool {
	return b.Lebt && b.Grund == ""
}
