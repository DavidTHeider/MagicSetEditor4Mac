// Der **Editor des gewählten Schlüsselworts**
// (#1062,
// #2009).
//
// # Was MSE2 hier tut
//
// Rechts neben der Liste stehen die Felder **des einen gewählten Wortes**:
// Schlüsselwort, Verwendung, Muster mit „Parameter einfügen", Erinnerungstext
// mit „Auf Parameter verweisen" und die Regeln
// (`KeywordsPanel::initControls`, `src/gui/set/keywords_panel.cpp`, Zeilen
// 34–91, Fassung f9d9356 vom 2021-07-06). Hier ist es dieselbe Anordnung in
// einer Spalte; die Liste dazu steht in `Schluesselwortliste.tsx`.
//
// # Warum als eigene Datei
//
// `Schluesselwoerter.tsx` beantwortet die Frage „was tut diese Seite": laden,
// wählen, anlegen, schreiben, löschen. Wie **ein** Schlüsselwort aussieht, ist
// eine andere Frage (§2) — dieselbe Aufteilung wie zwischen `Karteneditor` und
// `Feldzeile`. Die Datei hieß `Schluesselwortzeile.tsx`, solange jedes Wort
// seine Zeile mit allen Feldern hatte (UXR-F3).
import { useRef, type RefObject } from 'react'

import type { SchluesselwortDTO } from '../api'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Parameterwahl } from './Parameterwahl'
import {
  eingabe,
  einfuegenAnDerMarke,
  Schluesselwortfeld,
} from './Schluesselwortfeld'
import { useEntwurf } from '../haken/useEntwurf'
import { useParameterhilfen } from '../haken/useParameterhilfen'
import { useTexte } from '../sprache/kontext'

interface SchluesselworteditorEigenschaften {
  wort: SchluesselwortDTO
  modi: string[]
  angehalten: boolean

  // namensfeld ist der Griff auf das Namensfeld — **den hält die Seite**, weil
  // sie nach dem Anlegen den Fokus dorthin setzt, sobald der Editor des neuen
  // Wortes steht.
  namensfeld?: RefObject<HTMLInputElement | null>

  // setID und stand tragen die Parameterhilfen (#1342): sie hängen am Set und
  // müssen nach jeder Änderung am Muster neu gerechnet werden.
  setID: string
  stand: number

  // bearbeiten schreibt und löscht das Schlüsselwort dieses Editors.
  //
  // **Immer da, seit der Editor nur noch eigene trägt** — UXR-07 (#1915). Die
  // des Spiels stehen als Text (`Schluesselwortauskunft`); bis dahin fehlte
  // `bearbeiten` bei ihnen, und die Felder waren abgeschaltet.
  bearbeiten: {
    // `bekannt` kommt vom Entwurf (U02b, #2179); der Modus hat keinen und
    // schreibt ohne Bedingung.
    schreiben: (feld: string, wert: string, bekannt?: readonly string[]) => void
    loeschen: () => void
  }
}

