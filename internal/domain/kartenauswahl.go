package domain

import "fmt"

// **Eine Teilmenge ausgeben** — Werkstück 4 von Phase 3.6
// (#1167).
//
// # ported, und die Fundstelle ist genau
//
//	enum ExportCardSelectionType
//	{  EXPORT_SEL_ENTIRE_SET
//	,  EXPORT_SEL_SUBSET
//	,  EXPORT_SEL_CUSTOM
//	};
//
// `src/gui/card_select_window.hpp`, Zeilen 20–24, Fassung f9d9356 vom
// 2021-07-06. Die Wahl selbst baut MSE2 so:
//
//	void SetWindow::selectionChoices(ExportCardSelectionChoices& out) {
//	  out.push_back(make_intrusive<ExportCardSelectionChoice>(*set)); // entire set
//	  FOR_EACH(p, panels) { p->selectionChoices(out); }
//	  out.push_back(make_intrusive<ExportCardSelectionChoice>()); // custom
//	}
//
// `src/gui/set/window.cpp`, Zeilen 383–389. MSE2 stellt die Wahl vor **vier** seiner sieben Ausgabewege: Drucken
// (`onFilePrint`), Druckvorschau (`onFilePrintPreview`), HTML-Ausgabe
// (`onFileExportHTML`) und **mehrere Bilder** (`onFileExportImages`) —
// `src/gui/set/window.cpp`, Zeilen 656–712, Fassung f9d9356 vom 2021-07-06.
//
// **Nicht** vor dem Einzelbild (`onFileExportImage` nimmt die aufgeschlagene
// Karte), nicht vor Apprentice und nicht vor MWS — und **nicht** beim
// Speichern des Sets: ein gespeichertes Teilset wäre ein anderes Set.
//
// # Was hier anders ist, und warum
//
// **Die mittlere Wahl gibt es nicht.** `EXPORT_SEL_SUBSET` ist bei MSE2 eine
// Liste, die ein Reiter beisteuert; im Auslieferungsstand tut das genau einer
// — `RandomPackPanel` mit „export generated packs"
// (`src/gui/set/random_pack_panel.cpp`, Zeilen 627–633). Hier gibt es diesen
// Beitrag nicht, und eine Wahl ohne Inhalt wäre eine leere Zeile in jedem
// Ausgabefenster.
//
// **Übrig bleiben zwei**, und sie sind der Nullwert und sein Gegenteil: keine
// Auswahl heißt das ganze Set, eine Auswahl heißt genau diese Karten.

// Kartenauswahl ist die Teilmenge eines Sets, die ausgegeben werden soll.
type Kartenauswahl struct {
	// Karten sind die gewählten Kennungen.
	//
	// **Leer heißt: das ganze Set.** Das ist MSE2s `EXPORT_SEL_ENTIRE_SET`,
	// und es ist der Nullwert — ein Aufrufer, der die Auswahl nicht kennt,
	// bekommt damit das gewohnte Verhalten.
	Karten []int64
}

// Filtert sagt, ob diese Auswahl überhaupt etwas einschränkt.
func (a Kartenauswahl) Filtert() bool { return len(a.Karten) > 0 }

// Anwenden gibt die gewählten Karten — **in der Reihenfolge des Sets**.
//
// # Warum nicht in der Reihenfolge der Nennung
//
// Ein Druckbogen legt die Karten in Setreihenfolge. Wer die Auswahl in ihrer
// eigenen anwendete, bekäme einen Bogen, dessen Karten anders liegen, je
// nachdem, in welcher Reihenfolge jemand angehakt hat — und zweimal
// dieselbe Auswahl gäbe zweimal ein anderes Ergebnis.
//
// **Eine Kennung, die es im Set nicht gibt, fällt weg.** Sie zu melden wäre
// die andere Möglichkeit und die schlechtere: die Auswahl kommt aus einer
// Liste, die jemand vor dem Löschen einer Karte angehakt hat (§8).
func (a Kartenauswahl) Anwenden(karten []*CardRow) []*CardRow {
	if !a.Filtert() {
		return karten
	}

	gewaehlt := make(map[int64]bool, len(a.Karten))
	for _, kennung := range a.Karten {
		gewaehlt[kennung] = true
	}

	genommen := make([]*CardRow, 0, len(gewaehlt))

	for _, karte := range karten {
		if gewaehlt[karte.ID] {
			genommen = append(genommen, karte)
		}
	}

	return genommen
}

// Pruefe weist eine Auswahl ab, die keine Karte des Sets trifft.
//
// **Und nicht: dann eben das ganze Set.** Wer eine Auswahl mitgibt, hat sich
// entschieden; ein Bogen über 300 Karten wäre die teuerste Art, einen
// Tippfehler zu beantworten.
func (a Kartenauswahl) Pruefe(karten []*CardRow) error {
	if !a.Filtert() {
		return nil
	}

	if len(a.Anwenden(karten)) == 0 {
		return fmt.Errorf(
			"%w: keine der %d gewählten Karten gehört zu diesem Set",
			ErrNotFound, len(a.Karten))
	}

	return nil
}
