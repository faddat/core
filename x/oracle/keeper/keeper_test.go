package keeper

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	core "github.com/terra-money/core/types"
	"github.com/terra-money/core/x/oracle/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestExchangeRate(t *testing.T) {
	input := CreateTestInput(t)

	cnyExchangeRate := sdk.NewDecWithPrec(839, int64(OracleDecPrecision)).MulInt64(core.MicroUnit)
	gbpExchangeRate := sdk.NewDecWithPrec(4995, int64(OracleDecPrecision)).MulInt64(core.MicroUnit)
	krwExchangeRate := sdk.NewDecWithPrec(2838, int64(OracleDecPrecision)).MulInt64(core.MicroUnit)
	lunaExchangeRate := sdk.NewDecWithPrec(3282384, int64(OracleDecPrecision)).MulInt64(core.MicroUnit)

	// Set & get rates
	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, core.MicroCNYDenom, cnyExchangeRate)
	rate, err := input.OracleKeeper.GetLunaExchangeRate(input.Ctx, core.MicroCNYDenom)
	require.NoError(t, err)
	require.Equal(t, cnyExchangeRate, rate)

	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, core.MicroGBPDenom, gbpExchangeRate)
	rate, err = input.OracleKeeper.GetLunaExchangeRate(input.Ctx, core.MicroGBPDenom)
	require.NoError(t, err)
	require.Equal(t, gbpExchangeRate, rate)

	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, core.MicroKRWDenom, krwExchangeRate)
	rate, err = input.OracleKeeper.GetLunaExchangeRate(input.Ctx, core.MicroKRWDenom)
	require.NoError(t, err)
	require.Equal(t, krwExchangeRate, rate)

	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, core.MicroLunaDenom, lunaExchangeRate)
	rate, _ = input.OracleKeeper.GetLunaExchangeRate(input.Ctx, core.MicroLunaDenom)
	require.Equal(t, sdk.OneDec(), rate)

	input.OracleKeeper.DeleteLunaExchangeRate(input.Ctx, core.MicroKRWDenom)
	_, err = input.OracleKeeper.GetLunaExchangeRate(input.Ctx, core.MicroKRWDenom)
	require.Error(t, err)

	numExchangeRates := 0
	handler := func(denom string, exchangeRate sdk.Dec) (stop bool) {
		numExchangeRates = numExchangeRates + 1
		return false
	}
	input.OracleKeeper.IterateLunaExchangeRates(input.Ctx, handler)

	require.True(t, numExchangeRates == 3)
}

func TestIterateLunaExchangeRates(t *testing.T) {
	input := CreateTestInput(t)

	cnyExchangeRate := sdk.NewDecWithPrec(839, int64(OracleDecPrecision)).MulInt64(core.MicroUnit)
	gbpExchangeRate := sdk.NewDecWithPrec(4995, int64(OracleDecPrecision)).MulInt64(core.MicroUnit)
	krwExchangeRate := sdk.NewDecWithPrec(2838, int64(OracleDecPrecision)).MulInt64(core.MicroUnit)
	lunaExchangeRate := sdk.NewDecWithPrec(3282384, int64(OracleDecPrecision)).MulInt64(core.MicroUnit)

	// Set & get rates
	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, core.MicroCNYDenom, cnyExchangeRate)
	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, core.MicroGBPDenom, gbpExchangeRate)
	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, core.MicroKRWDenom, krwExchangeRate)
	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, core.MicroLunaDenom, lunaExchangeRate)

	input.OracleKeeper.IterateLunaExchangeRates(input.Ctx, func(denom string, rate sdk.Dec) (stop bool) {
		switch denom {
		case core.MicroCNYDenom:
			require.Equal(t, cnyExchangeRate, rate)
		case core.MicroGBPDenom:
			require.Equal(t, gbpExchangeRate, rate)
		case core.MicroKRWDenom:
			require.Equal(t, krwExchangeRate, rate)
		case core.MicroLunaDenom:
			require.Equal(t, lunaExchangeRate, rate)
		}
		return false
	})

}

func TestRewardPool(t *testing.T) {
	input := CreateTestInput(t)

	fees := sdk.NewCoins(sdk.NewCoin(core.MicroSDRDenom, sdk.NewInt(1000)))
	acc := input.AccountKeeper.GetModuleAccount(input.Ctx, types.ModuleName)
	err := FundAccount(input, acc.GetAddress(), fees)
	if err != nil {
		panic(err) // never occurs
	}

	KFees := input.OracleKeeper.GetRewardPool(input.Ctx)
	require.Equal(t, fees, KFees)
}

