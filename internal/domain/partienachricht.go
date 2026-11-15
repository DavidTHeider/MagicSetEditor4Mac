package domain

// **Was über die Leitung einer Partie geht** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Vier Arten, und sie sind nicht dasselbe
//
// Das Issue verlangt die Trennung wörtlich: „bestätigte Events,
// Kontrollnachrichten, Chat und flüchtige Präsenz getrennt … keine
// Persistenz, Hashkette oder Replayaufnahme" — der letzte Halbsatz gilt der
// Präsenz.
//
// **Wer sie in eine Gestalt zwänge, bekäme an jeder Stelle Felder, die die
// Hälfte der Zeit nichts bedeuten** — und ein Mauszeiger läge in der
// Hashkette. Was eine Art tut, steht deshalb an **ihr** und nicht beim
// Aufrufer: `WirdAufbewahrt` und `GehtInDieKette` sind Eigenschaften der Art,
// keine Entscheidungen der Leitung.
//
// # Und warum das hier steht und nicht in `service`
//
// **Weil es eine fachliche Aussage ist.** Ob ein Mauszeiger aufbewahrt wird,
// hat mit Netzwerk nichts zu tun — es hat damit zu tun, was eine Partie ist.
// Dieselbe Überlegung wie bei `Zweck` eine Datei weiter.

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Nachrichtenart sagt, was eine Nachricht ist.
type Nachrichtenart string

// Die vier Arten, die über die Leitung einer Partie gehen.
const (
	// ArtEreignis ist ein **bestätigter** Spielzug.
	//
	// **Nur diese Art geht in die Hashkette.** Sie ist das, was eine Partie
	// ausmacht; alles andere begleitet sie.
	ArtEreignis Nachrichtenart = "ereignis"

	// ArtKontrolle ist Beitritt, Bereitschaft, Pause, Abbruch.
	//
	// **Dauerhaft, aber nicht Teil des Spielstands**: wer wann beigetreten
	// ist, gehört ins Protokoll — es ändert die Karten auf dem Tisch nicht.
	ArtKontrolle Nachrichtenart = "kontrolle"

	// ArtChat ist, was Menschen einander sagen.
	//
	// **Dauerhaft und mit Sichtrechten** (#1462: „Replay/Chat/Logs eines
	// Teilnehmers enthalten nur seine freigegebenen Informationen").
	ArtChat Nachrichtenart = "chat"

	// ArtPraesenz ist Mauszeiger, Ziehen, Schweben, ein Ping.
	//
	// **Flüchtig**: keine Folge, kein Kettenglied, keine Aufbewahrung. Sie
	// verschwindet, wenn eine Verbindung abreißt, und das ist richtig so —
	// ein Mauszeiger, der nach einem Verbindungsabbruch stehenbliebe, zeigte
	// auf jemanden, der nicht mehr da ist.
	ArtPraesenz Nachrichtenart = "praesenz"
)

// Nachrichtenarten sind die vier, in fester Reihenfolge.
func Nachrichtenarten() []Nachrichtenart {
	return []Nachrichtenart{ArtEreignis, ArtKontrolle, ArtChat, ArtPraesenz}
}

// Pruefe sagt, ob das eine der vier ist.
func (a Nachrichtenart) Pruefe() error {
	for _, bekannt := range Nachrichtenarten() {
		if a == bekannt {
			return nil
		}
	}

	return fmt.Errorf("%w: %q ist keine Nachrichtenart einer Partie",
		ErrValidation, string(a))
}

// WirdAufbewahrt sagt, ob eine Nachricht dieser Art ins Protokoll gehört.
func (a Nachrichtenart) WirdAufbewahrt() bool {
	return a != ArtPraesenz
}

// GehtInDieKette sagt, ob eine Nachricht dieser Art den Spielstand fortschreibt.
//
// **Nur das Ereignis.** Wer Chat oder Kontrolle mit hineinnähme, machte aus
// „hallo" einen Spielzug — und zwei Teilnehmer, von denen einer den Chat
// abgeschaltet hat, bekämen verschiedene Ketten.
func (a Nachrichtenart) GehtInDieKette() bool {
	return a == ArtEreignis
}

