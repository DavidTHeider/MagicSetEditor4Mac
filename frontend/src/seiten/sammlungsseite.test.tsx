// **Die Sammlungsseite als Sicht** —
// #1200.
//
// Was dastand, lud die ganze Liste und bildete sie unverändert ab: kein
// Suchen, kein Filtern, kein Sortieren — und der Kartenbezug war eine UUID.
import { keineHerkunft } from '../apiherkunft'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import type { ExemplarDTO } from '../api'
import { Sammlung } from './Sammlung'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(() => {
  raeumeApi()
})

const texte = {
  suche: 'Suchen',
  sortierung: 'Sortieren nach',
  ausfuehrung: 'Ausführung',
  bearbeiten: 'Bearbeiten',
}

// exemplar gibt ein Exemplar mit den genannten Abweichungen.
function exemplar(teile: Partial<ExemplarDTO> = {}): ExemplarDTO {
  return {
    id: 1,
    fassung: 1,
    printing: '8a1b0d5e',
    druckID: '',
    vorlaeufig: false,
    menge: 1,
    zustand: 'nearMint',
    sprache: 'en',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name: '',
    setcode: '',
    nummer: '',
    lagerort: '',
    wunsch: false,
    bezeichnung: '8a1b0d5e',
    herkunft: keineHerkunft,
    ...teile,
  }
}

// zwei gibt eine benannte und eine unbenannte Zeile.
const bolt = exemplar({
  id: 1,
  printing: '8a1b0d5e',
  name: 'Lightning Bolt',
  setcode: 'm11',
  nummer: '149',
  lagerort: '',
  wunsch: false,
  bezeichnung: 'Lightning Bolt · M11 #149',
  herkunft: keineHerkunft,
  finish: 'foil',
})

const ohneNamen = exemplar({ id: 2, printing: 'zzz-9', bezeichnung: 'zzz-9' })

// zeige belegt die Fassade mit einer **echten** Sicht in Go.
//
// **Seit Werkstück #1500
// filtert und ordnet die Go-Seite.** Der Beleg bildet das nach, so weit diese
// Fälle es brauchen — und prüft damit zugleich, dass die Sicht wirklich
// hinausgeht. Was die Sicht **tut**, steht in `internal/domain`.
function fenster(exemplare: ExemplarDTO[]) {
  return (suche: string, ausfuehrung: string, sortierung: string) => {
    let gesehen = exemplare.filter(
      (eines) =>
        suche === '' ||
        eines.bezeichnung.toLowerCase().includes(suche.toLowerCase()),
    )

    if (ausfuehrung !== 'alle') {
      gesehen = gesehen.filter((eines) => eines.finish === ausfuehrung)
    }

    if (sortierung === 'menge') {
      gesehen = [...gesehen].sort((links, rechts) => rechts.menge - links.menge)
    }

    return Promise.resolve({
      exemplare: gesehen,
      gesamt: gesehen.length,
      insgesamt: exemplare.length,
      stueck: exemplare.length,
      fehler: null,
    })
  }
}

function zeige(exemplare: ExemplarDTO[], zusatz = {}) {
  belegeApi({
    Sammlung: () => Promise.resolve({ exemplare, fehler: null }),
    Sammlungsfenster: fenster(exemplare),
    Portfolio: () =>
      Promise.resolve({ zeilen: [], ohnePreis: 0, fehler: null }),
    Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
    ...zusatz,
  })

  render(
    <Sprachraum sprache="de">
      <Sammlung bestandGeaendert={vi.fn()} />
    </Sprachraum>,
  )
}

