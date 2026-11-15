// **Ein Sitz an einem Profil** — Werkstück G8 von Phase 5.5
// (#1465).
//
// # Wogegen hier geprüft wird
//
//   - **Beim Wechsel ist die fremde Hand weg** — die Abnahme verlangt es
//     „im Fenster, nicht nur in einem Lauf": was Sicht A zeigte, steht in
//     Sicht B nicht mehr da.
//   - Ein Sitz **mit** Profil trägt dessen Namen in der Sichtwahl; zwei
//     gleichnamige Sitze sind dadurch auseinanderzuhalten.
//   - Ein **Befund** über ein verschwundenes Profil steht da — als Hinweis
//     und nicht als Fehler.
import { render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { Partietisch } from './Partietisch'
import { sitzwort } from './sitzwort'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einNutzer } from '../test/anmeldebelegung'
import { einLauf, einObjekt, einePartie, einSitz, einZug } from './partieprobe'
import { kataloge } from '../sprache/kataloge'
import { probeprofil } from '../test/profil'

const eigene = kataloge.de.seiten.spielen

afterEach(() => {
  raeumeApi()
})

// **Das Profil trennt zwei Sitze, die gleich heißen.**
it('nennt das Profil neben dem Namen', () => {
  const namen = new Map([['2', 'Turnier']])

  expect(sitzwort(einSitz({ profilID: '2' }), namen)).toBe('Anna (Turnier)')
  expect(sitzwort(einSitz(), namen)).toBe('Anna')

  // **Und ein Profil, das es nicht mehr gibt, steht nicht da** — was fehlt,
  // sagt der Befund und nicht diese Zeile.
  expect(sitzwort(einSitz({ profilID: '9' }), namen)).toBe('Anna')
})

// **Ein Befund ist ein Hinweis und keine Störung.**
it('zeigt einen Befund als Hinweis und nicht als Fehler', () => {
  const partie = einePartie()
  const lauf = einLauf({
    partie,
    befunde: [
      'der Sitz "sitz1" ist an ein Profil gebunden, das es nicht mehr gibt (9)',
    ],
  })

  render(
    <Sprachraum sprache="de">
      <Partietisch lauf={lauf} partie={partie} />
    </Sprachraum>,
  )

  expect(screen.getByText(eigene.sitzprofilBefunde)).toBeInTheDocument()
  expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  expect(screen.getByRole('status')).toHaveTextContent('nicht mehr gibt')
})

// **Und ohne Befund steht kein Kasten da.**
it('zeigt ohne Befund keinen Hinweis', () => {
  const partie = einePartie()

  render(
    <Sprachraum sprache="de">
      <Partietisch lauf={einLauf({ partie })} partie={partie} />
    </Sprachraum>,
  )

  expect(screen.queryByText(eigene.sitzprofilBefunde)).not.toBeInTheDocument()
})

// **Beim Wechsel ist die fremde Hand weg.**
//
// # Warum das hier steht und nicht nur in Go
//
// **Weil die Abnahme es ausdrücklich verlangt**: „nachgewiesen im Fenster,
// nicht nur in einem Lauf". Die Projektion rechnet je Sitz — aber gesehen
// wird sie am Bildschirm, und nur dort zeigt sich, dass nach dem Wechsel
// nichts von vorhin stehen bleibt.
//
// **Zweimal gezeichnet und nicht einmal umgeschaltet**: so kommt die zweite
// Sicht als das an, was sie ist — eine **andere** Antwort des Kerns, nicht
// ein gefiltertes Bild der ersten.
it('zeigt nach dem Sichtwechsel die fremde Hand nicht mehr', () => {
  const sichtVonAnna = einePartie({
    fuer: 'sitz1',
    objekte: [
      einObjekt({
        kennung: 'sitz1/k1/0',
        zone: 'hand',
        name: 'Wald',
        geheim: false,
      }),
      einObjekt({
        kennung: 'sitz2/k1/0',
        beherrscher: 'sitz2',
        eigner: 'sitz2',
        zone: 'hand',
        geheim: true,
      }),
    ],
  })

  const { unmount } = render(
    <Sprachraum sprache="de">
      <Partietisch
        lauf={einLauf({ partie: sichtVonAnna })}
        partie={sichtVonAnna}
      />
    </Sprachraum>,
  )

  expect(screen.getAllByText(/Wald/).length).toBeGreaterThan(0)

  unmount()

  // **Die Sicht von Bert kennt Annas Hand nicht** — der Kern hat sie gar
  // nicht erst geschickt.
  const sichtVonBert = einePartie({
    fuer: 'sitz2',
    objekte: [
      einObjekt({
        kennung: 'sitz1/k1/0',
        zone: 'hand',
        name: '',
        geheim: true,
      }),
      einObjekt({
        kennung: 'sitz2/k1/0',
        beherrscher: 'sitz2',
        eigner: 'sitz2',
        zone: 'hand',
        name: 'Insel',
        geheim: false,
      }),
    ],
  })

  render(
    <Sprachraum sprache="de">
      <Partietisch
        lauf={einLauf({ partie: sichtVonBert })}
        partie={sichtVonBert}
      />
    </Sprachraum>,
  )

  expect(screen.queryByText(/Wald/)).not.toBeInTheDocument()
  expect(screen.getAllByText(/Insel/).length).toBeGreaterThan(0)
})

