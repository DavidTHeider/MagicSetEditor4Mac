// **Der gezeichnete Tisch** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Eine Wahrheit, zwei Erscheinungen
//
// **Die Knöpfe über der Leinwand sind keine Verzierung — sie sind der Tisch.**
// Was PixiJS zeichnet, ist ein Bild; was hier im DOM steht, ist das, womit
// bedient wird: ein Knopf je Karte, an der Stelle, an der die Karte gerade
// liegt, mit ihrem Namen, ihrer Zone und ihrem Zustand als Beschriftung.
//
// **Und genau deshalb bleibt der Tisch ohne Grafikkarte bedienbar.** Fällt
// der Renderer aus, werden dieselben Knöpfe sichtbar gesetzt statt
// durchsichtig — kein zweiter Weg, keine zweite Fachimplementierung (#1460:
// „unsichtbare semantische DOM-Overlays sind zulässig, keine zweite
// Karten-Fachimplementierung").
import type { Menueeintrag } from './Menueliste'
import type { Klick } from '../spiel/kartengriff'
import { useZugklick } from '../haken/useZugklick'
import { useCallback, useId, useMemo, useRef, useState } from 'react'

import { bewegungErwuenscht } from '../tisch/bewegung'
import { mitZug } from '../tisch/tischplan'
import { drehung, Kartenknopf, Kartenmenue, stelle } from './Tischkartenknopf'
import { objektAn, zoneAn } from '../tisch/treffer'
import { useTexte } from '../sprache/kontext'
import { Tischabzugfeld } from './Tischabzugfeld'
import { Tischsteuerung } from './Tischsteuerung'
import { Tischzonenmarken } from './Tischzonenmarken'
import { useAbzugOhneZug, useZugende } from './tischzugende'
import { useTisch } from '../tisch/useTisch'
import { kurzzonenwort, zonenwort } from '../seiten/partiewoerter'

import type { Bildquelle } from '../tisch/maler'
import type { Tischanschluss } from '../tisch/useTisch'
import type { Tischplan } from '../tisch/tischplan'
import type { Wirkungsart } from '../tisch/tischwirkungen'
import {
  angefasst,
  geschoben,
  losgelassen,
  type Ziehen,
  type Zugziel,
} from './tischziehen'
import { Abschnittstitel } from './Abschnittstitel'
import { amZeiger, griff, hindernissatz } from './tischflaechenhilfen'

export interface Tischflaecheneigenschaften {
  plan: Tischplan
  bild: Bildquelle
  malerei: Tischanschluss['malerei']

  gewaehlt?: string
  tischbild?: string
  aufTischbild?: (name: string) => void

  // hintergrund ist die Adresse eines eigenen Tischbilds.
  hintergrund?: string

  // eigenesBild öffnet den Dateidialog — fehlt es, gibt es die Wahl nicht.
  eigenesBild?: () => void

  // eigenesBildWeg nimmt das eigene Bild zurück.
  eigenesBildWeg?: () => void

  // sitzkopf gibt, was am Sitz auf der Leinwand steht — PLAY-03.
  //
  // **Gereicht und nicht gebaut.** Diese Fläche bekommt einen **Plan** und
  // keine Partie; Namen, Leben und Gift liegen dort, wo das Spiel liegt. Wer
  // es weglässt, bekommt einen Tisch ohne Sitzköpfe — den Zustand vor
  // PLAY-03, und keinen kaputten.
  sitzkopf?: (sitz: string) => string

  // beschriftungsstufe ist der Wert von `tisch.beschriftung` — ob und wie
  // groß Zonennamen und Sitzköpfe auf der Leinwand stehen.
  beschriftungsstufe?: string

  // wirkungen sagt je Kennung, was gerade geschehen ist — PLAY-04.AK-02
  // (#1712).
  wirkungen?: Map<string, Wirkungsart>

  // bewegt sagt, ob Karten fliegen sollen — die Systemeinstellung hat die
  // stärkere Stimme, und das entscheidet `bewegungErwuenscht`.
  bewegt?: boolean
  aufKarte?: (kennung: string, klick: Klick) => void

