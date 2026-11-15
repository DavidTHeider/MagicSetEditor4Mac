package domain

import "strings"

// Value ist der Wert eines Kartenfelds.
//
// **Text, immer.** Die Typisierung steht im Spielpaket, nicht hier
// ([0002](../../docs/entscheidungen/0002-sqlite-als-arbeitsformat.md)): ein
// MSE-Spiel definiert seine Felder selbst, und eine Domäne, die für jedes
// fremde Spiel einen eigenen Typ bräuchte, müsste für jedes migrieren.
type Value struct {
	Field string
	Text  string

	// Nachgezogen sagt, dass das Wertskript eines **anderen** Felds diesen
	// Wert rechnete (#2396). Kein Fenster hat ihn bearbeitet; die Bedingung
	// bekannter Werte (U02b, #2179) gilt deshalb nicht für ihn
	// (Gesamtreview über 0.9.0, Befund 1). **Nur auf dem Weg in den
	// Bestand** — gespeichert wird sie nicht, und hinaus geht sie nicht.
	Nachgezogen bool `json:"-"`
}

// RawBlock ist ein Bereich einer Paketdatei, den wir nicht deuten.
//
// **In der Domäne und nicht in `service` oder `storage`.** Der erste Anlauf
// definierte ihn in beiden — als `service.Rohblock` und `storage.Rohblock`, mit
// dem Kommentar, Gos strukturelle Typisierung bringe sie an der Schnittstelle
// zusammen. Das ist falsch: sie gilt für Schnittstellen, nicht für benannte
// Typen in Parameterpositionen. `[]storage.Rohblock` erfüllte eine Zusage über
// `[]service.Rohblock` nie, Feld für Feld gleich oder nicht.
//
// Aufgefallen ist das erst dem Integrationstest im Composition Root. Die
// Fachtests beider Schichten liefen gegen ihren je eigenen Typ und waren grün —
// genau die Fehlerklasse, vor der AGENTS.md §3 warnt: eine Vorrichtung, die
// großzügiger ist als der Betrieb.
type RawBlock struct {
	File string
	Line int
	Text string
}

// Card ist eine Karte.
//
// Die Werte liegen als **geordnete Liste**, nicht als Abbildung. Der Grund ist
// die Rundlaufzusage: die Reihenfolge in der Datei ist die, in der sie
// zurückgeschrieben werden muss — und eine Abbildung in Go hat gar keine.
type Card struct {
	Values []Value
}

// Value gibt den Wert eines Felds; der zweite Rückgabewert sagt, ob es ihn gibt.
//
// Zwei Rückgabewerte und nicht nur einer: „das Feld ist leer" und „das Feld
// gibt es nicht" sind zwei verschiedene Aussagen, und wer sie zusammenwirft,
// kann einen absichtlich geleerten Wert nicht von einem fehlenden trennen.
func (c *Card) Value(feld string) (string, bool) {
	for _, wert := range c.Values {
		if wert.Field == feld {
			return wert.Text, true
		}
	}

	return "", false
}

// SetValue setzt einen Feldwert und behält dabei die Reihenfolge.
//
// Ein vorhandenes Feld wird an seiner Stelle geändert, ein neues hinten
// angehängt. Wer stattdessen löscht und neu anhängt, ändert die Reihenfolge —
// und damit die Datei, ohne dass jemand es angeordnet hätte.
func (c *Card) SetValue(feld, text string) {
	for i := range c.Values {
		if c.Values[i].Field == feld {
			c.Values[i].Text = text

			return
		}
	}

	c.Values = append(c.Values, Value{Field: feld, Text: text})
}

// Validate prüft eine Karte gegen die Felddefinitionen ihres Spiels.
//
// **Ein unbekanntes Feld ist kein Fehler.** Ein Set kann mit einem neueren
// Spielpaket entstanden sein, oder mit einem, das jemand selbst erweitert hat.
// Der Wert wird mitgeführt, nicht abgewiesen — sonst verlöre ein Rundlauf ihn.
func (c *Card) Validate(spiel Game) error {
	if befund := c.pruefe(spiel); befund != nil {
		return befund
	}

	return nil
}

