// **Spielen an einem fremden Tisch** — Werkstück G5 von Phase 5.5
// (#1462).
//
// **Aus `partiegast.test.tsx` herausgelöst**, als sie an die Zeilengrenze aus
// §2 stieß. Der Schnitt folgt der Frage: dort geht es ums **Hinsetzen** —
// beitreten, nachsehen, aufstehen —, hier ums **Spielen**: was ankommt, was
// hinausgeht und was ein Gast ausdrücklich nicht darf.
import { screen, waitFor, within } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { einObjekt, einePartie, element } from './partieprobe'
import { stelle } from '../test/handlungsprobe'
import { ereignisTischbild, ereignisZug } from '../haken/usePartiegast'
import { ereignisNebenbei } from '../haken/usePartienebenbei'
import { kataloge } from '../sprache/kataloge'
import { effekteNachholen } from '../test/effekte'
import {
  gastApi,
  laufzeitbelegen,
  waehleWirt,
  zeigeGast,
} from './partiegastprobe'

const eigene = kataloge.de.seiten.spielen

// **Ein gemeldeter Zug steht im Verlauf.**
//
// **Neben dem Bild und nicht statt seiner**: das Bild sagt, wie der Tisch
// aussieht, der Zug, was geschehen ist. Ohne ihn gäbe es keinen Verlauf.
it('zeigt einen gemeldeten Zug', async () => {
  const horcher = laufzeitbelegen()

  gastApi({
    PartieBeitreten: vi.fn().mockResolvedValue({
      laeuft: true,
      partie: 'p1',
      gegenueber: 'abcd',
      fehler: null,
    }),
  })

  const hand = zeigeGast()

  await waehleWirt(hand)
  await hand.click(screen.getByRole('button', { name: eigene.gastBeitreten }))

  await waitFor(() => {
    horcher.get(ereignisTischbild)?.('p1', {
      partie: einePartie(),
      erlaubt: ['objekt.tappen'],
      handlungen: [],
    })
  })

  await waitFor(() => {
    horcher.get(ereignisZug)?.('p1', {
      folge: '1',
      art: 'objekt.tappen',
      oeffentlich: {},
      privat: {},
      kette: 'glied',
    })
  })

  // **Mit dem Wort der Handlung** (#1836) — und gesucht in der Liste, denn
  // „Tappen" heißt auch ein Knopf.
  const liste = element(
    (await screen.findByRole('heading', { name: eigene.ereignisse }))
      .parentElement,
  )

  expect(
    await within(liste).findByText(
      kataloge.de.seiten.handlungen.handlungsname['objekt.tappen'],
    ),
  ).toBeInTheDocument()
})

// **Nachsehen fragt den Stand, ohne etwas zu ändern.**
it('fragt den Stand beim Nachsehen', async () => {
  const gefragt = vi.fn().mockResolvedValue({
    laeuft: true,
    partie: 'p1',
    gegenueber: 'abcd-ef01',
    fehler: null,
  })

  gastApi({ PartieBeitrittStand: gefragt })

  const hand = zeigeGast()

  await hand.click(screen.getByRole('button', { name: eigene.gastNachsehen }))

  await waitFor(() => {
    expect(gefragt).toHaveBeenCalled()
  })

  expect(await screen.findByText('abcd-ef01')).toBeInTheDocument()
})

// **Ein Zug am Gasttisch geht als Wunsch hinaus.**
//
// **Und er bekommt kein Ergebnis zurück**: was daraus wird, entscheidet der
// Wirt und meldet es. Wer hier auf eine Antwort wartete, wartete auf etwas,
// das über eine andere Tür kommt.
it('schickt einen Zug als Wunsch hinaus', async () => {
  const horcher = laufzeitbelegen()

  const gewuenscht = vi.fn().mockResolvedValue({
    laeuft: true,
    partie: 'p1',
    gegenueber: 'abcd',
    fehler: null,
  })

  gastApi({
    PartieBeitreten: vi.fn().mockResolvedValue({
      laeuft: true,
      partie: 'p1',
      gegenueber: 'abcd',
      fehler: null,
    }),
    PartieWuenschen: gewuenscht,
    Spieltextur: vi.fn().mockResolvedValue({ adresse: '', fehler: null }),
    Rueckentextur: vi.fn().mockResolvedValue({ adresse: '', fehler: null }),
  })

  const hand = zeigeGast()

  await waehleWirt(hand)
  await hand.click(screen.getByRole('button', { name: eigene.gastBeitreten }))

  await waitFor(() => {
    horcher.get(ereignisTischbild)?.('p1', {
      partie: einePartie(),
      erlaubt: ['zone.blick'],
      handlungen: [],
    })
  })

  // **Die Zonenhandlung wird aufgeklappt und gestellt** — die Zone steht in
  // ihren Feldern (#1623),
  // und die stehen seit UX-08 erst auf Zuruf da.
  await screen.findByRole('button', {
    name: kataloge.de.seiten.handlungen.handlungsname['zone.blick'],
  })
  await stelle(hand, 'zone.blick')

  await waitFor(() => {
    expect(gewuenscht).toHaveBeenCalled()
  })

  // **Die Befehlskennung ist frisch** — sonst hielte der Wirt den zweiten
  // Wunsch für eine Wiederholung des ersten.
  const [kennung, art] = gewuenscht.mock.calls[0] as [string, string]

  expect(art).toBe('zone.blick')
  expect(kennung).not.toBe('')
})

