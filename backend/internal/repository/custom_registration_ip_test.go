package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/setting"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func newRegistrationIPRepo(t *testing.T) (*userRepository, *dbent.Client) {
	t.Helper()
	repo, client := newUserEntRepo(t)
	migration, err := os.ReadFile("../../migrations/260_custom_registration_ips.sql")
	require.NoError(t, err)
	_, err = client.ExecContext(context.Background(), string(migration))
	require.NoError(t, err)
	return repo, client
}

func registrationIPUser(email string) *service.User {
	return &service.User{Email: email, Username: email, PasswordHash: "hash", Role: service.RoleUser, Status: service.StatusActive}
}

func setRegistrationIPLimit(t *testing.T, client *dbent.Client, value string) {
	t.Helper()
	ctx := context.Background()
	_, err := client.Setting.Delete().Where(setting.KeyEQ(service.SettingKeyRegistrationIPLimitEnabled)).Exec(ctx)
	require.NoError(t, err)
	_, err = client.Setting.Create().SetKey(service.SettingKeyRegistrationIPLimitEnabled).SetValue(value).Save(ctx)
	require.NoError(t, err)
}

func TestRegistrationIPRejectsDuplicateAcrossCreatePaths(t *testing.T) {
	repo, client := newRegistrationIPRepo(t)
	ctx := service.WithRegistrationIP(context.Background(), "192.0.2.1")
	require.NoError(t, repo.CreateWithEmailAliasGuard(ctx, registrationIPUser("first@gmail.com")))
	require.ErrorIs(t, repo.Create(ctx, registrationIPUser("oauth@gmail.com")), service.ErrRegistrationIPLimit)
	require.ErrorIs(t, repo.CreateWithEmailAliasGuardAndDomainLimit(ctx, registrationIPUser("user@business.example"), "business.example"), service.ErrRegistrationIPLimit)
	count, err := client.User.Query().Count(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.NoError(t, repo.Create(service.WithRegistrationIP(context.Background(), "192.0.2.2"), registrationIPUser("other@gmail.com")))
	require.NoError(t, repo.Create(context.Background(), registrationIPUser("admin-created@gmail.com")))
}

func TestRegistrationIPCanonicalAddresses(t *testing.T) {
	for _, addresses := range [][2]string{{"192.0.2.1", "::ffff:192.0.2.1"}, {"2001:db8::1", "2001:0db8:0:0:0:0:0:1"}} {
		t.Run(addresses[0], func(t *testing.T) {
			repo, _ := newRegistrationIPRepo(t)
			require.NoError(t, repo.Create(service.WithRegistrationIP(context.Background(), addresses[0]), registrationIPUser("first@gmail.com")))
			require.ErrorIs(t, repo.Create(service.WithRegistrationIP(context.Background(), addresses[1]), registrationIPUser("second@gmail.com")), service.ErrRegistrationIPLimit)
		})
	}
}

func TestRegistrationIPDisabledStillRecordsSources(t *testing.T) {
	repo, client := newRegistrationIPRepo(t)
	ctx := service.WithRegistrationIP(context.Background(), "192.0.2.1")
	setRegistrationIPLimit(t, client, "false")
	require.NoError(t, repo.Create(ctx, registrationIPUser("first@gmail.com")))
	require.NoError(t, repo.Create(ctx, registrationIPUser("second@gmail.com")))
	setRegistrationIPLimit(t, client, "true")
	require.ErrorIs(t, repo.Create(ctx, registrationIPUser("third@gmail.com")), service.ErrRegistrationIPLimit)
}

func TestRegistrationIPOuterRollbackReleasesClaim(t *testing.T) {
	repo, client := newRegistrationIPRepo(t)
	ctx := service.WithRegistrationIP(context.Background(), "192.0.2.1")
	tx, err := client.Tx(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.Create(dbent.NewTxContext(ctx, tx), registrationIPUser("rolledback@gmail.com")))
	require.NoError(t, tx.Rollback())
	require.NoError(t, repo.Create(ctx, registrationIPUser("retry@gmail.com")))
}

func TestRegistrationIPOrdinaryDeletionRetainsClaim(t *testing.T) {
	repo, _ := newRegistrationIPRepo(t)
	ctx := service.WithRegistrationIP(context.Background(), "192.0.2.1")
	user := registrationIPUser("first@gmail.com")
	require.NoError(t, repo.Create(ctx, user))
	require.NoError(t, repo.Delete(context.Background(), user.ID))
	require.ErrorIs(t, repo.Create(ctx, registrationIPUser("second@gmail.com")), service.ErrRegistrationIPLimit)
}

func TestRegistrationIPOAuthCompensationReleasesOnlyItsClaim(t *testing.T) {
	repo, client := newRegistrationIPRepo(t)
	ctx := service.WithRegistrationIP(context.Background(), "192.0.2.1")
	first := registrationIPUser("first@gmail.com")
	require.NoError(t, repo.Create(ctx, first))
	setRegistrationIPLimit(t, client, "false")
	second := registrationIPUser("second@gmail.com")
	require.NoError(t, repo.Create(ctx, second))
	require.NoError(t, repo.RollbackRegistration(ctx, second.ID))
	setRegistrationIPLimit(t, client, "true")
	require.ErrorIs(t, repo.Create(ctx, registrationIPUser("third@gmail.com")), service.ErrRegistrationIPLimit)
	require.NoError(t, repo.RollbackRegistration(ctx, first.ID))
	require.NoError(t, repo.Create(ctx, registrationIPUser("retry@gmail.com")))
}

func TestRegistrationIPMissingTableFailsClosedAndRollsBack(t *testing.T) {
	repo, client := newRegistrationIPRepo(t)
	_, err := client.ExecContext(context.Background(), "DROP TABLE custom_registration_ips")
	require.NoError(t, err)
	err = repo.Create(service.WithRegistrationIP(context.Background(), "192.0.2.1"), registrationIPUser("first@gmail.com"))
	require.ErrorIs(t, err, service.ErrRegistrationIPUnavailable)
	count, err := client.User.Query().Count(context.Background())
	require.NoError(t, err)
	require.Zero(t, count)
}

func TestRegistrationIPInvalidSourceOrSettingFailsClosed(t *testing.T) {
	for index, address := range []string{"", "invalid", "0.0.0.0", "::", "ff02::1", "fe80::1%eth0"} {
		t.Run(fmt.Sprintf("invalid-%d", index), func(t *testing.T) {
			repo, _ := newRegistrationIPRepo(t)
			err := repo.Create(service.WithRegistrationIP(context.Background(), address), registrationIPUser("first@gmail.com"))
			require.ErrorIs(t, err, service.ErrRegistrationIPUnavailable)
		})
	}
	repo, client := newRegistrationIPRepo(t)
	setRegistrationIPLimit(t, client, "invalid")
	err := repo.Create(service.WithRegistrationIP(context.Background(), "192.0.2.1"), registrationIPUser("first@gmail.com"))
	require.ErrorIs(t, err, service.ErrRegistrationIPUnavailable)
}

func TestRegistrationIPConcurrentRequestsOneAccount(t *testing.T) {
	repo, client := newRegistrationIPRepo(t)
	repo.sql.(*sql.DB).SetMaxOpenConns(1)
	ctx := service.WithRegistrationIP(context.Background(), "192.0.2.1")
	results := make(chan error, 12)
	var workers sync.WaitGroup
	for index := range 12 {
		workers.Add(1)
		go func(index int) {
			defer workers.Done()
			results <- repo.Create(ctx, registrationIPUser(fmt.Sprintf("user%d@gmail.com", index)))
		}(index)
	}
	workers.Wait()
	close(results)
	succeeded := 0
	for err := range results {
		if err == nil {
			succeeded++
		} else {
			require.ErrorIs(t, err, service.ErrRegistrationIPLimit)
		}
	}
	require.Equal(t, 1, succeeded)
	count, err := client.User.Query().Count(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
