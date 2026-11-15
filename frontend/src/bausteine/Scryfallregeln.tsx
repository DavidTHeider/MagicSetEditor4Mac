// **Regelhinweise und die Manakostenprüfung**
// (#888).
//
// # Warum beides in einem Kasten
//
// Beide sind **Auskünfte**: sie ändern nichts an der Karte. Sie neben den
// Bildkasten zu stellen, der schreibt, wäre eine Vermischung — hier steht,
// was Scryfall **weiß**, dort, was es in die Karte bringt.
//
// # Warum die Unterscheidung offiziell/Anmerkung so deutlich ist
//
// Scryfall führt zwei Quellen: `wotc` sind offizielle Regelhinweise,
// `scryfall` sind Anmerkungen des Teams — oft zu einem **inoffiziellen**
// Format wie Duel Commander. Wer beide gleich darstellte, gäbe eine Meinung
// als Regel aus.
import { useState } from 'react'

import { api } from '../api'
import type { RegelauskunftAntwort } from '../apiscryfall'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Knopf } from './Knopf'

interface ScryfallregelnEigenschaften {
  // kartenname ist der Name, mit dem gesucht wird.
  kartenname: string

  // manakosten ist der Wert des Kostenfelds — leer, wenn die Karte keines hat.
  manakosten: string
}

// Scryfallregeln zeigt die Hinweise zu einer Karte.
export function Scryfallregeln({
  kartenname,
  manakosten,
}: ScryfallregelnEigenschaften): React.JSX.Element {
  const texte = useTexte().seiten.karten.regeln

  const [auskunft, setzeAuskunft] = useState<RegelauskunftAntwort | undefined>(
    undefined,
  )
  const [manabefund, setzeManabefund] = useState<string | undefined>(undefined)
  const [laeuft, setzeLaeuft] = useState(false)

  // **Der Fehler bleibt in diesem Kasten** (§6: „Fehler erscheinen am
  // verursachenden Feld").
  //
  // Der erste Anlauf reichte ihn an den Karteneditor weiter, und der zeigt
  // eine Meldung **nur am Feld, das sie verursacht hat** — eine Auskunft
  // verursacht keines. Die Meldung fiel damit still aus: kein Fehler im
  // Ablauf, nur nichts zu sehen. Aufgefallen ist es dem Test, nicht dem Auge.
  const [meldung, setzeMeldung] = useState<string | undefined>(undefined)

  const melden = setzeMeldung

  // **Ein Vorgang je Kasten** (§7 Regel 4): zwei gleichzeitige Abrufe gingen
  // beide gegen dieselbe Drossel und lieferten am Ende den früheren Stand.
  const holen = async () => {
    setzeLaeuft(true)

    try {
      const geholt = await api().ScryfallRegeln(kartenname)

      if (geholt.fehler !== undefined && geholt.fehler !== null) {
        melden(geholt.fehler.meldung)

        return
      }

      melden(undefined)
      setzeAuskunft(geholt)
    } finally {
      setzeLaeuft(false)
    }
  }

  const pruefen = async () => {
    if (manakosten === '') {
      setzeManabefund(texte.manaLeer)

      return
    }

    setzeLaeuft(true)

    try {
      const geprueft = await api().ScryfallManakosten(manakosten)

      if (geprueft.fehler !== undefined && geprueft.fehler !== null) {
        melden(geprueft.fehler.meldung)

        return
      }

      melden(undefined)
      setzeManabefund(
        geprueft.unbekannt.length === 0
          ? texte.manaGut
          : einsetzen(texte.manaSchlecht, {
              symbole: geprueft.unbekannt.join(', '),
            }),
      )
    } finally {
      setzeLaeuft(false)
    }
  }

  return (
    <section className="mt-4 border-t border-kante pt-4">
      <h3 className="text-sm font-semibold">{texte.ueberschrift}</h3>

      <div className="mt-2 flex flex-wrap gap-2">
        <Knopf
          angehalten={laeuft || kartenname.trim() === ''}
          aufKlick={() => {
            void holen()
          }}
        >
          {texte.holen}
        </Knopf>

        <Knopf
          angehalten={laeuft}
          aufKlick={() => {
            void pruefen()
          }}
        >
          {texte.manaPruefen}
        </Knopf>
      </div>

      {meldung !== undefined && (
        <p className="mt-2 text-xs text-fehler" role="alert">
          {meldung}
        </p>
      )}

      {manabefund !== undefined && (
        <p className="mt-2 text-xs text-leise">{manabefund}</p>
      )}

      {auskunft !== undefined && (
        <Hinweisliste auskunft={auskunft} texte={texte} />
      )}
    </section>
  )
}

// Hinweisliste zeigt die geholten Hinweise.
//
// **Eine eigene Komponente**, weil `Scryfallregeln` mit ihr über die
// Zeilengrenze ginge (§2) — und weil der Schnitt einer Frage folgt: dort steht,
// **wie** geholt wird, hier, **wie es aussieht**.
function Hinweisliste({
  auskunft,
  texte,
}: {
  auskunft: RegelauskunftAntwort
  texte: ReturnType<typeof useTexte>['seiten']['karten']['regeln']
}): React.JSX.Element {
  if (auskunft.hinweise.length === 0) {
    return <p className="mt-2 text-xs text-leise">{texte.keine}</p>
  }

  return (
    <div className="mt-2">
      <ul className="flex flex-col gap-2">
        {auskunft.hinweise.map((einer, stelle) => (
          <li
            className="rounded-sm border border-kante p-2 text-xs"
            key={`${einer.datum}-${String(stelle)}`}
          >
            <p className="font-semibold">
              {einer.offiziell ? texte.offiziell : texte.anmerkung} ·{' '}
              {einer.datum}
            </p>
            <p className="mt-1">{einer.text}</p>
          </li>
        ))}
      </ul>

      <p className="mt-2 text-xs text-leise">
        {texte.anmerkungErklaerung}{' '}
        <a
          className="underline focus-visible:outline-2"
          href={auskunft.adresse}
          rel="noreferrer"
          target="_blank"
        >
          {texte.quelle}
        </a>
      </p>

      {auskunft.alterStunden >= 1 && (
        <p className="mt-1 text-xs text-leise">
          {einsetzen(texte.alter, {
            stunden: String(Math.floor(auskunft.alterStunden)),
          })}
        </p>
      )}
    </div>
  )
}
