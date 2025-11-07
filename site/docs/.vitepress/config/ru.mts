import {createRequire} from 'module'
import {defineConfig, type DefaultTheme} from 'vitepress'
import {GetLatestRelease} from "../../../src/version/data.mjs";

const require = createRequire(import.meta.url)
const pkg = require('vitepress/package.json')

const version = await GetLatestRelease()

export const ru = defineConfig({
    lang: 'ru-RU',
    description: 'Лучший инструмент для работы с protobuf.',

    themeConfig: {
        nav: nav(),

        sidebar: {
            '/ru/guide/': {base: '/ru/guide/', items: sidebarGuide()},
        },

        editLink: {
            pattern: 'https://github.com/easyp-tech/site/edit/main/docs/:path',
            text: 'Редактировать страницу'
        },

        footer: {
            message: 'Распространяется под лицензией Apache Version 2.0.',
            copyright: 'Авторские права © 2024 представлены Эдгаром Сипки и Василием Близнецовым'
        },

        outline: {label: 'Содержание страницы'},

        docFooter: {
            prev: 'Предыдущая страница',
            next: 'Следующая страница'
        },

        lastUpdated: {
            text: 'Обновлено'
        },

        darkModeSwitchLabel: 'Оформление',
        lightModeSwitchTitle: 'Переключить на светлую тему',
        darkModeSwitchTitle: 'Переключить на тёмную тему',
        sidebarMenuLabel: 'Меню',
        returnToTopLabel: 'Вернуться к началу',
        langMenuLabel: 'Изменить язык'
    }
})

function nav(): DefaultTheme.NavItem[] {
    return [
        {
            text: 'Руководство',
            link: '/ru/guide/introduction/what-is',
            activeMatch: '/ru/guide/'
        },
        {
            text: 'Команда',
            link: '/ru/team',
            activeMatch: '/ru/team'
        },
        {
            text: 'Поддержать 🎁',
            link: '/ru/donate',
            activeMatch: '/ru/donate'
        },
        {
            text: version,
            items: [
                {
                    text: 'Изменения',
                    link: 'https://github.com/easyp-tech/easyp/blob/main/CHANGELOG.md'
                }
            ]
        }
    ]
}

