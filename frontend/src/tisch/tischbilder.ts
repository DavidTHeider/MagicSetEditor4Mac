// **Die Tischbilder** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Sechs Sätze, und keiner stammt von woanders
//
// **#1460 verlangt „neutrale und rechtlich verwendbare Hintergründe" und
// nennt sechs Gestaltungsvorschläge** — Holz, dunkles Holz, Picknick,
// Dungeon, Taverne, Turniertisch. Hier sind es **Farbsätze** und keine
// Bilddateien: was selbst gewählt ist, kann niemandem gehören (§11), und drei
// Zahlen je Satz wiegen nichts.
//
// **Wer ein Bild will, bringt eines mit.** Der Weg dafür steht daneben; was
// er zeigt, ist dann seine Sache.
//
// # Warum die Kante nicht mitwandert
//
// **Weil sie die Zugänglichkeit trägt.** Zwei Holztöne nebeneinander halten
// keine 3:1 — die Linie tut es, und sie hält gegen **jeden** der sechs Sätze
// (`tischbilder.test.ts` rechnet es nach, statt dass jemand es aufschreibt).

// Tischbildname ist einer der sechs.
export type Tischbildname =
  'holz' | 'dunkles-holz' | 'picknick' | 'dungeon' | 'taverne' | 'turniertisch'

// Tischbild sind die drei Flächenfarben eines Satzes, je Erscheinungsbild.
export interface Tischbild {
  name: Tischbildname
  hell: { grund: number; sitz: number; eigen: number }
  dunkel: { grund: number; sitz: number; eigen: number }
}

// tischbilder sind die sechs.
//
// # Die Namen sagen den Ton, nicht die Helligkeit
//
// **Im hellen Erscheinungsbild ist jeder Tisch hell, im dunklen jeder
// dunkel** — sonst hielte die Kante nicht. „Dunkles Holz" ist deshalb ein
// **wärmerer, satterer** Ton als „Holz" und nicht ein dunklerer; wer es
// dunkler machte, bekäme im hellen Modus 2,8:1 gegen die Linie, und die
// Zugänglichkeit hinge an der Wahl eines Hintergrunds.
//
// **Das ist der Preis dafür, dass die Kante die Zugänglichkeit trägt**, und
// er gehört ausgesprochen: wer einen wirklich dunklen Tisch will, schaltet
// das Erscheinungsbild um.
//
// Der Lauf daneben rechnet jeden Satz nach — er ist die Zusage, nicht diese
// Zeile.
// **Als Tupel und nicht als Reihe.** `noUncheckedIndexedAccess` machte aus
// `tischbilder[0]` sonst ein `Tischbild | undefined`, und der Rückfall
// bräuchte einen Zweig, den kein Lauf je beträte — toter Code (§3). Ein
// Tupel mit mindestens einem Glied sagt dem Typprüfer, was die Liste ohnehin
// ist.
export const tischbilder: [Tischbild, ...Tischbild[]] = [
  {
    name: 'holz',
    hell: { grund: 0xd4bb94, sitz: 0xdcc4a0, eigen: 0xe4cfae },
    dunkel: { grund: 0x221b13, sitz: 0x2a2218, eigen: 0x332a1e },
  },
  {
    name: 'dunkles-holz',
    hell: { grund: 0xd9ba90, sitz: 0xe1c49e, eigen: 0xe9ceac },
    dunkel: { grund: 0x1a140e, sitz: 0x221b13, eigen: 0x2a2218 },
  },
  {
    name: 'picknick',
    hell: { grund: 0xd8b6a8, sitz: 0xe0c0b2, eigen: 0xe8cabd },
    dunkel: { grund: 0x241812, sitz: 0x2c1f18, eigen: 0x34261e },
  },
  {
    name: 'dungeon',
    hell: { grund: 0xc4c0b8, sitz: 0xccc9c2, eigen: 0xd5d2cc },
    dunkel: { grund: 0x1c1c1a, sitz: 0x242422, eigen: 0x2c2c2a },
  },
  {
    name: 'taverne',
    hell: { grund: 0xdcc0a4, sitz: 0xe3c9b0, eigen: 0xead3bd },
    dunkel: { grund: 0x201810, sitz: 0x281f16, eigen: 0x30271c },
  },
  {
    name: 'turniertisch',
    hell: { grund: 0xb8c4b0, sitz: 0xc2ccba, eigen: 0xccd5c5 },
    dunkel: { grund: 0x161c16, sitz: 0x1e241e, eigen: 0x262c26 },
  },
]

// tischbild sucht einen Satz — und fällt auf Holz zurück.
//
// **Auf Holz und nicht auf nichts**: ein gespeicherter Name, den es nicht mehr
// gibt, ließe den Tisch sonst ohne Farbe dastehen. Derselbe Fall wie eine
// unbekannte Zone (§8) — benannt wird er nicht, weil er niemandem hilft: die
// Wahl liegt danach sichtbar wieder auf Holz.
export function tischbild(name: string): Tischbild {
  return tischbilder.find((eines) => eines.name === name) ?? tischbilder[0]
}
