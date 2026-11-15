// **Die Steuerleiste des Tisches** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Eine eigene Datei
//
// **Der Schnitt folgt der Frage** (§2): `Tischflaeche.tsx` beantwortet, *was*
// auf dem Tisch liegt und wie man es anfasst — hier steht, womit man den
// **Blick** darauf einstellt. Der Anlass war die Zeilengrenze; dass beides
// zusammenfiel, ist der übliche Fall.
//
// # Eine Leiste unter der Leinwand
//
// **Wie im Artboard *Tisch*** — seit dem 2026-09-24 (Befund F09). Sie stand
// mit einer eigenen Überschrift über dem Tisch; die Überschrift steht jetzt
// nur noch für Vorlesehilfen vorn im Abschnitt, und die Ausgabe des
// Tischbilds hängt an einem Knopf dieser Leiste.
//
// # Im schmalen Fenster eine Zeile
//
// **Entscheidung zu #2526**: wo Tisch und Spalte untereinander stehen, brach
// die Leiste in zwei Zeilen um, und weil sie im Tischort liegt, fehlte die
// zweite dem Zug darunter. Dort stehen Hintergrund, eigenes Bild und die
// Ausgabe hinter „Darstellung"; Blick und Ziel bleiben in der Zeile, wie im
// Artboard. **Die Schwelle ist die der Spalten** (`@4xl` am Abschnitt,
// `Partietisch.tsx`) — eine eigene ließe die Leiste neben dem Tisch hinter
// dem Knopf oder darunter wieder umbrechen.
import { useState, type ReactNode } from 'react'

import { Popover } from './Popover'
import { tischbilder } from '../tisch/tischbilder'
import { useTexte } from '../sprache/kontext'
import { useTisch } from '../tisch/useTisch'

type Spieltexte = ReturnType<typeof useTexte>['seiten']['spielen']

export interface Steuerungseigenschaften {
  tisch: ReturnType<typeof useTisch>
  tischbild?: string | undefined
  aufTischbild?: (name: string) => void

  // hatEigenes sagt, ob ein eigenes Bild abgelegt ist.
  hatEigenes?: boolean

  eigenesBild?: () => void
  eigenesBildWeg?: () => void

  // ausgabe sagt, ob die Ausgabe des Tischbilds offensteht;
  // ausgabeUmschalten öffnet und schließt sie.
  ausgabe: boolean
  ausgabeUmschalten: () => void
}

