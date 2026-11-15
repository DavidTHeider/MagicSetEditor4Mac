// **Ein Leerzustand sagt in höchstens 140 Zeichen, was als Nächstes kommt** —
// UX-15 (#1771):
// kein Absatz im ersten Blick über 140 Zeichen. QA01
// (#2382)
// fand den der Turniere mit 179 in der WebKit-Abnahme `ux15-seitenlaengen` —
// und drei weitere, die sie nie sah: *Decks*, *Sammlung* und die
// Symbolbibliothek sind im Prüfbestand nicht leer. Genau diese Texte sieht
// ein neuer Nutzer zuerst.
//
// # Was er hält
//
// Jeder Schlüssel `leer` oder `…Leer` im **deutschen** Katalog. **Was er
// nicht hält**: die übrigen Sprachen — ihre Länge folgt der Übersetzung, und
// Französisch liegt verlässlich darüber — und Absätze, die nicht so heißen;
// die misst `ux15-seitenlaengen` am gerenderten Fenster.
import { expect, it } from 'vitest'

import { kataloge } from './kataloge'

const grenze = 140

function leertexte(knoten: unknown, pfad: string, funde: string[]): void {
  if (typeof knoten === 'string') {
    const name = pfad.split('.').at(-1) ?? ''
    if (/^leer$|Leer$/.test(name) && knoten.length > grenze) {
      funde.push(`${pfad} (${String(knoten.length)})`)
    }

    return
  }

  if (typeof knoten === 'object' && knoten !== null) {
    for (const [schluessel, wert] of Object.entries(knoten)) {
      leertexte(wert, pfad === '' ? schluessel : `${pfad}.${schluessel}`, funde)
    }
  }
}

it('hält jeden Leertext des deutschen Katalogs unter 140 Zeichen', () => {
  const funde: string[] = []
  leertexte(kataloge.de, '', funde)

  expect(funde).toEqual([])
})

it('findet einen zu langen Leertext auch tief im Katalog', () => {
  const funde: string[] = []
  leertexte(
    { seiten: { probe: { leer: 'x'.repeat(141), titel: 'x'.repeat(200) } } },
    '',
    funde,
  )

  expect(funde).toEqual(['seiten.probe.leer (141)'])
})
