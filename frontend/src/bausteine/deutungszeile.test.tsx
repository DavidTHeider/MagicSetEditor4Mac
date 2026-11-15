// **Was eine Zeile sagt, die nicht aufgeht** — G01
// (#2248).
//
// Gefunden am nativen Fenster (`g01einsteiger`): eine Zeile ohne Karte stand
// mit dem rohen Fehler da, deutsch in jeder Sprache; die Bilanz zählte sie zu
// denen, die eine Entscheidung brauchen; und die Auswahl des Drucks ragte
// über den Rand der Schublade.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { Listeneinfuhr } from './Listeneinfuhr'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { ohneDublette } from '../test/einfuhr'

import type { Listendeutungszeile } from '../antwortenlisteneinfuhr'

afterEach(() => {
  raeumeApi()
})

const t = kataloge.en.seiten.sammlung

function zeile(teile: Partial<Listendeutungszeile>): Listendeutungszeile {
  return {
    roh: '1 Probe',
    zeilennummer: 1,
    menge: 1,
    name: 'Probe',
    setcode: '',
    nummer: '',
    finish: 'nonfoil',
    treffer: [],
    drucke: 0,
    eindeutig: false,
    befund: '',
    weitere: false,
    fehler: null,
    ...teile,
  }
}

const druck = (nummer: string) => ({
  printing: `p${nummer}`,
  name: 'Counterspell',
  setcode: 'lea',
  setname: 'Limited Edition Alpha',
  nummer,
  seltenheit: 'common',
  bild: '',
})

// zeige deutet die Zeilen **im englischen Fenster** — ein deutscher Befund
// fiele dort sofort auf.
async function zeige(zeilen: Listendeutungszeile[]) {
  belegeApi({
    ListeVorschauen: ohneDublette,
    ListeDeuten: () =>
      Promise.resolve({
        deutung: { zeilen, befunde: [], alterSekunden: 0 },
        fehler: null,
      }),
  })

  render(
    <Sprachraum sprache="en">
      <Listeneinfuhr aufAufgenommen={() => undefined} />
    </Sprachraum>,
  )

  await userEvent.type(screen.getByLabelText(t.listeFeld), '1 Probe')
  await userEvent.click(screen.getByRole('button', { name: t.listeDeuten }))
}

it('sagt in der Sprache des Fensters, dass keine Karte so heißt', async () => {
  await zeige([
    zeile({
      zeilennummer: 1,
      roh: '1 Gibtesnichtxyz',
      befund: 'kein Druck heißt so',
    }),
  ])

  expect(await screen.findByText(t.listeKeinDruck)).toBeInTheDocument()
  expect(screen.queryByText('kein Druck heißt so')).toBeNull()
})

it('übersetzt den Ausfall einer Zeile, statt ihn roh zu zeigen', async () => {
  await zeige([
    zeile({
      befund: 'nicht gefunden: roh',
      fehler: {
        code: 'unbekannt',
        meldung: 'Roh auf Deutsch',
        schluessel: 'unbekannt',
      },
    }),
  ])

  expect(
    await screen.findByText(kataloge.en.fehler.unbekannt),
  ).toBeInTheDocument()
  expect(screen.queryByText('nicht gefunden: roh')).toBeNull()
})

it('sagt, dass es mehr Drucke gibt, als eine Seite nennt', async () => {
  await zeige([
    zeile({
      treffer: [druck('1'), druck('2')],
      drucke: 175,
      weitere: true,
      befund: 'mehr als 175 Drucke heißen so',
    }),
  ])

  expect(
    await screen.findByText(t.listeZuViele.replace('{drucke}', '175')),
  ).toBeInTheDocument()
  expect(screen.queryByText('mehr als 175 Drucke heißen so')).toBeNull()
})

it('zeigt, was die Quelle sonst noch meldet, wie es kommt', async () => {
  await zeige([
    zeile({
      treffer: [druck('1')],
      drucke: 1,
      eindeutig: true,
      befund: 'Scryfall nennt eine unbekannte Seltenheit',
    }),
  ])

  expect(
    await screen.findByText('Scryfall nennt eine unbekannte Seltenheit'),
  ).toBeInTheDocument()
})

it('zählt Zeilen ohne Karte nicht zu denen, die eine Entscheidung brauchen', async () => {
  await zeige([
    zeile({
      zeilennummer: 1,
      treffer: [druck('1')],
      drucke: 1,
      eindeutig: true,
    }),
    zeile({ zeilennummer: 2, treffer: [druck('1'), druck('2')], drucke: 2 }),
    zeile({ zeilennummer: 3, befund: 'kein Druck heißt so' }),
  ])

  expect(
    await screen.findByText(
      t.listeOffen.replace('{offen}', '1').replace('{gesamt}', '3'),
    ),
  ).toBeInTheDocument()
  expect(screen.getByText(t.listeOhneKarteEins)).toBeInTheDocument()
})

it('nennt mehrere Zeilen ohne Karte mit ihrer Zahl', async () => {
  await zeige([
    zeile({
      zeilennummer: 1,
      treffer: [druck('1')],
      drucke: 1,
      eindeutig: true,
    }),
    zeile({ zeilennummer: 2, befund: 'kein Druck heißt so' }),
    zeile({ zeilennummer: 3, befund: 'kein Druck heißt so' }),
  ])

  expect(
    await screen.findByText(t.listeOhneKarte.replace('{anzahl}', '2')),
  ).toBeInTheDocument()
  expect(
    screen.queryByText(t.listeAlleEindeutig.replace('{gesamt}', '3')),
  ).toBeNull()
})

// **Die Auswahl bleibt in der Schublade**: jsdom misst keine Breite, also
// wird geprüft, was sie begrenzt — der nachweis am Fenster steht im Beleg
// `g01einsteiger`.
it('begrenzt die Auswahl des Drucks auf die Breite der Zeile', async () => {
  await zeige([zeile({ treffer: [druck('1'), druck('2')], drucke: 2 })])

  const wahl = await screen.findByRole('combobox', { name: t.listeDruckWahl })
  expect(wahl).toHaveClass('min-w-0', 'max-w-full')
  expect(wahl.closest('label')).toHaveClass('min-w-0', 'max-w-full')
})
