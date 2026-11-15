// **Die Vorrichtungen der bezogenen Stände** — geteilt von
// `bezugsstaende.test.tsx` und `bezugsvorschau.test.tsx` (G04,
// #2251),
// seit die Läufe über die Zeilengrenze aus §2 gewachsen waren.
import { belegeSeite, oeffneKasten, texte, zeige } from './paketseite'

import type { Api } from '../api'
import type { BezugsvorschauAntwort, PaketbezugDTO } from '../paketbezuege'

export const t = texte.seiten.pakete.bezuege

export const bezogen = 'a'.repeat(40)
export const spitze = 'c8e38c9b138c5e9c66199fd099a353bce6ff5ec7'

export function bezug(teile: Partial<PaketbezugDTO> = {}): PaketbezugDTO {
  return {
    name: 'M15-Magic-Pack',
    adresse: 'https://github.com/MagicSetEditorPacks/M15-Magic-Pack',
    stand: bezogen,
    geprueft: bezogen,
    pakete: 324,
    vorige: false,
    vorigerStand: '',
    lage: 'ungefragt',
    spitze: '',
    spitzeZeit: '',
    vorsprung: 0,
    fehler: null,
    ...teile,
  }
}

export function vorschau(
  teile: Partial<BezugsvorschauAntwort> = {},
): BezugsvorschauAntwort {
  return {
    name: 'M15-Magic-Pack',
    adresse: '',
    von: bezogen,
    zu: spitze,
    lage: 'ahead',
    vorsprung: 63,
    rueckstand: 0,
    pakete: [
      { paket: 'magic-m15.mse-style', neu: 1, geaendert: 2, entfernt: 0 },
    ],
    uebrige: 4,
    gekuerzt: false,
    sets: ['Mein Set'],
    bestaetigungNoetig: false,
    gruende: [],
    vorigeErsetzt: false,
    ...teile,
  }
}

export async function aufschlagen(belegung: Partial<Api>) {
  belegeSeite(belegung)
  zeige()
  await oeffneKasten(t.ueberschrift)
}
