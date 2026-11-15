// **Woher ein Deck oder ein Exemplar kam** — Werkstück 12 von Phase 4.12
// (#1351).
//
// # Die Frage, die eine Herkunft beantwortet
//
// **„Das gehört mir — aber woher kommt es?"** So steht es in der
// Kommentierung: ein Deck, das künftig getauscht werden kann, und eine
// Sammlung ebenso. Beide gehören danach dem, der sie hat; entstanden sind sie
// woanders.
//
// # Warum sie **fehlen** darf
//
// **Weil keine Herkunft schon eine Auskunft ist**: es ist hier entstanden.
// Eine Zeile für „von mir selbst" trüge nichts bei und wäre die Zeile, die
// beim Anlegen einmal vergessen wird — und dann sähe ein eigenes Deck aus wie
// ein übernommenes. Dieselbe Überlegung steht in
// `0014_wertherkunft.sql`: **eine falsche Nennung ist schlimmer als keine.**
//
// # Warum die Quelle als **Text** mitreist und nicht nur als Kennung
//
// **Weil eine Kennung allein nach einem Löschen nichts mehr sagt.** Wer das
// Profil, aus dem etwas kam, später entfernt, soll weiter lesen können, wie es
// hieß. Dieselbe Entscheidung wie bei [Exemplar] und [Kartenverweis], wo Name,
// Setcode und Nummer neben der Druckkennung stehen.
//
// # Was hier **nicht** steht
//
// **Keine Kette.** Gibt A an B und B an C, dann sagt die Herkunft bei C
// „von B" und nicht „von B, ursprünglich von A". Der Grund steht in
// [0071](../../docs/entscheidungen/0071-woher-ein-deck-kommt.md): eine Kette
// müsste mit dem Deck **mitreisen**, und was mitreist, hat der Gebende
// geschrieben. Was dieser Bestand aufschreibt, ist, was er **selbst gesehen**
// hat.
//
// # Portierungsmarke: **new**
//
// **Und hier ist gesucht worden** (§15): `grep -ri deck` über
// `/Users/david/mse2-quelle/src` gibt am 2026-09-06 **einen** Treffer, und der
// ist ein deutsches Wort in einem Kommentar — MSE2 kennt kein Deck. Über
// `/Users/david/magic-set-generator/src` gibt es ebenfalls einen: ein `TODO`
// über `args.decklists`, nicht umgesetzt. Es gibt kein Altverhalten, das zu
// portieren wäre.

package domain

// Herkunftsart sagt, **wie** etwas hierhergekommen ist.
type Herkunftsart string

const (
	// HerkunftProfil heißt: aus dem Profil eines anderen übernommen.
	HerkunftProfil Herkunftsart = "profil"

	// HerkunftEinfuhr heißt: eingelesen — aus einer Datei oder einer Liste.
	HerkunftEinfuhr Herkunftsart = "einfuhr"
)

// Herkunftsarten gibt beide.
//
// **Als Funktion und nicht als Paketvariable**: eine Aufzählung, die ein
// Aufrufer umschreiben kann, ist keine.
func Herkunftsarten() []Herkunftsart {
	return []Herkunftsart{HerkunftProfil, HerkunftEinfuhr}
}

// Pruefe weist eine erfundene Art ab.
//
// **Auch die leere** — anders als bei [Sichtbarkeit], wo leer `privat`
// bedeutet. Hier gibt es keinen Vorgabewert: wer eine Herkunft anlegt, weiß,
// woher; wer es nicht weiß, legt keine an.
func (a Herkunftsart) Pruefe() error {
	for _, bekannt := range Herkunftsarten() {
		if a == bekannt {
			return nil
		}
	}

	return NewFieldError("art", SchluesselUnbekannteHerkunftsart,
		"Die Herkunft %q kennt dieses Programm nicht.", string(a))
}

// Herkunft sagt, woher ein Deck oder ein Exemplar kam.
type Herkunft struct {
	Art Herkunftsart

	// Quelle ist die **lesbare** Bezeichnung — ein Profilname, ein Dateiname.
	//
	// **Abgeschrieben und nicht nachgeschlagen**: sie hält den Stand des
	// Übernahmezeitpunkts fest. Wer die Quelle später umbenennt, ändert damit
	// nicht, woher etwas kam.
	Quelle string

	// QuellProfil ist die Kennung des Profils, aus dem es kam — 0 bei einer
	// Einfuhr.
	//
	// **Ohne Fremdschlüssel im Bestand** ([0071](../../docs/entscheidungen/0071-woher-ein-deck-kommt.md)):
	// wird das Profil gelöscht, bleibt die Auskunft stehen. Ein `CASCADE`
	// nähme sie weg, ein `RESTRICT` verhinderte das Löschen.
	QuellProfil int64

	// Uebernommen ist der Zeitpunkt in RFC 3339.
	Uebernommen string
}

// Validate prüft eine Herkunft.
func (h Herkunft) Validate() error {
	if fehler := h.Art.Pruefe(); fehler != nil {
		return fehler
	}

	if h.Quelle == "" {
		return NewFieldError("quelle", SchluesselPflichtfeld,
			"Eine Herkunft braucht die Bezeichnung dessen, woher sie kommt.")
	}

	// **Genau eines von beidem** — dieselbe Bauform wie beim [Deckplatz]: die
	// Kennung gehört zur Übernahme aus einem Profil und zu nichts sonst. Wer
	// sie an einer Einfuhr stehen ließe, behauptete ein Profil, das an dem
	// Weg gar nicht beteiligt war.
	if (h.Art == HerkunftProfil) != (h.QuellProfil > 0) {
		return NewFieldError("quellProfil", SchluesselHerkunftOhneProfil,
			"Eine Herkunft aus einem Profil nennt es, eine Einfuhr nennt keines.")
	}

	return nil
}
