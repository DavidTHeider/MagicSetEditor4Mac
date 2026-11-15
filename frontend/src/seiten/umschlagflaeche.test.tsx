// **Die beiden Flächen des portablen Umschlags** — Werkstück N4 von
// Phase 5.4
// (#1457).
//
// # Wonach hier gefragt wird
//
// Danach, dass die Vorschau **vor** der Übernahme steht, dass der Abdruck
// seinen Satz danebenhat — er sagt *wer*, nicht *ob man trauen sollte* — und
// dass ein Umschlag ohne Unterschrift keine leere Zeile hinterlässt.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { stillerUmschlaglauf } from '../test/umschlaglauf'
import { Umschlagannahme } from './Umschlagannahme'
import { kurzform } from '../sprache/kurzform'

const texte = kataloge.de.seiten.sets

const angesehen = {
  objekte: [
    { art: 'set', bezeichnung: 'set.mse-set', groesse: 12, braucht: [] },
  ],
  urheber: 'k51qzi5',
  fingerabdruck: 'AB CD EF',
  pfad: '/tmp/u.msepackage',
  abgebrochen: false,
  fehler: null,
}

function zeige(lauf = stillerUmschlaglauf()) {
  return render(
    <Sprachraum sprache="de">
      <Umschlagannahme umschlag={lauf} />
    </Sprachraum>,
  )
}

// **Der Knopf an der Setkarte steht seit UXR-04 (#1912) in ihrem Menü** —
// geprüft in `setkarte.test.tsx`, samt Setkennung, Vorgangsname und der
// Meldung an der Karte.

describe('die Annahme', () => {
  it('zeigt vor der Wahl nur den Weg dorthin', () => {
    zeige()

    expect(
      screen.getByRole('button', { name: texte.umschlagWaehlen }),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: texte.umschlagUebernehmen }),
    ).not.toBeInTheDocument()
  })

  it('fragt nach einem Umschlag', async () => {
    const waehlen = vi.fn()

    zeige(stillerUmschlaglauf({ waehlen }))

    await userEvent.click(
      screen.getByRole('button', { name: texte.umschlagWaehlen }),
    )

    expect(waehlen).toHaveBeenCalled()
  })

  it('zeigt Urheber, Abdruck und Inhalt — und den Satz zum Abdruck', () => {
    zeige(stillerUmschlaglauf({ angesehen }))

    expect(screen.getByText('AB CD EF')).toBeInTheDocument()
    expect(screen.getByText('set.mse-set')).toBeInTheDocument()

    // **Ohne diesen Satz läse sich eine Kennung wie eine
    // Unbedenklichkeitsbescheinigung.**
    expect(
      screen.getByText(kurzform(texte.umschlagAbdruckHinweis)),
    ).toBeInTheDocument()
  })

  it('lässt den Abdruck weg, wo keiner steht', () => {
    zeige(stillerUmschlaglauf({ angesehen: { ...angesehen, urheber: '' } }))

    expect(
      screen.queryByText(kurzform(texte.umschlagAbdruckHinweis)),
    ).not.toBeInTheDocument()
    expect(screen.getByText('set.mse-set')).toBeInTheDocument()
  })

  it('reicht den Pfad durch, den die Vorschau kennt', async () => {
    const uebernehmen = vi.fn()

    zeige(stillerUmschlaglauf({ angesehen, uebernehmen }))

    await userEvent.click(
      screen.getByRole('button', { name: texte.umschlagUebernehmen }),
    )

    expect(uebernehmen).toHaveBeenCalledWith('/tmp/u.msepackage')
  })

  it('räumt auf, wer es sich anders überlegt', async () => {
    const raeumen = vi.fn()

    zeige(stillerUmschlaglauf({ angesehen, raeumen }))

    await userEvent.click(
      screen.getByRole('button', { name: texte.umschlagVerwerfen }),
    )

    expect(raeumen).toHaveBeenCalled()
  })

  it('hält beide Knöpfe an, solange die Übernahme läuft', () => {
    zeige(
      stillerUmschlaglauf({
        angesehen,
        laeuft: (was: string) => was === 'umschlag-uebernehmen',
      }),
    )

    expect(
      screen.getByRole('button', { name: texte.umschlagUebernehmen }),
    ).toBeDisabled()
    expect(
      screen.getByRole('button', { name: texte.umschlagVerwerfen }),
    ).toBeDisabled()
  })

  it('hält die Wahl an, solange sie läuft', () => {
    zeige(
      stillerUmschlaglauf({
        laeuft: (was: string) => was === 'umschlag-waehlen',
      }),
    )

    expect(
      screen.getByRole('button', { name: texte.umschlagWaehlen }),
    ).toBeDisabled()
  })

  // **Was das Ablegen meldet, steht an seiner Setkarte** (UXR-04, #1912) —
  // hier stehen nur die Meldungen der Annahme. `fuer` sagt, wem ein Fehler
  // gehört.
  it('zeigt die Meldungen der Annahme und nicht die des Ablegens', () => {
    const { rerender } = zeige(stillerUmschlaglauf({ fehler: 'kaputt' }))

    expect(screen.getByRole('alert')).toHaveTextContent('kaputt')

    rerender(
      <Sprachraum sprache="de">
        <Umschlagannahme
          umschlag={stillerUmschlaglauf({ fehler: 'kaputt', fuer: '7' })}
        />
      </Sprachraum>,
    )
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()

    rerender(
      <Sprachraum sprache="de">
        <Umschlagannahme
          umschlag={stillerUmschlaglauf({
            geschnuert: { ...angesehen },
            fuer: '7',
          })}
        />
      </Sprachraum>,
    )
    expect(screen.queryByRole('status')).not.toBeInTheDocument()

    rerender(
      <Sprachraum sprache="de">
        <Umschlagannahme umschlag={stillerUmschlaglauf({ uebernommen: '9' })} />
      </Sprachraum>,
    )
    expect(screen.getByRole('status')).toHaveTextContent(
      texte.umschlagUebernommen,
    )
  })
})
