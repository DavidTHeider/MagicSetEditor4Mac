// Frame, navigation, language selection.
export const allgemein = {
  // **Count words** (#1218) — a number and its word in the matching form.
  zaehlwoerter: {
    karten: { eine: 'карта', wenige: 'карты', viele: 'карт' },

    // **Und die Seiten** (#1219):
    // der Druckbogen meldete »1 Karten auf 1 Seiten«.
    seiten: { eine: 'страница', wenige: 'страницы', viele: 'страниц' },

    // And the three objects of the overview (#1422).
    sets: { eine: 'сет', wenige: 'сета', viele: 'сетов' },
    decks: { eine: 'колода', wenige: 'колоды', viele: 'колод' },
    pakete: { eine: 'пакет', wenige: 'пакета', viele: 'пакетов' },
    eigenePakete: {
      eine: 'собственный пакет',
      wenige: 'собственных пакета',
      viele: 'собственных пакетов',
    },
    stile: { eine: 'стиль', wenige: 'стиля', viele: 'стилей' },
    positionen: { eine: 'позиция', wenige: 'позиции', viele: 'позиций' },
    exemplare: {
      eine: 'экземпляр',
      wenige: 'экземпляра',
      viele: 'экземпляров',
    },

    // **Und die Drucke einer Sichtfreigabe** (UXR-F18, #2033).
    drucke: {
      eine: 'печать',
      wenige: 'печати',
      viele: 'печатей',
    },
  },
  bestandGesichert: 'Копия базы данных сохранена в',
  zahlenbereich: 'Число от {kleinstens} до {groesstens}.',
  zahlenAb: 'Число от {kleinstens} и больше.',
  ganzeZahlBereich: 'Целое число от {kleinstens} до {groesstens}.',
  ganzeZahlAb: 'Целое число от {kleinstens} и больше.',
  nurZiffern:
    'Только цифры, без точки и запятой — сумма указывается в наименьшей единице.',

  // **Ein Doppelklick im Finder, der nicht aufging** (#891).
  dateiNichtGeoeffnet: 'Не удалось открыть файл {pfad}. {grund}',
  anwendungsname: 'MagicSetEditor4Mac',
  sprache: {
    bezeichnung: 'Язык',
    de: 'Deutsch',
    en: 'English',
    es: 'Español',
    fr: 'Français',
    it: 'Italiano',
    ja: '日本語',
    ru: 'Русский',
    zh: '简体中文',
  },
  // **Die fünf Arbeitsbereiche, der Einstieg und die Dienstzone** — UX-02
  // (#1674).
  //
  // **`bereiche` und `dienste` sind die zwei Überschriften der Leiste**, nicht
  // Bereiche selbst — sie sagen, was darunter steht.
  bereiche: {
    // **„Home" und nicht „Start"** — so heißt die Seite in dieser Sprache
    // (`en/seitenstart.ts`), und der Bereich trägt ihren Namen.
    start: 'Главная',
    gestalten: 'Дизайн',
    sammeln: 'Коллекция',
    spielen: 'Игра',
    handeln: 'Обмен',
    ressourcen: 'Ресурсы',
    programm: 'Программа',

    ueberschrift: 'Рабочие области',
  },

  navigation: {
    bezeichnung: 'Основная навигация',
    sprungZumInhalt: 'Перейти к содержимому',
    seitenleisteEinblenden: 'Показать боковую панель',
    seitenleisteAusblenden: 'Скрыть боковую панель',
    seitenleisteKuerzel: '⌥⌘S',
    weg: 'Ты здесь',
  },

  // When something does not draw — F01 (#2340).
  fehlergrenze: {
    bereichTitel: 'Не удалось отрисовать эту область',
    bereichSatz:
      'Номер {kennung}. Остальная база данных не затронута; ошибка записана в журнал под этим номером.',
    neuLaden: 'Перезагрузить область',
    diagnoseOeffnen: 'Открыть диагностику',
    wurzelTitel: 'Не удалось отрисовать окно',
    wurzelSatz:
      'Номер {kennung}. Твоя база данных не затронута; ошибка записана в журнал под этим номером.',
    fensterNeuLaden: 'Перезагрузить окно',
  },
  meldungen: {
    methodeFehlt: 'Метод {name} не подключён.',
    vorschauLeer: 'Предпросмотр вернулся без изображения и без ошибки.',
    kontoLeer: 'Учётная запись вернулась без пользователя и без ошибки.',
  },
  // Das Band über dem Inhalt: was zu keiner Seite gehört (#180).
  meldungsband: {
    bereich: 'Сообщения окна',
    setNichtGesichert: 'Сет «{name}» не сохранён — {grund}',
    setNichtGesichertOhneName: 'Сет не сохранён — {grund}',
    bestandNichtGesichert: 'Копия базы данных не создана — {grund}',
    verlaufGescheitert: 'Не удалось отменить или повторить действие — {grund}',
    darstellungGescheitert: 'Не удалось сбросить оформление — {grund}',
  },
  // Die neue Fassung, wenn eine da ist (#1100).
  vertraeglichkeit: {
    vertraeglich: 'Эта версия может открыть твою библиотеку.',
    unvertraeglich: 'Эта версия не может открыть твою библиотеку.',
    ungeklaert:
      'Эта версия не сообщает, может ли она открыть твою библиотеку: она старше этой проверки. Сделай резервную копию библиотеки перед переходом и восстанови её, если что-то пойдёт не так.',
    bereit: 'Переход подготовлен: копия создана и прочитана обратно.',
    vorbereiten: 'Подготовить переход',
    hindernisse: {
      keinManifest: 'Выпуск ничего не сообщает о себе.',
      plattform: 'Версия собрана для другого компьютера.',
      wanderungFehlt:
        'В твоей библиотеке есть изменение, неизвестное этой версии.',
      summeAnders: 'Одно из изменений имеет в ней другое содержимое.',
      vorDerGrundlinie:
        'Твоя библиотека старше её базовой линии — сначала открой её в промежуточной версии.',
      platz: 'Недостаточно свободного места.',
      vorgaengeLaufen: 'Что-то ещё выполняется.',
      keineSicherung: 'Резервная копия не создана.',
      sicherungUnlesbar: 'Не удалось прочитать резервную копию обратно.',
      halter:
        'Другая программа держит библиотеку открытой — закрой её перед переходом.',
    },
  },
  fassungsband: {
    bereich: 'Новая версия',
    neu: 'Вышла версия {fassung}.',
    seite: 'Что изменилось',
    laden: 'Загрузить',
    laeuft: 'Загрузка …',
    fortschritt: 'Загружено {gelesen} из {ganz}',
    fortschrittOhneLaenge: 'Загружено: {gelesen}',
    geladen:
      'Версия {fassung} находится в {pfad}. Чтобы установить её, открой файл в Finder и перетащи приложение в папку «Программы».',
    weglegen: 'Закрыть это сообщение',
  },
  wiederanlauf: {
    bereich: 'Последний сеанс',
    satz: 'Последний сеанс, начатый {zeit}, завершился неожиданно.',
    satzOhneZeit: 'Последний сеанс завершился неожиданно.',
    mitSpur:
      'Есть след сбоя; диагностический пакет включит его, когда ты его создашь.',
    ohneSpur: 'Следа сбоя нет — вероятно, приложение было завершено извне.',
    ausblenden: 'Скрыть',
  },

  // Der Arbeitsplatz eines Sets: seine Laschen und sein leerer Zustand (#288).
  arbeitsplatz: {
    leiste: {
      bezeichnung: 'Показатели набора',
      befund: 'С замечаниями',
      nichtGeprueft: 'не проверено',
      nichtGelesen: 'не прочитано',
      geprueft: 'проверено {wann}',
    },
    laschen: 'Разделы сета',

    gruppen: {
      bauen: 'Создание карт',
    },
    schubladen: {
      eigenschaften: {
        knopf: 'Свойства',
        titel: 'Свойства сета',
        laschen: 'Разделы свойств',
        hinweis: 'То, что относится ко всем картам этого сета.',
      },
      werkzeuge: {
        knopf: 'Инструменты',
        titel: 'Инструменты',
        laschen: 'Инструменты сета',
        hinweis: 'Статистика, случайный бустер и консоль для этого сета.',
      },
      schliessen: 'Закрыть',
    },

    keinSet: 'Нет открытого сета',
    keinSetErklaerung:
      'Выбери сет в разделе «Сеты», создай новый или импортируй существующий — здесь появятся его карты, статистика и лист печати.',
    zuDenSets: 'Перейти к сетам',
    zuDenPaketen: 'Перейти к пакетам',
    zuDenKarten: 'Перейти к картам',
    ohneKartenTitel: 'В этом сете пока нет карт',
    ohneKarten:
      'Эта вкладка работает с картами сета. Карты появляются в разделе «Карты» — созданные, вставленные или импортированные.',
    ohneNamen: 'Открытый сет',
  },
  palette: {
    bezeichnung: 'Палитра команд',
    eingabe: 'Куда перейти?',
    keineTreffer: 'Ничего не найдено. Попробовать другое слово?',
    oeffnen: 'Палитра команд',
    seiten: 'Страницы',
    befehle: 'Команды',
    befehlEingabe: 'Куда перейти или какую команду выполнить?',
    tun: 'Выполнить',
    finden: 'Найти',
    karteErstellen: 'Создать карту',
    setWaehlen: 'Создать или открыть сет',
    deckSpielen: 'Быстрая игра с {name}',
    tauschenMit: 'Обменяться картами с {name}',
    sicherungErstellen: 'Создать резервную копию',
    schnittstellenEinrichten:
      'Настроить интерфейсы (командная строка, MCP, REST)',
    woKarte: 'Дизайн · {set}',
    woSets: 'Дизайн · Сеты',
    woPartie: 'Игра · Начать партию',
    woHandel: 'Обмен · Партнёр',
    woEinstellungen: 'Настройки · {gruppe}',
    alsKarte: 'Карта · {set}',
    alsSet: 'Сет · {karten}',
    alsDeck: 'Колода · {spiel}',
    alsGegenueber: 'Партнёр · подтверждён',
    fuss: '↑↓ выбрать · ↵ выполнить · esc закрыть',
  },
} as const
