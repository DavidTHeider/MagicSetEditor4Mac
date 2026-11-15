package domain

// **Ein Ereignis trägt die Sichtkennungen seines Augenblicks** — #2480,
// Vermerk 0201, Punkt 5.
//
// # Warum als Ergänzung und nicht als Feld
//
// **Weil ein Feld eine Spalte bräuchte** — und eine Schemaänderung ist nach
// dem Plan für 0.9.0 ein eigener Zug. Die Zuordnung steht deshalb als
// Ergänzung an einen Empfänger, der kein Sitz ist: [Ereignis.FuerMitleser]
// gibt eine Ergänzung nur ihrem Empfänger, und diesen gibt es am Tisch
// nicht. Die Kette rechnet über den öffentlichen Teil und sieht sie nicht.
//
// # Was sie nicht kann
//
// **Ein Ereignis aus der Zeit davor trägt keine Zuordnung** und nennt seine
// Kennungen weiter. Es gibt sie nur in Partien, die vor 0.9.0 begannen.

import (
	"encoding/json"
	"maps"
)

// sichtnamenEmpfaenger ist der Empfänger der Zuordnung — **kein Sitz**, und
// deshalb geht sie nie hinaus.
const sichtnamenEmpfaenger = "#sichtnamen"

// Benennen gibt jedem Ereignis die Sichtkennungen der Objekte, die es nennt
// — so, wie sie nach dem Befehl heißen, und für ein Objekt, das es danach
// nicht mehr gibt, so, wie es vorher hieß.
func Benennen(vorher, danach Partie, ereignisse []Ereignis) []Ereignis {
	namen := vorher.sichtnamen()
	maps.Copy(namen, danach.sichtnamen())

	benannt := make([]Ereignis, len(ereignisse))

	for stelle, ereignis := range ereignisse {
		genannt := map[string]any{}

		sammle := func(wert string) {
			if name, da := namen[wert]; da {
				genannt[wert] = name
			}
		}

		besuche(ereignis.Oeffentlich, sammle)

		for _, ergaenzung := range ereignis.Privat {
			besuche(ergaenzung.Inhalt, sammle)
		}

		if len(genannt) > 0 {
			ereignis.Privat = append(append([]Ergaenzung{}, ereignis.Privat...), Ergaenzung{
				Empfaenger: sichtnamenEmpfaenger,
				Inhalt:     map[string]any{"namen": genannt},
			})
		}

		benannt[stelle] = ereignis
	}

	return benannt
}

// Erfahren sagt, ob mit diesem Ereignis jemand etwas nur für sich erfuhr —
// **die Zuordnung zählt nicht**, sie erfährt niemand.
func (e Ereignis) Erfahren() bool {
	for _, ergaenzung := range e.Privat {
		if ergaenzung.Empfaenger != sichtnamenEmpfaenger {
			return true
		}
	}

	return false
}

// zuordnung liest die Sichtkennungen eines Ereignisses — leer, wenn es keine
// trägt.
func (e Ereignis) zuordnung() map[string]string {
	zuordnung := map[string]string{}

	for _, ergaenzung := range e.Privat {
		if ergaenzung.Empfaenger != sichtnamenEmpfaenger {
			continue
		}

		genannt, _ := ergaenzung.Inhalt["namen"].(map[string]any) //nolint:errcheck // ohne Tabelle keine Zuordnung
		for kennung, name := range genannt {
			if text, ist := name.(string); ist {
				zuordnung[kennung] = text
			}
		}
	}

	return zuordnung
}

// besuche ruft `jeder` für jede Zeichenkette in einem Wert — auch in Listen,
// Tabellen und in Werten, die erst ihre JSON-Gestalt zeigt.
func besuche(wert any, jeder func(string)) {
	switch gestalt := wert.(type) {
	case nil:
	case string:
		jeder(gestalt)
	case []any:
		for _, teil := range gestalt {
			besuche(teil, jeder)
		}
	case map[string]any:
		for schluessel, teil := range gestalt {
			jeder(schluessel)
			besuche(teil, jeder)
		}
	default:
		if traeger := alsJSON(gestalt); kannTragen(traeger) {
			besuche(traeger, jeder)
		}
	}
}

// kannTragen sagt, ob ein Wert eine Kennung tragen kann — **und nur dann
// geht es weiter**: eine Zahl gibt über JSON wieder eine Zahl, und wer sie
// noch einmal besuchte, käme nie an.
func kannTragen(wert any) bool {
	switch wert.(type) {
	case string, []any, map[string]any:
		return true
	}

	return false
}

// umbenannt gibt einen Wert, in dem jede Kennung ihre Sichtkennung trägt —
// **eine neue Tabelle**, das Ereignis bleibt, was es ist.
func umbenannt(wert any, namen map[string]string) any {
	switch gestalt := wert.(type) {
	case nil:
		return nil
	case string:
		if name, da := namen[gestalt]; da {
			return name
		}

		return gestalt
	case []any:
		kopie := make([]any, len(gestalt))
		for stelle, teil := range gestalt {
			kopie[stelle] = umbenannt(teil, namen)
		}

		return kopie
	case map[string]any:
		kopie := make(map[string]any, len(gestalt))
		for schluessel, teil := range gestalt {
			name, _ := umbenannt(schluessel, namen).(string) //nolint:errcheck // eine Zeichenkette bleibt eine
			kopie[name] = umbenannt(teil, namen)
		}

		return kopie
	default:
		if traeger := alsJSON(gestalt); kannTragen(traeger) {
			return umbenannt(traeger, namen)
		}

		// **Was keine Kennung tragen kann, bleibt, wie es ist** — eine Zahl
		// bleibt eine Zahl desselben Typs. Was sich nicht kodieren lässt,
		// geht als nichts hinaus.
		if alsJSON(gestalt) == nil {
			return nil
		}

		return gestalt
	}
}

// alsJSON gibt einen Wert in der Gestalt, die er über den Draht hat.
//
// **Für das, was im Speicher noch einen Go-Typ trägt** — eine Liste von
// Angriffen etwa. Was sich nicht kodieren lässt, kommt als nichts zurück:
// die Kette hätte es vorher schon abgewiesen ([Ereignis.Abdruck]).
func alsJSON(wert any) any {
	roh, fehler := json.Marshal(wert)
	if fehler != nil {
		return nil
	}

	var gelesen any

	_ = json.Unmarshal(roh, &gelesen) //nolint:errcheck // eben selbst kodiert

	return gelesen
}
