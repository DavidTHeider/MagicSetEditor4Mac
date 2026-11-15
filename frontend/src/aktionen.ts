// **Der Aktionskatalog** — UX-10.AK-01
// (#1682),
// nach dem Artboard *BelegMuster* (`docs/prototyp/project/BelegMuster.dc.html`).
//
// # Was er festlegt
//
// Je Aktion das **Symbol**, das **Kürzel** und die **Bestätigung**, die vor
// ihrer Wirkung steht. Kurzlabel und Bedeutung — der Tooltip — stehen in den
// Katalogen unter `bausteine.aktionskatalog`, denn sie sind Sprache (§9); den
// zugänglichen Namen gibt die Aufschrift, und wo sie die Sache nicht nennt,
// `bezeichnung` am Knopf.
//
// # adapted (§15)
//
// MSE2 hält dieselben Angaben **je Befehl**: ein Bild, die Menüaufschrift
// samt Kürzel, einen Hilfesatz, eine Werkzeugaufschrift und einen Tooltip —
// und ob die Werkzeugleiste die Aufschrift neben dem Bild zeigt
// (`make_menu_item_tr` und `add_tool_tr`, `src/gui/util.cpp`, Zeilen 262–300,
// Fassung f9d9356 vom 2021-07-06; die Texte in den Abschnitten `menu`,
// `help`, `tool` und `tooltip` einer `.mse-locale`). Hier gelten sie **je
// Art der Aktion**: ein Löschen sieht überall gleich aus, wo MSE2 jedem
// Befehl sein eigenes Bild gab. Das ist die Abweichung, und sie ist der
// Zweck des Katalogs.
//
// # Acht und nicht sieben
//
// Das Artboard führt „Kopieren — ein zweites daneben". **Hier heißt das
// Duplizieren**, seit Phase 3.7 (`seitenkarten.ts`): „Kopieren" ist die
// Zwischenablage, und beide stehen auf der Kartenseite nebeneinander. Ein Wort
// für zwei Handlungen wäre genau die Mehrdeutigkeit, die dieser Katalog
// verhindern soll — die Abweichung steht in `docs/PROTOTYP-0.6.0.md`.
//
// # Was er nicht zusagt
//
// **Dass jedes Kürzel überall wirkt.** ⌘N legt ein Set an und kein Deck. Ein
// Knopf kündigt sein Kürzel deshalb nur an, wo sein Aufrufer es verdrahtet hat
// (`mitKuerzel`), und `aktionskatalog.test.ts` führt diese Stellen samt dem
// Lauf, der das Kürzel drückt.

export const aktionen = [
  'anlegen',
  'bearbeiten',
  'sichern',
  'zuruecknehmen',
  'duplizieren',
  'kopieren',
  'entfernen',
  'loeschen',
] as const

export type Aktion = (typeof aktionen)[number]

// Bestaetigung ist, was zwischen dem Klick und der Wirkung steht.
//
//   - `keine` — die Wirkung ist umkehrbar oder nimmt nichts weg.
//   - `ruecknahme` — sie geschieht sofort, und ⌘Z holt sie zurück.
//   - `rueckfrage` — eine Frage an der Klickstelle (`Entscheidung`); der
//     Fokus steht auf dem Abbrechen.
//   - `namen` — dieselbe Frage, und bestätigt wird erst mit dem getippten
//     Namen. Für das, was am meisten mitnimmt: ein Set mit allen Karten.
export type Bestaetigung = 'keine' | 'ruecknahme' | 'rueckfrage' | 'namen'

export interface Aktionsart {
  symbol: string

  // tasten ist das Kürzel in der Schreibweise von `aria-keyshortcuts`.
  tasten: string

  // zeichen ist dasselbe Kürzel, wie macOS es in einem Menü schreibt.
  zeichen: string

  // bestaetigung ist die **Vorgabe**. Wo eine Stelle mehr verlangt — ein
  // Set löschen heißt, seinen Namen zu tippen —, steht das in der Tafel der
  // löschenden Bindungen (`aktionskatalog.test.ts`).
  bestaetigung: Bestaetigung
}

export const aktionsarten: Readonly<Record<Aktion, Aktionsart>> = {
  anlegen: {
    symbol: '+',
    tasten: 'Meta+N',
    zeichen: '⌘N',
    bestaetigung: 'keine',
  },
  bearbeiten: {
    symbol: '✎',
    tasten: 'Enter',
    zeichen: '↵',
    bestaetigung: 'keine',
  },
  sichern: {
    symbol: '⤓',
    tasten: 'Meta+S',
    zeichen: '⌘S',
    bestaetigung: 'keine',
  },
  zuruecknehmen: {
    symbol: '⎌',
    tasten: 'Meta+Z',
    zeichen: '⌘Z',
    bestaetigung: 'keine',
  },
  duplizieren: {
    symbol: '⧉',
    tasten: 'Meta+D',
    zeichen: '⌘D',
    bestaetigung: 'keine',
  },
  kopieren: {
    symbol: '⎘',
    tasten: 'Meta+C',
    zeichen: '⌘C',
    bestaetigung: 'keine',
  },
  entfernen: {
    symbol: '−',
    tasten: 'Backspace',
    zeichen: '⌫',
    bestaetigung: 'ruecknahme',
  },
  loeschen: {
    symbol: '✕',
    tasten: 'Shift+Backspace',
    zeichen: '⇧⌫',
    bestaetigung: 'rueckfrage',
  },
}
