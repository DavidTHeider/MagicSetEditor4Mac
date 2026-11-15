// **Ein Kasten, der zu ist, bis jemand ihn öffnet** — Werkstück C von Phase
// 4.10 (#1325).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useEffect, useState } from 'react'
import { expect, it, vi } from 'vitest'

import { Schreibrechtraum } from '../haken/schreibrecht'
import { kataloge } from '../sprache/kataloge'
import { Aufklappkasten } from './Aufklappkasten'

it('zeigt seinen Inhalt erst, wenn er geöffnet wird', async () => {
  render(
    <Aufklappkasten titel="Werkzeug">
      <p>Der Inhalt</p>
    </Aufklappkasten>,
  )

  // **Zu ist der Anfangszustand**, und das ist der ganze Zweck: was man
  // einmal einstellt, soll nicht die Stelle belegen, an der die Hauptsache
  // steht.
  expect(screen.queryByText('Der Inhalt')).toBeNull()

  const schalter = screen.getByRole('button', { name: 'Werkzeug' })
  expect(schalter).toHaveAttribute('aria-expanded', 'false')

  await userEvent.click(schalter)

  expect(screen.getByText('Der Inhalt')).toBeVisible()
  expect(schalter).toHaveAttribute('aria-expanded', 'true')

  // **Und wieder zu.** Ein Schalter, der nur in eine Richtung geht, ist eine
  // Sackgasse. **Zu heißt verborgen**, nicht fort — sonst ginge verloren, was
  // darin eingegeben war (unten).
  await userEvent.click(schalter)

  expect(screen.getByText('Der Inhalt')).not.toBeVisible()
  expect(schalter).toHaveAttribute('aria-expanded', 'false')
})

// **`aria-controls` zeigt auf den Kasten, den man sieht.**
//
// Der Grund steht bei `Kartenwerkzeuge.tsx`: ein `steuert` auf eine Kennung,
// die nicht im Baum steht, hat axe schon einmal angestrichen. Solange der
// Kasten zu ist, sieht ihn niemand — also darf auch das Attribut nicht
// dastehen, weder vor dem ersten Öffnen noch nach dem Zuklappen.
it('nennt nur einen Kasten, den man sieht', async () => {
  render(
    <Aufklappkasten titel="Werkzeug">
      <p>Der Inhalt</p>
    </Aufklappkasten>,
  )

  const schalter = screen.getByRole('button', { name: 'Werkzeug' })
  expect(schalter).not.toHaveAttribute('aria-controls')

  await userEvent.click(schalter)

  const kennung = schalter.getAttribute('aria-controls')
  expect(kennung).not.toBeNull()
  expect(document.getElementById(kennung ?? '')).not.toBeNull()

  await userEvent.click(schalter)
  expect(schalter).not.toHaveAttribute('aria-controls')
})

// **Zuklappen verliert keine Eingabe** — UX-08 S11.NAT-03
// (#1680):
// *„Aufklappen verändert keine Eingaben."* Bis zum 2026-09-25 hing der
// Kasten seinen Inhalt beim Zuklappen aus — und mit ihm den Namen eines
// halb angelegten Decks oder die Kennung eines Gegenübers.
it('behält, was darin eingegeben war, über Zu- und Aufklappen', async () => {
  function Formular() {
    const [name, setzeName] = useState('')

    return (
      <input
        aria-label="Name"
        value={name}
        onChange={(ereignis) => {
          setzeName(ereignis.target.value)
        }}
      />
    )
  }

  render(
    <Aufklappkasten titel="Deck anlegen">
      <Formular />
    </Aufklappkasten>,
  )

  const schalter = screen.getByRole('button', { name: 'Deck anlegen' })
  await userEvent.click(schalter)
  await userEvent.type(screen.getByRole('textbox', { name: 'Name' }), 'Aggro')

  await userEvent.click(schalter)
  expect(screen.queryByRole('textbox', { name: 'Name' })).toBeNull()

  await userEvent.click(schalter)
  expect(screen.getByRole('textbox', { name: 'Name' })).toHaveValue('Aggro')
})

