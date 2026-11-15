export const dateiansicht = {
  ueberschrift: 'Il file',
  laedt: 'Analisi del file in corso …',
  arten: {
    paket: 'Pacchetto MSE',
    bild: 'Immagine',
    vektor: 'Immagine vettoriale',
    pdf: 'PDF',
    text: 'Testo',
    archiv: 'Archivio',
    unbekannt: 'Tipo sconosciuto',
  },
  name: 'Nome',
  art: 'Tipo',
  groesse: 'Dimensione',
  geaendert: 'Modificato',
  ort: 'Posizione',
  masse: '{breite} × {hoehe} punti',
  seiten: '{anzahl} pagine',
  eintraege: '{anzahl} elementi',
  gekuerzt: 'Viene mostrato l’inizio — il file continua.',
  unsicher:
    '{anzahl} elementi portano fuori dal file o sono collegamenti. Vengono contati e mai aperti.',
  unsicherEintrag: 'non aperto',
  verschluesselt:
    'Un backup protetto da password. Il suo contenuto si vede solo quando lo ripristini.',
  zuGross: 'Troppo grande per un’anteprima. Aprilo con la sua app.',
  unlesbar:
    'Il contenuto non può essere letto. Nome, dimensione e ora sono corretti.',
  keineVorschau:
    'Per questo tipo non c’è nessuna anteprima: potrebbe essere un programma. Mostralo nel Finder se ti serve.',
  imFinder: 'Mostra nel Finder',
  oeffnen: 'Apri',
  imFinderFuer: 'Mostra {name} nel Finder',
  oeffnenFuer: 'Apri {name} con la sua app',
  ansehen: 'Visualizza',
  ansehenFuer: 'Visualizza {name}',
  zuklappen: 'Chiudi la visualizzazione',
} as const
