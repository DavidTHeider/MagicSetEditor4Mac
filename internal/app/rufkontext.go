package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Kontext eines Rufs aus dem Fenster** — BK-01
// (#2109).
//
// # Zwei Kontexte, und jeder trägt etwas anderes
//
// **Das Ende gehört dem Ruf.** Wails bricht ihn ab, wenn die Oberfläche ihn
// aufgibt oder das Fenster sich schließt — und am Ende jedes Rufs ohnehin.
//
// **Die Werte gehören zuerst dem Ruf und dann dem Fenster.** Was
// [App.Bereit] am Fenster ablegt, trug bis BK-01 jeder Ruf, und eine Fläche,
// die ihren Wurzelkontext selbst ausstattet (`storage.MitGezeigtemProfil`),
// soll das weiter können. Wer nur den Ruf nähme, verlöre es still.
//
// **Und die Vorgangskennung ist neu** — je Ruf eine. Bis BK-01 trug jeder
// Ruf die Kennung, die [App.Bereit] gewürfelt hatte: [MitNeuemVorgang] würfelt
// nur, wenn noch keine da ist, und am Fenster war immer schon eine.
type rufkontext struct {
	context.Context

	fenster context.Context
}

// Value sucht zuerst im Ruf und dann am Fenster.
func (k rufkontext) Value(schluessel any) any {
	if wert := k.Context.Value(schluessel); wert != nil {
		return wert
	}

	return k.fenster.Value(schluessel)
}

// rufkontext verbindet den Kontext eines Rufs mit dem des Fensters.
//
// **Eine Kennung, die der Ruf schon trägt, bleibt** — wer von außen eine
// bestimmte mitgibt, will sie wiederfinden.
func (a *App) rufkontext(ruf context.Context) context.Context {
	fenster := a.fenster
	if fenster == nil {
		fenster = context.Background()
	}

	verbunden := rufkontext{Context: ruf, fenster: fenster}
	if VorgangAus(ruf) != "" {
		return verbunden
	}

	return MitVorgang(verbunden, domain.NeueVorgangskennung())
}
