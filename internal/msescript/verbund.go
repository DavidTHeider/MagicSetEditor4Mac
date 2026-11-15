package msescript

import "strconv"

// Verbund ist **der Rückweg von `combined_editor`** — aus einem Feldskript
// ablesen, welche Felder es zusammensetzt, und einen Wert wieder in sie zerlegen
// (#696).
//
// # Warum das hier steht und nicht im Dienst
//
// Weil `<sep>…</sep>` eine Übereinkunft **dieser** Sprache ist: sie entsteht in
// [verbundeneFelder] und wird nur hier wieder aufgelöst. Ein Dienst, der die
// Marke selbst suchte, hätte dieselbe Regel ein zweites Mal — und die zweite
// liefe irgendwann auseinander.
//
// # Was hier bewusst nicht steht
//
// MSE schreibt die Teile **im Aufruf selbst** zurück und benachrichtigt das Set
// (`src/script/functions/editor.cpp`, Zeilen 78–92, Fassung f9d9356 vom
// 2021-07-06). Das setzt einen Auswerter voraus, der schreibt; dieser hier
// rechnet nur. Deshalb ist der Rückweg getrennt: [VerbundAus] liest ihn ab,
// **wer** schreibt, entscheidet der Dienst.
type Verbund struct {
	// Felder sind die Namen hinter `card.…`, in der Reihenfolge des Aufrufs
	// — und **so geschrieben, wie sie im Skript stehen**: `rule_text`, nicht
	// `rule text`. Wer sie einem Feld zuordnet, tut das über [GliedTrifft].
	Felder []string

	// Trenner sind die `separator`-Angaben, einer weniger als Felder.
	Trenner []string
}

// VerbundAus liest den Rückweg aus einem Feldskript.
//
// # Der Aufruf steht selten allein
//
// So sieht `text` in `magic.mse-game` wirklich aus:
//
//	if mainframe_walker() then mainframe_walker_text_script()
//	else if alt_text() then alt_text_script()
//	else combined_editor(field1: card.rule_text, separator: "<line>\n</line>",
//	                     field2: card.flavor_text)
//
// Deshalb wird der **ganze Baum** durchsucht. Eine frühere Fassung verlangte
// den Aufruf als ganzen Ausdruck — und fand ihn an keinem einzigen der 249
// Felder von `magic.mse-game`.
//
// # Was den Rückweg verschließt
//
// **Zwei Aufrufe.** Welcher gilt, entscheidet sich beim Rechnen; wer einen
// davon nähme, schriebe die Hälfte der Karten in die falschen Felder.
//
// **Ein gerechnetes Argument.** Jedes `field#` muss ein Zugriff auf `card`
// sein und jeder `separator#` eine feste Zeichenkette. `pt` und `type`
// schreiben `separator: language().pt_separator`; was dabei herauskommt, weiß
// erst der Auswerter.
//
// # Was er nicht prüft
//
// **Ob dieser Zweig auch gerechnet wird.** Trägt eine Karte einen
// Planeswalker-Rahmen, gilt `mainframe_walker_text_script()`, und die Teile
// stehen woanders. Der Rückweg nennt trotzdem `rule text` und `flavor text` —
// die Felder, aus denen der Text **gewöhnlich** entsteht. Sie anzubieten ist
// in jedem Fall näher an der Karte als ein Kasten, dessen Inhalt nirgends
// ankommt; MSE entscheidet es zur Laufzeit, weil dort der Aufruf selbst
// zurückschreibt (`src/script/functions/editor.cpp`, Zeilen 78–92, Fassung
// f9d9356 vom 2021-07-06).
func VerbundAus(skript string) (Verbund, bool) {
	aufruf := einzigerVerbund(skript)
	if aufruf == nil {
		return Verbund{}, false
	}

	verbund := Verbund{
		Felder:  kartenfelder(aufruf),
		Trenner: festeTrenner(aufruf),
	}

	if len(verbund.Felder) == 0 || len(verbund.Trenner) < len(verbund.Felder)-1 {
		return Verbund{}, false
	}

	return verbund, true
}

// VerbundfelderAus nennt die Felder, aus denen ein Skript etwas zusammensetzt
// — **auch dann, wenn sich der Wert nicht zurück spalten lässt**.
//
// # Warum das eine zweite Frage ist
//
// `type` und `pt` schreiben `separator: language().type_separator`. Welcher
// Trenner dazwischen steht, weiß erst der Auswerter — [VerbundAus] gibt
// deshalb auf, und **spalten** kann man den Wert wirklich nicht.
//
// Die Oberfläche fragt aber etwas anderes: *woraus entsteht dieses Feld?* Die
// Antwort steht im Aufruf, gleich wie der Trenner zustande kommt. Wer sie
// verschwiege, böte ein Feld an, das keine Eingabe annimmt — und `type` ist
// das Feld, das jeder als Erstes ausfüllen will
// (#687).
func VerbundfelderAus(skript string) []string {
	aufruf := einzigerVerbund(skript)
	if aufruf == nil {
		return nil
	}

	return kartenfelder(aufruf)
}

// einzigerVerbund gibt den einen `combined_editor` eines Skripts — oder nichts.
func einzigerVerbund(skript string) *Knoten {
	baum, err := Zerteile(skript)
	if err != nil {
		return nil
	}

	var gefunden []*Knoten

	verbundaufrufe(baum, &gefunden)

	if len(gefunden) != 1 {
		return nil
	}

	return gefunden[0]
}

