// **Die Recherche im Sucharbeitsplatz** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Warum hier und nicht auf einer eigenen Seite
//
// **Weil das Issue es ausschließt**: „Kein neuer Hauptbereich/Shortcut und
// keine Paralleloberfläche zum bestehenden Sucharbeitsplatz." Wer etwas
// sucht, sucht an einer Stelle — die Frage „meine Karten oder das Regelwerk?"
// ist eine Wahl innerhalb der Suche und keine zweite Anwendung.
//
// # Der Modus und der Grund stehen über den Treffern
//
// **Weil man beides an ihnen nicht sieht.** Eine lexikalische Antwort ist
// eine andere Auskunft als eine hybride, und eine leere Liste heißt sechs
// verschiedene Dinge — fünf davon sagen jemandem, was zu tun ist.
import { useId, useState } from 'react'

import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { Bestandsfreigabe } from '../bausteine/Bestandsfreigabe'
import { Knopf } from '../bausteine/Knopf'
import { api } from '../fassade'
import type { RechercheAntwort } from '../apirecherche'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { Recherchetreffer } from './Recherchetreffer'
import { Kontextfreigabe } from './Kontextfreigabe'
import { derLeerfilter, feldklassen } from './recherchefeld'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface WissensrechercheEigenschaften {
  // oeffnen bringt einen Treffer aus diesem Bestand in den Karteneditor.
  oeffnen: (setID: string, karteID: string) => void
}

export function Wissensrecherche({ oeffnen }: WissensrechercheEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.recherche
  const vorgang = useVorgang()
  const ueberschrift = useId()

  const [fall, setzeFall] = useState('karten')
  const [frage, setzeFrage] = useState('')
  const [ergebnis, setzeErgebnis] = useState<RechercheAntwort>()
  const [fehler, setzeFehler] = useState<string>()

  // **Leer, und das ist die Vorgabe** (A13.4, #1452).
  const [freigegeben, setzeFreigegeben] = useState<string[]>([])

  const suchen = () =>
    vorgang.starte('recherche', async () => {
      try {
        const antwort = await api().Recherchieren(
          fall,
          frage,
          10,
          { ...derLeerfilter, freigegeben },
          '',
        )

        // **Der Umschlag wird geöffnet und nicht gelesen**: ein `fehler`
        // darin wirft, und der Satz kommt aus demselben Katalog wie jeder
        // andere. Wer ihn von Hand herausnähme, hätte einen zweiten Weg zu
        // derselben Meldung.
        setzeErgebnis(auspacken(antwort, () => antwort))
        setzeFehler(undefined)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  return (
    <section
      className="mt-8 flex flex-col gap-3 border-t border-kante pt-6"
      aria-labelledby={ueberschrift}
    >
      <h3 id={ueberschrift} className="text-lg font-semibold">
        {eigene.rechercheTitel}
      </h3>

      <Erklaerungsabsatz
        text={eigene.rechercheErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      <div className="flex flex-wrap items-end gap-3">
        <label className="flex flex-col gap-1 text-sm">
          {eigene.rechercheFall}

          <select
            className={feldklassen}
            value={fall}
            onChange={(ereignis) => {
              // **Ein Wechsel räumt das Ergebnis mit** (§7, Regel 1): eine
              // Trefferliste aus dem Regelwerk unter der Überschrift „Karten"
              // wäre eine Antwort auf eine Frage, die niemand gestellt hat.
              setzeFall(ereignis.target.value)
              setzeErgebnis(undefined)
            }}
          >
            <option value="karten">{eigene.rechercheFallKarten}</option>
            <option value="regeln">{eigene.rechercheFallRegeln}</option>
            <option value="rulings">{eigene.rechercheFallRulings}</option>
          </select>
        </label>

        <label className="flex max-w-xl flex-1 flex-col gap-1 text-sm">
          {eigene.rechercheFrage}

          <input
            type="search"
            className={feldklassen}
            value={frage}
            onChange={(ereignis) => {
              setzeFrage(ereignis.target.value)
            }}
            onKeyDown={(ereignis) => {
              if (ereignis.key === 'Enter') {
                void suchen()
              }
            }}
          />
        </label>

        <Knopf
          gewicht="haupt"
          aufKlick={() => void suchen()}
          angehalten={vorgang.laeuft('recherche')}
        >
          {eigene.rechercheSuchen}
        </Knopf>
      </div>

      {/* **Nur beim Kartenfall** (A13.4, #1452): eine Regelrecherche sucht
          gar nicht im eigenen Bestand, und ein Kreuz ohne Wirkung wäre eine
          Zusage, die niemand einlöst — `domain` weist eine solche Freigabe
          ausdrücklich ab. */}
      {fall === 'karten' && (
        <Bestandsfreigabe
          freigegeben={freigegeben}
          waehlen={(neu) => {
            setzeFreigegeben(neu)
            setzeErgebnis(undefined)
          }}
        />
      )}

      {ergebnis && <Recherchetreffer ergebnis={ergebnis} oeffnen={oeffnen} />}

      {/* **Das Zusammenstellen für ein Modell steht hinter seinem Titel** —
          UXR-05 (#1913). Es ist der seltenste Weg der Seite und trug fünf
          ihrer elf Felder (R3 aus `docs/UIREVIEW.md`: keine Formularwand).
          Zu heißt verborgen, nicht fort: was darin eingetragen war, steht
          beim Wiederaufklappen noch da. */}
      <Aufklappkasten titel={eigene.freigabeTitel}>
        <Kontextfreigabe fall={fall} frage={frage} freigegeben={freigegeben} />
      </Aufklappkasten>

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}
    </section>
  )
}
