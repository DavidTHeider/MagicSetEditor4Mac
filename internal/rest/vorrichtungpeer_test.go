package rest_test

// **Die Peerwege des willigen Doppels** — Werkstück N1 von Phase 5.4
// (#1454).
//
// **Eine eigene Datei**, weil `vorrichtungbestand_test.go` mit ihnen über die
// Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort steht, was ein
// Bestand hergibt, hier, womit sich dieses Programm ausweist.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) Peeridentitaet(context.Context) (rest.Peeridentitaet, error) {
	return rest.Peeridentitaet{Stand: "bereit", Kennung: "bprobe"}, nil
}

func (williger) PeeridentitaetSicherstellen(context.Context) (rest.Peeridentitaet, error) {
	return rest.Peeridentitaet{Stand: "bereit", Kennung: "bprobe"}, nil
}

// **Die Rotation nennt die Vorgängerin** — ohne sie könnte niemand einem
// Kontakt sagen, welche Kennung abgelöst wurde.
func (williger) PeeridentitaetRotieren(context.Context) (rest.Peeridentitaet, error) {
	return rest.Peeridentitaet{Stand: "bereit", Kennung: "bneu", Vorher: "bprobe"}, nil
}
