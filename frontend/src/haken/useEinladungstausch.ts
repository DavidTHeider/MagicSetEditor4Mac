// **Was der Einladungstausch tut** — NET-01
// (#1705).
//
// # Die Regeln aus §7, und wo sie hier stehen
//
// **Regel 1 und 2**: die Vorschau gehört zu **dem** Code, der geprüft wurde.
// Wer die Eingabe ändert, räumt sie ab, und eine Antwort, die zu einem
// früheren Code gehört, wird nicht mehr eingetragen.
//
// **Regel 3**: nach einer geglückten Annahme lädt die Liste nach — über
// `nachAufnahme`; scheitert das Nachladen, bleibt die Annahme geschehen.
//
// **Regel 4**: ein Vorgang je Schritt. Die Annahme teilt sich keinen Riegel
// mit den Zeilen der Liste: sie legt einen Kontakt an, den es dort noch nicht
// gibt — und einen, den es dort gibt, fasst sie nur an, wenn er nicht
// gesperrt ist.
import { useRef, useState } from 'react'

import { Bindungsfehler } from '../api'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

import type {
  Einladungsvorschau,
  Kontakteinladung,
} from '../apikontakteinladung'
import type { Einloesemeldung } from '../apipeerkontakt'
import type { Katalog } from '../sprache/kataloge'

type Einladungstexte = Katalog['seiten']['technik']['einladung']

// Einladungsort sagt, an welchem Schritt ein Fehler steht (§6: am Feld).
export type Einladungsort = 'erzeugen' | 'pruefen' | 'annehmen'

export interface Einladungstausch {
  ausgestellt: Kontakteinladung | undefined
  eingabe: string
  setzeEingabe: (eingabe: string) => void
  vorschau: Einladungsvorschau | undefined
  // ohneIdentitaet sagt, dass der Schritt an der fehlenden eigenen Identität
  // scheiterte — dann bietet die Stelle sie an (G01, `g01fan`).
  fehler:
    | {
        ort: Einladungsort
        text: string
        ohneIdentitaet: boolean
      }
    | undefined
  meldung: string | undefined
  laeuft: (ort: Einladungsort) => boolean
  erzeugen: (name: string, zweck: string) => void
  pruefen: () => void
  verwerfen: () => void
  // identitaetAngelegt räumt den Fehler, der das Anlegen auslöste, und prüft
  // den Code, an dem er stand, noch einmal — **nur, wenn beide noch gelten**:
  // eine inzwischen geänderte Eingabe ist eine andere Einladung (§7, Regel 2;
  // Befund 6 des Zweitblicks über Phase G, zweiter Teil).
  identitaetAngelegt: (
    ausloeser: NonNullable<Einladungstausch['fehler']>,
  ) => void
  annehmen: (
    vorschau: Einladungsvorschau,
    eigenerName: string,
    verglichen: boolean,
    freigeben: boolean,
    meinName: string,
  ) => void
}

