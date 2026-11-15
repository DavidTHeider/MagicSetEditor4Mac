package app

// **Ein Entwurf des Fensters schreibt bedingt** — U02b
// (#2179).
//
// Das Fenster schickt mit seinem Entwurf die Werte, die es von diesem Feld
// kennt: den Ausgangswert, die selbst geschickten, die gesehenen
// (`frontend/src/haken/useEntwurf.ts`). Steht im Bestand ein anderer, hat ein
// anderer Klient dazwischen geschrieben — MCP, CLI, REST —, und der Entwurf
// wird **nicht** angenommen: `fremderWert` am Feld, ein Konflikt. Geprüft wird
// in der Transaktion, die schreibt (`storage.MitBekanntenWerten`).
//
// **Nur hier**: die Flächen ohne Fenster schreiben weiter unbedingt (AK-4) —
// sie haben keinen Entwurf, gegen den ein fremder Wert ein Konflikt wäre. Die
// Einordnung der Wails-Naht: verträglich erweiternd; sie ist keine Fläche
// nach `docs/entscheidungen/0033`.
//
// **Leer heißt ohne Bedingung**: ein Schreiben, das keinen Entwurf hat — die
// Wahl des Modus eines Schlüsselworts, ein abgelegtes Symbol —, schickt keine
// Werte. **Ein Einschub aus der Symbolpalette hat einen**: er geht durch den
// Entwurf des Feldes (Befund 1 des Zweitblicks über Phase 0.7.0-I).

import "context"

// FeldSetzenBedingt schreibt einen Feldwert wie [App.FeldSetzen] — **nur, wenn
// das Feld noch einen der Werte trägt, die das Fenster kennt**.
func (a *App) FeldSetzenBedingt(
	ctx context.Context, setID, karteID, feld, wert string, bekannt []string,
) FeldAntwort {
	ctx = a.imProfil(ctx)

	return a.FeldSetzen(mitBekannten(ctx, bekannt), setID, karteID, feld, wert)
}

// SetwertSetzenBedingt tut dasselbe für eine Setangabe.
func (a *App) SetwertSetzenBedingt(
	ctx context.Context, setID, feld, wert string, bekannt []string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.SetwertSetzen(mitBekannten(ctx, bekannt), setID, feld, wert)
}

// SchluesselwortSetzenBedingt tut dasselbe für ein Feld eines Schlüsselworts.
func (a *App) SchluesselwortSetzenBedingt(
	ctx context.Context, setID, wortID, feld, wert string, bekannt []string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.SchluesselwortSetzen(mitBekannten(ctx, bekannt), setID, wortID, feld, wert)
}

// WortWaehlenBedingt tut dasselbe für eine Wahl aus der Auswahlliste unter
// einem Kartenfeld — Befund 1 des Nachlaufs über Phase 0.7.0-I: sie rechnet
// aus dem gespeicherten Wert und schrieb ohne Bedingung, auch über einen
// fremden Wert, über den der Entwurf noch nicht entschieden hatte.
func (a *App) WortWaehlenBedingt(
	ctx context.Context, setID, karteID, feld string, bereich int, weg []int, bekannt []string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.WortWaehlen(mitBekannten(ctx, bekannt), setID, karteID, feld, bereich, weg)
}
