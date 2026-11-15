// The mapping of ⌘0 to ⌘9 — UX-02.AK-03 (#1674), split out like the German file.
export const einstellungsbelegung = {
  belegungUeberschrift: 'Assegnazione da ⌘0 a ⌘9',
  belegungHinweis:
    'Quale pagina si apre con una cifra insieme a ⌘. Per impostazione predefinita sono le aree della barra laterale; “Come prima della 0.6.0” ripristina l’assegnazione fino alla 0.5.0, la cui tabella si trova nel registro delle modifiche.',
  belegungFuer: 'Pagina per {kuerzel}',
  belegungFrei: '—',
  belegungVor06: 'Come prima della 0.6.0',
  belegungVorgabe: 'Ripristina il predefinito',
  kuerzel: {
    palette: 'Apri la palette dei comandi',
    suche: 'Cerca nel set',
    zurueck: 'Annulla una modifica',
    vor: 'Ripristina una modifica',
    sichern: 'Salva il set',
    sichernUnter: 'Salva il set come …',
    oeffnen: 'Apri un set …',
    neu: 'Crea un nuovo set',
    schliessen: 'Chiudi il set',
    drucken: 'Vai al foglio di stampa',
    navigator: 'Mostra o nascondi il navigatore',
    inspector: 'Mostra o nascondi l’ispettore',

    // Back and forward between pages — UX-03.AK-02 (#1675).
    seiteZurueck: 'Pagina precedente',
    seiteVor: 'Pagina successiva',
  },
} as const
