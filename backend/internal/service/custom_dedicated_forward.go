package service

import "context"

func (s *OpenAIGatewayService) CheckCustomDedicatedAccount(ctx context.Context, account *Account) error {
	return s.customDedicated.Check(ctx, account, nil)
}

func (s *GatewayService) CheckCustomDedicatedAccount(ctx context.Context, account *Account) error {
	return s.customDedicated.Check(ctx, account, nil)
}
