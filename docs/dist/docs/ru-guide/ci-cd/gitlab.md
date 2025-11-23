# GitLab CI

## В процессе разработки

Документация по интеграции EasyP с GitLab CI находится в процессе подготовки.  
Планируемые разделы:
- Быстрый старт (пример `.gitlab-ci.yml` для линтинга и проверки несовместимых изменений)
- Кэширование зависимостей и ускорение сборок
- Генерация кода и проверка актуальности артефактов
- Работа с приватными Git-репозиториями в `deps`
- Матрицы (parallel jobs) для разных языков и плагинов
- Автоматизация релизов по тегам

Если вам нужен пример прямо сейчас, базовый минимальный вариант может выглядеть так:

```yaml
stages:
  - validate
  - generate

variables:
  GO_VERSION: "1.22"

validate:
  stage: validate
  image: golang:${GO_VERSION}
  before_script:
    - go install github.com/easyp-tech/easyp/cmd/easyp@latest
    - echo "EasyP version:"
    - easyp version
    # При необходимости: git fetch --all --tags
  script:
    - easyp lint
    - easyp breaking --against git-ref=origin/main
  only:
    changes:
      - "**/*.proto"
      - "easyp.yaml"

generate:
  stage: generate
  image: golang:${GO_VERSION}
  before_script:
    - go install github.com/easyp-tech/easyp/cmd/easyp@latest
  script:
    - easyp generate
    - |
      if [ -n "$(git status --porcelain)" ]; then
        echo "Сгенерированный код не актуален. Закоммитьте изменения локально."
        git diff --name-only
        exit 1
      fi
  only:
    - main

```

Дополнительные рекомендации будут добавлены позже.