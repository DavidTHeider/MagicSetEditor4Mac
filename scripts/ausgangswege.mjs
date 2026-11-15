// **Der Vertrag der Ausgangswerte für Spielen und Handeln** — P0
// (#2336),
// Befund C-22 aus dem Audit zur Produktionsreife 0.9.0.
//
// # Wofür
//
// Die Pakete P4 bis P13 bauen Tisch und Markttisch um, und jedes nennt ein
// Ziel in Handlungen (S1 ≤ 3, H1 ≤ 5, …). **Ein Ziel ohne Ausgangswert ist
// eine Behauptung** — und ein Nachherlauf, der anders zählt als der
// Vorherlauf, vergleicht zwei Dinge. Deshalb steht hier, **welche Wege** ein
// Datensatz tragen muss und **was** eine Zeile ausmacht; die Drehbücher
// `ausgangswerte09spielen.mjs` und `ausgangswerte09handeln.mjs` füllen ihn,
// und derselbe Vertrag prüft den Nachherlauf.
//
// # Was eine Zeile ist
//
// Ein Weg mit **Art** (`maus` oder `tastatur`) und **Eingang**, gezählt nach
// `huellenwege.mjs`: Handlungen, Sichtwechsel, Tabstellen, Fehlversuche,
// Maschinenzeit, **bedingte Griffe** — und die **beobachtete Wirkung**. Ohne Wirkung ist ein Weg
// gegangen, nicht gemessen. **Oder** er ist `NOT_RUN` bzw. `FEHLT` — mit
// Grund: ein Weg, den es nicht gibt (S11, Rückgängig), steht als `FEHLT` da
// und nicht als Lücke.
//
// # Was er nicht hält
//
// **Ob die Zahl stimmt.** Er hält die Vollständigkeit und die Gestalt; dass
// die Probe richtig zählt, belegt ihr Lauf mit seinen Bildern.

// pflichtwege sind die Wege aus Audit C, Abschnitt 3, je mit ihren Arten.
//
// **S14 fehlt mit Absicht**: das einmalige Aufnehmen eines Gegenübers ist
// kein Weg je Partie und hat seinen Beleg in `belege/personawege-g01.md`.
export const pflichtwege = [
  { weg: 'S1', eingang: 'bereich', arten: ['maus', 'tastatur'] },
  { weg: 'S1', eingang: 'deck', arten: ['maus'] },
  { weg: 'S2', eingang: 'tisch', arten: ['maus', 'tastatur'] },
  { weg: 'S3', eingang: 'tisch', arten: ['maus', 'tastatur'] },
  { weg: 'S4', eingang: 'tisch', arten: ['maus'] },
  { weg: 'S5', eingang: 'tisch', arten: ['maus', 'tastatur'] },
  { weg: 'S6', eingang: 'tisch', arten: ['maus', 'tastatur'] },
  { weg: 'S7', eingang: 'tisch', arten: ['maus'] },
  { weg: 'S8', eingang: 'tisch', arten: ['maus'] },
  { weg: 'S9', eingang: 'tisch', arten: ['maus'] },
  { weg: 'S10', eingang: 'tisch', arten: ['maus', 'tastatur'] },
  { weg: 'S11', eingang: 'tisch', arten: ['maus'] },
  { weg: 'S12', eingang: 'bereich', arten: ['maus'] },
  { weg: 'S13', eingang: 'bereich', arten: ['maus'] },
  { weg: 'H1', eingang: 'bereich', arten: ['maus', 'tastatur'] },
  { weg: 'H2', eingang: 'sammlung', arten: ['maus'] },
  { weg: 'H3', eingang: 'bereich', arten: ['maus'] },
  { weg: 'H4', eingang: 'bereich', arten: ['maus'] },
  { weg: 'H5', eingang: 'handel', arten: ['maus'] },
  { weg: 'H6', eingang: 'handel', arten: ['maus'] },
]

const schluessel = (z) => `${z.weg}/${z.eingang}/${z.art}`

// pruefeDatensatz nennt, was einem Datensatz fehlt — leer heißt vollständig.
export function pruefeDatensatz(datensatz) {
  const fehler = []

  if (!/^[0-9a-f]{12,}$/.test(datensatz?.bauabdruck ?? '')) {
    fehler.push('der Datensatz nennt keinen Bauabdruck — ohne ihn ist er keinem Bündel zuzuordnen')
  }

  const zeilen = new Map((datensatz?.wege ?? []).map((z) => [schluessel(z), z]))

  for (const pflicht of pflichtwege) {
    for (const art of pflicht.arten) {
      const k = `${pflicht.weg}/${pflicht.eingang}/${art}`
      const zeile = zeilen.get(k)

      if (zeile === undefined) {
        fehler.push(`${k}: fehlt im Datensatz`)
      } else if (zeile.status === 'NOT_RUN' || zeile.status === 'FEHLT') {
        if (!zeile.grund) fehler.push(`${k}: ${zeile.status} ohne Grund`)
      } else if (zeile.wirkung !== true) {
        fehler.push(`${k}: gegangen, aber keine Wirkung beobachtet`)
      } else if (!Number.isInteger(zeile.handlungen) || zeile.handlungen < 1) {
        fehler.push(`${k}: keine Handlung gezählt`)
      } else if (!Array.isArray(zeile.bedingt)) {
        fehler.push(`${k}: nennt keine bedingten Griffe — der Zähler ist älter als der Vertrag`)
      }
    }
  }

  return fehler
}

// vereinige legt die Teildatensätze zusammen — Spielen, Netz und Handeln
// laufen als drei Drehbücher, gemessen wird ein Bündel. **Aus verschiedenen
// Bündeln** wären es zwei Messungen, und ihr Vergleich hieße nichts.
export function vereinige(teile) {
  const abdruecke = [...new Set(teile.map((t) => t?.bauabdruck ?? ''))]
  const fehler = abdruecke.length > 1 ? [`die Teildatensätze stammen aus verschiedenen Bündeln: ${abdruecke.join(', ')}`] : []

  return { datensatz: { bauabdruck: abdruecke[0] ?? '', wege: teile.flatMap((t) => t?.wege ?? []) }, fehler }
}
