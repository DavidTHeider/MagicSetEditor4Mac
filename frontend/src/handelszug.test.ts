// **Ein Zug je Entscheidung** — P13
// (#2381),
// Befund C-16, Vermerk 0200.
//
// Was ein Klick geht: die Schritte seiner Gruppe, die der Vertrag als nächste
// nennt, bis das Gegenüber dran ist — **höchstens bis zu dem Schritt, den
// sein Knopf nennt**, und **nur von dem Stand aus, den der Mensch sah**
// (Befund 1 des Zweitblicks über Phase 8). In den Ausgang gelegt wird erst,
// wenn ein Stand verlassen wird, damit eine Stimme mit ihrem Stand reist und
// nicht als zweite Nachricht derselben Art drüben „bekannt" heißt.
import { afterEach, describe, expect, it } from 'vitest'

import { ziehe } from './handelszug'
import { du, einDienst, einStand, ich } from './test/handelszugprobe'
import { belegeApi, raeumeApi } from './test/api'

import type { Zugauftrag } from './handelszug'
import type { Handelsvorgang as Vorgangsstand } from './apihandelsleitung'

afterEach(() => {
  raeumeApi()
})

// zug belegt einen Dienst, der den gesehenen Stand bestätigt, und geht den
// Zug.
async function zug(
  folge: Record<string, Vorgangsstand>,
  auftrag: Omit<Zugauftrag, 'gegenueber'>,
) {
  const dienst = einDienst(folge, {}, auftrag.stand)
  belegeApi(dienst.api as never)
  const ausgang = await ziehe({ ...auftrag, gegenueber: du })

  return { ausgang, rufe: dienst.rufe }
}

const angebotenMitDir = einStand({
  zustand: 'angeboten',
  beschluss: 'zustimmen',
  zustimmungen: [du],
  naechster: 'zustimmen',
})

