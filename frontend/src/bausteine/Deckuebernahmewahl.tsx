// **Aus welchem Profil, welches Deck — gewählt** — UX-08.AK-04
// (#1680):
// *„Normalnutzende müssen keine Netzwerk- oder Transaktionskennungen
// eingeben, soweit diese technisch ableitbar sind."*
//
// # Was vorher war
//
// **Zwei Felder für zwei Kennungen**, und die des Decks stand im eigenen
// Profil nirgends. Der Grund war ehrlich: die Decks von drüben zu zeigen
// hieße, über die Profilgrenze zu lesen. **Er fällt mit
// `service.Deckuebernahme.DeckeZurUebernahme`**: die Liste steht hinter
// derselben Erlaubnis wie die Übernahme, geprüft, bevor gelesen wird
// (Vermerk 0071) — wer nicht hinein darf, sieht keine Decks, sondern das
// Nein.
//
// # Warum das eigene Profil fehlt
//
// **Weil aus ihm nichts übernommen wird** — `DeckUebernehmen` weist es ab.
// Eine Wahl, die zu einer Absage führt, ist keine.
import { useId } from 'react'

import { Ladefehler, Wirdgeladen } from './Ladehinweise'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'

import type { DeckDTO } from '../apidecks'
import type { Ladezustand } from '../haken/useLaden'

// wahlklassen ist eine Auswahlliste — **mit sichtbarem Fokusring** (§6) und
// dem Grund der Fläche, damit sie im dunklen Erscheinungsbild lesbar bleibt.
const wahlklassen =
  'rounded-sm border border-kante bg-flaeche px-2 py-1 text-text ' +
  'focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

interface Deckuebernahmewahleigenschaften {
  profil: string
  deck: string
  setzeProfil: (profil: string) => void
  setzeDeck: (deck: string) => void
}

export function Deckuebernahmewahl({
  profil,
  deck,
  setzeProfil,
  setzeDeck,
}: Deckuebernahmewahleigenschaften) {
  const eigene = useTexte().seiten.decks
  const profilfeld = useId()
  const deckfeld = useId()

  const profile = useLaden(async () => {
    const [nutzer, gezeigt] = await Promise.all([
      api().Nutzer(),
      api().GezeigtesProfil(),
    ])
    const eigenes = auspacken(gezeigt, () => gezeigt.profil)?.id

    return auspacken(nutzer, () => nutzer.nutzer).flatMap((mensch) =>
      mensch.profile
        .filter((eines) => eines.id !== eigenes)
        .map((eines) => ({
          id: eines.id,
          beschriftung: einsetzen(eigene.uebernahmeProfilEintrag, {
            nutzer: mensch.name,
            profil: eines.name,
          }),
        })),
    )
  }, [])

  // **Erst mit einem Profil wird gefragt** — und die Antwort kann ein Nein
  // sein: dann steht es da, wo sonst die Decks stünden.
  const decks = useLaden(async () => {
    if (profil === '') {
      return []
    }

    const antwort = await api().DeckeZurUebernahme(profil)

    return auspacken(antwort, () => antwort.decks)
  }, [profil])

  if (profile.daten === undefined) {
    return profile.fehler === undefined ? (
      <Wirdgeladen />
    ) : (
      <Ladefehler fehler={profile.fehler} wiederholen={profile.nachladen} />
    )
  }

  if (profile.daten.length === 0) {
    return <p className="text-sm text-leise">{eigene.uebernahmeKeinProfil}</p>
  }

  return (
    <div className="flex flex-wrap items-end gap-3">
      <label className="flex flex-col gap-1" htmlFor={profilfeld}>
        <span className="text-sm font-medium">
          {eigene.uebernahmeProfilwahl}
        </span>
        <select
          id={profilfeld}
          value={profil}
          onChange={(ereignis) => {
            setzeProfil(ereignis.target.value)
            setzeDeck('')
          }}
          className={wahlklassen}
        >
          <option value="">{eigene.uebernahmeProfilWaehlen}</option>
          {profile.daten.map((eines) => (
            <option key={eines.id} value={eines.id}>
              {eines.beschriftung}
            </option>
          ))}
        </select>
      </label>

      {profil !== '' && (
        <Deckliste
          decks={decks}
          deck={deck}
          setzeDeck={setzeDeck}
          feld={deckfeld}
        />
      )}
    </div>
  )
}

// Deckliste ist die zweite Wahl — mit ihren drei Zuständen.
function Deckliste({
  decks,
  deck,
  setzeDeck,
  feld,
}: {
  decks: Ladezustand<DeckDTO[]>
  deck: string
  setzeDeck: (deck: string) => void
  feld: string
}) {
  const eigene = useTexte().seiten.decks

  if (decks.daten === undefined) {
    return decks.fehler === undefined ? (
      <Wirdgeladen />
    ) : (
      <Ladefehler fehler={decks.fehler} wiederholen={decks.nachladen} />
    )
  }

  if (decks.daten.length === 0) {
    return <p className="text-sm text-leise">{eigene.uebernahmeKeinDeck}</p>
  }

  return (
    <label className="flex flex-col gap-1" htmlFor={feld}>
      <span className="text-sm font-medium">{eigene.uebernahmeDeckwahl}</span>
      <select
        id={feld}
        value={deck}
        onChange={(ereignis) => {
          setzeDeck(ereignis.target.value)
        }}
        className={wahlklassen}
      >
        <option value="">{eigene.uebernahmeDeckWaehlen}</option>
        {decks.daten.map((eines) => (
          <option key={eines.id} value={eines.id}>
            {eines.name}
          </option>
        ))}
      </select>
    </label>
  )
}