// pruefe ist Validate mit dem Befund als greifbarem Wert.
//
// Getrennt, damit `Set.Validate` Schlüssel und Feldbezug übernehmen kann, ohne
// sie über `errors.As` wieder herausholen zu müssen — ein Umweg, dessen
// Fehlschlagzweig niemand je beträte und der deshalb ungeprüft bliebe.
func (c *Card) pruefe(spiel Game) *FieldError {
	// `doppelteFelder` fängt den Nil-Empfänger und kehrt mit einem Befund um;
	// die Schleife darunter erreicht ihn deshalb nie. Das steht hier, weil
	// `c.Values` an einer nil-Karte sonst abstürzte — über eine Wails-Bindung
	// nicht als Fehler, sondern als beendete Anwendung.
	if befund := c.doppelteFelder(); befund != nil {
		return befund
	}

	for _, wert := range c.Values {
		definition, bekannt := spiel.Field(wert.Field)
		if !bekannt {
			continue
		}

		if befund := pruefeGegenDefinition(definition, wert.Text); befund != nil {
			return befund
		}
	}

	return nil
}

// doppelteFelder prüft, was ohne Spielpaket zu prüfen ist.
func (c *Card) doppelteFelder() *FieldError {
	if c == nil {
		return NewFieldError("", SchluesselValidierung, "Eine Karte fehlt.")
	}

	return doppelteWerte(c.Values, "an dieser Karte")
}

// pruefeGegenDefinition prüft einen Wert gegen seine Felddefinition.
func pruefeGegenDefinition(definition Field, text string) *FieldError {
	if definition.Kind != KindChoice || text == "" {
		return nil
	}

	for _, moeglich := range definition.Choices {
		if moeglich == text {
			return nil
		}
	}

	return NewFieldError(definition.Name, SchluesselValidierung,
		"%q ist keine der Möglichkeiten des Felds %q.", text, definition.Name)
}

// Set ist ein Kartensatz.
type Set struct {
	Name  string
	Game  string
	Style string

	// Origin ist der Pfad, aus dem dieses Set gelesen wurde.
	//
	// `docs/ARCHITEKTUR.md` sagt ihn zu, und das Schema führt eine Spalte dafür
	// — gefüllt wurde sie bis zum Zweitblick über Phase 2 nie. Eine Spalte, die
	// immer ihren Standardwert trägt, ist eine Zusage ohne Einlösung.
	//
	// **Leer ist erlaubt und heißt: aus keiner Datei.** Ein Set, das in der
	// Anwendung entsteht, hat keine Herkunft, und ihm eine zu erfinden wäre
	// eine Falschangabe. Der Import füllt sie immer; wer sie liest, prüft
	// deshalb auf leer statt sich auf einen Pfad zu verlassen.
	Origin string

	// Values sind die Angaben des Sets selbst — Setname, Symbol, Copyright.
	Values []Value
	Cards  []*Card

	// Keywords sind die Schlüsselwörter, die dieses Set mitbringt
	// (#1062).
	//
	// **In der Reihenfolge der Datei**, und aus demselben Grund wie bei den
	// Kartenwerten: sie entscheidet, welches Schlüsselwort zuerst auf einen
	// Text passt (`src/data/keyword.cpp`, Zeilen 148–208, Fassung f9d9356 vom
	// 2021-07-06).
	//
	// **Die des Spiels stehen hier nicht.** Sie kommen aus dem Spielpaket, sind
	// unveränderlich und gehören keinem Set — MSE unterscheidet sie über
	// `Keyword::fixed` (`keyword.hpp:104`).
	Keywords []*Keyword

	// Packs sind die Packarten, die dieses Set selbst mitbringt
	// (#1081).
	//
	// **Sie gehen denen des Spiels vor.** MSE sucht zuerst hier
	// (`src/data/pack.cpp`, Zeilen 373–392, Fassung f9d9356 vom 2021-07-06),
	// und das Feld heißt dort „Additional/replacement pack types" — beides.
	Packs []*PackArt

	// Kopf sind die fünf Formatangaben des Setkopfs (#1325).
	//
	// **Sie stehen nicht in `Values`**: dort führt das **Spiel**, was ein Set
	// beschreibt; hier steht, was das **Format** über jede Datei sagt.
	Kopf Setkopf

	// Quelle ist die Datei, aus der dieses Set gelesen wurde — Byte für Byte.
	//
	// **Nil ist erlaubt und heißt: aus keiner Datei.** Ein Set, das in der
	// Anwendung entsteht, hat keine, und die Byte-Zusage gilt dort nicht — es
	// gibt nichts, womit zu vergleichen wäre
	// ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)).
	Quelle *Quelle
}

