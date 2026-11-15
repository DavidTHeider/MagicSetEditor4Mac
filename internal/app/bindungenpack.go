package app

import "context"

// **Booster öffnen** — welche Packarten ein Set kennt, und eine ziehen.
//
// Eine eigene Datei, seit `bindungenkarten.go` mit dem Rufkontext aus BK-01
// (#2109) über 400 Zeilen ging (§2).

// Packarten gibt die Packarten des Spiels hinter einem Set
// (#1072).
//
// **Ein Vorgang je Bezugsobjekt** (§7, Regel 4): ein zweiter Aufruf bricht den
// ersten ab. Die Rechnung geht über **jede** Karte des Sets — 3,3 s bei 107
// Karten im Fenster, gemessen am 2026-08-29 —, und wer das Set wechselt, will
// die Antwort des vorigen nicht mehr
// (#1288).
//
// **Der abgebrochene Lauf meldet `abgebrochen` und keinen Fehler**; `useLaden`
// im Fenster lässt den vorigen Stand stehen und holt einmal nach.
func (b *Bindungen) Packarten(ctx context.Context, setID string) PackAntwort {
	ctx, fertig := b.app(ctx).packlauf.beginne(b.ruf(ctx))
	defer fertig()

	return b.app(ctx).Packarten(ctx, setID)
}

// PackZiehen zieht einen Booster.
//
// **Der Same als Text** — über die Wails-Bindung käme ein `uint64` als
// JavaScript-Zahl an, und die verliert oberhalb von 2⁵³ die letzten Stellen.
func (b *Bindungen) PackZiehen(ctx context.Context, setID, art, same string) PackziehungAntwort {
	return b.app(ctx).PackZiehen(b.ruf(ctx), setID, art, same)
}
