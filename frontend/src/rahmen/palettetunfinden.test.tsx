// **Tun und Finden in der Befehlspalette** — UX-02.AK-02
// (#1674).
//
// # Der Anlass
//
// **Die Palette fand Seiten und Fachbefehle, aber keine Objekte** — kein
// Set, kein Deck, keine Karte über ihren Namen; so stand es im Kommentar zum
// Issue am 2026-09-21. Wer „Aschenwächter" suchte, musste wissen, dass es
// eine Seite *Suche* gibt.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import { Palette } from './Palette'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { deckprobe } from '../seiten/deckprobe'
import { eineListe, einKontakt } from '../seiten/peerkontaktprobe'

import type { Api } from '../api'

const texte = kataloge.de.allgemein.palette

afterEach(() => {
  raeumeApi()
})

function belege(teile: Partial<Api> = {}) {
  const Sets = vi.fn(() =>
    Promise.resolve({
      sets: [
        {
          id: '7',
          name: 'Werkstatt der Dämmerung',
          spiel: 'magic',
          stil: 'm15',
          herkunft: '',
          karten: 41,
          herkunftFehlt: false,
        },
      ],
      fehler: null,
    }),
  )

  belegeApi({
    Befehlskatalog: () => Promise.resolve({ befehle: [], fehler: null }),
    Sets,
    Decks: () =>
      Promise.resolve({
        decks: [deckprobe({ id: '3', name: 'Aschenrat' })],
        fehler: null,
      }),
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([
          einKontakt({
            kennung: 'peer-1',
            eigenerName: 'Turfer',
            stand: 'bestaetigt',
          }),
          einKontakt({ kennung: 'peer-2', eigenerName: 'Fremd' }),
        ]),
      ),
    Suchen: () =>
      Promise.resolve({
        treffer: [
          {
            karteID: '12',
            setID: '7',
            nummer: 4,
            kartenname: 'Aschenwächter',
            feld: 'name',
            wert: 'Aschenwächter',
          },
          {
            karteID: '12',
            setID: '7',
            nummer: 4,
            kartenname: 'Aschenwächter',
            feld: 'rule text',
            wert: 'Asche zu Asche',
          },
        ],
        gekuerzt: false,
        fehler: null,
      }),
    ...teile,
  })

  return { Sets }
}

function zeige(mitWegen = true) {
  const wechseln = vi.fn()
  const schliessen = vi.fn()
  const wege = {
    setID: '7',
    waehleSet: vi.fn(),
    oeffneKarte: vi.fn(),
  }

  render(
    <Sprachraum sprache="de">
      <Anschlussraum aktuell="start" wechseln={wechseln}>
        <Palette
          schliessen={schliessen}
          waehlen={wechseln}
          waehleBefehl={vi.fn()}
          wege={mitWegen ? wege : undefined}
        />
      </Anschlussraum>
    </Sprachraum>,
  )

  return { wechseln, schliessen, wege }
}

// **„Schnittstellen einrichten“ findet, wer danach tippt** — AP-10 (#2367),
// B-19. Ohne Eingabe stehen nur die ersten vier Handlungen da
// (`Ohneeingabe`); „MCP“ oder „REST“ führt hin.
describe('Die Schnittstellen in der Palette', () => {
  it.each(['MCP', 'REST', 'Schnittstellen'])(
    'stehen unter »Tun«, wenn jemand %s tippt',
    async (wort) => {
      belege()
      zeige()

      await userEvent.type(await screen.findByRole('textbox'), wort)

      const tun = await screen.findByRole('list', { name: texte.tun })
      expect(
        await within(tun).findByRole('button', {
          name: new RegExp(
            texte.schnittstellenEinrichten.replace(/[()]/g, '.'),
          ),
        }),
      ).toBeInTheDocument()
    },
  )
})