export function Schluesselworteditor({
  wort,
  modi,
  angehalten,
  namensfeld,
  setID,
  stand,
  bearbeiten,
}: SchluesselworteditorEigenschaften) {
  const texte = useTexte().seiten.karten.schluesselwoerter

  // **Ein Wort ohne Namen heißt, wie die Liste es nennt** — angelegt ist es
  // leer, und ein Knopf „Löschen — " ohne Namen wüsste niemand zu deuten.
  const name = wort.wort === '' ? texte.ohneName : wort.wort

  // **Nur für ein eigenes** — und der Editor trägt nur eigene. Ein festes
  // gehört dem Spielpaket; seine Parameter zu bearbeiten hieße, das Paket zu
  // ändern, und das geht über die Paketwerkstatt.
  const hilfen = useParameterhilfen(setID, wort.id, stand)

  // **Die Griffe liegen hier**, nicht im Feld: eine Funktion, die
  // `ref.current` liest, darf nicht beim Zeichnen weitergereicht werden.
  const musterfeld = useRef<HTMLInputElement>(null)
  const erinnerungsfeld = useRef<HTMLInputElement>(null)

  // **Je Feld ein Entwurf** (U06, #2122) — hier und nicht im Feld, damit
  // der Einschub aus dem Menü ihn erreicht. Die Regel steht in `useEntwurf`.
  const entwuerfe = {
    wort: useEntwurf(wort.wort, (wert, bekannt) => {
      bearbeiten.schreiben('wort', wert, bekannt)
    }),
    muster: useEntwurf(wort.muster, (wert, bekannt) => {
      bearbeiten.schreiben('muster', wert, bekannt)
    }),
    erinnerung: useEntwurf(wort.erinnerung, (wert, bekannt) => {
      bearbeiten.schreiben('erinnerung', wert, bekannt)
    }),
    regeln: useEntwurf(wort.regeln, (wert, bekannt) => {
      bearbeiten.schreiben('regeln', wert, bekannt)
    }),
  }

  return (
    <div
      role="group"
      aria-label={`${texte.bearbeiten} — ${name}`}
      className="flex flex-col gap-2 rounded-sm border border-kante p-3"
    >
      <div className="flex flex-wrap items-center gap-2">
        <Schluesselwortfeld
          aufschrift={texte.wort}
          wert={wort.wort}
          angehalten={angehalten}
          entwurf={entwuerfe.wort}
          {...(namensfeld === undefined ? {} : { feld: namensfeld })}
        />

        <Schluesselwortfeld
          aufschrift={texte.muster}
          wert={wort.muster}
          angehalten={angehalten}
          entwurf={entwuerfe.muster}
          feld={musterfeld}
          /* **Ein Schlüsselwort ohne greifendes Muster erkennt nichts.** MSE
             zeigt es rot und schaltet es ab — „that would make MSE hang"
             (`Keyword::prepare`, `src/data/keyword.cpp`, Zeilen 205–209,
             Fassung f9d9356 vom 2021-07-06). Der Satz stand bis Werkstück U3
             von Phase 5.1 (#1414) als namenloser `<span>` unter der Zeile;
             §6 verlangt ihn am **verursachenden** Feld, und das ist das
             Muster. */
          befund={wort.gueltig ? undefined : texte.ungueltig}
          menue={
            <Parameterwahl
              aufschrift={texte.parameterEinfuegen}
              eintraege={hilfen.arten}
              angehalten={angehalten}
              aufWahl={(text) => {
                einfuegenAnDerMarke(
                  musterfeld,
                  text,
                  entwuerfe.muster.uebernehmen,
                )
              }}
            />
          }
        />

        {modi.length > 0 && (
          <label className="flex flex-col gap-1 text-sm">
            {texte.modus}
            {/* **Der gespeicherte Modus ist der Schlüssel** — eine Rücknahme
                erreicht die Wahl, ohne den Editor neu zu bauen (U06). */}
            <select
              key={wort.modus}
              className={eingabe}
              defaultValue={wort.modus}
              disabled={angehalten}
              onChange={(ereignis) => {
                bearbeiten.schreiben('modus', ereignis.target.value)
              }}
            >
              {modi.map((modus) => (
                <option key={modus} value={modus}>
                  {modus}
                </option>
              ))}
            </select>
          </label>
        )}

        <Aktionsknopf
          aktion="loeschen"
          gewicht="neben"
          bezeichnung={`${texte.loeschen} — ${name}`}
          angehalten={angehalten}
          aufKlick={bearbeiten.loeschen}
        >
          {texte.loeschen}
        </Aktionsknopf>
      </div>

      <Schluesselwortfeld
        aufschrift={texte.erinnerung}
        wert={wort.erinnerung}
        angehalten={angehalten}
        breit
        entwurf={entwuerfe.erinnerung}
        feld={erinnerungsfeld}
        menue={
          <Parameterwahl
            aufschrift={texte.parameterVerweisen}
            eintraege={hilfen.verweise}
            angehalten={angehalten}
            aufWahl={(text) => {
              einfuegenAnDerMarke(
                erinnerungsfeld,
                text,
                entwuerfe.erinnerung.uebernehmen,
              )
            }}
          />
        }
      />

      <Schluesselwortfeld
        aufschrift={texte.regeln}
        wert={wort.regeln}
        angehalten={angehalten}
        breit
        entwurf={entwuerfe.regeln}
      />

      {/* **Ein Verweisskript, das nicht lief, sagt es hier** (§8). Das Menü
          bleibt: die übrigen Formen stehen weiter zur Wahl. */}
      {hilfen.befunde.map((befund) => (
        <span key={befund} className="text-sm text-hinweis">
          {befund}
        </span>
      ))}
    </div>
  )
}
