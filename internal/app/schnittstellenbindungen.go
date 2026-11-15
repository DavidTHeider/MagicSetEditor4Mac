package app

import "context"

// **Schnittstellen am Fenster** — OPS-09 (#1907). Warum nur hier
// geschrieben wird, steht im Kopf von `schnittstellen.go`.

// Schnittstellen liest den Stand von Kommandozeile und Werkzeugserver.
func (b *Bindungen) Schnittstellen(ctx context.Context) SchnittstellenAntwort {
	return b.app(ctx).Schnittstellen(b.ruf(ctx))
}

// SchnittstelleEinrichten richtet eine Schnittstelle ein.
func (b *Bindungen) SchnittstelleEinrichten(ctx context.Context, name string) SchnittstellenAntwort {
	return b.app(ctx).SchnittstelleEinrichten(b.ruf(ctx), name)
}

// SchnittstelleWiderrufen nimmt den eigenen Eintrag wieder heraus.
func (b *Bindungen) SchnittstelleWiderrufen(ctx context.Context, name string) SchnittstellenAntwort {
	return b.app(ctx).SchnittstelleWiderrufen(b.ruf(ctx), name)
}

// SchnittstellenAlleEinrichten richtet mit einem Ruf ein, was fehlt — G08.
func (b *Bindungen) SchnittstellenAlleEinrichten(ctx context.Context) SchnittstellenAlleAntwort {
	return b.app(ctx).SchnittstellenAlleEinrichten(b.ruf(ctx))
}

// SchnittstellenAlleWiderrufen nimmt mit einem Ruf zurück, was eingerichtet ist.
func (b *Bindungen) SchnittstellenAlleWiderrufen(ctx context.Context) SchnittstellenAlleAntwort {
	return b.app(ctx).SchnittstellenAlleWiderrufen(b.ruf(ctx))
}

// Netzschnittstellen liest die Schnittstellen dieses Rechners — G08.
func (b *Bindungen) Netzschnittstellen(ctx context.Context) NetzschnittstellenAntwort {
	return b.app(ctx).Netzschnittstellen(b.ruf(ctx))
}
