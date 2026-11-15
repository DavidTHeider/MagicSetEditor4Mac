package app

// **Die Bindungen der Sammelvorschau** — DOM-03.AK-01
// (#1698).
//
// **Sie schreiben nichts** (§16, Regel 6): was eine Einfuhr zusammenführte,
// sieht, wer einliest, bevor es geschieht — und wählt dann.

import "context"

// ListeVorschauen sagt, was das Einlesen der entschiedenen Zeilen täte.
func (a *App) ListeVorschauen(ctx context.Context, zeilen []Listeneinfuhrzeile) Sammelvorschauantwort {
	ctx = a.imProfil(ctx)

	if a.listeneinfuhr == nil {
		return Sammelvorschauantwort{Fehler: Uebersetze(ctx, keineListeneinfuhr())}
	}

	vorschau, err := a.listeneinfuhr.Vorschauen(ctx, zeilen)
	if err != nil {
		return Sammelvorschauantwort{Fehler: Uebersetze(ctx, err)}
	}

	return Sammelvorschauantwort{Vorschau: vorschau}
}

// SammlungCSVVorschauen sagt, was das Einlesen eines CSV täte.
func (a *App) SammlungCSVVorschauen(ctx context.Context, text string) Sammelvorschauantwort {
	ctx = a.imProfil(ctx)

	if a.sammlungsausfuhr == nil {
		return Sammelvorschauantwort{Fehler: Uebersetze(ctx, keineSammlungsausfuhr())}
	}

	vorschau, err := a.sammlungsausfuhr.CSVVorschauen(ctx, text)
	if err != nil {
		return Sammelvorschauantwort{Fehler: Uebersetze(ctx, err)}
	}

	return Sammelvorschauantwort{Vorschau: vorschau}
}

// ListeVorschauen sagt, was das Einlesen der entschiedenen Zeilen täte.
func (b *Bindungen) ListeVorschauen(ctx context.Context, zeilen []Listeneinfuhrzeile) Sammelvorschauantwort {
	return b.app(ctx).ListeVorschauen(b.ruf(ctx), zeilen)
}

// SammlungCSVVorschauen sagt, was das Einlesen eines CSV täte.
func (b *Bindungen) SammlungCSVVorschauen(ctx context.Context, text string) Sammelvorschauantwort {
	return b.app(ctx).SammlungCSVVorschauen(b.ruf(ctx), text)
}
