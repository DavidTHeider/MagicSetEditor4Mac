// **Was die Kontakte aus einem Adressbuch tun** — ID-04
// (#1703).
// Wie sie aussehen, steht in `seiten/Adressbuch.tsx` (§2).
//
// # Zwei Schritte
//
// **Einlesen schreibt nichts**: es zeigt, was in der Datei steht und wie jede
// Karte zu den vorhandenen Kontakten steht. **Übernehmen schreibt nur die
// Auswahl** — und wählt je Karte, welche Wege mitgehen (AK-01).
import { useState } from 'react'

import type {
  KartenvorschauDTO,
  KontaktDTO,
  KontaktwegDTO,
} from '../apikontakte'
import { api } from '../fassade'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'
import { useVorgang, type Vorgang } from './useVorgang'

// Karteinwahl ist eine Karte der Vorschau mit dem, was gewählt ist.
export interface Karteinwahl {
  karte: KartenvorschauDTO
  an: boolean
  wege: readonly boolean[]
}

// adressbuchvorgang ist **ein** Schlüssel für Einlesen und Übernehmen
// (Zweitblick über ID-04, Befund 3): liefe ein Einlesen neben einer
// Übernahme, räumte deren späte Antwort eine Vorschau ab, die nicht zu ihr
// gehört (§7, Regel 2).
export const adressbuchvorgang = 'adressbuch'

export interface Adressbuchstand {
  kontakte: readonly KontaktDTO[]
  laedt: boolean
  ladefehler: string | undefined
  fehler: string | undefined
  meldung: string | undefined
  auswahl: readonly Karteinwahl[] | undefined
  befunde: readonly string[]
  vorgang: Vorgang
  einlesen: () => void
  karteWaehlen: (stelle: number, an: boolean) => void
  wegWaehlen: (stelle: number, weg: number, an: boolean) => void
  // uebernehmen schreibt die Auswahl, die die Vorschau zeigt.
  uebernehmen: (auswahl: readonly Karteinwahl[]) => void
  verwerfen: () => void
  loeschen: (kontakt: KontaktDTO) => void
}

// waehlbar sagt, ob eine Karte etwas zu übernehmen hat — **eine
// unveränderte nicht**, eine vermutliche Dublette erst, wenn jemand es will.
function vorgewaehlt(karte: KartenvorschauDTO): boolean {
  return karte.abgleich === 'neu' || karte.abgleich === 'geaendert'
}

// alsUebernahme macht aus einer gewählten Karte, was Go schreibt: **ersetzt
// wird gegen die gelesene Fassung** (§7, Regel 2).
function alsUebernahme(wahl: Karteinwahl): KontaktDTO {
  const { kontakt, vorhanden, abgleich } = wahl.karte
  const wege: KontaktwegDTO[] = kontakt.wege.filter(
    (_, stelle) => wahl.wege[stelle],
  )

  // **Eine Vermutung ersetzt wie eine Änderung** (Zweitblick über ID-04,
  // Befund 1) — neu angelegt wäre sie die Dublette, die AK-02 ausschließt.
  if (abgleich !== 'neu' && vorhanden !== null) {
    return { ...kontakt, wege, id: vorhanden.id, fassung: vorhanden.fassung }
  }

  return { ...kontakt, wege }
}

export function useAdressbuch(): Adressbuchstand {
  const texte = useTexte()
  const eigene = texte.seiten.adressbuch
  const vorgang = useVorgang()
  const [runde, setzeRunde] = useState(0)
  const [fehler, setzeFehler] = useState<string>()
  const [meldung, setzeMeldung] = useState<string>()
  const [auswahl, setzeAuswahl] = useState<Karteinwahl[]>()
  const [befunde, setzeBefunde] = useState<readonly string[]>([])

  const stand = useLaden(
    async () => {
      const antwort = await api().Kontakte()

      return auspacken(antwort, () => antwort.kontakte)
    },
    [runde],
    true,
  )

  const einlesen = () => {
    void vorgang.starte(adressbuchvorgang, async () => {
      try {
        const antwort = await api().KontakteEinlesen()
        auspacken(antwort, () => null)
        setzeFehler(undefined)
        setzeMeldung(undefined)

        // **Ein Abbruch ändert nichts** — auch nicht an einer offenen
        // Vorschau (AK-03).
        if (antwort.abgebrochen) {
          return
        }

        setzeBefunde(antwort.befunde)
        setzeAuswahl(
          antwort.karten.map((karte) => ({
            karte,
            an: vorgewaehlt(karte),
            wege: karte.kontakt.wege.map(() => true),
          })),
        )
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  const aendern = (stelle: number, neu: (wahl: Karteinwahl) => Karteinwahl) => {
    setzeAuswahl((vorher) =>
      vorher?.map((wahl, hier) => (hier === stelle ? neu(wahl) : wahl)),
    )
  }

  const uebernehmen = (gezeigt: readonly Karteinwahl[]) => {
    const gewaehlt = gezeigt.filter((wahl) => wahl.an)

    void vorgang.starte(adressbuchvorgang, async () => {
      try {
        const antwort = await api().KontakteUebernehmen(
          gewaehlt.map(alsUebernahme),
        )
        auspacken(antwort, () => null)
        setzeFehler(undefined)
        setzeAuswahl(undefined)
        setzeBefunde([])
        setzeMeldung(
          einsetzen(eigene.uebernommen, {
            angelegt: String(antwort.angelegt),
            ersetzt: String(antwort.ersetzt),
          }),
        )
        setzeRunde((bisher) => bisher + 1)
      } catch (ursache: unknown) {
        // **Nach einem gescheiterten Schreiben bleibt die Auswahl** (§7,
        // Regel 3): wer sie ändern will, soll sie nicht neu treffen müssen.
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  const loeschen = (kontakt: KontaktDTO) => {
    void vorgang.starte(kontakt.id, async () => {
      try {
        auspacken(
          await api().KontaktLoeschen(kontakt.id, kontakt.fassung),
          () => null,
        )
        setzeFehler(undefined)
        setzeRunde((bisher) => bisher + 1)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  return {
    kontakte: stand.daten ?? [],
    laedt: stand.laedt,
    ladefehler: stand.fehler,
    fehler,
    meldung,
    auswahl,
    befunde,
    vorgang,
    einlesen,
    karteWaehlen: (stelle, an) => {
      aendern(stelle, (wahl) => ({ ...wahl, an }))
    },
    wegWaehlen: (stelle, weg, an) => {
      aendern(stelle, (wahl) => ({
        ...wahl,
        wege: wahl.wege.map((bisher, hier) => (hier === weg ? an : bisher)),
      }))
    },
    uebernehmen,
    verwerfen: () => {
      setzeAuswahl(undefined)
      setzeBefunde([])
    },
    loeschen,
  }
}
