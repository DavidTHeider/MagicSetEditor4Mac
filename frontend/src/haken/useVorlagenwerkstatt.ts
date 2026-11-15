// **Was die Vorlagenwerkstatt tut** — Werkstück U7 von Phase 5.1
// (#1418).
//
// `Vorlagen.tsx` zeigt; dieser Haken entscheidet (§2). Er hält die Wahl der
// Datei, die Rückfrage vor dem Verwerfen, die Testkarte, den Zeiger und den
// Sprung — alles, was Navigator, Zentrum und Inspector gemeinsam brauchen.
//
// # Die vier Regeln, je Zustand
//
// - **Die Wahl** räumt beim Wechsel, was an ihr hängt: Puffer, Befund,
//   Testkarte und Zeiger sind über den Schlüssel der Datei oder des Pakets
//   geführt und gelten nur, solange er stimmt (Regel 1).
// - **Antworten** tragen sich über `useLaden` und den Puffer nur ein, wenn
//   ihr Stand noch der aktuelle ist (Regel 2).
// - **Geschrieben ist geschrieben** — der Puffer lädt nach einem Schreiben
//   nicht nach; nur der Symbolindex liest neu, denn Definitionen können
//   dazugekommen sein (Regel 3).
// - **Ein Schreibvorgang je Datei** (Regel 4) hält der Puffer.
//
// # Der Sprung in eine andere Datei
//
// „Zur Definition" kann in eine Datei führen, die noch nicht offen ist. Der
// Sprung wird dann **vorgemerkt** und ausgeführt, sobald die Datei steht —
// beim Zeichnen, hinter einem Vergleich, wie die React-Dokumentation es für
// abgeleiteten Zustand vorsieht. Ein Effekt liefe dem Feld eine Runde
// hinterher und spränge in einen Text, der noch der alte ist.
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { symbolAnStelle } from '../bausteine/symbolstelle'
import { useLaden } from './useLaden'
import { usePufferbefund } from './usePufferbefund'
import { useVerzoegert } from './useVerzoegert'
import { useVorlagenpuffer } from './useVorlagenpuffer'
import {
  usePaketsymbole,
  useProbekarten,
  useSymbolverwendungen,
} from './useWerkstattquellen'

import type {
  PaketDTO,
  PaketsymbolDTO,
  ProbekarteDTO,
  Skriptstelle,
  SymbolverwendungDTO,
} from '../api'
import type { Sprungziel } from '../bausteine/quelltextsprung'
import type { Symbolstelle } from '../bausteine/symbolstelle'
import type { Pufferbefund } from './usePufferbefund'
import type { Vorlagenpuffer } from './useVorlagenpuffer'

export interface Dateiwahl {
  paket: string
  datei: string

  // beiwerk sagt, dass die Datei kein Text ist (FMT-02, #1691).
  beiwerk?: boolean
}

// Sprung ist ein vorgemerkter Sprung in eine Datei, die erst noch aufgeht.
interface Sprung {
  datei: string
  zeile: number
  spalte: number
}

// verwendungsfrist ist die Ruhe des Zeigers, bevor Verwendungen gesucht
// werden — kurz genug, dass es wie eine Antwort wirkt, lang genug, dass ein
// Wort zu Ende getippt ist.
const verwendungsfrist = 300

export interface Vorlagenwerkstatt {
  pakete: string[]

  // eintraege sind die Pakete samt Art, Name und Herkunftsart — für die
  // eigene Kopie (FMT-02.AK-04, `usePaketkopie`).
  eintraege: PaketDTO[]

  // paketeGeladen sagt, dass die Paketliste einmal angekommen ist — vorher
  // ist sie unbekannt und nicht leer (UX-08.NAT-01).
  paketeGeladen: boolean
  paketeFehler: string | undefined
  paketeNeuLaden: () => void

  wahl: Dateiwahl | null
  waehlen: (paket: string, datei: string, beiwerk?: boolean) => void

  // hauptdatei ist die des gewählten Pakets — sie lässt sich weder
  // umbenennen noch entfernen (FMT-02, #1691).
  hauptdatei: string

  // fassung und baumNeuLaden halten den Baum nach einem Handgriff frisch.
  fassung: number
  baumNeuLaden: () => void

