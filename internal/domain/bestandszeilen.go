package domain

// Die Gestalten, in denen ein Set aus dem Bestand kommt.
//
// **Eine eigene Datei**, weil `set.go` sonst über die Zeilengrenze wüchse
// (AGENTS.md §2). Der Schnitt folgt der Frage: dort das Set und seine Regeln,
// hier die Zeilen, wie sie aus einer Abfrage zurückkommen.

// SetHead ist ein Set ohne seine Karten — für Listen.
//
// **In der Domäne und nicht in `storage`.** `internal/app` liest über eine
// Schnittstelle, und Gos strukturelle Typisierung gilt für Schnittstellen, nicht
// für benannte Typen in Parameterpositionen: eine Zusage über `[]storage.Setkopf`
// zwänge `app`, `storage` zu importieren — und das darf nur der Composition
// Root (AGENTS.md §2).
//
// Dieselbe Lehre wie bei `RawBlock`, und sie hat dort eine Schnittstelle
// gekostet, die niemand erfüllen konnte.
type SetHead struct {
	ID    int64
	Name  string
	Game  string
	Style string

	// Origin ist der Pfad, aus dem das Set gelesen wurde; leer heißt „aus keiner
	// Datei" (siehe `Set.Origin`).
	Origin string

	// Cards ist die **Zahl** der Karten, nicht die Karten. Eine Liste, die 3.000
	// Karten mitlädt, um eine Zahl zu zeigen, lässt jemanden warten.
	Cards int
}

// CardLine ist eine Karte, wie eine **Liste** sie zeigt (#891).
//
// **Ohne Feldwerte.** `Black.mse-set` hat 3.788 Karten; sie mit jedem Wert zu
// laden waren 4.761.182 Bytes JSON je Aufruf — gemessen am 2026-08-25. Die
// Liste braucht davon einen Namen je Karte.
type CardLine struct {
	ID     int64
	Number int

	// Name ist der Wert des Feldes `name` — leer, wenn es keinen gibt.
	Name string

	// Fields ist die **Zahl** der Feldwerte, nicht die Werte.
	//
	// Die Liste zeigt „ohne Felder" an einer Karte, die keine hat: der
	// Unterschied zwischen „heißt nicht" und „ist leer".
	Fields int
}

// CardField ist **ein** Feld **einer** Karte (#891).
//
// Die Massenänderung zeigt neben jeder angehakten Karte den Wert, um den es
// geht — und braucht dafür genau diesen einen Wert, nicht die Karte.
type CardField struct {
	ID int64

	// Text ist der Wert; er ist leer, wenn die Karte das Feld nicht hat. Anders
	// als bei `CardRow.Value` steht hier kein zweiter Rückgabewert: die Abfrage
	// gibt nur Karten zurück, die den Wert **haben**.
	Text string
}

// CardRow ist eine Karte mit ihren Feldwerten, wie sie aus dem Bestand kommt.
type CardRow struct {
	ID int64

	// SetID ist das Set, zu dem diese Karte gehört.
	//
	// **Sie steht hier, weil der Stil am Set hängt.** `Vorschau` und
	// `Bildausfuhr` bekommen eine Kartenkennung und sonst nichts; ohne diese
	// Angabe fänden sie das Stilpaket nicht und zeigten den Behelfsplan
	// (#206).
	SetID int64

	Number int
	Values []Value
}

// Notizenfeld ist der Schlüssel, unter dem die Notizen einer Karte stehen.
//
// **Ein Feldwert und keine Spalte.** Ein `card:`-Block trägt sein `notes` wie
// jeden anderen Eintrag, und der Einleser macht daraus einen [Value]; die
// Rundlaufzusage verlangt, dass er dort **bleibt** — die Liste ist die, die
// zurückgeschrieben wird.
const Notizenfeld = "notes"

// Notizen gibt die Notizen dieser Karte — `card.notes`.
//
//	class Card … {
//	  IndexMap<FieldP, ValueP> data;
//	  /// Notes for this card
//	  String notes;
//	}
//
// `src/data/card.hpp`, Zeilen 36–40, Fassung `f9d9356` vom 2021-07-06.
//
// **Kein Kartenfeld des Spiels, sondern eine Eigenschaft jeder Karte** —
// deshalb steht es in keiner `card field:`-Liste. `magic.mse-game` liest es an
// drei Stellen, darunter die Kartennummer
// (#270).
//
// **Eine Ableitung und kein zweites Feld** — Werkstück 6 von Phase 3.7,
// Befund Z1-01 (#1168).
// Ein eigenes Feld neben demselben Wert in [CardRow.Values] ließe eine
// Reihenfolge entscheiden, die nirgends zugesagt ist. MSE hat nur eine
// Darstellung:
// `REFLECT(notes)` steht vor `REFLECT_NAMELESS(data)` (`src/data/card.cpp`,
// Zeilen 91 und 95), und `GetMember` nimmt den ersten Treffer.
//
// **Leer statt fehlend**: MSE führt `notes` an jeder Karte.
func (c *CardRow) Notizen() string {
	text, _ := c.Value(Notizenfeld)

	return text
}

// Value gibt den Wert eines Felds; der zweite Rückgabewert sagt, ob es ihn gibt.
//
// Dieselbe Aussage wie bei `Card.Value` und aus demselben Grund: „das Feld ist
// leer" und „das Feld gibt es nicht" sind zwei verschiedene Dinge.
func (c *CardRow) Value(feld string) (string, bool) {
	for _, wert := range c.Values {
		if wert.Field == feld {
			return wert.Text, true
		}
	}

	return "", false
}
