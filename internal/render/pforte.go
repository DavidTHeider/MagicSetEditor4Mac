package render

import "context"

// **Eine Rasterung wartet, bis sie dran ist** — Folgebefund F3 des Lastlaufs
// QA-08 (#2020),
// Marke `new`: MSE2 zeichnet auf dem Rechner eines Menschen, einen Dienst mit
// mehreren Aufrufern gibt es dort nicht.
//
// # Warum hier und nicht in jedem, der rastert
//
// Eine Karte, ein Druckbogen, ein Kartenbild, die Vorschau im Fenster, die
// Karten am Spieltisch, die Rasterteile einer SVG: alle gehen durch
// [Rasterer.Zeichne], und alle Flächen — Fenster, Kommandozeile, MCP und
// REST — bauen **denselben** Rasterer (`app.NeuesZeichenwerk`). Ein Platz in
// jedem Dienst wären sechs Stellen, und die siebte, die rastert, bekäme
// keinen. Bis zu diesem Befund hatte `RessourceRendern` eine Zahl in den
// Grenzen und keinen Aufrufer: niemand hat je einen Platz belegt.
//
// # Was `render` dabei nicht weiß
//
// **Was ein Platz ist und wie viele es gibt.** Der Rasterer fragt eine
// [Pforte], bevor er Bildpunkte anlegt, und gibt sie wieder zurück, wenn er
// fertig ist oder abbricht. Der Haushalt der Maschine steht in
// `internal/service` (`Rasterpforte`) — dieses Paket kennt nichts aus dem
// Projekt (§2).

// Pforte lässt eine Rasterung erst beginnen, wenn sie dran ist.
type Pforte interface {
	// Belegen wartet auf einen Platz — oder bis der Kontext endet — und gibt
	// die Rückgabe des Platzes. **Der Fehler ist der des Kontexts**, kein
	// eigener: ein Aufrufer, der sich abgemeldet hat, bekommt dasselbe, was er
	// bei einem Abbruch mitten im Zeichnen bekäme.
	Belegen(ctx context.Context) (freigeben func(), err error)
}

// MitPforte hängt die Pforte an, die jede Rasterung durchläuft.
//
// **Ohne Pforte ist jede Rasterung frei** — so wie bisher, und so bleibt es
// für jeden, der keinen Haushalt kennt. Gesetzt wird sie einmal beim Bau, nie
// im Betrieb: der Rasterer wird von mehreren Läufen zugleich benutzt.
func (r *Rasterer) MitPforte(pforte Pforte) *Rasterer {
	r.pforte = pforte

	return r
}

// platz belegt den Platz dieser Rasterung — und gibt ihn zurück, wenn das
// Ergebnis da ist oder der Lauf endet.
func (r *Rasterer) platz(ctx context.Context) (freigeben func(), err error) {
	if r.pforte == nil {
		return func() {}, nil
	}

	return r.pforte.Belegen(ctx)
}
