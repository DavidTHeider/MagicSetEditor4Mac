package mtgjson

import (
	"encoding/json"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Schreibe gibt Sets der Domäne als MTG JSON aus.
//
// # Was die Ausgabe zusagt, und was nicht
//
// **Die Kartenwerte, nicht die Bytes.** Der Rundlauf über `.mse-set` (#48) ist
// Byte für Byte zu halten, weil die Datei dort die Quelle ist. MTG JSON ist es
// nicht: es beschreibt Karten, nicht ihre Darstellung, und was MSE führt — Stil,
// Styling, Rohblöcke — hat dort keinen Platz. Zugesagt ist deshalb:
// lesen → schreiben → lesen ergibt dieselben Kartenwerte.
//
// **Was dabei verlorengeht, wird benannt** und nicht stillschweigend
// weggelassen: die Befunde stehen am Ergebnis.
//
// # Die Gestalt
//
// **Ein Set je Datei** — eine *Individual Set*-Datei mit dem Umschlag aus `meta`
// und `data`, den MTGJSON seit Fassung 5 schreibt. Gelesen wird auch die
// *AllSets*-Form; geschrieben wird sie nicht, und das ist eine Entscheidung:
// wer ein Set ausgibt, gibt eines aus, und eine Sammeldatei mit einem Eintrag
// wäre eine Form ohne Anlass. Wer mehrere braucht, schreibt mehrere Dateien in
// ein Verzeichnis — genau die Gestalt, die der Leser ohnehin nimmt.
//
// Ein Datum in `meta` schreiben wir **nicht**: es wäre eine Angabe über die
// Daten, die wir nicht haben, und ein Rundlauf mit wechselndem Datum ließe sich
// nicht vergleichen.
func (z Zuordnung) Schreibe(set *domain.Set) ([]byte, []domain.Befund) {
	gebaut, befunde := z.setAus(set)

	umschlag := map[string]any{
		"meta": map[string]any{"version": "5.2.2"},
		"data": gebaut,
	}

	// **Kann nicht scheitern.** Der Baum besteht aus Abbildungen, Listen,
	// Zeichenketten und rohem JSON, das aus einem geglückten `Unmarshal`
	// stammt. Ein Fehlerzweig, den kein Aufruf betreten kann, wäre eine
	// Behauptung darüber, was im Ernstfall geschieht (§3).
	inhalt, _ := json.MarshalIndent(umschlag, "", "  ") //nolint:errcheck // siehe darüber

	return append(inhalt, '\n'), befunde
}

// setAus baut ein Set samt seinen Karten.
func (z Zuordnung) setAus(set *domain.Set) (map[string]any, []domain.Befund) {
	// **Derselbe Ersatz wie beim Lesen** (`befundort.go`). Der Zweitblick über
	// Phase 4.2 schloss die Klasse „jeder Befund nennt sein Set" für die
	// Leserfamilie; der Nachlauf hat angestrichen, dass der Schreibweg sie
	// weiterhin roh zusammensetzte — ein Set ohne `code` und mit leerem Namen
	// gab einen Ort, der mit dem Trenner beginnt.
	rohcode, _ := set.Value("code")
	code := benannt(rohcode, benannt(set.Name, ohneCodeUndNamen))

	karten := make([]map[string]any, 0, len(set.Cards))

	var befunde []domain.Befund

	gemeldet := map[string]bool{}

	for _, karte := range set.Cards {
		// **Eine zusammengelegte Karte wird wieder getrennt** (Zweitblick über
		// Phase V, 2026-08-24). Das Lesen macht aus zwei MTGJSON-Karten eine
		// MSE-Karte; ohne diese Umkehrung gäbe das Schreiben `"name 2"` als
		// MTGJSON-Schlüssel heraus, und den gibt es dort nicht.
		//
		// **Vor der Zuordnung**, weil ` 2` ein MSE-Feldname ist: danach heißt
		// das Feld `manaCost`, und `casting cost 2` wäre dort nie angekommen.
		vorn, hinten := TrenneKartenwerte(karte.Values)

		gebaut, hinweise := z.karteAus(&domain.Card{Values: vorn})

		if len(hinten) == 0 {
			karten = append(karten, gebaut)
		} else {
			gebautHinten, weitere := z.karteAus(&domain.Card{Values: hinten})
			hinweise = append(hinweise, weitere...)

			VerbindeSeiten(gebaut, gebautHinten)
			karten = append(karten, gebaut, gebautHinten)
		}

		// **Je Feld einmal, nicht je Karte.** Derselbe Grund wie beim Lesen: ein
		// Verlust betrifft in der Regel jede Karte, und 295 gleichlautende
		// Zeilen sind keine Auskunft mehr.
		for _, hinweis := range hinweise {
			if gemeldet[hinweis.Feld] {
				continue
			}

			gemeldet[hinweis.Feld] = true

			befunde = append(befunde, domain.Befund{
				Ort:     code + " → " + benannt(hinweis.Feld, feldOhneNamen),
				Meldung: hinweis.Meldung,
			})
		}
	}

	return map[string]any{
		"code":  code,
		"name":  set.Name,
		"cards": karten,
	}, befunde
}

// karteAus baut eine Karte aus ihren Feldwerten.
func (z Zuordnung) karteAus(karte *domain.Card) (map[string]any, []Hinweis) {
	gebaut := map[string]any{}

	var hinweise []Hinweis

	zugeordnet := map[string]bool{}

	for _, paar := range z.Felder {
		if paar.Feld == "" {
			continue
		}

		zugeordnet[paar.Feld] = true

		text, da := karte.Value(paar.Feld)
		if !da || text == "" {
			continue
		}

		gebaut[paar.zielSchluessel()] = paar.zurueck(text)

		if paar.Verlust != "" {
			hinweise = append(hinweise, Hinweis{Feld: paar.Feld, Meldung: paar.Verlust})
		}
	}

	// Was die Zuordnung nicht kennt, ist ein **mitgeführter** JSON-Schlüssel —
	// beim Lesen roh abgelegt und hier roh zurückgeschrieben. Was sich nicht als
	// JSON lesen lässt, ist ein Feld, das jemand in der Anwendung angelegt hat;
	// es geht als Zeichenkette hinaus.
	for _, wert := range karte.Values {
		if zugeordnet[wert.Field] {
			continue
		}

		var roh json.RawMessage
		if err := json.Unmarshal([]byte(wert.Text), &roh); err == nil {
			gebaut[wert.Field] = roh

			continue
		}

		gebaut[wert.Field] = wert.Text
	}

	return gebaut, hinweise
}

// zielSchluessel ist der JSON-Schlüssel, in den dieses Paar zurückschreibt.
//
// Ohne eigenes `Ziel` der erste aus `JSON`. Ein Paar, das mehrere Schlüssel
// verbraucht, hat beim Lesen alle zu **einem** Wert verbunden — geschrieben
// wird er in genau einen, und welchen, sagt `Ziel`.
func (p Feldpaar) zielSchluessel() string {
	if p.Ziel != "" {
		return p.Ziel
	}

	return p.JSON[0]
}

// zurueck formt einen Kartenfeldwert in den JSON-Wert zurück.
func (p Feldpaar) zurueck(text string) any {
	if p.Rueckwaerts == nil {
		return text
	}

	return p.Rueckwaerts(text)
}

// Manasymbole macht aus `2R` wieder `{2}{R}`.
//
// **Die Gegenrichtung von `Manakosten`, und sie ist die unsicherere.** Eine Zahl
// ist ein Symbol, ein Buchstabe ist ein Symbol — das ist eindeutig. Alles
// andere ist es nicht: `W/U` könnte ein hybrides Symbol sein oder zwei, und
// welches, sagt die MSE-Schreibweise nicht. Solche Kosten gehen als **eine**
// Klammer hinaus; der Rundlauf hält, die Bedeutung ist geraten.
func Manasymbole(text string) any {
	if text == "" {
		return ""
	}

	if strings.IndexFunc(text, unbekanntesZeichen) >= 0 {
		return "{" + text + "}"
	}

	var bau strings.Builder

	zahl := ""

	for _, zeichen := range text {
		if zeichen >= '0' && zeichen <= '9' {
			zahl += string(zeichen)

			continue
		}

		if zahl != "" {
			bau.WriteString("{" + zahl + "}")
			zahl = ""
		}

		bau.WriteString("{" + string(zeichen) + "}")
	}

	if zahl != "" {
		bau.WriteString("{" + zahl + "}")
	}

	return bau.String()
}

// unbekanntesZeichen sagt, ob ein Zeichen weder Ziffer noch Buchstabe ist.
func unbekanntesZeichen(zeichen rune) bool {
	switch {
	case zeichen >= '0' && zeichen <= '9':
		return false
	case zeichen >= 'A' && zeichen <= 'Z':
		return false
	case zeichen >= 'a' && zeichen <= 'z':
		return false
	}

	return true
}

// WoerterZurueck macht aus `Golem Soldier` wieder `["Golem", "Soldier"]`.
func WoerterZurueck(text string) any {
	if text == "" {
		return []string{}
	}

	return strings.Fields(text)
}

// SeltenheitZurueck macht aus `mythic rare` wieder `mythic`.
func SeltenheitZurueck(text string) any {
	if text == "mythic rare" {
		return "mythic"
	}

	return text
}
