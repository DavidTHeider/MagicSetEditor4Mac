// **Stil wechseln, Stiloptionen setzen und zurücksetzen** — eine Gruppe des
// Vertragsrundgangs (B01, #2220).
//
// Das Saatset steht auf `probe-kunst`, das **keine** Stiloptionen führt; die
// Saat nennt sie nur an `probe-schlicht` (`saat/LIESMICH.md`). Deshalb wechselt
// diese Gruppe dorthin, setzt eine Wahloption, nimmt sie zurück und **wechselt
// am Ende auf den Stil zurück, auf dem das Set stand** — die übrigen Gruppen
// zeichnen dann wieder mit dem Kunstfenster (`kunst`), das `BildEinsetzen`
// braucht.
//
// Braucht `k.set` aus der ersten Gruppe.

// wahl nimmt den Stil mit Optionen, wenn die Saat ihn kennt, sonst irgendeinen
// anderen des eigenen Spiels — ein Wechsel auf den Stil, auf dem das Set schon
// steht, prüfte nichts.
const wahl = (k) => k.aStile.includes('probe-schlicht')
  ? 'probe-schlicht'
  : (k.aStile.find((s) => s !== k.set.stil && s.startsWith(`${k.set.spiel}-`)) ?? k.aStile[0])

export const aStil = {
  name: 'Stil und Stiloptionen (A)',
  rufe: [
    { name: 'Stile', argumente: (k) => [k.set.id], nimm: 'a.stile.map((s) => s.paket)', nach: (w, k) => { k.aStile = w } },
    { name: 'StilWechseln', argumente: (k) => [k.set.id, wahl(k)] },
    {
      // **Eine Wahloption mit mindestens zwei Wahlen**: `StiloptionSetzen` prüft
      // gegen das Paket, und eine Wahl, die schon gilt, änderte nichts.
      name: 'Stiloptionen', argumente: (k) => [k.set.id],
      nimm: "a.optionen.filter((o) => o.art === 'choice' && o.moeglichkeiten.length > 1).map((o) => ({ name: o.name, wert: o.moeglichkeiten.at(-1).wert }))",
      nach: (w, k) => { k.aOption = w[0] },
    },
    { name: 'StiloptionSetzen', argumente: (k) => [k.set.id, k.aOption.name, k.aOption.wert] },
    { name: 'StiloptionZuruecksetzen', argumente: (k) => [k.set.id, k.aOption.name] },
    // **Der Stil, wie ihn das Set vorher trug** — in der Schreibweise, in der
    // `Sets` ihn nannte; `stilpaketnamen` findet beide Formen.
    { name: 'StilWechseln', argumente: (k) => [k.set.id, k.set.stil] },
  ],
}
