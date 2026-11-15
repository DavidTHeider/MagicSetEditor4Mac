// **Der Kasten, mit dem ein Dienst geschaltet wird** — Werkstück N6 von
// Phase 5.4
// (#1572).
//
// # Wonach hier gefragt wird
//
// Danach, dass **nichts von selbst startet**, dass die Adresse aus dem Feld
// wirklich ankommt, dass ein laufender Dienst seinen **Zugang** zeigt — er
// entsteht beim Starten und stünde sonst nirgends — und dass der Satz über
// die Verschlüsselung sagt, was gilt.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Dienstposten } from './Dienstposten'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { laeuft, steht } from '../test/dienststaende'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.technik

function zeige(art: 'rest' | 'mcp' = 'rest') {
  return render(
    <Sprachraum sprache="de">
      <Dienstposten art={art} />
    </Sprachraum>,
  )
}

describe('solange nichts läuft', () => {
  it('zeigt die Felder und startet von selbst nichts', async () => {
    const starten = vi.fn()
    belegeApi({
      RESTStand: vi.fn().mockResolvedValue(steht),
      RESTStarten: starten,
    })

    zeige()

    expect(
      await screen.findByRole('button', { name: texte.dienstStarten }),
    ).toBeInTheDocument()

    // **Das ist die Zusage aus 0115**: eine sichtbare Zustimmung je Öffnung.
    expect(starten).not.toHaveBeenCalled()
  })

  it('reicht Adresse und Port durch, wie sie dastehen', async () => {
    const starten = vi.fn().mockResolvedValue(laeuft)
    belegeApi({
      RESTStand: vi.fn().mockResolvedValue(steht),
      RESTStarten: starten,
    })

    zeige()

    const adresse = await screen.findByLabelText(texte.dienstRESTAdresse)
    await userEvent.clear(adresse)
    await userEvent.type(adresse, '0.0.0.0')

    const port = screen.getByLabelText(texte.dienstRESTPort)
    await userEvent.clear(port)
    await userEvent.type(port, '8080')

    await userEvent.click(
      screen.getByRole('button', { name: texte.dienstStarten }),
    )

    expect(starten).toHaveBeenCalledWith('0.0.0.0', 8080)
  })

  // **Was kein Port sein kann, steht am Feld** und geht nicht hinaus (F03,
  // #2269) — ein leeres Feld, „8080x", „80.5", „1e3", „-80" und alles über
  // 65535. Vorher wurde daraus `-1`, abgeschickt, und die Absage kam als
  // allgemeine Meldung über dem Formular. Die Fälle stehen in
  // `dienstpostenport.test.tsx`.
})

describe('während er läuft', () => {
  it('zeigt Adresse, Zugang und die erreichbaren Namen', async () => {
    belegeApi({ RESTStand: vi.fn().mockResolvedValue(laeuft) })

    zeige()

    expect(await screen.findByText('https://[::]:54591')).toBeInTheDocument()

    // **Der Zugang steht da, weil er sonst nirgends stünde**: er entsteht
    // beim Starten und stirbt mit dem Posten.
    expect(screen.getByText('die-marke')).toBeInTheDocument()

    expect(screen.getByText('192.168.178.51:54591')).toBeInTheDocument()
    expect(screen.getByText('mein.mac')).toBeInTheDocument()
    expect(screen.getByText(texte.dienstMitTLS)).toBeInTheDocument()
  })

  it('sagt bei einer Bindung auf der Schleife, dass Klartext hier reicht', async () => {
    belegeApi({
      RESTStand: vi.fn().mockResolvedValue({
        ...laeuft,
        stand: {
          ...laeuft.stand,
          tls: false,
          adresse: 'http://127.0.0.1:54742',
          adressen: [],
          namen: [],
        },
      }),
    })

    zeige()

    expect(await screen.findByText(texte.dienstOhneTLS)).toBeInTheDocument()
    expect(screen.queryByText(texte.dienstMitTLS)).not.toBeInTheDocument()
    expect(screen.queryByText(texte.dienstErreichbar)).not.toBeInTheDocument()
  })

  it('beendet ihn auf Klick', async () => {
    const beenden = vi.fn().mockResolvedValue(steht)
    belegeApi({
      RESTStand: vi.fn().mockResolvedValue(laeuft),
      RESTBeenden: beenden,
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.dienstBeenden }),
    )

    expect(beenden).toHaveBeenCalled()
  })

  it('zeigt den Fingerabdruck und kopiert ihn', async () => {
    // **Weil der Satz daneben dazu auffordert**, ihn zu vergleichen: das
    // Zeugnis ist selbst ausgestellt, und ein Gegenüber hat sonst nichts.
    const schreiben = vi.fn().mockResolvedValue(undefined)
    Object.assign(navigator, { clipboard: { writeText: schreiben } })

    belegeApi({ RESTStand: vi.fn().mockResolvedValue(laeuft) })

    zeige()

    expect(await screen.findByText('4e0b-621d')).toBeInTheDocument()

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.dienstFingerabdruckKopieren,
      }),
    )

    expect(schreiben).toHaveBeenCalledWith('4e0b-621d')
    expect(
      await screen.findByRole('button', { name: texte.dienstKopiert }),
    ).toBeInTheDocument()

    // **Und der Zugangsknopf bleibt, was er war** — zwei Knöpfe, ein
    // Merker, und ohne diese Zeile stünde zweimal „Kopiert".
    expect(
      screen.getByRole('button', { name: texte.dienstMarkeKopieren }),
    ).toBeInTheDocument()
  })

  it('zeigt keinen Fingerabdruck, wo keiner gilt', async () => {
    // **Die Gegenprobe**: auf der Schleife gibt es kein Zeugnis, und eine
    // leere Zeile „Fingerabdruck:" sähe aus wie ein verlorener.
    belegeApi({
      RESTStand: vi.fn().mockResolvedValue({
        stand: {
          ...laeuft.stand,
          adresse: 'http://127.0.0.1:8123',
          tls: false,
          fingerabdruck: '',
        },
        fehler: null,
      }),
    })

    zeige()

    expect(await screen.findByText('die-marke')).toBeInTheDocument()
    expect(
      screen.queryByText(`${texte.dienstFingerabdruck}: `),
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', {
        name: texte.dienstFingerabdruckKopieren,
      }),
    ).not.toBeInTheDocument()
  })

  it('merkt sich nichts, wenn das Kopieren scheitert', async () => {
    // **Weil „Kopiert" sonst über einer leeren Zwischenablage stünde.**
    Object.assign(navigator, {
      clipboard: { writeText: vi.fn().mockRejectedValue(new Error('nein')) },
    })

    belegeApi({ RESTStand: vi.fn().mockResolvedValue(laeuft) })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.dienstFingerabdruckKopieren,
      }),
    )

    expect(
      await screen.findByRole('button', {
        name: texte.dienstFingerabdruckKopieren,
      }),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: texte.dienstKopiert }),
    ).not.toBeInTheDocument()
  })

  it('kopiert den Zugang', async () => {
    const schreiben = vi.fn().mockResolvedValue(undefined)
    Object.assign(navigator, { clipboard: { writeText: schreiben } })

    belegeApi({ RESTStand: vi.fn().mockResolvedValue(laeuft) })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.dienstMarkeKopieren }),
    )

    expect(schreiben).toHaveBeenCalledWith('die-marke')
    expect(
      await screen.findByRole('button', { name: texte.dienstKopiert }),
    ).toBeInTheDocument()
  })
})

