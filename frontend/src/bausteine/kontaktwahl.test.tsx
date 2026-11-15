// **Wen es betrifft — gewählt, ab fünfzehn Gegenübern mit Suche** — UXR-F1
// (#2007),
// R7 aus `docs/UIREVIEW.md`: *„Auswahl ohne Suche ab 15"*.
//
// Bis vierzehn bestätigte Gegenüber stehen als Optionsfelder da: man sieht
// alle, und der Fingerabdruck steht dabei. Ab fünfzehn wären das fünfzehn
// Zeilen und mehr, die niemand überblickt — dann ist es eine Auswahl, in die
// man den Namen tippt.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { gegenueber } from '../test/gegenueberbestand'
import { Kontaktwahl } from './Kontaktwahl'

import type { Peerkontakt } from '../apipeerkontakt'

const eigene = kataloge.de.bausteine

afterEach(() => {
  raeumeApi()
})

// namen sind `anzahl` Vornamen mit Nummer — „Anna01", „Anna02" …
const bestaetigte = (anzahl: number, anders: Partial<Peerkontakt> = {}) =>
  Array.from({ length: anzahl }, (_, stelle) =>
    gegenueber(`Anna${String(stelle + 1).padStart(2, '0')}`, {
      stand: 'bestaetigt',
      ...anders,
    }),
  )

function zeige(
  kontakte: Peerkontakt[],
  teile: { gewaehlt?: string; recht?: string } = {},
) {
  const waehlen = vi.fn()

  belegeApi({
    Peerkontakte: () => Promise.resolve({ kontakte, zwecke: [] }),
  })

  render(
    <Sprachraum sprache="de">
      <Kontaktwahl
        titel="Wer leiht?"
        leer="Niemand da."
        gewaehlt={teile.gewaehlt ?? ''}
        waehlen={waehlen}
        {...(teile.recht === undefined ? {} : { recht: teile.recht })}
      />
    </Sprachraum>,
  )

  return waehlen
}

describe('Die Kontaktwahl', () => {
  it('zeigt bis vierzehn Gegenüber als Optionsfelder', async () => {
    const waehlen = zeige(bestaetigte(14))

    expect(await screen.findAllByRole('radio')).toHaveLength(14)
    expect(screen.queryByRole('combobox')).toBeNull()

    await userEvent.click(screen.getByRole('radio', { name: /^Anna03 · / }))

    expect(waehlen).toHaveBeenCalledExactlyOnceWith('banna03')
  })

  it('ist ab fünfzehn Gegenübern eine Auswahl mit Suche', async () => {
    const waehlen = zeige(bestaetigte(15))

    const wahl = await screen.findByRole('combobox', { name: 'Wer leiht?' })

    expect(screen.queryAllByRole('radio')).toHaveLength(0)
    expect(wahl).toHaveValue('')
    expect(wahl).toHaveAttribute('placeholder', eigene.kontaktwahlSuchen)

    await userEvent.click(wahl)
    await userEvent.keyboard('anna12{Enter}')

    expect(waehlen).toHaveBeenCalledExactlyOnceWith('banna12')
  })

  it('zeigt den Gewählten mit Namen und Fingerabdruck', async () => {
    zeige(bestaetigte(15), { gewaehlt: 'banna07' })

    expect(
      await screen.findByRole('combobox', { name: 'Wer leiht?' }),
    ).toHaveValue(
      eigene.kontaktwahlEintrag
        .replace('{name}', 'Anna07')
        .replace('{fingerabdruck}', '1111-2222-3333-4444-5555-6666-7777-8888'),
    )
  })

  // **Gezählt wird, wer dasteht**: wo ein Zweck verlangt ist, nur die, denen
  // er freigegeben ist.
  it('zählt nur die, die den verlangten Zweck dürfen', async () => {
    const kontakte = [
      ...bestaetigte(15, { rechte: [] }),
      ...bestaetigte(2, { rechte: ['partie'], eigenerName: 'Berta' }).map(
        (eines, stelle) => ({ ...eines, kennung: `bb${String(stelle)}` }),
      ),
    ]
    zeige(kontakte, { recht: 'partie' })

    expect(await screen.findAllByRole('radio')).toHaveLength(2)
    expect(screen.queryByRole('combobox')).toBeNull()
  })

  // **Ohne eigenen Namen gilt der, den das Gegenüber sich selbst gibt** — in
  // beiden Gestalten der Wahl.
  it.each([
    ['als Optionsfeld', 1],
    ['in der Auswahl mit Suche', 15],
  ])(
    'nennt ein Gegenüber ohne eigenen Namen bei seinem eigenen, %s',
    async (_, anzahl) => {
      const kontakte = bestaetigte(anzahl)
      kontakte[0] = gegenueber('Zora', {
        stand: 'bestaetigt',
        eigenerName: '',
        selbstgewaehlterName: 'Zoras Mac',
      })
      zeige(kontakte, { gewaehlt: 'bzora' })

      if (anzahl === 1) {
        expect(
          await screen.findByRole('radio', { name: /^Zoras Mac · / }),
        ).toBeChecked()
      } else {
        const feld = await screen.findByRole('combobox', {
          name: 'Wer leiht?',
        })

        expect((feld as HTMLInputElement).value).toMatch(/^Zoras Mac · /)
      }
    },
  )

  it('sagt, wenn niemand dasteht', async () => {
    zeige([])

    expect(await screen.findByText('Niemand da.')).toBeInTheDocument()
  })

  it('wartet und meldet einen Fehlschlag mit dem nächsten Versuch', async () => {
    belegeApi({
      Peerkontakte: () => Promise.reject(new Error('kein Draht')),
    })
    render(
      <Sprachraum sprache="de">
        <Kontaktwahl
          titel="Wer leiht?"
          leer="Niemand da."
          gewaehlt=""
          waehlen={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(
      await screen.findByRole('button', {
        name: kataloge.de.bausteine.zustaende.wiederholen,
      }),
    ).toBeInTheDocument()
  })
})