// **Am fremden Tisch steht nichts, was dem Wirt gehört** — gefunden am
// nativen Fenster mit zwei Instanzen (Hüllenprobe `nat04`, S12.NAT-04).
//
// Bis dahin zeichnete der Gasttisch die Führung des Wirts mit: Pausieren,
// Fortsetzen und Beenden, die beim Gast nichts taten (hier stand ein Lauf,
// der genau das festhielt — eine Zusage ohne Einlösung, §10), „Tisch öffnen"
// für einen Tisch, den er nicht freigeben kann, und Rückschau und
// Protokollfreigabe, die den **eigenen** Bestand nach der fremden Partie
// fragten — und „Der gesuchte Eintrag wurde nicht gefunden" zeigten.
it('zeigt am fremden Tisch nichts, was dem Wirt gehört', async () => {
  const horcher = laufzeitbelegen()

  const gelesen = vi.fn()

  gastApi({
    PartieBeitreten: vi.fn().mockResolvedValue({
      laeuft: true,
      partie: 'p1',
      gegenueber: 'abcd',
      fehler: null,
    }),
    PartieVerlauf: gelesen,
    PartieOffenlegungen: gelesen,
    // **In der Gestalt der echten Antwort** (#2526): hier stand
    // `{ adresse: '' }`, und der Tisch las `textur.befunde` aus nichts — ein
    // Fehler, den die Verneinung unten nur nicht abwartete.
    Spieltextur: vi.fn().mockResolvedValue(eineTextur('karte')),
    Rueckentextur: vi.fn().mockResolvedValue(eineTextur('ruecken')),
    Fremdkartenbild: vi.fn().mockResolvedValue(eineTextur('fremd')),
  })

  const hand = zeigeGast()

  await waehleWirt(hand)
  await hand.click(screen.getByRole('button', { name: eigene.gastBeitreten }))

  await waitFor(() => {
    horcher.get(ereignisTischbild)?.('p1', {
      partie: einePartie(),
      erlaubt: [],
      handlungen: [],
    })
  })

  // **Erst der Tisch, dann die Verneinung** (§3).
  expect(await screen.findByRole('application')).toBeInTheDocument()

  for (const knopf of [
    eigene.pausieren,
    eigene.fortsetzen,
    eigene.beenden,
    eigene.freigabeStarten,
  ]) {
    expect(
      screen.queryByRole('button', { name: knopf }),
    ).not.toBeInTheDocument()
  }

  // **Und keine Sichtwahl** — ein Gast sitzt an seinem Sitz; die Wahl rief
  // `ansehen`, und das tut am fremden Tisch nichts.
  expect(screen.queryByLabelText(eigene.sicht)).not.toBeInTheDocument()

  expect(gelesen).not.toHaveBeenCalled()

  // **Gewartet wird auf die Wirkung** (§3): der Hinweis ohne beschleunigte
  // Grafik erscheint, wenn der Tisch zu zeichnen versucht hat, und erst
  // danach ist eine fehlende Meldung ein Beleg. Bis #2526 prüfte die Zeile
  // vorher — und ein Texturfehler kam erst danach.
  const hinweis = await screen.findByText(eigene.tischOhneWebgl)
  await effekteNachholen()
  expect(
    screen.getAllByRole('alert').filter((meldung) => meldung !== hinweis),
  ).toEqual([])
})

