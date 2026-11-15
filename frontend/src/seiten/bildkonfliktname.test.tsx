// **Der Namenskonflikt heißt auf der Einstellungsseite, wie er auf der
// Bildseite heißt** — Sichtprüfung am nativen Fenster für OPS-02
// (#1719):
// unter *Daten und Sicherung* standen „bildkonflikt" und „nummerierenneu"
// roh da. Seine Namen kommen aus dem Katalog der Bildseite
// (`einstellungsnamenswege_test.go`: „aus eigener Seite") — und die
// Einstellungsseite fragte dort nicht nach.
import { render, screen, within } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { Einstellungsgruppe } from './einstellungsgebiete'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

const bilder = kataloge.de.seiten.bilder

it('nennt den Namenskonflikt und seine Werte mit ihren Namen', () => {
  render(
    <Sprachraum sprache="de">
      <Einstellungsgruppe
        name="Daten und Sicherung"
        eintraege={[
          {
            schluessel: 'bildkonflikt',
            wert: 'nummerierenneu',
            wirksam: 'nummerierenneu',
            // Alle vier, wie Go sie schickt: vier Werte sind seit UXR-F2
            // (#2008) vier Segmente.
            werte: [
              'altbehalten',
              'ueberschreiben',
              'nummerieren',
              'nummerierenneu',
            ],
            bereich: 'bildausfuhr',
          },
        ]}
        laeuft={() => false}
        setze={vi.fn()}
      />
    </Sprachraum>,
  )

  const reihe = screen.getByRole('radiogroup', { name: bilder.konflikt })

  expect(
    within(reihe).getByRole('radio', { name: bilder.konflikte.nummerierenneu }),
  ).toBeChecked()
  expect(
    within(reihe)
      .getAllByRole('radio')
      .map((feld) => feld.closest('label')?.textContent),
  ).toEqual([
    bilder.konflikte.altbehalten,
    bilder.konflikte.ueberschreiben,
    bilder.konflikte.nummerieren,
    bilder.konflikte.nummerierenneu,
  ])
  expect(screen.queryByText('bildkonflikt')).toBeNull()
})
