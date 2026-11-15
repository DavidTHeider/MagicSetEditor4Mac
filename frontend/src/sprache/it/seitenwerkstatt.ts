// The card workspace: workspace, quick filter, zoom, overlays and the problem list —
// part of the `karten` catalog, split off because `seitenkarten.ts` sits at the line limit.
export const werkstatt = {
  arbeitsplatz: 'Area di lavoro delle carte',
  inspectorBezeichnung: 'Campi e controlli',
  // **Marks on cards** (#1208).
  schnellfilter: {
    bezeichnung: 'Filtra le carte',
    platzhalter: 'Filtra — ad es. type:creature -goblin',
    leeren: 'Cancella il filtro',
    hilfe:
      'Gli spazi separano le parti, «…» le tiene unite, campo:valore lega a un campo, un - iniziale nega. Maiuscole e minuscole non contano.',
    treffer: '{gezeigt} di {gesamt} carte',
  },
  zoom: {
    bezeichnung: 'Dimensione dell’anteprima',
    anpassen: 'Adatta',
    druck: 'Dimensione di stampa',
    ohneDruck:
      'Il modello non indica una risoluzione, quindi non c’è una dimensione di stampa.',
  },
  // **Das Kartenbild in den Finder ziehen** — AP-12 (#2369).
  ziehen: {
    beschriftung: 'Trascina l’immagine',
    hinweis:
      'Trascina l’immagine nel Finder per salvare la carta come PNG. {weg} fa lo stesso.',
  },
  overlays: {
    bezeichnung: 'Sovrapposizioni',
    bildgrenzen: 'Bordi delle immagini',
    beschnitt: 'Abbondanza',
    diagnose: 'Diagnostica',
    beschnittHinweis:
      'Ciò che sta dentro la linea sopravvive a un taglio di 3 mm.',
  },
  probleme: {
    ueberschrift: 'Problemi',
    keine: 'Nessun problema su questa carta.',
    zumFeld: 'Vai al campo {feld}',
    ohneFeld: 'nessun campo',
  },
  // **Paging** (#1290) — ported from MSE's Card menu.
  blaettern: {
    vorige: 'Carta precedente',
    naechste: 'Carta successiva',
    stelle: '{stelle} di {gesamt}',
    kuerzelVorige: '⇞',
    kuerzelNaechste: '⇟',
  },
}
