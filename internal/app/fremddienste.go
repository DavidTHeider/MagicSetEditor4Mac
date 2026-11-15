// **Was die Anwendung über fremde Dienste weiß** — Werkstück 2 von Phase 4.8
// (#1321).
//
// **Eine eigene Datei, und der Anlass war die Zeilengrenze** (§2): mit dem
// Dienstschalter ging `anwendung.go` darüber. Der Schnitt folgt trotzdem der
// Frage und nicht der Zahl — dort steht, **welche** Nähte es gibt, hier,
// welche davon **nach außen** greifen und wer sie fragen darf.
//
// **Warum das zusammengehört:** jede dieser Nähte entsteht über denselben
// einen Klienten ([Scryfallklient]), hält dieselbe Drossel ein und hängt seit
// Phase 4.8 an demselben Schalter. Sie in der langen Feldliste zu verstreuen
// hieß, dass die Zusage „aus heißt null Anfragen" an sechs Stellen mitgedacht
// werden müsste.

package app

// fremddienste ist, was die Anwendung über fremde Dienste weiß.
type fremddienste struct {
	// dienstwahl sagt, welcher fremde Dienst gefragt werden darf.
	//
	// **`nil` heißt: es gibt keinen Bestand, in dem ein Schalter stünde** —
	// der Unterschied zu „abgeschaltet" steht an [App.wahl].
	dienstwahl Dienstwahl

	// scryfall ist der Zugang zum fremden Dienst.
	//
	// **Im Composition Root gebaut, nicht im Verfahren** — sonst brauchte
	// jeder Test ein Zuhause, ein Verzeichnis und eine Gegenstelle.
	scryfall Scryfallquelle

	// scryfallsuche ist die Kartensuche bei Scryfall (#1167).
	scryfallsuche  Suchquelle
	scryfallbilder Bildquelle

	scryfallanlage Scryfallanleger
	scryfallsaetze Satzquelle

	// verweise bezieht entworfene Karten auf echte Drucke (#1167).
	verweise Verweisquelle

	// bund ist der Schlüsselbund für die Zugangsdaten — Werkstück 8 von
	// Phase 4.8.
	//
	// **`nil` heißt: der Anmeldebund von macOS.** Ein Test setzt einen
	// eigenen; ihn hier vorauszusetzen hieße, dass jeder Lauf einen anlegt
	// (§3).
	bund Geheimnisspeicher
}
