package fenster

// **Der Schluss der Hülle, unabhängig von einzelnen Paniken** — Befund 3 des
// Zweitblicks über Phase 1 von 0.9.0 (F02, #2341). Bis dahin fragte der
// Schlusshaken die Fenstergröße im selben Riegel, in dem er den Bestand
// schloss: warf die Frage, entfiel der ganze Schluss, und hing sie — in
// Wails' `InvokeSync`, das nach einer inneren Panik nie zurückkehrt
// (`pkg/application/mainthread.go`, v3.0.0-beta.27) —, hing er mit.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// fenstergroesse fragt die Größe des Fensters — **höchstens bis zum Ende von
// `frist`**, und eine Panik darin ist keine Antwort. Falsch heißt: keine
// Größe; gemerkt wird dann nichts.
//
// **Die Frage läuft weiter, wenn sie hängt** — aber niemand wartet mehr auf
// sie. Beim Schluss ist das der kleinere Verlust.
func fenstergroesse(frist context.Context, frage func() (int, int)) (breite, hoehe int, da bool) {
	antwort := make(chan [2]int, 1)
	gescheitert := make(chan struct{})

	go func() {
		if !riegeln("die Fenstergröße", func() {
			b, h := frage()
			antwort <- [2]int{b, h}
		}) {
			close(gescheitert)
		}
	}()

	select {
	case groesse := <-antwort:
		return groesse[0], groesse[1], true
	case <-gescheitert:
		return 0, 0, false
	case <-frist.Done():
		return 0, 0, false
	}
}

// schlussFahren ist der Schluss der Hülle: die Größe fragen, den Bestand
// schließen, das Gerätemodell beenden — **jeder Schritt in seinem eigenen
// Riegel**.
//
// **Wirft der Schluss, läuft er noch einmal, ohne Bequemlichkeit**: zweimal
// schließen ist keines ([app.App.Schliessen]), und der zweite findet vor, was
// der erste nicht mehr abräumen konnte. So geht der Bestand auch dann
// geordnet zu, wenn etwas beim Merken der Größe oder davor wirft.
func schlussFahren(
	frist context.Context, groesse func() (int, int),
	schliessen func(context.Context, func(context.Context)) error,
	merken func(context.Context, int, int), geraetBeenden func(),
) {
	knapp, aufhoeren := context.WithTimeout(frist, app.Bestenfallsfrist)
	breite, hoehe, da := fenstergroesse(knapp, groesse)

	aufhoeren()

	bestenfalls := func(ctx context.Context) {
		if da {
			merken(ctx, breite, hoehe)
		}
	}

	if !riegeln("der Schluss", func() { _ = schliessen(frist, bestenfalls) }) { //nolint:errcheck // der Schluss schreibt seinen Verlust selbst ins Protokoll
		riegeln("der Schluss, zweiter Versuch", func() { _ = schliessen(frist, nil) }) //nolint:errcheck // ebenso
	}

	riegeln("das Ende von fm serve", geraetBeenden)
}
