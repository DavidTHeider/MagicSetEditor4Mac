// **Ein Knopf, der angehalten wird, behält seinen Platz im Tabweg** —
// #1990.
//
// Gemessen am 2026-10-01 in der Abnahme-VM mit VoiceOver (#1871): nach
// *Add card*, nach *Create code* und nach *Choose picture …* mit Escape
// landete das nächste ⇥ bei „Skip to content". Der Knopf wird während seines
// Vorgangs `disabled`, und **ein fokussiertes Element, das `disabled` wird,
// verliert den Fokus an den Seitenanfang** — WebKit tut das, jsdom nicht.
// Die Tests stellen es deshalb mit `blur()` nach.
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it } from 'vitest'

import { Knopf } from './Knopf'

function Probe({ angehalten }: { angehalten: boolean }) {
  return (
    <>
      <input aria-label="davor" />
      <Knopf angehalten={angehalten} aufKlick={() => undefined}>
        Anlegen
      </Knopf>
    </>
  )
}

it('gibt den Fokus zurück, den er beim Anhalten an den Seitenanfang verlor', async () => {
  const nutzer = userEvent.setup()
  const { rerender } = render(<Probe angehalten={false} />)
  const knopf = screen.getByRole('button', { name: 'Anlegen' })

  knopf.focus()
  await nutzer.keyboard('{Enter}')

  // **Der Fokusverlust, wie WebKit ihn macht** — jsdom gibt den Fokus eines
  // Knopfs, der schon `disabled` ist, bei `blur()` nicht ab; der Endzustand
  // ist derselbe.
  knopf.blur()
  rerender(<Probe angehalten />)
  expect(document.body).toHaveFocus()

  rerender(<Probe angehalten={false} />)

  expect(knopf).toHaveFocus()
})

// **Wer inzwischen woanders ist, bleibt dort** — etwa, wer während des
// Vorgangs mit ⇥ weitergegangen ist.
it('lässt den Fokus, wo er inzwischen steht', async () => {
  const nutzer = userEvent.setup()
  const { rerender } = render(<Probe angehalten={false} />)
  const knopf = screen.getByRole('button', { name: 'Anlegen' })
  const feld = screen.getByRole('textbox', { name: 'davor' })

  knopf.focus()
  await nutzer.keyboard('{Enter}')

  rerender(<Probe angehalten />)
  feld.focus()

  rerender(<Probe angehalten={false} />)

  expect(feld).toHaveFocus()
})

// **Ein Klick mit dem Zeiger gibt in WebKit keinen Fokus** (`fokus.ts`), und
// danach soll ihn der Knopf auch nicht an sich ziehen.
it('nimmt keinen Fokus, den er beim Auslösen nicht hatte', () => {
  const { rerender } = render(<Probe angehalten={false} />)
  const knopf = screen.getByRole('button', { name: 'Anlegen' })

  fireEvent.click(knopf)

  rerender(<Probe angehalten />)
  rerender(<Probe angehalten={false} />)

  expect(knopf).not.toHaveFocus()
})
