// **Der Tisch** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Wogegen hier geprüft wird
//
// **Dass der Tisch die Sicht **eines** Sitzes zeigt** — und dass der
// Sitzwechsel eine **neue** Sicht holt, statt die alte umzurechnen. Wer sie
// umrechnete, hätte das Geheimnis des einen im Speicher des anderen.
//
// **Und dass die Mischpflicht jedem dasteht.** Sie ist die öffentliche
// Hälfte eines Blicks; stünde sie nur beim Betroffenen, wäre sie eine Notiz
// an sich selbst.
import { screen, within } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import { einLauf, einObjekt, einePartie, volleAuskunft } from './partieprobe'
import { zeige } from './partietischprobe'
import { stelle } from '../test/handlungsprobe'
import { kurzform } from '../sprache/kurzform'
import { einsetzen } from '../sprache/einsetzen'

const eigene = kataloge.de.seiten.spielen

afterEach(() => {
  raeumeApi()
})

// **Der Wirt wird genannt, bevor jemand mitspielt.**
//
// **Das Issue verlangt es wörtlich**: „diese Vertrauensgrenze wird vor
// Beitritt offengelegt". Ein Hinweis, der erst nach dem Mitspielen käme,
// wäre keiner.
it('nennt den Wirt und seine Vertrauensgrenze', async () => {
  const hand = zeige(einLauf())

  // **Die Kopfzeile und nicht irgendeine Fundstelle**: das Wort steht auch
  // im Hinweis darunter, und ein `getByText` über beide sagte nichts
  // darüber, dass der **Name** dasteht.
  expect(
    screen.getByText(new RegExp(`${eigene.wirt}: Anna`)),
  ).toBeInTheDocument()

  // **Die Grenze steht beim Namen, hinter einem Fragezeichen** — seit dem
  // 2026-09-24 (Befund F09): als Absatz nahm sie jedem Tisch zwei Zeilen
  // Leinwand. Geprüft wird, dass sie dort **ankommt**.
  await hand.click(
    screen.getByRole('button', {
      name: einsetzen(kataloge.de.bausteine.feldhilfe, { feld: eigene.wirt }),
    }),
  )

  expect(await screen.findByRole('tooltip')).toHaveTextContent(
    kurzform(eigene.wirtHinweis),
  )
})

// **Ein Sitzwechsel holt eine neue Sicht.**
it('holt beim Sitzwechsel die Sicht des anderen', async () => {
  const angesehen = vi.fn()

  const hand = zeige(einLauf({ ansehen: angesehen }))

  await hand.selectOptions(screen.getByLabelText(eigene.sicht), 'sitz2')

  expect(angesehen).toHaveBeenCalledWith('p1', 'sitz2')
})

// **Jede Zone steht da — auch die leeren.**
//
// **Auch die leeren**, denn „nichts zu sehen" ist eine Auskunft: wer den
// Friedhof nicht fände, wüsste nicht, ob er leer ist oder ob die Seite ihn
// verschweigt.
it('zeigt jede Zone, leer oder nicht', () => {
  zeige(einLauf())

  // **Innerhalb der Zonenliste gesucht**: dieselben Wörter stehen auch im
  // Auswahlfeld der Handlungen, und ein Fund dort sagte über den Tisch
  // nichts.
  const liste = screen.getByRole('heading', { name: eigene.zonen })
    .nextElementSibling as HTMLElement

  for (const wort of [
    eigene.zoneHand,
    eigene.zoneTisch,
    eigene.zoneStapel,
    eigene.zoneBibliothek,
    eigene.zoneFriedhof,
    eigene.zoneExil,
  ]) {
    expect(within(liste).getByText(wort)).toBeInTheDocument()
  }

  // **Die Kommandozone gehört der Variante** — PLAY-03.AK-01
  // (#1711):
  // an einem gewöhnlichen Tisch wäre sie ein Feld, das nie etwas trägt.
  expect(within(liste).queryByText(eigene.zoneKommando)).not.toBeInTheDocument()

  expect(within(liste).getAllByText(eigene.zoneLeer)).toHaveLength(5)
})

// **Und sie steht da, sobald etwas darin liegt** — ein Emblem entsteht in
// der Kommandozone (CR 114.2), auch im Sandkasten. Eine Zone auszublenden,
// in der etwas liegt, hieße eine Karte zu verstecken (§8).
it('zeigt die Kommandozone, sobald etwas darin liegt', () => {
  zeige(
    einLauf(),
    einePartie({
      objekte: [einObjekt({ kennung: 'e1', zone: 'kommando', geheim: false })],
    }),
  )

  const liste = screen.getByRole('heading', { name: eigene.zonen })
    .nextElementSibling as HTMLElement

  expect(within(liste).getByText(eigene.zoneKommando)).toBeInTheDocument()
})

// **Eine getappte Karte sagt es, und eine offene trägt ihren Namen.**
it('beschriftet eine offene Karte mit Namen und Zustand', () => {
  zeige(
    einLauf(),
    einePartie({
      objekte: [
        einObjekt({
          zone: 'tisch',
          name: 'Wald',
          geheim: false,
          getappt: true,
        }),
      ],
    }),
  )

  // **Ein Teilstück und nicht der ganze Name** — seit `M12` (#1614) steht
  // der Beherrscher mit an der Karte (CR 110.2), und er gehört in den
  // zugänglichen Namen: wer nur hört, soll wissen, wessen Karte da liegt.
  expect(
    screen.getByRole('button', {
      name: new RegExp(`Wald \\(${eigene.getappt}\\)`),
    }),
  ).toBeInTheDocument()
})

