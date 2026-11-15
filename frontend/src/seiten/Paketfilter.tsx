// Der Filterbalken über der Paketliste (#244).
//
// **Was er tut, steht hier; wie er aussieht, in den Klassen** (§2).
//
// # Vier Steuerelemente
//
// Suchen, nach Art filtern, ordnen. Jedes beantwortet eine eigene Frage — *wo
// ist X*, *zeig mir nur Y*, *was ist alt* —, und keine davon lässt sich mit
// einem der anderen stellen.
//
// **Das vierte kam mit DOM-01** (#1696): die **Herkunft** — mitgeliefert,
// bezogen, übernommen, eigene. Es beantwortet eine Frage, die keines der
// drei stellt: *was habe ich selbst hereingeholt*. Das Artboard `Pakete`
// zeichnet sie als Knöpfe, und so standen sie hier bis UXR-17 (#1925).
// **Jetzt ein Auswahlfeld wie die Art**: mit „jede" sind es fünf Werte, und
// R7 aus `docs/UIREVIEW.md` gibt Segmente bis vier — fünf Knöpfe zählten
// zudem im Bild mit (R2). Die Abweichung steht in `PROTOTYP-0.6.0.md`.
//
// **Untereinander und nicht nebeneinander**: der Filter steht im Navigator,
// und der ist 240 Punkte breit.
import type { Paketherkunftsart } from '../api'
import type { Ordnung, Sichtwahl } from './paketsicht'
import { Auswahlfilter } from '../bausteine/Auswahlfilter'
import { useTexte } from '../sprache/kontext'

interface Eigenschaften {
  wahl: Sichtwahl

  // arten sind die Arten, die im Bestand vorkommen — nicht die sieben
  // möglichen (der Grund steht bei [arten]).
  arten: readonly string[]

  aendern: (wahl: Sichtwahl) => void
}

const eingabe =
  'rounded-md border border-kante bg-flaeche px-2 py-1 text-text ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring '

export function Paketfilter({ wahl, arten, aendern }: Eigenschaften) {
  const texte = useTexte()

  // **Die vier Ordnungen stehen hier als Paare** — Wert und Aufschrift. Der
  // Wert ist Vokabular des Programms und wird nicht übersetzt; die Aufschrift
  // kommt aus dem Katalog (§9).
  const ordnungen: readonly [Ordnung, string][] = [
    ['gruppen', texte.seiten.pakete.katalogOrdnungGruppen],
    ['name', texte.seiten.pakete.katalogOrdnungName],
    ['version', texte.seiten.pakete.katalogOrdnungVersion],
    ['herkunft', texte.seiten.pakete.katalogOrdnungHerkunft],
  ]

  // **Die Herkunft ebenso** — der Wert ist Vokabular des Programms
  // (`domain.Paketherkunftsart`), die Aufschrift kommt aus dem Katalog.
  const herkunftsarten = texte.seiten.pakete.plan.herkunftsarten
  const herkuenfte: readonly ['' | Paketherkunftsart, string][] = [
    ['', texte.seiten.pakete.plan.herkunftAlle],
    ['mitgeliefert', herkunftsarten.mitgeliefert],
    ['bezogen', herkunftsarten.bezogen],
    ['uebernommen', herkunftsarten.uebernommen],
    ['eigen', herkunftsarten.eigen],
  ]

  return (
    <div className="flex flex-col gap-3">
      {/* **Die Auswahl nennt sich und lässt sich lösen** — AP-8 (#2365): wer
          aus der Aufmerksamkeit kommt, sieht, warum nur diese Pakete
          dastehen. */}
      {wahl.auswahl !== undefined && (
        <Auswahlfilter
          name={wahl.auswahl.name}
          loesen={texte.seiten.pakete.auswahlLoesen}
          aufLoesen={() => {
            aendern({ ...wahl, auswahl: undefined })
          }}
        />
      )}

      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium">
          {texte.seiten.pakete.katalogSuche}
        </span>

        {/* **`type="search"` und nicht `text`.** Das Feld bekommt damit unter
            macOS das Löschkreuz und die gewohnte Behandlung durch
            Hilfstechnik; §6 verlangt, macOS-Eigenheiten zu bedienen. */}
        <input
          type="search"
          value={wahl.suche}
          placeholder={texte.seiten.pakete.katalogSucheHinweis}
          onChange={(ereignis) => {
            aendern({ ...wahl, suche: ereignis.target.value })
          }}
          className={eingabe}
        />
      </label>

      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium">
          {texte.seiten.pakete.katalogArt}
        </span>

        <select
          value={wahl.art}
          onChange={(ereignis) => {
            aendern({ ...wahl, art: ereignis.target.value })
          }}
          className={eingabe}
        >
          <option value="">{texte.seiten.pakete.katalogArtAlle}</option>

          {arten.map((art) => (
            <option key={art} value={art}>
              {art}
            </option>
          ))}
        </select>
      </label>

      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium">
          {texte.seiten.pakete.plan.herkunft}
        </span>

        <select
          value={wahl.herkunft}
          onChange={(ereignis) => {
            // **Die Zusicherung steht hier und nicht am Zustand** — dieselbe
            // wie bei der Ordnung: die fünf Werte stehen darüber, und jede
            // andere Zeichenkette käme nur aus einem Fehler in dieser Datei.
            aendern({
              ...wahl,
              herkunft: ereignis.target.value as '' | Paketherkunftsart,
            })
          }}
          className={eingabe}
        >
          {herkuenfte.map(([wert, aufschrift]) => (
            <option key={wert} value={wert}>
              {aufschrift}
            </option>
          ))}
        </select>
      </label>

      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium">
          {texte.seiten.pakete.katalogOrdnung}
        </span>

        <select
          value={wahl.ordnung}
          onChange={(ereignis) => {
            // **Die Zusicherung steht hier und nicht am Zustand.** Ein
            // `<select>` gibt `string`; die vier möglichen Werte stehen zehn
            // Zeilen darüber, und jede andere Zeichenkette käme nur aus einem
            // Fehler in genau dieser Datei.
            aendern({ ...wahl, ordnung: ereignis.target.value as Ordnung })
          }}
          className={eingabe}
        >
          {ordnungen.map(([wert, aufschrift]) => (
            <option key={wert} value={wert}>
              {aufschrift}
            </option>
          ))}
        </select>
      </label>
    </div>
  )
}
