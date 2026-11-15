// **Die Symbolbibliothek** — DOM-02
// (#1697).
//
// Unter dem Symboleditor im Fach *Symbol*: Symbole des Profils mit Vorschau,
// Suche und Marken, dazu Aufnehmen, Einlesen und je Symbol Einsetzen,
// Duplizieren, Umbenennen, Ausgeben und Löschen (AK-01). **Die
// Symbolschriften stehen getrennt** und nur zur Ansicht — sie sind Pakete
// und gehören dem Rechner.
//
// *Was* sie tut, steht im Haken (`useSymbolbibliothek`), hier nur, wie es
// aussieht (§2).
import { useId } from 'react'

import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Entscheidung } from '../bausteine/Entscheidung'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Knopf } from '../bausteine/Knopf'
import { Ladefehler } from '../bausteine/Ladehinweise'
import { Markenfilterknopf } from '../bausteine/Markenfilterkasten'
import { nachMarken, useMarkenfilter } from '../haken/useMarkenfilter'
import type { BibliothekssymbolDTO } from '../apisymbolbibliothek'
import {
  aufnahme,
  type Bibliotheksanschluss,
  einlesung,
  useSymbolbibliothek,
} from '../haken/useSymbolbibliothek'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Symbolbibliothekseintrag } from './Symbolbibliothekseintrag'
import { Symbolschriften } from './Symbolschriften'

interface SymbolbibliothekEigenschaften extends Bibliotheksanschluss {
  setID: string

  // bearbeiten öffnet ein Symbol im Editor darüber (AK-02) — ohne gibt es
  // den Knopf nicht.
  bearbeiten?: ((symbol: BibliothekssymbolDTO) => void) | undefined
}

export function Symbolbibliothek({
  setID,
  bearbeiten,
  ...anschluss
}: SymbolbibliothekEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten.symbolbibliothek
  const bibliothek = useSymbolbibliothek(setID, anschluss)
  const marken = useMarkenfilter('symbol')
  const namensfehler = useId()

  const { befund, symbole, frage } = bibliothek
  const amNamen = befund?.wo === aufnahme && befund.amNamen
  const gezeigte = nachMarken(
    symbole.daten ?? [],
    (symbol) => symbol.id,
    marken.getroffen,
  )

  return (
    // **Der Name steht am Bereich und nicht als Überschrift** — UXR-07
    // (#1915): das Fach *Symbol* legt die Bibliothek in einen Kasten, dessen
    // Knopf schon so heißt, und zwei Namen für dieselbe Sache sind einer zu
    // viel (`Aufklappkasten`). Der Kasten trägt auch Rand und Abstand.
    <section aria-label={eigene.ueberschrift} className="flex flex-col gap-3">
      <Erklaerungsabsatz text={eigene.erklaerung} />

      <div className="flex flex-wrap items-end gap-2">
        <label className="flex flex-col gap-1 text-sm">
          {eigene.name}
          <input
            type="text"
            value={bibliothek.name}
            aria-invalid={amNamen}
            aria-describedby={amNamen ? namensfehler : undefined}
            onChange={(ereignis) => {
              bibliothek.setzeName(ereignis.target.value)
            }}
            className="rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
        </label>
        <Aktionsknopf
          aktion="anlegen"
          angehalten={bibliothek.laeuft(aufnahme)}
          aufKlick={() => void bibliothek.aufnehmen()}
        >
          {eigene.aufnehmen}
        </Aktionsknopf>
        <Knopf
          angehalten={bibliothek.laeuft(einlesung)}
          aufKlick={() => void bibliothek.einlesen()}
        >
          {eigene.einlesen}
        </Knopf>
      </div>

      {amNamen && (
        <p id={namensfehler} className="text-sm text-fehler">
          {befund.text}
        </p>
      )}
      {befund !== undefined &&
        !amNamen &&
        (befund.wo === aufnahme || befund.wo === einlesung) && (
          <p role="alert" className="text-sm text-fehler">
            {befund.text}
          </p>
        )}
      {bibliothek.hinweis !== undefined && (
        <p role="status" className="text-sm text-leise">
          {bibliothek.hinweis}
        </p>
      )}

      <div className="flex flex-wrap items-end gap-2">
        <label className="flex flex-col gap-1 text-sm">
          {eigene.suche}
          <input
            type="search"
            value={bibliothek.suche}
            onChange={(ereignis) => {
              bibliothek.setzeSuche(ereignis.target.value)
            }}
            className="rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
        </label>
        <Markenfilterknopf art="symbol" stand={marken} />
      </div>

      {symbole.fehler !== undefined ? (
        <Ladefehler fehler={symbole.fehler} wiederholen={symbole.nachladen} />
      ) : (
        symbole.daten !== undefined &&
        (gezeigte.length === 0 ? (
          bibliothek.suche === '' && marken.getroffen === undefined ? (
            <p className="max-w-prose text-sm text-leise">{eigene.leer}</p>
          ) : (
            // **Mit dem Weg zurück** (G02, #2249) — der Satz riet, eines
            // von beiden zu ändern, und ließ beide stehen.
            <div className="flex max-w-prose flex-col items-start gap-2 text-sm">
              <p className="text-leise">{eigene.keinTreffer}</p>
              <Knopf
                aufKlick={() => {
                  bibliothek.setzeSuche('')
                  marken.setzeFilter({ marken: [], verknuepfung: 'und' })
                }}
              >
                {texte.bausteine.listenansicht.filterZuruecksetzen}
              </Knopf>
            </div>
          )
        ) : (
          <ul aria-label={eigene.liste} className="flex flex-col gap-2">
            {gezeigte.map((symbol) => (
              <Symbolbibliothekseintrag
                key={symbol.id}
                symbol={symbol}
                setID={setID}
                bibliothek={bibliothek}
                marken={marken}
                bearbeiten={bearbeiten}
              />
            ))}
          </ul>
        ))
      )}

      {frage !== undefined && (
        <Entscheidung
          frage={einsetzen(eigene.loeschenFrage, { name: frage.name })}
          erklaerung={
            frage.verwendet.length === 0
              ? eigene.loeschenOhneFolgen
              : einsetzen(eigene.loeschenFolgen, {
                  sets: frage.verwendet.join(', '),
                })
          }
          bestaetigen={eigene.loeschenJa}
          abbrechen={eigene.loeschenNein}
          aktion="loeschen"
          laeuft={bibliothek.laeuft(frage.id)}
          aufBestaetigen={() => void bibliothek.loeschen(frage)}
          aufAbbrechen={() => {
            bibliothek.fragen(undefined)
          }}
        />
      )}

      <Symbolschriften />
    </section>
  )
}
