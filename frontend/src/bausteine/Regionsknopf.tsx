// **Der Knopf über einer Spalte des Arbeitsplatzes** — UXR-02
// (#1910).
//
// # Warum ein eigenes Bauteil
//
// **Weil `Arbeitsplatz.tsx` mit ihm über die Zeilengrenze ginge** (§2). Der
// Schnitt folgt der Frage: dort steht, **wann** eine Spalte steht und wohin
// sie weicht; hier, **wie** ihr Knopf aussieht und was er dazu sagt.
//
// # Warum klein, und wo der Satz steht
//
// Das Artboard *Karten* zeichnet über jeder Spalte nur einen kleinen Kopf
// mit ihrem Namen und ihrem Kürzel (`region-kopf`), und *Enge* lässt den
// Inspector „an der Kante warten". Bis UXR-02 stand darüber eine eigene Zeile
// mit zwei Knöpfen in voller Größe, und im Mindestfenster dazu der Satz „Zu
// schmal für alle drei Spalten …", der die Zeile in zwei umbrach — rund 88
// Punkte über einem Zentrum, das auf der Kartenseite gerechnete 265 hatte.
//
// **Der Satz bleibt, und er beschreibt jetzt den Knopf**: die Vorlesehilfe
// liest ihn als Beschreibung (`aria-describedby` auf einen verborgenen
// Knoten), der Zeiger sieht ihn im Titel. Ohne Enge nennt der Titel das
// Kürzel — wie der Knopf der Seitenleiste „(⌥⌘S)".
//
// # Warum kein `Tooltip`
//
// **Er verbraucht Escape, solange er steht** (WCAG 1.4.13). Wer die Spalte
// über dem Zentrum mit der Maus öffnet, dem steht er danach unter dem
// ruhenden Zeiger — und das erste Escape schloss den Hinweis statt der
// Spalte. `arbeitsplatzenge.test.tsx` wurde damit rot.
import { useId, type RefObject } from 'react'

import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'

import type { Arbeitsplatzregion } from '../haken/useArbeitsplatzgedaechtnis'

// kuerzel ist je Region die Taste, wie macOS sie schreibt und wie
// `aria-keyshortcuts` sie liest. **Gedeutet wird sie in `rahmen/tastatur.ts`**;
// `arbeitsplatzleiste.test.tsx` schickt beide Schreibweisen durch die
// Deutung, damit hier keine Kopie entsteht, die niemand nachführt.
export const regionskuerzel: Readonly<
  Record<Arbeitsplatzregion, { zeichen: string; tasten: string }>
> = {
  navigator: { zeichen: '⌥⌘[', tasten: 'Alt+Meta+[' },
  inspector: { zeichen: '⌥⌘]', tasten: 'Alt+Meta+]' },
}

interface RegionsknopfEigenschaften {
  region: Arbeitsplatzregion

  // offen sagt, ob die Spalte steht — neben dem Zentrum oder über ihm.
  offen: boolean

  // eng sagt, dass diese Spalte bei der Breite weicht.
  eng: boolean

  griff: RefObject<HTMLButtonElement | null>
  steuert: string
  angehalten: boolean
  aufKlick: () => void
}

export function Regionsknopf({
  region,
  offen,
  eng,
  griff,
  steuert,
  angehalten,
  aufKlick,
}: RegionsknopfEigenschaften) {
  const texte = useTexte().bausteine.arbeitsplatz
  const { zeichen, tasten } = regionskuerzel[region]
  const navigator = region === 'navigator'

  const aufschrift = navigator
    ? offen
      ? texte.navigatorZu
      : texte.navigatorAuf
    : offen
      ? texte.inspectorZu
      : texte.inspectorAuf
  const satz = navigator ? texte.navigatorEng : texte.inspectorEng

  const satzID = useId()

  return (
    <>
      <Knopf
        gewicht="leise"
        klein
        griff={griff}
        titel={`${eng ? satz : aufschrift} (${zeichen})`}
        beschriebenVon={eng ? satzID : undefined}
        kuerzel={tasten}
        ausgeklappt={offen}
        steuert={steuert}
        angehalten={angehalten}
        aufKlick={aufKlick}
      >
        {aufschrift}
      </Knopf>

      {/* **Verborgen und trotzdem gelesen**: eine Beschreibung darf auf
          einen verborgenen Knoten zeigen, und so steht der Satz nicht in der
          Zeile. */}
      {eng && (
        <span id={satzID} hidden>
          {satz}
        </span>
      )}
    </>
  )
}
