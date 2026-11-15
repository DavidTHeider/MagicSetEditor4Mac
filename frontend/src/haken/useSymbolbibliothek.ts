// **Was die Symbolbibliothek tut** — DOM-02
// (#1697).
//
// Der Haken hält Suche, Liste, die offene Löschfrage und die Befunde; die
// Seite zeichnet sie (`seiten/Symbolbibliothek.tsx`). Was ein verwendetes
// Symbol ist, sagt Go — hier wird nur gefragt und gezeigt.
//
// # Die vier Regeln aus §7
//
// **Regel 1 und 2 löst der Schlüssel** der Seite: ein anderes Set baut sie
// neu, und eine Antwort für das vorige trifft keinen Zustand mehr. Die
// Bibliothek selbst gehört dem Profil und nicht dem Set. **Regel 3**: neu
// gelesen wird nach einem geglückten Schreiben, nach einem gescheiterten
// nicht. **Regel 4**: ein Vorgang je Symbol, und je einer für Aufnehmen und
// Einlesen.
import { useState } from 'react'

import { api } from '../api'
import type { BibliothekssymbolDTO } from '../apisymbolbibliothek'
import { feldbefund, meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { useLaden, type Ladezustand } from './useLaden'
import { useVorgang } from './useVorgang'

// Befund ist eine Meldung und der Ort, an dem sie steht — ein Symbol, das
// Aufnehmen oder das Einlesen.
export interface Befund {
  wo: string
  text: string

  // amNamen sagt, dass sie zum Namensfeld dieses Ortes gehört (§6).
  amNamen: boolean
}

export interface Symbolbibliothekstand {
  suche: string
  setzeSuche: (suche: string) => void
  symbole: Ladezustand<BibliothekssymbolDTO[]>

  name: string
  setzeName: (name: string) => void

  befund: Befund | undefined
  hinweis: string | undefined

  // frage ist das Symbol, dessen Löschen gerade gefragt wird.
  frage: BibliothekssymbolDTO | undefined
  fragen: (symbol: BibliothekssymbolDTO | undefined) => void

  laeuft: (wo: string) => boolean
  aufnehmen: () => Promise<void>
  einlesen: () => Promise<void>
  einsetzen: (symbol: BibliothekssymbolDTO) => Promise<void>
  kopieren: (symbol: BibliothekssymbolDTO) => Promise<void>
  ausgeben: (symbol: BibliothekssymbolDTO) => Promise<void>
  umbenennen: (symbol: BibliothekssymbolDTO, name: string) => Promise<boolean>
  zuruecknehmen: (symbol: BibliothekssymbolDTO) => Promise<void>
  loeschen: (symbol: BibliothekssymbolDTO) => Promise<void>
}

// aufnahme und einlesung sind die Orte der beiden Wege ohne Symbol.
export const aufnahme = 'aufnehmen'
export const einlesung = 'einlesen'

// Bibliotheksanschluss ist, was das Fach der Bibliothek sagt und von ihr
// hört.
export interface Bibliotheksanschluss {
  // eingesetzt sagt, dass das Set ein anderes Symbol trägt.
  eingesetzt: () => void

  // zurueckgeholt sagt, dass ein Symbol seine vorige Fassung zurückbekam —
  // der Editor liest es dann neu, wenn er es gerade bearbeitet.
  zurueckgeholt?: ((id: string) => void) | undefined

  // runde zählt, wie oft von außen nachzulesen ist — nach dem Sichern im
  // Editor und nach dem Nachziehen.
  runde?: number | undefined
}

export function useSymbolbibliothek(
  setID: string,
  { eingesetzt, zurueckgeholt, runde = 0 }: Bibliotheksanschluss,
): Symbolbibliothekstand {
  const texte = useTexte()
  const eigene = texte.seiten.karten.symbolbibliothek
  const vorgang = useVorgang()

  const [suche, setzeSuche] = useState('')
  const [name, setzeName] = useState('')
  const [befund, setzeBefund] = useState<Befund>()
  const [hinweis, setzeHinweis] = useState<string>()
  const [frage, fragen] = useState<BibliothekssymbolDTO>()

  // **Mit `behalten`**: beim Tippen bleibt die vorige Liste stehen, bis die
  // neue da ist, statt bei jedem Buchstaben zu verschwinden.
  const symbole = useLaden(
    async () => {
      const antwort = await api().Symbolbibliothek(suche)

      return auspacken(antwort, () => antwort.symbole)
    },
    [suche, runde],
    true,
  )

  // schreiben führt einen Weg aus, meldet ihn und liest nach — **nur nach
  // einem geglückten** (§7, Regel 3).
  const schreiben = async (
    wo: string,
    weg: () => Promise<string | undefined>,
  ) => {
    await vorgang.starte(wo, async () => {
      try {
        const erfolg = await weg()
        setzeBefund(undefined)
        setzeHinweis(erfolg)
        symbole.nachladen()
      } catch (ursache: unknown) {
        setzeHinweis(undefined)
        setzeBefund({
          wo,
          text: meldungAus(ursache, texte),
          amNamen: feldbefund(ursache, 'name', texte) !== undefined,
        })
      }
    })
  }

  return {
    suche,
    setzeSuche,
    symbole,
    name,
    setzeName,
    befund,
    hinweis,
    frage,
    fragen,
    laeuft: vorgang.laeuft,

    aufnehmen: () =>
      schreiben(aufnahme, async () => {
        const antwort = await api().BibliothekssymbolAufnehmen(setID, name)
        const aufgenommen = auspacken(antwort, () => antwort.name)
        setzeName('')

        return einsetzen(eigene.aufgenommen, { name: aufgenommen })
      }),

    // **Ein weggeklickter Dialog ist kein Fehler** — und auch kein Anlass,
    // eine Meldung von vorhin zu löschen.
    einlesen: () =>
      schreiben(einlesung, async () => {
        const antwort = await api().BibliothekssymbolEinlesen()
        const gelesen = auspacken(antwort, () => antwort)

        return gelesen.abgebrochen
          ? undefined
          : einsetzen(eigene.eingelesen, { name: gelesen.name })
      }),

    einsetzen: (symbol) =>
      schreiben(symbol.id, async () => {
        const antwort = await api().BibliothekssymbolEinsetzen(symbol.id, setID)
        auspacken(antwort, () => antwort.datei)
        eingesetzt()

        return einsetzen(eigene.eingesetzt, { name: symbol.name })
      }),

    kopieren: (symbol) =>
      schreiben(symbol.id, async () => {
        const antwort = await api().BibliothekssymbolKopieren(
          symbol.id,
          einsetzen(eigene.kopieName, { name: symbol.name }),
        )
        const kopie = auspacken(antwort, () => antwort.name)

        return einsetzen(eigene.aufgenommen, { name: kopie })
      }),

    ausgeben: (symbol) =>
      schreiben(symbol.id, async () => {
        const antwort = await api().BibliothekssymbolAusgeben(
          symbol.id,
          symbol.name,
        )
        const ausgegeben = auspacken(antwort, () => antwort)

        return ausgegeben.abgebrochen
          ? undefined
          : einsetzen(eigene.ausgegeben, { pfad: ausgegeben.pfad })
      }),

    // umbenennen sagt, ob es geglückt ist — die Zeile schließt ihr Feld
    // nur dann.
    umbenennen: async (symbol, neu) => {
      let geglueckt = false

      await schreiben(symbol.id, async () => {
        const antwort = await api().BibliothekssymbolUmbenennen(symbol.id, neu)
        auspacken(antwort, () => antwort.name)
        geglueckt = true

        return undefined
      })

      return geglueckt
    },

    zuruecknehmen: (symbol) =>
      schreiben(symbol.id, async () => {
        const antwort = await api().BibliothekssymbolZuruecknehmen(symbol.id)
        auspacken(antwort, () => antwort.veraltet)
        zurueckgeholt?.(symbol.id)

        return einsetzen(eigene.zurueckgeholt, { name: symbol.name })
      }),

    // **Bestätigt wird hier und nicht in Go** — die Frage hat die Folgen
    // schon genannt, und dieselben nennt die Liste.
    loeschen: (symbol) =>
      schreiben(symbol.id, async () => {
        const antwort = await api().BibliothekssymbolLoeschen(symbol.id, true)
        auspacken(antwort, () => antwort.sets)
        fragen(undefined)

        return undefined
      }),
  }
}
