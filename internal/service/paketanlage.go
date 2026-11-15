// **Ein Spiel- und ein Stilpaket anlegen** —
// #1350.
//
// # Was hier entschieden wird und nicht in `internal/msepaket`
//
// Das Blatt kann einen Baum als Datei schreiben. **Was in einem neuen Paket
// steht**, ist fachlich und steht deshalb hier — samt der drei Fragen, die das
// Issue offen ließ:
//
//  1. **Wo ein neues Paket liegt**: im verwalteten Verzeichnis, wie ein
//     bezogenes. Nicht aus Bequemlichkeit — [Paketverwalter.ortPruefen] weist
//     alles andere ab, und [msepaket.Paketwurzeln] macht nur verwaltete Orte
//     zur Wurzel. Ein Paket daneben läge da und wäre für Stil, Bild und
//     Schrift unauffindbar.
//  2. **Was ein leeres Paket mitbringt**: drei Kartenfelder und ein Setfeld.
//     MSE2 verlangt nichts davon — `Game::validate`
//     (`src/data/game.cpp`, Zeilen 66–88, Fassung f9d9356 vom 2021-07-06)
//     baut Statistiken und einen Pack selbst und kommt mit null Feldern
//     durch. **Ein Spiel ohne Felder ist trotzdem nutzlos**: eine Karte
//     darauf hat nichts, was man eingeben könnte.
//  3. **Ob ein Set seine Vorlage ausführen kann** — „aus diesem Set ein
//     Stilpaket machen" ist nicht gebaut. Es ist der nächste Schritt und
//     nicht dieser; was hier entsteht, ist die Vorlage, auf der ein Set
//     überhaupt erst anfängt.
//
// # Warum Spiel und Stil zusammen
//
// Die Abnahme verlangt, dass sich auf dem neuen Spiel **ein Set anlegen**
// lässt — und ein Set trägt Spiel **und** Stil (`domain.Set`). Ein Spielpaket
// allein wäre eine Fähigkeit, die man nicht benutzen kann.

package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Am Verwalter und nicht in einem eigenen Dienst.**
//
// Ein Paket anzulegen *ist* Paketverwaltung: es endet in
// [Paketverwalter.Installieren], damit gelten dieselbe Ortsprüfung, dieselbe
// Konfliktregel und derselbe Katalogeintrag wie für ein bezogenes Paket — und
// es braucht dieselben Wurzeln, um das Spiel eines Stils zu finden. Ein
// zweiter Dienst daneben wäre ein zweiter Weg in dasselbe Verzeichnis und
// eine zweite Naht, die jede Fläche verdrahten müsste.

// Neuanlage ist der Wunsch nach einem neuen Paket.
type Neuanlage struct {
	// Art ist `mse-game`, `mse-style`, `mse-include` oder
	// `mse-export-template` ([anlegbar]).
	Art string

	// Name wird zum Verzeichnisnamen und zum `short name`.
	Name string

	// Vollname ist der `full name` — was in der Paketliste steht.
	Vollname string

	// Spiel ist das Spielpaket, zu dem ein Stil gehört — für ein Spiel leer.
	Spiel string
}

// Anlegen schreibt das Paket und trägt es in den Katalog.
//
// **Erst in ein Wegwerfverzeichnis, dann installieren.** Ein halb geschriebenes
// Paket im verwalteten Verzeichnis wäre eines, das der Katalog beim nächsten
// Auffrischen als Befund führt — und niemand wüsste, ob es jemand angefangen
// oder ein Fehler hinterlassen hat (§8).
func (v *Paketverwalter) Anlegen(ctx context.Context, wunsch Neuanlage) (domain.Paket, error) {
	if err := v.pruefe(wunsch); err != nil {
		return domain.Paket{}, err
	}

	verzeichnis, name := wunsch.Name, msepaket.Art(wunsch.Art).Hauptdatei()
	if brauchtSpiel(wunsch.Art) {
		verzeichnis = stilverzeichnis(wunsch.Spiel, wunsch.Name)
	}

	inhalt, err := v.inhalt(wunsch)
	if err != nil {
		return domain.Paket{}, err
	}

	// **Erst den Inhalt, dann das Verzeichnis.** Ein Stil auf einem Spiel, das
	// es nicht gibt, hinterließe sonst ein leeres Verzeichnis im Wegwerfort.
	//
	// **Das Paketverzeichnis heißt, wie es heißen soll** — `Installieren`
	// liest den Namen aus dem letzten Pfadglied. Der Wegwerfname steht
	// deshalb am **Elternverzeichnis** und nicht am Paket; beim ersten Anlauf
	// stand er am Paket, und das Ergebnis hieß `mse4-anlage-eigenes`.
	//
	// **Ein eigener Ort je Bau** (`wegwerfpaket`) — ein fester Name ließ zwei
	// Anlagen nebeneinander einander das Verzeichnis wegräumen. Zweimal
	// derselbe Name bleibt ein Konflikt — bei `Installieren`.
	bau, ort, err := wegwerfpaket(v.bauort(), "mse4-anlage-", verzeichnis+"."+wunsch.Art)
	if err != nil {
		return domain.Paket{}, fmt.Errorf("das Paket vorbereiten: %w", err)
	}

	defer func() { _ = os.RemoveAll(bau) }() //nolint:errcheck // Wegwerfverzeichnis

	if err := v.ablegen(filepath.Join(ort, name), inhalt); err != nil {
		return domain.Paket{}, fmt.Errorf("das Paket schreiben: %w", err)
	}

	return v.einbinden(ctx, ort, false)
}

