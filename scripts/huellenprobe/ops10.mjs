// **Bereiche und Funktionen abschalten, am nativen Fenster** — OPS-10
// (#1908):
//
// 1. vorher steht *Spielen* in der Leiste,
// 2. in den Einstellungen unter *Bereiche und Funktionen* auf „aus“ — und
//    der Eintrag ist fort, ohne Neustart,
// 3. ⌘4 führt trotzdem hin, und dort steht, warum nichts dasteht, samt Weg
//    zurück,
// 4. die Kommandozeile weist denselben Bereich ab,
// 5. wieder „an“ — und *Spielen* steht wieder da.

const leistentext = (p) => p.seite("return document.querySelector('#seitenleiste nav')?.innerText ?? ''")

// setzeWahl legt einen Schalter um — **seit UXR-19 (#1927) ein Schalter und
// keine Auswahlliste** (R7). Über `data-einstellung`, und nur, wenn er
// nicht schon so steht: ein Klick legt um, er setzt keinen Wert.
const setzeWahl = (schluessel, wert) => `const s = document.querySelector('input[role=switch][data-einstellung=${JSON.stringify(schluessel)}]'); if (!s || !s.checkVisibility()) return false; if (s.checked !== (${JSON.stringify(wert)} === 'an')) s.click(); return true`

export default async function ops10(p) {
  const ort = p.ort()
  await p.starte({ ort })

  p.notiz('1 — vorher steht Spielen in der Leiste')
  await p.warte('die Leiste steht', async () => (await leistentext(p)).includes('Spielen'), 20000)
  await p.bild('vorher')

  p.notiz('2 — Einstellungen › Bereiche und Funktionen › Spielen aus')
  await p.taste(',', 'befehl')
  await p.wartetext('Bereiche und Funktionen', 20000)
  await p.klickeText('Bereiche und Funktionen')
  await p.wartetext('Decks, Turniere, Partien und der Tisch', 20000)
  await p.bild('schalter')
  p.pruefe('die Wahl für Spielen steht da', await p.seite(setzeWahl('schalter.spielen', 'aus')))
  await p.warte('Spielen ist aus der Leiste fort', async () => !(await leistentext(p)).includes('Spielen'), 20000)
  p.pruefe('die übrigen Bereiche stehen weiter da', (await leistentext(p)).includes('Sammeln'))
  await p.bild('abgeschaltet')

  p.notiz('3 — ⌘4 führt hin, und die Seite sagt, warum nichts dasteht')
  await p.taste('4', 'befehl')
  await p.wartetext('ist abgeschaltet', 20000)
  p.pruefe('der Weg zurück steht da', await p.seite("return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Bereiche und Funktionen öffnen')"))
  await p.bild('seite')

  p.notiz('4 — die Kommandozeile weist denselben Bereich ab')
  let abgewiesen = ''
  try {
    p.mse(ort, 'deck', 'liste')
  } catch (fehler) {
    abgewiesen = String(fehler.stderr ?? fehler.message ?? fehler)
  }
  p.pruefe('mse deck liste nennt den abgeschalteten Bereich', abgewiesen.includes('dieser Bereich ist abgeschaltet') && abgewiesen.includes('spielen'), abgewiesen)

  p.notiz('5 — der Weg zurück, und wieder an')
  await p.klickeText('Bereiche und Funktionen öffnen')
  await p.wartetext('Decks, Turniere, Partien und der Tisch', 20000)
  p.pruefe('die Wahl für Spielen steht wieder da', await p.seite(setzeWahl('schalter.spielen', 'an')))
  await p.warte('Spielen steht wieder in der Leiste', async () => (await leistentext(p)).includes('Spielen'), 20000)
  await p.bild('wieder-an')

  p.pruefe('⌘Q beendet', await p.beende())
}
