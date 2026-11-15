// **Die Paketwerkstatt für die Flächen ohne Fenster**
// (#1164).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort. Eine Vorlage zu ändern
// ist gerade auf der Kommandozeile und über MCP naheliegend — wer ein Skript
// in `card_fields` ändert, will es danach mit `mse zeichnen` ansehen.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Werkstattnaht ist die Paketwerkstatt, wie eine Fläche ohne Fenster sie ruft.
type Werkstattnaht struct{ werkstatt *service.Paketwerkstatt }

// WerkstattnahtUeber baut sie über dem Paketverzeichnis.
//
// **Ohne Bestand.** Eine Paketdatei liegt auf der Platte und nicht in der
// Datenbank; ein Griff danach wäre eine Abhängigkeit ohne Anlass.
func WerkstattnahtUeber() (*Werkstattnaht, error) {
	pakete, err := Paketverzeichnis()
	if err != nil {
		return nil, err
	}

	return &Werkstattnaht{werkstatt: service.NeuePaketwerkstatt(pakete)}, nil
}

// Dateien nennt die bearbeitbaren Dateien eines Pakets.
func (w *Werkstattnaht) Dateien(_ context.Context, paket string) ([]string, error) {
	return w.werkstatt.Dateien(paket)
}

// Lesen gibt eine Paketdatei samt ihrem Stand.
func (w *Werkstattnaht) Lesen(_ context.Context, paket, datei string) (inhalt, stand string, fehler error) {
	return w.werkstatt.Lesen(paket, datei)
}

// PaketSkripteGeprueft nennt Zeile, Spalte und Satz des ersten Syntaxfehlers
// in den Skripten eines Paketdateitextes — **Zeile 0**, wenn keiner darin
// steht.
//
// # Warum drei Zahlen und kein Verbund
//
// Weil `internal/mcp` ein **Blatt** ist (§2): es kennt `internal/domain` und
// sonst nichts aus diesem Projekt. Ein `*service.Skriptbefund` als Rückgabe
// zöge `service` in den MCP-Server; ein eigener Verbund in `mcp` wäre eine
// dritte Gestalt derselben Auskunft. Drei Rückgaben sind hier das kleinere
// Übel — und die Grenze `Zeile == 0` steht in beiden Erklärungen.
//
// **Geprüft wird der übergebene Text**, nicht die installierte Datei: eine
// Änderung soll sich prüfen lassen, bevor sie geschrieben wird.
//
// **Eine Funktion und keine Methode.** Sie braucht die Werkstatt nicht — der
// Text kommt mit. Als Methode zwänge sie jeden Aufrufer, erst eine Naht über
// das Paketverzeichnis zu bauen (und deren Fehlerweg zu behandeln), oder sie
// lüde ihn ein, einen Nullwert zu nehmen, der nur so lange trägt, wie
// niemand hier ein Feld anfasst.
func PaketSkripteGeprueft(inhalt string) (zeile, spalte int, meldung string) {
	befund := service.PaketSkriptfehler(inhalt)
	if befund == nil {
		return 0, 0, ""
	}

	return befund.Zeile, befund.Spalte, befund.Meldung
}

// Schreiben setzt eine Paketdatei zurück, wenn der Stand noch stimmt.
//
// **Der Kontext reist mit.** Der Schreibweg wartet auf die Sperre des Pakets,
// und ein abgebrochener Aufruf soll das Warten nicht abwarten.
func (w *Werkstattnaht) Schreiben(
	ctx context.Context, paket, datei, inhalt, erwartet string,
) (string, error) {
	return w.werkstatt.Schreiben(ctx, paket, datei, inhalt, erwartet)
}

// Texte reicht die Dateien samt Inhalt durch — für den Symbolindex (U7).
func (w *Werkstattnaht) Texte(_ context.Context, paket string) ([]service.Paketdateitext, error) {
	return w.werkstatt.Texte(paket)
}

// DateienUndBeiwerk nennt Vorlagen und Beiwerk getrennt (FMT-02, #1691).
func (w *Werkstattnaht) DateienUndBeiwerk(_ context.Context, paket string) (dateien, beiwerk []string, fehler error) {
	return w.werkstatt.DateienUndBeiwerk(paket)
}

// Anlegen legt eine neue Vorlage an — der Dateibaum (FMT-02, #1691).
func (w *Werkstattnaht) Anlegen(ctx context.Context, paket, datei, inhalt string) (string, error) {
	return w.werkstatt.Anlegen(ctx, paket, datei, inhalt)
}

// Einfuegen legt eine Datei von der Platte ins Paket.
func (w *Werkstattnaht) Einfuegen(ctx context.Context, paket, quelle, datei string) error {
	return w.werkstatt.Einfuegen(ctx, paket, quelle, datei)
}

// Umbenennen gibt einer Datei einen neuen Namen.
func (w *Werkstattnaht) Umbenennen(ctx context.Context, paket, alt, neu, erwartet string) error {
	return w.werkstatt.Umbenennen(ctx, paket, alt, neu, erwartet)
}

// Entfernen nimmt eine Datei aus dem Paket — in die Sicherung.
func (w *Werkstattnaht) Entfernen(ctx context.Context, paket, datei, erwartet string) error {
	return w.werkstatt.Entfernen(ctx, paket, datei, erwartet)
}
