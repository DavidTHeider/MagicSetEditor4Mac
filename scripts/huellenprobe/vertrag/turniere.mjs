// **Turniere** — eine Gruppe des Vertragsrundgangs (B01, #2220). **Ein
// Einzelspielturnier**: ein Match ist ein Spiel, das Ergebnis 1 : 0. Braucht
// `k.set` und `k.deck`; legt ein Einzelspielturnier mit zwei Teilnehmenden an,
// paart eine Runde, trägt ein Ergebnis ein, schließt ab, öffnet wieder.

const turnierfassung = 'a.turnier.fassung'

export const turniere = {
  name: 'Turniere',
  rufe: [
    { name: 'Turniere', argumente: (k) => [k.set.spiel] },
    { name: 'TurnierAnlegen', argumente: (k) => ['Vertragsturnier', k.set.spiel, 'aus dem Rundgang', true, ''], nimm: 'a.turnier.id', nach: (w, k) => { k.turnier = w } },
    { name: 'TurnierAnmelden', argumente: (k) => [k.turnier, 'Ada', k.deck ?? ''] },
    { name: 'TurnierAnmelden', argumente: (k) => [k.turnier, 'Bo', k.deck ?? ''] },
    { name: 'TurnierPaarungsvorschlag', argumente: (k) => [k.turnier] },
    { name: 'TurnierRundeLegen', argumente: (k) => [k.turnier], nimm: 'a.runde.paarungen.map((x) => x.id)', nach: (w, k) => { k.paarung = w[0] } },
    { name: 'Turnier', argumente: (k) => [k.turnier], nimm: turnierfassung, nach: (w, k) => { k.turnierfassung = w } },
    { name: 'TurnierKorrekturfolgen', argumente: (k) => [k.turnier, k.paarung, { spieleA: 1, spieleB: 0, unentschieden: 0 }] },
    { name: 'TurnierErgebnisSetzen', argumente: (k) => [k.turnier, k.turnierfassung, k.paarung, { spieleA: 1, spieleB: 0, unentschieden: 0 }] },
    { name: 'TurnierRangliste', argumente: (k) => [k.turnier] },
    { name: 'TurnierProtokoll', argumente: (k) => [k.turnier] },
    { name: 'TurnierMatch', argumente: (k) => [k.paarung] },
    { name: 'TurnierKopieren', argumente: (k) => [k.turnier, 'Vertragsturnier Kopie', true], nimm: 'a.turnier.id', nach: (w, k) => { k.turnierKopie = w } },
    { name: 'Turnier', argumente: (k) => [k.turnier], nimm: turnierfassung, nach: (w, k) => { k.turnierfassung = w } },
    { name: 'TurnierAbschliessen', argumente: (k) => [k.turnier, k.turnierfassung] },
    { name: 'Turnier', argumente: (k) => [k.turnier], nimm: turnierfassung, nach: (w, k) => { k.turnierfassung = w } },
    { name: 'TurnierWiedereroeffnen', argumente: (k) => [k.turnier, k.turnierfassung] },
    { name: 'TurnierLoeschen', argumente: (k) => [k.turnierKopie] },
  ],
}
