
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

        return data[0].name;
    } catch (error) {
        console.error(error);
        return 'unknown version';
    }
}