function sidebarGuide(): DefaultTheme.SidebarItem[] {
    return [
        {
            text: 'Введение',
            collapsed: false,
            items: [
                {text: 'Для чего EasyP?', link: 'introduction/what-is'},
                {text: 'Установка EasyP cli', link: 'introduction/install'},
                {text: 'Быстрый старт', link: 'introduction/quickstart'},
                {text: 'Новости', link: 'introduction/news'},
            ]
        },
        {
            text: 'Easyp CLI',
            collapsed: false,
            items: [
                {
                    text: 'Linter',
                    link: 'cli/linter/linter',
                    collapsed: true,
                    items: [
                        {text: 'DIRECTORY_SAME_PACKAGE', link: 'cli/linter/rules/directory-same-package'},
                        {text: 'PACKAGE_DEFINED', link: 'cli/linter/rules/package-defined'},
                        {text: 'PACKAGE_DIRECTORY_MATCH', link: 'cli/linter/rules/package-directory-match'},
                        {text: 'PACKAGE_SAME_DIRECTORY', link: 'cli/linter/rules/package-same-directory'},

                        {text: 'ENUM_FIRST_VALUE_ZERO', link: 'cli/linter/rules/enum-first-value-zero'},
                        {text: 'ENUM_NO_ALLOW_ALIAS', link: 'cli/linter/rules/enum-no-allow-alias'},
                        {text: 'ENUM_PASCAL_CASE', link: 'cli/linter/rules/enum-pascal-case'},
                        {text: 'ENUM_VALUE_UPPER_SNAKE_CASE', link: 'cli/linter/rules/enum-value-upper-snake-case'},
                        {text: 'FIELD_LOWER_SNAKE_CASE', link: 'cli/linter/rules/field-lower-snake-case'},
                        {text: 'IMPORT_NO_PUBLIC', link: 'cli/linter/rules/import-no-public'},
                        {text: 'IMPORT_NO_WEAK', link: 'cli/linter/rules/import-no-weak'},
                        {text: 'IMPORT_USED', link: 'cli/linter/rules/import-used'},
                        {text: 'MESSAGE_PASCAL_CASE', link: 'cli/linter/rules/message-pascal-case'},
                        {text: 'ONEOF_LOWER_SNAKE_CASE', link: 'cli/linter/rules/oneof-lower-snake-case'},
                        {text: 'PACKAGE_LOWER_SNAKE_CASE', link: 'cli/linter/rules/package-lower-snake-case'},
                        {text: 'PACKAGE_SAME_CSHARP_NAMESPACE', link: 'cli/linter/rules/package-same-csharp-namespace'},
                        {text: 'PACKAGE_SAME_GO_PACKAGE', link: 'cli/linter/rules/package-same-go-package'},
                        {
                            text: 'PACKAGE_SAME_JAVA_MULTIPLE_FILES',
                            link: 'cli/linter/rules/package-same-java-multiple-files'
                        },
                        {text: 'PACKAGE_SAME_JAVA_PACKAGE', link: 'cli/linter/rules/package-same-java-package'},
                        {text: 'PACKAGE_SAME_PHP_NAMESPACE', link: 'cli/linter/rules/package-same-php-namespace'},
                        {text: 'PACKAGE_SAME_RUBY_PACKAGE', link: 'cli/linter/rules/package-same-ruby-package'},
                        {text: 'PACKAGE_SAME_SWIFT_PREFIX', link: 'cli/linter/rules/package-same-swift-prefix'},
                        {text: 'RPC_PASCAL_CASE', link: 'cli/linter/rules/rpc-pascal-case'},
                        {text: 'SERVICE_PASCAL_CASE', link: 'cli/linter/rules/service-pascal-case'},

                        {text: 'ENUM_VALUE_PREFIX', link: 'cli/linter/rules/enum-value-prefix'},
                        {text: 'ENUM_ZERO_VALUE_SUFFIX', link: 'cli/linter/rules/enum-zero-value-suffix'},
                        {text: 'FILE_LOWER_SNAKE_CASE', link: 'cli/linter/rules/file-lower-snake-case'},
                        {text: 'RPC_REQUEST_RESPONSE_UNIQUE', link: 'cli/linter/rules/rpc-request-response-unique'},
                        {text: 'RPC_REQUEST_STANDARD_NAME', link: 'cli/linter/rules/rpc-request-standard-name'},
                        {text: 'RPC_RESPONSE_STANDARD_NAME', link: 'cli/linter/rules/rpc-response-standard-name'},
                        {text: 'PACKAGE_VERSION_SUFFIX', link: 'cli/linter/rules/package-version-suffix'},
                        {text: 'PROTOVALIDATE', link: 'cli/linter/rules/protovalidate'},
                        {text: 'SERVICE_SUFFIX', link: 'cli/linter/rules/service-suffix'},

                        {text: 'COMMENT_ENUM', link: 'cli/linter/rules/comment-enum'},
                        {text: 'COMMENT_ENUM_VALUE', link: 'cli/linter/rules/comment-enum-value'},
                        {text: 'COMMENT_FIELD', link: 'cli/linter/rules/comment-field'},
                        {text: 'COMMENT_MESSAGE', link: 'cli/linter/rules/comment-message'},
                        {text: 'COMMENT_ONEOF', link: 'cli/linter/rules/comment-oneof'},
                        {text: 'COMMENT_RPC', link: 'cli/linter/rules/comment-rpc'},
                        {text: 'COMMENT_SERVICE', link: 'cli/linter/rules/comment-service'},

                        {text: 'RPC_NO_CLIENT_STREAMING', link: 'cli/linter/rules/rpc-no-client-streaming'},
                        {text: 'RPC_NO_SERVER_STREAMING', link: 'cli/linter/rules/rpc-no-server-streaming'},

                        {text: 'PACKAGE_NO_IMPORT_CYCLE', link: 'cli/linter/rules/package-no-import-cycle'},
                    ],
                },
                {
                    text: 'Package Manager',
                    link: 'cli/package-manager/package-manager',
                },
                {
                    text: 'Generator',
                    link: 'cli/generator/generator',
                },
                {
                    text: 'Breaking Changes checks',
                    link: 'cli/breaking-changes/breaking-changes',
                },
                {
                    text: 'Auto completion',
                    link: 'cli/auto-completion/auto-completion',
                },
            ]
        },
        {
            text: 'CI/CD',
            collapsed: false,
            items: [
                {text: 'Github Actions', link: 'ci-cd/github-actions'},
                {text: 'Gitlab', link: 'ci-cd/gitlab'},
            ]
        },
        {
            text: 'Migration guide',
            collapsed: false,
            items: [
                {text: 'Migrate from Buf CLI', link: 'migration/buf-cli'},
                {text: 'Migrate from Prototool', link: 'migration/prototool'},
                {text: 'Migrate from Protolock', link: 'migration/protolock'},
                {text: 'Migrate from protoc', link: 'migration/protoc'},
            ]
        }
    ]
}
