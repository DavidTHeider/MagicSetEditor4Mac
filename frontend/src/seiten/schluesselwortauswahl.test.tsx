// **Welches Schlüsselwort gewählt ist** — UXR-F3
// (#2009).
//
// # Was MSE2 hier tut
//
// Die Liste hält eine Auswahl, und der Editor daneben zeigt genau dieses
// Wort (`KeywordsPanel::onKeywordSelect`, `src/gui/set/keywords_panel.cpp`,
// Zeilen 356–380, Fassung f9d9356 vom 2021-07-06):
//
// - **Beim Öffnen ist nichts gewählt**, und der Editor ist leer — die
//   Auswahl entsteht erst mit dem ersten Klick (`ItemList::onChangeSet`
//   wählt nichts).
// - **Hinzufügen wählt das neue Wort** (`KeywordList::onAction`, Zeilen
//   63–66 in `src/gui/control/keyword_list.cpp`).
// - **Entfernen wählt das Wort an derselben Stelle** — das folgende, beim
//   letzten das davor, beim einzigen keines (ebd., Zeilen 67–73, und
//   `ItemList::selectItemPos`, `src/gui/control/item_list.cpp`, Zeilen
//   87–100).
//
// # Was hier anders ist — `adapted`
//
// Die Wörter des Spiels stehen nicht in der Liste (sie bleiben die benannte
// Liste darunter), also kennt das Entfernen des letzten eigenen Wortes kein
// „folgendes aus dem Spiel": dann ist nichts gewählt. Und das neue Wort
// bekommt den Fokus **im Namen**, weil der das Erste ist, was man ausfüllt.
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { raeumeApi } from '../test/api'
import {
  editor,
  fuenf,
  liste,
  texte,
  waehle,
  wort,
  zeige,
  zeile,
} from '../test/schluesselwoerter'

afterEach(() => {
  raeumeApi()
})

const felder = () =>
  document.querySelectorAll('input:not([type=hidden]), select, textarea')

describe('Die Wahl eines eigenen Schlüsselworts', () => {
  it('lässt beim Öffnen nichts gewählt und sagt, was zu tun ist', async () => {
    zeige(fuenf())

    const gitter = await liste()

    expect(
      within(gitter).queryAllByRole('row', { selected: true }),
    ).toHaveLength(0)
    expect(screen.getByText(texte.waehleEines)).toBeInTheDocument()
    expect(screen.queryByRole('textbox')).toBeNull()
  })

  it('zeigt den Editor des gewählten Worts und nur den', async () => {
    zeige(fuenf())

    await waehle('Kicker')

    const bearbeiten = await editor('Kicker')
    expect(
      within(bearbeiten).getByRole('textbox', { name: texte.wort }),
    ).toHaveValue('Kicker')
    expect(
      within(bearbeiten).getByRole('textbox', { name: texte.erinnerung }),
    ).toHaveValue('Erinnerung an Kicker')
    expect(screen.queryByText(texte.waehleEines)).toBeNull()
    expect(screen.getAllByRole('textbox', { name: texte.wort })).toHaveLength(1)
    expect((await zeile('Kicker')).ariaSelected).toBe('true')
  })

  it('wechselt den Editor mit der Wahl', async () => {
    zeige(fuenf())

    await waehle('Kicker')
    await editor('Kicker')
    await waehle('Sog')

    expect(
      within(await editor('Sog')).getByRole('textbox', { name: texte.wort }),
    ).toHaveValue('Sog')
    expect(screen.queryByRole('group', { name: /Kicker/ })).toBeNull()
  })

  // **R3 aus `docs/UIREVIEW.md`: höchstens zwölf Felder** — auch mit fünf
  // Wörtern und beiden Parameterlisten. Vorher waren es fünf je Wort.
  it('zeichnet mit fünf Wörtern höchstens zwölf Felder', async () => {
    zeige(fuenf(), {
      belegen: {
        Parameterhilfen: () =>
          Promise.resolve({
            arten: [{ name: 'zahl', beschreibung: '', einfuegen: '<z>' }],
            verweise: [
              {
                nummer: 1,
                art: 'zahl',
                name: '',
                beschreibung: '',
                einfuegen: '{p}',
              },
            ],
            befunde: [],
            fehler: null,
          }),
      },
    })

    await waehle('Kicker')
    await screen.findByRole('combobox', { name: texte.parameterVerweisen })

    expect(felder().length).toBe(7)
    expect(felder().length).toBeLessThanOrEqual(12)
  })
})

