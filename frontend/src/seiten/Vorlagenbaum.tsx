// **Der Vorlagenbaum** — Pakete und ihre Dateien, der Navigator der Vorlagenwerkstatt (Werkstück U7 von
// Phase 5.1, #1418).
//
// **`Vorlagenbaum` und nicht `Paketbaum`**, weil `seiten/paketbaum.ts` schon
// der Gruppenbaum der Paketseite ist — und auf einer Platte, die Groß- und
// Kleinschreibung nicht unterscheidet, löste `./Paketbaum` auf die falsche
// Datei auf. Gefunden beim ersten Testlauf.
//
// # Warum eine Liste mit Aufklappknöpfen und kein `role="tree"`
//
// Dieselbe Überlegung wie bei `Paketzeilen.tsx`: ein ARIA-Baum verlangt
// eine Tippsuche, Kürzel zum Auf- und Zuklappen und eine eigene
// Fokusverwaltung — viel Fläche, auf der sich etwas verstecken kann. Hier
// stehen Pakete als Überschriften mit einem Aufklappknopf und darunter ihre
// Dateien als Knöpfe; `↑` und `↓` laufen über alle Knöpfe, `Home` und `End`
// an die Enden.
//
// **Eine Tabstelle für den ganzen Baum** — derselbe Vertrag wie bei der
// Paketliste (#1290): gemessen mit `arbeitswegmass.mjs` war die Datei
// »style« des Messpakets die **30.** Tabstelle, weil jeder Paketknopf ein
// eigener Halt war. Jetzt hält nur ein Knopf `tabIndex 0` — die gewählte
// Datei, sonst die erste Datei des ersten bearbeitbaren Pakets —, und wer
// von dort weiterwill, nimmt die Pfeile. **Bearbeitbare Pakete stehen
// oben**, mitgelieferte darunter: was sich hier nicht ändern lässt, soll
// nicht vor dem stehen, wofür die Seite da ist.
//
// # Wann alles aufgeklappt beginnt
//
// **Solange der Baum klein ist**, denn sonst kostete der Kernweg `paket`
// eine Handlung mehr (UX-ARCHITEKTUR, Ziel ≤ 5). Hier stand „Wer viele
// Pakete hat, filtert oder klappt zu" — und mit den 526 Paketen des
// Full-Magic-Pack zeichnete die Werkstatt 35.769 Knöpfe, ehe jemand etwas
// tun konnte (UXR-18, #1926). **Ein großer Baum beginnt zugeklappt**, sein
// Beiwerk steht hinter einer eigenen Zeile, und keine Liste zeigt mehr als
// ein Stück auf einmal; die Regel steht in `vorlagenbaumfaltung.ts`.
//
// **Der Filter sucht Paket- und Dateinamen.** Passt der Paketname, bleiben
// alle seine Dateien; passt nur eine Datei, bleibt sie mit ihrem Paket.
import { Wirdgeladen } from '../bausteine/Ladehinweise'
import { useCallback, useId, useRef, useState } from 'react'

import { useVorlagenbaum } from '../haken/useVorlagenbaum'
import { Knopf } from '../bausteine/Knopf'
import { useTexte } from '../sprache/kontext'
import { Weiterknopf, sichtbar } from './Vorlagenbaumteile'
import { Vorlagenzweig, dateischluessel } from './Vorlagenzweig'
import { Stueck, plane } from './vorlagenbaumfaltung'

import type { KeyboardEvent } from 'react'
import type { Paketkopie } from '../haken/usePaketkopie'
import type { Zweig } from '../haken/useVorlagenbaum'
import type { Klappwahl, Stuecke } from './vorlagenbaumfaltung'
import type { Dateianstoss, Zeilengriffe } from './Vorlagenzweig'

export interface Dateiwahl {
  paket: string
  datei: string

  // beiwerk sagt, dass die Datei kein Text ist — die Werkstatt zeigt dann
  // keinen Editor, sondern eine Tafel (FMT-02, #1691).
  beiwerk?: boolean
}

interface VorlagenbaumEigenschaften {
  pakete: readonly string[]

  // geladen und fehlschlag trennen „noch nicht da" von „keine" (UX-08.NAT-01):
  // vorher steht „Wird geladen" da, nach einem Fehlschlag nichts — der steht
  // mit seinem nächsten Versuch über der Seite.
  geladen: boolean
  fehlschlag: boolean
  gewaehlt: Dateiwahl | null
  waehlen: (paket: string, datei: string, beiwerk?: boolean) => void

  // fassung stößt das Nachladen an — nach einem Handgriff am Dateibaum.
  fassung?: number

  // geaendert sagt, dass die gewählte Datei ungespeichert ist — der Baum
  // nennt es an ihrem Knopf, damit es auch eine Sprachausgabe hört.
  geaendert: boolean

