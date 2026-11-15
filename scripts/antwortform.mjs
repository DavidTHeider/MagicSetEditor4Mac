// **Die Form einer Antwort prüfen** — C01
// (#2226).
//
// Eine Form beschreibt, was das Fenster von einer Bindung erwartet — gelesen
// aus seinen TypeScript-Typen (`scripts/antwortformen.mjs`). `pruefeForm`
// hält eine **echte** Antwort dagegen: was fehlt, was den falschen Typ hat
// und was zusätzlich dasteht.
//
// **Als Quelltext und als Funktion zugleich**: der Vertragsrundgang prüft in
// der Seite (die Antwort muss dann nicht über die Leitung), die Tests prüfen
// in Node — dieselbe Funktion, nicht zwei.
//
// # Was rot ist und was nicht
//
// - **rot**: ein Pflichtfeld fehlt, ein Wert hat einen anderen Typ, `null`
//   steht, wo die Form es nicht erlaubt, ein Literal passt zu keiner
//   Möglichkeit;
// - **gezählt, nicht rot**: ein Feld, das die Form nicht kennt (das Fenster
//   übergeht es), und eine Stelle, deren Form `unbekannt` ist.

export const pruefeFormQuelle = `function pruefeForm(wert, form) {
  const abweichungen = []
  const zusaetzlich = []
  let unbekannt = 0
  const art = (w) => (w === null ? 'null' : Array.isArray(w) ? 'liste' : typeof w)
  const passt = (w, f) => {
    const probe = []
    pruefe(w, f, '', probe, [], { zaehle: () => {} })
    return probe.length === 0
  }
  const pruefe = (w, f, pfad, rot, extra, zaehler) => {
    const ort = pfad === '' ? '' : pfad
    switch (f.art) {
      case 'unbekannt': zaehler.zaehle(); return
      case 'null': if (w !== null) rot.push(ort + ': ' + art(w) + ' statt null'); return
      case 'string': case 'number': case 'boolean':
        if (typeof w !== f.art) rot.push(ort + ': ' + art(w) + ' statt ' + f.art); return
      case 'literal': if (w !== f.wert) rot.push(ort + ': ' + JSON.stringify(w) + ' statt ' + JSON.stringify(f.wert)); return
      case 'liste':
        if (!Array.isArray(w)) { rot.push(ort + ': ' + art(w) + ' statt liste'); return }
        w.forEach((e, i) => pruefe(e, f.von, ort + '[' + i + ']', rot, extra, zaehler)); return
      case 'abbildung':
        if (w === null || typeof w !== 'object' || Array.isArray(w)) { rot.push(ort + ': ' + art(w) + ' statt abbildung'); return }
        for (const [k, e] of Object.entries(w)) pruefe(e, f.von, (ort ? ort + '.' : '') + k, rot, extra, zaehler)
        return
      case 'oder':
        if (!f.von.some((m) => passt(w, m))) {
          const einfach = f.von.filter((m) => m.art !== 'null')
          if (einfach.length === 1 && w !== null) { pruefe(w, einfach[0], pfad, rot, extra, zaehler); return }
          rot.push(ort + ': ' + (typeof w === 'string' ? JSON.stringify(w) : art(w)) + ' passt zu keiner Möglichkeit')
        } else {
          const getroffen = f.von.find((m) => passt(w, m))
          pruefe(w, getroffen, pfad, rot, extra, zaehler)
        }
        return
      case 'objekt': {
        if (w === null || typeof w !== 'object' || Array.isArray(w)) { rot.push(ort + ': ' + art(w) + ' statt objekt'); return }
        for (const [name, feld] of Object.entries(f.felder)) {
          const unter = (ort ? ort + '.' : '') + name
          if (!(name in w) || w[name] === undefined) { if (!feld.optional) rot.push(unter + ': fehlt'); continue }
          pruefe(w[name], feld.form, unter, rot, extra, zaehler)
        }
        for (const name of Object.keys(w)) if (!(name in f.felder)) extra.push((ort ? ort + '.' : '') + name)
        return
      }
      default: rot.push(ort + ': unbekannte Formart ' + f.art)
    }
  }
  pruefe(wert, form, '', abweichungen, zusaetzlich, { zaehle: () => { unbekannt++ } })
  return { abweichungen, zusaetzlich, unbekannt }
}`

// eslint-disable-next-line no-new-func -- dieselbe Funktion wie in der Seite
export const pruefeForm = new Function(`${pruefeFormQuelle}\nreturn pruefeForm`)()