export function useEinladungstausch(
  nachAufnahme: () => void,
): Einladungstausch {
  const texte = useTexte()
  const eigene = texte.seiten.technik.einladung
  const vorgang = useVorgang()

  const [ausgestellt, setzeAusgestellt] = useState<Kontakteinladung>()
  const [eingabe, setzeRoheEingabe] = useState('')
  const [vorschau, setzeVorschau] = useState<Einladungsvorschau>()
  const [fehler, setzeFehlerstand] = useState<Einladungstausch['fehler']>()

  // **Der Fehler, der gerade gilt**, für eine Antwort, die nach dem Zeichnen
  // ankommt — wie `aktuelleEingabe` darunter.
  const fehlerJetzt = useRef<Einladungstausch['fehler']>(undefined)
  const setzeFehler = (neu: Einladungstausch['fehler']) => {
    fehlerJetzt.current = neu
    setzeFehlerstand(neu)
  }
  const [meldung, setzeMeldung] = useState<string>()

  // **Die Eingabe, zu der eine Antwort passen muss** (Regel 2) — als
  // Referenz, weil die Antwort nach dem Zeichnen ankommt.
  const aktuelleEingabe = useRef('')

  // **Und ein Zähler je Eingabe** (Nachlauf über Phase 6.7, N6): derselbe
  // Code ein zweites Mal ist ein neuer Stand — der Wert allein verwechselte
  // ihn mit dem alten, und eine verspätete Annahme räumte den neuen Vergleich.
  const eingabestand = useRef(0)

  const scheitern = (ort: Einladungsort, ursache: unknown) => {
    setzeFehler({
      ort,
      text: meldungAus(ursache, texte),
      ohneIdentitaet:
        ursache instanceof Bindungsfehler &&
        ursache.schluessel === 'ohnePeeridentitaet',
    })
  }

  const setzeEingabe = (neu: string) => {
    aktuelleEingabe.current = neu
    eingabestand.current += 1
    setzeRoheEingabe(neu)
    setzeVorschau(undefined)
    setzeFehler(undefined)
  }

  const erzeugen = (name: string, zweck: string) => {
    void vorgang.starte('erzeugen', async () => {
      setzeFehler(undefined)

      try {
        const antwort = await api().KontakteinladungErstellen(name, zweck)
        setzeAusgestellt(auspacken(antwort, () => antwort.einladung))
      } catch (ursache) {
        scheitern('erzeugen', ursache)
      }
    })
  }

  const pruefen = () => {
    const code = aktuelleEingabe.current
    const stand = eingabestand.current

    void vorgang.starte('pruefen', async () => {
      setzeFehler(undefined)
      setzeMeldung(undefined)

      try {
        const antwort = await api().KontakteinladungPruefen(code)
        const gelesen = auspacken(antwort, () => antwort.vorschau)

        if (eingabestand.current === stand) {
          setzeVorschau(gelesen)
        }
      } catch (ursache) {
        if (eingabestand.current === stand) {
          scheitern('pruefen', ursache)
        }
      }
    })
  }

  const verwerfen = () => {
    setzeEingabe('')
  }

  const identitaetAngelegt = (
    ausloeser: NonNullable<Einladungstausch['fehler']>,
  ) => {
    // **Nur, wenn der auslösende Fehler noch gilt.** Jede Änderung der
    // Eingabe räumt ihn (`setzeEingabe`) — gilt er noch, ist es dieselbe
    // Einladung. **Die Identität bleibt** in jedem Fall.
    if (fehlerJetzt.current !== ausloeser) {
      return
    }

    setzeFehler(undefined)

    if (ausloeser.ort === 'pruefen') {
      pruefen()
    }
  }

  const annehmen = (
    offen: Einladungsvorschau,
    eigenerName: string,
    verglichen: boolean,
    freigeben: boolean,
    meinName: string,
  ) => {
    // **Die Vorschau, die der Mensch vor sich hatte**, kommt vom Aufrufer —
    // sie ist es, gegen die er verglichen hat.
    const code = aktuelleEingabe.current
    const stand = eingabestand.current

    void vorgang.starte('annehmen', async () => {
      setzeFehler(undefined)

      try {
        const antwort = await api().KontakteinladungAnnehmen(
          code,
          eigenerName,
          verglichen ? offen.vergleichscode : '',
          verglichen && freigeben,
          meinName,
        )
        const gemeldet = auspacken(antwort, () => antwort.gemeldet)
        const name = eigenerName.trim() === '' ? offen.name : eigenerName.trim()

        // **Geschrieben ist geschrieben** (Regel 3): die Liste lädt nach,
        // auch wenn das Feld inzwischen einen anderen Code trägt. **Aber
        // Feld und Meldung gehören dem Code, der jetzt dasteht** (Regel 2;
        // Zweitblick über Phase 6.7, Z6).
        nachAufnahme()

        if (eingabestand.current !== stand) {
          return
        }

        setzeEingabe('')
        setzeMeldung(
          [
            einsetzen(verglichen ? eigene.bestaetigt : eigene.aufgenommen, {
              name,
            }),
            meldesatz(gemeldet, name, eigene),
          ]
            .filter((satz) => satz !== '')
            .join(' '),
        )
      } catch (ursache) {
        if (eingabestand.current === stand) {
          scheitern('annehmen', ursache)
        }
      }
    })
  }

  return {
    ausgestellt,
    eingabe,
    setzeEingabe,
    vorschau,
    fehler,
    meldung,
    laeuft: (ort) => vorgang.laeuft(ort),
    erzeugen,
    pruefen,
    verwerfen,
    identitaetAngelegt,
    annehmen,
  }
}

// meldesatz sagt, was die Meldung beim Einladenden ergab (NET-01b) — leer,
// wenn es nichts zu sagen gibt.
//
// **Ein Grund, den der Katalog nicht kennt, wird der allgemeine**: eine
// neuere Gegenstelle darf ein Ergebnis schicken, das es hier noch nicht gibt.
function meldesatz(
  gemeldet: Einloesemeldung | undefined,
  name: string,
  eigene: Einladungstexte,
): string {
  switch (gemeldet?.stand) {
    case 'gemeldet':
      return einsetzen(eigene.gemeldet, { name })
    case 'ohne':
      return einsetzen(eigene.nichtGemeldet, { name })
    case 'gescheitert': {
      const eintrag = Object.entries(eigene.meldegruende).find(
        ([schluessel]) => schluessel === gemeldet.grund,
      )

      return einsetzen(eigene.meldungGescheitert, {
        name,
        grund:
          eintrag === undefined ? eigene.meldegruende.gescheitert : eintrag[1],
      })
    }
    default:
      return ''
  }
}
