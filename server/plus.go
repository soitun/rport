package chserver

import (
	"errors"
	"fmt"

	rportplus "github.com/cloudradar-monitoring/rport/plus"
	"github.com/cloudradar-monitoring/rport/plus/capabilities/oauth"
	"github.com/cloudradar-monitoring/rport/plus/capabilities/status"
	"github.com/cloudradar-monitoring/rport/share/files"
	"github.com/cloudradar-monitoring/rport/share/logger"
)

var (
	ErrPlusNotEnabled = errors.New("rport-plus not enabled")
)

// EnablePlusIfLicensed will initialize a new plus manager and request registration of the desired
// capabilities
func EnablePlusIfLicensed(cfg *Config, filesAPI files.FileAPI) (plusManager rportplus.Manager, err error) {
	logger := logger.NewLogger("rport-plus", cfg.Logging.LogOutput, cfg.Logging.LogLevel)

	if !cfg.PlusEnabled() {
		logger.Infof("not enabled")
		return nil, ErrPlusNotEnabled
	}

	plusManager, err = rportplus.NewPlusManager(&cfg.PlusConfig, logger, filesAPI)
	if err != nil {
		return nil, err
	}
	logger.Infof("plus manager initialized")

	err = RegisterPlusCapabilities(plusManager, cfg, logger)
	if err != nil {
		return nil, err
	}
	return plusManager, nil
}

// RegisterPluginCapabilitities registers the rport-plus additional capabilities.
// All plus capabilities must be added here.
func RegisterPlusCapabilities(plusManager rportplus.Manager, cfg *Config, logger *logger.Logger) (err error) {
	if cfg.PlusOAuthEnabled() {

		_, err := plusManager.RegisterCapability(rportplus.PlusOAuthCapability, &oauth.Capability{
			Config: cfg.PlusConfig.OAuthConfig,
			Logger: logger,
		})
		if err != nil {
			return fmt.Errorf("unable to register oauth plugin capability: %w", err)
		}

		// now validate the registered capability using the capability itself
		v := plusManager.GetConfigValidator(rportplus.PlusOAuthCapability)
		if v != nil {
			err = v.ValidateConfig()
			if err != nil {
				return fmt.Errorf("invalid oauth configuration: %w", err)
			}
		}

		logger.Infof("oauth capability registered")
	}

	// always register the plus status capability
	_, err = plusManager.RegisterCapability(rportplus.PlusStatusCapability, &status.Capability{
		Config: nil,
		Logger: logger,
	})
	if err != nil {
		return fmt.Errorf("unable to register plus status capability: %w", err)
	}
	logger.Infof("plus status capability registered")

	return nil
}