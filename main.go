package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"

	gar "cloud.google.com/go/artifactregistry/apiv1"
	garproto "cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	flag "github.com/spf13/pflag"
	"google.golang.org/api/iterator"
)

var (
	project  = flag.StringP("project", "p", "", "Project containing the Google Artifact Registry")
	location = flag.StringP("location", "l", "", "Location of the Google Artifact Registry")
	repo     = flag.StringP("repo", "r", "", "Repository containing the Docker images")
	format   = flag.StringP("format", "f", "gib", "Output format: bytes, kib, kb, mib, mb, gib, gb")

	// optional filter
	includeTagFilter   = flag.StringP("include-tags", "i", "", "Include only tags matching the regex")
	excludeTagFilter   = flag.StringP("exclude-tags", "e", "", "Exclude tags matching the regex")
	includeImageFilter = flag.StringP("include-images", "I", "", "Include only images matching the regex")
	excludeImageFilter = flag.StringP("exclude-images", "E", "", "Exclude images matching the regex")
)

type image struct {
	name string
	size int64
}

func main() {
	flag.Parse()

	if *project == "" || *location == "" || *repo == "" {
		log.Fatal("-project, -location and -repo must be set")
	}

	var err error
	var includeTagRegex *regexp.Regexp
	var excludeTagRegex *regexp.Regexp
	var includeImageRegex *regexp.Regexp
	var excludeImageRegex *regexp.Regexp

	if *includeTagFilter != "" {
		includeTagRegex, err = regexp.Compile(*includeTagFilter)
		if err != nil {
			log.Fatalf("Error parsing --include-tags regex: %q: %s", *includeTagFilter, err)
		}
	}
	if *excludeTagFilter != "" {
		excludeTagRegex, err = regexp.Compile(*excludeTagFilter)
		if err != nil {
			log.Fatalf("Error parsing --exclude-tags regex: %q: %s", *excludeTagFilter, err)
		}
	}
	if *includeImageFilter != "" {
		includeImageRegex, err = regexp.Compile(*includeImageFilter)
		if err != nil {
			log.Fatalf("Error parsing --include-images regex: %q: %s", *includeImageFilter, err)
		}
	}
	if *excludeImageFilter != "" {
		excludeImageRegex, err = regexp.Compile(*excludeImageFilter)
		if err != nil {
			log.Fatalf("Error parsing --exclude-images regex: %q: %s", *excludeImageFilter, err)
		}
	}

	repoId := fmt.Sprintf("projects/%s/locations/%s/repositories/%s", *project, *location, *repo)
	fmt.Fprintf(os.Stderr, "Analyzing %s ...\n", repoId)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := gar.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	images := client.ListDockerImages(ctx, &garproto.ListDockerImagesRequest{Parent: repoId})

	imageStats := make(map[string]int64)
	for {
		img, err := images.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		name := strings.Split(img.GetUri(), "@")[0]

		// image filters
		if includeImageRegex != nil && !includeImageRegex.MatchString(name) {
			continue
		}
		if excludeImageRegex != nil && excludeImageRegex.MatchString(name) {
			continue
		}

		// tag filters
		if includeTagRegex != nil && !matchesAnyTag(img.GetTags(), includeTagRegex) {
			continue
		}
		if excludeTagRegex != nil && matchesAnyTag(img.GetTags(), excludeTagRegex) {
			continue
		}

		imageStats[name] += int64(img.GetImageSizeBytes())
	}

	var stats []image
	for name, size := range imageStats {
		stats = append(stats, image{name, size})
	}

	// sort by size, largest first
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].size > stats[j].size
	})

	tot := int64(0)
	for _, img := range stats {
		tot += int64(img.size)
		fmt.Printf("%-12s\t%s\n", printUnits(img.size), img.name)
	}
	fmt.Printf("%-12s\t.\n", printUnits(tot))
}

func printUnits(bytes int64) string {
	switch *format {
	case "kib":
		return fmt.Sprintf("%.2f KiB", toKiB(bytes))
	case "kb":
		return fmt.Sprintf("%.2f KB", toKB(bytes))
	case "mib":
		return fmt.Sprintf("%.2f MiB", toMiB(bytes))
	case "mb":
		return fmt.Sprintf("%.2f MB", toMB(bytes))
	case "gib":
		return fmt.Sprintf("%.2f GiB", toGiB(bytes))
	case "gb":
		return fmt.Sprintf("%.2f GB", toGB(bytes))
	default: // bytes
		return fmt.Sprintf("%d bytes", bytes)
	}
}

func toKB(bytes int64) float64 {
	return float64(bytes) / 1000.0 // 1 KB = 1000 bytes
}

func toKiB(bytes int64) float64 {
	return float64(bytes) / 1024.0 // 1 KiB = 1024 bytes
}

func toMB(bytes int64) float64 {
	return float64(bytes) / 1000000.0 // 1 MB = 1000^2 bytes
}

func toMiB(bytes int64) float64 {
	return float64(bytes) / (1 << 20) // 1 MiB = 2^20 bytes
}

func toGB(bytes int64) float64 {
	return float64(bytes) / 1e9 // 1 GB = 1,000,000,000 bytes
}

func toGiB(bytes int64) float64 {
	return float64(bytes) / (1 << 30) // 1 GiB = 2^30 bytes
}

func matchesAnyTag(tags []string, regex *regexp.Regexp) bool {
	for _, tag := range tags {
		if regex.MatchString(tag) {
			return true
		}
	}
	return false
}
