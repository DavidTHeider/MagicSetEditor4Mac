// **Ein Port, der keiner sein kann, wird am Feld abgewiesen** — F03 aus dem
// Review zu J05
// (#2269),
// AGENTS.md §6: Fehler erscheinen am verursachenden Feld.
//
// # Was bis dahin geschah
//
// „8080x" war schon seit F03 (#2245) nicht mehr Port 8080 — aber die
// Abweisung kam **über einen Umweg**: das Fenster machte daraus `-1`, schickte
// es, und `domain.PruefeBindung` antwortete mit einem Satz, der als allgemeine
// Meldung **über dem Formular** stand. Das Feld selbst trug weder
// `aria-invalid` noch eine Verbindung zum Satz, und der Ruf ging hinaus.
//
// # Wogegen hier geprüft wird
//
// Was syntaktisch kein Port ist — nicht nur Ziffern, oder mehr als 65535, oder
// leer —, steht als Befund **am Feld**, und es geht nichts hinaus. Die
// Gegenproben stehen daneben: ein Port im Bereich geht durch (auch `0`, die
// Vorbelegung: „einen freien suchen"), die Meldung nimmt sich zurück, und was
// Go sachlich abweist, bleibt die allgemeine Meldung — das Feld kennt nur die
// Gestalt, nicht die Lage.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Dienstposten } from './Dienstposten'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { laeuft, steht } from '../test/dienststaende'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.technik

// satz ist der Befund am Feld — derselbe Wortlaut wie bei jeder ganzen Zahl
// mit Grenzen; ein eigener Satz je Feld wäre ein Satz mehr in acht Sprachen.
const satz = einsetzen(kataloge.de.allgemein.ganzeZahlBereich, {
  kleinstens: '0',
  groesstens: '65535',
})

function zeige(art: 'rest' | 'mcp' = 'rest') {
  return render(
    <Sprachraum sprache="de">
      <Dienstposten art={art} />
    </Sprachraum>,
  )
}

// startbereit belegt den Stand „steht" und gibt den Spion auf dem Start.
function startbereit() {
  const starten = vi.fn().mockResolvedValue(laeuft)

  belegeApi({
    RESTStand: vi.fn().mockResolvedValue(steht),
    RESTStarten: starten,
  })

  return starten
}

// tippe setzt den Port und gibt das Feld.
async function tippe(roh: string) {
  const port = await screen.findByLabelText(texte.dienstRESTPort)

  await userEvent.clear(port)

  if (roh !== '') {
    await userEvent.type(port, roh)
  }

  return port
}