  // menueFuer gibt das Kontextmenü einer Karte — Rechtsklick und ⇧F10, wie
  // in der Liste (P4, #2372). Ohne Menü steht der Knopf allein.
  menueFuer?: (
    kennung: string,
  ) => { bezeichnung: string; eintraege: Menueeintrag[] } | undefined

  // tastenFuer und taste sind die Einzeltasten einer Karte (C-09).
  tastenFuer?: (kennung: string) => string
  taste?: (kennung: string, taste: string) => boolean

  // aufZug meldet, dass jemand eine Karte in eine andere Zone gezogen hat.
  //
  // **Und seit dem 2026-09-21 auch, dass er sie auf eine andere gelegt oder
  // aus ihrem Haufen genommen hat** — CR 400.5
  // (#1756).
  // Was gemeint ist, sagt `Zugziel.art`; die Fläche entscheidet es nicht.
  //
  // **Erst beim Loslassen und nur dann** — #1460: „Pointer-Drag bewegt nur
  // Darstellung; Drop sendet validierten Command."
  aufZug?: (ziel: Zugziel) => void

  // ablegen sagt, dass der Tisch **nicht** selbst ordnet — dann hält ein
  // Zug über freien Platz die Lage fest (PLAY-04.AK-03).
  ablegen?: boolean | undefined

  // sicht ist der Sitz, aus dem der Plan gerechnet ist; an sie bindet das
  // Ausgabefeld seine Vorschau (PLAY-05.AK-02).
  sicht?: string
}

