# GitHub Workflows

This directory contains GitHub Actions workflows for automated building and releasing of Redi and Rejs.

## Release Workflow (release.yml)

The release workflow automatically builds and publishes releases when a version tag is pushed to the repository.

### Trigger

The workflow is triggered when you push a tag that matches the pattern `v*.*.*` (semantic versioning):

```bash
git tag v1.0.0
git push origin v1.0.0
git push origin --tag
```

### What it does

1. **Multi-platform builds**: Builds both `redi` and `rejs` for:
   - Linux AMD64 and ARM64
   - macOS Intel (AMD64) and Apple Silicon (ARM64)  
   - Windows AMD64

2. **Creates release archives**:
   - `.tar.gz` files for Unix-like systems (Linux, macOS)
   - `.zip` files for Windows
   - Each archive contains both `redi` and `rejs` executables

3. **Binary testing**: Tests each built binary to ensure it works correctly

4. **Creates GitHub release**:
   - Automatic release notes with download instructions
   - All platform archives attached as downloadable assets
   - SHA256 checksums for verification
   - Marked as prerelease if version contains a hyphen (e.g., `v1.0.0-beta`)

### Workflow Jobs

#### 1. Build Job
- **Matrix strategy**: Builds for all platform combinations
- **Artifacts**: Uploads build artifacts for the release job
- **Optimization**: Uses Go build flags `-s -w` to reduce binary size
- **Version injection**: Embeds the git tag version into the binary

#### 2. Release Job
- **Depends on**: Build job completion
- **Downloads**: All build artifacts from the build job
- **Creates**: GitHub release with proper release notes
- **Uploads**: All archives and checksums as release assets

#### 3. Test Binaries Job
- **Platform testing**: Tests binaries on actual target platforms
- **Version verification**: Ensures `--version` flag works correctly
- **Quality assurance**: Catches platform-specific issues

### Release Assets

For each release, you get these downloadable files:

**Combined Package (includes both redi and rejs):**
- `redi-v1.0.0-linux-amd64.tar.gz`
- `redi-v1.0.0-linux-arm64.tar.gz`
- `redi-v1.0.0-darwin-amd64.tar.gz`
- `redi-v1.0.0-darwin-arm64.tar.gz`
- `redi-v1.0.0-windows-amd64.zip`

**Additional Files:**
- `checksums.txt` - SHA256 checksums for all archives

**Note:** Each archive contains both `redi` (web server) and `rejs` (JavaScript runtime) executables. No additional files are included for cleaner packaging.

### Local Testing

Before creating a release, you can test the build process locally:

```bash
# Test multi-platform builds
./scripts/test-build.sh

# Check that binaries work
./build/redi-v1.0.0-test-darwin-arm64 --version
./build/rejs-v1.0.0-test-darwin-arm64 --version
```

### Release Process

1. **Prepare the release**:
   ```bash
   # Make sure all changes are committed
   git add .
   git commit -m "Prepare release v1.0.0"
   git push origin main
   ```

2. **Create and push the tag**:
   ```bash
   # Create the version tag
   git tag v1.0.0
   
   # Push the tag to trigger the workflow
   git push origin v1.0.0
   ```

3. **Monitor the workflow**:
   - Go to the "Actions" tab in your GitHub repository
   - Watch the "Release" workflow execute
   - Check for any build failures or issues

4. **Verify the release**:
   - Go to the "Releases" page in your GitHub repository
   - Download and test a few binaries
   - Verify checksums if needed

### Permissions Required

The workflow requires these GitHub repository permissions:
- **Contents: write** - To create releases and upload assets
- **Actions: read** - To access workflow artifacts

These are typically enabled by default for repository owners and admins.

### Recent Updates

**Artifact Actions v4**: The workflow has been updated to use `actions/upload-artifact@v4` and `actions/download-artifact@v4` to comply with GitHub's deprecation of v3. This ensures compatibility with future GitHub Actions runner environments.

**Windows Archive Extraction**: Fixed PowerShell extraction logic to avoid file conflicts when testing Windows binaries by using separate extraction directories for each archive.

**Release Name Generation**: Improved release name generation to handle edge cases where version extraction fails, preventing releases from showing as "-" in the GitHub interface.

**Clean Packaging**: Archives now contain only the executable files without README.md or LICENSE files, providing cleaner, smaller downloads. Documentation is available via GitHub repository links in the release notes.

**Combined Archives**: Both `redi` and `rejs` executables are now packaged together in a single archive per platform, simplifying downloads and ensuring users always get both tools.
