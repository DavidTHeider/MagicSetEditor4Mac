// **Die Spieltexturen im Fenster** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Was er hält
//
// **Dass sieben gleiche Karten eine Textur benutzen.** Der Speicher daneben
// führt die Zahlen; dieser Haken sorgt dafür, dass ein Bauteil sie beim
// Erscheinen anmeldet und beim Verschwinden abmeldet — und dass zwischen
// beidem **einmal** geholt wird.
//
// **Und dass die Bilder da sind, bevor die Karte kommt** — `vorladen` holt
// den Rücken und das eigene Deck in der niedrigen Tischgüte.
//
// # Die Regeln aus §7
//
//  1. **Ein Wechsel räumt** (Regel 1): wer die Partie schließt, räumt die
//     Texturen mit. Sie gehören zu dieser Partie und zu diesem Profil.
//  2. **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (Regel 2):
//     `raeume` zählt eine **Runde** hoch, und eine Antwort aus einer alten
//     wird nicht mehr eingetragen. Der Beleg allein genügte nicht — er kann
//     zwischenzeitlich neu entstanden sein, und dann trüge eine Antwort aus
//     dem vorigen Profil den Schlüssel des neuen.
//  3. **Ein Vorgang je Textur** (Regel 4): zwei Rufe für dasselbe Bild wären
//     zwei Rasterläufe für ein Ergebnis. Genau das verhindert `belegen`.
import { useCallback, useEffect, useRef, useState } from 'react'

import type { TexturAntwort, Texturwunsch } from '../apitextur'
import { api, auspacken } from '../api'
import { meldungAus } from '../meldung'
import { Texturspeicher, texturschluessel } from './texturspeicher'
import { useTexte } from '../sprache/kontext'

export interface Texturlauf {
  // belege meldet ein Ding an und holt die Textur, wenn sie neu ist.
  belege: (wunsch: Texturwunsch) => void

  // gib meldet es wieder ab.
  gib: (wunsch: Texturwunsch) => void

  // adresse gibt die Adresse — leer, solange keine da ist.
  adresse: (wunsch: Texturwunsch) => string

  // zweiseitig sagt, ob die Karte hinter einem Wunsch zwei Seiten hat —
  // `false`, solange nichts da ist.
  zweiseitig: (wunsch: Texturwunsch) => boolean

  // vorladen holt Bilder, die noch niemand zeigt.
  //
  // **Ohne Anmeldung**: was vorgeladen ist, hält niemand fest — es liegt im
  // Fach in Go und wird beim ersten Zeigen gefunden statt gezeichnet.
  vorladen: (wuensche: Texturwunsch[]) => void

  raeume: () => void

  // befunde sagen, **warum** eine Karte behilflich aussieht — je Satz einmal.
  befunde: string[]

  fehler: string | undefined
}

