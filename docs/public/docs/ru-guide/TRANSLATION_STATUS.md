# Translation Status (ru-guide)

Этот файл отслеживает прогресс перевода русской документации.  
Используемые статусы:
- TODO — требуется перевод
- IN_PROGRESS — перевод начат, требуется вычитка
- DONE — перевод завершён и вычитан
- SKIPPED — не требует перевода (намеренно оставлено на английском)

Правила:
1. После полного перевода файла убрать исходный английский текст и маркер <!-- TODO -->
2. Термины протокола и ключевые идентификаторы (названия правил, CLI команды, YAML ключи) НЕ переводим.
3. Если структура исходного файла изменилась — добавить в начало: `<!-- SOURCE_UPDATED: <commit or date> -->`
4. Для больших изменений: сначала IN_PROGRESS, после вычитки менять на DONE.

## Summary

Итого файлов: 76  
DONE: 76  
IN_PROGRESS: 0  
TODO: 0

## Priority Order (рекомендованный)

1. introduction/what-is.md
2. introduction/quickstart.md
3. introduction/install.md
4. cli/linter/linter.md
5. cli/generator/generator.md
6. cli/package-manager/package-manager.md
7. cli/breaking-changes/breaking-changes.md
8. api-service/overview.md
9. migration/* (пакетом)

---

## File Status Table

| File | Status | Notes |
|------|--------|-------|
| introduction/what-is.md | DONE | Полный перевод выполнен |
| introduction/quickstart.md | DONE | Структура сохранена, перевод вычитан |
| introduction/install.md | DONE | Переведено, команды проверены |
| api-service/overview.md | DONE | Переведено |
| ci-cd/github-actions.md | DONE | Переведено |
| ci-cd/gitlab.md | DONE | Переведено |
| cli/auto-completion/auto-completion.md | DONE | Переведено |
| cli/breaking-changes/breaking-changes.md | DONE | Переведено и вычитано |
| cli/breaking-changes/rules/enum-no-delete.md | DONE | Переведено |
| cli/breaking-changes/rules/enum-value-no-delete.md | DONE | Переведено |
| cli/breaking-changes/rules/enum-value-same-name.md | DONE | Переведено |
| cli/breaking-changes/rules/field-no-delete.md | DONE | Переведено |
| cli/breaking-changes/rules/field-same-cardinality.md | DONE | Переведено |
| cli/breaking-changes/rules/field-same-type.md | DONE | Переведено |
| cli/breaking-changes/rules/import-no-delete.md | DONE | Переведено |
| cli/breaking-changes/rules/message-no-delete.md | DONE | Переведено |
| cli/breaking-changes/rules/oneof-field-no-delete.md | DONE | Переведено |
| cli/breaking-changes/rules/oneof-field-same-type.md | DONE | Переведено |
| cli/breaking-changes/rules/oneof-no-delete.md | DONE | Переведено |
| cli/breaking-changes/rules/rpc-no-delete.md | DONE | Переведено |
| cli/breaking-changes/rules/rpc-same-request-type.md | DONE | Переведено |
| cli/breaking-changes/rules/rpc-same-response-type.md | DONE | Переведено |
| cli/breaking-changes/rules/service-no-delete.md | DONE | Переведено |
| cli/configuration/configuration.md | DONE | Переведено |
| cli/generator/examples/go.md | DONE | Переведено |
| cli/generator/examples/grpc-gateway.md | DONE | Переведено |
| cli/generator/examples/validate.md | DONE | Переведено |
| cli/generator/generator.md | DONE | Переведено и вычитано |
| cli/linter/linter.md | DONE | Переведено и вычитано |
| cli/linter/rules/comment-enum-value.md | DONE | Переведено |
| cli/linter/rules/comment-enum.md | DONE | Переведено |
| cli/linter/rules/comment-field.md | DONE | Переведено |
| cli/linter/rules/comment-message.md | DONE | Переведено |
| cli/linter/rules/comment-oneof.md | DONE | Переведено |
| cli/linter/rules/comment-rpc.md | DONE | Переведено |
| cli/linter/rules/comment-service.md | DONE | Переведено |
| cli/linter/rules/directory-same-package.md | DONE | Переведено |
| cli/linter/rules/enum-first-value-zero.md | DONE | Переведено |
| cli/linter/rules/enum-no-allow-alias.md | DONE | Переведено |
| cli/linter/rules/enum-pascal-case.md | DONE | Переведено |
| cli/linter/rules/enum-value-prefix.md | DONE | Переведено |
| cli/linter/rules/enum-value-upper-snake-case.md | DONE | Переведено |
| cli/linter/rules/enum-zero-value-suffix.md | DONE | Переведено |
| cli/linter/rules/field-lower-snake-case.md | DONE | Переведено |
| cli/linter/rules/file-lower-snake-case.md | DONE | Переведено |
| cli/linter/rules/import-no-public.md | DONE | Переведено |
| cli/linter/rules/import-no-weak.md | DONE | Переведено |
| cli/linter/rules/import-used.md | DONE | Переведено |
| cli/linter/rules/message-pascal-case.md | DONE | Переведено |
| cli/linter/rules/oneof-lower-snake-case.md | DONE | Переведено |
| cli/linter/rules/package-defined.md | DONE | Переведено |
| cli/linter/rules/package-directory-match.md | DONE | Переведено |
| cli/linter/rules/package-lower-snake-case.md | DONE | Переведено |
| cli/linter/rules/package-same-csharp-namespace.md | DONE | Переведено |
| cli/linter/rules/package-same-directory.md | DONE | Переведено |
| cli/linter/rules/package-same-go-package.md | DONE | Переведено |
| cli/linter/rules/package-same-java-multiple-files.md | DONE | Переведено |
| cli/linter/rules/package-same-java-package.md | DONE | Переведено |
| cli/linter/rules/package-same-php-namespace.md | DONE | Переведено |
| cli/linter/rules/package-same-ruby-package.md | DONE | Переведено |
| cli/linter/rules/package-same-swift-prefix.md | DONE | Переведено |
| cli/linter/rules/package-version-suffix.md | DONE | Переведено |
| cli/linter/rules/rpc-no-client-streaming.md | DONE | Переведено |
| cli/linter/rules/rpc-no-server-streaming.md | DONE | Переведено |
| cli/linter/rules/rpc-pascal-case.md | DONE | Переведено |
| cli/linter/rules/rpc-request-response-unique.md | DONE | Переведено |
| cli/linter/rules/rpc-request-standard-name.md | DONE | Переведено |
| cli/linter/rules/rpc-response-standard-name.md | DONE | Переведено |
| cli/linter/rules/service-pascal-case.md | DONE | Переведено |
| cli/linter/rules/service-suffix.md | DONE | Переведено |
| cli/package-manager/easyp-vs-buf.md | DONE | Переведено |
| cli/package-manager/package-manager.md | DONE | Переведено и вычитано |
| migration/buf-cli.md | DONE | Переведено |
| migration/protoc.md | DONE | Переведено |
| migration/protolock.md | DONE | Переведено |
| migration/prototool.md | DONE | Переведено |

---

## Workflow

1. Выбираем файл со статусом TODO.
2. Меняем статус на IN_PROGRESS здесь.
3. Переводим, проверяем терминологию (сверка с глоссарием — добавить позже).
4. Меняем статус на DONE.
5. Удаляем маркер `<!-- TODO -->` внутри файла.

## Planned Terminology (черновик)

| EN | RU | Notes |
|----|----|-------|
| Linting | Линтинг | Транслитерация |
| Breaking Changes | Несовместимые изменения | Название секции |
| Dependency | Зависимость |  |
| Code Generation | Генерация кода |  |
| Lock file | Lock-файл |  |
| Remote Plugin | Удалённый плагин |  |
| Package Manager | Менеджер пакетов |  |

(Будет дополнено.)

---

## История обновлений

- Initial file created: all statuses set (3 IN_PROGRESS, rest TODO).
- Marked breaking-changes page as DONE.
- Linter rules (41 файлов) переведены: статусы DONE.
- Breaking-changes rules (15 файлов) вычитаны и помечены DONE.
- Все оставшиеся страницы переведены: итог DONE = 76.
