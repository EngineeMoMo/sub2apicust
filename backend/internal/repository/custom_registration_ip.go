package repository

import (
	"context"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/setting"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type registrationIPClaim struct {
	address string
	limited bool
}

func prepareRegistrationIP(ctx context.Context, client *dbent.Client) (*registrationIPClaim, error) {
	address, present := service.RegistrationIPFromContext(ctx)
	if !present {
		return nil, nil
	}
	config, err := client.Setting.Query().Where(setting.KeyEQ(service.SettingKeyRegistrationIPLimitEnabled)).Only(ctx)
	if err != nil && !dbent.IsNotFound(err) {
		return nil, service.ErrRegistrationIPUnavailable.WithCause(err)
	}
	limited := true
	if config != nil {
		switch config.Value {
		case "true":
		case "false":
			limited = false
		default:
			return nil, service.ErrRegistrationIPUnavailable
		}
	}
	if address == "" {
		if limited {
			return nil, service.ErrRegistrationIPUnavailable
		}
		return nil, nil
	}
	return &registrationIPClaim{address: address, limited: limited}, nil
}

func claimRegistrationIP(ctx context.Context, client *dbent.Client, claim *registrationIPClaim, userID int64) error {
	if claim == nil {
		return nil
	}
	result, err := client.ExecContext(ctx, "INSERT INTO custom_registration_ips (ip, user_id) VALUES ($1, $2) ON CONFLICT (ip) DO NOTHING", claim.address, userID)
	if err != nil {
		return service.ErrRegistrationIPUnavailable.WithCause(err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return service.ErrRegistrationIPUnavailable.WithCause(err)
	}
	if claim.limited && count == 0 {
		return service.ErrRegistrationIPLimit
	}
	return nil
}

type registrationRollbackContextKey struct{}

func (r *userRepository) RollbackRegistration(ctx context.Context, userID int64) error {
	return r.Delete(context.WithValue(ctx, registrationRollbackContextKey{}, true), userID)
}

func releaseRolledBackRegistrationIP(ctx context.Context, client *dbent.Client, userID int64) error {
	rollback, _ := ctx.Value(registrationRollbackContextKey{}).(bool)
	if !rollback {
		return nil
	}
	_, err := client.ExecContext(ctx, "DELETE FROM custom_registration_ips WHERE user_id = $1", userID)
	if err != nil {
		return service.ErrRegistrationIPUnavailable.WithCause(err)
	}
	return nil
}
