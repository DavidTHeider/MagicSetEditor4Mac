// **Was neben einem übernommenen Bild stehen muss** —
// #1025.
//
// Scryfall verlangt, nachgelesen am 2026-08-23 unter
// `https://scryfall.com/docs/api`:
//
// > When using the art_crop, list the artist name and copyright elsewhere in
// > the same interface presenting the art crop […] Users should be able to
// > identify the artist and source of the image somehow.
//
// **Diese Fläche ist die Oberfläche, die den Ausschnitt zeigt** — die
// Kartenvorschau, und zwar auch dann noch, wenn der Zuschnittkasten längst zu
// ist. Bis hierher stand die Nennung nur an ihm; eine Bedingung, die das
// Schließen eines Kastens nicht überlebt, ist keine.
//
// **Der Satz kommt fertig aus Go** (`service.Nennungen`) und wird hier nicht
// gebaut: dort steht, woran eine Kartenseite bei Scryfall erkannt wird, und
// dort steht das Copyright als Konstante. Eine zweite Fassung liefe
// auseinander, sobald jemand eine davon nachzöge.
import type { NennungDTO } from '../formen'
import { useTexte } from '../sprache/kontext'

export function Nennung({ nennungen }: { nennungen: NennungDTO[] }) {
  const texte = useTexte()

  if (nennungen.length === 0) {
    return null
  }

  return (
    <ul className="mt-2 space-y-1 text-xs text-leise">
      {nennungen.map((nennung) => (
        <li key={nennung.adresse}>
          {nennung.text}{' '}
          <a
            className="underline focus-visible:outline-2"
            href={nennung.adresse}
            rel="noreferrer"
            target="_blank"
          >
            {texte.seiten.karten.scryfall.quelle}
          </a>
        </li>
      ))}
    </ul>
  )
}