// **Am fremden Tisch fragt das Fenster den Wirt** (`fremd`) — P9 (#2377),
// Vermerk 0198. Bis dahin gab es dort keine Vorderseiten: die Kennungen am
// Tisch waren Zeilen des Wirts, und hier nachgeschlagen fänden sie eine
// **andere** Karte (gefunden mit zwei Fenstern, `nat04`, S12.NAT-04). Jetzt
// schickt der Wirt keine Zeilen mehr, sondern Bildkennungen; eine Zeile ohne
// Bildkennung (ein Wirt vor 0.9.0) nimmt der fremde Tisch nicht. Rücken
// gehören dem Spiel und kommen weiter aus diesem Bestand.
//
// **Mit der Partie, deren Bild die Kennungen trägt** (Zweitblick über
// Phase 7, Befund 4): ein Fenster kann einen eigenen Tisch führen und an
// einem fremden sitzen, und Go fragt nur den, der zu dieser Partie gehört.
export function useSpieltexturen(fremd = false, partie = ''): Texturlauf {
  const texte = useTexte()

  // **Ein Speicher je Haken und kein Modulspeicher.** Ein Speicher neben der
  // Anwendung überlebte den Profilwechsel und zeigte die Karten von vorhin.
  const speicher = useRef(new Texturspeicher())

  // **Ein Zähler, um neu zu zeichnen.** Der Speicher ist kein Zustand von
  // React; ohne dieses Zeichen sähe niemand, dass eine Textur angekommen ist.
  const [, neuZeichnen] = useState(0)
  const [fehler, setzeFehler] = useState<string>()
  const [befunde, setzeBefunde] = useState<string[]>([])

  const hole = useCallback(
    async (wunsch: Texturwunsch, schluessel: string, runde: number) => {
      try {
        const antwort = await holeTextur(wunsch, partie)

        auspacken(antwort, () => undefined)

        if (!speicher.current.setze(schluessel, antwort.textur, runde)) {
          return
        }

        if (antwort.textur.befunde.length > 0) {
          setzeBefunde((bisher) => vereinigt(bisher, antwort.textur.befunde))
        }

        neuZeichnen((stand) => stand + 1)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    },
    [texte, partie],
  )

  const belege = useCallback(
    (wunsch: Texturwunsch) => {
      if (ohneBild(wunsch, fremd)) {
        return
      }

      const schluessel = texturschluessel(wunsch)

      if (!speicher.current.belegen(schluessel)) {
        return
      }

      void hole(wunsch, schluessel, speicher.current.runde())
    },
    [hole, fremd],
  )

  const vorladen = useCallback(
    (wuensche: Texturwunsch[]) => {
      for (const wunsch of wuensche) {
        if (
          ohneBild(wunsch, fremd) ||
          speicher.current.kennt(texturschluessel(wunsch))
        ) {
          continue
        }

        // **Vorgeladen wird ohne Anmeldung** — der Speicher merkt sich den
        // Schlüssel, damit nicht zweimal geholt wird, aber niemand hält ihn
        // fest. Wer ihn später zeigt, findet ihn da.
        const schluessel = texturschluessel(wunsch)

        speicher.current.merke(schluessel)

        void hole(wunsch, schluessel, speicher.current.runde())
      }
    },
    [hole, fremd],
  )

  const gib = useCallback(
    (wunsch: Texturwunsch) => {
      if (ohneBild(wunsch, fremd)) {
        return
      }

      speicher.current.gib(texturschluessel(wunsch))
    },
    [fremd],
  )

  const adresse = useCallback(
    (wunsch: Texturwunsch) =>
      speicher.current.hole(texturschluessel(wunsch))?.adresse ?? '',
    [],
  )

  const zweiseitig = useCallback(
    (wunsch: Texturwunsch) =>
      speicher.current.hole(texturschluessel(wunsch))?.zweiseitig ?? false,
    [],
  )

  const raeume = useCallback(() => {
    speicher.current.raeume()
    setzeFehler(undefined)
    setzeBefunde([])
    neuZeichnen((stand) => stand + 1)
  }, [])

  // **Und beim Verschwinden des Tisches ebenso.** Ein Speicher, der einen
  // Seitenwechsel überlebte, hielte Bilder, die niemand mehr zeigt.
  //
  // **Der Griff wird im Effekt gelesen und nicht beim Zeichnen** — eine
  // Referenz ist kein Zustand, und wer sie beim Zeichnen anfasst, bekommt
  // einen Baum, der sich nicht erneuert (`react-hooks`, und zu Recht).
  useEffect(() => {
    const jetzt = speicher.current

    return () => {
      jetzt.raeume()
    }
  }, [])

  return {
    belege,
    gib,
    adresse,
    zweiseitig,
    vorladen,
    raeume,
    befunde,
    fehler,
  }
}

// holeTextur fragt den Weg, der zu einem Wunsch gehört — den eigenen Bestand
// mit der Zeile, den Wirt mit Objekt und Bildkennung (P9, #2377), den Rücken
// über das Spiel.
function holeTextur(
  wunsch: Texturwunsch,
  partie: string,
): Promise<TexturAntwort> {
  if (wunsch.art === 'ruecken') {
    return api().Rueckentextur(wunsch.spiel, wunsch.guete)
  }

  // **Objekt und Bildkennung kommen hier nur zusammen an** — `ohneBild` hat
  // einen Wunsch ohne Zeile und ohne Bildkennung schon aussortiert.
  const { karteID, objekt = '', bild = '', seite, guete } = wunsch

  if (karteID === '') {
    return api().Fremdkartenbild(partie, objekt, bild, seite, guete)
  }

  return api().Spieltextur(karteID, seite, guete)
}

// ohneBild sagt, ob es zu diesem Wunsch gar kein Bild geben kann.
//
// **Ein verdecktes Objekt und ein echter Druck tragen keine Kartenkennung**,
// und ein Ruf mit leerer Kennung wäre eine Abweisung, die niemand braucht.
//
// **Der Rücken hat diese Frage nicht** — er hängt an einem Spiel, und das
// steht in jeder Partiesicht.
function ohneBild(wunsch: Texturwunsch, fremd: boolean): boolean {
  if (wunsch.art === 'ruecken') {
    return wunsch.spiel === ''
  }

  // **Am fremden Tisch zählt die Bildkennung** (P9, #2377): ein Wirt vor
  // 0.9.0 schickt seine eigenen Zeilen ohne sie, und hier nachgeschlagen
  // fänden sie andere Karten. **Am eigenen** genügt eines von beiden — eine
  // Karte aus dem Deck eines Gastes trägt keine Zeile dieses Bestands, und
  // dann zeichnet der Gast (P9b).
  const ohneKennung = (wunsch.bild ?? '') === ''

  return fremd ? ohneKennung : ohneKennung && wunsch.karteID === ''
}

// vereinigt hängt an, was noch nicht dasteht.
//
// **Je Satz einmal**: sechzig Karten mit derselben fehlenden Schrift sind ein
// Befund und nicht sechzig.
function vereinigt(bisher: string[], neue: string[]): string[] {
  const zusammen = [...bisher]

  for (const satz of neue) {
    if (!zusammen.includes(satz)) {
      zusammen.push(satz)
    }
  }

  return zusammen
}
