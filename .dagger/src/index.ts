import {
  argument,
  Container,
  dag,
  Directory,
  File,
  func,
  object,
  Secret,
} from "@dagger.io/dagger"

const GO_IMAGE =
  "golang:1.26.5-bookworm@sha256:53eeac89074db483fdf0ab3be1df32bf6e47562263d2d0d6baa7f26acb4957dd"
const JQ_IMAGE =
  "ghcr.io/jqlang/jq:1.8.2@sha256:b9c68867e5766576263a222e91db3de422d802069c7af70440e667a95344e486"

const SOURCE_EXCLUDES = [
  ".cache",
  ".cache/**",
  ".dagger/node_modules",
  ".dagger/node_modules/**",
  ".dagger-inputs",
  ".dagger-inputs/**",
  "dist",
  "dist/**",
  "site/node_modules",
  "site/node_modules/**",
]

const ASSET_OUTPUTS = [
  "araihu-assets.json",
  "site/internal/brand/assets/goshtoso-icon-transparent.svg",
  "site/internal/brand/assets/goshtoso-logo-transparent.svg",
] as const

@object()
export class GoshtosoCharts {
  /** Run the exact required root, site, Muamba, race, vet, and workflow checks. */
  @func({ cache: "never" })
  async ci(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    trustDomain: string,
    runNonce: string,
  ): Promise<string> {
    return this.goProject(source, trustDomain, runNonce)
      .withExec([
        "go", "tool", "muamba", "sync", "--strict", "--cache-dir", ".cache/muamba",
      ])
      .withExec([
        "go", "tool", "muamba", "verify", "--strict", "--cache-dir", ".cache/muamba",
      ])
      .withExec([
        "go", "tool", "muamba", "generate-go", "--strict", "--check",
        "--dir", "assets", "--output", "muamba_gen.go",
      ])
      .withExec(["git", "diff", "--exit-code"])
      .withExec(["go", "test", "./...", "-count=1"])
      .withExec([
        "go", "test", "-race", "./assets", "./components/interactive/wordcloud", "-count=1",
      ])
      .withExec(["go", "vet", "./..."])
      .withExec(["bash", "-euo", "pipefail", "-c", "cd site && go vet ./..."])
      .withExec(["./scripts/materialize-dagger-input_test.sh"])
      .withExec(["go", "test", "./.dagger/tools/...", "-count=1"])
      .withExec([
        "go", "run", "github.com/rhysd/actionlint/cmd/actionlint@v1.7.11",
      ])
      .withExec(["printf", "Goshtoso Charts CI passed\n"])
      .stdout()
  }

