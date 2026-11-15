// **Was vom Stand eines Handels vorn steht** — UXR-15
// (#1923),
// herausgelöst aus `Handelsvorgang` mit P13 (#2381), als der Kasten an die
// Zeilengrenze kam (§2).
//
// **Der Stand wird angesagt** (§6): er ändert sich auf einen Klick hin, und
// wer die Seite nicht sieht, erführe es sonst nicht.
import { Satz } from './Satz'
import { useTexte } from '../sprache/kontext'
import { artsatz, standsatz } from './handelsstaende'

import type { Handelsvorgang as Vorgangsstand } from '../apihandelsleitung'

export function Handelsstandanzeige({
  stand,
  nenne,
}: {
  stand: Vorgangsstand

  // nenne sagt, wie eine Partei heißt — „du", das Gegenüber beim Namen.
  nenne: (partei: string) => string
}) {
  const eigene = useTexte().seiten.handeln

  return (
    <div role="status" className="flex flex-col gap-1 text-sm text-leise">
      <p>
        {eigene.buchungZustand}: {standsatz(eigene, stand.zustand)}
      </p>

      <p>
        {eigene.vorgangArt}: {artsatz(eigene, stand.art)}
      </p>

      {/* **Nur ein Tausch bewegt Bestand** (TRADE-03.AK-02, #1717) —
          und eine Einigung ist eine Vereinbarung, kein Wechsel. */}
      {stand.art !== 'tausch' && <p>{eigene.artOhneBestand}</p>}

      {/* **In einem Absatz** — im Spaltenbehälter darüber würde jedes
          Textstück des Satzes zu einer eigenen Zeile. */}
      {stand.zustand === 'angenommen' && (
        <p>
          <Satz text={eigene.vorgangVereinbarung} />
        </p>
      )}

      {/* **Die Stimmen gehören zu einem offenen Beschluss** — ohne ihn
          stünde „Angenommen" neben „Zugestimmt: niemand". */}
      {stand.beschluss !== '' && (
        <p>
          {eigene.vorgangZustimmungen}:{' '}
          {stand.zustimmungen.length === 0
            ? eigene.vorgangKeine
            : stand.zustimmungen.map(nenne).join(', ')}
        </p>
      )}

      {stand.ende && <p>{eigene.vorgangEnde}</p>}
    </div>
  )
}
