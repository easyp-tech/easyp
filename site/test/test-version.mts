import { GetLatestRelease } from '../src/version/data.mjs';

async function test() {
    console.log('Testing GetLatestRelease function...');

    try {
        const version = await GetLatestRelease();
        console.log(`Latest release tag: ${version}`);

        if (version === 'unknown version') {
            console.warn('Warning: Could not retrieve version from git tags');
        } else {
            console.log('✓ Successfully retrieved version from local git repository');
        }
    } catch (error) {
        console.error('Error during test:', error);
    }
}

test();
