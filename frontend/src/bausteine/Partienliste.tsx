// **Der Weg zurück in die eigene Partie** —
// #1600.
//
// # Warum ein Sitz gewählt wird und nicht einer geraten
//
// **Weil man an einem Tisch, den man verlassen hat, den Platz wechseln
// können muss.** Beim **Anfangen** entscheidet das gezeigte Profil (G8,
// #1465) — dort hat man die Sitze gerade selbst gebunden. Beim
// Wiedereinsteigen liegt der Fall anders: die Partie kann aus einem anderen
// Profil stammen, mit Namen, die man vor zwei Tagen getippt hat. Ein Weg, der
// dann stillschweigend den Wirt nähme, setzte jemanden an einen Platz, den er
// nicht gewählt hat.
//
// # Und warum die Liste keine Karten zeigt
//
// **Weil eine Liste keine Sicht ist.** Was in einer Partie liegt, rechnet die
// Projektion je Sitz (G1); diese Liste nennt Sitze, Stand und die **Zahl** der
// Objekte — keinen Kartennamen.
import { Ladefehler, Wirdgeladen } from './Ladehinweise'
import { useMemo, useState } from 'react'

import { Faltliste } from './Faltliste'
import { Knopf } from './Knopf'
import { Partienauswahlleiste, keineAuswahl } from './Partienauswahlleiste'
import { Partienzeile } from './Partienzeile'
import { Verwaltungsmeldung } from './Partienverwaltung'
import { Einlesekennung } from '../haken/usePartieverwaltung'
import { partienauswahl, staendeIn } from './partienauswahl'
import { useSprache, useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Markenfilterknopf } from './Markenfilterkasten'
import { nachMarken, useMarkenfilter } from '../haken/useMarkenfilter'

import type { Partiezeile } from '../apipartie'
import type { Verwaltungsstand } from '../haken/usePartieverwaltung'

// Vorn ist, wie viele Partien zu Anfang dastehen, und Schritt, wie viele
// „Weitere zeigen" dazulegt — UXR-13
// (#1921).
//
// **Vier, weil R2 zwanzig gleichrangige Bedienelemente im Bild zulässt**
// (`docs/UIREVIEW.md`): neben der Liste zeichnet die Seite neun — Kopf,
// Sitze, Anfang —, die Liste um ihre Zeilen drei — „Weitere zeigen", das
// Einlesen und seine Hilfe —, und jede Zeile trägt zwei. Zusammen zwanzig,
// gezählt unter jsdom mit `test/spielbestand.ts`; mit Marken und einer Auswahl
// kommen der Filter und das Zurücksetzen dazu. Die Reihenfolge ist
// die des Bestands — die zuletzt geänderte zuerst ([app.Partien]) —, vorn
// steht also, woran zuletzt gespielt wurde.
const Vorn = 4
const Schritt = 10

interface Listeneigenschaften {
  partien: readonly Partiezeile[]

  // geladen und wiederholen trennen „noch nicht da" von „keine" — UX-08.NAT-01
  // und S15.NAT-02: vorher behauptet die Liste nichts, und ein Fehlschlag
  // bietet den nächsten Versuch an.
  geladen: boolean
  fehler: string | undefined
  wiederholen: () => void

  // einsteigen nimmt Partie und Sitz — in dieser Reihenfolge, wie `ansehen`.
  einsteigen: (partieID: string, sitz: string) => void

  // verwaltung ist, was mit einer Partie als Ganzem geht (PLAY-03.AK-05,
  // #1711).
  verwaltung: Verwaltungsstand
}

