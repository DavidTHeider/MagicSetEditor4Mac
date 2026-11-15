// **Ein Tooltip, der auch ohne Maus erscheint** — Werkstück U2 von Phase 5.1
// (#1413).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { Tooltip } from './Tooltip'
import { merkeEscape } from '../test/escape'
import { Verdeckung } from '../verdeckung'

function Probe() {
  return (
    <>
      <Tooltip text="Legt eine Kopie hinter diese Karte." verzoegerung={0}>
        {(griffe) => (
          <button type="button" {...griffe}>
            Duplizieren
          </button>
        )}
      </Tooltip>
      <button type="button">Danach</button>
    </>
  )
}

describe('Tooltip', () => {
  it('beschreibt den Auslöser, auch solange er verborgen ist', () => {
    render(<Probe />)

    const knopf = screen.getByRole('button', { name: 'Duplizieren' })
    expect(knopf).toHaveAccessibleDescription(
      'Legt eine Kopie hinter diese Karte.',
    )
    expect(screen.getByRole('tooltip', { hidden: true })).not.toBeVisible()
  })

  it('erscheint beim Fokus und verschwindet beim Verlassen', async () => {
    render(<Probe />)

    await userEvent.tab()
    expect(screen.getByRole('tooltip')).toBeVisible()

    await userEvent.tab()
    expect(screen.getByRole('tooltip', { hidden: true })).not.toBeVisible()
  })

  it('erscheint unter dem Zeiger und verschwindet, wenn er geht', async () => {
    render(<Probe />)

    const knopf = screen.getByRole('button', { name: 'Duplizieren' })
    await userEvent.hover(knopf)
    expect(await screen.findByRole('tooltip')).toBeVisible()

    await userEvent.unhover(knopf)
    expect(screen.getByRole('tooltip', { hidden: true })).not.toBeVisible()
  })

  // **Escape schließt ihn** (WCAG 1.4.13), ohne dass der Fokus geht.
  it('schließt mit Escape', async () => {
    render(<Probe />)

    await userEvent.tab()
    expect(screen.getByRole('tooltip')).toBeVisible()

    await userEvent.keyboard('{Escape}')
    expect(screen.getByRole('tooltip', { hidden: true })).not.toBeVisible()
    expect(screen.getByRole('button', { name: 'Duplizieren' })).toHaveFocus()
  })

  // **Ein sichtbarer Hinweis verbraucht das Escape, das ihn schließt** — ein
  // verborgener nicht. Sonst schlösse eine Schublade mit demselben Druck, der
  // nur den Hinweis wegnehmen sollte; und ohne die zweite Hälfte ginge sie
  // nie mehr zu (UX-08, #1680).
  it('verbraucht Escape nur, solange er zu sehen ist', async () => {
    render(<Probe />)

    await userEvent.tab()

    const erstes = merkeEscape()
    await userEvent.keyboard('{Escape}')
    expect(erstes()).toBe(true)

    const zweites = merkeEscape()
    await userEvent.keyboard('{Escape}')
    expect(zweites()).toBe(false)
  })

  // **Unter einer Schublade verbraucht er nichts** (UX-08, #1680) — sonst
  // nähme ein Hinweis, der beim Öffnen noch stand, ihr das erste Escape.
  it('lässt Escape durch, solange seine Seite verdeckt ist', async () => {
    render(
      <Verdeckung value>
        <Probe />
      </Verdeckung>,
    )

    await userEvent.tab()
    expect(screen.getByRole('tooltip')).toBeVisible()

    const verbraucht = merkeEscape()
    await userEvent.keyboard('{Escape}')

    expect(verbraucht()).toBe(false)
  })

  // **Ein Zeiger, der nur vorbeigeht, reißt nichts auf.**
  it('wartet unter dem Zeiger die Verzögerung ab', async () => {
    render(
      <Tooltip text="Erklärung" verzoegerung={100000}>
        {(griffe) => (
          <button type="button" {...griffe}>
            Knopf
          </button>
        )}
      </Tooltip>,
    )

    const knopf = screen.getByRole('button', { name: 'Knopf' })
    await userEvent.hover(knopf)
    expect(screen.getByRole('tooltip', { hidden: true })).not.toBeVisible()

    await userEvent.unhover(knopf)
    expect(screen.getByRole('tooltip', { hidden: true })).not.toBeVisible()
  })
})

