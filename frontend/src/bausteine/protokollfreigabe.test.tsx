// **Wer das Protokoll später liest** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
//   - Die **drei Grenzen** stehen da: wirkt erst später, holt nichts zurück,
//     betrifft diesen Rechner. Eine Zusage, die niemand liest, wirkt wie ein
//     Mangel (§6, §10).
//   - Die Empfängerliste wird **zerlegt** und der Stern als Wort gezeigt.
//   - Gelöscht wird nur, was beendet ist — der Knopf steht sonst nicht da.
import { render, screen, waitFor, within } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { Protokollfreigabe } from './Protokollfreigabe'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { kurzform } from '../sprache/kurzform'

const eigene = kataloge.de.seiten.spielen

// zeige zeichnet den Abschnitt und gibt die Hand.
function zeige(beendet = false, eigenerSitz = 'anna') {
  render(
    <Sprachraum sprache="de">
      <Protokollfreigabe
        kennung="p1"
        eigenerSitz={eigenerSitz}
        beendet={beendet}
      />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

// leereAntwort ist, was die Naht ohne Einwilligung schickt.
const leereAntwort = { offenlegungen: [], fehler: null }

it('nennt den Abschnitt bei seinem Namen und zeigt die drei Grenzen', () => {
  belegeApi({ PartieOffenlegungen: vi.fn() })

  zeige(true)

  expect(
    screen.getByRole('region', { name: eigene.protokoll }),
  ).toBeInTheDocument()

  for (const satz of [
    eigene.protokollNurSpaeter,
    eigene.protokollKeineRueckholung,
    eigene.protokollAufbewahrung,
    eigene.protokollLoeschenErklaerung,
  ]) {
    expect(screen.getByText(kurzform(satz))).toBeInTheDocument()
  }
})

// **Ohne Klick geht nichts hinaus** — und der Beleg ist die Liste, die noch
// gar nicht dasteht (§3).
it('holt die Freigaben nicht von selbst', () => {
  const geholt = vi.fn()
  belegeApi({ PartieOffenlegungen: geholt })

  zeige()

  expect(screen.queryByText(eigene.protokollNiemand)).not.toBeInTheDocument()
  expect(geholt).not.toHaveBeenCalled()
})

it('holt die Freigaben auf Zuruf und sagt, wenn keine da ist', async () => {
  const geholt = vi.fn().mockResolvedValue(leereAntwort)
  belegeApi({ PartieOffenlegungen: geholt })

  const hand = zeige()

  await hand.click(
    screen.getByRole('button', { name: eigene.protokollAnsehen }),
  )

  expect(await screen.findByText(eigene.protokollNiemand)).toBeInTheDocument()
  expect(geholt).toHaveBeenCalledWith('p1')
})

// **Der Stern wird zu einem Wort** — wer ihn stehenließe, verlangte von einem
// Menschen, die Form auf der Leitung zu kennen.
it('zeigt jede Freigabe mit ihren Empfängern', async () => {
  belegeApi({
    PartieOffenlegungen: vi.fn().mockResolvedValue({
      offenlegungen: [
        { von: 'anna', an: ['bert'] },
        { von: 'cem', an: ['*'] },
      ],
      fehler: null,
    }),
  })

  const hand = zeige()

  await hand.click(
    screen.getByRole('button', { name: eigene.protokollAnsehen }),
  )

  expect(
    await screen.findByText(
      `${eigene.protokollSitz} anna — ${eigene.protokollFuer} bert`,
    ),
  ).toBeInTheDocument()

  expect(
    screen.getByText(
      `${eigene.protokollSitz} cem — ${eigene.protokollFuer} ${eigene.protokollAlle}`,
    ),
  ).toBeInTheDocument()
})

// **Die Liste wird zerlegt und Leeres fällt weg.**
it('schickt die Empfänger einzeln und ohne leere Stücke', async () => {
  const freigegeben = vi.fn().mockResolvedValue(leereAntwort)
  belegeApi({ PartieOffenlegen: freigegeben })

  const hand = zeige()

  await hand.type(
    screen.getByLabelText(eigene.protokollEmpfaenger),
    'bert, cem, ',
  )

  await hand.click(
    screen.getByRole('button', { name: eigene.protokollFreigeben }),
  )

  await waitFor(() => {
    expect(freigegeben).toHaveBeenCalledWith('p1', 'anna', ['bert', 'cem'])
  })
})

// **Ohne Empfänger geht nichts hinaus** — und ohne eigenen Sitz auch nicht:
// freigeben kann nur, wem etwas gehört.
it('hält den Knopf an, solange nichts dasteht', async () => {
  belegeApi({ PartieOffenlegen: vi.fn() })

  zeige()

  expect(
    screen.getByRole('button', { name: eigene.protokollFreigeben }),
  ).toBeDisabled()

  const ohneSitz = zeige(false, '')

  // **Und ohne eigenen Sitz auch nicht, obwohl etwas dasteht.**
  const [, zweitesFeld] = screen.getAllByLabelText(eigene.protokollEmpfaenger)
  const [, zweiterKnopf] = screen.getAllByRole('button', {
    name: eigene.protokollFreigeben,
  })

  if (zweitesFeld === undefined || zweiterKnopf === undefined) {
    throw new Error('der zweite Abschnitt fehlt')
  }

  await ohneSitz.type(zweitesFeld, 'bert')

  expect(zweiterKnopf).toBeDisabled()
})

it('nimmt eine Freigabe zurück', async () => {
  const zurueck = vi.fn().mockResolvedValue(leereAntwort)
  belegeApi({ PartieOffenlegungZuruecknehmen: zurueck })

  const hand = zeige()

  await hand.click(
    screen.getByRole('button', { name: eigene.protokollZuruecknehmen }),
  )

  await waitFor(() => {
    expect(zurueck).toHaveBeenCalledWith('p1', 'anna')
  })
})

// **Eine laufende Partie wird nicht gelöscht** — der Knopf steht gar nicht da.
it('bietet das Löschen nur bei einer beendeten Partie an', () => {
  belegeApi({ PartieVergessen: vi.fn() })

  zeige(false)

  expect(
    screen.queryByRole('button', { name: eigene.protokollLoeschen }),
  ).not.toBeInTheDocument()
})

it('löscht das Protokoll und sagt es', async () => {
  const geloescht = vi.fn().mockResolvedValue(leereAntwort)
  belegeApi({ PartieVergessen: geloescht })

  const hand = zeige(true)

  await hand.click(
    screen.getByRole('button', { name: eigene.protokollLoeschen }),
  )

  // **Erst nach der Frage** (UX-10.AK-04, #1682) — ein Klick genügte bis zum
  // 2026-09-25, und die Erklärung stand nur darunter. **Abbrechen löscht
  // nichts**, und danach steht die Erklärung wieder unter dem Knopf.
  expect(geloescht).not.toHaveBeenCalled()
  await hand.click(
    screen.getByRole('button', {
      name: kataloge.de.bausteine.aktionen.abbrechen,
    }),
  )
  expect(geloescht).not.toHaveBeenCalled()
  await hand.click(
    screen.getByRole('button', { name: eigene.protokollLoeschen }),
  )
  await hand.click(
    within(
      screen.getByRole('group', { name: eigene.protokollLoeschenFrage }),
    ).getByRole('button', { name: eigene.protokollLoeschenJa }),
  )

  expect(await screen.findByText(eigene.protokollGeloescht)).toBeInTheDocument()

  expect(geloescht).toHaveBeenCalledWith('p1')
})

// **Ein Fehler steht als Meldung da und wirft die Liste nicht weg** (§7,
// Regel 3).
it('meldet einen Fehler und lässt das Gelesene stehen', async () => {
  const geholt = vi
    .fn()
    .mockResolvedValueOnce({
      offenlegungen: [{ von: 'anna', an: ['bert'] }],
      fehler: null,
    })
    .mockResolvedValueOnce({
      offenlegungen: [],
      fehler: { code: 'konflikt', meldung: 'ging nicht' },
    })

  belegeApi({ PartieOffenlegungen: geholt })

  const hand = zeige()

  const knopf = screen.getByRole('button', { name: eigene.protokollAnsehen })

  await hand.click(knopf)
  expect(
    await screen.findByText(
      `${eigene.protokollSitz} anna — ${eigene.protokollFuer} bert`,
    ),
  ).toBeInTheDocument()

  await hand.click(knopf)

  expect(await screen.findByRole('alert')).toHaveTextContent('ging nicht')

  expect(
    screen.getByText(
      `${eigene.protokollSitz} anna — ${eigene.protokollFuer} bert`,
    ),
  ).toBeInTheDocument()
})

// **Zurückgenommen ist zurückgenommen** (§7, Regel 3; #1970): kam die Liste
// danach nicht, sagt der Abschnitt das an ihrer Stelle — und nicht
// „niemand", was die leere Antwort behauptete.
it('nimmt ein Zurücknehmen mit Nachladefehler als geschehen', async () => {
  belegeApi({
    PartieOffenlegungZuruecknehmen: vi.fn().mockResolvedValue({
      offenlegungen: [],
      nachladefehler: {
        code: 'validierung',
        schluessel: '',
        meldung: 'Die Liste klemmt',
      },
      fehler: null,
    }),
  })

  const hand = zeige()

  await hand.click(
    screen.getByRole('button', { name: eigene.protokollZuruecknehmen }),
  )

  expect(await screen.findByRole('alert')).toHaveTextContent('Die Liste klemmt')
  expect(screen.queryByText(eigene.protokollNiemand)).toBeNull()
})
