// The bulk-change preview — work package U6 of phase 5.1 (#1417). See the
// German file for why it lives apart from `seitenkarten.ts`.
export const massenvorschauEn = {
  vorschau: 'Aperçu',
  vorschauErklaerung:
    'L’aperçu montre, carte par carte, ce qui changerait. Rien n’est écrit avant « Appliquer » — et seulement si personne n’a écrit entre-temps.',
  karte: 'Carte',
  alt: 'Avant',
  neu: 'Après',
  keinWertBisher: 'aucune valeur',
  bleibtGleich: 'inchangé — déjà défini ainsi',
  zusammenfassung:
    '{geaendert} carte(s) vont changer, {gleich} restent telles quelles.',
  nichtsZuTun:
    'Toutes les cartes sélectionnées ont déjà cette valeur. Il n’y a rien à modifier.',
  konflikt:
    'Ces cartes ont changé depuis l’aperçu. Rien n’a été écrit — récupère de nouveau l’aperçu :',
  inzwischenGeaendert: 'modifiée depuis',
  angewendet: '{anzahl} carte(s) modifiée(s).',
} as const