// **Der lesbare Bezug steht vorn, die Kennung daneben.**
it('zeigt den lesbaren Kartenbezug statt der Kennung', async () => {
  zeige([bolt])

  // In der Zeile steht der Bezug; die Kennung erst im Inspector (U8).
  expect(
    await screen.findByText('Lightning Bolt · M11 #149'),
  ).toBeInTheDocument()
  expect(screen.queryByText('8a1b0d5e')).not.toBeInTheDocument()

  await userEvent.click(screen.getByText('Lightning Bolt · M11 #149'))
  const inspector = within(
    screen.getByRole('complementary', {
      name: kataloge.de.seiten.sammlung.inspector,
    }),
  )
  expect(inspector.getByText('8a1b0d5e')).toHaveClass('font-mono')
})

// **Ohne Namen steht die Kennung allein** — und nicht zweimal.
it('zeigt ohne Namen nur die Kennung', async () => {
  zeige([ohneNamen])

  await waitFor(() => {
    expect(screen.getAllByText('zzz-9')).toHaveLength(1)
  })
})

// **Die Suche nimmt weg, was nicht passt.**
it('sucht über den Bezug', async () => {
  zeige([bolt, ohneNamen])

  await screen.findByText('Lightning Bolt · M11 #149')

  await userEvent.type(screen.getByLabelText(texte.suche), 'bolt')

  await waitFor(() => {
    expect(screen.queryByText('zzz-9')).not.toBeInTheDocument()
  })

  expect(screen.getByText('Lightning Bolt · M11 #149')).toBeInTheDocument()
})

// **Und der Filter über die Ausführung ebenso.**
it('filtert nach Ausführung', async () => {
  zeige([bolt, ohneNamen])

  await screen.findByText('Lightning Bolt · M11 #149')

  await userEvent.selectOptions(
    screen.getByLabelText(texte.ausfuehrung),
    'nonfoil',
  )

  await waitFor(() => {
    expect(
      screen.queryByText('Lightning Bolt · M11 #149'),
    ).not.toBeInTheDocument()
  })
})

// **Nimmt die Sicht alles weg, sagt sie es** (§6): ein leerer Zustand erklärt
// den nächsten Schritt, statt „Keine Daten" zu zeigen.
it('erklärt eine Sicht ohne Treffer', async () => {
  zeige([bolt])

  await screen.findByText('Lightning Bolt · M11 #149')

  await userEvent.type(screen.getByLabelText(texte.suche), 'gibtesnicht')

  expect(await screen.findByText(/Kein Exemplar passt/)).toBeInTheDocument()

  // **Und der nächste Schritt ist ein Knopf** (UX-08.AK-02, Artboard
  // *BelegZustaende*: „Gefiltert leer"): er nimmt Suche und Filter zurück,
  // und das Exemplar steht wieder da.
  await userEvent.click(
    screen.getByRole('button', {
      name: kataloge.de.seiten.sammlung.ohneTrefferZuruecksetzen,
    }),
  )

  expect(
    await screen.findByText('Lightning Bolt · M11 #149'),
  ).toBeInTheDocument()
  expect(screen.getByLabelText(texte.suche)).toHaveValue('')
})

// **Eine Sicht, die etwas wegnimmt, nimmt es in einem Schritt zurück — auch
// mit Treffern** (DOM-03.AK-03, #1698): derselbe Knopf wie im leeren
// Zustand, damit niemand erst alles wegfiltern muss, um alles zu sehen.
it('nimmt eine einschränkende Sicht auch mit Treffern zurück', async () => {
  zeige([bolt, ohneNamen])

  await screen.findByText('Lightning Bolt · M11 #149')
  const zuruecksetzen = {
    name: kataloge.de.seiten.sammlung.ohneTrefferZuruecksetzen,
  }
  expect(screen.queryByRole('button', zuruecksetzen)).toBeNull()

  await userEvent.type(screen.getByLabelText(texte.suche), 'bolt')
  await waitFor(() => {
    expect(screen.queryByText('zzz-9')).not.toBeInTheDocument()
  })

  await userEvent.click(screen.getByRole('button', zuruecksetzen))

  expect(await screen.findByText('zzz-9')).toBeInTheDocument()
  expect(screen.getByLabelText(texte.suche)).toHaveValue('')
  expect(screen.queryByRole('button', zuruecksetzen)).toBeNull()
})

