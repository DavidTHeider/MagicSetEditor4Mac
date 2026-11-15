// **Nur zeichnen, was zu sehen ist** —
// #891,
// Werkstück 5 von Phase Y.
//
// # Was gemessen wurde
//
// `Black.mse-set` hat 3.788 Karten. Vor diesem Haken standen **15.541
// DOM-Knoten** in der Kartenlasche, und **jeder Klick** in die Liste — eine
// Karte wählen, eine anhaken — kostete eine **Langaufgabe von 73 bis 132 ms**
// (vier bzw. drei Proben, gemessen am 2026-08-25 in der laufenden Anwendung
// über einen `PerformanceObserver` auf `longtask`).
//
// Das ist der Unterschied zwischen „reagiert" und „hakt": ab 50 ms ist eine
// Eingabe als Verzögerung zu spüren, und ein Bild dauert 16 ms.
//
// # ported
//
// MSE2 zeichnet seine Kartenliste ebenso wenig vollständig. `ItemList` ist ein
// `wxListView` mit `wxLC_VIRTUAL` (`src/gui/control/item_list.cpp:17`); es
// bekommt über `SetItemCount` nur die **Zahl** der Einträge (ebenda, Zeile 183)
// und beantwortet für jede tatsächlich gezeichnete Zeile
// `CardListBase::OnGetItemText(pos, col)` (`src/gui/control/card_list.cpp:307`).
// Fassung f9d9356 vom 2021-07-06.
//
// # Was er nicht kann
//
// **Zeilen verschiedener Höhe.** Er rechnet mit **einer** Höhe, gemessen an der
// ersten gezeichneten Zeile. Die Kartenliste hat sie: eine Zeile, abgeschnitten
// statt umgebrochen. Wer das ändert, nimmt diesem Haken die Grundlage.
import { useCallback, useRef, useState } from 'react'

// rand ist der Überhang je Seite, in Zeilen.
//
// **Er ist der Grund, warum die Tastatur weiter durch die Liste läuft.** Wer
// zur letzten sichtbaren Zeile tabbt, scrollt sie ins Bild — und das nächste
// Ziel muss schon dastehen, bevor der Browser es sucht. Ohne Überhang endete
// der Tabulator am Fensterrand, und alles darunter wäre mit der Tastatur nicht
// mehr erreichbar (§6).
const rand = 8

// probezeilen ist, was gezeichnet wird, solange nichts gemessen ist.
const probezeilen = 1

export interface Sichtfenster {
  // behaelter wird an das scrollende Element gehängt.
  //
  // **Er gibt eine Aufräumfunktion zurück**, und die gehört durchgereicht: an
  // dem Knoten hängen ein `scroll`-Horcher und ein `ResizeObserver`. Wer die
  // Referenz einwickelt und nichts zurückgibt, lässt beide am abgeräumten
  // Knoten hängen (#1290).
  behaelter: (knoten: HTMLElement | null) => (() => void) | undefined

  // zeilenmass wird an die **erste gezeichnete** Zeile gehängt.
  zeilenmass: (knoten: HTMLElement | null) => void

  erste: number

  // letzte ist **ausschließlich** — wie das Ende eines Slice.
  letzte: number

  // obenHoch und untenHoch sind die Abstandhalter, die den Rollbalken tragen.
  obenHoch: number
  untenHoch: number

  // zeilenhoehe ist, was die erste gezeichnete Zeile misst — **0**, solange
  // nichts gemessen wurde.
  //
  // **Sie steht hier, weil ein Aufrufer sie braucht**, der zu einer Stelle
  // rollen will (#1290).
  // Aus `obenHoch / erste` ließe sie sich rechnen — außer bei `erste === 0`,
  // und das ist der Normalfall.
  zeilenhoehe: number
}

interface Mass {
  zeile: number
  sicht: number
  oben: number
}

