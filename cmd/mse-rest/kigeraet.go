package main

import "context"

// KIGeraetemodellFuerSitzung schaltet Apples Gerätemodell für diesen Dienst
// ab oder an — AI-05.AK-04 (#2042). **Die Sitzung ist der Prozess**: die
// Anwendung hält die Naht je Prozess, nicht je Aufruf.
func (p *preise) KIGeraetemodellFuerSitzung(ctx context.Context, an bool) (bool, error) {
	antwort := p.kianwendung().KIGeraetemodellFuerSitzung(ctx, an)
	if antwort.Fehler != nil {
		return antwort.An, antwort.Fehler.AlsFehler()
	}

	return antwort.An, nil
}
