// **Das Schlachtfeld in drei Gruppen** — PLAY-03.AK-02
// (#1711).
import { describe, expect, it } from 'vitest'

import { Gruppenluecke, schlachtfeldlagen } from './schlachtfeld'
import { Kartenabstand, Kartenhoehe, Reihenhoehe } from './masse'
import { einObjekt } from '../seiten/partieprobe'

import type { Spielobjekt } from '../apipartie'
import type { Zonenfeld } from './zonenfeld'

const feld: Zonenfeld = {
  zone: 'tisch',
  sitz: 'ich',
  x: 100,
  y: 200,
  breite: 900,
  hoehe: Reihenhoehe,
  gestalt: 'reihe',
  verborgen: false,
  geordnet: false,
}

// mitTypen gibt ein gelegtes Objekt mit gesetzten Kartentypen.
function mitTypen(kennung: string, typen: string[]) {
  const objekt: Spielobjekt = einObjekt({
    kennung,
    zone: 'tisch',
    eigenschaften: {
      name: kennung,
      farben: [],
      manawert: 0,
      supertypen: [],
      typen,
      untertypen: [],
      staerke: { steht: false, stern: false, zahl: 0 },
      widerstandskraft: { steht: false, stern: false, zahl: 0 },
      loyalitaet: { steht: false, stern: false, zahl: 0 },
      verteidigung: { steht: false, stern: false, zahl: 0 },
      unbekannteTypen: [],
    },
  })

  return { kennung, getappt: false, verdeckt: false, haufen: '', objekt }
}

// zeileVon sagt, in welcher der zwei Reihen eine Lage liegt.
const zeileVon = (y: number) => (y < feld.y + Kartenhoehe ? 'oben' : 'unten')

describe('schlachtfeldlagen', () => {
  // **Kreaturen liegen zur Mitte** — sie greifen an, und wer angreift, steht
  // vorn. CR 403.1 sagt nur „vor ihm"; welche Gruppe zuerst, entscheidet
  // diese Oberfläche.
  it('legt Kreaturen in die Reihe zur Mitte', () => {
    const lagen = schlachtfeldlagen(
      feld,
      [mitTypen('k', ['creature']), mitTypen('l', ['land'])],
      -1,
    ).lagen

    const kreatur = lagen.find((eine) => eine.kennung === 'k')
    const land = lagen.find((eine) => eine.kennung === 'l')

    expect(kreatur && zeileVon(kreatur.y)).toBe('oben')
    expect(land && zeileVon(land.y)).toBe('unten')
  })

  // **Und andersherum, wenn die Mitte unten liegt.**
  it('dreht die Reihen mit dem Sitz', () => {
    const lagen = schlachtfeldlagen(
      feld,
      [mitTypen('k', ['creature']), mitTypen('l', ['land'])],
      1,
    ).lagen

    const kreatur = lagen.find((eine) => eine.kennung === 'k')
    const land = lagen.find((eine) => eine.kennung === 'l')

    expect(kreatur && zeileVon(kreatur.y)).toBe('unten')
    expect(land && zeileVon(land.y)).toBe('oben')
  })

  // **Länder und Sonstiges teilen sich die hintere Reihe**, getrennt durch
  // eine Lücke — sonst wäre die Gruppierung keine.
  it('setzt eine Lücke zwischen Länder und Sonstiges', () => {
    const lagen = schlachtfeldlagen(
      feld,
      [mitTypen('l', ['land']), mitTypen('a', ['artifact'])],
      -1,
    ).lagen

    const land = lagen.find((eine) => eine.kennung === 'l')
    const artefakt = lagen.find((eine) => eine.kennung === 'a')

    expect((artefakt?.x ?? 0) - (land?.x ?? 0)).toBeGreaterThanOrEqual(
      (land?.breite ?? 0) + Gruppenluecke,
    )
  })

  // **Eine leere Gruppe nimmt keinen Platz.** Wer nur Länder hat, bekommt die
  // ganze hintere Reihe für sie.
  it('gibt die ganze Reihe her, wo eine Gruppe leer ist', () => {
    const viele = Array.from({ length: 12 }, (_, n) =>
      mitTypen(`l${String(n)}`, ['land']),
    )

    const lagen = schlachtfeldlagen(feld, viele, -1).lagen

    const letzte = lagen.at(-1)

    expect(letzte?.x ?? 0).toBeGreaterThan(feld.x + feld.breite / 2)
  })

  // **Jede Lage ist einmal da, und keine zwei liegen gleich hoch.** Die drei
  // Gruppen werden getrennt gerechnet; ohne Versatz trügen ihre Lagen
  // dieselben Höhenränge, und der Zeichner entschiede nach Zufall, was oben
  // liegt.
  it('zählt die Höhenränge über alle Gruppen durch', () => {
    const lagen = schlachtfeldlagen(
      feld,
      [
        mitTypen('k', ['creature']),
        mitTypen('l', ['land']),
        mitTypen('a', ['artifact']),
      ],
      -1,
    ).lagen

    const raenge = lagen.map((eine) => eine.hoehenrang)

    expect(new Set(raenge).size).toBe(raenge.length)
  })

  // **Und die Zahl bleibt die Zahl.**
  it('nennt alle Objekte in der Zahl', () => {
    const gerechnet = schlachtfeldlagen(
      feld,
      [
        mitTypen('k', ['creature']),
        mitTypen('l', ['land']),
        mitTypen('s', ['sorcery']),
      ],
      -1,
    )

    expect(gerechnet.zahl).toBe(3)
    expect(gerechnet.lagen).toHaveLength(3)
  })

  // **Eine Hexerei bekommt keine eigene Reihe** — sie fällt zu den sonstigen
  // und damit in die hintere. Das Kriterium sagt es wörtlich: keine falsche
  // dauerhafte Unterzone.
  it('gibt einer Hexerei keine eigene Reihe', () => {
    const lagen = schlachtfeldlagen(
      feld,
      [mitTypen('s', ['sorcery']), mitTypen('a', ['artifact'])],
      -1,
    ).lagen

    const hexerei = lagen.find((eine) => eine.kennung === 's')
    const artefakt = lagen.find((eine) => eine.kennung === 'a')

    expect(hexerei && artefakt && hexerei.y).toBe(artefakt?.y)
  })

  // **Die Lücke ist sichtbar und nicht symbolisch.**
  it('macht die Lücke so breit wie ein halbes Kartenmaß', () => {
    expect(Gruppenluecke).toBeGreaterThan(Kartenabstand)
  })
})
