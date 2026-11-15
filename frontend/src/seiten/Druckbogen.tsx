// Der Druckbogen eines Sets.
//
// **Was sie tut, steht hier; wie sie eingestellt wird, in `Bogeneinstellungen`**
// (AGENTS.md §2).
//
// Anders als die Bildausgabe läuft hier **kein** Stapel in der Oberfläche: ein
// PDF entsteht als Ganzes, und ein halbes wäre keines. Der Lauf ist ein Aufruf,
// und was er kostet, sagt der Knopf, während er läuft.
import { useState } from 'react'

import { api, auspacken, type BogenAntwort } from '../api'
import { useVorgang } from '../haken/useVorgang'
import { istAbgebrochen, meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { zaehlwort } from '../sprache/zaehlwort'
import { Ausgabeauswahl } from '../bausteine/Ausgabeauswahl'
import { Ausgabefuss } from '../bausteine/Ausgabefuss'
import { Knopf } from '../bausteine/Knopf'
import { Wieeswird } from '../bausteine/Wieeswird'
import { useBogenprobelauf } from '../haken/useBogenprobelauf'
import { Bogeneinstellungen } from './Bogeneinstellungen'
import { Bogenzahlen, Bogenzeile } from './Bogenzahlen'
import { Dateiverweis } from '../bausteine/Dateiverweis'
import { Laufstand } from '../bausteine/Laufstand'

interface DruckbogenEigenschaften {
  // setID ist **gesetzt** — der Arbeitsplatz rendert diese Lasche nur mit
  // offenem Set und zeigt sonst seinen eigenen leeren Zustand
  // (#288).
  // Ein Wächter „kein Set gewählt" wäre hier ein Zweig, den kein Aufruf
  // betreten kann (§3).
  setID: string
}

export function Druckbogen({ setID }: DruckbogenEigenschaften) {
  const texte = useTexte()
  const vorgang = useVorgang()

  const [papier, setzePapier] = useState('a4')
  const [anschnitt, setzeAnschnitt] = useState(0)
  const [aufloesung, setzeAufloesung] = useState(300)
  const [marken, setzeMarken] = useState(true)

  // **Welche Karten hinausgehen** (#1167) — leer heißt: das ganze Set.
  const [auswahl, setzeAuswahl] = useState<readonly string[]>([])

  const [bogen, setzeBogen] = useState<BogenAntwort>()
  const [fehler, setzeFehler] = useState<string>()

  // **Wie es wird, bevor etwas geschrieben wird** (UXR-F7, #2015): derselbe
  // Dienst mit denselben Wahlen wie der Knopf darunter. Eine Wahl, die sich
  // ändert, verwirft die laufende Frage (§7, Regel 2).
  const probe = useBogenprobelauf(
    setID,
    papier,
    anschnitt,
    aufloesung,
    marken,
    auswahl,
  )

  // **Ein Vorgang je Set** (§7, Regel 4): zwei Bögen desselben Sets gleichzeitig
  // gingen auf dieselbe Datei.
  const legen = async (wessen: string) => {
    await vorgang.starte(`bogen-${wessen}`, async () => {
      setzeFehler(undefined)
      setzeBogen(undefined)

      try {
        const antwort = await api().BogenLegen(
          wessen,
          papier,
          anschnitt,
          aufloesung,
          marken,
          [...auswahl],
        )

        auspacken(antwort, () => undefined)

        if (antwort.abgebrochen) {
          return
        }

        setzeBogen(antwort)
      } catch (ursache: unknown) {
        // **Wer anhält, bekommt keine Fehlermeldung dafür** (FMT-03) — der
        // Bogen entsteht erst am Ende, und der Laufstand sagt, dass keiner
        // geschrieben wurde.
        if (!istAbgebrochen(ursache)) {
          setzeFehler(meldungAus(ursache, texte))
        }
      }
    })
  }

  const laeuft = vorgang.laeuft(`bogen-${setID}`)

  return (
    <>
      {/* **Überschrift und Erklärung trägt die Fläche des Assistenten**
          (UXR-09, #1917) — `Ausgabeflaeche`, unter dem Namen, den die Wahl
          „Wohin?" daneben trägt. */}
      <Bogeneinstellungen
        papier={papier}
        anschnitt={anschnitt}
        aufloesung={aufloesung}
        marken={marken}
        laeuft={laeuft}
        setzePapier={setzePapier}
        setzeAnschnitt={setzeAnschnitt}
        setzeAufloesung={setzeAufloesung}
        setzeMarken={setzeMarken}
      />

      {/* **Welche Karten hinausgehen** (#1167). Sie steht **unter** den
          Maßen, weil sie die zweite Frage ist: erst wie der Bogen aussieht,
          dann was darauf kommt — und MSE2 fragt sie ebenso erst im
          Ausgabefenster. */}
      <div className="mt-4">
        <Ausgabeauswahl
          setID={setID}
          gewaehlt={auswahl}
          setzeGewaehlt={setzeAuswahl}
          angehalten={laeuft}
        />
      </div>

      {/* **„Wie es wird"** (UXR-F7, #2015) steht vor dem Knopf: erst sehen, dann
          schreiben. */}
      <Wieeswird
        wartet={probe.daten === undefined && probe.fehler === undefined}
        veraltet={probe.laedt}
        fehler={probe.fehler}
      >
        {probe.daten !== undefined && <Bogenzahlen probe={probe.daten} />}
      </Wieeswird>

      {/* **Der Knopf, der legt, führt** (UXR-09, R1 aus `docs/UIREVIEW.md`) —
          in der Fußleiste, die beim Rollen stehen bleibt (AP-10, B-20).
          Daneben, was er legt, **solange die Zahl zur Wahl gehört**; beim
          Rechnen schweigt sie, statt eine alte zu zeigen. Vorgelesen wird
          sie in „Wie es wird" — hier stünde sie ein zweites Mal. */}
      <Ausgabefuss>
        <Knopf
          gewicht="haupt"
          angehalten={laeuft}
          aufKlick={() => {
            void legen(setID)
          }}
        >
          {texte.seiten.druckbogen.legen}
        </Knopf>

        {probe.daten !== undefined && !probe.laedt && (
          <span aria-hidden="true" className="text-sm text-leise">
            <Bogenzeile probe={probe.daten} />
          </span>
        )}
      </Ausgabefuss>

      <Laufstand art="bogen" />

      {fehler !== undefined && (
        <p role="alert" className="mt-3 text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      )}

      {bogen && (
        <div role="status" className="mt-4">
          <p className="text-leise">
            {`${zaehlwort(bogen.karten, texte.allgemein.zaehlwoerter.karten)} ${texte.seiten.druckbogen.fertig} ${zaehlwort(bogen.seiten, texte.allgemein.zaehlwoerter.seiten)}: ${bogen.pfad}`}
          </p>

          <Dateiverweis pfad={bogen.pfad} />

          {bogen.befunde.length > 0 && (
            <>
              <p className="mt-3 max-w-prose text-leise">
                {texte.seiten.druckbogen.befunde}
              </p>
              <ul className="mt-1 space-y-1 font-mono text-sm">
                {bogen.befunde.map((befund) => (
                  <li key={befund}>{befund}</li>
                ))}
              </ul>
            </>
          )}
        </div>
      )}
    </>
  )
}
