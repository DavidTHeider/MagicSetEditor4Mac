// **Eine Kartenliste im Fenster** — Werkstück 3 von Phase 4.5
// (#1292).
//
// # Was hier geprüft wird
//
// **Der zweite Schritt** (§16, Regel 6): dass eine offene Zeile eine Wahl
// verlangt und nicht geraten wird. „Counterspell“ ohne Setangabe passt bei
// Scryfall auf 72 Drucke (nachgelesen am 2026-08-30) — welchen jemand besitzt,
// weiß nur er.
import { render, screen, waitFor } from '@testing-library/react'
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
  wahl: 'Welcher Druck?',
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

// offene gibt eine Zeile mit zwei Drucken.
const offene: Listendeutungszeile = {
  roh: '2 Counterspell',
  zeilennummer: 2,
  menge: 2,
  name: 'Counterspell',
  setcode: '',
  nummer: '',
  finish: 'nonfoil',
  treffer: [
    {
      printing: 'druck-2',
      name: 'Counterspell',
      setcode: 'lea',
      setname: 'Limited Edition Alpha',
      nummer: '54',
      seltenheit: 'common',
      bild: '',
    },
    {
      printing: 'druck-3',
      name: 'Counterspell',
      setcode: 'dsc',
      setname: 'Duskmourn Commander',
      nummer: '114',
      seltenheit: 'common',
      bild: '',
    },
  ],
  drucke: 2,
  eindeutig: false,
  befund: '',
  weitere: false,
  fehler: null,
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

// aufgenommen ist der Rückruf, den der Kasten nach dem Schreiben ruft.
//
// **Eine benannte Funktion und kein `() => undefined` im JSX**: eine
// Void-Ausdruck-Zuweisung in einer Eigenschaft ist eslint zu Recht suspekt.
function aufgenommen(): void {
  // absichtlich leer — was danach nachlädt, prüft `sammlungsseite.test.tsx`
}

// tippeUndDeute füllt das Feld und drückt den Knopf.
async function tippeUndDeute(text: string) {
  await userEvent.type(screen.getByLabelText(texte.feld), text)
  await userEvent.click(screen.getByRole('button', { name: texte.deuten }))
}

// **Eine eindeutige Zeile braucht keine Wahl.**
it('zeigt eine eindeutige Zeile ohne Auswahl', async () => {
  zeige([eindeutige])

  await tippeUndDeute('4 Lightning Bolt (M11) 149')

  // **Der Bezug und nicht der Name**: die rohe Zeile nennt ihn auch, und ein
  // Ausdruck über »Lightning Bolt« träfe beide. Was diese Zeile belegt, ist
  // dass der **zugeordnete Druck** dasteht.
  expect(
    await screen.findByText('Lightning Bolt · Magic 2011 #149'),
  ).toBeInTheDocument()

  expect(
    screen.queryByRole('combobox', { name: texte.wahl }),
  ).not.toBeInTheDocument()
})

// **Eine offene Zeile verlangt eine Wahl — und wird nicht geraten.**
//
// Das ist der Satz, für den es den zweiten Schritt gibt (§16, Regel 6).
it('verlangt bei mehreren Drucken eine Entscheidung', async () => {
  const einlesen = zeige([offene])

  await tippeUndDeute('2 Counterspell')

  const wahl = await screen.findByRole('combobox', { name: texte.wahl })
  expect(wahl).toBeInTheDocument()

  // **Nichts ist vorgewählt.** Ein vorgewählter Druck wäre eine Entscheidung,
  // die niemand getroffen hat.
  expect((wahl as HTMLSelectElement).value).toBe('')

  await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))

  expect(einlesen).not.toHaveBeenCalled()

  // **Und der Beleg, dass jemand zuhört** (§3): dieselbe Taste nach einer Wahl
  // ruft wirklich. Ohne ihn träfe die Verneinung darüber auch dann zu, wenn
  // der Knopf gar nicht hinge.
  await userEvent.selectOptions(wahl, 'druck-3')
  await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))

  await waitFor(() => {
    expect(einlesen).toHaveBeenCalledTimes(1)
  })
})

// **Was gewählt ist, geht mit — samt lesbarem Bezug.**
it('schickt den gewählten Druck samt Namen', async () => {
  const einlesen = vi.fn(() =>
    Promise.resolve({
      bericht: { angelegt: [1], erhoeht: 0, stueck: 2, befunde: [] },
      fehler: null,
    }),
  )

  zeige([offene], einlesen)

  await tippeUndDeute('2 Counterspell')

  await userEvent.selectOptions(
    await screen.findByRole('combobox', { name: texte.wahl }),
    'druck-2',
  )
  await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))

  await waitFor(() => {
    expect(einlesen).toHaveBeenCalledWith(
      [
        {
          printing: 'druck-2',
          menge: 2,
          zustand: 'nearMint',
          sprache: 'en',
          finish: 'nonfoil',
          name: 'Counterspell',
          setcode: 'lea',
          nummer: '54',
        },
      ],
      'addieren',
    )
  })
})

