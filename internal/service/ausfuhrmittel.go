package service

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"

// Ausfuhrmittel sind die Fähigkeiten und Orte **einer** Ausfuhr —
// #736.
//
// Der Nullwert ist eine Ausfuhr, die nichts schreiben darf — und das ist der
// richtige Nullwert: ohne genannten Ort gibt es die schreibenden Namen nicht
// (#216,
// #657,
// #658).
//
// # Warum das ein Wert ist und kein Feld am Ausführer
//
// Es gibt **einen** `Vorlagenausfuehrer` je Anwendung
// (`vorlagenliste.go`, Zeile 235). Bis zum 2026-08-21 setzte `App.schreibeVorlage`
// seine Felder je Aufruf — Bildziel, Beigabenort, Setpaket, Verzeichnis. Die
// Oberfläche sperrt über `useVorgang` **je Set**; zwei Ausfuhren verschiedener
// Sets liefen damit ausdrücklich nebeneinander, und der zweite Lauf überschrieb
// dem ersten sein Ziel. `go test -race` zeigte drei Rennzustände.
//
// **Eine Sperre nahm der Sache die Wirkung, nicht die Gestalt.** Ein Wert je
// Lauf nimmt ihr beides: es gibt keinen geteilten veränderlichen Zustand mehr,
// und zwei Ausfuhren laufen wirklich nebeneinander.
//
// # Was hier steht und was am Ausführer bleibt
//
// Hier steht, was **je Ausfuhr** wechselt. Am Ausführer bleibt, was die
// Anwendung einmal einrichtet: die Symbolschriften und der Rasterer — sie
// hängen an den Paketwurzeln, nicht an einem Ziel.
//
// # Und warum kein Verweis auf den Ausführer
//
// Weil ein Wert, der auf ein veränderliches Ding zeigt, dessen Veränderlichkeit
// erbt. Der `ausfuhrlauf` hielt bis zum 2026-08-21 einen Zeiger auf den
// Ausführer; damit half auch eine Kopie der Felder nichts.
type Ausfuhrmittel struct {
	// Bilder, Bildziel und Bildbreite erlauben `write_image_file` über einer
	// Karte — siehe `exportbild.go`.
	Bilder     Bildschreiber
	Bildziel   string
	Bildbreite int

	// Beigaben erlaubt `copy_file`, `write_text_file` und `write_set_file` —
	// siehe `exportbeigaben.go`.
	Beigaben Beigabenschreiber

	// Symbolzeichnungen und Setpaket erlauben `symbol_variation` und
	// `write_image_file` über einem Bildwert — siehe `exportsetsymbol.go`.
	//
	// `Setpaket` ist der **Ort** des Sets auf der Platte: `set.symbol` nennt
	// eine `.mse-symbol` darin.
	Symbolzeichnungen Symbolquelle
	Setpaket          string

	// Stile gibt ein Stilpaket nach seinem Namen — leer heißt: der Stil des
	// Sets. Daraus werden `set.stylesheet` und `card.stylesheet`
	// (`exportstil.go`, #2039).
	Stile Ausfuhrstile

	// Verzeichnis ist der **relative** Name des Beigabenverzeichnisses —
	// `spoiler-files`. Leer heißt „die Vorlage legt keines an", und dann darf
	// sie gar nicht schreiben
	// (#732).
	Verzeichnis string
}

// Ausfuhrstile gibt ein Stilpaket nach seinem Namen; leer heißt der Stil des
// Sets. **Eine Funktion und keine Schnittstelle**: die Ausfuhr fragt genau
// eines, und ein Lauf ohne Bestand setzt keine. Nicht [Stilquelle] — die gibt
// dem Kartenplan den Stil **einer Karte** samt Kopf.
type Ausfuhrstile func(paket string) (msepaket.Stil, bool)

// gemerkt gibt dieselben Stile, aber **jeden Namen nur einmal gefragt** — die
// Quelle öffnet ihr Paket bei jedem Ruf. E04
// (#2241):
// `magic-namecheck` über Black fragte je Karte und brauchte 28,6 s gegen
// MSE2s 6,2 s. Gemerkt wird auch ein »nicht da« — es bliebe bei jedem Ruf
// dasselbe. Nicht nebenläufig: eine Ausfuhr rechnet ihre Karten der Reihe
// nach.
func (stile Ausfuhrstile) gemerkt() Ausfuhrstile {
	if stile == nil {
		return nil
	}

	type antwort struct {
		stil msepaket.Stil
		da   bool
	}

	gefragt := map[string]antwort{}

	return func(paket string) (msepaket.Stil, bool) {
		if gemerkt, bekannt := gefragt[paket]; bekannt {
			return gemerkt.stil, gemerkt.da
		}

		stil, da := stile(paket)
		gefragt[paket] = antwort{stil: stil, da: da}

		return stil, da
	}
}
