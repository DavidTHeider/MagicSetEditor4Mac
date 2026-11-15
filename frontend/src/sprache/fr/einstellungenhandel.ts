// **The trading limits** — work item `T1` of phase 5.6
// (#1466).
//
// A separate file, for the same reason as `einstellungennetz.ts` next to it:
// `seiteneinstellungen.ts` would cross the 400-line limit (§2). It stays
// **one** catalogue — the parts are spread in there.

// handelsnamen are the labels of the five limits.
export const handelsnamen = {
  'richtlinie.handel': 'Autoriser les échanges',
  'richtlinie.handel.markt':
    'Autoriser les échanges via un fournisseur externe',
  'richtlinie.handel.menge': 'Quantité maximale par ligne',
  'richtlinie.handel.wert':
    'Valeur maximale d’une offre dans la plus petite unité (centimes pour EUR)',
  'richtlinie.handel.frist': 'Durée de réservation maximale en minutes',
}

// handelshilfe is the explaining sentence per limit.
export const handelshilfe = {
  'richtlinie.handel':
    'Désactive d’un coup le partage, le prêt de decks, l’échange et le marché. **Une limite, pas une préférence** : ce qui est désactivé ici ne peut être réactivé dans aucun profil.',
  'richtlinie.handel.markt':
    'Par défaut : **désactivé**. Le marché s’adresse à un fournisseur externe et touche à l’argent — tout le reste reste entre deux personnes.',
  'richtlinie.handel.menge':
    'Le nombre maximal d’exemplaires qu’une seule ligne d’une offre peut mentionner. **Vide signifie aucune limite.**',
  'richtlinie.handel.wert':
    'La valeur totale maximale d’une offre. S’applique uniquement lorsque l’évaluation **fait partie du contrat** — une simple vue du marché ne la limite pas. **Vide signifie aucune limite.**',
  'richtlinie.handel.frist':
    'La durée pendant laquelle un engagement peut réserver des exemplaires. **Vide signifie aucune limite.**',
}

// handelswerte are the two positions of the switches.
export const handelswerte = {
  'richtlinie.handel.aus': 'Désactivé',
  'richtlinie.handel.an': 'Activé',
  'richtlinie.handel.markt.aus': 'Désactivé',
  'richtlinie.handel.markt.an': 'Activé',
}