// **Eine eindeutige Zeile geht ohne Zutun mit.**
it('nimmt eindeutige Zeilen ohne weitere Wahl auf', async () => {
  const einlesen = vi.fn(() =>
    Promise.resolve({
      bericht: { angelegt: [1], erhoeht: 0, stueck: 4, befunde: [] },
      fehler: null,
    }),
  )

  zeige([eindeutige], einlesen)

  await tippeUndDeute('4 Lightning Bolt (M11) 149')

  await screen.findByText('Lightning Bolt · Magic 2011 #149')
  await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))

  await waitFor(() => {
    expect(einlesen).toHaveBeenCalledTimes(1)
  })

  // **Über den Aufruf und nicht über `mock.calls`**: die Argumente sind hier
  // `unknown[]`, und `.at(0)?.at(0)` ist eslint zu Recht ein Ausdruck zu viel.
  expect(einlesen).toHaveBeenCalledWith(
    [
      {
        printing: 'druck-1',
        menge: 4,
        zustand: 'nearMint',
        sprache: 'en',
        finish: 'nonfoil',
        name: 'Lightning Bolt',
        setcode: 'm11',
        nummer: '149',
      },
    ],
    'addieren',
  )
})

// **Der Bericht steht danach da.**
it('sagt, was aufgenommen wurde', async () => {
  zeige(
    [eindeutige],
    vi.fn(() =>
      Promise.resolve({
        bericht: { angelegt: [1], erhoeht: 2, stueck: 7, befunde: [] },
        fehler: null,
      }),
    ),
  )

  await tippeUndDeute('4 Lightning Bolt (M11) 149')
  await screen.findByText('Lightning Bolt · Magic 2011 #149')
  await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))

  expect(
    await screen.findByText('1 neu, 2 erhöht, 7 Stück aufgenommen.'),
  ).toBeInTheDocument()
})

// **Wie viele Zeilen offen sind, steht da** (§10: keine stille Kappung).
it('nennt die Zahl der offenen Zeilen', async () => {
  zeige([eindeutige, offene])

  await tippeUndDeute('zwei Zeilen')

  expect(
    await screen.findByText(
      '1 von 2 Zeilen brauchen eine Entscheidung — wähle den Druck, den du besitzt.',
    ),
  ).toBeInTheDocument()
})

// **Ein Fehler beim Deuten steht am Kasten und nicht im Nichts** (§6).
it('zeigt einen Fehler beim Zuordnen', async () => {
  belegeApi({
    ListeVorschauen: ohneDublette,
    ListeDeuten: () =>
      Promise.resolve({
        deutung: null,
        fehler: {
          code: 'validierung',
          meldung: 'Liste ist leer',
          feld: '',
          schluessel: '',
        },
      }),
  })

  render(
    <Sprachraum sprache="de">
      <Listeneinfuhr aufAufgenommen={aufgenommen} />
    </Sprachraum>,
  )

  await tippeUndDeute('x')

  expect(await screen.findByRole('alert')).toHaveTextContent('Liste ist leer')
})

// **Ein Befund des Zerteilers steht am Kasten** (§8).
//
// Eine Zeile, die schon der Zerteiler nicht lesen konnte, hat keine
// Zeilendeutung — sie kommt als Befund neben der Liste an.
it('zeigt die Befunde des Zerteilers', async () => {
  belegeApi({
    ListeVorschauen: ohneDublette,
    ListeDeuten: () =>
      Promise.resolve({
        deutung: {
          zeilen: [eindeutige],
          befunde: ['Zeile 2: "7" ist eine Zahl und kein Kartenname'],
          alterSekunden: 0,
        },
        fehler: null,
      }),
  })

  render(
    <Sprachraum sprache="de">
      <Listeneinfuhr aufAufgenommen={aufgenommen} />
    </Sprachraum>,
  )

  await tippeUndDeute('4 Lightning Bolt (M11) 149\n7')

  expect(
    await screen.findByText('Zeile 2: "7" ist eine Zahl und kein Kartenname'),
  ).toBeInTheDocument()
})

// **Ein Befund an einer Zeile steht neben ihr.**
//
// Der Unterschied zum Befund des Zerteilers darüber: dieser gehört zu einer
// **gedeuteten** Zeile — »kein Druck heißt so«.
it('zeigt den Befund einer einzelnen Zeile', async () => {
  zeige([
    {
      ...eindeutige,
      treffer: [],
      drucke: 0,
      eindeutig: false,
      befund: 'kein Druck heißt so',
      weitere: false,
      fehler: null,
    },
  ])

  await tippeUndDeute('1 Gibtesnicht')

  // **Der Satz aus dem Katalog, nicht der aus Go** (G01): der stand deutsch
  // in jeder Sprache da (`deutungszeile.test.tsx`).
  expect(
    await screen.findByText('Keine Karte heißt so — prüfe die Schreibweise.'),
  ).toBeInTheDocument()
})

// **Wenn die Auswahl gekappt ist, steht die Gesamtzahl daneben** (§10).
//
// Die Liste zeigt höchstens zwanzig Drucke; »Counterspell« hat 72. Ohne diese
// Zahl sähe eine gekappte Auswahl wie eine vollständige aus.
it('nennt die Gesamtzahl, wenn die Auswahl gekappt ist', async () => {
  zeige([{ ...offene, drucke: 72 }])

  await tippeUndDeute('2 Counterspell')

  expect(await screen.findByText('72 Drucke')).toBeInTheDocument()
})
