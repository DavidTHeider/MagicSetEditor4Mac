// **Passt die neue Fassung zu diesem Bestand?** — Werkstück O6 von Phase 5.2
// (#1430).
//
// **Am Bauteil und nicht am Band**: es stellt dar, was das Band geholt hat.
// Dass der Ladeknopf bei einer unverträglichen Fassung wegbleibt, prüft
// `fassungsband.test.tsx` — dort gehört er hin.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Vertraeglichkeit } from './Vertraeglichkeit'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

import type { VertraeglichkeitAntwort } from '../apivertraeglichkeit'
import { zeichnetNichts } from '../test/leererbaum'

const texte = kataloge.de.allgemein.vertraeglichkeit

function urteil(
  teil: Partial<VertraeglichkeitAntwort> = {},
): VertraeglichkeitAntwort {
  return {
    fassung: '0.6.0',
    urteil: 'vertraeglich',
    bereit: false,
    hindernisse: [],
    fehler: null,
    ...teil,
  }
}

function zeige(
  gezeigt: VertraeglichkeitAntwort | undefined = urteil(),
  aufVorbereiten = vi.fn(),
  fehler?: string,
) {
  render(
    <Sprachraum sprache="de">
      <Vertraeglichkeit
        urteil={gezeigt}
        aufVorbereiten={aufVorbereiten}
        fehler={fehler}
      />
    </Sprachraum>,
  )
}

describe('Die Verträglichkeit', () => {
  it('sagt, dass die Fassung den Bestand öffnen kann', () => {
    zeige()

    expect(screen.getByText(texte.vertraeglich)).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: texte.vorbereiten }),
    ).toBeInTheDocument()
  })

  // **Jedes Hindernis wird genannt, mit seiner Stelle** — wer nur das erste
  // behebt, soll nicht überrascht werden.
  it('nennt jedes Hindernis samt Stelle', () => {
    zeige(
      urteil({
        urteil: 'unvertraeglich',
        hindernisse: [
          { art: 'wanderungFehlt', stelle: '0052_deckzuordnungen.sql' },
          { art: 'plattform', stelle: 'linux/amd64' },
        ],
      }),
    )

    expect(screen.getByText(texte.unvertraeglich)).toBeInTheDocument()
    expect(
      screen.getByText(
        `${texte.hindernisse.wanderungFehlt} (0052_deckzuordnungen.sql)`,
      ),
    ).toBeInTheDocument()
    expect(
      screen.getByText(`${texte.hindernisse.plattform} (linux/amd64)`),
    ).toBeInTheDocument()

    // **Und kein Knopf, der sicher nein sagt.**
    expect(
      screen.queryByRole('button', { name: texte.vorbereiten }),
    ).not.toBeInTheDocument()
  })

  // **Ungeklärt ist kein Nein** — der Weg von Hand steht daneben.
  it('führt bei einer Freigabe ohne Manifest in den Weg von Hand', () => {
    zeige(
      urteil({
        urteil: 'ungeklaert',
        hindernisse: [{ art: 'keinManifest', stelle: '' }],
      }),
    )

    expect(screen.getByText(texte.ungeklaert)).toBeInTheDocument()
    expect(screen.getByText(texte.hindernisse.keinManifest)).toBeInTheDocument()
  })

  it('meldet den Wunsch, den Wechsel vorzubereiten', async () => {
    const aufVorbereiten = vi.fn()

    zeige(urteil(), aufVorbereiten)

    await userEvent.click(
      screen.getByRole('button', { name: texte.vorbereiten }),
    )

    expect(aufVorbereiten).toHaveBeenCalledTimes(1)
  })

  // **Und ein vorbereiteter Wechsel sagt es** — samt dem Knopf, der weg ist.
  it('sagt, wenn der Wechsel vorbereitet ist', () => {
    zeige(urteil({ bereit: true }))

    expect(screen.getByText(texte.bereit)).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: texte.vorbereiten }),
    ).not.toBeInTheDocument()
  })

  it('zeigt einen Fehler als Meldung', () => {
    zeige(undefined, vi.fn(), 'Es ist ein unerwarteter Fehler aufgetreten.')

    expect(
      screen.getByText('Es ist ein unerwarteter Fehler aufgetreten.'),
    ).toBeInTheDocument()
  })

  // **Ohne Urteil steht nichts da** — auch nicht, solange noch geholt wird.
  it('zeigt ohne Urteil nichts', () => {
    const { container } = render(
      <Sprachraum sprache="de">
        <Vertraeglichkeit
          urteil={undefined}
          aufVorbereiten={vi.fn()}
          fehler={undefined}
        />
      </Sprachraum>,
    )

    expect(zeichnetNichts(container)).toBe(true)
  })

  // **Und ein leeres Urteil auch nicht**: es heißt „es gibt keine neuere
  // Fassung", und darüber ist nichts zu sagen.
  it('zeigt bei leerem Urteil nichts', () => {
    const { container } = render(
      <Sprachraum sprache="de">
        <Vertraeglichkeit
          urteil={urteil({ urteil: '' })}
          aufVorbereiten={vi.fn()}
          fehler={undefined}
        />
      </Sprachraum>,
    )

    expect(zeichnetNichts(container)).toBe(true)
  })

  // **Eine unbekannte Marke steht im Wortlaut da** — sie ist immer noch eine
  // Auskunft.
  it('zeigt eine unbekannte Marke im Wortlaut', () => {
    zeige(
      urteil({
        urteil: 'irgendwas',
        hindernisse: [{ art: 'nochwas', stelle: '' }],
      }),
    )

    expect(screen.getByText('irgendwas')).toBeInTheDocument()
    expect(screen.getByText('nochwas')).toBeInTheDocument()
  })
})
