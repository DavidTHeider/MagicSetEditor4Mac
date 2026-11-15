// **Die Schlüsselwortseite**
// (#1062).
//
// # Wofür
//
// `magic.mse-game` führt **334** Schlüsselwörter; ein Set bringt selten mehr
// als eine Handvoll eigene mit. Bis hierher war keines davon zu sehen: die
// `keyword:`-Blöcke eines Sets überlebten den Rundlauf als Rohblock, und die
// des Spiels standen nur auf der Karte.
//
// # Was hier geprüft wird
//
// Der Weg **durch** die Naht: was die Bindung schickt, wird zu Feldern; die
// des Spiels sind abgeschaltet, die eigenen nicht; und ein Klick geht mit
// Kennung und Feldnamen zurück.
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { raeumeApi } from '../test/api'
import { editor, texte, waehle, wort, zeige } from '../test/schluesselwoerter'

afterEach(() => {
  raeumeApi()
})

describe('Die Schlüsselwortseite', () => {
  it('zeigt die eigenen zum Bearbeiten', async () => {
    zeige([wort({ id: '7', wort: 'Beben', muster: 'Beben', modus: 'custom' })])

    // **Erst gewählt, dann bearbeitet** (UXR-F3, #2009): ein Editor für das
    // gewählte Wort statt einer für jedes.
    await waehle('Beben')

    const name = await screen.findByLabelText(texte.wort)
    expect(name).toHaveValue('Beben')
    expect(name).toBeEnabled()

    expect(screen.getByLabelText(texte.modus)).toHaveValue('custom')
  })

  // **Die des Spiels sind unveränderlich, und die Oberfläche sagt es** —
  // MSE zeigt sie grau und lässt sie nicht ändern (`KeywordList::OnGetItemAttr`
  // und `KeywordsPanel::onKeywordSelect`, `src/gui/control/keyword_list.cpp`
  // Zeilen 225–235 und `src/gui/set/keywords_panel.cpp` Zeilen 355–370,
  // Fassung f9d9356 vom 2021-07-06).
  //
  // **Als Text und nicht als abgeschaltetes Feld** — seit UXR-07 (#1915).
  // Hier stand „abgeschaltet und nicht nur grau: `disabled` ist eine
  // Auskunft“; dieselbe Auskunft gibt jetzt ein Satz über der Liste, und
  // die 368 Wörter von `magic.mse-game` zeichnen keine 1.840 Felder mehr
  // (`schluesselwoerterfeste.test.tsx`).
  it('zeigt die des Spiels erst auf Wunsch, und dann als Text', async () => {
    zeige([wort({ wort: 'Fliegend', muster: 'Fliegend', fest: true })])

    expect(screen.queryByText('Fliegend')).not.toBeInTheDocument()

    await userEvent.click(
      await screen.findByRole('button', { name: '1 aus dem Spiel' }),
    )

    expect(await screen.findByText(texte.ausDemSpiel)).toBeInTheDocument()
    expect(screen.getByRole('listitem')).toHaveTextContent('Fliegend')
    expect(screen.queryByLabelText(texte.wort)).not.toBeInTheDocument()
  })

  it('sagt es, wenn das Set keine eigenen mitbringt', async () => {
    zeige([wort({ wort: 'Fliegend', fest: true })])

    expect(await screen.findByText(texte.keine)).toBeInTheDocument()
  })

  it('schreibt ein Feld beim Verlassen', async () => {
    const { SchluesselwortSetzen } = zeige([
      wort({ id: '7', wort: 'Beben', muster: 'alt' }),
    ])

    await waehle('Beben')
    const feld = await screen.findByLabelText(texte.muster)
    await userEvent.clear(feld)
    await userEvent.type(feld, 'neu')
    await userEvent.tab()

    await waitFor(() => {
      expect(SchluesselwortSetzen).toHaveBeenCalledWith(
        '3',
        '7',
        'muster',
        'neu',
        expect.any(Array),
      )
    })
  })

  // **Ein unveränderter Wert wird nicht geschrieben.** Wer ein Feld anklickt
  // und wieder verlässt, legte sonst einen Schreibvorgang an, der nichts
  // ändert — dieselbe Regel wie an einem Kartenfeld.
  it('schreibt einen unveränderten Wert nicht', async () => {
    const { SchluesselwortSetzen } = zeige([
      wort({ id: '7', wort: 'Beben', muster: 'alt' }),
    ])

    await waehle('Beben')
    await userEvent.click(await screen.findByLabelText(texte.muster))
    await userEvent.tab()

    expect(SchluesselwortSetzen).not.toHaveBeenCalled()
  })

  it('legt eines an und löscht eines', async () => {
    const { SchluesselwortAnlegen, SchluesselwortLoeschen } = zeige([
      wort({ id: '7', wort: 'Beben' }),
    ])

    await userEvent.click(
      await screen.findByRole('button', { name: texte.anlegen }),
    )

    await waitFor(() => {
      expect(SchluesselwortAnlegen).toHaveBeenCalledWith('3')
    })

    // **Das neue ist gewählt; gelöscht wird das andere** — beide stehen in
    // der Liste, und gelöscht wird, was man wählt.
    await waehle('Beben')

    // **Auf den bedienbaren Knopf gewartet, nicht auf den Ruf** (#1809):
    // nach dem Anlegen lädt die Liste nach, und solange der Vorgang läuft,
    // ruhen die Knöpfe.
    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: `${texte.loeschen} — Beben` }),
      ).toBeEnabled()
    })

    await userEvent.click(
      screen.getByRole('button', { name: `${texte.loeschen} — Beben` }),
    )

    await waitFor(() => {
      expect(SchluesselwortLoeschen).toHaveBeenCalledWith('3', '7')
    })
  })

  // **Jedes der vier Textfelder schreibt sein eigenes Feld.** Ein Test, der
  // nur eines fährt, ließe die anderen drei ungeprüft — und ein vertauschter
  // Feldname fiele erst dem auf, der einen Erinnerungstext im Namensfeld
  // wiederfindet.
  it('schreibt jedes Feld unter seinem Namen', async () => {
    const { SchluesselwortSetzen } = zeige([wort({ id: '7', wort: 'Beben' })])

    await waehle('Beben')

    const felder: [string, string][] = [
      [texte.wort, 'wort'],
      [texte.muster, 'muster'],
      [texte.erinnerung, 'erinnerung'],
      [texte.regeln, 'regeln'],
    ]

    for (const [aufschrift, feld] of felder) {
      const eingabe = await screen.findByLabelText(aufschrift)
      await userEvent.clear(eingabe)
      await userEvent.type(eingabe, 'x')
      await userEvent.tab()

      await waitFor(() => {
        expect(SchluesselwortSetzen).toHaveBeenCalledWith(
          '3',
          '7',
          feld,
          'x',
          expect.any(Array),
        )
      })
    }
  })

  it('setzt den Modus über die Auswahl', async () => {
    const { SchluesselwortSetzen } = zeige([
      wort({ id: '7', wort: 'Beben', modus: 'core' }),
    ])

    await waehle('Beben')
    await userEvent.selectOptions(
      await screen.findByLabelText(texte.modus),
      'custom',
    )

    await waitFor(() => {
      expect(SchluesselwortSetzen).toHaveBeenCalledWith(
        '3',
        '7',
        'modus',
        'custom',
        null,
      )
    })
  })

  // **Ohne Modi im Spiel gibt es keine Auswahl** — erfunden wird keine.
  it('lässt die Auswahl aus, wo das Spiel keine Modi kennt', async () => {
    zeige([wort({ id: '7', wort: 'Beben' })], { modi: [] })

    await waehle('Beben')
    await screen.findByLabelText(texte.wort)

    expect(screen.queryByLabelText(texte.modus)).not.toBeInTheDocument()
  })

  it('sagt es, wenn ein Muster nichts erkennt', async () => {
    zeige([wort({ id: '7', wort: 'Kaputt', gueltig: false })])

    // **In der Liste und im Editor** — wer ein Wort ohne Wirkung sucht, sieht
    // es schon in der Zeile (MSE zeigt sie rot), und am Muster steht, warum.
    const zeile = await screen.findByRole('row', { name: 'Kaputt' })
    expect(within(zeile).getByText(texte.ungueltig)).toBeInTheDocument()

    await waehle('Kaputt')
    expect(
      within(await editor('Kaputt')).getByText(texte.ungueltig),
    ).toBeInTheDocument()
  })

  it('meldet einen Fehler beim Schreiben', async () => {
    zeige([wort({ id: '7', wort: 'Beben', muster: 'alt' })], {
      modi: [],
      belegen: {
        SchluesselwortSetzen: () =>
          Promise.resolve({
            fehler: {
              code: 'konflikt' as const,
              schluessel: 'probe',
              meldung: 'Dieses Set hat sich inzwischen geändert.',
            },
          }),
      },
    })

    await waehle('Beben')
    const feld = await screen.findByLabelText(texte.muster)
    await userEvent.clear(feld)
    await userEvent.type(feld, 'neu')
    await userEvent.tab()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Dieses Set hat sich inzwischen geändert.',
    )
  })
})

