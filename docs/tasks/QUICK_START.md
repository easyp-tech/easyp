# Быстрый старт миграции документации

## 🚀 Шаг 1: Создание базовой инфраструктуры (30 минут)

### 1.1 Создать роуты для документации

```tsx
// src/routes/DocsRoutes.tsx
import { Routes, Route } from 'react-router-dom'
import DocsLayout from '../components/docs/DocsLayout'
import MarkdownPage from '../components/docs/MarkdownPage'

export default function DocsRoutes() {
    return (
        <Routes>
            <Route path="/docs" element={<DocsLayout />}>
                <Route index element={<MarkdownPage path="introduction/what-is" />} />
                <Route path="guide/:category/:page" element={<MarkdownPage />} />
                <Route path="guide/:category/:subcategory/:page" element={<MarkdownPage />} />
            </Route>
        </Routes>
    )
}
```

### 1.2 Добавить роуты в App.tsx

```tsx
// src/App.tsx
import DocsRoutes from './routes/DocsRoutes'

// В компоненте App добавить:
<Route path="/docs/*" element={<DocsRoutes />} />
```

## 🎨 Шаг 2: Создание компонентов документации (45 минут)

### 2.1 DocsLayout компонент

```tsx
// src/components/docs/DocsLayout.tsx
import { Outlet } from 'react-router-dom'
import DocsSidebar from './DocsSidebar'

export default function DocsLayout() {
    return (
        <div className="flex min-h-screen">
            <DocsSidebar />
            <main className="flex-1 p-8">
                <Outlet />
            </main>
        </div>
    )
}
```

### 2.2 MarkdownPage компонент

```tsx
// src/components/docs/MarkdownPage.tsx
import { useParams } from 'react-router-dom'
import { useState, useEffect } from 'react'
import { MarkdownContent } from '../../lib/markdown'

export default function MarkdownPage({ path }: { path?: string }) {
    const params = useParams()
    const [content, setContent] = useState('')
    
    const mdPath = path || `${params.category}/${params.page}`
    
    useEffect(() => {
        // Загрузка markdown файла
        fetch(`/docs/guide/${mdPath}.md`)
            .then(res => res.text())
            .then(setContent)
    }, [mdPath])
    
    return <MarkdownContent content={content} />
}
```

## 📝 Шаг 3: Тестирование с первым документом (15 минут)

### 3.1 Подготовить тестовый файл

Скопировать `public/docs/guide/introduction/what-is.md` и проверить:
- Рендеринг HTML блоков
- Работу ссылок
- Отображение кода

### 3.2 Добавить стили

```tsx
// src/main.tsx или src/index.css
import '@/lib/markdown/styles.css'
```

### 3.3 Запустить и проверить

```bash
npm run dev
# Открыть http://localhost:5173/docs
```

## ✅ Чек-лист первого запуска

- [ ] Страница `/docs` открывается без ошибок
- [ ] Markdown контент отображается
- [ ] Custom blocks (tip, warning) рендерятся правильно
- [ ] Код подсвечивается
- [ ] Ссылки кликабельны

## 🔧 Следующие шаги

После успешного запуска базовой версии:

1. **Добавить навигацию:**
   - Создать `sidebar.json` с структурой документации
   - Реализовать компонент `DocsSidebar`

2. **Улучшить загрузку контента:**
   - Использовать динамический import для .md файлов
   - Добавить loading state
   - Обработать 404 ошибки

3. **Интегрировать TOC:**
   - Добавить `TableOfContents` компонент
   - Показывать TOC справа от контента

## 📌 Команды для быстрого старта

```bash
# Создать структуру папок
mkdir -p src/components/docs
mkdir -p src/routes
mkdir -p src/config

# Установить зависимости (если еще не установлены)
npm install markdown-to-jsx gray-matter prism-react-renderer

# Запустить dev сервер
npm run dev
```

## 🆘 Troubleshooting

### Проблема: Markdown не загружается
**Решение:** Проверить путь к файлам в `public/docs/`

### Проблема: Стили не применяются
**Решение:** Убедиться что `@/lib/markdown/styles.css` импортирован

### Проблема: Ссылки не работают
**Решение:** Проверить интеграцию с React Router в `MarkdownContent`

## 📊 Метрики успеха первого этапа

- ✅ Хотя бы одна страница документации работает
- ✅ Навигация между страницами функционирует
- ✅ Контент читаем и стилизован
- ✅ Нет критических ошибок в консоли

---

**Время на реализацию:** ~2 часа
**Результат:** Рабочий прототип системы документации