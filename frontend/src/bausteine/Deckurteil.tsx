// **Prüfung und Abgleich eines Decks** — Werkstück 7 von Phase 4.6
// (#1293).
//
// # Die eine Zusage, die diese Fläche halten muss
//
// **`ungeprüft` ist kein `nicht zulässig`** (Vermerk 0061). Eine selbst
// entworfene Karte ist in keinem Format der Welt gelistet; sie rot zu färben
// hieße, jedem eigenen Deck zu sagen, es sei kaputt. Deshalb steht die Zahl
// der ungeprüften Karten **neben** dem Urteil und nicht darin — und der Satz
// daneben sagt, was sie heißt.
import { einsetzen } from '../sprache/einsetzen'
import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'

import type { Deckstand } from '../haken/useDecks'
import type { Deckpruefung } from '../haken/useDeckpruefung'
import type { Ladezustand } from '../haken/useLaden'
import type { DeckDTO, HabenstandAntwort } from '../apidecks'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

export function Deckurteil({
  stand,
  deck,
  pruefung,
  abgleich,
}: {
  stand: Deckstand
  deck: DeckDTO
  pruefung: Deckpruefung

  // abgleich ist, was vom Deck in der Sammlung liegt — **gelesen von der
  // Seite**, nicht erst auf Knopfdruck (AP-9, `useDeckabgleich`): die Leiste
  // über dem Deck führt hierher und nennt dieselbe Zahl.
  abgleich: Ladezustand<HabenstandAntwort | null>
}) {
  const texte = useTexte()
  const eigene = texte.seiten.decks

  const bestand = abgleich.daten ?? null
  const woerter = new Map([
    ['legal', eigene.legal],
    ['not_legal', eigene.not_legal],
    ['restricted', eigene.restricted],
    ['banned', eigene.banned],
    ['ungeprüft', eigene.ungeprüft],
  ])

  const urteil = pruefung.urteil
  const befunde = pruefung.befunde

  return (
    <div className="flex flex-col gap-3">
      {pruefung.laedt && (
        <p role="status" className="text-sm text-leise">
          {eigene.pruefungLaeuft}
        </p>
      )}

      {pruefung.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {pruefung.fehler}
        </p>
      )}

      {/* **Ohne Format steht der Satz und nicht der Fehler.** Die Prüfung
          fragt dann gar nicht (siehe [useDeckpruefung]); ein Knopf, der eine
          Frage stellt, auf die es keine Antwort gibt, ruht. */}
      {pruefung.ohneFormat && <p>{eigene.pruefungOhneFormat}</p>}

      <div className="flex flex-wrap gap-3">
        <Knopf
          angehalten={
            stand.vorgang.laeuft(`deck:${deck.id}`) ||
            pruefung.laedt ||
            pruefung.ohneFormat
          }
          aufKlick={() => {
            pruefung.nachladen()
          }}
        >
          {eigene.pruefen}
        </Knopf>

        {/* **Neu fragen** — die Sammlung kann sich geändert haben, ohne dass
            das Deck es tut. */}
        <Knopf
          angehalten={abgleich.laedt}
          aufKlick={() => {
            abgleich.nachladen()
          }}
        >
          {eigene.bestand}
        </Knopf>
      </div>

      {abgleich.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {abgleich.fehler}
        </p>
      )}

      {urteil !== null && (
        <div className="flex flex-col gap-2">
          <p className="font-medium">
            {urteil.zulaessig ? eigene.zulaessig : eigene.unzulaessig}
          </p>

          {urteil.ungeprueft > 0 && (
            <>
              <p>
                {einsetzen(eigene.ungeprueftZahl, {
                  zahl: String(urteil.ungeprueft),
                })}
              </p>

              <Erklaerungsabsatz
                text={eigene.ungepruefterklaerung}
                klasse="text-sm text-leise"
              />
            </>
          )}

          <ul className="flex flex-col gap-1">
            {urteil.groessen.map((groesse) => (
              <li key={groesse.meldung} role="alert">
                {groesse.meldung}
              </li>
            ))}
          </ul>

          <ul className="flex flex-col gap-1">
            {/* **Mit Grund und Quelle** — das verlangt die Abnahme des
                Issues ausdrücklich, und Befund 7 des Zweitblicks über Phase
                4.6 hat angestrichen, dass das Fenster beide verschwieg,
                obwohl die Naht sie trägt. */}
            {urteil.urteile.map((eines) => (
              <li key={eines.platzID} className="flex flex-wrap gap-3">
                <span className="tabular-nums">{eines.menge}</span>
                <span>{eines.bezeichnung}</span>
                <span>{wort(woerter, eines.zulaessigkeit)}</span>

                {eines.quelle !== '' && (
                  <span className="text-sm text-leise">{eines.quelle}</span>
                )}

                {eines.grund !== '' && (
                  <span className="text-sm text-leise">{eines.grund}</span>
                )}
              </li>
            ))}
          </ul>

          <ul className="flex flex-col gap-1">
            {befunde.map((befund) => (
              <li key={befund} className="text-sm">
                {befund}
              </li>
            ))}
          </ul>
        </div>
      )}

      {bestand !== null && (
        <div className="flex flex-col gap-2">
          <p className="font-medium">
            {bestand.fehlend === 0
              ? eigene.bestandVollstaendig
              : einsetzen(eigene.bestandFehlt, {
                  zahl: String(bestand.fehlend),
                })}
          </p>

          <ul className="flex flex-col gap-1">
            {bestand.zeilen.map((zeile) => (
              <li key={zeile.printing + zeile.grund} className="flex gap-3">
                <span className="tabular-nums">
                  {zeile.vorhanden}/{zeile.gebraucht}
                </span>
                <span>
                  {zeile.grund === '' ? zeile.bezeichnung : zeile.grund}
                </span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}

// wort übersetzt eines der fünf Zulässigkeitswörter.
//
// **Über eine Abbildung und nicht über einen Index**: Scryfall darf ein
// sechstes schicken, und dann steht es unübersetzt da statt leer (§9).
// **Und über die fünf, nicht über den ganzen Katalog** — sonst träfe ein
// Wort wie `leiste` einen Abschnitt des Katalogs statt eines Satzes.
function wort(
  woerter: ReadonlyMap<string, string>,
  zulaessigkeit: string,
): string {
  return woerter.get(zulaessigkeit) ?? zulaessigkeit
}
