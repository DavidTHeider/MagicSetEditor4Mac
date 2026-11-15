// **Womit sich dieses Profil ausweist** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Wogegen hier geprüft wird
//
// **Dass nichts angelegt wird, solange niemand es will.** Der Kasten ist
// zugeklappt, und auch aufgeklappt fragt er nur — der Knopf zum Anlegen steht
// erst da, wenn es nichts gibt, und das Ablösen hat eine Rückfrage.
//
// **Und dass der Fingerabdruck dasteht.** Er ist das, was zwei Menschen
// vergleichen; eine Kennung allein läse niemand vor.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Peeridentitaet } from './Peeridentitaet'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { kurzform } from '../sprache/kurzform'

const eigene = kataloge.de.seiten.technik

afterEach(() => {
  raeumeApi()
})

// eineLage baut eine Antwort.
function eineLage(anders: Record<string, unknown> = {}) {
  return {
    stand: 'bereit',
    kennung: 'bkennungdieserprobe',
    fingerabdruck: '1111-2222-3333-4444-5555-6666-7777-8888',
    seit: '2026-09-14T03:00:00Z',
    vorher: '',
    vorherFingerabdruck: '',
    befunde: [],
    fehler: null,
    ...anders,
  }
}

// zeige stellt den Kasten auf und schlägt ihn auf.
async function zeige(): Promise<ReturnType<typeof userEvent.setup>> {
  render(
    <Sprachraum sprache="de">
      <Peeridentitaet />
    </Sprachraum>,
  )

  const hand = userEvent.setup()

  // **Der Knopf heißt nicht wie die Überschrift** — UXR-24 (#1932): derselbe
  // Satz stand zweimal übereinander, einmal als Überschrift, einmal als Knopf.
  expect(screen.getAllByText(eigene.peerTitel)).toHaveLength(1)

  await hand.click(screen.getByRole('button', { name: eigene.peerZeigen }))

  return hand
}

// **Zugeklappt wird nicht gefragt.**
//
// **Die Gegenprobe zum Aufklappen** (§10): ein Kasten, der beim Zeichnen
// fragte, ginge bei jedem Seitenwechsel an den Schlüsselbund — und der zeigt
// unter Umständen einen Dialog.
it('fragt zugeklappt nicht nach der Identität', () => {
  const gefragt = vi.fn(() => Promise.resolve(eineLage()))

  belegeApi({ Peeridentitaet: gefragt })

  render(
    <Sprachraum sprache="de">
      <Peeridentitaet />
    </Sprachraum>,
  )

  expect(gefragt).not.toHaveBeenCalled()
})

// **Aufgeschlagen steht der Fingerabdruck da — und die Kennung darunter.**
it('zeigt Fingerabdruck, Kennung und Zeitpunkt', async () => {
  belegeApi({ Peeridentitaet: () => Promise.resolve(eineLage()) })

  await zeige()

  expect(
    await screen.findByText('1111-2222-3333-4444-5555-6666-7777-8888'),
  ).toBeInTheDocument()
  expect(screen.getByText('bkennungdieserprobe')).toBeInTheDocument()
  expect(screen.getByText(eigene.peerStandBereit)).toBeInTheDocument()
  expect(
    screen.getByText(kurzform(eigene.peerFingerabdruckHinweis)),
  ).toBeInTheDocument()
})

// **Was diese Identität nicht ist** — NET-02, S13.NAT-05 (#1706): weder
// das Apple-Konto noch der Name dieses Rechners. Wer beides dafür hielt,
// suchte den Fehler an der falschen Stelle.
it('sagt, dass die Identität weder Apple-Konto noch Rechnername ist', async () => {
  belegeApi({ Peeridentitaet: () => Promise.resolve(eineLage()) })

  await zeige()

  expect(await screen.findByText(eigene.peerNichtKonto)).toBeInTheDocument()
})

// **Ohne Identität steht der Knopf zum Anlegen da — und sonst keiner.**
it('bietet das Anlegen nur an, wenn es keine gibt', async () => {
  const angelegt = vi.fn(() => Promise.resolve(eineLage()))

  belegeApi({
    Peeridentitaet: () =>
      Promise.resolve(eineLage({ stand: 'ohne', kennung: '', seit: '' })),
    PeeridentitaetSicherstellen: angelegt,
  })

  const hand = await zeige()

  expect(await screen.findByText(eigene.peerStandOhne)).toBeInTheDocument()
  expect(
    screen.queryByRole('button', { name: eigene.peerRotieren }),
  ).not.toBeInTheDocument()

  await hand.click(screen.getByRole('button', { name: eigene.peerAnlegen }))

  expect(angelegt).toHaveBeenCalled()
})

