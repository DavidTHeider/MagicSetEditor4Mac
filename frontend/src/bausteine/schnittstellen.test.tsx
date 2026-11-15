// **Kommandozeile und Werkzeugserver per Knopf** — OPS-09, Schnitt 1
// (#1907).
//
// AK-01: vorher stehen Datei und Zeilen. AK-03: jeder Stand ist benannt, und
// nur, was geht, ist anklickbar.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Schnittstellen } from './Schnittstellen'
import type { SchnittstellenAntwort } from '../antworten'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.technik.schnittstellen

function stand(codex: string, befehlszeile = 'offen'): SchnittstellenAntwort {
  return {
    schnittstellen: [
      {
        name: 'befehlszeile',
        art: '',
        stand: befehlszeile,
        datei: '/Users/ich/.zprofile',
        zeilen: [
          '# >>> MagicSetEditor4Mac',
          'export PATH="$HOME/.local/bin:$PATH"',
          '# <<< MagicSetEditor4Mac',
        ],
      },
      {
        name: 'codex',
        art: '',
        stand: codex,
        datei: '/Users/ich/.codex/config.toml',
        zeilen: [
          '# >>> MagicSetEditor4Mac',
          '[mcp_servers.magicseteditor4mac]',
          '# <<< MagicSetEditor4Mac',
        ],
      },
    ],
    nurEintrag: '',
    fehler: null,
  }
}

function zeige() {
  return render(
    <Sprachraum sprache="de">
      <Schnittstellen />
    </Sprachraum>,
  )
}

function zeileVon(name: string) {
  return screen.getByRole('listitem', { name })
}

it('zeigt Datei und Zeilen vorher und richtet ein', async () => {
  const einrichten = vi.fn(() => Promise.resolve(stand('eingerichtet')))
  belegeApi({
    Schnittstellen: () => Promise.resolve(stand('offen')),
    SchnittstelleEinrichten: einrichten,
  })

  zeige()

  const codex = await screen.findByRole('listitem', {
    name: eigene.namen.codex,
  })
  expect(within(codex).getByText(eigene.staende.offen)).toBeInTheDocument()
  // **Das Zuhause als `~`** seit UXR-19 (#1927); der ganze Pfad steht im
  // Titel, `schnittstellenkarte.test.tsx` hält beides.
  expect(within(codex).getByText('~/.codex/config.toml')).toBeInTheDocument()
  expect(
    within(codex).getByText(/\[mcp_servers\.magicseteditor4mac\]/),
  ).toBeInTheDocument()

  await userEvent.click(
    within(codex).getByRole('button', { name: eigene.einrichten }),
  )

  expect(einrichten).toHaveBeenCalledWith('codex')
  expect(
    await within(zeileVon(eigene.namen.codex)).findByText(
      eigene.staende.eingerichtet,
    ),
  ).toBeInTheDocument()
})

it('widerruft nur, was eigen ist, und richtet nichts Fremdes ein', async () => {
  const widerrufen = vi.fn(() => Promise.resolve(stand('offen', 'offen')))
  belegeApi({
    Schnittstellen: () => Promise.resolve(stand('fremd', 'eingerichtet')),
    SchnittstelleWiderrufen: widerrufen,
  })

  zeige()

  const codex = await screen.findByRole('listitem', {
    name: eigene.namen.codex,
  })
  // **Nicht gesperrt, sondern nicht da** — seit UXR-19 (#1927) steht nur,
  // was etwas tut; AK-03 von OPS-09 („nur, was geht, ist anklickbar") hält
  // das ebenso, und die Karte trägt keine zwei toten Knöpfe mehr.
  expect(
    within(codex).queryByRole('button', { name: eigene.einrichten }),
  ).toBeNull()
  expect(
    within(codex).queryByRole('button', { name: eigene.widerrufen }),
  ).toBeNull()

  const befehlszeile = zeileVon(eigene.namen.befehlszeile)
  await userEvent.click(
    within(befehlszeile).getByRole('button', { name: eigene.widerrufen }),
  )

  expect(widerrufen).toHaveBeenCalledWith('befehlszeile')
  expect(
    await within(zeileVon(eigene.namen.befehlszeile)).findByText(
      eigene.staende.offen,
    ),
  ).toBeInTheDocument()
})

it('nennt, was beim Einrichten scheitert, und einen Stand, den es nicht kennt, roh', async () => {
  belegeApi({
    Schnittstellen: () => Promise.resolve(stand('neu')),
    SchnittstelleEinrichten: () =>
      Promise.resolve({
        schnittstellen: stand('offen').schnittstellen,
        nurEintrag: '',
        fehler: {
          code: 'konflikt',
          schluessel: '',
          meldung: 'ein fremder Eintrag',
        },
      }),
  })

  zeige()

  const befehlszeile = await screen.findByRole('listitem', {
    name: eigene.namen.befehlszeile,
  })
  expect(
    within(zeileVon(eigene.namen.codex)).getByText('neu'),
  ).toBeInTheDocument()

  await userEvent.click(
    within(befehlszeile).getByRole('button', { name: eigene.einrichten }),
  )

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'ein fremder Eintrag',
  )
})

it('zeigt ein Ziel, das es nicht kennt, mit seinem Namen und sagt, wenn das Lesen scheitert', async () => {
  belegeApi({
    Schnittstellen: () =>
      Promise.resolve({
        schnittstellen: [
          { name: 'spaeter', art: '', stand: 'offen', datei: '/x', zeilen: [] },
        ],
        nurEintrag: '',
        fehler: null,
      }),
  })

  const { unmount } = zeige()

  expect(
    await screen.findByRole('listitem', { name: 'spaeter' }),
  ).toBeInTheDocument()
  unmount()

  belegeApi({
    Schnittstellen: () =>
      Promise.resolve({
        schnittstellen: [],
        nurEintrag: '',
        fehler: { code: 'unbekannt', schluessel: '', meldung: 'kein Zuhause' },
      }),
  })

  zeige()

  expect(await screen.findByRole('alert')).toHaveTextContent('kein Zuhause')
})
