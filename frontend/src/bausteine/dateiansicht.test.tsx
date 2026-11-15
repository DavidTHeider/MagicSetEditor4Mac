// **Eine Datei ansehen, ohne dass etwas geschieht** — UX-13
// (#1685).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { DateiansichtDTO } from '../apidateiansicht'
import { Dateiansicht } from './Dateiansicht'
import { Dateiverweis } from './Dateiverweis'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.bausteine.dateiansicht

afterEach(() => {
  raeumeApi()
})

function ansicht(teil: Partial<DateiansichtDTO>): DateiansichtDTO {
  return {
    name: 'liste.txt',
    pfad: '/Nutzer/ich/liste.txt',
    art: 'text',
    vorschau: 'text',
    groesse: 14,
    geaendert: '2026-09-26T10:00:00Z',
    verzeichnis: false,
    oeffnen: true,
    text: '',
    gekuerzt: false,
    bild: '',
    breite: 0,
    hoehe: 0,
    seiten: 0,
    eintraege: [],
    unsicher: 0,
    verschluesselt: false,
    zuGross: false,
    unlesbar: false,
    fehler: null,
    ...teil,
  }
}

function zeige(pfad = '/Nutzer/ich/liste.txt') {
  return render(
    <Sprachraum sprache="de">
      <Dateiansicht pfad={pfad} />
    </Sprachraum>,
  )
}

describe('Die Dateiansicht', () => {
  it('zeigt einen Text als Text und sagt, dass er weitergeht', async () => {
    belegeApi({
      DateiAnsehen: vi
        .fn()
        .mockResolvedValue(
          ansicht({ text: '4 Blitzschlag\n<script>', gekuerzt: true }),
        ),
    })

    zeige()

    expect(await screen.findByText(/4 Blitzschlag/)).toBeInTheDocument()
    expect(screen.getByText(texte.gekuerzt)).toBeInTheDocument()
    expect(screen.getByText('Text')).toBeInTheDocument()
    expect(screen.getByText('14 B')).toBeInTheDocument()

    // **Das Skript ist Text und kein Knoten.**
    expect(document.querySelector('script')).toBeNull()
  })

  it('zeigt ein Bild über img mit seinen Maßen', async () => {
    belegeApi({
      DateiAnsehen: vi.fn().mockResolvedValue(
        ansicht({
          name: 'karte.png',
          art: 'bild',
          vorschau: 'bild',
          bild: 'data:image/png;base64,AAAA',
          breite: 375,
          hoehe: 523,
        }),
      ),
    })

    zeige('/Nutzer/ich/karte.png')

    const bild = await screen.findByRole('img', { name: 'karte.png' })
    expect(bild).toHaveAttribute('src', 'data:image/png;base64,AAAA')
    expect(screen.getByText('375 × 523 Punkte')).toBeInTheDocument()
  })

  it('listet ein Paket und markiert, was nicht betreten wird', async () => {
    belegeApi({
      DateiAnsehen: vi.fn().mockResolvedValue(
        ansicht({
          name: 'magic.mse-game',
          art: 'paket',
          vorschau: 'liste',
          verzeichnis: true,
          oeffnen: false,
          eintraege: [
            { name: 'game', groesse: 18, verzeichnis: false, unsicher: false },
            { name: 'bilder', groesse: 0, verzeichnis: true, unsicher: false },
            { name: 'fremd', groesse: 0, verzeichnis: false, unsicher: true },
          ],
          unsicher: 1,
        }),
      ),
    })

    zeige('/Nutzer/ich/magic.mse-game')

    expect(await screen.findByText('3 Einträge')).toBeInTheDocument()
    expect(screen.getByText('bilder/')).toBeInTheDocument()
    expect(screen.getByText(/1 Einträge führen/)).toBeInTheDocument()
    expect(screen.getByText(texte.unsicherEintrag)).toBeInTheDocument()

    // **Ein Paket wird nicht geöffnet** — der Knopf fehlt.
    expect(
      screen.queryByRole('button', { name: /mit seinem Programm/ }),
    ).toBeNull()
  })

  it.each([
    [
      'eine Sicherung mit Kennwort',
      { verschluesselt: true },
      texte.verschluesselt,
    ],
    ['eine zu große Datei', { zuGross: true }, texte.zuGross],
    ['eine unlesbare Datei', { unlesbar: true }, texte.unlesbar],
    [
      'eine unbekannte Art',
      { art: 'unbekannt' as const, vorschau: 'keine' as const, oeffnen: false },
      texte.keineVorschau,
    ],
  ])('sagt bei %s, warum die Vorschau fehlt', async (_, teil, satz) => {
    belegeApi({ DateiAnsehen: vi.fn().mockResolvedValue(ansicht(teil)) })

    zeige()

    expect(await screen.findByText(satz)).toBeInTheDocument()
  })

  it('nennt die Seiten eines PDF', async () => {
    belegeApi({
      DateiAnsehen: vi.fn().mockResolvedValue(
        ansicht({
          name: 'bogen.pdf',
          art: 'pdf',
          vorschau: 'keine',
          seiten: 5,
        }),
      ),
    })

    zeige('/Nutzer/ich/bogen.pdf')

    expect(await screen.findByText('5 Seiten')).toBeInTheDocument()
    expect(screen.queryByText(texte.keineVorschau)).toBeNull()
  })

  it('zeigt einen Fehler statt der Ansicht', async () => {
    belegeApi({
      DateiAnsehen: vi.fn().mockResolvedValue(
        ansicht({
          fehler: {
            code: 'nichtGefunden',
            meldung: 'liste.txt gibt es nicht',
            schluessel: '',
          },
        }),
      ),
    })

    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'liste.txt gibt es nicht',
    )
  })

  it('gibt die Datei an den Finder und an ihr Programm', async () => {
    const finder = vi.fn().mockResolvedValue({ fehler: null })
    const oeffnen = vi.fn().mockResolvedValue({
      fehler: { code: 'validierung', meldung: 'kein Programm', schluessel: '' },
    })

    belegeApi({
      DateiAnsehen: vi.fn().mockResolvedValue(ansicht({ text: 'x' })),
      DateiImFinderZeigen: finder,
      DateiMitProgrammOeffnen: oeffnen,
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: 'liste.txt im Finder zeigen' }),
    )
    await waitFor(() => {
      expect(finder).toHaveBeenCalledWith('/Nutzer/ich/liste.txt')
    })

    // **Erst die Antwort, dann der nächste Knopf** — solange der Finder
    // antwortet, ist er angehalten (§7, Regel 4).
    await antwortAbwarten()

    await userEvent.click(
      screen.getByRole('button', {
        name: 'liste.txt mit seinem Programm öffnen',
      }),
    )

    // **Ein Fehler von macOS kommt an** — am Knopf, nicht verschluckt.
    expect(await screen.findByRole('alert')).toHaveTextContent('kein Programm')
    expect(oeffnen).toHaveBeenCalledWith('/Nutzer/ich/liste.txt')
  })
})

