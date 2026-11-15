// **Ein Turnier als Datei** — DOM-04.AK-03
// (#1699),
// Vermerk [0140](../../docs/entscheidungen/0140-ein-turnier-als-datei.md):
// „Import/Export … ist möglich; eine Kopie übernimmt keine fremden
// Spielerzustimmungen oder alten Transaktionskennungen."
//
// # Was in der Datei steht
//
// **Der Verlauf, wie ihn ein Mensch nachspielen könnte**: Kopf, Teilnehmende
// samt eingereichter Liste, Runden, Paarungen, Ergebnisse und — falls
// gesetzt — das Abschlussdatum. Das Format reist mit seinem **Namen**; beim
// Einlesen sucht der Dienst es im eigenen Profil.
//
// # Was ausdrücklich nicht darin steht
//
// **Keine Kennung dieses Bestands** — weder die des Turniers noch die der
// Teilnehmenden, Runden, Paarungen, Decks oder Drucke. Eine Kennung sagt nur
// in dem Bestand etwas, aus dem sie stammt; im nächsten verwiese sie auf
// etwas anderes oder ins Leere. Die Paarungen zählen deshalb die
// Teilnehmenden **ab 1 in der Reihenfolge der Datei**. Nicht darin steht
// auch das Protokoll: es erzählt, was in **diesem** Bestand geschah, und das
// eingelesene Turnier beginnt sein eigenes mit „eingelesen".
//
// Portierungsmarke **new** — MSE2 kennt keine Turniere (Vermerk 0138).

package domain

import "encoding/json"

// TurnieraustauschFassung ist die Fassung, die dieses Programm schreibt.
const TurnieraustauschFassung = 1

// TurnieraustauschHoechstens ist die größte Datei, die gelesen wird.
//
// **Vier Mebibyte, und die Zahl ist gemessen**: ein Turnier mit 64
// Teilnehmenden, je 75 Zeilen Deckliste und acht vollen Runden wiegt
// **575.139 Bytes** (gemessen am 2026-09-27 über [FormeTurnier] eines so
// gebauten Turniers). Die Grenze ist damit gut das Siebenfache — ein Riegel
// gegen eine Datei, die jemand groß gemacht hat, und keine Bemessung.
const TurnieraustauschHoechstens = 4 << 20

// TurnieraustauschTiefe ist die tiefste Verschachtelung dieses Formats.
//
// **Zwölf**: das Format selbst ist **sieben** Stufen tief (Umschlag, Turnier,
// Runden, Runde, Paarungen, Paarung, Ergebnis) — derselbe Abstand wie bei der
// Partie ([PartieaustauschTiefe]).
const TurnieraustauschTiefe = 12

// Turnieraustausch ist der Umschlag.
type Turnieraustausch struct {
	Fassung int            `json:"fassung"`
	Turnier Turnierausgabe `json:"turnier"`
}

// Turnierausgabe ist ein Turnier, wie es in der Datei steht.
type Turnierausgabe struct {
	Name          string              `json:"name"`
	Spiel         string              `json:"spiel"`
	Format        string              `json:"format,omitempty"`
	Einzelspiel   bool                `json:"einzelspiel,omitempty"`
	Notiz         string              `json:"notiz,omitempty"`
	Abgeschlossen string              `json:"abgeschlossen,omitempty"`
	Teilnehmer    []Teilnehmerausgabe `json:"teilnehmer"`
	Runden        []Rundenausgabe     `json:"runden"`
}

// Teilnehmerausgabe ist ein Name samt eingereichter Liste.
type Teilnehmerausgabe struct {
	Name         string          `json:"name"`
	Ausgestiegen bool            `json:"ausgestiegen,omitempty"`
	Deckabzug    []Abzugsausgabe `json:"deckabzug,omitempty"`
}

// Abzugsausgabe ist eine Zeile der Liste — **ohne** die lokale Zeile des
// Drucks; `printing` ist die Kennung des Drucks bei seiner Quelle und gilt
// überall.
type Abzugsausgabe struct {
	Bezeichnung string      `json:"bezeichnung"`
	Printing    string      `json:"printing,omitempty"`
	Menge       int         `json:"menge"`
	Bereich     Deckbereich `json:"bereich"`
}

