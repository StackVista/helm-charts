// Bump a published chart's prerelease version on a branch via ONE GitHub-signed
// commit (createCommitOnBranch), retrying if the branch advances meanwhile.
//
// Run from an actions/github-script step:
//   await require(`${process.env.GITHUB_WORKSPACE}/scripts/ci/bump-suse-prerelease.cjs`)
//     ({ github, core, chart: 'stable/suse-observability' });
//
// The injected `github` Octokit must be authenticated as the helm-charts-internal
// App (commits it creates are GitHub-signed, satisfying require_signed_commits and
// bypassing branch protection). Reads/commits go through the API, so no local git.
//
//   x.y.z-pre.N -> x.y.z-pre.(N+1)   |   x.y.z -> x.y.(z+1)-pre.1

const COMMIT_MUTATION = `
mutation($repo: String!, $branch: String!, $oid: GitObjectID!, $msg: String!,
         $cp: String!, $cc: Base64String!, $rp: String!, $rc: Base64String!) {
  createCommitOnBranch(input: {
    branch:          { repositoryNameWithOwner: $repo, branchName: $branch }
    expectedHeadOid: $oid
    message:         { headline: $msg }
    fileChanges:     { additions: [{ path: $cp, contents: $cc }, { path: $rp, contents: $rc }] }
  }) { commit { url } }
}`;

function bumpVersion(v) {
  let m = v.match(/^(\d+\.\d+\.\d+)-pre\.(\d+)$/);
  if (m) return `${m[1]}-pre.${Number(m[2]) + 1}`;
  m = v.match(/^(\d+)\.(\d+)\.(\d+)$/);
  if (m) return `${m[1]}.${m[2]}.${Number(m[3]) + 1}-pre.1`;
  throw new Error(`Unsupported chart version "${v}"; expected x.y.z or x.y.z-pre.N`);
}

// A stale expectedHeadOid is the one error we retry. GitHub returns type
// STALE_DATA with a message like 'Expected branch to point to "<oid>" but it did
// not.'; match the type plus that canonical wording narrowly (not a broad
// /expected.*but/, which also catches validation errors that never succeed).
function isStaleHead(err) {
  const errs = err?.errors || err?.response?.errors || [];
  const hit = (e) =>
    String(e?.type).toUpperCase() === 'STALE_DATA' ||
    /point to .*but it did not|is at .*but expected/i.test(e?.message || '');
  return errs.length > 0 ? errs.every(hit) : hit({ message: err?.message });
}

module.exports = async ({ github, core, chart, branch = 'master', maxAttempts = 5 }) => {
  const [owner, repo] = (process.env.GITHUB_REPOSITORY || 'StackVista/helm-charts-internal').split('/');
  const chartPath = `${chart}/Chart.yaml`;
  const readmePath = `${chart}/README.md`;

  const readAt = async (path, ref) => {
    const { data } = await github.rest.repos.getContent({ owner, repo, path, ref });
    return Buffer.from(data.content, 'base64').toString('utf8');
  };

  for (let attempt = 1; ; attempt++) {
    // Re-read the live tip each attempt so a retry re-bumps on top of whatever
    // just landed instead of replaying a stale version.
    const { data: br } = await github.rest.repos.getBranch({ owner, repo, branch });
    const headOid = br.commit.sha;

    const chartYaml = await readAt(chartPath, headOid);
    const readme = await readAt(readmePath, headOid);

    const curVer = (chartYaml.match(/^version:\s*(.+?)\s*$/m) || [])[1];
    if (!curVer) throw new Error(`No 'version:' field in ${chartPath}`);
    const newVer = bumpVersion(curVer);
    const newChart = chartYaml.replace(/^(version:\s*).+$/m, `$1${newVer}`);
    const newReadme = readme.replace(/^(Current chart version is ).*$/m, `$1\`${newVer}\``);
    // Chart.yaml always changes; if the README didn't, its version line is
    // missing/reworded — surface it instead of silently shipping a stale README.
    if (newReadme === readme) core.warning(`No 'Current chart version is' line updated in ${readmePath}`);

    core.info(`Bumping ${chart} ${curVer} -> ${newVer} on ${branch}@${headOid.slice(0, 12)} (attempt ${attempt}/${maxAttempts})`);

    try {
      const res = await github.graphql(COMMIT_MUTATION, {
        repo: `${owner}/${repo}`, branch, oid: headOid,
        msg: `Updating '${chart}' helm chart version to ${newVer} [skip ci]`,
        cp: chartPath, cc: Buffer.from(newChart).toString('base64'),
        rp: readmePath, rc: Buffer.from(newReadme).toString('base64'),
      });
      core.info(`Created commit: ${res.createCommitOnBranch.commit.url}`);
      return;
    } catch (err) {
      if (isStaleHead(err) && attempt < maxAttempts) {
        core.info('Branch advanced since the tip was read; re-deriving and retrying…');
        continue;
      }
      throw err;
    }
  }
};

module.exports.bumpVersion = bumpVersion;
module.exports.isStaleHead = isStaleHead;
