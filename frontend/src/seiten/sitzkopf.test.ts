// **Was am Sitz steht** — PLAY-03
// (#1711).
import { describe, expect, it } from 'vitest'

import { einePartie, einSitz } from './partieprobe'
import { kataloge } from '../sprache/kataloge'
import { sitzkopf, sitzkopfFuer } from './partiewoerter'

const eigene = kataloge.de.seiten.spielen

const partie = einePartie({
  sitze: [
    einSitz({
      kennung: 'ada',
      name: 'Ada',
      leben: 20,
      vergiftung: 0,
      profilID: 'p1',
    }),
    einSitz({ kennung: 'bo', name: 'Bo', leben: 17, vergiftung: 2 }),
  ],
  fuer: 'ada',
})

describe('sitzkopf', () => {
  it('nennt Namen und Leben', () => {
    const satz = sitzkopf(partie, eigene, 'ada', (sitz) => sitz.name)

    expect(satz).toContain('Ada')
    expect(satz).toContain('20')
    expect(satz).toContain(eigene.standLeben)
  })

  // **Gift steht nur, wo es welches gibt.** Eine Null an jedem Sitz wäre eine
  // Auskunft über nichts — und sie stünde auf jedem Tisch, auch auf denen, wo
  // niemand je vergiftet wird.
  it('nennt Gift nur, wo es welches gibt', () => {
    expect(sitzkopf(partie, eigene, 'ada', (sitz) => sitz.name)).not.toContain(
      eigene.standGift,
    )
    expect(sitzkopf(partie, eigene, 'bo', (sitz) => sitz.name)).toContain(
      eigene.standGift,
    )
  })

  // **Ein Sitz, den die Partie nicht führt, bekommt seine Kennung** (§8):
  // besser eine Kennung als ein leeres Schild, das aussieht wie ein Fehler
  // im Zeichner.
  it('gibt für einen unbekannten Sitz die Kennung', () => {
    expect(sitzkopf(partie, eigene, 'niemand', (sitz) => sitz.name)).toBe(
      'niemand',
    )
  })
})

describe('sitzkopfFuer', () => {
  // **Was sich am Tisch nicht ändert, wird gebunden.** Der Zeichner bekommt
  // eine Funktion über die Kennung; Partie, Katalog und Profilnamen stehen
  // schon fest.
  it('gibt eine Funktion über die Sitzkennung', () => {
    const fuer = sitzkopfFuer(partie, eigene, new Map())

    expect(fuer('bo')).toContain('Bo')
    expect(fuer('bo')).toContain(eigene.standGift)
  })

  // **Und der Name kommt von außen** — damit auf der Leinwand derselbe steht
  // wie in der Sichtwahl daneben (G8: zwei Sitze dürfen gleich heißen, und
  // das Profil trennt sie).
  it('nimmt den Namen, den der Aufrufer gibt', () => {
    const fuer = sitzkopfFuer(partie, eigene, new Map([['p1', 'Mein Profil']]))

    expect(fuer('ada')).toContain('Ada (Mein Profil)')
  })
})

describe('wer am Zug ist', () => {
  // **Der Sitzkopf sagt es, und zwar nur an einem Sitz.** Der Entwurf
  // zeichnet es so („am Zug" gegen „wartet"); dieses Fenster nennt nur den,
  // der dran ist — „wartet" an drei von vier Sitzen wäre dreimal dieselbe
  // Auskunft.
  it('nennt den aktiven Sitz', () => {
    const amZug = einePartie({
      sitze: partie.sitze,
      fuer: 'ada',
      zug: { ...partie.zug, aktiv: 'bo' },
    })

    const fuer = sitzkopfFuer(amZug, eigene, new Map())

    expect(fuer('bo')).toContain(eigene.zugAktiv)
    expect(fuer('ada')).not.toContain(eigene.zugAktiv)
  })

  // **Ohne Zug steht es an keinem.** Ein Sandkasten aus G1 bis G8 geht nicht
  // zugweise zu; dort stünde sonst an einem Sitz eine Auskunft über etwas,
  // das es nicht gibt.
  it('nennt ihn nicht, wo es keinen Zug gibt', () => {
    const ohneZug = einePartie({
      sitze: partie.sitze,
      fuer: 'ada',
      zug: { ...partie.zug, aktiv: '' },
    })

    const fuer = sitzkopfFuer(ohneZug, eigene, new Map())

    expect(fuer('ada')).not.toContain(eigene.zugAktiv)
    expect(fuer('bo')).not.toContain(eigene.zugAktiv)
  })
})

// **Strahlung und Steuer stehen am Sitz** — PLAY-03, CR 728.1 und CR 903.8
// (#1711).
//
// **Nur, wo es sie gibt** — dieselbe Regel wie beim Gift: eine Null wäre
// eine Auskunft über nichts, und sie stünde auf jedem Tisch.
describe('die Zahlen einer Commanderpartie am Sitz', () => {
  const amTisch = einePartie({
    sitze: [
      einSitz({
        kennung: 'ada',
        name: 'Ada',
        leben: 33,
        strahlung: 3,
        commandersteuer: 2,
      }),
      einSitz({ kennung: 'bo', name: 'Bo' }),
    ],
    fuer: 'ada',
  })

  it('nennt Strahlung und Steuer, wo es sie gibt', () => {
    const satz = sitzkopf(amTisch, eigene, 'ada', (sitz) => sitz.name)

    expect(satz).toContain(eigene.standStrahlung)
    expect(satz).toContain('3')
    expect(satz).toContain(eigene.standSteuer)
    expect(satz).toContain('2')
  })

  it('lässt sie weg, wo es sie nicht gibt', () => {
    const satz = sitzkopf(amTisch, eigene, 'bo', (sitz) => sitz.name)

    expect(satz).not.toContain(eigene.standStrahlung)
    expect(satz).not.toContain(eigene.standSteuer)
  })

  // **Der Commanderschaden steht hier ausdrücklich nicht.** Eine Zeile auf
  // der Leinwand trägt keine Liste — sie steht an jedem Sitz und wird bei
  // vier Sitzen viermal gelesen. Er steht in der Spielerleiste, und das ist
  // eine Abweichung vom Artboard mit Grund (`docs/PROTOTYP-0.6.0.md`).
  it('trägt den Commanderschaden nicht auf die Leinwand', () => {
    const mitSchaden = einePartie({
      sitze: [
        einSitz({
          kennung: 'ada',
          name: 'Ada',
          commanderschaden: { 'bo/drache/0': 21 },
        }),
      ],
      fuer: 'ada',
    })

    expect(
      sitzkopf(mitSchaden, eigene, 'ada', (sitz) => sitz.name),
    ).not.toContain('21')
  })
})
