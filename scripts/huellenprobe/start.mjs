// **Starten und Beenden am nativen Fenster** — ENG-02.AK-04, Kernweg 1
// (#1727),
// und UX-14.AK-01 (#1686):
// das Fenster geht maximiert auf — und nicht im Vollbild —, ⌘Q beendet, mit
// „wie zuletzt" kommt die Größe von vorhin zurück, und eine, die nicht auf
// den Bildschirm passt, wird geklemmt (UX-14.AK-04). **Und der Abnahmebau
// selbst ist verschlossen**: Wails' MCP-Server weist ohne Schlüssel ab, und
// ohne die Probe startet er gar nicht (N1 des Nachlaufs über Phase 6.3).
import { spawnSync } from 'node:child_process'
import { existsSync } from 'node:fs'
import { join, resolve } from 'node:path'

const buendel = resolve(import.meta.dirname, '..', '..', 'build/bin/MagicSetEditor4Mac.app')

const nah = (a, b) => Math.abs(a - b) <= 2

// stehend wartet, bis der Rahmen sich nicht mehr bewegt — macOS zieht ein
// Fenster in einer Bewegung auf, und mitten darin misst man eine Zwischengröße.
async function stehend(p, frist = 5000) {
  const ende = Date.now() + frist
  let vorher = await p.fenster()
  while (Date.now() < ende) {
    await new Promise((los) => setTimeout(los, 250))
    const jetzt = await p.fenster()
    if (jetzt.sichtbar && JSON.stringify(jetzt.rahmen) === JSON.stringify(vorher.rahmen)) return jetzt
    vorher = jetzt
  }
  return vorher
}

export default async function start(p) {
  const ort = p.ort()

  p.notiz('Die Vorgabe: maximiert')
  await p.starte({ ort })
  const lage = await stehend(p)
  // **Die Höhe auf vier Punkte**: die Arbeitsfläche maß in derselben Sitzung
  // 838 und 841 Punkte — die Leiste oben ist nicht immer gleich hoch.
  const voll = nah(lage.rahmen.B, lage.arbeitsflaeche.B) && Math.abs(lage.rahmen.H - lage.arbeitsflaeche.H) <= 4
  p.pruefe('das Fenster füllt die Arbeitsfläche', lage.gezoomt && voll,
    `Rahmen ${String(lage.rahmen.B)} × ${String(lage.rahmen.H)}, Arbeitsfläche ${String(lage.arbeitsflaeche.B)} × ${String(lage.arbeitsflaeche.H)}, gezoomt ${String(lage.gezoomt)}`)
  p.pruefe('und nicht im Vollbild', !lage.vollbild)
  await p.bild('start-maximiert')
  const ohne = await fetch(`http://127.0.0.1:${String(p.port)}/mcp`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ jsonrpc: '2.0', id: 1, method: 'tools/list' }),
  })
  p.pruefe("Wails' MCP-Server weist ohne Schlüssel ab", ohne.status === 401, `Status ${String(ohne.status)}`)
  p.pruefe('⌘Q beendet die Anwendung', await p.beende())

  p.notiz('„Wie zuletzt": die Größe beim Schließen kommt wieder')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
  await p.starte({ ort })
  await p.groesse(1100, 760)
  const vorher = await p.warte('das Fenster ist 1.100 × 760', async () => {
    const f = await p.fenster()
    return nah(f.rahmen.B, 1100) ? f : false
  })
  p.zeige(`vor dem Schließen: ${String(vorher.rahmen.B)} × ${String(vorher.rahmen.H)}`)
  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
  await p.starte({ ort })
  const nachher = await stehend(p)
  p.pruefe('dieselbe Größe wie beim Schließen', nah(nachher.rahmen.B, vorher.rahmen.B) && nah(nachher.rahmen.H, vorher.rahmen.H),
    `${String(nachher.rahmen.B)} × ${String(nachher.rahmen.H)}`)
  p.pruefe('und nicht maximiert', !nachher.gezoomt)
  p.pruefe('Beenden über die Leiste', await p.beende({ ueber: 'leiste' }))

  // **Eine gemerkte Größe, die größer ist als der Bildschirm** — UX-14.AK-04:
  // so kommt sie an, wenn sie auf einem großen Bildschirm entstand und der
  // abgezogen ist. Auf Retina griff die Klemmung bis zum Zweitblick über
  // Phase 6.3 nie (S2): Wails liefert die Fläche in Bildpunkten.
  p.notiz('Eine zu große gemerkte Größe passt auf den Bildschirm')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', '3000')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', '2000')
  await p.starte({ ort })
  const geklemmt = await stehend(p)
  p.pruefe('das Fenster passt auf die Arbeitsfläche', geklemmt.rahmen.B <= geklemmt.arbeitsflaeche.B + 1 && geklemmt.rahmen.H <= geklemmt.arbeitsflaeche.H + 1,
    `Rahmen ${String(geklemmt.rahmen.B)} × ${String(geklemmt.rahmen.H)}, Arbeitsfläche ${String(geklemmt.arbeitsflaeche.B)} × ${String(geklemmt.arbeitsflaeche.H)}`)
  p.pruefe('⌘Q beendet die Anwendung', await p.beende())

  // **Ohne die Probe startet der Abnahmebau nicht** — mit einem frischen
  // Bestand, damit auch ein Rückfall nichts Echtes berührt.
  p.notiz('Ohne die Probe startet der Abnahmebau nicht')
  const leer = p.ort()
  const lauf = spawnSync('open', ['-n', '-W', '-a', buendel, '--env', `MSE_BESTAND=${leer.bestand}`, '--env', `MSE_SCHLUESSELBUND=${leer.bund}`], { timeout: 20000 })
  if (lauf.status === null) {
    // **Ein Rückfall bleibt nicht stehen** — der Pfad ist der dieses Bündels.
    spawnSync('pkill', ['-f', join(buendel, 'Contents/MacOS/MagicSetEditor4Mac')])
  }
  p.pruefe('er endet von selbst', lauf.status !== null, lauf.status === null ? 'nach 20 s noch da' : 'beendet')
  p.pruefe('und legt keinen Bestand an', !existsSync(leer.bestand))
}