  /**
   * Validate an immutable Arai Hu release, update only allowlisted fallbacks,
   * prove idempotency, and return the files consumed by create-pull-request.
   */
  @func({ cache: "never" })
  async updateAraihuAssets(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    payload: File,
    githubToken: Secret,
    trustDomain: string,
    runNonce: string,
  ): Promise<Directory> {
    const input = await this.readStringObject(payload, [
      "assets_repository",
      "assets_revision",
      "release",
      "release_url",
      "release_sha256",
      "release_json_sha256",
      "source_repository",
    ])
    this.validateSourceRepository(input.source_repository)
    const assetsRepository = input.assets_repository
    const assetsRevision = input.assets_revision
    const release = input.release
    const releaseUrl = input.release_url
    const releaseSha256 = input.release_sha256
    const releaseJsonSha256 = input.release_json_sha256
    this.validateAssetsIdentity(
      assetsRepository,
      assetsRevision,
      release,
      releaseUrl,
      releaseSha256,
      releaseJsonSha256,
    )

    const archive = dag.http(releaseUrl, {
      checksum: `sha256:${releaseSha256}`,
      name: `araihu-assets-${release}.tar.gz`,
    })
    const updaterArgs = [
      "-release-dir", "/tmp/araihu-assets-release",
      "-assets-repository", assetsRepository,
      "-assets-revision", assetsRevision,
      "-release", release,
      "-release-url", releaseUrl,
      "-release-sha256", releaseSha256,
      "-release-json-sha256", releaseJsonSha256,
    ]

    const updated = this.goProject(source, trustDomain, runNonce)
      .withSecretVariable("GITHUB_TOKEN", githubToken)
      .withFile("/tmp/araihu-assets-release.tar.gz", archive)
      .withExec(["./scripts/reject-archive-links_test.sh"])
      .withExec([
        "go", "run", "./.dagger/tools/verifytag",
        "-release", release,
        "-revision", assetsRevision,
      ])
      .withExec([
        "bash", "-euo", "pipefail", "-c",
        `while IFS= read -r member; do
          case "$member" in
            /*|../*|*/../*|*/..) echo "unsafe archive member: $member" >&2; exit 1 ;;
          esac
        done < <(tar --list --gzip --file /tmp/araihu-assets-release.tar.gz)
        ./scripts/reject-archive-links.sh /tmp/araihu-assets-release.tar.gz
        mkdir /tmp/araihu-assets-release
        tar --extract --gzip --file /tmp/araihu-assets-release.tar.gz \
          --directory /tmp/araihu-assets-release --no-same-owner --no-same-permissions`,
      ])
      .withExec(["go", "run", "./cmd/araihu-assets-update", ...updaterArgs])
      .withExec([
        "bash", "-euo", "pipefail", "-c", "git diff --binary > /tmp/first.diff",
      ])
      .withExec(["go", "run", "./cmd/araihu-assets-update", ...updaterArgs])
      .withExec([
        "bash", "-euo", "pipefail", "-c", "git diff --binary > /tmp/second.diff && cmp /tmp/first.diff /tmp/second.diff",
      ])
      .withExec([
        "bash", "-euo", "pipefail", "-c",
        `while IFS= read -r path; do
          case "$path" in
            araihu-assets.json|site/internal/brand/assets/goshtoso-icon-transparent.svg|site/internal/brand/assets/goshtoso-logo-transparent.svg) ;;
            *) echo "updater changed non-allowlisted path: $path" >&2; exit 1 ;;
          esac
        done < <(git diff --name-only)`,
      ])
      .withExec([
        "go", "test", "./internal/araihuassets", "./cmd/araihu-assets-update", "-count=1",
      ])

    return dag.directory()
      .withFile(ASSET_OUTPUTS[0], updated.file(`/work/${ASSET_OUTPUTS[0]}`))
      .withFile(ASSET_OUTPUTS[1], updated.file(`/work/${ASSET_OUTPUTS[1]}`))
      .withFile(ASSET_OUTPUTS[2], updated.file(`/work/${ASSET_OUTPUTS[2]}`))
  }

  /** Resolve the exact deployment version without causing an external effect. */
  @func({ cache: "never" })
  async resolveDeployVersion(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    refType: string,
    refName: string,
    sourceSha: string,
    runNonce: string,
  ): Promise<string> {
    this.validateRunNonce(runNonce)
    return this.base(source, "deploy", runNonce)
      .withExec(["./scripts/resolve-deploy-version_test.sh"])
      .withExec(["./scripts/resolve-deploy-version.sh", refType, refName, sourceSha])
      .stdout()
  }

  /** Resolve a version and dispatch the verified identity to araihu/fly-deploy. */
  @func({ cache: "never" })
  async dispatchFly(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    payload: File,
    token: Secret,
    runNonce: string,
  ): Promise<string> {
    const input = await this.readStringObject(payload, [
      "ref_type",
      "ref_name",
      "source_sha",
      "source_run_id",
      "source_repository",
    ])
    this.validateSourceRepository(input.source_repository)
    const refType = input.ref_type
    const refName = input.ref_name
    const sourceSha = input.source_sha
    const sourceRunId = input.source_run_id
    this.validateRunNonce(runNonce)
    if (!/^[1-9][0-9]*$/.test(sourceRunId)) {
      throw new Error("source run ID must be a positive decimal GitHub run ID")
    }

    const version = (await this.base(source, "deploy", runNonce)
      .withExec(["./scripts/resolve-deploy-version_test.sh"])
      .withExec(["./scripts/resolve-deploy-version.sh", refType, refName, sourceSha])
      .stdout()).trim()

    return this.base(source, "deploy", runNonce)
      .withSecretVariable("FLY_DEPLOY_DISPATCH_TOKEN", token)
      .withExec([
        "go", "run", "./.dagger/tools/flydispatch",
        "-source-sha", sourceSha,
        "-source-run-id", sourceRunId,
        "-version", version,
      ])
      .stdout()
  }

