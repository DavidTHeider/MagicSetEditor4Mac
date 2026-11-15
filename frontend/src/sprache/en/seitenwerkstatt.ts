// The card workspace: workspace, quick filter, zoom, overlays and the problem list —
// part of the `karten` catalog, split off because `seitenkarten.ts` sits at the line limit.
export const werkstatt = {
  arbeitsplatz: 'Card workspace',
  inspectorBezeichnung: 'Fields and checks',
  // **Marks on cards** (#1208).
  schnellfilter: {
    bezeichnung: 'Filter cards',
    platzhalter: 'Filter — e.g. type:creature -goblin',
    leeren: 'Clear filter',
    hilfe:
      'Whitespace separates parts, "…" keeps them together, field:value binds to a field, a leading - negates. Case does not matter.',
    treffer: '{gezeigt} of {gesamt} cards',
  },
  zoom: {
    bezeichnung: 'Preview size',
    anpassen: 'Fit',
    druck: 'Print size',
    ohneDruck: 'The template names no resolution, so there is no print size.',
  },
  // **Das Kartenbild in den Finder ziehen** — AP-12 (#2369).
  ziehen: {
    beschriftung: 'Drag image',
    hinweis:
      'Drag the image to the Finder to save the card as a PNG. {weg} does the same.',
  },
  overlays: {
    bezeichnung: 'Overlays',
    bildgrenzen: 'Image bounds',
    beschnitt: 'Bleed',
    diagnose: 'Diagnostics',
    beschnittHinweis: 'What is inside the line survives a 3 mm cut.',
  },
  probleme: {
    ueberschrift: 'Problems',
    keine: 'No problems on this card.',
    zumFeld: 'Go to field {feld}',
    ohneFeld: 'no field',
  },
  // **Paging** (#1290) — ported from MSE's Card menu.
  blaettern: {
    vorige: 'Previous card',
    naechste: 'Next card',
    stelle: '{stelle} of {gesamt}',
    kuerzelVorige: '⇞',
    kuerzelNaechste: '⇟',
  },
}