export function Tischflaeche({
  plan: roherPlan,
  bild,
  malerei,
  gewaehlt,
  tischbild,
  aufTischbild,
  hintergrund,
  eigenesBild,
  eigenesBildWeg,
  sitzkopf,
  beschriftungsstufe,
  wirkungen,
  bewegt = true,
  aufKarte,
  menueFuer,
  tastenFuer,
  taste,
  aufZug,
  ablegen,
  sicht = '',
}: Tischflaecheneigenschaften) {
  const [ziehen, setzeZiehen] = useState<Ziehen | undefined>(undefined)
  const zugklick = useZugklick()

  // **Der gezogene Plan geht an den Maler, der rohe an den Wirt.** Was
  // während des Ziehens zu sehen ist, hat niemand außer diesem Fenster
  // vereinbart.
  const plan = useMemo(
    () => (ziehen ? mitZug(roherPlan, ziehen) : roherPlan),
    [roherPlan, ziehen],
  )
  const texte = useTexte()
  const eigene = texte.seiten.spielen
  const ueberschrift = useId()
  const leinwand = useRef<HTMLDivElement | null>(null)
  // **Die Systemeinstellung hat die stärkere Stimme** — wer
  // `prefers-reduced-motion` setzt, bekommt keine Bewegung, auch wenn die
  // Anwendung sie anböte.
  //
  // **Gefragt wird das Fenster und nicht die Leinwand** — deren Referenz im
  // Render zu lesen ist genau das, was React verbietet, und die Antwort wäre
  // dieselbe: `prefers-reduced-motion` gilt für das Fenster.
  const fliegt = bewegungErwuenscht(globalThis.window, bewegt)

  // **Die Zonennamen kommen aus demselben Katalog wie überall** —
  // PLAY-03.AK-06
  // (#1711).
  //
  // **Gehalten, weil der Zeichner daran hängt**: `useTisch` zeichnet in einem
  // Effekt, und eine bei jedem Bild neue Funktion machte aus jeder Eingabe
  // einen neuen Lauf. Dieselbe Überlegung wie bei `plan` darüber.
  // **Über einem Stapel die Kurzform** (`zonenschilder.ts`).
  const beschriftung = useCallback(
    (zone: string, kurz?: boolean) =>
      kurz === true ? kurzzonenwort(eigene, zone) : zonenwort(eigene, zone),
    [eigene],
  )

  const tisch = useTisch({
    plan,
    bild,
    malerei,
    gewaehlt,
    tischbild,
    hintergrund,
    bewegt: fliegt,
    beschriftung,
    sitzkopf,
    beschriftungsstufe,
    wirkungen,
    leinwand,
  })
  const { kamera, masse } = tisch

  // **Ein Zug endet auch neben dem Tisch, und währenddessen entsteht kein
  // Tischbild** — Befund 3 des Zweitblicks über Phase 6.9 (`tischzugende.ts`).
  const zieht = ziehen !== undefined
  useZugende(zieht, setzeZiehen)
  const abzug = useAbzugOhneZug(zieht, tisch.abzug)

  // **Ohne Renderer stehen die Knöpfe sichtbar da.** Sie tragen dann keine
  // Stelle auf der Leinwand mehr, sondern fließen als gewöhnliche Liste —
  // dieselben Knöpfe, dieselben Namen, dieselbe Bedienung.
  const gezeichnet = tisch.hindernis === undefined

  // **Die Ausgabe steht hinter einem Knopf** — Befund F09 (Artboard
  // *Tisch*). Offen zwischen Steuerung und Leinwand nähme sie dem Tisch
  // Zeilen, auch wenn niemand ein Bild will.
  const [ausgabe, setzeAusgabe] = useState(false)

  return (
    // **Die Fläche füllt, was ihr Behälter ihr gibt** — Befund F09: die
    // Leinwand wächst, Steuerung und Ausgabe stehen darunter, wie im
    // Artboard *Tisch*.
    <section
      aria-labelledby={ueberschrift}
      className="flex min-h-0 flex-1 flex-col gap-2"
    >
      {/* **Die Überschrift steht vorn und nur für Vorlesehilfen** — das
          Artboard zeichnet keine; sie benennt den Abschnitt, und sie gehört
          vor seinen Inhalt und nicht unter die Leinwand. */}
      <Abschnittstitel id={ueberschrift} className="sr-only">
        {eigene.tischBild}
      </Abschnittstitel>

      {tisch.hindernis ? (
        <p
          role="alert"
          className="max-w-prose rounded-sm border border-kante bg-betont p-3 text-sm"
        >
          {hindernissatz(eigene, tisch.hindernis)}
        </p>
      ) : undefined}

      {/* **Eine Spielfläche mit eigener Tastaturbedienung.** `application`
          ist die Rolle, die ARIA genau dafür kennt, und sie gehört auf einen
          Behälter — eine Leinwand nimmt keinen Fokus, und ein Knopf ist nicht
          die Fläche. Dieselbe Ausnahme und derselbe Grund wie in
          `Dialog.tsx`: die Rolle trägt, was die Tastatur erwartet. */}
      {/* eslint-disable-next-line jsx-a11y/no-noninteractive-element-interactions */}
      <div
        role="application"
        // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
        tabIndex={0}
        aria-label={eigene.tischBild}
        // **Gescrollt wird nur, wo nicht gezeichnet wird**: ohne Leinwand
        // stehen die Knöpfe sichtbar untereinander, und eine feste Höhe
        // schnitte die hinteren ab.
        className={
          gezeichnet
            ? 'relative min-h-40 flex-1 overflow-hidden rounded-sm border border-kante focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'
            : 'relative min-h-40 flex-1 overflow-y-auto rounded-sm border border-kante focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'
        }
        onKeyDown={(ereignis) => {
          griff(ereignis.key, tisch)
        }}
        onPointerDown={(ereignis) => {
          const ort = amZeiger(ereignis, kamera, masse)
          zugklick.gedrueckt(ereignis)

          setzeZiehen(angefasst(ereignis.button, objektAn(roherPlan, ort), ort))
        }}
        onPointerMove={(ereignis) => {
          if (!ziehen) return

          setzeZiehen(geschoben(ziehen, amZeiger(ereignis, kamera, masse)))
        }}
        onPointerUp={(ereignis) => {
          if (!ziehen) return

          // **Erst zurücksetzen, dann melden.** Was danach zu sehen ist,
          // kommt vom Wirt — und wenn er ablehnt, liegt die Karte wieder da,
          // wo sie war, ohne dass jemand sie zurückrechnen müsste (#1460:
          // „Abgelehnter Drop stellt nur das betroffene Objekt zurück").
          setzeZiehen(undefined)
          if (!zugklick.losgelassen(ereignis)) return

          const ort = amZeiger(ereignis, kamera, masse)

          const ziel = losgelassen(
            roherPlan,
            ziehen,
            zoneAn(roherPlan, ort),
            objektAn(roherPlan, ort),

            // **Wo der Tisch nicht selbst ordnet, bleibt die Karte liegen**
            // — PLAY-04.AK-03
            // (#1712).
            ablegen,
          )

          // **⌥ heißt frei** (P5, #2373) — der Kern schiebt dann, statt zu
          // deuten, wo das Profil es erlaubt.
          if (ziel) {
            aufZug?.(
              ziel.art === 'zone' && ereignis.altKey
                ? { ...ziel, frei: true }
                : ziel,
            )
          }
        }}
        onPointerCancel={() => {
          setzeZiehen(undefined)
        }}
      >
        {/* **Der Ort der Leinwand, nicht sie selbst** — UI-N01 (#2089): die
            Leinwand legt `tischmalerei` je Anlauf hinein und nimmt sie beim
            Abräumen wieder heraus. React hält nur, was PixiJS nie anfasst. */}
        <div
          ref={leinwand}
          aria-hidden="true"
          className="block h-full w-full"
        />

        {gezeichnet ? (
          <Tischzonenmarken
            felder={plan.felder}
            kamera={kamera}
            masse={masse}
          />
        ) : undefined}

        <ul
          className={
            gezeichnet
              ? 'absolute inset-0 m-0 list-none p-0'
              : 'flex flex-wrap gap-2 p-3'
          }
        >
          {/* **Im gezeichneten Stapel nur die oberste** (#2472): die übrigen
              lägen 0,6 Punkte darunter als verdeckte Ziele; die Liste ohne
              Zeichnung führt jede einzeln. */}
          {plan.lagen
            .filter((lage) => !gezeichnet || !lage.untergelegt)
            .map((lage) => (
              <li
                key={lage.kennung}
                style={gezeichnet ? stelle(lage, kamera, masse) : undefined}
                className={gezeichnet ? 'absolute' : undefined}
              >
                <Kartenmenue
                  menue={menueFuer?.(lage.kennung)}
                  gezeichnet={gezeichnet}
                >
                  <Kartenknopf
                    lage={lage}
                    eigene={eigene}
                    gezeichnet={gezeichnet}
                    gewaehlt={gewaehlt === lage.kennung}
                    aufDruck={(klick) => {
                      if (zugklick.nimmKlick()) aufKarte?.(lage.kennung, klick)
                    }}
                    tasten={tastenFuer?.(lage.kennung) ?? ''}
                    aufTaste={(wort) => taste?.(lage.kennung, wort) ?? false}
                    {...(gezeichnet ? { gedreht: drehung(lage, kamera) } : {})}
                  />
                </Kartenmenue>
              </li>
            ))}
        </ul>
      </div>

      <Tischsteuerung
        tisch={tisch}
        tischbild={tischbild}
        hatEigenes={hintergrund !== undefined && hintergrund !== ''}
        ausgabe={ausgabe}
        ausgabeUmschalten={() => {
          setzeAusgabe((offen) => !offen)
        }}
        {...(aufTischbild ? { aufTischbild } : {})}
        {...(eigenesBild ? { eigenesBild } : {})}
        {...(eigenesBildWeg ? { eigenesBildWeg } : {})}
      />

      {/* **Die Ausgabe steht bei der Steuerung** — PLAY-05.AK-03
          (#1713):
          beide beantworten dieselbe Frage, nämlich was man vom Tisch sieht.
          Der Abzug nimmt den Blick, den die Steuerung eingestellt hat. */}
      {ausgabe ? (
        <Tischabzugfeld plan={plan} abzug={abzug} sicht={sicht} />
      ) : undefined}

      {plan.heimatlos.length > 0 ? (
        <p className="max-w-prose text-sm text-leise">
          {eigene.tischHeimatlos} {plan.heimatlos.join(', ')}
        </p>
      ) : undefined}
    </section>
  )
}
