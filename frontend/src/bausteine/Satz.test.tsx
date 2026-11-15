// **Ein Satz zeigt seine Hervorhebungen, nicht ihre Zeichen** —
// #699.
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Satz } from './Satz'

describe('Satz', () => {
  it('macht aus zwei Sternchen eine Hervorhebung', () => {
    render(<Satz text="Sagt, **welche Felder** eine Karte hat." />)

    expect(screen.getByText('welche Felder').tagName).toBe('STRONG')
    expect(screen.queryByText(/\*\*/)).not.toBeInTheDocument()
  })

  it('macht aus Rückstrichen einen Code', () => {
    render(<Satz text="Sie liegt in einer Datei `.mse-set`." />)

    expect(screen.getByText('.mse-set').tagName).toBe('CODE')
  })

  it('lässt beides in einem Satz gelten', () => {
    render(<Satz text="**Ein** Wort und `ein` Code." />)

    expect(screen.getByText('Ein').tagName).toBe('STRONG')
    expect(screen.getByText('ein').tagName).toBe('CODE')
  })

  // **Was nicht paart, bleibt stehen.** Ein einzelnes Sternchen ist ein
  // Sternchen — wer es fräße, änderte einen Satz, den jemand so gemeint hat.
  it('lässt ein einzelnes Zeichen stehen', () => {
    render(<Satz text="Ein * und ein ` bleiben." />)

    expect(screen.getByText('Ein * und ein ` bleiben.')).toBeInTheDocument()
  })

  // **Zweimal dasselbe Wort steht zweimal da.** Der Schlüssel ist die Stelle;
  // wäre er der Text, zeichnete React das zweite nicht.
  it('zeigt dieselbe Hervorhebung zweimal', () => {
    render(<Satz text="**A** und **A**." />)

    expect(screen.getAllByText('A')).toHaveLength(2)
  })
})