export function Tischsteuerung({
  tisch,
  tischbild,
  aufTischbild,
  hatEigenes = false,
  eigenesBild,
  eigenesBildWeg,
  ausgabe,
  ausgabeUmschalten,
}: Steuerungseigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen
  const mitte = { x: tisch.masse.breite / 2, y: tisch.masse.hoehe / 2 }
  const [darstellung, setzeDarstellung] = useState(false)

  // zielgruppe sind die Sprünge des Blicks — neben dem Tisch in der Zeile,
  // darunter hinter „Darstellung".
  const zielgruppe: ReactNode = (
    <div
      role="group"
      aria-label={eigene.tischZiel}
      className="flex items-center gap-1"
    >
      <Steuerknopf titel={eigene.tischAuswahl} aufDruck={tisch.zeigeAuswahl} />
      <Steuerknopf
        titel={eigene.tischEigenerSitz}
        aufDruck={tisch.zeigeEigenenSitz}
      />
    </div>
  )

  // darstellungsteile sind Hintergrund und eigenes Bild — in der Zeile oder
  // hinter „Darstellung", je nach Platz.
  const darstellungsteile: ReactNode = (
    <>
      {aufTischbild ? (
        <label className="flex items-center gap-2 text-sm">
          <span className="font-medium">{eigene.tischbild}</span>
          <select
            value={tischbild ?? 'holz'}
            onChange={(ereignis) => {
              aufTischbild(ereignis.target.value)
            }}
            className="rounded-sm border border-kante bg-flaeche px-2 py-1 text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            {tischbilder.map((eines) => (
              <option key={eines.name} value={eines.name}>
                {bildwort(eigene, eines.name)}
              </option>
            ))}
          </select>
        </label>
      ) : undefined}

      {eigenesBild ? (
        <Steuerknopf titel={eigene.tischEigenes} aufDruck={eigenesBild} />
      ) : undefined}

      {/* **Wegnehmen steht nur da, wo etwas wegzunehmen ist.** Ein Knopf,
          der nichts tut, sagt nicht, dass nichts da ist — er sagt, dass
          etwas nicht ging. */}
      {eigenesBildWeg && hatEigenes ? (
        <Steuerknopf titel={eigene.tischEigenesWeg} aufDruck={eigenesBildWeg} />
      ) : undefined}
    </>
  )

  return (
    <div className="flex flex-wrap items-center gap-x-4 gap-y-2 rounded-sm border border-kante bg-flaeche px-2 py-1.5">
      {/* **Blick und Ziel brechen je als Ganzes um** (P04, #2167): in einer
          Reihe stand „Auf meinen Sitz" bei schmalem Fenster allein in der
          nächsten Zeile — gemessen am nativen Fenster (`politur`). */}
      <div className="flex flex-wrap items-center gap-1">
        <div
          role="group"
          aria-label={eigene.tischBlick}
          className="flex items-center gap-1"
        >
          <Steuerknopf
            titel={eigene.tischEinpassen}
            aufDruck={tisch.passeEin}
          />
          <Steuerknopf
            titel={eigene.tischNaeher}
            aufDruck={() => {
              tisch.zoome(1.25, mitte)
            }}
          />
          <Steuerknopf
            titel={eigene.tischWeiter}
            aufDruck={() => {
              tisch.zoome(0.8, mitte)
            }}
          />
          <Steuerknopf
            titel={eigene.tischDrehen}
            aufDruck={() => {
              tisch.drehe(90)
            }}
          />
        </div>

        {/* **Ein Strich zwischen Blick und Ziel** — wie im Artboard: links,
            wie man hinsieht, rechts, wohin. **Unter dem Tisch steht das Ziel
            hinter „Darstellung"** (#2526): gemessen bei 1024 × 700 rutschte
            „Darstellung" sonst allein in die zweite Zeile. */}
        <span
          data-tischleiste="ziel"
          className="hidden items-center gap-1 @4xl:flex"
        >
          <span aria-hidden="true" className="mx-1 h-4 w-px bg-kante" />
          {zielgruppe}
        </span>
      </div>

      {/* **Neben dem Tisch in der Zeile** — dort ist Platz, und der Zug
          steht in der Spalte daneben. */}
      <div
        data-tischleiste="breit"
        className="hidden flex-1 flex-wrap items-center gap-x-4 gap-y-2 @4xl:flex"
      >
        <div className="flex flex-wrap items-center gap-1">
          {darstellungsteile}
        </div>

        {/* **`aria-expanded` sagt, ob sie offen ist** — der Knopf heißt in
            beiden Lagen gleich, und nur so hört man den Unterschied. */}
        <span className="ml-auto">
          <Steuerknopf
            titel={eigene.abzugTitel}
            aufDruck={ausgabeUmschalten}
            offen={ausgabe}
          />
        </span>
      </div>

      {/* **Darunter hinter „Darstellung"** (#2526) — die Ausgabe schließt
          das Popover, denn sie klappt unter der Leiste auf, und das Popover
          stünde sonst über ihr. */}
      <span data-tischleiste="schmal" className="ml-auto @4xl:hidden">
        <Popover
          offen={darstellung}
          titel={eigene.tischDarstellung}
          schliessen={() => {
            setzeDarstellung(false)
          }}
          ausloeser={
            <Steuerknopf
              titel={eigene.tischDarstellung}
              aufDruck={() => {
                setzeDarstellung((offen) => !offen)
              }}
              offen={darstellung}
            />
          }
        >
          {zielgruppe}
          {/* **Links ausgerichtet** — gestreckt standen die Knöpfe mittig unter der
              Auswahl. */}
          <div className="flex flex-col items-start gap-2">
            {darstellungsteile}
            <Steuerknopf
              titel={eigene.abzugTitel}
              aufDruck={() => {
                setzeDarstellung(false)
                ausgabeUmschalten()
              }}
              offen={ausgabe}
            />
          </div>
        </Popover>
      </span>
    </div>
  )
}

// bildwort übersetzt den Namen eines Tischbilds.
//
// **Über den Katalog und nicht über eine Tabelle daneben** — derselbe Weg wie
// `zonenwort` (§9): eine zweite Liste liefe auseinander, und die kürzere wäre
// wieder die stumme.
//
// **Und warum es einen Rückfall gibt**: dieselbe Zusage wie bei `zonenwort`
// — ein Name ohne Eintrag lässt die Wahl sonst mit einer leeren Zeile
// dastehen, und die Kennung ist schlechter als ein Wort und besser als
// nichts (§8).
export function bildwort(eigene: Spieltexte, name: string): string {
  const schluessel = `tischbild${name
    .split('-')
    .map((teil) => teil.charAt(0).toUpperCase() + teil.slice(1))
    .join('')}`

  return (eigene as unknown as Record<string, string>)[schluessel] ?? name
}

function Steuerknopf({
  titel,
  aufDruck,
  offen,
}: {
  titel: string
  aufDruck: () => void

  // offen gibt es nur bei einem Knopf, der etwas auf- und zuklappt.
  offen?: boolean
}) {
  return (
    <button
      type="button"
      aria-expanded={offen}
      onClick={aufDruck}
      className="rounded-sm px-2 py-1 text-sm text-marke focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
    >
      {titel}
    </button>
  )
}