// **Ein Muster, das nichts erkennt, sagt es am Muster** — Werkstück U3 von
// Phase 5.1 (#1414).
//
// Die Regel steht in `service.musterGreift`, portiert aus `Keyword::prepare`
// (`valid = !match_re.matches(_(""))`, `src/data/keyword.cpp`, Zeilen
// 205–209, Fassung f9d9356 vom 2021-07-06). Bis heute stand der Satz als
// namenloser `<span>` unter der Zeile — §6 verlangt ihn **am verursachenden
// Feld**, und das ist das Muster und nicht der Name.
it('hängt den Befund an das Musterfeld', async () => {
  zeige([wort({ id: '7', wort: 'Kaputt', gueltig: false })])

  await waehle('Kaputt')
  const feld = await screen.findByLabelText(texte.muster)

  expect(feld).toHaveAttribute('aria-invalid', 'true')
  expect(feld).toHaveAccessibleDescription(texte.ungueltig)

  // **Und der Name ist nicht schuld.** Ein Befund an jedem Feld der Zeile
  // wäre dieselbe Auskunft wie gar keiner.
  expect(screen.getByLabelText(texte.wort)).toHaveAttribute(
    'aria-invalid',
    'false',
  )
})

it('lässt ein greifendes Muster ohne Befund', async () => {
  zeige([wort({ id: '7', wort: 'Beben' })])

  await waehle('Beben')
  const feld = await screen.findByLabelText(texte.muster)

  expect(feld).toHaveAttribute('aria-invalid', 'false')
  expect(screen.queryByText(texte.ungueltig)).toBeNull()
})
