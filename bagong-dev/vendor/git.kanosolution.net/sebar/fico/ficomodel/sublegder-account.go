package ficomodel

import "git.kanosolution.net/sebar/tenantcore/tenantcoremodel"

type SubledgerAccount struct {
	AccountType tenantcoremodel.TrxModule
	AccountID   string
}

func NewSubAccount(acType tenantcoremodel.TrxModule, acID string) SubledgerAccount {
	return SubledgerAccount{AccountType: acType, AccountID: acID}
}

func (s *SubledgerAccount) IsValid(kind tenantcoremodel.TrxModule) bool {
	return s.AccountType == kind && s.AccountID != ""
}
