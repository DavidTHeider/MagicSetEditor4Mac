// The card workspace: workspace, quick filter, zoom, overlays and the problem list —
// part of the `karten` catalog, split off because `seitenkarten.ts` sits at the line limit.
export const werkstatt = {
  arbeitsplatz: 'Espacio de trabajo de cartas',
  inspectorBezeichnung: 'Campos y comprobaciones',
  // **Marks on cards** (#1208).
  schnellfilter: {
    bezeichnung: 'Filtrar cartas',
    platzhalter: 'Filtrar — p. ej. type:creature -goblin',
    leeren: 'Borrar el filtro',
    hilfe:
      'Los espacios separan las partes, «…» las mantiene juntas, campo:valor se limita a un campo y un - inicial niega. Las mayúsculas y minúsculas no importan.',
    treffer: '{gezeigt} de {gesamt} cartas',
  },
  zoom: {
    bezeichnung: 'Tamaño de la vista previa',
    anpassen: 'Ajustar',
    druck: 'Tamaño de impresión',
    ohneDruck:
      'La plantilla no indica ninguna resolución, así que no hay tamaño de impresión.',
  },
  // **Das Kartenbild in den Finder ziehen** — AP-12 (#2369).
  ziehen: {
    beschriftung: 'Arrastrar imagen',
    hinweis:
      'Arrastra la imagen al Finder para guardar la carta como PNG. {weg} hace lo mismo.',
  },
  overlays: {
    bezeichnung: 'Superposiciones',
    bildgrenzen: 'Límites de imagen',
    beschnitt: 'Sangrado',
    diagnose: 'Diagnóstico',
    beschnittHinweis:
      'Lo que queda dentro de la línea sobrevive a un corte de 3 mm.',
  },
  probleme: {
    ueberschrift: 'Problemas',
    keine: 'Ningún problema en esta carta.',
    zumFeld: 'Ir al campo {feld}',
    ohneFeld: 'sin campo',
  },
  // **Paging** (#1290) — ported from MSE's Card menu.
  blaettern: {
    vorige: 'Carta anterior',
    naechste: 'Carta siguiente',
    stelle: '{stelle} de {gesamt}',
    kuerzelVorige: '⇞',
    kuerzelNaechste: '⇟',
  },
}
