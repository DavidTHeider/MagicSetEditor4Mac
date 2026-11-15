package domain

// **Wie ein Gast zurückkommt** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Drei Wege und eine Sperre
//
// Das Issue nennt sie: „kleine Lücke über Events, große über autorisierten
// Snapshot plus Folgestrom. Doppelte/geänderte IDs, Lücken, falsche Hashes
// und alte Epochs blockieren statt weiterzuspielen."
//
// # Warum die Entscheidung hier steht und nicht in der Leitung
//
// **Weil sie auf beiden Seiten dieselbe sein muss.** Ein Gast rechnet aus,
// was er braucht; ein Wirt rechnet aus, was er schickt — und wenn beide zu
// verschiedenen Ergebnissen kämen, wartete der eine auf etwas, das der andere
// nie sendet. Eine Rechnung, zwei Aufrufer.
//
// # Und warum die Grenze nicht eine Zahl ist
//
// **„Klein" und „groß" hängen daran, was der Wirt noch hat.** Wer sie an eine
// feste Zahl bände — „bis fünfzig Ereignisse ist es klein" —, bekäme bei
// einem kurzen Protokoll eine Lücke, die er für klein hält, und der Gast
// liefe in ein Loch. Maßgeblich ist die **älteste** Folge, die noch vorliegt.

import "fmt"

// Wiederaufnahmewunsch sagt, wo ein Gast stehengeblieben ist.
type Wiederaufnahmewunsch struct {
	Partie string

	// Epoche und Folge sind sein letzter Stand — beide null bei einem, der
	// noch nie dabei war.
	Epoche uint64
	Folge  uint64

	// Kette ist das letzte Glied, das er gesehen hat.
	//
	// **Es ist der Beleg, dass seine Geschichte dieselbe ist.** Zwei Stände
	// können dieselbe Folge tragen und verschiedene Züge enthalten; die
	// Zahlen merken es nicht, die Kette schon.
	Kette string

	// Mitgebracht ist das Deck, mit dem sich ein eingeladener Gast setzt
	// (P9b, #2377) — **nur beim ersten Mal**, und ein Wirt vor 0.9.0 liest
	// das Feld nicht.
	Mitgebracht *Mitgebracht `json:",omitempty"`

	// Kann sagt, was dieser Gast über das Spielen hinaus beantwortet — etwa
	// [FaehigkeitBild] (Nachlauf zum Zweitblick über Phase 7, R4). **In jedem
	// Wunsch**, denn ein Wiederhinsetzen ist eine neue Sitzung; ein Gast vor
	// 0.9.0 schickt das Feld nicht, und ein Wirt vor 0.9.0 liest es nicht.
	Kann []string `json:",omitempty"`
}

// FaehigkeitBild erklärt, dass ein Gast Bildwünsche beantwortet (P9b).
const FaehigkeitBild = "bild"

// Wirtsstand ist, was der Wirt hat.
type Wirtsstand struct {
	Epoche uint64
	Folge  uint64
	Kette  string

	// AelteteFolge ist die älteste, die noch vorliegt.
	//
	// **Sie zieht die Grenze zwischen klein und groß.** Wer weiter zurückliegt,
	// bekommt einen Abzug — nicht weil die Lücke eine Zahl überschreitet,
	// sondern weil die Ereignisse dazu nicht mehr da sind.
	AelteteFolge uint64
}

// Wiederaufnahmeart ist einer der vier Ausgänge.
type Wiederaufnahmeart int

const (
	// WegWeiter — es fehlt nichts.
	WegWeiter Wiederaufnahmeart = iota

	// WegEreignisse — die Lücke wird nachgereicht.
	WegEreignisse

	// WegAbzug — der Stand wird ganz geschickt, danach der Folgestrom.
	WegAbzug

	// WegAbbruch — hier wird nicht weitergespielt.
	WegAbbruch
)

// String gibt der Art einen Namen.
func (a Wiederaufnahmeart) String() string {
	switch a {
	case WegWeiter:
		return "weiter"
	case WegEreignisse:
		return "Ereignisse"
	case WegAbzug:
		return "Abzug"
	case WegAbbruch:
		return "Abbruch"
	}

	return "unbekannt"
}

