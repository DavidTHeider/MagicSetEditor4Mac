// **Die Karte einer Schnittstelle, ruhiger** — UXR-19
// (#1927).
//
// # Der Anlass
//
// **Beobachtet an der Hüllenprobe `ops09b`** (#1961), festgehalten im Issue:
// fünf gleich hohe Karten mit je zwei Knöpfen, der Stand als Fließtext,
// und `/Users/…/Library/Application Support/Claude/claude_desktop_config.json`
// brach über zwei Zeilen.
//
// # Was er hält
//
// Der Stand ist eine **Marke**, ein Satz steht nur, wo er mehr sagt; das
// Zuhause steht als `~` und der ganze Pfad im Titel; und es stehen **nur die
// Knöpfe** da, die etwas tun können.
import { render, screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { Schnittstellen } from './Schnittstellen'
import type { SchnittstellenAntwort } from '../antworten'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.technik.schnittstellen

// ziele sind fünf Karten, je Stand eine.
function ziele(): SchnittstellenAntwort {
  const eine = (name: string, stand: string, datei: string) => ({
    name,
    art: '',
    stand,
    datei,
    zeilen: [],
  })

  return {
    schnittstellen: [
      eine('befehlszeile', 'eingerichtet', '/Users/ich/.zprofile'),
      eine('codex', 'offen', '/Users/ich/.codex/config.toml'),
      eine(
        'claudedesktop',
        'veraltet',
        '/Users/ich/Library/Application Support/Claude/claude_desktop_config.json',
      ),
      eine('claudecode', 'fremd', '/Users/ich/.claude.json'),
      eine('dienst', 'nichtGefunden', '/opt/dienst.plist'),
    ],
    nurEintrag: '',
    fehler: null,
  }
}

async function zeige() {
  belegeApi({ Schnittstellen: () => Promise.resolve(ziele()) })

  render(
    <Sprachraum sprache="de">
      <Schnittstellen />
    </Sprachraum>,
  )

  await screen.findByRole('listitem', { name: eigene.namen.codex })
}

const karte = (name: keyof typeof eigene.namen) =>
  screen.getByRole('listitem', { name: eigene.namen[name] })

const knoepfe = (name: keyof typeof eigene.namen) =>
  within(karte(name))
    .queryAllByRole('button')
    .map((knopf) => knopf.textContent)

describe('die Karte einer Schnittstelle', () => {
  it('zeigt nur die Knöpfe, die etwas tun', async () => {
    await zeige()

    expect(knoepfe('befehlszeile')).toEqual([eigene.widerrufen])
    expect(knoepfe('codex')).toEqual([eigene.einrichten])
    expect(knoepfe('claudedesktop')).toEqual([
      eigene.einrichten,
      eigene.widerrufen,
    ])
    expect(knoepfe('claudecode')).toEqual([])
    expect(knoepfe('dienst')).toEqual([])
  })

  it('zeigt den Stand als Marke und den Satz nur, wo er mehr sagt', async () => {
    await zeige()

    expect(
      within(karte('claudedesktop')).getByText(eigene.staende.veraltet),
    ).toHaveClass('rounded-full')
    expect(
      within(karte('claudedesktop')).getByText(eigene.staendeSaetze.veraltet),
    ).toBeVisible()
    expect(
      within(karte('claudecode')).getByText(eigene.staendeSaetze.fremd),
    ).toBeVisible()

    // **Eingerichtet braucht keinen Satz** — die Marke sagt alles.
    expect(
      within(karte('befehlszeile')).getByText(eigene.staende.eingerichtet),
    ).toHaveClass('rounded-full')
    for (const satz of Object.values(eigene.staendeSaetze)) {
      expect(karte('befehlszeile')).not.toHaveTextContent(satz)
    }
  })

  it('schreibt das Zuhause als ~ und nennt den ganzen Pfad im Titel', async () => {
    await zeige()

    expect(
      within(karte('claudedesktop')).getByText(
        '~/Library/Application Support/Claude/claude_desktop_config.json',
      ),
    ).toHaveAttribute(
      'title',
      '/Users/ich/Library/Application Support/Claude/claude_desktop_config.json',
    )

    // **Ein Pfad außerhalb des Zuhauses bleibt, wie er ist.**
    expect(
      within(karte('dienst')).getByText('/opt/dienst.plist'),
    ).toBeInTheDocument()
  })
})
