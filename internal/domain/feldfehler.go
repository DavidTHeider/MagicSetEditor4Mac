// **Der Befund an einem einzelnen Feld** — abgetrennt von `fehler.go`, als
// die Datei mit dem zweiten Faktor über die Zeilengrenze aus §2 kam
// (#1351,
// Werkstück 7).
//
// **Der Schnitt folgt der Frage**: nebenan stehen die **Klassen** — welche
// Fehlerarten es gibt und welche Schlüssel dazugehören —, hier der **Träger**,
// der einen davon an ein Feld hängt.

package domain

import "fmt"

// FieldError ist ein Befund an einem einzelnen Feld.
//
// Er trägt drei Dinge, und jedes hat seinen Grund: das **Feld**, damit die
// Meldung dort erscheint, wo sie entstanden ist (AGENTS.md §6); den
// **Schlüssel**, damit sie sich übersetzen lässt; und den **Satz**, damit sie
// auch ohne Katalog etwas sagt.
type FieldError struct {
	Field   string
	Key     Schluessel
	Message string

	// Werte füllt die Platzhalter des Katalogsatzes — siehe
	// [FieldError.MitWerten]. Leer bei den meisten Befunden.
	Werte map[string]string

	// grund ist die verdeckte Ursache — siehe [FieldError.MitGrund].
	grund error
}

// NewFieldError baut einen Feldbefund aus einer Vorlage und ihren Werten.
//
// **Ein nicht angemeldeter Schlüssel wird hier zu SchluesselUnbekannt**, und der
// Satz bleibt. Das ist die Stelle, an der die Zusage aus §9 eingelöst wird:
// `Schluessel` ist ein offener Zeichenkettentyp, der Vertragstest prüft nur die
// **angemeldeten** — ohne diese Normalisierung könnte Go also einen Schlüssel
// setzen, den kein Katalog kennt, und in einer englischen Sitzung bliebe der
// deutsche Rückfallsatz stehen. Ein unbekannter Schlüssel ist ein
// Programmierfehler; er darf nicht zu einer halb übersetzten Oberfläche werden.
func NewFieldError(feld string, schluessel Schluessel, vorlage string, werte ...any) *FieldError {
	if !IstAngemeldet(schluessel) {
		schluessel = SchluesselUnbekannt
	}

	return &FieldError{
		Field:   feld,
		Key:     schluessel,
		Message: fmt.Sprintf(vorlage, werte...),
	}
}

func (f *FieldError) Error() string { return f.Message }

// Unwrap ordnet den Befund seiner Fehlerart zu — **nach seinem Schlüssel**.
//
// **Nicht immer [ErrValidation]**: ein Feldbefund mit `SchluesselKonflikt`
// ist ein Konflikt, und `errors.Is(err, ErrConflict)` muss ihn finden. Der
// Schlüssel sagt bereits, welche Art vorliegt — ihn zu setzen und dann zu
// ignorieren, hieße ihn zweimal zu führen.
//
// **Und dazu der verdeckte Grund**, wenn einer dasteht — [FieldError.MitGrund].
func (f *FieldError) Unwrap() []error { return []error{f.art(), f.grund} }

// art ist die Fehlerart zum Schlüssel.
func (f *FieldError) art() error {
	// **Drei Schlüssel meinen einen doppelten Eintrag** — ENG-05.AK-03
	// (#1730): eine verletzte Eindeutigkeit steht am Feld, und sie ist ein
	// Konflikt, damit REST weiter mit 409 antwortet.
	switch f.Key {
	case SchluesselDoppelt, SchluesselDoppelterName, SchluesselDoppeltesExemplar:
		return ErrDoppelt
	}

	// **Sieben Schlüssel meinen einen Konflikt** — seit DOM-04.AK-04 (#1699)
	// auch ein abgeschlossenes Turnier, seit DOM-01.AK-03 (#1696) ein
	// geändertes Paket, das ersetzt werden soll, seit DOM-02 (#1697) ein
	// Symbol, das Sets tragen, seit #1979 ein Sitz, der gegangen ist, seit
	// U02b (#2179) ein fremder Wert vor der Annahme. `karteInDeck` ist keine
	// Validierung: die Eingabe ist in Ordnung, der Bestand steht ihr entgegen
	// (#1369).
	if f.Key == SchluesselKonflikt || f.Key == SchluesselKarteInDeck ||
		f.Key == SchluesselTurnierAbgeschlossen || f.Key == SchluesselPaketGeaendert ||
		f.Key == SchluesselSymbolVerwendet || f.Key == SchluesselSitzGegangen ||
		f.Key == SchluesselFremderWert {
		return ErrConflict
	}

	// **Und einer meint einen fehlenden Schritt.** „Hier fehlt der zweite
	// Faktor" ist keine Abweisung der Eingabe: was dastand, war richtig, es
	// fehlt nur noch etwas (Werkstück 7, #1351).
	if f.Key == SchluesselZweiterFaktor {
		return ErrZweiterFaktor
	}

	// **Und einer meint „gibt es nicht".**
	//
	// **Dieselbe Klasse wie beim Konflikt eine Zeile darüber**, und derselbe
	// Fehler: ein Feldbefund mit `nichtGefunden` meldete sich als
	// Validierung, und `errors.Is(err, ErrNotFound)` war rot. Gefunden am
	// Export personenbezogener Daten (Werkstück 8, #1351).
	if f.Key == SchluesselNichtGefunden {
		return ErrNotFound
	}

	return ErrValidation
}

// MitWerten hängt einem Befund die Werte für seine Platzhalter an.
//
// # Warum sie nicht im Satz stehen
//
// Weil der Satz übersetzt wird. Der Katalog des Fensters führt für jeden
// Schlüssel einen eigenen Text mit Platzhaltern; stünden die Namen nur im
// deutschen Satz aus Go, ginge die englische Fassung ohne sie hinaus
// (#940 hat
// das für die übrigen Meldungen gelöst, für Feldbefunde bis
// #1369
// nicht).
func (f *FieldError) MitWerten(werte map[string]string) *FieldError {
	f.Werte = werte

	return f
}

// MitGrund hängt einem Befund seine technische Ursache an — für das
// Protokoll und nur dafür, wie [MitGrund] (ENG-05.AK-03, #1730).
//
// **Warum ein Befund einen braucht**: eine verletzte Eindeutigkeit gehört ans
// Feld (§6), und der Treibertext, der sie meldete, ins Protokoll (§8). Ohne
// ihn stünde dort nur, *dass* etwas doppelt war, und nicht, in welcher Tafel.
func (f *FieldError) MitGrund(grund error) *FieldError {
	f.grund = grund

	return f
}