// Wiederaufnahmeweg ist die Antwort samt ihrem Grund.
type Wiederaufnahmeweg struct {
	Art Wiederaufnahmeart

	// AbFolge ist die erste Folge, die nachgereicht wird — nur bei
	// [WegEreignisse].
	AbFolge uint64

	// Grund sagt, warum. **Auch beim Weiterspielen**: ein Protokoll, in dem
	// nur die Abbrüche einen Satz tragen, erzählt die halbe Geschichte.
	Grund string
}

// Wiederaufnahme entscheidet, wie ein Gast zurückkommt.
func Wiederaufnahme(
	wunsch Wiederaufnahmewunsch, stand Wirtsstand,
) Wiederaufnahmeweg {
	if fehler := pruefeEpoche(wunsch, stand); fehler != "" {
		return Wiederaufnahmeweg{Art: WegAbbruch, Grund: fehler}
	}

	switch {
	// **Ein Gast, der weiter ist als der Wirt, hat mit jemand anderem
	// gespielt** — oder mit einem Wirt, der seinen Stand verloren hat. Beides
	// ist kein Fall zum Weiterspielen.
	case wunsch.Folge > stand.Folge:
		return Wiederaufnahmeweg{
			Art: WegAbbruch,
			Grund: fmt.Sprintf(
				"der Gast steht bei Folge %d, der Wirt bei %d — das ist nicht "+
					"dieselbe Partie", wunsch.Folge, stand.Folge),
		}

	case wunsch.Folge == stand.Folge:
		return gleicherStand(wunsch, stand)
	}

	// **Jetzt fehlt etwas.** Ob es noch vorliegt, sagt die älteste Folge.
	if wunsch.Folge+1 < stand.AelteteFolge {
		return Wiederaufnahmeweg{
			Art: WegAbzug,
			Grund: fmt.Sprintf(
				"ab Folge %d liegt nichts mehr vor, der Gast steht bei %d",
				stand.AelteteFolge, wunsch.Folge),
		}
	}

	return Wiederaufnahmeweg{
		Art:     WegEreignisse,
		AbFolge: wunsch.Folge + 1,
		Grund: fmt.Sprintf("es fehlen die Folgen %d bis %d",
			wunsch.Folge+1, stand.Folge),
	}
}

// gleicherStand entscheidet, wo die Zahlen übereinstimmen.
//
// **Hier und nur hier trägt die Kette.** Der Wirt hält genau dieses Glied;
// weicht es ab, sind zwei Geschichten bei derselben Zahl angekommen.
func gleicherStand(
	wunsch Wiederaufnahmewunsch, stand Wirtsstand,
) Wiederaufnahmeweg {
	if wunsch.Kette != stand.Kette {
		return Wiederaufnahmeweg{
			Art: WegAbbruch,
			Grund: "bei derselben Folge trägt der Gast ein anderes Kettenglied " +
				"als der Wirt — das sind zwei verschiedene Verläufe",
		}
	}

	return Wiederaufnahmeweg{Art: WegWeiter, Grund: "der Gast ist auf Stand"}
}

// pruefeEpoche gibt einen Grund, wenn die Epoche nicht passt — sonst leer.
func pruefeEpoche(wunsch Wiederaufnahmewunsch, stand Wirtsstand) string {
	// **Ein Neuling hat keine Epoche.** Er war noch nie dabei; für ihn
	// entscheidet allein, ob es etwas nachzureichen gibt.
	if wunsch.Epoche == 0 && wunsch.Folge == 0 {
		return ""
	}

	switch {
	case wunsch.Epoche < stand.Epoche:
		return fmt.Sprintf(
			"der Gast kommt aus Epoche %d, der Wirt ist in %d — dieser Lauf "+
				"ist vorbei", wunsch.Epoche, stand.Epoche)

	case wunsch.Epoche > stand.Epoche:
		return fmt.Sprintf(
			"der Gast nennt Epoche %d, der Wirt ist erst in %d — nur der Wirt "+
				"zählt sie hoch", wunsch.Epoche, stand.Epoche)
	}

	return ""
}
