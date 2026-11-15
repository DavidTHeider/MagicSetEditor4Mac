package domain

// ProzessVeraltet sagt, dass ein laufender Prozess älter ist als der Bestand,
// den er offen hält — eine neuere Fassung hat ihn unter ihm gewandert (M02,
// #2346,
// Befund A-11). **Er schreibt nicht mehr und endet**; ein eingerichteter Dienst
// startet danach unter launchd mit der neuen Fassung neu, jeder andere Prozess
// von Hand.
//
// **Die Klasse ist [ErrBelegt]**: für den Aufrufer heißt es „gleich noch
// einmal" — dann antwortet der Prozess der neuen Fassung. Was wanderte, trägt
// [NeuereFassung].
type ProzessVeraltet struct {
	Fassung *NeuereFassung
}

func (p *ProzessVeraltet) Error() string {
	return "dieser Prozess ist älter als der Bestand und endet — " + p.Fassung.Error()
}

// Unwrap gibt die Klasse.
func (p *ProzessVeraltet) Unwrap() error { return ErrBelegt }
