// **Ein Kartenfoto lesen lassen** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Warum der Kasten in der Sammlung steht
//
// **Weil die Frage dort gestellt wird.** Wer ein Foto einer Karte hat, will
// wissen, welches Papier das ist — und genau das ist die Frage der Sammlung.
// Im Karteneditor stünde sie neben einer Karte, die man schon hat.
//
// **Und unter der Listeneinfuhr**, weil beide dasselbe tun: die eine deutet
// getippte Zeilen, die andere ein Bild. Beide schreiben nichts.
//
// # Warum es keinen Übernahmeknopf gibt
//
// **Weil das Issue es ausschließt**: „Menge/Exemplar wird **nicht
// automatisch** angelegt." Ein Knopf „ins Regal" wäre die automatische
// Anlage mit einem Klick davor — und bei 72 passenden Drucken wäre er eine
// Wette. Wer ein Exemplar eintragen will, nimmt es im Formular darüber auf.
import { useState } from 'react'

import { KIErkennungsbefund } from './KIErkennungsbefund'
import { KIKostenfreigabe } from './KIKostenfreigabe'
import { KIModellwahl } from './KIModellwahl'
import { aliasAus, anbieterAus, kostetGeld } from './kiwahl'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useKIModelle } from '../haken/useKIModelle'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

import type { KIErkennungAntwort } from '../antwortenerkennung'

export function KIKartenerkennung() {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen
  const vorgang = useVorgang()

  const [offen, setzeOffen] = useState(false)
  const [gefragt, setzeGefragt] = useState(false)

  const katalog = useKIModelle(gefragt)

  const [gewaehlt, setzeGewaehlt] = useState('')
  const [freigabe, setzeFreigabe] = useState(false)

  const [antwort, setzeAntwort] = useState<KIErkennungAntwort>()
  const [fehler, setzeFehler] = useState<string>()

  const laeuft = vorgang.laeuft('erkennung')

  // ausloesen entscheidet, ob vorher nach einer Freigabe zu fragen ist —
  // dieselbe Entscheidung wie beim Vorschlag ([kostetGeld]).
  function ausloesen() {
    if (kostetGeld(katalog.modelle, gewaehlt)) {
      setzeFreigabe(true)

      return
    }

    void erkennen('')
  }

  async function erkennen(nachweis: string) {
    setzeFehler(undefined)

    // **Ein Vorgang je Kasten** (§7, Regel 4): ein zweiter Dateidialog über
    // dem ersten wäre ein Dialog über einem Dialog (§6).
    await vorgang.starte('erkennung', async () => {
      try {
        const gelesen = await api().KIKarteErkennen({
          // **Leer, damit die Bindung nach der Datei fragt.** Der Dialog ist
          // Fenstersache; an den Flächen ohne Fenster steht der Pfad im
          // Aufruf.
          pfad: '',
          anbieter: anbieterAus(gewaehlt),
          alias: aliasAus(gewaehlt),
          nachweis,
        })

        setzeAntwort(auspacken(gelesen, () => gelesen))
      } catch (ursache) {
        setzeAntwort(undefined)
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  return (
    <details
      className="rounded-sm border border-kante px-3 py-2"
      onToggle={(ereignis) => {
        const auf = ereignis.currentTarget.open

        setzeOffen(auf)

        if (auf) {
          setzeGefragt(true)
        }
      }}
    >
      {/* **Zugeklappt, bis jemand ihn aufschlägt** — erst dann wird der
          Modellkatalog gelesen (§16, Regel 6). */}
      <summary className="cursor-pointer text-sm font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring">
        {eigene.kiErkennungTitel}
      </summary>

      {offen && (
        <div className="mt-3 flex flex-col gap-3">
          <p className="max-w-prose text-sm text-leise">
            {eigene.kiErkennungSatz}
          </p>

          <KIModellwahl
            gefragt={gefragt}
            modelle={katalog.modelle}
            laedt={katalog.laedt}
            fehler={katalog.fehler}
            gewaehlt={gewaehlt}
            waehlen={setzeGewaehlt}
          />

          <button
            type="button"
            className="self-start rounded-sm border border-kante px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
            disabled={laeuft || gewaehlt === ''}
            onClick={ausloesen}
          >
            {eigene.kiErkennungLesen}
          </button>

          {freigabe && (
            <KIKostenfreigabe
              anbieter={anbieterAus(gewaehlt)}
              freigegeben={(nachweis) => {
                setzeFreigabe(false)
                void erkennen(nachweis)
              }}
              verzichtet={() => {
                setzeFreigabe(false)
              }}
            />
          )}

          {laeuft && (
            <p role="status" className="text-sm text-leise">
              {eigene.kiErkennungLaeuft}
            </p>
          )}

          {fehler !== undefined && (
            <p role="alert" className="text-sm text-fehler">
              {fehler}
            </p>
          )}

          {antwort === undefined ? (
            <p className="text-sm text-leise">{eigene.kiErkennungLeer}</p>
          ) : (
            <KIErkennungsbefund antwort={antwort} />
          )}
        </div>
      )}
    </details>
  )
}
