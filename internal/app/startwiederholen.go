package app

import "context"

// StartWiederholen wiederholt einen gescheiterten Start und gibt den neuen
// Befund — F05 (#2344).
//
// **Der Knopf zu „belegt", „voll" und „gesperrt"**: diese Lagen vergehen,
// wenn der andere Prozess fertig, Platz geschaffen oder der Ort freigegeben
// ist — und dann soll derselbe Bestand aufgehen, ohne die Anwendung neu zu
// starten.
//
// **Unter der Sperre des Tauschs**: ein Zurückspielen, das gleichzeitig
// liefe, öffnete denselben Bestand ein zweites Mal. **Nur nach einem
// gescheiterten Start** — ein offener Bestand bleibt, wie er ist.
//
// **Über [App.wiederOeffnen] und nicht über `Starten`**: nach dem Beginn des
// Schlusses wird nicht mehr geöffnet, und ein Schluss, der **während** des
// Öffnens begann und nicht auf die Sperre warten konnte, findet den Bestand
// danach geschlossen. Bis zum Zweitblick über Phase 1 von 0.9.0 (Befund 2)
// prüfte diese Tür nur davor, und der eben geöffnete Bestand blieb offen.
func (a *App) StartWiederholen(ctx context.Context) Startbefund {
	ctx = a.imProfil(ctx)

	a.lauf.wechsel.Lock()
	defer a.lauf.wechsel.Unlock()

	if a.startfehler != nil {
		_ = a.wiederOeffnen(ctx) //nolint:errcheck // der Befund trägt ihn
	}

	return a.Startbefund()
}