// **Ein Satz kommt an und steht da, ohne den Tisch zu ändern.**
//
// **Beides in einem Lauf**, denn beides gehört zusammen: ein Chat, der den
// Stand änderte, wäre ein Fehler, und einer, den niemand sieht, wäre keiner
// — bis zum 2026-09-16 sammelte der Haken ihn in eine Liste, die kein
// Bauteil las.
it('zeigt einen Satz, ohne den Tisch zu ändern', async () => {
  const horcher = laufzeitbelegen()

  gastApi({
    PartieBeitreten: vi.fn().mockResolvedValue({
      laeuft: true,
      partie: 'p1',
      gegenueber: 'abcd',
      fehler: null,
    }),
  })

  const hand = zeigeGast()

  await waehleWirt(hand)
  await hand.click(screen.getByRole('button', { name: eigene.gastBeitreten }))

  const melde = horcher.get(ereignisNebenbei)
  expect(melde).toBeDefined()

  // **Erst sitzen, dann reden** (§3, auf die Wirkung): das Gespräch nimmt
  // nur, was zur Partie des Platzes gehört, und die kennt der Gast hier aus
  // dem Stand des Beitritts. Ein Satz davor ginge verloren — unter Last am
  // 2026-10-04 in `make ci` gesehen.
  expect(
    await screen.findByText(eigene.gastSitzt, { exact: false }),
  ).toBeInTheDocument()

  await waitFor(() => {
    melde?.('p1', {
      art: 'chat',
      von: 'anna',
      text: 'hallo',
      zeigerart: '',
      objekt: '',
      x: 0,
      y: 0,
    })
  })

  expect(await screen.findByText('hallo')).toBeInTheDocument()
  expect(screen.getByText('anna')).toBeInTheDocument()

  // **Der Tisch bleibt leer** — ein Chat ändert den Stand nicht.
  expect(screen.getByText(eigene.gastOhneTisch)).toBeInTheDocument()
})

// **Der fremde Tisch fragt den Wirt nach dem Bild** — P9 (#2377), Vermerk
// 0198. Bis dahin sagte er nur, warum die Kartenbilder fehlen: die Kennungen
// am Tisch waren die des Wirts. Jetzt trägt eine sichtbare Karte eine
// Bildkennung, und das Fenster fragt mit ihr — **über die Liste**, denn unter
// jsdom zeichnet die Leinwand nicht. Hier steht, dass `fremd` bis zum Haken
// reicht; welche Wünsche wohin gehen, hält `haken/fremdetexturen.test.tsx`.
it('fragt am fremden Tisch den Wirt nach dem Bild einer sichtbaren Karte', async () => {
  const horcher = laufzeitbelegen()
  const fremd = vi.fn().mockResolvedValue({
    textur: {
      schluessel: 'fremd-b1',
      adresse: '/spielbild/fremd-b1.png',
      guete: 'tisch',
      seite: 'vorne',
      breite: 2,
      hoehe: 3,
      befunde: [],
      zweiseitig: false,
    },
    fehler: null,
  })
  gastApi({
    PartieBeitreten: vi.fn().mockResolvedValue({
      laeuft: true,
      partie: 'p1',
      gegenueber: 'abcd',
      fehler: null,
    }),
    Rueckentextur: vi.fn().mockResolvedValue({ adresse: '', fehler: null }),
    Fremdkartenbild: fremd,
  })

  const hand = zeigeGast()

  await waehleWirt(hand)
  await hand.click(screen.getByRole('button', { name: eigene.gastBeitreten }))

  await waitFor(() => {
    horcher.get(ereignisTischbild)?.('p1', {
      partie: einePartie({
        objekte: [
          einObjekt({
            kennung: 'o1',
            name: 'Wald',
            zone: 'tisch',
            karteID: '',
            bild: 'b1',
            geheim: false,
          }),
        ],
      }),
      erlaubt: [],
      handlungen: [],
    })
  })

  await hand.click(
    await screen.findByRole('button', { name: eigene.tischListe }),
  )

  await waitFor(() => {
    expect(fremd).toHaveBeenCalledWith(
      'p1',
      'o1',
      'b1',
      'vorne',
      expect.any(String),
    )
  })
})

// eineTextur ist eine Antwort auf einen Bildwunsch des Tisches.
function eineTextur(schluessel: string) {
  return {
    textur: {
      schluessel,
      adresse: `/spielbild/${schluessel}.png`,
      guete: 'tisch',
      seite: 'vorne',
      breite: 2,
      hoehe: 3,
      befunde: [],
      zweiseitig: false,
    },
    fehler: null,
  }
}
