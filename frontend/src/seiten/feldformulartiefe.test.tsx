// **Drei Tiefen im Kartenformular** — UX-05.AK-01 und AK-02
// (#1677).
//
// # Wogegen hier geprüft wird
//
// **Dass die Tiefen ineinanderliegen.** Wer von `normal` nach `experte`
// geht, sieht mehr — nie anderes. Die Rechnung dahinter steht in
// `feldtiefe.test.ts`; hier steht, was daraus im Baum wird.
//
// **Und dass ein Befund in einem verdeckten Feld dasteht.** AK-02 verlangt
// eine „direkte Öffnungsaktion"; ohne sie stünde eine Karte da, die sich
// nicht speichern lässt, und niemand sähe, woran es liegt.
import { createRef } from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Feldformular } from './Feldformular'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { FeldbeschreibungDTO } from '../api'
import type { Feldtiefe } from './feldtiefe'

const texte = kataloge.de.seiten.karten

function beschreibung(teile: Partial<FeldbeschreibungDTO> = {}) {
  return {
    name: 'name',
    aufschrift: 'Name',
    art: 'text',
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    editierbar: true,
    gespeichert: true,
    teilfelder: [],
    reihe: '',
    stelle: 0,
    ...teile,
  }
}

// dreiFelder gibt je ein Feld in jeder Tiefe.
const dreiFelder = [
  { feld: 'name', wert: 'Aschenwächter' },
  { feld: 'farbe', wert: 'gruen' },
  { feld: 'stylesheet', wert: 'schlicht' },
]

const nachName = new Map([
  ['name', beschreibung()],
  ['farbe', beschreibung({ name: 'farbe', editierbar: false })],
])

function zeige(
  tiefe: Feldtiefe,
  teile: {
    probleme?: { feld?: string; meldung: string }[]
    aufTiefe?: (tiefe: Feldtiefe) => void
  } = {},
) {
  render(
    <Sprachraum sprache="de">
      <Feldformular
        felder={dreiFelder}
        nachName={nachName}
        tiefe={tiefe}
        probleme={teile.probleme ?? []}
        aufTiefe={teile.aufTiefe ?? vi.fn()}
        zeile={(feld) => <span key={feld.feld}>{feld.feld}</span>}
        formular={createRef<HTMLFormElement>()}
      />
    </Sprachraum>,
  )
}

describe('das Kartenformular in drei Tiefen', () => {
  it('zeigt in normal nur, was das Spiel erklärt und wer ausfüllt', () => {
    zeige('normal')

    expect(screen.getByText('name')).toBeInTheDocument()
    expect(screen.queryByText('farbe')).not.toBeInTheDocument()
    expect(screen.queryByText('stylesheet')).not.toBeInTheDocument()
  })

  it('nimmt in erweitert die skripterzeugten dazu', () => {
    zeige('erweitert')

    expect(screen.getByText('name')).toBeInTheDocument()
    expect(screen.getByText('farbe')).toBeInTheDocument()
    expect(screen.queryByText('stylesheet')).not.toBeInTheDocument()
  })

  it('nimmt in experte die Buchführung dazu', () => {
    zeige('experte')

    expect(screen.getByText('name')).toBeInTheDocument()
    expect(screen.getByText('farbe')).toBeInTheDocument()
    expect(screen.getByText('stylesheet')).toBeInTheDocument()
  })
})

describe('ein Befund unter einer verdeckten Tiefe', () => {
  const satz = einsetzen(texte.befundVerdeckt, { tiefe: texte.tiefeExperte })

  it('sagt, wo er liegt, und bietet den Sprung an', async () => {
    const aufTiefe = vi.fn()

    zeige('normal', {
      probleme: [{ feld: 'stylesheet', meldung: 'kaputt' }],
      aufTiefe,
    })

    expect(screen.getByText(satz)).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: texte.befundSpringen }),
    )

    expect(aufTiefe).toHaveBeenCalledWith('experte')
  })

  // **Die Gegenprobe**: liegt der Befund im Sichtbaren, steht der Satz nicht
  // da. Sonst sagte der Lauf oben nur, dass irgendein Hinweis erscheint.
  it('sagt nichts, wo der Befund schon zu sehen ist', () => {
    zeige('normal', { probleme: [{ feld: 'name', meldung: 'kaputt' }] })

    expect(screen.queryByText(satz)).not.toBeInTheDocument()
  })

  it('sagt nichts ohne Befund', () => {
    zeige('normal')

    expect(screen.queryByText(satz)).not.toBeInTheDocument()
  })
})