// **Die Ordnung lässt sich umstellen.**
it('sortiert nach Menge', async () => {
  zeige([
    exemplar({ id: 1, bezeichnung: 'Aaa', menge: 1 }),
    exemplar({ id: 2, bezeichnung: 'Bbb', menge: 9 }),
  ])

  await screen.findByText('Aaa')

  await userEvent.selectOptions(
    screen.getByLabelText(texte.sortierung),
    'menge',
  )

  // **Die erste Zeile trägt danach den größeren Stapel.** Geprüft am Text der
  // Zeile und nicht an einer Indexbehauptung: `toHaveTextContent` sagt
  // dasselbe und braucht keine Zusicherung über `undefined`.
  await waitFor(() => {
    // Zeile 1 ist die Kopfzeile; die erste Datenzeile trägt Bbb.
    expect(screen.getAllByRole('row')[1]).toHaveTextContent('Bbb')
  })
})

// **Druck, Sprache und Ausführung sind änderbar** — sie bestimmen die
// Identität mit, und eine falsch erfasste Sprache ließ sich nur durch Löschen
// und Neuanlegen berichtigen.
it('lässt Druck, Sprache und Ausführung ändern', async () => {
  const aendern = vi.fn().mockResolvedValue(null)
  zeige([bolt], { ExemplarAendern: aendern })

  // **Die Zeile ist die Wahl, der Inspector die Bearbeitung** (U8, #1419).
  await userEvent.click(await screen.findByText('Lightning Bolt · M11 #149'))
  const inspector = within(
    screen.getByRole('complementary', {
      name: kataloge.de.seiten.sammlung.inspector,
    }),
  )
  await userEvent.click(
    inspector.getByRole('button', { name: texte.bearbeiten }),
  )
  await userEvent.clear(inspector.getByLabelText('Sprache'))
  await userEvent.type(inspector.getByLabelText('Sprache'), 'de')
  await userEvent.click(
    inspector.getByRole('button', {
      name: kataloge.de.seiten.sammlung.aenderungSichern,
    }),
  )
  await waitFor(() => {
    expect(aendern).toHaveBeenCalledWith(
      expect.objectContaining({ sprache: 'de', printing: '8a1b0d5e' }),
    )
  })
})

// **Ein Konflikt bietet den heutigen Stand an** — UX-08.AK-02
// (#1680):
// jemand hat das Exemplar inzwischen geändert, und „Neu laden" holt es.
it('bietet nach einem Konflikt „Neu laden" an', async () => {
  const fensterweg = vi.fn(fenster([bolt]))
  zeige([bolt], {
    Sammlungsfenster: fensterweg,
    ExemplarAendern: () =>
      Promise.resolve({
        code: 'konflikt',
        schluessel: 'konflikt',
        meldung: 'Der Eintrag wurde zwischenzeitlich geändert.',
      }),
  })

  await userEvent.click(await screen.findByText('Lightning Bolt · M11 #149'))
  const inspector = within(
    screen.getByRole('complementary', {
      name: kataloge.de.seiten.sammlung.inspector,
    }),
  )
  await userEvent.click(
    inspector.getByRole('button', { name: texte.bearbeiten }),
  )
  await userEvent.click(
    inspector.getByRole('button', {
      name: kataloge.de.seiten.sammlung.aenderungSichern,
    }),
  )

  const neuLaden = await screen.findByRole('button', {
    name: kataloge.de.bausteine.zustaende.neuLaden,
  })
  const vorher = fensterweg.mock.calls.length

  await userEvent.click(neuLaden)

  await waitFor(() => {
    expect(fensterweg.mock.calls.length).toBeGreaterThan(vorher)
  })
  expect(
    screen.queryByText('Der Eintrag wurde zwischenzeitlich geändert.'),
  ).toBeNull()
})