// bauort ist das Verzeichnis, in dem ein Paket entsteht, bevor es installiert
// wird.
//
// **Ein Feld und keine Konstante** — dieselbe Bauform wie `schluesselbund.Bund`
// und aus demselben Grund: so lässt sich der Fehlerweg fahren, ohne die Platte
// des Rechners vollzuschreiben. Leer heißt: der Wegwerfort des Systems.
func (v *Paketverwalter) bauort() string {
	if v.bau != "" {
		return v.bau
	}

	return os.TempDir()
}

// pruefe weist ab, was kein Paket werden kann.
func (v *Paketverwalter) pruefe(wunsch Neuanlage) error {
	if !anlegbar[wunsch.Art] {
		return domain.NewFieldError("art", domain.SchluesselValidierung,
			"Anlegen lässt sich ein Spiel (mse-game), ein Stil (mse-style), ein "+
				"Include (mse-include) oder eine Exportvorlage (mse-export-template), "+
				"nicht %q.", wunsch.Art)
	}

	// **Ein Name, der kein Verzeichnisname ist, wird abgewiesen.** Ein
	// Schrägstrich legte das Paket woanders an, ein Punkt machte aus der
	// Endung eine andere — beides außerhalb dessen, was der Verwalter kennt.
	if !istPaketname(wunsch.Name) {
		return domain.NewFieldError("name", domain.SchluesselPaketnameUngueltig,
			"Ein Paketname besteht aus Buchstaben, Ziffern, Bindestrichen und "+
				"Unterstrichen; %q geht nicht.", wunsch.Name)
	}

	if !brauchtSpiel(wunsch.Art) {
		return nil
	}

	if strings.TrimSpace(wunsch.Spiel) == "" {
		return domain.NewFieldError("spiel", domain.SchluesselPflichtfeld,
			"Ein Stil und eine Exportvorlage gehören zu einem Spiel; ohne dessen "+
				"Namen weiß niemand, welche Felder sie meinen.")
	}

	// **Der Spielname wird anders geprüft als der eigene, und das ist der
	// Unterschied zwischen Schreiben und Lesen.**
	//
	// `Name` wird zu einem Verzeichnis, das **wir** anlegen — dort gilt die
	// geschlossene Liste. `Spiel` benennt ein Paket, das **schon da ist**:
	// eingebunden über [Paketverwalter.Installieren] oder
	// [Paketverwalter.Uebernehmen], und beide nehmen den Namen so, wie er auf
	// der Platte steht. Ein `späß.mse-game` kann also im Katalog stehen — und
	// die geschlossene Liste wies dann den Stil darauf ab, obwohl das Spiel
	// benutzbar ist. Befund 2 des Nachlaufs über Phase 4.10.
	//
	// **Was bleibt, ist die Wegsicherheit.** `spieldatei` hängt den Namen an
	// eine Paketwurzel; ein `..` oder ein Schrägstrich darin läse eine Datei
	// außerhalb jeder Wurzel — angestrichen von `gosec` (G304), zu Recht. Die
	// Prüfung dafür ist enger als die Zeichenliste und trifft genau das.
	if !istWegsicher(wunsch.Spiel) {
		return domain.NewFieldError("spiel", domain.SchluesselValidierung,
			"Ein Spielname darf keinen Schrägstrich und keinen Punkt tragen; "+
				"%q geht nicht.", wunsch.Spiel)
	}

	return nil
}

