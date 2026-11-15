package app

import "context"

// **Die Sicherung des ganzen Bestands** — schreiben, wählen, beschauen,
// zurückspielen.
//
// Eine eigene Datei, seit `bindungen.go` mit dem Rufkontext aus BK-01 (#2109)
// über 400 Zeilen ging (§2). Nicht zu verwechseln mit `bindungensichern.go`:
// dort wird **ein Set** gesichert, hier der Bestand.

// BestandSichern schreibt eine Sicherung der ganzen Datenbank.
//
// #799.
//
// **Ohne Ziel**: gefragt wird im nativen Speichern-Dialog, und was dort
// gewählt wird, weiß das Fenster nicht besser als macOS. **Mit Kennwort und
// Klartextwunsch** seit OPS-06 (#1723) — beides leer heißt: wie bisher.
func (b *Bindungen) BestandSichern(ctx context.Context, kennwort string, klartext bool) Sicherungsantwort {
	return b.app(ctx).BestandSichern(b.ruf(ctx), kennwort, klartext)
}

// SicherungWaehlen fragt nach einer Sicherung und liest, was darin steht.
//
// **Sie übernimmt nichts** — die Rückfrage steht im Fenster (§6), und erst
// [Bindungen.SicherungZurueckspielen] tauscht.
func (b *Bindungen) SicherungWaehlen(ctx context.Context) Sicherungsauswahl {
	return b.app(ctx).SicherungWaehlen(b.ruf(ctx))
}

// SicherungBeschauen liest eine gewählte Sicherung mit ihrem Kennwort —
// OPS-06 (#1723).
func (b *Bindungen) SicherungBeschauen(ctx context.Context, pfad, kennwort string) Sicherungsauswahl {
	return b.app(ctx).SicherungBeschauen(b.ruf(ctx), pfad, kennwort)
}

// SicherungZurueckspielen ersetzt den Bestand durch die gewählte Sicherung —
// eine verschlüsselte mit ihrem Kennwort, sonst leer.
func (b *Bindungen) SicherungZurueckspielen(ctx context.Context, pfad, kennwort string) Rueckspielantwort {
	return b.app(ctx).SicherungZurueckspielen(b.ruf(ctx), pfad, kennwort)
}

// StartWiederholen wiederholt einen gescheiterten Start — F05 (#2344).
//
// **Frei wie [Bindungen.Startbefund]**: ohne offenen Bestand gibt es keine
// Sitzung, deren Tor den Ruf halten könnte, und gerade dann wird er gerufen.
func (b *Bindungen) StartWiederholen(ctx context.Context) Startbefund {
	return b.frei().StartWiederholen(b.ruf(ctx))
}
