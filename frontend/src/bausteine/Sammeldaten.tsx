// **Die Sammeldaten von Scryfall**
// (#888).
//
// # Warum dieser Kasten auf der Paketseite steht
//
// Weil er dasselbe tut wie sie: er holt etwas Großes von außen und legt es auf
// die Platte. Auf der Kartenseite stünde eine Verwaltungsaufgabe zwischen der
// Arbeit an einer Karte.
import { useState } from 'react'

import { api } from '../api'
import type { SammeldatenAntwort } from '../apiscryfall'
import { einsetzen } from '../sprache/einsetzen'
import { alsZeitpunkt } from '../zeitpunkt'
import { useDarstellung, useZeitzone, useTexte } from '../sprache/kontext'
import { useLaden } from '../haken/useLaden'
import { Knopf } from './Knopf'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

// Sammeldaten zeigt den Stand und holt die Datei.
export function Sammeldaten(): React.JSX.Element {
  const texte = useTexte().seiten.pakete.sammeldaten
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()

  const [geholt, setzeGeholt] = useState<SammeldatenAntwort | undefined>(
    undefined,
  )
  const [laeuft, setzeLaeuft] = useState(false)
  const [meldung, setzeMeldung] = useState<string | undefined>(undefined)

  // **Über [useLaden] und nicht von Hand** (§7, Regel 2).
  //
  // Der erste Anlauf führte einen eigenen `gilt`-Wächter im Effekt — das ist
  // die zweite Antwort auf eine Frage, die dieser Haken beantwortet, und die
  // zweite läuft irgendwann anders. Er liest nur die Platte und fragt Scryfall
  // nach der angebotenen Größe; ohne Netz kommt er trotzdem, mit dem, was hier
  // liegt.
  const geladen = useLaden(() => api().ScryfallSammeldaten(), [])

  // **Der geholte Stand gewinnt über den geladenen.** Nach einem Abruf ist er
  // der neuere; vorher gibt es ihn nicht.
  const stand = geholt ?? geladen.daten

  const holen = async () => {
    setzeLaeuft(true)

    try {
      const antwort = await api().ScryfallSammeldatenHolen()

      if (antwort.fehler !== undefined && antwort.fehler !== null) {
        setzeMeldung(antwort.fehler.meldung)

        return
      }

      setzeMeldung(undefined)
      setzeGeholt(antwort)
    } finally {
      setzeLaeuft(false)
    }
  }

  // **Ohne eigene Überschrift** — seit UXR-17 (#1925) steht der Kasten
  // eingeklappt unter dem Arbeitsplatz der Pakete, und dessen Schalter trägt
  // den Namen.
  return (
    <section>
      <Erklaerungsabsatz
        text={texte.erklaerung}
        klasse="max-w-prose text-leise"
      />

      <p className="mt-2 text-sm">
        {stand?.vorhanden === true
          ? einsetzen(texte.vorhanden, { mb: stand.mb.toFixed(1) })
          : texte.fehlt}
      </p>

      {/* **Wie alt die liegende Datei ist** (F04, #2246) — ohne Netz die
          einzige Auskunft darüber; der Stand von Scryfall steht darunter,
          wenn es antwortet. */}
      {stand?.vorhanden === true && stand.geholt !== '' && (
        <p className="text-sm text-leise">
          {einsetzen(texte.geholt, {
            zeit: alsZeitpunkt(stand.geholt, darstellung, zeitzone),
          })}
        </p>
      )}

      {stand?.angebotenMB !== undefined && stand.angebotenMB > 0 && (
        <p className="text-sm text-leise">
          {einsetzen(texte.angeboten, {
            mb: stand.angebotenMB.toFixed(1),
            stand: alsZeitpunkt(stand.stand, darstellung, zeitzone),
          })}
        </p>
      )}

      {meldung !== undefined && (
        <p className="mt-2 text-sm text-fehler" role="alert">
          {meldung}
        </p>
      )}

      <div className="mt-3">
        <Knopf
          angehalten={laeuft}
          aufKlick={() => {
            void holen()
          }}
        >
          {laeuft ? texte.laedt : texte.holen}
        </Knopf>
      </div>
    </section>
  )
}