describe('Tun', () => {
  it('bietet ohne Eingabe die Handlungen an, die gerade gehen', async () => {
    belege()
    zeige()

    const tun = await screen.findByRole('list', { name: texte.tun })
    await within(tun).findByRole('button', { name: /Karten tauschen/ })

    expect(
      within(tun)
        .getAllByRole('button')
        .map((knopf) => knopf.firstChild?.textContent),
    ).toEqual([
      texte.karteErstellen,
      'Schnell spielen mit Aschenrat',
      'Karten tauschen mit Turfer',
      texte.sicherungErstellen,
    ])

    // **Ohne Eingabe wird nichts gefunden** — das wäre der ganze Bestand.
    expect(screen.queryByRole('list', { name: texte.finden })).toBeNull()
  })

  // **Nur bestätigte Gegenüber**: gehandelt wird mit keinem anderen.
  it('lässt ein unbestätigtes Gegenüber weg', async () => {
    belege()
    zeige()

    await userEvent.type(screen.getByRole('textbox'), 'tauschen')

    const tun = await screen.findByRole('list', { name: texte.tun })
    expect(
      within(tun).getByRole('button', { name: /Turfer/ }),
    ).toBeInTheDocument()
    expect(within(tun).queryByRole('button', { name: /Fremd/ })).toBeNull()
  })

  it('schließt die Palette und führt zum Spielen', async () => {
    belege()
    const { wechseln, schliessen } = zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: /Schnell spielen mit Aschenrat/,
      }),
    )

    expect(schliessen).toHaveBeenCalled()
    expect(wechseln).toHaveBeenCalledWith('spielen')
  })
})

describe('Finden', () => {
  it('findet eine Karte beim Namen — einmal, auch wenn sie mehrfach trifft', async () => {
    belege()
    const { wege, schliessen } = zeige()

    await userEvent.type(screen.getByRole('textbox'), 'asch')

    const finden = await screen.findByRole('list', { name: texte.finden })
    const karte = await within(finden).findByRole('button', {
      name: /Aschenwächter/,
    })
    expect(
      within(finden).getAllByRole('button', { name: /Aschenwächter/ }),
    ).toHaveLength(1)
    expect(karte).toHaveTextContent('Karte · Werkstatt der Dämmerung')

    await userEvent.click(karte)

    expect(schliessen).toHaveBeenCalled()
    expect(wege.oeffneKarte).toHaveBeenCalledWith('7', '12')
  })

  it('findet ein Set, ein Deck und ein Gegenüber beim Namen', async () => {
    belege()
    zeige()

    await userEvent.type(screen.getByRole('textbox'), 'a')

    const finden = await screen.findByRole('list', { name: texte.finden })
    await within(finden).findByRole('button', { name: /Aschenrat/ })

    expect(
      within(finden).getByRole('button', { name: /Werkstatt/ }),
    ).toBeInTheDocument()
  })
})

describe('die Pfeiltasten', () => {
  it('wandern vom Feld in die Zeilen und zurück', async () => {
    belege()
    zeige()

    const feld = screen.getByRole('textbox')
    await screen.findByRole('button', { name: /Karten tauschen/ })

    feld.focus()
    await userEvent.keyboard('{ArrowDown}')
    expect(document.activeElement).toHaveTextContent(texte.karteErstellen)

    await userEvent.keyboard('{ArrowDown}')
    expect(document.activeElement).toHaveTextContent('Schnell spielen mit')

    await userEvent.keyboard('{ArrowUp}{ArrowUp}')
    expect(document.activeElement).toBe(feld)

    // **Über das Feld hinaus geht es nicht** — oben ist oben.
    await userEvent.keyboard('{ArrowUp}')
    expect(document.activeElement).toBe(feld)
  })
})

describe('ohne Wege', () => {
  // **Wie in den Läufen, die nur Seiten und Befehle prüfen**: dann fragt die
  // Palette den Bestand gar nicht.
  it('fragt den Bestand nicht und zeigt weder Tun noch Finden', async () => {
    const { Sets } = belege()
    zeige(false)

    await userEvent.type(screen.getByRole('textbox'), 'sets')

    expect(
      await screen.findByRole('list', { name: texte.seiten }),
    ).toBeInTheDocument()
    expect(Sets).not.toHaveBeenCalled()
    expect(screen.queryByRole('list', { name: texte.tun })).toBeNull()
  })
})
