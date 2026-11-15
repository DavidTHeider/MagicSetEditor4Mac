// Die Griffe, an denen ein Schreiber hängt — und was sich davon austauschen
// lässt.
//
// **Eine eigene Datei, weil sie eine eigene Frage beantwortet** (§2): nicht
// *wie* eine Datei ins Paket kommt — das steht in `schreibweg.go` —, sondern
// *woran* der Weg dorthin hängt. Beides in einer Datei ging über die
// Zeilengrenze, als die Nachbardatei zur Schnittstelle wurde
// (#1289,
// Zweitblick über Phase 4.2).

package msepaket

import (
	"fmt"
	"io"
	"os"
)

// nachbardatei ist der Griff auf die Datei, die neben dem Ziel entsteht.
//
// **Eine Schnittstelle und kein `*os.File`**, damit ihre Fehlerwege betretbar
// sind — dieselbe Gestalt wie bei [archivBauen] und aus demselben Grund. Ein
// Fehlschlag beim Schreiben, beim Synchronisieren oder beim Schließen ist auf
// einer gesunden Platte nicht zu bestellen; ohne diese Naht bliebe „**das**
// abgebrochene Schreiben" ungeprüft, und genau so heißt das Werkstück
// (#1289).
type nachbardatei interface {
	io.Writer

	// Name ist der Weg, unter dem sie liegt — der Austausch braucht ihn.
	Name() string

	Sync() error
	Close() error
}

// schreibhilfen sind die Griffe, die ein Lauf austauschen kann.
//
// **Zusammen und nicht als zwei Parameter**: sie werden immer gemeinsam
// gesetzt, und eine Aufzählung von Funktionswerten in einer Signatur ist an der
// Aufrufstelle nicht mehr zu lesen.
type schreibhilfen struct {
	umbenennen func(alt, neu string) error
	nachbar    func(verzeichnis string) (nachbardatei, error)

	// schliessen gibt das gelesene Archiv frei.
	//
	// **Auch das ist ein Fehlergriff, und er war der letzte unbetretbare.**
	// `insArchiv` muss das Archiv schließen, bevor es umbenennt — unter Windows
	// ist eine offene Datei nicht zu ersetzen —, und scheitert das, steht die
	// Nachbardatei fertig da und der Tausch findet nicht statt. Ein Archiv, das
	// sich nicht schließen lässt, kommt auf einer gesunden Platte nicht
	// zustande; angestrichen vom Nachlauf über Phase 4.2, weil der Kopf des
	// Injektionslaufs „jede erreichbare Abbruchstelle" behauptete.
	schliessen func(io.Closer) error
}

// echteHilfen sind die Griffe des Betriebs.
func echteHilfen() schreibhilfen {
	return schreibhilfen{
		umbenennen: os.Rename,
		nachbar:    echterNachbar,
		schliessen: func(griff io.Closer) error { return griff.Close() }, //nolint:wrapcheck // der Griff reicht durch
	}
}

// echterNachbar legt die Nachbardatei an.
//
// **Über [os.CreateTemp]** und nicht unter einem ausgerechneten Namen: zwei
// Läufe zugleich träfen sich sonst darin, und was am Ende dastünde, wäre die
// Mischung zweier Stände (§7, Regel 4).
func echterNachbar(verzeichnis string) (nachbardatei, error) {
	datei, err := os.CreateTemp(verzeichnis, ".mse-schreibt-*")
	if err != nil {
		return nil, fmt.Errorf("die Nachbardatei anlegen: %w", err)
	}

	return datei, nil
}

// ablegen bringt eine Nachbardatei auf die Platte und schließt sie.
//
// **Mit [os.File.Sync].** Ohne ihn steht nach einem Stromausfall die
// umbenannte Datei da und ihr Inhalt nicht — das Umbenennen ist atomar, das
// Schreiben davor war es nie.
func ablegen(neben nachbardatei, inhalt []byte) error {
	var fehler error

	if inhalt != nil {
		_, fehler = neben.Write(inhalt)
	}

	if fehler == nil {
		fehler = neben.Sync()
	}

	if schliessfehler := neben.Close(); fehler == nil {
		fehler = schliessfehler
	}

	return fehler
}
