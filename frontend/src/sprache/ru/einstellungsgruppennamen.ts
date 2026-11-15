// **Die neun Gruppen der Einstellungsseite** — UX-04.AK-01
// (#1676).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge. Ein Katalog darf sich auf mehrere Dateien
// verteilen, solange er als **einer** ankommt (§9).
export const einstellungsgruppennamen = {
  gruppen: {
    allgemein: 'Основные',
    bereiche: 'Разделы и функции',
    erscheinungsbild: 'Внешний вид',
    nutzer: 'Пользователь и профиль',
    fach: 'Предметные функции',
    ki: 'ИИ',
    verbindungen: 'Подключения',
    dienste: 'Сервисы',
    daten: 'Данные и резервные копии',
    diagnose: 'Диагностика',
  },

  untertitel: {
    allgemein: 'Язык, запуск, сочетания клавиш',
    bereiche: 'Рабочие разделы, функции, зрелость',
    erscheinungsbild: 'Тема, шрифт, окно',
    nutzer: 'Вход, разблокировка',
    fach: 'Карты, коллекция, игра',
    ki: 'Провайдеры, модели, расходы',
    verbindungen: 'Сеть, удалённые стороны, CLI, MCP, REST',
    dienste: 'Сервисы карт, цены, рынок',
    daten: 'Хранилище, пути, резервные копии',
    diagnose: 'Журналы, измерения, технические данные',
  },

  abschnitte: {
    fenster: 'Окно и боковая панель',
    spalten: 'Колонки и представления',
    handel: 'Ограничения обмена',
    protokoll: 'Журнал',
    messung: 'Измерение',
    speicher: 'Кеши и поисковый индекс',
    adressen: 'Адреса',
    ordner: 'Папки и имена файлов',
    eingriffe: 'Возврат, порядок, сброс',
    mitnehmen: 'Перенести настройки',
  },
} as const