export function useSichtfenster(anzahl: number): Sichtfenster {
  const [mass, setzeMass] = useState<Mass>({ zeile: 0, sicht: 0, oben: 0 })
  const knoten = useRef<HTMLElement | null>(null)
  const zeile = useRef<HTMLElement | null>(null)

  const nachmessen = useCallback(() => {
    const behaelter = knoten.current

    if (behaelter === null) {
      return
    }

    const zeilenhoehe = zeile.current?.getBoundingClientRect().height ?? 0
    const sicht = behaelter.clientHeight
    const oben = behaelter.scrollTop

    // **Nur bei echter Änderung setzen.** Ein `scroll` feuert je Bild; ein
    // Zustand, der sich dabei jedes Mal neu setzt, zeichnet die Liste bei
    // jedem Bild neu — genau das, was hier vermieden werden soll.
    setzeMass((bisher) => {
      // **Eine Höhe von null überschreibt keine gemessene.**
      //
      // Beim Rollen wechselt die **erste** Zeile ihre Identität; React löst
      // die Referenz von der alten (`null`) und hängt sie an die neue. Wer
      // dazwischen auf null zurückfiele, zeichnete für einen Herzschlag die
      // **ganze** Liste — bei jedem Rollen, und genau das soll hier aufhören.
      const gemessen = {
        zeile: zeilenhoehe > 0 ? zeilenhoehe : bisher.zeile,
        sicht,
        oben,
      }

      return bisher.zeile === gemessen.zeile &&
        bisher.sicht === gemessen.sicht &&
        bisher.oben === gemessen.oben
        ? bisher
        : gemessen
    })
  }, [])

  // **Die Horcher hängen an der Referenz, nicht an einem Effekt** (React 19).
  //
  // Ein Effekt daneben bräuchte einen Wächter auf „noch kein Knoten" — einen
  // Zweig, den kein Aufruf betritt, weil React die Referenz vor jedem Effekt
  // setzt. Ein Zweig, den niemand betritt, ist eine Behauptung darüber, was
  // geschieht, und niemand prüft sie nach (§3).
  const behaelter = useCallback(
    (neu: HTMLElement | null) => {
      knoten.current = neu

      if (neu === null) {
        return undefined
      }

      // **Gemessen wird nach dem Zeichnen**: die erste Zeile steht erst dann,
      // und ihre Höhe ist die Grundlage für alles hier.
      nachmessen()

      neu.addEventListener('scroll', nachmessen, { passive: true })

      const wacht = new ResizeObserver(nachmessen)
      wacht.observe(neu)

      return () => {
        neu.removeEventListener('scroll', nachmessen)
        wacht.disconnect()
      }
    },
    [nachmessen],
  )

  const zeilenmass = useCallback(
    (neu: HTMLElement | null) => {
      zeile.current = neu
      nachmessen()
    },
    [nachmessen],
  )

  // **Ohne Maß wird alles gezeichnet**, und das ist kein Notbehelf, sondern
  // der Weg zum ersten Maß: die erste Zeile muss dastehen, damit ihre Höhe
  // messbar ist. Unter jsdom bleibt es dabei — dort hat nichts eine Höhe —,
  // und jede Prüfung sieht weiter die ganze Liste.
  if (mass.zeile <= 0 || mass.sicht <= 0) {
    return {
      behaelter,
      zeilenmass,
      erste: 0,
      letzte: Math.min(anzahl, probezeilen),
      obenHoch: 0,
      untenHoch: 0,
      zeilenhoehe: mass.zeile,
    }
  }

  // **Das Fenster wird an der Liste gekappt, nicht nur an ihrem Ende.**
  //
  // `erste` kommt aus dem Rollstand, und der gehört zu der Liste, die vorhin
  // dastand. Wird sie **kürzer**, ohne dass der Behälter neu entsteht, zeigte
  // `erste` hinter ihr Ende: `slice(992, 107)` ist leer, und der Abstandhalter
  // oben hielte die Höhe von 992 Zeilen — der Rollbalken käme nie zurück.
  //
  // Angestrichen vom Zweitblick über Phase Y als Befund 1. **Sein Auslöser —
  // der Setwechsel — greift nicht**: `Karten.tsx` hängt die ganze Spalte an
  // `karten.daten.length > 0`, und `useLaden` räumt beim Wechsel, also
  // entsteht der Behälter neu (nachgemessen in der laufenden Anwendung: der
  // `<ul>`-Knoten ist danach ein anderer). Die Rechnung bleibt trotzdem
  // ungedeckt, und eine Liste, die **ohne** Neuaufbau schrumpft, träfe sie.
  const sichtbar = Math.ceil(mass.sicht / mass.zeile)
  const fenster = sichtbar + 2 * rand
  const erste = Math.max(
    0,
    Math.min(Math.floor(mass.oben / mass.zeile) - rand, anzahl - fenster),
  )
  const letzte = Math.min(anzahl, erste + fenster)

  return {
    behaelter,
    zeilenmass,
    erste,
    letzte,
    obenHoch: erste * mass.zeile,
    untenHoch: (anzahl - letzte) * mass.zeile,
    zeilenhoehe: mass.zeile,
  }
}