it('meldet einen Fehler und bleibt stehen', async () => {
  belegeApi({
    RESTStand: vi.fn().mockResolvedValue(steht),
    RESTStarten: vi.fn().mockResolvedValue({
      ...steht,
      fehler: {
        code: 'validation',
        meldung: 'localhost ist keine Adresse',
        schluessel: '',
      },
    }),
  })

  zeige()

  await userEvent.click(
    await screen.findByRole('button', { name: texte.dienstStarten }),
  )

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'localhost ist keine Adresse',
  )

  // **Die Felder bleiben** — wer eine Adresse tippt, die abgewiesen wird,
  // soll sie berichtigen können und nicht neu tippen müssen (§7, Regel 3).
  expect(screen.getByLabelText(texte.dienstRESTAdresse)).toBeInTheDocument()
})

it('bedient auch den zweiten Dienst', async () => {
  const starten = vi.fn().mockResolvedValue({
    ...laeuft,
    stand: { ...laeuft.stand, art: 'mcp' },
  })

  belegeApi({
    MCPStand: vi.fn().mockResolvedValue({
      ...steht,
      stand: { ...steht.stand, art: 'mcp' },
    }),
    MCPStarten: starten,
  })

  zeige('mcp')

  expect(await screen.findByText(texte.dienstMCPTitel)).toBeInTheDocument()

  await userEvent.click(
    screen.getByRole('button', { name: texte.dienstStarten }),
  )

  expect(starten).toHaveBeenCalled()
})

it('meldet einen Lesefehler des Stands', async () => {
  belegeApi({
    RESTStand: vi.fn().mockResolvedValue({
      ...steht,
      fehler: {
        code: 'unbekannt',
        meldung: 'die Platte schweigt',
        schluessel: '',
      },
    }),
  })

  zeige()

  expect(await screen.findByText(texte.dienstRESTTitel)).toBeInTheDocument()
})

it('nimmt einen getippten Port als Zahl — auch den, der nicht null ist', async () => {
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
  await userEvent.type(port, '9000')

  await userEvent.click(
    screen.getByRole('button', { name: texte.dienstStarten }),
  )

  // **Der zweite Dienst nimmt denselben Weg** — dieselbe Zeile, derselbe
  // Zweig. Ohne diesen Fall bliebe die MCP-Hälfte der Weiche unbetreten.
  expect(starten).toHaveBeenCalledWith('', 9000)
})

it('beendet auch den zweiten Dienst', async () => {
  const beenden = vi.fn().mockResolvedValue(steht)
  belegeApi({
    MCPStand: vi.fn().mockResolvedValue({
      ...laeuft,
      stand: { ...laeuft.stand, art: 'mcp' },
    }),
    MCPBeenden: beenden,
  })

  zeige('mcp')

  await userEvent.click(
    await screen.findByRole('button', { name: texte.dienstBeenden }),
  )

  expect(beenden).toHaveBeenCalled()
})
