// **Die Schublade** — UX-08.AK-01 und S06.NAT-02
// (#1680,
// #1686):
// ein Seitenbereich übernimmt den Fokus und gibt ihn beim Schließen an den
// Auslöser zurück.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'

import { Schublade } from './Schublade'
import { Sprachraum } from '../sprache/kontext'
import { effekteNachholen } from '../test/effekte'

function Probe() {
  const [offen, setzeOffen] = useState(false)

  return (
    <>
      <button
        type="button"
        onClick={() => {
          setzeOffen(true)
        }}
      >
        Eigenschaften
      </button>

      {/* **Ein Feld draußen** — es steht für die Befehlspalette, die über
          einer offenen Schublade aufgehen kann. */}
      <input type="text" aria-label="Palette" />

      {offen && (
        <Schublade
          titel="Eigenschaften des Sets"
          schliessenBezeichnung="Schließen"
          schliessen={() => {
            setzeOffen(false)
          }}
        >
          <input type="text" aria-label="Stilbogen" />

          {/* **Ein Feld, das Escape selbst braucht** — wie die Suchleiste der
              Konsole: es leert sich und sagt es mit `preventDefault`. */}
          <input
            type="text"
            aria-label="Suche"
            onKeyDown={(ereignis) => {
              if (ereignis.key === 'Escape') {
                ereignis.preventDefault()
              }
            }}
          />
        </Schublade>
      )}
    </>
  )
}

function zeichnen() {
  render(
    <Sprachraum sprache="de">
      <Probe />
    </Sprachraum>,
  )
}

describe('Schublade', () => {
  // **Sie schiebt sich herein, statt zu springen** — AP-4 (#2361), B-11:
  // Deckkraft und Lage von rechts in `--dauer-mittel` (`einschieben`), der
  // Schleier taucht auf.
  it('schiebt sich in der mittleren Dauer herein', async () => {
    zeichnen()

    await userEvent.click(screen.getByRole('button', { name: 'Eigenschaften' }))

    expect(
      screen.getByRole('dialog', { name: 'Eigenschaften des Sets' }),
    ).toHaveClass('einschieben')
    expect(document.querySelector('.mantine-Drawer-overlay')).toHaveClass(
      'auftauchen',
    )
  })

  it('ist ein benannter, modaler Bereich und nimmt den Fokus', async () => {
    zeichnen()

    await userEvent.click(screen.getByRole('button', { name: 'Eigenschaften' }))

    const schublade = screen.getByRole('dialog', {
      name: 'Eigenschaften des Sets',
    })
    expect(schublade).toHaveAttribute('aria-modal', 'true')

    // **Der Fokus ist drin** — auf dem ersten Bedienelement, nicht irgendwo
    // dahinter.
    await waitFor(() => {
      expect(schublade).toContainElement(document.activeElement as HTMLElement)
    })

    expect(
      screen.getByRole('heading', { name: 'Eigenschaften des Sets' }),
    ).toBeInTheDocument()
  })

  // **Der Inhalt ist keine Beschreibung.** Mantines `Drawer.Body` hängt sich
  // über `aria-describedby` an den Kasten (`ModalBaseContent.mjs`, Zeile 37,
  // Mantine 9.6.1), und dann liest VoiceOver beim Öffnen die **ganze**
  // Stilseite als Beschreibung vor, bevor man eine Taste drücken kann.
  // Genannt wird die Schublade über ihren Titel — und nur über ihn.
  it('trägt den Titel als Namen und keinen Inhalt als Beschreibung', async () => {
    zeichnen()

    await userEvent.click(screen.getByRole('button', { name: 'Eigenschaften' }))

    const schublade = screen.getByRole('dialog', {
      name: 'Eigenschaften des Sets',
    })
    expect(schublade).not.toHaveAttribute('aria-describedby')
    expect(schublade).toHaveAccessibleDescription('')
  })

  it('schließt mit Escape und gibt den Fokus an den Auslöser zurück', async () => {
    zeichnen()

    const ausloeser = screen.getByRole('button', { name: 'Eigenschaften' })
    await userEvent.click(ausloeser)
    await effekteNachholen()
    await userEvent.keyboard('{Escape}')

    await waitFor(() => {
      expect(screen.queryByRole('dialog')).toBeNull()
    })

    await waitFor(() => {
      expect(ausloeser).toHaveFocus()
    })
  })

  // **Escape schließt das Innerste zuerst.** Wer darin die Taste für sich
  // braucht — die Suchleiste der Konsole, eine Rückfrage, ein Menü —,
  // verbraucht sie, und die Schublade bleibt offen. Mantine schlösse sie in
  // der Capture-Phase, **bevor** das innere Element die Taste überhaupt sieht
  // (`use-modal.mjs`, Zeile 15, Mantine 9.6.1).
  it('bleibt offen, wenn etwas darin das Escape verbraucht', async () => {
    zeichnen()

    await userEvent.click(screen.getByRole('button', { name: 'Eigenschaften' }))
    await effekteNachholen()

    screen.getByLabelText('Suche').focus()
    await userEvent.keyboard('{Escape}')

    // **Nachgeholt, bevor verneint wird** — die Schublade schließt eine
    // Aufgabe später, und ein Schließen, das noch aussteht, sähe hier wie
    // „offen" aus.
    await effekteNachholen()
    expect(screen.getByRole('dialog')).toBeInTheDocument()

    // **Der Beleg, dass jemand zuhört** (§3): dasselbe Escape von einem Feld,
    // das es nicht braucht, schließt sie.
    screen.getByLabelText('Stilbogen').focus()
    await userEvent.keyboard('{Escape}')

    await waitFor(() => {
      expect(screen.queryByRole('dialog')).toBeNull()
    })
  })

  // **Und nur, wenn der Fokus in ihr steht.** Geht die Befehlspalette über
  // ihr auf, gehört ein Escape der Palette — sonst gingen beide mit einem
  // Druck zu.
  it('bleibt offen, solange der Fokus woanders steht', async () => {
    zeichnen()

    await userEvent.click(screen.getByRole('button', { name: 'Eigenschaften' }))
    await effekteNachholen()

    screen.getByLabelText('Palette').focus()
    await userEvent.keyboard('{Escape}')

    await effekteNachholen()
    expect(screen.getByRole('dialog')).toBeInTheDocument()

    // **Der Beleg, dass jemand zuhört**: mit dem Fokus darin schließt sie.
    screen.getByLabelText('Stilbogen').focus()
    await userEvent.keyboard('{Escape}')

    await waitFor(() => {
      expect(screen.queryByRole('dialog')).toBeNull()
    })
  })

  it('schließt über den benannten Knopf im Kopf', async () => {
    zeichnen()

    const ausloeser = screen.getByRole('button', { name: 'Eigenschaften' })
    await userEvent.click(ausloeser)
    await userEvent.click(screen.getByRole('button', { name: 'Schließen' }))

    expect(screen.queryByRole('dialog')).toBeNull()

    await waitFor(() => {
      expect(ausloeser).toHaveFocus()
    })
  })

  it('hält den Fokus darin, wenn Tab über das letzte Element hinausgeht', async () => {
    zeichnen()

    await userEvent.click(screen.getByRole('button', { name: 'Eigenschaften' }))

    const schublade = screen.getByRole('dialog')

    for (let schritt = 0; schritt < 5; schritt++) {
      await userEvent.tab()

      expect(schublade).toContainElement(document.activeElement as HTMLElement)
    }
  })
})