// **Und der Profilname steht wirklich in der Sichtwahl** — von den eigenen
// Profilen bis zum Eintrag.
//
// # Warum das neben dem Lauf über [sitzwort] steht
//
// **Weil jener nur die Zeile prüft, die den Namen baut.** Ob die Namen
// überhaupt ankommen, hängt an `useEigeneProfile` und an der Abbildung
// daneben — und genau dort war in dieser Datei bis zum 2026-09-17 keine
// einzige Antwort belegt: die Abbildung blieb leer, und der Lauf war grün
// (§10, „eine Zusage, die zur Hälfte hält").
it('nennt das Profil in der Sichtwahl des Tisches', async () => {
  belegeApi({
    Nutzer: () =>
      Promise.resolve({
        nutzer: [
          {
            ...einNutzer(),
            profile: [probeprofil({ id: '10', name: 'Turnier' })],
          },
        ],
        fehler: null,
      }),
    WerArbeitet: () => Promise.resolve({ nutzerID: '1', fehler: null }),
  })

  const partie = einePartie({
    sitze: [
      einSitz({ profilID: '10' }),
      einSitz({ kennung: 'sitz2', deckabdruck: 'def' }),
    ],
  })

  render(
    <Sprachraum sprache="de">
      <Partietisch lauf={einLauf({ partie })} partie={partie} />
    </Sprachraum>,
  )

  expect(
    await screen.findByRole('option', { name: 'Anna (Turnier)' }),
  ).toBeInTheDocument()

  // **Und der ungebundene Sitz trägt nur seinen Namen** — sonst sagte die
  // Liste zweimal dasselbe, und die Trennung wäre keine.
  expect(screen.getByRole('option', { name: 'Anna' })).toBeInTheDocument()
})

// **Und ein gelöschtes Profil verschwindet aus der Sichtwahl, sobald der
// Befund dasteht.**
//
// # Der Widerspruch, für den es diesen Lauf gibt
//
// **Gesehen am 2026-09-17 an der laufenden Anwendung**: der Hinweis sagte
// „dieses Profil gibt es nicht mehr", und einen Zentimeter darüber führte die
// Sichtwahl seinen Namen. Die Namen standen seit dem Öffnen des Tisches da,
// der Befund kam bei jedem Lesen frisch — zwei Sätze, die einander
// widersprechen (§10).
//
// **Geprüft wird der Weg und nicht die Zeile**: dieselbe Vorrichtung
// antwortet zweimal verschieden, und erst der Befund holt die zweite Antwort.
it('nimmt ein gelöschtes Profil aus der Sichtwahl', async () => {
  let profile = [probeprofil({ id: '10', name: 'Turnier' })]

  belegeApi({
    Nutzer: () =>
      Promise.resolve({
        nutzer: [{ ...einNutzer(), profile }],
        fehler: null,
      }),
    WerArbeitet: () => Promise.resolve({ nutzerID: '1', fehler: null }),
  })

  const partie = einePartie({
    sitze: [
      einSitz({ profilID: '10' }),
      einSitz({ kennung: 'sitz2', name: 'Bert' }),
    ],
  })

  const { rerender } = render(
    <Sprachraum sprache="de">
      <Partietisch lauf={einLauf({ partie })} partie={partie} />
    </Sprachraum>,
  )

  expect(
    await screen.findByRole('option', { name: 'Anna (Turnier)' }),
  ).toBeInTheDocument()

  // **Das Profil ist weg — und der Kern sagt es als Befund.**
  profile = []

  const befund =
    'der Sitz "sitz1" ist an ein Profil gebunden, das es nicht mehr gibt (10)'

  rerender(
    <Sprachraum sprache="de">
      <Partietisch
        lauf={einLauf({ partie, befunde: [befund] })}
        partie={partie}
      />
    </Sprachraum>,
  )

  expect(
    await screen.findByRole('option', { name: 'Anna' }),
  ).toBeInTheDocument()

  expect(
    screen.queryByRole('option', { name: 'Anna (Turnier)' }),
  ).not.toBeInTheDocument()
})