// verbundaufrufe sammelt jeden `combined_editor` im Baum.
//
// **Auch die in Bedingungen und Argumenten** — der Aufruf steht im Pack im
// dritten Zweig einer Verzweigung, und ein Baumgang, der nur die erste Ebene
// ansähe, fände ihn nicht.
//
// **Ohne Wächter auf `nil`**: der Zerteiler hängt kein leeres Kind an, auch
// nicht an eine Bedingung ohne `else` — nachgemessen an vier Formen. Ein
// Wächter, den kein Aufruf erreicht, sagt nichts (§10).
func verbundaufrufe(knoten *Knoten, gefunden *[]*Knoten) {
	if istVerbund(knoten) {
		*gefunden = append(*gefunden, knoten)
	}

	for _, kind := range knoten.Kinder {
		verbundaufrufe(kind, gefunden)
	}
}

// istVerbund sagt, ob dieser Knoten ein `combined_editor` ist.
//
// **Nur ein Name wird gerufen.** `card.combined_editor(…)` ruft ein Glied der
// Karte; was es tut, weiß niemand.
func istVerbund(knoten *Knoten) bool {
	if knoten.Form != FormAufruf {
		return false
	}

	gerufen := knoten.Kinder[0]

	return gerufen.Form == FormName &&
		(gerufen.Text == "combined_editor" || gerufen.Text == "forward_editor")
}

// kartenfelder liest `field`, `field1`, `field2`, … als Zugriffe auf die Karte.
//
// Die Reihe ist dieselbe wie in [durchnummeriert] — und sie endet an der
// ersten Lücke. Ein Argument, das **kein** `card.…` ist, beendet sie
// ebenfalls: was gerechnet wird, lässt sich nicht zurückschreiben.
func kartenfelder(aufruf *Knoten) []string {
	var reihe []string

	for stelle := 0; ; stelle++ {
		kind := benanntesKind(aufruf, nummeriert("field", stelle))
		if kind == nil {
			if stelle > 0 {
				return reihe
			}

			continue
		}

		if kind.Form != FormZugriff || len(kind.Kinder) != 1 {
			return nil
		}

		if unter := kind.Kinder[0]; unter.Form != FormName || unter.Text != "card" {
			return nil
		}

		reihe = append(reihe, kind.Text)
	}
}

// festeTrenner liest `separator`, `separator1`, … als Zeichenketten.
func festeTrenner(aufruf *Knoten) []string {
	var reihe []string

	for stelle := 0; ; stelle++ {
		kind := benanntesKind(aufruf, nummeriert("separator", stelle))
		if kind == nil {
			if stelle > 0 {
				return reihe
			}

			continue
		}

		if kind.Form != FormZeichenkette {
			return nil
		}

		reihe = append(reihe, kind.Text)
	}
}

// nummeriert bildet `name`, `name1`, `name2`, … — MSE hängt die Zahl erst ab
// dem zweiten Durchgang an, und die Vorlagen schreiben `field1` und `field2`.
func nummeriert(name string, stelle int) string {
	if stelle == 0 {
		return name
	}

	return name + strconv.Itoa(stelle)
}

// benanntesKind gibt das Argument mit diesem Namen — oder nichts.
func benanntesKind(aufruf *Knoten, name string) *Knoten {
	for stelle := range aufruf.Namen {
		if aufruf.Namen[stelle] == name && stelle < len(aufruf.Kinder) {
			return aufruf.Kinder[stelle]
		}
	}

	return nil
}

// Zerlege spaltet einen zusammengesetzten Wert in seine Teile.
//
// Die Regel steht in `src/script/functions/editor.cpp`, Zeilen 66–76, Fassung
// f9d9356 vom 2021-07-06:
//
//	pos = value.find("<sep");
//	while (pos != npos) {
//	  value_parts.push_back(value.substr(0, pos));
//	  value = value.substr(min(match_close_tag_end(value,pos), value.size()));
//	  pos = value.find("<sep");
//	}
//	value_parts.push_back(value);
//	value_parts.resize(values.size());
//
// **`<sep` trifft auch `<sep-soft>`** — dieselbe Marke, weicher gesetzt. Und
// die Zahl der Teile ist am Ende die Zahl der Felder: fehlende sind leer,
// überzählige fallen weg. Das letzte ist MSEs offene Frage
// (`// TODO: what if there are more value_parts than values?`) und hier
// genauso beantwortet, damit ein Wert nicht in ein Feld läuft, das der Aufruf
// nicht nennt.
func (v Verbund) Zerlege(wert string) []string {
	teile := make([]string, 0, len(v.Felder))

	for {
		anfang := marke(wert)
		if anfang < 0 {
			break
		}

		teile = append(teile, wert[:anfang])
		wert = wert[min(schlussende(wert, anfang), len(wert)):]
	}

	teile = append(teile, wert)

	for len(teile) < len(v.Felder) {
		teile = append(teile, "")
	}

	return teile[:len(v.Felder)]
}

// marke findet die nächste `<sep…`-Marke.
func marke(wert string) int {
	for stelle := 0; stelle+4 <= len(wert); stelle++ {
		if wert[stelle:stelle+4] == "<sep" {
			return stelle
		}
	}

	return -1
}

// GliedTrifft ordnet einen Skriptnamen einem Feldnamen zu.
//
// **`rule_text` findet `rule text`** — MSEs `canonical_name_compare`, dieselbe
// Regel wie beim Lesen eines Glieds. Sie steht hier ein zweites Mal
// **exportiert**, weil der Rückweg außerhalb dieses Pakets eingelöst wird und
// die Zuordnung sonst dort erraten würde.
func GliedTrifft(gesucht, glied string) bool {
	return gliedtrifft(gesucht, glied)
}
