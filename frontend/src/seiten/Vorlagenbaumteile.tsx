// **Teile des Vorlagenbaums** — der Filter, das Angebot einer eigenen Kopie
// und „Weitere zeigen". Eine eigene Datei, weil `Vorlagenbaum.tsx` mit dem
// Beiwerk (FMT-02, #1691) über die Zeilengrenze aus §2 ging.
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { Paketkopie } from '../haken/usePaketkopie'
import type { Zweig } from '../haken/useVorlagenbaum'
import type { Zeilengriffe } from './Vorlagenzweig'

// knopfklassen tragen alle Knöpfe des Baums.
export const knopfklassen =
  'w-full rounded-sm px-2 py-1 text-left text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

// passt vergleicht ohne Groß- und Kleinschreibung.
function passt(name: string, anfrage: string): boolean {
  return name.toLowerCase().includes(anfrage.toLowerCase())
}

// sichtbar gibt den Zweig mit den Dateien, die zur Anfrage passen — oder
// `null`, wenn nichts passt.
export function sichtbar(zweig: Zweig, anfrage: string): Zweig | null {
  if (anfrage === '' || passt(zweig.name, anfrage)) {
    return zweig
  }

  const dateien = zweig.dateien.filter((datei) => passt(datei, anfrage))
  const beiwerk = zweig.beiwerk.filter((datei) => passt(datei, anfrage))

  return dateien.length + beiwerk.length === 0
    ? null
    : { ...zweig, dateien, beiwerk }
}

// Kopieangebot steht unter einem Paket, das sich nur lesen lässt: der Grund
// und der Weg zur eigenen Kopie — samt ihrem Fehler, am Paket selbst.
export function Kopieangebot({
  paket,
  kopie,
}: {
  paket: string
  kopie: Pick<Paketkopie, 'kopieren' | 'laeuft' | 'fehler'>
}) {
  const texte = useTexte().seiten.vorlagen
  const fehler = kopie.fehler(paket)

  return (
    <div className="flex flex-col items-start gap-1 pl-6">
      <p className="text-sm text-leise">{texte.nurLesbar}</p>
      <Aktionsknopf
        aktion="kopieren"
        angehalten={kopie.laeuft(paket)}
        aufKlick={() => void kopie.kopieren(paket)}
      >
        {texte.kopieren}
      </Aktionsknopf>
      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}
    </div>
  )
}

// Weiterknopf öffnet das nächste Stück einer langen Liste (UXR-18, #1926) —
// und sagt, wie viel davon schon dasteht. Er ist ein Knopf des Baums wie
// jeder andere: die Pfeile erreichen ihn, die Tabstelle kann auf ihm liegen.
export function Weiterknopf({
  gezeigt,
  gesamt,
  griffe,
  aufKlick,
}: {
  gezeigt: number
  gesamt: number
  griffe: Zeilengriffe
  aufKlick: () => void
}) {
  const texte = useTexte().seiten.vorlagen

  return (
    <li>
      <button
        type="button"
        {...griffe}
        onClick={aufKlick}
        className={[knopfklassen, 'text-xs text-leise'].join(' ')}
      >
        {einsetzen(texte.weitereZeigen, {
          gezeigt: String(gezeigt),
          gesamt: String(gesamt),
        })}
      </button>
    </li>
  )
}