func TestParams(t *testing.T) {
	input := CreateTestInput(t)

	// Test default params setting
	input.OracleKeeper.SetParams(input.Ctx, types.DefaultParams())
	params := input.OracleKeeper.GetParams(input.Ctx)
	require.NotNil(t, params)

	// Test custom params setting
	votePeriod := uint64(10)
	voteThreshold := sdk.NewDecWithPrec(33, 2)
	oracleRewardBand := sdk.NewDecWithPrec(1, 2)
	rewardDistributionWindow := uint64(10000000000000)
	slashFraction := sdk.NewDecWithPrec(1, 2)
	slashWindow := uint64(1000)
	minValidPerWindow := sdk.NewDecWithPrec(1, 4)
	whitelist := types.DenomList{
		{Name: core.MicroSDRDenom, TobinTax: types.DefaultTobinTax},
		{Name: core.MicroKRWDenom, TobinTax: types.DefaultTobinTax},
	}

	// Should really test validateParams, but skipping because obvious
	newParams := types.Params{
		VotePeriod:               votePeriod,
		VoteThreshold:            voteThreshold,
		RewardBand:               oracleRewardBand,
		RewardDistributionWindow: rewardDistributionWindow,
		Whitelist:                whitelist,
		SlashFraction:            slashFraction,
		SlashWindow:              slashWindow,
		MinValidPerWindow:        minValidPerWindow,
	}
	input.OracleKeeper.SetParams(input.Ctx, newParams)

	storedParams := input.OracleKeeper.GetParams(input.Ctx)
	require.NotNil(t, storedParams)
	require.Equal(t, storedParams, newParams)
}

func TestFeederDelegation(t *testing.T) {
	input := CreateTestInput(t)

	// Test default getters and setters
	delegate := input.OracleKeeper.GetFeederDelegation(input.Ctx, ValAddrs[0])
	require.Equal(t, Addrs[0], delegate)

	input.OracleKeeper.SetFeederDelegation(input.Ctx, ValAddrs[0], Addrs[1])
	delegate = input.OracleKeeper.GetFeederDelegation(input.Ctx, ValAddrs[0])
	require.Equal(t, Addrs[1], delegate)
}

func TestIterateFeederDelegations(t *testing.T) {
	input := CreateTestInput(t)

	// Test default getters and setters
	delegate := input.OracleKeeper.GetFeederDelegation(input.Ctx, ValAddrs[0])
	require.Equal(t, Addrs[0], delegate)

	input.OracleKeeper.SetFeederDelegation(input.Ctx, ValAddrs[0], Addrs[1])

	var delegators []sdk.ValAddress
	var delegates []sdk.AccAddress
	input.OracleKeeper.IterateFeederDelegations(input.Ctx, func(delegator sdk.ValAddress, delegate sdk.AccAddress) (stop bool) {
		delegators = append(delegators, delegator)
		delegates = append(delegates, delegate)
		return false
	})

	require.Equal(t, 1, len(delegators))
	require.Equal(t, 1, len(delegates))
	require.Equal(t, ValAddrs[0], delegators[0])
	require.Equal(t, Addrs[1], delegates[0])
}

func TestMissCounter(t *testing.T) {
	input := CreateTestInput(t)

	// Test default getters and setters
	counter := input.OracleKeeper.GetMissCounter(input.Ctx, ValAddrs[0])
	require.Equal(t, uint64(0), counter)

	missCounter := uint64(10)
	input.OracleKeeper.SetMissCounter(input.Ctx, ValAddrs[0], missCounter)
	counter = input.OracleKeeper.GetMissCounter(input.Ctx, ValAddrs[0])
	require.Equal(t, missCounter, counter)

	input.OracleKeeper.DeleteMissCounter(input.Ctx, ValAddrs[0])
	counter = input.OracleKeeper.GetMissCounter(input.Ctx, ValAddrs[0])
	require.Equal(t, uint64(0), counter)
}

func TestIterateMissCounters(t *testing.T) {
	input := CreateTestInput(t)

	// Test default getters and setters
	counter := input.OracleKeeper.GetMissCounter(input.Ctx, ValAddrs[0])
	require.Equal(t, uint64(0), counter)

	missCounter := uint64(10)
	input.OracleKeeper.SetMissCounter(input.Ctx, ValAddrs[1], missCounter)

	var operators []sdk.ValAddress
	var missCounters []uint64
	input.OracleKeeper.IterateMissCounters(input.Ctx, func(delegator sdk.ValAddress, missCounter uint64) (stop bool) {
		operators = append(operators, delegator)
		missCounters = append(missCounters, missCounter)
		return false
	})

	require.Equal(t, 1, len(operators))
	require.Equal(t, 1, len(missCounters))
	require.Equal(t, ValAddrs[1], operators[0])
	require.Equal(t, missCounter, missCounters[0])
}

