package app

import "context"

// **Ein Paket als `.mse-installer` weitergeben und einen einlesen** —
// FMT-02.AK-04
// (#1691).

// PaketWeitergeben fragt nach einem Ort und schreibt das Paket als Installer.
func (b *Bindungen) PaketWeitergeben(ctx context.Context, paket string) Paketweitergabe {
	return b.app(ctx).PaketWeitergeben(b.ruf(ctx), paket)
}

// InstallerEinlesen fragt nach einem Installer und liest seine Pakete ein.
func (b *Bindungen) InstallerEinlesen(ctx context.Context) UebernahmeAntwort {
	return b.app(ctx).InstallerEinlesen(b.ruf(ctx))
}
