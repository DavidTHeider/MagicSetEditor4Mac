// **Was unterwegs ist, und wann es das nicht mehr ist** — UX-03.AK-01
// (#1675).
//
// # Warum eine Bühne und kein `renderHook`
//
// **Weil der Raum zwei Enden hat.** `uebergib` wechselt die Seite, und die
// Seite entscheidet, ob der Anschluss stehen bleibt — das ist eine Aussage
// über die Verbindung und nicht über eine Funktion. Die Bühne baut sie so
// auf, wie `Anwendung` es tut: der Stand der Seite liegt **über** dem Raum.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'

import { Anschlussraum, useAnschluss, useAufgenommen } from './anschlussraum'

import type { Anschluss } from './anschluss'
import type { Seitenschluessel } from './navigation'

const einDruck: Anschluss = {
  art: 'deck',
  name: 'Aschenwächter',
  printing: 'p-4',
}

const einAngebot: Anschluss = {
  art: 'handel',
  name: 'Glutschmiede',
  printing: 'p-9',
  guete: 'nearMint',
  sprache: 'de',
  finish: 'nonfoil',
}

// Buehne stellt den Raum so auf, wie `Anwendung` es tut.
function Buehne({ start = 'sammlung' }: { start?: Seitenschluessel }) {
  const [aktuell, setzeAktuell] = useState<Seitenschluessel>(start)

  return (
    <Anschlussraum aktuell={aktuell} wechseln={setzeAktuell}>
      <p>{aktuell}</p>
      <Anzeige wechseln={setzeAktuell} />
    </Anschlussraum>
  )
}

function Anzeige({
  wechseln,
}: {
  wechseln: (seite: Seitenschluessel) => void
}) {
  const anschluss = useAnschluss()

  return (
    <div>
      <p>{anschluss.offen?.name ?? 'nichts'}</p>

      <button
        type="button"
        onClick={() => {
          anschluss.uebergib(einDruck)
        }}
      >
        ins deck
      </button>
      <button
        type="button"
        onClick={() => {
          anschluss.uebergib(einAngebot)
        }}
      >
        zum tausch
      </button>
      <button
        type="button"
        onClick={() => {
          anschluss.verbrauche()
        }}
      >
        verbrauchen
      </button>
      <button
        type="button"
        onClick={() => {
          wechseln('einstellungen')
        }}
      >
        woanders hin
      </button>
    </div>
  )
}

describe('der Anschlussraum', () => {
  it('nimmt das Objekt mit und geht dorthin, wo es hingehört', async () => {
    render(<Buehne />)

    await userEvent.click(screen.getByRole('button', { name: 'ins deck' }))

    expect(screen.getByText('decks')).toBeInTheDocument()
    expect(screen.getByText('Aschenwächter')).toBeInTheDocument()
  })

  it('führt den Handel auf seine eigene Fläche', async () => {
    render(<Buehne />)

    await userEvent.click(screen.getByRole('button', { name: 'zum tausch' }))

    expect(screen.getByText('handeln')).toBeInTheDocument()
    expect(screen.getByText('Glutschmiede')).toBeInTheDocument()
  })

  // **§7, Regel 1**: ein Wechsel auf eine dritte Fläche räumt den offenen
  // Kontext. Wer es sich anders überlegt, soll den Druck nicht beim nächsten
  // Blick aufs Deck wiederfinden.
  it('räumt, was unterwegs war, sobald jemand woanders hingeht', async () => {
    render(<Buehne />)

    await userEvent.click(screen.getByRole('button', { name: 'ins deck' }))
    await userEvent.click(screen.getByRole('button', { name: 'woanders hin' }))

    expect(screen.getByText('einstellungen')).toBeInTheDocument()
    expect(screen.getByText('nichts')).toBeInTheDocument()
  })

  // **Die Gegenprobe zur Zeile darüber**: derselbe Wechsel, nur auf die
  // Zielfläche — und dort muss er stehen bleiben. Ohne sie sagte der Lauf
  // oben nur, dass irgendein Wechsel räumt.
  it('lässt stehen, was auf seiner eigenen Fläche ankommt', async () => {
    render(<Buehne start="decks" />)

    await userEvent.click(screen.getByRole('button', { name: 'ins deck' }))
    await userEvent.click(screen.getByRole('button', { name: 'ins deck' }))

    expect(screen.getByText('Aschenwächter')).toBeInTheDocument()
  })

  it('räumt, was die Zielfläche aufgenommen hat', async () => {
    render(<Buehne />)

    await userEvent.click(screen.getByRole('button', { name: 'ins deck' }))
    await userEvent.click(screen.getByRole('button', { name: 'verbrauchen' }))

    expect(screen.getByText('nichts')).toBeInTheDocument()
  })

  // **Ohne Raum eine taube Stelle.** Ein Bauteil, das außerhalb geprüft wird,
  // soll nicht daran scheitern — dieselbe Überlegung wie bei `useAblage`.
  it('ist außerhalb des Raums taub', async () => {
    render(<Anzeige wechseln={() => undefined} />)

    await userEvent.click(screen.getByRole('button', { name: 'ins deck' }))
    await userEvent.click(screen.getByRole('button', { name: 'verbrauchen' }))

    expect(screen.getByText('nichts')).toBeInTheDocument()
  })
})

// Aufnehmer zählt, wie oft er aufgenommen hat.
function Aufnehmer() {
  const anschluss = useAnschluss()
  const [mal, setzeMal] = useState<string[]>([])

  useAufgenommen(anschluss.offen, (angekommen) => {
    setzeMal((bisher) => [...bisher, angekommen.name])
  })

  return (
    <div>
      <p>{mal.join(' · ') || 'nichts'}</p>
      <button
        type="button"
        onClick={() => {
          anschluss.uebergib(einDruck)
        }}
      >
        ins deck
      </button>
      <button
        type="button"
        onClick={() => {
          anschluss.uebergib(einAngebot)
        }}
      >
        zum tausch
      </button>
    </div>
  )
}

describe('useAufgenommen', () => {
  it('nimmt jeden Anschluss genau einmal auf', async () => {
    function Zweitbuehne() {
      const [aktuell, setzeAktuell] = useState<Seitenschluessel>('sammlung')

      return (
        <Anschlussraum aktuell={aktuell} wechseln={setzeAktuell}>
          <Aufnehmer />
        </Anschlussraum>
      )
    }

    render(<Zweitbuehne />)

    // **Zweimal derselbe Anschluss** — die Fläche zeichnet dazwischen neu,
    // und aufgenommen wird er trotzdem nur einmal.
    await userEvent.click(screen.getByRole('button', { name: 'ins deck' }))
    await userEvent.click(screen.getByRole('button', { name: 'ins deck' }))

    expect(screen.getByText('Aschenwächter')).toBeInTheDocument()

    // **Ein anderer kommt an und wird aufgenommen** — sonst prüfte der Lauf
    // oben nur, dass gar nichts geschieht.
    await userEvent.click(screen.getByRole('button', { name: 'zum tausch' }))

    expect(screen.getByText('Aschenwächter · Glutschmiede')).toBeInTheDocument()
  })

  it('nimmt nichts auf, wo nichts unterwegs ist', () => {
    render(<Aufnehmer />)

    expect(screen.getByText('nichts')).toBeInTheDocument()
  })
})
