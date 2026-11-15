// **Ein Set lässt sich löschen — nach einer Bestätigung mit seinem Namen**
// (#1218).
//
// # Warum der Name und nicht ein „Ja"
//
// Weil es keinen Papierkorb gibt (Vermerk 0054). Ein „Ja" lässt sich
// wegklicken, ohne gelesen zu haben; ein Name, den man abtippt, nicht.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import type { SetDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { Setloeschen } from './Setloeschen'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

const probeset: SetDTO = {
  id: '7',
  name: 'Abnahme 3.5',
  spiel: 'probe',
  stil: 'probe-einzeiler',
  karten: 3,
  herkunft: '',
  herkunftFehlt: false,
}

function zeige(
  loeschen: (setID: string, name: string) => void = () => undefined,
  abbrechen: () => void = vi.fn(),
) {
  render(
    <Sprachraum sprache="de">
      <Setloeschen
        satz={probeset}
        laeuft={false}
        loeschen={loeschen}
        abbrechen={abbrechen}
        befunde={[]}
        fehler={undefined}
      />
    </Sprachraum>,
  )
}

// **Gefragt wird über das Menü der Setkarte** seit UXR-04 (#1912) — ob die
// Frage dasteht, entscheidet die Karte (`setkarte.test.tsx`); hier steht,
// was sie fragt. Bis dahin trug sie ihren eigenen Knopf „Löschen …".
it('setzt den Fokus in das Feld, sobald sie dasteht', () => {
  zeige()

  expect(screen.getByLabelText(texte.seiten.sets.loeschenFeld)).toHaveFocus()
})

it('nennt das Set und seine Kartenzahl in der Warnung', () => {
  zeige()

  const warnung = screen.getByRole('alert')
  expect(warnung).toHaveTextContent('Abnahme 3.5')
  expect(warnung).toHaveTextContent('3 Karten')
})

// **Und bei einer Karte steht die Einzahl.**
//
// Gefunden bei der Sichtprüfung im laufenden Fenster: dort stand „mit allen 1
// Karten". Ein Test konnte es nicht sehen, denn die Zeichenfolge war ja da —
// der Wächter über Texte im Quelltext prüft, dass keiner dasteht, nicht ob
// der im Katalog stimmt.
it('sagt »1 Karte« und nicht »1 Karten«', () => {
  render(
    <Sprachraum sprache="de">
      <Setloeschen
        satz={{ ...probeset, karten: 1 }}
        laeuft={false}
        loeschen={() => undefined}
        abbrechen={vi.fn()}
        befunde={[]}
        fehler={undefined}
      />
    </Sprachraum>,
  )

  expect(screen.getByRole('alert')).toHaveTextContent('samt 1 Karte.')
})

it('hält den Knopf zu, solange der Name nicht stimmt', async () => {
  const gerufen = vi.fn()
  zeige(gerufen)

  await userEvent.type(
    screen.getByLabelText(texte.seiten.sets.loeschenFeld),
    'Abnahme 3',
  )

  const knopf = screen.getByRole('button', {
    name: texte.seiten.sets.loeschenJa,
  })
  expect(knopf).toBeDisabled()

  await userEvent.click(knopf)
  expect(gerufen).not.toHaveBeenCalled()

  // **Und jetzt der Beleg, dass der Knopf überhaupt hört** (§3): dasselbe
  // Feld, derselbe Knopf, nur der Name vollständig. Ohne diesen zweiten Teil
  // wäre die Verneinung darüber auch dann grün, wenn niemand zuhört.
  await userEvent.type(
    screen.getByLabelText(texte.seiten.sets.loeschenFeld),
    'ct.5',
  )
  await userEvent.clear(screen.getByLabelText(texte.seiten.sets.loeschenFeld))
  await userEvent.type(
    screen.getByLabelText(texte.seiten.sets.loeschenFeld),
    'Abnahme 3.5',
  )
  await userEvent.click(
    screen.getByRole('button', { name: texte.seiten.sets.loeschenJa }),
  )

  await waitFor(() => {
    expect(gerufen).toHaveBeenCalledWith('7', 'Abnahme 3.5')
  })
})

it('nimmt die Groß- und Kleinschreibung genau', async () => {
  const gerufen = vi.fn()
  zeige(gerufen)

  await userEvent.type(
    screen.getByLabelText(texte.seiten.sets.loeschenFeld),
    'abnahme 3.5',
  )

  expect(
    screen.getByRole('button', { name: texte.seiten.sets.loeschenJa }),
  ).toBeDisabled()
})

// **Abbrechen gibt die Frage an die Karte zurück** — sie hängt sie aus, und
// mit ihr geht der getippte Name (`setkarte.test.tsx`).
it('gibt beim Abbrechen Bescheid', async () => {
  const abbrechen = vi.fn()
  zeige(undefined, abbrechen)

  await userEvent.type(
    screen.getByLabelText(texte.seiten.sets.loeschenFeld),
    'Abnahme 3.5',
  )
  await userEvent.click(
    screen.getByRole('button', { name: texte.seiten.sets.loeschenNein }),
  )

  expect(abbrechen).toHaveBeenCalledTimes(1)
})

it('zeigt die Befunde eines geglückten Löschens', () => {
  render(
    <Sprachraum sprache="de">
      <Setloeschen
        satz={probeset}
        laeuft={false}
        loeschen={() => undefined}
        abbrechen={vi.fn()}
        befunde={['Die Bilder ließen sich nicht wegräumen.']}
        fehler={undefined}
      />
    </Sprachraum>,
  )

  expect(
    screen.getByText('Die Bilder ließen sich nicht wegräumen.'),
  ).toBeInTheDocument()
})

it('zeigt den Fehler am Feld', () => {
  render(
    <Sprachraum sprache="de">
      <Setloeschen
        satz={probeset}
        laeuft={false}
        loeschen={() => undefined}
        abbrechen={vi.fn()}
        befunde={[]}
        fehler="Das Set heißt anders."
      />
    </Sprachraum>,
  )

  expect(screen.getByText('Das Set heißt anders.')).toBeInTheDocument()
})
