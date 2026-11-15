// **Der Bereich Spielen am Fenster** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Wogegen hier geprüft wird
//
// **Dass das Fenster den Spielkern wirklich erreicht.** Ein Wächter zählt,
// dass eine Bindung gerufen wird (`internal/werkzeuge/bindungsaufrufer_test.go`);
// dass beim Klick das Richtige hinausgeht, steht nur hier.
//
// **Und dass ein Geheimnis eines bleibt.** Die Bibliothek ist auch für ihren
// Eigner verdeckt — am Fenster heißt das: da steht kein Kartenname, auch
// nicht klein, auch nicht in einem Titel.
import { screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import { einDeck, element, eineAntwort, einObjekt } from './partieprobe'
import { mitDecks, steigeEin, zeige } from './spielenprobe'

const eigene = kataloge.de.seiten.spielen

afterEach(() => {
  raeumeApi()
})

// **Ohne ein Deck gibt es keinen Tisch — und der Leerzustand sagt, warum.**
//
// **Eines genügt**: dasselbe Deck darf an zwei Sitzen liegen, und ein Sitz
// allein ist eine Spielart (Goldfish). Hier stand bis zum 2026-09-24 „zwei
// Decks" — eine Schranke aus dem ersten Werkstück (G1), die niemand mehr
// begründete; die Probepartie der WebKit-Probe spielt genau so.
it('erklärt, dass es ein Deck braucht', async () => {
  belegeApi({
    Decks: () => Promise.resolve({ decks: [], fehler: null }),
    Formate: () => Promise.resolve({ formate: [], fehler: null }),
  })

  zeige()

  expect(await screen.findByText(eigene.leerDecks)).toBeInTheDocument()

  // **Und er heißt nicht wie die Seite** — die heißt seit dem 2026-09-24
  // „Eine Partie anfangen", und zweimal dasselbe sagte nicht, was fehlt.
  expect(screen.getByText(eigene.leerDecksTitel)).toBeInTheDocument()
})

// **Und der nächste Schritt ist ein Knopf, kein Weg zum Merken** — UXR-13
// (#1921). Am Probebestand hatte die Seite keinen Hauptknopf (R1 in
// `docs/UIREVIEW.md`): der Leerzustand sagte „Lege eines unter Decks an",
// und wer es tun wollte, suchte die Decks in der Leiste.
it('führt ohne Deck zu den Decks', async () => {
  belegeApi({
    Decks: () => Promise.resolve({ decks: [], fehler: null }),
    Formate: () => Promise.resolve({ formate: [], fehler: null }),
  })
  const wechseln = vi.fn()

  const hand = zeige(wechseln)

  const knopf = await screen.findByRole('button', {
    name: eigene.leerDecksZiel,
  })

  // **Er führt** (R1): Hauptgewicht, und auf der Seite ohne Deck kein
  // zweiter.
  expect(knopf).toHaveClass('bg-marke')

  await hand.click(knopf)

  expect(wechseln).toHaveBeenCalledWith('decks')
})

// **Ein einziges Deck reicht für zwei Sitze.**
it('fängt mit einem einzigen Deck an', async () => {
  const angelegt = vi.fn(() => Promise.resolve(eineAntwort()))

  mitDecks({
    Decks: () =>
      Promise.resolve({ decks: [einDeck('1', 'Aggro')], fehler: null }),
    PartieAnlegen: angelegt,
  })

  const hand = zeige()

  // **Die Namen stehen vorbelegt da** (P6, #2374) — sie werden ersetzt.
  const erster = await screen.findByLabelText(`${eigene.sitzName} 1`)
  const zweiter = screen.getByLabelText(`${eigene.sitzName} 2`)
  await hand.clear(erster)
  await hand.type(erster, 'Anna')
  await hand.clear(zweiter)
  await hand.type(zweiter, 'Bert')

  for (const feld of screen.getAllByLabelText(eigene.sitzDeck)) {
    await hand.selectOptions(feld, '1')
  }

  await hand.click(screen.getByRole('button', { name: eigene.anfangen }))

  await waitFor(() => {
    expect(angelegt).toHaveBeenCalled()
  })
})

// **Was hinausgeht, ist der Entwurf — mit den gewählten Decks.**
//
// **Geprüft wird das Argument und nicht nur der Ruf** (§10): ein Formular,
// das jeden Sitz auf dasselbe Deck stellte, riefe genauso oft.
it('schickt die gewählten Sitze an den Kern', async () => {
  const angelegt = vi.fn(() => Promise.resolve(eineAntwort()))

  mitDecks({ PartieAnlegen: angelegt })

  const hand = zeige()

  // **Die Namen stehen vorbelegt da** (P6, #2374) — sie werden ersetzt.
  const erster = await screen.findByLabelText(`${eigene.sitzName} 1`)
  const zweiter = screen.getByLabelText(`${eigene.sitzName} 2`)
  await hand.clear(erster)
  await hand.type(erster, 'Anna')
  await hand.clear(zweiter)
  await hand.type(zweiter, 'Bert')

  const felder = screen.getAllByLabelText(eigene.sitzDeck)

  await hand.selectOptions(element(felder[0]), '1')
  await hand.selectOptions(element(felder[1]), '2')

  await hand.click(screen.getByRole('button', { name: eigene.anfangen }))

  await waitFor(() => {
    expect(angelegt).toHaveBeenCalledWith({
      spiel: 'magic',
      regelwerk: { kennung: '', fassung: 0, strenge: '', abgewaehlt: [] },
      wirt: '',
      sitze: [
        {
          kennung: 'sitz1',
          name: 'Anna',
          deckID: '1',
          leiheID: '',
          profilID: '',
          eingeladen: '',
        },
        {
          kennung: 'sitz2',
          name: 'Bert',
          deckID: '2',
          leiheID: '',
          profilID: '',
          eingeladen: '',
        },
      ],
    })
  })
})

// **Ein dritter Sitz kommt hinzu und geht wieder.**
it('nimmt einen dritten Sitz auf und wieder weg', async () => {
  mitDecks()

  const hand = zeige()

  await hand.click(
    await screen.findByRole('button', { name: eigene.sitzHinzu }),
  )

  expect(screen.getAllByLabelText(eigene.sitzDeck)).toHaveLength(3)

  const wegnehmen = screen.getAllByRole('button', { name: eigene.sitzWeg })

  await hand.click(element(wegnehmen[2]))

  expect(screen.getAllByLabelText(eigene.sitzDeck)).toHaveLength(2)

  // **Bis auf einen herunter** — Werkstück G3 (#1460): allein üben ist eine
  // Spielart und kein Fehler. **CR 100.1** (Fassung 2026-08-19) sagt dazu,
  // dass die Regeln für zwei oder mehr Spieler gelten; ein Goldfish ist damit
  // ein Aufbau und kein Spiel, und genau so steht es im Assistenten.
  //
  // **Hier stand bis zum 2026-09-17 das Gegenteil** — „bei zweien gibt es
  // keinen Wegnehmen-Knopf mehr" —, und damit war Solo Goldfish nicht
  // erreichbar.
  await hand.click(
    element(screen.getAllByRole('button', { name: eigene.sitzWeg })[1]),
  )

  expect(screen.getAllByLabelText(eigene.sitzDeck)).toHaveLength(1)

  // **Und beim letzten ist Schluss**: ohne Sitz gibt es keinen Aufbau.
  expect(
    screen.queryByRole('button', { name: eigene.sitzWeg }),
  ).not.toBeInTheDocument()
})

// **Und die Meldung des Einsteigens steht am Kopf und nicht am Anfangsknopf.**
//
// # Die Gegenprobe zum Lauf darunter
//
// **Ohne sie sagte er wenig** (§10): dass eine Meldung „beim Knopf" steht,
// ist erst dann eine Aussage, wenn eine **andere** Meldung woanders steht.
// Sonst hielte auch ein Fenster her, das jede Meldung an denselben Ort legt
// — genau der Zustand, den der Befund vom 2026-09-17 angestrichen hat.
it('legt die Meldung des Einsteigens an den Seitenkopf', async () => {
  mitDecks({
    Partien: () =>
      Promise.resolve({
        partien: [
          {
            id: 'p1',
            fassung: '3',
            spiel: 'magic',
            stand: 'laeuft',
            sitze: [
              { kennung: 'sitz1', name: 'Anna' },
              { kennung: 'sitz2', name: 'Bert' },
            ],
            objekte: 12,
            zug: 2,
            eigenerSitz: '',
            dran: false,
          },
        ],
        fehler: null,
      }),
    Partie: () =>
      Promise.resolve({
        partie: eineAntwort().partie,
        erlaubt: [],
        befunde: [],
        fehler: {
          code: 'nichtGefunden',
          schluessel: '',
          meldung: 'Diese Partie gibt es nicht mehr.',
          feld: '',
        },
      }),
  })

  const hand = zeige()

  await steigeEin(hand, 'Anna · Bert', 'Anna')

  const meldung = await screen.findByRole('alert')

  expect(meldung).toHaveTextContent('Diese Partie gibt es nicht mehr.')

  // **Und sie steht gerade nicht beim Anfangsknopf** — sie gehört nicht dazu.
  expect(
    screen.getByRole('button', { name: eigene.anfangen }).closest('section'),
  ).not.toContainElement(meldung)
})

// **Ein Fehler beim Anlegen steht als Meldung da und nicht als leerer Tisch.**
it('meldet, was beim Anlegen schiefging', async () => {
  mitDecks({
    PartieAnlegen: () =>
      Promise.resolve({
        partie: eineAntwort().partie,
        erlaubt: [],
        fehler: {
          code: 'validierung',
          schluessel: '',
          meldung: 'Das Deck fehlt.',
          feld: '',
        },
      }),
  })

  const hand = zeige()

  // **Die Namen stehen vorbelegt da** (P6, #2374) — sie werden ersetzt.
  const erster = await screen.findByLabelText(`${eigene.sitzName} 1`)
  const zweiter = screen.getByLabelText(`${eigene.sitzName} 2`)
  await hand.clear(erster)
  await hand.type(erster, 'Anna')
  await hand.clear(zweiter)
  await hand.type(zweiter, 'Bert')

  const felder = screen.getAllByLabelText(eigene.sitzDeck)

  await hand.selectOptions(element(felder[0]), '1')
  await hand.selectOptions(element(felder[1]), '2')

  await hand.click(screen.getByRole('button', { name: eigene.anfangen }))

  const meldung = await screen.findByRole('alert')

  expect(meldung).toHaveTextContent('Das Deck fehlt.')

  // **Und sie steht beim Knopf, der sie ausgelöst hat** — CLAUDE.md §6:
  // „Fehler erscheinen am verursachenden Feld, nicht als anonymer Hinweis."
  //
  // **Der Befund, für den diese Zusicherung dasteht**: bis zum 2026-09-17
  // stand sie am **Seitenkopf**, und dazwischen lag die ganze Regelschau aus
  // G4 — an der laufenden Anwendung in einem 370 Pixel breiten Fenster
  // **fünf Bildschirmhöhen**. Der Lauf darüber war grün: er fragte nur, ob
  // **irgendwo** eine Meldung steht. Gesehen hat es die Sichtprüfung (§3).
  expect(
    screen.getByRole('button', { name: eigene.anfangen }).closest('section'),
  ).toContainElement(meldung)

  // **Und der Tisch bleibt zu** — eine gescheiterte Anlage hinterlässt keine
  // halbe Partie.
  expect(
    screen.queryByRole('heading', { name: eigene.tischTitel }),
  ).not.toBeInTheDocument()
})

// **Die Bibliothek trägt am Fenster keinen Kartennamen.**
//
// **Auch nicht für ihren Eigner** — das ist die Zusage, an der am meisten
// hängt, und sie ist am Fenster als Erstes zu brechen.
it('zeigt eine verdeckte Karte ohne ihren Namen', async () => {
  mitDecks({
    PartieAnlegen: () =>
      Promise.resolve(
        eineAntwort({
          objekte: [einObjekt({ karte: '', name: '', geheim: true })],
        }),
      ),
  })

  const hand = zeige()

  // **Die Namen stehen vorbelegt da** (P6, #2374) — sie werden ersetzt.
  const erster = await screen.findByLabelText(`${eigene.sitzName} 1`)
  const zweiter = screen.getByLabelText(`${eigene.sitzName} 2`)
  await hand.clear(erster)
  await hand.type(erster, 'Anna')
  await hand.clear(zweiter)
  await hand.type(zweiter, 'Bert')

  const felder = screen.getAllByLabelText(eigene.sitzDeck)

  await hand.selectOptions(element(felder[0]), '1')
  await hand.selectOptions(element(felder[1]), '2')

  await hand.click(screen.getByRole('button', { name: eigene.anfangen }))

  // **Die Zonenliste ist seit G3 die zweite Darstellung** (#1460): der Tisch
  // steht vorn, und beide zugleich legten jede Karte zweimal in die
  // Tabulatorreihenfolge.
  await hand.click(
    await screen.findByRole('button', { name: eigene.tischListe }),
  )

  expect(
    await screen.findByRole('button', { name: eigene.geheim }),
  ).toBeInTheDocument()

  expect(screen.queryByText('Wald')).not.toBeInTheDocument()

  await antwortAbwarten()
})
