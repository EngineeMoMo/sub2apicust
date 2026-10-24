package service

import (
	"context"
	"fmt"
)

type CustomKiroRefreshRepository interface {
	UpdateKiroOAuthCredentialsIfUnchanged(context.Context, int64, map[string]any, *int64, map[string]any) (bool, error)
}

func (api *OAuthRefreshAPI) persistCustomKiroRotation(ctx context.Context, account *Account, credentials map[string]any) (*Account, bool, error) {
	repository, ok := api.accountRepo.(CustomKiroRefreshRepository)
	if !ok {
		return nil, false, &providerCycleContainmentRefreshError{err: fmt.Errorf("kiro 凭据条件更新仓库未配置")}
	}
	applied, err := repository.UpdateKiroOAuthCredentialsIfUnchanged(ctx, account.ID, account.Credentials, account.ProxyID, credentials)
	if err != nil {
		return nil, false, &providerCycleContainmentRefreshError{err: fmt.Errorf("%w: Kiro 刷新凭据保存失败", errOAuthRefreshCredentialPersist)}
	}
	latest, err := api.accountRepo.GetByID(ctx, account.ID)
	if err != nil || latest == nil {
		return nil, false, &providerCycleContainmentRefreshError{err: fmt.Errorf("kiro 刷新后读取持久化凭据失败")}
	}
	return latest, applied, nil
}
