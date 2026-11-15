// **Welche Karten der Leistungsvergleich zeichnet** — QA-04.AK-04
// (#1735).
//
// # Aus der Zufallsstichprobe von QA-02
//
// **Gezogen wird nicht neu.** Die Karten kommen aus dem gespeicherten Plan
// `docs/belege/kanonstichprobe.json` (`docs/MESSMETHODIK.md` §4) — derselben
// Kohorte, an der der Abstand zu MSE2 gemessen ist. Je Regel werden die
// Stellen des Plans nach der Vorlage gefiltert und dann **gleichmäßig**
// verteilt gewählt, mit derselben Formel wie die feste Kohorte:
// `i · (len−1) / (anzahl−1)` trifft die erste und die letzte und verteilt den
// Rest. Eine Schicht, die weniger Karten hat als verlangt, gibt alle — und
// der Beleg nennt, wie viele es waren.
//
// # Die Vorlage einer Karte
//
// **Steht in der Setdatei**: `stylesheet:` auf oberster Ebene gilt für jede
// Karte, ein `stylesheet:` eine Ebene tiefer in einem `card:`-Block für diese
// eine (`src/data/card.cpp`, `IMPLEMENT_REFLECTION(Card)`, Zeile 77; die
// Felder der Karte selbst, darunter `name`, folgen namenlos in Zeile 95;
// Fassung f9d9356 vom 2021-07-06). Gelesen werden nur Felder **genau eine** Ebene
// tief — `name:` zwei Ebenen tief gehört zu etwas anderem.

/** kartenvorlagen liest je Karte einer Setdatei Name und Vorlage, in Dateireihenfolge. */
export function kartenvorlagen(settext) {
  const zeilen = settext.replace(/^\ufeff/, '').split('\n')
  const karten = []
  let vorgabe = null
  let karte = null
  for (const zeile of zeilen) {
    if (!zeile.startsWith('\t')) {
      karte = null
      if (zeile.startsWith('stylesheet:')) vorgabe = zeile.slice('stylesheet:'.length).trim()
      if (zeile.startsWith('card:')) {
        karte = { name: null, vorlage: null }
        karten.push(karte)
      }
      continue
    }
    if (karte === null || zeile.startsWith('\t\t')) continue
    const trenner = zeile.indexOf(':')
    const feld = zeile.slice(1, trenner)
    const wert = zeile.slice(trenner + 1).trim()
    if (feld === 'name' && karte.name === null) karte.name = wert
    if (feld === 'stylesheet') karte.vorlage = wert
  }
  return karten.map((k) => ({ name: k.name, vorlage: k.vorlage ?? vorgabe }))
}

/** gleichmaessig wählt `anzahl` Werte, über die Liste verteilt, die Ränder eingeschlossen. */
export function gleichmaessig(werte, anzahl) {
  if (werte.length <= anzahl) return [...werte]
  if (anzahl === 1) return [werte[0]]
  return Array.from({ length: anzahl }, (_, i) => werte[Math.floor((i * (werte.length - 1)) / (anzahl - 1))])
}

/**
 * waehleKarten wendet die Regeln auf den Plan an. `vorlagen` gibt je Satz die
 * Karten aus [kartenvorlagen]; eine Regel nennt Schicht, Satz, wahlweise eine
 * Vorlage und die Anzahl.
 */
export function waehleKarten(plan, vorlagen, regeln) {
  const auswahl = []
  for (const regel of regeln) {
    const schicht = plan.schichten.find((s) => s.satz === regel.satz)
    if (!schicht) throw new Error(`der Plan kennt den Satz ${regel.satz} nicht`)
    const karten = vorlagen[regel.satz] ?? []
    for (const stelle of schicht.stichprobe) {
      if (stelle >= karten.length) {
        throw new Error(`${regel.satz}: Stelle ${stelle} liegt hinter der letzten Karte (${karten.length}) — die Kartenliste passt nicht zum Plan`)
      }
    }
    const passend = schicht.stichprobe.filter((stelle) => regel.vorlage === undefined || karten[stelle].vorlage === regel.vorlage)
    for (const stelle of gleichmaessig(passend, regel.anzahl)) {
      auswahl.push({ schicht: regel.schicht, satz: regel.satz, stelle, name: karten[stelle].name, vorlage: karten[stelle].vorlage })
    }
  }
  return auswahl
}