  rueckfrage: Dateiwahl | null

  // verwerfen nimmt das Ziel der Rückfrage entgegen — die Seite hat es in der
  // Hand, und ein Griff in den Zustand hier wäre ein Zweig ohne Fall.
  verwerfen: (ziel: Dateiwahl) => void
  bleiben: () => void

  puffer: Vorlagenpuffer
  pruefung: Pufferbefund

  // veraltet ist die Stelle des Befunds, wenn die Vorschau den letzten
  // gültigen Stand zeigt und nicht den getippten.
  veraltet: Skriptstelle | null

  karteID: string
  setzeKarteID: (karteID: string) => void
  karten: ProbekarteDTO[]
  kartenFehler: string | undefined

  symbole: PaketsymbolDTO[]

  // namen sind die Vorschläge des Felds — die Namen der Umgebung einer
  // Testkarte und die Definitionen des Pakets.
  namen: string[]

  symbol: Symbolstelle | null
  setzeZeiger: (stelle: number) => void
  verwendungen: SymbolverwendungDTO[]
  verwendungenLaden: boolean

  sprungziel: Sprungziel | null
  zurStelle: (stelle: Skriptstelle) => void
  oeffnen: (datei: string, zeile: number, spalte: number) => void
}

// **`fremdstand` zählt fremde Schreibvorgänge** (#156) — Befund 3 des
// Nachlaufs zum Gesamtreview über 0.6.0 (#1660): bezog die Kommandozeile ein
// Paket, fehlte es im Baum bis zum eigenen Neuladen.
//
// **Die gelesene Datei hängt nicht daran.** Der Datenstand zählt den
// Bestand, nicht die Dateien eines Pakets — eine Datei, die ein anderer
// umschreibt, bewegt ihn nicht. Was der Puffer dagegen hat, ist sein
// `stand`: wer über einen fremden Stand speichert, bekommt den Konflikt
// gemeldet und verliert seinen Entwurf nicht (`useVorlagenpuffer`).
export function useVorlagenwerkstatt(fremdstand = 0): Vorlagenwerkstatt {
  // **Die Paketliste kommt von der Paketseite** — dieselbe Bindung, dieselbe
  // Antwort. Weitergereicht wird der volle Pfad: Pakete liegen verschachtelt,
  // und der Name allein erreicht keines. **Behalten**, damit der Baum beim
  // Nachladen stehen bleibt.
  const paketstand = useLaden(
    async () => {
      const antwort = await api().Pakete()

      return auspacken(antwort, () => antwort.pakete)
    },
    [fremdstand],
    true,
  )

  const [wahl, setzeWahl] = useState<Dateiwahl | null>(null)
  const [rueckfrage, setzeRueckfrage] = useState<Dateiwahl | null>(null)
  const [anstehend, setzeAnstehend] = useState<Sprung | null>(null)
  const [sprungziel, setzeSprungziel] = useState<Sprungziel | null>(null)

  const [fassung, setzeFassung] = useState(0)

  const paket = wahl?.paket ?? ''
  // **Beiwerk wird nicht gelesen** — es ist kein Text, und der Puffer meldete
  // es als Fehler (FMT-02, #1691).
  const datei = wahl === null || wahl.beiwerk === true ? '' : wahl.datei
  const schluessel = `${paket}\n${datei}`

  const puffer = useVorlagenpuffer(paket, datei)
  const pruefung = usePufferbefund(
    puffer.laedt ? '' : puffer.inhalt,
    schluessel,
  )

  const [karte, setzeKarte] = useState({ paket: '', id: '' })
  const karteID = karte.paket === paket ? karte.id : ''
  const probekarten = useProbekarten(paket)
  const symbole = usePaketsymbole(paket, puffer.stand)

  const [zeiger, setzeZeigerStand] = useState({ schluessel: '', stelle: -1 })
  const stelle = zeiger.schluessel === schluessel ? zeiger.stelle : -1
  const symbol =
    stelle < 0 ? null : symbolAnStelle(puffer.inhalt, stelle, symbole.symbole)
  const gesucht = useVerzoegert(symbol?.name ?? '', verwendungsfrist)
  const verwendungen = useSymbolverwendungen(paket, gesucht)

  // **Die Namen der Umgebung** kommen von einer Karte — der gewählten
  // Testkarte oder der ersten, die es gibt. Ohne Karte bleiben die
  // Definitionen des Pakets.
  const kandidat = karteID !== '' ? karteID : (probekarten.karten[0]?.id ?? '')
  const skriptnamen = useLaden(async () => {
    if (kandidat === '') {
      return []
    }

    const antwort = await api().Skriptnamen(kandidat)

    return auspacken(antwort, () => antwort.namen)
  }, [kandidat])

  // **Der vorgemerkte Sprung wird ausgeführt, sobald die Datei steht.**
  if (
    anstehend !== null &&
    wahl !== null &&
    wahl.datei === anstehend.datei &&
    !puffer.laedt &&
    puffer.lesefehler === undefined
  ) {
    setzeAnstehend(null)
    setzeSprungziel({
      zeile: anstehend.zeile,
      spalte: anstehend.spalte,
      stand: (sprungziel?.stand ?? 0) + 1,
    })
  }

  const springen = (zeile: number, spalte: number) => {
    setzeSprungziel({ zeile, spalte, stand: (sprungziel?.stand ?? 0) + 1 })
  }

  const waehlen = (neuesPaket: string, neueDatei: string, beiwerk = false) => {
    if (
      wahl !== null &&
      wahl.paket === neuesPaket &&
      wahl.datei === neueDatei
    ) {
      return
    }

    // **Ungespeichertes wird nicht still verworfen** — die Rückfrage steht
    // im Zentrum, nicht in einem Dialog (§6).
    const ziel = { paket: neuesPaket, datei: neueDatei, beiwerk }

    if (puffer.geaendert) {
      setzeRueckfrage(ziel)

      return
    }

    setzeWahl(ziel)
  }

  const oeffnen = (zielDatei: string, zeile: number, spalte: number) => {
    if (wahl !== null && wahl.datei === zielDatei) {
      springen(zeile, spalte)

      return
    }

    setzeAnstehend({ datei: zielDatei, zeile, spalte })
    waehlen(paket, zielDatei)
  }

  const befundstelle = pruefung.befund?.stelle ?? null

  return {
    pakete: (paketstand.daten ?? []).map((eintrag) => eintrag.pfad),
    eintraege: paketstand.daten ?? [],
    paketeGeladen: paketstand.daten !== undefined,
    paketeFehler: paketstand.fehler,
    paketeNeuLaden: paketstand.nachladen,
    wahl,
    waehlen,
    hauptdatei:
      (paketstand.daten ?? []).find((eintrag) => eintrag.pfad === paket)
        ?.hauptdatei ?? '',
    fassung,
    baumNeuLaden: () => {
      setzeFassung((war) => war + 1)
    },
    rueckfrage,
    verwerfen: (ziel: Dateiwahl) => {
      // **Verworfen ist verworfen** (UX-03.AK-04): der Entwurf der Datei, die
      // jemand verlässt, kommt auch beim Zurückkehren nicht wieder.
      puffer.verwerfen()
      setzeWahl(ziel)
      setzeRueckfrage(null)
    },
    bleiben: () => {
      setzeRueckfrage(null)
      setzeAnstehend(null)
    },
    puffer,
    pruefung,
    veraltet:
      befundstelle !== null && pruefung.gueltig !== puffer.inhalt
        ? befundstelle
        : null,
    karteID,
    setzeKarteID: (id: string) => {
      setzeKarte({ paket, id })
    },
    karten: probekarten.karten,
    kartenFehler: probekarten.fehler,
    symbole: symbole.symbole,
    namen: [...new Set([...(skriptnamen.daten ?? []), ...symbole.namen])],
    symbol,
    setzeZeiger: (neueStelle: number) => {
      setzeZeigerStand({ schluessel, stelle: neueStelle })
    },
    verwendungen: verwendungen.verwendungen,
    verwendungenLaden: verwendungen.laedt || gesucht !== (symbol?.name ?? ''),
    sprungziel,
    zurStelle: (ziel: Skriptstelle) => {
      springen(ziel.zeile, ziel.spalte)
    },
    oeffnen,
  }
}