// **Eine Karte ohne Namen zeigt ihre Kartenkennung.**
//
// **Der Zweig, der sonst nie betreten wird**: ein Objekt darf einen Druck
// nennen, ohne einen Namen zu tragen — ein Spielstein zum Beispiel.
it('zeigt die Kartenkennung, wo kein Name steht', () => {
  zeige(
    einLauf(),
    einePartie({
      objekte: [
        einObjekt({ zone: 'tisch', karte: 'k7', name: '', geheim: false }),
      ],
    }),
  )

  expect(screen.getByRole('button', { name: /k7/ })).toBeInTheDocument()
})

// **Die Mischpflicht steht mit Namen und Zone da.**
it('nennt eine offene Mischpflicht', () => {
  zeige(
    einLauf(),
    einePartie({ mischpflichten: [{ sitz: 'sitz2', zone: 'bibliothek' }] }),
  )

  expect(
    screen.getByRole('heading', { name: eigene.mischpflichtTitel }),
  ).toBeInTheDocument()
  expect(
    screen.getByText(`Bert · ${eigene.zoneBibliothek}`),
  ).toBeInTheDocument()
})

// **Und ohne eine steht gar nichts da** — seit dem 2026-09-24. Hier stand
// bis dahin ein Satz, der sagte, dass keine aussteht; er nahm in der Spalte
// neben dem Tisch drei Zeilen über der gewählten Karte, in jedem Zug, und
// einen nächsten Schritt nannte er nicht (§6). Das Artboard *Tisch* führt
// die Mischpflichten gar nicht.
it('zeigt keine Mischpflichten, solange keine aussteht', () => {
  zeige(einLauf())

  expect(
    screen.queryByRole('heading', { name: eigene.mischpflichtTitel }),
  ).not.toBeInTheDocument()
})

// **Eine unbekannte Zone und ein unbekannter Stand fallen auf ihr Wort
// zurück.**
//
// **Der Rückfall ist Absicht** (§8, „fremde Daten sind der Normalfall"): ein
// Kern, der eine neue Zone einführte, ließe das Fenster sonst mit einer
// leeren Zeile dastehen.
it('zeigt ein unbekanntes Wort unverändert', () => {
  zeige(
    einLauf(),
    einePartie({
      stand: 'schwebt',
      mischpflichten: [{ sitz: 'sitz1', zone: 'kiste' }],
    }),
  )

  expect(screen.getByText(/schwebt/)).toBeInTheDocument()
  expect(screen.getByText('Anna · kiste')).toBeInTheDocument()
})

// **Und wer die Partie schließt, räumt auf.**
it('räumt beim Schließen', async () => {
  const geraeumt = vi.fn()

  const hand = zeige(einLauf({ raeumen: geraeumt }))

  await hand.click(screen.getByRole('button', { name: eigene.verlassen }))

  expect(geraeumt).toHaveBeenCalled()
})

// **Eine gewählte Karte steht als gewählt da — und nach dem Wechsel nicht
// mehr.**
//
// **§7, Regel 1**: der Sitz ist das Bezugsobjekt; was für den einen gewählt
// war, gehört dem anderen nicht.
it('hebt die Wahl beim Sitzwechsel auf', async () => {
  const hand = zeige(
    einLauf(),
    einePartie({
      objekte: [einObjekt({ zone: 'tisch', name: 'Wald', geheim: false })],
    }),
  )

  const karte = screen.getByRole('button', { name: /Wald/ })

  await hand.click(karte)

  expect(karte).toHaveAttribute('aria-pressed', 'true')

  await hand.selectOptions(screen.getByLabelText(eigene.sicht), 'sitz2')

  expect(screen.getByRole('button', { name: /Wald/ })).toHaveAttribute(
    'aria-pressed',
    'false',
  )
})

// **Ein Sitz, den es nicht gibt, steht mit seiner Kennung da.**
//
// **Der Rückfall ist Absicht** (§8): eine Mischpflicht für einen Sitz, den
// diese Sicht nicht führt, ist besser mit Kennung als gar nicht — ein leerer
// Eintrag sähe aus wie ein Fehler der Seite.
it('zeigt eine unbekannte Sitzkennung unverändert', () => {
  zeige(
    einLauf(),
    einePartie({ mischpflichten: [{ sitz: 'fremd', zone: 'bibliothek' }] }),
  )

  expect(
    screen.getByText(`fremd · ${eigene.zoneBibliothek}`),
  ).toBeInTheDocument()
})

// **Nach einem Schieben ist die Karte nicht mehr gewählt.**
//
// **§7, Regel 1 über die Naht zwischen Tisch und Handlungen**: der Tisch
// hält die Wahl, die Handlungen lösen sie — und dass beide zusammenpassen,
// sieht man nur, wenn sie zusammen laufen.
it('hebt die Wahl nach einem Schieben auf', async () => {
  const hand = zeige(
    einLauf({ handlungen: [volleAuskunft('wald')] }),
    einePartie({
      objekte: [
        einObjekt({
          kennung: 'wald',
          zone: 'tisch',
          name: 'Wald',
          geheim: false,
        }),
      ],
    }),
  )

  await hand.click(screen.getByRole('button', { name: /Wald/ }))

  expect(screen.getByRole('button', { name: /Wald/ })).toHaveAttribute(
    'aria-pressed',
    'true',
  )

  // **Das Schieben wird aufgeklappt und gestellt** — seine Felder stehen
  // seit UX-08 erst auf Zuruf da.
  await stelle(hand, 'objekt.schieben')

  expect(screen.getByRole('button', { name: /Wald/ })).toHaveAttribute(
    'aria-pressed',
    'false',
  )
})
