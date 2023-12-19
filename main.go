package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	gar "cloud.google.com/go/artifactregistry/apiv1"
	garproto "cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"google.golang.org/api/iterator"
)

var (
	project  = flag.String("project", "", "Project containing the Google Artifact Registry")
	location = flag.String("location", "", "Location of the Google Artifact Registry")
	repo     = flag.String("repo", "", "Repository containing the Docker images")
	format   = flag.String("format", "gib", "Output format: bytes, kib, kb, mib, mb, gib, gb")
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

	repoId := fmt.Sprintf("projects/%s/locations/%s/repositories/%s", *project, *location, *repo)
	fmt.Fprintf(os.Stderr, "Analyzing %s ...\n", repoId)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := gar.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	req := &garproto.ListDockerImagesRequest{Parent: repoId}
	images := client.ListDockerImages(ctx, req)

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
