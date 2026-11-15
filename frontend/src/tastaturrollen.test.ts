// **Pfeiltasten rollen einen Bereich** — N3 des Nachlaufs über Phase 6.13
// (#1906).
import { afterEach, expect, it, vi } from 'vitest'

import { installiereTastaturrollen } from './tastaturrollen'

let abbauen: (() => void) | undefined

afterEach(() => {
  abbauen?.()
  document.body.innerHTML = ''
})

// bereich baut einen Rollbereich, der breiter ist als sein Fenster.
function bereich(
  teile: { rolle?: string; tabstelle?: number; breit?: boolean } = {},
) {
  const element = document.createElement('div')
  if (teile.rolle !== undefined) element.setAttribute('role', teile.rolle)
  element.tabIndex = teile.tabstelle ?? 0
  Object.defineProperty(element, 'scrollWidth', {
    value: teile.breit === false ? 100 : 2400,
  })
  Object.defineProperty(element, 'clientWidth', { value: 240 })
  const rollen = vi.fn()
  element.scrollBy = rollen
  document.body.append(element)

  return { element, rollen }
}

function taste(
  element: HTMLElement,
  key: string,
  mehr: KeyboardEventInit = {},
) {
  const ereignis = new KeyboardEvent('keydown', {
    key,
    bubbles: true,
    cancelable: true,
    ...mehr,
  })
  element.dispatchEvent(ereignis)

  return ereignis
}

it('rollt einen benannten Bereich mit Tabstelle nach rechts und links', () => {
  abbauen = installiereTastaturrollen(document)
  const { element, rollen } = bereich({ rolle: 'region' })

  const rechts = taste(element, 'ArrowRight')
  taste(element, 'ArrowLeft')

  expect(rollen).toHaveBeenNthCalledWith(1, { left: 40 })
  expect(rollen).toHaveBeenNthCalledWith(2, { left: -40 })
  expect(rechts.defaultPrevented).toBe(true)
})

it('lässt alles andere in Ruhe', () => {
  abbauen = installiereTastaturrollen(document)
  const ohneRolle = bereich()
  const ohneTabstelle = bereich({ rolle: 'region', tabstelle: -1 })
  const schmal = bereich({ rolle: 'region', breit: false })
  const mitTaste = bereich({ rolle: 'region' })

  taste(ohneRolle.element, 'ArrowRight')
  taste(ohneTabstelle.element, 'ArrowRight')
  taste(schmal.element, 'ArrowRight')
  taste(mitTaste.element, 'ArrowRight', { metaKey: true })
  taste(mitTaste.element, 'ArrowDown')

  for (const fall of [ohneRolle, ohneTabstelle, schmal, mitTaste]) {
    expect(fall.rollen).not.toHaveBeenCalled()
  }
})

it('hört nach dem Abbauen auf', () => {
  installiereTastaturrollen(document)()
  const { element, rollen } = bereich({ rolle: 'region' })

  taste(element, 'ArrowRight')

  expect(rollen).not.toHaveBeenCalled()
})