  // kopie legt die eigene Kopie eines Pakets an, das sich nur lesen lässt —
  // FMT-02.AK-04 (#1691), `usePaketkopie`. Ohne sie sagt der Baum nur, dass
  // es sich nicht ändern lässt.
  kopie?: Pick<Paketkopie, 'kopieren' | 'laeuft' | 'fehler'>

  // zuDenPaketen führt dorthin, wo ein Paket entsteht — G02 (#2249). Ohne
  // sie nennt der Baum den Weg nur.
  zuDenPaketen?: () => void

  // dateianstoss öffnet die Handgriffe einer Datei — AP-5 (#2362). Ohne die
  // Werkstatt darum hat der Baum keine, und der Rechtsklick bleibt bei macOS.
  dateianstoss?: Dateianstoss
}

const ohneAnstoss: Dateianstoss = () => undefined

export function Vorlagenbaum({
  pakete,
  geladen,
  fehlschlag,
  gewaehlt,
  waehlen,
  geaendert,
  kopie,
  fassung = 0,
  zuDenPaketen,
  dateianstoss = ohneAnstoss,
}: VorlagenbaumEigenschaften) {
  const alle = useTexte()
  const texte = alle.seiten.vorlagen
  const sucheLeeren = alle.bausteine.listenansicht.sucheLeeren
  const kennung = useId()
  const [anfrage, setzeAnfrage] = useState('')
  const [klappwahl, setzeKlappwahl] = useState<Klappwahl>({})
  const [stuecke, setzeStuecke] = useState<Stuecke>({})
  const baum = useVorlagenbaum(pakete, fassung)

  const gefiltert = baum.zweige
    .map((zweig) => sichtbar(zweig, anfrage))
    .filter((zweig): zweig is Zweig => zweig !== null)
  const zweige = [
    ...gefiltert.filter((zweig) => !zweig.nurLesbar),
    ...gefiltert.filter((zweig) => zweig.nurLesbar),
  ]
  const plan = plane(zweige, { gewaehlt, klappwahl, stuecke })

  // **Die eine Tabstelle des Baums.** `fokus` ist der zuletzt betretene
  // Knopf; solange es ihn gibt, bleibt er der Halt. Sonst die gewählte
  // Datei, sonst die erste Datei des ersten bearbeitbaren Pakets, sonst das
  // erste Paket.
  const [fokus, setzeFokus] = useState<string | null>(null)
  const vorhanden = new Set<string>()

  for (const zweigplan of plan.zweige) {
    const { paket } = zweigplan.zweig

    vorhanden.add(`p:${paket}`)
    for (const datei of [...zweigplan.dateien, ...zweigplan.beiwerk]) {
      vorhanden.add(dateischluessel(paket, datei))
    }
    vorhanden.add(`w:d:${paket}`)
    vorhanden.add(`b:${paket}`)
    vorhanden.add(`w:b:${paket}`)
  }
  vorhanden.add('w:pakete')

  const ersteDatei = plan.zweige.flatMap((zweigplan) =>
    zweigplan.dateien
      .slice(0, 1)
      .map((datei) => dateischluessel(zweigplan.zweig.paket, datei)),
  )[0]
  const halt =
    [
      fokus,
      gewaehlt === null
        ? null
        : dateischluessel(gewaehlt.paket, gewaehlt.datei),
      ersteDatei ?? null,
      zweige[0] === undefined ? null : `p:${zweige[0].paket}`,
    ].find((kandidat) => kandidat !== null && vorhanden.has(kandidat)) ?? null

  // klappen dreht eine Zeile um — und merkt es sich, damit es gilt, auch
  // wenn der Baum klein oder groß wird.
  const klappen = (schluessel: string, offen: boolean) => {
    setzeKlappwahl((war) => ({ ...war, [schluessel]: !offen }))
  }

  // **Nach „Weitere zeigen" steht der Fokus auf dem ersten neuen Eintrag** —
  // dort, wo jemand weiterlesen will. `springen` hält den Wunsch bis zum
  // Zeichnen, und die Referenz des neuen Halts erfüllt ihn.
  const springen = useRef(false)
  const fokussiere = useCallback((knopf: HTMLButtonElement | null) => {
    if (knopf !== null && springen.current) {
      springen.current = false
      knopf.focus()
    }
  }, [])

  const weiter = (liste: string, gezeigt: number, erster: string) => {
    setzeStuecke((war) => ({ ...war, [liste]: gezeigt + Stueck }))
    springen.current = true
    setzeFokus(erster)
  }

  // **Die Pfeiltasten laufen über die Knöpfe des Baums** — in der Reihenfolge,
  // in der sie stehen; ein zugeklapptes Paket hat keine Dateiknöpfe. Der
  // Horcher hängt an jedem Knopf, nicht an der Liste: eine Liste ist kein
  // Bedienelement, und jsx-a11y weist einen Horcher dort zu Recht ab.
  const aufTaste = (ereignis: KeyboardEvent<HTMLButtonElement>) => {
    const richtungen: Record<string, number> = { ArrowDown: 1, ArrowUp: -1 }
    // **Nur die Knoten** — WK02 (#2103): eine Kopieaktion steht auch im
    // Baum, ist aber eine eigene Tabstelle ohne Horcher. Lief der Pfeil in
    // sie hinein, führte keiner mehr heraus, und Ende landete dort.
    const knoepfe = [
      ...document.querySelectorAll<HTMLButtonElement>(
        `[data-baum="${kennung}"] [data-baumknoten]`,
      ),
    ]
    const schritt = richtungen[ereignis.key]
    let ziel = -1

    if (schritt !== undefined) {
      ziel = Math.min(
        knoepfe.length - 1,
        Math.max(0, knoepfe.indexOf(ereignis.currentTarget) + schritt),
      )
    } else if (ereignis.key === 'Home') {
      ziel = 0
    } else if (ereignis.key === 'End') {
      ziel = knoepfe.length - 1
    }

    if (ziel >= 0) {
      ereignis.preventDefault()
      knoepfe[ziel]?.focus()
    }
  }

  const griffe = (schluessel: string): Zeilengriffe => ({
    'data-baumknoten': true,
    tabIndex: halt === schluessel ? 0 : -1,
    ref: halt === schluessel ? fokussiere : undefined,
    onFocus: () => {
      setzeFokus(schluessel)
    },
    onKeyDown: aufTaste,
  })

  const { naechstesPaket } = plan

  return (
    <nav data-baum={kennung} className="flex flex-col gap-2">
      <div className="flex flex-col gap-1">
        <label htmlFor={kennung} className="sr-only">
          {texte.baumSuche}
        </label>
        <input
          id={kennung}
          type="search"
          value={anfrage}
          placeholder={texte.baumSuchePlatzhalter}
          onChange={(ereignis) => {
            setzeAnfrage(ereignis.target.value)
          }}
          onKeyDown={(ereignis) => {
            if (ereignis.key === 'Escape' && anfrage !== '') {
              ereignis.preventDefault()
              setzeAnfrage('')
            }
          }}
          className="min-w-0 rounded-md border border-kante bg-flaeche px-2 py-1 text-sm placeholder:italic placeholder:text-leise focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        />
      </div>

      {!geladen && !fehlschlag && <Wirdgeladen />}

      {/* **Mit dem Weg dorthin** (G02, #2249) — knapp, denn der Baum steht
          in der schmalen Spalte. */}
      {geladen && pakete.length === 0 && (
        <div className="flex flex-col items-start gap-2 text-sm">
          <p className="text-leise">{texte.baumOhnePakete}</p>

          {zuDenPaketen !== undefined && (
            <Knopf aufKlick={zuDenPaketen}>
              {alle.allgemein.arbeitsplatz.zuDenPaketen}
            </Knopf>
          )}
        </div>
      )}

      {pakete.length > 0 && baum.laedt && zweige.length === 0 && (
        <p className="text-sm text-leise">{texte.dateienLaden}</p>
      )}

      {pakete.length > 0 && !baum.laedt && zweige.length === 0 && (
        <div className="flex flex-col items-start gap-2 text-sm">
          <p className="text-leise">{texte.baumLeer}</p>

          {anfrage !== '' && (
            <Knopf
              aufKlick={() => {
                setzeAnfrage('')
              }}
            >
              {sucheLeeren}
            </Knopf>
          )}
        </div>
      )}

      {/* **Einmal gesagt, nicht je Paket**: liegt hier nichts Bearbeitbares,
          steht der Weg dorthin über dem Baum — und nicht siebenmal darunter. */}
      {!baum.laedt &&
        zweige.length > 0 &&
        zweige.every((zweig) => zweig.nurLesbar) && (
          <p className="text-sm text-leise">{texte.baumNurMitgeliefert}</p>
        )}

      <ul className="flex flex-col gap-1">
        {plan.zweige.map((zweigplan) => (
          <Vorlagenzweig
            key={zweigplan.zweig.paket}
            plan={zweigplan}
            gewaehlt={gewaehlt}
            geaendert={geaendert}
            kopie={kopie}
            griffe={griffe}
            waehlen={waehlen}
            klappen={klappen}
            weiter={weiter}
            dateianstoss={dateianstoss}
          />
        ))}
        {naechstesPaket !== undefined && (
          <Weiterknopf
            gezeigt={plan.zweige.length}
            gesamt={zweige.length}
            griffe={griffe('w:pakete')}
            aufKlick={() => {
              weiter('pakete', plan.zweige.length, `p:${naechstesPaket}`)
            }}
          />
        )}
      </ul>
    </nav>
  )
}