// **Der Passen-Knopf am Tisch stellt den Befehl** — Werkstück G9 von
// Phase 5.5
// (#1592).
//
// # Warum das hier steht und nicht nur im Lauf über die Leiste
//
// **Weil jener die Leiste allein zeichnet.** Dass der Tisch sie **einhängt**
// und ihren Rückruf mit `zug.passen` verbindet, steht nur hier — und genau
// das ist die Naht, an der ein Knopf sonst ins Leere klickt (§10).
it('stellt beim Passen den Zugbefehl', async () => {
  const befohlen = vi.fn()

  const partie = einePartie({
    fuer: 'sitz1',
    zug: einZug({ nummer: 1, aktiv: 'sitz1', prioritaet: 'sitz1' }),
  })

  render(
    <Sprachraum sprache="de">
      <Partietisch
        lauf={einLauf({ partie, befehlen: befohlen })}
        partie={partie}
      />
    </Sprachraum>,
  )

  const hand = nutzerereignis.setup()

  await hand.click(screen.getByRole('button', { name: eigene.zugPassen }))

  expect(befohlen).toHaveBeenCalledWith('zug.passen', {})
})

// **Der Aufgeben-Knopf am Tisch stellt den Befehl** — CR 104.3a (M17,
// #1618).
//
// **Und er ist etwas anderes als „Partie schließen"**: das schließt das
// Fenster, dies beendet das Spiel. Der Lauf prüft beide, weil ein
// vertauschter Rückruf sonst niemandem auffiele.
it('stellt beim Aufgeben den Zugbefehl und schließt nicht', async () => {
  const befohlen = vi.fn()
  const geraeumt = vi.fn()

  const partie = einePartie({ fuer: 'sitz1' })

  render(
    <Sprachraum sprache="de">
      <Partietisch
        lauf={einLauf({ partie, befehlen: befohlen, raeumen: geraeumt })}
        partie={partie}
      />
    </Sprachraum>,
  )

  const hand = nutzerereignis.setup()

  await hand.click(screen.getByRole('button', { name: eigene.aufgeben }))
  // **Seit P1b (#2370) erst nach der Rückfrage.**
  await hand.click(screen.getByRole('button', { name: eigene.aufgebenJa }))

  expect(befohlen).toHaveBeenCalledWith('zug.aufgeben', {})
  expect(geraeumt).not.toHaveBeenCalled()

  await hand.click(screen.getByRole('button', { name: eigene.verlassen }))

  expect(geraeumt).toHaveBeenCalledTimes(1)
})

// **Die zwei Abkürzungsknöpfe am Tisch stellen ihre Befehle** — CR 732 und
// CR 724 (M1,
// #1603).
//
// **Aus demselben Grund wie beim Passen**: dass der Tisch die Leiste
// **einhängt**, steht nur hier. Ein Knopf, dessen Rückruf ins Leere geht,
// sieht aus wie einer, der funktioniert (§10).
it('stellt die Befehle der Abkürzung und des Zugendes', async () => {
  const befohlen = vi.fn()

  const partie = einePartie({
    fuer: 'sitz1',
    zug: einZug({ nummer: 1, aktiv: 'sitz1', prioritaet: 'sitz1' }),
  })

  render(
    <Sprachraum sprache="de">
      <Partietisch
        lauf={einLauf({ partie, befehlen: befohlen })}
        partie={partie}
      />
    </Sprachraum>,
  )

  const hand = nutzerereignis.setup()

  await hand.click(screen.getByRole('button', { name: eigene.zugAbkuerzen }))
  await hand.click(screen.getByRole('button', { name: eigene.zugBeenden }))

  expect(befohlen).toHaveBeenCalledWith('zug.abkuerzen', {})
  expect(befohlen).toHaveBeenCalledWith('zug.beenden', {})
})
