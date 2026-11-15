// The texts of the **export assistant** — where to, and what each
// destination needs.
//
// Split out of `seiten.ts` when the assistant brought the file close to the
// 400-line limit (§2, UXR-09, #1917).

export const ausgeben = {
  titel: 'Export',
  erklaerung:
    'Five ways to get a set out of this application — and what each one is for.',
  wohin: 'Where to?',
  keinZiel: 'No destination chosen yet',
  nichtGemerkt:
    'The target applies, but it was not remembered — it will not be preselected next time:',
  keinZielErklaerung:
    'Choose under “Where to?” in which form the set should go out — what to set for it then appears here.',
  ohneKarten:
    'Images and the print sheet would have nothing to show. Cards come from “Cards” — created, pasted or imported.',
  ziele: {
    setdatei: {
      name: 'Set file',
      kurz: 'The set itself, reopenable — an `.mse-set` file that Magic Set Editor 2 reads too.',
      lang: 'The shortcut is ⌘S. **Not in it** are decks, copies, prices, tags and the origin of individual values — they stay in the database, because MSE2 drops unknown entries when it saves.',
    },
    bilder: {
      kurz: 'One image file per card, to use elsewhere — a forum, a spreadsheet, a game program.',
    },
    druckbogen: {
      kurz: 'The cards side by side on paper, with cut marks.',
    },
    vorlage: {
      kurz: 'Text instead of image: a list, a table, a file for another program.',
      lang: 'Which templates exist is up to the installed package.',
    },
    mtgjson: {
      kurz: 'The card format from mtgjson.com — for tools that read it.',
      lang: 'Not to keep working in this application: whatever MTG JSON cannot hold is listed after the export.',
    },
  },
  vorlagenEingrenzen: 'Filter templates',
  weitereOptionen: 'More options ({zahl})',
  // **„What you will get"** — der Schritt vor dem Schreiben (UXR-F7, #2015).
  probelauf: {
    ueberschrift: 'What you will get',
    rechnet: 'Calculating …',
    hinweis: 'Nothing is written until you press the button below.',
    bogen: {
      jeSeite: '{zahl} per page · {seiten}',
      masse: '{karten} · paper {breite} × {hoehe} mm · {dpi} dpi',
      letzte: 'On the last page: {karten}.',
    },
    vorlage: {
      datei: 'File: {name} ({groesse})',
      karten: 'Evaluated over {karten}.',
      beigaben: 'Files next to it: {zahl} (folder “{ordner}”, {groesse})',
      mindestens: 'at least {groesse}',
      befunde: 'Noticed during evaluation:',
    },
  },
}
