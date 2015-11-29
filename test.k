; converted from npm's package json (https://raw.githubusercontent.com/npm/npm/master/package.json)

version = 3.5.1
name = npm
description = a package manager for JavaScript
preferGlobal = true
homepage = https\://docs.npmjs.com/
author = Isaac Z. Schlueter <i@izs.me> (http\://blog.izs.me)
main = ./lib/npm.js
bin = ./bin/npm-cli.js
license = Artistic-2.0

[keywords]
= install
= modules
= package manager
= package.json

[config]
publishtest = false

[repository]
type = git
url = https\://github.com/npm/npm

[bugs]
url = http\://github.com/npm/npm/issues

[directories]
man = ./man
bin = ./bin
doc = ./doc
lib = ./lib

[dependencies]
slide = ~1.1.6
text-table = ~0.2.0
ansistyles = ~0.1.3
glob = ~5.0.15
npm-install-checks = ~2.0.1
retry = ~0.8.0
lockfile = ~1.0.1
read-package-tree = ~5.1.2
realize-package-specifier = ~3.0.1
async-some = ~1.0.2
config-chain = ~1.1.9
graceful-fs = ~4.1.2
lodash\.uniq = ~3.2.2
archy = ~1.0.0
dezalgo = ~1.0.3
sorted-object = ~1.0.0
fs-write-stream-atomic = ~1.0.4
nopt = ~3.0.4
normalize-package-data = ~2.3.5
npm-registry-client = ~7.0.9
unique-filename = ~1.0.0
which = ~1.2.0
ansicolors = ~0.3.2
has-unicode = ~1.0.1
npm-package-arg = ~4.1.0
read-installed = ~4.0.3
lodash\.clonedeep = ~3.0.2
osenv = ~0.1.3
read-package-json = ~2.0.2
fstream-npm = ~1.0.7
inflight = ~1.0.4
ini = ~1.3.4
init-package-json = ~1.9.1
aproba = ~1.0.1
fs-vacuum = ~1.2.7
lodash\.without = ~3.2.1
normalize-git-url = ~3.0.1
semver = ~5.0.3
sha = ~2.0.1
columnify = ~1.5.2
node-gyp = ~3.2.0
npm-cache-filename = ~1.0.2
opener = ~1.4.1
hosted-git-info = ~2.1.4
mkdirp = ~0.5.1
read = ~1.0.7
unpipe = ~1.0.0
npmlog = ~2.0.0
read-cmd-shim = ~1.0.1
tar = ~2.2.1
uid-number = 0.0.6
abbrev = ~1.0.7
editor = ~1.0.0
fstream = ~1.0.8
inherits = ~2.0.1
wrappy = ~1.0.1
write-file-atomic = ~1.1.3
chownr = ~1.0.1
npm-user-validate = ~0.1.2
path-is-inside = ~1.0.1
lodash\.union = ~3.1.0
once = ~1.3.2
validate-npm-package-name = ~2.2.2
cmd-shim = ~2.0.1
iferr = ~0.1.5
rimraf = ~2.4.3
umask = ~1.1.0

[bundleDependencies]
= abbrev
= ansi-regex
= ansicolors
= ansistyles
= aproba
= archy
= async-some
= chownr
= cmd-shim
= columnify
= config-chain
= debuglog
= dezalgo
= editor
= fs-vacuum
= fs-write-stream-atomic
= fstream
= fstream-npm
= glob
= graceful-fs
= has-unicode
= hosted-git-info
= iferr
= inflight
= inherits
= ini
= init-package-json
= lockfile
= lodash._baseindexof
= lodash._baseuniq
= lodash._bindcallback
= lodash._cacheindexof
= lodash._createcache
= lodash._getnative
= lodash.clonedeep
= lodash.isarguments
= lodash.isarray
= lodash.keys
= lodash.restparam
= lodash.union
= lodash.uniq
= lodash.without
= mkdirp
= node-gyp
= nopt
= normalize-git-url
= normalize-package-data
= npm-cache-filename
= npm-install-checks
= npm-package-arg
= npm-registry-client
= npm-user-validate
= npmlog
= once
= opener
= osenv
= path-is-inside
= read
= read-cmd-shim
= read-installed
= read-package-json
= read-package-tree
= readdir-scoped-modules
= realize-package-specifier
= request
= retry
= rimraf
= semver
= sha
= slide
= sorted-object
= strip-ansi
= tar
= text-table
= uid-number
= umask
= unique-filename
= unpipe
= validate-npm-package-license
= validate-npm-package-name
= which
= wrappy
= write-file-atomic

[devDependencies]
marked-man = ~0.1.5
nock = ~1.9.0
npm-registry-mock = ~1.0.1
readable-stream = ~2.0.2
sprintf-js = ~1.0.3
deep-equal = ~1.0.1
marked = ~0.3.5
npm-registry-couchapp = ~2.6.11
require-inject = ~1.3.0
standard = ~5.3.1
tap = ~2.2.0

[scripts]
test-all = standard && npm run test-legacy && npm run test-tap
test-legacy = node ./test/run.js
test = standard && npm run test-tap
prepublish = node bin/npm-cli.js prune --prefix\=. --no-global && rimraf test/*/*/node_modules && make -j4 doc
preversion = bash scripts/update-authors.sh && git add AUTHORS && git commit -m "update AUTHORS" || true
tap = tap --timeout 240
test-tap = npm run tap -- "test/tap/*.js"
test-node = "$NODE" ./test/run.js && "$NODE" "node_modules/.bin/tap" --timeout 240 "test/tap/*.js"
dumpconf = env | grep npm | sort | uniq
