package domain

// **Was an einem Objekt zu tun ist** — P3
// (#2371),
// Befund C-10. Welche Befehlsarten es gibt, weiß `service`; hier steht die
// Gestalt der Auskunft und die Frage, ob ein Sitz die Karte eines Objekts
// überhaupt kennt.

// Objekthandlungen sagt, was ein Sitz an einem Objekt gerade tun kann.
//
// **Primär und Zweit** sind die Handlungen für Klick und Doppelklick (P4);
// leer heißt, dass es keine gibt und das Kontextmenü der Weg ist.
//
// **Mit Feldnamen, denn sie geht über die Leitung** — im [Tischbild] an den
// Gast, gerechnet beim Wirt.
type Objekthandlungen struct {
	Objekt     string   `json:"objekt"`
	Handlungen []string `json:"handlungen"`
	Primaer    string   `json:"primaer"`
	Zweit      string   `json:"zweit"`
}

// KennenDieKarte sagt, ob `fuer` die Karte dieses Objekts sehen darf —
// dieselbe Regel wie die Sicht ([Partie.Sicht]): eine fremde Hand, jede
// Bibliothek und ein verdeckt liegendes fremdes Objekt sind geheim.
func KennenDieKarte(objekt Spielobjekt, fuer string) bool {
	return !objekt.fuerSicht(fuer).Geheim
}

// HatGrundlandart sagt, ob ein Objekt einen der fünf Grundlandtypen trägt —
// und damit die innewohnende Manafähigkeit aus CR 305.6.
func (p Partie) HatGrundlandart(objekt Spielobjekt) bool {
	_, fehler := p.grundlandart(objekt)

	return fehler == nil
}
