// **Der Puffer der Vorlagenwerkstatt** — Werkstück U7 von Phase 5.1
// (#1418).
//
// # Was er hält
//
// Den gelesenen Stand einer Paketdatei (`basis`, `stand`), den getippten
// (`inhalt`) und das Schreiben zurück. Bis U7 lag das in `Vorlagendatei.tsx`,
// und der Paketwechsel räumte es über einen `key`. Im Arbeitsplatz lesen
// **zwei** Regionen denselben Puffer — der Editor im Zentrum, Vorschau und
// Diagnose im Inspector —, und ein `key` kann nur ein Bauteil abräumen.
// Deshalb hält der Haken den Entwurf **unter dem Schlüssel der Datei**: ein
// Entwurf, dessen Schlüssel nicht mehr der gewählte ist, gilt nicht (§7,
// Regel 1), und eine Antwort trägt sich nur ein, wenn ihr Schlüssel noch der
// aktuelle ist (Regel 2).
//
// # Die vier Regeln
//
// 1. Ein Dateiwechsel räumt den Entwurf — über den Schlüssel, nicht über
//    einen Effekt.
// 2. `useLaden` zählt den Lesevorgang mit; das Schreiben vergleicht den
//    Schlüssel nach der Antwort mit dem aktuellen.
// 3. Geschrieben ist geschrieben: nach einem geglückten Schreiben wird der
//    neue Stand eingetragen und nichts nachgeladen; nach einem gescheiterten
//    bleibt der Text stehen, und die Meldung sagt, was zu tun ist.
// 4. `useVorgang` lässt je Datei einen Schreibvorgang zu.
//
// # Und der Entwurf überlebt das Verlassen der Seite
//
// **UX-03.AK-04** (#1675):
// *„Entwürfe werden sicher behandelt."* Bis zum 2026-09-24 lebte der
// getippte Stand im Zustand dieses Hakens, und wer die Werkstatt über die
// Leiste, ⌘K oder ein Kürzel verließ, fand ihn beim Zurückkommen nicht mehr
// — ohne Rückfrage; die gab es nur beim Wechsel der Datei. Jetzt steht er
// unter dem Schlüssel der Datei in [aufgehoben], bis er gespeichert oder
// verworfen wird. **Ein Profilwechsel räumt alle** (`vergissVorlagenentwuerfe`
// in `App.tsx`): ein Entwurf aus einem anderen Profil wären fremde Daten.
//
// **Was er nicht hält**: ein Beenden der Anwendung. Er lebt im Speicher
// dieses Fensters.
import { useCallback, useEffect, useRef, useState } from 'react'

import { Bindungsfehler } from '../bindungsfehler'
import { api } from '../fassade'
import { meldungAus } from '../meldung'
import { auspacken } from '../umschlag'
import { useTexte } from '../sprache/kontext'
import { useLaden } from './useLaden'
import { useVorgang } from './useVorgang'

// Entwurf ist der getippte Stand samt dem, wogegen er gemessen wird.
interface Entwurf {
  schluessel: string
  inhalt: string
  basis: string
  stand: string
  gespeichert: boolean
}

// aufgehoben hält die ungespeicherten Entwürfe je Datei über das Abbauen
// der Seite hinaus.
const aufgehoben = new Map<string, Entwurf>()

// vergissVorlagenentwuerfe räumt alle — beim Profilwechsel.
export function vergissVorlagenentwuerfe(): void {
  aufgehoben.clear()
}

export interface Vorlagenpuffer {
  inhalt: string

  // basis ist der zuletzt gelesene oder geschriebene Inhalt — der Stand, gegen
  // den der Diff läuft.
  basis: string

  // stand ist die Prüfsumme, die beim Schreiben zurückgeht.
  stand: string

  geaendert: boolean
  gespeichert: boolean
  laedt: boolean
  schreibt: boolean

  // lesefehler ist der Fehler beim Aufschlagen — dann gibt es kein Feld, nur
  // die Meldung (§7, Regel 3: kein halber Stand).
  lesefehler: string | undefined

  // fehler ist der Fehler beim Schreiben — der Text bleibt stehen.
  fehler: string | undefined

  setzeInhalt: (inhalt: string) => void
  speichern: () => Promise<void>
  neuLaden: () => void

  // behalten löst einen Konflikt zugunsten des eigenen Texts: die fremde
  // Fassung wird die Basis, der Text bleibt, und Sichern geht gegen ihren
  // Stand — FMT-02.AK-03 (#1691).
  behalten: () => Promise<void>

  // verwerfen nimmt den Entwurf dieser Datei zurück — auch den aufgehobenen.
  verwerfen: () => void
}

