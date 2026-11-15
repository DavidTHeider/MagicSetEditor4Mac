// **⌘C, ⌘X und ⌘V über Karten** — Werkstück 1 von Phase 3.7
// (#1168).
//
// # ported: die drei hängen am **fokussierten** Bedienelement
//
//	#define CUT_COPY_PASTE(op,return) \
//	  int id = focused_control(this); \
//	  if      (id == ID_EDITOR)    { return editor->op();    } \
//	  else if (id == ID_CARD_LIST) { return card_list->op(); } \
//	  else if (id == ID_NOTES)     { return notes->op();     } \
//	  else                         { return false;           }
//
// `src/gui/set/cards_panel.cpp`, Zeilen 371–382, Fassung f9d9356 vom
// 2021-07-06. Deshalb steht diese Deutung an der **Liste** und nicht am
// Fenster: wer im Kartentext oder in den Notizen steht und ⌘C drückt, meint
// den Text — und bekommt ihn, weil der Browser ihn selbst kopiert.
//
// # adapted: ⌘V gewinnt hier **nicht** über den Fokus
//
//	// always alow pasting cards, even if something else is selected
//	bool CardsPanel::canPaste() const {
//	  if (card_list->canPaste()) return true;
//
// (dieselbe Datei, Zeilen 384–390). MSE2 kann das, weil es das Klemmbrett nach
// seinem **Format** fragt (`CardsDataObject::format`). Über die Wails-Naht
// steht nur Text zur Verfügung; ob er Karten trägt, sagt erst das Zerteilen —
// und das ist bereits das Einfügen. Ein ⌘V im Kartentext bliebe damit
// entweder immer beim Text oder verschluckte ihn immer.
//
// **Getrennt vom Ereignis**, wie `rahmen/tastatur.ts` und aus demselben Grund:
// eine Tabelle aus Taste und Bedeutung ist zu lesen, ein Test über
// `fireEvent.keyDown` nicht.

// Ablagetat ist, was eine Tastenkombination über Karten bedeutet.
export type Ablagetat = 'kopieren' | 'ausschneiden' | 'einfuegen'

// Ablagelage ist alles, was die Deutung braucht.
export interface Ablagetastenlage {
  taste: string
  meta: boolean

  // imEingabefeld sagt, ob der Fokus in einem Textfeld steht.
  //
  // **Auch innerhalb der Liste gibt es eines**: die Notizen stehen unter ihr,
  // und ein Feld, das eine Karte umbenennt, kann dazukommen. Wer die Frage
  // hier nicht stellt, nimmt dem Textfeld sein ⌘C.
  imEingabefeld: boolean

  // nurLesend sagt, dass hier nur gelesen werden darf — dann bleibt ⌘C, und
  // ⌘X und ⌘V bedeuten nichts (#2288): sie schreiben ins Set, und im Menü
  // sind sie für einen Leser gesperrt.
  nurLesend?: boolean
}

export function ablagetaste(lage: Ablagetastenlage): Ablagetat | null {
  if (!lage.meta || lage.imEingabefeld) {
    return null
  }

  const gewohnt: Record<string, Ablagetat> = {
    c: 'kopieren',
    x: 'ausschneiden',
    v: 'einfuegen',
  }

  const tat = gewohnt[lage.taste.toLowerCase()] ?? null

  return lage.nurLesend === true && tat !== 'kopieren' ? null : tat
}
