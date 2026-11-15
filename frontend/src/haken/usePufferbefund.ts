// **Der Befund des Puffers und sein letzter gültiger Stand** — Werkstück U7
// von Phase 5.1 (#1418).
//
// # Warum der gültige Stand hier liegt
//
// Die Abnahme verlangt: „Fehler in Zeile/Feld zeigt exakte Stelle, letzte
// gültige Vorschau bleibt als veraltet markiert; Reparatur ersetzt sie ohne
// Neustart." Die Vorschau kann sich den letzten gültigen Text nicht selbst
// merken — sie bekommt nur den aktuellen. Also merkt ihn der Haken, der weiß,
// **wann** ein Text gültig war: wenn die Prüfung für genau diesen Text ohne
// Stelle zurückkam.
//
// # Warum nicht `useSkriptbefund`
//
// Der bis U7 gab nur den Befund und behielt ihn, während der nächste Lauf
// rechnete — richtig für die Marke im Feld, aber niemand konnte sagen, **zu
// welchem** Text er gehört. Hier kommt der Text mit der Antwort zurück.
//
// # Der Stand wird beim Zeichnen nachgeführt, nicht in einem Effekt
//
// „Storing information from previous renders" aus der React-Dokumentation:
// ein `setState` während des Zeichnens, hinter einem Vergleich. Ein Effekt
// zeichnete zweimal und liefe dem Bild eine Runde hinterher.
import { useState } from 'react'

import { api } from '../fassade'
import { useLaden } from './useLaden'

import type { Skriptstelle } from '../api'

export interface Skriptbefund {
  stelle: Skriptstelle | null
  // meldung ist der deutsche Satz aus Go — der Rückfall, wenn der Katalog den
  // Schlüssel nicht kennt (§9).
  meldung: string
  schluessel: string
  werte: string[]
}

export interface Pufferbefund {
  // befund ist die letzte Antwort — sie bleibt stehen, während die nächste
  // rechnet, damit die Marke nicht flackert.
  befund: Skriptbefund | undefined

  // geprueft sagt, ob `befund` zum aktuellen Text gehört.
  geprueft: boolean

  // gueltig ist der letzte Text, den die Prüfung ohne Stelle gesehen hat —
  // `undefined`, solange es keinen gab.
  gueltig: string | undefined
}

// **`schluessel` benennt die Datei.** Der gültige Stand gehört zu ihr; nach
// einem Dateiwechsel gilt der alte nicht mehr (§7, Regel 1) — sonst zeichnete
// die Vorschau den Text der vorigen Datei unter dem Namen der neuen.
//
// **Ein leerer Text wird nicht geprüft.** Er steht da, solange die Datei
// lädt; ein Ruf dafür wäre einer ohne Frage.
export function usePufferbefund(inhalt: string, schluessel = ''): Pufferbefund {
  const stand = useLaden<{ befund: Skriptbefund; inhalt: string }>(
    async () => {
      if (inhalt === '') {
        return {
          befund: { stelle: null, meldung: '', schluessel: '', werte: [] },
          inhalt,
        }
      }

      const antwort = await api().PaketdateiPruefen(inhalt)

      return {
        befund: {
          stelle: antwort.stelle,
          meldung: antwort.meldung,
          schluessel: antwort.schluessel,
          werte: antwort.werte,
        },
        inhalt,
      }
    },
    [inhalt, schluessel],
    true,
  )

  const geprueft = !stand.laedt && stand.daten?.inhalt === inhalt

  const [gueltig, setzeGueltig] = useState<{
    schluessel: string
    text: string
  }>()
  const eigener = gueltig?.schluessel === schluessel ? gueltig.text : undefined

  if (
    geprueft &&
    inhalt !== '' &&
    stand.daten?.befund.stelle === null &&
    eigener !== inhalt
  ) {
    setzeGueltig({ schluessel, text: inhalt })
  }

  return { befund: stand.daten?.befund, geprueft, gueltig: eigener }
}
