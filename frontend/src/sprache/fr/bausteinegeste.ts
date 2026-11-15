// Die Ansagen einer Geste — GESTE (#2358); Ausgangssprache ist Deutsch.
export const geste = {
  gegriffen:
    'Saisi. Cible : {ziel} ({stelle} sur {von}). Les flèches choisissent la cible, Entrée dépose, Échap annule.',
  ziel: 'Cible : {ziel} ({stelle} sur {von}).',
  abgelegt: 'Déposé : {ziel}.',
  abgebrochen: 'Annulé. Rien n’a été déplacé.',
  keinZiel: 'Il n’y a pour l’instant aucune cible pour cet objet.',
  zielFehlt: 'Cette cible n’existe plus — rien n’a été déplacé.',
  veraltet: 'L’état a changé entre-temps — rien n’a été déplacé.',
} as const
