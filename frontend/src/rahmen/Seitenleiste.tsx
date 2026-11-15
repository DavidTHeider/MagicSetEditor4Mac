// **Die Seitenleiste ist eine eigene Fläche.** Sie steht auf dem
// Anwendungsgrund wie ein Blatt darauf; ohne diesen Unterschied hält eine
// Trennlinie allein zwei Bereiche auseinander — und eine Linie ist keine
// Ordnung
// (#288).
//
// **Aus `Rahmen.tsx` herausgezogen**, weil die Datei an der Zeilengrenze
// stand (§2) — derselbe Grund wie bei `useLeistenhalt` und `useLeistenfokus`.
// Der Schnitt folgt der Frage: *was* der Rahmen tut, steht dort; *wie die
// Leiste aussieht*, steht hier.
import type { RefObject } from 'react'

import type { Seitenschluessel } from '../navigation'
import { Leistenzone } from './Leistenzone'
import { useVerborgeneSeiten } from '../abschaltung'
import { useSchalterreife } from '../haken/useSchalterreife'
import { useLeistenabstand } from '../haken/useLeistenabstand'
import { leisteneintraege, mitBelegung } from '../leistenordnung'
import { vorgabebelegung, type Ziffernbelegung } from './ziffernbelegung'
import { useTexte } from '../sprache/kontext'
import type { RovierendeZeile } from '../haken/useRovierendeZeile'
import { Kontextleiste } from './Kontextleiste'

// **Die Klassen stehen benannt und nicht als Ausdruck im JSX.** Ein
// Zeichenkettenausdruck dort ist ein Befund — `keintextimjsx.test.ts`
// unterscheidet nicht, ob er einen Satz oder Klassennamen trägt, und das
// ist richtig so: er kann es nicht.
const leistengrund =
  'flex w-60 shrink-0 flex-col border-r border-kante bg-flaeche'

// **Im Fluss steht sie fest und ist so hoch wie das Fenster** — UX-02
// (#1674).
//
// # Was das behebt
//
// **Der Fuß mit ⌘K stand am Ende des Dokuments statt am Fuß der Leiste.** Der
// Rahmen ist `min-h-screen`, und die Leiste ist sein Flexkind: ohne eigene
// Höhe wuchs sie mit dem Inhalt mit. Gemessen am 2026-09-21 an der laufenden
// Anwendung auf der Startseite bei 1.280 × 800: die Leiste war **24.875**
// Bildpunkte hoch, und „Befehlspalette öffnen" saß bei y = 24.809. Sichtbar
// war es nur, solange ein Dialog offen war und den Hintergrund sperrte — der
// hielt die Seite auf Fensterhöhe.
//
// **Kein Test hat das gesehen und keiner konnte es**: jsdom rechnet kein
// Layout (`test/einrichtung.ts`). Das ist die Sorte Befund, für die §3 die
// Sichtprüfung verlangt — und sie hat ihn beim ersten Blick gefunden.
//
// **`sticky` und nicht `fixed`**: die Leiste bleibt ein Flexkind und behält
// ihre Spalte; `fixed` nähme sie aus dem Fluss, und der Inhalt rückte darunter.
const imFluss = `${leistengrund} sticky top-0 h-screen`

// **Über dem Inhalt statt neben ihm**, sobald das Fenster schmal ist: so
// bleibt dem Inhalt die volle Breite, und die Leiste legt sich für den
// Augenblick darüber.
// **Als Schublade trägt sie eine kräftigere Kante statt eines Schattens** —
// Vermerk 0065, Festlegung 3, und Festlegung 9 sagt es für den Drawer
// ausdrücklich.
//
// **`border-r-2` und nicht ein zweiter Ton**: die Leiste liegt über dem
// Inhalt, und was sie von ihm trennt, ist ihre Kante — im Entwurf steht
// `border-left: 2px solid var(--kante)` am Drawer, hier spiegelbildlich.
// **Als Schublade ist sie am Fenster festgemacht**, nicht am Rahmen: `fixed`
// misst gegen das Fenster, `absolute` gegen den Rahmen — und der ist so hoch
// wie der Inhalt. Derselbe Befund wie oben, eine Zeile tiefer.
const alsSchublade = `${leistengrund} fixed inset-y-0 left-0 z-10 border-r-2`

// **Oben steht sie als Zeile über dem Inhalt** — UX-14.AK-02
// (#1686):
// *„horizontale Anordnung ist ebenfalls möglich und erhält dieselben
// Aktionen, Kürzel, Fokuspfade und Aktivmarkierungen."* **Dieselben Teile in
// derselben Reihenfolge** — Name, wer arbeitet, die Bereiche, die Palette —,
// damit Tab denselben Weg nimmt. **Nie als Schublade**: eine Zeile nimmt dem
// Inhalt keine Breite, und schmal wird sie waagerecht rollbar.
const alsZeile =
  'sticky top-0 z-10 flex w-full shrink-0 flex-wrap items-stretch border-b border-kante bg-flaeche'
// **`py-3` und nicht `py-4`** — UXR-02 (#1910): die senkrechte Leiste rollte
// im Mindestfenster (`leistenmass.test.tsx`), und das Artboard setzt den
// Namen mit 12 und 10 Punkten Rand (`.nav-kopf`).
const namensklassen =
  'border-b border-kante px-4 py-3 text-sm font-semibold tracking-tight'
