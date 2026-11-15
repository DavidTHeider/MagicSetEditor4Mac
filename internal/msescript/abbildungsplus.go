package msescript

import "strconv"

// Das **Verketten zweier Abbildungen** — `+`
// (#271).
//
// Eigene Datei, weil `auswerter_verknuepfung.go` mit ihr über 400 Zeilen ginge
// (§2).

// verketten hängt die rechte Abbildung an die linke.
//
//	ScriptValueP ScriptConcatCollection::getMember(const String& name) const {
//	  ScriptValueP member = a->getMember(name);
//	  if (member->type() != SCRIPT_ERROR) return member;
//	  …
//	  return b->getMember(name);
//	}
//
// `src/script/value.cpp`, Zeilen 551–562, Fassung f9d9356 vom 2021-07-06.
// **Die linke Seite wird zuerst gefragt**, und beide Glieder bleiben stehen.
//
// # Nicht „rechts gewinnt"
//
// `map := map + [stylesheet: …]` sieht in den Paketen aus wie ein
// Überschreiben. Am Orakel ist es das Gegenteil
// ([0026](../../docs/entscheidungen/0026-mse2-rendert-die-referenz-selbst.md)):
//
//	to_string(([a: 1] + [a: 2]).a)      → "1"
//	to_string(length([a: 1] + [a: 2]))  → "2"
//
// **Gemessen** (§10): über den Packlauf treffen
// **6.464** Verbindungen zweier Abbildungen auf mindestens einen doppelten
// Schlüssel (#271).
//
// **Ohne Prüfung der Arten.** Der einzige Aufrufer stellt sie schon: `+` ruft
// hier nur für zwei Abbildungen. Ein Wächter wäre ein Zweig, den kein Aufruf
// betreten kann (§3).
func (a *auswerter) verketten(links, rechts Wert) Wert {
	ergebnis := Wert{
		Art:       ArtAbbildung,
		Eintraege: make([]Eintrag, 0, len(links.Eintraege)+len(rechts.Eintraege)),
	}

	// **Zahlenschlüssel werden umnummeriert, nicht überschrieben.**
	//
	//	if (name.ToLong(&index) && index - itemsInA >= 0 && index - itemsInA < b->itemCount()) {
	//	    return b->getMember(String() << (index - itemsInA));
	//
	// `ScriptConcatCollection::getMember`, `src/script/value.cpp`, Zeilen
	// 551–562, Fassung f9d9356 vom 2021-07-06. Zwei Listen **verketten** sich
	// also; nur benannte Schlüssel verschmelzen.
	//
	// Bis #390
	// wurde auch über die Zahlen verschmolzen: `[1] + [2]` ergab **ein** Glied,
	// weil beide „0" heißen. Damit war jede Liste, die eine Vorlage über eine
	// Schleife baut, ein Glied lang.
	zahlen := zaehleZahlschluessel(links)

	ergebnis.Eintraege = append(ergebnis.Eintraege, links.Eintraege...)

	// **Ein benannter Schlüssel wird angehängt, nicht ersetzt.** [Wert.Eintrag]
	// gibt den **ersten** Treffer, und damit gewinnt die linke Seite — so wie
	// `getMember` es tut.
	for i := range rechts.Eintraege {
		eintrag := rechts.Eintraege[i]

		if stelle, istZahl := stellenname(eintrag.Name); istZahl {
			eintrag = Eintrag{Name: strconv.Itoa(stelle + zahlen), Wert: eintrag.Wert}
		}

		ergebnis.Eintraege = append(ergebnis.Eintraege, eintrag)
	}

	return ergebnis
}

// stellenname sagt, ob ein Schlüssel eine Stelle in einer Liste ist.
func stellenname(name string) (int, bool) {
	stelle, err := strconv.Atoi(name)

	return stelle, err == nil
}

// zaehleZahlschluessel gibt die Zahl der Glieder mit einem Stellenschlüssel.
//
// **Nicht die Länge der Abbildung**: eine Abbildung darf beides führen, und
// MSEs `itemCount()` zählt nur die stellenweisen Glieder (`value`, nicht
// `key_value` — `src/script/value.cpp`, `ScriptCustomCollection`).
//
// Fassung f9d9356 vom 2021-07-06.
func zaehleZahlschluessel(wert Wert) int {
	var zahlen int

	for i := range wert.Eintraege {
		if _, istZahl := stellenname(wert.Eintraege[i].Name); istZahl {
			zahlen++
		}
	}

	return zahlen
}