describe('Der Verweis auf eine geschriebene Datei', () => {
  // **Angesehen wird erst auf Wunsch** — eine Ausfuhr von 20 MB wird nicht
  // bei jedem Schreiben ein zweites Mal gelesen.
  it('liest die Datei erst, wenn jemand sie ansehen will', async () => {
    const ansehen = vi.fn().mockResolvedValue(ansicht({ text: 'Inhalt' }))
    belegeApi({ DateiAnsehen: ansehen })

    render(
      <Sprachraum sprache="de">
        <Dateiverweis pfad="/Nutzer/ich/liste.txt" />
      </Sprachraum>,
    )

    const knopf = screen.getByRole('button', { name: 'liste.txt ansehen' })
    expect(ansehen).not.toHaveBeenCalled()

    await userEvent.click(knopf)
    expect(await screen.findByText('Inhalt')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: texte.zuklappen }))
    expect(screen.queryByText('Inhalt')).toBeNull()
  })

  it('zeigt einen gewöhnlichen Ordner ohne Vorschau und ohne Grund', async () => {
    belegeApi({
      DateiAnsehen: vi.fn().mockResolvedValue(
        ansicht({
          name: 'bilder',
          art: 'unbekannt',
          vorschau: 'keine',
          verzeichnis: true,
          oeffnen: false,
        }),
      ),
    })

    zeige('/Nutzer/ich/bilder')

    expect(
      await screen.findByRole('button', { name: 'bilder im Finder zeigen' }),
    ).toBeInTheDocument()
    expect(screen.queryByText(texte.keineVorschau)).toBeNull()
    expect(screen.queryByText(texte.groesse)).toBeNull()
  })
})
