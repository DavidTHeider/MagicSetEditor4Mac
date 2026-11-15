package app

import "context"

// **Der Dateibaum der Werkstatt am Fenster** — FMT-02.AK-01
// (#1691).
// Eine eigene Datei, weil `bindungenkarten.go` an der Zeilengrenze steht.

// PaketdateiAnlegen legt eine neue Vorlage an.
func (b *Bindungen) PaketdateiAnlegen(ctx context.Context, paket, datei, inhalt string) Paketdateistand {
	return b.app(ctx).PaketdateiAnlegen(b.ruf(ctx), paket, datei, inhalt)
}

// PaketdateiEinfuegen fragt nach einer Datei und legt sie ins Paket.
func (b *Bindungen) PaketdateiEinfuegen(ctx context.Context, paket, ordner string) Paketdateieinfuegung {
	return b.app(ctx).PaketdateiEinfuegen(b.ruf(ctx), paket, ordner)
}

// PaketdateiUmbenennen gibt einer Datei des Pakets einen neuen Namen.
func (b *Bindungen) PaketdateiUmbenennen(ctx context.Context, paket, alt, neu, stand string) SchreibAntwort {
	return b.app(ctx).PaketdateiUmbenennen(b.ruf(ctx), paket, alt, neu, stand)
}

// PaketdateiEntfernen nimmt eine Datei aus dem Paket.
func (b *Bindungen) PaketdateiEntfernen(ctx context.Context, paket, datei, stand string) SchreibAntwort {
	return b.app(ctx).PaketdateiEntfernen(b.ruf(ctx), paket, datei, stand)
}
