// Der Rahmen: Navigation, Befehlspalette, Tastenkürzel.
//
// **Die Seitenliste kommt aus `navigation.ts`**, nicht aus einer Kopie daneben
// (AGENTS.md §3, sinngemäß auch #38). Eine zu kurze Aufzählung ist grün, und
// niemand sieht es.
import {
  type ComponentProps,
  type ReactNode,
  useCallback,
  useEffect,
  useState,
} from 'react'

import { type Seitenschluessel } from '../navigation'
import { useTexte } from '../sprache/kontext'
import { Brotkrume } from '../bausteine/Brotkrume'
import { brotkrumen, type Bezugsobjekt, type Krume } from '../brotkrumen'
import { Seitenweggeber } from '../seitenweg'
import { Fassungsband } from './Fassungsband'
import { Wiederanlaufband } from './Wiederanlaufband'
import { Meldungsband, type Bandmeldung } from './Meldungsband'
import { Palette } from './Palette'
import { Rahmendialoge } from './Rahmendialoge'
import type { BefehlDTO } from '../apibefehle'
import { inEingabe } from '../eingabefeld'
import type { Ziffernbelegung } from './ziffernbelegung'
import { deuten, istGewohnt } from './tastatur'
import { Seitenleistenknopf } from './Seitenleistenknopf'
import { Sprunglink } from './Sprunglink'
import { Inhalt } from './Inhalt'
import { Seitenleiste } from './Seitenleiste'
import { useLeistenfokus } from '../haken/useLeistenfokus'
import { useRuecknahmeziel } from '../ruecknahme'
import { useLeistenhalt } from '../haken/useLeistenhalt'
import { useSchmalesFenster } from '../haken/useSchmalesFenster'

interface RahmenEigenschaften {
  aktuell: Seitenschluessel
  wechseln: (seite: Seitenschluessel) => void
  suchen?: () => void

  // zuruecknehmen und wiederherstellen liegen hinter ⌘Z und ⇧⌘Z (#52).
  //
  // **Als Eigenschaften und nicht als eigener Haken im Rahmen**: welches Set
  // gemeint ist, weiß `App` — der Rahmen kennt keine Bezugsobjekte und soll
  // keine kennen.
  zuruecknehmen?: () => void
  wiederherstellen?: () => void

  // sichern und sichernUnter liegen hinter ⌘S und ⇧⌘S (#162) — und hinter den
  // gleichnamigen Einträgen der nativen Menüleiste.
  sichern?: () => void
  sichernUnter?: () => void

  // **⌘O, ⌘W und ⌘P** (#891) — dieselben Handlungen wie in der Ablage.
  //
  // **Wahlfrei wie `sichern`**: die Browservorschau hat kein Set und keinen
  // Dateidialog, und eine Vorgabe hier wäre eine zweite Stelle, an der steht,
  // was diese Tasten bedeuten.
  oeffnen?: () => void
  neu?: () => void
  schliessen?: () => void
  drucken?: () => void

  // **⌘[ und ⌘] gehen den Seitengang** (UX-03.AK-02, #1675).
  seiteZurueck?: () => void
  seiteVor?: () => void

  // meldung ist das, was zu keiner Seite gehört (#180).
  meldung?: Bandmeldung | undefined
  meldungSchliessen?: () => void

  // seitenleiste sagt, ob die Leiste steht, und wie man sie umschaltet —
  // Werkstück U14 von Phase 5.1 (#1490). Ohne Angabe steht sie.
  // `waagerecht` legt sie oben als Zeile (UX-14.AK-02, #1686).
  seitenleiste?:
    { offen: boolean; umschalten: () => void; waagerecht?: boolean } | undefined

  // objekt ist, woran gerade gearbeitet wird — UX-03
  // (#1675).
  //
  // **Gereicht und nicht geholt**: der Rahmen kennt keine Bezugsobjekte
  // (#52), und das Fenster ist die einzige Stelle, die das offene Set führt.
  objekt?: Bezugsobjekt | undefined

