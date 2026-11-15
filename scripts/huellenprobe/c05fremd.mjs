// **Was ein anderer Prozess schreibt, erreicht das offene Fenster** — C05
// (#2230).
//
// # Der Anlass
//
// Im Gastlauf von A03 wirkte `EinstellungSetzen`, über die Bindung an der
// Oberfläche vorbei gerufen, erst nach dem Neuladen. **Das ist kein Weg des
// Betriebs**: im Fenster setzt die Oberfläche ihren eigenen Stand, und jede
// andere Fläche — Kommandozeile, MCP, REST — ist ein eigener Prozess. Für
// sie fragt das Fenster alle zwei Sekunden `PRAGMA data_version`
// (`useBestandsstand`, [0017](../../docs/entscheidungen/0017-die-prozesse-bemerken-einander.md)).
// Dieses Drehbuch belegt den Weg, den es wirklich gibt.
//
// # Der Vertrag
//
// - Eine **Einstellung**, die `mse` umlegt, gilt im offenen Fenster binnen
//   zehn Sekunden — hin und zurück.
// - Ein **Kartenfeld**, das `mse` ändert, steht binnen zehn Sekunden in der
//   offenen Kartenliste.
//
// **Was es nicht belegt**: eine Änderung, während jemand im selben Feld
// tippt — welche gewinnt, entscheidet das Fenster beim Speichern, und das
// prüft dieses Drehbuch nicht.

const bild = 'return document.documentElement.dataset.erscheinungsbild ?? ""'

export default async function c05fremd(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'dunkel')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  await p.starte({ ort })

  p.notiz('Eine Einstellung aus einem anderen Prozess')
  p.pruefe('vorher: dunkel', (await p.seite(bild)) === 'dunkel')
  for (const wahl of ['hell', 'dunkel']) {
    const beginn = Date.now()
    p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', wahl)
    await p.warte(`das Fenster zeigt ${wahl}`, async () => (await p.seite(bild)) === wahl, 10000)
    p.pruefe(`${wahl}: nach ${String(Date.now() - beginn)} ms im Fenster`, true)
  }

  p.notiz('Ein Kartenfeld aus einem anderen Prozess')
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  await p.wartetext('Langtext', 20000)
  const beginn = Date.now()
  p.mse(ort, 'feld', '1', '1', 'name', 'Fremd geschrieben')
  await p.wartetext('Fremd geschrieben', 10000)
  p.pruefe(`der neue Name steht nach ${String(Date.now() - beginn)} ms in der Liste`, true)
  await p.bild('fremd-geschrieben')

  p.pruefe('⌘Q beendet', await p.beende())
}
