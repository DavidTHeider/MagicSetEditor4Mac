// **Drei Gruppen auf dem Schlachtfeld** — PLAY-03.AK-02
// (#1711).
import { describe, expect, it } from 'vitest'

import { einObjekt } from '../seiten/partieprobe'
import { gruppenordnung, permanentengruppe } from './permanentengruppe'

// mitTypen gibt ein Objekt mit gesetzten Kartentypen.
const mitTypen = (typen: string[]) =>
  einObjekt({
    zone: 'tisch',
    eigenschaften: {
      name: 'Probe',
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

describe('permanentengruppe', () => {
  it('legt eine Kreatur zu den Kreaturen', () => {
    expect(permanentengruppe(mitTypen(['creature']))).toBe('kreatur')
  })

  it('legt ein Land zu den Ländern', () => {
    expect(permanentengruppe(mitTypen(['land']))).toBe('land')
  })

  // **Ein Kreaturenland ist eine Kreatur.** Es kann angreifen, und wer
  // angreift, steht vorn. Die Regel sagt dazu nichts — CR 403.1 nennt nur
  // „vor ihm"; die Wahl gehört dieser Oberfläche.
  it('zählt ein Kreaturenland zu den Kreaturen', () => {
    expect(permanentengruppe(mitTypen(['creature', 'land']))).toBe('kreatur')
  })

  it('legt Artefakt und Verzauberung zu den sonstigen', () => {
    expect(permanentengruppe(mitTypen(['artifact']))).toBe('sonstige')
    expect(permanentengruppe(mitTypen(['enchantment']))).toBe('sonstige')
  })

  // **Hexerei und Spontanzauber bekommen keine eigene Gruppe** — das
  // Kriterium sagt es wörtlich. Sie fallen zu den sonstigen, und damit gibt
  // es für sie keinen Platz, den sie behielten.
  it('gibt Hexerei und Spontanzauber keine eigene Gruppe', () => {
    expect(permanentengruppe(mitTypen(['sorcery']))).toBe('sonstige')
    expect(permanentengruppe(mitTypen(['instant']))).toBe('sonstige')
    expect(gruppenordnung).not.toContain('zauber')
  })

  // **Ohne Eigenschaften bleibt nur `sonstige`.** Auf dem Schlachtfeld kommt
  // das nicht vor — es ist eine öffentliche Zone (CR 400.2) —, und ein
  // verdecktes Permanent trägt seine 2/2-Kreatur schon aus dem Kern
  // (CR 708.2a).
  it('nimmt ein Objekt ohne Eigenschaften zu den sonstigen', () => {
    expect(permanentengruppe(einObjekt({ zone: 'tisch' }))).toBe('sonstige')
  })

  // **Die Ordnung ist von hinten nach vorn** — Länder an der eigenen Kante,
  // Kreaturen zur Mitte.
  it('ordnet Länder hinten und Kreaturen vorn', () => {
    expect(gruppenordnung).toEqual(['land', 'sonstige', 'kreatur'])
  })
})