// **Der Zeiger darf auf den Text** — WCAG 1.4.13, „Hoverable": was durch
// Zeigen erscheint, muss sich auch überfahren lassen, ohne zu verschwinden.
// Befund 4 des Zweitblicks über Phase 5.1
// (#1405):
// bis zum 2026-09-10 verbarg das Verlassen des Auslösers ihn sofort, und der
// Text selbst hatte keine Zeigerbehandlung — der Weg dorthin führte durch
// nichts.
describe('Tooltip und der Zeiger', () => {
  it('bleibt stehen, wenn der Zeiger auf den Text wandert', async () => {
    render(<Probe />)

    await userEvent.hover(screen.getByRole('button', { name: 'Duplizieren' }))

    const text = await screen.findByRole('tooltip')
    expect(text).toBeVisible()

    await userEvent.hover(text)
    expect(screen.getByRole('tooltip')).toBeVisible()
  })

  it('geht, sobald der Zeiger auch den Text verlässt', async () => {
    render(<Probe />)

    await userEvent.hover(screen.getByRole('button', { name: 'Duplizieren' }))
    await userEvent.hover(await screen.findByRole('tooltip'))

    await userEvent.unhover(screen.getByRole('tooltip'))

    expect(screen.getByRole('tooltip', { hidden: true })).not.toBeVisible()
  })
})

// **Ein Druck räumt ihn** (#2068). Er liegt unter seinem Auslöser und fängt
// den Zeiger — in einer Reihe von Knöpfen also auf dem nächsten. Wer klickt
// und gleich darunter weiterklickt, traf bis zum 2026-10-05 den Hinweis.
describe('Tooltip und ein Druck', () => {
  it('verschwindet beim Druck auf den Auslöser, auch wenn der Fokus folgt', async () => {
    render(<Probe />)

    const knopf = screen.getByRole('button', { name: 'Duplizieren' })
    await userEvent.hover(knopf)
    expect(await screen.findByRole('tooltip')).toBeVisible()

    await userEvent.click(knopf)

    expect(knopf).toHaveFocus()
    expect(screen.getByRole('tooltip', { hidden: true })).not.toBeVisible()
  })

  it('verschwindet beim Druck auf den Text selbst', async () => {
    render(<Probe />)

    await userEvent.hover(screen.getByRole('button', { name: 'Duplizieren' }))
    const text = await screen.findByRole('tooltip')

    await userEvent.click(text)

    expect(screen.getByRole('tooltip', { hidden: true })).not.toBeVisible()
  })

  it('kommt mit dem nächsten Zeiger wieder', async () => {
    render(<Probe />)

    const knopf = screen.getByRole('button', { name: 'Duplizieren' })
    await userEvent.click(knopf)
    await userEvent.unhover(knopf)
    expect(screen.getByRole('tooltip', { hidden: true })).not.toBeVisible()

    await userEvent.hover(knopf)
    expect(await screen.findByRole('tooltip')).toBeVisible()
  })

  it('kommt mit dem nächsten Fokus von der Tastatur wieder', async () => {
    render(<Probe />)

    await userEvent.click(screen.getByRole('button', { name: 'Duplizieren' }))
    await userEvent.tab()
    await userEvent.tab({ shift: true })

    expect(screen.getByRole('tooltip')).toBeVisible()
  })
})

// **Am Fragezeichen bleibt er** (#2068): dort ist der Hinweis die Antwort
// auf den Druck, und die Feldhilfe sagt es mit `druck="bleibt"`.
it('bleibt beim Druck, wo der Druck ihn sehen will', async () => {
  render(
    <Tooltip text="Erklärung" verzoegerung={0} druck="bleibt">
      {(griffe) => (
        <button type="button" {...griffe}>
          ?
        </button>
      )}
    </Tooltip>,
  )

  const knopf = screen.getByRole('button', { name: '?' })
  await userEvent.hover(knopf)
  await userEvent.click(knopf)

  expect(screen.getByRole('tooltip')).toBeVisible()
})