// Quelle ist die Datei, aus der ein Set gelesen wurde.
//
// # Warum die Bytes und nicht die Rohzeile je Wert
//
// Eine Rohzeile rettet Einrückung und Abstände **einer Zeile** — nicht, was
// zwischen den Zeilen steht: Kommentare, Leerzeilen, ein zweites `set info:`.
// Genau davon handelt
// #88.
type Quelle struct {
	// Datei ist der Name **im Paket**, nicht auf der Platte — `set`.
	Datei string

	// Inhalt sind die Bytes, wie sie dastanden: mit BOM, mit gemischten
	// Zeilenenden, ohne letzten Umbruch, in welcher Kodierung auch immer (§8).
	Inhalt []byte

	// Herkunft ist das Paketverzeichnis, aus dem gelesen wurde — oder leer.
	//
	// **Es steht hier, weil es beim Schreiben gebraucht wird**: das Beiwerk
	// eines Sets — Bilder, Symbole — liegt dort und nicht im Bestand.
	Herkunft string
}

// Value gibt eine Setangabe.
func (s *Set) Value(feld string) (string, bool) {
	for _, wert := range s.Values {
		if wert.Field == feld {
			return wert.Text, true
		}
	}

	return "", false
}

// ValidateBasics prüft, was **ohne** das Spielpaket prüfbar ist.
//
// Getrennt von `Validate`, weil der Import stattfindet, bevor feststeht, ob das
// zugehörige Spiel überhaupt installiert ist.
//
// # Was sie deckt
//
// Ein fehlendes Spiel, doppelte oder namenlose Setwerte und ein Feld, das an
// derselben Karte zweimal vorkommt.
//
// # Was sie durchlässt, und warum
//
// **Eine leere Herkunft.** Sie ist erlaubt und heißt „aus keiner Datei" — ein
// Set, das in der Anwendung entsteht, hat keine.
//
// **Einen leeren Setnamen.** Kein gesichtetes Paket sagt, dass ein Set einen
// Titel tragen muss; ihn abzuweisen hieße, eine Regel zu erfinden, und die
// erste fremde Datei ohne Titel wäre unlesbar. Auch eine Ablehnung braucht
// Deckung (§10) — sie steht bei `TestEinLeererSetnameBleibtErlaubt`.
//
// Dass `service.namenAus` beim Import auf den Dateinamen zurückfällt, ist keine
// Regel, sondern die Wahl eines **Anzeigenamens**: eine Liste, in der eine
// Zeile leer bliebe, sagte niemandem, welches Set gemeint ist.
//
// # Warum die Regeln hier stehen und nicht dort, wo sie zuerst auffielen
//
// Doppelte Setwerte wies bis dahin der Primärschlüssel von `set_values` ab; das
// Urteil stimmte, aber es entstand in `storage`, das die Regel nicht kennt. Der
// Setname kam aus `service.namenAus`, das auf den Dateinamen zurückfällt. Beide
// Regeln galten also, solange **ein bestimmter Weg** genommen wurde — ein
// zweiter Schreiber käme daran vorbei, und kein Test würde rot.
func (s *Set) ValidateBasics() error {
	if strings.TrimSpace(s.Game) == "" {
		return NewFieldError("game", SchluesselPflichtfeld, "Ein Set nennt kein Spiel.")
	}

	// **Die Setwerte gehören dazu** (#89). Ihr Urteil fiel bis hierher im
	// Primärschlüssel von `set_values` — es stimmte, aber es entstand in
	// `storage`, das die Regel nicht kennt. Wer die Spalte einmal ohne
	// Primärschlüssel neu schnitte, verlöre die Prüfung, ohne dass ein Test rot
	// wird. Und die Meldung von dort sagt nicht, **welches** Feld doppelt ist.
	if befund := doppelteWerte(s.Values, "an diesem Set"); befund != nil {
		return befund
	}

	for nummer, karte := range s.Cards {
		if befund := karte.doppelteFelder(); befund != nil {
			return NewFieldError(befund.Field, befund.Key,
				"Karte %d: %s", nummer+1, befund.Message)
		}
	}

	return nil
}

