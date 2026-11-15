package domain

// **Was ein Gegenüber bekommt, um über ein fremdes Netz zu wählen** —
// Werkstück N3 von Phase 5.4
// (#1456).
//
// # Warum die Einladung eine Fassung trägt
//
// **Weil das Drahtformat unter uns sich ändern darf.** tailcat sagt es selbst:
// „types, functions, and the wire format may all change". Bricht es, kommt gar
// keine Verbindung zustande — der signierte Handschlag aus N1 prüft **seine**
// Fassung und weist ab, aber er läuft nie, weil darunter nichts zustande kam.
// Ein Mensch läse „nicht erreichbar", und das wäre ein Verbindungsfehler, der
// in Wahrheit ein Fassungsfehler ist: die stille Form, vor der §8 warnt.
//
// **Also trägt die Einladung die Fassung**, und sie wird geprüft, **bevor**
// irgendetwas ins Netz geht.
//
// # Was die Zahl sagt und was nicht
//
// **Sie sagt einzig: diese beiden Fassungen können einander erreichen.** Sie
// ist keine zweite Protokollfassung neben der des Handschlags — der bleibt
// zuständig für alles, was **über** der Leitung liegt.
//
// **Sie steigt, wenn tailcats Drahtformat bricht, und nur dann.** Wer tailcat
// hebt, sieht in dessen Änderungsprotokoll nach und hebt sie mit. Ein Wächter
// kann das nicht — er wüsste nicht, was „bricht" heißt.
//
// **Und sie schützt vor dem bekannten Bruch, nicht vor dem unbemerkten.** Eine
// Einladung derselben Zahl, deren tailcat sich trotzdem anders verhält, endet
// weiter in einem Verbindungsfehler. Das steht hier, weil eine Zusage ohne
// ihre Grenze schlimmer ist als keine (§10).
//
// # Warum in `domain` und nicht im Blatt
//
// **Weil sie eine fachliche Regel ist und kein Transport.** Ob eine Einladung
// gilt, entscheidet dieses Programm — nicht die Bibliothek, die den Tunnel
// aufmacht. Und `service` prüft sie, bevor es ein Blatt überhaupt ruft.

import (
	"fmt"
	"strconv"
	"strings"
)

// Einladungsmarke steht am Anfang jeder Einladung.
//
// **Damit eine fremde Zeichenkette nicht versehentlich als Einladung
// durchgeht** — und damit ein Mensch, der sie in einer Nachricht sieht, weiß,
// wozu sie gehört.
const Einladungsmarke = "mse4mac"

// Einladungsfassung ist die Fassung, die dieses Programm ausgibt und annimmt.
const Einladungsfassung = 1

// Einladung ist, was ein Gegenüber zum Wählen braucht.
type Einladung string

// NeueEinladung setzt eine Adresse in die heutige Fassung.
func NeueEinladung(adresse string) Einladung {
	return Einladung(fmt.Sprintf(
		"%s:%d:%s", Einladungsmarke, Einladungsfassung, adresse))
}

// ZerlegeEinladung gibt die Adresse — oder sagt, warum nicht.
//
// **Am zweiten Doppelpunkt geteilt und nicht an jedem**: die Adresse darf
// selbst welche enthalten, und wer sie an allen teilte, verstümmelte sie.
func ZerlegeEinladung(einladung Einladung) (adresse string, fehler error) {
	teile := strings.SplitN(string(einladung), ":", 3)
	if len(teile) != 3 || teile[0] != Einladungsmarke {
		return "", fmt.Errorf(
			"%w: das ist keine Einladung dieses Programms", ErrValidation)
	}

	fassung, umwandlungsfehler := strconv.Atoi(teile[1])
	if umwandlungsfehler != nil || fassung < 1 {
		return "", fmt.Errorf(
			"%w: die Einladung nennt keine Fassung", ErrValidation)
	}

	if fassung > Einladungsfassung {
		return "", fmt.Errorf("%w (%d gegen %d)",
			ErrEinladungZuNeu, fassung, Einladungsfassung)
	}

	if teile[2] == "" {
		return "", fmt.Errorf(
			"%w: die Einladung nennt keine Adresse", ErrValidation)
	}

	return teile[2], nil
}

// IstEinladung sagt, ob eine Zeichenkette eine Einladung sein will.
//
// **„Sein will" und nicht „ist"**: geprüft wird die Marke, nicht die Fassung
// und nicht die Adresse. Wer danach zerlegt, erfährt den Rest — und zwar mit
// einem Satz, der sagt, was falsch ist.
//
// # Wofür
//
// **Damit ein Wähler weiß, welchen Weg er nimmt.** Im Adressfeld eines
// Kontakts steht entweder eine Adresse im eigenen Netz („192.168.1.7:41234")
// oder eine Einladung über ein fremdes. Beides ist dasselbe Feld, weil es
// dieselbe Frage beantwortet: **wo erreiche ich dieses Gegenüber?**
func IstEinladung(text string) bool {
	return strings.HasPrefix(text, Einladungsmarke+":")
}
