package render

import (
	"image"
	"strconv"
	"strings"
)

// ausBaum führt einen [Bildbaum] aus und gibt das fertige Bild.
//
// # Was heute ausgeführt wird
//
// **Das Blatt.** Eine Verknüpfung wird **benannt** und ihr Ausgangsbild
// gezeichnet — nicht verschwiegen und nicht weggelassen (§8).
//
// # Warum ein Ausgangsbild und keine leere Fläche
//
// Eine Verknüpfung wie `recolor_image` färbt ein vorhandenes Bild um. Wer sie
// nicht kann, hat die Wahl zwischen **nichts** und **dem Bild in seiner
// Ausgangsfarbe**. Das Zweite ist näher an dem, was gemeint war, und es ist
// sichtbar falsch statt unsichtbar fehlend — ein Rahmen in der falschen Farbe
// fällt jedem auf, eine fehlende Ebene niemandem.
//
// **Der Befund geht trotzdem mit.** Wer die Zahl der Befunde als Maß nimmt,
// soll nicht dadurch besser dastehen, dass etwas halb getan wurde.
//
// # Warum die Verknüpfungen noch nicht gebaut sind
//
// Sie brauchen je eine eigene Zusage samt Goldbildvergleich (§3), und die fünf
// häufigsten machen 96 Prozent aus
// (#270). Der
// Baum ist die Voraussetzung dafür; sie einzeln zu bauen ist die Arbeit
// danach.
func (r *Rasterer) ausBaum(baum Bildbaum, gefragt image.Point) (quelle image.Image, befund string) {
	quelle, maengel := r.ausBaumMaengel(baum, gefragt)

	return quelle, strings.Join(maengel, "; ")
}

// ausBaumMaengel führt einen Bildbaum aus wie [Rasterer.ausBaum] und gibt
// **jeden Befund einzeln** — für eine Ebene, deren Mängel je ein Befund mit
// eigener Art werden (QA-03, #1656). Die zusammengefügte Form bleibt für die
// Wege, die eine einzige Meldung weiterreichen.
func (r *Rasterer) ausBaumMaengel(baum Bildbaum, gefragt image.Point) (quelle image.Image, maengel []string) {
	// **Ein Lauf bekommt seine eigene Kopie** —
	// #748.
	//
	// Die gefragte Größe und die beiseitegelegten Befunde gehören zu **diesem**
	// Bildbaum. Am Rasterer lägen sie falsch, denn der ist geteilt:
	// `internal/app` reicht **einen** an Vorschau, Symbolpalette, Bildausfuhr,
	// Druckbogen und Vorlagenausfuhr, und zwei Ausfuhren verschiedener Sets
	// laufen ausdrücklich nebeneinander (#741).
	//
	// **Eine Kopie und keine Sperre**: eine Sperre nähme der Sache die Wirkung,
	// nicht die Gestalt — und sie nähme genau die Parallelität zurück, die #741
	// hergestellt hat. Dieselbe Antwort wie dort, eine Ebene tiefer.
	//
	// **Der Ebenenspeicher wird nicht mitkopiert**: er ist ein Zeiger und
	// bleibt geteilt, mit eigener Sperre (`ebenenspeicher.go`).
	if r.gemeldeteGroesse != nil {
		r.gemeldeteGroesse(gefragt)
	}

	lauf := *r
	lauf.gefragteGroesse = gefragt
	lauf.baumbefunde = nil

	// **Die gefragte Größe reist mit** — MSE gibt sie als `Options` an jeden
	// Erzeuger weiter, und `BlankImage` macht daraus sein durchsichtiges Bild
	// (#509).
	bild, eigener := lauf.ausBaumTief(baum, 0)

	return bild, mit(lauf.baumbefunde, eigener)
}

// Baumtiefe ist die Grenze, ab der ein Bildbaum als zyklisch gilt.
//
// # Warum es eine Grenze braucht
//
// Ein [Bildbaum] ist ein Werttyp, und trotzdem lässt sich über eine
// rückverweisende Slice ein Selbstkind bauen. Die Werte aus `msescript` sind
// heute azyklisch, aber **der Zeichenplan erzwingt das weder technisch noch
// vertraglich** — und ein Plan kommt aus einem fremden Paket.
//
// Angestrichen vom Zweitblick über die Phasen B bis D: „rekursive Daten werden
// ohne Zyklus- oder Tiefenschutz an einer Paketgrenze übernommen."
//
// **Die Zahl ist eine Grenze, keine Herleitung.** Aus der Zahl der
// Verknüpfungen folgt keine Tiefe — dieselbe Verknüpfung darf beliebig oft
// vorkommen.
//
// Was bleibt, ist eine **Betriebsgrenze**: der tiefste Baum in
// `M15-Magic-Pack` ist vier Verknüpfungen tief (`new-blends`, vierfarbige
// Hybride). Vierzig lässt jedem denkbaren Paket Luft und fängt einen Ring
// trotzdem, bevor der Stapel überläuft. **Es ist keine Zykluserkennung**, und
// das steht hier, damit niemand sie dafür hält
// (#318).
//
// # Der Vertrag gilt für jeden Abläufer
//
// Nicht nur für [Rasterer.ausBaumTief]: [Bildbaum.Beschreibung],
// [Bildbaum.Gebaut] und [Bildbaum.Blaetter] zählen dieselbe Tiefe mit. Sie
// taten es bis #318 **nicht** und liefen an demselben Ring bis zum
// Stapelüberlauf — der beendet den ganzen Lauf und sieht nach keinem Befund
// aus.
//
// Wer einen weiteren Abläufer schreibt, hält ihn ein. Die Gegenprobe steht in
// `tiefenvertrag_test.go`: sie baut den Ring über die geteilte Slice und prüft,
// dass jeder der drei **zurückkehrt**.
const Baumtiefe = 40

