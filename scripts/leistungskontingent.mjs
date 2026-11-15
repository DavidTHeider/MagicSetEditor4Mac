// **Das Guthaben von MSE4Mac im Leistungsvergleich** — Z1 des Zweitblicks über
// Phase 6.11.
//
// # Warum es abgeschaltet wird
//
// **Ein Leistungsvergleich misst nicht die Drosselung.** `mse-mcp` und
// `mse-rest` führen ein Guthaben je Principal (`grenzen.kontingente`,
// `internal/domain/kontingentschalter.go`), und seit Zeichnen schnell ist
// (#2005), ist es nach 70 Karten leer: der Batchfall im Beleg vom 2026-10-03
// zeichnete 24 von 70 Bildern, und das Dokument stellte seine 2,4 s neben die
// 3,0 s von MSE2 für alle 70. Abgeschaltet war das Guthaben nur im Langlauf.
//
// Gegenstand ist die Drossel **nur im Teil `last`** — dort ist sie die Frage —;
// für jeden anderen Teil ist sie eine Bremse ohne Gegenüber. MSE2 und 0.5.0
// haben keines: der Schalter gehört allein dem Teilnehmer `heute`.
//
// # Einmal je Lauf, und zurück
//
// Das Abschalten steht **im Bestand** (`mse einstellungen setzen`) und gilt
// beim Start eines Dienstes. Es geschieht einmal vor dem ersten Teil und wird
// vor dem Teil `last` und am Ende zurückgestellt — **auf den Stand, den der
// Lauf vorgefunden hat**, auch wenn ein Teil abbricht. Die Rückstellung für
// einen Abbruch durch ein Signal verdrahtet der Läufer (`leistungsvergleich.mjs`);
// gegen `SIGKILL` hilft keine Verdrahtung, und der Bestand ist ein eigener
// unter `ARBEIT`, kein echter.

/** Die Teile, deren Gegenstand die Drossel ist — nur dort gilt das Guthaben. */
export const drosselteile = ['last']

const werte = ['an', 'aus']

/**
 * kontingentstandAus liest den Stand aus der Antwort von
 * `mse einstellungen --json` (`cmd/mse/bestandsbefehle.go`,
 * `einstellungenZeigen`): `an` oder `aus`. **Was fehlt oder fremd ist, bleibt
 * ein Fehler** — ein geratener Stand würde später zurückgeschrieben.
 */
export function kontingentstandAus(antwort) {
  let eintrag
  try {
    eintrag = JSON.parse(antwort).einstellungen?.find((e) => e.schluessel === 'grenzen.kontingente')
  } catch {
    eintrag = undefined
  }
  if (!eintrag) throw new Error('grenzen.kontingente steht nicht in der Auskunft der Einstellungen')
  const stand = eintrag.wirksam ?? eintrag.wert
  if (!werte.includes(stand)) throw new Error(`grenzen.kontingente hat den Wert ${JSON.stringify(stand)}, erwartet an oder aus`)
  return stand
}

/**
 * kontingentschalter gibt `aus` und `zurueck` über `lies` und `setze`.
 *
 * **`aus` merkt sich den alten Stand, bevor es setzt**, und vermerkt die
 * Änderung vor dem Setzen: ein Setzen, das scheitert, kann dennoch geschrieben
 * haben, und `zurueck` stellt dann zurück. **Beide sind wiederholbar**: ein
 * zweites `aus` ändert nichts, ein zweites `zurueck` ebenso.
 */
export function kontingentschalter({ lies, setze }) {
  let alt = null
  return {
    async aus() {
      if (alt !== null) return
      const vorgefunden = await lies()
      alt = vorgefunden
      await setze('aus')
    },
    async zurueck() {
      if (alt === null) return
      await setze(alt)
      alt = null
    },
  }
}

/**
 * schalterFuer gibt den Schalter für den Vorrat eines Teilnehmers — das
 * Lesen und das Setzen laufen über dessen `mse`, und ein Fehler dabei bricht
 * ab, statt eine Messung ohne bekannten Stand zu beginnen.
 */
export function schalterFuer(t) {
  const pruefe = (m, was) => {
    if (m.fehler) throw new Error(`${t.name}: die Kontingente lassen sich nicht ${was} — ${m.fehler}\n${m.fehlerText ?? ''}`)
    return m
  }
  return kontingentschalter({
    lies: async () => kontingentstandAus(pruefe(await t.kontingenteStand(t.vorrat), 'lesen').aus),
    setze: async (stand) => { pruefe(await t.kontingente(t.vorrat, stand), `auf ${stand} stellen`) },
  })
}

/**
 * fahreTeile fährt die Teile der Reihe nach, jeden mit dem Guthaben, das zu
 * ihm gehört, und stellt am Ende zurück — auch wenn einer abbricht.
 */
export async function fahreTeile(teile, schalter, fahre) {
  try {
    for (const teil of teile) {
      if (drosselteile.includes(teil)) await schalter.zurueck()
      else await schalter.aus()
      await fahre(teil)
    }
  } finally {
    await schalter.zurueck()
  }
}