describe('Ein Schlüsselwort anlegen', () => {
  it('wählt das neue Wort und setzt den Fokus in seinen Namen', async () => {
    const probe = zeige(fuenf())

    await userEvent.click(
      await screen.findByRole('button', { name: texte.anlegen }),
    )

    const neu = await screen.findByRole('row', { name: texte.ohneName })
    expect(probe.SchluesselwortAnlegen).toHaveBeenCalledWith('3')
    await waitFor(() => {
      expect(neu).toHaveAttribute('aria-selected', 'true')
    })
    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: texte.wort })).toHaveFocus()
    })
    expect(screen.getByRole('textbox', { name: texte.wort })).toHaveValue('')
    expect(within(await liste()).getAllByRole('row')).toHaveLength(7)
  })

  // **Ohne Wort wählt es das erste** — und die Liste steht erst jetzt da.
  it('legt in einem leeren Set das erste an', async () => {
    zeige([])

    expect(await screen.findByText(texte.keine)).toBeInTheDocument()
    await userEvent.click(
      await screen.findByRole('button', { name: texte.anlegen }),
    )

    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: texte.wort })).toHaveFocus()
    })
    expect(screen.queryByText(texte.keine)).toBeNull()
  })

  // **Scheitert es, ist nichts gewählt und die Meldung steht da** (§7,
  // Regel 3: nach einem gescheiterten Schreiben wird nicht nachgeladen).
  it('meldet ein Scheitern und wählt nichts', async () => {
    zeige(fuenf(), {
      belegen: {
        SchluesselwortAnlegen: () =>
          Promise.resolve({
            wortID: '',
            fehler: {
              code: 'konflikt' as const,
              schluessel: 'probe',
              meldung: 'Dieses Set hat sich inzwischen geändert.',
            },
          }),
      },
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.anlegen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Dieses Set hat sich inzwischen geändert.',
    )
    expect(
      within(await liste()).queryAllByRole('row', { selected: true }),
    ).toHaveLength(0)
  })
})

describe('Ein Schlüsselwort löschen', () => {
  async function loesche(name: string) {
    await waehle(name)
    const knopf = await screen.findByRole('button', {
      name: `${texte.loeschen} — ${name}`,
    })
    await waitFor(() => {
      expect(knopf).toBeEnabled()
    })
    await userEvent.click(knopf)
  }

  // **Das folgende, wie MSE** — dasselbe Wort an derselben Stelle.
  it('wählt danach das folgende Wort und nimmt den Fokus mit', async () => {
    const probe = zeige(fuenf())

    await loesche('Rausch')

    await waitFor(() => {
      expect(probe.SchluesselwortLoeschen).toHaveBeenCalledWith('3', '3')
    })
    await waitFor(() => {
      expect(screen.queryByRole('row', { name: 'Rausch' })).toBeNull()
    })
    expect((await zeile('Sog')).ariaSelected).toBe('true')
    await waitFor(() => {
      expect(screen.getByRole('row', { name: 'Sog' })).toHaveFocus()
    })
    await editor('Sog')
  })

  it('wählt beim letzten Wort das davor', async () => {
    zeige(fuenf())

    await loesche('Wacht')

    await waitFor(() => {
      expect(screen.queryByRole('row', { name: 'Wacht' })).toBeNull()
    })
    expect((await zeile('Sog')).ariaSelected).toBe('true')
    await waitFor(() => {
      expect(screen.getByRole('row', { name: 'Sog' })).toHaveFocus()
    })
  })

  // **Beim einzigen bleibt nichts gewählt**, und der Fokus geht dorthin, wo
  // man weitermacht: zum Anlegen. Ohne ihn stünde er am `body`, denn der
  // Knopf, der ihn hatte, ist weg.
  it('lässt beim einzigen Wort nichts gewählt und führt zum Anlegen', async () => {
    zeige([wort({ id: '1', wort: 'Beben' })])

    await loesche('Beben')

    expect(await screen.findByText(texte.keine)).toBeInTheDocument()
    expect(screen.queryByRole('group')).toBeNull()
    await waitFor(() => {
      expect(screen.getByRole('button', { name: texte.anlegen })).toHaveFocus()
    })
  })

  it('meldet ein Scheitern und löscht nichts', async () => {
    zeige(fuenf(), {
      belegen: {
        SchluesselwortLoeschen: () =>
          Promise.resolve({
            fehler: {
              code: 'konflikt' as const,
              schluessel: 'probe',
              meldung: 'Dieses Set hat sich inzwischen geändert.',
            },
          }),
      },
    })

    await loesche('Rausch')

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Dieses Set hat sich inzwischen geändert.',
    )
    expect((await zeile('Rausch')).ariaSelected).toBe('true')
    expect(within(await liste()).getAllByRole('row')).toHaveLength(6)
  })
})