export function useVorlagenpuffer(
  paket: string,
  datei: string,
): Vorlagenpuffer {
  const texte = useTexte()
  const schluessel = `${paket}\n${datei}`
  // **Der aktuelle Schlüssel für die Zeit nach dem `await`** — gesetzt im
  // Effekt, nicht beim Zeichnen (die Regel gegen Refs im Zeichnen gilt), und
  // gelesen erst, wenn eine Antwort zurück ist.
  const aktuell = useRef(schluessel)
  useEffect(() => {
    aktuell.current = schluessel
  }, [schluessel])

  const gelesen = useLaden(async () => {
    if (datei === '') {
      return null
    }

    const antwort = await api().PaketdateiLesen(paket, datei)

    return auspacken(antwort, () => ({
      inhalt: antwort.inhalt,
      stand: antwort.stand,
    }))
  }, [paket, datei])

  const [entwurf, setzeEntwurf] = useState<Entwurf>()
  const [schreibfehler, setzeSchreibfehler] = useState<{
    schluessel: string
    meldung: string
  }>()
  const vorgang = useVorgang()

  const gilt =
    entwurf?.schluessel === schluessel ? entwurf : aufgehoben.get(schluessel)
  const basis = gilt?.basis ?? gelesen.daten?.inhalt ?? ''
  const stand = gilt?.stand ?? gelesen.daten?.stand ?? ''
  const inhalt = gilt?.inhalt ?? basis

  const setzeInhalt = useCallback(
    (neu: string) => {
      const naechster: Entwurf = {
        schluessel,
        inhalt: neu,
        basis,
        stand,
        gespeichert: false,
      }

      aufgehoben.set(schluessel, naechster)
      setzeEntwurf(naechster)
    },
    [basis, schluessel, stand],
  )

  const speichern = useCallback(async () => {
    const eigener = schluessel

    const neu = await vorgang.starte(eigener, async () => {
      // **Der Aufruf steht im Versuch.** Draußen entkäme eine abgelehnte
      // Zusage — eine abgerissene Bindung, eine fehlende Methode — an
      // `vorgang.starte` vorbei, und die Seite zeigte nichts.
      try {
        const antwort = await api().PaketdateiSchreiben(
          paket,
          datei,
          inhalt,
          stand,
        )

        return auspacken(antwort, () => antwort.stand)
      } catch (ursache: unknown) {
        // **Ein Konflikt bekommt seinen eigenen Satz** — „Der Eintrag wurde
        // zwischenzeitlich geändert" sagte nicht, dass der getippte Text noch
        // dasteht und was damit zu tun ist.
        // **Unter dem eigenen Schlüssel abgelegt, nicht geprüft**: wer
        // inzwischen eine andere Datei aufgeschlagen hat, sieht ihn nicht —
        // `fehler` gibt nur den Fehler des aktuellen Schlüssels heraus.
        setzeSchreibfehler({
          schluessel: eigener,
          meldung:
            ursache instanceof Bindungsfehler && ursache.code === 'konflikt'
              ? texte.seiten.vorlagen.konflikt
              : meldungAus(ursache, texte),
        })

        return undefined
      }
    })

    // **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (Regel 2): wer
    // inzwischen eine andere Datei aufgeschlagen hat, bekommt sie nicht
    // eingetragen.
    if (neu === undefined || aktuell.current !== eigener) {
      return
    }

    setzeSchreibfehler(undefined)
    aufgehoben.delete(eigener)
    setzeEntwurf({
      schluessel: eigener,
      inhalt,
      basis: inhalt,
      stand: neu,
      gespeichert: true,
    })
  }, [datei, inhalt, paket, schluessel, stand, texte, vorgang])

  const neuLaden = useCallback(() => {
    aufgehoben.delete(schluessel)
    setzeEntwurf(undefined)
    setzeSchreibfehler(undefined)
    gelesen.nachladen()
  }, [gelesen, schluessel])

  // **Die fremde Fassung wird die Basis, der eigene Text bleibt.** Danach
  // zeigen die Änderungen, was gegenüber ihr überschrieben wird — und nichts
  // wird still überschrieben: gesichert wird erst auf Knopfdruck.
  //
  // **Eine späte Antwort trägt sich nicht ein** (Regel 2), auch nicht unter
  // ihrem eigenen Schlüssel: wer die Datei verlassen hat, hat den Entwurf
  // womöglich verworfen, und `aufgehoben` brächte ihn beim Wiederkommen
  // zurück.
  const behalten = useCallback(async () => {
    const eigener = schluessel
    const mein = inhalt

    try {
      const antwort = await api().PaketdateiLesen(paket, datei)
      const frisch = auspacken(antwort, () => ({
        inhalt: antwort.inhalt,
        stand: antwort.stand,
      }))

      if (aktuell.current !== eigener) {
        return
      }

      const naechster: Entwurf = {
        schluessel: eigener,
        inhalt: mein,
        basis: frisch.inhalt,
        stand: frisch.stand,
        gespeichert: false,
      }

      aufgehoben.set(eigener, naechster)
      setzeEntwurf(naechster)
      setzeSchreibfehler(undefined)
    } catch (ursache: unknown) {
      if (aktuell.current === eigener) {
        setzeSchreibfehler({
          schluessel: eigener,
          meldung: meldungAus(ursache, texte),
        })
      }
    }
  }, [datei, inhalt, paket, schluessel, texte])

  // **Mit dem Entwurf geht sein Fehler** — ein Konflikt über einen Text, den
  // es nicht mehr gibt, stünde sonst beim Wiederkommen über der frischen Datei.
  const verwerfen = useCallback(() => {
    aufgehoben.delete(schluessel)
    setzeEntwurf(undefined)
    setzeSchreibfehler((fehler) =>
      fehler?.schluessel === schluessel ? undefined : fehler,
    )
  }, [schluessel])

  return {
    inhalt,
    basis,
    stand,
    geaendert: inhalt !== basis,
    gespeichert: gilt?.gespeichert ?? false,
    laedt: gelesen.laedt,
    schreibt: vorgang.laeuft(schluessel),
    lesefehler: gelesen.fehler,
    fehler:
      schreibfehler?.schluessel === schluessel
        ? schreibfehler.meldung
        : undefined,
    setzeInhalt,
    speichern,
    neuLaden,
    behalten,
    verwerfen,
  }
}
