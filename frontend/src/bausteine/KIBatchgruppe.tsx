// **Die Lagen eines Batches und die atomare Übernahme** — Werkstück A13.1b
// von Phase 5.3
// (#1452).
//
// # Ein eigenes Bauteil
//
// **Weil `KIBatchkasten` sonst über die Zeilengrenze aus §2 ginge** — und
// weil der Schnitt der Frage folgt: dort die Bedienung des Batches, hier
// **was daraus wurde** und was davon geschrieben wird.
//
// # Warum ein Knopf und nicht einer je Zeile
//
// **Weil das die Zusage des Issues ist**: „selektive Übernahme als definierte
// atomare **Gruppe**". Ein Knopf je Zeile wäre ein Ruf je Zeile — und damit
// eine Transaktion je Karte, mit genau dem Teilbestand, den A13.1b
// ausschließt. Was ausgewählt wird, steht an den Häkchen; **wann** es
// geschrieben wird, entscheidet der eine Knopf.
//
// # Und warum gescheiterte Ziele keinen Haken haben
//
// **Weil es nichts zu übernehmen gibt.** Sie stehen trotzdem da, mit ihrem
// Grund: ein Batch, der nur die geglückten zeigt, verschweigt, welche Karte
// nichts bekommen hat.
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

import type { KIBatchAntwort, KIGruppenwahl } from '../antwortenbatch'

// marke ist der Schlüssel eines Häkchens — Ziel und Feld.
//
// **Ein Schrägstrich genügt**: eine Auftragskennung trägt keinen, und ein
// Feldname ebenso wenig (`domain.PruefeFeldname`).
function marke(kennung: string, feld: string) {
  return `${kennung}/${feld}`
}

interface KIBatchgruppeEigenschaften {
  batch: KIBatchAntwort

  // geaendert läuft nach einer geglückten Gruppenübernahme.
  geaendert: () => void
}

export function KIBatchgruppe({
  batch,
  geaendert,
}: KIBatchgruppeEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen

  const vorgang = useVorgang()

  // **Zu Anfang ist alles angehakt** — dieselbe Überlegung wie bei der
  // Vorschlagsliste: abhaken geht schneller als anhaken, und geschrieben wird
  // ohnehin erst auf Klick.
  //
  // # Warum eine flache Menge und keine Abbildung je Ziel
  //
  // **Weil eine Abbildung einen Zweig erzeugt, den niemand betreten kann.**
  // Der erste Anlauf hielt `Record<string, string[]>` und griff mit
  // `bisher[kennung] ?? []` zu — der Rückfall lag an zwei Stellen und war
  // unerreichbar, weil nur Ziele umgeschaltet werden, die auch darin stehen.
  // Ein Zweig ohne Fall ist kein Schutz, sondern eine Lücke in der Abdeckung
  // (§3).
  const [gewaehlt, setzeGewaehlt] = useState<Set<string>>(
    () =>
      new Set(
        batch.ziele
          .filter((lage) => lage.fehler === null)
          .flatMap((lage) =>
            lage.felder.map((feld) => marke(lage.kennung, feld.feld)),
          ),
      ),
  )

  const [fehler, setzeFehler] = useState<string>()
  const [geschrieben, setzeGeschrieben] = useState<number>()

  // derLauf ist der Schlüssel, unter dem [useVorgang] mitzählt — **derselbe**
  // beim Starten und beim Fragen. Zwei verschiedene hießen: der laufende
  // Zustand ist nie sichtbar.
  const derLauf = `gruppe/${batch.kennung}`

  function umschalten(kennung: string, feld: string) {
    setzeGewaehlt((bisher) => {
      const jetzt = new Set(bisher)

      if (jetzt.has(marke(kennung, feld))) {
        jetzt.delete(marke(kennung, feld))
      } else {
        jetzt.add(marke(kennung, feld))
      }

      return jetzt
    })
  }

  // wahl sind die Ziele, von denen wirklich etwas übernommen wird.
  //
  // **Ein Ziel ohne Haken fällt heraus** und wird nicht als leere Auswahl
  // mitgeschickt: leer ist in `service` ein Fehler, und der brächte die ganze
  // Gruppe zu Fall, weil jemand ein Häkchen weggenommen hat.
  const wahl: KIGruppenwahl[] = batch.ziele
    .map((lage) => ({
      kennung: lage.kennung,
      felder: lage.felder
        .map((feld) => feld.feld)
        .filter((feld) => gewaehlt.has(marke(lage.kennung, feld))),
    }))
    .filter((eine) => eine.felder.length > 0)

  async function uebernehmen() {
    setzeFehler(undefined)
    setzeGeschrieben(undefined)

    await vorgang.starte(derLauf, async () => {
      try {
        const antwort = await api().KIGruppeUebernehmen(wahl)
        const bericht = auspacken(antwort, () => antwort)

        setzeGeschrieben(bericht.karten)

        // **Erst melden, dann nachladen** (§7, Regel 3): geschrieben ist
        // geschrieben, und ein Lesefehler danach nimmt es nicht zurück.
        geaendert()
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  return (
    <div className="flex flex-col gap-2 border-t border-kante pt-2">
      <h5 className="text-sm font-medium">{eigene.kiBatchLagen}</h5>

      <ul className="flex flex-col gap-2">
        {batch.ziele.map((lage) => (
          <li key={lage.zielkarte} className="text-sm">
            <span className="font-medium">{lage.zielkarte}</span>{' '}
            {lage.fehler === null ? lage.modell : lage.fehler.meldung}
            {lage.fehler === null && (
              <ul className="mt-1 flex flex-col gap-0.5 pl-4">
                {lage.felder.map((feld) => (
                  <li key={feld.feld}>
                    <label className="flex items-center gap-2">
                      <input
                        type="checkbox"
                        checked={gewaehlt.has(marke(lage.kennung, feld.feld))}
                        onChange={() => {
                          umschalten(lage.kennung, feld.feld)
                        }}
                      />
                      <span className="font-medium">{feld.feld}</span>
                      <span className="text-leise">
                        {feld.werte.join(', ')}
                      </span>
                    </label>
                  </li>
                ))}
              </ul>
            )}
            {lage.befunde.length > 0 && (
              <ul className="mt-0.5 list-disc pl-8 text-leise">
                {lage.befunde.map((befund) => (
                  <li key={befund}>{befund}</li>
                ))}
              </ul>
            )}
          </li>
        ))}
      </ul>

      <p className="max-w-prose text-sm text-leise">
        {eigene.kiBatchGanzOderGar}
      </p>

      <button
        type="button"
        className="self-start rounded-sm bg-marke px-3 py-1 text-sm text-marketext focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
        disabled={wahl.length === 0 || vorgang.laeuft(derLauf)}
        onClick={() => {
          void uebernehmen()
        }}
      >
        {vorgang.laeuft(derLauf)
          ? eigene.kiBatchGruppeLaeuft
          : einsetzen(eigene.kiBatchGruppe, { karten: String(wahl.length) })}
      </button>

      {geschrieben !== undefined && (
        <p role="status" className="text-sm text-leise">
          {einsetzen(eigene.kiBatchGeschrieben, {
            karten: String(geschrieben),
          })}
        </p>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}
    </div>
  )
}