// ausBaumTief führt einen Baum aus und zählt dabei die Tiefe mit.
//
// # Ein Befund geht hier nicht mehr verloren
//
// Jede Einbild-Umformung holt ihr Kind so:
//
//	unten, tiefer := r.ausBaumTief(baum.Kinder[0].Baum, tiefe+1)
//	if unten == nil {
//	    return nil, tiefer
//	}
//
// `tiefer` wurde damit **nur im Fehlschlag** weitergereicht. Ein Kind, das ein
// Bild *und* einen Befund liefert — `recolor_image` ohne Farbe gibt sein
// Ausgangsbild zurück und meldet den Mangel —, verlor seinen Befund an jedem
// Elternteil, das gelang
// (#473).
//
// **Die Antwort steht hier und nicht in den zehn Aufrufern.** Kommt aus einem
// Knoten ein Bild **und** ein Befund, wird der Befund beiseitegelegt und der
// Rückgabewert geleert; [Rasterer.ausBaumMaengel] sammelt sie am Ende ein.
// Damit kann kein Elternteil ihn mehr verlieren — auch keines, das noch
// niemand geschrieben hat.
//
// **Warum nicht zehnmal die Meldung des Kindes an die eigene hängen**: das
// wäre der Fall gewesen, nicht die Klasse. Die elfte Umformung hätte den Fehler wieder gehabt, und
// nichts hätte es gemeldet (§10).
func (r *Rasterer) ausBaumTief(baum Bildbaum, tiefe int) (quelle image.Image, befund string) {
	bild, eigener := r.ausBaumRoh(baum, tiefe)

	if bild != nil && eigener != "" {
		r.baumbefunde = append(r.baumbefunde, eigener)

		return bild, ""
	}

	return bild, eigener
}

// ausBaumRoh ist der Verteiler über die Knotenarten.
func (r *Rasterer) ausBaumRoh(baum Bildbaum, tiefe int) (quelle image.Image, befund string) {
	if tiefe > Baumtiefe {
		return nil, "der Bildbaum ist tiefer als " + strconv.Itoa(Baumtiefe) +
			" Verknüpfungen — kein Paket baut so tief, und ein Ring sähe genauso aus"
	}

	if baum.Leer() {
		return nil, "die Ebene nennt kein Bild"
	}

	if baum.IstBlatt() {
		blatt, err := r.bilddaten(baum.Pfad)
		if err != nil {
			return nil, "das Bild " + baum.Pfad + " ist nicht lesbar: " + err.Error()
		}

		return blatt, ""
	}

	if baum.Knotenart == "symbol_variation" {
		return r.symbolbild(baum)
	}

	// **`blank_image` hat berechtigt kein Kind** — es *ist* der Erzeuger, so
	// wie `datei` und `symbol_variation`
	// (#509).
	if baum.Knotenart == "blank_image" {
		return r.leeresBild(), ""
	}

	// **Und `built_in_image` ebenso** — sein Name *ist* die Quelle.
	if baum.Knotenart == "built_in_image" {
		return r.eingebautesBild(baum)
	}

	// **Und das Schachbrett eines leeren Bildfelds** — es braucht nur seine
	// Größe ([Platzhalterknoten]).
	if baum.Knotenart == Platzhalterknoten {
		return r.platzhalter(), ""
	}

	if len(baum.Kinder) == 0 {
		return nil, "die Verknüpfung " + baum.Knotenart + " hat kein Ausgangsbild"
	}

	// **Die Liste in `bildbaum.go` trägt die Ausführung selbst** — seit dem
	// 2026-09-03; warum, steht dort.
	// **Und `fuehreAus` wird geprüft, nicht vorausgesetzt.** Die vier
	// Erzeuger stehen ohne Ausführung in der Liste und werden oben
	// abgefangen; erreichte einer diese Stelle — weil jemand seinen Zweig
	// entfernt —, wäre ein Nil-Aufruf eine Panik statt eines Befunds (§8).
	if eintrag, gebaut := gebauteVerknuepfungen[baum.Knotenart]; gebaut &&
		eintrag.fuehreAus != nil {
		return eintrag.fuehreAus(r, baum, tiefe)
	}

	unten, tiefer := r.ausBaumTief(baum.Kinder[0].Baum, tiefe+1)
	if unten == nil {
		return nil, tiefer
	}

	return unten, "die Verknüpfung " + baum.Knotenart +
		" ist noch nicht gebaut; gezeichnet wird das Ausgangsbild " +
		baum.Kinder[0].Baum.Beschreibung()
}
