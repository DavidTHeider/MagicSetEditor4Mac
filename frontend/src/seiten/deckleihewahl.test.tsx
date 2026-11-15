// **Ziel und Umschlag einer Leihe wählt ein Dateidialog** — UXR-F15
// (#2030).
//
// # Wogegen hier geprüft wird
//
// **Gegen die Leihe, wie sie bis dahin war**: *Ziel* und *Umschlag* standen
// als Pfad in einem Freitextfeld, und ein Pfad von Hand ist auf macOS kein
// Bedienelement (R7 in `docs/UIREVIEW.md`).
//
//   - **Gewählt und nicht getippt**: kein Textfeld mehr; der Dialog öffnet
//     auf Zuruf, nicht von selbst.
//   - **Der Pfad steht lesbar daneben** und lässt sich ändern — durch einen
//     zweiten Dialog.
//   - **Abbrechen lässt den vorigen Stand stehen**, und es ist keine
//     Meldung wert. Scheitert der Dialog, steht der Grund am Feld.
//   - **Solange ein Dialog offen ist, geht kein zweiter auf** (§7, Regel 4).
import { screen, waitFor, within } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import {
  eigene,
  gewaehlt,
  leer,
  leiheApi,
  oeffneAnnahme,
  umschlaggruppe,
  umschlagGewaehlt,
  waehleBertUndDeck,
  waehleUmschlag,
  waehleZiel,
  weggeklickt,
  zeige,
  zielgruppe,
} from '../test/leiheprobe'

it('fragt den Ort nicht von selbst und bietet kein Textfeld', async () => {
  const frage = vi.fn()
  leiheApi({ DeckleiheZielWaehlen: frage, UmschlagWaehlen: frage })

  const hand = zeige()

  expect(frage).not.toHaveBeenCalled()
  expect(
    screen.queryByRole('textbox', { name: eigene.leiheZiel }),
  ).not.toBeInTheDocument()

  // **Auch das Annehmen tippt keinen Pfad.**
  await oeffneAnnahme(hand)

  expect(
    screen.queryByRole('textbox', { name: eigene.leiheUmschlag }),
  ).not.toBeInTheDocument()
  expect(frage).not.toHaveBeenCalled()
})

it('sagt, dass noch kein Ort gewählt ist, und bietet die Wahl', async () => {
  leiheApi({ DeckleiheZielWaehlen: vi.fn() })

  const hand = zeige()

  expect(within(zielgruppe()).getByText(eigene.leiheZielKein)).toBeVisible()
  expect(
    within(zielgruppe()).getByRole('button', { name: eigene.leiheZielWaehlen }),
  ).toBeEnabled()

  await oeffneAnnahme(hand)

  expect(
    within(umschlaggruppe()).getByText(eigene.leiheUmschlagKein),
  ).toBeVisible()
  expect(
    within(umschlaggruppe()).getByRole('button', {
      name: eigene.leiheUmschlagWaehlen,
    }),
  ).toBeEnabled()
})

it('zeigt den gewählten Pfad daneben und ändert ihn durch eine neue Wahl', async () => {
  const wahl = vi
    .fn()
    .mockResolvedValueOnce(gewaehlt('/Users/ada/Leihen/erste.msepackage'))
    .mockResolvedValueOnce(gewaehlt('/Users/ada/Leihen/zweite.msepackage'))

  leiheApi({ DeckleiheZielWaehlen: wahl })

  const hand = zeige()

  await waehleZiel(hand, '/Users/ada/Leihen/erste.msepackage')

  // **Der Satz „noch kein Ort" geht, und die Aufschrift wechselt**: wer
  // schon gewählt hat, will einen anderen, nicht „einen" Ort.
  expect(
    within(zielgruppe()).queryByText(eigene.leiheZielKein),
  ).not.toBeInTheDocument()

  await hand.click(
    within(zielgruppe()).getByRole('button', { name: eigene.leiheZielAnderes }),
  )

  expect(
    await within(zielgruppe()).findByText(
      '/Users/ada/Leihen/zweite.msepackage',
    ),
  ).toBeVisible()
  expect(
    within(zielgruppe()).queryByText('/Users/ada/Leihen/erste.msepackage'),
  ).not.toBeInTheDocument()
  expect(wahl).toHaveBeenCalledTimes(2)
})