describe('ein Port, der keiner sein kann', () => {
  it.each([
    '8080x',
    '80.5',
    '1e3',
    '-80',
    '65536',

    // Jenseits von 2⁵³ − 1 gäbe `Number` eine Zahl, die nicht die getippte ist.
    '99999999999999999999',
  ])('steht am Feld und geht nicht hinaus: %j', async (roh) => {
    const starten = startbereit()

    zeige()

    const port = await tippe(roh)

    await userEvent.click(
      screen.getByRole('button', { name: texte.dienstStarten }),
    )

    expect(port).toHaveAttribute('aria-invalid', 'true')
    expect(port).toHaveAccessibleDescription(satz)

    // **Genau eine Meldung, und es ist die am Feld**: eine zweite, allgemeine
    // über dem Formular wäre der Umweg, den es nicht mehr gibt.
    expect(screen.getByRole('alert')).toHaveTextContent(satz)
    expect(starten).not.toHaveBeenCalled()
  })

  // **Ein leeres Feld ist nach dem Klick ein Fehler** — wer absendet, meint
  // einen Wert; die Vorbelegung `0` heißt „einen freien suchen" und ein
  // leeres Feld nicht (§6: der Knopf bleibt bedienbar und sagt, was fehlt).
  it('steht am Feld, wenn der Port leer abgeschickt wird', async () => {
    const starten = startbereit()

    zeige()

    const port = await tippe('')

    await userEvent.click(
      screen.getByRole('button', { name: texte.dienstStarten }),
    )

    expect(port).toHaveAttribute('aria-invalid', 'true')
    expect(port).toHaveAccessibleDescription(satz)
    expect(starten).not.toHaveBeenCalled()
  })

  it('meldet sich schon beim Tippen und nicht erst beim Klick', async () => {
    startbereit()

    zeige()

    const port = await tippe('8080x')

    expect(port).toHaveAttribute('aria-invalid', 'true')
    expect(port).toHaveAccessibleDescription(satz)
  })

  // **Mitten im Ändern ist ein leeres Feld noch kein Fehler** — wer eine Zahl
  // ersetzt, löscht sie zuerst (so hält es auch `Zahlenfeld`).
  it('beschwert sich nicht über ein Feld, das gerade geleert wurde', async () => {
    startbereit()

    zeige()

    const port = await tippe('')

    expect(port).not.toHaveAttribute('aria-invalid', 'true')
  })

  it('nimmt die Meldung zurück, sobald der Port stimmt', async () => {
    const starten = startbereit()

    zeige()

    const port = await tippe('8080x')

    await userEvent.click(
      screen.getByRole('button', { name: texte.dienstStarten }),
    )

    await tippe('8080')

    expect(port).not.toHaveAttribute('aria-invalid', 'true')
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: texte.dienstStarten }),
    )

    expect(starten).toHaveBeenCalledWith('', 8080)
  })

  // **Gegenprobe**: der Bereich ist 0 bis 65535 — die Grenzen gehören dazu,
  // und `0` ist die Vorbelegung. Eine Prüfung, die `0` abwiese, hielte den
  // Knopf in der Grundstellung an.
  it.each([
    ['0', 0],
    ['1', 1],
    ['65535', 65535],
    [' 8080 ', 8080],
  ])('lässt %j durch', async (roh, erwartet) => {
    const starten = startbereit()

    zeige()

    const port = await tippe(roh)

    await userEvent.click(
      screen.getByRole('button', { name: texte.dienstStarten }),
    )

    expect(port).not.toHaveAttribute('aria-invalid', 'true')
    expect(starten).toHaveBeenCalledWith('', erwartet)
  })

  it('lässt die Vorbelegung ungeändert durch', async () => {
    const starten = startbereit()

    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.dienstStarten }),
    )

    expect(starten).toHaveBeenCalledWith('', 0)
  })

  // **Der zweite Dienst geht denselben Weg** — dasselbe Bauteil, dieselbe
  // Prüfung; der Fall steht da, damit eine spätere Weiche ihn nicht trennt.
  it('gilt für den zweiten Dienst ebenso', async () => {
    const starten = vi.fn().mockResolvedValue(laeuft)

    belegeApi({
      MCPStand: vi.fn().mockResolvedValue({
        ...steht,
        stand: { ...steht.stand, art: 'mcp' },
      }),
      MCPStarten: starten,
    })

    zeige('mcp')

    const port = await screen.findByLabelText(texte.dienstMCPPort)

    await userEvent.clear(port)
    await userEvent.type(port, '8080x')
    await userEvent.click(
      screen.getByRole('button', { name: texte.dienstStarten }),
    )

    expect(port).toHaveAttribute('aria-invalid', 'true')
    expect(port).toHaveAccessibleDescription(satz)
    expect(starten).not.toHaveBeenCalled()
  })

  // **Was Go sachlich abweist, bleibt die allgemeine Meldung.** Das Feld
  // kennt die Gestalt einer Zahl, nicht die Lage des Rechners: ein Port, der
  // belegt ist, ist ein gültiger Port.
  it('lässt die Absage des Dienstes eine allgemeine Meldung bleiben', async () => {
    const starten = vi.fn().mockResolvedValue({
      ...steht,
      fehler: {
        code: 'validation',
        meldung: 'der Port ist belegt',
        schluessel: '',
      },
    })

    belegeApi({
      RESTStand: vi.fn().mockResolvedValue(steht),
      RESTStarten: starten,
    })

    zeige()

    const port = await tippe('80')

    await userEvent.click(
      screen.getByRole('button', { name: texte.dienstStarten }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'der Port ist belegt',
    )
    expect(port).not.toHaveAttribute('aria-invalid', 'true')
    expect(starten).toHaveBeenCalledWith('', 80)
  })
})
