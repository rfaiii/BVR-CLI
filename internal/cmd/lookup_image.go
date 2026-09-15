package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/richavery/bvr-cli/internal/catalog"
	"github.com/richavery/bvr-cli/internal/imagelookup"
	"github.com/spf13/cobra"
)

var lookupImageCmd = &cobra.Command{
	Use:   "lookup-image [image-path]",
	Short: "Identify a handbag, purse, sunglass, or accessory from a photo using local AI",
	Long: `Identify a handbag, purse, sunglass, or accessory from a photo using a local
Ollama vision model and a local product catalog.

The image is sent to a local vision model (default: qwen2.5vl:3b) which
extracts observable attributes as structured JSON. The extracted attributes
are then matched against a local SQLite catalog using FTS text search and
optional embedding similarity.

Image-based identification is NOT authentication.`,
	Example: `  # Identify an item from a photo
  bvr lookup-image ./photos/handbag-front.jpg

  # Use a custom catalog database
  BVR_CATALOG_DB=./data/my-catalog.db bvr lookup-image ./photos/sunglasses.jpg

  # Override the vision model
  BVR_VISION_MODEL=moondream bvr lookup-image ./photos/unknown-tote.jpg`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imagePath := args[0]

		cfg := imagelookup.FromEnv()

		// Open the catalog database (if it exists).
		repo, err := catalog.Open(cfg.CatalogDB)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not open catalog database %q: %v\n", cfg.CatalogDB, err)
			fmt.Fprintln(os.Stderr, "  Candidates will not be retrieved. Create a catalog to enable matching.")
		} else {
			defer repo.Close()
		}

		// Wire up the service.
		svc := imagelookup.NewService(cfg)
		if repo != nil {
			svc.Catalog = repo
			svc.Cache = repo
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
		defer cancel()

		fmt.Fprintln(os.Stderr, "Checking Ollama...")
		if err := svc.Vision.Health(ctx); err != nil {
			return fmt.Errorf("Ollama is not running: %w\n  Start it with: ollama serve", err)
		}

		fmt.Fprintln(os.Stderr, "Loading vision model (first call may be slow)...")
		if err := svc.Vision.LoadModel(ctx, cfg.VisionModel); err != nil {
			return fmt.Errorf("load vision model %q: %w", cfg.VisionModel, err)
		}

		fmt.Fprintln(os.Stderr, "Verifying image...")
		if err := imagelookup.ValidateImage(imagePath, cfg.MaxImageMB*1024*1024); err != nil {
			return fmt.Errorf("invalid image %q: %w", imagePath, err)
		}

		fmt.Fprintln(os.Stderr, "Inspecting logos, materials, hardware, and design cues...")
		result, err := svc.Lookup(ctx, imagePath)
		if err != nil {
			return fmt.Errorf("lookup failed: %w", err)
		}

		fmt.Print(imagelookup.RenderResult(result))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lookupImageCmd)
}
