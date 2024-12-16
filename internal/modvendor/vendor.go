package modvendor

import (
	"context"
	"fmt"
	"log/slog"

	cp "github.com/otiai10/copy"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/api/config"
	"github.com/easyp-tech/easyp/internal/api/factories"
	"github.com/easyp-tech/easyp/internal/mod/models"
)

func Run(cliCtx *cli.Context) error {
	ctx := context.Background()
	slog.Info("Start vendor")

	cfg, err := config.ReadConfig(cliCtx)
	if err != nil {
		return fmt.Errorf("ReadConfig: %w", err)
	}

	moduleReflect, err := factories.NewModuleReflect()
	if err != nil {
		return fmt.Errorf("factories.NewModuleReflect: %w", err)
	}

	vendorPath := "easyp_vendor"

	for _, dep := range cfg.Deps {
		modulePath, err := moduleReflect.GetModulePath(ctx, dep)
		if err != nil {
			return fmt.Errorf("GetModulePath: %w", err)
		}
		module := models.NewModule(dep)
		slog.Info("dep", "dep", dep, "path", modulePath, "module", module)

		if err := cp.Copy(modulePath, vendorPath); err != nil {
			return fmt.Errorf("Copy: %w", err)
		}
		//if err := cp.Copy()
		//depDir := path.Join(vendorPath, module.Name)
		//if err := os.MkdirAll(depDir, 0766); err != nil {
		//	return fmt.Errorf("os.MkdirAll: %w", err)
		//}
	}

	_ = cfg
	_ = moduleReflect

	return nil
}
