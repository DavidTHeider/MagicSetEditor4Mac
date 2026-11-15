// **Ein Zahlenfeld, das seinen Bereich am Feld sagt** — Werkstück F von Phase
// 4.10 (#1325).
//
// # Warum es das braucht
//
// Vier Zahlenfelder trugen `min` und `max` — und **meldeten nichts**.
// Gemessen am 2026-09-03 im laufenden Fenster: nach `99999` in „Breite in
// Bildpunkten" gab `checkValidity()` `false` und `validationMessage` einen
// fertigen Satz; im Fenster stand **kein** `aria-invalid`, **kein**
// `role=alert` und kein sichtbarer Hinweis. Der Wert blieb stehen.
//
// **`min` und `max` ohne Formular sind eine Zusage, die niemand einlöst**
// (§10): der Browser meldet sie beim Absenden, und hier wird nichts
// abgesendet.
//
// # Was es tut
//
// Es hält den getippten Text selbst — und **meldet nur, was gilt**. Ein Wert
// außerhalb des Bereichs erreicht den Aufrufer nicht; er stünde sonst in
// einer Rechnung, die niemand geprüft hat (§7, Regel 3).
//
// **Ein leeres Feld ist kein Fehler.** Wer eine Zahl ändert, löscht sie
// zuerst; ein Hinweis nach dem ersten Rücktaster meldete etwas, das gerade
// geschieht.
//
// # Was Werkstück U3 dazugelegt hat (#1414)
//
//   - **Beschriftung, Hinweis und Befund kommen aus [Feld]** — dieselbe Naht
//     wie bei `Pflichtfeld`. Was hier bleibt, ist das Urteil über eine Zahl.
//   - **`groesstens` darf fehlen.** `domain.Format.zahlenPruefen` prüft die
//     vier Zahlen eines Formats nur gegen `< 0`, und
//     `service.Sammlungssichten.Dubletten` weist `behalten < 0` ab — beides
//     ohne Obergrenze. Eine hier erfundene wiese einen gültigen Wert ab (§11).
//   - **`ganzzahlig` ist keine Verzierung**: die Felder dahinter sind in Go
//     `int`. Eine Kommazahl käme dort als Umwandlungsfehler an, also als Satz
//     über JSON statt als Befund am Feld.
//   - **`befund`** trägt den Satz der Anwendung ans Feld — wie bei
//     `Pflichtfeld`, und aus demselben Grund: „die Obergrenze liegt unter der
//     Untergrenze" weiß nur die Domäne.
import { useState } from 'react'

import { Feld } from './Feld'

import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

interface ZahlenfeldEigenschaften {
  beschriftung: string
  wert: number
  kleinstens: number

  // groesstens fehlt, wo die Domäne keine Obergrenze kennt.
  groesstens?: number

  // ganzzahlig sagt, dass hinter dem Feld ein `int` steht.
  ganzzahlig?: boolean

  // schritt ist die Rasterung — `0.5` für Millimeter, `1` für Bildpunkte.
  schritt?: number

  angehalten?: boolean

  // beschreibung ist die Kennung eines Hilfesatzes, der ohnehin am Feld
  // steht. **Der Hinweis tritt neben ihn und nicht an seine Stelle**: was das
  // Feld bedeutet, gilt weiter, auch wenn der Wert nicht passt.
  beschreibung?: string

  // befund ist ein Fehler, den die Anwendung **zu diesem Feld** gemeldet hat.
  //
  // **Er schlägt das eigene Urteil**: wer eine Zahl im Bereich eingegeben hat,
  // die der Dienst trotzdem abweist, soll den Grund lesen und nicht den
  // Bereich.
  befund?: string | undefined

  geaendert: (wert: number) => void