// istWegsicher sagt, ob ein **gelesener** Name gefahrlos an eine Wurzel
// gehängt werden darf.
//
// **Ein Schrägstrich legte das Paket woanders an, ein Punkt machte aus der
// Endung eine andere, ein `..` ginge eine Ebene höher** — alle drei führen aus
// dem verwalteten Verzeichnis heraus, das §11 verlangt. Das ist die ganze
// Zusage: **Wegsicherheit, nicht Wohlgeformtheit.**
//
// **Sie gilt für den Spielnamen eines neuen Stils**, denn der benennt ein
// Paket, das schon da ist — eingebunden oder übernommen, und beide nehmen den
// Namen so, wie er auf der Platte steht. Ihn hier so eng zu prüfen wie einen
// selbst angelegten wies einen benutzbaren Stil ab (Nachlauf über Phase 4.10).
func istWegsicher(name string) bool {
	return name != "" && name == strings.TrimSpace(name) &&
		!strings.ContainsAny(name, `/\.`)
}

// istPaketname sagt, ob eine Zeichenkette als **neu angelegter**
// Verzeichnisname taugt.
//
// # Was abgewiesen wird und warum
//
// **Ein Schrägstrich legte das Paket woanders an, ein Punkt machte aus der
// Endung eine andere, ein `..` ginge eine Ebene höher** — alle drei führen aus
// dem verwalteten Verzeichnis heraus, das §11 verlangt.
//
// # Warum die Liste geschlossen ist und nicht offen
//
// **Weil die Meldung es versprach und die Prüfung es nicht hielt.** Hier stand
// eine Prüfung, die nur Schrägstrich, Punkt und Randleerraum abwies, und
// daneben ein Satz, der „Buchstaben, Ziffern und Bindestriche" nannte:
// `stil!` und `späß` gingen durch. Ein Kopf, der mehr verspricht als sein
// Rumpf hält, ist schlimmer als gar keiner (§10) — Befund 2 des Zweitblicks
// über Phase 4.10.
//
// **Gezählt, nicht geschätzt** (§10): über alle **526** Pakete von
// `Full-Magic-Pack` kommen in einem Paketnamen genau zwei Zeichen außer
// Buchstaben und Ziffern vor — `-` und `_`. Kein Leerzeichen, kein Punkt,
// nichts sonst. Die frühere Zusage hatte den Unterstrich vergessen; sechs
// Pakete tragen einen.
//
// **Und das Nein zu Umlauten ist eine Entscheidung dieses Projekts**, keine
// von MSE2: dort ist der Paketname der Dateiname, und gelesen wird, was
// dasteht. Ein Name reist aber in `depends on:` und in `stylesheet:` mit,
// über Dateisysteme hinweg — und dieses Programm wird für Windows und Linux
// mitübersetzt (§12).
func istPaketname(name string) bool {
	if name == "" {
		return false
	}

	for _, zeichen := range name {
		erlaubt := zeichen == '-' || zeichen == '_' ||
			(zeichen >= '0' && zeichen <= '9') ||
			(zeichen >= 'a' && zeichen <= 'z') ||
			(zeichen >= 'A' && zeichen <= 'Z')

		if !erlaubt {
			return false
		}
	}

	return true
}

// inhalt baut die Hauptdatei des Pakets.
func (v *Paketverwalter) inhalt(wunsch Neuanlage) ([]byte, error) {
	switch wunsch.Art {
	case "mse-game":
		return msepaket.Schreibe(spielbaum(wunsch)), nil
	case "mse-include":
		return msepaket.Schreibe(includebaum(wunsch)), nil
	case "mse-export-template":
		// **Das Spiel muss da sein** — MSE2 öffnet die Vorlage sonst nicht.
		if _, err := v.spieldatei(wunsch.Spiel); err != nil {
			return nil, err
		}

		return msepaket.Schreibe(exportbaum(wunsch)), nil
	}

	felder, err := v.spielfelder(wunsch.Spiel)
	if err != nil {
		return nil, err
	}

	return msepaket.Schreibe(stilbaum(wunsch, felder)), nil
}