func TestAggregatePrevoteAddDelete(t *testing.T) {
	input := CreateTestInput(t)

	hash := types.GetAggregateVoteHash("salt", "100ukrw,1000uusd", sdk.ValAddress(Addrs[0]))
	aggregatePrevote := types.NewAggregateExchangeRatePrevote(hash, sdk.ValAddress(Addrs[0]), 0)
	input.OracleKeeper.SetAggregateExchangeRatePrevote(input.Ctx, sdk.ValAddress(Addrs[0]), aggregatePrevote)

	KPrevote, err := input.OracleKeeper.GetAggregateExchangeRatePrevote(input.Ctx, sdk.ValAddress(Addrs[0]))
	require.NoError(t, err)
	require.Equal(t, aggregatePrevote, KPrevote)

	input.OracleKeeper.DeleteAggregateExchangeRatePrevote(input.Ctx, sdk.ValAddress(Addrs[0]))
	_, err = input.OracleKeeper.GetAggregateExchangeRatePrevote(input.Ctx, sdk.ValAddress(Addrs[0]))
	require.Error(t, err)
}

func TestAggregatePrevoteIterate(t *testing.T) {
	input := CreateTestInput(t)

	hash := types.GetAggregateVoteHash("salt", "100ukrw,1000uusd", sdk.ValAddress(Addrs[0]))
	aggregatePrevote1 := types.NewAggregateExchangeRatePrevote(hash, sdk.ValAddress(Addrs[0]), 0)
	input.OracleKeeper.SetAggregateExchangeRatePrevote(input.Ctx, sdk.ValAddress(Addrs[0]), aggregatePrevote1)

	hash2 := types.GetAggregateVoteHash("salt", "100ukrw,1000uusd", sdk.ValAddress(Addrs[1]))
	aggregatePrevote2 := types.NewAggregateExchangeRatePrevote(hash2, sdk.ValAddress(Addrs[1]), 0)
	input.OracleKeeper.SetAggregateExchangeRatePrevote(input.Ctx, sdk.ValAddress(Addrs[1]), aggregatePrevote2)

	i := 0
	bigger := bytes.Compare(Addrs[0], Addrs[1])
	input.OracleKeeper.IterateAggregateExchangeRatePrevotes(input.Ctx, func(voter sdk.ValAddress, p types.AggregateExchangeRatePrevote) (stop bool) {
		if (i == 0 && bigger == -1) || (i == 1 && bigger == 1) {
			require.Equal(t, aggregatePrevote1, p)
			require.Equal(t, voter.String(), p.Voter)
		} else {
			require.Equal(t, aggregatePrevote2, p)
			require.Equal(t, voter.String(), p.Voter)
		}

		i++
		return false
	})
}

func TestAggregateVoteAddDelete(t *testing.T) {
	input := CreateTestInput(t)

	aggregateVote := types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{
		{Denom: "foo", ExchangeRate: sdk.NewDec(-1)},
		{Denom: "foo", ExchangeRate: sdk.NewDec(0)},
		{Denom: "foo", ExchangeRate: sdk.NewDec(1)},
	}, sdk.ValAddress(Addrs[0]))
	input.OracleKeeper.SetAggregateExchangeRateVote(input.Ctx, sdk.ValAddress(Addrs[0]), aggregateVote)

	KVote, err := input.OracleKeeper.GetAggregateExchangeRateVote(input.Ctx, sdk.ValAddress(Addrs[0]))
	require.NoError(t, err)
	require.Equal(t, aggregateVote, KVote)

	input.OracleKeeper.DeleteAggregateExchangeRateVote(input.Ctx, sdk.ValAddress(Addrs[0]))
	_, err = input.OracleKeeper.GetAggregateExchangeRateVote(input.Ctx, sdk.ValAddress(Addrs[0]))
	require.Error(t, err)
}

