// **⌘Q mit laufenden eigenen Läufen, in der Frist** — BK-07a.AK-04 und
// BK-02.AK-03 (#2115,
// #2110).
//
// 1. **Regulär**: ein Horchposten läuft, das Fenster hat eine eigene Größe,
//    ⌘Q. Gemessen wird, wie lange das Programm braucht, ob die Größe im
//    Bestand steht und ob sie nach einem Neustart wieder gilt.
// 2. **Mit gesperrtem Bestand**: ein fremder Schreiber hält die
//    Schreibsperre, das Fenster bekommt eine andere Größe, ⌘Q. Das Merken
//    kann nicht schreiben — das Programm endet trotzdem in der Frist, die
//    alte Größe bleibt, und das Protokoll nennt den Verlust.
//
// **Was sie nicht belegt**: einen Peer- oder Partielauf über ein zweites
// Programm (`nat04` fährt ihn) und die KI-Läufe — sie gehören ihrem Ruf
// (Vermerk 0170) und brauchen einen Anbieter.
import { DatabaseSync } from 'node:sqlite'
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'

const bindung = (p, ausdruck) => p.seite(`const b = window.go.app.Bindungen; return await ${ausdruck}`)

const gemerkt = (p, ort) => {
  const zeilen = p.mse(ort, 'einstellungen').split('\n')
  const wert = (name) => (zeilen.find((z) => z.startsWith(name)) ?? '').split(/\s+/)[1] ?? ''

  return { breite: wert('fensterbreite'), hoehe: wert('fensterhoehe') }
}

export default async function bk07a(p) {
  const ort = p.ort()

  p.notiz('1 — regulär: ein Horchposten läuft, eigene Größe, ⌘Q')
  // **„Wie zuletzt"** — die Vorgabe ist „maximiert", und dann gilt keine
  // gemerkte Größe (`domain.FensterstartMaximiert`; gemessen: 1471 pt nach
  // dem Neustart, als diese Zeile fehlte).
  p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
  await p.starte({ ort })
  p.pruefe('eine Identität steht', !(await bindung(p, 'b.PeeridentitaetSicherstellen()')).fehler)
  const horchen = await bindung(p, "b.HorchStarten('127.0.0.1:0')")
  p.pruefe('der Horchposten läuft', horchen.stand?.laeuft === true, horchen.stand?.adresse ?? JSON.stringify(horchen.fehler))
  await p.groesse(1300, 850)
  await new Promise((fertig) => { setTimeout(fertig, 500) })

  let anfang = Date.now()
  p.pruefe('⌘Q beendet', await p.beende())
  const regulaer = Date.now() - anfang
  p.zeige(`bis der Prozess weg war: ${regulaer} ms`)
  p.pruefe('in der Frist des Schlusses (5 s) und einer Sekunde Spiel', regulaer < 6000)

  const nachRegulaer = gemerkt(p, ort)
  p.zeige(`gemerkt: ${nachRegulaer.breite} × ${nachRegulaer.hoehe}`)
  // **Die Höhe, die das Fenster meldet**, nicht die gesetzte: gemessen
  // 1300 × 843 nach `set_size` 1300 × 850 — die Breite stimmt genau.
  p.pruefe('die Größe steht im Bestand', nachRegulaer.breite === '1300' && Math.abs(Number(nachRegulaer.hoehe) - 850) <= 10)

  p.notiz('2 — nach dem Neustart gilt die gemerkte Größe')
  await p.starte({ ort })
  // **`innerWidth` und nicht `outerWidth`**: in WKWebView ist `outerWidth`
  // null (gemessen); seitlich hat das Fenster keinen Rahmen.
  const breite = await p.seite('return window.innerWidth')
  p.zeige(`das Fenster ist ${breite} pt breit`)
  p.pruefe('so breit wie gemerkt', Math.abs(breite - 1300) <= 2)

  p.notiz('3 — mit gesperrtem Bestand: das Merken geht verloren, das Programm endet trotzdem')
  await p.groesse(1100, 760)
  await new Promise((fertig) => { setTimeout(fertig, 500) })
  const fremd = new DatabaseSync(ort.bestand)
  fremd.exec('BEGIN IMMEDIATE')

  let gesperrt
  try {
    anfang = Date.now()
    p.pruefe('⌘Q beendet', await p.beende())
    gesperrt = Date.now() - anfang
  } finally {
    fremd.exec('ROLLBACK')
    fremd.close()
  }

  p.zeige(`bis der Prozess weg war: ${gesperrt} ms`)
  p.pruefe('in der Frist des Schlusses (5 s) und einer Sekunde Spiel', gesperrt < 6000)

  const nachGesperrt = gemerkt(p, ort)
  p.zeige(`gemerkt: ${nachGesperrt.breite} × ${nachGesperrt.hoehe}`)
  p.pruefe('die alte Größe blieb — ein Bestenfalls-Verlust', nachGesperrt.breite === '1300')

  const protokoll = readFileSync(join(dirname(ort.bestand), 'mse.log'), 'utf8')
  const zeile = protokoll.split('\n').filter((z) => z.includes('Schluss') || z.includes('verloren')).at(-1) ?? ''
  p.zeige(`im Protokoll: ${zeile.slice(0, 200)}`)
  p.pruefe('das Protokoll nennt den Verlust', /der Schluss kam nicht in seiner Frist zu Ende|beim Schließen ging etwas verloren/.test(protokoll))
  // **Und nur ihn**: das Merken bekommt höchstens eine Sekunde, die eigenen
  // Läufe enden trotzdem (`Bestenfallsfrist`, gemessen am ersten Lauf).
  p.pruefe('die eigenen Läufe kamen trotzdem zu Ende', !zeile.includes('das Beenden der eigenen Läufe'))
}