// **Gezeichnet wird beim ersten Öffnen, und einmal** — ein Inhalt, der den
// Bestand fragt (`Gegenueberkasten`), fragt nicht, solange niemand ihn sehen
// will, und nicht bei jedem Aufklappen neu.
it('zeichnet seinen Inhalt beim ersten Öffnen, und dann nicht wieder', async () => {
  const gezeichnet = vi.fn()

  function Inhalt() {
    useEffect(() => {
      gezeichnet()
    }, [])

    return <p>Der Inhalt</p>
  }

  render(
    <Aufklappkasten titel="Werkzeug">
      <Inhalt />
    </Aufklappkasten>,
  )
  expect(gezeichnet).not.toHaveBeenCalled()

  const schalter = screen.getByRole('button', { name: 'Werkzeug' })
  await userEvent.click(schalter)
  await userEvent.click(schalter)
  await userEvent.click(schalter)

  expect(gezeichnet).toHaveBeenCalledTimes(1)
})

// **Ein Kasten, der nur zum Anlegen da ist, öffnet sich einem Leser nicht**
// — G10 (#2276): er zeigte sonst ein Formular, das am Ende nicht
// abzuschicken ist.
it('bleibt für einen Leser zu, wenn er schreibt, und sagt warum', async () => {
  render(
    <Schreibrechtraum nurLesend>
      <Aufklappkasten titel="Neues Paket anlegen" schreibt>
        <p>Das Formular</p>
      </Aufklappkasten>
      <Aufklappkasten titel="Ansicht">
        <p>Zum Lesen</p>
      </Aufklappkasten>
    </Schreibrechtraum>,
  )

  const anlegen = screen.getByRole('button', { name: 'Neues Paket anlegen' })
  expect(anlegen).toHaveAttribute('aria-disabled', 'true')
  expect(anlegen).toHaveAccessibleDescription(
    kataloge.de.bausteine.nurlesen.knopf,
  )
  expect(anlegen).toHaveAttribute('title', kataloge.de.bausteine.nurlesen.knopf)

  await userEvent.click(anlegen)
  await userEvent.click(screen.getByRole('button', { name: 'Ansicht' }))

  // **Der Kasten daneben geht auf** — der Beleg, dass der Klick ankam.
  expect(screen.getByText('Zum Lesen')).toBeVisible()
  expect(anlegen).toHaveAttribute('aria-expanded', 'false')
  expect(screen.queryByText('Das Formular')).toBeNull()
})

// **Ein Wunsch schlägt ihn auf, einmal** — AP-10 (#2367): „Schnittstellen
// einrichten“ aus der Befehlspalette landet am offenen Kasten mit dem Fokus
// auf seinem Schalter. Wer ihn danach schließt, schließt ihn; erst ein neuer
// Wunsch öffnet ihn wieder.
it('schlägt auf einen Wunsch auf, gibt dem Schalter den Fokus und tut es je Wunsch einmal', async () => {
  const kasten = (aufschlagen: string | undefined) => (
    <Aufklappkasten titel="Werkzeug" aufschlagen={aufschlagen}>
      <p>Der Inhalt</p>
    </Aufklappkasten>
  )
  const { rerender } = render(kasten(undefined))
  const schalter = screen.getByRole('button', { name: 'Werkzeug' })
  expect(schalter).toHaveAttribute('aria-expanded', 'false')

  rerender(kasten('werkzeuge#1'))
  expect(schalter).toHaveAttribute('aria-expanded', 'true')
  expect(schalter).toHaveFocus()

  await userEvent.click(schalter)
  rerender(kasten('werkzeuge#1'))
  expect(schalter).toHaveAttribute('aria-expanded', 'false')

  rerender(kasten('werkzeuge#2'))
  expect(schalter).toHaveAttribute('aria-expanded', 'true')
})

it('schlägt auf, wenn er mit einem Wunsch entsteht', () => {
  render(
    <Aufklappkasten titel="Werkzeug" aufschlagen="werkzeuge#1">
      <p>Der Inhalt</p>
    </Aufklappkasten>,
  )

  expect(screen.getByText('Der Inhalt')).toBeVisible()
  expect(screen.getByRole('button', { name: 'Werkzeug' })).toHaveFocus()
})