// **Abbrechen im Dialog lässt den vorigen Stand stehen** — und der Pfad, der
// danach hinausgeht, ist der vorige.
it('lässt beim Abbrechen das Ziel stehen und legt dorthin ab', async () => {
  const wahl = vi
    .fn()
    .mockResolvedValueOnce(gewaehlt('/tmp/erste.msepackage'))
    .mockResolvedValueOnce(weggeklickt)

  const angeboten = vi.fn().mockResolvedValue(leer)

  leiheApi({ DeckleiheZielWaehlen: wahl, DeckleiheAnbieten: angeboten })

  const hand = zeige()

  await waehleBertUndDeck(hand)
  await waehleZiel(hand, '/tmp/erste.msepackage')

  await hand.click(
    within(zielgruppe()).getByRole('button', { name: eigene.leiheZielAnderes }),
  )
  await waitFor(() => {
    expect(wahl).toHaveBeenCalledTimes(2)
  })

  // **Ein Abbruch ist keine Meldung wert.**
  expect(within(zielgruppe()).queryByRole('alert')).not.toBeInTheDocument()
  expect(within(zielgruppe()).getByText('/tmp/erste.msepackage')).toBeVisible()

  // **Auf den bedienbaren Knopf gewartet**, nicht auf den Ruf: solange der
  // Dialog offen war, ruhte er.
  await waitFor(() => {
    expect(
      within(zielgruppe()).getByRole('button', {
        name: eigene.leiheZielAnderes,
      }),
    ).toBeEnabled()
  })

  await hand.click(screen.getByRole('button', { name: eigene.leiheAnbieten }))

  await waitFor(() => {
    expect(angeboten).toHaveBeenCalledWith(
      expect.anything(),
      '/tmp/erste.msepackage',
    )
  })
})

it('hält das Verleihen an, bis ein Ort gewählt ist', async () => {
  leiheApi({
    DeckleiheZielWaehlen: vi.fn().mockResolvedValue(weggeklickt),
    DeckleiheAnbieten: vi.fn(),
  })

  const hand = zeige()

  await waehleBertUndDeck(hand)

  expect(
    screen.getByRole('button', { name: eigene.leiheAnbieten }),
  ).toBeDisabled()

  // **Ein weggeklickter Dialog gibt keinen Ort.**
  await hand.click(
    within(zielgruppe()).getByRole('button', { name: eigene.leiheZielWaehlen }),
  )
  await waitFor(() => {
    expect(
      within(zielgruppe()).getByRole('button', {
        name: eigene.leiheZielWaehlen,
      }),
    ).toBeEnabled()
  })

  expect(
    screen.getByRole('button', { name: eigene.leiheAnbieten }),
  ).toBeDisabled()
  expect(within(zielgruppe()).getByText(eigene.leiheZielKein)).toBeVisible()
})

// **Scheitert der Dialog, steht der Grund am Feld** (§6) — und der vorige
// Pfad bleibt.
it('meldet einen Fehler des Dialogs am Ziel und behält den Pfad', async () => {
  const wahl = vi
    .fn()
    .mockResolvedValueOnce(gewaehlt('/tmp/erste.msepackage'))
    .mockResolvedValueOnce({
      pfad: '',
      abgebrochen: false,
      fehler: {
        code: 'unbekannt',
        schluessel: '',
        meldung: 'Kein Fenster zum Fragen',
      },
    })

  leiheApi({ DeckleiheZielWaehlen: wahl })

  const hand = zeige()

  await waehleZiel(hand, '/tmp/erste.msepackage')

  await hand.click(
    within(zielgruppe()).getByRole('button', { name: eigene.leiheZielAnderes }),
  )

  expect(await within(zielgruppe()).findByRole('alert')).toHaveTextContent(
    'Kein Fenster zum Fragen',
  )
  expect(within(zielgruppe()).getByText('/tmp/erste.msepackage')).toBeVisible()

  // **Der nächste Versuch räumt die Meldung** — sie gehörte dem vorigen.
  wahl.mockResolvedValueOnce(gewaehlt('/tmp/dritte.msepackage'))

  await hand.click(
    within(zielgruppe()).getByRole('button', { name: eigene.leiheZielAnderes }),
  )

  expect(
    await within(zielgruppe()).findByText('/tmp/dritte.msepackage'),
  ).toBeVisible()
  expect(within(zielgruppe()).queryByRole('alert')).not.toBeInTheDocument()
})

// **Ein Dialog je Wahl** (§7, Regel 4): ein zweiter Klick, solange der erste
// offen ist, öffnet keinen zweiten.
it('öffnet keinen zweiten Dialog, solange der erste offen ist', async () => {
  let antworten: (wert: ReturnType<typeof gewaehlt>) => void = () => undefined

  const wahl = vi.fn().mockImplementation(
    () =>
      new Promise<ReturnType<typeof gewaehlt>>((fertig) => {
        antworten = fertig
      }),
  )

  leiheApi({ DeckleiheZielWaehlen: wahl })

  const hand = zeige()
  const knopf = within(zielgruppe()).getByRole('button')

  await hand.click(knopf)

  await waitFor(() => {
    expect(knopf).toBeDisabled()
  })

  await hand.click(knopf)
  expect(wahl).toHaveBeenCalledTimes(1)

  antworten(gewaehlt('/tmp/spaet.msepackage'))

  expect(
    await within(zielgruppe()).findByText('/tmp/spaet.msepackage'),
  ).toBeVisible()
  expect(knopf).toBeEnabled()
})

