package perp_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/NibiruChain/nibiru/collections/keys"

	simapp2 "github.com/NibiruChain/nibiru/simapp"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/NibiruChain/nibiru/x/common"
	"github.com/NibiruChain/nibiru/x/perp"
	"github.com/NibiruChain/nibiru/x/perp/types"
	"github.com/NibiruChain/nibiru/x/testutil/sample"
)

func TestGenesis(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := simapp2.NewTestNibiruApp(false)
		ctxUncached := app.NewContext(false, tmproto.Header{})
		ctx, _ := ctxUncached.CacheContext()

		// create some params
		app.PerpKeeper.SetParams(ctx, types.Params{
			Stopped:                 true,
			FeePoolFeeRatio:         sdk.MustNewDecFromStr("0.00001"),
			EcosystemFundFeeRatio:   sdk.MustNewDecFromStr("0.000005"),
			LiquidationFeeRatio:     sdk.MustNewDecFromStr("0.000007"),
			PartialLiquidationRatio: sdk.MustNewDecFromStr("0.00001"),
			TwapLookbackWindow:      15 * time.Minute,
		})

		// create some positions
		for i := int64(0); i < 100; i++ {
			p := types.Position{
				TraderAddress:                       sample.AccAddress().String(),
				Pair:                                common.PairGovStable,
				Size_:                               sdk.NewDec(i + 1),
				Margin:                              sdk.NewDec(i * 2),
				OpenNotional:                        sdk.NewDec(i * 100),
				LastUpdateCumulativePremiumFraction: sdk.NewDec(5 * 100),
				BlockNumber:                         i,
			}
			app.PerpKeeper.Positions.Insert(ctx, keys.Join(p.Pair, keys.String(p.TraderAddress)), p)
		}

		// create some prepaid bad debt
		for i := 0; i < 10; i++ {
			denom := fmt.Sprintf("%d", i)
			amount := sdk.NewInt(int64(i))
			app.PerpKeeper.PrepaidBadDebt.Insert(ctx, keys.String(denom), types.PrepaidBadDebt{
				Denom:  denom,
				Amount: amount,
			})
		}

		// whitelist some addrs
		for i := 0; i < 5; i++ {
			app.PerpKeeper.Whitelist.Insert(ctx, keys.String(sample.AccAddress().String()))
		}

		// export genesis
		genState := perp.ExportGenesis(ctx, app.PerpKeeper)
		// create new context and init genesis
		ctx, _ = ctxUncached.CacheContext()
		perp.InitGenesis(ctx, app.PerpKeeper, *genState)

		// export again to ensure they match
		genStateAfterInit := perp.ExportGenesis(ctx, app.PerpKeeper)
		require.Equal(t, genState, genStateAfterInit)
	})
}
