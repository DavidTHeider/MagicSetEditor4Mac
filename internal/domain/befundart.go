package domain

// Befundart sagt, **wem** ein Befund gehört
// (#513).
//
// # Warum ein Befundzähler ohne sie nichts sagt
//
// Der Lauf über alle 23.168 kanonischen Karten meldete am 2026-08-18 noch 174
// Befunde, und sie waren zwei völlig verschiedene Dinge:
//
//   - **159** „in der Schrift Beleren Small Caps Bold fehlen Zeichen" — die
//     Zeichen stehen in der Schriftdatei nicht drin, die auf **diesem** Rechner
//     liegt.
//   - **15** „das Bild /…/true ist nicht lesbar" — die Vorlage rechnet einen
//     Pfad aus, den es nicht gibt. Am Orakel geprüft: `"a" + true` gibt auch
//     bei MSE `"atrue"`, und MSE findet dort ebenso wenig ein Bild.
//
// **Keiner von beiden ist ein Portierungsfehler**, und beide addierten sich in
// dieselbe Zahl. Ein Zähler, der zwei verschiedene Dinge summiert, sagt bei
// keinem, was zu tun ist (§10).
//
// # Was die Trennung **nicht** ist
//
// **Kein Wegdefinieren.** Jeder Befund bleibt sichtbar, mit derselben Meldung
// an derselben Stelle; er wird nur einer Art zugeordnet und getrennt gezählt.
// Wer eine Art benutzt, um eine unbequeme Meldung loszuwerden, hat die
// Buchführung geändert und nicht das Programm.
type Befundart string

const (
	// ArtPortierung ist die Vorgabe: **dieses Programm kann etwas nicht**, was
	// MSE kann. Der Deckel darauf gehört auf null.
	ArtPortierung Befundart = "Portierung"

	// ArtVorlage heißt: **die Vorlage rechnet ins Leere**, und MSE täte es
	// genauso. Nachweispflichtig — wer sie setzt, sagt dazu, woran er das
	// geprüft hat.
	ArtVorlage Befundart = "Vorlage"

	// ArtUmgebung heißt: **es hängt an diesem Rechner** — eine Schrift, die
	// nicht installiert ist, eine Datei, die nicht daliegt. Auf einem anderen
	// Rechner steht die Zahl anders.
	ArtUmgebung Befundart = "Umgebung"

	// ArtEingabe heißt: **jemand hat etwas geschrieben, das nicht aufgeht** —
	// eine Zeile einer eingefügten Kartenliste, ein Name, den es nicht gibt,
	// eine Stückzahl von null.
	//
	// # Warum sie dazukam
	//
	// Die drei darüber sagen alle etwas über **die Anwendung oder den
	// Rechner**: was dieses Programm nicht kann, was in der Vorlage steht, was
	// hier nicht installiert ist. Keine sagt „die Eingabe war es".
	//
	// Ohne sie bekäme jede unlesbare Zeile einer eingefügten Liste den
	// Nullwert [ArtPortierung] — also die Aussage, MSE2 könne das und wir
	// nicht. Das ist falsch und wäre nicht auffällig: der Befund läse sich
	// richtig, und nur seine Einordnung wäre es nicht
	// (#1292).
	//
	// **Sie ist die einzige, die niemand beheben kann außer der Person selbst.**
	ArtEingabe Befundart = "Eingabe"
)

// Oder gibt die Art, oder die Vorgabe.
//
// **Der Nullwert ist [ArtPortierung]**, und das ist Absicht: wer einen Befund
// baut, ohne sich zu entscheiden, bekommt die strengste Einordnung. Andersherum
// verschwänden neue Befunde still in einer Sammelart.
func (a Befundart) Oder() Befundart {
	if a == "" {
		return ArtPortierung
	}

	return a
}
