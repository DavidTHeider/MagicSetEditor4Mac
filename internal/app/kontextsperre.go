package app

import "context"

// sperreMitKontext nimmt eine Sperre — oder gibt auf, wenn `ctx` endet
// (Zweitblick über A und B, Befund 4).
//
// **`sync.RWMutex` kennt keinen Kontext**, und ein abgebrochener Ruf, der an
// einer gehaltenen Sperre stand, kam erst zurück, wenn sie frei wurde. Hier
// wartet eine Goroutine an seiner Stelle; gibt der Ruf auf, gibt sie die
// Sperre gleich wieder frei, sobald sie sie bekommt — sonst hielte niemand
// sie, und niemand gäbe sie zurück.
//
// **Erst der schnelle Weg**: ist die Sperre frei, kostet es keine Goroutine.
func sperreMitKontext(ctx context.Context, versuche func() bool, nimm, gib func()) bool {
	if versuche() {
		return true
	}

	erhalten := make(chan struct{})

	go func() {
		nimm()
		close(erhalten)
	}()

	select {
	case <-erhalten:
		return true
	case <-ctx.Done():
		go func() {
			<-erhalten
			gib()
		}()

		return false
	}
}
