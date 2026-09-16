const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { spawnSync } = require('node:child_process');

const script = path.join(__dirname, 'updatecli_squash_push.sh');
const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'updatecli-auth-test-'));
try {
  const bin = path.join(tmp, 'bin');
  fs.mkdirSync(bin);
  fs.writeFileSync(path.join(bin, 'git'), `#!/usr/bin/env bash
set -euo pipefail
[[ "$1" == "-c" && "$2" == "credential.helper=" ]]
[[ "$3" == "-c" && "$4" == "http.https://x-access-token@github.com/example/private.git.extraheader=" ]]
[[ "$5" == "--config-env=http.https://x-access-token@github.com/example/private.git.extraheader=UPDATECLI_GIT_AUTH_HEADER" ]]
[[ "$UPDATECLI_GIT_AUTH_HEADER" == "AUTHORIZATION: basic $(printf 'x-access-token:%s' "$GH_TOKEN" | base64 | tr -d '\\n')" ]]
shift 5
printf '%s\\n' "$*" >> "$TEST_LOG"
case "$1" in
  ls-remote) exit "$REMOTE_STATUS" ;;
  clone) exit 17 ;;
  *) echo "unexpected git command" >&2; exit 99 ;;
esac
`, { mode: 0o755 });

  for (const [remoteStatus, checkOnly] of [[0, false], [2, false], [128, false], [0, true], [128, true]]) {
    const log = path.join(tmp, `calls-${remoteStatus}-${checkOnly}`);
    const clone = path.join(tmp, `clone-${remoteStatus}-${checkOnly}`);
    fs.mkdirSync(clone);
    fs.writeFileSync(path.join(clone, 'sentinel'), 'preserve before preflight');
    const result = spawnSync('bash', [script, 'working', 'master', 'test'], {
      encoding: 'utf8',
      env: {
        ...process.env, PATH: `${bin}:${process.env.PATH}`,
        GH_TOKEN: 'test-token-never-log', GITHUB_REPOSITORY: 'example/private',
        PUSH_CLONE_DIRECTORY: clone, TEST_LOG: log,
        REMOTE_STATUS: String(remoteStatus),
        CHECK_ACCESS_ONLY: String(checkOnly),
      },
    });
    if (checkOnly) {
      assert.equal(result.status, remoteStatus, result.stderr);
      assert.equal(fs.readFileSync(log, 'utf8'),
        'ls-remote --exit-code --heads https://x-access-token@github.com/example/private.git refs/heads/master\n');
      assert.ok(fs.existsSync(path.join(clone, 'sentinel')));
      continue;
    }
    assert.equal(result.status, remoteStatus === 0 ? 17 : remoteStatus === 2 ? 0 : 128,
      result.stderr);
    const calls = fs.readFileSync(log, 'utf8');
    assert.match(calls, /ls-remote --exit-code --heads https:\/\/x-access-token@github.com\/example\/private.git refs\/heads\/working/);
    assert.equal(calls.includes('clone --branch master'), remoteStatus === 0);
    assert.equal(fs.existsSync(path.join(clone, 'sentinel')), remoteStatus !== 0);
    assert.ok(!`${calls}${result.stdout}${result.stderr}`.includes('test-token-never-log'));
    if (remoteStatus === 2) assert.match(result.stdout, /does not exist/);
    if (remoteStatus === 128) {
      assert.match(result.stderr, /cannot read example\/private with the follow-up GH_TOKEN/);
      assert.ok(!result.stdout.includes('no changes'));
    }
  }
  console.log('Updatecli Git authentication/preflight tests passed');
} finally {
  fs.rmSync(tmp, { recursive: true, force: true });
}
