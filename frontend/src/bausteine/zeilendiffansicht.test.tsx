import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Zeilendiff } from './Zeilendiff'

const texte = kataloge.de.seiten.vorlagen

function zeige(alt: string, neu: string) {
  return render(
    <Sprachraum sprache="de">
      <Zeilendiff alt={alt} neu={neu} />
    </Sprachraum>,
  )
}

describe('Die Ansicht des Zeilen-Diffs', () => {
  it('zählt zusammen und zeigt nur die Umgebung der Änderung', () => {
    zeige('a\nb\nc\nd\ne\nf\ng', 'a\nb\nc\nd\nE\nf\ng')

    expect(screen.getByText('1 Zeile(n) neu, 1 entfernt.')).toBeInTheDocument()

    const zeilen = screen.getAllByRole('listitem')
    expect(zeilen.map((zeile) => zeile.textContent)).toEqual([
      '44 d',
      `5−${texte.zeileWeg}e`,
      `5+${texte.zeileNeu}E`,
      '66 f',
    ])
  })

  it('sagt es, wenn nichts anders ist', () => {
    zeige('a\nb', 'a\nb')

    expect(screen.getByText(texte.aenderungenKeine)).toBeInTheDocument()
    expect(screen.queryByRole('list')).not.toBeInTheDocument()
  })

  it('trennt zwei Hünke', () => {
    zeige('a\nb\nc\nd\ne\nf', 'A\nb\nc\nd\ne\nF')

    expect(screen.getAllByRole('list')).toHaveLength(2)
    // **Ein Hunk rollt quer, also erreicht er die Tastatur** (G10) — als
    // Region mit Namen, wie `Diagnose.tsx`.
    expect(screen.getByRole('region', { name: 'Änderung 2' })).toHaveAttribute(
      'tabindex',
      '0',
    )
  })
})
