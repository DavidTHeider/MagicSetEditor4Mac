// **Suchen, Ersetzen, Marken, Autoersetzung, Sprachpakete** — eine Gruppe
// des Vertragsrundgangs (B01, #2220). Braucht `k.set`, `k.karten` und
// `k.textfeld` aus der ersten Gruppe.

const frage = (k, stand = '') => ({ text: 'Vertrag', ersatz: 'Kontrakt', feld: k.textfeld, grossKleinBeachten: false, ganzesWort: false, stand })

export const werkzeuge = {
  name: 'Suchen, Ersetzen, Marken, Autoersetzung, Sprachpakete',
  rufe: [
    { name: 'Suchfelder', argumente: (k) => [k.set.id] },
    { name: 'Suchen', argumente: (k) => ['Vertrag', '', k.set.id] },
    { name: 'Statistik', argumente: (k) => [k.set.id] },
    { name: 'Verteilung', argumente: (k) => [k.set.id, k.textfeld] },
    { name: 'Kartenlistenaufbau', argumente: (k) => [k.set.id] },
    { name: 'ListensortierungSetzen', argumente: (k) => [k.set.id, k.textfeld, true] },
    { name: 'ErsetzungVorschau', argumente: (k) => [k.set.id, frage(k)] },
    { name: 'ErsetzungAusfuehren', argumente: (k) => [k.set.id, frage(k)] },
    { name: 'AblageKopieren', argumente: (k) => [k.set.id, [k.karten[0]]] },
    { name: 'AblageEinfuegen', argumente: (k) => [k.set.id], nimm: 'a.karten', nach: (w, k) => { k.eingefuegt = w[0] } },
    { name: 'AblageAusschneiden', argumente: (k) => [k.set.id, [k.eingefuegt]] },
    { name: 'Autoersetzungen', argumente: (k) => [k.set.id] },
    { name: 'AutoersetzungSetzen', argumente: (k) => [k.set.id, 'vtr', 'Vertrag', true, true] },
    { name: 'AutoersetzungSchalten', argumente: (k) => [k.set.id, false] },
    { name: 'AutoersetzungSchalten', argumente: (k) => [k.set.id, true] },
    { name: 'AutoersetzungLoeschen', argumente: (k) => [k.set.id, 'vtr'] },
    { name: 'Markenkatalog' },
    { name: 'ObjektmarkeSetzen', argumente: (k) => ['karte', k.karten[0], 'vertrag'] },
    { name: 'MarkeVergeben', argumente: (k) => ['karte', [k.karten[0]], 'vertrag-zwei'] },
    { name: 'MarkenJeObjekt', argumente: () => ['karte'] },
    { name: 'MarkenFiltern', argumente: () => ['karte', ['vertrag'], 'und'] },
    { name: 'MarkeUmbenennen', argumente: () => ['vertrag-zwei', 'vertrag', true] },
    { name: 'ObjektmarkeEntfernen', argumente: (k) => ['karte', k.karten[0], 'vertrag'] },
    { name: 'Sprachpakete' },
    // **Eine Sprachkennung nach BCP 47** — Niederdeutsch, das es hier nicht gibt.
    { name: 'SprachpaketAnlegen', argumente: () => ['nds', 'Plattdüütsch', 'de'] },
    { name: 'Sprachpaket', argumente: () => ['nds'] },
    { name: 'SprachpaketAendern', argumente: () => ['nds', { name: 'Plattdüütsch', stand: 'entwurf', lizenz: 'GPL-2.0-or-later', herkunft: 'Rundgang' }, {}] },
    { name: 'SprachpaketLoeschen', argumente: () => ['nds'] },
  ],
}
