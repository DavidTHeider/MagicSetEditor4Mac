// **Die zusätzlichen Sprachen beim Start** — LOC-01.AK-04
// (#1687).
//
// Was er hält: eine gespeicherte Sprache gilt nach einem Neustart, ohne Wahl
// gilt die des Systems, und das Dokument trägt die **Schrift** mit, wo sie
// die Glyphen entscheidet. Ein Han-Zeichen hat in Japan, auf dem Festland und
// in Taiwan verschiedene Formen; `lang="zh"` lässt offen, welche WebKit zieht.
//
// Was er nicht hält: wie die Glyphen **aussehen**. Das zeigt nur das Fenster,
// und der Abzug dazu steht im Beleg von LOC-01.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { App } from './App'
import { kataloge, type Sprache } from './sprache/kataloge'
import { belegeApi, raeumeApi } from './test/api'

const urspruenglich = Object.getOwnPropertyDescriptor(
  window.navigator,
  'language',
)

afterEach(() => {
  raeumeApi()
  document.documentElement.lang = 'de'

  if (urspruenglich) {
    Object.defineProperty(window.navigator, 'language', urspruenglich)
  }
})

// starte baut die Anwendung mit einer gespeicherten Sprache und einer des
// Systems.
function starte(gespeichert: string, system: string) {
  Object.defineProperty(window.navigator, 'language', {
    value: system,
    configurable: true,
  })

  belegeApi({
    Einstellungen: () =>
      Promise.resolve({
        einstellungen: [
          {
            schluessel: 'sprache',
            wert: gespeichert,
            wirksam: gespeichert,
            werte: ['de', 'en', 'es', 'fr', 'it', 'ja', 'ru', 'zh'],
            bereich: 'anwendung',
          },
        ],
      }),
    SpracheSetzen: () => Promise.resolve({}),
  })

  render(<App />)
}

// startseite wartet auf die Überschrift der Startseite in einer Sprache.
async function startseite(sprache: Sprache) {
  return screen.findByRole('heading', {
    name: kataloge[sprache].seiten.start.ueberschrift,
  })
}

describe('Die zusätzlichen Sprachen beim Start', () => {
  it('gilt nach einem Neustart in der gespeicherten Sprache', async () => {
    starte('ja', 'de-DE')

    expect(await startseite('ja')).toBeInTheDocument()
    expect(document.documentElement.lang).toBe('ja')
  })

  it('zeichnet Chinesisch mit den vereinfachten Formen', async () => {
    starte('zh', 'de-DE')

    expect(await startseite('zh')).toBeInTheDocument()
    expect(document.documentElement.lang).toBe('zh-Hans')
  })

  it('nimmt ohne Wahl die Sprache des Systems', async () => {
    starte('', 'ru-RU')

    expect(await startseite('ru')).toBeInTheDocument()
    expect(document.documentElement.lang).toBe('ru')
  })

  // **Traditionelles Chinesisch führt das Fenster nicht** — die Entscheidung
  // steht in `docs/SPRACHINVENTAR.md`. Ein System in `zh-TW` bekommt das
  // vereinfachte, nicht das deutsche: näher liegt es allemal, und die Wahl
  // in den Einstellungen nennt die Schrift.
  it('gibt einem traditionell chinesischen System das vereinfachte', async () => {
    starte('', 'zh-TW')

    expect(await startseite('zh')).toBeInTheDocument()
  })
})
