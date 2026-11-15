// **The measurement settings** — OPS-05
// (#1722).
export const messnamen = {
  'messung.spur': 'Trace',
  'messung.abtastung': 'Part des appels enregistrés, en pourcentage',
  'messung.spurgrenze': 'Au plus ce nombre d’entrées',
  'messung.aufbewahrung': 'Les entrées restent ce nombre de minutes',
  'messung.profile': 'Profils de performance conservés',
  'zwischenspeicher.grenze': 'Limite par cache sur disque en Mo',
} as const

export const messwerte = {
  'messung.spur.aus': 'Désactivée',
  'messung.spur.an': 'Activée',
} as const

export const messhinweise = {
  'messung.spur':
    'Une ligne par appel au service et au serveur d’outils — avec identifiant d’opération, chemin, durée et issue, sans contenu. Elle réside dans la mémoire du processus et s’applique dès le prochain démarrage.',
  'messung.abtastung':
    'À 10, la trace enregistre un appel sur dix — compté exactement, pas au hasard. Vide signifie 100.',
  'messung.spurgrenze':
    'Quand la trace est pleine, l’entrée la plus ancienne disparaît. De 10 à 10000, vide signifie 1000.',
  'messung.aufbewahrung':
    'Les entrées plus anciennes sont supprimées. De 1 à 1440 minutes, vide signifie 60.',
  'messung.profile':
    'Les plus récents restent à côté de la base ; les plus anciens disparaissent avec le prochain profil. De 1 à 20, vide signifie 3.',
  'zwischenspeicher.grenze':
    'Pour les réponses et les images de Scryfall, chacune séparément. Au démarrage, les plus anciennes disparaissent jusqu’à ce que cela tienne. De 1 à 8192, vide signifie 512.',
} as const
