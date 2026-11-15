// **Ähnliche Karten im Editor** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Gesucht wird mit dem Text, nicht mit dem Namen
//
// **Das ist der Unterschied zur Volltextsuche daneben.** Zwei Karten können
// denselben Namen tragen, und dieselbe Karte hat in zwei Sprachen zwei — A2
// sagt es als eigene Regel, und A15 verbietet ausdrücklich die
// „Deduplizierung nach Namen".
//
// # Und die Karte selbst steht nicht in der Liste
//
// **Sie wäre immer der erste Treffer.** Ein Ergebnis, das mit der Frage
// beginnt, beantwortet sie nicht.
import { useId, useState } from 'react'

import { Bestandsfreigabe } from './Bestandsfreigabe'
import { Knopf } from './Knopf'
import { api } from '../fassade'
import type { RechercheAntwort } from '../apirecherche'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { Recherchetreffer } from '../seiten/Recherchetreffer'
import { derLeerfilter } from '../seiten/recherchefeld'
import { auspacken } from '../umschlag'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { KIEinrichtungsverweis } from './KIEinrichtungsverweis'

interface AehnlicheEigenschaften {
  karteID: string
}

export function AehnlicheKarten({ karteID }: AehnlicheEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.recherche
  const vorgang = useVorgang()
  const ueberschrift = useId()

  const [ergebnis, setzeErgebnis] = useState<RechercheAntwort>()
  const [fehler, setzeFehler] = useState<string>()

  // **Leer, und das ist die Vorgabe** (A13.4, #1452): ohne ausdrückliche
  // Auswahl bleibt der eigene Bestand außen — und dann findet diese Suche
  // nur Karten, die es auch bei Scryfall gibt.
  const [freigegeben, setzeFreigegeben] = useState<string[]>([])

  const suchen = () =>
    vorgang.starte('aehnlich', async () => {
      try {
        // **Ohne Stand, und die Antwort nennt ihn.** Den gerechneten Stand
        // einer Karte rechnet der Bestand; das Fenster bekommt ihn erst mit
        // der Antwort zu sehen — und gibt ihn für die nächste Seite zurück.
        const antwort = await api().AehnlicheKarten(
          Number(karteID),
          '',
          10,
          { ...derLeerfilter, freigegeben },
          '',
        )

        setzeErgebnis(auspacken(antwort, () => antwort))
        setzeFehler(undefined)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  return (
    <section
      className="flex flex-col gap-2 border-t border-kante pt-4"
      aria-labelledby={ueberschrift}
    >
      <h3 id={ueberschrift} className="font-semibold">
        {eigene.aehnlichTitel}
      </h3>

      <Erklaerungsabsatz
        text={eigene.aehnlichErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      {/* **Die Freigabe steht über dem Knopf** (§6): eine Entscheidung, die
          man erst nach dem Auslösen sähe, ist keine. */}
      <Bestandsfreigabe
        freigegeben={freigegeben}
        waehlen={(neu) => {
          // **Ein Wechsel räumt das Ergebnis mit** (§7, Regel 1): eine
          // Trefferliste ohne den eigenen Bestand unter einem gesetzten
          // Kreuz wäre eine Antwort auf eine andere Frage.
          setzeFreigegeben(neu)
          setzeErgebnis(undefined)
        }}
      />

      <Knopf
        gewicht="leise"
        aufKlick={() => void suchen()}
        angehalten={vorgang.laeuft('aehnlich')}
      >
        {eigene.aehnlichSuchen}
      </Knopf>

      {ergebnis === undefined ? (
        <p className="text-sm text-leise">{eigene.aehnlichLeer}</p>
      ) : (
        // **Ohne `oeffnen`**: die Karte, zu der gesucht wurde, ist schon
        // offen — ein Weg „dorthin" führte von hier nach hier.
        <Recherchetreffer ergebnis={ergebnis} oeffnen={undefined} />
      )}

      {/* **Mit dem Weg zur Einrichtung** (AI-01.AK-04, #1742): eine

          KI-Meldung heißt meist, dass ein Anbieter oder Modell fehlt. */}

      {fehler !== undefined && (
        <div className="flex flex-col items-start gap-1">
          <p role="alert" className="text-sm text-fehler">
            {fehler}
          </p>
          <KIEinrichtungsverweis />
        </div>
      )}
    </section>
  )
}