export function Partienliste({
  partien,
  geladen,
  fehler,
  wiederholen,
  einsteigen,
  verwaltung,
}: Listeneigenschaften) {
  const texte = useTexte()
  const sprache = useSprache()
  const eigene = texte.seiten.spielen

  const [auswahl, setzeAuswahl] = useState(keineAuswahl)

  // **Marken an Partien** — UX-12 (#1684): gefiltert wird nach der Wahl
  // der Leiste und danach nach dem, was Go für die Marken nennt.
  const marken = useMarkenfilter('partie')
  const { getroffen } = marken
  const gezeigt = useMemo(
    () =>
      nachMarken(
        partienauswahl(partien, auswahl, sprache),
        (partie) => partie.id,
        getroffen,
      ),
    [partien, auswahl, getroffen, sprache],
  )

  const staende = useMemo(() => staendeIn(partien), [partien])

  const markiert = Object.values(marken.je).some(
    (vergeben) => vergeben.length > 0,
  )

  return (
    <section className="flex flex-col gap-3">
      <h2 className="text-lg font-semibold">{eigene.partienTitel}</h2>

      {fehler !== undefined && (
        <Ladefehler fehler={fehler} wiederholen={wiederholen} />
      )}

      {!geladen && fehler === undefined && <Wirdgeladen />}

      {/*
        **Die Leiste erscheint ab zwei Partien** — über einer einzigen Zeile
        kann sie nur etwas wegnehmen, und drei Bedienelemente darüber wären
        genau der aufgeblähte Bildschirm aus #1771.
      */}
      {partien.length > 1 && (
        <Partienauswahlleiste
          auswahl={auswahl}
          setzeAuswahl={setzeAuswahl}
          staende={staende}
          gezeigt={gezeigt.length}
          alle={partien.length}
        />
      )}

      {/*
        **Der Markenfilter erst, wenn eine Partie eine Marke trägt** — UXR-13.
        Vorher kann er nur wegnehmen, wie die Leiste über einer einzigen
        Partie. **Und solange eine Marke gewählt ist**, damit sich die Wahl
        zurücknehmen lässt, auch wenn die letzte Marke inzwischen fort ist.
      */}
      {partien.length > 0 && (markiert || marken.filter.marken.length > 0) && (
        <Markenfilterknopf art="partie" stand={marken} />
      )}

      {geladen && leerzustand(partien.length, gezeigt.length, eigene)}

      {gezeigt.length > 0 && (
        <Faltliste
          eintraege={gezeigt}
          vorn={Vorn}
          schritt={Schritt}
          klasse="flex flex-col gap-2"
          zeile={(partie) => (
            <Partienzeile
              key={partie.id}
              partie={partie}
              einsteigen={einsteigen}
              verwaltung={verwaltung}
              marken={marken.je[partie.id] ?? []}
              nachladen={marken.nachladen}
            />
          )}
        />
      )}

      {/*
        **Der Satz und das Einlesen stehen unter der Liste** — UXR-13, wie im
        Artboard: was man sucht, ist eine Partie, und sie soll gleich unter
        der Überschrift stehen. Das Einlesen gehört zu keiner Partie, sondern
        legt eine neue an; es steht auch dann da, wenn die Liste leer ist —
        genau dann braucht man es.
      */}
      <Erklaerungsabsatz
        text={eigene.partienErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      <div className="flex flex-wrap items-center gap-3">
        <Knopf
          gewicht="leise"
          schreibt
          angehalten={verwaltung.laeuft(Einlesekennung)}
          aufKlick={verwaltung.einlesen}
        >
          {eigene.partienEinlesen}
        </Knopf>

        <Erklaerungsabsatz
          text={eigene.partienEinlesenErklaerung}
          klasse="max-w-prose text-xs text-leise"
        />

        <Verwaltungsmeldung kennung={Einlesekennung} lauf={verwaltung} />
      </div>
    </section>
  )
}

// leerzustand sagt, warum nichts dasteht — **und es sind zwei Gründe.**
//
// **„Du hast keine Partie" und „keine passt dazu" sind nicht dasselbe** (§6):
// der erste nennt den nächsten Schritt, der zweite die Auswahl, die ihn
// verdeckt. Wer beide in einen Satz legte, schickte jemanden zum Anlegen, der
// neunzehn Partien hat.
function leerzustand(
  alle: number,
  gezeigt: number,
  eigene: { partienLeer: string; partienNichtsGefunden: string },
) {
  if (gezeigt > 0) {
    return undefined
  }

  return (
    <p className="text-sm text-leise">
      {alle === 0 ? eigene.partienLeer : eigene.partienNichtsGefunden}
    </p>
  )
}
