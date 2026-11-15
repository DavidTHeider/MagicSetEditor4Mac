package app

// **Der Netzweg an den Flächen ohne Fenster** — Werkstück N5 von Phase 5.4
// (#1571),
// Werkstück 2.
//
// **Weil `mse`, `mse-mcp` und `mse-rest` Klienten wie das Fenster sind.** Was
// sie zeigen, ist der Weg, den sie **selbst** nehmen — dieselben
// Einstellungen, dieselbe Deutung, dieselbe Herkunft. Eine zweite Rechnung
// dort wäre die Gestalt, in der ein Proxy an einer Fläche gilt und an der
// anderen nicht (§16).
//
// **Und sie ist klein wie [PeerflaecheUeber]**: der Netzweg braucht die
// Einstellungen und sonst nichts.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// NetzflaecheUeber baut eine Anwendung, die nach dem Netzweg gefragt werden
// kann.
func NetzflaecheUeber(db *sql.DB) *App {
	anwendung := NeueApp("")
	anwendung.db = db
	// **Mit demselben Verwalter wie das Fenster** und nicht mit dem rohen
	// Speicher: die Ebenen aus
	// [0079](../../docs/entscheidungen/0079-einstellungsebenen-und-portabler-import.md)
	// gelten hier genauso, und ein Netzweg, der sie überginge, zeigte einen
	// anderen Wert als den wirksamen.
	anwendung.MitEinstellungen(
		einstellungsverwalterUeber(db))

	return anwendung
}

// DienstflaecheUeber baut die Anwendung, die ein Dienst für sein Zeugnis
// braucht — Peeridentität **und** Einstellungen (NET-03, #1707).
//
// **Derselbe Verwalter wie am Fenster**, aus demselben Grund wie oben: die
// Pfade des Zeugnisses stehen auf der Ebene des Geräts, und ein Dienst, der
// die Ebenen überginge, läse einen anderen Wert als den wirksamen.
func DienstflaecheUeber(db *sql.DB) *App {
	anwendung := PeerflaecheUeber(db)
	anwendung.MitEinstellungen(einstellungsverwalterUeber(db))

	return anwendung
}

// mitGezeigtemProfil hängt das gezeigte Profil des Bestands an — für den
// Proxy der Flächen ohne Fenster (Befund 1 des Zweitblicks über Phase D,
// `netzvorgaben.go`). **Hier und nicht dort**: `internal/schichten` erlaubt
// den Import von `internal/storage` nur in einer Aufzählung, und diese Datei
// steht schon darin — dieselbe Überlegung wie bei [gezeigtesProfil].
func mitGezeigtemProfil(ctx context.Context, db *sql.DB) (context.Context, error) {
	return storage.MitGezeigtemProfil(ctx, db)
}
