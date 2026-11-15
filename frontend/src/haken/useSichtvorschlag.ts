// **Die Sicht des Gegenübers anfragen und ihren Vorschlag übernehmen** —
// P11 (#2379),
// Befund C-15.
//
// # Ein Weg ohne Datei und ohne Text
//
// **Über die Leitung** (`HandelSichtAnfragen`), dann **derselbe Abgleich**
// wie mit einer Datei (`Tauschabgleich`) — Angebot und Wünsche getrennt,
// denn eine Wunschzeile ist kein Angebot. Was der Vorschlag nennt, wird zu
// Zeilen: was das Gegenüber gibt, mit dem Zustand seiner Zeile; was du gibst,
// nur mit dem Druck — welches Exemplar weggeht, sagst du in „Zusagen“.
//
// # Nebenläufig
//
// **Eine Antwort gehört zu ihrer Anfrage** (§7, Regel 2), und wer den Handel
// verlässt, bekommt keine mehr hierher. **Übernommen ist übernommen**: ein
// zweites Mal legt keine Zeile doppelt an.
//
// # Mit einer Sicht, die schon geholt ist
//
// **Am Markttisch fragt der Tisch selbst** (#2494, H3): sein rechtes Fach ist
// die freigegebene Sicht. Wer sie hereinreicht (`geholt`), bekommt den
// Abgleich, sobald sie da ist — ohne Klick, und ohne dass sie ein zweites
// Mal über die Leitung geht.
import { useEffect, useState } from 'react'

import { api } from '../fassade'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'
import { useLauf } from './useLauf'

import type { Handelsposition } from '../apihandelsleitung'
import type { Sichtkarte } from '../apisichtfreigabe'
import type {
  Tauschabgleichantwort,
  Tauschkandidat,
} from '../apitauschabgleich'

export interface Sichtergebnis {
  sicht: Sichtkarte[]
  abgleich: Tauschabgleichantwort
}

// Geholt ist eine Sicht, die ein anderer schon geholt hat — `zeilen` fehlt,
// solange sie lädt.
export interface Geholt {
  zeilen: Sichtkarte[] | undefined
}

interface Eigenschaften {
  ich: string
  gegenueber: string
  positionen: Handelsposition[]
  setzePositionen: (positionen: Handelsposition[]) => void
  geholt?: Geholt | undefined
}

export function useSichtvorschlag({
  ich,
  gegenueber,
  positionen,
  setzePositionen,
  geholt,
}: Eigenschaften) {
  const texte = useTexte()
  const lauf = useLauf()
  const [laeuft, setzeLaeuft] = useState(false)
  const [fehler, setzeFehler] = useState<string>()
  const [gefragt, setzeGefragt] = useState<Sichtergebnis>()

  // **Übernommen gilt für ein Ergebnis** — ein neues, gefragt oder vom Tisch
  // nachgeladen, steht wieder offen.
  const [uebernommen, setzeUebernommen] = useState<{
    fuer: Sichtergebnis
    zahl: number
  }>()

  useEffect(
    () => () => {
      lauf.naechster()
    },
    [lauf],
  )

  const anfragen = async () => {
    const nummer = lauf.naechster()
    setzeLaeuft(true)
    setzeFehler(undefined)
    setzeGefragt(undefined)

    let ausgang: Sichtergebnis | { fehler: string }
    try {
      const antwort = await api().HandelSichtAnfragen(gegenueber)
      ausgang = await abgleichen(
        gegenueber,
        auspacken(antwort, () => antwort.zeilen),
      )
    } catch (ursache) {
      ausgang = { fehler: meldungAus(ursache, texte) }
    }

    if (!lauf.istAktuell(nummer)) {
      return
    }

    setzeLaeuft(false)

    if ('fehler' in ausgang) {
      setzeFehler(ausgang.fehler)
    } else {
      setzeGefragt(ausgang)
    }
  }

  // **Steht die Sicht schon da, wird sie abgeglichen, sobald sie kommt** —
  // und noch einmal, wenn der Tisch neu fragt. Über `useLaden`, den Wächter
  // jedes Ladeeffekts (§7, Regel 2).
  const zeilen = geholt?.zeilen
  const vomTisch = useLaden(
    () =>
      zeilen === undefined
        ? Promise.resolve(undefined)
        : abgleichen(gegenueber, zeilen),
    [zeilen, gegenueber],
  )

  const stand =
    geholt === undefined
      ? { laeuft, fehler, ergebnis: gefragt }
      : {
          laeuft: vomTisch.laedt,
          fehler: vomTisch.fehler,
          ergebnis: vomTisch.daten,
        }

  const ergebnis = stand.ergebnis
  const vorgeschlagen =
    ergebnis === undefined ? [] : zeilenAus(ergebnis, ich, gegenueber)

  const uebernehmen = (fuer: Sichtergebnis) => {
    const neue = vorgeschlagen.filter(
      (zeile) =>
        !positionen.some(
          (da) =>
            da.von === zeile.von &&
            da.printing === zeile.printing &&
            da.guete === zeile.guete &&
            da.sprache === zeile.sprache &&
            da.finish === zeile.finish,
        ),
    )

    setzePositionen([...positionen, ...neue])
    setzeUebernommen({ fuer, zahl: neue.length })
  }

  return {
    ...stand,
    vorgeschlagen,
    uebernommen:
      uebernommen !== undefined && uebernommen.fuer === ergebnis
        ? uebernommen.zahl
        : undefined,
    anfragen,
    uebernehmen,
  }
}

// abgleichen gleicht eine Sicht mit dem eigenen Bestand ab — **Angebot und
// Wünsche getrennt**, denn eine Wunschzeile ist kein Angebot.
async function abgleichen(
  gegenueber: string,
  sicht: Sichtkarte[],
): Promise<Sichtergebnis> {
  const abgleich = await api().Tauschabgleich(
    gegenueber,
    {
      angebot: sicht.filter((zeile) => zeile.wunsch !== true),
      wuensche: sicht.filter((zeile) => zeile.wunsch === true),
    },
    1,
    0,
  )

  return { sicht, abgleich: auspacken(abgleich, () => abgleich) }
}

// zeilenAus macht aus dem Vorschlag Handelszeilen — **ohne Vorschlag** (es
// gibt nichts, was das Gegenüber von dir sucht) das, was du bekommen
// könntest.
function zeilenAus(
  { abgleich }: Sichtergebnis,
  ich: string,
  gegenueber: string,
): Handelsposition[] {
  const gebe = abgleich.hatVorschlag ? abgleich.vorschlag.gebe.zeilen : []
  const bekomme = abgleich.hatVorschlag
    ? abgleich.vorschlag.bekomme.zeilen
    : abgleich.bekomme

  return [
    ...gebe.map((kandidat) => zeileFuer(ich, kandidat)),
    ...bekomme.map((kandidat) => zeileFuer(gegenueber, kandidat)),
  ]
}

// zeileFuer macht aus einem Treffer eine Handelszeile — **mit seiner
// Variante**, auf beiden Seiten (Befund 2 des Zweitblicks über Phase 8):
// ohne Sprache scheiterte drüben der Zugang beim Verbuchen, und ein Nachschlagen
// über den Druck allein nahm bei zwei Varianten die erste.
function zeileFuer(von: string, kandidat: Tauschkandidat): Handelsposition {
  return {
    von,
    printing: kandidat.angebotDruck,
    druckId: 0,
    menge: kandidat.menge,
    guete: kandidat.angebotZustand,
    sprache: kandidat.angebotSprache,
    finish: kandidat.angebotFinish,
    abbild: '',
    bewertung: 0,
  }
}