  // gilt meldet bei jeder Eingabe, ob der getippte Text **abgeschickt werden
  // darf** — für ein Formular, das erst beim Absenden schreibt (UX-06.AK-03,
  // #1678).
  //
  // **Weil `geaendert` dort nicht reicht**: ein Wert außerhalb des Bereichs
  // erreicht den Aufrufer nicht, und der Knopf daneben schickte sonst still
  // den letzten gültigen — eine Menge von 1, wo 0 dasteht. Leer gilt hier
  // nicht: wer absendet, meint einen Wert.
  gilt?: (gilt: boolean) => void
}

export function Zahlenfeld({
  beschriftung,
  wert,
  kleinstens,
  groesstens,
  ganzzahlig,
  schritt,
  angehalten,
  beschreibung,
  befund,
  geaendert,
  gilt,
}: ZahlenfeldEigenschaften) {
  const texte = useTexte()

  // **Der getippte Text und nicht die Zahl.** `Number('')` ist `0`, und `0`
  // liegt bei den meisten Feldern unter der Grenze — ein leeres Feld sähe
  // damit falsch aus, statt leer zu sein.
  const [getippt, setzeGetippt] = useState<string>()

  // imBereich sagt, ob ein getippter Text eine Zahl ist, die gelten darf.
  //
  // **Sie steht innen und nicht daneben**, weil sie `kleinstens`,
  // `groesstens` und `ganzzahlig` braucht — als freie Funktion stünden die
  // drei an jeder Rufstelle noch einmal.
  const imBereich = (text: string): boolean => {
    const zahl = Number(text)

    return (
      Number.isFinite(zahl) &&
      zahl >= kleinstens &&
      (groesstens === undefined || zahl <= groesstens) &&
      // **Sicher und nicht nur ganz** (J05, #2269): jenseits von 2⁵³ − 1
      // rundet `Number` auf die nächste darstellbare Zahl, und die ist ganz.
      (ganzzahlig !== true || Number.isSafeInteger(zahl))
    )
  }

  const gezeigt = getippt ?? String(wert)
  const passt = gezeigt.trim() === '' || imBereich(gezeigt)

  const meldung =
    befund ??
    (passt
      ? undefined
      : einsetzen(bereichssatz(texte.allgemein, groesstens, ganzzahlig), {
          kleinstens: String(kleinstens),
          groesstens: String(groesstens),
        }))

  return (
    <Feld
      beschriftung={beschriftung}
      {...(meldung === undefined ? {} : { befund: meldung })}
      {...(beschreibung === undefined ? {} : { beschreibungsID: beschreibung })}
    >
      {(verbindung) => (
        <input
          {...verbindung}
          type="number"
          min={kleinstens}
          max={groesstens}
          step={schritt}
          value={gezeigt}
          disabled={angehalten}
          className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60 aria-[invalid=true]:border-fehler"
          onChange={(ereignis) => {
            const eingabe = ereignis.target.value

            setzeGetippt(eingabe)

            const giltEingabe = eingabe.trim() !== '' && imBereich(eingabe)
            gilt?.(giltEingabe)

            if (giltEingabe) {
              geaendert(Number(eingabe))
            }
          }}
        />
      )}
    </Feld>
  )
}

// bereichssatz wählt den Satz, der zu den Grenzen dieses Feldes passt.
//
// **Vier Fälle und vier Sätze, kein Zusammensetzen.** „Eine ganze Zahl" und
// „zwischen … und …" in zwei Bausteinen stünden in einer der beiden Sprachen
// schief (§9).
function bereichssatz(
  allgemein: {
    zahlenbereich: string
    zahlenAb: string
    ganzeZahlBereich: string
    ganzeZahlAb: string
  },
  groesstens: number | undefined,
  ganzzahlig: boolean | undefined,
): string {
  if (ganzzahlig === true) {
    return groesstens === undefined
      ? allgemein.ganzeZahlAb
      : allgemein.ganzeZahlBereich
  }

  return groesstens === undefined ? allgemein.zahlenAb : allgemein.zahlenbereich
}
