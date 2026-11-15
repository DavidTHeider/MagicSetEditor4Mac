import { useState } from 'react'

import type { KIVorschlagAntwort } from '../antwortenmarkt'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

interface VorschlagslisteEigenschaften {
  vorschlag: KIVorschlagAntwort
  eigene: ReturnType<typeof useTexte>['seiten']['einstellungen']
  laeuft: boolean
  uebernehmen: (felder: string[]) => void
  verwerfen: () => void
}

// Vorschlagsliste zeigt, was das Modell gegeben hat.
//
// **Ein eigenes Bauteil**, weil `KIVorschlag` sonst über die Zeilengrenze
// aus §2 ginge — und weil der Schnitt der Frage folgt: dort die Bedienung,
// hier die Anzeige.
export function Vorschlagsliste({
  vorschlag,
  eigene,
  laeuft,
  uebernehmen,
  verwerfen,
}: VorschlagslisteEigenschaften) {
  // **Zu Anfang ist alles angehakt** — und das ist keine stille Übernahme:
  // geschrieben wird erst auf Klick, und daneben steht je Feld, was sich
  // ändert. Ein Vorschlag, bei dem man acht Haken setzen muss, bevor
  // irgendetwas geht, ist Arbeit ohne Ertrag; abhaken geht schneller als
  // anhaken, und beides ist sichtbar.
  const [gewaehlt, setzeGewaehlt] = useState(() =>
    vorschlag.felder.map((feld) => feld.feld),
  )

  return (
    <div className="flex flex-col gap-2">
      {/* **Wer geantwortet hat, steht am Vorschlag** — das Issue verlangt
          es (A8, #1447): „Provider-/Modellname im Vorschlag sichtbar". Der
          Alias allein sagte es nicht, seit es zwei Anbieter gibt. */}
      <p className="text-sm text-leise">
        {einsetzen(eigene.kiVorschlagAntwortVon, {
          anbieter: vorschlag.anbieter,
          modell: vorschlag.modell,
        })}
      </p>

      <p className="text-sm text-leise">
        {einsetzen(eigene.kiVorschlagAblauf, { ablauf: vorschlag.ablauf })}
      </p>

      {vorschlag.felder.length === 0 ? (
        <p className="text-sm text-leise">{eigene.kiVorschlagNichts}</p>
      ) : (
        <ul className="flex flex-col gap-2 text-sm">
          {vorschlag.felder.map((feld) => (
            <li key={feld.feld} className="flex flex-col gap-0.5">
              {/* **Ein Haken je Feld** (A11, #1450): übernommen wird, was
                  angehakt ist. Die Beschriftung nennt das Feld — ohne sie
                  hieße jeder Haken „Kästchen". */}
              <label className="flex items-start gap-2">
                <input
                  type="checkbox"
                  className="mt-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                  checked={gewaehlt.includes(feld.feld)}
                  aria-label={einsetzen(eigene.kiVorschlagAuswahl, {
                    feld: feld.feld,
                  })}
                  onChange={(ereignis) => {
                    setzeGewaehlt(
                      ereignis.target.checked
                        ? [...gewaehlt, feld.feld]
                        : gewaehlt.filter((name) => name !== feld.feld),
                    )
                  }}
                />

                <span className="flex flex-col">
                  <span className="font-medium">{feld.feld}</span>
                  <span>{feld.werte.join(', ')}</span>

                  {/* **Was heute dasteht, steht daneben** — das Issue
                      verlangt „Feld-Diff": ein Vorschlag ohne den alten Wert
                      verlangt, dass jemand ihn im Kopf hat. */}
                  <span className="text-leise">
                    {feld.bisher === ''
                      ? eigene.kiVorschlagBisherLeer
                      : einsetzen(eigene.kiVorschlagBisher, {
                          wert: feld.bisher,
                        })}
                  </span>
                </span>
              </label>
            </li>
          ))}
        </ul>
      )}

      {/* **Was das Schema nicht durchgelassen hat, steht daneben** — und
          wird nicht geschrieben. */}
      {vorschlag.befunde.length > 0 && (
        <ul className="flex flex-col gap-1 text-sm text-leise">
          {vorschlag.befunde.map((befund) => (
            <li key={befund}>{befund}</li>
          ))}
        </ul>
      )}

      {/* **Keine Auswahl heißt: nichts zu übernehmen** — und der Knopf sagt
          das, statt still nichts zu tun (§6). */}
      {vorschlag.felder.length > 0 && gewaehlt.length === 0 && (
        <p className="text-sm text-leise">{eigene.kiVorschlagNichtsGewaehlt}</p>
      )}

      <div className="flex gap-2">
        <button
          type="button"
          className="rounded-sm border border-kante px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
          disabled={laeuft || gewaehlt.length === 0}
          onClick={() => {
            uebernehmen(gewaehlt)
          }}
        >
          {eigene.kiVorschlagUebernehmen}
        </button>

        <button
          type="button"
          className="rounded-sm border border-kante px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
          onClick={verwerfen}
        >
          {eigene.kiVorschlagAbbrechen}
        </button>
      </div>
    </div>
  )
}
