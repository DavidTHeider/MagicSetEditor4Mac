// **Eine Deckliste wird ein Deck** — Werkstück T8 von Phase 5.6
// (#1473).
//
// # Warum es zwei Schritte sind und nicht einer
//
// **Weil niemand eine Liste einfügt und danach wissen will, was daraus wurde.**
// Die Vorschau schreibt nichts; sie sagt, welche Zeile welchen Druck meint und
// welche offen bleibt. Erst das Übernehmen legt.
//
// # Und warum der Stand mitgeht
//
// **Weil zwischen beidem ein Mensch liegt** (§7, Regel 2). Wer in dieser Zeit
// eine Karte ins Deck legt, hat ein anderes Deck vor sich als das, über das
// entschieden wurde — dann wird nichts gelegt, und es steht da.
//
// **Derselbe Riegel macht den zweiten Klick harmlos**: er geht gegen einen
// Stand, den es nicht mehr gibt, und legt nichts doppelt.
import { useRef, useState } from 'react'

import { Knopf } from './Knopf'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { derSideboardbereich } from '../apideckeinfuhr'

import type { DeckDTO } from '../apidecks'
import type {
  Deckeinfuhrbericht,
  Deckeinfuhrzeile,
  Deckvorschau,
} from '../apideckeinfuhr'

interface Deckeinfuhreigenschaften {
  // deck ist das offene Deck.
  //
  // **Nicht wahlfrei**: ohne offenes Deck gibt es nichts zu füllen, und die
  // Seite zeigt den Kasten dann gar nicht. Ein Zweig „kein Deck" hier wäre
  // einer, den im Betrieb nie jemand betritt (§10).
  deck: DeckDTO

  // gelegt läuft, nachdem Zeilen ins Deck gekommen sind.
  gelegt: () => void
}

