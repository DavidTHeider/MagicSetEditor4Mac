// **Eine Auswahl, die sich nennt und sich lösen lässt** — AP-8
// (#2365).
// Wer aus dem Überblick in eine Liste kommt, sieht genau die Objekte, die die
// Zahl meinte — und warum nur sie: der Satz steht über der Liste, ein Knopf
// gibt alle zurück. Paketliste und Setliste teilen ihn.
import { Knopf } from './Knopf'

export function Auswahlfilter({
  name,
  loesen,
  aufLoesen,
}: {
  // name ist der Satz, der die Auswahl nennt — „1 Set mit Problemen“.
  name: string
  loesen: string
  aufLoesen: () => void
}) {
  return (
    <div className="flex flex-col items-start gap-1">
      <p role="status" className="text-sm font-medium">
        {name}
      </p>
      <Knopf aufKlick={aufLoesen}>{loesen}</Knopf>
    </div>
  )
}
