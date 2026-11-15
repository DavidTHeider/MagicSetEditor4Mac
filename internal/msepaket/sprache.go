package msepaket

// Sprachpaket sind die Beschriftungen aus einem `.mse-locale`.
//
// **Sie kommen aus dem Paket, nicht aus unseren Katalogen** (AGENTS.md §9). Was
// MSE selbst an Übersetzungen mitbringt, wird von dort gelesen; es nachzubauen
// hieße, dieselbe Aussage zweimal zu führen — und die zweite läuft weg.
type Sprachpaket struct {
	Name string

	// felder bildet Spiel → Feldname → Beschriftung ab.
	felder map[string]map[string]string

	// pakete bildet Paketname → Beschriftung ab.
	pakete map[string]string
}

// Feldbeschriftung gibt die Beschriftung eines Feldes.
//
// **Ohne Übersetzung kommt der Originalname zurück, nicht die leere
// Zeichenkette.** Eine fehlende Übersetzung kostet eine unübersetzte
// Beschriftung, nicht eine leere — dasselbe Verhältnis wie beim Rückfallsatz
// einer Fehlermeldung.
func (s Sprachpaket) Feldbeschriftung(spiel, feld string) string {
	if beschriftung, vorhanden := s.felder[Kanonisch(spiel)][Kanonisch(feld)]; vorhanden && beschriftung != "" {
		return beschriftung
	}

	return feld
}

// Paketbeschriftung gibt die Beschriftung eines Pakets.
func (s Sprachpaket) Paketbeschriftung(paket string) string {
	if beschriftung, vorhanden := s.pakete[paket]; vorhanden && beschriftung != "" {
		return beschriftung
	}

	return paket
}

// Sprachmodell liest ein Sprachpaket.
func (p *Paket) Sprachmodell() Sprachpaket {
	sprache := Sprachpaket{
		Name:   p.KurzerName,
		felder: map[string]map[string]string{},
		pakete: map[string]string{},
	}

	for _, spielblock := range p.Kopf.Alle("game") {
		for _, spiel := range spielblock.Kinder {
			if spiel.Zeilenart != ArtEintrag {
				continue
			}

			beschriftungen := map[string]string{}
			for _, feld := range spiel.Kinder {
				if feld.Zeilenart == ArtEintrag {
					beschriftungen[Kanonisch(feld.Schluessel)] = feld.Wert
				}
			}

			sprache.felder[Kanonisch(spiel.Schluessel)] = beschriftungen
		}
	}

	for _, paketblock := range p.Kopf.Alle("package") {
		for _, eintrag := range paketblock.Kinder {
			if eintrag.Zeilenart == ArtEintrag {
				sprache.pakete[eintrag.Schluessel] = eintrag.Wert
			}
		}
	}

	return sprache
}