describe('ein Zug', () => {
  it('bietet an, stimmt zu und schickt beides in einer Nachricht', async () => {
    const { ausgang, rufe } = await zug(
      {
        anbieten: einStand({
          zustand: 'angeboten',
          beschluss: 'zustimmen',
          naechster: 'zustimmen',
        }),
        zustimmen: einStand({
          zustand: 'angeboten',
          beschluss: 'zustimmen',
          zustimmungen: [ich],
          lage: 'wartet',
        }),
      },
      {
        stand: einStand({ naechster: 'anbieten' }),
        schritt: 'anbieten',
        bis: 'zustimmen',
      },
    )

    expect(rufe).toEqual([
      'stand',
      'anbieten',
      'zustimmen',
      'ausgang',
      'zustellen',
    ])
    expect(ausgang.gegangen).toEqual(['anbieten', 'zustimmen'])
    expect(ausgang.stand.zustimmungen).toEqual([ich])
    expect(ausgang.getragen).toBe(1)
    expect(ausgang.gescheitert).toBeUndefined()
  })

  it('geht nach dem Zustimmen nicht ins Vorbereiten — das ist eine andere Entscheidung', async () => {
    const { ausgang, rufe } = await zug(
      {
        zustimmen: einStand({
          zustand: 'angenommen',
          naechster: 'vorbereiten',
          phase: 'zusagen',
        }),
      },
      { stand: angebotenMitDir, schritt: 'zustimmen', bis: 'bestaetigen' },
    )

    expect(rufe).toEqual(['stand', 'zustimmen', 'ausgang', 'zustellen'])
    expect(ausgang.stand.zustand).toBe('angenommen')
  })

  it('bereitet vor und bestätigt die Übergabe — die Stimme reist mit dem Stand', async () => {
    const { rufe } = await zug(
      {
        vorbereiten: einStand({
          zustand: 'vorbereitet',
          beschluss: 'uebergeben',
          naechster: 'uebergeben',
          phase: 'zusagen',
        }),
        uebergeben: einStand({
          zustand: 'vorbereitet',
          beschluss: 'uebergeben',
          zustimmungen: [ich],
          lage: 'wartet',
        }),
      },
      {
        stand: einStand({ zustand: 'angenommen', naechster: 'vorbereiten' }),
        schritt: 'vorbereiten',
        bis: 'uebergeben',
      },
    )

    expect(rufe).toEqual([
      'stand',
      'vorbereiten',
      'uebergeben',
      'ausgang',
      'zustellen',
    ])
  })

  it('übergibt, verbucht, meldet den Abgleich und bestätigt den eigenen Abschluss', async () => {
    const { ausgang, rufe } = await zug(
      {
        uebergeben: einStand({
          zustand: 'uebergabeOffen',
          naechster: 'verbuchen',
          phase: 'verbuchen',
        }),
        verbuchen: einStand({
          zustand: 'lokalVerbucht',
          naechster: 'abgleichen',
          phase: 'verbuchen',
        }),
        abgleichen: einStand({
          zustand: 'abgleichOffen',
          beschluss: 'bestaetigen',
          naechster: 'bestaetigen',
          phase: 'verbuchen',
        }),
        [`bestaetigen:${ich}`]: einStand({
          zustand: 'abgleichOffen',
          beschluss: 'bestaetigen',
          zustimmungen: [ich],
          lage: 'wartet',
          phase: 'verbuchen',
        }),
      },
      {
        stand: einStand({
          zustand: 'vorbereitet',
          beschluss: 'uebergeben',
          zustimmungen: [du],
          naechster: 'uebergeben',
        }),
        schritt: 'uebergeben',
        bis: 'bestaetigen',
      },
    )

    // **Die Übergabe geht in den Ausgang, bevor gebucht wird** — danach
    // trüge der Stand nur noch „bei mir ist gebucht", und drüben spielt die
    // Leitung höchstens einen Schritt nach.
    expect(rufe).toEqual([
      'stand',
      'uebergeben',
      'ausgang',
      'freigabe:verbuchen:',
      'verbuchen',
      'stand',
      'ausgang',
      'abgleichen',
      'stand',
      `freigabe:bestaetigen:${ich}`,
      `bestaetigen:${ich}`,
      'stand',
      'ausgang',
      'zustellen',
    ])
    expect(ausgang.gegangen).toEqual([
      'uebergeben',
      'verbuchen',
      'abgleichen',
      'bestaetigen',
    ])
  })

  it('legt eine Übergabe, die von drüben kam, in den Ausgang, bevor gebucht wird', async () => {
    // **Gefunden am nativen Fenster** (`p13abschluss09`): Annas Stimme zur
    // Übergabe reiste mit dem Schritt nach „vorbereitet" und zählte drüben
    // nicht (`NimmtStand`: eine Stimme gehört zu ihrem Beschluss). Die
    // Übergabe schloss deshalb bei Anna, als Berts Stimme kam — ohne dass Bert
    // es je erfuhr, und ein Abschlussschritt wird drüben nicht nachgespielt.
    const { rufe } = await zug(
      { verbuchen: einStand({ zustand: 'lokalVerbucht', lage: 'wartet' }) },
      {
        stand: einStand({ zustand: 'uebergabeOffen', naechster: 'verbuchen' }),
        schritt: 'verbuchen',
        bis: 'bestaetigen',
      },
    )

    expect(rufe).toEqual([
      'stand',
      'ausgang',
      'freigabe:verbuchen:',
      'verbuchen',
      'stand',
      'ausgang',
      'zustellen',
    ])
  })

  it('trägt die Stimme des Gegenübers allein ein und schickt den Beleg', async () => {
    const { ausgang, rufe } = await zug(
      {
        [`bestaetigen:${du}`]: einStand({
          zustand: 'beidseitigVerbucht',
          zustimmungen: [ich, du],
          lage: 'abgeschlossen',
          ende: true,
        }),
      },
      {
        stand: einStand({
          zustand: 'abgleichOffen',
          beschluss: 'bestaetigen',
          zustimmungen: [ich],
          lage: 'wartet',
        }),
        schritt: 'bestaetigen',
        bis: 'bestaetigen',
        wer: du,
      },
    )

    expect(rufe).toEqual([
      'stand',
      `freigabe:bestaetigen:${du}`,
      `bestaetigen:${du}`,
      'stand',
      'ausgang',
      'zustellen',
    ])
    expect(ausgang.stand.zustand).toBe('beidseitigVerbucht')
  })
})

