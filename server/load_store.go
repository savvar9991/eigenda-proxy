package server

import (
	"context"
	"fmt"

	"github.com/Layr-Labs/eigenda-proxy/metrics"
	"github.com/Layr-Labs/eigenda-proxy/store"
	"github.com/Layr-Labs/eigenda-proxy/store/generated_key/eigenda"
	"github.com/Layr-Labs/eigenda-proxy/store/generated_key/memstore"
	"github.com/Layr-Labs/eigenda-proxy/store/precomputed_key/redis"
	"github.com/Layr-Labs/eigenda-proxy/store/precomputed_key/s3"
	"github.com/Layr-Labs/eigenda-proxy/verify"
	"github.com/Layr-Labs/eigenda/api/clients"
	"github.com/ethereum/go-ethereum/log"
)

// TODO - create structured abstraction for dependency injection vs. overloading stateless functions

// populateTargets ... creates a list of storage backends based on the provided target strings
func populateTargets(targets []string, s3 store.PrecomputedKeyStore, redis *redis.Store) []store.PrecomputedKeyStore {
	stores := make([]store.PrecomputedKeyStore, len(targets))

	for i, f := range targets {
		b := store.StringToBackendType(f)

		switch b {
		case store.RedisBackendType:
			if redis == nil {
				panic(fmt.Sprintf("Redis backend is not configured but specified in targets: %s", f))
			}
			stores[i] = redis

		case store.S3BackendType:
			if s3 == nil {
				panic(fmt.Sprintf("S3 backend is not configured but specified in targets: %s", f))
			}
			stores[i] = s3

		case store.EigenDABackendType, store.MemoryBackendType:
			panic(fmt.Sprintf("Invalid target for fallback: %s", f))

		case store.Unknown:
			fallthrough

		default:
			panic(fmt.Sprintf("Unknown fallback target: %s", f))
		}
	}

	return stores
}

// LoadStoreManager ... creates storage backend clients and instruments them into a storage routing abstraction
func LoadStoreManager(ctx context.Context, cfg CLIConfig, log log.Logger, m metrics.Metricer) (store.IManager, error) {
	// create S3 backend store (if enabled)
	var err error
	var s3Store store.PrecomputedKeyStore
	var redisStore *redis.Store

	if cfg.EigenDAConfig.S3Config.Bucket != "" && cfg.EigenDAConfig.S3Config.Endpoint != "" {
		log.Info("Using S3 backend")
		s3Store, err = s3.NewS3(cfg.EigenDAConfig.S3Config)
		if err != nil {
			return nil, fmt.Errorf("failed to create S3 store: %w", err)
		}
	}

	if cfg.EigenDAConfig.RedisConfig.Endpoint != "" {
		log.Info("Using Redis backend")
		// create Redis backend store
		redisStore, err = redis.NewStore(&cfg.EigenDAConfig.RedisConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Redis store: %w", err)
		}
	}

	// create cert/data verification type
	daCfg := cfg.EigenDAConfig
	vCfg := daCfg.VerifierConfig

	verifier, err := verify.NewVerifier(&vCfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create verifier: %w", err)
	}

	if vCfg.VerifyCerts {
		log.Info("Certificate verification with Ethereum enabled")
	} else {
		log.Warn("Verification disabled")
	}

	// create EigenDA backend store
	var eigenDA store.GeneratedKeyStore
	if cfg.EigenDAConfig.MemstoreEnabled {
		log.Info("Using memstore backend for EigenDA")
		eigenDA, err = memstore.New(ctx, verifier, log, cfg.EigenDAConfig.MemstoreConfig)
	} else {
		var client *clients.EigenDAClient
		log.Info("Using EigenDA backend")
		client, err = clients.NewEigenDAClient(log.With("subsystem", "eigenda-client"), daCfg.EdaClientConfig)
		if err != nil {
			return nil, err
		}

		eigenDA, err = eigenda.NewStore(
			client,
			verifier,
			log,
			&eigenda.StoreConfig{
				MaxBlobSizeBytes:     cfg.EigenDAConfig.MemstoreConfig.MaxBlobSizeBytes,
				EthConfirmationDepth: cfg.EigenDAConfig.VerifierConfig.EthConfirmationDepth,
				StatusQueryTimeout:   cfg.EigenDAConfig.EdaClientConfig.StatusQueryTimeout,
			},
		)
	}

	if err != nil {
		return nil, err
	}

	// create secondary storage router
	fallbacks := populateTargets(cfg.EigenDAConfig.FallbackTargets, s3Store, redisStore)
	caches := populateTargets(cfg.EigenDAConfig.CacheTargets, s3Store, redisStore)
	secondary := store.NewSecondaryManager(log, m, caches, fallbacks)

	if secondary.Enabled() { // only spin-up go routines if secondary storage is enabled
		// NOTE: in the future the number of threads could be made configurable via env
		log.Debug("Starting secondary write loop(s)", "count", cfg.EigenDAConfig.AsyncPutWorkers)

		for i := 0; i < cfg.EigenDAConfig.AsyncPutWorkers; i++ {
			go secondary.WriteSubscriptionLoop(ctx)
		}
	}

	log.Info("Creating storage router", "eigenda backend type", eigenDA != nil, "s3 backend type", s3Store != nil)
	return store.NewManager(eigenDA, s3Store, log, secondary)
}
