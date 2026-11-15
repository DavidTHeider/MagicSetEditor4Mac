// **Die Zugriffstoken eines Profils im Fenster** — Werkstück 10 von
// Phase 4.12
// (#1351),
// Punkt 13.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Zugriffstoken } from './Zugriffstoken'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { alsSatz } from '../test/auszeichnung'

import type { TokenDTO } from '../api'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.nutzer

// einToken ist die Vorlage, aus der die Fälle ihre Abweichung bauen.
function einToken(teile: Partial<TokenDTO> = {}): TokenDTO {
  return {
    id: '1',
    profilID: '10',
    bezeichnung: 'Nachtlauf',
    principal: 'Rechner',
    scope: 'lesen',
    ablauf: '',
    angelegt: '2026-09-05T12:00:00Z',
    abgelaufen: false,
    letzteNutzung: '',
    vorgaengerBis: '',
    partie: '',
    sitz: '',
    fassung: '1',
    ...teile,
  }
}

function zeige(teile: Record<string, unknown> = {}) {
  belegeApi({
    Token: vi.fn(() => Promise.resolve({ token: [einToken()] })),
    Tokenablaufwahlen: vi.fn(() =>
      Promise.resolve({
        wahlen: [
          { name: 'tag', ablauf: '2026-09-09T12:00:00Z' },
          { name: 'monat', ablauf: '2026-10-08T12:00:00Z' },
          { name: 'jahr', ablauf: '2027-09-08T12:00:00Z' },
          { name: 'nie', ablauf: '' },
        ],
        fehler: null,
      }),
    ),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Zugriffstoken profilID="10" />
    </Sprachraum>,
  )
}

describe('Die Zugriffstoken eines Profils', () => {
  it('zeigt, was da ist', async () => {
    zeige()

    expect(await screen.findByText('Nachtlauf')).toBeInTheDocument()
    expect(screen.getByText(eigene.tokenUnbefristet)).toBeInTheDocument()
  })

  it('sagt vor dem Anlegen, dass es den Klartext nur einmal gibt', async () => {
    // **Der Satz steht VOR dem Knopf und nicht nach ihm.** Wer erst hinterher
    // erfährt, dass es keinen zweiten Blick gibt, hat ihn schon weggeklickt.
    zeige()

    // **Über `ohneAuszeichnung` und nicht über den rohen Katalogtext.**
    //
    // **`Satz` löst `**genau einmal**` in ein eigenes Element auf**, und ein
    // Vergleich mit dem Katalogtext fände ihn danach nicht mehr — bis zum
    // 2026-09-06 stand hier der rohe Satz, und der Lauf sicherte damit zu,
    // dass die Sternchen **stehen bleiben**.
    expect(
      await screen.findByText(alsSatz(eigene.tokenEinmalig)),
    ).toBeInTheDocument()
  })

  it('zeigt den Klartext nach dem Anlegen — und erst dann', async () => {
    const anlegen = vi.fn(() =>
      Promise.resolve({
        token: einToken({ id: '2' }),
        klartext: 'mse_2_GEHEIM',
      }),
    )

    zeige({ TokenAnlegen: anlegen })

    // **Vorher steht er nirgends.** Ohne diesen Fall wäre die Zusicherung
    // darunter auch dann grün, wenn er die ganze Zeit dastünde (§10).
    expect(await screen.findByText('Nachtlauf')).toBeInTheDocument()
    expect(screen.queryByText('mse_2_GEHEIM')).not.toBeInTheDocument()

    await userEvent.type(
      screen.getByLabelText(eigene.tokenBezeichnung),
      'Bogenbau',
    )
    await userEvent.type(
      screen.getByLabelText(eigene.tokenPrincipal),
      'Baurechner',
    )
    await userEvent.click(
      screen.getByRole('button', { name: eigene.tokenAnlegen }),
    )

    expect(await screen.findByText('mse_2_GEHEIM')).toBeInTheDocument()

    expect(anlegen).toHaveBeenCalledWith(
      expect.objectContaining({
        profilID: '10',
        bezeichnung: 'Bogenbau',
        principal: 'Baurechner',
        scope: 'lesen',
      }),
    )
  })

  it('legt eines an, das schreiben darf und abläuft', async () => {
    // **Der Fall, ohne den das Formular nur seine Vorgaben belegt**: Erlaubnis
    // und Ablauf sind die zwei Angaben, die ein Token von einem anderen
    // unterscheiden — beide gehen hier hinein und müssen unten ankommen.
    const anlegen = vi.fn(() =>
      Promise.resolve({ token: einToken({ id: '2' }), klartext: 'mse_2_X' }),
    )

    zeige({ TokenAnlegen: anlegen })

    await userEvent.type(
      await screen.findByLabelText(eigene.tokenBezeichnung),
      'Setzer',
    )
    await userEvent.type(screen.getByLabelText(eigene.tokenPrincipal), 'Cron')
    await userEvent.selectOptions(
      screen.getByLabelText(eigene.tokenScope),
      'schreiben',
    )

    // **Ein eigener Zeitpunkt gibt das freie Feld frei** — Werkstück O2
    // (#1426); die Vorschläge stehen davor.
    await userEvent.selectOptions(
      screen.getByLabelText(eigene.tokenAblaufwahl),
      'eigen',
    )
    await userEvent.type(
      screen.getByLabelText(eigene.tokenAblauf),
      '2027-01-01T00:00:00Z',
    )
    await userEvent.click(
      screen.getByRole('button', { name: eigene.tokenAnlegen }),
    )

    expect(anlegen).toHaveBeenCalledWith(
      expect.objectContaining({
        bezeichnung: 'Setzer',
        scope: 'schreiben',
        ablauf: '2027-01-01T00:00:00Z',
      }),
    )
  })

  it('legt ohne Bezeichnung gar nicht erst an', async () => {
    const anlegen = vi.fn()

    zeige({ TokenAnlegen: anlegen })

    expect(
      await screen.findByRole('button', { name: eigene.tokenAnlegen }),
    ).toBeDisabled()

    // **Der Beleg, dass der Knopf danach wirklich trägt** — ohne ihn stünde
    // dieser Fall auch dann grün da, wenn er dauerhaft angehalten wäre.
    await userEvent.type(
      screen.getByLabelText(eigene.tokenBezeichnung),
      'Bogenbau',
    )

    // **Und ohne Principal auch nicht** — Werkstück O2 (#1426): ein Token
    // nennt, für wen es arbeitet.
    expect(
      screen.getByRole('button', { name: eigene.tokenAnlegen }),
    ).toBeDisabled()

    await userEvent.type(screen.getByLabelText(eigene.tokenPrincipal), 'Cron')

    expect(
      screen.getByRole('button', { name: eigene.tokenAnlegen }),
    ).toBeEnabled()

    expect(anlegen).not.toHaveBeenCalled()
  })

  it('setzt die Erlaubnis um und lädt danach neu', async () => {
    const aendern = vi.fn(() =>
      Promise.resolve({
        token: einToken({ scope: 'schreiben', fassung: '2' }),
      }),
    )
    const lesen = vi.fn(() => Promise.resolve({ token: [einToken()] }))

    zeige({ Token: lesen, TokenAendern: aendern })

    await userEvent.selectOptions(
      await screen.findByRole('combobox', {
        name: `${eigene.tokenScope} — Nachtlauf`,
      }),
      'schreiben',
    )

    expect(aendern).toHaveBeenCalledWith(
      expect.objectContaining({ id: '1', scope: 'schreiben', fassung: '1' }),
    )

    // **Nachladen ist Bestenfalls-Arbeit** (§7, Regel 3), aber es findet
    // statt: der Stand nach dem Schreiben ist der des Bestands.
    await vi.waitFor(() => {
      expect(lesen).toHaveBeenCalledTimes(2)
    })
  })

  // **Erst nach der Frage** (UX-10.AK-04, #1682): wer den Token benutzt,
  // kommt danach nicht mehr herein. Bis zum 2026-09-25 genügte ein Klick —
  // und die einzige Frage der Zeile galt dem Erneuern.
  it('zieht eines zurück, wenn die Frage bejaht ist', async () => {
    const loeschen = vi.fn(() => Promise.resolve({}))

    zeige({ TokenLoeschen: loeschen })

    await userEvent.click(
      await screen.findByRole('button', { name: eigene.tokenLoeschen }),
    )
    expect(loeschen).not.toHaveBeenCalled()

    await userEvent.click(
      within(
        screen.getByRole('group', {
          name: eigene.tokenLoeschenFrage.replace('{name}', 'Nachtlauf'),
        }),
      ).getByRole('button', { name: eigene.tokenLoeschenJa }),
    )

    expect(loeschen).toHaveBeenCalledWith('1')
  })

  it('sagt, wenn eines abgelaufen ist', async () => {
    // **Die Anwendung rechnet das nicht selbst aus** — `abgelaufen` kommt aus
    // Go. Vier Flächen, die dieselbe Zeitspanne rechnen, laufen an der
    // Zeitzone auseinander.
    zeige({
      Token: vi.fn(() =>
        Promise.resolve({
          token: [
            einToken({ ablauf: '2026-01-01T00:00:00Z', abgelaufen: true }),
          ],
        }),
      ),
    })

    expect(await screen.findByText(eigene.tokenAbgelaufen)).toBeInTheDocument()
  })

  it('nennt den Ablauf, solange er nicht überschritten ist', async () => {
    // **Drei Zustände und nicht zwei**: abgelaufen, unbefristet, und „läuft
    // ab am". Ohne diesen Fall stünde der dritte ungeprüft da.
    zeige({
      Token: vi.fn(() =>
        Promise.resolve({
          token: [einToken({ ablauf: '2027-01-01T00:00:00Z' })],
        }),
      ),
    })

    expect(await screen.findByText('2027-01-01T00:00:00Z')).toBeInTheDocument()
  })

  it('meldet auch, was beim Lesen schiefgeht', async () => {
    // **Der andere der beiden Fehlerwege.** Einer kommt vom Schreiben, einer
    // vom Laden — und die Meldung steht an derselben Stelle. Ohne diesen Fall
    // wäre der zweite Zweig unbetreten (§3).
    zeige({
      Token: vi.fn(() =>
        Promise.resolve({
          fehler: { code: 'konflikt', schluessel: 'konflikt' },
        }),
      ),
    })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.konflikt,
    )
  })

  it('erklärt den leeren Zustand statt „keine Daten" zu zeigen', async () => {
    zeige({ Token: vi.fn(() => Promise.resolve({ token: [] })) })

    expect(await screen.findByText(eigene.tokenLeer)).toBeInTheDocument()
  })

  it('meldet, was schiefgeht', async () => {
    zeige({
      TokenLoeschen: vi.fn(() =>
        Promise.resolve({
          fehler: { code: 'konflikt', schluessel: 'konflikt' },
        }),
      ),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: eigene.tokenLoeschen }),
    )
    await userEvent.click(
      within(
        screen.getByRole('group', {
          name: eigene.tokenLoeschenFrage.replace('{name}', 'Nachtlauf'),
        }),
      ).getByRole('button', { name: eigene.tokenLoeschenJa }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.konflikt,
    )
  })
})