func TestAggregateVoteIterate(t *testing.T) {
	input := CreateTestInput(t)

	aggregateVote1 := types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{
		{Denom: "foo", ExchangeRate: sdk.NewDec(-1)},
		{Denom: "foo", ExchangeRate: sdk.NewDec(0)},
		{Denom: "foo", ExchangeRate: sdk.NewDec(1)},
	}, sdk.ValAddress(Addrs[0]))
	input.OracleKeeper.SetAggregateExchangeRateVote(input.Ctx, sdk.ValAddress(Addrs[0]), aggregateVote1)

	aggregateVote2 := types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{
		{Denom: "foo", ExchangeRate: sdk.NewDec(-1)},
		{Denom: "foo", ExchangeRate: sdk.NewDec(0)},
		{Denom: "foo", ExchangeRate: sdk.NewDec(1)},
	}, sdk.ValAddress(Addrs[1]))
	input.OracleKeeper.SetAggregateExchangeRateVote(input.Ctx, sdk.ValAddress(Addrs[1]), aggregateVote2)

	i := 0
	bigger := bytes.Compare(address.MustLengthPrefix(Addrs[0]), address.MustLengthPrefix(Addrs[1]))
	input.OracleKeeper.IterateAggregateExchangeRateVotes(input.Ctx, func(voter sdk.ValAddress, p types.AggregateExchangeRateVote) (stop bool) {
		if (i == 0 && bigger == -1) || (i == 1 && bigger == 1) {
			require.Equal(t, aggregateVote1, p)
			require.Equal(t, voter.String(), p.Voter)
		} else {
			require.Equal(t, aggregateVote2, p)
			require.Equal(t, voter.String(), p.Voter)
		}

		i++
		return false
	})
}

func TestTobinTaxGetSet(t *testing.T) {
	input := CreateTestInput(t)

	tobinTaxes := map[string]sdk.Dec{
		core.MicroSDRDenom: sdk.NewDec(1),
		core.MicroUSDDenom: sdk.NewDecWithPrec(1, 3),
		core.MicroKRWDenom: sdk.NewDecWithPrec(123, 3),
		core.MicroMNTDenom: sdk.NewDecWithPrec(1423, 4),
	}

	for denom, tobinTax := range tobinTaxes {
		input.OracleKeeper.SetTobinTax(input.Ctx, denom, tobinTax)
		factor, err := input.OracleKeeper.GetTobinTax(input.Ctx, denom)
		require.NoError(t, err)
		require.Equal(t, tobinTaxes[denom], factor)
	}

	input.OracleKeeper.IterateTobinTaxes(input.Ctx, func(denom string, tobinTax sdk.Dec) (stop bool) {
		require.Equal(t, tobinTaxes[denom], tobinTax)
		return false
	})

	input.OracleKeeper.ClearTobinTaxes(input.Ctx)
	for denom := range tobinTaxes {
		_, err := input.OracleKeeper.GetTobinTax(input.Ctx, denom)
		require.Error(t, err)
	}
}

func TestValidateFeeder(t *testing.T) {
	// initial setup
	input := CreateTestInput(t)
	addr, val := ValAddrs[0], ValPubKeys[0]
	addr1, val1 := ValAddrs[1], ValPubKeys[1]
	amt := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	sh := staking.NewHandler(input.StakingKeeper)
	ctx := input.Ctx

	// Validator created
	_, err := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.NoError(t, err)
	_, err = sh(ctx, NewTestMsgCreateValidator(addr1, val1, amt))
	require.NoError(t, err)
	staking.EndBlocker(ctx, input.StakingKeeper)

	require.Equal(
		t, input.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(input.StakingKeeper.GetParams(ctx).BondDenom, InitTokens.Sub(amt))),
	)
	require.Equal(t, amt, input.StakingKeeper.Validator(ctx, addr).GetBondedTokens())
	require.Equal(
		t, input.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr1)),
		sdk.NewCoins(sdk.NewCoin(input.StakingKeeper.GetParams(ctx).BondDenom, InitTokens.Sub(amt))),
	)
	require.Equal(t, amt, input.StakingKeeper.Validator(ctx, addr1).GetBondedTokens())

	require.NoError(t, input.OracleKeeper.ValidateFeeder(input.Ctx, sdk.AccAddress(addr), sdk.ValAddress(addr)))
	require.NoError(t, input.OracleKeeper.ValidateFeeder(input.Ctx, sdk.AccAddress(addr1), sdk.ValAddress(addr1)))

	// delegate works
	input.OracleKeeper.SetFeederDelegation(input.Ctx, sdk.ValAddress(addr), sdk.AccAddress(addr1))
	require.NoError(t, input.OracleKeeper.ValidateFeeder(input.Ctx, sdk.AccAddress(addr1), sdk.ValAddress(addr)))
	require.Error(t, input.OracleKeeper.ValidateFeeder(input.Ctx, sdk.AccAddress(Addrs[2]), sdk.ValAddress(addr)))

	// only active validators can do oracle votes
	validator, found := input.StakingKeeper.GetValidator(input.Ctx, sdk.ValAddress(addr))
	require.True(t, found)
	validator.Status = stakingtypes.Unbonded
	input.StakingKeeper.SetValidator(input.Ctx, validator)
	require.Error(t, input.OracleKeeper.ValidateFeeder(input.Ctx, sdk.AccAddress(addr1), sdk.ValAddress(addr)))
}
