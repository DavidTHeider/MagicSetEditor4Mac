// **Womit sich dieses Profil gegenüber anderen ausweist** — Werkstück N1 von
// Phase 5.4
// (#1454).
//
// # Warum der Fingerabdruck groß dasteht
//
// **Weil er das ist, was zwei Menschen vergleichen.** Die Kennung ist 52
// Zeichen lang und trägt den ganzen Schlüssel; niemand liest sie am Telefon
// vor. Ein Anzeigename ist kein Nachweis — das sagt der Satz darunter, und
// das Issue verlangt es.
//
// # Und warum das Ablösen eine Rückfrage hat
//
// **Weil es jede Bestätigung entwertet**, die ein Kontakt gegeben hat. Das
// ist keine Einstellung, sondern eine Handlung mit Folgen bei anderen — und
// §6 verlangt die Frage dort, wo geklickt wird, nicht in einem Dialog.
import { useId, useState } from 'react'

import { Entscheidung } from '../bausteine/Entscheidung'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Knopf } from '../bausteine/Knopf'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

import type { Peeridentitaetsantwort } from '../apipeer'

export function Peeridentitaet() {
  const texte = useTexte()
  const eigene = texte.seiten.technik
  const vorgang = useVorgang()
  const ueberschrift = useId()

  const [gefragt, setzeGefragt] = useState(false)
  const [geaendert, setzeGeaendert] = useState(0)
  const [fehler, setzeFehler] = useState<string>()
  const [fragt, setzeFragt] = useState(false)

  // **Über den Ladehaken** (§7): er führt den Lauf mit und trägt eine Antwort
  // nur ein, wenn sie noch zum aktuellen Stand gehört.
  //
  // **Und erst, wenn jemand aufschlägt.** Der Weg geht an den Schlüsselbund
  // dieses Rechners; zugeklappt zu fragen hieße, ihn bei jedem Blick auf
  // diese Seite anzufassen — dieselbe Gestalt wie bei [useKIModelle], und
  // aus demselben Grund.
  const lage = useLaden<Peeridentitaetsantwort | undefined>(async () => {
    if (!gefragt) {
      return undefined
    }

    const antwort = await api().Peeridentitaet()

    // **Der Fehler kommt im Umschlag und nicht als Sturz** — `auspacken`
    // wirft ihn, und der Ladehaken macht daraus seine Meldung.
    return auspacken(antwort, () => antwort)
  }, [gefragt, geaendert])

  // handeln fährt einen der zwei schreibenden Wege.
  //
  // **Ein Vorgang für beide** (§7, Regel 4): anlegen und ablösen gehen an
  // dieselbe Identität, und zwei Riegel ließen beide zugleich laufen.
  const handeln = (weg: () => Promise<Peeridentitaetsantwort>): Promise<void> =>
    vorgang.starte('peeridentitaet', async () => {
      setzeFehler(undefined)

      try {
        auspacken(await weg(), () => undefined)

        // **Geschrieben ist geschrieben** (§7, Regel 3): nachgeladen wird
        // erst nach dem geglückten Schreiben, und ein Fehler dabei nimmt die
        // Änderung nicht zurück.
        setzeGeaendert((stand) => stand + 1)
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  const laeuft = vorgang.laeuft('peeridentitaet')
  const stand = lage.daten?.stand ?? ''

  return (
    <section aria-labelledby={ueberschrift} className="flex flex-col gap-3">
      <h3 id={ueberschrift} className="text-lg font-semibold">
        {eigene.peerTitel}
      </h3>

      <Erklaerungsabsatz
        text={eigene.peerErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      {/* **Was sie nicht ist** (NET-02, S13.NAT-05, #1706) — wer sie für
          das Apple-Konto oder den Rechnernamen hielt, suchte einen Fehler an
          der falschen Stelle. */}
      <p className="max-w-prose text-sm text-leise">{eigene.peerNichtKonto}</p>

      {!gefragt && (
        <Knopf
          gewicht="leise"
          aufKlick={() => {
            setzeGefragt(true)
          }}
        >
          {eigene.peerZeigen}
        </Knopf>
      )}

      {lage.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {lage.fehler}
        </p>
      )}

      {lage.daten !== undefined && (
        <>
          <p className="text-sm">{standsatz(eigene, stand)}</p>

          {lage.daten.kennung !== '' && (
            <dl className="flex flex-col gap-1 text-sm">
              <dt className="font-medium">{eigene.peerFingerabdruck}</dt>
              <dd className="font-mono">{lage.daten.fingerabdruck}</dd>
              <dt className="font-medium">{eigene.peerKennung}</dt>
              <dd className="break-all font-mono text-xs text-leise">
                {lage.daten.kennung}
              </dd>
            </dl>
          )}

          {lage.daten.kennung !== '' && (
            <Erklaerungsabsatz
              text={eigene.peerFingerabdruckHinweis}
              klasse="max-w-prose text-sm text-leise"
            />
          )}

          {lage.daten.seit !== '' && (
            <p className="text-sm text-leise">
              {einsetzen(eigene.peerSeit, { zeitpunkt: lage.daten.seit })}
            </p>
          )}

          {lage.daten.vorher !== '' && (
            <p className="text-sm text-leise">
              {einsetzen(eigene.peerVorher, {
                kennung: lage.daten.vorherFingerabdruck,
              })}
            </p>
          )}

          {lage.daten.befunde.map((befund) => (
            <p key={befund} role="status" className="max-w-prose text-sm">
              {befund}
            </p>
          ))}

          {fehler !== undefined && (
            <p role="alert" className="text-sm text-fehler">
              {fehler}
            </p>
          )}

          <div className="flex flex-wrap gap-2">
            {stand === 'ohne' && (
              <Aktionsknopf
                aktion="anlegen"
                gewicht="haupt"
                angehalten={laeuft}
                aufKlick={() => {
                  void handeln(() => api().PeeridentitaetSicherstellen())
                }}
              >
                {eigene.peerAnlegen}
              </Aktionsknopf>
            )}

            {stand !== 'ohne' && !fragt && (
              <Knopf
                gewicht="leise"
                angehalten={laeuft}
                aufKlick={() => {
                  setzeFragt(true)
                }}
              >
                {eigene.peerRotieren}
              </Knopf>
            )}
          </div>

          {fragt && (
            <Entscheidung
              frage={eigene.peerRotierenFrage}
              bestaetigen={eigene.peerRotierenJa}
              abbrechen={eigene.peerRotierenNein}
              laeuft={laeuft}
              aufBestaetigen={() => {
                setzeFragt(false)
                void handeln(() => api().PeeridentitaetRotieren())
              }}
              aufAbbrechen={() => {
                setzeFragt(false)
              }}
            />
          )}
        </>
      )}
    </section>
  )
}

// standsatz gibt den Satz zu einem der vier Stände.
//
// **Eine Nachschlagetabelle und kein Typ**: der Stand kommt aus Go und ist
// hier eine gewöhnliche Zeichenkette — dieselbe Überlegung wie bei `satzZu`
// in `Technik.tsx`.
//
// **Der Rückfall ist der Stand selbst.** Ein fünfter, den dieses Fenster noch
// nicht kennt, ist immer noch eine Auskunft; eine leere Zeile wäre keine.
function standsatz(
  eigene: {
    peerStandOhne: string
    peerStandBereit: string
    peerStandVerloren: string
    peerStandFremd: string
  },
  stand: string,
): string {
  const saetze: Record<string, string> = {
    ohne: eigene.peerStandOhne,
    bereit: eigene.peerStandBereit,
    verloren: eigene.peerStandVerloren,
    fremd: eigene.peerStandFremd,
  }

  return saetze[stand] ?? stand
}