  // ziffern ist die Belegung von ⌘0 bis ⌘9 (UX-02.AK-03, #1674) — dieselbe
  // für die Tastatur und für die Leiste, damit beide dasselbe sagen.
  ziffern?: Ziffernbelegung | undefined

  // palettenwege lassen die Befehlspalette Handlungen und Objekte anbieten
  // (UX-02.AK-02, #1674) — gereicht wie `objekt`, aus demselben Grund.
  palettenwege?: ComponentProps<typeof Palette>['wege']

  children: ReactNode
}

// **Die Leiste links oder oben** — UX-14.AK-02 (#1686): oben stehen Rahmen
// und Inhalt untereinander statt nebeneinander.
const rahmenSenkrecht = 'relative flex min-h-screen bg-grund text-text'
const rahmenQuer = 'relative flex min-h-screen flex-col bg-grund text-text'

export function Rahmen({
  aktuell,
  wechseln,
  suchen,
  zuruecknehmen,
  wiederherstellen,
  sichern,
  sichernUnter,
  oeffnen,
  neu,
  schliessen,
  drucken,
  seiteZurueck,
  seiteVor,
  meldung,
  meldungSchliessen,
  seitenleiste,
  objekt,
  palettenwege,
  ziffern,
  children,
}: RahmenEigenschaften) {
  const texte = useTexte()

  // **Wohin ⌘Z zielt, fragt die Taste bei jedem Druck** — wie der Menüeintrag.
  const ruecknahme = useRuecknahmeziel()

  // **Ein schmales Fenster verdeckt die Einstellung, es ändert sie nicht**
  // (#1630).
  // Ob die Leiste stehen soll, ist eine Entscheidung des Menschen und liegt
  // im Bestand (U14, #1490); die Breite des Fensters ist keine.
  const schmal = useSchmalesFenster()

  // **Die Schublade merkt sich die Seite, auf der sie geöffnet wurde** — und
  // steht offen, solange dieselbe aufgeschlagen ist. Schlägt jemand eine
  // andere auf, ist sie zu, ohne dass es jemand abräumen müsste (§7, Regel 1).
  //
  // **Abgeleitet und nicht nachgeführt**: ein Effekt, der beim Wechsel
  // `false` setzt, wäre ein zweites Zeichnen für eine Antwort, die schon
  // feststeht — und `react-hooks/set-state-in-effect` weist ihn zu Recht ab.
  const [schubladeAuf, setzeSchubladeAuf] = useState<Seitenschluessel>()

  const gespeichertOffen = seitenleiste?.offen ?? true
  const gespeichertUmschalten = seitenleiste?.umschalten
  const waagerecht = seitenleiste?.waagerecht === true
  const leisteOffen = schmal ? schubladeAuf === aktuell : gespeichertOffen

  const leisteUmschalten = useCallback(() => {
    if (schmal) {
      setzeSchubladeAuf((offenAuf) =>
        offenAuf === aktuell ? undefined : aktuell,
      )

      return
    }

    gespeichertUmschalten?.()
  }, [aktuell, gespeichertUmschalten, schmal])

  const { leistenknopf, leistenbereich } = useLeistenfokus(leisteOffen)

  // **Eine Tabstelle für die ganze Leiste** (#1488): der Halt liegt auf der
  // aufgeschlagenen Seite und folgt ihr; ↑/↓ wandern (`leistentaste`).
  const halt = useLeistenhalt(aktuell)
  const [paletteOffen, setzePaletteOffen] = useState(false)

  // **Was die Seite über ihren Weg sagt** (UX-03.AK-02, `seitenweg.tsx`) —
  // die Partie über dem Tisch kennt nur die Spielen-Seite.
  const [seitenweg, setzeSeitenweg] = useState<readonly Krume[]>()

  // **Die Kommandozentrale** (#1423) — offen über genau einem Befehl, den
  // die Palette gewählt hat; sie geht dabei zu (kein Dialog über einem Dialog).
  const [befehl, setzeBefehl] = useState<BefehlDTO | undefined>(undefined)

  // Als eigener Rückruf, damit `Palette` ihn nicht bei jedem Neuzeichnen als
  // anderen sieht und ihren Escape-Horcher ab- und wieder anmeldet.
  const schliessePalette = useCallback(() => {
    setzePaletteOffen(false)
  }, [])

  const aufTaste = useCallback(
    (ereignis: KeyboardEvent) => {
      const treffer = deuten(
        {
          taste: ereignis.key,
          code: ereignis.code,
          meta: ereignis.metaKey,
          ctrl: ereignis.ctrlKey,
          alt: ereignis.altKey,
          umschalt: ereignis.shiftKey,
          imEingabefeld: inEingabe(ereignis.target),
        },
        ziffern,
      )
      if (!treffer) {
        return
      }

      // **Erst dann abfangen, wenn das Kürzel wirklich gilt.** Wer jede
      // ⌘-Kombination abfängt, nimmt dem Betriebssystem ⌘Q und ⌘W weg — und die
      // sagt §6 ausdrücklich zu.
      ereignis.preventDefault()

      if (treffer.art === 'palette') {
        setzePaletteOffen((offen) => !offen)

        return
      }

      if (treffer.art === 'zurueck' || treffer.art === 'vor') {
        // **Nichts tun, wenn niemand etwas anzubieten hat.** Abgefangen ist die
        // Taste trotzdem: ⌘Z gehört ab hier der Anwendung, und ein Durchreichen
        // an den Browser täte an dieser Stelle nichts Sinnvolles.
        //
        // **Ein gemeldeter Weg geht vor** — derselbe wie beim Menüeintrag
        // (`ruecknahmeweg.ts`): der Tisch (P8, #2376) und die offene
        // Zeichnung. Bis dahin traf die Taste immer den Set-Verlauf.
        const weg = ruecknahme.gemeldetes()
        const handlung =
          treffer.art === 'zurueck'
            ? (weg?.zuruecknehmen ?? zuruecknehmen)
            : (weg?.wiederherstellen ?? wiederherstellen)
        handlung?.()

        return
      }

      if (treffer.art === 'sichern' || treffer.art === 'sichernUnter') {
        // **Nichts tun, wenn kein Set offen ist** — und die Taste trotzdem
        // abfangen. ⌘S gehört ab hier der Anwendung; ein Durchreichen an den
        // Browser böte an dieser Stelle „Seite sichern" an, und das ist die
        // falsche Antwort auf eine gewohnte Taste (#162).
        const handlung = treffer.art === 'sichern' ? sichern : sichernUnter
        handlung?.()

        return
      }

      // **Öffnen, Schließen, Drucken** (#891) und der Seitengang (UX-03) —
      // und die Taste wird auch dann abgefangen, wenn niemand etwas anzubieten
      // hat. ⌘P böte sonst „Seite drucken" des Browsers an, und das ist die
      // falsche Antwort auf eine gewohnte Taste.
      if (istGewohnt(treffer)) {
        const handlungen = {
          oeffnen,
          neu,
          schliessen,
          drucken,
          seiteZurueck,
          seiteVor,
        }

        handlungen[treffer.art]?.()

        return
      }

      if (treffer.art === 'suche') {
        // **⌘F führt zur Suche**, wenn niemand etwas anderes vorgibt. Eine
        // Taste, die je nach Seite etwas anderes tut, ist keine Gewohnheit
        // mehr.
        if (suchen) {
          suchen()
        } else {
          wechseln('suche')
        }

        return
      }

      // **Die Regionen gehören dem Arbeitsplatz** (#1413): er horcht selbst
      // am Fenster, denn nur er weiß, ob er einen Navigator hat. Hier ist die
      // Taste abgefangen und sonst nichts.
      if (treffer.art === 'navigator' || treffer.art === 'inspector') {
        return
      }

      if (treffer.art === 'seitenleiste') {
        ereignis.preventDefault()
        leisteUmschalten()

        return
      }

      wechseln(treffer.seite)
    },
    [
      suchen,
      wechseln,
      ziffern,
      ruecknahme,
      zuruecknehmen,
      wiederherstellen,
      sichern,
      sichernUnter,
      oeffnen,
      neu,
      schliessen,
      drucken,
      seiteZurueck,
      seiteVor,
      leisteUmschalten,
    ],
  )

  useEffect(() => {
    window.addEventListener('keydown', aufTaste)

    return () => {
      window.removeEventListener('keydown', aufTaste)
    }
  }, [aufTaste])

  return (
    // **`bg-grund` und `text-text` kommen aus `stil.css`** (Vermerk 0065)
    // **und brauchen kein `dark:`**: die Merkmale wechseln selbst mit dem
    // Erscheinungsbild.
    <div className={waagerecht ? rahmenQuer : rahmenSenkrecht}>
      <Sprunglink />

      {/* **Die Seitenleiste ist eine eigene Fläche.** Sie steht auf dem
          Anwendungsgrund wie ein Blatt darauf; ohne diesen Unterschied hält
          eine Trennlinie allein zwei Bereiche auseinander — und eine Linie ist
          keine Ordnung
          (#288). */}
      {/* **Oben gibt es nichts einzuklappen** (UX-14.AK-02): eine Zeile
          nimmt dem Inhalt keine Breite. */}
      {!waagerecht && (
        <Seitenleistenknopf
          offen={leisteOffen}
          umschalten={leisteUmschalten}
          griff={leistenknopf}
        />
      )}
      <Seitenleiste
        ziffern={ziffern}
        aktuell={aktuell}
        wechseln={wechseln}
        offen={leisteOffen}
        schmal={schmal}
        bereich={leistenbereich}
        halt={halt}
        waagerecht={waagerecht}
        paletteOeffnen={() => {
          setzePaletteOffen(true)
        }}
      />

      <div className="flex min-w-0 flex-1 flex-col">
        {/* **Über dem Inhalt und unter der Navigation**: eine Meldung, die zu
        keiner Seite gehört, gehört auch an keine Stelle in ihr (#180). */}
        <Meldungsband
          meldung={meldung}
          schliessen={meldungSchliessen ?? (() => undefined)}
        />

        {/* **Und darunter die neue Fassung, wenn eine da ist** (#1100). Sie
            gehört zu keiner Seite und trägt eine eigene Handlung — deshalb
            ein eigenes Band und keine zweite Meldung im ersten. */}
        <Fassungsband />
        <Wiederanlaufband />

        {/* **Wo du bist** — UX-03.AK-02: die fachliche Hierarchie, nicht die
            Komponentenstruktur. Über dem Kopf und unter dem Band, so wie der
            Entwurf sie zeichnet. */}
        <Brotkrume
          krumen={brotkrumen(texte, aktuell, objekt, seitenweg)}
          wechseln={wechseln}
        />

        <Inhalt aktuell={aktuell}>
          <Seitenweggeber setze={setzeSeitenweg}>{children}</Seitenweggeber>
        </Inhalt>
      </div>

      <Rahmendialoge
        paletteOffen={paletteOffen}
        befehl={befehl}
        schliessePalette={schliessePalette}
        waehlen={(seite) => {
          setzePaletteOffen(false)
          wechseln(seite)
        }}
        waehleBefehl={(gewaehlt) => {
          setzePaletteOffen(false)
          setzeBefehl(gewaehlt)
        }}
        schliesseBefehl={() => {
          setzeBefehl(undefined)
        }}
        wege={palettenwege}
      />
    </div>
  )
}
