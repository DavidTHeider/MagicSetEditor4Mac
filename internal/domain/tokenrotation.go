// **Rotation, Übergangsfrist und Ablaufwahlen** — Werkstück O2 von Phase 5.2
// (#1426),
// nach [Vermerk 0080](../../docs/entscheidungen/0080-maschinenzugaenge-und-widerrufbare-sitzungen.md).
//
// # Was Rotation hier heißt
//
// **Ein neues Geheimnis am selben Token**, und nicht ein zweites Token. Der
// Unterschied ist die Buchführung: wer ein neues anlegt und das alte
// zurückzieht, hat zwischen beiden Schritten entweder eine Lücke oder zwei
// gültige Zugänge — und wer den zweiten Schritt vergisst, hat sie für immer.
// Ein Token, das sein Geheimnis wechselt, behält seine Kennung, seine
// Bezeichnung und seinen Platz in jeder Liste.

package domain

import (
	"strings"
	"time"
)

// Tokenuebergang ist die Frist, in der das **vorige** Geheimnis nach einer
// Rotation noch angenommen wird.
//
// # Warum eine Stunde
//
// **Weil die Frist zwei Dinge zugleich sein muss und keines davon lange
// dauert.** Sie soll einen laufenden Auftrag nicht mitten im Lauf aussperren
// — und das Längste, was diese Anwendung in einem Zug tut (ein Massenlauf,
// eine Bestandssicherung), dauert Minuten. Und sie soll dem Menschen reichen,
// der das neue Geheimnis in die Einrichtung seines Gegenübers trägt und den
// Prozess neu startet; das ist Handarbeit von wenigen Minuten, mit
// reichlichem Zuschlag für eine Unterbrechung dazwischen.
//
// **Länger wäre keine Übergangsfrist, sondern ein zweiter Zugang.** Rotiert
// wird, weil das alte Geheimnis weg soll; jede weitere Stunde ist Zeit, in
// der es noch geht.
//
// **Und die Stunde ist eine Obergrenze, kein Versprechen.** Wer noch einmal
// rotiert, wirft den Vorgänger sofort weg (es gibt genau einen), und wer das
// Token zurückzieht, sperrt beide auf der Stelle.
const Tokenuebergang = time.Hour

// VorgaengerGilt sagt, ob das vorige Geheimnis zum genannten Zeitpunkt noch
// angenommen wird.
//
// **Ein unlesbarer Wert gilt nicht** — und das ist die andere Richtung als
// bei [Zugriffstoken.Abgelaufen], mit demselben Grund: die sichere Auslegung
// ist die, die keinen Zugang gewährt.
func (z Zugriffstoken) VorgaengerGilt(jetzt time.Time) bool {
	if strings.TrimSpace(z.VorgaengerBis) == "" {
		return false
	}

	bis, err := time.Parse(time.RFC3339, z.VorgaengerBis)
	if err != nil {
		return false
	}

	return jetzt.Before(bis)
}

// Ablaufwahl ist ein benannter Vorschlag für den Ablauf eines Tokens.
//
// **Der Ablauf ist ausgerechnet und kommt fertig heraus.** Sonst rechnete
// jede der vier Flächen dieselbe Spanne aus, und die Ergebnisse liefen an der
// Zeitzone auseinander — dieselbe Überlegung wie bei `TokenDTO.Abgelaufen`.
type Ablaufwahl struct {
	// Name ist der Schlüssel — die Oberfläche schlägt ihren Text darunter
	// nach, die Kommandozeile gibt ihn aus.
	Name string

	// Ablauf ist der Zeitpunkt in RFC 3339; leer heißt „läuft nicht ab".
	Ablauf string
}

// Ablaufwahlen gibt die vier Vorschläge, gerechnet aus der übergebenen Uhr.
//
// # Warum diese vier
//
//   - **Ein Tag** für den einmaligen Auftrag und für das Ausprobieren. Ein
//     Zugang, den jemand heute braucht, soll den heutigen Tag nicht
//     überleben.
//   - **Dreißig Tage** für die wiederkehrende Automatisierung. Ein Monat ist
//     der Takt, in dem jemand seine Einrichtung ohnehin wieder ansieht — und
//     die längste Spanne, über die er noch weiß, wofür ein Token da war.
//   - **Ein Jahr** für den Zugang, der zur Installation gehört und nicht zu
//     einer Aufgabe. Er soll nicht monatlich stören, aber ein vergessenes
//     Token soll innerhalb eines Jahres von selbst stumpf werden.
//   - **Nie** für den Zugang auf diesem Rechner, dessen Lebensdauer die der
//     Installation ist. Das Modell kann es (`ablauf` leer), die
//     Kommandozeile kann es — es am Fenster zu verschweigen wäre eine
//     Unwahrheit über die Flächen (§16) und triebe niemanden zu kürzeren
//     Fristen, sondern nur an die Kommandozeile.
//
// **Kalendarisch gerechnet und nicht in Stunden** ([time.Time.AddDate]): ein
// Jahr sind nicht 365 Tage, wenn ein Schalttag dazwischenliegt, und ein
// Ablauf, der um einen Tag danebenliegt, ist an einem Monatsende ärgerlich.
func Ablaufwahlen(jetzt time.Time) []Ablaufwahl {
	inUTC := jetzt.UTC()

	return []Ablaufwahl{
		{Name: "tag", Ablauf: inUTC.AddDate(0, 0, 1).Format(time.RFC3339)},
		{Name: "monat", Ablauf: inUTC.AddDate(0, 0, 30).Format(time.RFC3339)},
		{Name: "jahr", Ablauf: inUTC.AddDate(1, 0, 0).Format(time.RFC3339)},
		{Name: "nie", Ablauf: ""},
	}
}

// Gehuetetes sind die gehüteten Formen eines Tokens — die geltende und die
// des Vorgängers aus einer Rotation.
//
// **Kein Klartext.** Was hier steht, ist die Ausgabe der Hutfunktion; damit
// meldet sich niemand an. Die Gestalt steht in `domain`, weil sie zwischen
// `service` (der prüft) und `storage` (der ablegt) reist — und `storage` darf
// `service` nicht kennen (§2).
type Gehuetetes struct {
	// Aktuell ist die Form, die immer gilt.
	Aktuell string

	// Vorgaenger ist die des vorigen Geheimnisses — leer, wenn nie rotiert
	// wurde. Ob sie noch zählt, sagt [Zugriffstoken.VorgaengerGilt].
	Vorgaenger string
}