// Kettengliedlaenge ist die Länge eines Kettenglieds in Zeichen.
//
// **Ein SHA-256 in Hex** — dieselbe Kette wie in G1
// (#1458).
const Kettengliedlaenge = 64

// Partienachricht ist, was über die Leitung geht.
type Partienachricht struct {
	Art Nachrichtenart

	// Partie ist die Kennung der Partie, zu der sie gehört.
	Partie string

	// Epoche zählt, wie oft der Wirt neu angefangen hat.
	//
	// **Sie trennt zwei Läufe derselben Partie** (#1462: „Resume mit
	// Session-ID, Epoch, Sequence"): eine Nachricht aus einer alten Epoche
	// gehört zu einem Stand, den es nicht mehr gibt, und wird nicht etwa
	// nachgetragen — sie wird abgewiesen.
	Epoche uint64

	// Folge zählt innerhalb einer Epoche.
	//
	// **Null bei Präsenz** — sie wird nicht aufbewahrt und hat deshalb keine
	// Stelle in einer Reihenfolge.
	Folge uint64

	// Von ist die Sitzkennung des Absenders.
	Von string

	// Inhalt ist die Nutzlast — ihre Gestalt hängt an der Art.
	//
	// **Als `json.RawMessage` und nicht als `[]byte`**: sonst ginge er als
	// Base64 über die Leitung, und wer mitliest, sähe statt eines Zuges eine
	// Zeichenkette. So bleibt eine Partienachricht von außen lesbar — was
	// beim Nachstellen eines Fehlers den Unterschied macht.
	Inhalt json.RawMessage

	// Kette ist das Glied, das dieses Ereignis an das vorige bindet.
	//
	// **Nur bei einem Ereignis**, und dort verbindlich.
	Kette string
}

// Pruefe hält eine Nachricht gegen ihre Art.
func (n Partienachricht) Pruefe() error {
	if fehler := n.Art.Pruefe(); fehler != nil {
		return fehler
	}

	switch {
	case strings.TrimSpace(n.Partie) == "":
		return fmt.Errorf("%w: eine Nachricht ohne Partie gehört nirgendwohin",
			ErrValidation)

	case strings.TrimSpace(n.Von) == "":
		return fmt.Errorf("%w: eine Nachricht ohne Absender lässt sich nicht "+
			"zuordnen", ErrValidation)

	case n.Epoche == 0:
		return fmt.Errorf("%w: eine Nachricht ohne Epoche gehört zu keinem Lauf",
			ErrValidation)
	}

	return n.pruefeArt()
}

// pruefeArt hält Folge und Kettenglied gegen die Art.
func (n Partienachricht) pruefeArt() error {
	if !n.Art.WirdAufbewahrt() {
		return n.pruefeFluechtig()
	}

	if n.Folge == 0 {
		return fmt.Errorf("%w: eine aufbewahrte Nachricht braucht eine Folge — "+
			"sonst hat sie keine Stelle in der Reihenfolge", ErrValidation)
	}

	if !n.Art.GehtInDieKette() {
		if n.Kette != "" {
			return fmt.Errorf("%w: %q gehört nicht in die Hashkette und trägt "+
				"deshalb kein Kettenglied", ErrValidation, string(n.Art))
		}

		return nil
	}

	if len(n.Kette) != Kettengliedlaenge {
		return fmt.Errorf("%w: ein Ereignis braucht ein Kettenglied von %d "+
			"Zeichen, hier sind es %d", ErrValidation, Kettengliedlaenge,
			len(n.Kette))
	}

	return nil
}

// pruefeFluechtig hält fest, dass Präsenz nichts mitschleppt.
func (n Partienachricht) pruefeFluechtig() error {
	switch {
	case n.Folge != 0:
		return fmt.Errorf("%w: flüchtige Präsenz hat keine Stelle in einer "+
			"Reihenfolge", ErrValidation)

	case n.Kette != "":
		return fmt.Errorf("%w: flüchtige Präsenz gehört nicht in die Hashkette",
			ErrValidation)
	}

	return nil
}
