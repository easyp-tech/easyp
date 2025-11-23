import { getAdjacentPages } from './src/utils/getAdjacentPages.ts'

const path = 'guide/introduction/what-is'
const result = getAdjacentPages(path)
console.log('Result for', path, ':', JSON.stringify(result, null, 2))

const path2 = 'guide/cli/linter/rules/package-same-java-multiple-files'
const result2 = getAdjacentPages(path2)
console.log('Result for', path2, ':', JSON.stringify(result2, null, 2))
