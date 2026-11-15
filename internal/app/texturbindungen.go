package app

import "context"

// **Die Texturbindungen** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Warum sie nur am Fenster stehen
//
// **Weil die Adresse nur dort etwas bedeutet.** Ein Spielbild wird über die
// Assetleitung ausgeliefert, die Wails für **dieses** Fenster betreibt; wer
// die Adresse anderswo bekäme, bekäme einen Weg ins Leere.
//
// **Und die kopflose Gestalt derselben Sache gibt es schon**: *Ein Kartenbild
// schreiben* (`mse zeichnen`) steht an allen vier Flächen und schreibt eine
// **Datei**, die jemand behält. Genau diese Trennung steht seit
// #1270 an
// der Vorschau: „ein Bild zum Hinsehen, keines zum Behalten."

// Spieltextur gibt die Adresse des Kartenbildes für den Spieltisch.
//
// **Rein lesend.** Sie zeichnet, und das kostet Rechenzeit — geschrieben wird
// dabei nichts: kein Bestand, keine Datei, keine Zeile.
func (b *Bindungen) Spieltextur(ctx context.Context, karteID, seite, guete string) TexturAntwort {
	return b.app(ctx).Spieltextur(b.ruf(ctx), karteID, seite, guete)
}

// Rueckentextur gibt die Adresse des Kartenrückens eines Spiels.
//
// **Ein eigener Weg, weil ein verdecktes Objekt keine Kartenkennung trägt** —
// die Gründe stehen an [App.Rueckentextur].
func (b *Bindungen) Rueckentextur(ctx context.Context, spiel, guete string) TexturAntwort {
	return b.app(ctx).Rueckentextur(b.ruf(ctx), spiel, guete)
}

// Fremdkartenbild gibt das Bild eines Objekts, das dieses Fenster nicht selbst
// zeichnet — erfragt beim Wirt des fremden Tisches (P9, #2377, Vermerk 0198).
//
// **Rein lesend** wie die beiden darüber: das Bild landet im Fach dieses
// Fensters und nirgends sonst. **Nur am Fenster**, aus demselben Grund — und
// weil es eine laufende Gastsitzung voraussetzt, die nur das Fenster hält.
func (b *Bindungen) Fremdkartenbild(
	ctx context.Context, partie, objekt, bild, seite, guete string,
) TexturAntwort {
	return b.app(ctx).Fremdkartenbild(b.ruf(ctx), partie, objekt, bild, seite, guete)
}
