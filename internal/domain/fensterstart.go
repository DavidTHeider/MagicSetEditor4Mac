package domain

// **Wie das Fenster aufgeht** — UX-14.AK-01
// (#1686).
//
// # Was das Kriterium verlangt
//
// *„Beim ersten Start ist das Fenster standardmäßig maximiert, nicht
// automatisch im macOS-Vollbildmodus. Nutzer können Normalgröße,
// Maximierung oder Wiederherstellung des letzten Zustands einstellen."*
//
// **Drei Werte, und kein vierter.** Ein „vollbild" stünde da als Einladung zu
// genau dem, was der erste Satz ausschließt: Vollbild nimmt unter macOS die
// Menüleiste mit, und diese Anwendung bedient die Menüleiste (§6), statt sie
// zu verstecken. Wer Vollbild will, hat dafür `⌃⌘F` — das ist eine Handlung
// des Fensters und keine Vorgabe.
//
// # Warum das hier steht und nicht in `main.go`
//
// **Weil es eine Entscheidung ist und keine Wails-Angabe.** Die Hülle
// (`internal/fenster`) übersetzt sie seit Wails 3 in die Fensteroptionen
// und maximiert beim ersten `WindowRuntimeReady`; welche Werte es gibt und
// welcher gilt, wenn niemand etwas gesagt hat, ist eine Frage dieses
// Programms.

import "strconv"

// Die drei Werte von [EinstellungFensterstart].
const (
	// FensterstartMaximiert füllt den Bildschirm ohne Vollbild — die Vorgabe.
	//
	// **Sie ist die Vorgabe, weil dieses Fenster eine Arbeitsfläche ist.** Ein
	// Kartenarbeitsplatz zeigt Navigator, Karte und Inspector nebeneinander;
	// bei 1.280 Punkten geht das gerade, bei 1.800 bequem. Wer es kleiner
	// will, stellt es kleiner — wer es größer will, musste bisher bei jedem
	// Start ziehen.
	FensterstartMaximiert = "maximiert"

	// FensterstartNormal öffnet in der eingebauten Größe.
	FensterstartNormal = "normal"

	// FensterstartZuletzt stellt wieder her, wie das Fenster zuletzt stand.
	FensterstartZuletzt = "zuletzt"
)

// EinstellungFensterbreite und EinstellungFensterhoehe halten, wie groß das
// Fenster zuletzt war — für [FensterstartZuletzt].
//
// **Zwei Zahlen und keine Wahlliste**, wie `vorgabe.hochladepfad`. Eine
// Fenstergröße wählt niemand in einem Auswahlfeld; sie entsteht, indem man
// am Fensterrand zieht. Für sie gilt deshalb [PruefeFenstermass] statt einer
// Aufzählung.
//
// **In der Einstellungsreihe stehen sie trotzdem**: `PruefeEinstellung` wiese
// sonst jeden Schreibversuch ab, und die Ebenentabelle dürfte sie nicht
// führen. Sichtbar sind sie damit auch — und das ist Absicht: wer „wie
// zuletzt" wählt, soll sehen, was *zuletzt* ist.
const (
	EinstellungFensterbreite = "fensterbreite"
	EinstellungFensterhoehe  = "fensterhoehe"
)

// Fenstermindestmass und Fensterhoechstmass grenzen ein, was als
// Fenstergröße zurückgeschrieben werden darf.
//
// # Warum es eine Grenze braucht
//
// **Weil eine Zahl aus dem Fenster kommt und eine Zahl im Bestand steht.**
// Ein Bildschirm, der zwischen zwei Sitzungen wechselt, kann eine Größe
// hinterlassen, die auf dem nächsten nicht mehr passt — ein Fenster von
// 6.016 Punkten auf einem 1.440er Bildschirm wäre beim Start nicht mehr zu
// fassen. Das Höchstmaß ist die Breite zweier heutiger Bildschirme; das
// Mindestmaß ist die Höhe, die die Hülle als `MinHeight` setzt
// (`internal/fenster/bildschirm.go`).
//
// **Es ist keine Plausibilisierung des Monitors** — die steht seit
// UX-14.AK-04 in der Hülle (`internal/fenster/bildschirm.go`, `passend`), die
// eine gemerkte Größe auf die Arbeitsfläche klemmt. Dies hier hält nur, dass
// im Bestand keine sinnlose Zahl liegt.
const (
	Fenstermindestmass = 700
	Fensterhoechstmass = 8192
)

// PruefeFenstermass sagt, ob eine Zahl als Fenstergröße taugt.
func PruefeFenstermass(wert string) error {
	punkte, fehler := strconv.Atoi(wert)
	if fehler != nil {
		return NewFieldError("wert", SchluesselValidierung,
			"%q ist kein Fenstermaß — erwartet wird eine ganze Zahl in Punkten.",
			wert)
	}

	if punkte < Fenstermindestmass || punkte > Fensterhoechstmass {
		return NewFieldError("wert", SchluesselValidierung,
			"%d Punkte liegen außerhalb von %d bis %d.",
			punkte, Fenstermindestmass, Fensterhoechstmass)
	}

	return nil
}