describe('ein Zug, gebunden an das Gesehene', () => {
  it('nennt bei jedem Schritt den Abdruck des Stands, den er zuletzt kannte', async () => {
    // **#2496**: der Dienst prüft ihn gegen den Stand, auf den er den
    // Schritt anwendet — ein Gegenangebot dazwischen ist dann ein Konflikt
    // und keine Zustimmung zu einer ungesehenen Revision.
    const anfang = einStand({ naechster: 'anbieten', abdruck: 'a-1' })
    const dienst = einDienst(
      {
        anbieten: einStand({
          zustand: 'angeboten',
          beschluss: 'zustimmen',
          naechster: 'zustimmen',
          abdruck: 'a-2',
        }),
        zustimmen: einStand({
          zustand: 'angeboten',
          zustimmungen: [ich],
          lage: 'wartet',
          abdruck: 'a-2',
        }),
      },
      {},
      anfang,
    )
    belegeApi(dienst.api as never)

    await ziehe({
      stand: anfang,
      schritt: 'anbieten',
      bis: 'zustimmen',
      gegenueber: du,
    })

    expect(dienst.gesehen).toEqual(['a-1', 'a-2'])
  })

  it('geht nichts, wenn sich der Stand seit dem Ansehen geändert hat', async () => {
    // **Befund 1 des Zweitblicks über Phase 8**: ein Gegenangebot kam
    // zwischen Anzeige und Klick — „Zustimmen" hätte einer Revision
    // zugestimmt, die niemand gesehen hat.
    const neu = einStand({ ...angebotenMitDir, revision: 2, abdruck: 'neu' })
    const dienst = einDienst({}, {}, neu)
    belegeApi(dienst.api as never)

    const ausgang = await ziehe({
      stand: angebotenMitDir,
      schritt: 'zustimmen',
      bis: 'zustimmen',
      gegenueber: du,
    })

    expect(dienst.rufe).toEqual(['stand'])
    expect(ausgang.gegangen).toEqual([])
    expect(ausgang.gescheitert).toEqual({
      teil: 'veraltet',
      schritt: 'zustimmen',
      ursache: null,
    })
    expect(ausgang.stand).toBe(neu)
  })

  it.each([
    ['eine neue Stimme', { zustimmungen: [du, 'k-dritter'] }],
    ['ein anderer Zustand', { zustand: 'verhandlung' }],
    ['ein anderer nächster Schritt', { naechster: 'gegenangebot' }],
    ['eine andere Lage', { lage: 'wartet' }],
  ])('sieht %s als Änderung', async (_fall, teil) => {
    const dienst = einDienst({}, {}, einStand({ ...angebotenMitDir, ...teil }))
    belegeApi(dienst.api as never)

    const ausgang = await ziehe({
      stand: angebotenMitDir,
      schritt: 'zustimmen',
      bis: 'zustimmen',
      gegenueber: du,
    })

    expect(ausgang.gescheitert?.teil).toBe('veraltet')
  })

  it('geht nicht über den Schritt hinaus, den sein Knopf nennt', async () => {
    // **„Übergabe bestätigen" verbucht nicht**, auch wenn die Stimme des
    // Gegenübers inzwischen da ist und die Übergabe schließt.
    const { ausgang, rufe } = await zug(
      {
        uebergeben: einStand({
          zustand: 'uebergabeOffen',
          naechster: 'verbuchen',
          phase: 'verbuchen',
        }),
      },
      {
        stand: einStand({
          zustand: 'vorbereitet',
          beschluss: 'uebergeben',
          naechster: 'uebergeben',
        }),
        schritt: 'uebergeben',
        bis: 'uebergeben',
      },
    )

    expect(ausgang.gegangen).toEqual(['uebergeben'])
    expect(rufe).toEqual(['stand', 'uebergeben', 'ausgang', 'zustellen'])
  })

  it('geht nichts, wenn sich der Stand nicht lesen lässt', async () => {
    const dienst = einDienst({}, { stand: 'Der Bestand ist belegt.' })
    belegeApi(dienst.api as never)

    const ausgang = await ziehe({
      stand: angebotenMitDir,
      schritt: 'zustimmen',
      bis: 'zustimmen',
      gegenueber: du,
    })

    expect(dienst.rufe).toEqual(['stand'])
    expect(ausgang.gescheitert).toEqual(
      expect.objectContaining({ teil: 'lesen', schritt: '' }),
    )
  })
})
