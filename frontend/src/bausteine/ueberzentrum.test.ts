// **Wem Escape gehört** — die Regel aus `ueberzentrum.ts`, Fall für Fall
// (Zweitblick über Phase 6.3, M1).
import { afterEach, describe, expect, it, onTestFinished } from 'vitest'

import { escapeSchliesst, imUeberZentrum, verdecktVon } from './ueberzentrum'

afterEach(() => {
  document.body.replaceChildren()
})

const escape = (teile: KeyboardEventInit = {}) =>
  new KeyboardEvent('keydown', { key: 'Escape', ...teile })

describe('escapeSchliesst', () => {
  it('schließt eine sichtbare Spalte, wenn der Fokus nirgends steht', () => {
    expect(escapeSchliesst(escape(), true)).toBe(true)
  })

  it('schließt keine ruhende Spalte', () => {
    expect(escapeSchliesst(escape(), false)).toBe(false)
  })

  it('schließt nicht während einer Eingabe im Werden', () => {
    expect(escapeSchliesst(escape({ isComposing: true }), true)).toBe(false)
  })

  it('schließt nicht auf eine andere Taste', () => {
    expect(escapeSchliesst(escape({ key: 'Enter' }), true)).toBe(false)
  })

  it('schließt, wenn der Fokus in der Spalte steht, und nicht in einem Feld daneben', () => {
    const spalte = document.createElement('aside')
    spalte.dataset.ueber = 'zentrum'
    const drin = document.createElement('button')
    spalte.append(drin)
    const daneben = document.createElement('input')
    document.body.append(spalte, daneben)

    drin.focus()
    expect(imUeberZentrum(document.activeElement)).toBe(true)
    expect(escapeSchliesst(escape(), true)).toBe(true)

    daneben.focus()
    expect(escapeSchliesst(escape(), true)).toBe(false)
  })

  // **Ohne aktives Element** — WebKit meldet dann `body`, ein Dokument
  // ohne Körper `null`; beides heißt: nirgends.
  it('zählt ein fehlendes aktives Element als nirgends', () => {
    const vorher = Object.getOwnPropertyDescriptor(
      Document.prototype,
      'activeElement',
    )
    Object.defineProperty(document, 'activeElement', {
      configurable: true,
      get: () => null,
    })
    onTestFinished(() => {
      Reflect.deleteProperty(document, 'activeElement')
      if (vorher !== undefined) {
        Object.defineProperty(Document.prototype, 'activeElement', vorher)
      }
    })

    expect(imUeberZentrum(null)).toBe(false)
    expect(escapeSchliesst(escape(), true)).toBe(true)
  })
})

// **Was die Spalte verdeckt** — N12 aus dem Nachlauf über Phase 6.3.
describe('verdecktVon', () => {
  const kasten = (
    links: number,
    oben: number,
    breite: number,
    hoehe: number,
  ) => {
    const element = document.createElement('div')
    element.getBoundingClientRect = () =>
      new DOMRect(links, oben, breite, hoehe)

    return element
  }
  const spalte = kasten(0, 0, 240, 600)

  it('verdeckt, wessen Mitte unter ihr liegt', () => {
    expect(verdecktVon(kasten(20, 10, 180, 30), spalte)).toBe(true)
  })

  it('verdeckt nicht, was daneben liegt', () => {
    expect(verdecktVon(kasten(400, 10, 200, 30), spalte)).toBe(false)
  })

  it('verdeckt nicht, was nur mit dem Rand unter ihr liegt', () => {
    expect(verdecktVon(kasten(200, 10, 200, 30), spalte)).toBe(false)
  })

  it('verdeckt nicht, was darunter oder darüber liegt', () => {
    expect(verdecktVon(kasten(20, 700, 180, 30), spalte)).toBe(false)
    expect(verdecktVon(kasten(20, -100, 180, 30), spalte)).toBe(false)
  })

  it('verdeckt kein Element ohne Fläche', () => {
    expect(verdecktVon(kasten(20, 10, 0, 30), spalte)).toBe(false)
    expect(verdecktVon(kasten(20, 10, 180, 0), spalte)).toBe(false)
  })
})
