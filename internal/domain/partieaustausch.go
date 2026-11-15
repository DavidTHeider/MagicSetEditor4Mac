// **Eine Partie als Datei** — PLAY-03.AK-05
// (#1711):
// „Partieverwaltung bietet Anlegen, Speichern/Fortsetzen, Suche, Sortierung,
// Filter und **Export/Import**."
//
// # Warum ein Umschlag und nicht die Partie selbst
//
// **Weil eine Datei ohne Fassungsangabe nicht altern kann** — dieselbe
// Überlegung wie beim Einstellungsaustausch ([Austausch]): wer die Gestalt
// später ändert, hat keine Möglichkeit, eine ältere Datei zu erkennen, und
// eine Datei aus der Zukunft läse er falsch statt gar nicht.
//
// # Was ausdrücklich nicht in der Datei steht
//
// **Der Ereignisstrom und die Einwilligungen.** Beides gehört zu der Partie,
// aus der ausgeführt wurde, und nicht zu der, die daraus entsteht — der
// Grund steht bei [Kopiere], und die Einfuhr geht durch genau diese
// Funktion.

package domain

import "encoding/json"

// PartieaustauschFassung ist die Fassung, die dieses Programm schreibt.
const PartieaustauschFassung = 1

// PartieaustauschHoechstens ist die größte Datei, die gelesen wird.
//
// **Vier Mebibyte, und die Zahl ist gemessen**: eine Partie mit vier Sitzen,
// hundert Karten je Deck und vierhundert Objekten wiegt **20.522 Bytes**
// (gemessen am 2026-09-22 über `json.Marshal` einer so gebauten [Partie]).
// Die Grenze ist damit das Zweihundertfache des Erwarteten — sie ist keine
// Bemessung, sondern ein Riegel gegen eine Datei, die jemand groß gemacht
// hat, damit dieses Programm sie in den Arbeitsspeicher legt.
const PartieaustauschHoechstens = 4 << 20

// PartieaustauschTiefe ist die tiefste Verschachtelung dieses Formats.
//
// **Zwölf, und auch die ist gemessen**: dieselbe Partie ist **sieben**
// Stufen tief, der Umschlag legt eine darüber. Der Abstand ist Luft für
// Felder, die noch dazukommen — und trotzdem eine Grenze, denn ein tief
// geschachteltes JSON kostet beim Lesen Stapel.
const PartieaustauschTiefe = 12

// Partieaustausch ist der Umschlag.
type Partieaustausch struct {
	Fassung int    `json:"fassung"`
	Partie  Partie `json:"partie"`
}

// FormePartie schreibt den Umschlag.
//
// **Eingerückt und mit Zeilenende.** Eine ausgeführte Partie landet in einem
// Verzeichnis, wird per Hand weitergereicht und irgendwann in einem Editor
// geöffnet; eine einzige Zeile mit zwanzigtausend Zeichen wäre dort
// unbrauchbar, und `git diff` über zwei davon sagte nichts.
//
// **Ohne Fehlerrückgabe, und das ist geprüft und nicht bequem** — dieselbe
// Form und derselbe Grund wie bei `storage.alsPartiejson`: in eine [Partie]
// gehen Zeichenketten, Zahlen, Wahrheitswerte und Abbildungen davon, und
// `encoding/json` kann daran nicht scheitern. Ein Zweig dafür stünde
// ungeprüft da und sähe aus wie eine geprüfte Zusage (§10).
func FormePartie(partie Partie) []byte {
	roh, _ := json.MarshalIndent( //nolint:errcheck,errchkjson // siehe oben
		Partieaustausch{Fassung: PartieaustauschFassung, Partie: partie},
		"", "  ")

	return append(roh, '\n')
}

// ZerlegePartie liest ihn.
//
// **Drei Ergebnisse und nicht zwei** — wie bei [Zerlege]: was gelesen wurde,
// was daran auffiel, und ob es überhaupt ging.
func ZerlegePartie(roh []byte) (Partieaustausch, []Einfuhrbefund, error) {
	if len(roh) > PartieaustauschHoechstens {
		return Partieaustausch{}, nil, NewFieldError("datei",
			SchluesselValidierung,
			"Diese Datei ist %d Bytes groß; gelesen werden höchstens %d.",
			len(roh), PartieaustauschHoechstens)
	}

	tiefe, fehler := tiefeVon(roh)
	if fehler != nil {
		return Partieaustausch{}, nil, NewFieldError("datei",
			SchluesselValidierung, "Diese Datei ist kein JSON: %v", fehler)
	}

	if tiefe > PartieaustauschTiefe {
		return Partieaustausch{}, nil, NewFieldError("datei",
			SchluesselValidierung,
			"Diese Datei ist %d Stufen tief verschachtelt; dieses Format hat %d.",
			tiefe, PartieaustauschTiefe)
	}

	return ausPartiefeldern(roh)
}

// ausPartiefeldern liest die Felder der obersten Ebene.
func ausPartiefeldern(roh []byte) (Partieaustausch, []Einfuhrbefund, error) {
	var felder map[string]json.RawMessage

	if fehler := json.Unmarshal(roh, &felder); fehler != nil {
		return Partieaustausch{}, nil, NewFieldError("datei",
			SchluesselValidierung,
			"Diese Datei trägt kein Objekt: %v", fehler)
	}

	var austausch Partieaustausch

	if fehler := json.Unmarshal(mitVorgabe(felder["fassung"], "0"),
		&austausch.Fassung); fehler != nil {
		return Partieaustausch{}, nil, NewFieldError("fassung",
			SchluesselValidierung,
			"Das Feld »fassung« ist keine Zahl: %v", fehler)
	}

	if austausch.Fassung < 1 || austausch.Fassung > PartieaustauschFassung {
		return Partieaustausch{}, nil, NewFieldError("fassung",
			SchluesselValidierung,
			"Diese Datei ist in Fassung %d geschrieben; dieses Programm liest "+
				"Fassung 1 bis %d.", austausch.Fassung, PartieaustauschFassung)
	}

	if fehler := json.Unmarshal(mitVorgabe(felder["partie"], "{}"),
		&austausch.Partie); fehler != nil {
		return Partieaustausch{}, nil, NewFieldError("partie",
			SchluesselValidierung,
			"Das Feld »partie« hat nicht die erwartete Gestalt: %v", fehler)
	}

	// **Geprüft wird hier und nicht beim Aufrufer.** Eine Datei von
	// irgendwoher ist fremde Eingabe (§8, §16); wer sie ungeprüft
	// weiterreichte, legte eine Partie an, die keine Regel je gesehen hat.
	if fehler := austausch.Partie.Pruefe(); fehler != nil {
		return Partieaustausch{}, nil, fehler
	}

	return austausch, fremdeFelder(felder, "fassung", "partie"), nil
}