it('wählt den Umschlag über den Dialog und nimmt ihn von dort an', async () => {
  const wahl = vi
    .fn()
    .mockResolvedValueOnce(umschlagGewaehlt('/tmp/erste.msepackage'))
    .mockResolvedValueOnce(umschlagGewaehlt('/tmp/zweite.msepackage'))
  const genommen = vi.fn().mockResolvedValue(leer)

  leiheApi({ UmschlagWaehlen: wahl, DeckleiheAnnehmen: genommen })

  const hand = zeige()

  await oeffneAnnahme(hand)

  // **Ohne Umschlag nichts zu tun.**
  expect(
    screen.getByRole('button', { name: eigene.leiheAnnehmen }),
  ).toBeDisabled()

  await waehleUmschlag(hand, '/tmp/erste.msepackage')

  expect(
    screen.getByRole('button', { name: eigene.leiheAnnehmen }),
  ).toBeEnabled()

  // **Ändern heißt: ein zweiter Dialog.**
  await hand.click(
    within(umschlaggruppe()).getByRole('button', {
      name: eigene.leiheUmschlagAnderer,
    }),
  )
  await within(umschlaggruppe()).findByText('/tmp/zweite.msepackage')

  await hand.click(screen.getByRole('button', { name: eigene.leiheAnnehmen }))

  await waitFor(() => {
    expect(genommen).toHaveBeenCalledWith('/tmp/zweite.msepackage')
  })
})

it('lässt beim Abbrechen den Umschlag stehen', async () => {
  const wahl = vi
    .fn()
    .mockResolvedValueOnce(umschlagGewaehlt('/tmp/fremd.msepackage'))
    .mockResolvedValueOnce({
      ...umschlagGewaehlt(''),
      abgebrochen: true,
    })
  const genommen = vi.fn().mockResolvedValue(leer)

  leiheApi({ UmschlagWaehlen: wahl, DeckleiheAnnehmen: genommen })

  const hand = zeige()

  await oeffneAnnahme(hand)
  await waehleUmschlag(hand, '/tmp/fremd.msepackage')

  await hand.click(
    within(umschlaggruppe()).getByRole('button', {
      name: eigene.leiheUmschlagAnderer,
    }),
  )
  await waitFor(() => {
    expect(wahl).toHaveBeenCalledTimes(2)
  })
  await waitFor(() => {
    expect(
      within(umschlaggruppe()).getByRole('button', {
        name: eigene.leiheUmschlagAnderer,
      }),
    ).toBeEnabled()
  })

  expect(within(umschlaggruppe()).queryByRole('alert')).not.toBeInTheDocument()
  expect(
    within(umschlaggruppe()).getByText('/tmp/fremd.msepackage'),
  ).toBeVisible()

  await hand.click(screen.getByRole('button', { name: eigene.leiheAnnehmen }))

  await waitFor(() => {
    expect(genommen).toHaveBeenCalledWith('/tmp/fremd.msepackage')
  })
})

// **Ein Umschlag, der sich nicht lesen lässt, wird nicht gewählt** — der
// Grund steht am Feld, und ein vorher gewählter bleibt.
it('meldet einen unlesbaren Umschlag am Feld', async () => {
  const wahl = vi
    .fn()
    .mockResolvedValueOnce(umschlagGewaehlt('/tmp/gut.msepackage'))
    .mockResolvedValueOnce({
      ...umschlagGewaehlt(''),
      fehler: {
        code: 'unbekannt',
        schluessel: '',
        meldung: 'Die Unterschrift trägt nicht',
      },
    })

  leiheApi({ UmschlagWaehlen: wahl, DeckleiheAnnehmen: vi.fn() })

  const hand = zeige()

  await oeffneAnnahme(hand)
  await waehleUmschlag(hand, '/tmp/gut.msepackage')

  await hand.click(
    within(umschlaggruppe()).getByRole('button', {
      name: eigene.leiheUmschlagAnderer,
    }),
  )

  expect(await within(umschlaggruppe()).findByRole('alert')).toHaveTextContent(
    'Die Unterschrift trägt nicht',
  )
  expect(
    within(umschlaggruppe()).getByText('/tmp/gut.msepackage'),
  ).toBeVisible()
  expect(
    screen.getByRole('button', { name: eigene.leiheAnnehmen }),
  ).toBeEnabled()
})