// spielfelder liest die Kartenfelder des Spiels, zu dem ein Stil gehört.
//
// **Gelesen und nicht angenommen.** Ein Stil, der Felder auslegt, die es nicht
// gibt, zeichnet nichts — und einer, der vorhandene auslässt, lässt sie leer.
// Beides sähe aus wie ein Fehler in der Vorlage.
func (v *Paketverwalter) spielfelder(spiel string) ([]saatfeld, error) {
	roh, err := v.spieldatei(spiel)
	if err != nil {
		return nil, err
	}

	baum, zerteilfehler := msepaket.Zerteile("game", roh)
	if zerteilfehler != nil {
		return nil, domain.NewFieldError("spiel", domain.SchluesselValidierung,
			"Das Spiel %q ist nicht lesbar: %v", spiel, zerteilfehler)
	}

	var felder []saatfeld

	for _, knoten := range baum.Wurzel {
		if msepaket.Kanonisch(knoten.Schluessel) != "card_field" {
			continue
		}

		// **Name und Art, nicht nur der Name.** MSE2 liest `font:` allein an
		// einem `TextStyle` (`src/data/field/text.cpp`, Zeilen 143–159,
		// Fassung f9d9356 vom 2021-07-06); ein `ImageStyle` kennt nur
		// `default` (`src/data/field/image.cpp`, Zeilen 24–27). Ein `font`
		// unter einem Bildfeld gibt beim Öffnen „Unexpected key: 'font'" —
		// nachgemessen mit dem Orakel (§15).
		var gefunden saatfeld

		for _, kind := range knoten.Kinder {
			switch msepaket.Kanonisch(kind.Schluessel) {
			case "name":
				gefunden.name = kind.Wert
			case "type":
				gefunden.art = kind.Wert
			}
		}

		if gefunden.name != "" {
			gefunden.hoehe = hoeheFuer(gefunden.name)
			felder = append(felder, gefunden)
		}
	}

	if len(felder) == 0 {
		return nil, domain.NewFieldError("spiel", domain.SchluesselValidierung,
			"Das Spiel %q führt kein Kartenfeld; ein Stil darauf legte nichts aus.",
			spiel)
	}

	return felder, nil
}

// spieldatei sucht die `game`-Datei eines Spiels in **allen** Wurzeln.
//
// **Nicht nur im verwalteten Verzeichnis.** Ein Stil darf auf einem bezogenen
// Spiel aufsetzen — `magic.mse-game` liegt nicht dort, wo installiert wird,
// und ein Stil, der nur auf selbst angelegten Spielen ginge, verfehlte den
// Fall, für den es dieses Issue gibt.
func (v *Paketverwalter) spieldatei(spiel string) ([]byte, error) {
	for _, wurzel := range v.wurzeln {
		pfad := filepath.Join(wurzel, spiel+".mse-game", "game")

		// **Mit dem Budget einer Paketdatei** (Befund 4 des Zweitblicks über
		// Phase 6.1): ein Spielpaket kommt von außen. `spiel` ist durch
		// [istPaketname] gegangen — keine Schrägstriche, keine Punkte, kein
		// `..` —, und die Wurzeln sind die eigenen.
		roh, err := msepaket.LeseDateiBegrenzt(pfad)
		if err == nil {
			return roh, nil
		}

		// **Zu groß ist nicht „nicht da"**: wer weitersuchte, meldete am Ende
		// ein fehlendes Spiel, obwohl es daliegt.
		if errors.Is(err, msepaket.ErrZuGross) {
			return nil, err
		}
	}

	return nil, domain.NewFieldError("spiel", domain.SchluesselNichtGefunden,
		"Das Spiel %q liegt in keiner Paketwurzel; ein Stil darauf hätte "+
			"nichts, wonach er sich richten könnte.", spiel)
}

// heute gibt das Datum, in der Form, die MSE2 als Fassung liest.
//
// **Ein Datum und keine Zählnummer**: so schreiben es die Pakete dieses
// Projekts (`saat/probe.mse-game`) und die bezogenen. `Version::fromString`
// liest beides, aber ein Datum sagt einem Menschen, wie alt eine Vorlage ist.
func heute() string { return time.Now().Format("2006-01-02") }

// stilverzeichnis setzt den Verzeichnisnamen eines Stilpakets zusammen.
//
// # Warum das Spiel vorn steht
//
// **So heißen alle Stilpakete.** MSE selbst filtert seine Auswahlliste danach
// (`game->name() + "-*"`, `src/gui/new_window.cpp`, Zeile 65, Fassung f9d9356
// vom 2021-07-06), und `magic-set-generator` schreibt `stylesheet:` als
// **Kurzform** — das Paket heißt `<spiel>-<stil>.mse-style` (§11). Auch
// [pruefeStil] misst daran, ob ein Stil zu einem Set passt.
//
// # Warum beide Schreibweisen gehen
//
// **Weil der Name, den man eingibt, nicht der ist, den man danach braucht.**
// Wer `schlicht` anlegt, bekommt `eigenes-schlicht` — und muss genau das beim
// Anlegen eines Sets angeben. Der erste Anlauf nahm nur das Suffix, und der
// Flächenlauf ist prompt darüber gestolpert: angelegt `schlicht`, verlangt
// `eigenes-schlicht`.
//
// Ein zweites Voranstellen wäre `eigenes-eigenes-schlicht` — deshalb wird
// geprüft statt gerechnet.
func stilverzeichnis(spiel, name string) string {
	if strings.HasPrefix(name, spiel+"-") {
		return name
	}

	return spiel + "-" + name
}
