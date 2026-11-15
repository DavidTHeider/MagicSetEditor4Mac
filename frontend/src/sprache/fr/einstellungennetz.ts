// **The network settings** — work item N5 of phase 5.4
// (#1571).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge — dieselbe Gestalt wie `einstellungenzugaenge.ts`
// daneben. **Ein Katalog bleibt es trotzdem**: die drei Teile werden dort
// hineingestreut, und der Vertragstest gegen die Go-Seite sieht einen.

// netznamen sind die Beschriftungen der sechs Einstellungen.
export const netznamen = {
  'netz.proxy.http': 'Proxy pour http',
  'netz.proxy.https': 'Proxy pour https',
  'netz.proxy.aus': 'Accéder sans proxy',
  'netz.frist.verbindung': 'Délai : connexion (s)',
  'netz.frist.kopf': 'Délai : en-tête de réponse (s)',
  'netz.frist.stillstand': 'Délai : absence de progression (s)',
  'netz.schnittstelle': 'Interface pour la recherche sur le réseau local',
  'netz.lauschen': 'Où écoutent les services propres',
  'dienst.tls.zertifikat': 'Certificat propre des services',
  'dienst.tls.schluessel': 'Clé du certificat propre',
  'dienst.tls.kette': 'Certificats intermédiaires',
  'dienst.tls.anker': 'Ancres de confiance du certificat',
  'grenzen.kontingente': 'Quotas des services',
}

// netzhilfe ist der erklärende Satz je Einstellung.
export const netzhilfe = {
  'netz.proxy.http':
    'Le proxy par lequel passent les adresses `http://`. **Vide signifie que `HTTP_PROXY` de l’environnement s’applique** — toute valeur saisie ici a priorité. Avec le schéma, par exemple `http://proxy.entreprise:3128`. **Sans mot de passe** — il serait stocké en clair sur le disque ; un proxy avec authentification est repris de `HTTPS_PROXY` dans l’environnement.',
  'netz.proxy.https':
    'La même chose pour les adresses `https://`, à partir de `HTTPS_PROXY`. **Séparé**, car de nombreux réseaux n’ont un proxy que pour l’un des deux.',
  'netz.proxy.aus':
    'Ce qui ne passe **pas** par le proxy — noms d’hôtes et plages d’adresses, séparés par des virgules, comme `NO_PROXY`. L’adresse de bouclage reste de toute façon à l’écart ; un Ollama sur une **autre** machine de ton réseau doit figurer ici.',
  'netz.frist.verbindung':
    'La durée maximale d’une connexion, TLS compris. Vide signifie 15 secondes.',
  'netz.frist.kopf':
    'Le temps dont dispose un serveur pour envoyer son en-tête de réponse. Vide signifie 30 secondes.',
  'netz.frist.stillstand':
    'La durée pendant laquelle une réponse en cours peut rester sans progresser. **C’est celui-ci qu’il faut augmenter** lorsqu’un gros téléchargement sur une liaison lente s’interrompt — pas les deux précédents. Vide signifie 60 secondes.',
  'netz.schnittstelle':
    'Interface réseau par laquelle passent les requêtes de la recherche sur le réseau local et l’annonce — par exemple `en0`. **Vide signifie automatique** : toute interface adaptée avec IPv4 et multicast. Un nom qui n’existe pas pour le moment est traité comme vide et consigné comme constat dans le journal.',
  'netz.lauschen':
    'Adresses sur lesquelles REST et MCP écoutent, séparées par des virgules. **Vide signifie : ce Mac uniquement** (`127.0.0.1`). `0.0.0.0` signifie toutes les interfaces et reste seule. **Si une adresse quitte cet ordinateur, TLS et jeton s’appliquent** ; le choix est plus simple sous « Outils sans fenêtre ». Prend effet au prochain démarrage d’un service.',
  'dienst.tls.zertifikat':
    'Le certificat que les services REST et MCP présentent en dehors de l’adresse de bouclage, au format PEM. **Vide signifie : le certificat lié à l’identité de ce profil.** Avec ta propre autorité de certification, les clients vérifient le nom et la chaîne au lieu de l’empreinte. S’applique au prochain démarrage du service.',
  'dienst.tls.schluessel':
    'La clé privée du certificat, au format PEM. **Lisible uniquement par toi (0600)**, sinon le service ne démarre pas. Le remplacement des deux fichiers prend effet sans redémarrage à la prochaine poignée de main.',
  'dienst.tls.kette':
    'Certificats intermédiaires, s’ils ne figurent pas déjà dans le certificat. Vide signifie : aucun.',
  'dienst.tls.anker':
    'La racine de ton autorité, par rapport à laquelle la chaîne est vérifiée avant le démarrage. Vide signifie : les ancres du système.',
  'grenzen.kontingente':
    'Ce qu’un accès machine peut faire par minute via REST et MCP — par accès et pour tous ensemble. **Désactivés, c’est pour un environnement contrôlé** où personne ne prend rien à personne, comme un essai de charge. L’accès, TLS, les places par type de travail, les limites du corps et de la réponse, les délais et le blocage contre les tentatives de devinette restent. Prend effet au prochain démarrage d’un service.',
}

// netzbereich ist die Überschrift des Bereichs samt Hinweis.
export const netzbereich = {
  netz: 'Comment ce programme sort sur le réseau',
  netzHinweis:
    'Ces réglages s’appliquent à **tous** les services à la fois — y compris la ligne de commande, le serveur MCP et le service. Vide signifie que ce que l’environnement indique s’applique. Ce qui s’applique actuellement, et d’où cela vient, est affiché sous Réseau et connectivité.',
}

// netzwerte sind die Namen der Werte — QA-08.AK-03 (#1739).
export const netzwerte = {
  'grenzen.kontingente.an': 'Activés',
  'grenzen.kontingente.aus': 'Désactivés',
}
