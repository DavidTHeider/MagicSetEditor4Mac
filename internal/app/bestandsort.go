// **Wo eine Fläche ohne Fenster ihren Bestand sucht**
// (#886).
//
// # Warum es das gibt
//
// `mse` und `mse-mcp` arbeiten auf demselben Bestand wie das Fenster — das ist
// der Normalfall und bleibt die Vorgabe. Zwei Fälle brauchen etwas anderes:
//
//  1. **Ein Skript, das nichts anfassen soll.** Wer eine Stapelverarbeitung
//     ausprobiert, will sie nicht am gewachsenen Bestand üben.
//  2. **Ein Abnahmelauf.** Ein Test, der die gebauten Programme fährt, führe
//     sonst gegen die Daten der Nutzenden — und das ist kein Test, sondern ein
//     Eingriff.
//
// # Und das Fenster
//
// **Hier stand, das Fenster habe ihn nicht** — „eine Anwendung, deren Daten
// davon abhängen, womit man sie gestartet hat, ist eine Falle". Seit dem
// 2026-09-14 fragt `main.go` ihn trotzdem: die Sichtprüfung braucht einen
// eigenen Bestand, und ohne den Schalter schrieb sie in den echten. Wer die
// Anwendung doppelklickt, hat keine Umgebung gesetzt und bekommt den
// Vorgabeort — die Falle entsteht nur für den, der sie aufstellt.
//
// # Was zum Bestand gehört — und was zum Rechner
//
// **Dem Bestand folgen Datenbank, eingesetzte Bilder und das Protokoll des
// Fensters** (QA-10.NAT-01, #1741). Die Bilder folgten bis dahin nicht: ein
// Bestand mit eigenem Ort legte sie neben einen anderen.
//
// **Pakete, ihre Schriften und der Zwischenspeicher von Scryfall bleiben beim
// Rechner** — sie gehören ihm (`docs/AUTORISIERUNG.md`), und der erste Fall
// oben braucht sie: ein Skript an einem Wegwerfbestand zeichnet mit den
// Paketen dieses Rechners. Wer auch sie trennen will, trennt das Zuhause.

package app

import "os"

// Bestandsumgebung ist die Umgebungsvariable, die den Ort setzt.
const Bestandsumgebung = "MSE_BESTAND"

// BestandsortFuerFlaechen gibt den Ort, an dem eine Fläche ohne Fenster ihren
// Bestand sucht.
//
// **Die Umgebung schlägt den Vorgabeort** — und nur, wenn sie gesetzt **und
// nicht leer** ist: ein leerer Wert ist ein vergessenes `export`, kein Wunsch
// nach einer Datei namens „".
func BestandsortFuerFlaechen() (string, error) {
	if ort := os.Getenv(Bestandsumgebung); ort != "" {
		return ort, nil
	}

	return Datenbankpfad()
}
