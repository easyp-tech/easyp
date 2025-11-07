import { execSync } from 'child_process';

// Get last git tag from local repository
export async function GetLatestRelease(): Promise<string> {
    try {
        // Get the latest tag from git
        // git describe --tags --abbrev=0 returns the most recent tag
        const latestTag = execSync('git describe --tags --abbrev=0', {
            encoding: 'utf-8',
            stdio: ['pipe', 'pipe', 'ignore'] // ignore stderr to avoid noise in console
        }).trim();

        if (!latestTag) {
            throw new Error('No tags found in repository');
        }

        return latestTag;
    } catch (error) {
        // If no tags exist or git command fails, try alternative approach
        try {
            // Get all tags sorted by version
            const allTags = execSync('git tag -l --sort=-v:refname', {
                encoding: 'utf-8',
                stdio: ['pipe', 'pipe', 'ignore']
            }).trim();

            if (allTags) {
                // Return the first tag (most recent by version)
                const tags = allTags.split('\n').filter(tag => tag.trim());
                if (tags.length > 0) {
                    return tags[0];
                }
            }
        } catch (secondError) {
            console.error('Failed to get git tags:', secondError);
        }

        console.error('Failed to get latest release tag:', error);
        return 'unknown version';
    }
}
