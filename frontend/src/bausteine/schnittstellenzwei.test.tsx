// **Claude Desktop, Claude Code und der Dienst als LaunchAgent** — OPS-09b
// (#1961).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { Schnittstellen } from './Schnittstellen'
import type { SchnittstellenAntwort } from '../antworten'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.technik.schnittstellen

function stand(desktop: string, nurEintrag = ''): SchnittstellenAntwort {
  return {
    schnittstellen: [
      {
        name: 'claudedesktop',
        art: 'json',
        stand: desktop,
        datei:
          '/Users/ich/Library/Application Support/Claude/claude_desktop_config.json',
        zeilen: ['"mcpServers": {', '  "magicseteditor4mac": {}', '}'],
      },
      {
        name: 'claudecode',
        art: 'befehl',
        stand: 'offen',
        datei: '/Users/ich/.claude.json',
        zeilen: ['$ claude mcp add -s user magicseteditor4mac /x/mse-mcp'],
      },
      {
        name: 'dienst',
        art: 'agent',
        stand: 'offen',
        datei:
          '/Users/ich/Library/LaunchAgents/io.github.davidtheider.magicseteditor4mac.dienst.plist',
        zeilen: ['<plist>', '$ launchctl bootstrap gui/501 x'],
      },
    ],
    nurEintrag,
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

it('nennt die drei Ziele und was je Art geschieht', async () => {
  belegeApi({ Schnittstellen: () => Promise.resolve(stand('offen')) })

  zeige()

  const desktop = await screen.findByRole('listitem', {
    name: eigene.namen.claudedesktop,
  })
  expect(within(desktop).getByText(eigene.vorschau.json)).toBeInTheDocument()

  const code = screen.getByRole('listitem', { name: eigene.namen.claudecode })
  expect(within(code).getByText(eigene.vorschau.befehl)).toBeInTheDocument()

  const dienst = screen.getByRole('listitem', { name: eigene.namen.dienst })
  expect(within(dienst).getByText(eigene.vorschau.agent)).toBeInTheDocument()
})

it('sagt, wenn nur der eigene Eintrag ging', async () => {
  belegeApi({
    Schnittstellen: () => Promise.resolve(stand('eingerichtet')),
    SchnittstelleWiderrufen: () =>
      Promise.resolve(stand('offen', 'claudedesktop')),
  })

  zeige()

  const desktop = await screen.findByRole('listitem', {
    name: eigene.namen.claudedesktop,
  })
  await userEvent.click(
    within(desktop).getByRole('button', { name: eigene.widerrufen }),
  )

  expect(
    await screen.findByText(
      eigene.nurEintrag.replace('{name}', eigene.namen.claudedesktop),
    ),
  ).toBeInTheDocument()
})

it('nennt ein unbekanntes Ziel beim Namen, auch im Hinweis', async () => {
  const fremd = stand('offen', 'neu')
  belegeApi({
    Schnittstellen: () => Promise.resolve(fremd),
    SchnittstelleEinrichten: () => Promise.resolve(fremd),
  })

  zeige()

  const desktop = await screen.findByRole('listitem', {
    name: eigene.namen.claudedesktop,
  })
  await userEvent.click(
    within(desktop).getByRole('button', { name: eigene.einrichten }),
  )

  expect(
    await screen.findByText(eigene.nurEintrag.replace('{name}', 'neu')),
  ).toBeInTheDocument()
})

// **Eine abgelehnte Zusage kommt als Meldung an** — seit OPS-10 (#1908) weist
// eine Bindung so ab, wenn ihr Bereich abgeschaltet ist.
it('zeigt eine abgelehnte Zusage als Meldung', async () => {
  belegeApi({
    Schnittstellen: () => Promise.resolve(stand('offen')),
    SchnittstelleEinrichten: () => Promise.reject(new Error('abgewiesen')),
  })

  zeige()

  const desktop = await screen.findByRole('listitem', {
    name: eigene.namen.claudedesktop,
  })
  await userEvent.click(
    within(desktop).getByRole('button', { name: eigene.einrichten }),
  )

  expect(await screen.findByRole('alert')).toBeInTheDocument()
})
