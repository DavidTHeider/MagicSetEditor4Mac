package peer

// **Was eine Hülle sein muss, und was unterschrieben wird.**
//
// # Warum die kanonische Form nicht die JSON-Form ist
//
// **Weil JSON mehrere Gestalten für dieselbe Nachricht hat** — Reihenfolge der
// Schlüssel, Leerraum, Zahlenschreibweise. Unterschrieben wird deshalb nicht
// das JSON, sondern eine Form, die es nur einmal gibt.

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die drei Fachprotokolle aus dem Issue.
//
// **Getrennt und nicht eine Fläche mit einem Feld darin.** Ein Peer, der nur
// Karten tauschen will, spricht `/mse4mac/share/1` — und muss dafür nicht
// verstehen, was eine Partie ist. Die Trennung ist die Grenze, an der später
// Rechte hängen (N1.4).
//
// **Die Fassung steht im Namen.** So legt libp2p Protokolle fest, und es hat
// denselben Grund wie hier: zwei Fassungen sind zwei Protokolle, und wer
// beide kann, bietet beide an — statt sich auf eine zu einigen.
const (
	ProtokollPaarung = "/mse4mac/pair/1"
	ProtokollTeilen  = "/mse4mac/share/1"
	ProtokollSpiel   = "/mse4mac/game/1"
)

const (
	// Hoechstlaenge ist die größte Nachricht, die durch diese Hülle geht.
	//
	// **Ein Megabyte, und der Grund ist der Zuschnitt.** Diese Hülle trägt
	// **Steuerung** — einen Handschlag, eine Anfrage, die Beschreibung
	// eines Objekts. Ein ganzes Set reist als Übertragung in Stücken (N4)
	// und nicht als eine Nachricht. Ein Megabyte lässt einer Karte samt
	// ihren Feldwerten reichlich Platz und hält zugleich die Zuteilung
	// eines Empfängers beschränkt, der noch niemandem vertraut.
	Hoechstlaenge = 1 << 20

	// Hoechstwort ist die Grenze für jedes Feld des Kopfes.
	//
	// **Eine Kennung wiegt 52 Zeichen**, eine Sitzungskennung weniger. 128
	// ist der Platz, den ein Kopf braucht, und nicht der, den er nehmen
	// könnte.
	Hoechstwort = 128

	// bereich trennt die Unterschrift einer Hülle von jeder anderen.
	//
	// **Damit derselbe Schlüssel zwei Dinge unterschreiben kann.** Der
	// Handschlag (N1.3) unterschreibt mit demselben Schlüssel; ohne einen
	// eigenen Bereich ließe sich eine Hülle als Handschlag ausgeben oder
	// umgekehrt — die Bytes wären dieselben, nur die Bedeutung eine andere.
	bereich = "mse4mac/huelle/1"
)

// dieProtokolle ist die Positivliste.
//
// **Eine Liste und keine Musterprüfung**: `/mse4mac/*/1` ließe ein viertes
// Protokoll entstehen, ohne dass jemand es entschieden hätte.
var dieProtokolle = map[string]struct{}{
	ProtokollPaarung: {},
	ProtokollTeilen:  {},
	ProtokollSpiel:   {},
}

// zuGross gibt den einen Befund für beide Grenzen.
func zuGross(gewicht int) error {
	return fmt.Errorf(
		"%w: die Nachricht wiegt %d Byte, erlaubt sind %d",
		domain.ErrValidation, gewicht, Hoechstlaenge)
}