// Rundenausgabe ist eine Runde samt ihren Paarungen.
type Rundenausgabe struct {
	Nummer    int               `json:"nummer"`
	Paarungen []Paarungsausgabe `json:"paarungen"`
}

// Paarungsausgabe zählt die Teilnehmenden ab 1 — **B fehlt bei einem Bye**.
type Paarungsausgabe struct {
	A        int              `json:"a"`
	B        int              `json:"b,omitempty"`
	Ergebnis *Ergebnisausgabe `json:"ergebnis,omitempty"`
}

// Ergebnisausgabe ist ein Spielstand.
type Ergebnisausgabe struct {
	SpieleA       int `json:"spieleA"`
	SpieleB       int `json:"spieleB"`
	Unentschieden int `json:"unentschieden,omitempty"`
}

// FormeTurnier schreibt den Umschlag — eingerückt und mit Zeilenende, aus
// demselben Grund wie [FormePartie].
//
// **Ohne Fehlerrückgabe**, und aus demselben Grund: in eine [Turnierausgabe]
// gehen nur Zeichenketten, Zahlen und Wahrheitswerte.
func FormeTurnier(ausgabe Turnierausgabe) []byte {
	roh, _ := json.MarshalIndent( //nolint:errcheck,errchkjson // siehe oben
		Turnieraustausch{Fassung: TurnieraustauschFassung, Turnier: ausgabe},
		"", "  ")

	return append(roh, '\n')
}

// ZerlegeTurnier liest ihn — was gelesen wurde, was daran auffiel, und ob es
// überhaupt ging.
func ZerlegeTurnier(roh []byte) (Turnieraustausch, []Einfuhrbefund, error) {
	if len(roh) > TurnieraustauschHoechstens {
		return Turnieraustausch{}, nil, NewFieldError("datei", SchluesselValidierung,
			"Diese Datei ist %d Bytes groß; gelesen werden höchstens %d.",
			len(roh), TurnieraustauschHoechstens)
	}

	tiefe, fehler := tiefeVon(roh)
	if fehler != nil {
		return Turnieraustausch{}, nil, NewFieldError("datei", SchluesselValidierung,
			"Diese Datei ist kein JSON: %v", fehler)
	}

	if tiefe > TurnieraustauschTiefe {
		return Turnieraustausch{}, nil, NewFieldError("datei", SchluesselValidierung,
			"Diese Datei ist %d Stufen tief verschachtelt; dieses Format hat %d.",
			tiefe, TurnieraustauschTiefe)
	}

	return ausTurnierfeldern(roh)
}

// ausTurnierfeldern liest die Felder der obersten Ebene.
func ausTurnierfeldern(roh []byte) (Turnieraustausch, []Einfuhrbefund, error) {
	var felder map[string]json.RawMessage

	if fehler := json.Unmarshal(roh, &felder); fehler != nil {
		return Turnieraustausch{}, nil, NewFieldError("datei", SchluesselValidierung,
			"Diese Datei trägt kein Objekt: %v", fehler)
	}

	var austausch Turnieraustausch

	if fehler := json.Unmarshal(mitVorgabe(felder["fassung"], "0"), &austausch.Fassung); fehler != nil {
		return Turnieraustausch{}, nil, NewFieldError("fassung", SchluesselValidierung,
			"Das Feld »fassung« ist keine Zahl: %v", fehler)
	}

	if austausch.Fassung < 1 || austausch.Fassung > TurnieraustauschFassung {
		return Turnieraustausch{}, nil, NewFieldError("fassung", SchluesselValidierung,
			"Diese Datei ist in Fassung %d geschrieben; dieses Programm liest "+
				"Fassung 1 bis %d.", austausch.Fassung, TurnieraustauschFassung)
	}

	if fehler := json.Unmarshal(mitVorgabe(felder["turnier"], "{}"), &austausch.Turnier); fehler != nil {
		return Turnieraustausch{}, nil, NewFieldError("turnier", SchluesselValidierung,
			"Das Feld »turnier« hat nicht die erwartete Gestalt: %v", fehler)
	}

	// **Geprüft wird hier und nicht beim Aufrufer** — eine Datei von
	// irgendwoher ist fremde Eingabe (§8, §16).
	if fehler := austausch.Turnier.pruefe(); fehler != nil {
		return Turnieraustausch{}, nil, fehler
	}

	return austausch, fremdeFelder(felder, "fassung", "turnier"), nil
}
