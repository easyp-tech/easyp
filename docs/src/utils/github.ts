
interface GithubTag {
    name: string;
}

export async function GetLatestRelease(): Promise<string> {
    try {
        const response = await fetch('https://api.github.com/repos/easyp-tech/easyp/tags');

        if (!response.ok) {
            throw new Error(`GitHub API returned status code ${response.status}`);
        }

        const data = await response.json() as GithubTag[];

        if (data.length === 0) {
            throw new Error('No tags found for this repository');
        }

        // Filter out pre-release versions (containing -rc, -alpha, -beta, etc.)
        // Only return stable releases
        const stableReleases = data.filter(tag => {
            const version = tag.name.toLowerCase();
            return !version.includes('-rc') &&
                !version.includes('-alpha') &&
                !version.includes('-beta') &&
                !version.includes('-pre') &&
                !version.includes('-dev');
        });

        if (stableReleases.length === 0) {
            // If no stable releases found, fall back to the first tag
            return data[0].name;
        }

        return stableReleases[0].name;
    } catch (error) {
        console.error(error);
        return 'unknown version';
    }
}
