// **Ein Mensch gibt einen bezahlten Auftrag frei** —
// #1539.
//
// # Der Befund
//
// **Am laufenden Fenster ging kein kostenpflichtiger Auftrag hinaus.** Die
// Kostenwache verlangt die ausdrückliche Freigabe eines Menschen
// ([0078](../../../docs/entscheidungen/0078-der-wirksame-akteur.md)), und
// ausgestellt hat sie **keine Bindung**. Der Satz „braucht die ausdrückliche
// Freigabe eines Menschen" stand am Fenster, und es gab keinen Weg, sie zu
// geben.
//
// # Zwei Tore, und sie sind nicht dasselbe
//
//   - **Das Budget** ist die Grenze eines Zeitraums. Es wird unter „Kosten"
//     gesetzt und gilt, bis jemand es ändert.
//   - **Die Freigabe** gilt **einem** Auftrag und läuft nach fünf Minuten ab.
//
// Wer nur eines hätte, hätte entweder eine Vollmacht oder eine Grenze, die
// niemand gelesen hat. Deshalb nennt dieser Kasten beides: er zeigt, was
// heute noch verfügbar ist, und nimmt für **diesen** Auftrag einen
// Höchstbetrag entgegen.
//
// # Warum er ohne Budget keinen Knopf anbietet
//
// **Weil der Ruf sonst am Journal scheiterte** — mit einer Meldung über ein
// Budget, während jemand auf einen Knopf gedrückt hat, der Freigabe hieß.
// §6: ein leerer Zustand erklärt den nächsten Schritt.
import { useState } from 'react'

import { alsGeld } from '../geld'
import { Feld } from './Feld'
import { ziffernbetrag } from '../ziffernbetrag'
import { api } from '../fassade'
import { dienstnamen } from './dienstnamen'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useLaden } from '../haken/useLaden'
import { useDarstellung, useTexte } from '../sprache/kontext'

interface KIKostenfreigabeEigenschaften {
  // anbieter ist das Konto, auf das gebucht wird — `anthropic`, `openai`.
  anbieter: string

  // freigegeben bekommt die Kennung des Nachweises.
  freigegeben: (nachweis: string) => void

  // verzichtet meldet, dass niemand freigegeben hat.
  verzichtet: () => void
}

// derZeitraum ist der, über den die Freigabe gerechnet wird.
//
// **Der Tag und nicht der Monat**: es ist die engere der beiden Grenzen, und
// dieselbe, die `Kostenwache.Reservieren` ohne Angabe nimmt.
const derZeitraum = 'tag'

export function KIKostenfreigabe({
  anbieter,
  freigegeben,
  verzichtet,
}: KIKostenfreigabeEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen
  const darstellung = useDarstellung()

  // **Der Anbieter steht mit seinem Namen da und nicht mit seiner Kennung.**
  //
  // Im Fenster stand „rechnet bei openai" — die Kennung, mit der das Programm
  // rechnet, nicht der Name, den ein Mensch kennt. Eigennamen werden nicht
  // übersetzt (0063), und deshalb stehen sie in `dienstnamen.ts` und in
  // keinem Sprachkatalog.
  const name = dienstnamen[anbieter] ?? anbieter

  const [betrag, setzeBetrag] = useState('')
  const [laeuft, setzeLaeuft] = useState(false)
  const [fehler, setzeFehler] = useState<string>()

  const stand = useLaden(
    () => api().Kostenstand(anbieter, derZeitraum),
    [anbieter],
  )

  // **Ohne gelesenen Stand kein Knopf.** Solange nichts dasteht, ist „kein
  // Budget" und „noch nicht gelesen" dasselbe Bild — und das sind zwei
  // Auskünfte.
  const grenze = stand.daten?.grenze
  const verfuegbar = stand.daten?.verfuegbar
  const gedeckt = grenze !== undefined && grenze.wert > 0

  // **Die Währung kommt vom Konto und nicht aus einem Feld.**
  //
  // Sie steht im Budget, das jemand unter „Kosten" gesetzt hat; wer sie hier
  // tippen ließe, bekäme beim ersten Vertipper die Absage „die Grenze gilt in
  // EUR, die Kosten fallen in EURO an" — und suchte sie am falschen Ort.
  const waehrung = verfuegbar?.waehrung ?? ''

  // **Nur Ziffern** — »1.000« wäre für `Number` ein Cent (N5).
  const zahl = ziffernbetrag(betrag)
  const brauchbar = zahl !== undefined && zahl > 0

  async function erteilen(hoechstbetrag: number) {
    setzeFehler(undefined)
    setzeLaeuft(true)

    try {
      const antwort = await api().KIFreigabeErteilen(
        anbieter,
        hoechstbetrag,
        waehrung,
      )

      freigegeben(auspacken(antwort, () => antwort.kennung))
    } catch (ursache) {
      setzeFehler(meldungAus(ursache, texte))
    } finally {
      setzeLaeuft(false)
    }
  }

  return (
    <div className="mt-3 rounded-sm border border-kante p-3">
      <h4 className="text-sm font-medium">{eigene.kiFreigabeTitel}</h4>

      <p className="mt-1 max-w-prose text-sm text-leise">
        {einsetzen(eigene.kiFreigabeSatz, { anbieter: name })}
      </p>

      {verfuegbar !== undefined && (
        <p className="mt-1 text-sm text-leise">
          {einsetzen(eigene.kiFreigabeVerfuegbar, {
            betrag: alsGeld(verfuegbar.wert, verfuegbar.waehrung, darstellung),
          })}
        </p>
      )}

      {stand.daten !== undefined && !gedeckt && (
        <p className="mt-2 max-w-prose text-sm text-leise">
          {einsetzen(eigene.kiFreigabeOhneBudget, { anbieter: name })}
        </p>
      )}

      {gedeckt && (
        <div className="mt-2">
          <Feld
            beschriftung={eigene.kiFreigabeBetrag}
            {...(betrag === '' || zahl !== undefined
              ? {}
              : { befund: texte.allgemein.nurZiffern })}
          >
            {(verbindung) => (
              <input
                {...verbindung}
                className="w-40 rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                inputMode="numeric"
                value={betrag}
                onChange={(ereignis) => {
                  setzeBetrag(ereignis.target.value)
                }}
              />
            )}
          </Feld>
        </div>
      )}

      <div className="mt-3 flex flex-wrap gap-2">
        {gedeckt && (
          <button
            type="button"
            className="rounded-sm bg-marke px-3 py-1 text-sm text-marketext focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
            disabled={!brauchbar || laeuft}
            onClick={
              zahl === undefined
                ? undefined
                : () => {
                    void erteilen(zahl)
                  }
            }
          >
            {laeuft ? eigene.kiFreigabeLaeuft : eigene.kiFreigabeErteilen}
          </button>
        )}

        <button
          type="button"
          className="rounded-sm border border-kante px-3 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          onClick={verzichtet}
        >
          {eigene.kiFreigabeVerzichten}
        </button>
      </div>

      {(fehler ?? stand.fehler) !== undefined && (
        <p role="alert" className="mt-2 text-sm text-fehler">
          {fehler ?? stand.fehler}
        </p>
      )}
    </div>
  )
}
