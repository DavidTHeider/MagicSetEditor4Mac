// Die Einstellungen der KI — die lokale Laufzeit, Apples Gerätemodell und
// die Adresse (A7, #1446; AI-05, #2042). **Aus `seiteneinstellungen.ts`
// herausgelöst**, als der Katalog über die Zeilengrenze aus §2 ging.

export const kinamen = {
  'ki.laufzeit': 'Environnement d’exécution IA local',
  'ki.geraetemodell': 'Modèle sur l’appareil d’Apple',
  'ki.vorrangmodell': 'Modèle Ollama pour la priorité',
  'ki.adresse': 'Adresse de l’environnement d’exécution IA',
} as const

export const kihinweise = {
  'ki.laufzeit':
    'Indique si un environnement d’exécution IA est sollicité. **Désactivé** signifie qu’aucun appel ne part — pas même vers cet ordinateur. Rien n’est installé ni lancé d’ici.',
  'ki.geraetemodell':
    'Indique si le modèle sur l’appareil de macOS est sollicité pour les petites tâches de texte — nom de carte, texte d’ambiance. **Automatique** : s’il est activé sur ce Mac (Apple Intelligence active, licence de `fm` acceptée), il est intégré et sollicité en premier ; les données ne quittent pas l’ordinateur. **Désactivé** vaut jusqu’à ce que tu le changes ; le désactiver pour une session se fait sous **IA**.',
  'ki.vorrangmodell':
    'Le modèle Ollama qu’interroge une suggestion avec **priorité**, après le modèle sur l’appareil. **Vide** signifie le premier capable de répondre sous une forme fixe, et Ollama cite d’abord le dernier téléchargé. Un modèle choisi qu’Ollama ne cite plus est signalé, et non remplacé en silence par un autre.',
  'ki.adresse':
    'L’endroit où l’environnement écoute. Vide signifie l’adresse par défaut sur cet ordinateur. Une adresse qui **n’est pas** sur cet ordinateur classe tout modèle qui se trouve derrière comme non local — avec « Travailler uniquement en local », rien n’est demandé du tout.',
} as const

export const kiwerte = {
  'ki.laufzeit.aus': 'Désactivé',
  'ki.laufzeit.an': 'Activé',
  'ki.geraetemodell.automatisch': 'Automatique',
  'ki.geraetemodell.aus': 'Désactivé',
} as const