// **Oben trägt die Zeile den Namen nur für den Vorleser**: sichtbar steht er
// in der Titelleiste und in der Menüleiste, und die Zeile braucht die Breite.
const namensklassenQuer = 'sr-only'
const navigationSenkrecht = 'flex-1 overflow-y-auto p-3'
// **Umbrechen statt rollen** (UX-14.AK-03): ein Eintrag hinter dem Rand
// wäre abgeschnitten, und das Kriterium verlangt, dass nichts es ist.
const navigationQuer = 'flex flex-1 flex-wrap items-center px-2'
const fussSenkrecht = 'border-t border-kante p-3'
const fussQuer = 'flex items-center border-l border-kante px-2'

interface Seitenleisteneigenschaften {
  aktuell: Seitenschluessel
  wechseln: (seite: Seitenschluessel) => void
  offen: boolean

  // schmal sagt, dass die Leiste **über** dem Inhalt liegt statt neben ihm —
  // #1630.
  // Im Fluss nähme sie dem Inhalt bei 370 Bildpunkten alles bis auf 89.
  schmal: boolean

  bereich: RefObject<HTMLDivElement | null>
  halt: RovierendeZeile
  paletteOeffnen: () => void

  // ziffern ist die Belegung von ⌘0 bis ⌘9 — die Leiste zeigt, was sie sagt
  // (UX-02.AK-03, #1674).
  ziffern?: Ziffernbelegung | undefined

  // waagerecht legt die Leiste oben als Zeile (UX-14.AK-02, #1686).
  waagerecht?: boolean
}

export function Seitenleiste({
  aktuell,
  wechseln,
  offen,
  schmal,
  bereich,
  halt,
  paletteOeffnen,
  ziffern = vorgabebelegung,
  waagerecht = false,
}: Seitenleisteneigenschaften) {
  const texte = useTexte()
  const namen = texte.allgemein.bereiche

  // **Oben hält sie dem Fokus den Platz frei**, den sie verdeckt.
  useLeistenabstand(bereich, waagerecht)

  // **Zwei Zonen, ein Trenner dazwischen** — der Entwurf zeichnet es so, und
  // UX-02.AK-04 verlangt es: die Dienstzone steht erreichbar da, **ohne
  // einen fachlichen Hauptbereich zu verbrauchen**.
  //
  // **Und ohne, was abgeschaltet ist** (OPS-10, #1908): ein Eintrag, dessen
  // Ziel verborgen ist, steht nicht da.
  const verborgen = useVerborgeneSeiten()
  const reife = useSchalterreife()
  const eintraege = mitBelegung(leisteneintraege, ziffern).filter(
    (eintrag) => !verborgen.has(eintrag.ziel),
  )
  const fachlich = eintraege.filter((eintrag) => eintrag.alsBereich)
  const dienste = eintraege.filter((eintrag) => !eintrag.alsBereich)

  return (
    <div
      id="seitenleiste"
      ref={bereich}
      data-lage={waagerecht ? 'waagerecht' : 'senkrecht'}
      hidden={!offen && !waagerecht}
      className={waagerecht ? alsZeile : schmal ? alsSchublade : imFluss}
    >
      {/* **Der Name der Anwendung steht oben links** — die Stelle, an der man
          ihn unter macOS sucht. */}
      <p className={waagerecht ? namensklassenQuer : namensklassen}>
        {texte.allgemein.anwendungsname}
      </p>

      {/* **Wer wo als was arbeitet** — Werkstück U4 von Phase 5.1 (#1415):
          die Leiste gehört zum Rahmen wie die Seitenliste, denn sie gilt auf
          jeder Seite. */}
      <Kontextleiste wechseln={wechseln} waagerecht={waagerecht} />

      <nav
        aria-label={texte.allgemein.navigation.bezeichnung}
        className={waagerecht ? navigationQuer : navigationSenkrecht}
      >
        <Leistenzone
          eintraege={fachlich}
          reife={reife}
          ueberschrift={namen.ueberschrift}
          trenner={false}
          versatz={0}
          gesamt={leisteneintraege.length}
          aktuell={aktuell}
          wechseln={wechseln}
          halt={halt}
          waagerecht={waagerecht}
        />

        <Leistenzone
          eintraege={dienste}
          ueberschrift={namen.programm}
          trenner
          versatz={fachlich.length}
          gesamt={leisteneintraege.length}
          aktuell={aktuell}
          wechseln={wechseln}
          halt={halt}
          waagerecht={waagerecht}
        />
      </nav>

      <div className={waagerecht ? fussQuer : fussSenkrecht}>
        <button
          type="button"
          onClick={paletteOeffnen}
          className="flex h-8 w-full items-center justify-between gap-2 rounded-md px-3 text-left text-sm text-leise hover:bg-betont focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        >
          {/* **Oben nur das Kürzel sichtbar** (UX-14.AK-02): der Name bleibt
              für den Vorleser, und er enthält, was man sieht (WCAG 2.5.3). */}
          {/* Gekürzt bleibt der Name als Titel lesbar — H03 (#2257). */}
          <span
            className={waagerecht ? 'sr-only' : 'truncate'}
            title={waagerecht ? undefined : texte.allgemein.palette.oeffnen}
          >
            {texte.allgemein.palette.oeffnen}
          </span>
          <span aria-hidden={!waagerecht} className="shrink-0 text-xs">
            ⌘K
          </span>
        </button>
      </div>
    </div>
  )
}
