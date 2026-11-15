// The bulk-change preview — work package U6 of phase 5.1 (#1417). See the
// German file for why it lives apart from `seitenkarten.ts`.
export const massenvorschauEn = {
  vorschau: 'Anteprima',
  vorschauErklaerung:
    'L’anteprima mostra, per ogni carta, che cosa cambierebbe. Non viene scritto nulla finché non scegli “Applica” — e solo se nel frattempo nessuno ha scritto.',
  karte: 'Carta',
  alt: 'Prima',
  neu: 'Dopo',
  keinWertBisher: 'nessun valore',
  bleibtGleich: 'resta invariato — è già così',
  zusammenfassung: 'Carte da modificare: {geaendert}, invariate: {gleich}.',
  nichtsZuTun:
    'Tutte le carte selezionate hanno già questo valore. Non c’è nulla da modificare.',
  konflikt:
    'Queste carte sono cambiate dopo l’anteprima. Non è stato scritto nulla — ricarica l’anteprima:',
  inzwischenGeaendert: 'modificata nel frattempo',
  angewendet: 'Carte modificate: {anzahl}.',
} as const
