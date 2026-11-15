// **Was eine Karte ist, steht an ihr** — Werkstück M2 der Regelmaschine
// (#1604).
//
// # Wogegen hier geprüft wird
//
// **Dass Typ und Werte am Kartenknopf stehen** — und zwar die **gelebten**:
// eine 2/3 mit zwei +1/+1-Marken ist 4/5 (CR 122.1a), und die Rechnung
// geschieht in der Domäne. Das Fenster zeigt, was ankommt.
//
// **Und dass ein Geheimnis eines bleibt.** Wer die Werte einer verdeckten
// Karte läse, wüsste, welche es ist.
import { screen } from '@testing-library/react'
import { expect, it } from 'vitest'

import { einLauf, einObjekt, einePartie } from './partieprobe'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { zeige } from './partietischprobe'

const eigene = kataloge.de.seiten.spielen

// einBaer gibt eine offene Karte mit Eigenschaften.
const einBaer = (steht = true) =>
  einObjekt({
    kennung: 'b',
    zone: 'tisch',
    name: 'Grizzlybär',
    geheim: false,
    eigenschaften: {
      name: 'Grizzlybär',
      farben: ['G'],
      manawert: 2,
      supertypen: [],
      typen: ['creature'],
      untertypen: ['Bear'],
      staerke: { steht, stern: false, zahl: 4 },
      widerstandskraft: { steht, stern: false, zahl: 5 },
      loyalitaet: { steht: false, stern: false, zahl: 0 },
      verteidigung: { steht: false, stern: false, zahl: 0 },
      unbekannteTypen: [],
    },
  })

it('nennt Typ und gelebte Werte am Kartenknopf', async () => {
  const hand = zeige(einLauf(), einePartie({ objekte: [einBaer()] }), false)

  await hand.click(screen.getByRole('button', { name: eigene.tischListe }))

  expect(screen.getByText('creature — Bear')).toBeInTheDocument()
  expect(screen.getByText('4/5')).toBeInTheDocument()
})

// **Ein Land trägt keine Werte** — CR 208.3, und `steht: false` sagt es.
it('zeigt keine Werte, wo keine stehen', async () => {
  const hand = zeige(
    einLauf(),
    einePartie({ objekte: [einBaer(false)] }),
    false,
  )

  await hand.click(screen.getByRole('button', { name: eigene.tischListe }))

  expect(screen.getByText('creature — Bear')).toBeInTheDocument()
  expect(screen.queryByText('4/5')).not.toBeInTheDocument()
})

it('zeigt an einer geheimen Karte nichts davon', async () => {
  const verdeckt = einObjekt({
    kennung: 'v',
    zone: 'tisch',
    geheim: true,
    eigenschaften: null,
  })

  const hand = zeige(einLauf(), einePartie({ objekte: [verdeckt] }), false)

  await hand.click(screen.getByRole('button', { name: eigene.tischListe }))

  expect(screen.getByText(eigene.geheim)).toBeInTheDocument()
  expect(screen.queryByText('creature — Bear')).not.toBeInTheDocument()
})

// **Ein Typ, den das Regelwerk nicht kennt, steht als Befund da** — §8, und
// die dritte Abnahme von M2: kein stiller Ausfall.
it('nennt einen unbekannten Kartentyp, statt ihn zu verschlucken', async () => {
  const fremd = einObjekt({
    kennung: 'f',
    zone: 'tisch',
    name: 'Zauberding',
    geheim: false,
    eigenschaften: {
      name: 'Zauberding',
      farben: [],
      manawert: 0,
      supertypen: [],
      typen: ['creature'],
      untertypen: [],
      staerke: { steht: false, stern: false, zahl: 0 },
      widerstandskraft: { steht: false, stern: false, zahl: 0 },
      loyalitaet: { steht: false, stern: false, zahl: 0 },
      verteidigung: { steht: false, stern: false, zahl: 0 },
      unbekannteTypen: ['Zauberding'],
    },
  })

  const hand = zeige(einLauf(), einePartie({ objekte: [fremd] }), false)

  await hand.click(screen.getByRole('button', { name: eigene.tischListe }))

  // **Ohne Untertypen steht kein Gedankenstrich da** (CR 205.3a) — und das
  // ist zugleich der Zweig, den der Lauf darüber nicht betritt.
  expect(screen.getByText('creature')).toBeInTheDocument()

  expect(
    screen.getByText(
      einsetzen(eigene.unbekannterTyp, { woerter: 'Zauberding' }),
    ),
  ).toBeInTheDocument()
})

// **Ein Stern bleibt ein Stern** — CR 208.2: was er bedeutet, sagt eine
// Fähigkeit (CR 604.3), und eine Zahl an seiner Stelle wäre erfunden.
it('zeigt einen Stern als Stern', async () => {
  const stern = einObjekt({
    kennung: 's',
    zone: 'tisch',
    name: 'Nachtmahr',
    geheim: false,
    eigenschaften: {
      name: 'Nachtmahr',
      farben: ['B'],
      manawert: 5,
      supertypen: [],
      typen: ['creature'],
      untertypen: ['Nightmare', 'Horse'],
      staerke: { steht: true, stern: true, zahl: 0 },
      widerstandskraft: { steht: true, stern: true, zahl: 0 },
      loyalitaet: { steht: false, stern: false, zahl: 0 },
      verteidigung: { steht: false, stern: false, zahl: 0 },
      unbekannteTypen: [],
    },
  })

  const hand = zeige(einLauf(), einePartie({ objekte: [stern] }), false)

  await hand.click(screen.getByRole('button', { name: eigene.tischListe }))

  expect(screen.getByText('*/*')).toBeInTheDocument()
})

// **Der Name einer Karte steht in ihren Eigenschaften und nicht am Objekt.**
//
// `Sichtobjekt.Name` trägt ihn nur, wo **keine** Karte dahintersteht — so
// schickt es der Kern. Am 2026-09-24, mit dem kanonischen Set „Green" am
// Tisch, stand in dieser Liste „p0-1" statt „A-Canopy Tactician": die
// Leinwand war am 2026-09-17 behoben worden, die Liste nicht. Die Läufe hier
// setzten den Namen bis dahin an **beide** Stellen und sahen es nie (§3,
// eine Vorrichtung, die großzügiger ist als der Betrieb).
it('nennt eine Karte beim Namen aus ihren Eigenschaften', async () => {
  const baer = { ...einBaer(), name: '', karte: 'p0-1' }

  const hand = zeige(einLauf(), einePartie({ objekte: [baer] }), false)

  await hand.click(screen.getByRole('button', { name: eigene.tischListe }))

  expect(screen.getByText('Grizzlybär')).toBeInTheDocument()
  expect(screen.queryByText('p0-1')).not.toBeInTheDocument()
})
