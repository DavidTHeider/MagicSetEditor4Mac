// **Sets und Karten** — die erste Gruppe des Vertragsrundgangs (B01, #2220).
//
// Sie legt die Kennungen an, die die übrigen Gruppen brauchen: `k.set` (das
// Saatset), `k.karten`, `k.neueKarte`, `k.textfeld`. **Die Reihenfolge ist
// Teil des Vertrags**: erst lesen, dann anlegen und ändern, dann
// zurücknehmen; gelöscht wird erst ganz am Ende des Rundgangs.

export const karten = {
  name: 'Sets und Karten',
  rufe: [
    { name: 'Sets', nimm: 'a.sets.map((s) => ({ id: s.id, spiel: s.spiel, stil: s.stil }))', nach: (w, k) => { k.set = w[0] } },
    { name: 'Setangaben', argumente: (k) => [k.set.id] },
    { name: 'Technikangaben', argumente: (k) => [k.set.id] },
    { name: 'Karten', argumente: (k) => [k.set.id], nimm: 'a.karten.map((c) => c.id)', nach: (w, k) => { k.karten = w } },
    { name: 'Kartenzeilen', argumente: (k) => [k.set.id] },
    // **Unicode kommt unverändert an** (C01, #2226) — Umlaute, Anführungen,
    // ein CJK-Zeichen und eines außerhalb der Grundebene, aus der Saat
    // gelesen, wie es dort steht (`saat/probe-werkstatt.mse-set/set`).
    {
      name: 'Karten', kohorte: 'unicode', argumente: (k) => [k.set.id],
      nimm: "a.karten.flatMap((c) => c.werte).find((w) => w.feld === 'name' && w.wert.startsWith('Sonderzeichen'))?.wert ?? ''",
      nach: (w) => { if (w !== 'Sonderzeichen ÄÖÜ ß «» ✦ ☙ 漢字 🃏') throw new Error(`gelesen ${JSON.stringify(w)}`) },
    },
    {
      name: 'Kartenfelder', argumente: (k) => [k.set.id],
      nimm: "a.felder.filter((f) => f.editierbar && f.art === 'text').map((f) => f.name)",
      nach: (w, k) => { k.textfeld = w.includes('name') ? 'name' : w[0] },
    },
    { name: 'Kartenfeldwerte', argumente: (k) => [k.set.id, k.textfeld] },
    { name: 'Karte', argumente: (k) => [k.karten[0]] },
    { name: 'Vorschau', argumente: (k) => [k.karten[0], 1] },
    { name: 'Kartenbilder', argumente: (k) => [k.set.id, [k.karten[0]]] },
    { name: 'Vorbehalte', argumente: (k) => [k.set.id] },
    { name: 'Schluesselwoerter', argumente: (k) => [k.set.id] },
    { name: 'Skriptnamen', argumente: (k) => [k.karten[0]] },
    { name: 'Skript', argumente: (k) => [k.karten[0], '1 + 1'] },
    { name: 'Wortlisten', argumente: (k) => [k.karten[0], k.textfeld] },
    { name: 'Symbolpalette', argumente: (k) => [k.karten[0], k.textfeld] },
    { name: 'KarteAnlegen', argumente: (k) => [k.set.id], nimm: 'a.karteID', nach: (w, k) => { k.neueKarte = w } },
    { name: 'FeldSetzen', argumente: (k) => [k.set.id, k.neueKarte, k.textfeld, 'Vertragskarte', []] },
    { name: 'FelderVorschau', argumente: (k) => [k.set.id, [k.neueKarte], k.textfeld, 'Vertragskarte zwei'] },
    {
      name: 'FelderSetzenMitErwartung',
      argumente: (k) => [k.set.id, [{ karteID: k.neueKarte, wert: 'Vertragskarte', gabEs: true }], k.textfeld, 'Vertragskarte zwei'],
    },
    { name: 'KarteKopieren', argumente: (k) => [k.set.id, k.neueKarte], nimm: 'a.karteID', nach: (w, k) => { k.kopie = w } },
    { name: 'KartenTauschen', argumente: (k) => [k.set.id, k.neueKarte, k.kopie] },
    { name: 'Verlauf', argumente: (k) => [k.set.id], nimm: 'a.schritte.map((s) => s.vorgang).filter(Boolean)', nach: (w, k) => { k.vorgang = w.at(-1) } },
    { name: 'Zuruecknehmen', argumente: (k) => [k.set.id, ''] },
    { name: 'Wiederherstellen', argumente: (k) => [k.set.id, ''] },
    { name: 'SetwertSetzen', argumente: (k) => [k.set.id, 'title', 'Vertragsset', []] },
    { name: 'SchluesselwortAnlegen', argumente: (k) => [k.set.id], nimm: 'a.wortID', nach: (w, k) => { k.wort = w } },
    { name: 'SchluesselwortSetzen', argumente: (k) => [k.set.id, k.wort, 'wort', 'Vertragswort', []] },
    { name: 'SchluesselwortLoeschen', argumente: (k) => [k.set.id, k.wort] },
    { name: 'KarteLoeschen', argumente: (k) => [k.set.id, k.kopie, true] },
    { name: 'SetBenutzt', argumente: (k) => [k.set.id] },
    { name: 'Datenstand' },
    { name: 'Startbefund' },
  ],
}
