module.exports = {
    "branches": [
        { "name": "master" },
        { "name": "main" },
        { "name": "beta", "channel": "beta", "prerelease": "beta" },
    ],
    "plugins": [
        "@semantic-release/commit-analyzer",
        "@semantic-release/release-notes-generator",
        "@semantic-release/github",
    ],
}