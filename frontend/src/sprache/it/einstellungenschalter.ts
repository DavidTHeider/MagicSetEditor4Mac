// **Bereiche und Funktionen** — OPS-10
// (#1908).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge; die Namen und Werte werden dort hineingestreut.
// Ein Bereich heißt hier wie sein Eintrag in der Leiste.
export const seitenschalter = {
  schalter: {
    erklaerung:
      'Disattivata, un’area scompare dalla barra laterale e dalla palette, e i suoi percorsi vengono rifiutati su ogni superficie. I suoi dati restano: riattivarla ripristina tutto.',
    hinweis:
      'Un interruttore nasconde, non protegge: per proteggere qualcosa, usa ruoli e condivisioni.',
    reife: {
      stabil: 'stabile',
      vorschau: 'Anteprima',
      auslaufend: 'in dismissione',
    },
    saetze: {
      sammeln: 'Collezione, possesso e liste dei desideri',
      spielen: 'Mazzi, tornei, partite e il tavolo',
      handeln: 'Scambi, condivisioni e buste',
      ressourcen: 'Pacchetti di modelli e modelli',
      ki: 'Suggerimenti, immagini, indice di ricerca, account e costi',
      netz: 'Identità, contatti, inviti e raggiungibilità',
      markt: 'Offerte, prezzi e account dei marketplace',
      turniere: 'Tornei con turni, abbinamenti e risultati',
      dienst:
        'REST per script: un servizio in esecuzione si ferma alla disattivazione',
      werkzeugserver:
        'MCP via HTTP: un server in esecuzione si ferma alla disattivazione',
    },
    abgeschaltet: {
      titel: '{seite} è disattivata',
      erklaerung:
        'Riattiva l’area in Impostazioni › Aree e funzioni. I suoi dati sono ancora lì, invariati.',
      oeffnen: 'Apri aree e funzioni',
    },
  },
} as const

// schalternamen sind die Beschriftungen der zehn Schalter.
export const schalternamen = {
  'schalter.sammeln': 'Colleziona',
  'schalter.spielen': 'Gioca',
  'schalter.handeln': 'Scambia',
  'schalter.ressourcen': 'Risorse',
  'schalter.ki': 'IA',
  'schalter.netz': 'Rete e controparti',
  'schalter.markt': 'Marketplace e prezzi',
  'schalter.turniere': 'Tornei',
  'schalter.dienst': 'Servizio per script',
  'schalter.werkzeugserver': 'Server di strumenti via HTTP',
}

// schalterwerte sind an und aus, je Schalter.
export const schalterwerte = {
  'schalter.sammeln.an': 'attiva',
  'schalter.sammeln.aus': 'disattiva',
  'schalter.spielen.an': 'attiva',
  'schalter.spielen.aus': 'disattiva',
  'schalter.handeln.an': 'attiva',
  'schalter.handeln.aus': 'disattiva',
  'schalter.ressourcen.an': 'attiva',
  'schalter.ressourcen.aus': 'disattiva',
  'schalter.ki.an': 'attiva',
  'schalter.ki.aus': 'disattiva',
  'schalter.netz.an': 'attiva',
  'schalter.netz.aus': 'disattiva',
  'schalter.markt.an': 'attiva',
  'schalter.markt.aus': 'disattiva',
  'schalter.turniere.an': 'attiva',
  'schalter.turniere.aus': 'disattiva',
  'schalter.dienst.an': 'attiva',
  'schalter.dienst.aus': 'disattiva',
  'schalter.werkzeugserver.an': 'attiva',
  'schalter.werkzeugserver.aus': 'disattiva',
}