// doppelteWerte prüft eine Werteliste auf doppelte und namenlose Felder.
//
// **Beides an einer Stelle**, weil es dieselbe Schleife ist und weil ein
// namenloses Feld sonst als „zweimal vorhanden" gemeldet würde, sobald zwei
// davon dastehen — eine Meldung, die auf die falsche Ursache zeigt.
func doppelteWerte(werte []Value, wo string) *FieldError {
	gesehen := make(map[string]bool, len(werte))

	for _, wert := range werte {
		if befund := feldnameBefund(wert.Field); befund != nil {
			return befund
		}

		if gesehen[wert.Field] {
			return NewFieldError(wert.Field, SchluesselDoppelt,
				"Das Feld %q kommt %s zweimal vor.", wert.Field, wo)
		}

		gesehen[wert.Field] = true
	}

	return nil
}

// PruefeFeldname weist einen Feldnamen ab, der keiner ist.
//
// # Warum diese Regel hier steht
//
// #89 fragt
// nicht „mehr prüfen", sondern **wo**: eine Regel, die nur an einer Fassade
// steht, kennt die Domäne nicht — und der nächste Schreiber geht daran vorbei.
//
// **Genau das war eingetreten.** `app.FeldSetzen` wies einen leeren Feldnamen
// ab, seit Phase 5; der MCP-Server nimmt einen anderen Weg
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)) und wies
// ihn nicht ab. Ein Sprachmodell konnte eine Zeile anlegen, die niemand
// wiederfindet — und beim Zurückschreiben stünde `: wert` in der Paketdatei.
//
// **Leerraum ist kein Name.** Ein Feld `" "` fände niemand wieder, und in der
// Paketdatei sähe seine Zeile aus wie ein Einrückungsfehler.
//
// Beschnitten wird nur zum **Prüfen**. Wer den Namen beim Schreiben stutzte,
// änderte eine Zeile, die niemand angefasst hat — gegen die Byte-Zusage aus
// #48.
func PruefeFeldname(feld string) error {
	if befund := feldnameBefund(feld); befund != nil {
		return befund
	}

	return nil
}

// feldnameBefund gibt denselben Befund als Feldfehler.
//
// **Zwei Signaturen für eine Regel**, und das ist Absicht: nach außen ein
// `error`, weil ein Aufrufer ihn umschließt; nach innen ein `*FieldError`, weil
// `doppelteWerte` ihn samt Feld und Schlüssel weiterreicht. Ein `errors.As`
// dazwischen wäre ein Umweg über eine Schnittstelle, um zurückzubekommen, was
// eine Zeile vorher dastand.
func feldnameBefund(feld string) *FieldError {
	if strings.TrimSpace(feld) == "" {
		return NewFieldError("feld", SchluesselPflichtfeld, "Ein Feld braucht einen Namen.")
	}

	return nil
}

// Validate prüft ein Set und seine Karten.
func (s *Set) Validate(spiel Game) error {
	if strings.TrimSpace(s.Game) == "" {
		return NewFieldError("game", SchluesselPflichtfeld, "Ein Set nennt kein Spiel.")
	}

	if s.Game != spiel.Name {
		return NewFieldError("game", SchluesselKonflikt,
			"Das Set nennt das Spiel %q, geprüft wird gegen %q.", s.Game, spiel.Name)
	}

	for nummer, karte := range s.Cards {
		befund := karte.pruefe(spiel)
		if befund == nil {
			continue
		}

		// Schlüssel und Feldbezug reisen mit. Ohne sie erschiene die Meldung
		// als anonymer Hinweis statt an dem Feld, an dem sie entstanden ist
		// (AGENTS.md §6) — und die Fehlerart ginge beim Umschließen verloren.
		return NewFieldError(befund.Field, befund.Key,
			"Karte %d: %s", nummer+1, befund.Message)
	}

	return nil
}
