package app

// **Wie ein Bild ans Fenster geht** — die Wahl der Verdichtungsstufe.
//
// Eigene Datei, weil `app.go` mit ihr über 400 Zeilen ginge (§2). Der Schnitt
// folgt der Frage: dort steht, **woraus** die Anwendung besteht, hier eine
// Entscheidung über den Weg eines Bildes.

import (
	"image"
	"image/png"
	"io"
)

// schnellPacken kodiert ein Bild als PNG — **auf Geschwindigkeit gestellt**
// (#672).
//
// # Warum die Stufe hier eine Rolle spielt
//
// Das Bild geht an ein Fenster im selben Rechner, nicht über ein Netz. Es wird
// einmal gezeigt und weggeworfen; jedes Byte, das die Verdichtung spart, kostet
// Zeit, die jemand als Ruckeln sieht.
//
// **Gemessen** über 120 Vorschauen einer M15-Karte: von 349 ms je Karte gingen
// rund 60 ms in `compress/flate` — ein Sechstel für Bytes, die niemand
// aufhebt.
//
// **PNG bleibt verlustfrei, gleich welche Stufe.** `BestSpeed` ändert die
// Größe der Datei, nicht ihren Inhalt; das Bild, das ankommt, ist Punkt für
// Punkt dasselbe.
//
// **Für Dateien auf der Platte gilt das nicht** — `service.Bildausfuhr` und
// `service.Bildablage` schreiben, was jemand behält, und packen weiter mit der
// Vorgabe.
func schnellPacken(ziel io.Writer, bild image.Image) error {
	kodierer := png.Encoder{CompressionLevel: png.BestSpeed}

	return kodierer.Encode(ziel, bild)
}
