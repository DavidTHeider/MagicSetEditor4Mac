package app

// **Der Horchstand an der Naht** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Warum die Adresse mitkommt und nicht nur ein Ja
//
// **Weil ein Mensch sie weitergeben muss.** Wer `:0` angibt, bekommt vom
// Betriebssystem einen Port; ohne die wirkliche Adresse in der Antwort wüsste
// niemand, welchen er dem Gegenüber nennen soll — und der Weg „Adresse von
// Hand" hätte an einem Ende keine Hand.

// Horchzustand sagt, ob und wo dieses Profil erreichbar ist.
type Horchzustand struct {
	// Laeuft heißt: es wird gerade gehorcht.
	Laeuft bool `json:"laeuft"`

	// Adresse ist, wo wirklich gehorcht wird — leer heißt: nirgends.
	//
	// **Wo gebunden wurde, nicht wo jemand ankommt.** Für „jede
	// Schnittstelle" steht hier `[::]:41234`, und das ist keine Adresse, zu
	// der sich wählen ließe — dafür gibt es [Horchzustand.Erreichbar].
	Adresse string `json:"adresse"`

	// ErreichbarNamen nennt die geprüften Rechnernamen dieses Rechners.
	//
	// **Sie stehen vor den Adressen, weil sie länger gelten.** Eine
	// IP-Adresse kommt in den meisten Netzen über DHCP, und eine Lease
	// endet: wer sie weitergibt, gibt eine Momentaufnahme weiter. Ein Name,
	// den das Netz auflöst, zeigt nach dem Wechsel auf die neue Adresse.
	//
	// **Geprüft heißt geprüft** (§10): er steht hier nur, wenn er wirklich
	// auf eine Adresse dieses Rechners auflöst. Ein geratener Name schickte
	// jemanden zu einem Fremden.
	ErreichbarNamen []string `json:"erreichbarNamen"`

	// ErreichbarAdressen nennt die Adressen, die ein Gegenüber eintippen kann.
	//
	// **Der Befund aus der Sichtprüfung vom 2026-09-14.** Hier stand nur
	// `Adresse`, und das Fenster sagte darunter „Erreichbar unter
	// `[::]:41234`" — wahr über die Bindung und unbrauchbar für den, der
	// wählen soll. Gemessen am selben Tag: der Rechner war unter
	// `192.168.178.51:41234` erreichbar, und diese Zeichenkette stand an
	// keiner Fläche.
	//
	// **Leer heißt: von außen nicht erreichbar.** Wer auf der Rückschleife
	// horcht, bekommt keine erfundene Adresse, sondern nichts — und die
	// Oberfläche sagt, was das bedeutet.
	ErreichbarAdressen []string `json:"erreichbarAdressen"`

	// Sichtbar heißt: dieses Profil kündigt sich im selben Netz an (N2,
	// #1455).
	//
	// **Ein zweiter Wahrheitswert und kein dritter Zustand**: erreichbar
	// und auffindbar sind zwei Dinge, und wer nur eine Zahl gäbe, müsste
	// „erreichbar, aber nicht auffindbar" hineindeuten.
	Sichtbar bool `json:"sichtbar"`
}

// Nachbar ist ein Gegenüber, das sich im selben Netz gemeldet hat.
type Nachbar struct {
	// Kennung ist, was es über sich sagt — **noch nicht geprüft**.
	Kennung string `json:"kennung"`

	// Fingerabdruck ist die Form, die zwei Menschen vergleichen.
	Fingerabdruck string `json:"fingerabdruck"`

	// Name ist der selbstgewählte — **kein Nachweis**.
	Name string `json:"name"`

	// Adresse ist, wohin gewählt würde.
	Adresse string `json:"adresse"`

	// Merkadresse ist, was in die Kontaktadresse gehört — der `.local`-Name,
	// wo es einen gibt, sonst dieselbe Adresse
	// (#1633).
	//
	// **Zwei Felder und nicht eines**: zum Wählen ist die Adresse schneller,
	// zum Merken ist der Name haltbarer — und wer nur eines zeigte, nähme
	// einem von beiden seinen Grund.
	Merkadresse string `json:"merkadresse"`

	// Bekannt heißt: dieses Profil führt ihn schon in seiner Kontaktliste.
	Bekannt bool `json:"bekannt"`

	// Stand ist seiner — leer, solange er unbekannt ist.
	Stand string `json:"stand"`

	// GemerkteAdresse ist, was die Kontaktliste über ihn weiß; und
	// AdresseGeaendert sagt, ob er woanders steht — dann gibt es eine
	// Adresse zu übernehmen (NET-02.NAT-02,
	// #1706).
	// **Entschieden im Dienst**, nicht im Fenster.
	GemerkteAdresse  string `json:"gemerkteAdresse"`
	AdresseGeaendert bool   `json:"adresseGeaendert"`
}

// NachbarnAntwort ist die Liste samt dem, was nicht zu deuten war.
type NachbarnAntwort struct {
	Nachbarn []Nachbar `json:"nachbarn"`

	// Verworfen zählt die Antworten, die keine Peerkennung trugen.
	//
	// **Sie steht da, damit „niemand gefunden" nicht wie „nichts gehört"
	// aussieht.** Auf `_mse4mac._tcp` antwortet dieses Programm; kommt etwas
	// anderes, ist das eine Auskunft über das Netz.
	Verworfen int `json:"verworfen"`

	Fehler *Fehler `json:"fehler"`
}

// Horchantwort ist der Stand samt seinem Fehler.
type Horchantwort struct {
	Stand  Horchzustand `json:"stand"`
	Fehler *Fehler      `json:"fehler"`
}

// ohneListen ist der Stand, wenn es keinen gibt.
//
// **Leere Listen und nicht `nil`.** Der Nullwert eines [Horchzustand] trägt
// zwei nil-Scheiben, und die gehen als `null` hinaus — das Fenster bekäme
// keine leere Liste, sondern gar keine. Genau die Klasse aus §3, und
// `leerelistenlaufzeit_test.go` hält sie.
//
// **An der Naht und nicht an jeder Rückgabestelle**: eine Fehlerantwort
// entsteht an zwei Stellen heute und an der dritten morgen.
func ohneListen() Horchzustand {
	return Horchzustand{
		ErreichbarNamen:    []string{},
		ErreichbarAdressen: []string{},
	}
}
