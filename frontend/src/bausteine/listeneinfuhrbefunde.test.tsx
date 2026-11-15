// **Was der Zweitblick über Phase 4.5 an der Listeneinfuhr fand** —
// Befunde 2 und 4
// (#1292).
//
// **Eine eigene Datei**, weil `Listeneinfuhr.test.tsx` an der Zeilengrenze aus
// §2 stand — und weil der Schnitt der Frage folgt: dort, *wie der Kasten sich
// verhält*, hier, *was er über Fehlschläge sagt*.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Listeneinfuhr } from './Listeneinfuhr'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { ohneDublette } from '../test/einfuhr'

import type { Listendeutungszeile } from '../antwortenlisteneinfuhr'

afterEach(() => {
  raeumeApi()
})

const texte = {
  feld: 'Kartenliste',
  deuten: 'Zuordnen',
  einlesen: 'Aufnehmen',
}

// eindeutige gibt eine Zeile mit genau einem Druck.
const eindeutige: Listendeutungszeile = {
  roh: '4 Lightning Bolt (M11) 149',
  zeilennummer: 1,
  menge: 4,
  name: 'Lightning Bolt',
  setcode: 'm11',
  nummer: '149',
  finish: 'nonfoil',
  treffer: [
    {
      printing: 'druck-1',
      name: 'Lightning Bolt',
      setcode: 'm11',
      setname: 'Magic 2011',
      nummer: '149',
      seltenheit: 'common',
      bild: '',
    },
  ],
  drucke: 1,
  eindeutig: true,
  befund: '',
  weitere: false,
  fehler: null,
}

// aufgenommen ist der Rückruf, den der Kasten nach dem Schreiben ruft.
function aufgenommen(): void {
  // absichtlich leer
}

// zeige stellt den Kasten mit einer belegten Fassade auf.
function zeige(zeilen: Listendeutungszeile[], einlesen = vi.fn()) {
  belegeApi({
    ListeVorschauen: ohneDublette,
    ListeDeuten: () =>
      Promise.resolve({
        deutung: { zeilen, befunde: [], alterSekunden: 0 },
        fehler: null,
      }),
    ListeEinlesen: einlesen,
  })

  render(
    <Sprachraum sprache="de">
      <Listeneinfuhr aufAufgenommen={aufgenommen} />
    </Sprachraum>,
  )

  return einlesen
}

// tippeUndDeute füllt das Feld und drückt den Knopf.
async function tippeUndDeute(text: string) {
  await userEvent.type(screen.getByLabelText(texte.feld), text)
  await userEvent.click(screen.getByRole('button', { name: texte.deuten }))
}

// **Was beim Aufnehmen nicht durchkam, steht da** — Befund 4.
it('zeigt die Zeilen, die nicht aufgenommen wurden', async () => {
  zeige(
    [eindeutige],
    vi.fn(() =>
      Promise.resolve({
        bericht: {
          angelegt: [1],
          erhoeht: 0,
          stueck: 4,
          befunde: ['Zeile 2: die gibt es schon'],
        },
        fehler: null,
      }),
    ),
  )

  await tippeUndDeute('4 Lightning Bolt (M11) 149')
  await screen.findByText('Lightning Bolt · Magic 2011 #149')
  await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))

  expect(
    await screen.findByText('Zeile 2: die gibt es schon'),
  ).toBeInTheDocument()
})

// **Ein bearbeiteter Text verwirft die alte Deutung** — Befund 2.
//
// Ohne das könnte jemand Liste A deuten, den Text durch Liste B ersetzen und
// dann »Aufnehmen« drücken: geschrieben würde A, obwohl sichtbar B im Feld
// steht.
it('verwirft die Deutung, wenn der Text sich ändert', async () => {
  const einlesen = vi.fn()

  zeige([eindeutige], einlesen)

  await tippeUndDeute('4 Lightning Bolt (M11) 149')
  await screen.findByText('Lightning Bolt · Magic 2011 #149')

  await userEvent.type(screen.getByLabelText(texte.feld), '\n2 Counterspell')

  // **Der Knopf ist weg** — es gibt nichts mehr, das entschieden wäre.
  expect(
    screen.queryByRole('button', { name: texte.einlesen }),
  ).not.toBeInTheDocument()

  expect(einlesen).not.toHaveBeenCalled()
})
