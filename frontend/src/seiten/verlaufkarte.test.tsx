// **Ein Kartenschritt sagt, ob die Karte kam oder ging** — G01
// (#2248).
//
// Gefunden am nativen Fenster (`g01einsteiger`): die erste Karte stand im
// Verlauf als »Karte angelegt oder gelöscht«, mit »kein Wert« vorher und
// nachher. Die Fassade gibt jetzt mit, ob die Karte vorher und nachher da
// war, und die Zeile sagt es.
import { render, screen, within } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Verlaufsliste } from './Verlaufsliste'

import type { VerlaufsschrittDTO } from '../api'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.verlauf

function kartenschritt(
  id: string,
  gabVorher: boolean,
  gabNachher: boolean,
): VerlaufsschrittDTO {
  return {
    id,
    zeitpunkt: '2026-10-10T22:50:49Z',
    handlung: 'karte',
    karteID: '7',
    feld: '',
    vorher: '',
    gabVorher,
    nachher: '',
    gabNachher,
    zurueckgenommen: false,
    urheber: 'fenster',
    herkunft: '',
    vorgang: '',
    umfang: 1,
  }
}

it('nennt Anlegen und Löschen beim Namen, mit der Karte vorher und nachher', async () => {
  belegeApi({
    Verlauf: () =>
      Promise.resolve({
        schritte: [
          kartenschritt('2', true, false),
          kartenschritt('1', false, true),
        ],
      }),
  })

  render(
    <Sprachraum sprache="de">
      <Verlaufsliste
        setID="1"
        karten={[{ id: '7', nummer: 0, name: 'Feuerfuchs', felder: 2 }]}
        verlaufsstand={0}
      />
    </Sprachraum>,
  )

  const zeilen = within(await screen.findByRole('table')).getAllByRole('row')
  const geloescht = zeilen[1]?.textContent ?? ''
  const angelegt = zeilen[2]?.textContent ?? ''

  expect(geloescht).toContain(texte.handlungKarteGeloescht)
  expect(geloescht).toContain(texte.karteVorhanden)
  expect(geloescht).toContain(texte.keineKarte)
  expect(angelegt).toContain(texte.handlungKarteAngelegt)
  expect(angelegt.indexOf(texte.keineKarte)).toBeLessThan(
    angelegt.indexOf(texte.karteVorhanden),
  )
  expect(angelegt).not.toContain(texte.keinWert)
})

it('nennt eine andere Handlung weiter beim Wort', async () => {
  belegeApi({
    Verlauf: () =>
      Promise.resolve({
        schritte: [
          { ...kartenschritt('3', true, true), handlung: 'reihenfolge' },
        ],
      }),
  })

  render(
    <Sprachraum sprache="de">
      <Verlaufsliste setID="1" karten={[]} verlaufsstand={0} />
    </Sprachraum>,
  )

  const zeilen = within(await screen.findByRole('table')).getAllByRole('row')
  expect(zeilen[1]?.textContent ?? '').toContain(texte.handlungReihenfolge)
})
