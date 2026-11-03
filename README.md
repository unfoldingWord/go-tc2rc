# go-tc2rc

Go library and CLI for converting [translationCore](https://software.translationcore.com/) (tc)
repositories to [Resource Container](https://resource-container.readthedocs.io/) (RC) repositories.

A tc repo is a single-book scripture project (subject "Aligned Bible"): a
`manifest.json`, one exported USFM file named after the repo (e.g.
`as_irv_1th_book.usfm`), and tool data (alignment JSON, `settings.json`) that
is not carried into the RC.

Combined with [go-rc2sb](https://github.com/unfoldingWord/go-rc2sb), this
completes the pipeline that turns any tc repo on DCS into a
[Scripture Burrito](https://docs.burrito.bible/) (SB):

```
tc2rc <tcDir> <rcDir>
rc2sb <rcDir> <sbDir>
```

## Library

```go
import tc2rc "github.com/unfoldingWord/go-tc2rc"

rep := tc2rc.Convert(ctx, inDir, outDir, tc2rc.Options{})
if !rep.OK {
    log.Fatal(rep.Error)
}
```

`Options.RepoName` overrides the DCS repo name used to locate the exported
USFM file when the input directory's basename differs from the repo name.

## CLI

```bash
go install github.com/unfoldingWord/go-tc2rc/cmd/tc2rc@latest

tc2rc [--repo-name <name>] <tcRepoDir> <rcOutDir>
```

A conversion report is written to `<rcOutDir>.report.json`. Exit code is 0
when a usable RC repo was produced, 1 otherwise.

## What the conversion does

1. **manifest.json → manifest.yaml** — Dublin Core fields are built from the
   tc manifest: `resource.id` becomes the RC identifier, `target_language`
   becomes the language block, `license` becomes rights, `tsv_relation`
   becomes relation, `source_translations` becomes source, and the subject is
   `Aligned Bible` (`Bible` when the text comes from chapter JSON, see 2).
2. **USFM selection** — tc repos usually hold two `.usfm` files: the tc
   export named after the repo and a copy named after the book. Selection
   priority: `<repoName>.usfm`, then `<book>.usfm` (e.g. `amo.usfm` for
   Amos), then the first `.usfm` file in the repo root.
   translationCore saves made before `tc_version` existed (`generator.name:
   tc-desktop`, `package_version: 1`) have no `.usfm` file; their text is in
   `<book>/<chapter>.json`, each mapping verse numbers to text. For these, the
   USFM is built from those files (`\id`, `\h`, `\toc1-3`, `\mt`, then `\c`,
   `\p` and one `\v` per verse). That text is unaligned, so the RC subject is
   `Bible` and `Report.ChapterDir` names the book directory used. A repo with
   neither a `.usfm` file nor chapter JSON fails to convert.
3. **Localized book title** — read from the USFM `\mt`, `\mt1`, or `\h`
   marker (in that order), falling back to `target_language.book.name`,
   `project.name`, then the canonical English name. Used for the RC project
   title (and the manifest title when `resource.name` is empty).
4. **File relocation** — the selected USFM is copied to `<BOOK>.usfm` (e.g.
   `1TH.usfm`); the single manifest project entry points at it with the
   correct sort, versification (`ufw`), and category (`bible-ot`/`bible-nt`).
   Alignment JSON directories and `settings.json` are not carried over.
5. **License/README** — `LICENSE.md` and `README.md` are copied through; when
   the tc repo has no license file, the standard unfoldingWord CC BY-SA 4.0
   `LICENSE.md` is written.

## Finding tc repos on DCS

```bash
curl "https://git.door43.org/api/v1/repos/search?metadataType=tc"
```

`scripts/fetch_tc_repos.py` pages through that endpoint and clones every tc
repo into a target directory:

```bash
python3 scripts/fetch_tc_repos.py --dest /tmp/tc_repos --limit 100
```

## Build & Test

```bash
go build ./...
go test ./...
go vet ./...
```
