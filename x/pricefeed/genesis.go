package pricefeed

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/NibiruChain/nibiru/x/pricefeed/keeper"
	"github.com/NibiruChain/nibiru/x/pricefeed/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	k.ActivePairsStore().
		AddActivePairs(ctx, genState.Params.Pairs)
	k.WhitelistOracles(ctx,
		[]sdk.AccAddress{
			sdk.MustAccAddressFromBech32("nibi1pzd5e402eld9kcc3h78tmfrm5rpzlzk6hnxkvu"),
			sdk.MustAccAddressFromBech32("nibi15cdcxznuwpuk5hw7t678wpyesy78kwy00qcesa"),
			sdk.MustAccAddressFromBech32("nibi1x5zknk8va44th5vjpg0fagf0lxx0rvurpmp8gs"),
		},
	)

	// If posted prices are not expired, set them in the store
	for _, pp := range genState.PostedPrices {
		if pp.Expiry.After(ctx.BlockTime()) {
			oracle := sdk.MustAccAddressFromBech32(pp.Oracle)
			_, err := k.PostRawPrice(ctx, oracle, pp.PairID, pp.Price, pp.Expiry)
			if err != nil {
				panic(err)
			}
		} else {
			panic(fmt.Errorf("failed to post prices for pair %v", pp.PairID))
		}
	}
	params := k.GetParams(ctx)

	// Set the current price (if any) based on what's now in the store
	for _, pair := range params.Pairs {
		if !k.ActivePairsStore().Get(ctx, pair) {
			continue
		}
		postedPrices := k.GetRawPrices(ctx, pair.String())

		if len(postedPrices) == 0 {
			continue
		}
		err := k.GatherRawPrices(ctx, pair.Token0, pair.Token1)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	var postedPrices []types.PostedPrice
	for _, assetPair := range k.GetPairs(ctx) {
		pp := k.GetRawPrices(ctx, assetPair.String())
		postedPrices = append(postedPrices, pp...)
	}
	genesis.PostedPrices = postedPrices

	return genesis
}
