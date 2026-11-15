package app

// **Die Nähte der Spieltexturen** — Werkstück G2 von Phase 5.5
// (#1459).
//
// **Eingebettet und nicht als zwei Felder in [App]**, weil jene Datei sonst
// über die Zeilengrenze aus §2 geht — derselbe Schnitt wie bei
// `uebersichtsnaehte` und `partienaht` daneben. Der Verbund erklärt nichts
// Neues; er fasst zusammen, was zusammengehört.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"

// texturnaht hält, was das Fenster zum Zeichnen eines Spielbildes braucht.
type texturnaht struct {
	// texturen zeichnet Kartenbilder für den Spieltisch — nil heißt: keine.
	texturen Texturwerk

	// spielbilder hält sie für die Assetleitung — nil heißt: keines.
	spielbilder *vorschauweg.Texturfach
}
