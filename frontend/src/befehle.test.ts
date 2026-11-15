// **Parameter, Argumente und der bereinigte Verlauf** — Werkstück U12 von
// Phase 5.1 (#1423).
import { describe, expect, it } from 'vitest'
import {
  argumenteAus,
  bereinigt,
  geheim,
  huebsch,
  parameterAus,
} from './befehle'

const schema = {
  type: 'object',
  properties: {
    set: { type: 'string', description: 'die Kennung' },
    anzahl: { type: 'integer' },
    trotzdem: { type: 'boolean' },
    art: { type: 'string', enum: ['png', 'svg'] },
    zeilen: { type: 'array', items: { type: 'object' } },
    frei: {},
  },
  required: ['set', 'anzahl', 'trotzdem'],
}

describe('die Parameter eines Schemas', () => {
  it('liest Art, Pflicht, Erklärung und Werte', () => {
    const parameter = parameterAus(schema)

    expect(parameter.map((p) => [p.name, p.art, p.pflicht])).toEqual([
      ['set', 'text', true],
      ['anzahl', 'zahl', true],
      ['trotzdem', 'wahrheit', true],
      ['art', 'wahl', false],
      ['zeilen', 'json', false],
      ['frei', 'text', false],
    ])
    expect(parameter[0]?.erklaerung).toBe('die Kennung')
    expect(parameter[3]?.werte).toEqual(['png', 'svg'])
  })

  it('kommt ohne Eigenschaften und mit Typenlisten zurecht', () => {
    expect(parameterAus({ type: 'object' })).toEqual([])
    expect(parameterAus({ properties: null })).toEqual([])
    expect(
      parameterAus({
        properties: { a: { type: ['number', 'null'] } },
        required: [1],
      }),
    ).toEqual([
      { name: 'a', art: 'zahl', pflicht: false, erklaerung: '', werte: [] },
    ])
  })
})

describe('die Argumente aus den Werten', () => {
  const parameter = parameterAus(schema)

  it('baut das Objekt und lässt Leeres weg', () => {
    const { argumente, befunde } = argumenteAus(parameter, {
      set: ' 7 ',
      anzahl: '3',
      trotzdem: true,
      art: 'svg',
      zeilen: '[{"a":1}]',
      frei: '',
    })

    expect(befunde).toEqual([])
    expect(argumente).toEqual({
      set: '7',
      anzahl: 3,
      trotzdem: true,
      art: 'svg',
      zeilen: [{ a: 1 }],
    })
  })

  it('nennt jedes Feld, das nicht taugt — und Shellzeichen bleiben Daten', () => {
    const { argumente, befunde } = argumenteAus(parameter, {
      set: 'a"; rm -rf / # $(x)',
      anzahl: 'drei',
      zeilen: '{',
    })

    // Eine Pflicht-Wahrheit ohne Haken ist falsch, kein Befund — sie fehlt nie.
    expect(befunde).toEqual([
      { name: 'anzahl', grund: 'zahl' },
      { name: 'zeilen', grund: 'json' },
    ])
    expect(argumente.set).toBe('a"; rm -rf / # $(x)')

    // Ein Haken in einem Textfeld ist kein Text — er fehlt.
    expect(argumenteAus(parameter, { set: true, anzahl: '1' }).befunde).toEqual(
      [{ name: 'set', grund: 'pflicht' }],
    )

    // Eine Pflicht-Wahrheit ohne Haken ist falsch, keine fehlende.
    expect(
      argumenteAus(parameter, { set: 'x', anzahl: '1' }).argumente,
    ).toEqual({
      set: 'x',
      anzahl: 1,
      trotzdem: false,
    })
  })
})

describe('der bereinigte Verlauf', () => {
  it('verdeckt Geheimnisse und lässt Daten stehen', () => {
    expect(geheim('token')).toBe(true)
    expect(geheim('Kennwort')).toBe(true)
    expect(geheim('set')).toBe(false)
    expect(bereinigt({ token: 'mse_x', set: '7' })).toEqual({
      token: '•••',
      set: '7',
    })
  })

  it('richtet JSON aus und lässt Text, wie er ist', () => {
    expect(huebsch('{"a":1}')).toEqual({ json: true, text: '{\n  "a": 1\n}' })
    expect(huebsch('kein json')).toEqual({ json: false, text: 'kein json' })
  })
})