export function Deckeinfuhr({ deck, gelegt }: Deckeinfuhreigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.decks
  const lauf = useVorgang()

  const [text, setzeText] = useState('')
  const [fehler, setzeFehler] = useState<string>()
  const [bericht, setzeBericht] = useState<Deckeinfuhrbericht | null>(null)
  const [vorschau, setzeVorschau] = useState<Deckvorschau | null>(null)

  // textstand zählt die Wechsel des Textes mit (§7, Regel 2): eine Vorschau
  // — oder ein Fehler — zu einem Text, der inzwischen ein anderer ist, wird
  // verworfen. Angestrichen im Nachlauf über Phase 6.6: `onChange` räumte nur
  // die **sichtbare** Vorschau, und eine unterwegs setzte sich danach.
  const textstand = useRef(0)

  const kennzeichen = 'deckeinfuhr'

  const ansehen = (): Promise<void> =>
    lauf.starte(kennzeichen, async () => {
      setzeFehler(undefined)
      setzeBericht(null)

      if (text.trim() === '') {
        setzeFehler(eigene.einfuhrLeer)

        return
      }

      const stand = textstand.current

      try {
        const antwort = await api().DecklisteVorschau(Number(deck.id), text)
        const angekommen = auspacken(antwort, () => antwort.vorschau)

        if (textstand.current === stand) {
          setzeVorschau(angekommen)
        }
      } catch (ursache) {
        if (textstand.current === stand) {
          setzeVorschau(null)
          setzeFehler(meldungAus(ursache, texte))
        }
      }
    })

  // **Die Vorschau kommt als Argument und nicht aus dem Zustand.** Der Knopf
  // steht erst da, wenn es eine gibt; ein Riegel hier wäre ein zweiter Zweig
  // für dieselbe Bedingung — und einer, den niemand betreten kann (§10).
  const uebernehmen = (gesehen: Deckvorschau): Promise<void> =>
    lauf.starte(kennzeichen, async () => {
      setzeFehler(undefined)
      setzeBericht(null)

      const zeilen = bestaetigte(gesehen)

      if (zeilen.length === 0) {
        setzeFehler(eigene.einfuhrNichtsOffen)

        return
      }

      try {
        const antwort = await api().DecklisteUebernehmen(
          Number(deck.id),
          gesehen.stand,
          zeilen,
        )

        setzeBericht(auspacken(antwort, () => antwort.bericht))

        // **Die Vorschau ist danach verbraucht** — ihr Stand gilt nicht mehr,
        // und ein zweiter Klick ginge gegen ein Deck von vorhin.
        setzeVorschau(null)
        gelegt()
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  return (
    <div className="flex flex-col gap-2">
      <Erklaerungsabsatz
        text={eigene.einfuhrErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      <label className="flex flex-col gap-1 text-sm">
        {eigene.einfuhrFeld}
        <textarea
          rows={6}
          value={text}
          className="rounded-sm border border-kante bg-transparent p-2 font-mono text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          onChange={(ereignis) => {
            setzeText(ereignis.target.value)
            textstand.current += 1

            // **Ein geänderter Text macht die Vorschau ungültig.** Sie gehört
            // zu dem, was dastand — nicht zu dem, was jetzt dasteht (§7,
            // Regel 1).
            setzeVorschau(null)
          }}
        />
      </label>

      <div className="flex flex-wrap gap-2">
        <Knopf
          gewicht="leise"
          angehalten={lauf.laeuft(kennzeichen)}
          aufKlick={() => {
            void ansehen()
          }}
        >
          {eigene.einfuhrVorschau}
        </Knopf>

        {vorschau !== null && (
          <Knopf
            angehalten={lauf.laeuft(kennzeichen)}
            aufKlick={() => {
              void uebernehmen(vorschau)
            }}
          >
            {eigene.einfuhrUebernehmen}
          </Knopf>
        )}
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {bericht !== null && (
        <p role="status" className="text-sm text-leise">
          {eigene.einfuhrGelegt}: {bericht.gelegt} — {bericht.stueck}{' '}
          {eigene.einfuhrStueck}
        </p>
      )}

      {vorschau !== null && <Vorschauliste vorschau={vorschau} />}
    </div>
  )
}

// Vorschauliste zeigt, was ins Deck käme — und was offen bleibt.
function Vorschauliste({ vorschau }: { vorschau: Deckvorschau }) {
  const eigene = useTexte().seiten.decks

  const offen = vorschau.zeilen.filter((zeile) => !zeile.uebernehmbar)

  return (
    <div className="flex flex-col gap-1 text-sm">
      <p role="status" className="text-leise">
        {eigene.einfuhrBilanz}: {vorschau.hauptdeck} / {vorschau.sideboard} —{' '}
        {offen.length} {eigene.einfuhrOffen}
      </p>

      <ul className="flex flex-col gap-0.5 font-mono text-xs">
        {vorschau.zeilen.map((zeile, stelle) => (
          <li
            // **Die Stelle ist der Schlüssel**, weil eine Liste dieselbe
            // Karte in zwei Bereichen nennen darf: Name und Druck sind
            // zusammen nicht eindeutig.
            key={`${String(stelle)}-${zeile.name}`}
            className={zeile.uebernehmbar ? '' : 'text-fehler'}
          >
            {zeile.uebernehmbar
              ? `${bereichsname(zeile.bereich, eigene)} · ${String(
                  zeile.menge,
                )} · ${zeile.name}`
              : `${zeile.name} — ${zeile.befund}`}
          </li>
        ))}
      </ul>

      {vorschau.befunde.map((befund) => (
        <p key={befund} className="text-xs text-fehler">
          {befund}
        </p>
      ))}
    </div>
  )
}

// bereichsname übersetzt den Bereich, den die Naht schickt.
//
// **Übersetzt und nicht durchgereicht**: `hauptdeck` steht so im Bestand und
// in jeder Fläche — auf dem Bildschirm hat es nichts zu suchen (§9).
function bereichsname(
  bereich: string,
  eigene: { hauptdeck: string; sideboard: string },
): string {
  return bereich === derSideboardbereich ? eigene.sideboard : eigene.hauptdeck
}

// bestaetigte macht aus der Vorschau, was gelegt wird.
function bestaetigte(vorschau: Deckvorschau): Deckeinfuhrzeile[] {
  return vorschau.zeilen
    .filter((zeile) => zeile.uebernehmbar)
    .map((zeile) => ({
      printing: zeile.printing,
      menge: zeile.menge,
      bereich: zeile.bereich,
    }))
}
