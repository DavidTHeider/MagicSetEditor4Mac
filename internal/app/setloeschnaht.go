// **Das Löschen eines Sets für die Flächen ohne Fenster** —
// #1218.
//
// **Aus demselben Grund exportiert wie [Einfuhrnaht]**: was zusammengesetzt
// wird, gehört an **eine** Stelle. Kommandozeile, MCP und REST dürfen
// `service` und `storage` nicht kennen; hätten sie je eine eigene Fassung
// davon, liefen drei Wurzeln auseinander — und die eine, die dann die
// Bildablage vergäße, ließe Verzeichnisse stehen, die niemandem mehr gehören.

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Setloeschnaht ist das Löschen, wie die Flächen ohne Fenster es sehen.
//
// **Sie gibt Befundtexte und keine [domain.Befund]**: MCP, REST und die
// Kommandozeile schreiben sie als Zeilen aus, und keine von ihnen hat für Datei
// und Art einen Platz. Das Fenster bekommt sie über [App.SetLoeschen] — auch
// dort als Text.
type Setloeschnaht struct{ loeschung *service.Setloeschung }

// SetloeschnahtUeber baut sie über einer Datenbank.
//
// **Ohne Bildablage, wenn das Zuhause fehlt.** Das Verzeichnis kommt aus dem
// Benutzerverzeichnis, und wenn das nicht zu ermitteln ist, soll das Löschen
// trotzdem gehen: die Zeilen sind das Wesentliche, die Bilder sind das
// Aufräumen. Der Fehler kommt als **Befund** zurück, nicht als Abbruch.
func SetloeschnahtUeber(db *sql.DB) *Setloeschnaht {
	loeschung := service.NeueSetloeschung(storage.NeuerSetSpeicher(db))

	if ort, err := Bilderverzeichnis(); err == nil {
		loeschung = loeschung.MitBildablage(service.NeueBildablage(ort))
	}

	return &Setloeschnaht{loeschung: loeschung}
}

// Loeschen nimmt das Set — die Gestalt, die das **Fenster** braucht.
//
// **Sie gibt [domain.Befund] und nicht Text**, weil die Übersetzung in
// Oberflächentext Sache von `app` ist und nicht der Naht (§2). [App.SetLoeschen]
// macht sie.
func (n *Setloeschnaht) Loeschen(ctx context.Context, setID int64, name string) (
	[]domain.Befund, error,
) {
	return n.loeschung.Loeschen(ctx, setID, name)
}

// SetLoeschen ist dieselbe Sache für die Flächen **ohne** Fenster.
//
// **Der Unterschied ist die Gestalt der Befunde und sonst nichts.** Sie steht
// hier und nicht dreimal in `cmd/`, damit Kommandozeile, MCP und REST
// dieselben Zeilen ausgeben — eine Übersetzung je Fläche liefe auseinander.
func (n *Setloeschnaht) SetLoeschen(ctx context.Context, setID int64, name string) (
	[]string, error,
) {
	befunde, fehler := n.Loeschen(ctx, setID, name)
	if fehler != nil {
		return nil, fehler
	}

	return befundtexte(befunde), nil
}
