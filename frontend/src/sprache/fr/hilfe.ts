import { betrieb } from './hilfebetrieb'
import { suche } from './hilfesuche'

export const hilfe = {
  betrieb,
  suche,
  // One help entry per page — the keys come from `navigation.ts`.
  seiten: {
    start: {
      ueberschrift: 'Accueil',
      text: 'Le point d’entrée. En haut, les chiffres de la bibliothèque — sets, cartes, paquets, exemplaires — puis, dessous, ce sur quoi tu as travaillé en dernier : ouvert, importé, exporté, et ce que tu gardes en haut. Ensuite ce qui demande de l’attention — sets avec des réserves, decks avec des cartes manquantes, paquets avec des constats de script — et ce qui est récupéré en arrière-plan. Les trois premières étapes s’affichent dans une fenêtre au premier lancement ; les visites guidées de l’aide les montrent à nouveau.',
    },
    sets: {
      ueberschrift: 'Sets',
      text: 'Tous les sets que cette application connaît. D’ici, on en ouvre un, on en crée un nouveau ou on importe un fichier `.mse-set`. Le set ouvert détermine ce que montrent ses onglets et les tiroirs « Propriétés » et « Outils ».',
    },
    pakete: {
      ueberschrift: 'Paquets',
      text: 'Jeux, feuilles de style, polices de symboles et modèles d’export. Ils sont récupérés ici — à une révision fixe et avec une somme de contrôle — puis installés. Sans eux, une carte n’a ni champs ni apparence.',
    },
    vorlagen: {
      ueberschrift: 'Modèles',
      text: 'Les fichiers texte d’un paquet de modèles — champs, scripts, dimensions. On peut les modifier ici sans éditeur de texte à côté. Une modification vaut pour tous les sets qui utilisent ce paquet ; l’état précédent reste comme sauvegarde à côté du paquet.',
    },
    sammlung: {
      ueberschrift: 'Collection',
      text: 'Ce que tu possèdes réellement — une ligne par impression, état, langue et finition. Elle n’a rien à voir avec tes sets : une carte conçue est un brouillon, un exemplaire est un morceau de carton. Les prix sont récupérés sur demande et sont des estimations dont personne ne répond.',
    },
    decks: {
      ueberschrift: 'Decks',
      text: 'Assemblages de tes propres cartes et d’impressions réelles. Un deck renvoie aux cartes au lieu de les copier, et appartient à un jeu — sans cela, il n’y aurait pas de format. Un format est ici un enregistrement et non une liste intégrée : il est créé, pas livré. La vérification interroge Scryfall pour les impressions réelles ; une carte que tu as conçue toi-même reste non vérifiée, ce qui ne veut pas dire illégale.',
    },
    turniere: {
      ueberschrift: 'Tournois',
      text: 'Un tournoi est un déroulement dans le temps : rondes, appariements, résultats. Le classement en est calculé et n’est stocké nulle part — corrige un résultat et le classement suit aussitôt. L’appariement suit le système suisse ; les quatre critères de départage et leurs chiffres viennent des Magic Tournament Rules. Inscrire un deck remet une liste : elle est copiée, et les modifications ultérieures du deck ne l’atteignent plus.',
    },
    spielen: {
      ueberschrift: 'Jouer',
      text: 'Une partie est une table avec des decks figés : ce qu’un deck devient au début de la partie ne change plus si quelqu’un le modifie ensuite. Ce qui se passe à la table est un flux d’événements — chacun dépend du précédent, et aucun n’est écrasé. Ce que tu vois est toujours la vue d’**un** siège : une bibliothèque est face cachée même pour son propriétaire, et qui la consulte doit ensuite la mélanger. L’hôte détient l’état complet ; la page le dit, car il faut le savoir avant de rejoindre une partie.',
    },
    handeln: {
      ueberschrift: 'Échanges',
      text: 'On échange avec un correspondant dont tu as comparé l’empreinte — sans compte et sans intermédiaire. La concordance est calculée sur **cette** machine : rien n’en sort, et ce qu’un correspondant voit de toi, c’est toi qui le décides sous « Réglages › Connexions ». L’enregistrement n’a lieu que lorsque les cartes ont réellement changé de mains ; ta bibliothèque change alors, et le journal note ce qui est allé où. **Comptabilisé chez toi ne veut pas dire comptabilisé des deux côtés** — deux bibliothèques sur deux machines ne forment pas une écriture commune, et l’état le dit. La valeur de ta collection figure à côté comme une observation, non comme une promesse.',
    },
    hilfe: {
      ueberschrift: 'Aide',
      text: 'Les termes issus de Magic Set Editor et que personne n’apporte avec soi, chacun en une phrase. Plus les visites guidées, les raccourcis clavier, pourquoi certains libellés restent en anglais et ce que cette application envoie vers l’extérieur.',
    },
    nutzer: {
      ueberschrift: 'Utilisateurs et profils',
      text: 'À qui appartient ce qui se trouve ici. Un utilisateur est la personne ; un profil est un environnement de travail de cette personne — ses propres sets, sa propre collection, ses propres réglages. Qui veut collectionner séparément pour un tournoi n’a pas besoin d’un second compte, mais d’un second profil.',
    },
    einstellungen: {
      ueberschrift: 'Réglages',
      text: 'La langue, l’apparence et les endroits où l’on cherche les paquets. Ce qui figure ici vaut pour tout le programme, non pour un set en particulier.',
    },
    karten: {
      ueberschrift: 'Cartes',
      text: 'Les cartes du set ouvert : à gauche la liste, au milieu l’aperçu, à droite les champs de la carte ouverte. Une modification d’un champ se voit aussitôt ; rien n’a besoin d’être confirmé.',
    },
    setangaben: {
      ueberschrift: 'Informations du set',
      text: 'Ce qui vaut pour tout le set : nom, description, symbole du set, auteur. Ces informations sont à la disposition de chaque carte — une feuille de style peut dessiner le symbole sur chacune.',
    },
    stil: {
      ueberschrift: 'Style',
      text: 'La feuille de style du set et ses options. La feuille dit à quoi ressemble une carte ; les options la règlent — par exemple sur un autre cadre ou une autre édition.',
    },
    suche: {
      ueberschrift: 'Recherche',
      text: 'Trouver des cartes d’après les valeurs de leurs champs, dans tous les sets ou dans un seul. La recherche porte sur le texte intégral, et un résultat mène directement à la carte.',
    },
    ersetzen: {
      ueberschrift: 'Remplacer',
      text: 'Remplacer un mot dans les champs de texte d’un set. Ce qui va changer est d’abord montré ; un ⌘Z annule toute l’opération.',
    },
    statistik: {
      ueberschrift: 'Statistiques',
      text: 'Ce que contient le set ouvert : combien de cartes, à quelle fréquence un champ est rempli, comment se répartit une valeur de champ. Pour vérifier qu’une série est complète.',
    },
    bilder: {
      ueberschrift: 'Images',
      text: 'Écrit les cartes du set sous forme de fichiers image — toutes d’un coup, avec un modèle de nom et une règle pour les noms identiques. **Attribuer** une image à une carte ne se fait pas ici, mais dans le champ image de la carte.',
    },
    druckbogen: {
      ueberschrift: 'Feuille d’impression',
      text: 'Les cartes côte à côte sur papier, en PDF avec des traits de coupe. Pensée pour l’imprimante et non pour être transmise — l’export sert à cela.',
    },
    schluesselwoerter: {
      ueberschrift: 'Mots-clés',
      text: 'Un mot-clé reconnaît son texte sur chaque carte et y ajoute le texte de rappel. Ceux du jeu sont fixes et figurent ici pour consultation ; les tiens ne valent que pour ce set et peuvent être ajoutés, modifiés et supprimés.',
    },
    pack: {
      ueberschrift: 'Pack aléatoire',
      text: 'Un booster tiré des cartes de ce set, selon les règles du paquet de jeu — les types de packs existants y sont indiqués. La même graine donne le même tirage ; pour montrer un résultat, transmets le nombre.',
    },
    symbol: {
      ueberschrift: 'Symbole',
      text: 'Le symbole de ce set — dessiné, non chargé. Une forme de base est le point de départ ; ensuite, chaque point peut être déplacé. La touche Tab passe d’un point à l’autre, les flèches les déplacent. Ce qui est dessiné ici apparaît aussitôt dans l’aperçu et entre dans le fichier à l’enregistrement du set.',
    },
    konsole: {
      ueberschrift: 'Console',
      text: 'Une expression MSE Script dans le contexte de la carte ouverte — le même environnement que pour le dessin. Elle ne modifie rien ; elle calcule et renvoie. Le résultat arrive sous forme de texte source, pour distinguer une chaîne d’un nom.',
    },
    ausgeben: {
      ueberschrift: 'Exporter',
      text: 'Cartes sous forme d’image, ou de texte via un modèle d’export. Le modèle provient d’un paquet et détermine ce qui en sort : une liste pour un forum, un tableau, un fichier pour un autre programme.',
    },
  },

  // **Über** (#1168).
  ueber: {
    ueberschrift: 'À propos de MagicSetEditor4Mac',
    fassung: 'Version',
    lizenzName: 'Licence',
    lizenzSatz:
      'Le code source de cette application est placé sous la GNU General Public License, version 2 ou ultérieure.',
    rechteName: 'Contenus de tiers',
    rechteSatz:
      'La licence couvre le code source, **pas** les cadres de cartes, les symboles de mana, les polices, les textes de règles, les noms et les marques. Ils appartiennent à leurs ayants droit — dont Wizards of the Coast — et non à ce projet.',
    quelleName: 'Code source',
    quelleSatz: 'github.com/DavidTHeider/MagicSetEditor4Mac',
  },

  weiterlesen: {
    ueberschrift: 'Pour aller plus loin',
    einleitung:
      'Deux documents font partie des sources de cette application — dans le dossier du projet, sous `docs/`, pas dans l’image disque que tu as installée. Cette application ne peut pas les ouvrir elle-même : elle n’a aucun chemin vers l’extérieur, et un lien qui ne mène nulle part serait pire que rien.',
    handbuchName: 'Manuel',
    handbuchSatz:
      'Ce qui figure ici en mots-clés, en phrases complètes : le premier lancement, ' +
      'la création d’une carte, l’export, les raccourcis — et que faire quand une carte ' +
      'reste vide. `docs/HANDBUCH.md`',
    administrationName: 'Données et sauvegardes',
    administrationSatz:
      'Où se trouvent les données, comment les sauvegarder et les restaurer, ce que ' +
      'contient le journal et ce qui se passe lors d’un changement de version — y compris ' +
      'la phrase selon laquelle une sauvegarde **sans mot de passe** est en clair. `docs/ADMINISTRATION.md`',
  },

  erststart: {
    ueberschrift: 'Au premier lancement',
    text:
      'Cette application n’est pas notariée. macOS la bloque une fois ; ' +
      'tu l’autorises dans Réglages Système, sous « Confidentialité et sécurité ».',
  },

  // Tours, and the areas that have no page (#1422).
  bereiche: {
    ueberschrift: 'Et un espace sans page dans la barre latérale',
    einleitung:
      'Tout ce que cette application sait faire ne figure pas dans la barre latérale. Ce qui vient de l’extérieur n’est pas un lieu — et n’a donc pas de page.',
    kiName: 'IA',
    kiSatz:
      'Un modèle de langage lit et modifie des cartes via le **serveur MCP** — un processus distinct sur la même bibliothèque, que tu enregistres dans ton outil d’IA. Ce qui sort ainsi et ce qui ne sort pas figure sous **Interfaces**.',
    kiKnopf: 'Aller aux interfaces',
  },
  fuehrungen: {
    ueberschrift: 'Visites guidées',
    einleitung:
      'Une visite guidée te fait parcourir l’application pas à pas et peut être répétée à tout moment. À côté, une phrase pour chacun explique comment l’application est organisée.',
    startName: 'Les premières étapes',
    startSatz:
      'Trois étapes jusqu’à la première carte : ouvrir un set, modifier des cartes, récupérer des modèles. Elles s’affichent d’elles-mêmes au premier lancement ; leur retour au prochain lancement est un réglage.',
    startZeigen: 'Afficher les premières étapes maintenant',
    arbeitsplatzName: 'L’espace de travail',
    arbeitsplatzSatz:
      'Cartes, modèles, collection et decks partagent une même organisation : panneau de navigation à gauche, inspecteur à droite, le centre entre les deux. **⌥⌘[** replie le panneau de navigation, **⌥⌘]** l’inspecteur — et ce que tu replies reste replié par profil. **⌥⌘S** replie et déplie la barre latérale pour que sa largeur revienne aux espaces de travail — cet état appartient à cet appareil et est enregistré par profil.',
    paletteName: 'La palette de commandes',
    paletteSatz:
      '**⌘K** l’ouvre au-dessus de chaque page : pages, sets et commandes, trouvés avec quelques lettres.',
    leisteName: 'La barre de contexte',
    leisteSatz:
      'Sous le nom de l’application, elle indique qui travaille où et en tant que quoi — personne, profil, rôle. Le changement se fait sur la page **Utilisateurs et profils** ; la barre y mène.',
  },

  begriffe: {
    set: {
      name: 'Set',
      satz: 'Une collection de cartes avec ses propres informations — nom, description, symbole. Il se trouve dans un fichier portant l’extension `.mse-set`.',
    },
    karte: {
      name: 'Carte',
      satz: 'L’objet dont il s’agit. Une carte porte des valeurs de champs — nom, coût, texte de règles — et est dessinée d’après une feuille de style.',
    },
    spiel: {
      name: 'Jeu',
      satz: 'Dit **quels champs** une carte possède. Un `.mse-game` les déclare : `magic` connaît le coût de mana et la rareté, un autre jeu en connaît d’autres.',
    },
    stil: {
      name: 'Feuille de style',
      satz: 'Dit **comment** une carte est dessinée : quel calque se place où, quelle image donne le cadre. Un `.mse-style` appartient toujours à un jeu.',
    },
    stiloption: {
      name: 'Option de style',
      satz: 'Un réglage de la feuille de style qui vaut pour tout le set ou pour une carte — par exemple le cadre d’une carte de draft ou une autre édition.',
    },
    symbolschrift: {
      name: 'Police de symboles',
      satz: 'Dessine les symboles de mana et autres. Ce n’est pas une police de caractères mais une collection de petites images avec des règles indiquant quel texte affiche laquelle.',
    },
    paket: {
      name: 'Paquet',
      satz: 'Le terme générique pour jeu, feuille de style, police de symboles et modèle d’export. Les paquets sont récupérés et installés ; sans eux, une carte reste vide.',
    },
    exportvorlage: {
      name: 'Modèle d’export',
      satz: 'Un petit programme qui transforme les cartes d’un set en **texte** — une liste pour un forum, un tableau, un fichier pour un autre programme.',
    },
    druckbogen: {
      name: 'Feuille d’impression',
      satz: 'Un PDF avec les cartes côte à côte sur papier, avec traits de coupe. Pour imprimer et découper — pas pour être transmis comme fichier.',
    },
    vorbehalt: {
      name: 'Réserve',
      satz: 'Une ligne d’un fichier ouvert que ce programme n’a pas interprétée. Elle n’est pas perdue : elle voyage inchangée et réapparaît à l’enregistrement.',
    },
    befund: {
      name: 'Constat',
      satz: 'Une indication sur quelque chose qui n’a pas fonctionné pendant le dessin — un calque manquant, un caractère sans symbole. La carte est dessinée malgré tout.',
    },
  },

  sprache: {
    ueberschrift: 'Pourquoi certains textes restent en anglais',
    einleitung:
      'Le libellé d’un champ de carte et la phrase qui le suit viennent du **paquet de jeu**, non de cette application. Ils sont tels que l’auteur du paquet les a écrits — généralement en anglais.',
    schluesselName: 'Le nom est la clé',
    schluesselSatz:
      'La valeur est enregistrée sous ce nom dans ton fichier `.mse-set`. Une traduction inventée serait plus jolie mais rendrait le fichier illisible pour quiconque y cherche quelque chose — c’est pourquoi cette application n’en invente aucune.',
    uebersetztName: 'Ce qui est traduit est traduit',
    uebersetztSatz:
      'Si un paquet fournit une traduction, elle est utilisée. Dans le `M15-Magic-Pack` téléchargé, cela concernait 25 champs de carte au 2026-08-21. Pour les 224 autres, il n’en fournissait aucune, et le mot est alors le même que dans Magic Set Editor 2.',
    fensterName: 'La fenêtre elle-même parle huit langues',
    fensterSatz:
      'Tout ce que cette application dit d’elle-même — boutons, messages, cette page — suit la langue des réglages. L’allemand est la langue source ; les sept autres sont traduites automatiquement et pas encore relues — le choix de la langue l’indique aussi.',
    leisteName: 'La barre des menus aussi',
    leisteSatz:
      'Le menu de l’application et les menus d’édition et de fenêtre emploient les mots que macOS utilise dans cette langue. **Couper**, **Copier** et **Coller** gardent leur lien avec le système.',
  },

  // **What this application knows about you** — Werkstück 13 of Phase 4.12
  // (#1351).
  datenschutz: {
    ueberschrift: 'Ce que cette application sait de toi',
    einleitung:
      'Depuis qu’il existe des utilisateurs et des profils, cette application enregistre des **données personnelles**. Voici lesquelles, où elles se trouvent, ce qui quitte cette machine — et ce qui ne la quitte expressément pas.',

    wasName: 'Ce qui est enregistré',
    wasSatz:
      'Tout est **facultatif** sauf le nom affiché. Pour la personne : **civilité**, **suffixe de nom**, **prénom**, **autres prénoms**, **nom de famille**, **slogan** et deux **biographies**, une courte et une longue. Pour le compte : **adresse**, **seconde adresse**, **numéro de téléphone** et, si tu en définis un, un **mot de passe**. Plus une **photo de profil** et un **badge** par profil.',

    woName: 'Où cela se trouve',
    woSatz:
      'Dans ta bibliothèque — un fichier SQLite dans ton dossier personnel — et dans des images à côté. **Trois choses se trouvent ailleurs**, à savoir dans le **trousseau** de macOS : le secret de ton second facteur, le secret d’une connexion chez un fournisseur, et les identifiants de services de cartes tiers. La bibliothèque se transmet, le trousseau non — d’où la séparation.',

    formName: 'Sous quelle forme',
    formSatz:
      'Un **mot de passe** n’est jamais stocké en clair : il est protégé avec **Argon2id**, et la forme protégée porte ses propres paramètres. Un **jeton d’accès** est stocké sous forme de somme de contrôle ; le texte en clair n’existe qu’une seule fois, juste après sa création. Tout le reste est stocké tel que tu l’as saisi — c’est ton fichier. **Et le fichier lui-même peut être chiffré** — sous **Réglages › Données et sauvegarde › Base de données**. Plus rien n’est alors lisible sur le disque ; la clé va dans le trousseau de ce Mac et **ne peut pas être récupérée**. Une **sauvegarde** reçoit son propre **mot de passe** — elle est alors chiffrée et s’ouvre pourtant sur n’importe quel ordinateur ; sans mot de passe, une base chiffrée ne la livre que sur demande explicite.',

    ausName: 'Ce qui quitte cette machine',
    ausSatz:
      'Lors d’une **connexion**, ton choix part vers le fournisseur que tu as choisi : `github.com` et `api.github.com`, `accounts.google.com` et `oauth2.googleapis.com`, `login.microsoftonline.com`, ou `twitter.com` et `api.twitter.com`. Seuls ton **nom et ton identifiant** chez lui sont demandés — **pas** ton adresse e-mail, ni tes contacts, ni tes publications. Trois choses reviennent et sont enregistrées : le fournisseur, ton identifiant chez lui et ton nom affiché chez lui.',

    nichtName: 'Ce qui ne sort pas',
    nichtSatz:
      'Les **éléments de ton nom, biographies, numéro de téléphone et adresse** ne quittent jamais cette machine. Ils se trouvent dans ta bibliothèque et ne sont envoyés nulle part. Il n’y a **aucune télémétrie**, aucun rapport de plantage, aucune statistique d’utilisation et aucun compte chez nous — il n’y a pas de « nous ».',

    dienstName: 'Et les services de cartes',
    dienstSatz:
      'La recherche de cartes interroge `api.scryfall.com` et — s’ils sont activés — `api.magicthegathering.io`, `graphql.mtgjson.com` ou `api.scrydex.com`. Seule la **requête de recherche** part, rien d’autre : ni nom d’utilisateur, ni identifiant, rien de ton profil. **Les vignettes des résultats**, la fenêtre les charge elle-même — depuis `cards.scryfall.io` et, avec magicthegathering.io, depuis `gatherer.wizards.com` et `gatherer-static.wizards.com` ; seule l’adresse de l’image y part. Les prix de toute la collection viennent — s’il est activé — de `mtgjson.com`, les offres du marché de `api.cardtrader.com` ; ce qui part, c’est l’identifiant de l’impression et, pour CardTrader, ton jeton d’accès, rien d’autre. Les paquets de modèles arrivent par `codeload.github.com`, les avis de version par `api.github.com`. Ce que chaque service peut faire ici et à quelles conditions figure sous **Réglages › Services › Services de cartes tiers** ; les liens qui s’y trouvent mènent à `scryfall.com`, `docs.magicthegathering.io`, `mtgjson.com`, `scrydex.com` et `www.cardtrader.com` — l’application n’appelle jamais elle-même ces pages, elles ne s’ouvrent dans ton navigateur qu’à ton clic. **Et les liens de marché sur une carte** (T8) : ils mènent à `www.cardmarket.com`, `www.cardtrader.com`, `www.tcgplayer.com`, `www.cardkingdom.com`, `gatherer.wizards.com` et `www.mtggoldfish.com` — **construits, pas appelés** : l’adresse est assemblée sur cette machine à partir du nom de la carte ou d’un identifiant de fournisseur, et elle ne part qu’à ton clic, dans le navigateur de ton choix. Aucune clé n’est nécessaire pour cela, et aucune ne figure dans ces adresses.',

    kiName: 'Et l’IA',
    kiSatz:
      'Un modèle **local** tourne sur cette machine et ne la quitte pas : l’application parle à `127.0.0.1` et à rien d’autre. Un **modèle cloud** est autre chose — qui active Anthropic, OpenAI ou xAI envoie à `api.anthropic.com`, `api.openai.com` ou `api.x.ai` la **tâche et les champs de la carte** concernée, ainsi que ta clé du trousseau. **Chez OpenAI, rien n’y reste** : l’interface conserve une réponse par défaut, et cette application le désactive explicitement. **De même chez xAI** : la voie qui conserve les réponses 30 jours n’est jamais utilisée. Rien de ton profil, aucune seconde carte, aucun set. Les deux sont **désactivés** tant que tu ne les actives pas, et sous **Réglages › Services › Travailler en local uniquement**, tout modèle cloud reste désactivé même si son propre interrupteur dit autre chose. Ce qui revient est **montré et non adopté** — c’est une personne qui confirme ce qui entre dans un champ. Ce qu’Anthropic peut faire ici et à quelles conditions figure sous **Réglages › Services › Services de cartes tiers** ; les liens qui s’y trouvent mènent à `www.anthropic.com`, `platform.claude.com`, `openai.com`, `developers.openai.com`, `x.ai` et `docs.x.ai` — l’application n’appelle jamais elle-même ces pages, elles ne s’ouvrent dans ton navigateur qu’à ton clic.',

    auskunftName: 'Accès et suppression',
    auskunftSatz:
      'Sous **Utilisateurs et profils › Mon compte**, tu exportes tes données sous forme de fichier — tout ce qui figure ici, en une seule fois. Et sous **Utilisateurs et profils**, tu supprimes un utilisateur — avec ses profils, son image et le secret de son second facteur. **Ce qui se trouve dans ses profils reste** : sets, decks, collection et tournois n’appartiennent alors plus à aucun profil et ne sont plus accessibles depuis l’interface. Ni l’un ni l’autre ne demande de réseau ni l’accord de quiconque.',
  },

  schnittstellen: {
    ueberschrift: 'Ce qui sort et ce qui entre',
    einleitung:
      'Cette application fonctionne hors ligne. **Vers le réseau**, elle ne va que si quelqu’un le déclenche : récupérer un modèle, une recherche de cartes et une connexion demandent un clic ; elle ne cherche de nouvelles versions que si un réglage l’autorise, et celui-ci est **désactivé** par défaut. Les adresses contactées et ce qui part figurent sous **Ce que cette application sait de toi**. À côté, il existe trois voies vers cet ordinateur que personne n’emprunte non plus de lui-même — et une qui s’exécute d’elle-même : si le démarrage échoue, elle affiche la raison dans une fenêtre.',
    bezugName: 'Récupérer des modèles',
    bezugSatz:
      'Sur clic, l’application récupère un paquet de modèles par HTTP, à une révision fixe et avec une somme de contrôle. Sans clic, il ne se passe rien.',
    scryfallName: 'Données de cartes de Scryfall',
    scryfallSatz:
      'Sur clic, l’application interroge `api.scryfall.com` pour une carte, un set ou une image de carte. Ce qui en revient est **montré et non adopté** — ce qui entre dans un champ est une décision humaine. Sans réseau, le dernier bon état est conservé, et il indique son ancienneté.',
    fassungName: 'Rechercher de nouvelles versions',
    fassungSatz:
      'Le seul endroit qui peut interroger sans clic — et **uniquement** s’il est activé sous **Réglages**. Le réglage par défaut est **désactivé**, et alors aucune requête ne part. Activé, il interroge `api.github.com` pour la dernière version publiée. Le téléchargement n’a lieu que sur clic, et **la signature et la somme de contrôle sont alors vérifiées avant qu’un fichier soit placé** ; rien n’est installé de soi-même.',
    mcpName: 'Le serveur MCP',
    mcpSatz:
      'Un programme à part nommé `mse-mcp`, avec lequel un modèle de langage peut travailler sur un set : lire cartes et champs, remplir des champs, créer, copier, supprimer et réordonner des cartes, importer un set, enregistrer un set et dessiner une carte. La mesure n’est pas la quantité mais la réversibilité : aucun outil n’écrase ce qui existe — là où un fichier existe déjà, il refuse — et chaque modification dit comment l’annuler. Lancé par un client comme Claude Desktop, il parle par l’entrée et la sortie standard : ni port, ni adresse. **En HTTP**, il n’écoute que lorsque tu le démarres dans **Réglages › Connexions**, et seulement là où l’indique **Où écoutent les services**.',
    mcpNichtAn:
      'Il ne tourne pas en même temps, mais il **est fourni** : dans le bundle, sous `MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp`, avec `mse`, la ligne de commande. **D’elle-même, cette application ne dépose rien hors de son bundle.** Dans **Réglages › Connexions › Outils sans fenêtre**, tu installes d’un clic ce que tu veux — le lien vers `mse` dans `~/.local/bin` avec l’entrée `PATH` dans `~/.zprofile`, l’entrée pour Codex, Claude Desktop ou Claude Code — et tu y retires aussi chaque entrée ; ce qui se trouve hors de son propre bloc, elle n’y touche pas. Ce que les outils modifient, cette fenêtre le remarque d’elle-même : elle demande l’état de la base toutes les deux secondes et recharge la page concernée.',
    dienstName: 'Le service pour scripts',
    dienstSatz:
      'Un programme à part nommé `mse-rest`, par lequel un script atteint les mêmes capacités que `mse` — lire sets et cartes, remplir un champ, importer ou enregistrer un set, dessiner une carte. Chaque requête exige un jeton. **Où il écoute, c’est toi qui décides** dans **Où écoutent les services** : par défaut, ce Mac seulement (`127.0.0.1`) ; si tu choisis une adresse qui quitte ce Mac, il exige TLS et un accès.',
    dienstNichtAn:
      'Il se trouve dans le bundle, sous `MagicSetEditor4Mac.app/Contents/MacOS/mse-rest`, et ne démarre pas de lui-même : qui ne veut pas de service n’en a pas. On le démarre dans **Réglages › Connexions** — ou comme LaunchAgent dans **Outils sans fenêtre** ; macOS le lance alors à chaque ouverture de session, jusqu’à ce que tu le retires. Qui le démarre reçoit l’adresse et le jeton. Les requêtes d’un navigateur sont refusées avant même que le jeton soit vérifié.',
    startfensterName: 'Le message en cas d’échec du démarrage',
    startfensterSatz:
      'La seule voie que l’application emprunte **d’elle-même**. Si la bibliothèque ne peut pas être ouverte, elle appelle `osascript` de macOS pour afficher la raison dans une fenêtre — sinon elle ne figurerait que sur un canal que personne ne voit après un double-clic. Le texte y entre comme valeur, non comme script.',
    dateienName: 'Fichiers sur cet ordinateur',
    dateienSatz:
      'Sets, paquets de modèles, images et PDF écrits. Ils se trouvent là où tu les places ; la bibliothèque se trouve dans ton dossier utilisateur.',
    sonstSatz:
      'Rien d’autre. **Aucune télémétrie**, aucun rapport de plantage, aucune statistique d’utilisation. Un compte existe — mais chez le fournisseur que tu choisis, pas chez nous ; ce qui part alors figure sous **Ce que cette application sait de toi**.',
    grenzenName: 'Limites',
    grenzenSatz:
      'Chaque accès machine dispose d’un **crédit**, et chaque appel coûte selon son type ; quand il est épuisé, la réponse le dit avec le temps d’attente. Les travaux lourds — rastérisation, import, export, opérations en masse — disposent d’un nombre fixe de places par classe, que la fenêtre, la ligne de commande, MCP et le service se partagent. Qui devine ne bloque que son propre justificatif.',
    grenzenZahlen:
      'Aujourd’hui : {burst} unités d’un coup, {nachfuellen} par minute ; une lecture coûte {lesen}, une écriture {schreiben}, un import {einfuhr} ; au plus {rendern} rastérisations à la fois ; {reserve} unités restent à la fenêtre.',
    grenzenOhneKontingente:
      '**Les quotas sont désactivés** (`grenzen.kontingente`) — pour un environnement contrôlé. Les places par type de travail, les limites du corps et de la réponse, les délais, le blocage contre les tentatives de devinette, l’accès et TLS restent en vigueur.',
  },
} as const
