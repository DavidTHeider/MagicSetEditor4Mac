// The log settings — OPS-04 (#1721).
import { messhinweise, messnamen, messwerte } from './einstellungenmessung'
import { netzwerte } from './einstellungennetz'

export const protokollnamen = {
  'protokoll.stufe': 'Ce que le journal consigne',
  'protokoll.datei': 'Fichier journal',
  'protokoll.system': 'Journal unifié de macOS',
  'protokoll.ausgabe': 'Sortie d’erreur standard',
  'protokoll.format': 'Format des lignes',
  'protokoll.grenze': 'Taille d’un fichier journal en Mo',
  'protokoll.aufbewahrung': 'Nombre de fichiers plus anciens conservés',
  'protokoll.komponenten': 'Détaillé uniquement pour',
  'protokoll.ort': 'Dossier du fichier journal',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messnamen,
} as const

export const protokollwerte = {
  'protokoll.stufe.fehler': 'Erreurs uniquement',
  'protokoll.stufe.warnung': 'Avertissements et erreurs',
  'protokoll.stufe.info': 'Événements (par défaut)',
  'protokoll.stufe.ausfuehrlich': 'Détaillé — chaque étape',
  'protokoll.datei.an': 'Activé',
  'protokoll.datei.aus': 'Désactivé',
  'protokoll.system.an': 'Activé',
  'protokoll.system.aus': 'Désactivé',
  'protokoll.ausgabe.aus': 'Désactivé',
  'protokoll.ausgabe.an': 'Activé',
  'protokoll.format.text': 'Texte lisible',
  'protokoll.format.json': 'JSON pour les outils',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messwerte,
  // **Und die Kontingente der Dienste** — QA-08.AK-03 (#1739): ihre Namen
  // stehen beim Netz, die Werte kommen hier dazu, weil `seiteneinstellungen.ts`
  // an der Zeilengrenze aus §2 steht.
  ...netzwerte,
} as const

export const protokollhinweise = {
  'protokoll.stufe':
    'Ce qui entre dans le journal. **Détaillé** aide à trouver une erreur — il ne conserve alors qu’un fichier de 1 Mo au plus et un fichier plus ancien, et le diagnostic indique qu’il est activé. S’applique au prochain démarrage.',
  'protokoll.datei':
    'Le fichier `mse.log` que tu joins à un rapport d’erreur. S’il ne peut pas être créé, les autres canaux continuent d’écrire, et le diagnostic en indique la raison.',
  'protokoll.system':
    'Le journal du système — à consulter dans l’app Console de macOS sous « MagicSetEditor4Mac ».',
  'protokoll.ausgabe':
    'Pour un lancement depuis un terminal : chaque ligne aussi à l’endroit où le terminal l’affiche.',
  'protokoll.format':
    '**JSON** est une ligne par événement avec des champs nommés — pour les outils qui collectent les journaux.',
  'protokoll.grenze':
    'À partir de cette taille, un nouveau fichier commence au prochain démarrage. Vide signifie 4 Mo ; les valeurs de 1 à 64 sont autorisées.',
  'protokoll.aufbewahrung':
    'Ce nombre de fichiers plus anciens reste à côté du fichier courant, de `mse.log.1` jusqu’au plus ancien. Vide signifie un ; les valeurs de 1 à 9 sont autorisées.',
  'protokoll.komponenten':
    'Uniquement avec **détaillé** : les étapes de quelles parties sont consignées, par exemple `app` pour les appels de la fenêtre. Vide signifie toutes ; les événements et les erreurs sont toujours consignés.',
  'protokoll.ort':
    'Un dossier commençant par / ou ~/. Vide signifie à côté de la base de données.',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messhinweise,
} as const
