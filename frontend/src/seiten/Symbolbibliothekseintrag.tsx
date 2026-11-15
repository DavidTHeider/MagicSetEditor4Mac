// **Ein Symbol der Bibliothek** — DOM-02
// (#1697).
//
// Vorschau, Name, wer es trägt, seine Marken und was sich mit ihm tun lässt.
// **Das Umbenennen steht in der Zeile** und nicht in einem Dialog (§6: keine
// verschachtelten Dialoge); sein Fehler steht an seinem Feld.
import { useId, useState } from 'react'

import type { BibliothekssymbolDTO } from '../apisymbolbibliothek'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Knopf } from '../bausteine/Knopf'
import { Markenzeile } from '../bausteine/Markenzeile'
import type { Markenfilterstand } from '../haken/useMarkenfilter'
import type { Symbolbibliothekstand } from '../haken/useSymbolbibliothek'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Symbolfassungen } from './Symbolfassungen'

interface EintragEigenschaften {
  symbol: BibliothekssymbolDTO
  setID: string
  bibliothek: Symbolbibliothekstand
  marken: Markenfilterstand

  // bearbeiten öffnet das Symbol im Editor — ohne gibt es den Knopf nicht.
  bearbeiten?: ((symbol: BibliothekssymbolDTO) => void) | undefined
}

export function Symbolbibliothekseintrag({
  symbol,
  setID,
  bibliothek,
  marken,
  bearbeiten,
}: EintragEigenschaften) {
  const eigene = useTexte().seiten.karten.symbolbibliothek
  const name = useId()
  const fehler = useId()

  // entwurf ist der neue Name — undefined, solange das Feld zu ist.
  const [entwurf, setzeEntwurf] = useState<string>()
  const [fassungenOffen, setzeFassungenOffen] = useState(false)

  const befund =
    bibliothek.befund?.wo === symbol.id ? bibliothek.befund : undefined
  const amNamen =
    befund !== undefined && befund.amNamen && entwurf !== undefined
  const laeuft = bibliothek.laeuft(symbol.id)

  return (
    <li
      aria-labelledby={name}
      className="flex flex-col gap-2 rounded-sm border border-kante p-2"
    >
      <div className="flex flex-wrap items-center gap-3">
        {symbol.vorschau === '' ? (
          <span aria-hidden="true" className="size-12 rounded-sm bg-kante" />
        ) : (
          <img
            src={symbol.vorschau}
            alt={einsetzen(eigene.vorschau, { name: symbol.name })}
            className="size-12 rounded-sm bg-bildhell"
          />
        )}
        <div className="flex flex-col">
          <span id={name} className="font-medium">
            {symbol.name}
          </span>
          <span className="text-sm text-leise">
            {symbol.verwendet.length === 0
              ? eigene.unverwendet
              : einsetzen(eigene.verwendet, {
                  sets: symbol.verwendet.join(', '),
                })}
          </span>
        </div>
      </div>

      <Markenzeile
        art="symbol"
        kennung={symbol.id}
        name={symbol.name}
        marken={marken.je[symbol.id] ?? []}
        nachladen={marken.nachladen}
      />

      {entwurf !== undefined && (
        <div className="flex flex-wrap items-end gap-2">
          <label className="flex flex-col gap-1 text-sm">
            {einsetzen(eigene.neuerName, { name: symbol.name })}
            <input
              type="text"
              value={entwurf}
              aria-invalid={amNamen}
              aria-describedby={amNamen ? fehler : undefined}
              onChange={(ereignis) => {
                setzeEntwurf(ereignis.target.value)
              }}
              className="rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            />
          </label>
          <Knopf
            gewicht="haupt"
            angehalten={laeuft}
            aufKlick={() => {
              void bibliothek.umbenennen(symbol, entwurf).then((geglueckt) => {
                if (geglueckt) {
                  setzeEntwurf(undefined)
                }
              })
            }}
          >
            {eigene.namenUebernehmen}
          </Knopf>
          <Knopf
            aufKlick={() => {
              setzeEntwurf(undefined)
            }}
          >
            {eigene.umbenennenAbbrechen}
          </Knopf>
        </div>
      )}

      <div className="flex flex-wrap gap-2">
        <Knopf
          gewicht="haupt"
          angehalten={laeuft}
          aufKlick={() => void bibliothek.einsetzen(symbol)}
        >
          {eigene.einsetzen}
        </Knopf>
        <Aktionsknopf
          aktion="duplizieren"
          angehalten={laeuft}
          aufKlick={() => void bibliothek.kopieren(symbol)}
        >
          {eigene.duplizieren}
        </Aktionsknopf>
        <Aktionsknopf
          aktion="bearbeiten"
          angehalten={laeuft || entwurf !== undefined}
          aufKlick={() => {
            setzeEntwurf(symbol.name)
          }}
        >
          {eigene.umbenennen}
        </Aktionsknopf>
        <Knopf
          angehalten={laeuft}
          aufKlick={() => void bibliothek.ausgeben(symbol)}
        >
          {eigene.ausgeben}
        </Knopf>
        {bearbeiten !== undefined && (
          <Aktionsknopf
            aktion="bearbeiten"
            aufKlick={() => {
              bearbeiten(symbol)
            }}
          >
            {eigene.bearbeiten}
          </Aktionsknopf>
        )}
        {symbol.vorige && (
          <Aktionsknopf
            aktion="zuruecknehmen"
            angehalten={laeuft}
            aufKlick={() => void bibliothek.zuruecknehmen(symbol)}
          >
            {eigene.zuruecknehmen}
          </Aktionsknopf>
        )}
        <Knopf
          gedrueckt={fassungenOffen}
          aufKlick={() => {
            setzeFassungenOffen(!fassungenOffen)
          }}
        >
          {fassungenOffen ? eigene.fassungenZu : eigene.fassungen}
        </Knopf>
        <Aktionsknopf
          aktion="loeschen"
          angehalten={laeuft}
          aufKlick={() => {
            bibliothek.fragen(symbol)
          }}
        >
          {eigene.loeschen}
        </Aktionsknopf>
      </div>

      {fassungenOffen && (
        <Symbolfassungen id={symbol.id} name={symbol.name} setID={setID} />
      )}

      {befund !== undefined &&
        (amNamen ? (
          <p id={fehler} className="text-sm text-fehler">
            {befund.text}
          </p>
        ) : (
          <p role="alert" className="text-sm text-fehler">
            {befund.text}
          </p>
        ))}
    </li>
  )
}