// pruefeKopf hält die Hülle gegen das, was eine Hülle ausmacht.
//
// **Ohne die Unterschrift**: sie wird hier nicht angesehen, denn diese Prüfung
// läuft auch beim Verpacken — und dort gibt es sie noch nicht.
func (h Huelle) pruefeKopf() error {
	if _, gilt := dieProtokolle[h.Protokoll]; !gilt {
		return fmt.Errorf(
			"%w: %q ist keines der Protokolle dieses Programms",
			domain.ErrValidation, h.Protokoll)
	}

	// **Eine feste Reihenfolge und keine Map**: ein Befund, dessen Wortlaut
	// bei jedem Lauf ein anderer ist, taugt in keinem Protokoll.
	for _, feld := range []struct {
		name string
		wert string
	}{
		{"Art", h.Art},
		{"ID", h.ID},
		{"Sitzung", h.Sitzung},
		{"Absender", h.Absender},
	} {
		if feld.wert == "" {
			return fmt.Errorf(
				"%w: ohne %s ist eine Hülle keine", domain.ErrValidation,
				feld.name)
		}

		if len(feld.wert) > Hoechstwort {
			return fmt.Errorf(
				"%w: %s wiegt %d Zeichen, erlaubt sind %d",
				domain.ErrValidation, feld.name, len(feld.wert), Hoechstwort)
		}
	}

	if h.Folge < 0 {
		return fmt.Errorf(
			"%w: eine Folge zählt ab null, hier ist sie %d",
			domain.ErrValidation, h.Folge)
	}

	// **`null` ist kein Rumpf.** Es ist gültiges JSON, und genau deshalb
	// käme es an einer bloßen Längenprüfung vorbei.
	if len(h.Rumpf) == 0 || bytes.Equal(h.Rumpf, []byte("null")) {
		return fmt.Errorf(
			"%w: ohne Rumpf ist eine Hülle keine", domain.ErrValidation)
	}

	if len(h.Rumpf) > Hoechstlaenge {
		return zuGross(len(h.Rumpf))
	}

	// **Hier wird der Rumpf als JSON belegt, und nur hier.**
	//
	// Alles danach — das Verdichten, das Schreiben — kann an ungültigem JSON
	// scheitern und **nur** daran. Steht die Prüfung an dieser einen Stelle,
	// haben die übrigen keinen Fehlerweg mehr; ohne sie hätten sie je einen,
	// den kein Lauf betreten kann (§10).
	if !json.Valid(h.Rumpf) {
		return fmt.Errorf(
			"%w: der Rumpf ist kein JSON", domain.ErrValidation)
	}

	return nil
}

// unterschriftsform gibt die Bytes, die unterschrieben werden.
//
// # Warum vor jedem Teil seine Länge steht
//
// **Weil zwei verschiedene Köpfe sonst eine Unterschrift teilten.** Aneinander
// gehängt ergäben `Sitzung="ab", Art="c"` und `Sitzung="a", Art="bc"` denselben
// Text — und wer eine Nachricht abfängt, verschöbe die Grenze und behielte die
// Unterschrift. Mit `<Länge>:<Bytes>` je Teil gibt es diese zweite Lesart
// nicht: die Ziffern vor dem Doppelpunkt sagen, wo der Teil endet.
//
// **Und der Rumpf steht zuletzt und roh darin.** Er ist beim Unterschreiben
// schon verdichtet; ein zweites Kodieren wäre eine zweite Gestalt derselben
// Nachricht, und dann hinge die Unterschrift an einer Gestalt statt an einem
// Inhalt.
func (h Huelle) unterschriftsform() []byte {
	var werk bytes.Buffer

	for _, teil := range [][]byte{
		[]byte(bereich),
		[]byte(h.Protokoll),
		[]byte(h.Art),
		[]byte(h.ID),
		[]byte(h.Sitzung),
		[]byte(h.Absender),
		[]byte(fmt.Sprint(h.Folge)),
		h.Rumpf,
	} {
		fmt.Fprintf(&werk, "%d:", len(teil))
		werk.Write(teil)
	}

	return werk.Bytes()
}

// verdichtet gibt den Rumpf in seiner kanonischen Form.
//
// **Ohne Fehlerweg, weil es keinen gibt**: [Huelle.pruefeKopf] hat den Rumpf
// mit `json.Valid` schon als JSON belegt, und `json.Compact` scheitert nur an
// ungültigem JSON.
func verdichtet(roh []byte) []byte {
	var eng bytes.Buffer

	_ = json.Compact(&eng, roh) //nolint:errcheck // siehe oben: pruefeKopf hat ihn belegt

	return eng.Bytes()
}

// alsJSON schreibt die Hülle so, wie sie hinausgeht.
//
// **Ohne das Maskieren von `<`, `>` und `&`.** `json.Marshal` tut es von
// selbst, und es täte es **auch im Rumpf** — dann stünden im Rumpf andere
// Bytes als die, die unterschrieben wurden, und keine Nachricht mit einem
// spitzen Klammerzeichen prüfte je.
//
// **Und ohne Fehlerweg**, aus demselben Grund wie bei [verdichtet]: jedes Feld
// ist eine Zeichenkette oder eine Zahl, und der Rumpf ist als JSON belegt.
func (h Huelle) alsJSON() []byte {
	var werk bytes.Buffer

	schreiber := json.NewEncoder(&werk)
	schreiber.SetEscapeHTML(false)

	_ = schreiber.Encode(h) //nolint:errcheck // siehe oben

	// **Der Encoder hängt eine Zeile an** — sie gehört nicht zur Nachricht,
	// und ein Empfänger, der sie mitzählte, käme auf eine andere Länge.
	return bytes.TrimRight(werk.Bytes(), "\n")
}
