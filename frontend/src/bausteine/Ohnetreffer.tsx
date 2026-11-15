// **Eine Suche ohne Treffer sagt es und nimmt sich zurück** — G02
// (#2249).
//
// Der Satz allein riet, das Feld zu leeren, und ließ es stehen; wer auf eine
// leere Fläche schaut, sucht den Weg zurück nicht oben im Suchfeld.
import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'

export function Ohnetreffer({
  satz,
  leeren,
}: {
  satz: string
  leeren: () => void
}) {
  const texte = useTexte()

  return (
    <div className="flex max-w-prose flex-col items-start gap-2 text-sm">
      <p className="text-leise">{satz}</p>
      <Knopf aufKlick={leeren}>
        {texte.bausteine.listenansicht.sucheLeeren}
      </Knopf>
    </div>
  )
}