  private goProject(source: Directory, trustDomain: string, runNonce: string): Container {
    const partition = this.validateTrustDomain(trustDomain)
    // hostinger-vps-pr and GitHub-hosted PRs execute on Engines separated from
    // trusted lanes. Workflow input cannot select that boundary: both PR
    // domains collapse to the constant cache namespace "pr" here.
    const cacheNamespace = partition === "fork" || partition === "internal" ? "pr" : partition
    this.validateRunNonce(runNonce)
    const project = this.base(source, partition, runNonce)
      .withEnvVariable("GOMODCACHE", "/go/pkg/mod")
      .withEnvVariable("GOCACHE", "/root/.cache/go-build")

    return project
      .withMountedCache(
        "/go/pkg/mod",
        dag.cacheVolume(`goshtoso-charts-${cacheNamespace}-go-mod-v1`),
      )
      .withMountedCache(
        "/root/.cache/go-build",
        dag.cacheVolume(`goshtoso-charts-${cacheNamespace}-go-build-v1`),
      )
      .withMountedCache(
        "/work/.cache/muamba",
        dag.cacheVolume(`goshtoso-charts-${cacheNamespace}-muamba-v1`),
      )
  }

  private base(source: Directory, trustDomain: string, runNonce: string): Container {
    const partition = this.validateTrustDomain(trustDomain)
    this.validateRunNonce(runNonce)
    const jq = dag.container().from(JQ_IMAGE).file("/jq")
    return dag.container()
      .from(GO_IMAGE)
      .withFile("/usr/local/bin/jq", jq, { permissions: 0o755 })
      .withDirectory("/work", source)
      .withWorkdir("/work")
      .withEnvVariable("GOSHTOSO_CHARTS_TRUST_DOMAIN", partition)
      .withEnvVariable("GOSHTOSO_CHARTS_RUN_NONCE", runNonce)
  }

  private validateTrustDomain(value: string): string {
    if (!/^(fork|internal|main|manual|assets|deploy|local)$/.test(value)) {
      throw new Error(`unsafe trust domain: ${value}`)
    }
    return value
  }

  private validateRunNonce(value: string): void {
    if (!/^[1-9][0-9]*-[1-9][0-9]*$/.test(value)) {
      throw new Error("run nonce must be github.run_id-github.run_attempt")
    }
  }

  private async readStringObject(
    payload: File,
    expectedKeys: readonly string[],
  ): Promise<Record<string, string>> {
    let parsed: unknown
    try {
      parsed = JSON.parse(await payload.contents())
    } catch {
      throw new Error("Dagger input payload must be valid JSON")
    }
    if (parsed === null || Array.isArray(parsed) || typeof parsed !== "object") {
      throw new Error("Dagger input payload must be a JSON object")
    }

    const record = parsed as Record<string, unknown>
    const actualKeys = Object.keys(record).sort()
    const requiredKeys = [...expectedKeys].sort()
    if (
      actualKeys.length !== requiredKeys.length ||
      actualKeys.some((key, index) => key !== requiredKeys[index])
    ) {
      throw new Error("Dagger input payload does not match the exact schema")
    }
    for (const key of requiredKeys) {
      if (typeof record[key] !== "string") {
        throw new Error(`Dagger input field ${key} must be a string`)
      }
    }
    return record as Record<string, string>
  }

  private validateSourceRepository(value: string): void {
    if (value !== "araihu/goshtoso-charts") {
      throw new Error("source repository must be araihu/goshtoso-charts")
    }
  }

  private validateAssetsIdentity(
    repository: string,
    revision: string,
    release: string,
    releaseUrl: string,
    releaseSha256: string,
    releaseJsonSha256: string,
  ): void {
    if (repository !== "araihu/assets") {
      throw new Error("assets repository must be araihu/assets")
    }
    if (!/^[0-9a-f]{40}$/.test(revision)) {
      throw new Error("assets revision must be a full lowercase Git SHA-1")
    }
    if (!/^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$/.test(release)) {
      throw new Error("assets release must be a stable vMAJOR.MINOR.PATCH tag")
    }
    if (!/^[0-9a-f]{64}$/.test(releaseSha256)) {
      throw new Error("release archive SHA-256 must be lowercase hexadecimal")
    }
    if (!/^[0-9a-f]{64}$/.test(releaseJsonSha256)) {
      throw new Error("release.json SHA-256 must be lowercase hexadecimal")
    }
    const expected = `https://github.com/araihu/assets/releases/download/${release}/araihu-assets-${release}.tar.gz`
    if (releaseUrl !== expected) {
      throw new Error("release URL does not match the immutable Arai Hu release shape")
    }
  }
}
