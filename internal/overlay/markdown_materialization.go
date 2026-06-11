package overlay

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func materializeMarkdownAsset(upstreamRepo string, asset OwnedOverlayAsset) (string, error) {
	if asset.MaterializedMarkdown == nil {
		return "", fmt.Errorf("asset %q is not configured as a materialized markdown asset", asset.Key)
	}
	if strings.TrimSpace(asset.MaterializedMarkdown.BaseSectionID) == "" {
		return "", fmt.Errorf("materialized markdown asset %q is missing a base section id", asset.Key)
	}
	baseSource := filepath.Join(upstreamRepo, filepath.FromSlash(asset.UpstreamPath))
	baseContent, err := readMaterializationSource(baseSource)
	if err != nil {
		return "", fmt.Errorf("cannot read base source for %q at %s: %w", asset.Key, baseSource, err)
	}
	sections := make([]string, 0, 1+len(asset.MaterializedMarkdown.SectionSources))
	sections = append(sections, renderMarkdownSection(asset.MaterializedMarkdown.BaseSectionID, baseContent))
	for _, section := range asset.MaterializedMarkdown.SectionSources {
		if strings.TrimSpace(section.SectionID) == "" {
			return "", fmt.Errorf("materialized markdown asset %q has an empty section id", asset.Key)
		}
		if strings.TrimSpace(section.SourcePath) == "" {
			return "", fmt.Errorf("materialized markdown asset %q has an empty section source path for %q", asset.Key, section.SectionID)
		}
		sourcePath := filepath.Join(upstreamRepo, filepath.FromSlash(section.SourcePath))
		content, err := readMaterializationSource(sourcePath)
		if err != nil {
			return "", fmt.Errorf("cannot read materialization section %q for %q at %s: %w", section.SectionID, asset.Key, sourcePath, err)
		}
		sections = append(sections, renderMarkdownSection(section.SectionID, content))
	}
	return normalizeLFTerminated(strings.Join(sections, "\n\n")), nil
}

func readMaterializationSource(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(normalizeLF(string(raw)), "\n"), nil
}

func renderMarkdownSection(sectionID, content string) string {
	content = strings.TrimRight(normalizeLF(content), "\n")
	return fmt.Sprintf("<!-- gentle-ai:%s -->\n%s\n<!-- /gentle-ai:%s -->", sectionID, content, sectionID)
}