// **Das Ablösen fragt vorher — und die Frage nennt die Folge.**
it('fragt vor dem Ablösen und löst erst danach ab', async () => {
  const rotiert = vi.fn(() => Promise.resolve(eineLage()))

  belegeApi({
    Peeridentitaet: () => Promise.resolve(eineLage()),
    PeeridentitaetRotieren: rotiert,
  })

  const hand = await zeige()

  await hand.click(
    await screen.findByRole('button', { name: eigene.peerRotieren }),
  )

  expect(screen.getByText(eigene.peerRotierenFrage)).toBeInTheDocument()
  expect(rotiert).not.toHaveBeenCalled()

  await hand.click(screen.getByRole('button', { name: eigene.peerRotierenJa }))

  expect(rotiert).toHaveBeenCalled()
})

// **Und ein Abbrechen löst nichts ab.**
it('löst nichts ab, wenn die Frage abgebrochen wird', async () => {
  const rotiert = vi.fn(() => Promise.resolve(eineLage()))

  belegeApi({
    Peeridentitaet: () => Promise.resolve(eineLage()),
    PeeridentitaetRotieren: rotiert,
  })

  const hand = await zeige()

  await hand.click(
    await screen.findByRole('button', { name: eigene.peerRotieren }),
  )
  await hand.click(
    screen.getByRole('button', { name: eigene.peerRotierenNein }),
  )

  expect(screen.queryByText(eigene.peerRotierenFrage)).not.toBeInTheDocument()
  expect(rotiert).not.toHaveBeenCalled()
})

// **Die Befunde stehen da, und die abgelöste Kennung dazu.**
it('nennt Befunde und die vorige Identität', async () => {
  belegeApi({
    Peeridentitaet: () =>
      Promise.resolve(
        eineLage({
          vorher: 'balteKennung',
          vorherFingerabdruck: '9999-8888-7777-6666-5555-4444-3333-2222',
          befunde: ['Diese Identität ist neu.'],
        }),
      ),
  })

  await zeige()

  expect(
    await screen.findByText('Diese Identität ist neu.'),
  ).toBeInTheDocument()
  expect(
    screen.getByText(/9999-8888-7777-6666-5555-4444-3333-2222/),
  ).toBeInTheDocument()
})

// **Ein unbekannter Stand steht als er selbst da.**
//
// **Und nicht als leere Zeile**: ein fünfter Stand, den dieses Fenster noch
// nicht kennt, ist immer noch eine Auskunft.
it('zeigt einen unbekannten Stand als sich selbst', async () => {
  belegeApi({
    Peeridentitaet: () => Promise.resolve(eineLage({ stand: 'kuenftig' })),
  })

  await zeige()

  expect(await screen.findByText('kuenftig')).toBeInTheDocument()
})

// **Ein Fehler beim Lesen steht als Meldung da.**
it('zeigt einen Fehler beim Lesen', async () => {
  belegeApi({
    Peeridentitaet: () =>
      Promise.resolve(
        eineLage({ fehler: { code: 'validierung', meldung: 'Kein Bestand.' } }),
      ),
  })

  await zeige()

  expect(await screen.findByRole('alert')).toHaveTextContent('Kein Bestand.')
})

// **Und ein Fehler beim Schreiben nimmt die Antwort nicht weg.**
//
// **Geschrieben ist geschrieben** (§7, Regel 3): was dastand, bleibt stehen,
// und die Meldung sagt, was schiefging.
it('zeigt einen Fehler beim Anlegen und behält, was dastand', async () => {
  belegeApi({
    Peeridentitaet: () =>
      Promise.resolve(eineLage({ stand: 'ohne', kennung: '', seit: '' })),
    PeeridentitaetSicherstellen: () =>
      Promise.resolve(
        eineLage({
          fehler: { code: 'konflikt', meldung: 'Der Schlüssel fehlt.' },
        }),
      ),
  })

  const hand = await zeige()

  await hand.click(
    await screen.findByRole('button', { name: eigene.peerAnlegen }),
  )

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Der Schlüssel fehlt.',
  )
  expect(screen.getByText(eigene.peerStandOhne)).toBeInTheDocument()
})
